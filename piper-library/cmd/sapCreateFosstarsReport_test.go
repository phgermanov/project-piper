//go:build unit
// +build unit

package cmd

import (
	"errors"
	"testing"
	"time"

	"fmt"
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/bmatcuk/doublestar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/fosstars"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/fosstars/mocks"
)

const npmStr = `{
  "name": "npm-dependency-check",
  "version": "1.0.0",
  "scripts": {
    "start": "node check"
  },
  "devDependencies": {
    "axios": "^0.17.1",
    "prompt": "1.0.0",
    "yargs": "^8.0.2"
  },
  "dependencies": {
    "read-package-json": "1.3.1",
    "npm-package-arg": "3.1.0",
    "commander": "2.6.0",
    "validator": "3.30.0"
  }
}`

const mavenStr = `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
  <modelVersion>4.0.0</modelVersion>
  <groupId>com.sap.phosphor</groupId>
  <artifactId>sbom-spring-starter-project</artifactId>
  <version>0.1.0-SNAPSHOT</version>
  <packaging>jar</packaging>
  <name>sBOM Starter: sbom-spring-starter-project</name>
  <description>Start project for sBOM Spring Boot</description>
  <properties>
    <java.version>1.8</java.version>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
    <maven-model.version>3.6.2</maven-model.version>
    <commons-io.version>2.8.0</commons-io.version>
  </properties>
  <dependencies>
<dependency>
      <groupId>commons-io</groupId>
      <artifactId>commons-io</artifactId>
      <version>${commons-io.version}</version>
    </dependency>
    <dependency>
      <groupId>org.apache.maven</groupId>
      <artifactId>maven-model</artifactId>
      <version>${maven-model.version}</version>
    </dependency>
  </dependencies>
</project>`

const dependencyGraphJsonStr = `{
  "graphName" : "sbom-spring-starter-project",
  "artifacts" : [ {
    "id" : "com.sap.phosphor:sbom-spring-starter-project:jar",
    "numericId" : 1,
    "groupId" : "com.sap.phosphor",
    "artifactId" : "sbom-spring-starter-project",
    "version" : "0.1.0-SNAPSHOT",
    "optional" : false,
    "scopes" : [ "compile" ],
    "types" : [ "jar" ]
  }, {
    "id" : "commons-io:commons-io:jar",
    "numericId" : 2,
    "groupId" : "commons-io",
    "artifactId" : "commons-io",
    "version" : "2.8.0",
    "optional" : false,
    "scopes" : [ "compile" ],
    "types" : [ "jar" ]
  }, {
    "id" : "org.apache.maven:maven-model:jar",
    "numericId" : 3,
    "groupId" : "org.apache.maven",
    "artifactId" : "maven-model",
    "version" : "3.6.2",
    "optional" : false,
    "scopes" : [ "compile" ],
    "types" : [ "jar" ]
  }, {
    "id" : "org.codehaus.plexus:plexus-utils:jar",
    "numericId" : 4,
    "groupId" : "org.codehaus.plexus",
    "artifactId" : "plexus-utils",
    "version" : "3.2.1",
    "optional" : false,
    "scopes" : [ "compile" ],
    "types" : [ "jar" ]
  } ],
  "dependencies" : [ {
    "from" : "com.sap.phosphor:sbom-spring-starter-project:jar",
    "to" : "commons-io:commons-io:jar",
    "numericFrom" : 0,
    "numericTo" : 1,
    "resolution" : "INCLUDED"
  }, {
    "from" : "org.apache.maven:maven-model:jar",
    "to" : "org.codehaus.plexus:plexus-utils:jar",
    "numericFrom" : 2,
    "numericTo" : 3,
    "resolution" : "INCLUDED"
  }, {
    "from" : "com.sap.phosphor:sbom-spring-starter-project:jar",
    "to" : "org.apache.maven:maven-model:jar",
    "numericFrom" : 0,
    "numericTo" : 2,
    "resolution" : "INCLUDED"
  } ]
}`

var dummyRating = fosstars.Rating{
	Id:                      "test",
	RatingDefinitionId:      "test",
	ModelRatingDefinitionId: "test",
	Created:                 "test",
	Value:                   1.3546,
	Confidence:              1.63524,
	Label:                   "test",
	NameSpace:               "test",
	Name:                    "test",
	RepositoryType:          "test",
	CoordinateValue:         "test"}

var testCases = []struct {
	testName      string
	s             fosstars.Fosstars
	expectedError error
}{
	{
		testName: "test with BuildDescriptor",
		s:        fosstars.Fosstars{BuildDescriptor: "/pom.xml", SCMUrl: "https://github.wdf.sap.corp/Phosphor/fosstars-piper-integration-sample", Branch: "master"},
	},
	{
		testName: "test with empty BuildDescriptor",
		s:        fosstars.Fosstars{BuildDescriptor: "", SCMUrl: "https://github.wdf.sap.corp/Phosphor/fosstars-piper-integration-sample", Branch: "master"},
	},
	{
		testName:      "test without SCMUrl",
		s:             fosstars.Fosstars{BuildDescriptor: "/pom.xml", SCMUrl: "", Branch: "master"},
		expectedError: errors.New("SCMUrl must not be empty"),
	},
	{
		testName:      "test without Branch",
		s:             fosstars.Fosstars{SCMUrl: "https://github.wdf.sap.corp/Phosphor/fosstars-piper-integration-sample", BuildDescriptor: "/pom.xml", Branch: ""},
		expectedError: errors.New("Branch must not be empty"),
	},
	{
		testName: "test includeTransitiveDependency:",
		s:        fosstars.Fosstars{BuildDescriptor: "/pom.xml", SCMUrl: "https://github.wdf.sap.corp/Phosphor/fosstars-piper-integration-sample", Branch: "master", IncludeTransitiveDependency: false},
	},
}

func TestFosstarsReport(t *testing.T) {
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			mockedInvocationClient := &mocks.InvocationClientInterface{}
			sapCreateFosstarsReportOptionsData := sapCreateFosstarsReportOptions{}
			mockedInvocationClient.Mock.On("GetRatings", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(func(sapFosstars *fosstars.Fosstars, fosstarsQueryServiceURL string, artifacts []string, duration time.Duration) map[string]*fosstars.Rating {
				return nil
			}).Once()
			artifacts := []string{}
			_, _, err := extractRatings(&tt.s, "", mockedInvocationClient, artifacts, sapCreateFosstarsReportOptionsData)

			if assert.NoError(t, err) {
				mockedInvocationClient.Mock.AssertNumberOfCalls(t, "GetRatings", 1)
			}
		})
	}
}

func TestIsMavenBuildDescriptor(t *testing.T) {
	sapFosstars := fosstars.Fosstars{BuildDescriptor: "/pom.xml", SCMUrl: "", Branch: ""}
	assert.True(t, isMavenBuildDescriptor(&sapFosstars), "Expecting boolean true")
	sapFosstarsNew := fosstars.Fosstars{BuildDescriptor: "/package.json", SCMUrl: "", Branch: ""}
	assert.False(t, isMavenBuildDescriptor(&sapFosstarsNew), "Expecting boolean false")
}

func TestIsNpmBuildDescriptor(t *testing.T) {
	sapFosstars := fosstars.Fosstars{BuildDescriptor: "/pom.xml", SCMUrl: "", Branch: ""}
	assert.False(t, isNpmBuildDescriptor(&sapFosstars), "Expecting boolean false")
	sapFosstarsNew := fosstars.Fosstars{BuildDescriptor: "/package.json", SCMUrl: "", Branch: ""}
	assert.True(t, isNpmBuildDescriptor(&sapFosstarsNew), "Expecting boolean true")
}

func TestIsMtaBuildDescriptor(t *testing.T) {
	sapFosstars := fosstars.Fosstars{BuildDescriptor: "/pom.xml", SCMUrl: "", Branch: ""}
	assert.False(t, isMtaBuildDescriptor(&sapFosstars), "Expecting boolean false")
	sapFosstarsNew := fosstars.Fosstars{BuildDescriptor: "/mta.yaml", SCMUrl: "", Branch: ""}
	assert.True(t, isMtaBuildDescriptor(&sapFosstarsNew), "Expecting boolean true")
}

func getFilesInPath(checkOutPath string, fileName string) ([]string, error) {
	matches, err := doublestar.Glob(checkOutPath + "**/**/" + fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to search for files %w", err)
	}
	return matches, nil
}

func TestExtractRatingsForNPMProjects(t *testing.T) {

	dummyRatingsData := map[string]*fosstars.Rating{
		"axios": &dummyRating,
	}

	mockedInvocationClient := &mocks.InvocationClientInterface{}
	sapCreateFosstarsReportOptionsData := sapCreateFosstarsReportOptions{}
	mockedInvocationClient.Mock.On("GetAllRatingsFromFosstars").Return(func() map[string]interface{} { return nil }).Once()
	mockedInvocationClient.Mock.On("GetRatings", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(func(sapFosstars *fosstars.Fosstars, fosstarsQueryServiceURL string, artifacts []string, duration time.Duration) map[string]*fosstars.Rating {
		return dummyRatingsData
	}).Once()

	workingDir, err := os.MkdirTemp("", "temp_directory_npm")
	previousWorkingDir := filepath.Join(workingDir, "")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating the temporary directory")
	}
	defer os.RemoveAll(workingDir) // clean up
	npmDir, err := os.MkdirTemp(workingDir, "-npm")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating temporary npm directory")
	}
	checkOutPath := filepath.Join(npmDir, "")

	dependencyGraphFilePath := filepath.Join(npmDir, "package.json")
	os.WriteFile(dependencyGraphFilePath, []byte(npmStr), 0644)
	sapFosstars := fosstars.Fosstars{
		SCMUrl:          "testUrl",
		Branch:          "testBranch",
		BuildDescriptor: "/package.json",
	}

	_, err = extractRatingsForNPMProjects(&sapFosstars, mockedInvocationClient, checkOutPath, previousWorkingDir, sapCreateFosstarsReportOptionsData)
	if assert.NoError(t, err) {
		matches, _ := getFilesInPath(previousWorkingDir, "fosstars-npm-dependency-check-fosstars.html")
		assert.Equal(t, 1, len(matches), "Expecting one HTML report file")
	}
}

func TestExtractRatingsForMtaProjects(t *testing.T) {

	yamlStr := `_schema-version: "2.0.0"
ID: fosstar-app
version: 0.1.0-SNAPSHOT
modules:
   - name: npm-project
     type: js
     path: npm-project
     build-parameters:
      builder: node.js`

	dummyRatingsData := map[string]*fosstars.Rating{
		"axios": &dummyRating,
	}

	mockedInvocationClient := &mocks.InvocationClientInterface{}
	sapCreateFosstarsReportOptionsData := sapCreateFosstarsReportOptions{}
	mockedInvocationClient.Mock.On("GetAllRatingsFromFosstars").Return(func() map[string]interface{} { return nil }).Times(2)
	mockedInvocationClient.Mock.On("GetRatings", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(func(sapFosstars *fosstars.Fosstars, fosstarsQueryServiceURL string, artifacts []string, duration time.Duration) map[string]*fosstars.Rating {
		return dummyRatingsData
	}).Times(2)

	workingDir, err := os.MkdirTemp("", "temp_directory")
	previousWorkingDir := filepath.Join(workingDir, "")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating the temporary directory")
	}
	defer os.RemoveAll(workingDir) // clean up
	mtaDir, err := os.MkdirTemp(workingDir, "-mta")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating temporary npm directory")
	}
	checkOutPath := filepath.Join(mtaDir, "")

	mtaFilePath := filepath.Join(mtaDir, "mta.yaml")
	os.WriteFile(mtaFilePath, []byte(yamlStr), 0755)

	npmDir := filepath.Join(mtaDir, "npm-project")
	os.MkdirAll(npmDir, 0755)

	npmFilePath := filepath.Join(npmDir, "package.json")
	os.WriteFile(npmFilePath, []byte(npmStr), 0755)

	sapFosstars := fosstars.Fosstars{
		SCMUrl:          "testUrl",
		Branch:          "testBranch",
		BuildDescriptor: "/mta.yaml",
	}

	_, err = extractRatingsForMtaProjects(&sapFosstars, mockedInvocationClient, checkOutPath, previousWorkingDir, sapCreateFosstarsReportOptionsData)
	if assert.NoError(t, err) {
		npmMatches, _ := getFilesInPath(previousWorkingDir, "fosstars-npm-dependency-check-fosstars.html")
		assert.Equal(t, 1, len(npmMatches), "Expecting one HTML report file for npm project")
	}
}

func TestExtractRatingsForMavenProjects(t *testing.T) {

	dummyRatingsData := map[string]*fosstars.Rating{
		"commons-io/commons-io": &dummyRating,
	}

	mockedInvocationClient := &mocks.InvocationClientInterface{}
	sapCreateFosstarsReportOptionsData := sapCreateFosstarsReportOptions{}
	mockedInvocationClient.Mock.On("GetAllRatingsFromFosstars").Return(func() map[string]interface{} { return nil }).Once()
	mockedInvocationClient.Mock.On("GetRatings", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(func(sapFosstars *fosstars.Fosstars, fosstarsQueryServiceURL string, artifacts []string, duration time.Duration) map[string]*fosstars.Rating {
		return dummyRatingsData
	}).Once()

	workingDir, err := os.MkdirTemp("", "temp_directory")
	previousWorkingDir := filepath.Join(workingDir, "")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating the temporary directory")
	}
	defer os.RemoveAll(workingDir) // clean up
	mavenDir, err := os.MkdirTemp(workingDir, "-maven")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating temporary npm directory")
	}
	checkOutPath := filepath.Join(mavenDir, "")

	mavenFilePath := filepath.Join(mavenDir, "pom.xml")
	os.WriteFile(mavenFilePath, []byte(mavenStr), 0644)
	sapFosstars := fosstars.Fosstars{
		SCMUrl:          "testUrl",
		Branch:          "testBranch",
		BuildDescriptor: "/pom.xml",
	}

	dummyDependencyTree := fosstars.DependencyTree{
		Artifact:           "commons-io:commons-io",
		Value:              1.234,
		Label:              "test",
		Confidence:         2.2334,
		Created:            "1622551715716",
		RatingDefinitionId: "test",
		RatingId:           "test",
		Children:           nil,
		Parent:             "test",
		Excluded:           false}

	dependencyTreeStr := `com.sap.phosphor:sbom-spring-starter-project:jar:0.1.0-SNAPSHOT
+- commons-io:commons-io:jar:2.8.0:compile
\- org.apache.maven:maven-model:jar:3.6.2:compile
   \- org.codehaus.plexus:plexus-utils:jar:3.2.1:compile`

	dependencyTreeFilePath := filepath.Join(mavenDir, "fosstar-generated-tree.txt")
	os.WriteFile(dependencyTreeFilePath, []byte(dependencyTreeStr), 0644)
	dependencyGraphJsonFilePath := filepath.Join(mavenDir, "dependency-graph.json")
	os.WriteFile(dependencyGraphJsonFilePath, []byte(dependencyGraphJsonStr), 0644)

	mockedDependencyTreeGenerator := &mocks.DependencyTreeGeneratorInterface{}
	mockedDependencyTreeGenerator.Mock.On("GetDependencyTreeFiles", mock.Anything, mock.Anything, mock.Anything).Return(func(pomFilePath string, globalSettingsFile string, buildQuality string) []string {
		return []string{dependencyTreeFilePath}
	}).Once()
	mockedDependencyTreeGenerator.Mock.On("AddRelativeBuidDescriptorPathToJson", mock.Anything, mock.Anything).Return(func(ratingDetailsJson []byte, buildDescriptorRelativePath string) []byte {
		return []byte(dependencyGraphJsonStr)
	}).Once()
	mockedDependencyTreeGenerator.Mock.On("GetDependencyTreeForNPM", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(artifactsRatingsMap map[string]*fosstars.Rating, rootArtifact string, excludedLibraries []string, excludeSAPInternalLibraries bool, excludeTestDevDependencies bool, devDependencies []string) *fosstars.DependencyTree {
			return nil
		}).Once()
	mockedDependencyTreeGenerator.Mock.On("ParseDependencyTreeFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(depGraphJsonFilePath string, artifactsRatingsMap map[string]*fosstars.Rating, pomArtifact string, excludedLibraries []string, excludeSAPInternalLibraries bool, excludeTestDevDependencies bool) *fosstars.DependencyTree {
			return &dummyDependencyTree
		}).Once()

	_, err = extractRatingsForMavenProjects(&sapFosstars, mockedInvocationClient, checkOutPath, previousWorkingDir, mockedDependencyTreeGenerator, sapCreateFosstarsReportOptionsData)
	if assert.NoError(t, err) {
		matches, _ := getFilesInPath(previousWorkingDir, "fosstars-sbom-spring-starter-project-fosstars.html")
		assert.Equal(t, 1, len(matches), "Expecting one HTML report file")
	}
}

func TestGetOptedSettingsForFosstrasReport(t *testing.T) {
	sapCreateFosstarsReportOptionsData := sapCreateFosstarsReportOptions{
		BuildDescriptor: "unknown.xml",
		SCMURL:          "test",
		Branch:          "master"}
	sapFosstars := getOptedSettingsForFosstrasReport(sapCreateFosstarsReportOptionsData)
	assert.Equal(t, "unknown.xml", sapFosstars.BuildDescriptor, "Expecting the correct build descriptor")
	assert.Equal(t, "test", sapFosstars.SCMUrl, "Expecting the correct SCM URL")
}

func TestGetLoggingInfo(t *testing.T) {
	sapCreateFosstarsReportOptionsData := sapCreateFosstarsReportOptions{
		BuildDescriptor: "unknown.xml",
		SCMURL:          "test",
		Branch:          "master"}
	sapFosstars := getLoggingInfo(sapCreateFosstarsReportOptionsData)
	assert.Equal(t, "unknown.xml", sapFosstars["Build Descriptor"].(string), "Expecting the correct build descriptor")
	assert.Equal(t, "test", sapFosstars["SCM Url"].(string), "Expecting the correct SCM URL")
}

func TestGetSCMCheckoutPathWithCheckout(t *testing.T) {
	sapCreateFosstarsReportOptionsData := sapCreateFosstarsReportOptions{
		BuildDescriptor: "unknown.xml",
		SCMURL:          "test",
		Branch:          "master",
		SkipSCMCheckOut: false,
	}
	sapFosstars := getOptedSettingsForFosstrasReport(sapCreateFosstarsReportOptionsData)
	_, err := getSCMCheckoutPath(sapCreateFosstarsReportOptionsData, "test", &sapFosstars)
	assert.True(t, err != nil, "Expecting error")
}

func TestGetSCMCheckoutPathWithCheckoutWithoutScmURL(t *testing.T) {
	sapCreateFosstarsReportOptionsData := sapCreateFosstarsReportOptions{
		BuildDescriptor: "unknown.xml",
		SCMURL:          "",
		Branch:          "master",
		SkipSCMCheckOut: false,
	}
	sapFosstars := getOptedSettingsForFosstrasReport(sapCreateFosstarsReportOptionsData)
	_, err := getSCMCheckoutPath(sapCreateFosstarsReportOptionsData, "test", &sapFosstars)
	assert.True(t, err != nil, "Expecting error")
}

func TestGetSCMCheckoutPath(t *testing.T) {
	sapCreateFosstarsReportOptionsData := sapCreateFosstarsReportOptions{
		BuildDescriptor: "unknown.xml",
		SCMURL:          "testURL",
		Branch:          "master",
		SkipSCMCheckOut: true,
	}
	sapFosstars := getOptedSettingsForFosstrasReport(sapCreateFosstarsReportOptionsData)
	checkoutPath, err := getSCMCheckoutPath(sapCreateFosstarsReportOptionsData, "test", &sapFosstars)
	assert.True(t, err == nil, "Expecting no error")
	assert.Equal(t, "test", checkoutPath, "Expecting the correct SCM path")
}

func TestHasBuildDescriptorAtRoot(t *testing.T) {
	workingDir, err := os.MkdirTemp("", "temp_directory_npm")
	checkOutPath := filepath.Join(workingDir, "")
	defer os.RemoveAll(workingDir) // clean up
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating the temporary directory")
	}

	buildDescriptorFilePath := filepath.Join(workingDir, "package.json")
	os.WriteFile(buildDescriptorFilePath, []byte(npmStr), 0644)
	assert.True(t, hasBuildDescriptorAtRoot(checkOutPath, "package.json"), "Expecting the file to exist")
	assert.False(t, hasBuildDescriptorAtRoot(checkOutPath, "pom.xml"), "Expecting no file")
}

func TestRunSapCreateFosstarsReportForNPMProjects(t *testing.T) {

	dummyRatingsData := map[string]*fosstars.Rating{
		"axios": &dummyRating,
	}

	mockedInvocationClient := &mocks.InvocationClientInterface{}
	mockedInvocationClient.Mock.On("GetAllRatingsFromFosstars").Return(func() map[string]interface{} { return nil }).Once()
	mockedInvocationClient.Mock.On("GetRatings", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(func(sapFosstars *fosstars.Fosstars, fosstarsQueryServiceURL string, artifacts []string, duration time.Duration) map[string]*fosstars.Rating {
		return dummyRatingsData
	}).Once()

	workingDir, err := os.MkdirTemp("", "temp_directory_npm")
	previousWorkingDir := filepath.Join(workingDir, "")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating the temporary directory")
	}
	defer os.RemoveAll(workingDir) // clean up
	npmDir, err := os.MkdirTemp(workingDir, "-npm")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating temporary npm directory")
	}
	checkOutPath := filepath.Join(npmDir, "")

	dependencyGraphFilePath := filepath.Join(npmDir, "package.json")
	os.WriteFile(dependencyGraphFilePath, []byte(npmStr), 0644)
	sapFosstars := fosstars.Fosstars{
		SCMUrl:          checkOutPath,
		Branch:          "testBranch",
		BuildDescriptor: "/package.json",
	}

	sapCreateFosstarsReportOptionsData := sapCreateFosstarsReportOptions{
		BuildDescriptor: "/package.json",
		SCMURL:          checkOutPath,
		Branch:          "testBranch",
		SkipSCMCheckOut: true,
	}

	_, err = runSapCreateFosstarsReport(sapCreateFosstarsReportOptionsData, &sapFosstars, previousWorkingDir, checkOutPath, mockedInvocationClient, &fosstars.DependencyTreeGenerator{})
	if assert.NoError(t, err) {
		matches, _ := getFilesInPath(previousWorkingDir, "fosstars-npm-dependency-check-fosstars.html")
		assert.Equal(t, 1, len(matches), "Expecting one HTML report file")
	}
}

func TestRunSapCreateFosstarsReportForMavenProjects(t *testing.T) {

	dummyRatingsData := map[string]*fosstars.Rating{
		"commons-io/commons-io": &dummyRating,
	}

	mockedInvocationClient := &mocks.InvocationClientInterface{}
	mockedInvocationClient.Mock.On("GetAllRatingsFromFosstars").Return(func() map[string]interface{} { return nil }).Once()
	mockedInvocationClient.Mock.On("GetRatings", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(func(sapFosstars *fosstars.Fosstars, fosstarsQueryServiceURL string, artifacts []string, duration time.Duration) map[string]*fosstars.Rating {
		return dummyRatingsData
	}).Once()

	workingDir, err := os.MkdirTemp("", "temp_directory")
	previousWorkingDir := filepath.Join(workingDir, "")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating the temporary directory")
	}
	defer os.RemoveAll(workingDir) // clean up
	mavenDir, err := os.MkdirTemp(workingDir, "-maven")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating temporary npm directory")
	}
	checkOutPath := filepath.Join(mavenDir, "")

	mavenFilePath := filepath.Join(mavenDir, "pom.xml")
	os.WriteFile(mavenFilePath, []byte(mavenStr), 0644)
	sapFosstars := fosstars.Fosstars{
		SCMUrl:          checkOutPath,
		Branch:          "testBranch",
		BuildDescriptor: "/pom.xml",
	}

	sapCreateFosstarsReportOptionsData := sapCreateFosstarsReportOptions{
		BuildDescriptor: "/pom.xml",
		SCMURL:          checkOutPath,
		Branch:          "testBranch",
		SkipSCMCheckOut: true,
	}

	dummyDependencyTree := fosstars.DependencyTree{
		Artifact:           "commons-io:commons-io",
		Value:              1.234,
		Label:              "test",
		Confidence:         2.2334,
		Created:            "1622551715716",
		RatingDefinitionId: "test",
		RatingId:           "test",
		Children:           nil,
		Parent:             "test",
		Excluded:           false}

	dependencyTreeStr := `com.sap.phosphor:sbom-spring-starter-project:jar:0.1.0-SNAPSHOT
+- commons-io:commons-io:jar:2.8.0:compile
\- org.apache.maven:maven-model:jar:3.6.2:compile
   \- org.codehaus.plexus:plexus-utils:jar:3.2.1:compile`

	dependencyTreeFilePath := filepath.Join(mavenDir, "fosstar-generated-tree.txt")
	os.WriteFile(dependencyTreeFilePath, []byte(dependencyTreeStr), 0644)
	dependencyGraphJsonFilePath := filepath.Join(mavenDir, "dependency-graph.json")
	os.WriteFile(dependencyGraphJsonFilePath, []byte(dependencyGraphJsonStr), 0644)

	mockedDependencyTreeGenerator := &mocks.DependencyTreeGeneratorInterface{}
	mockedDependencyTreeGenerator.Mock.On("GetDependencyTreeFiles", mock.Anything, mock.Anything, mock.Anything).Return(func(pomFilePath string, globalSettingsFile string, buildQuality string) []string {
		return []string{dependencyTreeFilePath}
	}).Once()
	mockedDependencyTreeGenerator.Mock.On("AddRelativeBuidDescriptorPathToJson", mock.Anything, mock.Anything).Return(func(ratingDetailsJson []byte, buildDescriptorRelativePath string) []byte {
		return []byte(dependencyGraphJsonStr)
	}).Once()
	mockedDependencyTreeGenerator.Mock.On("GetDependencyTreeForNPM", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(artifactsRatingsMap map[string]*fosstars.Rating, rootArtifact string, excludedLibraries []string, excludeSAPInternalLibraries bool, excludeTestDevDependencies bool, devDependencies []string) *fosstars.DependencyTree {
			return nil
		}).Once()
	mockedDependencyTreeGenerator.Mock.On("ParseDependencyTreeFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(depGraphJsonFilePath string, artifactsRatingsMap map[string]*fosstars.Rating, pomArtifact string, excludedLibraries []string, excludeSAPInternalLibraries bool, excludeTestDevDependencies bool) *fosstars.DependencyTree {
			return &dummyDependencyTree
		}).Once()

	_, err = runSapCreateFosstarsReport(sapCreateFosstarsReportOptionsData, &sapFosstars, previousWorkingDir, checkOutPath, mockedInvocationClient, mockedDependencyTreeGenerator)
	if assert.NoError(t, err) {
		matches, _ := getFilesInPath(previousWorkingDir, "fosstars-sbom-spring-starter-project-fosstars.html")
		assert.Equal(t, 1, len(matches), "Expecting one HTML report file")
	}
}

func TestRunSapCreateFosstarsReportForMtaProjects(t *testing.T) {

	yamlStr := `_schema-version: "2.0.0"
ID: fosstar-app
version: 0.1.0-SNAPSHOT
modules:
   - name: npm-project
     type: js
     path: npm-project
     build-parameters:
      builder: node.js`

	dummyRatingsData := map[string]*fosstars.Rating{
		"axios": &dummyRating,
	}

	mockedInvocationClient := &mocks.InvocationClientInterface{}
	mockedInvocationClient.Mock.On("GetAllRatingsFromFosstars").Return(func() map[string]interface{} { return nil }).Times(2)
	mockedInvocationClient.Mock.On("GetRatings", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(func(sapFosstars *fosstars.Fosstars, fosstarsQueryServiceURL string, artifacts []string, duration time.Duration) map[string]*fosstars.Rating {
		return dummyRatingsData
	}).Times(2)

	workingDir, err := os.MkdirTemp("", "temp_directory")
	previousWorkingDir := filepath.Join(workingDir, "")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating the temporary directory")
	}
	defer os.RemoveAll(workingDir) // clean up
	mtaDir, err := os.MkdirTemp(workingDir, "-mta")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating temporary npm directory")
	}
	checkOutPath := filepath.Join(mtaDir, "")

	mtaFilePath := filepath.Join(mtaDir, "mta.yaml")
	os.WriteFile(mtaFilePath, []byte(yamlStr), 0755)

	npmDir := filepath.Join(mtaDir, "npm-project")
	os.MkdirAll(npmDir, 0755)

	npmFilePath := filepath.Join(npmDir, "package.json")
	os.WriteFile(npmFilePath, []byte(npmStr), 0755)

	sapFosstars := fosstars.Fosstars{
		SCMUrl:          checkOutPath,
		Branch:          "testBranch",
		BuildDescriptor: "/mta.yaml",
	}

	sapCreateFosstarsReportOptionsData := sapCreateFosstarsReportOptions{
		BuildDescriptor: "/mta.yaml",
		SCMURL:          checkOutPath,
		Branch:          "testBranch",
		SkipSCMCheckOut: true,
	}

	_, err = runSapCreateFosstarsReport(sapCreateFosstarsReportOptionsData, &sapFosstars, previousWorkingDir, checkOutPath, mockedInvocationClient, &fosstars.DependencyTreeGenerator{})
	if assert.NoError(t, err) {
		npmMatches, _ := getFilesInPath(previousWorkingDir, "fosstars-npm-dependency-check-fosstars.html")
		assert.Equal(t, 1, len(npmMatches), "Expecting one HTML report file for npm project")
	}
}

func TestRunSapCreateFosstarsReportForMavenProjectsAutomatically(t *testing.T) {

	dummyRatingsData := map[string]*fosstars.Rating{
		"commons-io/commons-io": &dummyRating,
	}

	mockedInvocationClient := &mocks.InvocationClientInterface{}
	mockedInvocationClient.Mock.On("GetAllRatingsFromFosstars").Return(func() map[string]interface{} { return nil }).Once()
	mockedInvocationClient.Mock.On("GetRatings", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(func(sapFosstars *fosstars.Fosstars, fosstarsQueryServiceURL string, artifacts []string, duration time.Duration) map[string]*fosstars.Rating {
		return dummyRatingsData
	}).Once()

	workingDir, err := os.MkdirTemp("", "temp_directory")
	previousWorkingDir := filepath.Join(workingDir, "")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating the temporary directory")
	}
	defer os.RemoveAll(workingDir) // clean up
	mavenDir, err := os.MkdirTemp(workingDir, "-maven")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating temporary npm directory")
	}
	checkOutPath := filepath.Join(mavenDir, "")

	mavenFilePath := filepath.Join(mavenDir, "pom.xml")
	os.WriteFile(mavenFilePath, []byte(mavenStr), 0644)
	sapFosstars := fosstars.Fosstars{
		SCMUrl:                checkOutPath,
		Branch:                "testBranch",
		BuildAllMavenPomFiles: true,
	}

	sapCreateFosstarsReportOptionsData := sapCreateFosstarsReportOptions{
		SCMURL:                checkOutPath,
		Branch:                "testBranch",
		SkipSCMCheckOut:       true,
		BuildAllMavenPomFiles: true,
	}

	dummyDependencyTree := fosstars.DependencyTree{
		Artifact:           "commons-io:commons-io",
		Value:              1.234,
		Label:              "test",
		Confidence:         2.2334,
		Created:            "1622551715716",
		RatingDefinitionId: "test",
		RatingId:           "test",
		Children:           nil,
		Parent:             "test",
		Excluded:           false}

	dependencyTreeStr := `com.sap.phosphor:sbom-spring-starter-project:jar:0.1.0-SNAPSHOT
+- commons-io:commons-io:jar:2.8.0:compile
\- org.apache.maven:maven-model:jar:3.6.2:compile
   \- org.codehaus.plexus:plexus-utils:jar:3.2.1:compile`

	dependencyTreeFilePath := filepath.Join(mavenDir, "fosstar-generated-tree.txt")
	os.WriteFile(dependencyTreeFilePath, []byte(dependencyTreeStr), 0644)
	dependencyGraphJsonFilePath := filepath.Join(mavenDir, "dependency-graph.json")
	os.WriteFile(dependencyGraphJsonFilePath, []byte(dependencyGraphJsonStr), 0644)

	log.Entry().Infof("mavenFilePath: %v", mavenFilePath)
	mockedDependencyTreeGenerator := &mocks.DependencyTreeGeneratorInterface{}
	mockedDependencyTreeGenerator.Mock.On("GetAllPomFilesToBuildDependencyTree", mock.Anything).Return(func(checkOutPath string) []string { return []string{mavenFilePath} }).Once()
	mockedDependencyTreeGenerator.Mock.On("GetDependencyTreeFiles", mock.Anything, mock.Anything, mock.Anything).Return(func(pomFilePath string, globalSettingsFile string, buildQuality string) []string {
		return []string{dependencyTreeFilePath}
	}).Once()
	mockedDependencyTreeGenerator.Mock.On("AddRelativeBuidDescriptorPathToJson", mock.Anything, mock.Anything).Return(func(ratingDetailsJson []byte, buildDescriptorRelativePath string) []byte {
		return []byte(dependencyGraphJsonStr)
	}).Once()
	mockedDependencyTreeGenerator.Mock.On("GetDependencyTreeForNPM", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(artifactsRatingsMap map[string]*fosstars.Rating, rootArtifact string, excludedLibraries []string, excludeSAPInternalLibraries bool, excludeTestDevDependencies bool, devDependencies []string) *fosstars.DependencyTree {
			return nil
		}).Once()
	mockedDependencyTreeGenerator.Mock.On("ParseDependencyTreeFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(depGraphJsonFilePath string, artifactsRatingsMap map[string]*fosstars.Rating, pomArtifact string, excludedLibraries []string, excludeSAPInternalLibraries bool, excludeTestDevDependencies bool) *fosstars.DependencyTree {
			return &dummyDependencyTree
		}).Once()

	_, err = runSapCreateFosstarsReport(sapCreateFosstarsReportOptionsData, &sapFosstars, previousWorkingDir, checkOutPath, mockedInvocationClient, mockedDependencyTreeGenerator)
	if assert.NoError(t, err) {
		matches, _ := getFilesInPath(previousWorkingDir, "fosstars-sbom-spring-starter-project-fosstars.html")
		assert.Equal(t, 1, len(matches), "Expecting one HTML report file")
	}
}

func TestRunSapCreateFosstarsReportErrorWhileParsingAllFiles(t *testing.T) {

	malformedMavenStr := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
  <modelVersion>4.0.0</modelVersion>
  <groupId>com.sap.phosphor</groupId>
  <artifactId>sbom-spring-starter-project</artifactId>
  <version>0.1.0-SNAPSHOT</version>
  <packaging>jar</packaging>
  <name>sBOM Starter: sbom-spring-starter-project</name>
  <description>Start project for sBOM Spring Boot</description>
    <modules>
        <module>module1</module>
</project>`

	workingDir, err := os.MkdirTemp("", "temp_directory")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating the temporary directory")
	}
	defer os.RemoveAll(workingDir) // clean up
	mavenDir, err := os.MkdirTemp(workingDir, "-maven")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating temporary maven directory")
	}
	checkOutPath := filepath.Join(mavenDir, "")

	mavenFilePath := filepath.Join(mavenDir, "pom.xml")
	os.WriteFile(mavenFilePath, []byte(malformedMavenStr), 0644)

	sapFosstars := fosstars.Fosstars{
		SCMUrl:                checkOutPath,
		Branch:                "testBranch",
		BuildAllMavenPomFiles: true,
	}

	sapCreateFosstarsReportOptionsData := sapCreateFosstarsReportOptions{
		SCMURL:                checkOutPath,
		Branch:                "testBranch",
		SkipSCMCheckOut:       true,
		BuildAllMavenPomFiles: true,
	}

	_, err = runSapCreateFosstarsReport(sapCreateFosstarsReportOptionsData, &sapFosstars, "", checkOutPath, &fosstars.InvocationClient{}, &fosstars.DependencyTreeGenerator{})
	assert.Truef(t, err != nil, "Expecting error")
}

func TestRunSapCreateFosstarsReportErrorWhileParsing(t *testing.T) {

	malformedMavenStr := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
  <modelVersion>4.0.0</modelVersion>
  <groupId>com.sap.phosphor</groupId>
  <artifactId>sbom-spring-starter-project</artifactId>
  <version>0.1.0-SNAPSHOT</version>
  <packaging>jar</packaging>
  <name>sBOM Starter: sbom-spring-starter-project</name>
  <description>Start project for sBOM Spring Boot</description>
    <modules>
        <module>module1</module>
</project>`

	workingDir, err := os.MkdirTemp("", "temp_directory")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating the temporary directory")
	}
	defer os.RemoveAll(workingDir) // clean up
	mavenDir, err := os.MkdirTemp(workingDir, "-maven")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating temporary maven directory")
	}
	checkOutPath := filepath.Join(mavenDir, "")

	mavenFilePath := filepath.Join(mavenDir, "pom.xml")
	os.WriteFile(mavenFilePath, []byte(malformedMavenStr), 0644)

	sapFosstars := fosstars.Fosstars{
		SCMUrl:          checkOutPath,
		Branch:          "testBranch",
		BuildDescriptor: "/pom.xml",
	}

	sapCreateFosstarsReportOptionsData := sapCreateFosstarsReportOptions{
		SCMURL:          checkOutPath,
		Branch:          "testBranch",
		SkipSCMCheckOut: true,
		BuildDescriptor: "/pom.xml",
	}

	_, err = runSapCreateFosstarsReport(sapCreateFosstarsReportOptionsData, &sapFosstars, "", checkOutPath, &fosstars.InvocationClient{}, &fosstars.DependencyTreeGenerator{})
	assert.Truef(t, err != nil, "Expecting error")
}

func TestRunSapCreateFosstarsReportErrorUnkownFileDescriptor(t *testing.T) {

	sapFosstars := fosstars.Fosstars{
		SCMUrl:          "test",
		Branch:          "testBranch",
		BuildDescriptor: "/test.xml",
	}

	sapCreateFosstarsReportOptionsData := sapCreateFosstarsReportOptions{
		SCMURL:          "test",
		Branch:          "testBranch",
		SkipSCMCheckOut: true,
		BuildDescriptor: "/test.xml",
	}

	_, err := runSapCreateFosstarsReport(sapCreateFosstarsReportOptionsData, &sapFosstars, "", "", &fosstars.InvocationClient{}, &fosstars.DependencyTreeGenerator{})
	assert.Truef(t, err != nil, "Expecting error")
}

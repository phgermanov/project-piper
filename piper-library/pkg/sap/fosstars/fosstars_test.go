//go:build unit
// +build unit

package fosstars

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	piperutils "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/stretchr/testify/assert"
)

func TestGetYamlInfo(t *testing.T) {
	yamlStr := `_schema-version: "2.0.0"
ID: fosstar-app
version: 0.1.0-SNAPSHOT
modules:
   - name: fosstars-command-service
     type: java
     path: fosstars-command-service
     build-parameters:
      builder: maven`

	yamlDir, err := os.MkdirTemp(os.TempDir(), "-yaml")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating temporary yaml directory")
	}

	defer os.RemoveAll(yamlDir) // clean up

	yamlFile, err := os.CreateTemp(yamlDir, "mta-*.yaml")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating temporary yaml file")
	}

	if _, err := yamlFile.Write([]byte(yamlStr)); err != nil {
		log.Entry().WithError(err).Fatal("Error while writing to temporary yaml file")
	}

	yamlInfo, err := GetYamlInfo(yamlDir, yamlFile.Name())
	if assert.NoErrorf(t, err, "Error while calling GetYamlInfo with no modules") {
		assert.Equal(t, len(yamlInfo.MavenModulePaths), 0, "Expecting no paths")
		assert.Equal(t, len(yamlInfo.NpmModulePaths), 0, "Expecting no paths")
	}

	err = os.Mkdir(filepath.Join(yamlDir, "fosstars-command-service"), 0700)
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating fosstars-command-service directory")
	}

	pomFileTemp, err := os.Create(filepath.Join(yamlDir, "fosstars-command-service", "pom.xml"))
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating pom file")
	}

	pomFileTempExist, _ := piperutils.FileExists(pomFileTemp.Name())
	assert.Equal(t, pomFileTempExist, true, "Expecting pom file to be present")

	yamlInfoWithData, err := GetYamlInfo(yamlDir, yamlFile.Name())
	if assert.NoErrorf(t, err, "Error while calling GetYamlInfo with one maven module") {
		assert.Equal(t, len(yamlInfoWithData.MavenModulePaths), 1, "Expecting no paths")
		assert.Equal(t, len(yamlInfoWithData.NpmModulePaths), 0, "Expecting no paths")
	}
	//Negative Test
	yamlInfoWithData, err = GetYamlInfo(yamlDir, "")
	assert.True(t, err != nil, "Error expected for non-existing yaml file path")

}

func TestGetArtifactFromLine(t *testing.T) {
	artifact := GetArtifactFromLine("|  |  \\- com.squareup.moshi:moshi:jar:1.8.0:compile")
	assert.Equal(t, "com.squareup.moshi/moshi", artifact)
}

func TestValidateInput(t *testing.T) {
	sapFosstars := Fosstars{
		SCMUrl: "",
		Branch: "testBranch",
	}
	err := sapFosstars.ValidateInput()
	assert.True(t, err != nil, "Error expected for empty SCMUrl")
	assert.Equal(t, err.Error(), "SCMUrl must not be empty")
	sapFosstars = Fosstars{
		SCMUrl: "testUrl",
		Branch: "",
	}
	err = sapFosstars.ValidateInput()
	assert.True(t, err != nil, "Error expected for empty SCMUrl")
	assert.Equal(t, err.Error(), "Branch must not be empty")
	sapFosstars = Fosstars{
		SCMUrl: "testUrl",
		Branch: "testBranch",
	}
	err = sapFosstars.ValidateInput()
	assert.NoError(t, err)
}

func TestGetFileContentAsList(t *testing.T) {
	t.Parallel()

	yamlStr := `_schema-version: "2.0.0"
ID: fosstar-app
version: 0.1.0-SNAPSHOT
modules:
   - name: fosstars-command-service
     type: java
     path: fosstars-command-service
     build-parameters:
      builder: maven`

	t.Run("success case", func(t *testing.T) {
		t.Parallel()
		yamlDir, err := os.MkdirTemp(os.TempDir(), "-yaml")
		if err != nil {
			log.Entry().WithError(err).Fatal("Error while creating temporary yaml directory")
		}

		defer os.RemoveAll(yamlDir) // clean up

		yamlFile, err := os.CreateTemp(yamlDir, "mta-*.yaml")
		if err != nil {
			log.Entry().WithError(err).Fatal("Error while creating temporary yaml file")
		}

		if _, err := yamlFile.Write([]byte(yamlStr)); err != nil {
			log.Entry().WithError(err).Fatal("Error while writing to temporary yaml file")
		}

		lines, err := GetFileContentAsList(yamlFile.Name())
		assert.NoError(t, err, "Expected to succeed getting the artifacts")
		assert.True(t, len(lines) > 0, "Expected at least one line")
	})

	t.Run("error case - non-existing yaml file path", func(t *testing.T) {
		t.Parallel()
		_, err := GetFileContentAsList("/non-existing/path/file.yaml")
		assert.Error(t, err, "Error expected for non-existing yaml file path")
	})
}

func TestGetNpmInfo(t *testing.T) {
	npmStr := `{
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

	npmDir, err := os.MkdirTemp(os.TempDir(), "-npm")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating temporary npm directory")
	}

	defer os.RemoveAll(npmDir) // clean up

	npmFile, err := os.CreateTemp(npmDir, "package.json")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating temporary npm  file")
	}

	if _, err := npmFile.Write([]byte(npmStr)); err != nil {
		log.Entry().WithError(err).Fatal("Error while writing to temporary npm file")
	}
	sapFosstars := Fosstars{
		SCMUrl: "testUrl",
		Branch: "testBranch",
	}

	npmInfo, err := sapFosstars.GetNpmInfo(npmFile.Name())
	assert.NoError(t, err, "Expected to succeed getting the npmInfo")
	assert.Equal(t, "npm-dependency-check", npmInfo.Name)
	assert.Equal(t, "1.0.0", npmInfo.Version)
	assert.True(t, len(npmInfo.Artifacts) == 7, "Expecteded 7 artifacts")
	assert.True(t, len(npmInfo.Dependencies) == 4, "Expecteded 4 dependencies")
	assert.True(t, len(npmInfo.DevDependencies) == 3, "Expecteded 3 devDependencies")
	//Negative Test
	_, err = sapFosstars.GetNpmInfo("")
	assert.True(t, err != nil, "Error expected for non-existing npmFile file path")
}

func TestWriteToFile(t *testing.T) {
	dependencyTreeList := []string{}
	dependencyTreeList = append(dependencyTreeList, "com.sap.test:test-common:pom:1.0")
	dependencyTreeList = append(dependencyTreeList, "+- org.apache.activemq:artemis-journal:jar:2.13.0:compile")
	dependencyTreeList = append(dependencyTreeList, "|  - org.apache.activemq:activemq-artemis-native:jar:1.0.1:compile")
	dependencyTreeList = append(dependencyTreeList, "+- org.apache.activemq:artemis-selector:jar:2.13.0:compile")
	dependencyTreeList = append(dependencyTreeList, "+- org.apache.activemq:artemis-server:jar:2.13.0:compile")
	dependencyTreeList = append(dependencyTreeList, "|  +- org.jboss.logmanager:jboss-logmanager:jar:2.1.10.Final:compile")
	dependencyTreeList = append(dependencyTreeList, "|  +- org.apache.activemq:artemis-jdbc-store:jar:2.13.0:compile")
	dependencyTreeList = append(dependencyTreeList, "|  +- org.jctools:jctools-core:jar:2.1.2:compile")
	dependencyTreeList = append(dependencyTreeList, "|  - org.apache.commons:commons-configuration2:jar:2.7:compile")

	dependencyTreeDir, err := os.MkdirTemp(os.TempDir(), "-dependencyTree")
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating temporary dependencyTree directory")
	}

	defer os.RemoveAll(dependencyTreeDir) // clean up

	dependencyTreeFilePath := filepath.Join(dependencyTreeDir, "dependencyTree.txt")
	err = WriteToFile(dependencyTreeFilePath, dependencyTreeList)
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while writing to dependencyTree")
	}
	dependencyTreeFileExist, _ := piperutils.FileExists(dependencyTreeFilePath)
	assert.Equal(t, dependencyTreeFileExist, true, "Expecting dependencyTree file to be present")
	//Negative Path
	err = WriteToFile("", dependencyTreeList)
	assert.True(t, err != nil, "Error expected for invalid file path")
}

func TestCloneRepo(t *testing.T) {
	sapFosstars := Fosstars{
		BuildDescriptor: "",
		SCMUrl:          "https://github.wdf.sap.corp/Phosphor/fosstars-piper-integration-sample",
		Branch:          "master",
	}

	err := sapFosstars.CloneRepo("", "", "SampleCheckOutPath", time.Minute)
	assert.True(t, err != nil, "Error expected for invalid username and password as the anynymous access is restricted for github enterprise")

}

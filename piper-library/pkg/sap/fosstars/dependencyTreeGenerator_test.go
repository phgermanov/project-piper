//go:build unit
// +build unit

package fosstars

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const moduleMavenStr = `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 https://maven.apache.org/xsd/maven-4.0.0.xsd">
  <modelVersion>4.0.0</modelVersion>
  <groupId>com.sap.phosphor</groupId>
  <artifactId>sbom-spring-starter-project</artifactId>
  <version>0.1.0-SNAPSHOT</version>
  <packaging>jar</packaging>
  <name>sBOM Starter: sbom-spring-starter-project</name>
  <description>Start project for sBOM Spring Boot</description>
</project>`

func TestGetDirectDependencies(t *testing.T) {
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

	assert.Len(t, GetDirectDependencies(dependencyTreeList), 4, "Expecting the proper count")
}

func TestParseDependencyTreeFile(t *testing.T) {
	dependencyTreeGenerator := DependencyTreeGenerator{}
	var excludedLibrarires []string
	artifactsRatingsMap :=
		make(map[string]*Rating)
	artifactsRatingsMap["com.sap.test/test-common"] = &Rating{Id: "123445", RatingDefinitionId: "32435645765", Created: "2020-04-14T122344", Value: 0.23444, Confidence: 0.3, Label: "BAD", NameSpace: "", Name: "", RepositoryType: "", CoordinateValue: ""}
	artifactsRatingsMap["com.fasterxml.jackson.core/jackson-databind"] = &Rating{Id: "125687", RatingDefinitionId: "123545", Created: "2020-04-14T122344", Value: 0.43444, Confidence: 0.3, Label: "BAD", NameSpace: "", Name: "", RepositoryType: "", CoordinateValue: ""}
	artifactsRatingsMap["com.sap.test/test-sap-internal-child"] = &Rating{Id: "", RatingDefinitionId: "", Created: "", Label: "", NameSpace: "", Name: "", RepositoryType: "", CoordinateValue: ""}
	artifactsRatingsMap["com.google.code.findbugs/jsr305"] = &Rating{Id: "125688", RatingDefinitionId: "123546", Created: "2020-04-14T122344", Value: 0.43444, Confidence: 0.3, Label: "BAD", NameSpace: "", Name: "", RepositoryType: "", CoordinateValue: ""}
	// Error case
	_, err := dependencyTreeGenerator.ParseDependencyTreeFile("", artifactsRatingsMap, "com.sap.test/test-common", excludedLibrarires, true, true)
	assert.Error(t, err, "Error expected for non-existing dependency graph file path")
	// Happy path
	dependency_graph_json := " {\"artifacts\":[{\"id\":\"com.fasterxml.jackson.core:jackson-databind:jar\", \"numericId\":1,\"scopes\":[\"compile\"], \"Version\":\"2.10.0\"},{\"id\":\"com.sap.test:test-sap-internal-child:jar\",\"numericId\":2,\"scopes\":[\"compile\"]},{\"id\":\"com.google.code.findbugs:jsr305:jar\",\"numericId\":3,\"scopes\":[\"test\"]}] ,\"dependencies\":[{\"from\":\"com.sap.test:test-common\",\"to\":\"com.fasterxml.jackson.core:jackson-databind:jar\",\"numericFrom\":0,\"numericTo\":1}, {\"from\":\"com.sap.test:test-common\",\"to\":\"com.sap.test:test-sap-internal-child\",\"numericFrom\":0,\"numericTo\":2}, {\"from\":\"com.sap.test:test-common\",\"to\":\"com.google.code.findbugs:jsr305\",\"numericFrom\":0,\"numericTo\":3}]}"
	dir, err := os.MkdirTemp("", "temp_directory")
	require.NoError(t, err, "Error while creating the temporary directory")
	defer os.RemoveAll(dir) // clean up

	dependencyGraphFilePath := filepath.Join(dir, "test_depndency_graph.json")
	os.WriteFile(dependencyGraphFilePath, []byte(dependency_graph_json), 0644)
	parent, err := dependencyTreeGenerator.ParseDependencyTreeFile(dependencyGraphFilePath, artifactsRatingsMap, "com.sap.test/test-common", excludedLibrarires, true, true)
	require.NoError(t, err)
	assert.Equal(t, "com.sap.test/test-common", parent.Artifact)
	assert.Equal(t, "com.fasterxml.jackson.core/jackson-databind", parent.Children["com.fasterxml.jackson.core/jackson-databind"].Artifact)
	assert.Equal(t, "BAD", parent.Children["com.fasterxml.jackson.core/jackson-databind"].Label)
	assert.Equal(t, "2020-04-14T122344", parent.Children["com.fasterxml.jackson.core/jackson-databind"].Created)
	assert.Equal(t, "123545", parent.Children["com.fasterxml.jackson.core/jackson-databind"].RatingDefinitionId)
	assert.Equal(t, "125687", parent.Children["com.fasterxml.jackson.core/jackson-databind"].RatingId)
	assert.Equal(t, "2.10.0", parent.Children["com.fasterxml.jackson.core/jackson-databind"].Version)
	assert.False(t, parent.Children["com.fasterxml.jackson.core/jackson-databind"].Excluded)
	assert.True(t, parent.Children["com.sap.test/test-sap-internal-child"].Excluded)
	assert.True(t, parent.Children["com.google.code.findbugs/jsr305"].Excluded)
}

func TestGetDependencyTreeForNPM(t *testing.T) {
	dependencyTreeGenerator := DependencyTreeGenerator{}
	var excludedLibrarires []string
	devDependencies := []string{"axios"}
	artifactsRatingsMap :=
		make(map[string]*Rating)
	artifactsRatingsMap["axios"] = &Rating{Id: "123445", RatingDefinitionId: "32435645765", Created: "2020-04-14T122344", Value: 0.23444, Confidence: 0.3, Label: "BAD", NameSpace: "", Name: "", RepositoryType: "", CoordinateValue: ""}
	artifactsRatingsMap["commander"] = &Rating{Id: "125687", RatingDefinitionId: "123545", Created: "2020-04-14T122344", Value: 0.43444, Confidence: 0.3, Label: "BAD", NameSpace: "", Name: "", RepositoryType: "", CoordinateValue: ""}
	artifactsRatingsMap["@sap/test"] = &Rating{Id: "125688", RatingDefinitionId: "123545", Created: "2020-04-14T122344", Value: 0.43444, Confidence: 0.3, Label: "", NameSpace: "", Name: "", RepositoryType: "", CoordinateValue: ""}
	parent, err := dependencyTreeGenerator.GetDependencyTreeForNPM(artifactsRatingsMap, "npm-dependency-check", excludedLibrarires, true, true, devDependencies)

	require.NoError(t, err)
	assert.Equal(t, "npm-dependency-check", parent.Artifact)
	assert.Equal(t, "axios", parent.Children["axios"].Artifact)
	assert.Equal(t, "BAD", parent.Children["axios"].Label)
	assert.Equal(t, "2020-04-14T122344", parent.Children["axios"].Created)
	assert.Equal(t, "32435645765", parent.Children["axios"].RatingDefinitionId)
	assert.Equal(t, "123445", parent.Children["axios"].RatingId)
	assert.True(t, parent.Children["axios"].Excluded)
	assert.Equal(t, "123545", parent.Children["@sap/test"].RatingDefinitionId)
	assert.Equal(t, "125688", parent.Children["@sap/test"].RatingId)
	assert.Equal(t, "@sap/test", parent.Children["@sap/test"].Artifact)
	assert.Equal(t, "2020-04-14T122344", parent.Children["@sap/test"].Created)
	assert.True(t, parent.Children["@sap/test"].Excluded)
}

func TestGetChildParentRelationMap(t *testing.T) {
	dependency_graph_json := " {\"artifacts\":[{\"id\":\"com.fasterxml.jackson.core:jackson-databind:jar\",\"numericId\":1,\"scopes\":[\"compile\"]},{\"id\":\"com.sap.test:test-sap-internal-child:jar\",\"numericId\":2,\"scopes\":[\"compile\"]},{\"id\":\"com.google.code.findbugs:jsr305:jar\",\"numericId\":3,\"scopes\":[\"test\"]}] ,\"dependencies\":[{\"from\":\"com.sap.test:test-common:pom\",\"to\":\"com.fasterxml.jackson.core:jackson-databind\",\"numericFrom\":0,\"numericTo\":1}, {\"from\":\"com.sap.test:test-common:pom\",\"to\":\"com.sap.test:test-sap-internal-child\",\"numericFrom\":0,\"numericTo\":2}, {\"from\":\"com.sap.test:test-sap-internal-child\",\"to\":\"com.google.code.findbugs:jsr305\",\"numericFrom\":2,\"numericTo\":3}]}"
	dir, err := os.MkdirTemp("", "temp_directory")
	require.NoError(t, err, "Error while creating the temporary directory")
	defer os.RemoveAll(dir) // clean up

	// Error case
	_, err = GetChildParentRelationMap("", "com.sap.test/test-common")
	assert.Error(t, err, "Error expected for non-existing dependency graph file path")
	// Happy Path
	dependencyGraphFilePath := filepath.Join(dir, "test_GetChildParentRelationMap.json")
	os.WriteFile(dependencyGraphFilePath, []byte(dependency_graph_json), 0644)
	childParentMap, err := GetChildParentRelationMap(dependencyGraphFilePath, "com.sap.test/test-common")
	require.NoError(t, err)
	assert.Nil(t, childParentMap["com.fasterxml.jackson.core/jackson-databind"])
	assert.Equal(t, "com.sap.test/test-sap-internal-child", childParentMap["com.google.code.findbugs/jsr305"])
}

func TestGetFilesInPath(t *testing.T) {
	dependency_graph_json := " {\"artifacts\":[{\"id\":\"com.fasterxml.jackson.core:jackson-databind:jar\",\"numericId\":1,\"scopes\":[\"compile\"]},{\"id\":\"com.sap.test:test-sap-internal-child:jar\",\"numericId\":2,\"scopes\":[\"compile\"]},{\"id\":\"com.google.code.findbugs:jsr305:jar\",\"numericId\":3,\"scopes\":[\"test\"]}] ,\"dependencies\":[{\"from\":\"com.sap.test:test-common:pom\",\"to\":\"com.fasterxml.jackson.core:jackson-databind\",\"numericFrom\":0,\"numericTo\":1}, {\"from\":\"com.sap.test:test-common:pom\",\"to\":\"com.sap.test:test-sap-internal-child\",\"numericFrom\":0,\"numericTo\":2}, {\"from\":\"com.sap.test:test-sap-internal-child\",\"to\":\"com.google.code.findbugs:jsr305\",\"numericFrom\":2,\"numericTo\":3}]}"
	dir, err := os.MkdirTemp("", "temp_directory")
	require.NoError(t, err, "Error while creating the temporary directory")
	defer os.RemoveAll(dir) // clean up

	dependencyGraphFilePath := filepath.Join(dir, "test_GetChildParentRelationMap.json")
	os.WriteFile(dependencyGraphFilePath, []byte(dependency_graph_json), 0644)
	filePaths, err := getFilesInPath(dir, "test_GetChildParentRelationMap.json")
	require.NoError(t, err)
	assert.Len(t, filePaths, 1, "Expecting one file path to be present")
}

func TestGetPomArtifact(t *testing.T) {
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

	assert.Equal(t, GetPomArtifact(dependencyTreeList), "com.sap.test/test-common", "Expecting the com.sap.test/test-common")
}

func TestAddRelativeBuidDescriptorPathToJson(t *testing.T) {
	dependencyTreeGenerator := DependencyTreeGenerator{}
	dummy_json := " {\"artifacts\":[{\"id\":\"com.fasterxml.jackson.core:jackson-databind:jar\",\"numericId\":1,\"scopes\":[\"compile\"]},{\"id\":\"com.sap.test:test-sap-internal-child:jar\",\"numericId\":2,\"scopes\":[\"compile\"]},{\"id\":\"com.google.code.findbugs:jsr305:jar\",\"numericId\":3,\"scopes\":[\"test\"]}] ,\"dependencies\":[{\"from\":\"com.sap.test:test-common:pom\",\"to\":\"com.fasterxml.jackson.core:jackson-databind\",\"numericFrom\":0,\"numericTo\":1}, {\"from\":\"com.sap.test:test-common:pom\",\"to\":\"com.sap.test:test-sap-internal-child\",\"numericFrom\":0,\"numericTo\":2}, {\"from\":\"com.sap.test:test-sap-internal-child\",\"to\":\"com.google.code.findbugs:jsr305\",\"numericFrom\":2,\"numericTo\":3}]}"
	dummyRatingDetailsJson := []byte(dummy_json)
	buildDescriptorRelativePath := "dummyPath"
	ratingDetailsJsonWithFilePath, err := dependencyTreeGenerator.AddRelativeBuidDescriptorPathToJson(dummyRatingDetailsJson, buildDescriptorRelativePath)
	require.NoError(t, err)
	assert.Contains(t, string(ratingDetailsJsonWithFilePath), "\"filePath\":\"dummyPath\"", "Expecting the filePath to be available in json")
}

func TestGetAllPomFilesToBuildDependencyTreeWithoutOrphanPom(t *testing.T) {
	// init
	mavenStr := `<?xml version="1.0" encoding="UTF-8"?>
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
        <module>module2</module>
    </modules>
    <profiles>
      <profile>
        <id>java-inflector</id>
          <activation>
            <property>
              <name>env</name>
              <value>java</value>
            </property>
          </activation>
		  <modules>
             <module>profileModule</module>
          </modules>
      </profile>
    </profiles>
</project>`

	workingDir, err := os.MkdirTemp("", "temp_directory")
	require.NoError(t, err, "Error while creating the temporary directory")
	defer os.RemoveAll(workingDir) // clean up

	mavenDir, err := os.MkdirTemp(workingDir, "-maven")
	require.NoError(t, err, "Error while creating temporary maven directory")
	checkOutPath := filepath.Join(mavenDir, "")

	mavenFilePath := filepath.Join(mavenDir, "pom.xml")
	os.WriteFile(mavenFilePath, []byte(mavenStr), 0644)

	createTestMavenModule(mavenDir, "module1", moduleMavenStr)
	createTestMavenModule(mavenDir, "module2", moduleMavenStr)
	createTestMavenModule(mavenDir, "profileModule", moduleMavenStr)
	// test
	dependencyTreeGenerator := DependencyTreeGenerator{}
	pomFilesToBuildDependencyTree, _ := dependencyTreeGenerator.GetAllPomFilesToBuildDependencyTree(checkOutPath)
	// asserts
	assert.Len(t, pomFilesToBuildDependencyTree, 1, "Expecting one pom file path")
	assert.Contains(t, pomFilesToBuildDependencyTree, mavenFilePath, "Expecting correct pom file")
}

func TestGetAllPomFilesToBuildDependencyTreeWithOrphanPom(t *testing.T) {
	// init
	mavenStr := `<?xml version="1.0" encoding="UTF-8"?>
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
    <module>module2</module>
  </modules>
</project>`

	workingDir, err := os.MkdirTemp("", "temp_directory-")
	require.NoError(t, err, "Error while creating the temporary directory")
	defer os.RemoveAll(workingDir) // clean up

	mavenDir, err := os.MkdirTemp(workingDir, "maven-")
	require.NoError(t, err, "Error while creating temporary maven directory")

	checkOutPath := filepath.Join(mavenDir, "")

	mavenFilePath := filepath.Join(mavenDir, "pom.xml")
	os.WriteFile(mavenFilePath, []byte(mavenStr), 0644)
	require.FileExists(t, mavenFilePath)

	require.FileExists(t, createTestMavenModule(mavenDir, "module1", moduleMavenStr))
	require.FileExists(t, createTestMavenModule(mavenDir, "module2", moduleMavenStr))
	orphanModulePom := createTestMavenModule(mavenDir, "orphanModule", moduleMavenStr)
	require.FileExists(t, orphanModulePom)

	// test
	dependencyTreeGenerator := DependencyTreeGenerator{}
	pomFilesToBuildDependencyTree, _ := dependencyTreeGenerator.GetAllPomFilesToBuildDependencyTree(checkOutPath)
	// asserts
	assert.Len(t, pomFilesToBuildDependencyTree, 2, "Expecting two pom file paths")
	assert.Contains(t, pomFilesToBuildDependencyTree, mavenFilePath, "Expecting correct pom file")
	assert.Contains(t, pomFilesToBuildDependencyTree, orphanModulePom, "Expecting correct pom file")
}

func createTestMavenModule(projectDir string, moduleName string, moduleMavenStr string) string {
	moduleDir := filepath.Join(projectDir, moduleName)
	os.MkdirAll(moduleDir, 0755)

	modulePomPath := filepath.Join(moduleDir, "pom.xml")
	os.WriteFile(modulePomPath, []byte(moduleMavenStr), 0644)
	return modulePomPath
}

func TestGetAllPomFilesToBuildDependencyTreeErrorInSubModule(t *testing.T) {
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
	require.NoError(t, err, "Error while creating the temporary directory")
	defer os.RemoveAll(workingDir) // clean up

	mavenDir, err := os.MkdirTemp(workingDir, "-maven")
	require.NoError(t, err, "Error while creating the temporary directory")

	checkOutPath := filepath.Join(mavenDir, "")

	mavenFilePath := filepath.Join(mavenDir, "pom.xml")
	os.WriteFile(mavenFilePath, []byte(malformedMavenStr), 0644)

	dependencyTreeGenerator := DependencyTreeGenerator{}
	_, err = dependencyTreeGenerator.GetAllPomFilesToBuildDependencyTree(checkOutPath)
	assert.Error(t, err, "Expecting error")
}

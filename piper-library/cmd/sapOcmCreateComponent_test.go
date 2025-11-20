//go:build unit
// +build unit

package cmd

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	piperOsCmd "github.com/SAP/jenkins-library/cmd"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/stretchr/testify/assert"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/ocm"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/staging"
	"gopkg.in/yaml.v3"
)

type sapOcmCreateComponentMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func (s sapOcmCreateComponentMockUtils) prepareStagingRepository(cfg *sapOcmCreateComponentOptions, info ocm.ComponentInfo, commonPipelineEnvironment *sapOcmCreateComponentCommonPipelineEnvironment) (*staging.Staging, error) {
	ri := ocm.StagingRepoInfo{
		RepositoryURL: "https://mock.test.repository.cloud.sap",
		User:          "mock test user",
		Password:      "mock test password",
	}
	info.Set("RepoInfo", &ri)
	info["RepositoryURL"] = ocm.TrimHttpPrefix(ri.RepositoryURL)
	return nil, nil
}

func newSapOcmCreateComponentTestsUtils() sapOcmCreateComponentMockUtils {
	utils := sapOcmCreateComponentMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}

	return utils
}

const (
	testCompGitRepo    = "${OCI_NAME_0}"
	testCompGitRepoURL = "${gitURL}"
	testCompVersion    = "${artifactVersion}"
	testCompCommitID   = "${gitCommitID}"
	testImageName1     = "${OCI_NAME_0}"
	testImageName2     = "${OCI_NAME_1}"
)

var testConfigSingle = sapOcmCreateComponentOptions{
	ComponentConstructorPath:  "gen/" + ocm.ComponentConstructorFileName,
	ContainerRegistryURL:      "https://abc123.staging.repositories.cloud.sap",
	ContainerRegistryUser:     "kanikoUser",
	ContainerRegistryPassword: "kanikoSecret",
	ArtifactVersion:           testCompVersion,
	GitRepository:             testCompGitRepo,
	GitCommitID:               testCompCommitID,
	GitOrg:                    "open-component-model",
	GitURL:                    testCompGitRepoURL,
	ImageNames:                []string{testCompGitRepo},
	ImageNameTags:             []string{testCompGitRepo + ":" + testCompVersion},
	FailOnError:               true,
	Provider:                  "${provider}",
	ComponentName:             "${componentName}",
	GenDir:                    "gen",
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Helper functions:

// create the ocm input file from the given input data coming from user or output from previous steps
func executeComponentsGeneration(t *testing.T, testConfig *sapOcmCreateComponentOptions, ub SapOcmCreateComponentUtils) map[string]interface{} {
	var compMap map[string]interface{}
	info := &ocm.ComponentInfo{}

	mockStage()
	err := createOcmConfigForCredentials(ub, testConfig, *info)
	assert.NoError(t, err)
	err = createDirs(ub, testConfig.GenDir)
	assert.NoError(t, err)
	cpe := &sapOcmCreateComponentCommonPipelineEnvironment{}
	ocmFiles := []piperutils.Path{}
	err = prepareComponentsConstructor(ub, testConfig, *info, cpe, ocmFiles)
	assert.NoError(t, err)
	assert.NotNil(t, info)

	// Read the generated file
	compYamlPath := testConfig.ComponentConstructorPath
	if len(compYamlPath) == 0 {
		compYamlPath = ocm.ComponentConstructorFileName
	}

	compYamlFileExists, _ := ub.FileExists(compYamlPath)
	assert.True(t, compYamlFileExists)
	data, err := ub.ReadFile(compYamlPath)
	assert.NoError(t, err)

	// Unmarshal the YAML data into the map
	err = yaml.Unmarshal(data, &compMap)
	assert.NoError(t, err)

	return compMap
}

func getComponent(compMap map[string]interface{}) map[string]interface{} {
	comps := compMap["components"].([]interface{})
	return comps[0].(map[string]interface{})
}

func getComponentName(compMap map[string]interface{}) string {
	comp := getComponent(compMap)
	return comp["name"].(string)
}

func getProviderName(compMap map[string]interface{}) string {
	comp := getComponent(compMap)
	provider := comp["provider"].(map[string]interface{})
	return provider["name"].(string)
}

func getSource(comp map[string]interface{}) map[string]interface{} {
	sources := comp["sources"].([]interface{})
	return sources[0].(map[string]interface{})
}

func getResource(comp map[string]interface{}, index int) map[string]interface{} {
	resources := comp["resources"].([]interface{})
	return resources[index].(map[string]interface{})
}

func getResourceCount(comp map[string]interface{}) int {
	resources := comp["resources"].([]interface{})
	return len(resources)
}

func getFirstResource(comp map[string]interface{}) map[string]interface{} {
	return getResource(comp, 0)
}

func getAccess(srcOrResource map[string]interface{}) map[string]interface{} {
	access := srcOrResource["access"].(map[string]interface{})
	return access
}

func mockOrchestrator(t *testing.T) {
	err := os.Setenv("JENKINS_HOME", "test_jenkins_home") // pretend to be on Jenkins so that stage can be found
	assert.NoError(t, err)
	err = os.Setenv("JENKINS_URL", "http://test_jenkins_url")
	assert.NoError(t, err)
}

func unsetOrchestrator(t *testing.T) {
	err := os.Unsetenv("JENKINS_HOME")
	assert.NoError(t, err)
	err = os.Unsetenv("JENKINS_URL")
	assert.NoError(t, err)
}

func mockStage() {
	piperOsCmd.GeneralConfig.StageName = "Central Build"
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestRunSapOcmCreateComponent(t *testing.T) {
	t.Parallel()

	t.Run("single container image", func(t *testing.T) {
		utils := newSapOcmCreateComponentTestsUtils()

		// test and create a map from generated yaml to hold the YAML data
		compMap := executeComponentsGeneration(t, &testConfigSingle, utils)

		component := getComponent(compMap)
		assert.Equal(t, "${componentName}", component["name"])
		assert.Equal(t, testCompVersion, component["version"])
		assert.Equal(t, "${provider}", getProviderName(compMap))

		source := getSource(component)
		assert.Equal(t, "src", source["name"])
		assert.Equal(t, "filesystem", source["type"])
		sourceAccess := getAccess(source)
		assert.Equal(t, "gitHub", sourceAccess["type"])
		assert.Equal(t, testCompCommitID, sourceAccess["commit"])
		assert.Equal(t, testCompGitRepoURL, sourceAccess["repoUrl"])

		resource := getFirstResource(component)
		assert.Equal(t, 1, getResourceCount(component))
		assert.Equal(t, testCompGitRepo, resource["name"])
		assert.Equal(t, "${artifactVersion}", resource["version"])
		resourceAccess := getAccess(resource)
		assert.Equal(t, "ociArtifact", resourceAccess["type"])
		assert.Equal(t, "${OCI_REFERENCE_0}", resourceAccess["imageReference"])
	})

	t.Run("multi container images", func(t *testing.T) {
		testConfigMulti := testConfigSingle
		testConfigMulti.ImageNames = []string{testImageName1, testImageName2}
		testConfigMulti.ImageNameTags = []string{testImageName1 + ":" + testCompVersion, testImageName2 + ":" + testCompVersion}

		utils := newSapOcmCreateComponentTestsUtils()

		compMap := executeComponentsGeneration(t, &testConfigMulti, utils)
		component := getComponent(compMap)

		noResources := getResourceCount(component)
		assert.Equal(t, noResources, 2)
		for i := 0; i < noResources; i++ {
			resource := getResource(component, i)
			assert.Equal(t, testConfigMulti.ImageNames[i], resource["name"])
			assert.Equal(t, "${artifactVersion}", resource["version"])
			resourceAccess := getAccess(resource)
			assert.Equal(t, resourceAccess["type"], "ociArtifact")
			assert.Equal(t, "${OCI_REFERENCE_"+strconv.Itoa(i)+"}", resourceAccess["imageReference"])
		}
	})

	t.Run("configured names", func(t *testing.T) {
		utils := newSapOcmCreateComponentTestsUtils()
		compMap := executeComponentsGeneration(t, &testConfigSingle, utils)
		assert.Equal(t, getProviderName(compMap), testConfigSingle.Provider)
		assert.Equal(t, testConfigSingle.ComponentName, getComponentName(compMap))
	})

	t.Run("ocm config generated", func(t *testing.T) {
		utils := newSapOcmCreateComponentTestsUtils()
		home, err := os.UserHomeDir()
		assert.NoError(t, err)
		ocmConfigPath := home + "/.ocmconfig"
		compYamlFileExists, _ := utils.FileExists(ocmConfigPath)
		assert.False(t, compYamlFileExists)
		_ = executeComponentsGeneration(t, &testConfigSingle, utils)
		compYamlFileExists, _ = utils.FileExists(ocmConfigPath)
		assert.True(t, compYamlFileExists)

		data, err := utils.ReadFile(ocmConfigPath)
		assert.NoError(t, err)

		// Unmarshal the YAML data into the map
		var config ocm.Config
		err = yaml.Unmarshal(data, &config)
		assert.NoError(t, err)

		host, _ := strings.CutPrefix(testConfigSingle.ContainerRegistryURL, "https://")
		assert.Equal(t, "generic.config.ocm.software/v1", config.Type)
		assert.Equal(t, 1, len(config.Configurations))
		cred := config.Configurations[0]
		assert.Equal(t, "credentials.config.ocm.software", cred.Type)
		assert.Equal(t, 1, len(cred.Consumers))
		assert.Equal(t, ocm.Oci, cred.Consumers[0].Identity.Type)
		assert.Equal(t, host, cred.Consumers[0].Identity.Hostname)
		assert.Equal(t, 1, len(cred.Consumers[0].Credentials))
		assert.Equal(t, "Credentials", cred.Consumers[0].Credentials[0].Type)
		assert.Equal(t, testConfigSingle.ContainerRegistryUser, cred.Consumers[0].Credentials[0].Properties.Username)
		assert.Equal(t, testConfigSingle.ContainerRegistryPassword, cred.Consumers[0].Credentials[0].Properties.Password)

		// Test without separate OCM registry --> just need one credential:
		testConfigNoOcm := testConfigSingle
		utils = newSapOcmCreateComponentTestsUtils() // reset environment
		_ = executeComponentsGeneration(t, &testConfigNoOcm, utils)
		compYamlFileExists, _ = utils.FileExists(ocmConfigPath)
		assert.True(t, compYamlFileExists)
		// Unmarshal the YAML data into the map
		config = ocm.Config{} // empty structure
		data, err = utils.ReadFile(ocmConfigPath)
		assert.NoError(t, err)
		err = yaml.Unmarshal(data, &config)
		assert.NoError(t, err)
		cred = config.Configurations[0]
		assert.Equal(t, len(cred.Consumers), 1)
		assert.Equal(t, 1, len(cred.Consumers))
		assert.Equal(t, ocm.Oci, cred.Consumers[0].Identity.Type)
		assert.Equal(t, host, cred.Consumers[0].Identity.Hostname)
		assert.Equal(t, 1, len(cred.Consumers[0].Credentials))
		assert.Equal(t, "Credentials", cred.Consumers[0].Credentials[0].Type)
		assert.Equal(t, testConfigSingle.ContainerRegistryUser, cred.Consumers[0].Credentials[0].Properties.Username)
		assert.Equal(t, testConfigSingle.ContainerRegistryPassword, cred.Consumers[0].Credentials[0].Properties.Password)
	})

	t.Run("fail parameter", func(t *testing.T) {
		info := &ocm.ComponentInfo{}
		utils := newSapOcmCreateComponentTestsUtils()
		errorCfg := &sapOcmCreateComponentOptions{}
		errorCfg.FailOnError = true
		cpe := &sapOcmCreateComponentCommonPipelineEnvironment{}
		ocmFiles := []piperutils.Path{}
		err := prepareComponentsConstructor(utils, errorCfg, *info, cpe, ocmFiles)
		assert.NotNil(t, err)
	})

	templatedComponentsYaml := `components:
- name: ${gitHost}/${gitOrg}/${gitRepository}
  version: ${artifactVersion}
  provider:
    name: ${provider}
  sources:
  - name: sources
    version: ${artifactVersion}
    type: filesystem
    access:
      type: gitHub
      repoUrl: ${gitURL}
      commit: ${gitCommitID}
  resources:
  - name: ${imageName}
    type: ociImage
    version: ${artifactVersion}
    access:
      type: ociArtifact
      imageReference: ${registryURL}/${imageNameTag}
`

	t.Run("templated "+ocm.ComponentConstructorFileName, func(t *testing.T) {
		var compMap map[string]interface{}
		compYamlPath := filepath.Join(testConfigSingle.GenDir, ocm.ComponentConstructorFileName)
		settingsPath := filepath.Join(testConfigSingle.GenDir, ocm.SettingsFileName)
		utils := newSapOcmCreateComponentTestsUtils()
		mockOrchestrator(t)
		defer unsetOrchestrator(t)
		err := utils.WriteFile(compYamlPath, []byte(templatedComponentsYaml), 0644)
		assert.NoError(t, err)

		cpe := &sapOcmCreateComponentCommonPipelineEnvironment{}

		mockStage()
		err = runSapOcmCreateComponent(utils, &testConfigSingle, cpe)
		assert.NoError(t, err) // OCM CLI will fail (expected)

		// check that file was not overwritten
		data, err := utils.ReadFile(compYamlPath)
		assert.NoError(t, err)
		compYaml := string(data)
		assert.Equal(t, templatedComponentsYaml, compYaml)

		// check that settings were written correctly:
		data, err = utils.ReadFile(settingsPath)
		assert.NoError(t, err)
		assert.NotEmpty(t, string(data))
		err = yaml.Unmarshal(data, &compMap)
		assert.NoError(t, err)

		nameParts := strings.Split(testConfigSingle.ImageNameTags[0], ":")
		assert.Equal(t, compMap["containerRegistryURL"], testConfigSingle.ContainerRegistryURL)
		assert.Equal(t, compMap["gitCommitID"], testConfigSingle.GitCommitID)
		assert.Equal(t, compMap["gitOrg"], testConfigSingle.GitOrg)
		assert.Equal(t, compMap["gitRepository"], testConfigSingle.GitRepository)
		assert.Equal(t, compMap["gitURL"], testConfigSingle.GitURL)
		assert.Equal(t, compMap["imageNames"].([]interface{})[0], nameParts[0])
		assert.Equal(t, compMap["artifactVersion"], testConfigSingle.ArtifactVersion)
		_, ok := compMap["OcmUser"]
		assert.False(t, ok)
		_, ok = compMap["OcmPassword"]
		assert.False(t, ok)
		_, ok = compMap["ContainerRegistryUser"]
		assert.False(t, ok)
		_, ok = compMap["ContainerRegistryPassword"]
		assert.False(t, ok)
	})

	t.Run("missing required parameters - ContainerRegistryURL not set", func(t *testing.T) {
		mockOrchestrator(t)
		defer unsetOrchestrator(t)
		mockStage()

		utils := newSapOcmCreateComponentTestsUtils()
		cfg := sapOcmCreateComponentOptions{}
		err := runSapOcmCreateComponent(utils, &cfg, &sapOcmCreateComponentCommonPipelineEnvironment{})
		assert.NoError(t, err)
	})

	t.Run("invalid stage", func(t *testing.T) {
		mockOrchestrator(t)
		defer unsetOrchestrator(t)
		piperOsCmd.GeneralConfig.StageName = "doenstExist"
		utils := newSapOcmCreateComponentTestsUtils()
		cfg := testConfigSingle
		cfg.ContainerRegistryURL = ""
		err := runSapOcmCreateComponent(utils, &cfg, &sapOcmCreateComponentCommonPipelineEnvironment{})
		assert.NoError(t, err)
	})

	t.Run("successful component creation with Helm chart", func(t *testing.T) {
		mockOrchestrator(t)
		defer unsetOrchestrator(t)
		mockStage()
		utils := newSapOcmCreateComponentTestsUtils()
		cfg := testConfigSingle
		cfg.HelmChartURL = "https://example.com/helm-chart"
		compMap := executeComponentsGeneration(t, &cfg, utils)
		component := getComponent(compMap)
		assert.Equal(t, "${componentName}", component["name"])
		assert.Equal(t, testCompVersion, component["version"])
		assert.Equal(t, "${provider}", getProviderName(compMap))
	})

	t.Run("do not run on PRs", func(t *testing.T) {
		mockOrchestrator(t)
		defer unsetOrchestrator(t)
		_ = os.Setenv("CHANGE_ID", "fake PR")
		defer os.Unsetenv("CHANGE_ID")

		utils := newSapOcmCreateComponentTestsUtils()
		cfg := testConfigSingle
		err := runSapOcmCreateComponent(utils, &cfg, &sapOcmCreateComponentCommonPipelineEnvironment{})
		assert.NoError(t, err)
	})

	t.Run("do not run on wrong stage", func(t *testing.T) {
		mockOrchestrator(t)
		defer unsetOrchestrator(t)
		piperOsCmd.GeneralConfig.StageName = "doesntExist"

		utils := newSapOcmCreateComponentTestsUtils()
		cfg := testConfigSingle
		err := runSapOcmCreateComponent(utils, &cfg, &sapOcmCreateComponentCommonPipelineEnvironment{})
		assert.NoError(t, err)
	})

	t.Run("create NewSapOcmCreateComponentUtils()", func(t *testing.T) {
		utils := NewSapOcmCreateComponentUtils()
		assert.NotNil(t, utils)
	})

	t.Run("create NewSapOcmCreateComponentUtils()", func(t *testing.T) {
		cfg := sapOcmCreateComponentOptions{}
		assert.NotNil(t, cfg)
		assert.False(t, cfg.isValidStep())
		mockStage()
		assert.False(t, cfg.isValidStep())
		cfg.HelmChartURL = "https://example.com/helm-chart"
		assert.True(t, cfg.isValidStep())
	})

	t.Run("test isValidDockerConfigJSON()", func(t *testing.T) {
		utils := newSapOcmCreateComponentTestsUtils()
		dir := t.TempDir()
		file := filepath.Join(dir, "does-not-exist.json")
		err := isValidDockerConfigJSON(utils, file)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "dockerConfigJSON file does not exist")

		json := []byte(`content`)
		file = filepath.Join(dir, "not-a-json.json")
		err = utils.WriteFile(file, json, 0644)
		assert.NoError(t, err)
		err = isValidDockerConfigJSON(utils, file)
		assert.Error(t, err)

		json = []byte(`{"key2": "value"}`)
		file = filepath.Join(dir, "not-a-docker-config.json")
		err = utils.WriteFile(file, json, 0644)
		assert.NoError(t, err)
		err = isValidDockerConfigJSON(utils, file)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "dockerConfigJSON file does not contain 'auths' key")

		json = []byte(`{"auths":{"oci-repo-url":{"auth":"encoded-credentials"}}}`)
		file = filepath.Join(dir, "docker-config.json")
		err = utils.WriteFile(file, json, 0644)
		assert.NoError(t, err)
		err = isValidDockerConfigJSON(utils, file)
		assert.NoError(t, err)
	})

	t.Run("ocm config contains docker config", func(t *testing.T) {
		utils := newSapOcmCreateComponentTestsUtils()
		home, err := os.UserHomeDir()
		assert.NoError(t, err)

		// Check that the ocm config file does not exist.
		ocmConfigPath := filepath.Join(home, "/.ocmconfig")
		exists, err := utils.FileExists(ocmConfigPath)
		assert.NoError(t, err)
		assert.False(t, exists)

		// Generate the docker config file.
		dockerConfigPath := filepath.Join(home, "docker-config.json")
		exists, err = utils.FileExists(dockerConfigPath)
		assert.NoError(t, err)
		assert.False(t, exists)

		json := []byte(`{"auths":{"oci-repo-url":{"auth":"encoded-credentials"}}}`)
		err = utils.WriteFile(dockerConfigPath, json, 0644)
		assert.NoError(t, err)

		// Create new test config referring to docker config file (shallow copy is sufficient).
		newTestConfig := testConfigSingle
		newTestConfig.DockerConfigJSON = dockerConfigPath
		assert.NotEqual(t, testConfigSingle.DockerConfigJSON, dockerConfigPath)

		// Generate the ocm config file.
		info := ocm.ComponentInfo{}
		mockStage()
		err = createOcmConfigForCredentials(utils, &newTestConfig, info)
		assert.NoError(t, err)
		exists, err = utils.FileExists(ocmConfigPath)
		assert.True(t, exists)

		// Read the generated ocm config file.
		data, err := utils.ReadFile(ocmConfigPath)
		assert.NoError(t, err)
		var ocmConfig ocm.Config
		err = yaml.Unmarshal(data, &ocmConfig)
		assert.NoError(t, err)

		// Check that the generated ocm config file refers to the docker config file.
		assert.Equal(t, ocmConfig.Configurations[0].Repositories[0].Repository.DockerConfigFile, dockerConfigPath)
	})
}

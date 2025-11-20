package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/SAP/jenkins-library/cmd"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/config"
	jHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/ocm"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/staging"
	"gopkg.in/yaml.v3"
)

// ///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// SapOcmCreateComponentUtils interface for command execution, file operations and staging service interaction.
type SapOcmCreateComponentUtils interface {
	// prepareStagingRepository creates a staging repository in the staging service and stores the ocm.StagingRepoInfo in the ocm.ComponentInfo.
	prepareStagingRepository(cfg *sapOcmCreateComponentOptions, info ocm.ComponentInfo, commonPipelineEnvironment *sapOcmCreateComponentCommonPipelineEnvironment) (*staging.Staging, error)
	WriteFile(path string, content []byte, perm os.FileMode) error
	piperutils.FileUtils
	command.ExecRunner
}

// ///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// SapOcmCreateComponentUtilsBundle bundles command execution, file operations and staging service interaction. Implements SapOcmCreateComponentUtils.
type SapOcmCreateComponentUtilsBundle struct {
	*piperutils.Files
	*command.Command
}

// NewSapOcmCreateComponentUtils creates a new instance of SapOcmCreateComponentUtils. Command errors are rerouted to the logging framework.
func NewSapOcmCreateComponentUtils() SapOcmCreateComponentUtils {
	utils := SapOcmCreateComponentUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Always reroute command errors to logging framework
	utils.Stderr(log.Writer())
	return &utils
}

// prepareStagingRepository creates a staging repository in the staging service and stores the ocm.StagingRepoInfo in the ocm.ComponentInfo.
func (ub SapOcmCreateComponentUtilsBundle) prepareStagingRepository(cfg *sapOcmCreateComponentOptions, info ocm.ComponentInfo, commonPipelineEnvironment *sapOcmCreateComponentCommonPipelineEnvironment) (*staging.Staging, error) {
	log.Entry().Debugf("Found Staging GroupId: %s", cfg.StagingGroupID)

	stagingObj, err := ub.loginToStagingService(cfg.StagingGroupID) // how to mock this in tests?
	if err != nil {
		return stagingObj, err
	}

	log.Entry().Info("Create Staging OCM repository")
	stagingObj.OutputFile = "ocm-repo.json"
	repoJson, err := stagingObj.CreateStagingRepository()
	if err != nil {
		log.Entry().Errorf("Failed to create staging repository: %v", err)
		return stagingObj, err
	}

	var ri ocm.StagingRepoInfo
	err = json.Unmarshal([]byte(repoJson), &ri)
	if err != nil {
		log.Entry().Errorf("Failed to unmarshal repositoryInfo: %v", err)
		return stagingObj, err
	}
	log.RegisterSecret(ri.User)
	log.RegisterSecret(ri.Password)
	logGeneratedFile(ub, stagingObj.OutputFile) // file contains secrets!!!
	log.Entry().Debugf("New Staging repository created: %+v", ri)
	info.Set("RepoInfo", &ri)
	info["RepositoryURL"] = ocm.TrimHttpPrefix(ri.RepositoryURL)
	stagingObj.RepositoryId = ri.Repository

	if commonPipelineEnvironment.custom.isCustomComponentDescriptor {
		stagingObj.Metadata = "{\"isCustomComponentDescriptor\":true}"
		err = updateRepositoryMetadata(stagingObj)
		if err != nil {
			log.Entry().Errorf("Failed to updateRepositoryMetadata: %v", err)
			return stagingObj, err
		}
		log.Entry().Debugf("Updated repository metadata with: %s", stagingObj.Metadata)
	}

	return stagingObj, nil
}

// loginToStagingService logs into the staging service and returns a staging object.
func (ub SapOcmCreateComponentUtilsBundle) loginToStagingService(groupID string) (*staging.Staging, error) {
	stageCmd := SapCallStagingServiceCommand()
	metadata := sapCallStagingServiceMetadata()
	var stagingConfig sapCallStagingServiceOptions
	err := cmd.PrepareConfig(stageCmd, &metadata, metadata.Metadata.Name, &stagingConfig, config.OpenPiperFile)
	if err != nil {
		log.Entry().WithError(err).Errorf("Error getting staging config: %s", err.Error())
		return nil, err
	}

	log.Entry().Debugf("Staging profile name: %s, api-url: %s", stagingConfig.Profile, stagingConfig.Url)
	log.Entry().Debugf("Staging config %+v", stagingConfig)

	stagingObj := staging.Staging{
		TenantId:         stagingConfig.TenantID,
		TenantSecret:     stagingConfig.TenantSecret,
		Username:         stagingConfig.Username,
		Password:         stagingConfig.Password,
		Profile:          stagingConfig.Profile,
		RepositoryFormat: "ocm",
		Url:              stagingConfig.Url,
		HTTPClient:       &jHttp.Client{},
	}
	if groupID != "" {
		log.Entry().Debugf("login method called with existing group: %s", groupID)
		stagingObj.Group = groupID // login to existing group
	}

	log.Entry().Debugf("Login to staging service: %v", stagingObj)
	_, err = stagingObj.LoginAndReceiveAuthToken()
	if err != nil {
		log.Entry().WithError(err).Errorf("Error getting auth-token: %s", err.Error())
		return nil, err
	}

	if groupID == "" {
		log.Entry().Debug("No 'StagingGroupID' found, creating a new staging group.")
		group, er := stagingObj.CreateStagingGroup()
		if er != nil {
			return &stagingObj, er
		}
		log.Entry().Debugf("Staging group created: '%s'", group)
		stagingObj.SetGroup(group)
	}

	return &stagingObj, nil
}

// ///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// sapOcmCreateComponent is the entry point for the piper step.
func sapOcmCreateComponent(cfg sapOcmCreateComponentOptions, _ *telemetry.CustomData, commonPipelineEnvironment *sapOcmCreateComponentCommonPipelineEnvironment) {
	log.Entry().Info("Start creating OCM Component Version")
	log.Entry().Debugf("sapOcmCreateComponentOptions: %+v", cfg)
	log.Entry().Debugf("sapOcmCreateComponentCommonPipelineEnvironment: %+v", commonPipelineEnvironment)

	utils := NewSapOcmCreateComponentUtils()

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runSapOcmCreateComponent(utils, &cfg, commonPipelineEnvironment)
	if err != nil {
		if cfg.FailOnError {
			log.Entry().WithError(err).Fatal("OCM step execution failed")
		} else {
			log.Entry().WithError(err).Error("OCM step execution failed")
		}
	}
}

// runSapOcmCreateComponent is the main function for the OCM step.
func runSapOcmCreateComponent(ub SapOcmCreateComponentUtils, cfg *sapOcmCreateComponentOptions, commonPipelineEnvironment *sapOcmCreateComponentCommonPipelineEnvironment) error {
	info := ocm.NewComponentInfo() // create a new generic component info map
	cfg.copyToComponentInfo(info)  // copy all settings from sapOcmCreateComponentOptions to info

	log.Entry().Infof("Create OCM Component Version: %+v", cfg)
	log.Entry().Debugf("sapOcmCreateComponent called in stage: %s", ocm.StageName())

	provider, err := orchestrator.GetOrchestratorConfigProvider(nil)
	if err != nil {
		log.Entry().Errorf("Failed to create cfg provider: %s", err.Error())
		return err
	}

	if provider.IsPullRequest() {
		log.Entry().Warn("OCM called on a PR, check configuration")
		if cfg.EnableOnPr {
			log.Entry().Info("OCM enabled for PRs ('enableOnPR=true'), continue with step")
		} else {
			log.Entry().Info("OCM disabled for PRs ('enableOnPR=false'), skipping step")
			return nil
		}
	}

	// check if this is an ocm-supported build step and return otherwise
	if !cfg.isValidStep() {
		log.Entry().Error("OCM create component version not supported. Unsupported stage or build type. Exiting step")
		return nil
	}

	err = createDirs(ub, cfg.GenDir)
	if err != nil {
		return err
	}

	var ocmFiles []piperutils.Path
	err = writeSettingsFile(ub, cfg, info, commonPipelineEnvironment, ocmFiles)
	if err != nil {
		return err
	}

	// create a component-constructor.yaml file if it does not exist:
	err = prepareComponentsConstructor(ub, cfg, info, commonPipelineEnvironment, ocmFiles)
	if err != nil {
		return err
	}

	stagingObj, err := ub.prepareStagingRepository(cfg, info, commonPipelineEnvironment)
	if err != nil {
		return err
	}

	err = createOcmConfigForCredentials(ub, cfg, info)
	if err != nil {
		return err
	}

	err = callOcmCli(ub, cfg, info, commonPipelineEnvironment)
	if err != nil {
		return err
	}

	arg := fmt.Sprintf("%s//%s", info["RepositoryURL"], info["CV"])
	log.Entry().Infof("You can view the component using command: 'ocm get componentversion -o yaml %s", arg)

	err = executeCommand(ub, "get", "componentversion", "-o", "yaml", arg)
	if err != nil {
		log.Entry().Error(err.Error())
	}

	if info["RepoInfo"] != nil {
		// write repositoryInfo to output parameter
		jsonData, err := json.Marshal(info["RepoInfo"])
		if err != nil {
			log.Entry().WithError(err).Errorf("Error while Marshaling. %v", err)
		}
		commonPipelineEnvironment.custom.ocmStagingRepo = string(jsonData)
	}

	ocm.DebugContainerRegistry(cfg.ContainerRegistryURL, cfg.ContainerRegistryUser, cfg.ContainerRegistryPassword)
	ocm.DebugStaging(stagingObj)

	return nil
}

// copyToComponentInfo copies all settings from sapOcmCreateComponentOptions to info
func (o *sapOcmCreateComponentOptions) copyToComponentInfo(info ocm.ComponentInfo) {
	val := reflect.ValueOf(o).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Name
		fieldName = strings.ToLower(fieldName[:1]) + fieldName[1:]
		// Skip password fields
		if strings.Contains(strings.ToLower(fieldName), "password") {
			log.RegisterSecret(field.String())
			continue
		}
		if strings.Contains(strings.ToLower(fieldName), "username") {
			log.RegisterSecret(field.String())
		}
		// Skip boolean fields and empty strings
		if (field.Kind() == reflect.Bool) || (field.Kind() == reflect.String && field.String() == "") {
			continue
		}
		info[fieldName] = field.Interface()
	}
}

// isValidStep checks if the current stage is supported by the OCM step.
func (o *sapOcmCreateComponentOptions) isValidStep() bool {
	switch ocm.Stage() {
	case ocm.Build:
		if o.ComponentConstructorPath != "" {
			// let's assume this is a custom component build
			return true
		}
		log.Entry().Info("'ComponentConstructorPath' not set. This is no product component version build.")
		if o.ContainerRegistryURL != "" {
			// let's assume this is a container build
			return true
		}
		log.Entry().Info("'ContainerRegistryURL' not set. This is no container image build.")
		if o.HelmChartURL != "" {
			// let's assume this is a helm-chart build
			return true
		}
		log.Entry().Info("'HelmChartURL' not set. This is no helm-chart build.")
	default:
		log.Entry().Errorf("Unknown stage: %s - exiting", ocm.StageName())
		return false
	}
	return false
}

// validateInputParameters checks if all required parameters are set.
func (o *sapOcmCreateComponentOptions) validateInputParameters() error {
	var missingParameters []string
	if !ocm.IsSet(o.ArtifactVersion) {
		missingParameters = append(missingParameters, "ArtifactVersion")
	}
	if !ocm.IsSet(o.GitURL) {
		missingParameters = append(missingParameters, "GitURL")
	}
	if !ocm.IsSet(o.GitCommitID) {
		missingParameters = append(missingParameters, "GitCommitID")
	}
	if !ocm.IsSet(o.GitRepository) {
		missingParameters = append(missingParameters, "GitRepository")
	}
	if len(missingParameters) > 0 {
		errorMessage := fmt.Sprintf("Missing parameters: %s", strings.Join(missingParameters, ", "))
		log.Entry().Error(errorMessage)
		return errors.New(errorMessage)
	} else {
		return nil
	}
}

// componentName the component name is constructed from the GitOrg and GitRepository if no componentName is provided.
// It has to match the following regex: ^[a-z][-a-z0-9]*([.][a-z][-a-z0-9]*)*[.][a-z]{2,}(/[a-z][-a-z0-9_]*([.][a-z][-a-z0-9_]*)*)+$
func (o *sapOcmCreateComponentOptions) componentName() string {
	if ocm.IsSet(o.ComponentName) {
		return o.ComponentName
	}
	return strings.ToLower(o.GitOrg + ".sap.com/" + o.GitRepository)
}

// provider returns Provider from sapOcmCreateComponentOptions if set, otherwise: SAP SE (GitOrg).
func (o *sapOcmCreateComponentOptions) provider() string {
	if ocm.IsSet(o.Provider) {
		return o.Provider
	}
	return "SAP SE (" + o.GitOrg + ")"
}

// version returns the artifact version from sapOcmCreateComponentOptions with '_' replaced by '-'.
// This is required to match the semver format (pattern '^[v]?(0|[1-9]\\d*)(?:\\.(0|[1-9]\\d*))?(?:\\.(0|[1-9]\\d*))?(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$').
// Initial version of this functions replaced all '_' with '+', but '+' is not allowed in OCI tags.
func (o *sapOcmCreateComponentOptions) version() string {
	return strings.ReplaceAll(o.ArtifactVersion, "_", "-")
}

// containerRegistry returns the container registry URL without the http(s) prefix.
func (o *sapOcmCreateComponentOptions) containerRegistry() string {
	return ocm.TrimHttpPrefix(o.ContainerRegistryURL)
}

// ociImageResourceName returns the OCI image resource name for the given index and ensures that the name does not conflict with the helm-chart.
func (o *sapOcmCreateComponentOptions) ociImageResourceName(i int, info ocm.ComponentInfo) string {
	if info["HELM_CHART"] != o.ImageNames[i] {
		return o.ImageNames[i]
	}
	return o.ImageNames[i] + "-image"
}

// ///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// OCM helpers

// prepareComponentsConstructor prepares the component-constructor file.
func prepareComponentsConstructor(ub SapOcmCreateComponentUtils, cfg *sapOcmCreateComponentOptions, info ocm.ComponentInfo, commonPipelineEnvironment *sapOcmCreateComponentCommonPipelineEnvironment, ocmFiles []piperutils.Path) error {
	filePath := cfg.ComponentConstructorPath
	if !ocm.IsSet(filePath) {
		filePath = filepath.Join(cfg.GenDir, ocm.ComponentConstructorFileName)
		info.Set("componentConstructorPath", filePath)
		log.Entry().Debugf("No ocm component-constructor file provided, defaulting to %s", filePath)
		ocmFiles = append(ocmFiles, piperutils.Path{
			Name:      "OCM component-constructor",
			Target:    ocm.ComponentConstructorFileName,
			Mandatory: false,
			Scope:     "job",
		})
	}
	err := generateComponentsYaml(ub, cfg, info, filePath, commonPipelineEnvironment)
	if err != nil {
		return err
	}
	log.Entry().Infof("Using ocm input file: %s", filePath)
	return nil
}

// writeSettingsFile writes the settings file to the given path.
func writeSettingsFile(ub SapOcmCreateComponentUtils, cfg *sapOcmCreateComponentOptions, info ocm.ComponentInfo, commonPipelineEnvironment *sapOcmCreateComponentCommonPipelineEnvironment, ocmFiles []piperutils.Path) error {
	settingsPath := filepath.Join(cfg.GenDir, ocm.SettingsFileName)
	info["SettingsPath"] = settingsPath

	info.Set("componentName", cfg.componentName())
	info.Set("artifactVersion", cfg.version()) // make the piper version number semver compatible
	info.Set("provider", cfg.provider())
	info.Set("containerRegistry", cfg.containerRegistry())
	info.Set("CV", fmt.Sprintf("%s:%s", cfg.componentName(), cfg.version())) // component version

	// ensure to set 'HELM_CHART' before continuing with 'OCI_NAME_?', because otherwise we might run into issues:
	// `Error: duplicate resource identity "name"=` ...
	if ocm.IsSet(cfg.HelmChartURL) {
		// let's see if we can extract the chart-name and version from the URL
		helmRepository, helmChartName, helmChartVersion, er := ocm.ReadNameAndVersionFromUrl(cfg.HelmChartURL)
		if er != nil {
			log.Entry().Errorf("Failed to parse chart name and version: %s", er.Error())
			return er
		}
		info.Set("HELM_CHART", helmChartName)
		info.Set("HELM_VERSION", helmChartVersion)
		info.Set("HELM_REPOSITORY", helmRepository)
	}

	for i := 0; i < len(cfg.ImageNames); i++ {
		info.Set(fmt.Sprint("OCI_NAME_", i), cfg.ociImageResourceName(i, info)) // ensure 'HELM_CHART' does not conflict with 'OCI_NAME'
		info.Set(fmt.Sprint("OCI_VERSION_", i), cfg.version())
		info.Set(fmt.Sprint("OCI_REFERENCE_", i), cfg.containerRegistry()+"/"+cfg.ImageNameTags[i])
	}

	yamlData, err := yaml.Marshal(&info)
	if err != nil {
		log.Entry().WithError(err).Errorf("Error while Marshaling. %v", err)
		return err
	}

	err = ub.WriteFile(settingsPath, yamlData, 0644)
	if err != nil {
		log.Entry().WithError(err).Errorf("Could not write settings file: %s", settingsPath)
		return err
	}
	logGeneratedFile(ub, settingsPath)
	commonPipelineEnvironment.custom.settingsYaml = string(yamlData)

	ocmFiles = append(ocmFiles, piperutils.Path{
		Name:      "OCM settings",
		Target:    ocm.SettingsFileName,
		Mandatory: false,
		Scope:     "job",
	})

	return piperutils.PersistReportsAndLinks("sapOcmCreateComponent", cfg.GenDir, ub, ocmFiles, nil)
}

// callOcmCli calls the ocm cli to create a component version and transfer it to staging repository.
func callOcmCli(ub SapOcmCreateComponentUtils, cfg *sapOcmCreateComponentOptions, info ocm.ComponentInfo, commonPipelineEnvironment *sapOcmCreateComponentCommonPipelineEnvironment) error {
	if log.IsVerbose() {
		_ = executeCommand(ub, "version")
	}

	// ocm add componentversion
	ctfDir := filepath.Join(cfg.GenDir, "./ctf") // local directory where to create the component transfer archive
	var ocmArgs []string
	ocmArgs = ocm.Verbose(ocmArgs)
	ocmArgs = append(ocmArgs, "add", "componentversion", "--create", "--settings", info["SettingsPath"].(string), "--file", ctfDir)

	ocmArgs = append(ocmArgs, info["componentConstructorPath"].(string))
	err := executeCommand(ub, ocmArgs...) // ocm add componentversion --create --settings settings.yaml --file ctf
	if err != nil {
		return err
	}

	// ocm transfer
	ocmArgs = []string{}
	ocmURL := info["RepositoryURL"].(string)
	if !ocm.IsSet(ocmURL) {
		err = errors.New("could not find an OCI registry to transfer component registry, exiting")
		log.Entry().Logger.WithError(err)
		return err
	}
	log.Entry().Debug("Transfer OCM Component Version to OCI Registry: " + ocmURL)

	ocmArgs = ocm.Verbose(ocmArgs)
	ocmArgs = append(ocmArgs, "transfer", "componentversions")

	// when we're processing a custom component descriptor, we need to omit the ociArtifact access type, because we assume the oci artifact is handled properly by the user
	if commonPipelineEnvironment.custom.isCustomComponentDescriptor {
		ocmArgs = append(ocmArgs, "--disable-uploads")
	}

	cvRef := fmt.Sprintf("ctf::%s//%s", ctfDir, info["CV"])
	ocmArgs = append(ocmArgs, cvRef, ocmURL)
	err = executeCommand(ub, ocmArgs...) // ocm transfer ctf ./ctf <ocm-oci-registry>
	if err != nil {
		return err
	}
	log.Entry().Debug("Transfer to OCI Registry successful.")

	// ocm get componentversion
	// retrieve the component-descriptor, log it and store it as output
	result, err := executeGetOutput(ub, "get", "componentversion", "-o", "yaml", ctfDir)
	if err != nil {
		return err
	}
	cd := string(result)
	log.Entry().Infof("\n%s", cd) // ocm get componentversion -o yaml ...
	commonPipelineEnvironment.custom.componentDescriptor = cd

	return nil
}

// generateComponentsYaml creates a component-constructor file if it does not exist. Sets the content in the commonPipelineEnvironment.custom.componentConstructor.
func generateComponentsYaml(ub SapOcmCreateComponentUtils, cfg *sapOcmCreateComponentOptions, info ocm.ComponentInfo, filePath string, commonPipelineEnvironment *sapOcmCreateComponentCommonPipelineEnvironment) error {
	var err error
	commonPipelineEnvironment.custom.isCustomComponentDescriptor, err = ub.FileExists(filePath)
	if err != nil {
		log.Entry().WithError(err).Error()
		return err
	}
	if commonPipelineEnvironment.custom.isCustomComponentDescriptor {
		log.Entry().Debugf("Found an existing component-constructor file at %s", filePath)
		ba, err := ub.ReadFile(filePath)
		if err != nil {
			log.Entry().WithError(err).Errorf("Could not read file: %s", filePath)
			return err
		}
		commonPipelineEnvironment.custom.componentConstructor = string(ba)
		var comp ocm.Components
		if err = yaml.Unmarshal(ba, &comp); err != nil {
			log.Entry().WithError(err).Errorf("Could not unmarshal component-constructor file: %s", filePath)
			return err
		}
		err = comp.Replace(info)
		if err != nil {
			log.Entry().WithError(err).Error("Variable substitution in component-constructor failed")
			return err
		}
		info.Set("CV", fmt.Sprintf("%s:%s", comp.Name(), comp.Version())) // component version
	} else {
		err = cfg.validateInputParameters()
		if err != nil {
			return err
		}
		err = createDirs(ub, filepath.Dir(filePath))
		if err != nil {
			return err
		}
		log.Entry().Debugf("No component-constructor file found, generating one at %s", filePath)
		f, err := ub.Create(filePath)
		if err != nil {
			log.Entry().WithError(err).Errorf("Could not create file: %s", filePath)
			return err
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Entry().WithError(err).Errorf("Could not close file: %v", err)
			}
		}()

		comp := ocm.NewComponents()

		for i := 0; i < len(cfg.ImageNames); i++ {
			comp.AddOciResource(i)
		}

		if ocm.IsSet(cfg.HelmChartURL) {
			comp.AddHelmAccess(-1)
		}

		yamlData, err := yaml.Marshal(&comp)
		if err != nil {
			log.Entry().WithError(err).Errorf("Could not marshal component descriptor")
			return err
		}
		_, err = f.Write(yamlData)
		if err != nil {
			log.Entry().WithError(err).Errorf("Could not write component descriptor to file: %s", filePath)
			return err
		}
		logGeneratedFile(ub, filePath)
		commonPipelineEnvironment.custom.componentConstructor = string(yamlData)
	}

	return nil
}

// ///////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// OCM Credential Handling functions

// createOcmConfigForCredentials creates a .ocmconfig file with the credentials for the container registry and the staging repository.
func createOcmConfigForCredentials(ub SapOcmCreateComponentUtils, cfg *sapOcmCreateComponentOptions, info ocm.ComponentInfo) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	ocmconfig := path.Join(home, ".ocmconfig")
	exists, er := ub.FileExists(ocmconfig)
	if er != nil {
		log.Entry().WithError(err).Errorf("Error checking if .ocmconfig exists: %s", err.Error())
		return er
	}
	if exists {
		log.Entry().Info("Found a .ocmconfig file, not overwriting it")
		return nil
	}
	ocmConfig := ocm.NewConfig()

	if info.Has("RepoInfo") && ocm.IsSet(info.RepoInfo().RepositoryURL) && ocm.IsSet(info.RepoInfo().User) && ocm.IsSet(info.RepoInfo().Password) {
		// the staging repository for OCM component versions
		ocmConfig.AddCredential(ocm.Oci, info.RepoInfo().RepositoryURL, info.RepoInfo().User, info.RepoInfo().Password)
	}
	if ocm.IsSet(cfg.ContainerRegistryURL) && ocm.IsSet(cfg.ContainerRegistryUser) && ocm.IsSet(cfg.ContainerRegistryPassword) {
		// the build image (kanikoBuild)
		ocmConfig.AddCredential(ocm.Oci, cfg.ContainerRegistryURL, cfg.ContainerRegistryUser, cfg.ContainerRegistryPassword)
	}
	if ocm.IsSet(cfg.HelmChartURL) && ocm.IsSet(cfg.HelmRepositoryUsername) && ocm.IsSet(cfg.HelmRepositoryPassword) {
		ocmConfig.AddCredential(ocm.Helm, cfg.HelmChartURL, cfg.HelmRepositoryUsername, cfg.HelmRepositoryPassword)
	}
	if ocm.IsSet(cfg.DockerConfigJSON) {
		er = isValidDockerConfigJSON(ub, cfg.DockerConfigJSON)
		if er != nil {
			log.Entry().WithError(er).Errorf("Invalid docker config JSON: %s", er.Error())
			return er
		}

		ocmConfig.AddDockerConfig(cfg.DockerConfigJSON)
	}

	yamlData, err := yaml.Marshal(ocmConfig)
	if err != nil {
		log.Entry().Errorf("Error while Marshaling. %v", err)
	}
	err = ub.WriteFile(ocmconfig, yamlData, 0644)
	if err != nil {
		log.Entry().WithError(err).Errorf("Could not write OCM cfg file: %s", ocmconfig)
		return err
	}
	logGeneratedFile(ub, ocmconfig)
	return nil
}

// ///////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// logGeneratedFile logs the content of a generated file if verbose logging is enabled.
func logGeneratedFile(ub SapOcmCreateComponentUtils, filePath string) {
	if !log.IsVerbose() {
		return
	}
	log.Entry().Debugf("generated: %s", filePath)
	file, err := ub.Open(filePath)
	if err != nil {
		log.Entry().Infof("Can not read file %s", filePath)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Entry().Infof("Could not close file: %v", err)
		}
	}()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		log.Entry().Debug(scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		log.Entry().WithError(err).Errorf("Could not scan file: %s", filePath)
	}
}

// CreateDirs creates directories for a given file path.
func createDirs(ub SapOcmCreateComponentUtils, path string) error {
	dir := filepath.Clean(path)
	if dir != "." {
		log.Entry().Debugf("Creating dir: %s", dir)
		err := ub.MkdirAll(dir, 0770)
		if err != nil {
			log.Entry().WithError(err).Errorf("Could not create dir: %s", dir)
			return err
		}
	}
	return nil
}

// executeCommand executes 'ocm' command and logs the command and its output.
func executeCommand(ub SapOcmCreateComponentUtils, ocmArgs ...string) error {
	log.Entry().Infof("ocm %s", strings.Join(ocmArgs, " "))
	if log.IsVerbose() {
		ub.Stdout(log.Writer())
	}
	err := ub.RunExecutable("ocm", ocmArgs...)
	if err != nil {
		log.Entry().WithError(err).Error("Error executing ocm command")
		return err
	}
	return nil
}

// executeGetOutput executes 'ocm' command and returns the output.
func executeGetOutput(ub SapOcmCreateComponentUtils, ocmArgs ...string) ([]byte, error) {
	log.Entry().Infof("ocm %s", strings.Join(ocmArgs, " "))

	stdout := bytes.Buffer{}
	ub.Stdout(&stdout)
	err := ub.RunExecutable("ocm", ocmArgs...)
	if err != nil {
		log.Entry().WithError(err).Error("Error executing ocm command")
		var stderr *exec.ExitError
		ok := errors.As(err, &stderr)
		if ok {
			log.Entry().Error(stderr.Stderr)
		}
		return nil, err
	}
	return stdout.Bytes(), nil
}

// isValidDockerConfigJSON does a basic verification of the given docker config JSON file.
func isValidDockerConfigJSON(ub SapOcmCreateComponentUtils, pathDockerConfigJSON string) error {
	exists, err := ub.FileExists(pathDockerConfigJSON)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("dockerConfigJSON file does not exist")
	}

	dockerConfigContent, err := ub.FileRead(pathDockerConfigJSON)
	if err != nil {
		return err
	}

	dockerConfig := map[string]interface{}{}
	err = json.Unmarshal(dockerConfigContent, &dockerConfig)
	if err != nil {
		return err
	}

	if dockerConfig["auths"] == nil {
		return errors.New("dockerConfigJSON file does not contain 'auths' key")
	}

	return nil
}

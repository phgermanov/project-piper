package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"

	piperConfig "github.com/SAP/jenkins-library/pkg/config"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	piperOrchestrator "github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/pkg/errors"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/dwc"
)

const (
	artifactResolutionError   string = "error resolving artifact descriptor base"
	mtarExtractionDestination string = "extracted_mtar"
	dwcUIAppContentModule            = "dwc-ui-appcontent"
)

type sapDwCStageReleaseUtils interface {
	FileWrite(name string, data []byte, perm os.FileMode) error
	FileRemove(path string) error
	FileRead(path string) ([]byte, error)
	FileExists(filename string) (bool, error)
	MkdirAll(path string, perm os.FileMode) error
	Glob(pattern string) (matches []string, err error)
	ResolveDefaultResourceName(collector dwc.MetadataCollector) (string, error)
	ResolveContainerImageURL(promotedDockerImage, artifactVersion string) (string, string, error)
	UploadArtifact(artifactDescriptor dwc.ArtifactDescriptor) (*dwc.ArtifactUploadResponse, error)
	LoginToDwCCli(loginDescriptor dwc.LoginDescriptor) error
	InstallDwCCli(githubToken string) error
	GetMetadataEntry(element string) (string, error)
	Untar(tarArchive, destination string, extractionLevel int) error
	DetectOrchestrator() piperOrchestrator.Orchestrator
	LookupEnv(variable string) (string, bool)
	Unzip(tarArchive, destination string) error
}

type sapDwCStageReleaseUtilsBundle struct {
	dwc.UploadController
	dwc.ConfigResolver
	dwc.CLIBinaryResolver
	dwc.MetadataCollector
	*piperutils.Files
}

func newSapDwCStageReleaseUtils() sapDwCStageReleaseUtils {
	utils := sapDwCStageReleaseUtilsBundle{
		UploadController: dwc.NewUploadController(
			dwc.NewDefaultArtifactCompressor(),
			&piperutils.Files{},
			dwc.DefaultFilePatternMover{},
			&piperutils.Files{},
			dwc.DefaultGlobMatcher{},
			&dwc.DefaultStageWatchOrchestrator{},
			dwc.DefaultExecutorFactory{},
			dwc.DefaultCLICommandExecutor,
		),
		ConfigResolver: dwc.ConfigResolver{},
		Files:          &piperutils.Files{},
		CLIBinaryResolver: dwc.CLIBinaryResolver{
			FileDownloader:     &piperhttp.Client{},
			PermEditor:         &piperutils.Files{},
			CLIReleaseResolver: &dwc.DefaultReleaseResolver{},
		},
		MetadataCollector: &dwc.DefaultMetadataCollector{},
	}
	return &utils
}

func (utils *sapDwCStageReleaseUtilsBundle) Untar(tarArchive, destination string, extractionLevel int) error {
	return piperutils.Untar(tarArchive, destination, extractionLevel)
}

func (utils *sapDwCStageReleaseUtilsBundle) DetectOrchestrator() piperOrchestrator.Orchestrator {
	return piperOrchestrator.DetectOrchestrator()
}

func (utils *sapDwCStageReleaseUtilsBundle) LookupEnv(variable string) (string, bool) {
	return os.LookupEnv(variable)
}

func (utils *sapDwCStageReleaseUtilsBundle) Unzip(zipArchive, destination string) error {
	_, err := piperutils.Unzip(zipArchive, destination)
	return err
}

func sapDwCStageRelease(config sapDwCStageReleaseOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *sapDwCStageReleaseCommonPipelineEnvironment) {
	utils := newSapDwCStageReleaseUtils()
	if err := runSapDwCStageRelease(&config, telemetryData, commonPipelineEnvironment, utils); err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runSapDwCStageRelease(config *sapDwCStageReleaseOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *sapDwCStageReleaseCommonPipelineEnvironment, utils sapDwCStageReleaseUtils) error {
	log.Entry().Debugf("step config: %+v", config)
	globalConfiguration, artifactConfigurations, err := getConfigurations(config, utils)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("failed to parse configuration: %w", err)
	}
	if err := validateConfigurations(globalConfiguration, artifactConfigurations); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("failed to validate configuration: %w", err)
	}
	loginDescriptor, err := createLoginDescriptor(globalConfiguration, utils)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("failed to create login descriptor: %w", err)
	}
	if globalConfiguration.CliPath == "" {
		if err := utils.InstallDwCCli(globalConfiguration.GithubToken); err != nil {
			return fmt.Errorf("unable to install DwC CLI: %w", err)
		}
	} else {
		dwc.SetTargetBinary(globalConfiguration.CliPath)
		log.Entry().Debugf("using DwC CLI binary at %s", globalConfiguration.CliPath)
	}
	if err := utils.LoginToDwCCli(loginDescriptor); err != nil {
		return fmt.Errorf("unable to login to DwC CLI: %w", err)
	}
	uploadedArtifactIDs := make([]string, 0, len(artifactConfigurations))
	for _, artifactConfiguration := range artifactConfigurations {
		uiTargetPath := ""
		if artifactConfiguration.ExtractFromMTA && artifactConfiguration.UploadType == dwc.UploadTypeUI {
			if err := retrieveUIAppsFromMTA(globalConfiguration, &artifactConfiguration, utils); err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}
			uiTargetPath = "dist"
		}
		if artifactConfiguration.HasArchive {
			if artifactConfiguration.ArtifactType == dwc.ArtifactTypeHelm {
				if err := extractHelmChart(globalConfiguration, utils); err != nil {
					log.SetErrorCategory(log.ErrorConfiguration)
					return fmt.Errorf("unable to extract Helm chart for resource '%s': %w", artifactConfiguration.ResourceName, err)
				}
			} else if artifactConfiguration.UploadType == dwc.UploadTypeUI {
				if err := extractUIResources(globalConfiguration, utils, artifactConfiguration.ArchivePattern, uiTargetPath); err != nil {
					log.SetErrorCategory(log.ErrorConfiguration)
					return fmt.Errorf("unable to extract UI resources for resource '%s': %w", artifactConfiguration.ResourceName, err)
				}
			}
		}
		artifactDescriptor, err := createArtifactDescriptor(globalConfiguration, artifactConfiguration, utils)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return fmt.Errorf("failed to create artifact descriptor for resource '%s': %w", artifactConfiguration.ResourceName, err)
		}
		log.Entry().Debugf("successfully created artifact descriptor for resource %s", artifactConfiguration.ResourceName)

		uploadResponse, err := utils.UploadArtifact(artifactDescriptor)
		if err != nil {
			return fmt.Errorf("unable to upload artifact '%s' to DwC: %w", artifactDescriptor.GetResourceName(), err)
		}
		uploadedArtifactIDs = append(uploadedArtifactIDs, uploadResponse.ID)
	}
	// TODO: remove the following if statement if other integrations have been adjusted and are using dwcUploadedArtifactIDs instead of dwcUploadedArtifactID (Leon, Erik - 2024-08-30)
	if len(uploadedArtifactIDs) == 1 {
		commonPipelineEnvironment.custom.dwcUploadedArtifactID = uploadedArtifactIDs[0]
	}

	commonPipelineEnvironment.custom.dwcUploadedArtifactIDs = uploadedArtifactIDs

	log.Entry().Infof("Upload to DwC was successful :)")
	return nil
}

func getConfigurations(config *sapDwCStageReleaseOptions, utils sapDwCStageReleaseUtils) (dwc.GlobalConfiguration, []dwc.ArtifactConfiguration, error) {
	deriveAdditionalDownloadURLsEntries, err := parseConfigurationInterface[dwc.DeriveAdditionalDownloadURLsEntry](config.DeriveAdditionalDownloadURLs)
	if err != nil {
		return dwc.GlobalConfiguration{}, nil, fmt.Errorf("failed to parse deriveAdditionalDownloadURLs: %w", err)
	}
	globalConfiguration := dwc.GlobalConfiguration{
		ArtifactURLs:                    config.ArtifactURLs,
		ArtifactVersion:                 config.ArtifactVersion,
		CliPath:                         config.CliPath,
		DeriveAdditionalDownloadURLs:    deriveAdditionalDownloadURLsEntries,
		DownloadedArchivesPath:          config.DownloadedArchivesPath,
		GatewayCertificatePath:          config.GatewayCertificatePath,
		GatewayURL:                      config.GatewayURL,
		GithubToken:                     config.GithubToken,
		HelmChartDirectory:              config.HelmChartDirectory,
		HelmChartURL:                    config.HelmChartURL,
		MtarFilePath:                    config.MtarFilePath,
		MtarUIPath:                      config.MtarUIPath,
		ProjectName:                     config.ProjectName,
		RequiredSuccessfulStages:        config.RequiredSuccessfulStages,
		StagesToWatch:                   config.StagesToWatch,
		StageWatchPolicy:                config.StageWatchPolicy,
		ThemistoInstanceCertificatePath: config.ThemistoInstanceCertificatePath,
		ThemistoInstanceURL:             config.ThemistoInstanceURL,
		UseCertLogin:                    config.UseCertLogin,
		WatchResourceOfInterest:         config.WatchResourceOfInterest,
		VaultBasePath:                   config.VaultBasePath,
		VaultPipelineName:               config.VaultPipelineName,
		PipelineID:                      config.PipelineID,
		CumulusPipelineRunKey:           config.CumulusPipelineRunKey,
	}
	var artifactConfigurations []dwc.ArtifactConfiguration
	if len(config.Artifacts) > 0 {
		var err error
		artifactConfigurations, err = parseConfigurationInterface[dwc.ArtifactConfiguration](config.Artifacts)
		if err != nil {
			return dwc.GlobalConfiguration{}, nil, fmt.Errorf("failed to parse artifacts: %w", err)
		}
		applyArtifactConfigurationDefaults(config, artifactConfigurations, true)
		if len(artifactConfigurations) == 1 {
			if artifactConfigurations[0].ResourceName == "" {
				resourceName, err := utils.ResolveDefaultResourceName(utils)
				if err != nil {
					return dwc.GlobalConfiguration{}, nil, fmt.Errorf("unable to resolve default resourceName: %w", err)
				}
				artifactConfigurations[0].ResourceName = resourceName
				log.Entry().Infof("No resourceName has been provided. Resolved default value to %s", resourceName)
			}
		} else {
			err := setApplicableMTAArtifactMetadata(artifactConfigurations)
			if err != nil {
				return dwc.GlobalConfiguration{}, nil, err
			}
		}
	} else {
		if config.ResourceName == "" {
			resourceName, err := utils.ResolveDefaultResourceName(utils)
			if err != nil {
				return dwc.GlobalConfiguration{}, nil, fmt.Errorf("unable to resolve default resourceName: %w", err)
			}
			config.ResourceName = resourceName
			log.Entry().Infof("No resourceName has been provided. Resolved default value to %s", resourceName)
		}
		artifactConfigurations = []dwc.ArtifactConfiguration{
			{
				ResourceName: config.ResourceName,
			},
		}
		applyArtifactConfigurationDefaults(config, artifactConfigurations, false)
	}
	// get default value of helmChartDirectory from step metadata
	helmChartDirectoryDefaultValue := sapDwCStageReleaseMetadata().Spec.Inputs.Parameters[slices.IndexFunc(
		sapDwCStageReleaseMetadata().Spec.Inputs.Parameters,
		func(p piperConfig.StepParameters) bool { return p.Name == "helmChartDirectory" })].Default.(string)

	if globalConfiguration.HelmChartDirectory != helmChartDirectoryDefaultValue {
		// get default value of artifactFilesToUpload from step metadata
		artifactFilesToUploadDefaultValue := sapDwCStageReleaseMetadata().Spec.Inputs.Parameters[slices.IndexFunc(
			sapDwCStageReleaseMetadata().Spec.Inputs.Parameters,
			func(p piperConfig.StepParameters) bool { return p.Name == "artifactFilesToUpload" })].Default.([]string)

		for i, artifactConfiguration := range artifactConfigurations {
			if artifactConfiguration.ArtifactType == dwc.ArtifactTypeHelm && slices.Equal(artifactConfiguration.ArtifactFilesToUpload, artifactFilesToUploadDefaultValue) {
				artifactConfigurations[i].ArtifactFilesToUpload = append(artifactConfigurations[i].ArtifactFilesToUpload, globalConfiguration.HelmChartDirectory)
			}
		}
	}
	return globalConfiguration, artifactConfigurations, nil
}

// this function applies the default values from sapDwCStageReleaseOptions to the artifactConfigurations
// if applyOnlyDefaults is set to false, all the sapDwCStageReleaseOptions properties will be copied to the artifactConfiguration, even if the value differs from the default
func applyArtifactConfigurationDefaults(config *sapDwCStageReleaseOptions, artifactConfigurations []dwc.ArtifactConfiguration, applyOnlyDefaults bool) {
	configValue := reflect.ValueOf(config).Elem() // the value of the sapDwCStageReleaseOptions struct
	for artifactIndex := range artifactConfigurations {
		artifactValue := reflect.ValueOf(&artifactConfigurations[artifactIndex]).Elem() // the value of the current artifactConfiguration struct
		for fieldIndex := 0; fieldIndex < artifactValue.NumField(); fieldIndex++ {
			artifactFieldDefinition := artifactValue.Type().Field(fieldIndex) // the type definition of the current field in artifactConfiguration
			if artifactFieldDefinition.Name == "ResourceName" {               // resourceName has its own logic for resolving a default (see sapDwCStageReleaseUtils.ResolveDefaultResourceName)
				continue
			}
			artifactFieldValue := artifactValue.Field(fieldIndex) // the value of the current field in artifactConfiguration
			if configFieldDefinition, ok := configValue.Type().FieldByName(artifactFieldDefinition.Name); ok && artifactFieldValue.IsZero() && configFieldDefinition.Type == artifactFieldDefinition.Type {
				configFieldValue := configValue.FieldByName(artifactFieldDefinition.Name) // the value of the current field in sapDwCStageReleaseOptions
				if applyOnlyDefaults {
					propertyNameCamelCase := strings.ToLower(artifactFieldDefinition.Name[0:1]) + artifactFieldDefinition.Name[1:len(artifactFieldDefinition.Name)]
					configFieldDefault := sapDwCStageReleaseMetadata().Spec.Inputs.Parameters[slices.IndexFunc( // the default value of the current field
						sapDwCStageReleaseMetadata().Spec.Inputs.Parameters,
						func(p piperConfig.StepParameters) bool { return p.Name == propertyNameCamelCase })].Default
					if configFieldValue.Kind() == reflect.String || configFieldValue.Kind() == reflect.Bool { // for string and bool types, we can directly compare the values
						if configFieldValue.Interface() == configFieldDefault {
							artifactFieldValue.Set(configFieldValue)
						}
					} else if reflect.DeepEqual(configFieldValue.Interface(), configFieldDefault) { // for other types, we use reflect.DeepEqual to compare the values
						artifactFieldValue.Set(configFieldValue)
					}
				} else {
					artifactFieldValue.Set(configFieldValue)
				}
			}
		}
	}
}

func setApplicableMTAArtifactMetadata(artifactConfigurations []dwc.ArtifactConfiguration) error {
	var mtaConfiguration *dwc.ArtifactConfiguration = nil
	uiToBeExtractedFromMTA := false
	for i, artifactConfiguration := range artifactConfigurations {
		if artifactConfiguration.ArtifactType == dwc.ArtifactTypeMta {
			mtaConfiguration = &artifactConfigurations[i]
		}
		if artifactConfiguration.UploadType == dwc.UploadTypeUI && artifactConfiguration.ExtractFromMTA {
			uiToBeExtractedFromMTA = true
		}
	}
	if uiToBeExtractedFromMTA {
		if mtaConfiguration == nil {
			return errors.New("specifying 'extractFromMTA' without defining a MTA service artifact is not supported")
		}
		mtaConfiguration.UploadMetadata = append(mtaConfiguration.UploadMetadata, fmt.Sprintf("%s=true", dwc.SelectiveMtaModuleDeploymentMetadataKey))
	}
	return nil
}

func validateConfigurations(globalConfiguration dwc.GlobalConfiguration, artifactConfigurations []dwc.ArtifactConfiguration) error {
	if err := validateMandatoryStringParameters(globalConfiguration); err != nil {
		return fmt.Errorf("failed to parse parameters: %w", err)
	}
	if err := validateArtifactConfigurations(artifactConfigurations); err != nil {
		return fmt.Errorf("failed to validate artifact configurations: %w", err)
	}
	return nil
}

func validateArtifactConfigurations(artifacts []dwc.ArtifactConfiguration) error {
	resourceNames := make(map[string]struct{})
	for _, artifact := range artifacts {
		if _, ok := dwc.AvailableUploadTypes[artifact.UploadType]; !ok {
			return fmt.Errorf("artifact upload type is not supported: '%s'", artifact.UploadType)
		}
		if artifact.UploadType == dwc.UploadTypeService {
			if _, ok := dwc.AvailableArtifactTypes[artifact.ArtifactType]; !ok {
				return fmt.Errorf("artifact type is not supported: '%s'", artifact.ArtifactType)
			}
		} else if artifact.UploadType == dwc.UploadTypeOrbit {
			if artifact.ArtifactType != "" && artifact.ArtifactType != dwc.ArtifactTypeDocker && artifact.ArtifactType != dwc.ArtifactTypeDockerBuildReleaseMetadata {
				return fmt.Errorf("artifact type '%s' is not allowed for upload type 'orbit'; either omit artifactType or set it to '%s'", artifact.ArtifactType, dwc.ArtifactTypeDockerBuildReleaseMetadata)
			}
		} else {
			if artifact.ArtifactType != "" {
				return fmt.Errorf("artifact type is not allowed for upload type 'ui'")
			}
		}
		if artifact.ResourceName == "" {
			return fmt.Errorf("no resourceName provided for artifact type '%s'", artifact.ArtifactType)
		}
		if _, exists := resourceNames[artifact.ResourceName]; exists { // validate that resource name is unique among all artifacts
			return fmt.Errorf("duplicate resource name found: %s", artifact.ResourceName)
		}
		resourceNames[artifact.ResourceName] = struct{}{}
		if artifact.UploadType != dwc.UploadTypeUI {
			if err := validateAppNameProvidedExactlyOnceAndUnique(artifact.AppName, artifact.Apps); err != nil {
				return fmt.Errorf("invalid app name configuration for resource name %s: %w", artifact.ResourceName, err)
			}
		}
		// validate that artifact files to upload are provided for all artifact types except for MTAs (mta does not require file bundling)
		if len(artifact.ArtifactFilesToUpload) == 0 && artifact.ArtifactType != dwc.ArtifactTypeMta {
			return fmt.Errorf("no artifact files to upload provided for resource name '%s'", artifact.ResourceName)
		}
		if artifact.UploadType == dwc.UploadTypeUI && artifact.ArchivePattern == "" && artifact.HasArchive {
			return fmt.Errorf("no archive pattern provided for UI artifact with resource name '%s' but has archive", artifact.ResourceName)
		}
	}
	return nil
}

func validateMandatoryStringParameters(globalConfiguration dwc.GlobalConfiguration) error {
	stepMetadata := sapDwCStageReleaseMetadata()
	configReflect := reflect.Indirect(reflect.ValueOf(globalConfiguration))
	for _, parameter := range stepMetadata.Spec.Inputs.Parameters {
		if parameter.Mandatory && parameter.Type == "string" && parameter.Name != "" {
			//transform e.g. dwcResourceName to DwcResourceName
			capitalizedParameterName := fmt.Sprintf("%s%s", strings.ToUpper(parameter.Name[0:1]), parameter.Name[1:len(parameter.Name)])
			parameterValue := configReflect.FieldByName(capitalizedParameterName)
			if parameterValue.IsZero() {
				return fmt.Errorf("mandatory parameter %s not set", parameter.Name)
			}
		}
	}
	return nil
}

func validateAppNameProvidedExactlyOnceAndUnique(appName string, appsConfig []map[string]interface{}) error {
	apps, err := parseConfigurationInterface[dwc.App](appsConfig)
	if err != nil {
		return fmt.Errorf("failed to parse apps configuration: %w", err)
	}
	if appName == "" && len(apps) == 0 {
		return fmt.Errorf("neither parameter appName nor apps is set, however, providing one of them is mandatory")
	} else if appName != "" && len(apps) > 0 {
		return fmt.Errorf("both parameters appName and apps are set, however, only one of them can be provided")
	}

	if len(apps) > 0 {
		// validate that name for every app in apps is unique
		appNames := make(map[string]struct{})
		for _, app := range apps {
			if app.Name == "" {
				return fmt.Errorf("name is not set for at least one app in config list")
			} else {
				if _, exists := appNames[app.Name]; exists {
					return fmt.Errorf("duplicate app name found: %s", app.Name)
				}
				appNames[app.Name] = struct{}{}
			}
		}
	}

	return nil
}

func createArtifactDescriptor(globalConfiguration dwc.GlobalConfiguration, artifactConfiguration dwc.ArtifactConfiguration, utils sapDwCStageReleaseUtils) (dwc.ArtifactDescriptor, error) {
	var descriptor dwc.ArtifactDescriptor
	descriptorBase, err := resolveDescriptorBase(globalConfiguration, artifactConfiguration, utils)
	if err != nil {
		return nil, err
	}
	switch artifactConfiguration.UploadType {
	case dwc.UploadTypeService:
		var err error
		descriptor, err = resolveArtifactDescriptorFromArtifactType(globalConfiguration, artifactConfiguration, utils, descriptorBase)
		if err != nil {
			return nil, fmt.Errorf("error resolving artifact descriptor from artifactType: %w", err)
		}
	case dwc.UploadTypeUI:
		descriptor, err = dwc.NewUIArtifact(descriptorBase, artifactConfiguration.UiUploadBasePath)
		if err != nil {
			return nil, fmt.Errorf("error resolving artifact descriptor for UI artifact: %w", err)
		}
	case dwc.UploadTypeOrbit:
		image, tag, err := utils.ResolveContainerImageURL(artifactConfiguration.PromotedDockerImage, globalConfiguration.ArtifactVersion)
		if err != nil {
			return nil, fmt.Errorf("error resolving containerImageURL for uploadType %s: %w", dwc.UploadTypeOrbit, err)
		}
		containerImageLocator := fmt.Sprintf("%s:%s", image, tag)
		additionalDownloadURLs, err := deriveAdditionalDownloadURLs(globalConfiguration.DeriveAdditionalDownloadURLs, containerImageLocator)
		if err != nil {
			return nil, fmt.Errorf("error deriving additional download URLs: %w", err)
		}
		descriptor = dwc.NewOrbitArtifact(descriptorBase, containerImageLocator, globalConfiguration.HelmChartDirectory, additionalDownloadURLs)
	default:
		return nil, fmt.Errorf("unknown uploadType %s", artifactConfiguration.UploadType)
	}
	return descriptor, nil
}

func resolveArtifactDescriptorFromArtifactType(globalConfiguration dwc.GlobalConfiguration, artifactConfiguration dwc.ArtifactConfiguration, utils sapDwCStageReleaseUtils, descriptorBase dwc.DescriptorBase) (dwc.ArtifactDescriptor, error) {
	var descriptor dwc.ArtifactDescriptor
	var artifactUrl, containerImageLocator, image, tag string
	var additionalDownloadURLs map[string]string
	var err error

	switch artifactConfiguration.ArtifactType {
	case dwc.ArtifactTypeJava, dwc.ArtifactTypeMaven, dwc.ArtifactTypeMavenMta, dwc.ArtifactTypeMta:
		artifactUrl, err = extractArtifactTypeSpecificURLFromArtifactURLs(globalConfiguration.ArtifactURLs, artifactConfiguration.ArtifactType, artifactConfiguration.ArtifactPattern, artifactConfiguration.Repository)
		if err != nil {
			return nil, fmt.Errorf("failed to filter artifact URLs %s with pattern %s: %w", globalConfiguration.ArtifactURLs, artifactConfiguration.Repository, err)
		}
		additionalDownloadURLs, err = deriveAdditionalDownloadURLs(globalConfiguration.DeriveAdditionalDownloadURLs, artifactUrl)
		if err != nil {
			return nil, fmt.Errorf("error deriving additional download URLs: %w", err)
		}
	case dwc.ArtifactTypeDocker, dwc.ArtifactTypeDockerBuildReleaseMetadata:
		image, tag, err = utils.ResolveContainerImageURL(artifactConfiguration.PromotedDockerImage, globalConfiguration.ArtifactVersion)
		if err != nil {
			return nil, fmt.Errorf("error resolving containerImageURL for artifactType %s: %w", artifactConfiguration.ArtifactType, err)
		}
		containerImageLocator = fmt.Sprintf("%s:%s", image, tag)
		additionalDownloadURLs, err = deriveAdditionalDownloadURLs(globalConfiguration.DeriveAdditionalDownloadURLs, containerImageLocator)
		if err != nil {
			return nil, fmt.Errorf("error deriving additional download URLs: %w", err)
		}
	case dwc.ArtifactTypeHelm:
		image, tag, err = utils.ResolveContainerImageURL(artifactConfiguration.PromotedDockerImage, globalConfiguration.ArtifactVersion)
		if err != nil {
			return nil, fmt.Errorf("error resolving containerImageURL for artifactType %s: %w", artifactConfiguration.ArtifactType, err)
		}
	default:
		return nil, fmt.Errorf("unknown artifactType %s", artifactConfiguration.ArtifactType)
	}

	switch artifactConfiguration.ArtifactType {
	case dwc.ArtifactTypeJava, dwc.ArtifactTypeMaven:
		descriptor = dwc.NewJavaArtifact(descriptorBase, artifactUrl, additionalDownloadURLs)
	case dwc.ArtifactTypeMavenMta, dwc.ArtifactTypeMta:
		descriptor = dwc.NewMTAArtifact(descriptorBase, artifactUrl, additionalDownloadURLs)
	case dwc.ArtifactTypeDocker, dwc.ArtifactTypeDockerBuildReleaseMetadata:
		descriptor = dwc.NewDockerArtifact(descriptorBase, containerImageLocator, additionalDownloadURLs)
	case dwc.ArtifactTypeHelm:
		descriptor = dwc.NewHelmArtifact(descriptorBase, artifactConfiguration.OverwriteHelmDockerImage, artifactConfiguration.HelmValues, image, tag, globalConfiguration.HelmChartDirectory)
	default:
		return nil, fmt.Errorf("unknown artifactType %s", artifactConfiguration.ArtifactType)
	}
	return descriptor, nil
}

func resolveDescriptorBase(globalConfiguration dwc.GlobalConfiguration, artifactConfiguration dwc.ArtifactConfiguration, utils sapDwCStageReleaseUtils) (dwc.DescriptorBase, error) {
	watchPolicy, err := resolveStageWatchPolicy(globalConfiguration)
	if err != nil {
		return dwc.DescriptorBase{}, fmt.Errorf("%s: %w", artifactResolutionError, err)
	}
	uploadMetadata, err := parseUploadMetadata(artifactConfiguration.UploadMetadata)
	if err != nil {
		return dwc.DescriptorBase{}, fmt.Errorf("%s: %w", artifactResolutionError, err)
	}
	addCumulusMetadata(uploadMetadata, globalConfiguration)
	apps, err := parseConfigurationInterface[dwc.App](artifactConfiguration.Apps)
	if err != nil {
		return dwc.DescriptorBase{}, fmt.Errorf("%s: %w", artifactResolutionError, err)
	}

	return dwc.DescriptorBase{
		FileUtils:         utils,
		AppName:           artifactConfiguration.AppName,
		Apps:              apps,
		ResourceName:      artifactConfiguration.ResourceName,
		FilePatterns:      artifactConfiguration.ArtifactFilesToUpload,
		StagesToWatch:     globalConfiguration.StagesToWatch,
		WatchROI:          globalConfiguration.WatchResourceOfInterest,
		StageWatchPolicy:  watchPolicy,
		UploadMetadata:    uploadMetadata,
		MetadataCollector: utils,
	}, nil
}

func resolveStageWatchPolicy(globalConfiguration dwc.GlobalConfiguration) (dwc.StageWatchPolicy, error) {
	switch globalConfiguration.StageWatchPolicy {
	case dwc.StageWatchPolicyOverallSuccess:
		return dwc.OverallSuccessPolicy(), nil
	case dwc.StageWatchPolicySubsetSuccess:
		if len(globalConfiguration.RequiredSuccessfulStages) == 0 {
			return nil, fmt.Errorf("stageWatchPolicy %s was specified but requiredSuccessfulStages is empty", dwc.StageWatchPolicySubsetSuccess)
		}
		return dwc.SubsetSuccessPolicy(globalConfiguration.RequiredSuccessfulStages), nil
	case dwc.StageWatchPolicyAlwaysPass:
		return dwc.AlwaysPassPolicy(), nil
	case dwc.StageWatchPolicyAtLeastOneSuccessfulDeployment:
		return dwc.AtLeastOneSuccessfulDeploymentPolicy(), nil
	default:
		return nil, fmt.Errorf("unknown stageWatchPolicy %s", globalConfiguration.StageWatchPolicy)
	}
}

func createLoginDescriptor(config dwc.GlobalConfiguration, utils sapDwCStageReleaseUtils) (dwc.LoginDescriptor, error) {
	// try to use Gateway if themisto instance URL is not provided (should be the default)
	if config.ThemistoInstanceURL == "" {
		if config.GatewayURL == "" {
			return nil, errors.New("unable to resolve gatewayURL from config.")
		}
		var actionsIdTokenRequestToken, actionsIdTokenRequestUrl string
		orchestrator := utils.DetectOrchestrator()
		if orchestrator != piperOrchestrator.GitHubActions || config.UseCertLogin {
			if config.GatewayCertificatePath == "" {
				return nil, errors.New("unable to resolve gatewayCertificatePath from config. Make sure you considered https://pages.github.tools.sap/deploy-with-confidence/solar-system/onboarding/set-up-new-dwc-instance/cicd_environment for a proper CI/CD environment setup")
			}
		} else {
			var ok bool
			actionsIdTokenRequestToken, ok = utils.LookupEnv(dwc.GitHubActionsOIDCTokenRequestTokenEnvVar)
			if !ok {
				return nil, fmt.Errorf("environment variable %s not set", dwc.GitHubActionsOIDCTokenRequestTokenEnvVar)
			}
			actionsIdTokenRequestUrl, ok = utils.LookupEnv(dwc.GitHubActionsOIDCTokenRequestURLEnvVar)
			if !ok {
				return nil, fmt.Errorf("environment variable %s not set", dwc.GitHubActionsOIDCTokenRequestURLEnvVar)
			}
			// if the OIDC token is used for the login we need to use the non-mTLS gateway URL (the mTLS URL is the default)
			if config.GatewayURL == dwc.DefaultMtlsGatewayUrl {
				config.GatewayURL = dwc.DefaultGatewayUrl
			}
		}
		if config.ProjectName == "" {
			return nil, errors.New("unable to resolve projectName from config.")
		}
		descriptor := dwc.NewGatewayLoginDescriptor(config.GatewayURL, config.GatewayCertificatePath, config.ProjectName, actionsIdTokenRequestToken, actionsIdTokenRequestUrl, orchestrator, config.UseCertLogin)
		return descriptor, nil
	}

	// try to use themisto configuration as fallback
	log.Entry().Warnf("themistoInstanceURL provided. Trying to use the deprecated(!) themisto configuration.")
	if config.ThemistoInstanceCertificatePath == "" {
		return nil, errors.New("unable to resolve themistoInstanceCertificatePath from config. Make sure you considered https://pages.github.tools.sap/deploy-with-confidence/solar-system/onboarding/set-up-new-dwc-instance/cicd_environment for a proper CI/CD environment setup")
	}
	descriptor := dwc.NewThemistoLoginDescriptor(config.ThemistoInstanceURL, config.ThemistoInstanceCertificatePath)
	return descriptor, nil
}

func extractArtifactTypeSpecificURLFromArtifactURLs(artifactUrls []string, artifactType, artifactPattern, repository string) (string, error) {
	if len(artifactUrls) == 0 {
		return "", errors.New("length of artifactUrls is 0")
	}
	if artifactType == "" {
		return "", errors.New("artifactType not set")
	}
	if repository == "" && artifactPattern == "" {
		return "", errors.New("repository not set")
	}
	useRegex := false
	if artifactPattern == "" {
		log.Entry().Debugf("no artifact pattern provided by user. Using default pattern for specified build tool")
		switch artifactType {
		case dwc.ArtifactTypeJava, dwc.ArtifactTypeMaven:
			artifactPattern = fmt.Sprintf("**/%s/**/*.jar", repository)
		case dwc.ArtifactTypeMta, dwc.ArtifactTypeMavenMta:
			artifactPattern = fmt.Sprintf("**/%s/**/*.mtar", repository)
		default:
			artifactPattern = fmt.Sprintf("**/%s/**/*", repository)
		}
	} else {
		log.Entry().Debugf("artifact pattern (Regex) '%s' provided by user", artifactPattern)
		useRegex = true
	}

	log.Entry().Debugf("start filtering artifact URLs by pattern '%s'", artifactPattern)

	var regex *regexp.Regexp
	if useRegex {
		var err error
		regex, err = regexp.Compile(artifactPattern)
		if err != nil {
			return "", fmt.Errorf("failed to compile artifact pattern regex %s: %w", artifactPattern, err)
		}
	}

	filteredArtifactURLs, err := getFilteredArtifactURLs(artifactUrls, regex, artifactPattern)
	if err != nil {
		return "", err
	}
	if len(filteredArtifactURLs) != 1 {
		return "", fmt.Errorf("expected 1 extracted artifact type specific URL from artifact URLs for type %s using pattern %s, got %d, which are: %v", artifactType, artifactPattern, len(filteredArtifactURLs), filteredArtifactURLs)
	}
	return filteredArtifactURLs[0], nil
}

func getFilteredArtifactURLs(artifactUrls []string, regex *regexp.Regexp, artifactPattern string) ([]string, error) {
	var filteredArtifactURLs []string
	for _, url := range artifactUrls {
		var match bool
		if regex != nil {
			match = regex.MatchString(url)
		} else {
			var err error
			match, err = doublestar.Match(artifactPattern, url)
			if err != nil {
				return nil, fmt.Errorf("failed to match artifact URL %s: %w", url, err)
			}
		}
		if match {
			filteredArtifactURLs = append(filteredArtifactURLs, url)
			log.Entry().Debugf("artifact URL '%s' matches pattern", url)
		} else {
			log.Entry().Debugf("artifact URL '%s' does not match pattern", url)
		}
	}
	return filteredArtifactURLs, nil
}

func parseUploadMetadata(metadata []string) (map[string]string, error) {
	metadataMap := make(map[string]string)

	for _, meta := range metadata {
		log.Entry().Debugf("parse metadata: %s", meta)
		split := strings.SplitN(meta, "=", 2)
		if len(split) != 2 {
			return nil, errors.Errorf("failed to parse metadata value %v, must be in the form of key=value", meta)
		}
		key := split[0]
		value := split[1]
		log.Entry().Debugf("parsed metadata successfully: key=%s value=%s", key, value)

		metadataMap[key] = value
	}
	return metadataMap, nil
}

func parseConfigurationInterface[ConfigInterfaceType any](configInterface []map[string]interface{}) ([]ConfigInterfaceType, error) {
	var config []ConfigInterfaceType
	bytes, err := json.Marshal(configInterface) // marshal configInterface from []map[string]interface{} representation to json
	if err != nil {
		return nil, fmt.Errorf("marshalling config as json failed: %w", err)
	}
	err = json.Unmarshal(bytes, &config) // unmarshal config from json to []ConfigInterfaceType representation
	if err != nil {
		return nil, fmt.Errorf("unmarshalling config from json failed: %w", err)
	}
	return config, nil
}

func addCumulusMetadata(metadata map[string]string, globalConfiguration dwc.GlobalConfiguration) {
	if pipelineId := globalConfiguration.PipelineID; pipelineId != "" {
		metadata["pipelineId"] = pipelineId
	}
	if vaultBasePath := globalConfiguration.VaultBasePath; vaultBasePath != "" {
		metadata["vaultBasePath"] = vaultBasePath
	}
	if pipelineRunKey := globalConfiguration.CumulusPipelineRunKey; pipelineRunKey != "" {
		metadata["pipelineRunKey"] = pipelineRunKey
	}
}

func extractHelmChart(globalConfiguration dwc.GlobalConfiguration, utils sapDwCStageReleaseUtils) error {
	helmArchiveFileName := filepath.Base(globalConfiguration.HelmChartURL)
	if len(globalConfiguration.HelmChartURL) > 0 {
		helmArchivePath := filepath.Join(globalConfiguration.DownloadedArchivesPath, helmArchiveFileName)
		if err := extractArchive(utils, helmArchivePath, globalConfiguration.HelmChartDirectory, 1); err != nil {
			return fmt.Errorf("error extracting helm chart archive: %w", err)
		}
	}
	return nil
}

func extractUIResources(globalConfiguration dwc.GlobalConfiguration, utils sapDwCStageReleaseUtils, pattern, targetPath string) error {
	matches, err := utils.Glob(filepath.Join(globalConfiguration.DownloadedArchivesPath, pattern))
	if err != nil {
		return fmt.Errorf("error globbing UI archive pattern %s: %w", pattern, err)
	}
	if len(matches) != 1 {
		return fmt.Errorf("could not (uniquely) identify archive for UI upload. Found %d archives matching pattern %s: %v", len(matches), pattern, matches)
	}
	uiArchivePath := matches[0]
	if err := extractArchive(utils, uiArchivePath, targetPath, 1); err != nil {
		return fmt.Errorf("error extracting UI archive: %w", err)
	}
	return nil
}

func extractMTARArchive(globalConfiguration dwc.GlobalConfiguration, utils sapDwCStageReleaseUtils) error {
	mtarArchivePath := filepath.Join(globalConfiguration.DownloadedArchivesPath, globalConfiguration.MtarFilePath)
	if err := extractArchive(utils, mtarArchivePath, mtarExtractionDestination, 1); err != nil {
		return fmt.Errorf("error extracting mtar archive: %w", err)
	}
	return nil
}

func extractUIFilesFromMTARArchive(globalConfiguration dwc.GlobalConfiguration, utils sapDwCStageReleaseUtils) error {
	uiArchivePath := filepath.Join(mtarExtractionDestination, dwcUIAppContentModule, globalConfiguration.MtarUIPath, "data.zip")
	err := extractArchive(utils, uiArchivePath, globalConfiguration.DownloadedArchivesPath, 0)
	if err != nil {
		return fmt.Errorf("unable to extract archive '%s': %w", uiArchivePath, err)
	}
	return nil
}

func retrieveUIAppsFromMTA(globalConfiguration dwc.GlobalConfiguration, artifactConfiguration *dwc.ArtifactConfiguration, utils sapDwCStageReleaseUtils) error {
	if err := extractMTARArchive(globalConfiguration, utils); err != nil {
		return fmt.Errorf("unable to extract mtar archive for resource '%s': %w", artifactConfiguration.ResourceName, err)
	}
	if err := extractUIFilesFromMTARArchive(globalConfiguration, utils); err != nil {
		return fmt.Errorf("unable to extract UI app from mtar archive for resource '%s': %w", artifactConfiguration.ResourceName, err)
	}
	appName, err := getAppNameUnique(*artifactConfiguration)
	if err != nil {
		return fmt.Errorf("unable to get app name for resource '%s': %w", artifactConfiguration.ResourceName, err)
	}
	artifactConfiguration.HasArchive = true
	artifactConfiguration.ArchivePattern = fmt.Sprintf("%s.zip", appName)
	return nil
}

func extractArchive(utils sapDwCStageReleaseUtils, archivePath, targetPath string, extractionLevel int) error {
	archiveExists, err := utils.FileExists(archivePath)
	if err != nil {
		return fmt.Errorf("error checking if file %s exists: %w", archivePath, err)
	}
	if !archiveExists {
		return fmt.Errorf("archive %s does not exist", archivePath)
	}
	switch filepath.Ext(archivePath) {
	case ".tar":
		err = utils.Untar(archivePath, targetPath, extractionLevel)
	case ".tgz":
		err = utils.Untar(archivePath, targetPath, extractionLevel)
	case ".zip":
		err = utils.Unzip(archivePath, targetPath)
	case ".mtar":
		err = utils.Unzip(archivePath, targetPath)
	default:
		err = fmt.Errorf("unsupported archive type '%s'", filepath.Ext(archivePath))
	}
	if err != nil {
		return fmt.Errorf("error extracting archive %s: %w", archivePath, err)
	}
	log.Entry().Debugf("successfully extracted archive %s to %s", archivePath, targetPath)
	return nil
}

func getAppNameUnique(artifactConfiguration dwc.ArtifactConfiguration) (string, error) {
	if artifactConfiguration.AppName != "" {
		return artifactConfiguration.AppName, nil
	}
	if len(artifactConfiguration.Apps) == 1 {
		return artifactConfiguration.Apps[0]["name"].(string), nil
	}
	return "", fmt.Errorf("no unique app name provided")
}

func deriveAdditionalDownloadURLs(deriveAdditionalDownloadURLs []dwc.DeriveAdditionalDownloadURLsEntry, artifactUrl string) (map[string]string, error) {
	additionalDownloadURLs := make(map[string]string)
	for _, entry := range deriveAdditionalDownloadURLs {
		regex, err := regexp.Compile(entry.FindPattern)
		if err != nil {
			return nil, fmt.Errorf("failed to compile findPattern regex '%s' of deriveAdditionalDownloadURLs: %w", entry.FindPattern, err)
		}
		matches := regex.FindStringSubmatch(artifactUrl)
		if len(matches) == 0 {
			continue
		}
		replaceWith := entry.ReplaceWith
		for i := len(matches) - 1; i > 0; i-- {
			replaceWith = strings.ReplaceAll(replaceWith, fmt.Sprintf("$%d", i), matches[i]) // expand the capture group references in the replaceWith string
		}
		additionalDownloadURLs[entry.Key] = regex.ReplaceAllString(artifactUrl, replaceWith)
	}
	if len(additionalDownloadURLs) == 0 && len(deriveAdditionalDownloadURLs) > 0 {
		log.Entry().Warnf("None of the provided deriveAdditionalDownloadURLs entries matched the artifact URL '%s'. No additional download URLs will be added to the artifact.", artifactUrl)
	}
	return additionalDownloadURLs, nil
}

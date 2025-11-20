package dwc

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
	"helm.sh/helm/v3/pkg/chartutil"
	"k8s.io/utils/strings/slices"
)

const failedToResolveCurrentDirectory string = "failed to resolve current directory: %w"

type ArtifactDescriptor interface {
	getArtifactUploadFolderStructure() string
	getUploadFileName() string
	getFilePatterns() []string
	GetResourceName() string
	hasFilePatterns() bool
	prepareFiles() error
	buildUploadCommand() (dwcCommand, error)
	needsFileBundling() bool
	getStagesToWatch() []string
	hasStagesToWatch() bool
	watchROIOnly() bool
	getStageWatchPolicy() StageWatchPolicy
}

type descriptorFileUtils interface {
	FileWrite(name string, data []byte, perm os.FileMode) error
	FileRead(path string) ([]byte, error)
	FileRemove(path string) error
	MkdirAll(path string, perm os.FileMode) error
}

type App struct {
	Name                    string `json:"name"`
	NoEuporieTaskCollection bool   `json:"noEuporieTaskCollection"`
	NoRouteAssignment       bool   `json:"noRouteAssignment"`
	AllowStaticRoutes       bool   `json:"allowStaticRoutes"`
}

type ArtifactConfiguration struct {
	AppName                  string                   `json:"appName,omitempty"`
	Apps                     []map[string]interface{} `json:"apps,omitempty"`
	ArchivePattern           string                   `json:"archivePattern,omitempty"`
	ArtifactFilesToUpload    []string                 `json:"artifactFilesToUpload,omitempty"`
	ArtifactPattern          string                   `json:"artifactPattern,omitempty"`
	ArtifactType             string                   `json:"artifactType,omitempty"`
	ExtractFromMTA           bool                     `json:"extractFromMTA,omitempty"`
	HasArchive               bool                     `json:"hasArchive,omitempty"`
	HelmValues               []string                 `json:"helmValues,omitempty"`
	OverwriteHelmDockerImage bool                     `json:"overwriteHelmDockerImage,omitempty"`
	PromotedDockerImage      string                   `json:"promotedDockerImage,omitempty"`
	Repository               string                   `json:"repository,omitempty"`
	ResourceName             string                   `json:"resourceName,omitempty"`
	UiUploadBasePath         string                   `json:"uiUploadBasePath,omitempty"`
	UploadMetadata           []string                 `json:"uploadMetadata,omitempty"`
	UploadType               string                   `json:"uploadType,omitempty"`
}

type DescriptorBase struct {
	StageWatchPolicy  StageWatchPolicy
	FileUtils         descriptorFileUtils
	MetadataCollector MetadataCollector
	AppName           string
	Apps              []App
	ResourceName      string
	FilePatterns      []string
	StagesToWatch     []string
	UploadMetadata    map[string]string
	WatchROI          bool
	buildMetadata     string
}

func (base *DescriptorBase) getStageWatchPolicy() StageWatchPolicy {
	return base.StageWatchPolicy
}

func (base *DescriptorBase) watchROIOnly() bool {
	return base.WatchROI
}

func (base *DescriptorBase) prepareFiles() error { // base implementation is a no-op: can be overridden as a custom pre-bundle hook
	return nil
}

func (base *DescriptorBase) getStagesToWatch() []string {
	return base.StagesToWatch
}

func (base *DescriptorBase) GetResourceName() string {
	return base.ResourceName
}

func (base *DescriptorBase) hasStagesToWatch() bool {
	return len(base.StagesToWatch) != 0
}

func (base *DescriptorBase) collectBuildMetadata() error {
	metadata, err := resolveVectorMetadataBuildEntry(base.MetadataCollector)
	if err != nil {
		return fmt.Errorf("failed to collect build metadata: %w", err)
	}
	base.buildMetadata = metadata
	return nil
}

func (base *DescriptorBase) getUploadFileName() string {
	return "upload.zip"
}

func (base *DescriptorBase) getFilePatterns() []string {
	return base.FilePatterns
}

func (base *DescriptorBase) hasFilePatterns() bool {
	return len(base.FilePatterns) != 0
}

func (base *DescriptorBase) getArtifactUploadFolderStructure() string {
	return ""
}

func (base *DescriptorBase) needsFileBundling() bool {
	return true
}

type UIArtifact struct {
	*DescriptorBase
	UploadBasePath string
}

func (ui *UIArtifact) getArtifactUploadFolderStructure() string {
	return ui.UploadBasePath
}

func (ui *UIArtifact) buildUploadCommand() (dwcCommand, error) {
	if err := ui.collectBuildMetadata(); err != nil {
		return dwcCommand{}, err
	}
	return newUIUploadCommand(ui)
}

func NewUIArtifact(descriptorBase DescriptorBase, uploadBasePath string) (*UIArtifact, error) {
	if uploadBasePath == "" {
		if descriptorBase.AppName == "" {
			return nil, fmt.Errorf("uiUploadBasePath parameter not set but is mandatory if appName is not set")
		}
		uploadBasePath = fmt.Sprintf("webapps/%s/", descriptorBase.AppName)
		log.Entry().Infof("uiUploadBasePath parameter not set. Defaulting to %s", uploadBasePath)
	}
	return &UIArtifact{DescriptorBase: &descriptorBase,
		UploadBasePath: uploadBasePath,
	}, nil
}

type JavaArtifact struct {
	*DescriptorBase
	ArtifactURL            string
	AdditionalDownloadURLs map[string]string
}

func (java *JavaArtifact) buildUploadCommand() (dwcCommand, error) {
	if err := java.collectBuildMetadata(); err != nil {
		return dwcCommand{}, err
	}
	return newJavaUploadCommand(java)
}

func NewJavaArtifact(descriptorBase DescriptorBase, artifactURL string, additionalDownloadURLs map[string]string) *JavaArtifact {
	return &JavaArtifact{DescriptorBase: &descriptorBase, ArtifactURL: artifactURL, AdditionalDownloadURLs: additionalDownloadURLs}
}

type MTAArtifact struct {
	*DescriptorBase
	ArtifactURL            string
	AdditionalDownloadURLs map[string]string
}

func (mta *MTAArtifact) needsFileBundling() bool {
	return false
}

func (mta *MTAArtifact) buildUploadCommand() (dwcCommand, error) {
	if err := mta.collectBuildMetadata(); err != nil {
		return dwcCommand{}, err
	}
	return newMTAUploadCommand(mta)
}

func NewMTAArtifact(descriptorBase DescriptorBase, artifactURL string, additionalDownloadURLs map[string]string) *MTAArtifact {
	return &MTAArtifact{DescriptorBase: &descriptorBase, ArtifactURL: artifactURL, AdditionalDownloadURLs: additionalDownloadURLs}
}

type DockerArtifact struct {
	*DescriptorBase
	ContainerImageLocator  string
	AdditionalDownloadURLs map[string]string
}

// prepareFiles checks if the manifest file is specified via file pattern matching and if it exists on the fs.
func (docker *DockerArtifact) prepareFiles() error {
	for _, pattern := range docker.FilePatterns {
		if pattern == "manifest.yml" {
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf(failedToResolveCurrentDirectory, err)
			}
			if _, err := os.Stat(path.Join(wd, pattern)); errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("manifest.yml specified with artifactFilesToUpload, but it was not found: %w", err)
			} else if err != nil {
				return fmt.Errorf("manifest.yml specified with artifactFilesToUpload, but failed to check if it exists: %w", err)
			} else {
				return nil
			}
		}
	}
	return errors.New("manifest.yml was not specified with artifactFilesToUpload, but it is required for docker uploads")
}

func (docker *DockerArtifact) buildUploadCommand() (dwcCommand, error) {
	if err := docker.collectBuildMetadata(); err != nil {
		return dwcCommand{}, err
	}
	return newDockerUploadCommand(docker)
}

func NewDockerArtifact(descriptorBase DescriptorBase, containerImageLocator string, additionalDownloadURLs map[string]string) *DockerArtifact {
	return &DockerArtifact{DescriptorBase: &descriptorBase, ContainerImageLocator: containerImageLocator, AdditionalDownloadURLs: additionalDownloadURLs}
}

type OrbitArtifact struct {
	*DescriptorBase
	ContainerImageLocator  string
	HelmChartDirectory     string
	AdditionalDownloadURLs map[string]string
}

func (orbit *OrbitArtifact) buildUploadCommand() (dwcCommand, error) {
	if err := orbit.collectBuildMetadata(); err != nil {
		return dwcCommand{}, err
	}
	return newOrbitUploadCommand(orbit)
}

func (orbit *OrbitArtifact) prepareFiles() error {
	cfFilePatternDetected, helmFilePatternDetected := slices.Contains(orbit.FilePatterns, "cf"), slices.Contains(orbit.FilePatterns, orbit.HelmChartDirectory)
	if cfFilePatternDetected && helmFilePatternDetected { // handle piper defaults where both patterns are specified
		if err1, err2 := orbit.prepareHelmFiles(), orbit.prepareCfFiles(); err1 != nil && err2 != nil {
			return fmt.Errorf("your repository does neither contain a correct helm/ nor a correct cf/ folder to upload your custom orbit component. helm validation err: %s, cf validation err: %s", err1, err2)
		}
		return nil
	}
	if cfFilePatternDetected {
		return orbit.prepareCfFiles()
	}
	if helmFilePatternDetected {
		return orbit.prepareHelmFiles()
	}
	return fmt.Errorf("neither helm nor cf folder where specified as patterns for artifactFilesToUpload. But one of them is required depending on your orbit deployment strategy. (k8s/cf)")
}

func (orbit *OrbitArtifact) prepareHelmFiles() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf(failedToResolveCurrentDirectory, err)
	}
	if _, err := os.Stat(path.Join(wd, orbit.HelmChartDirectory)); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("directory '%s' specified with artifactFilesToUpload, but it does not exist: %w", orbit.HelmChartDirectory, err)
	} else if err != nil {
		return fmt.Errorf("directory '%s' specified with artifactFilesToUpload, but failed to check if it exists: %w", orbit.HelmChartDirectory, err)
	} else {
		return nil
	}
}

func (orbit *OrbitArtifact) prepareCfFiles() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf(failedToResolveCurrentDirectory, err)
	}
	if _, err := os.Stat(path.Join(wd, "cf/manifest.yml")); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cf specified with artifactFilesToUpload, but it does not contain manifest.yml or does not exist: %w", err)
	} else if err != nil {
		return fmt.Errorf("cf specified with artifactFilesToUpload, but failed to check if cf/manifest.yml exists: %w", err)
	} else {
		return nil
	}
}

func NewOrbitArtifact(descriptorBase DescriptorBase, containerImageLocator, helmChartDirectory string, additionalDownloadURLs map[string]string) *OrbitArtifact {
	return &OrbitArtifact{
		DescriptorBase:         &descriptorBase,
		ContainerImageLocator:  containerImageLocator,
		HelmChartDirectory:     helmChartDirectory,
		AdditionalDownloadURLs: additionalDownloadURLs,
	}
}

type HelmArtifact struct {
	*DescriptorBase
	ChartUtils               helmChartUtils
	ContainerImage, ImageTag string
	PatchImageSpec           bool
	ValueFiles               []string
	YAMLUtils                yamlUtils
	HelmChartDirectory       string
}

func (helm *HelmArtifact) prepareFiles() error {
	if len(helm.ValueFiles) > 0 {
		if err := handleHelmValueFiles(helm); err != nil {
			return err
		}
	} else {
		log.Entry().Debugf("No defined Helm value files. Using default: %s/%s", helm.HelmChartDirectory, ValuesFileName)
	}
	if helm.PatchImageSpec {
		err := patchImageSpec(helm)
		if err != nil {
			return err
		}
	}
	return nil
}

func patchImageSpec(helm *HelmArtifact) error {
	looksLikeAChart, err := helm.ChartUtils.IsChartDir(helm.HelmChartDirectory)
	if err != nil {
		return fmt.Errorf("error validating helm chart: %w", err)
	}
	if !looksLikeAChart {
		return fmt.Errorf("error validating helm chart. The directory '%s' seems not to be a valid helm chart", helm.HelmChartDirectory)
	}
	valuesFilePath := filepath.Join(helm.HelmChartDirectory, ValuesFileName)
	helmValues, err := helm.ChartUtils.ReadValuesFile(valuesFilePath)
	if err != nil {
		return fmt.Errorf("error parsing helm values: %w", err)
	}

	err = updateImageSpec(helm, helmValues, valuesFilePath)
	if err != nil {
		return err
	}

	yaml, err := helmValues.YAML()
	if err != nil {
		return fmt.Errorf("unable to patch image.repository and image.tag values in %s: %w", valuesFilePath, err)
	}
	log.Entry().Infof("Patched %s to:\n%s", valuesFilePath, yaml)

	if err := os.MkdirAll(filepath.Dir(valuesFilePath), os.ModePerm); err != nil {
		return fmt.Errorf(
			"unable to create directory %s which is required to store Helm values file: %w",
			filepath.Dir(valuesFilePath),
			err)
	}
	if err := helm.FileUtils.FileWrite(valuesFilePath, []byte(yaml), os.ModePerm); err != nil {
		return fmt.Errorf("unable to patch image.repository and image.tag values in %s: %w", valuesFilePath, err)
	}
	return nil
}

func updateImageSpec(helm *HelmArtifact, helmValues chartutil.Values, valuesFilePath string) error {
	valuesDecoded := helmValues.AsMap()
	imageSettings, ok := valuesDecoded["image"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unable to patch image.repository and image.tag values in %s. image identifier not found", valuesFilePath)
	}
	imageSettings["repository"] = helm.ContainerImage
	imageSettings["tag"] = helm.ImageTag
	return nil
}

func handleHelmValueFiles(helm *HelmArtifact) error {
	log.Entry().Debugf("defined Helm value files: %v", helm.ValueFiles)
	valuesYAML, err := readAndMergeFiles(helm)
	if err != nil {
		return err
	}
	return writeMergedValues(helm, valuesYAML)
}

func readAndMergeFiles(helm *HelmArtifact) (map[string]interface{}, error) {
	valuesYAML := map[string]interface{}{}
	for _, filePath := range helm.ValueFiles {
		currentMap, err := readAndUnmarshalFile(helm, filePath)
		if err != nil {
			return nil, err
		}
		valuesYAML = MergeMaps(valuesYAML, currentMap)
		if err := helm.FileUtils.FileRemove(filePath); err != nil {
			return nil, fmt.Errorf("failed to remove %s: %w", filePath, err)
		}
	}
	return valuesYAML, nil
}

func readAndUnmarshalFile(helm *HelmArtifact, filePath string) (map[string]interface{}, error) {
	currentMap := map[string]interface{}{}

	bytes, err := helm.FileUtils.FileRead(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	if err := helm.YAMLUtils.Unmarshal(bytes, &currentMap); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	return currentMap, nil
}

func writeMergedValues(helm *HelmArtifact, valuesYAML map[string]interface{}) error {
	valuesYAMLBytes, err := helm.YAMLUtils.Marshal(valuesYAML)
	if err != nil {
		return fmt.Errorf("failed to create YAML file: %w", err)
	}
	if err := helm.FileUtils.FileWrite(filepath.Join(helm.HelmChartDirectory, ValuesFileName), valuesYAMLBytes, 0644); err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}
	return nil
}

func (helm *HelmArtifact) buildUploadCommand() (dwcCommand, error) {
	if err := helm.collectBuildMetadata(); err != nil {
		return dwcCommand{}, err
	}
	return newHelmUploadCommand(helm)
}

func NewHelmArtifact(descriptorBase DescriptorBase, patchImageSpec bool, valueFiles []string, containerImage, imageTag, helmChartDirectory string) *HelmArtifact {
	return &HelmArtifact{
		DescriptorBase:     &descriptorBase,
		ChartUtils:         &DefaultHelmChartUtils{},
		PatchImageSpec:     patchImageSpec,
		ContainerImage:     containerImage,
		ImageTag:           imageTag,
		ValueFiles:         valueFiles,
		YAMLUtils:          &DefaultYAMLUtils{},
		HelmChartDirectory: helmChartDirectory,
	}
}

package cmd

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/versioning"
)

type sapDownloadArtifactUtils interface {
	command.ExecRunner
	maven.Utils
	getArtifactCoordinates(config *sapDownloadArtifactOptions) (versioning.Coordinates, error)
	untar(source, destination string, stripComponentLevel int) error
}

type sapDownloadArtifactUtilsBundle struct {
	*command.Command
	*piperhttp.Client
	*piperutils.Files
}

const (
	fullyQualifiedArtifactFormat         = "%v-%v.%v"
	fullyQualifiedArtifactFileNameFormat = "%v.%v"
	failureMessageFQDNArtifactName       = "failed to identify artifact full qualified name"
	failureMessageDownloadArtifact       = "failed to download artifact"
	genericErrorFormat                   = "%s: %w"
	buildToolIsNotSupported              = "build tool '%v' is not supported"
	downloadingArtifactLogFormat         = "Downloading '%s' artifact with name '%s' to '%s'"
)

func (s *sapDownloadArtifactUtilsBundle) untar(source, destination string, stripComponentLevel int) error {
	return piperutils.Untar(source, destination, stripComponentLevel)
}

func (s *sapDownloadArtifactUtilsBundle) getArtifactCoordinates(config *sapDownloadArtifactOptions) (versioning.Coordinates, error) {
	options := &versioning.Options{
		ProjectSettingsFile: config.ProjectSettingsFile,
		GlobalSettingsFile:  config.GlobalSettingsFile,
		M2Path:              config.M2Path,
	}

	artifact, err := versioning.GetArtifact(config.BuildTool, config.FilePath, options, s)
	if err != nil {
		return versioning.Coordinates{}, fmt.Errorf("failed to get artifact: %w", err)
	}

	coordinates, err := artifact.GetCoordinates()
	if err != nil {
		return versioning.Coordinates{}, fmt.Errorf("failed to get coordinates: %w", err)
	}

	return coordinates, nil
}

func newSapDownloadArtifactUtils(config *sapDownloadArtifactOptions) sapDownloadArtifactUtils {
	utils := sapDownloadArtifactUtilsBundle{
		Client:  &piperhttp.Client{},
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Configure HTTP Client
	utils.SetOptions(piperhttp.ClientOptions{TransportTimeout: time.Duration(config.Timeout) * time.Second, MaxRetries: config.MaxRetries})

	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func sapDownloadArtifact(config sapDownloadArtifactOptions, telemetryData *telemetry.CustomData, pipelineEnv *sapDownloadArtifactCommonPipelineEnvironment) {
	utils := newSapDownloadArtifactUtils(&config)

	if err := runSapDownloadArtifact(&config, telemetryData, utils, pipelineEnv); err != nil {
		log.Entry().WithError(err).Fatal("failed to execute step sapDownloadArtifact")
	}
}

func runSapDownloadArtifact(config *sapDownloadArtifactOptions, telemetryData *telemetry.CustomData, utils sapDownloadArtifactUtils, pipelineEnv *sapDownloadArtifactCommonPipelineEnvironment) error {
	pipelineEnv.custom.downloadTargetPath = config.TargetPath

	switch config.BuildTool {
	case "maven":
		return mavenSapDownloadArtifact(config, telemetryData, utils, pipelineEnv)
	case "npm":
		return npmDownloadArtifact(config, telemetryData, utils, pipelineEnv)
	case "mta":
		return mtaDownloadArtifact(config, telemetryData, utils, pipelineEnv)
	case "pip":
		return pipDownloadArtifact(config, telemetryData, utils, pipelineEnv)
	case "golang":
		return golangDownloadArtifact(config, telemetryData, utils, pipelineEnv)
	case "gradle":
		return gradleDownloadArtifact(config, telemetryData, utils, pipelineEnv)
	case "CAP":
		return capDownloadArtifacts(config, telemetryData, utils, pipelineEnv)
	default:
		if len(config.HelmChartURL) > 0 {
			return helmDownloadArtifact(config, telemetryData, utils, pipelineEnv)
		}
		log.Entry().Warn("There is nothing to download")
		return nil
	}
}

func helmDownloadArtifact(config *sapDownloadArtifactOptions, telemetryData *telemetry.CustomData, utils sapDownloadArtifactUtils, pipelineEnv *sapDownloadArtifactCommonPipelineEnvironment) error {
	filename := filepath.Base(config.HelmChartURL)
	path := filepath.Join(config.TargetPath, filename)

	log.Entry().Infof(downloadingArtifactLogFormat, "helm", filename, config.TargetPath)

	if len(config.HelmChartURL) == 0 {
		return fmt.Errorf("unable to identify helm artifact download url")
	}
	header := artifactoryHeaderAddition(config.ArtifactoryToken, config.FromStaging, config.HelmStagingRepositoryUsername, config.HelmStagingRepositoryPassword)
	if err := utils.DownloadFile(config.HelmChartURL, path, header, []*http.Cookie{}); err != nil {
		return fmt.Errorf(genericErrorFormat, failureMessageDownloadArtifact, err)
	}
	pipelineEnv.custom.localHelmChartPath = path

	if config.ExtractPackage {
		if err := untarPackage(path, config.TargetPath, utils); err != nil {
			return fmt.Errorf("package extraction failed: %w", err)
		}
	}

	return nil
}

func capDownloadArtifacts(config *sapDownloadArtifactOptions, telemetryData *telemetry.CustomData, utils sapDownloadArtifactUtils, pipelineEnv *sapDownloadArtifactCommonPipelineEnvironment) error {
	// Download helm charts
	if len(config.HelmChartURL) > 0 {
		if err := helmDownloadArtifact(config, telemetryData, utils, pipelineEnv); err != nil {
			return fmt.Errorf("helm download artifact failed: %w", err)
		}
	} else {
		log.Entry().Warn("HelmChartURL not set. Skipping helm charts downloading.")
	}

	// Download npm	packages
	filenames, err := commonMultipleFileDownloadHandler(config, utils)
	if err != nil {
		return fmt.Errorf(genericErrorFormat, failureMessageDownloadArtifact, err)
	}
	log.Entry().Infof("downloaded artifacts: %v", filenames)

	if config.ExtractPackage {
		for _, filename := range filenames {
			if err := untarPackage(filename, config.TargetPath, utils); err != nil {
				return fmt.Errorf("filename %s package extraction failed: %w", filename, err)
			}

			exists, err := utils.FileExists(filename)
			if !exists {
				return fmt.Errorf("filename %s does not exist after untar package", filename)
			}
			if err != nil {
				return fmt.Errorf("filename %s file exist check error: %w", filename, err)
			}
		}
	}

	return nil
}

func gradleDownloadArtifact(config *sapDownloadArtifactOptions, telemetryData *telemetry.CustomData, utils sapDownloadArtifactUtils, pipelineEnv *sapDownloadArtifactCommonPipelineEnvironment) error {
	// if helmChartURL is filled, downloads only the helm chart, but not the result of the technology build
	if len(config.HelmChartURL) > 0 {
		return helmDownloadArtifact(config, telemetryData, utils, pipelineEnv)
	}
	if len(config.Artifacts) == 0 {
		return fmt.Errorf("there are no artifacts available for downloading")
	}
	commonURLParts := []string{}
	commonURLParts = append(commonURLParts, strings.Split(config.GroupID, ".")...)
	commonURL := filepath.Join(commonURLParts...)

	header := artifactoryHeaderAddition(config.ArtifactoryToken, config.FromStaging, config.StagingServiceRepositoryUser, config.StagingServiceRepositoryPassword)

	for _, artifact := range config.Artifacts {
		name, ok := artifact["name"].(string)
		if !ok {
			return fmt.Errorf("failed to get artifact name")
		}
		id, ok := artifact["id"].(string)
		if !ok {
			return fmt.Errorf("failed to get artifact id")
		}
		log.Entry().Infof(downloadingArtifactLogFormat, "gradle", name, config.TargetPath)

		artifactURL := ""
		if config.FromStaging {
			artifactURL = fmt.Sprintf("%s/%s/%s/%s/%s", config.RepositoryURL, commonURL, id, config.Version, name)
		} else {
			artifactURL = identifyArtifactDownlURL(config, name)
		}
		if len(artifactURL) == 0 {
			return fmt.Errorf("failed to identify artifact url")
		}
		filename := filepath.Join(config.TargetPath, name)
		if err := utils.DownloadFile(artifactURL, filename, header, []*http.Cookie{}); err != nil {
			return fmt.Errorf(genericErrorFormat, failureMessageDownloadArtifact, err)
		}
	}
	return nil
}

func mavenSapDownloadArtifact(config *sapDownloadArtifactOptions, telemetryData *telemetry.CustomData, utils sapDownloadArtifactUtils, pipelineEnv *sapDownloadArtifactCommonPipelineEnvironment) error {
	// if helmChartURL is filled, downloads only the helm chart, but not the result of the technology build
	if len(config.HelmChartURL) > 0 {
		return helmDownloadArtifact(config, telemetryData, utils, pipelineEnv)
	}

	coordinates := versioning.Coordinates{
		GroupID:    config.GroupID,
		ArtifactID: config.ArtifactID,
		Version:    config.Version,
		Packaging:  config.Packaging,
	}

	if err := addMissingCoordinates(config, &coordinates, utils); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("failed to get missing coordinates: %w", err)
	}

	urlParts := []string{}
	if config.FromStaging {
		urlParts = append(urlParts, config.RepositoryURL)
	}

	urlParts = append(urlParts, strings.Split(coordinates.GroupID, ".")...)
	// fullyQualifiedArtifactNameFormat := "%v-%v.%v"

	if config.ArtifactWithDependencies {
		log.Entry().Infof("With dependencies identified as")
		urlParts = append(urlParts, coordinates.ArtifactID, coordinates.Version, fmt.Sprintf("%v-%v-%v-with-dependencies.%v", coordinates.ArtifactID, coordinates.Version, coordinates.Packaging, coordinates.Packaging))
	} else {
		urlParts = append(urlParts, coordinates.ArtifactID, coordinates.Version, fmt.Sprintf(fullyQualifiedArtifactFormat, coordinates.ArtifactID, coordinates.Version, coordinates.Packaging))
	}

	artifactIdentifier := strings.Join(urlParts, "/")
	log.Entry().Infof("Artifact download identifier is : '%v'", artifactIdentifier)

	artifactDownloadUrl := ""

	if config.FromStaging {
		artifactDownloadUrl = artifactIdentifier
	} else {
		artifactDownloadUrl = identifyArtifactDownlURL(config, artifactIdentifier)
	}

	filename := filepath.Join(config.TargetPath, fmt.Sprintf(fullyQualifiedArtifactFileNameFormat, coordinates.ArtifactID, coordinates.Packaging))

	if len(artifactDownloadUrl) == 0 {
		return fmt.Errorf("unable to identify '%v' artifact download url", config.BuildTool)
	}

	header := artifactoryHeaderAddition(config.ArtifactoryToken, config.FromStaging, config.StagingServiceRepositoryUser, config.StagingServiceRepositoryPassword)

	log.Entry().Infof(downloadingArtifactLogFormat, "maven", fmt.Sprintf(fullyQualifiedArtifactFileNameFormat, coordinates.ArtifactID, coordinates.Packaging), config.TargetPath)
	if err := utils.DownloadFile(artifactDownloadUrl, filename, header, []*http.Cookie{}); err != nil {
		return fmt.Errorf(genericErrorFormat, failureMessageDownloadArtifact, err)
	}

	return nil
}

func npmDownloadArtifact(config *sapDownloadArtifactOptions, telemetryData *telemetry.CustomData, utils sapDownloadArtifactUtils, pipelineEnv *sapDownloadArtifactCommonPipelineEnvironment) error {
	// if helmChartURL is filled, downloads only the helm chart, but not the result of the technology build
	if len(config.HelmChartURL) > 0 {
		return helmDownloadArtifact(config, telemetryData, utils, pipelineEnv)
	}

	urlParts := []string{}
	// fullyQualifiedArtifactNameFormat := "%v-%v.%v"
	// npm does not include build metadata (+git_sha) during npm publish to be added to npm registry as npm does not distinguish versions based on build metadata
	// see https://github.com/npm/npm/issues/12825
	if strings.Contains(config.Version, "+") {
		config.Version = strings.Split(config.Version, "+")[0]
	}
	if config.FromStaging {
		urlParts = append(urlParts, config.RepositoryURL)
		// if the artifact id is a scoped npm artifact (eg @test/myNpmPacage) then the packaged npm artifact does contain the scope
		//since npm trims it during npm publish so, need to remove it accordingly from config.ArtifactID
		if strings.Contains(config.ArtifactID, "@") {
			scopedArtifactIdentifier := config.ArtifactID[strings.Index(config.ArtifactID, "/")+1:]
			urlParts = append(urlParts, config.ArtifactID, "-", fmt.Sprintf(fullyQualifiedArtifactFormat, scopedArtifactIdentifier, config.Version, "tgz"))
		} else {
			urlParts = append(urlParts, config.ArtifactID, "-", fmt.Sprintf(fullyQualifiedArtifactFormat, config.ArtifactID, config.Version, "tgz"))
		}

	}

	artifactIdentifier := ""
	if config.FromStaging {
		artifactIdentifier = strings.Join(urlParts, "/")
	} else {
		resultNpm, err := returnArtifactIdentifier(fullyQualifiedArtifactFormat, config.ArtifactID, config.Version, config.BuildTool, "")
		if err != nil {
			return fmt.Errorf(genericErrorFormat, failureMessageFQDNArtifactName, err)
		}
		artifactIdentifier = resultNpm
	}

	filename, err := commonDownloadHandeler(config, utils, artifactIdentifier)
	if err != nil {
		return fmt.Errorf(genericErrorFormat, failureMessageDownloadArtifact, err)
	}

	if config.ExtractPackage {
		if err := untarPackage(filename, config.TargetPath, utils); err != nil {
			return fmt.Errorf("package extraction failed: %w", err)
		}
	}

	return nil
}

func mtaDownloadArtifact(config *sapDownloadArtifactOptions, telemetryData *telemetry.CustomData, utils sapDownloadArtifactUtils, pipelineEnv *sapDownloadArtifactCommonPipelineEnvironment) error {
	// if helmChartURL is filled, downloads only the helm chart, but not the result of the technology build
	if len(config.HelmChartURL) > 0 {
		return helmDownloadArtifact(config, telemetryData, utils, pipelineEnv)
	}

	artifactIdentifier := ""

	if config.FromStaging {
		artifactIdentifier = config.MtarPublishedURL
	} else {
		// mta artifacts follow the npm artifact file format "%v-%v.%v"
		resultMta, err := returnArtifactIdentifier(fullyQualifiedArtifactFormat, config.ArtifactID, config.Version, config.BuildTool, config.MtarPublishedURL)
		if err != nil {
			return fmt.Errorf(genericErrorFormat, failureMessageFQDNArtifactName, err)
		}
		artifactIdentifier = resultMta
	}

	if _, err := commonDownloadHandeler(config, utils, artifactIdentifier); err != nil {
		return fmt.Errorf(genericErrorFormat, failureMessageDownloadArtifact, err)
	}

	return nil
}

func pipDownloadArtifact(config *sapDownloadArtifactOptions, telemetryData *telemetry.CustomData, utils sapDownloadArtifactUtils, pipelineEnv *sapDownloadArtifactCommonPipelineEnvironment) error {
	// if helmChartURL is filled, downloads only the helm chart, but not the result of the technology build
	if len(config.HelmChartURL) > 0 {
		return helmDownloadArtifact(config, telemetryData, utils, pipelineEnv)
	}

	artifactIdentifier := ""

	urlParts := []string{}

	if config.FromStaging {
		// replace dashes with underscores as pip does that during pip publish to repository
		artifactName := strings.ReplaceAll(config.ArtifactID, "-", "_")
		artifactNameInUrl := strings.ReplaceAll(config.ArtifactID, "_", "-")
		urlParts = append(urlParts, strings.TrimRight(config.RepositoryURL, "/")) // sometimes Staging service generate URL with trailing slash
		urlParts = append(urlParts, "packages", artifactNameInUrl, config.Version, fmt.Sprintf(fullyQualifiedArtifactFormat, artifactName, config.Version, "tar.gz"))
		artifactIdentifier = strings.Join(urlParts, "/")
	} else {
		resultPip, err := returnArtifactIdentifier(fullyQualifiedArtifactFormat, config.ArtifactID, config.Version, config.BuildTool, "")
		if err != nil {
			return fmt.Errorf(genericErrorFormat, failureMessageFQDNArtifactName, err)
		}
		artifactIdentifier = resultPip
	}

	filename, err := commonDownloadHandeler(config, utils, artifactIdentifier)
	if err != nil {
		return fmt.Errorf(genericErrorFormat, failureMessageDownloadArtifact, err)
	}

	if config.ExtractPackage {
		if err := untarPackage(filename, config.TargetPath, utils); err != nil {
			return fmt.Errorf("package extraction failed: %w", err)
		}
	}

	return nil
}

func golangDownloadArtifact(config *sapDownloadArtifactOptions, telemetryData *telemetry.CustomData, utils sapDownloadArtifactUtils, pipelineEnv *sapDownloadArtifactCommonPipelineEnvironment) error {
	// if helmChartURL is filled, downloads only the helm chart, but not the result of the technology build
	if len(config.HelmChartURL) > 0 {
		return helmDownloadArtifact(config, telemetryData, utils, pipelineEnv)
	}

	if len(config.Artifacts) < 1 {
		return fmt.Errorf("no golang binaries available for download")
	}
	// Construct url parts which are common
	commonUrlParts := []string{}
	commonUrlParts = append(commonUrlParts, "go", config.GroupID, config.ArtifactID, config.Version)
	commonUrl := strings.Join(commonUrlParts, "/")

	header := artifactoryHeaderAddition(config.ArtifactoryToken, config.FromStaging, config.StagingServiceRepositoryUser, config.StagingServiceRepositoryPassword)

	// Iterate through the binary list and download each
	for _, binaryArtifact := range config.Artifacts {

		if binaryName, ok := binaryArtifact["name"].(string); ok {

			var err error
			var artifactDownloadUrl string

			log.Entry().Infof(downloadingArtifactLogFormat, "golang", binaryName, config.TargetPath)
			binaryPath := fmt.Sprintf("%s/%s", commonUrl, binaryName)

			if config.FromStaging {
				artifactDownloadUrl = fmt.Sprintf("%s/%s", config.RepositoryURL, binaryPath)
			} else {
				artifactDownloadUrl = identifyArtifactDownlURL(config, binaryName)
			}

			filename := filepath.Join(config.TargetPath, binaryName)

			if len(artifactDownloadUrl) != 0 {
				err = utils.DownloadFile(artifactDownloadUrl, filename, header, []*http.Cookie{})
			} else {
				err = fmt.Errorf("failed to identify artifactDownloadUrl")
			}
			if err != nil {
				return fmt.Errorf("failed to download %s: %w", binaryName, err)
			}
		}
	}

	return nil
}

func commonDownloadHandeler(config *sapDownloadArtifactOptions, utils sapDownloadArtifactUtils, artifactIdentifier string) (string, error) {
	log.Entry().Infof("Artifact download identifier is : '%v'", artifactIdentifier)

	artifactDownloadUrl := ""

	if config.FromStaging {
		artifactDownloadUrl = artifactIdentifier
	} else {
		if config.BuildTool == "pip" {
			// For pip, replace dashes with underscores in artifactIdentifier to match pip publish convention
			artifactName := strings.ReplaceAll(config.ArtifactID, "-", "_")
			artifactIdentifier = fmt.Sprintf(fullyQualifiedArtifactFormat, artifactName, config.Version, "tar.gz")
		}
		artifactDownloadUrl = identifyArtifactDownlURL(config, artifactIdentifier)
	}

	fileFormat := ""

	switch config.BuildTool {
	case "mta":
		fileFormat = "mtar"
	case "npm":
		fileFormat = "tgz"
	case "pip":
		fileFormat = "tar.gz"
	default:
		return "", fmt.Errorf("built tool not handled by common download handler")
	}

	fileName := filepath.Join(config.TargetPath, fmt.Sprintf(fullyQualifiedArtifactFileNameFormat, config.ArtifactID, fileFormat))

	if len(artifactDownloadUrl) == 0 {
		return "", fmt.Errorf("unable to identify '%v' Artifact download url", config.BuildTool)
	}

	header := artifactoryHeaderAddition(config.ArtifactoryToken, config.FromStaging, config.StagingServiceRepositoryUser, config.StagingServiceRepositoryPassword)

	log.Entry().Infof(downloadingArtifactLogFormat, config.BuildTool, fmt.Sprintf(fullyQualifiedArtifactFileNameFormat, config.ArtifactID, fileFormat), config.TargetPath)
	if err := utils.DownloadFile(artifactDownloadUrl, fileName, header, []*http.Cookie{}); err != nil {
		return "", fmt.Errorf(genericErrorFormat, failureMessageDownloadArtifact, err)
	}
	return fileName, nil

}

func commonMultipleFileDownloadHandler(config *sapDownloadArtifactOptions, utils sapDownloadArtifactUtils) ([]string, error) {
	var artifactDownloadUrls []string
	var fileFormat string
	fileNames := []string{}

	switch config.BuildTool {
	case "CAP":
		fileFormat = "tgz"
	default:
		return nil, fmt.Errorf("build tool not handled by common Multiple File Download Handler")
	}

	if config.FromStaging {
		artifactDownloadUrls = identifyAllArtifactDownloadURLs(config.ArtifactStagingDownloadURLs, fileFormat)
	} else {
		artifactDownloadUrls = identifyAllArtifactDownloadURLs(config.ArtifactDownloadURLs, fileFormat)
	}

	if len(artifactDownloadUrls) == 0 {
		return nil, fmt.Errorf("unable to identify npm packages from '%v'", config.BuildTool)
	}

	header := artifactoryHeaderAddition(config.ArtifactoryToken, config.FromStaging, config.StagingServiceRepositoryUser, config.StagingServiceRepositoryPassword)

	for _, url := range artifactDownloadUrls {
		splittedURL := strings.Split(url, "/")
		lastIdx := len(splittedURL) - 1
		fileName := splittedURL[lastIdx]
		path := filepath.Join(config.TargetPath, fileName)

		if err := utils.DownloadFile(url, path, header, []*http.Cookie{}); err != nil {
			return nil, fmt.Errorf(genericErrorFormat, failureMessageDownloadArtifact, err)
		}
		fileNames = append(fileNames, path)
	}

	return fileNames, nil
}

func artifactoryHeaderAddition(artifactoryToken string, fromStaging bool, stagingServiceRepositoryUser string, stagingServiceRepositoryPassword string) http.Header {
	header := http.Header{}
	if len(artifactoryToken) > 0 && !fromStaging {
		header = http.Header{"X-JFrog-Art-Api": []string{artifactoryToken}}
	} else if fromStaging && len(stagingServiceRepositoryUser) > 0 && len(stagingServiceRepositoryPassword) > 0 {
		auth := stagingServiceRepositoryUser + ":" + stagingServiceRepositoryPassword
		header = http.Header{"Authorization": []string{"Basic " + base64.StdEncoding.EncodeToString([]byte(auth))}}
	}

	return header

}

func returnArtifactIdentifier(fullyQualifiedArtifactFileNameFormat string, artifactID string, version string, buildTool string, mtarPublishedUrl string) (string, error) {
	switch buildTool {
	case "mta":
		// extract the name from staged url
		mtarUrlComponents := strings.Split(mtarPublishedUrl, "/")
		artifactIdentifier := mtarUrlComponents[len(mtarUrlComponents)-1]
		return artifactIdentifier, nil
	case "npm":
		artifactIdentifier := fmt.Sprintf(fullyQualifiedArtifactFormat, artifactID, version, "tgz")
		// if the artifact id is a scoped npm artifact (eg @test/myNpmPacage) then the packaged npm artifact does contain the scope
		//since npm trims it during npm publish so, need to remove it accordingly from config.ArtifactID
		if strings.Contains(artifactID, "@") {
			scopedArtifactIdentifier := artifactID[strings.Index(artifactID, "/")+1:]
			artifactIdentifier = fmt.Sprintf(fullyQualifiedArtifactFormat, scopedArtifactIdentifier, version, "tgz")
		}
		return artifactIdentifier, nil
	case "pip":
		artifactIdentifier := fmt.Sprintf(fullyQualifiedArtifactFormat, artifactID, version, "tar.gz")
		return artifactIdentifier, nil
	default:
		return "", fmt.Errorf("failed to identify artifacts fully qualified name for type: %v", buildTool)
	}
}

func identifyArtifactDownlURL(config *sapDownloadArtifactOptions, artifactIdentifier string) string {
	for _, url := range config.ArtifactDownloadURLs {
		if contains := strings.Contains(url, artifactIdentifier); contains {
			log.Entry().Infof("Artifact download url is %v", url)
			return url
		}
	}
	return ""
}

func identifyAllArtifactDownloadURLs(artifactDownloadURLs []string, fileFormat string) []string {
	downloadURLs := []string{}

	for _, url := range artifactDownloadURLs {
		if hasSuffix := strings.HasSuffix(url, fileFormat); hasSuffix {
			log.Entry().Infof("Artifact download url is %v", url)
			downloadURLs = append(downloadURLs, url)
		}
	}
	return downloadURLs
}

func addMissingCoordinates(config *sapDownloadArtifactOptions, coordinates *versioning.Coordinates, utils sapDownloadArtifactUtils) error {
	if len(config.GroupID) == 0 || len(config.ArtifactID) == 0 || len(config.Version) == 0 || len(config.Packaging) == 0 {

		descriptorCoordinates, err := utils.getArtifactCoordinates(config)
		if err != nil {
			return err
		}

		if descriptorCoordinates.Packaging == "maven-plugin" {
			descriptorCoordinates.Packaging = "jar"
		}

		if len(coordinates.GroupID) == 0 {
			coordinates.GroupID = descriptorCoordinates.GroupID
		}
		if len(coordinates.ArtifactID) == 0 {
			coordinates.ArtifactID = descriptorCoordinates.ArtifactID
		}
		if len(coordinates.Version) == 0 {
			coordinates.Version = descriptorCoordinates.Version
		}
		if len(coordinates.Packaging) == 0 {
			coordinates.Packaging = descriptorCoordinates.Packaging
		}
	}
	return nil
}

func untarPackage(fileName, targetPath string, utils sapDownloadArtifactUtils) error {
	log.Entry().Infof("Extracting artifact '%v' to '%v'", fileName, targetPath)
	return utils.untar(fileName, targetPath, 1)
}

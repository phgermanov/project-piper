package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	piperOsCmd "github.com/SAP/jenkins-library/cmd"
	"github.com/SAP/jenkins-library/pkg/config"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/events"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/cumulus"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/staging"
)

const (
	CAPRepoFormat           = "maven,npm"
	groupIDandOutputFileErr = "you must pass groupId and outputFile"
)

type stagingServiceUtils interface {
	GetClient() *piperhttp.Client
	FileExists(filename string) (bool, error)
	FileRead(path string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	GetStageConfig() (config.StepConfig, error)
	createRepositories(stagingObj staging.StagingInterface) (map[string]map[string]interface{}, error)
	closeGroup(stagingObj staging.StagingInterface) (map[string]interface{}, error)
	getStagedArtifactURLs(stagingObj staging.StagingInterface) ([]string, error)
}

type stagingServiceUtilsBundle struct {
	*piperhttp.Client
	*piperutils.Files
}

func (s *stagingServiceUtilsBundle) GetClient() *piperhttp.Client {
	return s.Client
}

func (s *stagingServiceUtilsBundle) GetStageConfig() (config.StepConfig, error) {
	return piperOsCmd.GetStageConfig()
}

func (s *stagingServiceUtilsBundle) createRepositories(stagingObj staging.StagingInterface) (map[string]map[string]interface{}, error) {
	return createRepositories(stagingObj)
}

func (s *stagingServiceUtilsBundle) closeGroup(stagingObj staging.StagingInterface) (map[string]interface{}, error) {
	return closeGroup(stagingObj)
}

func (c *stagingServiceUtilsBundle) getStagedArtifactURLs(stagingObj staging.StagingInterface) ([]string, error) {
	return getStagedArtifactURLs(stagingObj)
}

func newStagingServiceUtils(config *sapCallStagingServiceOptions) *stagingServiceUtilsBundle {
	utils := stagingServiceUtilsBundle{
		Client: &piperhttp.Client{},
		Files:  &piperutils.Files{},
	}
	// Configure HTTP Client
	utils.SetOptions(piperhttp.ClientOptions{TransportTimeout: time.Duration(config.Timeout) * time.Second})
	return &utils
}

func sapCallStagingService(config sapCallStagingServiceOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *sapCallStagingServiceCommonPipelineEnvironment) {
	utils := newStagingServiceUtils(&config)
	err := runSapCallStagingService(&config, telemetryData, commonPipelineEnvironment, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runSapCallStagingService(config *sapCallStagingServiceOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *sapCallStagingServiceCommonPipelineEnvironment, utils stagingServiceUtils) error {
	client := utils.GetClient()
	sapStaging := staging.Staging{
		TenantId:         config.TenantID,
		TenantSecret:     config.TenantSecret,
		Username:         config.Username,
		Password:         config.Password,
		Profile:          config.Profile,
		RepositoryFormat: config.RepositoryFormat,
		Url:              config.Url,
		Group:            config.GroupID,
		GroupIdFile:      config.GroupIDFile,
		OutputFile:       config.OutputFile,
		Metadata:         config.Metadata,
		RepositoryId:     config.RepositoryID,
		State:            config.State,
		Query:            config.Query,
		HTTPClient:       client,
		BuildTool:        config.BuildTool,
	}
	switch config.Action {
	case "createGroup":
		err := createGroup(&sapStaging)
		commonPipelineEnvironment.custom.stagingGroupID = sapStaging.Group
		return err
	case "createRepositories":
		stageConfig, err := utils.GetStageConfig()
		if err != nil {
			return fmt.Errorf("failed to get stage configuration: %w", err)
		}

		log.Entry().Debugf("stage config: %v", stageConfig)

		// use automatic detection in case no repositoryFormats are provided
		repoFormatFromBuildTool := identifyCorrectStagingRepoFormat(&sapStaging)
		if len(config.RepositoryFormats) == 0 {
			log.Entry().Debug("Automatic detection of staging repositories since no repositoryFormats provided.")

			config.RepositoryFormats = append(config.RepositoryFormats, strings.Split(repoFormatFromBuildTool, ",")...)

			// allow containerization and helm packaging in addition to build tool usage
			if !slices.Contains(config.RepositoryFormats, "docker") && ((stageConfig.Config["kanikoExecute"] != nil && stageConfig.Config["kanikoExecute"].(bool)) || (stageConfig.Config["cnbBuild"] != nil && stageConfig.Config["cnbBuild"].(bool))) {
				config.RepositoryFormats = append(config.RepositoryFormats, "docker")
			}
			if !slices.Contains(config.RepositoryFormats, "helm") && stageConfig.Config["helmExecute"] != nil && stageConfig.Config["helmExecute"].(bool) {
				config.RepositoryFormats = append(config.RepositoryFormats, "helm")
			}
		}
		sapStaging.RepositoryFormats = config.RepositoryFormats
		// results key reflects repository format
		results, err := utils.createRepositories(&sapStaging)
		if err != nil {
			return err
		}
		return configureCreateRepositories(results, repoFormatFromBuildTool, commonPipelineEnvironment, config, utils)
	case "createRepository":
		resultMap, err := createRepository(&sapStaging)
		if err == nil {
			configureCreateRepository(&sapStaging, resultMap, commonPipelineEnvironment, config, utils)
		}
		return err
	case "updateGroupMetadata":
		return updateGroupMetadata(&sapStaging)
	case "close":
		// ToDo: clean up files containing credentials
		if _, err := utils.closeGroup(&sapStaging); err != nil {
			return err
		}

		// at this point no new artifacts will be added to the staging repos,
		// so we can collect list of staged artifacts for use in other steps.
		stagedArtifactURLs, err := utils.getStagedArtifactURLs(&sapStaging)
		if err != nil {
			return err
		}
		commonPipelineEnvironment.custom.stagedArtifactURLs = stagedArtifactURLs
		log.Entry().Infof("Staged Artifact URL's are : '%v'", stagedArtifactURLs)

		return err
	case "promote":
		promotedArtifacts, err := promoteGroup(&sapStaging, config.ArtifactPattern)
		if err != nil {
			log.Entry().Debugf("commonPipelineEnvironment: %v", commonPipelineEnvironment)
			return err
		}

		commonPipelineEnvironment.custom.helmChartURL = promotedArtifacts.PromotedHelmChartURL
		commonPipelineEnvironment.custom.promotedArtifactURLs = promotedArtifacts.PromotedArtifactURLs
		commonPipelineEnvironment.custom.promotedDockerImages = promotedArtifacts.PromotedDockerImages
		log.Entry().Infof("Promoted Artifact URL's are : '%v'", commonPipelineEnvironment.custom.promotedArtifactURLs)
		log.Entry().Infof("Promoted Docker images are : '%v'", commonPipelineEnvironment.custom.promotedDockerImages)

		stagingServiceData := events.StagingServiceData{
			HeadCommitID:         config.HeadCommitID,
			CommitID:             config.CommitID,
			GitURL:               config.GitURL,
			NpmBuildArtifacts:    config.NpmBuildArtifacts,
			MavenBuildArtifacts:  config.MavenBuildArtifacts,
			MtaBuildArtifacts:    config.MtaBuildArtifacts,
			PromotedArtifactURLs: commonPipelineEnvironment.custom.promotedArtifactURLs,
		}
		events.SendPromotedArtifactsEvent(stagingServiceData)

		// Clear CPE to allow Release stage to use the artifactory credentials from the kubernetes cluster itself and not staging credentials
		commonPipelineEnvironment.container.registryURL = "toBeEmptied"
		commonPipelineEnvironment.container.repositoryUsername = "toBeEmptied"
		commonPipelineEnvironment.container.repositoryPassword = "toBeEmptied"
		commonPipelineEnvironment.custom.repositoryUsername = "toBeEmptied"
		commonPipelineEnvironment.custom.repositoryPassword = "toBeEmptied"
		commonPipelineEnvironment.custom.dockerConfigJSON = "toBeEmptied"

		status := cumulus.ReleaseStatus{Status: "promoted"}
		// ignore error since format is in our hands
		releaseStatus, _ := json.Marshal(status)
		commonPipelineEnvironment.custom.releaseStatus = string(releaseStatus)

		return err
	case "getRepositoryCredentials":
		return getRepositoryCredentials(&sapStaging)
	case "getRepositoryBom":
		return getRepositoryBom(&sapStaging)
	case "getGroupMetadata":
		return getGroupMetadata(&sapStaging)
	case "updateRepositoryMetadata":
		return updateRepositoryMetadata(&sapStaging)
	case "getRepositoryMetadata":
		return getRepositoryMetadata(&sapStaging)
	case "searchGroupMetadata":
		return searchGroupMetadata(&sapStaging)
	case "getGroupBom":
		return getGroupBom(&sapStaging)
	case "signGroup":
		return signGroup(&sapStaging)
	default:
		return errors.New("Wrong action")
	}
}

func configureCreateRepository(stagingObj staging.StagingInterface, resultMap map[string]interface{}, commonPipelineEnvironment *sapCallStagingServiceCommonPipelineEnvironment, config *sapCallStagingServiceOptions, utils stagingServiceUtils) {
	commonPipelineEnvironment.custom.stagingRepositoryID = resultMap["repository"].(string)
	commonPipelineEnvironment.custom.repositoryURL = resultMap["repositoryURL"].(string)
	commonPipelineEnvironment.custom.repositoryID = resultMap["repository"].(string)
	commonPipelineEnvironment.custom.repositoryFormat = stagingObj.GetRepositoryFormat()
	commonPipelineEnvironment.custom.repositoryUsername = resultMap["user"].(string)
	commonPipelineEnvironment.custom.repositoryPassword = resultMap["password"].(string)
	switch stagingObj.GetRepositoryFormat() {
	case "docker":
		configPath, err := staging.CreateDockerConfigJSON(resultMap["repositoryURL"].(string), resultMap["user"].(string), resultMap["password"].(string), config.DockerConfigJSON, utils)
		if err != nil {
			log.Entry().Error(err)
		}
		commonPipelineEnvironment.custom.dockerConfigJSON = configPath
		commonPipelineEnvironment.container.registryURL = fmt.Sprintf("https://%v", resultMap["repositoryURL"])
	case "helm":
		// ToDo
	case "go":
		// ToDo
	case "pypi":
		// ToDo
	}
}

func configureCreateRepositories(results map[string]map[string]interface{}, repoFormatFromBuildTool string, commonPipelineEnvironment *sapCallStagingServiceCommonPipelineEnvironment, config *sapCallStagingServiceOptions, utils stagingServiceUtils) error {
	for repoFormat, repoConfig := range results {
		if (repoFormatFromBuildTool == CAPRepoFormat && repoFormat == "npm") || (repoFormatFromBuildTool == repoFormat && repoFormat != "docker" && repoFormat != "helm") {
			commonPipelineEnvironment.custom.repositoryURL = repoConfig["repositoryURL"].(string)
			commonPipelineEnvironment.custom.repositoryFormat = repoFormat
			commonPipelineEnvironment.custom.repositoryUsername = repoConfig["user"].(string)
			commonPipelineEnvironment.custom.repositoryPassword = repoConfig["password"].(string)
		}

		switch repoFormat {
		case "docker":
			commonPipelineEnvironment.container.registryURL = fmt.Sprintf("https://%v", repoConfig["repositoryURL"])
			commonPipelineEnvironment.container.repositoryUsername = repoConfig["user"].(string)
			commonPipelineEnvironment.container.repositoryPassword = repoConfig["password"].(string)

			configPath, err := staging.CreateDockerConfigJSON(repoConfig["repositoryURL"].(string), repoConfig["user"].(string), repoConfig["password"].(string), config.DockerConfigJSON, utils)
			if err != nil {
				log.Entry().Error(err)
			}
			commonPipelineEnvironment.custom.dockerConfigJSON = configPath
		case "helm":
			commonPipelineEnvironment.custom.helmRepositoryURL = repoConfig["repositoryURL"].(string)
			commonPipelineEnvironment.custom.helmRepositoryUsername = repoConfig["user"].(string)
			commonPipelineEnvironment.custom.helmRepositoryPassword = repoConfig["password"].(string)
		case "maven":
			commonPipelineEnvironment.custom.mavenRepositoryURL = repoConfig["repositoryURL"].(string)
			commonPipelineEnvironment.custom.repositoryID = repoConfig["repository"].(string)
			commonPipelineEnvironment.custom.mavenRepositoryUsername = repoConfig["user"].(string)
			commonPipelineEnvironment.custom.mavenRepositoryPassword = repoConfig["password"].(string)
		case "npm":
			commonPipelineEnvironment.custom.npmRepositoryURL = repoConfig["repositoryURL"].(string)
			commonPipelineEnvironment.custom.npmRepositoryUsername = repoConfig["user"].(string)
			commonPipelineEnvironment.custom.npmRepositoryPassword = repoConfig["password"].(string)
		case "pypi":
			commonPipelineEnvironment.custom.pipRepositoryURL = repoConfig["repositoryURL"].(string)
			commonPipelineEnvironment.custom.pipRepositoryUsername = repoConfig["user"].(string)
			commonPipelineEnvironment.custom.pipRepositoryPassword = repoConfig["password"].(string)
		case "raw":
			commonPipelineEnvironment.custom.rawRepositoryURL = repoConfig["repositoryURL"].(string)
			commonPipelineEnvironment.custom.rawRepositoryUsername = repoConfig["user"].(string)
			commonPipelineEnvironment.custom.rawRepositoryPassword = repoConfig["password"].(string)
		}
	}
	return nil
}

func createGroup(stagingObj staging.StagingInterface) error {
	if stagingObj.GetProfile() == "" {
		return errors.New("You must pass profile")
	}
	_, err := stagingObj.LoginAndReceiveAuthToken()
	if err != nil {
		return err
	}
	groupID, err := stagingObj.CreateStagingGroup()
	if err != nil {
		return err
	}
	stagingObj.SetGroup(groupID)
	return nil
}

func createRepository(stagingObj staging.StagingInterface) (map[string]interface{}, error) {
	if (len(stagingObj.GetGroup())+len(stagingObj.GetGroupIdFile())) == 0 || (len(stagingObj.GetBuildTool()) == 0 && len(stagingObj.GetRepositoryFormat()) == 0) {
		return nil, errors.New("you must pass groupId and buildTool / repositoryFormat")
	}
	groupID, err := stagingObj.ReadGroupIdFromFile()
	stagingObj.SetGroup(groupID)
	if err != nil {
		return nil, err
	}
	_, err = stagingObj.LoginAndReceiveAuthToken()
	if err != nil {
		return nil, err
	}
	resultMap, err := stagingObj.CreateStagingRepositoryWithResultMap()
	if err != nil {
		return nil, err
	}
	return resultMap, nil
}

func createRepositories(stagingObj staging.StagingInterface) (map[string]map[string]interface{}, error) {
	if len(stagingObj.GetGroup()) == 0 || len(stagingObj.GetRepositoryFormats()) == 0 {
		log.SetErrorCategory(log.ErrorConfiguration)
		return nil, errors.New("you must pass groupId and repositoryFormats")
	}

	_, err := stagingObj.LoginAndReceiveAuthToken()
	if err != nil {
		return nil, err
	}
	resultMap, err := stagingObj.CreateMultipleStagingRepositories()
	if err != nil {
		return nil, err
	}
	return resultMap, nil
}

func getStagedArtifactURLs(stagingObj staging.StagingInterface) ([]string, error) {
	if (len(stagingObj.GetGroup()) + len(stagingObj.GetGroupIdFile())) == 0 {
		return nil, errors.New("you must pass groupId")
	}
	groupID, err := stagingObj.ReadGroupIdFromFile()
	stagingObj.SetGroup(groupID)
	if err != nil {
		return nil, err
	}
	_, err = stagingObj.LoginAndReceiveAuthToken()
	if err != nil {
		return nil, err
	}
	stagedArtifacts, err := stagingObj.GetStagedArtifactURLs()
	if err != nil {
		return nil, err
	}
	return stagedArtifacts, nil
}

func closeGroup(stagingObj staging.StagingInterface) (map[string]interface{}, error) {
	if (len(stagingObj.GetGroup()) + len(stagingObj.GetGroupIdFile())) == 0 {
		return nil, errors.New("You must pass groupId")
	}
	groupID, err := stagingObj.ReadGroupIdFromFile()
	stagingObj.SetGroup(groupID)
	if err != nil {
		return nil, err
	}
	_, err = stagingObj.LoginAndReceiveAuthToken()
	if err != nil {
		return nil, err
	}
	resultMap, err := stagingObj.CloseStagingGroup()
	if err != nil {
		return nil, err
	}
	return resultMap, nil
}

func promoteGroup(stagingObj staging.StagingInterface, artifactPattern string) (*staging.PromotedArtifacts, error) {
	if stagingObj.GetOutputFile() == "" || (len(stagingObj.GetGroup())+len(stagingObj.GetGroupIdFile())) == 0 {
		log.Entry().Debugf("Staging-service object: %v", stagingObj)
		return nil, errors.New(groupIDandOutputFileErr)
	}
	groupID, err := stagingObj.ReadGroupIdFromFile()
	log.Entry().Infof("Group ID: %s", groupID)
	stagingObj.SetGroup(groupID)
	if err != nil {
		log.Entry().Debugf("Failed to read group ID from file")
		return nil, err
	}
	_, err = stagingObj.LoginAndReceiveAuthToken()
	if err != nil {
		log.Entry().Debugf("Failed to login and receive auth token: %v", err)
		return nil, err
	}

	promotedArtifacts, err := stagingObj.PromoteGroup(artifactPattern)
	if err != nil {
		log.Entry().Debugf("Failed to promote group: %v", err)
	} else {
		log.Entry().Infof("Promoted artifacts: %v", promotedArtifacts)
	}

	return promotedArtifacts, err
}

func signGroup(stagingObj staging.StagingInterface) error {
	if (len(stagingObj.GetGroup()) + len(stagingObj.GetGroupIdFile())) == 0 {
		return errors.New("You must pass groupId")
	}
	groupID, err := stagingObj.ReadGroupIdFromFile()
	stagingObj.SetGroup(groupID)
	if err != nil {
		return err
	}
	_, err = stagingObj.LoginAndReceiveAuthToken()
	if err != nil {
		return err
	}
	_, err = stagingObj.SignGroup()
	if err != nil {
		return err
	}
	return nil
}

func updateGroupMetadata(stagingObj staging.StagingInterface) error {
	if (len(stagingObj.GetGroup())+len(stagingObj.GetGroupIdFile())) == 0 || stagingObj.GetMetadataField() == "" {
		return errors.New("You must pass groupId and metadata")
	}
	groupID, err := stagingObj.ReadGroupIdFromFile()
	stagingObj.SetGroup(groupID)
	if err != nil {
		return err
	}
	_, err = stagingObj.LoginAndReceiveAuthToken()
	if err != nil {
		return err
	}
	err = stagingObj.SetGroupMetadata()
	if err != nil {
		return err
	}
	return nil
}

func getRepositoryCredentials(stagingObj staging.StagingInterface) error {
	if stagingObj.GetOutputFile() == "" || (len(stagingObj.GetGroup())+len(stagingObj.GetGroupIdFile())) == 0 || stagingObj.GetRepositoryId() == "" {
		return errors.New("You must pass groupId, outputFile and repositoryId")
	}
	groupID, err := stagingObj.ReadGroupIdFromFile()
	stagingObj.SetGroup(groupID)
	if err != nil {
		return err
	}
	_, err = stagingObj.LoginAndReceiveAuthToken()
	if err != nil {
		return err
	}
	_, err = stagingObj.GetRepositoryCredentials()
	if err != nil {
		return err
	}
	return nil
}

func getRepositoryBom(stagingObj staging.StagingInterface) error {
	if (len(stagingObj.GetGroup())+len(stagingObj.GetGroupIdFile())) == 0 || stagingObj.GetOutputFile() == "" || stagingObj.GetRepositoryId() == "" {
		return errors.New("You must pass groupId, outputFile and repositoryId")
	}
	groupID, err := stagingObj.ReadGroupIdFromFile()
	stagingObj.SetGroup(groupID)
	if err != nil {
		return err
	}
	_, err = stagingObj.LoginAndReceiveAuthToken()
	if err != nil {
		return err
	}
	_, err = stagingObj.GetRepositoryBom()
	if err != nil {
		return err
	}
	return nil
}

func getGroupBom(stagingObj staging.StagingInterface) error {
	if (len(stagingObj.GetGroup())+len(stagingObj.GetGroupIdFile())) == 0 || stagingObj.GetOutputFile() == "" {
		return errors.New(groupIDandOutputFileErr)
	}
	groupID, err := stagingObj.ReadGroupIdFromFile()
	stagingObj.SetGroup(groupID)
	if err != nil {
		return err
	}
	_, err = stagingObj.LoginAndReceiveAuthToken()
	if err != nil {
		return err
	}
	_, err = stagingObj.GetGroupBom()
	if err != nil {
		return err
	}
	return nil
}

func getGroupMetadata(stagingObj staging.StagingInterface) error {
	if stagingObj.GetOutputFile() == "" || (len(stagingObj.GetGroup())+len(stagingObj.GetGroupIdFile())) == 0 {
		return errors.New(groupIDandOutputFileErr)
	}
	groupID, err := stagingObj.ReadGroupIdFromFile()
	stagingObj.SetGroup(groupID)
	if err != nil {
		return err
	}
	_, err = stagingObj.LoginAndReceiveAuthToken()
	if err != nil {
		return err
	}
	err = stagingObj.GetGroupMetadata()
	if err != nil {
		return err
	}
	return nil
}

func updateRepositoryMetadata(stagingObj staging.StagingInterface) error {
	if (len(stagingObj.GetGroup())+len(stagingObj.GetGroupIdFile())) == 0 || stagingObj.GetMetadataField() == "" || stagingObj.GetRepositoryId() == "" {
		return errors.New("You must pass groupId, metadata and repositoryId")
	}
	groupID, err := stagingObj.ReadGroupIdFromFile()
	stagingObj.SetGroup(groupID)
	if err != nil {
		return err
	}
	_, err = stagingObj.LoginAndReceiveAuthToken()
	if err != nil {
		return err
	}
	err = stagingObj.SetRepositoryMetadata()
	if err != nil {
		return err
	}
	return nil
}

func searchGroupMetadata(stagingObj staging.StagingInterface) error {
	if stagingObj.GetOutputFile() == "" || (len(stagingObj.GetGroup())+len(stagingObj.GetGroupIdFile())) == 0 || stagingObj.GetQuery() == "" {
		return errors.New("You must pass groupId, outputFile and query")
	}
	groupID, err := stagingObj.ReadGroupIdFromFile()
	stagingObj.SetGroup(groupID)
	if err != nil {
		return err
	}
	_, err = stagingObj.LoginAndReceiveAuthToken()
	if err != nil {
		return err
	}
	_, err = stagingObj.SearchMetadataGroup()
	if err != nil {
		return err
	}
	return nil
}

func identifyCorrectStagingRepoFormat(stagingObj staging.StagingInterface) string {
	return stagingObj.IdentifyCorrectStagingRepoFormat()
}

func getRepositoryMetadata(stagingObj staging.StagingInterface) error {
	if stagingObj.GetOutputFile() == "" || (len(stagingObj.GetGroup())+len(stagingObj.GetGroupIdFile())) == 0 || stagingObj.GetRepositoryId() == "" {
		return errors.New("You must pass groupId, outputFile and repositoryId")
	}
	groupID, err := stagingObj.ReadGroupIdFromFile()
	stagingObj.SetGroup(groupID)
	if err != nil {
		return err
	}
	_, err = stagingObj.LoginAndReceiveAuthToken()
	if err != nil {
		return err
	}
	err = stagingObj.GetRepositoryMetadata()
	if err != nil {
		return err
	}
	return nil
}

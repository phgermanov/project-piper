package staging

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/bmatcuk/doublestar/v4"
)

const (
	LoginEndpoint               = "/login"
	CloseEndpoint               = "/group/close"
	CreateGroupEndpoint         = "/group/create"
	MetadataGroupEndpoint       = "/group/metadata"
	PromoteEndpoint             = "/group/promote/async"
	GetStateEndpoint            = "/group/state"
	CreateRepositoryEndpoint    = "/repository/create"
	JsonFormat                  = "application/json"
	GetRepositoryCredentials    = "/repository/credentials/"
	GetRepositoryBom            = "/repository/BOM/"
	GetGroupBom                 = "/group/BOM"
	MetadataRepositoryEndpoint  = "/repository/metadata/"
	SearchMetadataGroupEndpoint = "/group/metadata/search"
	SignGroup                   = "/signing/group"
)

const (
	responseFromRequestMessage = "Response from request:"
	CAPRepoFormat              = "maven,npm"
	CAP                        = "CAP"
)

type StagingInterface interface {
	LoginAndReceiveAuthToken() (string, error)
	CreateMultipleStagingRepositories() (map[string]map[string]interface{}, error)
	CreateStagingGroup() (string, error)
	CreateStagingRepository() (string, error)
	CreateStagingRepositoryWithResultMap() (map[string]interface{}, error)
	CloseStagingGroup() (map[string]interface{}, error)
	PromoteGroup(string) (*PromotedArtifacts, error)
	SignGroup() (string, error)
	SetGroupMetadata() error
	GetGroupMetadata() error
	GetGroupBom() (string, error)
	GetRepositoryBom() (string, error)
	GetRepositoryCredentials() (string, error)
	GetRepositoryMetadata() error
	SetRepositoryMetadata() error
	GetStagedArtifactURLs() ([]string, error)
	SearchMetadataGroup() (string, error)
	getRequest(string, bool) (map[string]interface{}, string, error)
	postRequest([]uint8, string) (string, error)
	GetOutputFile() string
	GetProfile() string
	GetGroup() string
	SetGroup(string)
	GetGroupIdFile() string
	GetBuildTool() string
	ReadGroupIdFromFile() (string, error)
	GetMetadataField() string
	GetRepositoryId() string
	GetQuery() string
	GetRepositoryFormat() string
	GetRepositoryFormats() []string
	IdentifyCorrectStagingRepoFormat() string
}

type Staging struct {
	TenantId          string
	TenantSecret      string
	Username          string
	Password          string
	Profile           string
	Url               string
	Group             string
	OutputFile        string
	GroupIdFile       string
	Metadata          string
	Token             string
	RepositoryId      string
	Query             string
	State             string
	HTTPClient        piperhttp.Sender
	BuildTool         string
	RepositoryFormat  string
	RepositoryFormats []string
}

func (staging *Staging) LoginAndReceiveAuthToken() (string, error) {
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", staging.Username)
	data.Set("password", staging.Password)
	fullURL := staging.Url + LoginEndpoint
	if staging.Group != "" {
		data.Set("group", staging.Group)
	}
	opts := piperhttp.ClientOptions{
		Username: staging.TenantId,
		Password: staging.TenantSecret,
		Logger:   log.Entry(),
	}
	headers := map[string][]string{
		"Content-Type": {"application/x-www-form-urlencoded"},
		"Accept":       {JsonFormat},
	}
	staging.HTTPClient.SetOptions(opts)
	resp, err := staging.HTTPClient.SendRequest(http.MethodPost, fullURL, strings.NewReader(data.Encode()), headers, nil)
	if err != nil {
		return "", err
	}

	log.Entry().Info("Response code from post request to " + fullURL + " is " + resp.Status)
	log.Entry().Info("Login token was generated successfully")
	responseData, err := parseResponseBody(resp)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}
	mapJSON, err := byteSliceToJSONMap(responseData)
	if err != nil {
		return "", fmt.Errorf("failed to convert response data: %w", err)
	}
	defer resp.Body.Close()
	staging.Token = mapJSON["access_token"].(string)
	opts = piperhttp.ClientOptions{}
	staging.HTTPClient.SetOptions(opts)
	return mapJSON["access_token"].(string), nil
}
func (staging *Staging) CreateStagingGroup() (string, error) {

	reqBody, err := json.Marshal(map[string]string{
		"profile": staging.Profile,
	})
	if err != nil {
		return "", err
	}

	fullURL := staging.Url + CreateGroupEndpoint
	response, err := staging.postRequest(reqBody, fullURL)
	if err != nil {
		return "", err
	}
	log.Entry().Info(responseFromRequestMessage, response)

	if len(staging.OutputFile) > 0 {
		jsonFile, err := os.Create(staging.OutputFile)
		if err != nil {
			return "", err
		}
		jsonFile.WriteString(response)
		defer jsonFile.Close()
		log.Entry().Info("The groupId was written to the file")
	}

	result, err := byteSliceToJSONMap([]byte(response))
	if err != nil {
		return "", err
	}
	return result["group"].(string), nil
}

func (staging *Staging) CreateMultipleStagingRepositories() (map[string]map[string]interface{}, error) {
	results := map[string]map[string]interface{}{}
	var err error

	for _, repoFormat := range staging.RepositoryFormats {
		results[repoFormat], err = staging.createTypedStagingRepository(repoFormat)
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

func (staging *Staging) createTypedStagingRepository(repoFormat string) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	log.Entry().Debugf("Requesting staging repository of type '%s'", repoFormat)

	staging.SetRepositoryFormat(repoFormat)

	reqBody, err := json.Marshal(map[string]string{
		"repositoryFormat": repoFormat,
	})
	if err != nil {
		return result, fmt.Errorf("failed to marshal staging repository request: %w", err)
	}
	fullURL := staging.Url + CreateRepositoryEndpoint
	response, err := staging.postRequest(reqBody, fullURL)
	if err != nil {
		return result, fmt.Errorf("failed to create staging repository: %w", err)
	}
	log.Entry().Infof("Repository of type '%v' was created successfully", repoFormat)

	result, err = byteSliceToJSONMap([]byte(response))
	if err != nil {
		return result, err
	}
	return result, nil
}

func (staging *Staging) CreateStagingRepository() (string, error) {
	repositoryFormat := ""
	if len(staging.RepositoryFormat) > 0 {
		repositoryFormat = staging.RepositoryFormat
	} else if len(staging.BuildTool) > 0 {
		repositoryFormat = staging.IdentifyCorrectStagingRepoFormat()
	}

	if repositoryFormat == CAPRepoFormat {
		return "", fmt.Errorf("wrong repo format '%s', please use createRepositories action for build tool CAP", repositoryFormat)
	}

	log.Entry().Infof("Requesting staging repository of type '%s'", repositoryFormat)

	staging.SetRepositoryFormat(repositoryFormat)

	reqBody, err := json.Marshal(map[string]string{
		"repositoryFormat": repositoryFormat,
	})

	if err != nil {
		return "", err
	}
	fullURL := staging.Url + CreateRepositoryEndpoint
	response, err := staging.postRequest(reqBody, fullURL)
	if err != nil {
		return "", err
	}
	log.Entry().Info("Repository was created successfully")
	jsonFile, err := os.Create(staging.OutputFile)
	if err != nil {
		return "", err
	}
	jsonFile.WriteString(response)
	defer jsonFile.Close()
	log.Entry().Info("Repository info stored successfully in the file")
	return response, nil
}

func (staging *Staging) CreateStagingRepositoryWithResultMap() (map[string]interface{}, error) {
	response, err := staging.CreateStagingRepository()
	if err != nil {
		return nil, err
	}
	result, err := byteSliceToJSONMap([]byte(response))
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (staging *Staging) GetStagedArtifactURLs() ([]string, error) {
	var groupBOM GroupBOMResponse
	stagedArtifactsURLs := []string{}

	response, err := staging.GetGroupBom()
	if err != nil {
		return nil, err
	}

	log.Entry().Debug("Response from get BOM request: " + response)

	err = json.Unmarshal([]byte(response), &groupBOM)
	if err != nil {
		return nil, err
	}

	for _, repository := range groupBOM.Repositories {
		for _, component := range repository.Bom.Components {
			for _, asset := range component.Assets {
				stagedArtifactsURLs = append(stagedArtifactsURLs, asset.URL)
			}
		}
	}
	return stagedArtifactsURLs, nil
}

func (staging *Staging) CloseStagingGroup() (map[string]interface{}, error) {
	fullURL := staging.Url + CloseEndpoint
	response, err := staging.postRequest(nil, fullURL)
	if err != nil {
		return nil, err
	}
	log.Entry().Info(responseFromRequestMessage, response)

	responseMap, err := byteSliceToJSONMap([]byte(response))
	if err != nil {
		return nil, err
	}
	return responseMap, nil
}

func (staging *Staging) PromoteGroup(artifactPattern string) (*PromotedArtifacts, error) {
	promotedArtifacts := &PromotedArtifacts{}
	fullURL := staging.Url + PromoteEndpoint
	response, err := staging.postRequest(nil, fullURL)
	if err != nil {
		return nil, err
	}
	log.Entry().Info(responseFromRequestMessage, response)
	start := time.Now()
	for {
		time.Sleep(5 * time.Second)
		getURL := staging.Url + GetStateEndpoint
		responseMap, responseString, err := staging.getRequest(getURL, true)
		if err != nil {
			return nil, err
		}

		if responseMap["state"] != "promoting" {
			log.Entry().Info(responseString)

			if err := handleStates(staging, promotedArtifacts, responseMap["state"], start, responseString, artifactPattern); err != nil {
				return nil, err
			}

			break
		}
	}

	return promotedArtifacts, nil
}

func handleStates(staging *Staging, promotedArtifacts *PromotedArtifacts, state interface{}, startTime time.Time, responseString, artifactPattern string) (err error) {
	switch state {
	case "released":
		return handleReleasedState(staging, promotedArtifacts, responseString, artifactPattern)
	case "failed promote":
		return errors.New("promote failed")
	default:
		timeoutTime := startTime.Add(time.Duration(3) * time.Hour)
		if timeoutTime.Before(time.Now()) {
			return errors.New("promote timeout")
		}
	}

	return nil
}

func handleReleasedState(staging *Staging, promotedArtifacts *PromotedArtifacts, responseString, artifactPattern string) (err error) {
	var stagingPromote StagingPromote
	if err := json.Unmarshal([]byte(responseString), &stagingPromote); err != nil {
		return err
	}

	repositories := stagingPromote.ResponseFromPromote.Repositories
	for _, repository := range repositories {
		if len(artifactPattern) > 0 {
			if err := populatePromotedArtifactUrls(artifactPattern, repository, promotedArtifacts); err != nil {
				return err
			}
		} else {
			promotedArtifacts.PromotedArtifactURLs = append(promotedArtifacts.PromotedArtifactURLs, repository.Result...)
		}

		// 306-314: trying to determine promoted docker images
		if len(repository.List) == 0 {
			continue
		}

		populatePromotedDockerImages(repository, promotedArtifacts)
	}

	// if there are multiple repositories then we can assume its a multi build with helm
	// handle helm related cpe rewriting for custom/helmChartUrl
	if len(repositories) > 1 {
		if err := handleMultipleRepositories(staging, repositories, promotedArtifacts); err != nil {
			return err
		}
	}

	if err := writeToFile(staging.OutputFile, responseString); err != nil {
		return err
	}

	return nil
}

func handleMultipleRepositories(staging *Staging, repositories []Repository, promotedArtifacts *PromotedArtifacts) (err error) {
	log.Entry().Debug("Trying to identify Helm chart URL")
	if err := staging.handleHelmPromotedUrl(repositories, promotedArtifacts); err != nil {
		return err
	}

	// we remove the helm chart url from the promotedArtifact list because the cpe helmChartUrl will help download the chart and in sapDownloadArtifact
	// we dont want a match to hit the helm chart artifact match when downloading a technology artifact like (maven,npm..)
	if len(promotedArtifacts.PromotedHelmChartURL) > 0 {
		promotedArtifacts.PromotedArtifactURLs, err = removeURLFromPromotedList(promotedArtifacts.PromotedArtifactURLs, promotedArtifacts.PromotedHelmChartURL)
		if err != nil {
			return err
		}
	}

	return nil
}

func populatePromotedArtifactUrls(artifactPattern string, repository Repository, promotedArtifacts *PromotedArtifacts) error {
	for _, url := range repository.Result {
		filename := filepath.Base(url)
		ok, err := doublestar.Match(artifactPattern, filename)
		if err != nil {
			return fmt.Errorf("failed to match artifact: %w", err)
		}

		if ok {
			log.Entry().Debugf("keeping %s", url)
			promotedArtifacts.PromotedArtifactURLs = append(promotedArtifacts.PromotedArtifactURLs, url)

		} else {
			log.Entry().Debugf("ignoring %s", url)
		}
	}

	return nil
}

func populatePromotedDockerImages(repository Repository, promotedArtifacts *PromotedArtifacts) {
	for _, dockerArtifact := range repository.List {
		if dockerArtifact.Success {
			promotedArtifacts.PromotedDockerImages = append(promotedArtifacts.PromotedDockerImages, dockerArtifact.Image)
		}
	}
}

func (staging *Staging) handleHelmPromotedUrl(promotedRepositories []Repository, promotedArtifacts *PromotedArtifacts) error {
	// give a call to getGroupBOM
	response, err := staging.GetGroupBom()
	if err != nil {
		return err
	}

	log.Entry().Debug("Response from get BOM request: " + response)

	var groupBOM GroupBOMResponse
	err = json.Unmarshal([]byte(response), &groupBOM)
	if err != nil {
		return err
	}
	promotedHelmChartURL := findPromotedHelmChart(promotedRepositories, groupBOM)
	if promotedHelmChartURL != "" {
		promotedArtifacts.PromotedHelmChartURL = promotedHelmChartURL
	}
	return nil
}

func findPromotedHelmChart(promotedRepositories []Repository, groupBOM GroupBOMResponse) string {
	helmChartRepositoryId := getHelmChartRepositoryId(groupBOM)
	for _, promotedRepository := range promotedRepositories {
		if helmChartRepositoryId == promotedRepository.Repository {
			return promotedRepository.Result[0]
		}
	}
	return ""
}

func getHelmChartRepositoryId(groupBOM GroupBOMResponse) string {
	for _, repository := range groupBOM.Repositories {
		if repository.Bom.Format == "helm" &&
			len(repository.Bom.Components) == 1 &&
			len(repository.Bom.Components[0].Assets) == 1 {
			return repository.Repository
		}
	}
	return ""
}

func isHelmChart(repository GroupBOMRepository, promotedRepository Repository) bool {
	return repository.Bom.Format == "helm" &&
		repository.Repository == promotedRepository.Repository &&
		// we expect only one helm chart
		len(repository.Bom.Components) == 1 &&
		// we expect only one asset with the helm artifact
		len(repository.Bom.Components[0].Assets) == 1
}

func removeURLFromPromotedList(promotedArtifactURLs []string, url string) ([]string, error) {
	var i *int
	for index, promotedURL := range promotedArtifactURLs {
		if promotedURL == url {
			i = &index
			break
		}
	}
	if i == nil {
		return nil, fmt.Errorf("failed to delete url: %v not found", url)
	}
	return append(promotedArtifactURLs[:*i], promotedArtifactURLs[*i+1:]...), nil
}

func (staging *Staging) SetGroupMetadata() error {
	fullUrl := staging.Url + MetadataGroupEndpoint
	response, err := staging.postRequest([]uint8(staging.Metadata), fullUrl)
	log.Entry().Info(responseFromRequestMessage, response)
	if err != nil {
		return err
	}
	return nil
}
func (staging *Staging) GetGroupMetadata() error {
	fullURL := staging.Url + MetadataGroupEndpoint
	_, response, err := staging.getRequest(fullURL, false)
	if err != nil {
		return err
	}
	err = writeToFile(staging.OutputFile, response)
	if err != nil {
		return err
	}
	log.Entry().Info("Group metadata stored successfully in the file")
	return nil
}
func (staging *Staging) SetRepositoryMetadata() error {
	fullUrl := staging.Url + MetadataRepositoryEndpoint + staging.RepositoryId
	response, err := staging.postRequest([]uint8(staging.Metadata), fullUrl)
	log.Entry().Info(responseFromRequestMessage, response)
	if err != nil {
		return err
	}
	return nil
}
func (staging *Staging) GetRepositoryMetadata() error {
	fullURL := staging.Url + MetadataRepositoryEndpoint + staging.RepositoryId
	_, response, err := staging.getRequest(fullURL, false)
	if err != nil {
		return err
	}
	err = writeToFile(staging.OutputFile, response)
	if err != nil {
		return err
	}
	log.Entry().Info("Repository metadata stored successfully in the file")
	return nil
}
func (staging *Staging) SearchMetadataGroup() (string, error) {
	fullURL := staging.Url + SearchMetadataGroupEndpoint + "?q=" + staging.Query
	if staging.State != "" {
		fullURL = fullURL + "&state=" + staging.State
	}
	_, response, err := staging.getRequest(fullURL, false)
	if err != nil {
		return "", err
	}
	log.Entry().Info(response)
	err = writeToFile(staging.OutputFile, response)
	if err != nil {
		return "", err
	}
	log.Entry().Info("Searched metadata stored successfully in the file")
	return response, nil
}
func (staging *Staging) GetRepositoryCredentials() (string, error) {
	fullURL := staging.Url + GetRepositoryCredentials + staging.RepositoryId
	_, response, err := staging.getRequest(fullURL, false)
	if err != nil {
		return "", err
	}
	err = writeToFile(staging.OutputFile, response)
	if err != nil {
		return "", err
	}
	log.Entry().Info("Repository credentials stored successfully in the file")
	return response, nil
}
func (staging *Staging) GetRepositoryBom() (string, error) {
	fullURL := staging.Url + GetRepositoryBom + staging.RepositoryId
	_, response, err := staging.getRequest(fullURL, false)
	if err != nil {
		return "", err
	}
	err = writeToFile(staging.OutputFile, response)
	if err != nil {
		return "", err
	}
	log.Entry().Info("Repository bom stored successfully in the file")
	return response, nil
}
func (staging *Staging) GetGroupBom() (string, error) {
	fullURL := staging.Url + GetGroupBom
	_, response, err := staging.getRequest(fullURL, false)
	if err != nil {
		return "", err
	}
	err = writeToFile(staging.OutputFile, response)
	if err != nil {
		return "", err
	}
	log.Entry().Info("Group bom stored successfully in the file")
	return response, nil
}
func (staging *Staging) SignGroup() (string, error) {
	fullURL := staging.Url + SignGroup
	response, err := staging.postRequest([]uint8{}, fullURL)
	if err != nil {
		return "", err
	}
	log.Entry().Info(response)
	return response, nil
}
func writeToFile(fileName, response string) error {
	jsonFile, err := os.Create(filepath.Clean(fileName))
	if err != nil {
		return err
	}
	jsonFile.WriteString(response)
	defer jsonFile.Close()
	return nil
}
func (staging *Staging) getRequest(url string, state bool) (map[string]interface{}, string, error) {
	bearer := "Bearer " + staging.Token
	headers := map[string][]string{
		"Authorization": {bearer},
	}
	resp, err := staging.HTTPClient.SendRequest(http.MethodGet, url, nil, headers, nil)
	if err != nil {
		return nil, "", err
	}
	responseData, err := parseResponseBody(resp)
	if err != nil {
		return nil, "", err
	}
	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !statusOK {
		log.Entry().Info("The request to "+url+" failed with status code: ", resp.StatusCode)
		return nil, "", errors.New(string(responseData))
	}
	log.Entry().Info("Response code from get request to " + url + " is " + resp.Status)
	var mapJSON map[string]interface{}
	if state {
		mapJSON, err = byteSliceToJSONMap(responseData)
	}
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	return mapJSON, string(responseData), nil
}
func (staging *Staging) postRequest(reqBody []uint8, url string) (string, error) {
	bearer := "Bearer " + staging.Token
	headers := map[string][]string{
		"Authorization": {bearer},
		"Accept":        {JsonFormat},
		"Content-Type":  {JsonFormat},
	}
	resp, err := staging.HTTPClient.SendRequest(http.MethodPost, url, bytes.NewBuffer(reqBody), headers, nil)
	if err != nil {
		return "", err
	}
	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	responseData, err := parseResponseBody(resp)
	if !statusOK {
		log.Entry().Info("The request to "+url+" failed with status code: ", resp.StatusCode)
		return string(responseData), errors.New(string(responseData))
	}
	if err != nil {
		return "", err
	}
	log.Entry().Info("Response code from post request to " + url + " is " + resp.Status)

	defer resp.Body.Close()
	return string(responseData), nil
}
func parseResponseBody(resp *http.Response) ([]byte, error) {
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return responseData, nil
}
func byteSliceToJSONMap(slice []byte) (map[string]interface{}, error) {
	var mapJSON map[string]interface{}
	err := json.Unmarshal(slice, &mapJSON)
	if err != nil {
		return mapJSON, err
	}
	return mapJSON, nil
}

func (staging *Staging) ReadGroupIdFromFile() (string, error) {
	if len(staging.Group) > 0 {
		return staging.Group, nil
	}
	// Deprecated: to be removed -> solely rely on provided groupID
	jsonFile, err := os.Open(staging.GetGroupIdFile())
	if err != nil {
		return "", err
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	result, err := byteSliceToJSONMap([]byte(byteValue))
	if err != nil {
		return "", err
	}
	return result["group"].(string), err
}

func (staging *Staging) IdentifyCorrectStagingRepoFormat() string {
	switch staging.BuildTool {
	case "mta", "gradle":
		return "maven"
	case "golang":
		return "raw" // we only want to support uploading binaries, thus we request a raw repo
	case "pip":
		return "pypi"
	case CAP:
		return CAPRepoFormat
	default:
		return staging.BuildTool
	}
}

func (staging *Staging) GetOutputFile() string {
	return staging.OutputFile
}
func (staging *Staging) GetProfile() string {
	return staging.Profile
}
func (staging *Staging) GetGroup() string {
	return staging.Group
}
func (staging *Staging) SetGroup(group string) {
	staging.Group = group
}
func (staging *Staging) GetGroupIdFile() string {
	return staging.GroupIdFile
}
func (staging *Staging) GetBuildTool() string {
	return staging.BuildTool
}
func (staging *Staging) GetMetadataField() string {
	return staging.Metadata
}
func (staging *Staging) GetRepositoryId() string {
	return staging.RepositoryId
}
func (staging *Staging) GetQuery() string {
	return staging.Query
}
func (staging *Staging) GetRepositoryFormat() string {
	return staging.RepositoryFormat
}
func (staging *Staging) GetRepositoryFormats() []string {
	return staging.RepositoryFormats
}
func (staging *Staging) SetRepositoryFormat(repositoryFormat string) {
	staging.RepositoryFormat = repositoryFormat
}

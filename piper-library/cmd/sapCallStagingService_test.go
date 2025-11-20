//go:build unit
// +build unit

package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/config"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/staging"
)

type stagingMock struct {
	staging.StagingInterface
	group                      string
	groupIdFile                interface{}
	loginError                 error
	closeStagingGroupError     error
	closeStagingGroupResultMap map[string]interface{}
	createMultipleReposError   error
	getStagedArtifactsError    error
	readGroupIdFromFileError   error
	repositoryFormats          []string
}

func (s *stagingMock) CreateMultipleStagingRepositories() (map[string]map[string]interface{}, error) {
	if s.createMultipleReposError != nil {
		return map[string]map[string]interface{}{}, s.createMultipleReposError
	}
	return map[string]map[string]interface{}{"maven": {"testKey": "testVal"}}, nil
}

func (s *stagingMock) CreateStagingGroup() (string, error) {
	return "mock", nil
}
func (s *stagingMock) LoginAndReceiveAuthToken() (string, error) {
	if s.loginError != nil {
		return "", s.loginError
	}
	return "login", nil
}
func (s *stagingMock) CreateStagingRepository() (string, error) {
	return "mock", nil
}
func (s *stagingMock) CreateStagingRepositoryWithResultMap() (map[string]interface{}, error) {
	var mockedMap map[string]interface{}
	return mockedMap, nil
}
func (s *stagingMock) CloseStagingGroup() (map[string]interface{}, error) {
	var mockedMap map[string]interface{}
	if s.closeStagingGroupError != nil {
		return nil, s.closeStagingGroupError
	}

	if s.closeStagingGroupResultMap != nil {
		mockedMap = s.closeStagingGroupResultMap
	}
	return mockedMap, nil
}
func (s *stagingMock) GetGroup() string {
	return s.group
}
func (s *stagingMock) SetGroupMetadata() error {
	return nil
}
func (s *stagingMock) PromoteGroup(string) (*staging.PromotedArtifacts, error) {
	return &staging.PromotedArtifacts{
		PromotedArtifactURLs: []string{"https://common.repo.test/1/test-app"},
		PromotedDockerImages: []string{"docker.common.repo.test/test-image:1.22.3"},
		PromotedHelmChartURL: "https://common.repo.test/helm/test-0.1.0.tgz",
	}, nil
}
func (s *stagingMock) GetRepositoryCredentials() (string, error) {
	return "", nil
}
func (s *stagingMock) GetRepositoryBom() (string, error) {
	return "", nil
}
func (s *stagingMock) GetStagedArtifactURLs() ([]string, error) {
	if s.getStagedArtifactsError != nil {
		return nil, s.getStagedArtifactsError
	}
	return []string{
		"https://test.staging.service/stage/repository/16ec86818081/test-app",
		"https://test.staging.service/stage/repository/1d7a8cd48081/test.tgz",
	}, nil
}
func (s *stagingMock) GetGroupBom() (string, error) {
	return "", nil
}
func (s *stagingMock) SignGroup() (string, error) {
	return "", nil
}
func (s *stagingMock) GetGroupMetadata() error {
	return nil
}
func (s *stagingMock) SetRepositoryMetadata() error {
	return nil
}
func (s *stagingMock) GetRepositoryMetadata() error {
	return nil
}
func (s *stagingMock) SearchMetadataGroup() (string, error) {
	return "search", nil
}
func (s *stagingMock) GetProfile() string {
	return "mock"
}
func (s *stagingMock) GetOutputFile() string {
	return "output"
}
func (s *stagingMock) GetGroupIdFile() string {
	if s.groupIdFile != nil {
		return s.groupIdFile.(string)
	}
	return "idfile"
}
func (s *stagingMock) GetMetadataField() string {
	return "metadata"
}
func (s *stagingMock) SetRepositoryFormat() error {
	return nil
}
func (s *stagingMock) GetRepositoryFormat() string {
	return "format"
}
func (s *stagingMock) GetRepositoryFormats() []string {
	return s.repositoryFormats
}
func (s *stagingMock) GetRepositoryId() string {
	return "id"
}
func (s *stagingMock) SetGroup(group string) {
}
func (s *stagingMock) ReadGroupIdFromFile() (string, error) {
	if s.readGroupIdFromFileError != nil {
		return "", s.readGroupIdFromFileError
	}
	return "file", nil
}
func (s *stagingMock) GetQuery() string {
	return "query"
}
func (s *stagingMock) GetBuildTool() string {
	return "mock"
}

type stagingServiceUtilsMock struct {
	*piperhttp.Client
	*mock.FilesMock
	closeGroupError            error
	createRepoError            error
	getStagedArtifactURLsError error
	stageConfig                config.StepConfig
	stageConfigError           error
	stagingObject              staging.StagingInterface
}

func (s *stagingServiceUtilsMock) GetClient() *piperhttp.Client {
	return s.Client
}

func (s *stagingServiceUtilsMock) GetStageConfig() (config.StepConfig, error) {
	if s.stageConfigError != nil {
		return config.StepConfig{}, s.stageConfigError
	}
	return s.stageConfig, nil
}

func (s *stagingServiceUtilsMock) createRepositories(stagingObj staging.StagingInterface) (map[string]map[string]interface{}, error) {
	result := map[string]map[string]interface{}{}
	s.stagingObject = stagingObj
	if s.createRepoError != nil {
		return result, s.createRepoError
	}

	for _, repo := range stagingObj.GetRepositoryFormats() {
		result[repo] = map[string]interface{}{
			"repositoryURL": fmt.Sprintf("%v-repoUrl", repo),
			"repository":    repo,
			"user":          fmt.Sprintf("%v-user", repo),
			"password":      fmt.Sprintf("%v-password", repo),
		}
	}
	return result, nil
}

func (s *stagingServiceUtilsMock) closeGroup(stagingObj staging.StagingInterface) (map[string]interface{}, error) {
	if s.closeGroupError != nil {
		return nil, s.closeGroupError
	}

	closeStagingGroupResultMap := map[string]interface{}{
		"closed": true,
		"group":  "111",
		"repositories": []interface{}{
			map[string]interface{}{
				"repository":       "test-staging-repo-id",
				"repositoryClosed": true,
			},
		}}
	return closeStagingGroupResultMap, nil
}

func (s *stagingServiceUtilsMock) getStagedArtifactURLs(stagingObj staging.StagingInterface) ([]string, error) {
	if s.getStagedArtifactURLsError != nil {
		return nil, s.getStagedArtifactURLsError
	}

	repoArtifactURLs := []string{
		"https://test.staging.service/stage/repository/16ec86818081/test-app",
		"https://test.staging.service/stage/repository/1d7a8cd48081/test.tgz",
	}
	return repoArtifactURLs, nil
}

func newStagingServiceUtilsMock() *stagingServiceUtilsMock {
	return &stagingServiceUtilsMock{
		FilesMock: &mock.FilesMock{},
		Client:    &piperhttp.Client{},
	}
}

func TestRunSapCallStagingServiceMultiRepos(t *testing.T) {
	t.Parallel()

	t.Run("success case", func(t *testing.T) {
		t.Parallel()
		config := sapCallStagingServiceOptions{Action: "createRepositories", BuildTool: "maven"}
		cpe := sapCallStagingServiceCommonPipelineEnvironment{}
		utilsMock := newStagingServiceUtilsMock()
		err := runSapCallStagingService(&config, nil, &cpe, utilsMock)
		assert.NoError(t, err)
		assert.Equal(t, []string{"maven"}, utilsMock.stagingObject.GetRepositoryFormats())
	})

	t.Run("success case - additional kanikoExecute & helmExecute ", func(t *testing.T) {
		t.Parallel()
		config := sapCallStagingServiceOptions{Action: "createRepositories", BuildTool: "maven"}
		cpe := sapCallStagingServiceCommonPipelineEnvironment{}
		utilsMock := newStagingServiceUtilsMock()
		utilsMock.stageConfig.Config = map[string]interface{}{"kanikoExecute": true, "helmExecute": true}
		err := runSapCallStagingService(&config, nil, &cpe, utilsMock)
		assert.NoError(t, err)
		assert.Equal(t, []string{"maven", "docker", "helm"}, utilsMock.stagingObject.GetRepositoryFormats())
	})

	t.Run("success case - additional cnbBuild", func(t *testing.T) {
		t.Parallel()
		config := sapCallStagingServiceOptions{Action: "createRepositories", BuildTool: "maven"}
		cpe := sapCallStagingServiceCommonPipelineEnvironment{}
		utilsMock := newStagingServiceUtilsMock()
		utilsMock.stageConfig.Config = map[string]interface{}{"cnbBuild": true}
		err := runSapCallStagingService(&config, nil, &cpe, utilsMock)
		assert.NoError(t, err)
		assert.Equal(t, []string{"maven", "docker"}, utilsMock.stagingObject.GetRepositoryFormats())
	})

	t.Run("success case - repository formats", func(t *testing.T) {
		t.Parallel()
		tt := []struct {
			name     string
			config   sapCallStagingServiceOptions
			expected []string
		}{
			{name: "tool: mta", config: sapCallStagingServiceOptions{BuildTool: "mta"}, expected: []string{"maven"}},
			{name: "tool: gradle", config: sapCallStagingServiceOptions{BuildTool: "gradle"}, expected: []string{"maven"}},
			{name: "tool: maven", config: sapCallStagingServiceOptions{BuildTool: "maven"}, expected: []string{"maven"}},
			{name: "tool: golang", config: sapCallStagingServiceOptions{BuildTool: "golang"}, expected: []string{"raw"}},
			{name: "tool: pip", config: sapCallStagingServiceOptions{BuildTool: "pip"}, expected: []string{"pypi"}},
			{name: "tool: CAP", config: sapCallStagingServiceOptions{BuildTool: "CAP"}, expected: []string{"maven", "npm"}},
			{name: "provided repos", config: sapCallStagingServiceOptions{BuildTool: "pip", RepositoryFormats: []string{"maven", "raw"}}, expected: []string{"maven", "raw"}},
		}

		for _, test := range tt {
			test.config.Action = "createRepositories"
			cpe := sapCallStagingServiceCommonPipelineEnvironment{}
			utilsMock := newStagingServiceUtilsMock()
			err := runSapCallStagingService(&test.config, nil, &cpe, utilsMock)
			assert.NoError(t, err, test.name)
			assert.Equal(t, test.expected, test.config.RepositoryFormats, test.name)
			assert.Equal(t, test.expected, utilsMock.stagingObject.GetRepositoryFormats(), test.name)
		}

	})

	t.Run("success case - cpe", func(t *testing.T) {
		t.Parallel()
		config := sapCallStagingServiceOptions{Action: "createRepositories", BuildTool: "maven", RepositoryFormats: []string{"docker", "helm", "maven", "npm", "pypi", "raw"}}
		cpe := sapCallStagingServiceCommonPipelineEnvironment{}
		utilsMock := newStagingServiceUtilsMock()
		err := runSapCallStagingService(&config, nil, &cpe, utilsMock)
		assert.NoError(t, err)

		assert.Equal(t, "maven-repoUrl", cpe.custom.repositoryURL)
		assert.Equal(t, "maven", cpe.custom.repositoryFormat)
		assert.Equal(t, "maven-user", cpe.custom.repositoryUsername)
		assert.Equal(t, "maven-password", cpe.custom.repositoryPassword)

		assert.Equal(t, "https://docker-repoUrl", cpe.container.registryURL)
		assert.Equal(t, "docker-user", cpe.container.repositoryUsername)
		assert.Equal(t, "docker-password", cpe.container.repositoryPassword)

		assert.Equal(t, "helm-repoUrl", cpe.custom.helmRepositoryURL)
		assert.Equal(t, "helm-user", cpe.custom.helmRepositoryUsername)
		assert.Equal(t, "helm-password", cpe.custom.helmRepositoryPassword)

		assert.Equal(t, "maven-repoUrl", cpe.custom.mavenRepositoryURL)
		assert.Equal(t, "maven", cpe.custom.repositoryID)
		assert.Equal(t, "maven-user", cpe.custom.mavenRepositoryUsername)
		assert.Equal(t, "maven-password", cpe.custom.mavenRepositoryPassword)

		assert.Equal(t, "npm-repoUrl", cpe.custom.npmRepositoryURL)
		assert.Equal(t, "npm-user", cpe.custom.npmRepositoryUsername)
		assert.Equal(t, "npm-password", cpe.custom.npmRepositoryPassword)

		assert.Equal(t, "pypi-repoUrl", cpe.custom.pipRepositoryURL)
		assert.Equal(t, "pypi-user", cpe.custom.pipRepositoryUsername)
		assert.Equal(t, "pypi-password", cpe.custom.pipRepositoryPassword)

		assert.Equal(t, "raw-repoUrl", cpe.custom.rawRepositoryURL)
		assert.Equal(t, "raw-user", cpe.custom.rawRepositoryUsername)
		assert.Equal(t, "raw-password", cpe.custom.rawRepositoryPassword)
	})

	t.Run("success case - CAP", func(t *testing.T) {
		t.Parallel()
		config := sapCallStagingServiceOptions{Action: "createRepositories", BuildTool: "CAP", RepositoryFormats: []string{"docker", "helm", "maven", "npm", "pypi", "raw"}}
		cpe := sapCallStagingServiceCommonPipelineEnvironment{}
		utilsMock := newStagingServiceUtilsMock()
		err := runSapCallStagingService(&config, nil, &cpe, utilsMock)
		assert.NoError(t, err)

		assert.Equal(t, "npm-repoUrl", cpe.custom.repositoryURL)
		assert.Equal(t, "npm", cpe.custom.repositoryFormat)
		assert.Equal(t, "npm-user", cpe.custom.repositoryUsername)
		assert.Equal(t, "npm-password", cpe.custom.repositoryPassword)
	})

	t.Run("error case - no stage config", func(t *testing.T) {
		t.Parallel()
		config := sapCallStagingServiceOptions{Action: "createRepositories"}
		cpe := sapCallStagingServiceCommonPipelineEnvironment{}
		utilsMock := newStagingServiceUtilsMock()
		utilsMock.stageConfigError = fmt.Errorf("config error")
		err := runSapCallStagingService(&config, nil, &cpe, utilsMock)
		assert.EqualError(t, err, "failed to get stage configuration: config error")
	})

	t.Run("error case - repo creation failed", func(t *testing.T) {
		t.Parallel()
		config := sapCallStagingServiceOptions{Action: "createRepositories", RepositoryFormats: []string{"maven"}}
		cpe := sapCallStagingServiceCommonPipelineEnvironment{}
		utilsMock := newStagingServiceUtilsMock()
		utilsMock.createRepoError = fmt.Errorf("failed to create repo")
		err := runSapCallStagingService(&config, nil, &cpe, utilsMock)
		assert.EqualError(t, err, "failed to create repo")
	})
}

func TestCreateRepositories(t *testing.T) {
	t.Parallel()

	t.Run("success case", func(t *testing.T) {
		t.Parallel()
		staging := stagingMock{group: "111", repositoryFormats: []string{"maven"}}
		result, err := createRepositories(&staging)
		assert.NoError(t, err)
		assert.Equal(t, map[string]map[string]interface{}{"maven": {"testKey": "testVal"}}, result)
	})

	t.Run("error case - no group", func(t *testing.T) {
		t.Parallel()
		staging := stagingMock{group: "", repositoryFormats: []string{"maven"}}
		_, err := createRepositories(&staging)
		assert.EqualError(t, err, "you must pass groupId and repositoryFormats")
	})

	t.Run("error case - no repositoryFormats", func(t *testing.T) {
		t.Parallel()
		staging := stagingMock{group: "111"}
		_, err := createRepositories(&staging)
		assert.EqualError(t, err, "you must pass groupId and repositoryFormats")
	})

	t.Run("error case - login error", func(t *testing.T) {
		t.Parallel()
		staging := stagingMock{group: "111", repositoryFormats: []string{"maven"}, loginError: fmt.Errorf("login error")}
		_, err := createRepositories(&staging)
		assert.EqualError(t, err, "login error")
	})

	t.Run("error case - repo creation failed", func(t *testing.T) {
		t.Parallel()
		staging := stagingMock{group: "111", repositoryFormats: []string{"maven"}, createMultipleReposError: fmt.Errorf("create repos error")}
		_, err := createRepositories(&staging)
		assert.EqualError(t, err, "create repos error")
	})
}

func TestCloseGroup(t *testing.T) {
	t.Parallel()

	t.Run("success case", func(t *testing.T) {
		t.Parallel()
		staging := &stagingMock{group: "111"}
		_, err := closeGroup(staging)
		assert.NoError(t, err)
	})

	t.Run("error case - no group", func(t *testing.T) {
		t.Parallel()
		staging := &stagingMock{}
		staging.groupIdFile = ""
		_, err := closeGroup(staging)
		assert.EqualError(t, err, "You must pass groupId")
	})

	t.Run("error case - read group id from file error", func(t *testing.T) {
		t.Parallel()
		staging := &stagingMock{group: "111"}
		staging.readGroupIdFromFileError = fmt.Errorf("read group id from file error")
		_, err := closeGroup(staging)
		assert.EqualError(t, err, "read group id from file error")
	})

	t.Run("error case - login and receive auth token error", func(t *testing.T) {
		t.Parallel()
		staging := &stagingMock{group: "111"}
		staging.loginError = fmt.Errorf("auth token error")
		_, err := closeGroup(staging)
		assert.EqualError(t, err, "auth token error")
	})

	t.Run("error case - close staging group error", func(t *testing.T) {
		t.Parallel()
		staging := &stagingMock{group: "111"}
		staging.closeStagingGroupError = fmt.Errorf("close group error")
		_, err := closeGroup(staging)
		assert.EqualError(t, err, "close group error")
	})
}

func TestGetStagedArtifactURLs(t *testing.T) {
	t.Parallel()

	t.Run("success case", func(t *testing.T) {
		t.Parallel()
		staging := &stagingMock{group: "111"}
		artifactURLs, err := getStagedArtifactURLs(staging)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"https://test.staging.service/stage/repository/16ec86818081/test-app",
			"https://test.staging.service/stage/repository/1d7a8cd48081/test.tgz",
		}, artifactURLs)
	})

	t.Run("error case - no group", func(t *testing.T) {
		t.Parallel()
		staging := &stagingMock{group: "", groupIdFile: ""}
		_, err := getStagedArtifactURLs(staging)
		assert.EqualError(t, err, "you must pass groupId")
	})

	t.Run("error case - ReadGroupIdFromFile", func(t *testing.T) {
		t.Parallel()
		staging := &stagingMock{group: "111"}
		staging.readGroupIdFromFileError = fmt.Errorf("read group id from file error")
		_, err := getStagedArtifactURLs(staging)
		assert.EqualError(t, err, "read group id from file error")
	})

	t.Run("error case - LoginAndReceiveAuthToken", func(t *testing.T) {
		t.Parallel()
		staging := &stagingMock{group: "111"}
		staging.loginError = fmt.Errorf("login error")
		_, err := getStagedArtifactURLs(staging)
		assert.EqualError(t, err, "login error")
	})

	t.Run("error case - GetStagedArtifactURLs", func(t *testing.T) {
		t.Parallel()
		staging := &stagingMock{group: "111"}
		staging.getStagedArtifactsError = fmt.Errorf("get staged artifact urls error")
		_, err := getStagedArtifactURLs(staging)
		assert.EqualError(t, err, "get staged artifact urls error")
	})
}

func TestRunSapCallStagingService(t *testing.T) {
	t.Parallel()

	t.Run("create group", func(t *testing.T) {

		staging := &stagingMock{}
		err := createGroup(staging)
		assert.NoError(t, err)
		config := sapCallStagingServiceOptions{Action: "createGroup"}
		utilsMock := newStagingServiceUtilsMock()
		err = runSapCallStagingService(&config, nil, &sapCallStagingServiceCommonPipelineEnvironment{}, utilsMock)
		assert.EqualError(t, err, "You must pass profile")
	})

	t.Run("create repository", func(t *testing.T) {

		staging := &stagingMock{}
		_, err := createRepository(staging)
		assert.NoError(t, err)
		config := sapCallStagingServiceOptions{Action: "createRepository", GroupID: "dummy1", GroupIDFile: "dummy"}
		utilsMock := newStagingServiceUtilsMock()
		err = runSapCallStagingService(&config, nil, &sapCallStagingServiceCommonPipelineEnvironment{}, utilsMock)
		assert.EqualError(t, err, "you must pass groupId and buildTool / repositoryFormat")

	})

	t.Run("close (group) action - success case", func(t *testing.T) {
		expectedStagedArtifactsURLs := []string{
			"https://test.staging.service/stage/repository/16ec86818081/test-app",
			"https://test.staging.service/stage/repository/1d7a8cd48081/test.tgz",
		}

		config := sapCallStagingServiceOptions{Action: "close", GroupID: "dummy1", GroupIDFile: "dummy"}
		cpe := sapCallStagingServiceCommonPipelineEnvironment{}
		utilsMock := newStagingServiceUtilsMock()

		err := runSapCallStagingService(&config, nil, &cpe, utilsMock)
		assert.NoError(t, err)
		assert.Equal(t, expectedStagedArtifactsURLs, cpe.custom.stagedArtifactURLs)
	})

	t.Run("close (group) action - closeGroup returns an error - missed groupId", func(t *testing.T) {
		config := sapCallStagingServiceOptions{Action: "close"}
		utilsMock := newStagingServiceUtilsMock()
		utilsMock.closeGroupError = fmt.Errorf("You must pass groupId")

		err := runSapCallStagingService(&config, nil, &sapCallStagingServiceCommonPipelineEnvironment{}, utilsMock)
		assert.EqualError(t, err, "You must pass groupId")
	})

	t.Run("close (group) action - getStagedArtifactURLs returns an error", func(t *testing.T) {
		config := sapCallStagingServiceOptions{Action: "close", GroupID: "dummy1", GroupIDFile: "dummy"}
		utilsMock := newStagingServiceUtilsMock()
		utilsMock.getStagedArtifactURLsError = fmt.Errorf("some error happened in getStagedArtifactURLs")

		err := runSapCallStagingService(&config, nil, &sapCallStagingServiceCommonPipelineEnvironment{}, utilsMock)
		assert.EqualError(t, err, "some error happened in getStagedArtifactURLs")
	})

	t.Run("update metadata group", func(t *testing.T) {

		staging := &stagingMock{}
		err := updateGroupMetadata(staging)
		assert.NoError(t, err)
		config := sapCallStagingServiceOptions{Action: "updateGroupMetadata"}
		utilsMock := newStagingServiceUtilsMock()
		err = runSapCallStagingService(&config, nil, &sapCallStagingServiceCommonPipelineEnvironment{}, utilsMock)
		assert.EqualError(t, err, "You must pass groupId and metadata")

	})
	t.Run("promote group", func(t *testing.T) {

		staging := &stagingMock{}
		promotedArtifacts, err := promoteGroup(staging, "")
		assert.NoError(t, err)
		config := sapCallStagingServiceOptions{Action: "promote"}
		utilsMock := newStagingServiceUtilsMock()
		err = runSapCallStagingService(&config, nil, &sapCallStagingServiceCommonPipelineEnvironment{}, utilsMock)
		assert.EqualError(t, err, "you must pass groupId and outputFile")
		assert.Equal(t, "https://common.repo.test/1/test-app", promotedArtifacts.PromotedArtifactURLs[0])
		assert.Equal(t, "docker.common.repo.test/test-image:1.22.3", promotedArtifacts.PromotedDockerImages[0])
		assert.Equal(t, "https://common.repo.test/helm/test-0.1.0.tgz", promotedArtifacts.PromotedHelmChartURL)

	})
	t.Run("sign group", func(t *testing.T) {

		staging := &stagingMock{}
		err := signGroup(staging)
		assert.NoError(t, err)
		config := sapCallStagingServiceOptions{Action: "signGroup"}
		utilsMock := newStagingServiceUtilsMock()
		err = runSapCallStagingService(&config, nil, &sapCallStagingServiceCommonPipelineEnvironment{}, utilsMock)
		assert.EqualError(t, err, "You must pass groupId")

	})
	t.Run("get repo credentials", func(t *testing.T) {

		staging := &stagingMock{}
		err := getRepositoryCredentials(staging)
		assert.NoError(t, err)
		config := sapCallStagingServiceOptions{Action: "getRepositoryCredentials"}
		utilsMock := newStagingServiceUtilsMock()
		err = runSapCallStagingService(&config, nil, &sapCallStagingServiceCommonPipelineEnvironment{}, utilsMock)
		assert.EqualError(t, err, "You must pass groupId, outputFile and repositoryId")

	})
	t.Run("get repo bom", func(t *testing.T) {

		staging := &stagingMock{}
		err := getRepositoryBom(staging)
		assert.NoError(t, err)
		config := sapCallStagingServiceOptions{Action: "getRepositoryBom"}
		utilsMock := newStagingServiceUtilsMock()
		err = runSapCallStagingService(&config, nil, &sapCallStagingServiceCommonPipelineEnvironment{}, utilsMock)
		assert.EqualError(t, err, "You must pass groupId, outputFile and repositoryId")

	})
	t.Run("get group bom", func(t *testing.T) {

		staging := &stagingMock{}
		err := getGroupBom(staging)
		assert.NoError(t, err)
		config := sapCallStagingServiceOptions{Action: "getGroupBom"}
		utilsMock := newStagingServiceUtilsMock()
		err = runSapCallStagingService(&config, nil, &sapCallStagingServiceCommonPipelineEnvironment{}, utilsMock)
		assert.EqualError(t, err, "you must pass groupId and outputFile")

	})
	t.Run("get group metadata", func(t *testing.T) {

		staging := &stagingMock{}
		err := getGroupMetadata(staging)
		assert.NoError(t, err)
		config := sapCallStagingServiceOptions{Action: "getGroupMetadata"}
		utilsMock := newStagingServiceUtilsMock()
		err = runSapCallStagingService(&config, nil, &sapCallStagingServiceCommonPipelineEnvironment{}, utilsMock)
		assert.EqualError(t, err, "you must pass groupId and outputFile")

	})
	t.Run("set repository metadata", func(t *testing.T) {

		staging := &stagingMock{}
		err := updateRepositoryMetadata(staging)
		assert.NoError(t, err)
		config := sapCallStagingServiceOptions{Action: "updateRepositoryMetadata"}
		utilsMock := newStagingServiceUtilsMock()
		err = runSapCallStagingService(&config, nil, &sapCallStagingServiceCommonPipelineEnvironment{}, utilsMock)
		assert.EqualError(t, err, "You must pass groupId, metadata and repositoryId")

	})
	t.Run("get repository metadata", func(t *testing.T) {

		staging := &stagingMock{}
		err := getRepositoryMetadata(staging)
		assert.NoError(t, err)
		config := sapCallStagingServiceOptions{Action: "getRepositoryMetadata"}
		utilsMock := newStagingServiceUtilsMock()
		err = runSapCallStagingService(&config, nil, &sapCallStagingServiceCommonPipelineEnvironment{}, utilsMock)
		assert.EqualError(t, err, "You must pass groupId, outputFile and repositoryId")

	})
	t.Run("search metadata", func(t *testing.T) {

		staging := &stagingMock{}
		err := searchGroupMetadata(staging)
		assert.NoError(t, err)
		config := sapCallStagingServiceOptions{Action: "searchGroupMetadata"}
		utilsMock := newStagingServiceUtilsMock()
		err = runSapCallStagingService(&config, nil, &sapCallStagingServiceCommonPipelineEnvironment{}, utilsMock)
		assert.EqualError(t, err, "You must pass groupId, outputFile and query")

	})
	t.Run("error path", func(t *testing.T) {

		config := sapCallStagingServiceOptions{}
		utilsMock := newStagingServiceUtilsMock()
		err := runSapCallStagingService(&config, nil, &sapCallStagingServiceCommonPipelineEnvironment{}, utilsMock)

		assert.EqualError(t, err, "Wrong action")

	})
}

//go:build unit
// +build unit

package staging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

type stagingMockClient struct {
	username       string
	password       string
	httpMethod     string
	httpStatusCode int
	urlCalled      string
	requestBody    io.Reader
	responseBody   string
}

func (c *stagingMockClient) SetOptions(opts piperhttp.ClientOptions) {
	c.username = opts.Username
	c.password = opts.Password
}

func (c *stagingMockClient) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	c.httpMethod = method
	c.urlCalled = url
	if method == "POST" && url == "https://test.staging.service/stage/api/group/create" {
		respContent := `{"group":"group-1313123-test"}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(respContent)))}, nil
	}

	if method == "POST" && url == "https://test.staging.service/stage/api/repository/create" {
		respContent := `{"password":"test-pass","repository":"test-repo","user":"testCreateRepo","repositoryURL":"test-repo.staging.repositories.cloud.sap"}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(respContent)))}, nil
	}

	if method == "POST" && url == "https://test.staging.service/stage/api/login" {
		respContent := `{"access_token":"test-token","token_type":"bearer",  "group":"group-test"}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(respContent)))}, nil
	}

	if method == "POST" && url == "https://test.staging.service/stage/api/group/close" {
		respContent := `{"repositories":null,"closed":true,"group":"group-test-1321313"}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(respContent)))}, nil
	}
	if method == "GET" && url == "https://test.staging.service/stage/api/group/metadata" {
		respContent := `{"metadata":"test metadata","repositories":"","group":"group-test"}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(respContent)))}, nil
	}
	if method == "GET" && url == "https://test.staging.service/stage/api/repository/metadata/repo-test" {
		respContent := `{"metadata":"test metadata"}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(respContent)))}, nil
	}
	if method == "GET" && url == "https://test.staging.service/stage/api/repository/credentials/repo-test" {
		respContent := `{"password":"test","repository":"test-repo","user":"test"}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(respContent)))}, nil
	}
	if method == "GET" && url == "https://test.staging.service/stage/api/repository/BOM/repo-test" {
		respContent := `{"components":[{"artifact":"test","image":"aba3d5328081-20201113-144422322-511.test.sap/test:1605278662639","version":"1605278662639"}],"format":"docker"}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(respContent)))}, nil
	}
	if method == "GET" && url == "https://test.staging.service/stage/api/group/BOM" {
		respContent := `{
			"repositories":[
				{
					"BOM":{
						"components":[{"artifact":"test","image":"test","version":"1.0.0"}],
						"format":"docker"
					},
					"repository":"test"
				},{
					"BOM":{
						"components":[
							{
								"artifact":"test",
								"assets":[{"fileName":"test-app","relativePath":"test","url":"https://test.staging.service/stage/repository/16ec86818081/test-app"}],
								"version":null,
								"group":"test"
							}
						],
						"format":"raw"
					},
					"repository":"16ec86818081"
				},{
					"BOM":{
						"components":[
							{
								"artifact":"test",
								"assets":[{"fileName":"test","relativePath":"test","url":"https://test.staging.service/stage/repository/1d7a8cd48081/test.tgz"}],
								"version":"0.1.0","group":null
							}
						],
						"format":"helm"
					},
					"repository":"1d7a8cd48081"
				}
			],
			"group":"group-test"
		}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(respContent)))}, nil
	}
	if method == "GET" && url == "https://test.staging.service/stage/api/group/metadata/search?q=test" {
		respContent := `{"components":[{"artifact":"test","image":"aba3d5328081-20201113-144422322-511.test.sap/test:1605278662639","version":"1605278662639"}],"format":"docker"}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(respContent)))}, nil
	}
	if method == "POST" && url == "https://test.staging.service/stage/api/group/promote/async" {
		respContent := `{"responseFromPromote":{"repositories":[{"success":true,"repository":"test","result":["http://nexus.wdf.sap.corp:9091/nexus/content/repositories/deploy.milestones/com/sap/test/test/10.5/test-10.5.jar"]}],"released":true,"group":"group-test"},"state":"released"}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(respContent)))}, nil
	}
	if method == "POST" && url == "https://test.staging.service/stage/api/signing/group" {
		respContent := `{"result":"SUCCESSFUL","description":"Signing is triggered successfully"}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(respContent)))}, nil
	}
	if method == "GET" && url == "https://test.staging.service/stage/api/group/state" {
		respContent := `{
			"responseFromPromote":{
				"repositories":[
					{"success":true,"repository":"test","result":["http://nexus.wdf.sap.corp:9091/nexus/content/repositories/deploy.milestones/com/sap/test/test/10.5/test-10.5.jar"]},
					{"success":true,"repository":"1d7a8cd48081","result": ["https://common.repo.test/helm/test-0.1.0.tgz"]},
					{
						"success":true,
						"repository":"6a6b65788081",
						"list":[
							{
								"artifact":"test-artifact",
								"image":"docker.common.repo.test/test-image:1.22.3",
								"success":true,
								"version":"1.22.3"
							}
						]
					}
				],
				"released":true,
				"group":"group-test"
			},
			"state":"released"
		}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(respContent)))}, nil
	}
	return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte("")))}, nil
}

var files map[string][]byte

func writeFileMock(filename string, data []byte, perm os.FileMode) error {
	if files == nil {
		files = make(map[string][]byte)
	}
	files[filename] = data
	return nil
}
func TestStagingService(t *testing.T) {
	mockClient := stagingMockClient{}
	sapStaging := Staging{
		TenantId:     "test",
		TenantSecret: "test",
		Username:     "test",
		Password:     "test",
		Profile:      "test",
		BuildTool:    "docker",
		Url:          "https://test.staging.service/stage/api",
		OutputFile:   "test.json",
		RepositoryId: "repo-test",
		GroupIdFile:  "test.json",
		HTTPClient:   &mockClient,
		Query:        "test",
	}
	t.Run("Test Create Group", func(t *testing.T) {
		resp, err := sapStaging.CreateStagingGroup()
		assert.Equal(t, nil, err)
		assert.Equal(t, resp, "group-1313123-test")
		result, errJson := sapStaging.ReadGroupIdFromFile()
		assert.Equal(t, errJson, nil)
		assert.Equal(t, result, "group-1313123-test")
		os.Remove("test.json")
	})
	t.Run("Test Create Repository", func(t *testing.T) {
		resp, err := sapStaging.CreateStagingRepository()
		assert.Equal(t, nil, err)
		assert.Equal(t, resp, `{"password":"test-pass","repository":"test-repo","user":"testCreateRepo","repositoryURL":"test-repo.staging.repositories.cloud.sap"}`)
		result, errJson := readFile("test.json")
		assert.Equal(t, errJson, nil)
		assert.Equal(t, "test-repo", result["repository"])
		assert.Equal(t, "test-pass", result["password"])
		assert.Equal(t, "testCreateRepo", result["user"])
		assert.Equal(t, "test-repo.staging.repositories.cloud.sap", result["repositoryURL"])
		os.Remove("test.json")
	})
	t.Run("Test Create Repository - buildTool is CAP", func(t *testing.T) {
		sapStaging := Staging{
			BuildTool: "CAP",
		}
		resp, err := sapStaging.CreateStagingRepository()
		assert.Equal(t, "", resp)
		assert.EqualError(t, err, "wrong repo format 'maven,npm', please use createRepositories action for build tool CAP")
	})
	t.Run("Test Login", func(t *testing.T) {
		resp, err := sapStaging.LoginAndReceiveAuthToken()
		assert.Equal(t, nil, err)
		assert.Equal(t, "test-token", resp)
	})
	t.Run("Test Group Close", func(t *testing.T) {
		resp, err := sapStaging.CloseStagingGroup()
		assert.Equal(t, nil, err)
		expected := map[string]interface{}{
			"repositories": nil,
			"closed":       true,
			"group":        "group-test-1321313",
		}
		assert.Equal(t, expected, resp)
	})
	t.Run("Test Group Update Metadata", func(t *testing.T) {
		err := sapStaging.SetGroupMetadata()
		assert.Equal(t, nil, err)
	})
	t.Run("Test Group Get Metadata", func(t *testing.T) {
		err := sapStaging.GetGroupMetadata()
		assert.Equal(t, nil, err)
		result, errJson := readFile("test.json")
		assert.Equal(t, errJson, nil)
		assert.Equal(t, "group-test", result["group"])
		assert.Equal(t, "test metadata", result["metadata"])
		assert.Equal(t, "", result["repositories"])
		os.Remove("test.json")
	})
	t.Run("Test Repository Set Metadata", func(t *testing.T) {
		err := sapStaging.SetRepositoryMetadata()
		assert.Equal(t, nil, err)
	})
	t.Run("Test Repository Get Metadata", func(t *testing.T) {
		err := sapStaging.GetRepositoryMetadata()
		assert.Equal(t, nil, err)
		result, errJson := readFile("test.json")
		assert.Equal(t, errJson, nil)
		assert.Equal(t, "test metadata", result["metadata"])
		os.Remove("test.json")
	})
	t.Run("Test Get Repository Credentials", func(t *testing.T) {
		_, err := sapStaging.GetRepositoryCredentials()
		assert.Equal(t, nil, err)
		result, errJson := readFile("test.json")
		assert.Equal(t, errJson, nil)
		assert.Equal(t, "test", result["user"])
		assert.Equal(t, "test", result["password"])
		assert.Equal(t, "test-repo", result["repository"])
		os.Remove("test.json")
	})
	t.Run("Test Get Repository Bom", func(t *testing.T) {
		_, err := sapStaging.GetRepositoryBom()
		assert.Equal(t, nil, err)
		result, errJson := readFile("test.json")
		assert.Equal(t, errJson, nil)
		assert.Equal(t, "docker", result["format"])
		os.Remove("test.json")
	})
	t.Run("Test Get Staged Artifact URLs - success case", func(t *testing.T) {
		urls, err := sapStaging.GetStagedArtifactURLs()
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"https://test.staging.service/stage/repository/16ec86818081/test-app",
			"https://test.staging.service/stage/repository/1d7a8cd48081/test.tgz",
		}, urls)
	})
	t.Run("Test Get Group Bom", func(t *testing.T) {
		_, err := sapStaging.GetGroupBom()
		assert.Equal(t, nil, err)
		result, errJson := readFile("test.json")
		assert.Equal(t, errJson, nil)
		assert.Equal(t, "group-test", result["group"])
		os.Remove("test.json")
	})
	t.Run("Test Promote", func(t *testing.T) {
		promotedArtifacts, err := sapStaging.PromoteGroup("")
		assert.Equal(t, nil, err)
		result, errJson := readFile("test.json")
		assert.Equal(t, errJson, nil)
		assert.Equal(t, "released", result["state"])
		assert.Equal(t, "https://common.repo.test/helm/test-0.1.0.tgz", promotedArtifacts.PromotedHelmChartURL)
		assert.Equal(t, []string{"http://nexus.wdf.sap.corp:9091/nexus/content/repositories/deploy.milestones/com/sap/test/test/10.5/test-10.5.jar"}, promotedArtifacts.PromotedArtifactURLs)
		assert.Equal(t, []string{"docker.common.repo.test/test-image:1.22.3"}, promotedArtifacts.PromotedDockerImages)
		os.Remove("test.json")
	})
	t.Run("Test Search Metadata", func(t *testing.T) {
		_, err := sapStaging.SearchMetadataGroup()
		assert.Equal(t, nil, err)
		_, errJson := readFile("test.json")
		assert.Equal(t, errJson, nil)
		//assert.Equal(t,"docker",result["format"])
		os.Remove("test.json")
	})
	t.Run("Test Signing", func(t *testing.T) {
		resp, err := sapStaging.SignGroup()
		var jsonMap map[string]interface{}
		json.Unmarshal([]byte(resp), &jsonMap)
		assert.Equal(t, nil, err)
		assert.Equal(t, "SUCCESSFUL", jsonMap["result"])
	})
	t.Run("Test handleHelmPromotedUrl", func(t *testing.T) {
		tt := []struct {
			testCaseName         string
			promotedRepos        []Repository
			promotedArtifacts    *PromotedArtifacts
			expectedHelmChartURL string
		}{
			{
				testCaseName: "helm url exists",
				promotedRepos: []Repository{
					{
						Result:     []string{"https://common.repo.test/go/test-app"},
						Success:    true,
						Repository: "16ec86818081",
					},
					{
						Result:     []string{"https://common.repo.test/helm/test-0.1.0.tgz"},
						Success:    true,
						Repository: "1d7a8cd48081",
					},
				},
				promotedArtifacts: &PromotedArtifacts{
					PromotedArtifactURLs: []string{
						"https://common.repo.test/go/test-app",
						"https://common.repo.test/helm/test-0.1.0.tgz",
					},
				},
				expectedHelmChartURL: "https://common.repo.test/helm/test-0.1.0.tgz",
			},
			{
				testCaseName: "helm url doesn't exist",
				promotedRepos: []Repository{
					{
						Result:     []string{"https://common.repo.test/go/test-app-1"},
						Success:    true,
						Repository: "16ec86818081",
					},
					{
						Result:     []string{"https://common.repo.test/go/test-app-2"},
						Success:    true,
						Repository: "16ec86818082",
					},
				},
				promotedArtifacts: &PromotedArtifacts{
					PromotedArtifactURLs: []string{
						"https://common.repo.test/go/test-app-1",
						"https://common.repo.test/go/test-app-2",
					},
				},
				expectedHelmChartURL: "",
			},
		}

		for _, testCase := range tt {
			t.Run(testCase.testCaseName, func(t *testing.T) {
				err := sapStaging.handleHelmPromotedUrl(testCase.promotedRepos, testCase.promotedArtifacts)
				assert.Equal(t, testCase.expectedHelmChartURL, testCase.promotedArtifacts.PromotedHelmChartURL)
				assert.NoError(t, err)
			})
		}
	})
}

func readFile(nameOfFile string) (map[string]interface{}, error) {
	_, err := os.Stat("test.json")
	if err != nil {
		return nil, err
	}
	jsonFile, errOpen := os.Open("test.json")
	if err != nil {
		return nil, errOpen
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	result, errJson := byteSliceToJSONMap([]byte(byteValue))
	if err != nil {
		return nil, errJson
	}
	return result, nil
}

func TestRemoveURLFromPromotedList(t *testing.T) {
	tt := []struct {
		testCaseName            string
		promotedArtifactURLs    []string
		url                     string
		expectedPromotedURLsLen int
		expectedErr             error
	}{
		{
			testCaseName: "helm url exists",
			promotedArtifactURLs: []string{
				"https://common.repo.test/go/test-app",
				"https://common.repo.test/helm/test-0.1.0.tgz",
			},
			url:                     "https://common.repo.test/helm/test-0.1.0.tgz",
			expectedPromotedURLsLen: 1,
			expectedErr:             nil,
		},
		{
			testCaseName: "helm url exists (urls for go artifacts + helm url at last item of the slice)",
			promotedArtifactURLs: []string{
				"https://common.repo.test/go/1/test-app",
				"https://common.repo.test/go/2/test-app",
				"https://common.repo.test/helm/test-0.1.0.tgz",
			},
			url:                     "https://common.repo.test/helm/test-0.1.0.tgz",
			expectedPromotedURLsLen: 2,
			expectedErr:             nil,
		},
		{
			testCaseName: "helm url exists (helm url at first item of the slice)",
			promotedArtifactURLs: []string{
				"https://common.repo.test/helm/test-0.1.0.tgz",
				"https://common.repo.test/go/1/test-app",
				"https://common.repo.test/go/2/test-app",
			},
			url:                     "https://common.repo.test/helm/test-0.1.0.tgz",
			expectedPromotedURLsLen: 2,
			expectedErr:             nil,
		},
		{
			testCaseName: "helm url exists (helm url at the middle of the slice)",
			promotedArtifactURLs: []string{
				"https://common.repo.test/go/1/test-app",
				"https://common.repo.test/helm/test-0.1.0.tgz",
				"https://common.repo.test/go/2/test-app",
			},
			url:                     "https://common.repo.test/helm/test-0.1.0.tgz",
			expectedPromotedURLsLen: 2,
			expectedErr:             nil,
		},
		{
			testCaseName: "helm url exists (only url for helm artifact)",
			promotedArtifactURLs: []string{
				"https://common.repo.test/helm/test-0.1.0.tgz",
			},
			url:                     "https://common.repo.test/helm/test-0.1.0.tgz",
			expectedPromotedURLsLen: 0,
			expectedErr:             nil,
		},
		{
			testCaseName: "helm url doesn't exist",
			promotedArtifactURLs: []string{
				"https://common.repo.test/go/test-app",
			},
			url:                     "https://common.repo.test/helm/test-0.1.0.tgz",
			expectedPromotedURLsLen: 1,
			expectedErr:             fmt.Errorf("failed to delete url: https://common.repo.test/helm/test-0.1.0.tgz not found"),
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.testCaseName, func(t *testing.T) {
			promotedArtifactURLs, err := removeURLFromPromotedList(testCase.promotedArtifactURLs, testCase.url)
			if testCase.expectedErr == nil {
				assert.NoError(t, err)
				assert.Len(t, promotedArtifactURLs, testCase.expectedPromotedURLsLen)
			} else {
				assert.EqualError(t, err, testCase.expectedErr.Error())
				assert.Len(t, promotedArtifactURLs, 0)
			}
			assert.NotContains(t, "https://common.repo.test/helm/test-0.1.0.tg", promotedArtifactURLs)
		})
	}
}

func TestIdentifyCorrectStagingRepoFormat(t *testing.T) {
	var s Staging
	tt := []struct {
		buildTool, expectedFormat string
	}{
		{"maven", "maven"},
		{"golang", "raw"},
		{"pip", "pypi"},
		{"CAP", "maven,npm"},
		{"npm", "npm"},
	}

	for _, test := range tt {
		s.BuildTool = test.buildTool
		format := s.IdentifyCorrectStagingRepoFormat()
		if format != test.expectedFormat {
			t.Errorf("Expected %s, got %s", test.expectedFormat, format)
		}
	}
}

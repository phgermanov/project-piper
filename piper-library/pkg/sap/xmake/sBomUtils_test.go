//go:build unit
// +build unit

package xmake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestFetchSBomXML(t *testing.T) {
	// Set up a test server that returns a response with HTTP status 200 and a body containing "test"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<sbom></sbom>"))
	}))
	defer ts.Close()

	// Test case for a successful request
	expectedBody := []byte("<sbom></sbom>")
	actualBody, err := fetchSBomXML(ts.URL, "username", "password")
	if err != nil {
		t.Errorf("fetchSBomXML failed with error: %v", err)
	}
	if !bytes.Equal(actualBody, expectedBody) {
		t.Errorf("fetchSBomXML returned unexpected body: expected %v, got %v", expectedBody, actualBody)
	}

	// Test case for a request that returns a non-2xx HTTP status code
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	_, err = fetchSBomXML(ts.URL, "username", "password")
	if err == nil {
		t.Errorf("fetchSBomXML should have returned an error for non-2xx HTTP status code")
	}
}

func TestGetImagesAndCredentialsFromStageBOM(t *testing.T) {
	filesMock := mock.FilesMock{}

	filesMock.AddFile("sbom_one_image.json", []byte(`{
		"363001228081-20240215-175657891-598": {
			"components": [
				{
					"artifact": "com.sap.docker/ppiper",
					"image": "363001228081-20240215-175657891-598.intstaging.repositories.cloud.sap/com.sap.docker/ppiper:240104-20240215175557",
					"version": "240104-20240215175557"
				}
			],
			"credentials": {
				"password": "password123",
				"repository": "363001228081-20240215-175657891-598",
				"repositoryURL": "363001228081-20240215-175657891-598.intstaging.repositories.cloud.sap",
				"user": "user123"
			},
			"format": "docker"
		}
	}`))

	filesMock.AddFile("sbom_two_images.json", []byte(`{
		"363001228081-20240215-175657891-598": {
			"components": [
				{
					"artifact": "com.sap.docker/ppiper",
					"image": "363001228081-20240215-175657891-598.intstaging.repositories.cloud.sap/com.sap.docker/ppiper:240104-20240215175557",
					"version": "240104-20240215175557"
				},
				{
					"artifact": "com.sap.docker/ppiper2",
					"image": "363001228081-20240215-175657891-598.intstaging.repositories.cloud.sap/com.sap.docker/ppiper2:240104-20240215175557",
					"version": "240104-20240215175557"
				}
			],
			"credentials": {
				"password": "password123",
				"repository": "363001228081-20240215-175657891-598",
				"repositoryURL": "363001228081-20240215-175657891-598.intstaging.repositories.cloud.sap",
				"user": "user123"
			},
			"format": "docker"
		}
	}`))

	filesMock.AddFile("sbom_multiple_arches.json", []byte(`{
		"363001228081-20240215-175657891-598": {
			"components": [
				{
					"artifact": "com.sap.docker/ppiper",
					"image": "363001228081-20240215-175657891-598.intstaging.repositories.cloud.sap/com.sap.docker/ppiper:240104-20240215175557",
					"version": "240104-20240215175557"
				},
				{
					"artifact": "com.sap.docker/ppiper",
					"image": "363001228081-20240215-175657891-598.intstaging.repositories.cloud.sap/com.sap.docker/ppiper2:240104-20240215175557-amd64",
					"version": "240104-20240215175557"
				},
				{
					"artifact": "com.sap.docker/ppiper",
					"image": "363001228081-20240215-175657891-598.intstaging.repositories.cloud.sap/com.sap.docker/ppiper2:240104-20240215175557-aarch64",
					"version": "240104-20240215175557"
				}
			],
			"credentials": {
				"password": "password123",
				"repository": "363001228081-20240215-175657891-598",
				"repositoryURL": "363001228081-20240215-175657891-598.intstaging.repositories.cloud.sap",
				"user": "user123"
			},
			"format": "docker"
		}
	}`))

	filesMock.AddFile("sbom_without_images.json", []byte(`{
		"363000158081-20240215-17581313-294": {
			"components": [
				{
					"artifact": "ppiper",
					"assets": [
						{
							"extension": "pom",
							"fileName": "ppiper-240104-20240215175557.pom",
							"relativePath": "com/sap/docker/ppiper/240104-20240215175557/ppiper-240104-20240215175557.pom",
							"url": "https://intstaging.repositories.cloud.sap/stage/repository/363000158081-20240215-17581313-294/com/sap/docker/ppiper/240104-20240215175557/ppiper-240104-20240215175557.pom"
						}
					],
					"group": "com.sap.docker",
					"version": "240104-20240215175557"
				}
			],
			"credentials": {
				"password": "password456",
				"repository": "363000158081-20240215-17581313-294",
				"repositoryURL": "https://intstaging.repositories.cloud.sap/stage/repository/363000158081-20240215-17581313-294/",
				"user": "user456"
			},
			"format": "maven2"
		}
	}`))

	tt := []struct {
		name                          string
		jsonFilePath                  string
		expectedImageNames            []string
		expectedRepositoryCredentials repositoryCredentials
		expectedImageNameTags         []string
		expectedError                 error
	}{
		{
			name:                          "Success - one Docker image",
			jsonFilePath:                  "sbom_one_image.json",
			expectedImageNames:            []string{"com.sap.docker/ppiper"},
			expectedRepositoryCredentials: repositoryCredentials{Password: "password123", Repository: "363001228081-20240215-175657891-598", RepositoryURL: "363001228081-20240215-175657891-598.intstaging.repositories.cloud.sap", User: "user123"},
			expectedImageNameTags:         []string{"com.sap.docker/ppiper:240104-20240215175557"},
			expectedError:                 nil,
		},
		{
			name:                          "Success - two Docker images",
			jsonFilePath:                  "sbom_two_images.json",
			expectedImageNames:            []string{"com.sap.docker/ppiper", "com.sap.docker/ppiper2"},
			expectedRepositoryCredentials: repositoryCredentials{Password: "password123", Repository: "363001228081-20240215-175657891-598", RepositoryURL: "363001228081-20240215-175657891-598.intstaging.repositories.cloud.sap", User: "user123"},
			expectedImageNameTags:         []string{"com.sap.docker/ppiper:240104-20240215175557", "com.sap.docker/ppiper2:240104-20240215175557"},
			expectedError:                 nil,
		},
		{
			name:                          "Success - one image with multiple arches",
			jsonFilePath:                  "sbom_multiple_arches.json",
			expectedImageNames:            []string{"com.sap.docker/ppiper"},
			expectedRepositoryCredentials: repositoryCredentials{Password: "password123", Repository: "363001228081-20240215-175657891-598", RepositoryURL: "363001228081-20240215-175657891-598.intstaging.repositories.cloud.sap", User: "user123"},
			expectedImageNameTags:         []string{"com.sap.docker/ppiper:240104-20240215175557", "com.sap.docker/ppiper2:240104-20240215175557-amd64", "com.sap.docker/ppiper2:240104-20240215175557-aarch64"},
			expectedError:                 nil,
		},
		{
			name:                          "Fail - no Docker images",
			jsonFilePath:                  "sbom_without_images.json",
			expectedImageNames:            []string{},
			expectedRepositoryCredentials: repositoryCredentials{},
			expectedImageNameTags:         []string{},
			expectedError:                 fmt.Errorf("No images found in sbom"),
		},
	}

	for _, test := range tt {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// init
			jsonContent, err := filesMock.ReadFile(test.jsonFilePath)
			if err != nil {
				t.Errorf("Couldn't parse stageBOM file %s: %s", test.jsonFilePath, err)
			}
			var sbom map[string]interface{}
			err = json.Unmarshal(jsonContent, &sbom)
			if err != nil {
				t.Errorf("Couldn't parse stageBOM file %s: %s", test.jsonFilePath, err)
			}
			// test
			imageNames, imageNameTags, credentials, err := GetImagesAndCredentialsFromStageBOM(&sbom)
			// asserts
			if err != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedImageNames, imageNames)
				assert.Equal(t, test.expectedRepositoryCredentials, credentials)
				assert.Equal(t, test.expectedImageNameTags, imageNameTags)
			}
		})
	}
}

func TestGetArtifactURLs(t *testing.T) {
	tests := []struct {
		name           string
		stageRepo      stageRepository
		expectedResult []string
	}{
		{
			name: "Multiple components with URLs",
			stageRepo: stageRepository{
				Components: []*component{
					{
						Assets: []*asset{
							{FileName: "sbom1.xml", URL: "https://repo.example.com/sbom1.xml"},
						},
					},
					{
						Assets: []*asset{
							{FileName: "sbom2.json", URL: "https://repo.example.com/sbom2.json"},
						},
					},
				},
				Credentials: &repositoryCredentials{Repository: "repo1"},
			},
			expectedResult: []string{
				"https://repo.example.com/sbom1.xml",
				"https://repo.example.com/sbom2.json",
			},
		},
		{
			name: "No components, returns default",
			stageRepo: stageRepository{
				Components:  []*component{},
				Credentials: &repositoryCredentials{Repository: "repo2"},
			},
			expectedResult: []string{"repo2sbom"},
		},
		{
			name: "Component with empty URL is skipped",
			stageRepo: stageRepository{
				Components: []*component{
					{
						Assets: []*asset{
							{FileName: "sbom1.json", URL: ""},
						},
					},
					{
						Assets: []*asset{
							{FileName: "sbom2.json", URL: "https://repo.example.com/sbom2.json"},
						},
					},
				},
				Credentials: &repositoryCredentials{Repository: "repo3"},
			},
			expectedResult: []string{"https://repo.example.com/sbom2.json"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			result := GetArtifactURLs(tt.stageRepo)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestGetArtifactNames(t *testing.T) {
	tests := []struct {
		name           string
		stageRepo      stageRepository
		expectedResult []string
	}{
		{
			name: "Multiple components with file names",
			stageRepo: stageRepository{
				Components: []*component{
					{
						Assets: []*asset{
							{FileName: "ppiper-240104-20240215175557.json"},
						},
					},
					{
						Assets: []*asset{
							{FileName: "ppiper-240104-20240215175557-amd.json"},
						},
					},
				},
				Credentials: &repositoryCredentials{},
			},
			expectedResult: []string{
				"ppiper-240104-20240215175557.json",
				"ppiper-240104-20240215175557-amd.json",
			},
		},
		{
			name: "No components, returns default",
			stageRepo: stageRepository{
				Components:  []*component{},
				Credentials: &repositoryCredentials{},
			},
			expectedResult: []string{"sbom"},
		},
		{
			name: "Component with empty file name is skipped",
			stageRepo: stageRepository{
				Components: []*component{
					{
						Assets: []*asset{
							{FileName: ""},
						},
					},
					{
						Assets: []*asset{
							{FileName: "valid-sbom.json"},
						},
					},
				},
				Credentials: &repositoryCredentials{},
			},
			expectedResult: []string{"valid-sbom.json"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			result := GetArtifactNames(tt.stageRepo)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestGetArtifactNames_MultipleComponentsScenario(t *testing.T) {
	stageRepo := stageRepository{
		Components: []*component{
			{
				Assets: []*asset{
					{
						FileName:     "ppiper-240104-20240215175557.json",
						RelativePath: "com/sap/docker/ppiper/240104-20240215175557/ppiper-240104-20240215175557.json",
						URL:          "https://example/a.json",
					},
				},
			},
			{
				Assets: []*asset{
					{
						FileName:     "ppiper-240104-20240215175557-amd.json",
						RelativePath: "com/sap/docker/ppiper/240104-20240215175557/ppiper-240104-20240215175557-amd.json",
						URL:          "https://example/b.json",
					},
				},
			},
		},
		Credentials: &repositoryCredentials{},
	}
	expected := []string{
		"ppiper-240104-20240215175557.json",
		"ppiper-240104-20240215175557-amd.json",
	}
	actual := GetArtifactNames(stageRepo)
	assert.Equal(t, expected, actual)
}

func TestWriteSBomXmlForStageBuild(t *testing.T) {
	tests := []struct {
		name          string
		stageBom      map[string]interface{}
		setupServer   func() *httptest.Server
		expectedError string
		expectedFiles int
	}{
		{
			name: "Success - single raw repository with valid SBOM",
			stageBom: map[string]interface{}{
				"repo1": map[string]interface{}{
					"format": "raw",
					"credentials": map[string]interface{}{
						"user":       "testuser",
						"password":   "testpass",
						"repository": "testrepo",
					},
					"components": []interface{}{
						map[string]interface{}{
							"assets": []interface{}{
								map[string]interface{}{
									"fileName": "sbom.xml",
									"url":      "",
								},
							},
						},
					},
				},
			},
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("<sbom>test content</sbom>"))
				}))
			},
			expectedError: "",
			expectedFiles: 1,
		},
		{
			name: "Success - multiple raw repositories",
			stageBom: map[string]interface{}{
				"repo1": map[string]interface{}{
					"format": "raw",
					"credentials": map[string]interface{}{
						"user":       "testuser1",
						"password":   "testpass1",
						"repository": "testrepo1",
					},
					"components": []interface{}{
						map[string]interface{}{
							"assets": []interface{}{
								map[string]interface{}{
									"fileName": "sbom1.xml",
									"url":      "",
								},
							},
						},
					},
				},
				"repo2": map[string]interface{}{
					"format": "raw",
					"credentials": map[string]interface{}{
						"user":       "testuser2",
						"password":   "testpass2",
						"repository": "testrepo2",
					},
					"components": []interface{}{
						map[string]interface{}{
							"assets": []interface{}{
								map[string]interface{}{
									"fileName": "sbom2.json",
									"url":      "",
								},
							},
						},
					},
				},
			},
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("<sbom>test content</sbom>"))
				}))
			},
			expectedError: "",
			expectedFiles: 2,
		},
		{
			name: "Skip non-raw repositories",
			stageBom: map[string]interface{}{
				"repo1": map[string]interface{}{
					"format": "docker",
					"credentials": map[string]interface{}{
						"user":       "testuser",
						"password":   "testpass",
						"repository": "testrepo",
					},
				},
				"repo2": map[string]interface{}{
					"format": "maven2",
					"credentials": map[string]interface{}{
						"user":       "testuser",
						"password":   "testpass",
						"repository": "testrepo",
					},
				},
			},
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("<sbom>test content</sbom>"))
				}))
			},
			expectedError: "",
			expectedFiles: 0,
		},
		{
			name: "Error - HTTP fetch failure",
			stageBom: map[string]interface{}{
				"repo1": map[string]interface{}{
					"format": "raw",
					"credentials": map[string]interface{}{
						"user":       "testuser",
						"password":   "testpass",
						"repository": "testrepo",
					},
					"components": []interface{}{
						map[string]interface{}{
							"assets": []interface{}{
								map[string]interface{}{
									"fileName": "sbom.xml",
									"url":      "",
								},
							},
						},
					},
				},
			},
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))
			},
			expectedError: "failed to fetch stageBom file",
			expectedFiles: 0,
		},
		{
			name: "Error - invalid stageBom format",
			stageBom: map[string]interface{}{
				"invalid": "not a repository structure",
			},
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("<sbom>test content</sbom>"))
				}))
			},
			expectedError: "failed to unmarshal stageRepositories",
			expectedFiles: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Setup test server
			ts := tt.setupServer()
			defer ts.Close()

			// Update URLs in stageBom to point to test server
			if repos, ok := tt.stageBom["repo1"].(map[string]interface{}); ok {
				if components, ok := repos["components"].([]interface{}); ok {
					for _, comp := range components {
						if component, ok := comp.(map[string]interface{}); ok {
							if assets, ok := component["assets"].([]interface{}); ok {
								for _, ast := range assets {
									if asset, ok := ast.(map[string]interface{}); ok {
										asset["url"] = ts.URL
									}
								}
							}
						}
					}
				}
			}
			if repos, ok := tt.stageBom["repo2"].(map[string]interface{}); ok {
				if components, ok := repos["components"].([]interface{}); ok {
					for _, comp := range components {
						if component, ok := comp.(map[string]interface{}); ok {
							if assets, ok := component["assets"].([]interface{}); ok {
								for _, ast := range assets {
									if asset, ok := ast.(map[string]interface{}); ok {
										asset["url"] = ts.URL
									}
								}
							}
						}
					}
				}
			}

			// Create temporary directory for test
			tempDir := t.TempDir()
			originalWd, _ := os.Getwd()
			defer os.Chdir(originalWd)
			os.Chdir(tempDir)

			// Execute function
			result, err := WriteSBomXmlForStageBuild(tt.stageBom)

			// Assertions
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedFiles)

				// Verify files were created
				if tt.expectedFiles > 0 {
					sbomDir := filepath.Join(".", "sbom")
					_, err := os.Stat(sbomDir)
					assert.NoError(t, err, "SBOM directory should be created")
				}
			}
		})
	}
}

func TestWriteSBomXmlForStageBuild_InvalidJSON(t *testing.T) {
	// Test with invalid JSON that cannot be marshaled
	invalidStageBom := map[string]interface{}{
		"test": make(chan int), // channels cannot be marshaled to JSON
	}

	result, err := WriteSBomXmlForStageBuild(invalidStageBom)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse stageBom")
	assert.Nil(t, result)
}

func TestWriteSBomXmlForStageBuild_MultipleAssets(t *testing.T) {
	// Test with multiple assets per component
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<sbom>test content</sbom>"))
	}))
	defer ts.Close()

	stageBom := map[string]interface{}{
		"repo1": map[string]interface{}{
			"format": "raw",
			"credentials": map[string]interface{}{
				"user":       "testuser",
				"password":   "testpass",
				"repository": "testrepo",
			},
			"components": []interface{}{
				map[string]interface{}{
					"assets": []interface{}{
						map[string]interface{}{
							"fileName": "sbom1.xml",
							"url":      ts.URL,
						},
						map[string]interface{}{
							"fileName": "sbom2.json",
							"url":      ts.URL,
						},
					},
				},
			},
		},
	}

	// Create temporary directory for test
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	result, err := WriteSBomXmlForStageBuild(stageBom)

	assert.NoError(t, err)
	assert.Len(t, result, 2, "Should process both assets")

	// Verify both files were created
	sbomDir := filepath.Join(".", "sbom", "testrepo")
	file1 := filepath.Join(sbomDir, "sbom1.xml")
	file2 := filepath.Join(sbomDir, "sbom2.json")

	_, err1 := os.Stat(file1)
	_, err2 := os.Stat(file2)

	assert.NoError(t, err1, "First SBOM file should be created")
	assert.NoError(t, err2, "Second SBOM file should be created")
}

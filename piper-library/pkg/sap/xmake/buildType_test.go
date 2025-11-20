//go:build unit
// +build unit

package xmake

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestToBuildTypeFile(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		files := mock.FilesMock{}
		expectedFile := "build-type.json"
		s := BuildType{BuildType: "docker"}

		err := s.ToBuildTypeFile(&files)
		assert.NoError(t, err)

		exists, err := files.FileExists(expectedFile)
		assert.NoError(t, err)
		assert.True(t, exists)

		content, err := files.FileRead(expectedFile)
		assert.NoError(t, err)
		assert.Equal(t, `{"build-type":"docker"}`, string(content))
	})

	t.Run("error - write file", func(t *testing.T) {
		files := mock.FilesMock{FileWriteError: fmt.Errorf("write error")}
		expectedFile := fmt.Sprint("build-type.json")
		s := BuildType{BuildType: "docker"}
		err := s.ToBuildTypeFile(&files)
		assert.EqualError(t, err, fmt.Sprintf("failed to write %v: write error", expectedFile))
	})
}

func TestWriteBuildTypeJsonForStageBuild(t *testing.T) {

	t.Parallel()
	t.Run("success - docker build", func(t *testing.T) {
		// Mock the dependencies
		fileUtils := mock.FilesMock{}
		// Define the test input and output
		stageBomJSON := `{
							  "dd3a01b18081-20210709-040630878-903": {
									"components": [
									  {
										"artifact": "anything",
										"image": "docker.wdf.sap.corp:51021/something",
										"version": "anything"
									  }
									],
									"credentials": {
									  "password": "anything",
									  "repository": "anything",
									  "repositoryURL": "anything",
									  "user": "anything"
									},
									"format": "docker"
								  }
									  }
`
		var stageBom map[string]interface{}
		err := json.Unmarshal([]byte(stageBomJSON), &stageBom)
		assert.NoError(t, err)
		expectedBuildType := "docker"

		// Call the function and assert the result
		actualBuildType, err := WriteBuildTypeJsonForStageBuild(stageBom, &fileUtils)
		assert.NoError(t, err)
		assert.Equal(t, expectedBuildType, actualBuildType)
	})
	t.Run("success - not a docker build", func(t *testing.T) {
		// Mock the dependencies
		fileUtils := mock.FilesMock{}
		// Define the test input and output
		stageBomJSON := `{
							  "dd3a01b18081-20210709-040630878-903": {
									"components": [
									  {
										"artifact": "anything",
										"assets": [
										  {
											"classifier": "anything",
											"extension": "anything",
											"fileName": "anything",
											"relativePath": "anything",
											"url": "anything"
										  }
										],
										"group": "anything",
										"version": "anything"
									  }
									],
									"credentials": {
									  "password": "anything",
									  "repository": "anything",
									  "repositoryURL": "anything",
									  "user": "anything"
									},
									"format": "maven"
								  }
			  }
`
		var stageBom map[string]interface{}
		err := json.Unmarshal([]byte(stageBomJSON), &stageBom)
		assert.NoError(t, err)
		expectedBuildType := "bin"

		// Call the function and assert the result
		actualBuildType, err := WriteBuildTypeJsonForStageBuild(stageBom, &fileUtils)
		assert.NoError(t, err)
		assert.Equal(t, expectedBuildType, actualBuildType)
	})

	t.Run("fail - not a docker build", func(t *testing.T) {
		// Mock the dependencies
		files := mock.FilesMock{FileWriteError: fmt.Errorf("write error")}
		// Define the test input and output
		stageBomJSON := `{
							  "dd3a01b18081-20210709-040630878-903": {
									"components": [
									  {
										"artifact": "anything",
										"assets": [
										  {
											"classifier": "anything",
											"extension": "anything",
											"fileName": "anything",
											"relativePath": "anything",
											"url": "anything"
										  }
										],
										"group": "anything",
										"version": "anything"
									  }
									],
									"credentials": {
									  "password": "anything",
									  "repository": "anything",
									  "repositoryURL": "anything",
									  "user": "anything"
									},
									"format": "maven"
								  }
			  }
`
		var stageBom map[string]interface{}
		err := json.Unmarshal([]byte(stageBomJSON), &stageBom)
		assert.NoError(t, err)
		expectedBuildType := ""
		expectedFile := "build-type.json"

		actualBuildType, err := WriteBuildTypeJsonForStageBuild(stageBom, &files)
		assert.EqualError(t, err, fmt.Sprintf("failed to write %v: write error", expectedFile))
		assert.Equal(t, expectedBuildType, actualBuildType)
	})
}

func TestWriteBuildTypeToFile(t *testing.T) {

	t.Parallel()

	t.Run("success", func(t *testing.T) {
		files := mock.FilesMock{}
		expectedFile := "build-type.json"
		format := BuildType{BuildType: "bin"}

		err := WriteBuildTypeToFile(format, &files)
		assert.NoError(t, err)

		exists, err := files.FileExists(expectedFile)
		assert.NoError(t, err)
		assert.True(t, exists)

		content, err := files.FileRead(expectedFile)
		assert.NoError(t, err)
		assert.Equal(t, `{"build-type":"bin"}`, string(content))
	})

	t.Run("error - write file", func(t *testing.T) {
		files := mock.FilesMock{FileWriteError: fmt.Errorf("write error")}
		expectedFile := fmt.Sprint("build-type.json")
		format := BuildType{BuildType: "bin"}
		// Test error case
		// Create a fake BuildType with an invalid format
		err := WriteBuildTypeToFile(format, &files)
		assert.EqualError(t, err, fmt.Sprintf("failed to write %v: write error", expectedFile))
	})
}

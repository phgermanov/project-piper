//go:build unit
// +build unit

package cmd

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/SAP/jenkins-library/pkg/versioning"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type sapDownloadArtifactMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	config           sapDownloadArtifactOptions
	coordinates      versioning.Coordinates
	coordinateError  error
	downloadError    error
	nexusUrlNotFound error
	filename         string
	header           http.Header
	url              string
	untarError       error
}

func (s *sapDownloadArtifactMockUtils) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	if s.downloadError != nil {
		return s.downloadError
	}
	s.url = url
	s.filename = filename
	s.header = header
	return nil
}

func (s *sapDownloadArtifactMockUtils) getArtifactCoordinates(config *sapDownloadArtifactOptions) (versioning.Coordinates, error) {
	s.config = *config
	if s.coordinateError != nil {
		return versioning.Coordinates{}, s.coordinateError
	}
	return s.coordinates, nil
}

func (s *sapDownloadArtifactMockUtils) untar(src, dest string, stripComponentLevel int) error {
	if s.untarError != nil {
		return s.untarError
	}
	return nil
}

func newSapDownloadArtifactTestsUtils() sapDownloadArtifactMockUtils {
	utils := sapDownloadArtifactMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunSapDownloadArtifact(t *testing.T) {
	pipelineEnv := sapDownloadArtifactCommonPipelineEnvironment{}
	t.Parallel()

	t.Run("(MTA) success case - from Staging", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:        "mta",
			ArtifactID:       "sample1",
			Version:          "1.2.5",
			MtarPublishedURL: "https://my.reporisory.url/stage/repository/uniqueId/com/sap/test/sample1/1.2.5/sample1-1.2.5.mtar",
			FromStaging:      true,
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)

		assert.Equal(t, "sample1.mtar", utils.filename)
		assert.Equal(t, "https://my.reporisory.url/stage/repository/uniqueId/com/sap/test/sample1/1.2.5/sample1-1.2.5.mtar", utils.url)
	})

	t.Run("(MTA) success case - from Staging with the targetPath parameter", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:        "mta",
			ArtifactID:       "sample1",
			Version:          "1.2.5",
			MtarPublishedURL: "https://my.reporisory.url/stage/repository/uniqueId/com/sap/test/sample1/1.2.5/sample1-1.2.5.mtar",
			FromStaging:      true,
			TargetPath:       "./folder",
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)

		assert.Equal(t, "folder/sample1.mtar", utils.filename)
		assert.Equal(t, "https://my.reporisory.url/stage/repository/uniqueId/com/sap/test/sample1/1.2.5/sample1-1.2.5.mtar", utils.url)
	})

	t.Run("(MTA) success case - from Promote", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:            "mta",
			ArtifactID:           "sample1",
			Version:              "1.2.5",
			ArtifactDownloadURLs: []string{"https://common.repositories.cloud.sap/deploy-releases-hyperspace-maven/com/sap/test/sample1/1.2.5/sample1-1.2.5.mtar"},
			FromStaging:          false,
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)

		assert.Equal(t, "sample1.mtar", utils.filename)
		assert.Equal(t, "https://common.repositories.cloud.sap/deploy-releases-hyperspace-maven/com/sap/test/sample1/1.2.5/sample1-1.2.5.mtar", utils.url)
	})

	t.Run("(NPM) success case - from Staging", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:     "npm",
			ArtifactID:    "piper-validation-npm",
			Version:       "1.0.0",
			RepositoryURL: "https://my.reporisory.url/stage/repository/uniqueId",
			FromStaging:   true,
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)

		assert.Equal(t, "piper-validation-npm.tgz", utils.filename)
		assert.Equal(t, "https://my.reporisory.url/stage/repository/uniqueId/piper-validation-npm/-/piper-validation-npm-1.0.0.tgz", utils.url)
	})

	t.Run("(NPM) success case - from Staging with the targetPath parameter", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:      "npm",
			ArtifactID:     "@sap/piper-validation-npm",
			Version:        "1.0.0",
			RepositoryURL:  "https://my.reporisory.url/stage/repository/uniqueId",
			FromStaging:    true,
			ExtractPackage: true,
			TargetPath:     "folder",
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)

		assert.Equal(t, "folder/@sap/piper-validation-npm.tgz", utils.filename)
		assert.Equal(t, "https://my.reporisory.url/stage/repository/uniqueId/@sap/piper-validation-npm/-/piper-validation-npm-1.0.0.tgz", utils.url)
	})

	t.Run("(NPM) success case - from Staging with build info in Semver", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:     "npm",
			ArtifactID:    "piper-validation-npm",
			Version:       "14.0.0-20210913144601+e775c9a302e91983a804e6da19e31a53c6fc11e9",
			RepositoryURL: "https://my.reporisory.url/stage/repository/uniqueId",
			FromStaging:   true,
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)

		assert.Equal(t, "piper-validation-npm.tgz", utils.filename)
		assert.Equal(t, "https://my.reporisory.url/stage/repository/uniqueId/piper-validation-npm/-/piper-validation-npm-14.0.0-20210913144601.tgz", utils.url)
	})

	t.Run("(NPM) success case - from Staging with basic auth", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:                        "npm",
			ArtifactID:                       "piper-validation-npm",
			Version:                          "1.0.0",
			RepositoryURL:                    "https://my.reporisory.url/stage/repository/uniqueId",
			FromStaging:                      true,
			StagingServiceRepositoryUser:     "stagingUser",
			StagingServiceRepositoryPassword: "stagingPassword",
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)

		assert.Equal(t, http.Header{"Authorization": []string{"Basic c3RhZ2luZ1VzZXI6c3RhZ2luZ1Bhc3N3b3Jk"}}, utils.header)
	})

	t.Run("(NPM) success case - from Promote", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{
			BuildTool:            "npm",
			ArtifactID:           "piper-validation-npm",
			Version:              "1.0.0",
			RepositoryURL:        "https://my.reporisory.url/stage/repository/uniqueId",
			ArtifactDownloadURLs: []string{"https://my.reporisory.url/deploy.milestones/piper-validation-npm/-/piper-validation-npm-1.0.0.tgz"},
			FromStaging:          false,
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)

		assert.Equal(t, "piper-validation-npm.tgz", utils.filename)
		assert.Equal(t, "https://my.reporisory.url/deploy.milestones/piper-validation-npm/-/piper-validation-npm-1.0.0.tgz", utils.url)
	})

	t.Run("(NPM) success case - from Promote with the targetPath parameter", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{
			BuildTool:            "npm",
			ArtifactID:           "piper-validation-npm",
			Version:              "1.0.0",
			RepositoryURL:        "https://my.reporisory.url/stage/repository/uniqueId",
			ArtifactDownloadURLs: []string{"https://my.reporisory.url/deploy.milestones/piper-validation-npm/-/piper-validation-npm-1.0.0.tgz"},
			FromStaging:          false,
			TargetPath:           "./folder/",
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)

		assert.Equal(t, "folder/piper-validation-npm.tgz", utils.filename)
		assert.Equal(t, "https://my.reporisory.url/deploy.milestones/piper-validation-npm/-/piper-validation-npm-1.0.0.tgz", utils.url)
	})

	t.Run("(NPM) success case - from Promote with token", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{
			BuildTool:            "npm",
			ArtifactID:           "piper-validation-npm",
			Version:              "1.0.0",
			RepositoryURL:        "https://my.reporisory.url/stage/repository/uniqueId",
			ArtifactDownloadURLs: []string{"https://my.reporisory.url/deploy.milestones/piper-validation-npm/-/piper-validation-npm-1.0.0.tgz"},
			FromStaging:          false,
			ArtifactoryToken:     "myJFrogTestToken",
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)

		assert.Equal(t, http.Header{"X-JFrog-Art-Api": []string{"myJFrogTestToken"}}, utils.header)
	})

	t.Run("(NPM) bad case - with the extractPackage parameter", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:      "npm",
			ArtifactID:     "@sap/piper-validation-npm",
			Version:        "1.0.0",
			RepositoryURL:  "https://my.reporisory.url/stage/repository/uniqueId",
			FromStaging:    true,
			ExtractPackage: true,
		}
		utils := newSapDownloadArtifactTestsUtils()
		utils.untarError = fmt.Errorf("failed to untar")
		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.EqualError(t, err, "package extraction failed: failed to untar")
	})

	t.Run("(CAP) bad case - with the extractPackage parameter", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:  "CAP",
			ArtifactID: "testArtifact",
			ArtifactDownloadURLs: []string{
				"https://my.reporisory.url/deploy.milestones/piper-validation/-/testArtifact-1.0.0.tgz",
			},
			FromStaging:      false,
			ArtifactoryToken: "myJFrogTestToken",
			ExtractPackage:   true,
		}
		utils := newSapDownloadArtifactTestsUtils()
		utils.untarError = fmt.Errorf("failed to untar")
		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.EqualError(t, err, "filename testArtifact-1.0.0.tgz package extraction failed: failed to untar")
	})

	t.Run("(pip) success case - from Staging with basic auth", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:                        "pip",
			ArtifactID:                       "piper-validation-pip",
			Version:                          "1.0.0",
			RepositoryURL:                    "https://my.reporisory.url/stage/repository/uniqueId",
			FromStaging:                      true,
			StagingServiceRepositoryUser:     "stagingUser",
			StagingServiceRepositoryPassword: "stagingPassword",
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)

		assert.Equal(t, http.Header{"Authorization": []string{"Basic c3RhZ2luZ1VzZXI6c3RhZ2luZ1Bhc3N3b3Jk"}}, utils.header)
	})

	t.Run("(pip) success case - from Promote", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{
			BuildTool:            "pip",
			ArtifactID:           "piper-validation-pip",
			Version:              "1.0.0",
			RepositoryURL:        "https://my.reporisory.url/stage/repository/uniqueId",
			ArtifactDownloadURLs: []string{"https://my.reporisory.url/deploy.milestones/piper-validation-npm/1.0.0/piper_validation_pip-1.0.0.tar.gz"},
			FromStaging:          false,
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)

		assert.Equal(t, "piper-validation-pip.tar.gz", utils.filename)
		assert.Equal(t, "https://my.reporisory.url/deploy.milestones/piper-validation-npm/1.0.0/piper_validation_pip-1.0.0.tar.gz", utils.url)
	})

	t.Run("(pip) success case - from Promote with the targetPath parameter", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{
			BuildTool:            "pip",
			ArtifactID:           "piper-validation-pip",
			Version:              "1.0.0",
			RepositoryURL:        "https://my.reporisory.url/stage/repository/uniqueId",
			ArtifactDownloadURLs: []string{"https://my.reporisory.url/deploy.milestones/piper-validation-npm/1.0.0/piper_validation_pip-1.0.0.tar.gz"},
			FromStaging:          false,
			TargetPath:           "folder",
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)

		assert.Equal(t, "folder/piper-validation-pip.tar.gz", utils.filename)
		assert.Equal(t, "https://my.reporisory.url/deploy.milestones/piper-validation-npm/1.0.0/piper_validation_pip-1.0.0.tar.gz", utils.url)
	})

	t.Run("(pip) success case - from Promote with token", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{
			BuildTool:            "pip",
			ArtifactID:           "piper-validation-pip",
			Version:              "1.0.0",
			RepositoryURL:        "https://my.reporisory.url/stage/repository/uniqueId",
			ArtifactDownloadURLs: []string{"https://my.reporisory.url/deploy.milestones/piper-validation-npm/1.0.0/piper_validation_pip-1.0.0.tar.gz"},
			FromStaging:          false,
			ArtifactoryToken:     "myJFrogTestToken",
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)

		assert.Equal(t, http.Header{"X-JFrog-Art-Api": []string{"myJFrogTestToken"}}, utils.header)
	})

	t.Run("(pip) bad case - with the extractPackage parameter", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:                        "pip",
			ArtifactID:                       "piper-validation-pip",
			Version:                          "1.0.0",
			RepositoryURL:                    "https://my.reporisory.url/stage/repository/uniqueId",
			FromStaging:                      true,
			StagingServiceRepositoryUser:     "stagingUser",
			StagingServiceRepositoryPassword: "stagingPassword",
			ExtractPackage:                   true,
		}
		utils := newSapDownloadArtifactTestsUtils()
		utils.untarError = fmt.Errorf("failed to untar")

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.EqualError(t, err, "package extraction failed: failed to untar")
	})

	t.Run("(golang) success case - from Staging with 1 binary", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:     "golang",
			GroupID:       "com/sap/test",
			ArtifactID:    "sample-go",
			Version:       "1.2.5",
			RepositoryURL: "https://my.reporisory.url/stage/repository/uniqueId",
			FromStaging:   true,
			TargetPath:    "folder",
			Artifacts: []map[string]interface{}{
				{"name": "sample-linux.amd64"},
			},
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)
		assert.Equal(t, "folder/sample-linux.amd64", utils.filename)
		assert.Equal(t, "https://my.reporisory.url/stage/repository/uniqueId/go/com/sap/test/sample-go/1.2.5/sample-linux.amd64", utils.url)
	})

	t.Run("(golang) success case - from Staging with more than 1 binary", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:     "golang",
			GroupID:       "com/sap/test",
			ArtifactID:    "sample-go",
			Version:       "1.2.5",
			RepositoryURL: "https://my.reporisory.url/stage/repository/uniqueId",
			FromStaging:   true,
			Artifacts: []map[string]interface{}{
				{"name": "sample-linux.amd64"},
				{"name": "sample-windows.arm64.exe"},
			},
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)
		//utils.url contains last
		assert.Equal(t, "https://my.reporisory.url/stage/repository/uniqueId/go/com/sap/test/sample-go/1.2.5/sample-windows.arm64.exe", utils.url)
	})

	t.Run("(golang) success case - from Promote", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:            "golang",
			GroupID:              "com/sap/test",
			ArtifactID:           "sample-go",
			Version:              "1.2.5",
			RepositoryURL:        "https://my.reporisory.url/stage/repository/uniqueId",
			ArtifactDownloadURLs: []string{"https://my.reporisory.url/deploy.milestones/go/com/sap/test/sample-go/1.2.5/sample-linux.amd64"},
			FromStaging:          false,
			Artifacts: []map[string]interface{}{
				{"name": "sample-linux.amd64"},
			},
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)
		assert.Equal(t, "https://my.reporisory.url/deploy.milestones/go/com/sap/test/sample-go/1.2.5/sample-linux.amd64", utils.url)
	})

	t.Run("(golang) fail when no binaries available", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:     "golang",
			GroupID:       "com/sap/test",
			ArtifactID:    "sample-go",
			Version:       "1.2.5",
			RepositoryURL: "https://my.reporisory.url/stage/repository/uniqueId",
			FromStaging:   true,
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.Error(t, err)
	})

	t.Run("(maven) success case - from Staging", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{
			BuildTool:     "maven",
			GroupID:       "my.testGroup",
			ArtifactID:    "testArtifact",
			Version:       "1.0.0",
			Packaging:     "jar",
			RepositoryURL: "https://my.reporisory.url/stage/repository/uniqueId",
			FromStaging:   true,
			TargetPath:    "folder1/folder2",
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)

		assert.Equal(t, "folder1/folder2/testArtifact.jar", utils.filename)
		assert.Equal(t, "https://my.reporisory.url/stage/repository/uniqueId/my/testGroup/testArtifact/1.0.0/testArtifact-1.0.0.jar", utils.url)
	})

	t.Run("(maven) success case - from Promote without dependencies", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{
			BuildTool:            "maven",
			GroupID:              "my.testGroup",
			ArtifactID:           "testArtifact",
			Version:              "1.0.0",
			Packaging:            "jar",
			RepositoryURL:        "https://my.reporisory.url/stage/repository/uniqueId",
			ArtifactDownloadURLs: []string{"https://my.reporisory.url/deploy.milestones/my/testGroup/testArtifact/1.0.0/testArtifact-1.0.0.jar", "https://my.reporisory.url/deploy.milestones/my/testGroup/testArtifact/1.0.0/testArtifact-1.0.0-jar-with-dependencies.jar"},
			FromStaging:          false,
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)

		assert.Equal(t, "testArtifact.jar", utils.filename)
		assert.Equal(t, "https://my.reporisory.url/deploy.milestones/my/testGroup/testArtifact/1.0.0/testArtifact-1.0.0.jar", utils.url)
	})

	t.Run("(maven) success case - from Promote with dependencies", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{
			BuildTool:                "maven",
			GroupID:                  "my.testGroup",
			ArtifactID:               "testArtifact",
			Version:                  "1.0.0",
			Packaging:                "jar",
			RepositoryURL:            "https://my.reporisory.url/stage/repository/uniqueId",
			ArtifactDownloadURLs:     []string{"https://my.reporisory.url/deploy.milestones/my/testGroup/testArtifact/1.0.0/testArtifact-1.0.0.jar", "https://my.reporisory.url/deploy.milestones/my/testGroup/testArtifact/1.0.0/testArtifact-1.0.0-jar-with-dependencies.jar"},
			FromStaging:              false,
			ArtifactWithDependencies: true,
			TargetPath:               "./folder1/folder2",
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)

		assert.Equal(t, "folder1/folder2/testArtifact.jar", utils.filename)
		assert.Equal(t, "https://my.reporisory.url/deploy.milestones/my/testGroup/testArtifact/1.0.0/testArtifact-1.0.0-jar-with-dependencies.jar", utils.url)
	})

	t.Run("(maven) success case - from Promote with token", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{
			BuildTool:                "maven",
			GroupID:                  "my.testGroup",
			ArtifactID:               "testArtifact",
			Version:                  "1.0.0",
			Packaging:                "jar",
			RepositoryURL:            "https://my.reporisory.url/stage/repository/uniqueId",
			ArtifactDownloadURLs:     []string{"https://my.reporisory.url/deploy.milestones/my/testGroup/testArtifact/1.0.0/testArtifact-1.0.0.jar", "https://my.reporisory.url/deploy.milestones/my/testGroup/testArtifact/1.0.0/testArtifact-1.0.0-jar-with-dependencies.jar"},
			FromStaging:              false,
			ArtifactWithDependencies: true,
			ArtifactoryToken:         "myJFrogTestToken",
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)

		assert.Equal(t, http.Header{"X-JFrog-Art-Api": []string{"myJFrogTestToken"}}, utils.header)
	})

	t.Run("(maven) success case - from Staging with basic authentication", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{
			BuildTool:                        "maven",
			GroupID:                          "my.testGroup",
			ArtifactID:                       "testArtifact",
			Version:                          "1.0.0",
			Packaging:                        "jar",
			RepositoryURL:                    "https://my.reporisory.url/stage/repository/uniqueId",
			FromStaging:                      true,
			StagingServiceRepositoryUser:     "stagingUser",
			StagingServiceRepositoryPassword: "stagingPassword",
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.NoError(t, err)

		assert.Equal(t, http.Header{"Authorization": []string{"Basic c3RhZ2luZ1VzZXI6c3RhZ2luZ1Bhc3N3b3Jk"}}, utils.header)
	})

	t.Run("identify artifact download url", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:                "maven",
			GroupID:                  "my.testGroup",
			ArtifactID:               "testArtifact",
			Version:                  "1.0.0",
			Packaging:                "jar",
			RepositoryURL:            "https://my.reporisory.url/stage/repository/uniqueId",
			ArtifactDownloadURLs:     []string{"https://my.reporisory.url/deploy.milestones/my/testGroup/testArtifact/1.0.0/testArtifact-1.0.0.jar", "https://my.reporisory.url/deploy.milestones/my/testGroup/testArtifact/1.0.0/testArtifact-1.0.0-jar-with-dependencies.jar"},
			FromStaging:              false,
			ArtifactWithDependencies: true,
			ArtifactoryToken:         "myJFrogTestToken",
		}
		result := identifyArtifactDownlURL(&config, "testArtifact-1.0.0.jar")
		assert.Equal(t, result, "https://my.reporisory.url/deploy.milestones/my/testGroup/testArtifact/1.0.0/testArtifact-1.0.0.jar")

	})

	t.Run("(CAP) identify all artifact download URLs", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:  "CAP",
			GroupID:    "my.testGroup",
			ArtifactID: "testArtifact",
			ArtifactDownloadURLs: []string{
				"https://my.reporisory.url/deploy.milestones/piper-validation/-/testArtifact-1.0.0.tgz",
				"https://my.reporisory.url/deploy.milestones/piper-validation-npm/-/testArtifact-npm-1.0.0.tgz",
				"https://my.reporisory.url/deploy-releases-hyperspace-maven/my/testArtifact-1.0.0.jar",
				"https://my.reporisory.url/deploy-releases-hyperspace-maven/my/testArtifact-1.0.0.pom",
			},
		}
		result := identifyAllArtifactDownloadURLs(config.ArtifactDownloadURLs, "tgz")

		assert.ElementsMatch(t, result, []string{
			"https://my.reporisory.url/deploy.milestones/piper-validation/-/testArtifact-1.0.0.tgz",
			"https://my.reporisory.url/deploy.milestones/piper-validation-npm/-/testArtifact-npm-1.0.0.tgz",
		})
	})

	t.Run("CAP Download Artifacts - From Staging", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:  "CAP",
			GroupID:    "my.testGroup",
			ArtifactID: "testArtifact",
			ArtifactStagingDownloadURLs: []string{
				"https://my.reporisory.url/deploy.milestones/piper-validation/-/testArtifact-1.0.0.tgz",
				"https://my.reporisory.url/deploy.milestones/piper-validation-npm/-/testArtifact-npm-1.0.0.tgz",
			},
			FromStaging: true,
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := capDownloadArtifacts(&config, nil, &utils, &pipelineEnv)
		assert.NoError(t, err)
	})

	t.Run("(mta) identify artifact fully qualified name", func(t *testing.T) {
		t.Parallel()
		result, err := returnArtifactIdentifier("%v-%v.%v", "test", "1.0", "mta", "https://my.reporisory.url/deploy.milestones/my/testGroup/testArtifact/1.0.0/custom-1.0.mtar")
		assert.NoError(t, err)
		assert.Equal(t, result, "custom-1.0.mtar")
	})

	t.Run("(npm) identify artifact fully qualified name", func(t *testing.T) {
		t.Parallel()
		result, err := returnArtifactIdentifier("%v-%v.%v", "test", "1.0", "npm", "")
		assert.NoError(t, err)
		assert.Equal(t, result, "test-1.0.tgz")
	})

	t.Run("(pip) identify artifact fully qualified name", func(t *testing.T) {
		t.Parallel()
		result, err := returnArtifactIdentifier("%v-%v.%v", "test", "1.0", "pip", "")
		assert.NoError(t, err)
		assert.Equal(t, result, "test-1.0.tar.gz")
	})

	t.Run("(npm scoped) identify artifact fully qualified name", func(t *testing.T) {
		t.Parallel()
		result, err := returnArtifactIdentifier("%v-%v.%v", "@npm/test", "1.0", "npm", "")
		assert.NoError(t, err)
		assert.Equal(t, result, "test-1.0.tgz")
	})

	t.Run("(other)error - identify artifact fully qualified name", func(t *testing.T) {
		t.Parallel()
		_, err := returnArtifactIdentifier("%v-%v.%v", "test", "1.0", "maven", "https://my.reporisory.url/deploy.milestones/my/testGroup/testArtifact/1.0.0/custom-1.0.mtar")

		assert.EqualError(t, err, "failed to identify artifacts fully qualified name for type: maven")
	})

	t.Run("test common download handler from staging", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:                        "npm",
			ArtifactID:                       "piper-validation-npm",
			Version:                          "1.0.0",
			RepositoryURL:                    "https://my.reporisory.url/stage/repository/uniqueId",
			FromStaging:                      true,
			StagingServiceRepositoryUser:     "stagingUser",
			StagingServiceRepositoryPassword: "stagingPassword",
			TargetPath:                       "./folder",
		}
		utils := newSapDownloadArtifactTestsUtils()

		filename, err := commonDownloadHandeler(&config, &utils, "piper-validation-npm-1.0.0.tgz")

		assert.NoError(t, err)
		assert.Equal(t, "folder/piper-validation-npm.tgz", filename)
		assert.Equal(t, http.Header{"Authorization": []string{"Basic c3RhZ2luZ1VzZXI6c3RhZ2luZ1Bhc3N3b3Jk"}}, utils.header)
	})

	t.Run("(CAP) test common multiple file download handler - from Promote", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:  "CAP",
			ArtifactID: "testArtifact",
			Version:    "1.0.0",
			ArtifactDownloadURLs: []string{
				"https://my.reporisory.url/deploy.milestones/piper-validation/-/testArtifact-1.0.0.tgz",
				"https://my.reporisory.url/deploy.milestones/piper-validation-npm/-/testArtifact-npm-1.0.0.tgz",
				"https://my.reporisory.url/deploy-releases-hyperspace-maven/my/testArtifact-1.0.0.jar",
				"https://my.reporisory.url/deploy-releases-hyperspace-maven/my/testArtifact-1.0.0.pom",
			},
			ArtifactoryToken: "myJFrogTestToken",
			FromStaging:      false,
			TargetPath:       "./",
		}
		utils := newSapDownloadArtifactTestsUtils()

		filenames, err := commonMultipleFileDownloadHandler(&config, &utils)

		assert.NoError(t, err)
		assert.ElementsMatch(t, filenames, []string{
			"testArtifact-1.0.0.tgz",
			"testArtifact-npm-1.0.0.tgz",
		})
		assert.Equal(t, http.Header{"X-JFrog-Art-Api": []string{config.ArtifactoryToken}}, utils.header)
	})

	t.Run("(CAP) error - build tool not handled by common Multiple File Download Handler", func(t *testing.T) {
		t.Parallel()
		config := sapDownloadArtifactOptions{
			BuildTool:  "TestTool",
			ArtifactID: "testArtifact",
			Version:    "1.0.0",
		}
		utils := newSapDownloadArtifactTestsUtils()

		filenames, err := commonMultipleFileDownloadHandler(&config, &utils)
		assert.Nil(t, filenames)
		assert.EqualError(t, err, "build tool not handled by common Multiple File Download Handler")
	})

	t.Run("(CAP) error - failed to identify artifact download url - from promote ", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{
			BuildTool:            "CAP",
			ArtifactDownloadURLs: []string{},
			FromStaging:          false,
		}
		utils := newSapDownloadArtifactTestsUtils()
		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)
		assert.EqualError(t, err, "failed to download artifact: unable to identify npm packages from 'CAP'")
	})

	t.Run("(maven) error - failed to identify artifact download url from staging ", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{
			BuildTool:   "maven",
			GroupID:     "my.testGroup",
			ArtifactID:  "testArtifact",
			Version:     "1.0.0",
			Packaging:   "jar",
			FromStaging: false,
		}
		utils := newSapDownloadArtifactTestsUtils()
		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)
		assert.EqualError(t, err, "unable to identify 'maven' artifact download url")
	})

	t.Run("build tool CAP", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{
			BuildTool: "CAP",
			ArtifactDownloadURLs: []string{
				"https://my.reporisory.url/deploy.milestones/piper-validation/-/testArtifact-1.0.0.tgz",
				"https://my.reporisory.url/deploy.milestones/piper-validation-npm/-/testArtifact-npm-1.0.0.tgz",
			},
		}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		// assert
		assert.NoError(t, err)
	})

	t.Run("build tool docker", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{BuildTool: "docker"}
		utils := newSapDownloadArtifactTestsUtils()

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		// assert
		assert.NoError(t, err)
	})

	t.Run("(maven) error - failed to get coordinates", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{BuildTool: "maven"}
		utils := newSapDownloadArtifactTestsUtils()
		utils.coordinateError = fmt.Errorf("coordinateError")

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.EqualError(t, err, "failed to get missing coordinates: coordinateError")
	})

	t.Run("error - failed to download", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{
			BuildTool:     "maven",
			RepositoryURL: "https://my.reporisory.url/stage/repository/uniqueId",
			FromStaging:   true,
		}
		utils := newSapDownloadArtifactTestsUtils()
		utils.downloadError = fmt.Errorf("downloadError")

		err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)

		assert.EqualError(t, err, "failed to download artifact: downloadError")
	})

	t.Run("gradle (fail) - there are no artifacts for downloading", func(t *testing.T) {
		t.Parallel()
		config := &sapDownloadArtifactOptions{
			BuildTool: "gradle",
		}
		utils := newSapDownloadArtifactTestsUtils()
		err := runSapDownloadArtifact(config, nil, &utils, &pipelineEnv)
		assert.Equal(t, fmt.Errorf("there are no artifacts available for downloading"), err)
	})

	t.Run("gradle (success) - download multiple gradle artifacts", func(t *testing.T) {
		t.Parallel()
		config := &sapDownloadArtifactOptions{
			BuildTool:     "gradle",
			Artifacts:     []map[string]interface{}{{"id": "piper-validation", "name": "gradle-app-1-0.0.1.jar"}, {"id": "piper-validation", "name": "gradle-app-2-0.0.1.jar"}},
			ArtifactID:    "piper-validation",
			GroupID:       "com.example",
			Version:       "1.0.0",
			RepositoryURL: "https://my.reporisory.url/stage/repository/uniqueId",
			FromStaging:   true,
			TargetPath:    "folder",
		}
		utils := newSapDownloadArtifactTestsUtils()
		err := runSapDownloadArtifact(config, nil, &utils, &pipelineEnv)
		assert.NoError(t, err)
		assert.Equal(t, "folder/gradle-app-2-0.0.1.jar", utils.filename)
	})

	t.Run("different buildTools + helm", func(t *testing.T) {
		t.Parallel()
		pipelineEnv := sapDownloadArtifactCommonPipelineEnvironment{}
		expectedFilename := "test-helm-chart-1.0.0.tgz"
		expectedURL := "https://my.reporisory.url/stage/repository/uniqueId/test-helm-chart/test-helm-chart-1.0.0.tgz"
		config := sapDownloadArtifactOptions{
			Artifacts:                     []map[string]interface{}{{"name": "golang-app"}},
			ArtifactID:                    "piper-validation",
			Version:                       "1.0.0",
			RepositoryURL:                 "https://my.reporisory.url/stage/repository/uniqueId",
			HelmChartURL:                  "https://my.reporisory.url/stage/repository/uniqueId/test-helm-chart/test-helm-chart-1.0.0.tgz",
			HelmStagingRepositoryUsername: "stagingUser",
			HelmStagingRepositoryPassword: "stagingPassword",
			FromStaging:                   true,
			MtarPublishedURL:              "https://my.reporisory.url/stage/repository/uniqueId/test-helm-mtar/test-helm-mtar-1.0.0.tgz",
		}
		tt := []struct {
			testCase  string
			buildTool string
		}{
			{
				testCase:  "npm + helm",
				buildTool: "npm",
			},
			{
				testCase:  "pip + helm",
				buildTool: "pip",
			},
			{
				testCase:  "golang + helm",
				buildTool: "golang",
			},
			{
				testCase:  "maven + helm",
				buildTool: "maven",
			},
			{
				testCase:  "mta + helm",
				buildTool: "mta",
			},
			{
				testCase:  "docker + helm",
				buildTool: "docker",
			},
			{
				testCase:  "CAP + helm",
				buildTool: "CAP",
			},
			{
				testCase:  "gradle + helm",
				buildTool: "gradle",
			},
		}

		for _, test := range tt {
			test := test
			t.Run(test.testCase, func(t *testing.T) {
				t.Parallel()
				utils := newSapDownloadArtifactTestsUtils()
				config := config
				config.BuildTool = test.buildTool

				if test.buildTool == "CAP" {
					config.ArtifactStagingDownloadURLs = []string{
						"https://my.reporisory.url/deploy.milestones/piper-validation/-/testArtifact-1.0.0.tgz",
					}
				}

				err := runSapDownloadArtifact(&config, nil, &utils, &pipelineEnv)
				assert.NoError(t, err)

				if test.buildTool == "CAP" {
					assert.Equal(t, "testArtifact-1.0.0.tgz", utils.filename)
					assert.Equal(t, "https://my.reporisory.url/deploy.milestones/piper-validation/-/testArtifact-1.0.0.tgz", utils.url)
				} else {
					assert.Equal(t, expectedFilename, utils.filename)
					assert.Equal(t, expectedURL, utils.url)
				}
			})
		}
	})
}

func TestAddMissingCoordinates(t *testing.T) {
	t.Parallel()

	t.Run("success case", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{BuildTool: "maven"}
		utils := newSapDownloadArtifactTestsUtils()
		utils.coordinates = versioning.Coordinates{
			GroupID:    "testGroup",
			ArtifactID: "testArtifact",
			Version:    "1.0.0",
			Packaging:  "jar",
		}

		tt := []struct {
			coordinates versioning.Coordinates
			expected    versioning.Coordinates
		}{
			{
				coordinates: versioning.Coordinates{},
				expected:    versioning.Coordinates{GroupID: "testGroup", ArtifactID: "testArtifact", Version: "1.0.0", Packaging: "jar"},
			},
			{
				coordinates: versioning.Coordinates{GroupID: "myGroup", ArtifactID: "myArtifact", Version: "2.0.0", Packaging: "zip"},
				expected:    versioning.Coordinates{GroupID: "myGroup", ArtifactID: "myArtifact", Version: "2.0.0", Packaging: "zip"},
			},
		}

		for i, test := range tt {
			err := addMissingCoordinates(&config, &test.coordinates, &utils)

			assert.NoError(t, err, i)
			assert.Equal(t, test.expected, test.coordinates, i)
		}
	})

	t.Run("success case - maven-plugin", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{BuildTool: "maven"}
		coordinates := versioning.Coordinates{}
		utils := newSapDownloadArtifactTestsUtils()
		utils.coordinates = versioning.Coordinates{
			GroupID:    "testGroup",
			ArtifactID: "testArtifact",
			Version:    "1.0.0",
			Packaging:  "maven-plugin",
		}

		err := addMissingCoordinates(&config, &coordinates, &utils)

		assert.NoError(t, err)
		assert.Equal(t, "jar", coordinates.Packaging)
	})

	t.Run("error - coordinate error", func(t *testing.T) {
		t.Parallel()

		config := sapDownloadArtifactOptions{BuildTool: "maven"}
		utils := newSapDownloadArtifactTestsUtils()
		utils.coordinateError = fmt.Errorf("coordinateError")

		err := addMissingCoordinates(&config, &versioning.Coordinates{}, &utils)

		assert.EqualError(t, err, "coordinateError")
	})
}

func TestHelmDownloadArtifact(t *testing.T) {
	t.Parallel()
	pipelineEnv := sapDownloadArtifactCommonPipelineEnvironment{}
	tt := []struct {
		testCase         string
		config           sapDownloadArtifactOptions
		downloadErr      error
		expectedFilename string
		expectedURL      string
		expectedHeader   http.Header
		expectedErr      error
	}{
		{
			testCase: "good case - from Staging",
			config: sapDownloadArtifactOptions{
				HelmChartURL:                  "https://my.reporisory.url/stage/repository/uniqueId/test-helm-chart/test-helm-chart-1.0.0-12345.tgz",
				HelmStagingRepositoryUsername: "stagingUser",
				HelmStagingRepositoryPassword: "stagingPassword",
				FromStaging:                   true,
			},
			downloadErr:      nil,
			expectedFilename: "test-helm-chart-1.0.0-12345.tgz",
			expectedURL:      "https://my.reporisory.url/stage/repository/uniqueId/test-helm-chart/test-helm-chart-1.0.0-12345.tgz",
			expectedHeader:   http.Header{"Authorization": {"Basic c3RhZ2luZ1VzZXI6c3RhZ2luZ1Bhc3N3b3Jk"}},
			expectedErr:      nil,
		},
		{
			testCase: "good case - from Staging with the targetPath parameter",
			config: sapDownloadArtifactOptions{
				HelmChartURL:                  "https://my.reporisory.url/stage/repository/uniqueId/test-helm-chart/test-helm-chart-1.0.0-12345.tgz",
				HelmStagingRepositoryUsername: "stagingUser",
				HelmStagingRepositoryPassword: "stagingPassword",
				FromStaging:                   true,
				TargetPath:                    "folder",
			},
			downloadErr:      nil,
			expectedFilename: "folder/test-helm-chart-1.0.0-12345.tgz",
			expectedURL:      "https://my.reporisory.url/stage/repository/uniqueId/test-helm-chart/test-helm-chart-1.0.0-12345.tgz",
			expectedHeader:   http.Header{"Authorization": {"Basic c3RhZ2luZ1VzZXI6c3RhZ2luZ1Bhc3N3b3Jk"}},
			expectedErr:      nil,
		},
		{
			testCase: "bad case - from Staging",
			config: sapDownloadArtifactOptions{
				HelmChartURL:                  "",
				HelmStagingRepositoryUsername: "stagingUser",
				HelmStagingRepositoryPassword: "stagingPassword",
				FromStaging:                   true,
			},
			downloadErr:      nil,
			expectedFilename: "",
			expectedURL:      "",
			expectedErr:      fmt.Errorf("unable to identify helm artifact download url"),
		},
		{
			testCase: "good case - from Promote",
			config: sapDownloadArtifactOptions{
				ArtifactoryToken: "myJFrogTestToken",
				HelmChartURL:     "https://my.reporisory.url/deploy-hyperspace-helm/test-helm-chart/test-helm-chart-1.0.0-12345.tgz",
				FromStaging:      false,
			},
			downloadErr:      nil,
			expectedFilename: "test-helm-chart-1.0.0-12345.tgz",
			expectedURL:      "https://my.reporisory.url/deploy-hyperspace-helm/test-helm-chart/test-helm-chart-1.0.0-12345.tgz",
			expectedHeader:   http.Header{"X-JFrog-Art-Api": {"myJFrogTestToken"}},
			expectedErr:      nil,
		},
		{
			testCase: "bad case - failed to download artifact",
			config: sapDownloadArtifactOptions{
				ArtifactoryToken: "myJFrogTestToken",
				HelmChartURL:     "https://my.reporisory.url/deploy-hyperspace-helm/test-helm-chart/test-helm-chart-1.0.0-12345.tgz",
				FromStaging:      true,
			},
			downloadErr:      fmt.Errorf("some error"),
			expectedFilename: "",
			expectedURL:      "",
			expectedErr:      fmt.Errorf("failed to download artifact: some error"),
		},
	}

	for _, test := range tt {
		test := test
		t.Run(test.testCase, func(t *testing.T) {
			t.Parallel()
			utils := newSapDownloadArtifactTestsUtils()
			utils.downloadError = test.downloadErr
			err := helmDownloadArtifact(&test.config, nil, &utils, &pipelineEnv)
			if test.expectedErr != nil {
				assert.EqualError(t, test.expectedErr, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedFilename, utils.filename)
				assert.Equal(t, test.expectedURL, utils.url)
				assert.Equal(t, test.expectedFilename, pipelineEnv.custom.localHelmChartPath)
				assert.Equal(t, test.expectedHeader, utils.header)
			}
		})
	}
}

func TestGradleDownloadArtifact(t *testing.T) {
	t.Parallel()
	pipelineEnv := sapDownloadArtifactCommonPipelineEnvironment{}
	tt := []struct {
		name             string
		config           sapDownloadArtifactOptions
		downloadErr      error
		expectedFilename string
		expectedURL      string
		expectedHeader   http.Header
		expectedErr      error
	}{
		{
			name: "good case - from Staging",
			config: sapDownloadArtifactOptions{
				Artifacts:                        []map[string]interface{}{{"id": "gradle-app", "name": "gradle-app-1-0.0.1.jar"}, {"id": "gradle-app", "name": "gradle-app-2-0.0.1.jar"}},
				GroupID:                          "com.example",
				ArtifactID:                       "gradle-app",
				Version:                          "0.0.1",
				RepositoryURL:                    "https://my.reporisory.url/stage/repository/uniqueId/",
				StagingServiceRepositoryUser:     "stagingUser",
				StagingServiceRepositoryPassword: "stagingPassword",
				FromStaging:                      true,
			},
			downloadErr:      nil,
			expectedFilename: "gradle-app-2-0.0.1.jar",
			expectedURL:      "https://my.reporisory.url/stage/repository/uniqueId//com/example/gradle-app/0.0.1/gradle-app-2-0.0.1.jar",
			expectedHeader:   http.Header{"Authorization": {"Basic c3RhZ2luZ1VzZXI6c3RhZ2luZ1Bhc3N3b3Jk"}},
			expectedErr:      nil,
		},
		{
			name: "bad case - from Staging",
			config: sapDownloadArtifactOptions{
				FromStaging: true,
			},
			downloadErr:      nil,
			expectedFilename: "",
			expectedURL:      "",
			expectedErr:      fmt.Errorf("there are no artifacts available for downloading"),
		},
		{
			name: "good case - from Promote",
			config: sapDownloadArtifactOptions{
				Artifacts: []map[string]interface{}{{"id": "gradle", "name": "gradle-1-0.0.1.jar"}, {"id": "gradle", "name": "gradle-2-0.0.1.jar"}},
				ArtifactDownloadURLs: []string{
					"https://my.reporisory.url/deploy-hyperspace-maven/com/example/gradle/0.0.1/gradle-1-0.0.1.jar",
					"https://my.reporisory.url/deploy-hyperspace-maven/com/example/gradle/0.0.1/gradle-2-0.0.1.jar",
				},
				ArtifactoryToken: "myJFrogTestToken",
				FromStaging:      false,
			},
			downloadErr:      nil,
			expectedFilename: "gradle-2-0.0.1.jar",
			expectedURL:      "https://my.reporisory.url/deploy-hyperspace-maven/com/example/gradle/0.0.1/gradle-2-0.0.1.jar",
			expectedHeader:   http.Header{"X-JFrog-Art-Api": {"myJFrogTestToken"}},
			expectedErr:      nil,
		},
		{
			name: "bad case - failed to get artifact name",
			config: sapDownloadArtifactOptions{
				Artifacts:   []map[string]interface{}{{"id": "gradle", "namee": "gradle-1-0.0.1.jar"}},
				FromStaging: true,
			},
			expectedErr: fmt.Errorf("failed to get artifact name"),
		},
		{
			name: "bad case - failed to get artifact id",
			config: sapDownloadArtifactOptions{
				Artifacts:   []map[string]interface{}{{"ID": "gradle", "name": "gradle-1-0.0.1.jar"}},
				FromStaging: true,
			},
			expectedErr: fmt.Errorf("failed to get artifact id"),
		},
		{
			name: "bad case - failed to identify artifact URL",
			config: sapDownloadArtifactOptions{
				Artifacts: []map[string]interface{}{{"id": "gradle", "name": "gradle-2-0.0.1.jar"}},
				ArtifactDownloadURLs: []string{
					"https://my.reporisory.url/deploy-hyperspace-maven/com/example/gradle/0.0.1/gradle-1-0.0.1.jar",
				},
				ArtifactoryToken: "myJFrogTestToken",
				FromStaging:      false,
			},
			downloadErr:      nil,
			expectedFilename: "",
			expectedURL:      "",
			expectedErr:      fmt.Errorf("failed to identify artifact url"),
		},
		{
			name: "bad case - failed to download artifact",
			config: sapDownloadArtifactOptions{
				Artifacts: []map[string]interface{}{{"id": "gradle", "name": "gradle-1-0.0.1.jar"}, {"id": "gradle", "name": "gradle-2-0.0.1.jar"}},
				ArtifactDownloadURLs: []string{
					"https://my.reporisory.url/deploy-hyperspace-maven/com/example/gradle/0.0.1/gradle-1-0.0.1.jar",
					"https://my.reporisory.url/deploy-hyperspace-maven/com/example/gradle/0.0.1/gradle-2-0.0.1.jar",
				},
				ArtifactoryToken: "myJFrogTestToken",
				FromStaging:      false,
			},
			downloadErr:      fmt.Errorf("err"),
			expectedFilename: "",
			expectedURL:      "",
			expectedErr:      fmt.Errorf("failed to download artifact: err"),
		},
	}

	for _, test := range tt {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			utils := newSapDownloadArtifactTestsUtils()
			utils.downloadError = test.downloadErr
			err := gradleDownloadArtifact(&test.config, nil, &utils, &pipelineEnv)
			if test.expectedErr != nil {
				assert.EqualError(t, test.expectedErr, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedFilename, utils.filename)
				assert.Equal(t, test.expectedURL, utils.url)
				assert.Equal(t, test.expectedHeader, utils.header)
			}
		})
	}
}

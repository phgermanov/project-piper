//go:build unit
// +build unit

package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/orchestrator"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/stretchr/testify/assert"
)

func resetEnv(e []string) {
	for _, val := range e {
		tmp := strings.Split(val, "=")
		os.Setenv(tmp[0], tmp[1])
	}
}

func TestSapGenerateEnvironmentInfo(t *testing.T) {
	runInTempDir(t, "CPE values are filled", "temp_dir", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("JENKINS_URL", "FOO BAR BAZ")
		os.Setenv("BUILD_URL", "jaas.com/foo/bar/main/42")
		os.Setenv("GIT_BRANCH", "main")
		os.Setenv("GIT_COMMIT", "abcdef42713")
		os.Setenv("GIT_URL", "github.com/foo/bar")

		outPath := filepath.Join("temp_dir", "pogo", "env.json")
		log.Entry().Info(outPath)
		outputBuildSettingsPath := filepath.Join("temp_dir", "pogo", "build-settings.json")
		log.Entry().Info(outputBuildSettingsPath)

		options := sapGenerateEnvironmentInfoOptions{
			OutputPath:              outPath,
			OutputBuildSettingsPath: outputBuildSettingsPath,
			BuildSettingsInfo:       "{\"mavenBuild\":[{\"profiles\":[\"profile1\",\"profile2\"],\"createBOM\":true}]}",
			BuildTool:               "maven",
			ReleaseStatus:           `{"releaseStatus": "promoted"}`,
			GenerateFiles:           []string{"envInfo", "buildSettings", "releaseStatus"},
		}

		sapGenerateEnvironmentInfo(options, nil)

		releaseStatus, _ := filepath.Glob("release-status*")

		assert.FileExists(t, outPath)
		assert.FileExists(t, outputBuildSettingsPath)
		assert.FileExists(t, releaseStatus[0])
	})

	runInTempDir(t, "CPE values are not filled", "temp_dir", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("JENKINS_URL", "FOO BAR BAZ")
		os.Setenv("BUILD_URL", "jaas.com/foo/bar/main/42")
		os.Setenv("GIT_BRANCH", "main")
		os.Setenv("GIT_COMMIT", "abcdef42713")
		os.Setenv("GIT_URL", "github.com/foo/bar")

		outPath := filepath.Join("temp_dir", "pogo", "env.json")
		log.Entry().Info(outPath)
		outputBuildSettingsPath := filepath.Join("temp_dir", "pogo", "build-settings.json")
		log.Entry().Info(outputBuildSettingsPath)

		options := sapGenerateEnvironmentInfoOptions{
			OutputPath:              outPath,
			OutputBuildSettingsPath: outputBuildSettingsPath,
			GenerateFiles:           []string{"envInfo", "buildSettings", "releaseStatus"},
		}

		sapGenerateEnvironmentInfo(options, nil)

		assert.FileExists(t, outPath)
		assert.NoFileExists(t, outputBuildSettingsPath)
	})
}

func TestRunSapGenerateEnvironmentInfo(t *testing.T) {
	runInTempDir(t, "Happy path", "envInfo_temp_dir", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("JENKINS_URL", "FOO BAR BAZ")
		os.Setenv("BUILD_URL", "jaas.com/foo/bar/main/42")
		os.Setenv("GIT_BRANCH", "main")
		os.Setenv("GIT_COMMIT", "abcdef42713")
		os.Setenv("GIT_URL", "github.com/foo/bar")

		outPath := filepath.Join("envInfo_temp_dir", "pogo", "env.json")
		log.Entry().Info(outPath)
		options := sapGenerateEnvironmentInfoOptions{
			OutputPath: outPath,
		}

		err := runSapGenerateEnvironmentInfo(&options)

		assert.Nil(t, err)
		assert.FileExists(t, outPath)
	})

	runInTempDir(t, "Happy path (useCommitIDForCumulus is true)", "envInfo_temp_dir", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("JENKINS_URL", "FOO BAR BAZ")
		os.Setenv("BUILD_URL", "jaas.com/foo/bar/main/42")
		os.Setenv("GIT_BRANCH", "main")
		os.Setenv("GIT_COMMIT", "abcdef42713")
		os.Setenv("GIT_URL", "github.com/foo/bar")

		outPath := filepath.Join("envInfo_temp_dir", "pogo", "env.json")
		log.Entry().Info(outPath)
		options := sapGenerateEnvironmentInfoOptions{
			OutputPath:            outPath,
			UseCommitIDForCumulus: true,
			PipelineOptimization:  true,
		}

		err := runSapGenerateEnvironmentInfo(&options)

		assert.Nil(t, err)
		assert.FileExists(t, outPath)
	})

	runInTempDir(t, "Happy path (useCommitIDForCumulus+Scheduled+Optimized)", "envInfo_temp_dir", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		os.Setenv("JENKINS_URL", "FOO BAR BAZ")
		os.Setenv("BUILD_URL", "jaas.com/foo/bar/main/42")
		os.Setenv("GIT_BRANCH", "main")
		os.Setenv("GIT_COMMIT", "abcdef42713")
		os.Setenv("GIT_URL", "github.com/foo/bar")

		outPath := filepath.Join("envInfo_temp_dir", "pogo", "env.json")
		expectedOutPath := filepath.Join("envInfo_temp_dir", "pogo", "env-scheduled.json")
		log.Entry().Info(outPath)
		options := sapGenerateEnvironmentInfoOptions{
			OutputPath:            outPath,
			UseCommitIDForCumulus: true,
			PipelineOptimization:  true,
			Scheduled:             true,
		}

		err := runSapGenerateEnvironmentInfo(&options)

		assert.Nil(t, err)
		assert.FileExists(t, expectedOutPath)
	})

	runInTempDir(t, "No orchestrator", "envInfo_temp_dir", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		orchestrator.ResetConfigProvider()

		outPath := filepath.Join("envInfo_temp_dir", "pogo", "env.json")
		expectedOutPath := filepath.Join("envInfo_temp_dir", "pogo", "env-scheduled.json")
		log.Entry().Info(outPath)
		options := sapGenerateEnvironmentInfoOptions{
			OutputPath:            outPath,
			UseCommitIDForCumulus: true,
			PipelineOptimization:  true,
			Scheduled:             true,
		}
		err := runSapGenerateEnvironmentInfo(&options)

		assert.Nil(t, err)
		assert.FileExists(t, expectedOutPath)
	})
}

func runInTempDir(t *testing.T, nameOfRun, tempDirPattern string, run func(t *testing.T)) {
	dir, err := os.MkdirTemp("", tempDirPattern)
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}
	oldCWD, _ := os.Getwd()
	_ = os.Chdir(dir)
	// clean up tmp dir
	defer func() {
		_ = os.Chdir(oldCWD)
		_ = os.RemoveAll(dir)
	}()

	t.Run(nameOfRun, run)
}

func TestRunSapGenerateBuildSettings(t *testing.T) {
	runInTempDir(t, "Happy path for buildSettings file", "buildSettings_temp_dir", func(t *testing.T) {

		outputBuildSettingsPath := filepath.Join("buildSettings_temp_dir", "pogo", "build-settings.json")
		log.Entry().Info(outputBuildSettingsPath)
		options := sapGenerateEnvironmentInfoOptions{
			OutputBuildSettingsPath: outputBuildSettingsPath,
			BuildSettingsInfo:       "{\"mavenBuild\":[{\"profiles\":[\"profile1\",\"profile2\"],\"createBOM\":true}]}",
			BuildTool:               "maven",
		}

		err := runSapGenerateBuildSettings(&options)

		assert.Nil(t, err)
		assert.FileExists(t, outputBuildSettingsPath)
	})

	t.Run("No output path", func(t *testing.T) {
		options := sapGenerateEnvironmentInfoOptions{
			BuildSettingsInfo: "{}",
			BuildTool:         "maven",
		}
		err := runSapGenerateBuildSettings(&options)

		assert.EqualError(t, err, "failed to write JSON to given output path: open : no such file or directory")

	})
}

func TestRunSapGenerateReleaseStatus(t *testing.T) {
	runInTempDir(t, "good case", "release_status_temp_dir", func(t *testing.T) {
		err := runSapGenerateReleaseStatus([]byte(`{"releaseStatus":"promoted"}`))
		assert.NoError(t, err)
	})

	runInTempDir(t, "bad case", "release_status_temp_dir", func(t *testing.T) {
		err := runSapGenerateReleaseStatus([]byte(`"incorrectStatus":"promoted"}`))
		assert.EqualError(t, err, "invalid character ':' after top-level value")
	})
}

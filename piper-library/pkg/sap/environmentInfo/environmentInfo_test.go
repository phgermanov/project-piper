//go:build unit
// +build unit

package environmentInfo

import (
	"os"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/orchestrator"

	"github.com/SAP/jenkins-library/pkg/buildsettings"
	"github.com/stretchr/testify/assert"
)

func resetEnv(e []string) {
	for _, val := range e {
		tmp := strings.Split(val, "=")
		os.Setenv(tmp[0], tmp[1])
	}
}

func TestCreateEnvironmentInfo(t *testing.T) {
	t.Run("Azure", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		orchestrator.ResetConfigProvider()

		os.Setenv("BUILD_SOURCEBRANCH", "refs/heads/feat/test-azure")
		os.Setenv("AZURE_HTTP_USER_AGENT", "FOO BAR BAZ")
		os.Setenv("SYSTEM_TEAMFOUNDATIONCOLLECTIONURI", "https://pogo.sap/")
		os.Setenv("SYSTEM_TEAMPROJECT", "foo")
		os.Setenv("SYSTEM_DEFINITIONNAME", "bar")
		os.Setenv("BUILD_BUILDID", "42")
		os.Setenv("BUILD_SOURCEVERSION", "abcdef42713")
		os.Setenv("BUILD_REPOSITORY_URI", "github.com/foo/bar")

		envInfo := EnvironmentInfo{
			BuildEnv:        "Hyperspace_ADO_native_BuildStep",
			PipelineRunMode: "standard",
			Scheduled:       true,
		}
		provider, _ := orchestrator.GetOrchestratorConfigProvider(nil)
		i, err := CreateEnvironmentInfo(envInfo, provider)

		assert.Nil(t, err)
		assert.Equal(t, "https://pogo.sap/foo/bar/_build/results?buildId=42", i.BuildUrl)
		assert.Equal(t, "feat/test-azure", i.GitBranch)
		assert.Equal(t, "abcdef42713", i.GitCommit)
		assert.Equal(t, "github.com/foo/bar", i.GitUrl)
	})

	t.Run("GH Actions", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		orchestrator.ResetConfigProvider()

		os.Setenv("GITHUB_ACTIONS", "true")
		os.Setenv("GITHUB_REF_NAME", "feat/test-gh-actions")
		os.Setenv("GITHUB_RUN_ID", "42")
		os.Setenv("GITHUB_SHA", "abcdef42713")
		os.Setenv("GITHUB_SERVER_URL", "github.com")
		os.Setenv("GITHUB_REPOSITORY", "foo/bar")

		envInfo := EnvironmentInfo{
			BuildEnv:        "Hyperspace_GH_Actions_native_BuildStep",
			PipelineRunMode: "standard",
			Scheduled:       true,
		}
		provider, _ := orchestrator.GetOrchestratorConfigProvider(nil)
		i, err := CreateEnvironmentInfo(envInfo, provider)

		assert.Nil(t, err)
		assert.Equal(t, "github.com/foo/bar/actions/runs/42", i.BuildUrl)
		assert.Equal(t, "feat/test-gh-actions", i.GitBranch)
		assert.Equal(t, "abcdef42713", i.GitCommit)
		assert.Equal(t, "github.com/foo/bar", i.GitUrl)
	})

	t.Run("Jenkins", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		orchestrator.ResetConfigProvider()

		os.Setenv("JENKINS_URL", "FOO BAR BAZ")
		os.Setenv("BUILD_URL", "jaas.com/foo/bar/main/42")
		os.Setenv("BRANCH_NAME", "main")
		os.Setenv("GIT_COMMIT", "abcdef42713")
		os.Setenv("GIT_URL", "github.com/foo/bar")

		envInfo := EnvironmentInfo{
			BuildEnv:        "Hyperspace_Jenkins_native_BuildStep",
			PipelineRunMode: "standard",
			Scheduled:       true,
		}
		provider, _ := orchestrator.GetOrchestratorConfigProvider(nil)
		i, err := CreateEnvironmentInfo(envInfo, provider)

		assert.Nil(t, err)
		assert.Equal(t, "jaas.com/foo/bar/main/42", i.BuildUrl)
		assert.Equal(t, "main", i.GitBranch)
		assert.Equal(t, "abcdef42713", i.GitCommit)
		assert.Equal(t, "github.com/foo/bar", i.GitUrl)

		json, _ := i.ToJson()
		assert.Contains(t, string(json), "abcdef42713")
	})

	t.Run("No orchestrator", func(t *testing.T) {
		defer resetEnv(os.Environ())
		os.Clearenv()
		orchestrator.ResetConfigProvider()

		envInfo := EnvironmentInfo{
			BuildEnv:        "",
			PipelineRunMode: "standard",
			Scheduled:       false,
		}
		provider, _ := orchestrator.GetOrchestratorConfigProvider(nil)
		_, err := CreateEnvironmentInfo(envInfo, provider)

		assert.NoError(t, err)
	})
}

func TestCreateOrchestratorAgnosticBuildSettingsInfo(t *testing.T) {
	t.Parallel()

	testTableConfig := []struct {
		buildSettingsInfo string
		expected          []buildsettings.BuildOptions
	}{
		{
			buildSettingsInfo: "{\"mavenBuild\":[{\"profiles\":[\"profile1\",\"profile2\"],\"createBOM\":true}]}",
			expected:          []buildsettings.BuildOptions{{Profiles: []string{"profile1", "profile2"}, CreateBOM: true}},
		},
		{
			buildSettingsInfo: "{\"npmExecuteScripts\":[{\"profiles\":[\"profile1\",\"profile2\"],\"createBOM\":true}]}",
			expected:          []buildsettings.BuildOptions{{Profiles: []string{"profile1", "profile2"}, CreateBOM: true}},
		},
		{
			buildSettingsInfo: "{\"kanikoExecute\":[{\"dockerImage\":\"kaniko:0.1\"}]}",
			expected:          []buildsettings.BuildOptions{{DockerImage: "kaniko:0.1"}},
		},
		{
			buildSettingsInfo: "{\"mtaBuild\":[{\"profiles\":[\"release.build\"],\"publish\":true,\"globalSettingsFile\":\"http://nexus.test:8081/nexus/\"}]}",
			expected:          []buildsettings.BuildOptions{{Profiles: []string{"release.build"}, Publish: true, GlobalSettingsFile: "http://nexus.test:8081/nexus/"}},
		},
		{
			buildSettingsInfo: "{\"golangBuild\":[]}",
			expected:          []buildsettings.BuildOptions{},
		},
		{
			buildSettingsInfo: "{\"mavenBuild\":[{\"profiles\":[\"profile1\",\"profile2\"],\"createBOM\":true}],\"kanikoExecute\":[{\"dockerImage\":\"kaniko:0.1\"}]}",
			expected:          []buildsettings.BuildOptions{{Profiles: []string{"profile1", "profile2"}, CreateBOM: true}, {DockerImage: "kaniko:0.1"}},
		},
	}

	for _, testCase := range testTableConfig {
		info, err := CreateOrchestratorAgnosticBuildSettingsInfo(testCase.buildSettingsInfo)
		assert.Nil(t, err)
		if len(testCase.expected) > 1 {
			for _, bOptions := range testCase.expected {
				assert.Contains(t, info, bOptions)
			}
		} else {
			assert.Equal(t, testCase.expected, info)
		}
	}
}

package environmentInfo

import (
	"encoding/json"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/buildsettings"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/pkg/errors"
)

type EnvironmentInfo struct {
	BuildUrl        string `json:"BUILD_URL"`
	GitBranch       string `json:"GIT_BRANCH"`
	GitCommit       string `json:"GIT_COMMIT"`
	GitTagCommit    string `json:"GIT_TAG_COMMIT,omitempty"`
	GitUrl          string `json:"GIT_URL"`
	BuildEnv        string `json:"BUILD_ENV"`
	Scheduled       bool   `json:"isScheduled"`
	PipelineRunMode string `json:"PIPELINE_RUN_MODE"`
}

func (i *EnvironmentInfo) ToJson() ([]byte, error) {
	tmp, err := json.Marshal(i)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate valid JSON.")
	} else {
		return tmp, nil
	}
}

func CreateEnvironmentInfo(envInfo EnvironmentInfo, provider orchestrator.ConfigProvider) (EnvironmentInfo, error) {
	envInfo.BuildUrl = provider.BuildURL()
	if len(envInfo.GitBranch) == 0 {
		envInfo.GitBranch = provider.Branch()
	}
	if len(envInfo.GitUrl) == 0 {
		envInfo.GitUrl = provider.RepoURL()
	}
	if len(envInfo.GitCommit) == 0 {
		envInfo.GitCommit = provider.CommitSHA()
	}
	return envInfo, nil
}

func CreateOrchestratorAgnosticBuildSettingsInfo(buildSettingsInfo string) ([]buildsettings.BuildOptions, error) {
	dataBuildSettings := map[string][]buildsettings.BuildOptions{}
	err := json.Unmarshal([]byte(buildSettingsInfo), &dataBuildSettings)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal build options: %w", err)
	}

	combinedBuildOptions := []buildsettings.BuildOptions{}
	for _, opts := range dataBuildSettings {
		combinedBuildOptions = append(combinedBuildOptions, opts...)
	}

	return combinedBuildOptions, nil
}

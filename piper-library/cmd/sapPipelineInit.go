package cmd

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/cmd"
	jlConfig "github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/github"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/go-git/go-git/v5"
	"gopkg.in/yaml.v3"

	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/events"
)

func sapPipelineInit(config sapPipelineInitOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *sapPipelineInitCommonPipelineEnvironment, influx *sapPipelineInitInflux) {
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runSapPipelineInit(&config, telemetryData, commonPipelineEnvironment, influx)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}

}

func runSapPipelineInit(config *sapPipelineInitOptions, telemetryData *telemetry.CustomData, pipelineEnv *sapPipelineInitCommonPipelineEnvironment, influx *sapPipelineInitInflux) error {
	// collect provider values
	provider, err := orchestrator.GetOrchestratorConfigProvider(nil)
	if err != nil {
		log.Entry().WithError(err).Warning("Cannot infer config from CI environment")
	}

	var buildURL = provider.BuildURL()
	if len(config.BuildURL) > 0 {
		buildURL = config.BuildURL
	}

	// Telemetry
	telemetryData.IsScheduled = config.IsScheduled
	telemetryData.IsOptimized = config.PipelineOptimization
	// Pipeline Environment
	pipelineEnv.custom.scheduledRun = config.IsScheduled
	log.Entry().Infof("Pipeline is scheduled: %v", config.IsScheduled)
	pipelineEnv.custom.isOptimizedAndScheduled = config.PipelineOptimization && config.IsScheduled
	log.Entry().Infof("Pipeline optimized and scheduled: %v", config.PipelineOptimization && config.IsScheduled)

	var cumulusPipelineID string
	if config.PipelineID != "" {
		cumulusPipelineID = config.PipelineID
	} else {
		cumulusPipelineID = "N/A"
	}
	log.Entry().Infof("gcsBucketID/cumulusPipelineID: %v", cumulusPipelineID)
	pipelineEnv.custom.cumulusPipelineID = cumulusPipelineID
	pipelineEnv.custom.gcsBucketID = cumulusPipelineID // For future use-cases we will use gcsBucketID

	// call the function to get GH statistics and store the values
	gitInfo, err := getGitDetails()
	if err != nil {
		log.Entry().WithError(err).Warning("could not get git details")
	}

	// Get the branch via orchestrator package
	gitInfo.gitBranch = provider.Branch()

	var isProductiveBranch bool
	if config.ProductiveBranch != "" {
		isProductiveBranch, err = regexp.MatchString(config.ProductiveBranch, gitInfo.gitBranch)
		if err != nil {
			return fmt.Errorf("invalid regexp: %w", err)
		}
		pipelineEnv.custom.isProductiveBranch = isProductiveBranch
		log.Entry().Debugf("Running on productive branch: %v, branch: %v", isProductiveBranch, gitInfo.gitBranch)
	} else {
		log.Entry().Warn("could not get productive branch")
		pipelineEnv.custom.isProductiveBranch = false
	}

	statistics, err := getGitHubStatistics(config, err, gitInfo)
	if err != nil {
		log.Entry().Debugf("couldn't get GitHub statistics: %s", err)
	}

	// set initial event data for gcpPublishEvent in pipeline environment
	eventData, err := events.CreateEventData(provider, config.UseCommitIDForCumulus, config.PipelineOptimization, config.IsScheduled, config.PipelineID, gitInfo.gitCommitID, gitInfo.gitURL).ToJSON()
	if err != nil {
		log.Entry().Debugf("couldn't get event data: %s", err)
	} else {
		log.Entry().Debugf("exported event data: %s", eventData)
	}

	// set git-related pipeline environment variables
	pipelineEnv.git.url = gitInfo.gitURL
	pipelineEnv.git.instance = gitInfo.gitInstance
	pipelineEnv.git.organization = gitInfo.gitOrganization
	pipelineEnv.git.repository = gitInfo.gitRepository
	pipelineEnv.git.branch = gitInfo.gitBranch
	pipelineEnv.git.commitID = gitInfo.gitCommitID
	pipelineEnv.git.headCommitID = gitInfo.gitCommitID
	pipelineEnv.git.github_changes = statistics.Total
	pipelineEnv.git.github_additions = statistics.Additions
	pipelineEnv.git.github_deletions = statistics.Deletions
	pipelineEnv.git.github_filesChanged = statistics.Files
	pipelineEnv.custom.eventData = string(eventData)

	// Influx
	influx.jenkins_custom_data.fields.build_result = "SUCCESS"
	influx.jenkins_custom_data.fields.build_result_key = 1
	influx.pipeline_data.fields.build_url = buildURL
	influx.step_data.fields.build_url = buildURL
	influx.pipeline_data.fields.github_changes = statistics.Total
	influx.pipeline_data.fields.github_additions = statistics.Additions
	influx.pipeline_data.fields.github_deletions = statistics.Deletions
	influx.pipeline_data.fields.github_filesChanged = statistics.Files
	influx.pipeline_data.fields.cumulusPipelineID = cumulusPipelineID

	piperConfig, err := generatePiperConfig()
	if err != nil {
		return fmt.Errorf("failed to generate piper config: %w", err)
	}
	//Currently the piper config consists only of values needs from ppms i.e the ppms from sapCheckPPMSCompliance and sapCheckECCNCompliance
	// in future if we want to extend for other steps please keep in mind the OS env variable size limit may break CPE
	pipelineEnv.custom.piperConfig = piperConfig

	return nil
}

func getGitHubStatistics(config *sapPipelineInitOptions, err error, gitInfo gitDetails) (github.FetchCommitResult, error) {
	var statistics github.FetchCommitResult
	if err == nil && (gitInfo != gitDetails{}) && (config.GithubToken != "") {

		// collecting GitHub statistics, only works if gitInfo is not empty
		var gitAPI string
		if config.GithubAPIURL != "" {
			gitAPI = config.GithubAPIURL
		} else {
			gitAPI = gitInfo.gitApiURL
		}
		log.Entry().Debugf("fetching GitHub statistics for latest commit from %s", gitAPI)
		statistics, err = github.FetchCommitStatistics(
			&github.FetchCommitOptions{
				Owner:      gitInfo.gitOrganization,
				Repository: gitInfo.gitRepository,
				APIURL:     gitAPI,
				Token:      config.GithubToken,
				SHA:        gitInfo.gitCommitID,
			},
		)
		if err != nil {
			log.Entry().WithError(err).Warning("failed to fetch GitHub statistics")
			statistics = github.FetchCommitResult{}
		}
	}
	return statistics, err
}

func generatePiperConfig() (string, error) {
	// saving config files and merging them into 1 file. Hardcoded.
	cmd.SetConfigOptions(cmd.ConfigCommandOptions{
		Output:     "yaml",
		OutputFile: "sapCheckPPMSCompliance-configs.yaml",
		OpenFile:   jlConfig.OpenPiperFile,
		StepName:   "sapCheckPPMSCompliance",
	})
	err := cmd.GenerateConfig(nil, jlConfig.GetYAML)
	if err != nil {
		return "", fmt.Errorf("sapCheckPPMSCompliance: %v", err)
	}

	cmd.SetConfigOptions(cmd.ConfigCommandOptions{
		Output:     "yaml",
		OutputFile: "sapCheckECCNCompliance-configs.yaml",
		OpenFile:   jlConfig.OpenPiperFile,
		StepName:   "sapCheckECCNCompliance",
	})
	err = cmd.GenerateConfig(nil, jlConfig.GetYAML)
	if err != nil {
		return "", fmt.Errorf("sapCheckECCNCompliance: %v", err)
	}

	piperConfig, err := hardcodedMergeYamlFiles("sapCheckPPMSCompliance-configs.yaml", "sapCheckECCNCompliance-configs.yaml", "piper-config.yaml")
	if err != nil {
		return "", err
	}

	return piperConfig, nil
}

func hardcodedMergeYamlFiles(file1Path, file2Path, outputPath string) (string, error) {
	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2)

	file1Data, err := os.ReadFile(filepath.Clean(file1Path))
	if err != nil {
		return "", err
	}

	file2Data, err := os.ReadFile(filepath.Clean(file2Path))
	if err != nil {
		return "", err
	}

	type yamlLayout map[interface{}]interface{}

	var file1Map yamlLayout
	var file2Map yamlLayout

	if err := yaml.Unmarshal(file1Data, &file1Map); err != nil {
		return "", err
	}

	if err := yaml.Unmarshal(file2Data, &file2Map); err != nil {
		return "", err
	}

	ppmsID1 := ""
	ppmsID2 := ""
	switch v := file1Map["ppmsID"].(type) {
	case string:
		ppmsID1 = v
	}

	switch v := file2Map["ppmsID"].(type) {
	case string:
		ppmsID2 = v
	}

	var file3Map = yamlLayout{
		"steps": yamlLayout{
			"sapCheckPPMSCompliance": yamlLayout{
				"ppmsID": ppmsID1,
			},
			"sapCheckECCNCompliance": yamlLayout{
				"ppmsID": ppmsID2,
			},
		},
	}

	err = yamlEncoder.Encode(&file3Map)
	if err != nil {
		return "", nil
	}

	// err = ioutil.WriteFile(outputPath, b.Bytes(), 0644)
	// if err != nil {
	// 	return "",err
	// }

	return b.String(), nil
}

type gitDetails struct {
	gitBranch, gitInstance, gitURL, gitOrganization, gitRepository, gitCommitID, gitApiURL string
}

func getGitDetails() (gitDetails, error) {
	//Get information about GH statistics
	var gitInfo gitDetails

	repository, err := git.PlainOpen(".")
	if err != nil {
		log.Entry().WithError(err).Warning("could not read git repo")
		return gitInfo, err
	}
	//get git config
	cfg, err := repository.Config()
	if err != nil {
		log.Entry().WithError(err).Warning("could not read git config")
		return gitInfo, err
	}
	// retrieves the HEAD
	head, err := repository.Head()
	if err != nil {
		log.Entry().WithError(err).Warning("could not read git head")
		return gitInfo, err
	}
	// Git URL
	origin := cfg.Remotes["origin"]
	if origin != nil {
		gitInfo.gitURL = origin.URLs[0]
	} else {
		log.Entry().WithError(err).Warning("could not get gitURL from local repository")
		return gitInfo, err
	}
	// checks if it is cloned via ssh and modifies the url accordingly
	if strings.Contains(gitInfo.gitURL, "git@") {
		gitInfo.gitURL = strings.Replace(gitInfo.gitURL, ":", "/", 1)
		gitInfo.gitURL = strings.Replace(gitInfo.gitURL, "git@", "https://", 1)
	}
	gitInfo.gitURL = strings.TrimSuffix(gitInfo.gitURL, ".git")
	url, err := url.Parse(gitInfo.gitURL)
	if err != nil {
		log.Entry().WithError(err).Warning("could not parse URL")
		return gitInfo, err
	}
	gitInfo.gitInstance = url.Hostname()
	gitInfo.gitOrganization = strings.ReplaceAll(filepath.Dir(url.Path), "/", "")
	gitInfo.gitRepository = filepath.Base(url.Path)
	gitInfo.gitBranch = strings.TrimPrefix(string(head.Name()), "refs/heads/")
	gitInfo.gitCommitID = head.Hash().String()

	switch gitInfo.gitInstance {
	case "github.com":
		gitInfo.gitApiURL = "https://api.github.com/"
	case "github.wdf.sap.corp":
		gitInfo.gitApiURL = "https://github.wdf.sap.corp/api/v3/"
	case "github.tools.sap":
		gitInfo.gitApiURL = "https://github.tools.sap/api/v3/"
	default:
		gitInfo.gitApiURL = "https://github.wdf.sap.corp/api/v3/"
	}

	log.Entry().Debug("found the following git infos from the local repository:", gitInfo)
	return gitInfo, err
}

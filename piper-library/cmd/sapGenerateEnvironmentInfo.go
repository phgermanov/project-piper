package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/cumulus"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/environmentInfo"
)

func sapGenerateEnvironmentInfo(options sapGenerateEnvironmentInfoOptions, telemetryData *telemetry.CustomData) {
	for _, file := range options.GenerateFiles {
		switch file {
		case "envInfo":
			if err := runSapGenerateEnvironmentInfo(&options); err != nil {
				log.Entry().WithError(err).Fatal("Failed to generate SLC-25 compliance information.")
			}
		case "buildSettings":
			if err := runSapGenerateBuildSettings(&options); err != nil {
				log.Entry().WithError(err).Fatal("Failed to generate SLC-29 compliance information.")
			}
		case "releaseStatus":
			if len(options.ReleaseStatus) > 0 {
				if err := runSapGenerateReleaseStatus([]byte(options.ReleaseStatus)); err != nil {
					log.Entry().WithError(err).Fatal("Failed to generate release status file")
				}
			} else {
				log.Entry().Info("empty releaseStatus: skipping to write release status file")
			}
		case "piperConfig":
			if err := runSapGeneratePiperConfig(&options); err != nil {
				log.Entry().WithError(err).Fatal("Failed to generate piper config file")
			}

		}
	}
}

func runSapGeneratePiperConfig(options *sapGenerateEnvironmentInfoOptions) error {
	if len(options.PiperConfig) > 0 {
		err := os.MkdirAll(filepath.Dir(options.OutputPiperConfigPath), 0750)
		if err != nil {
			return errors.Wrap(err, "failed to create directory with given path")
		}

		err = os.WriteFile(options.OutputPiperConfigPath, []byte(options.PiperConfig), 0600)
		if err != nil {
			return errors.Wrap(err, "failed to write piper config to given output path")
		}

		log.Entry().Infof("Generated piper config for PPMS information in file: %s", options.OutputPiperConfigPath)
	} else {
		log.Entry().Info("empty piperConfig: skipping to write piper config file")
	}
	return nil
}

func runSapGenerateEnvironmentInfo(options *sapGenerateEnvironmentInfoOptions) error {
	provider, _ := orchestrator.GetOrchestratorConfigProvider(nil)
	envInfo := environmentInfo.EnvironmentInfo{
		BuildEnv:        options.BuildEnvironment,
		PipelineRunMode: cumulus.GetPipelineRunMode(provider.IsPullRequest(), options.UseCommitIDForCumulus, options.PipelineOptimization, options.Scheduled),
		Scheduled:       options.Scheduled,
		GitBranch:       options.GitBranch,
		GitCommit:       getCommitID(options),
		GitUrl:          options.GitHttpsURL,
	}

	// fill out the gitTagCommitId that is created if the versioning is cloud and a tag is created
	// from artifactPrepareVersion for versioning type cloud only: needed for DMS (PPMS)
	// only created with both headCommit and gitCommit are filled and are not equal
	if len(options.GitCommitID) > 0 && len(options.GitHeadCommitID) > 0 && options.GitCommitID != options.GitHeadCommitID {
		log.Entry().Debugf("Filling in tag commit ID %v ", options.GitCommitID)
		envInfo.GitTagCommit = options.GitCommitID
	}

	// Rely on Orchestrator to fill in values for env.json if above code is not successful.
	info, err := environmentInfo.CreateEnvironmentInfo(envInfo, provider)
	if err != nil {
		return err // Don't wrap error here. Created one level below
	}

	json, err := info.ToJson()
	log.Entry().Debugf("env.json contents %v", string(json))
	if err != nil {
		return err // Don't wrap error here. Already wrapped one level below
	}

	suffix := cumulus.GetPipelineRunModeSuffix(options.UseCommitIDForCumulus, options.PipelineOptimization, options.Scheduled)
	if suffix != "" {
		ext := filepath.Ext(options.OutputPath)
		fileName := fmt.Sprintf("%s-%s", strings.TrimSuffix(options.OutputPath, ext), suffix)
		options.OutputPath = fmt.Sprintf("%s%s", fileName, ext)
	}

	err = os.MkdirAll(filepath.Dir(options.OutputPath), 0750)
	if err != nil {
		return errors.Wrap(err, "failed to create output folder for JSON")
	}

	err = os.WriteFile(options.OutputPath, json, 0600)
	if err != nil {
		return errors.Wrap(err, "failed to write JSON to given output path")
	}

	log.Entry().Infof("Generated SLC-25 compliance information in file: %s", options.OutputPath)

	return nil
}

func getCommitID(options *sapGenerateEnvironmentInfoOptions) string {
	var envCommit string
	// Always fill env.json GitCommit field with head commit id
	// If not available, fill in commit id as a fallback
	if len(options.GitHeadCommitID) != 0 {
		envCommit = options.GitHeadCommitID
		log.Entry().Debugf("Head commit ID %v", options.GitHeadCommitID)
	} else {
		envCommit = options.GitCommitID
		log.Entry().Debugf("Commit ID %v", envCommit)
	}
	return envCommit
}

func runSapGenerateBuildSettings(options *sapGenerateEnvironmentInfoOptions) error {
	if len(options.BuildSettingsInfo) > 0 && len(options.BuildTool) > 0 {
		info, err := environmentInfo.CreateOrchestratorAgnosticBuildSettingsInfo(options.BuildSettingsInfo)
		if err != nil {
			return err // Don't wrap error here. Created one level below
		}

		json, err := json.Marshal(info)
		if err != nil {
			return errors.Wrap(err, "failed to generate valid JSON.")
		}

		os.MkdirAll(filepath.Dir(options.OutputBuildSettingsPath), 0750)
		err = os.WriteFile(options.OutputBuildSettingsPath, json, 0600)
		if err != nil {
			return errors.Wrap(err, "failed to write JSON to given output path")
		}

		log.Entry().Infof("Generated SLC-29 compliance information in file: %s", options.OutputBuildSettingsPath)
	} else {
		log.Entry().Info("empty buildSettingsInfo: skipping to write build settings file")
	}
	return nil
}

func runSapGenerateReleaseStatus(data []byte) error {
	var status cumulus.ReleaseStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return err
	}

	log.Entry().Info("Generating release status file")
	return status.ToFile(piperutils.Files{}, time.Now())
}

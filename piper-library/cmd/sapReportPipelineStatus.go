package cmd

import (
	"encoding/json"

	piperOsCmd "github.com/SAP/jenkins-library/cmd"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/reportpipelinestatus"
)

const PIPELINE_LOG_NAME = "pipelineLog"

type sapReportPipelineStatusUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The sapReportPipelineStatusUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type sapReportPipelineStatusUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to sapReportPipelineStatusUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// sapReportPipelineStatusUtilsBundle and forward to the implementation of the dependency.
}

func newSapReportPipelineStatusUtils() sapReportPipelineStatusUtils {
	utils := sapReportPipelineStatusUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func sapReportPipelineStatus(config sapReportPipelineStatusOptions, telemetryData *telemetry.CustomData, pipelineEnv *sapReportPipelineStatusCommonPipelineEnvironment) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newSapReportPipelineStatusUtils()

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runSapReportPipelineStatus(&config, pipelineEnv, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runSapReportPipelineStatus(config *sapReportPipelineStatusOptions, pipelineEnv *sapReportPipelineStatusCommonPipelineEnvironment, telemetryData *telemetry.CustomData, utils sapReportPipelineStatusUtils) error {
	var provider orchestrator.ConfigProvider
	var err error

	provider, err = orchestrator.GetOrchestratorConfigProvider(&orchestrator.Options{
		JenkinsUsername: config.JenkinsUser,
		JenkinsToken:    config.JenkinsToken,
		AzureToken:      config.AzureToken,
		GitHubToken:     config.GitHubToken,
	})
	if err != nil {
		log.Entry().Error(err)
		provider = &orchestrator.UnknownOrchestratorConfigProvider{}
	}

	// ---------------------------------------- Get Log File ----------------------------------
	// Get whole error logFile from orchestrator and persist it to local file system
	logFile, err := provider.FullLogs()
	gotLogFile := true // helper variable to estimate how many users have not maintained the orchestrator credentials
	if err != nil || string(logFile) == "" {
		// Error occurred
		log.Entry().WithError(err).Error("could not get log file, returning empty log")
		log.SetErrorCategory(log.ErrorInfrastructure)
		logFile = []byte{}
		gotLogFile = false
	}

	pipelineLog := ""
	if config.UseCommitIDForCumulus && config.IsOptimizedAndScheduled {
		pipelineLog = PIPELINE_LOG_NAME + "-scheduled.log"
	} else {
		pipelineLog = PIPELINE_LOG_NAME + ".log"
	}

	if err == nil {
		// No error occurred persisting log file to disk
		reportpipelinestatus.WriteLogFile(&logFile, pipelineLog)
	}
	// --------------------------------------- End Get Log File ----------------------------------

	errorDetails, pipelineStatus, stepDetails := reportpipelinestatus.ParseLogFile(&logFile)
	switch provider.OrchestratorType() {
	case "Jenkins", "GitHubActions":
		pipelineStatus = provider.BuildStatus()
	case "Azure":
		// Azure we determine the pipeline status based on the failures found in the logfile
	}

	var sendLogs bool
	if reportpipelinestatus.CategoryInList([]string{"all"}, config.SendErrorLogs) {
		log.Entry().Debugf("Sending logs for all error categories.")
		// TODO sends whole logfile even if no error happened!
		// FIX ME
		sendLogs = true
	} else {
		sendLogs = reportpipelinestatus.SendLogs(config.SendErrorLogs, errorDetails)
	}

	scheduledBuild := false
	if provider.BuildReason() == orchestrator.BuildReasonSchedule {
		scheduledBuild = true
	}

	pipelineData := &reportpipelinestatus.PipelineTelemetry{
		CorrelationId:           piperOsCmd.GeneralConfig.CorrelationID,
		Duration:                reportpipelinestatus.CalcDuration(provider.PipelineStartTime()), // Duration in Milliseconds since start of the pipeline
		Orchestrator:            provider.OrchestratorType(),
		OrchestratorVersion:     provider.OrchestratorVersion(),
		PipelineStartTime:       provider.PipelineStartTime().String(),
		BuildId:                 provider.BuildID(),
		JobName:                 provider.JobName(),
		PipelineStatus:          pipelineStatus,
		GitInstance:             config.GitInstance,
		JobURL:                  provider.JobURL(),
		CommitHash:              config.CommitID,
		Branch:                  config.Branch,
		GitOrganization:         config.GitOrganization,
		GitRepository:           config.GitRepository,
		GitUrl:                  config.GitURL,
		StepData:                stepDetails,
		ArtifactVersion:         config.ArtifactVersion,
		CumulusPipelineId:       config.GcsBucketID,
		BuildReason:             provider.BuildReason(),
		ScheduledBuild:          scheduledBuild,
		IsScheduledAndOptimized: config.IsOptimizedAndScheduled,
		IsProductiveBranch:      config.IsProductiveBranch,
		OrchestratorCredentials: gotLogFile,
	}

	prettyPipelineData, err := json.MarshalIndent(pipelineData, "", "    ")
	if err != nil {
		log.Entry().WithError(err).Warn("failed to generate json")
		prettyPipelineData = nil
	}
	log.Entry().Debugf("Collected pipeline data: %s", prettyPipelineData)

	eventBytes, err := reportpipelinestatus.SetEventOutcomeAndErrors(config.EventData, pipelineStatus, gotLogFile, provider.OrchestratorType())
	if err != nil {
		log.Entry().WithError(err).Warn("failed to set pipeline status and errors in event data")
	} else {
		pipelineEnv.custom.eventData = string(eventBytes)
		log.Entry().Debugf("Exported event data: %s", string(eventBytes))
	}

	if config.TelemetryReporting {
		log.Entry().Info("Telemetry reporting is active.")
		// config.TelemetryReporting should indicate if telemetry reporting is desired
		pipelineTelemetryData := telemetry.CustomData{}
		pipelineTelemetryData.Duration = pipelineData.Duration
		pipelineTelemetryData.PiperCommitHash = piperOsCmd.GitCommit

		if len(errorDetails) != 0 {
			// There are errors in the map, we return the first error we get
			pipelineTelemetryData.ErrorCategory = errorDetails[0]["Category"].(string)
			pipelineTelemetryData.ErrorCode = "1"
			// workaround - set empty error field to avoid error, needs to be fixed in OS piper pkg/telemetry
			details := logrus.Fields{}
			errDetails, _ := json.Marshal(&details)
			log.SetFatalErrorDetail(errDetails)

		} else {
			pipelineTelemetryData.ErrorCategory = ""
			pipelineTelemetryData.ErrorCode = "0"
		}

		pipelineTelemetryClient := &telemetry.Telemetry{}
		pipelineTelemetryClient.Initialize("pipelineTelemetry")
		pipelineTelemetryClient.SetData(&pipelineTelemetryData)
		pipelineTelemetryClient.LogStepTelemetryData()
	}

	if config.SplunkReporting {
		log.Entry().Infof("Splunk reporting is active.")
		var pipelineDataMap map[string]interface{}
		marshal, err := json.Marshal(pipelineData)
		if err != nil {
			log.Entry().WithError(err).Error("could not marshal determinedErrors")
		}
		err = json.Unmarshal(marshal, &pipelineDataMap)
		if err != nil {
			log.Entry().WithError(err).Error("could not unmarshal errorMap")
		}
		pipelineSplunkClient := &splunk.Splunk{}
		_ = pipelineSplunkClient.Initialize(piperOsCmd.GeneralConfig.CorrelationID, config.SplunkDsn, config.SplunkToken, config.SplunkIndex, sendLogs)
		_ = pipelineSplunkClient.SendPipelineStatus(pipelineDataMap, &logFile)

	}

	if !gotLogFile {
		log.Entry().Warn("The pipeline status couldn't be detected properly. This means that the pipeline status in Sirius will show up as " +
			"\"failure\". Please check your orchestrator credentials for this step.")
	}
	return nil
}

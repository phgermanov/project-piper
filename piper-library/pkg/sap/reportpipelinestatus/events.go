package reportpipelinestatus

import (
	"errors"

	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/events"
)

// SetEventOutcomeAndErrors adds the pipeline outcome and hardcoded errors to eventData for gcpPublishEvent
// the outcome should be "error" when there has been a pipeline-breaking error
// with non-fatal errors, the outcome should be "failure"
func SetEventOutcomeAndErrors(eventDataString, pipelineStatus string, gotLogFile bool, orchestrator string) ([]byte, error) {
	eventData, err := events.FromJSON([]byte(eventDataString))
	if err != nil {
		return []byte{}, errors.New("failed to unmarshal event data")
	}

	// if the log couldn't be retrieved, the provided credentials were invalid
	// in that case, on ADO and GHA the status couldn't be determined from the logs, and on Jenkins the status couldn't be retrieved from the API
	if !gotLogFile {
		pipelineStatus = "unknown"
	}

	eventOutcome, eventErrors := determineOutcomeAndErrors(pipelineStatus)
	eventData.Outcome = eventOutcome
	if eventErrors != nil {
		eventData.Errors = eventErrors
	}

	eventBytes, err := eventData.ToJSON()
	if err != nil {
		return []byte{}, errors.New("failed to marshal event data")
	}
	return eventBytes, nil
}

func determineOutcomeAndErrors(pipelineStatus string) (string, []string) {
	var pipelineOutcome string
	var pipelineErrors []string

	switch status := pipelineStatus; status {
	case orchestrator.BuildStatusInProgress:
		fallthrough
	case orchestrator.BuildStatusSuccess:
		pipelineOutcome = events.OutcomeSuccess
		pipelineErrors = nil
	case orchestrator.BuildStatusFailure:
		pipelineOutcome = events.OutcomeError
		pipelineErrors = []string{"Check the pipeline logs for errors."}
	// this case can currently only occur on Jenkins, as other orchestrators can't detect aborted status
	case orchestrator.BuildStatusAborted:
		pipelineOutcome = events.OutcomeAborted
		pipelineErrors = []string{"The pipeline has been aborted."}
	case "unknown":
		pipelineOutcome = events.OutcomeFailure
		pipelineErrors = []string{"The sapReportPipelineStatus step failed and couldn't determine the pipeline status properly. Please check if the orchestrator credentials for the step are valid."}
	default:
		pipelineOutcome = events.OutcomeFailure
		pipelineErrors = []string{"The sapReportPipelineStatus step couldn't determine the pipeline status properly."}
	}
	return pipelineOutcome, pipelineErrors
}

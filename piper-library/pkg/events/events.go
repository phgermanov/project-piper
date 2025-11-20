package events

import (
	"encoding/json"

	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/pkg/errors"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/cumulus"
)

// EventData as described here: https://github.tools.sap/hyperspace/event-bus
const (
	OutcomeSuccess = "success"
	OutcomeFailure = "failure"
	OutcomeError   = "error"
	OutcomeAborted = "aborted"
)

type EventData struct {
	PipelineRunId      string             `json:"pipelineRunId"`
	Url                string             `json:"url"`
	CommitId           string             `json:"commitId"`
	RepositoryUrl      string             `json:"repositoryUrl"`
	PipelineRunMode    string             `json:"pipelineRunMode"`
	CumulusInformation cumulusInformation `json:"cumulusInformation"`
	Outcome            string             `json:"outcome,omitempty"`
	Errors             []string           `json:"errors,omitempty"`
}

type cumulusInformation struct {
	PipelineId string `json:"pipelineId"`
	RunId      string `json:"pipelineRunKey,omitempty"`
}

func FromJSON(b []byte) (EventData, error) {
	eventData := EventData{}
	err := json.Unmarshal(b, &eventData)
	if err != nil {
		return EventData{}, errors.Wrapf(err, "failed to unmarshal JSON string to event data")
	}
	return eventData, nil
}

func (e EventData) ToJSON() ([]byte, error) {
	eventDataJSON, err := json.Marshal(e)
	if err != nil {
		return []byte{}, errors.Wrapf(err, "failed to marshal event data to JSON")
	}
	return eventDataJSON, nil
}

func CreateEventData(provider orchestrator.ConfigProvider, useCommitID bool, pipelineOptimization, isScheduled bool, pipelineID, commitID, repoURL string) EventData {
	return EventData{
		PipelineRunId:   provider.BuildURL(),
		Url:             provider.BuildURL(),
		CommitId:        commitID,
		RepositoryUrl:   repoURL,
		PipelineRunMode: cumulus.GetPipelineRunMode(provider.IsPullRequest(), useCommitID, pipelineOptimization, isScheduled),
		CumulusInformation: cumulusInformation{
			PipelineId: pipelineID,
		},
	}
}

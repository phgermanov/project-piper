//go:build unit
// +build unit

package reportpipelinestatus

import (
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/stretchr/testify/assert"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/events"
	"testing"
)

func Test_SetEventOutcomeAndErrors(t *testing.T) {
	tests := []struct {
		name            string
		pipelineStatus  string
		gotLogFile      bool
		orchestrator    string
		expectedOutcome string
	}{
		{
			name:            "pipelineStatus: success",
			pipelineStatus:  orchestrator.BuildStatusSuccess,
			gotLogFile:      true,
			orchestrator:    "Jenkins",
			expectedOutcome: events.OutcomeSuccess,
		},
		{
			name:            "pipelineStatus: aborted",
			pipelineStatus:  orchestrator.BuildStatusAborted,
			gotLogFile:      true,
			orchestrator:    "Jenkins",
			expectedOutcome: events.OutcomeAborted,
		},
		{
			name:            "pipelineStatus: failure",
			pipelineStatus:  orchestrator.BuildStatusFailure,
			gotLogFile:      true,
			orchestrator:    "Jenkins",
			expectedOutcome: events.OutcomeError,
		},
		{
			name:            "no log file available - Jenkins",
			pipelineStatus:  orchestrator.BuildStatusFailure,
			gotLogFile:      false,
			orchestrator:    "Jenkins",
			expectedOutcome: events.OutcomeFailure,
		},
		{
			name:            "no log file available - GitHub Actions",
			pipelineStatus:  orchestrator.BuildStatusFailure,
			gotLogFile:      false,
			orchestrator:    "GitHub Actions",
			expectedOutcome: events.OutcomeFailure,
		},
		{
			name:            "unexpected inProgress status",
			pipelineStatus:  orchestrator.BuildStatusInProgress,
			gotLogFile:      true,
			orchestrator:    "GitHub Actions",
			expectedOutcome: events.OutcomeSuccess,
		},
		{
			name:            "inProgress status",
			pipelineStatus:  orchestrator.BuildStatusInProgress,
			gotLogFile:      true,
			orchestrator:    "GitHub Actions",
			expectedOutcome: events.OutcomeSuccess,
		},
		{
			name:            "empty status",
			pipelineStatus:  "",
			gotLogFile:      true,
			orchestrator:    "GitHub Actions",
			expectedOutcome: events.OutcomeFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventBytes, err := SetEventOutcomeAndErrors("{}", tt.pipelineStatus, tt.gotLogFile, tt.orchestrator)
			eventData, err := events.FromJSON(eventBytes)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutcome, eventData.Outcome)
		})
	}
}

package events

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_getEventData(t *testing.T) {
	tests := []struct {
		name      string
		optimized bool
		scheduled bool
		golden    string
	}{
		{
			name:      "Test getEventData() - not optimized and not scheduled",
			optimized: false,
			scheduled: false,
			golden:    "standard",
		},
		{
			name:      "Test getEventData() - optimized and not scheduled",
			optimized: true,
			scheduled: false,
			golden:    "optimized1",
		},
		{
			name:      "Test getEventData() - optimized and scheduled",
			optimized: true,
			scheduled: true,
			golden:    "optimized2",
		},
		{
			name:      "Test getEventData() - scheduled and not optimized",
			optimized: false,
			scheduled: true,
			golden:    "scheduled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init
			// get unknown orchestrator provider so that results are predictable
			t.Setenv("JENKINS_HOME", "")
			t.Setenv("JENKINS_URL", "")
			t.Setenv("GITHUB_ACTION", "")
			t.Setenv("GITHUB_ACTIONS", "")
			provider, _ := orchestrator.GetOrchestratorConfigProvider(nil)
			want, err := os.ReadFile(filepath.Join("testdata", tt.golden+".golden"))
			assert.NoError(t, err, "failed to load golden file")
			// test
			eventData := CreateEventData(provider, false, tt.optimized, tt.scheduled, mock.Anything, mock.Anything, mock.Anything)
			got, err := eventData.ToJSON()
			// assert
			assert.NoError(t, err)
			assert.Equal(t, string(want), string(got))
		})
	}
}

func TestFromJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expectError bool
		expected    EventData
	}{
		{
			name:  "Valid JSON returns EventData",
			input: []byte(`{"pipelineRunId":"12345","url":"http://example.com","commitId":"abcde","repositoryUrl":"http://repo.com","pipelineRunMode":"optimized","cumulusInformation":{"pipelineId":"pipeline123","pipelineRunKey":"runKey123"},"outcome":"success","errors":["error1","error2"]}`),
			expected: EventData{
				PipelineRunId:   "12345",
				Url:             "http://example.com",
				CommitId:        "abcde",
				RepositoryUrl:   "http://repo.com",
				PipelineRunMode: "optimized",
				CumulusInformation: cumulusInformation{
					PipelineId: "pipeline123",
					RunId:      "runKey123",
				},
				Outcome: "success",
				Errors:  []string{"error1", "error2"},
			},
			expectError: false,
		},
		{
			name:        "Invalid JSON returns error",
			input:       []byte(`{invalid json}`),
			expected:    EventData{},
			expectError: true,
		},
		{
			name:        "Empty JSON returns empty EventData",
			input:       []byte(`{}`),
			expected:    EventData{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventData, err := FromJSON(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expected, eventData)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, eventData)
			}
		})
	}
}

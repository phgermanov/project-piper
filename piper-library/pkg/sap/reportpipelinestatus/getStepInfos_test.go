//go:build unit
// +build unit

package reportpipelinestatus

import (
	"reflect"
	"testing"
)

func Test_matchStepInfo(t *testing.T) {
	type args struct {
		line        string
		stepDetails *[]StepDetails
	}
	tests := []struct {
		name string
		args args
		want *[]StepDetails
	}{
		{
			name: "match step telemetry data without errors",
			args: args{
				line:        "[2022-02-12T11:58:32.572Z] info  testing - Step telemetry data:{\"PipelineUrlHash\":\"someHash123\",\"BuildUrlHash\":\"someHash456\",\"StageName\":\"test\",\"StepName\":\"testing\",\"ErrorCode\":\"0\",\"Duration\":\"42\",\"ErrorCategory\":\"infrastructure\",\"CorrelationID\":\"https://testing.internal.sap/\",\"CommitHash\":\"e7836a3e5a13\",\"Branch\":\"main\",\"GitOwner\":\"piper\",\"GitRepository\":\"piper-library\",\"ErrorDetail\":null}",
				stepDetails: &[]StepDetails{},
			},
			want: &[]StepDetails{{
				PipelineUrlHash: "someHash123",
				BuildUrlHash:    "someHash456",
				StageName:       "test",
				StepName:        "testing",
				ErrorCode:       "0",
				Duration:        "42",
				ErrorCategory:   "infrastructure",
				CorrelationID:   "https://testing.internal.sap/",
				CommitHash:      "e7836a3e5a13",
				Branch:          "main",
				GitOwner:        "piper",
				GitRepository:   "piper-library",
				ErrorDetail:     ErrorDetail{},
			}},
		},
		{
			name: "match step telemetry data with errors",
			args: args{
				line:        "[2022-02-12T11:58:32.572Z] info  testing - Step telemetry data:{\"PipelineUrlHash\":\"someHash123\",\"BuildUrlHash\":\"someHash456\",\"StageName\":\"test\",\"StepName\":\"testing\",\"ErrorCode\":\"0\",\"Duration\":\"42\",\"ErrorCategory\":\"infrastructure\",\"CorrelationID\":\"https://testing.internal.sap/\",\"CommitHash\":\"e7836a3e5a13\",\"Branch\":\"main\",\"GitOwner\":\"piper\",\"GitRepository\":\"piper-library\",\"ErrorDetail\":null}",
				stepDetails: &[]StepDetails{},
			},
			want: &[]StepDetails{{
				PipelineUrlHash: "someHash123",
				BuildUrlHash:    "someHash456",
				StageName:       "test",
				StepName:        "testing",
				ErrorCode:       "0",
				Duration:        "42",
				ErrorCategory:   "infrastructure",
				CorrelationID:   "https://testing.internal.sap/",
				CommitHash:      "e7836a3e5a13",
				Branch:          "main",
				GitOwner:        "piper",
				GitRepository:   "piper-library",
				ErrorDetail:     ErrorDetail{},
			}},
		},
		{
			name: "match step telemetry data with duplicate entries - no duplicates reported",
			args: args{
				line: "[2022-02-12T11:58:32.572Z] info  testing - Step telemetry data:{\"PipelineUrlHash\":\"someHash123\",\"BuildUrlHash\":\"someHash456\",\"StageName\":\"test\",\"StepName\":\"testing\",\"ErrorCode\":\"0\",\"Duration\":\"42\",\"ErrorCategory\":\"infrastructure\",\"CorrelationID\":\"https://testing.internal.sap/\",\"CommitHash\":\"e7836a3e5a13\",\"Branch\":\"main\",\"GitOwner\":\"piper\",\"GitRepository\":\"piper-library\",\"ErrorDetail\":null}",
				stepDetails: &[]StepDetails{{
					PipelineUrlHash: "someHash123",
					BuildUrlHash:    "someHash456",
					StageName:       "test",
					StepName:        "testing",
					ErrorCode:       "0",
					Duration:        "42",
					ErrorCategory:   "infrastructure",
					CorrelationID:   "https://testing.internal.sap/",
					CommitHash:      "e7836a3e5a13",
					Branch:          "main",
					GitOwner:        "piper",
					GitRepository:   "piper-library",
					ErrorDetail:     ErrorDetail{},
				}},
			},
			want: &[]StepDetails{{
				PipelineUrlHash: "someHash123",
				BuildUrlHash:    "someHash456",
				StageName:       "test",
				StepName:        "testing",
				ErrorCode:       "0",
				Duration:        "42",
				ErrorCategory:   "infrastructure",
				CorrelationID:   "https://testing.internal.sap/",
				CommitHash:      "e7836a3e5a13",
				Branch:          "main",
				GitOwner:        "piper",
				GitRepository:   "piper-library",
				ErrorDetail:     ErrorDetail{},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matchStepInfo(tt.args.line, tt.args.stepDetails)
			if !reflect.DeepEqual(tt.want, tt.args.stepDetails) {
				t.Errorf("matchStepInfo() = %v, want %v", tt.args.stepDetails, tt.want)
			}
		})
	}
}

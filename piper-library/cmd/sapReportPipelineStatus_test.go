//go:build unit
// +build unit

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
)

type sapReportPipelineStatusMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newSapReportPipelineStatusTestsUtils() sapReportPipelineStatusMockUtils {
	utils := sapReportPipelineStatusMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func Test_runSapReportPipelineStatus(t *testing.T) {
	t.Parallel()
	type args struct {
		config        *sapReportPipelineStatusOptions
		telemetryData *telemetry.CustomData
		utils         sapReportPipelineStatusUtils
	}
	tests := []struct {
		name                string
		args                args
		expectedPipelineLog string
		wantErr             assert.ErrorAssertionFunc
	}{
		{name: "all reporting activated no error",
			args: args{
				config: &sapReportPipelineStatusOptions{
					SplunkReporting:    true,
					SplunkDsn:          "localhost",
					SplunkToken:        "123456",
					SplunkIndex:        "testing",
					TelemetryReporting: true,
					SendErrorLogs:      []string{"infrastructure"},
				},
				telemetryData: &telemetry.CustomData{},
				utils:         newSapReportPipelineStatusTestsUtils(),
			},
			expectedPipelineLog: "pipelineLog.log",
			wantErr:             assert.NoError,
		},
		{name: "all reporting activated all errors no error",
			args: args{
				config: &sapReportPipelineStatusOptions{
					SplunkReporting:    true,
					SplunkDsn:          "localhost",
					SplunkToken:        "123456",
					SplunkIndex:        "testing",
					TelemetryReporting: true,
					SendErrorLogs:      []string{"all"},
				},
				telemetryData: &telemetry.CustomData{},
				utils:         newSapReportPipelineStatusTestsUtils(),
			},
			expectedPipelineLog: "pipelineLog.log",
			wantErr:             assert.NoError,
		},
		{name: "all reporting activated all errors no error (isOptimizedAndScheduled and useCommitIDForCumulus are true)",
			args: args{
				config: &sapReportPipelineStatusOptions{
					SplunkReporting:         true,
					SplunkDsn:               "localhost",
					SplunkToken:             "123456",
					SplunkIndex:             "testing",
					TelemetryReporting:      true,
					SendErrorLogs:           []string{"all"},
					IsOptimizedAndScheduled: true,
					UseCommitIDForCumulus:   true,
				},
				telemetryData: &telemetry.CustomData{},
				utils:         newSapReportPipelineStatusTestsUtils(),
			},
			expectedPipelineLog: "pipelineLog-scheduled.log",
			wantErr:             assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, runSapReportPipelineStatus(tt.args.config, &sapReportPipelineStatusCommonPipelineEnvironment{}, tt.args.telemetryData, tt.args.utils), fmt.Sprintf("runSapReportPipelineStatus(%v, %v, %v)", tt.args.config, tt.args.telemetryData, tt.args.utils))
			files, _ := filepath.Glob("pipelineLog*")
			for _, f := range files {
				assert.Equal(t, tt.expectedPipelineLog, f)
				_ = os.Remove(f)
			}
		})
	}
}

func Test_sapReportPipelineStatus(t *testing.T) {
	type args struct {
		config        sapReportPipelineStatusOptions
		telemetryData *telemetry.CustomData
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "success - dummy test as no changes to default"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sapReportPipelineStatus(tt.args.config, tt.args.telemetryData, &sapReportPipelineStatusCommonPipelineEnvironment{})
		})
	}
}

//go:build unit
// +build unit

package reportpipelinestatus

import (
	"reflect"
	"testing"
)

func TestParseLogFile(t *testing.T) {

	_ = []byte(`[2021-11-25T16:27:09.199Z] debug sapXmakeExecuteBuild - --------------------------------
[2021-11-25T16:27:09.199Z] error sapXmakeExecuteBuild - failed to fetch report: failed to fetch artifact: Artifact 'build-results.json' not found - failed to fetch artifact: Artifact 'build-results.json'
not found
[2022-02-13T14:30:21.519Z] warn  sapXmakeExecuteBuild - Could not read env variable GIT_URL using fallback value n/a
[2022-02-13T14:30:21.519Z] info  sapXmakeExecuteBuild - Step telemetry data:{"PipelineURLHash":"c1c3e5238b115ce8e9afd1ad102d772811259d02","BuildURLHash":"9ef2049e8d7454a32c2f02d1635f260f58e08481","StageName":"Central Build","StepName":"sapXmakeExecuteBuild","ErrorCode":"1","Duration":"30986","ErrorCategory":"undefined","CorrelationID":"https://someurl.internal.sap/job/someOrg/job/someRepo/job/master/6026/","CommitHash":"n/a","Branch":"main","GitOwner":"n/a","GitRepository":"n/a","ErrorDetail":{"category":"undefined","correlationId":"https://someurl.internal.sap/job/someOrg/job/someRepo/job/master/6026/","error":"job did not succeed","library":"","message":"step execution failed","result":"failure","stepName":"sapXmakeExecuteBuild","time":"2022-02-13T14:30:21.136996384Z"}}`)

	logfileWithErrors := []byte(`
[2022-02-13T14:30:21.519Z] info  sapXmakeExecuteBuild - Step telemetry data:{"PipelineURLHash":"c1c3e5238b115ce8e9afd1ad102d772811259d02","BuildURLHash":"9ef2049e8d7454a32c2f02d1635f260f58e08481","StageName":"Central Build","StepName":"sapXmakeExecuteBuild","ErrorCode":"1","Duration":"30986","ErrorCategory":"undefined","CorrelationID":"https://someurl.internal.sap/job/someOrg/job/someRepo/job/master/6026/","CommitHash":"n/a","Branch":"main","GitOwner":"n/a","GitRepository":"n/a","ErrorDetail":{"category":"undefined","correlationId":"https://someurl.internal.sap/job/someOrg/job/someRepo/job/master/6026/","error":"job did not succeed","library":"","message":"step execution failed","result":"failure","stepName":"sapXmakeExecuteBuild","time":"2022-02-13T14:30:21.136996384Z"}}`)

	logFileNoError := []byte(`[2022-02-13T14:29:33.500Z] warn  artifactPrepareVersion - Could not read env variable GIT_URL using fallback value n/a
[2022-02-13T14:29:33.500Z] info  artifactPrepareVersion - Step telemetry data:{"PipelineURLHash":"c1c3e5238b115ce8e9afd1ad102d772811259d02","BuildURLHash":"9ef2049e8d7454a32c2f02d1635f260f58e08481","StageName":"Init","StepName":"artifactPrepareVersion","ErrorCode":"0","Duration":"3349","ErrorCategory":"undefined","CorrelationID":"https://someurl.internal.sap/job/someOrg/job/someRepo/job/master/6026/","CommitHash":"n/a","Branch":"main","GitOwner":"n/a","GitRepository":"n/a","ErrorDetail":null}
[2022-02-13T14:29:33.500Z] debug artifactPrepareVersion - Sending telemetry data`)

	type args struct {
		logFile *[]byte
	}
	tests := []struct {
		name           string
		args           args
		errorsInLog    []map[string]interface{}
		stepsInLog     []map[string]interface{}
		pipelineStatus string
	}{
		{
			name:           "logFile without errors",
			args:           args{logFile: &logFileNoError},
			stepsInLog:     []map[string]interface{}{{"Branch": "main", "BuildUrlHash": "9ef2049e8d7454a32c2f02d1635f260f58e08481", "CommitHash": "n/a", "CorrelationID": "https://someurl.internal.sap/job/someOrg/job/someRepo/job/master/6026/", "Duration": "3349", "ErrorCategory": "undefined", "ErrorCode": "0", "ErrorDetail": nil, "GitOwner": "n/a", "GitRepository": "n/a", "PipelineUrlHash": "c1c3e5238b115ce8e9afd1ad102d772811259d02", "StageName": "Init", "StepName": "artifactPrepareVersion"}},
			pipelineStatus: "SUCCESS",
		},
		{
			name:           "logFile with errors",
			args:           args{logFile: &logfileWithErrors},
			stepsInLog:     []map[string]interface{}{{"Branch": "main", "BuildUrlHash": "9ef2049e8d7454a32c2f02d1635f260f58e08481", "CommitHash": "n/a", "CorrelationID": "https://someurl.internal.sap/job/someOrg/job/someRepo/job/master/6026/", "Duration": "30986", "ErrorCategory": "undefined", "ErrorCode": "1", "ErrorDetail": map[string]interface{}{"Category": "undefined", "CorrelationId": "https://someurl.internal.sap/job/someOrg/job/someRepo/job/master/6026/", "Error": "job did not succeed", "Message": "step execution failed", "Result": "failure", "StepName": "sapXmakeExecuteBuild", "Time": "2022-02-13T14:30:21.136996384Z"}, "GitOwner": "n/a", "GitRepository": "n/a", "PipelineUrlHash": "c1c3e5238b115ce8e9afd1ad102d772811259d02", "StageName": "Central Build", "StepName": "sapXmakeExecuteBuild"}},
			errorsInLog:    []map[string]interface{}{{"Category": "undefined", "CorrelationId": "https://someurl.internal.sap/job/someOrg/job/someRepo/job/master/6026/", "Error": "job did not succeed", "Message": "step execution failed", "Result": "failure", "StepName": "sapXmakeExecuteBuild", "Time": "2022-02-13T14:30:21.136996384Z"}},
			pipelineStatus: "FAILURE",
		},
		{
			name:           "empty logFile",
			args:           args{logFile: &[]byte{}},
			errorsInLog:    []map[string]interface{}{},
			stepsInLog:     []map[string]interface{}{},
			pipelineStatus: "FAILURE",
		},
		{
			name:           "nil logFile",
			args:           args{logFile: nil},
			errorsInLog:    []map[string]interface{}{},
			stepsInLog:     []map[string]interface{}{},
			pipelineStatus: "FAILURE",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			determinedErrors, pipelineStatus, determinedSteps := ParseLogFile(tt.args.logFile)
			if !reflect.DeepEqual(determinedErrors, tt.errorsInLog) {
				t.Errorf("ParseLogFile() determinedErrors \n got = %v,\n want =%v", determinedErrors, tt.errorsInLog)
			}
			if pipelineStatus != tt.pipelineStatus {
				t.Errorf("SendLogs() = %v, want %v", pipelineStatus, tt.pipelineStatus)
			}

			if !reflect.DeepEqual(determinedSteps, tt.stepsInLog) {
				t.Errorf("ParseLogFile() determinedSteps \n got = %v,\n want =%v", determinedSteps, tt.stepsInLog)
			}
		})
	}
}

func TestSendLogs(t *testing.T) {
	type args struct {
		sendErrorLogs []string
		details       []map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "true, category in list",
			args: args{
				sendErrorLogs: []string{"build"},
				details:       []map[string]interface{}{{"Category": "build"}, {"Category": "undefined"}},
			},
			want: true,
		},
		{
			name: "false, category not in list",
			args: args{
				sendErrorLogs: []string{"compliance"},
				details:       []map[string]interface{}{{"Category": "build"}, {"Category": "undefined"}},
			},
			want: false,
		},
		{
			name: "true, category twice in list",
			args: args{
				sendErrorLogs: []string{"build"},
				details:       []map[string]interface{}{{"Category": "build"}, {"Category": "build"}},
			},
			want: true,
		},
		{
			name: "false, no error categories in list",
			args: args{
				sendErrorLogs: []string{"build"},
				details:       []map[string]interface{}{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SendLogs(tt.args.sendErrorLogs, tt.args.details); got != tt.want {
				t.Errorf("SendLogs() = %v, want %v", got, tt.want)
			}
		})
	}
}

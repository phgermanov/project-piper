//go:build unit
// +build unit

package reportpipelinestatus

import (
	"os"
	"reflect"
	"testing"
)

func Test_readErrorJson(t *testing.T) {
	type args struct {
		filePath string
	}

	tests := []struct {
		name            string
		args            args
		want            ErrorDetail
		wantErr         bool
		createTempFile  bool
		tempFileContent string
	}{
		{
			name: "readErrorJson successful",
			want: ErrorDetail{
				Message: "successful read file",
			},
			createTempFile:  true,
			tempFileContent: "{\"Message\":\"successful read file\"}",
			wantErr:         false,
		},
		{
			name:           "reads json file failure",
			createTempFile: false,
			wantErr:        true,
		},
		{
			name:            "unmarshal ErrorDetail failure",
			createTempFile:  true,
			tempFileContent: "{Message:unsuccessful read file}", // Malformed json object
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tempFile *os.File
			var err error
			if tt.createTempFile {
				tempFile, err = os.CreateTemp("", "*errorDetails.json")
				if err != nil {
					t.Fatal("failed to create temporary file")
				}
				if _, err := tempFile.Write([]byte(tt.tempFileContent)); err != nil {
					t.Fatal("could not write content to temp file")
				}

			} else {
				tempFile, err = os.CreateTemp("", "")
				if err != nil {
					t.Fatal("failed to create temporary file")
				}
			}

			got, err := readErrorJson(tempFile.Name())
			if (err != nil) != tt.wantErr {
				t.Errorf("readErrorJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readErrorJson() got = %v, want %v", got, tt.want)
			}
			err = os.RemoveAll(tempFile.Name())
			if err != nil {
				t.Fatal("failed to remove temp files")
			}
		})
	}
}

func Test_matchError(t *testing.T) {
	type args struct {
		line         string
		errorDetails *[]ErrorDetail
	}
	tests := []struct {
		name string
		args args
		want *[]ErrorDetail
	}{
		{
			name: "match",
			args: args{
				line:         "info  someStep - fatal error: errorDetails{\"category\":\"undefined\",\"correlationId\":\"https://some.url.sap/\",\"error\":\"fatal testing\",\"library\":\"\",\"message\":\"step execution failed\",\"result\":\"failure\",\"stepName\":\"someStep\"}",
				errorDetails: &[]ErrorDetail{},
			},
			want: &[]ErrorDetail{{
				Message:       "step execution failed",
				Error:         "fatal testing",
				Category:      "undefined",
				Result:        "failure",
				CorrelationId: "https://some.url.sap/",
				StepName:      "someStep",
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matchError(tt.args.line, tt.args.errorDetails)
			if !reflect.DeepEqual(tt.want, tt.args.errorDetails) {
				t.Errorf("matchStepInfo() = %v, want %v", tt.args.errorDetails, tt.want)
			}
		})
	}
}

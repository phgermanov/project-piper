//go:build unit
// +build unit

package dwc

import (
	"fmt"
	"testing"
	"time"
)

func TestNewWaitForDeploymentCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		stage    string
		vectorID string
		want     dwcCommand
	}{
		{
			name:     "valid command build",
			stage:    "sampleStage",
			vectorID: "1234",
			want: append(deploymentVectorBaseCommand,
				waitForVectorDeploymentSubcommand,
				fmt.Sprintf(vectorIDFlag, "1234"),
				fmt.Sprintf(stageFlag, "sampleStage"),
				fmt.Sprintf(outputFlag, outputFormatJSON),
			),
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got := newWaitForDeploymentCommand(testCase.stage, testCase.vectorID)
			verifyDwCCommand(t, got, testCase.want, len(deploymentVectorBaseCommand)+1)
		})
	}
}

func TestNewAddVectorUsageCommand(t *testing.T) {
	t.Parallel()
	expiry := time.Now().UTC().Add(time.Duration(stageWatchLockMinutes) * time.Minute).Format(time.RFC3339)
	tests := []struct {
		name      string
		landscape string
		vectorID  string
		expiry    string
		want      dwcCommand
	}{
		{
			name:      "valid command build",
			landscape: "sampleLandscape",
			vectorID:  "1234",
			expiry:    expiry,
			want: append(deploymentVectorBaseCommand,
				addVectorUsageSubcommand,
				fmt.Sprintf(usageNameFlag, pipelineUsageName),
				fmt.Sprintf(expiresAtFlag, expiry),
				fmt.Sprintf(vectorIDFlag, "1234"),
				fmt.Sprintf(landscapeFlag, "sampleLandscape"),
			),
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got := newAddVectorUsageCommand(testCase.landscape, testCase.vectorID, testCase.expiry)
			verifyDwCCommand(t, got, testCase.want, len(deploymentVectorBaseCommand)+1)
		})
	}
}

func TestNewRemoveVectorUsageCommand(t *testing.T) {
	t.Parallel()
	expiry := time.Now().UTC().Add(time.Duration(stageWatchLockMinutes) * time.Minute).Format(time.RFC3339)
	tests := []struct {
		name      string
		landscape string
		vectorID  string
		expiry    string
		want      dwcCommand
	}{
		{
			name:      "valid command build",
			landscape: "sampleLandscape",
			vectorID:  "1234",
			expiry:    expiry,
			want: append(deploymentVectorBaseCommand,
				removeVectorUsageSubcommand,
				fmt.Sprintf(usageNameFlag, pipelineUsageName),
				fmt.Sprintf(expiresAtFlag, expiry),
				fmt.Sprintf(vectorIDFlag, "1234"),
				fmt.Sprintf(landscapeFlag, "sampleLandscape"),
			),
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got := newRemoveVectorUsageCommand(testCase.landscape, testCase.vectorID, testCase.expiry)
			verifyDwCCommand(t, got, testCase.want, len(deploymentVectorBaseCommand)+1)
		})
	}
}

func TestNewWatchVectorDeploymentCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		landscape        string
		vectorID         string
		createDescriptor func(t *testing.T) ArtifactDescriptor
		want             dwcCommand
	}{
		{
			name:      "valid command build watching ROI only",
			landscape: "sampleLandscape",
			vectorID:  "1234",
			createDescriptor: func(t *testing.T) ArtifactDescriptor {
				t.Helper()
				descriptor := newArtifactDescriptorMock(t)
				descriptor.On("watchROIOnly").Return(true)
				descriptor.On("GetResourceName").Return("random/resource/ref")
				return descriptor
			},
			want: append(deploymentVectorBaseCommand,
				watchVectorDeploymentSubcommand,
				fmt.Sprintf(logsFlag, watchVectorDeploymentLogMode),
				fmt.Sprintf(vectorIDFlag, "1234"),
				fmt.Sprintf(landscapeFlag, "sampleLandscape"),
				fmt.Sprintf(timeoutFlag, stageWatchLockMinutes),
				fmt.Sprintf(roiFlag, "random/resource/ref"),
			),
		},
		{
			name:      "valid command build watching all resources",
			landscape: "sampleLandscape",
			vectorID:  "1234",
			createDescriptor: func(t *testing.T) ArtifactDescriptor {
				t.Helper()
				descriptor := newArtifactDescriptorMock(t)
				descriptor.On("watchROIOnly").Return(false)
				return descriptor
			},
			want: append(deploymentVectorBaseCommand,
				watchVectorDeploymentSubcommand,
				fmt.Sprintf(logsFlag, watchVectorDeploymentLogMode),
				fmt.Sprintf(vectorIDFlag, "1234"),
				fmt.Sprintf(landscapeFlag, "sampleLandscape"),
				fmt.Sprintf(timeoutFlag, stageWatchLockMinutes),
			),
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got := newWatchVectorDeploymentCommand(testCase.landscape, testCase.vectorID, testCase.createDescriptor(t))
			verifyDwCCommand(t, got, testCase.want, len(deploymentVectorBaseCommand)+1)
		})
	}
}

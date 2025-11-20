//go:build unit
// +build unit

package dwc

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/strings/slices"
	"strings"
	"testing"
)

func TestDefaultStageWatchOrchestrator_WatchVectorDeployments(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		orchestrator             *DefaultStageWatchOrchestrator
		promotionResults         []PromotionResultEntry
		executorFactory          BlockingExecutorFactory
		createArtifactDescriptor func() ArtifactDescriptor
		createDwCCommandExecutor func() CLICommandExecutor
		generateWantedOutput     func() []string // content order is non-deterministic therefore we validate substrings
		wantErr                  bool
	}{
		{
			name:         "canonical watch",
			orchestrator: &DefaultStageWatchOrchestrator{},
			promotionResults: []PromotionResultEntry{
				{
					Stage:    "Dev/AllStableFeatures",
					Status:   promotionResultStatusCreated,
					VectorId: "Uagoop0i",
					Error:    "",
				},
				{
					Stage:    "Dev/DeliverRel",
					Status:   promotionResultStatusCreated,
					VectorId: "lskljfhjA0",
					Error:    "",
				},
				{
					Stage:    "DeliverREL/RELprodEU10",
					Status:   promotionResultStatusCreated,
					VectorId: "lskljfhjBBBA0",
					Error:    "",
				},
			},
			createArtifactDescriptor: func() ArtifactDescriptor {
				descriptor := &artifactDescriptorMock{}
				descriptor.On("GetResourceName").Return("https://github.tools.sap/deploy-with-confidence/sponde/main")
				descriptor.On("watchROIOnly").Return(true)
				descriptor.On("getStageWatchPolicy").Return(OverallSuccessPolicy())
				return descriptor
			},
			executorFactory: &blockingExecutorFactoryMock{},
			createDwCCommandExecutor: func() CLICommandExecutor {
				cmdRunner := &dwCCommandExecutorMock{}
				cmdRunner.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(
					func(executorFactory BlockingExecutorFactory, targetCmd dwcCommand, cliResponseTargetReference interface{}) error {
						if slices.Contains(targetCmd, waitForVectorDeploymentSubcommand) {
							responsePointer := cliResponseTargetReference.(*WaitForDeploymentResponse)
							*responsePointer = WaitForDeploymentResponse{
								Landscape: "dev-eu10",
							}
						}
						if slices.Contains(targetCmd, watchVectorDeploymentSubcommand) {
							strPointer := cliResponseTargetReference.(*string)
							*strPointer = "live-logs-written"
						}
						return nil
					})
				return cmdRunner.Execute
			},
			generateWantedOutput: func() []string {
				statusLine1 := []byte(fmt.Sprintf(stageWatchPrefixROITpl, "https://github.tools.sap/deploy-with-confidence/sponde/main", "Dev/AllStableFeatures", "Uagoop0i", watchResultPrefixLineSuccess))
				statusLine2 := []byte(fmt.Sprintf(stageWatchPrefixROITpl, "https://github.tools.sap/deploy-with-confidence/sponde/main", "Dev/DeliverRel", "lskljfhjA0", watchResultPrefixLineSuccess))
				statusLine3 := []byte(fmt.Sprintf(stageWatchPrefixROITpl, "https://github.tools.sap/deploy-with-confidence/sponde/main", "DeliverREL/RELprodEU10", "lskljfhjBBBA0", watchResultPrefixLineSuccess))
				targetByteBlock1 := append(assembleBytesForStageWatchOutputPrefixReader(statusLine1), []byte("live-logs-written")...)
				targetByteBlock2 := append(assembleBytesForStageWatchOutputPrefixReader(statusLine2), []byte("live-logs-written")...)
				targetByteBlock3 := append(assembleBytesForStageWatchOutputPrefixReader(statusLine3), []byte("live-logs-written")...)
				return []string{string(targetByteBlock1), string(targetByteBlock2), string(targetByteBlock3), "\n"}
			},
			wantErr: false,
		},
		{
			name:         "watch with no changes to any stage",
			orchestrator: &DefaultStageWatchOrchestrator{},
			promotionResults: []PromotionResultEntry{
				{
					Stage:    "Dev/AllStableFeatures",
					Status:   promotionResultStatusSuccess,
					VectorId: "Uagoop0i",
				},
				{
					Stage:    "Dev/DeliverRel",
					Status:   promotionResultStatusSuccess,
					VectorId: "lskljfhjA0",
				},
				{
					Stage:    "DeliverREL/RELprodEU10",
					Status:   promotionResultStatusSuccess,
					VectorId: "lskljfhjBBBA0",
				},
			},
			createArtifactDescriptor: func() ArtifactDescriptor {
				descriptor := &artifactDescriptorMock{}
				descriptor.On("GetResourceName").Return("https://github.tools.sap/deploy-with-confidence/sponde/main")
				descriptor.On("watchROIOnly").Return(true)
				descriptor.On("getStageWatchPolicy").Return(OverallSuccessPolicy())
				return descriptor
			},
			executorFactory: &blockingExecutorFactoryMock{},
			createDwCCommandExecutor: func() CLICommandExecutor {
				cmdRunner := &dwCCommandExecutorMock{}
				return cmdRunner.Execute
			},
			generateWantedOutput: func() []string {
				statusLine1 := []byte(fmt.Sprintf(stageWatchPrefixROITpl, "https://github.tools.sap/deploy-with-confidence/sponde/main", "Dev/AllStableFeatures", "Uagoop0i", watchResultPrefixLineSuccess))
				statusLine2 := []byte(fmt.Sprintf(stageWatchPrefixROITpl, "https://github.tools.sap/deploy-with-confidence/sponde/main", "Dev/DeliverRel", "lskljfhjA0", watchResultPrefixLineSuccess))
				statusLine3 := []byte(fmt.Sprintf(stageWatchPrefixROITpl, "https://github.tools.sap/deploy-with-confidence/sponde/main", "DeliverREL/RELprodEU10", "lskljfhjBBBA0", watchResultPrefixLineSuccess))
				targetByteBlock1 := append(assembleBytesForStageWatchOutputPrefixReader(statusLine1), []byte(promotionResultStatusSuccessMsgTpl)...)
				targetByteBlock2 := append(assembleBytesForStageWatchOutputPrefixReader(statusLine2), []byte(promotionResultStatusSuccessMsgTpl)...)
				targetByteBlock3 := append(assembleBytesForStageWatchOutputPrefixReader(statusLine3), []byte(promotionResultStatusSuccessMsgTpl)...)
				return []string{string(targetByteBlock1), string(targetByteBlock2), string(targetByteBlock3), "\n"}
			},
			wantErr: false,
		},
		{
			name:         "watch with promotion error",
			orchestrator: &DefaultStageWatchOrchestrator{},
			promotionResults: []PromotionResultEntry{
				{
					Stage:    "Dev/AllStableFeatures",
					Status:   promotionResultStatusError,
					VectorId: "Uagoop0i",
					Error:    "Some error along the way",
				},
				{
					Stage:    "Dev/DeliverRel",
					Status:   promotionResultStatusCreated,
					VectorId: "lskljfhjA0",
					Error:    "",
				},
				{
					Stage:    "DeliverREL/RELprodEU10",
					Status:   promotionResultStatusCreated,
					VectorId: "lskljfhjBBBA0",
					Error:    "",
				},
			},
			createArtifactDescriptor: func() ArtifactDescriptor {
				descriptor := &artifactDescriptorMock{}
				descriptor.On("GetResourceName").Return("https://github.tools.sap/deploy-with-confidence/sponde/main")
				descriptor.On("watchROIOnly").Return(true)
				descriptor.On("getStageWatchPolicy").Return(OverallSuccessPolicy())
				return descriptor
			},
			executorFactory: &blockingExecutorFactoryMock{},
			createDwCCommandExecutor: func() CLICommandExecutor {
				cmdRunner := &dwCCommandExecutorMock{}
				cmdRunner.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(
					func(executorFactory BlockingExecutorFactory, targetCmd dwcCommand, cliResponseTargetReference interface{}) error {
						if slices.Contains(targetCmd, waitForVectorDeploymentSubcommand) {
							responsePointer := cliResponseTargetReference.(*WaitForDeploymentResponse)
							*responsePointer = WaitForDeploymentResponse{
								Landscape: "dev-eu10",
							}
						}
						if slices.Contains(targetCmd, watchVectorDeploymentSubcommand) {
							strPointer := cliResponseTargetReference.(*string)
							*strPointer = "live-logs-written"
						}
						return nil
					})
				return cmdRunner.Execute
			},
			generateWantedOutput: func() []string {
				targetErr := &stageWatchError{
					stage: "Dev/AllStableFeatures",
					inner: fmt.Errorf(promotionResultStatusErrorMsgTpl, promotionResultStatusError, "Some error along the way"),
				}
				preErrorStream := append([]byte(targetErr.Error()), lineBreak)
				statusLine1 := []byte(fmt.Sprintf(stageWatchPrefixROITpl, "https://github.tools.sap/deploy-with-confidence/sponde/main", "Dev/AllStableFeatures", "Uagoop0i", watchResultPrefixLineFailure))
				statusLine2 := []byte(fmt.Sprintf(stageWatchPrefixROITpl, "https://github.tools.sap/deploy-with-confidence/sponde/main", "Dev/DeliverRel", "lskljfhjA0", watchResultPrefixLineSuccess))
				statusLine3 := []byte(fmt.Sprintf(stageWatchPrefixROITpl, "https://github.tools.sap/deploy-with-confidence/sponde/main", "DeliverREL/RELprodEU10", "lskljfhjBBBA0", watchResultPrefixLineSuccess))
				targetByteBlock1 := append(assembleBytesForStageWatchOutputPrefixReader(statusLine1), append([]byte(targetErr.Error()), lineBreak)...)
				targetByteBlock2 := append(assembleBytesForStageWatchOutputPrefixReader(statusLine2), []byte("live-logs-written")...)
				targetByteBlock3 := append(assembleBytesForStageWatchOutputPrefixReader(statusLine3), []byte("live-logs-written")...)
				return []string{string(preErrorStream), string(targetByteBlock1), string(targetByteBlock2), string(targetByteBlock3), "\n"}
			},
			wantErr: true,
		},
		{
			name:         "orbit watch",
			orchestrator: &DefaultStageWatchOrchestrator{},
			promotionResults: []PromotionResultEntry{
				{
					Stage:    "orbit/dev-eu10-k8s",
					Status:   promotionResultStatusCreated,
					VectorId: "Uagoop0i",
					Error:    "",
				},
				{
					Stage:    "orbit/dev-eu10-k9s",
					Status:   promotionResultStatusSuccess,
					VectorId: "lskljfhjA0",
					Error:    "",
				},
				{
					Stage:    "DeliverREL/RELprodEU10",
					Status:   promotionResultStatusCreated,
					VectorId: "lskljfhjBBBA0",
					Error:    "",
				},
			},
			createArtifactDescriptor: func() ArtifactDescriptor {
				descriptor := &artifactDescriptorMock{}
				descriptor.On("GetResourceName").Return("https://github.tools.sap/deploy-with-confidence/sponde/main")
				descriptor.On("watchROIOnly").Return(false)
				descriptor.On("getStageWatchPolicy").Return(OverallSuccessPolicy())
				return descriptor
			},
			executorFactory: &blockingExecutorFactoryMock{},
			createDwCCommandExecutor: func() CLICommandExecutor {
				cmdRunner := &dwCCommandExecutorMock{}
				cmdRunner.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(
					func(executorFactory BlockingExecutorFactory, targetCmd dwcCommand, cliResponseTargetReference interface{}) error {
						if slices.Contains(targetCmd, waitForVectorDeploymentSubcommand) {
							responsePointer := cliResponseTargetReference.(*WaitForDeploymentResponse)
							*responsePointer = WaitForDeploymentResponse{
								Landscape: "dev-eu10",
							}
						}
						if slices.Contains(targetCmd, watchVectorDeploymentSubcommand) {
							strPointer := cliResponseTargetReference.(*string)
							*strPointer = "live-logs-written"
						}
						return nil
					})
				return cmdRunner.Execute
			},
			generateWantedOutput: func() []string {
				statusLine1 := []byte(fmt.Sprintf(stageWatchPrefixTpl, "orbit/dev-eu10-k8s", "Uagoop0i", watchResultPrefixLineUnknown))
				statusLine2 := []byte(fmt.Sprintf(stageWatchPrefixTpl, "orbit/dev-eu10-k9s", "lskljfhjA0", watchResultPrefixLineSuccess))
				statusLine3 := []byte(fmt.Sprintf(stageWatchPrefixTpl, "DeliverREL/RELprodEU10", "lskljfhjBBBA0", watchResultPrefixLineSuccess))
				targetByteBlock1 := append(assembleBytesForStageWatchOutputPrefixReader(statusLine1), []byte(fmt.Sprintf(orbitDeploymentMsgTpl, "orbit/dev-eu10-k8s"))...)
				targetByteBlock2 := append(assembleBytesForStageWatchOutputPrefixReader(statusLine2), []byte(promotionResultStatusSuccessMsgTpl)...)
				targetByteBlock3 := append(assembleBytesForStageWatchOutputPrefixReader(statusLine3), []byte("live-logs-written")...)
				return []string{string(targetByteBlock1), string(targetByteBlock2), string(targetByteBlock3), "\n"}
			},
			wantErr: false,
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			checkBuff := &bytes.Buffer{}
			err := testCase.orchestrator.WatchVectorDeployments(testCase.promotionResults, checkBuff, testCase.createDwCCommandExecutor(), testCase.executorFactory, testCase.createArtifactDescriptor())
			if (err != nil) != testCase.wantErr {
				t.Fatalf("WatchVectorDeployments() error = %v, wantErr %v", err, testCase.wantErr)
			}
			checkBuffStr := checkBuff.String()
			wanted := testCase.generateWantedOutput()
			for _, want := range wanted {
				assert.Contains(t, checkBuffStr, want)
			}
			if len(checkBuffStr) != len(strings.Join(wanted, "")) {
				t.Fatalf("Expected stageWatchOutput to be of length %d, got %d. output is %s, wanted is all of %v", len(checkBuffStr), len(strings.Join(wanted, "")), checkBuffStr, wanted)
			}
		})
	}
}

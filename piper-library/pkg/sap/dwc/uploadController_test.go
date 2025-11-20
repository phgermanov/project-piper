//go:build unit
// +build unit

package dwc

import (
	"path"
	"testing"

	piperMocks "github.com/SAP/jenkins-library/pkg/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/strings/slices"
)

func TestUploadController_UploadArtifact(t *testing.T) {
	t.Parallel()
	const (
		wd                    = "/usr/local"
		tmpdir                = "/usr/local/tmpDirOnFS"
		uploadFolderStructure = "/some/path"
	)
	var (
		filePatterns            = []string{"A", "B"}
		foundFilePatternMatches = []string{"path/A", "path/sub/B"}
		uploadFolderPath        = path.Join(tmpdir, uploadFolderStructure)
		stagesToWatch           = []string{"dev/A", "dev/B"}
	)
	tests := []struct {
		name                    string
		configureController     func(t *testing.T) UploadController
		generateControllerInput func(t *testing.T) ArtifactDescriptor
		want                    *ArtifactUploadResponse
		wantErr                 bool
	}{
		{
			name: "canonical upload",
			configureController: func(t *testing.T) UploadController {
				t.Helper()
				fsMock := &piperMocks.FilesMock{CurrentDir: wd}
				treeEditor := newMockFileTreeEditor(t)
				treeEditor.On("TempDir", wd, "*").Return(tmpdir, nil)
				treeEditor.On("MkdirAll", uploadFolderPath, mock.Anything).Return(nil) // the tree editor should create the upload folder path incl. the root
				matcher := &mockGlobMatcher{}
				matcher.On("Glob", mock.Anything).Return(
					func(pattern string) []string {
						if pattern == "A" {
							return []string{"path/A"}
						} else if pattern == "B" {
							return []string{"path/sub/B"}
						}
						return nil
					},
					func(pattern string) error {
						return nil
					},
				)
				mover := newMockFilePatternMover(t)
				mover.On("move", foundFilePatternMatches, uploadFolderPath, mock.Anything).Return(nil) // All file matches should be moved to the target folder
				compressor := newMockArtifactCompressor(t)
				compressor.On("compressArtifact", mock.Anything, tmpdir).Return(nil) // the compressor should be called on the tempDir root
				execFactory := &blockingExecutorFactoryMock{}
				executor := &piperMocks.ExecMockRunner{}
				execFactory.On("CreateExecutor", mock.Anything, mock.Anything).Return(executor)
				cmdRunner := newDwCCommandExecutorMock(t)
				uploadResponse := ArtifactUploadResponse{
					AppName:       "test-service",
					ID:            "12345-678901234",
					CreatedVector: "01877e61-1bab-3f47-f0a3-123456789fe5",
					Type:          ArtifactTypeDocker,
					PromotionResult: []PromotionResultEntry{
						{
							Stage:    "dev/A",
							Status:   promotionResultStatusCreated,
							VectorId: "Aisah6Eo",
						},
						{
							Stage:    "dev/B",
							Status:   promotionResultStatusCreated,
							VectorId: "Komuid1g",
						},
					},
				}
				cmdRunner.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(
					func(executorFactory BlockingExecutorFactory, targetCmd dwcCommand, cliResponseTargetReference interface{}) error {
						if slices.Contains(targetCmd, uploadBaseCommand[1]) {
							uploadResp := cliResponseTargetReference.(*ArtifactUploadResponse)
							*uploadResp = uploadResponse
						}
						return nil
					})
				orchestrator := newMockStageWatchOrchestrator(t)
				orchestrator.On("WatchVectorDeployments", uploadResponse.PromotionResult, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil) // Vector Deployments should be watched with the filtered stages to watch
				return NewUploadController(compressor, treeEditor, mover, fsMock, matcher, orchestrator, execFactory, cmdRunner.Execute)
			},
			generateControllerInput: func(t *testing.T) ArtifactDescriptor {
				t.Helper()
				artifactDescriptor := &artifactDescriptorMock{}
				artifactDescriptor.On("hasFilePatterns").Return(true)
				artifactDescriptor.On("prepareFiles").Return(nil)
				artifactDescriptor.On("needsFileBundling").Return(true)
				artifactDescriptor.On("getArtifactUploadFolderStructure").Return(uploadFolderStructure)
				artifactDescriptor.On("getFilePatterns").Return(filePatterns)
				artifactDescriptor.On("getUploadFileName").Return("upload.zip")
				artifactDescriptor.On("hasStagesToWatch").Return(true)
				artifactDescriptor.On("getStagesToWatch").Return(stagesToWatch)
				artifactDescriptor.On("buildUploadCommand").Return(dwcCommand(uploadBaseCommand), nil)
				return artifactDescriptor
			},
			wantErr: false,
			want: &ArtifactUploadResponse{
				AppName:       "test-service",
				ID:            "12345-678901234",
				CreatedVector: "01877e61-1bab-3f47-f0a3-123456789fe5",
				Type:          ArtifactTypeDocker,
				PromotionResult: []PromotionResultEntry{
					{
						Stage:    "dev/A",
						Status:   promotionResultStatusCreated,
						VectorId: "Aisah6Eo",
					},
					{
						Stage:    "dev/B",
						Status:   promotionResultStatusCreated,
						VectorId: "Komuid1g",
					},
				},
			},
		},
		{
			name: "upload operation that fails watching vector deployments",
			configureController: func(t *testing.T) UploadController {
				t.Helper()
				fsMock := &piperMocks.FilesMock{CurrentDir: wd}
				treeEditor := newMockFileTreeEditor(t)
				treeEditor.On("TempDir", wd, "*").Return(tmpdir, nil)
				treeEditor.On("MkdirAll", uploadFolderPath, mock.Anything).Return(nil) // the tree editor should create the upload folder path incl. the root
				matcher := &mockGlobMatcher{}
				matcher.On("Glob", mock.Anything).Return(
					func(pattern string) []string {
						if pattern == "A" {
							return []string{"path/A"}
						} else if pattern == "B" {
							return []string{"path/sub/B"}
						}
						return nil
					},
					func(pattern string) error {
						return nil
					},
				)
				mover := newMockFilePatternMover(t)
				mover.On("move", foundFilePatternMatches, uploadFolderPath, mock.Anything).Return(nil) // All file matches should be moved to the target folder
				compressor := newMockArtifactCompressor(t)
				compressor.On("compressArtifact", mock.Anything, tmpdir).Return(nil) // the compressor should be called on the tempDir root
				execFactory := &blockingExecutorFactoryMock{}
				executor := &piperMocks.ExecMockRunner{}
				execFactory.On("CreateExecutor", mock.Anything, mock.Anything).Return(executor)
				cmdRunner := newDwCCommandExecutorMock(t)
				uploadResponse := ArtifactUploadResponse{
					AppName:       "test-service",
					ID:            "12345-678901234",
					CreatedVector: "01877e61-1bab-3f47-f0a3-123456789fe5",
					Type:          ArtifactTypeDocker,
					PromotionResult: []PromotionResultEntry{
						{
							Stage:    "dev/A",
							Status:   promotionResultStatusCreated,
							VectorId: "Aisah6Eo",
						},
						{
							Stage:    "dev/B",
							Status:   promotionResultStatusCreated,
							VectorId: "Komuid1g",
						},
					},
				}
				cmdRunner.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(
					func(executorFactory BlockingExecutorFactory, targetCmd dwcCommand, cliResponseTargetReference interface{}) error {
						if slices.Contains(targetCmd, uploadBaseCommand[1]) {
							uploadResp := cliResponseTargetReference.(*ArtifactUploadResponse)
							*uploadResp = uploadResponse
						}
						return nil
					})
				orchestrator := newMockStageWatchOrchestrator(t)
				orchestrator.On("WatchVectorDeployments", uploadResponse.PromotionResult, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("Some error that popped up")) // Vector Deployments should be watched with the filtered stages to watch
				return NewUploadController(compressor, treeEditor, mover, fsMock, matcher, orchestrator, execFactory, cmdRunner.Execute)
			},
			generateControllerInput: func(t *testing.T) ArtifactDescriptor {
				t.Helper()
				artifactDescriptor := &artifactDescriptorMock{}
				artifactDescriptor.On("hasFilePatterns").Return(true)
				artifactDescriptor.On("prepareFiles").Return(nil)
				artifactDescriptor.On("needsFileBundling").Return(true)
				artifactDescriptor.On("getArtifactUploadFolderStructure").Return(uploadFolderStructure)
				artifactDescriptor.On("getFilePatterns").Return(filePatterns)
				artifactDescriptor.On("getUploadFileName").Return("upload.zip")
				artifactDescriptor.On("hasStagesToWatch").Return(true)
				artifactDescriptor.On("getStagesToWatch").Return(stagesToWatch)
				artifactDescriptor.On("buildUploadCommand").Return(dwcCommand(uploadBaseCommand), nil)
				return artifactDescriptor
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "upload operation that fails preparing files",
			configureController: func(t *testing.T) UploadController {
				t.Helper()
				fsMock := &piperMocks.FilesMock{CurrentDir: wd}
				treeEditor := newMockFileTreeEditor(t)
				matcher := &mockGlobMatcher{}
				mover := newMockFilePatternMover(t)
				compressor := newMockArtifactCompressor(t)
				execFactory := &blockingExecutorFactoryMock{}
				cmdRunner := newDwCCommandExecutorMock(t)
				orchestrator := newMockStageWatchOrchestrator(t)
				return NewUploadController(compressor, treeEditor, mover, fsMock, matcher, orchestrator, execFactory, cmdRunner.Execute)
			},
			generateControllerInput: func(t *testing.T) ArtifactDescriptor {
				t.Helper()
				artifactDescriptor := &artifactDescriptorMock{}
				artifactDescriptor.On("hasFilePatterns").Return(true)
				artifactDescriptor.On("prepareFiles").Return(errors.New("error while preparing files"))
				return artifactDescriptor
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "upload operation that fails bundling files",
			configureController: func(t *testing.T) UploadController {
				t.Helper()
				fsMock := &piperMocks.FilesMock{CurrentDir: wd}
				treeEditor := newMockFileTreeEditor(t)
				treeEditor.On("TempDir", wd, "*").Return("", errors.New("failed to create tempdir"))
				matcher := &mockGlobMatcher{}
				mover := newMockFilePatternMover(t)
				compressor := newMockArtifactCompressor(t)
				execFactory := &blockingExecutorFactoryMock{}
				cmdRunner := newDwCCommandExecutorMock(t)
				orchestrator := newMockStageWatchOrchestrator(t)
				return NewUploadController(compressor, treeEditor, mover, fsMock, matcher, orchestrator, execFactory, cmdRunner.Execute)
			},
			generateControllerInput: func(t *testing.T) ArtifactDescriptor {
				t.Helper()
				artifactDescriptor := &artifactDescriptorMock{}
				artifactDescriptor.On("hasFilePatterns").Return(true)
				artifactDescriptor.On("prepareFiles").Return(nil)
				artifactDescriptor.On("needsFileBundling").Return(true)
				return artifactDescriptor
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "upload operation that fails because needs file bundling but artifactFilesToUpload is missing",
			configureController: func(t *testing.T) UploadController {
				t.Helper()
				fsMock := &piperMocks.FilesMock{CurrentDir: wd}
				treeEditor := newMockFileTreeEditor(t)
				matcher := &mockGlobMatcher{}
				mover := newMockFilePatternMover(t)
				compressor := newMockArtifactCompressor(t)
				execFactory := &blockingExecutorFactoryMock{}
				cmdRunner := newDwCCommandExecutorMock(t)
				orchestrator := newMockStageWatchOrchestrator(t)
				return NewUploadController(compressor, treeEditor, mover, fsMock, matcher, orchestrator, execFactory, cmdRunner.Execute)
			},
			generateControllerInput: func(t *testing.T) ArtifactDescriptor {
				t.Helper()
				artifactDescriptor := &artifactDescriptorMock{}
				artifactDescriptor.On("hasFilePatterns").Return(false)
				artifactDescriptor.On("needsFileBundling").Return(true)
				return artifactDescriptor
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "upload operation that fails building upload command",
			configureController: func(t *testing.T) UploadController {
				t.Helper()
				fsMock := &piperMocks.FilesMock{CurrentDir: wd}
				treeEditor := newMockFileTreeEditor(t)
				treeEditor.On("TempDir", wd, "*").Return(tmpdir, nil)
				treeEditor.On("MkdirAll", uploadFolderPath, mock.Anything).Return(nil) // the tree editor should create the upload folder path incl. the root
				matcher := &mockGlobMatcher{}
				matcher.On("Glob", mock.Anything).Return(
					func(pattern string) []string {
						if pattern == "A" {
							return []string{"path/A"}
						} else if pattern == "B" {
							return []string{"path/sub/B"}
						}
						return nil
					},
					func(pattern string) error {
						return nil
					},
				)
				mover := newMockFilePatternMover(t)
				mover.On("move", foundFilePatternMatches, uploadFolderPath, mock.Anything).Return(nil) // All file matches should be moved to the target folder
				compressor := newMockArtifactCompressor(t)
				compressor.On("compressArtifact", mock.Anything, tmpdir).Return(nil) // the compressor should be called on the tempDir root
				execFactory := &blockingExecutorFactoryMock{}
				executor := &piperMocks.ExecMockRunner{}
				execFactory.On("CreateExecutor", mock.Anything, mock.Anything).Return(executor)
				cmdRunner := newDwCCommandExecutorMock(t)
				orchestrator := newMockStageWatchOrchestrator(t)
				return NewUploadController(compressor, treeEditor, mover, fsMock, matcher, orchestrator, execFactory, cmdRunner.Execute)
			},
			generateControllerInput: func(t *testing.T) ArtifactDescriptor {
				t.Helper()
				artifactDescriptor := &artifactDescriptorMock{}
				artifactDescriptor.On("hasFilePatterns").Return(true)
				artifactDescriptor.On("prepareFiles").Return(nil)
				artifactDescriptor.On("needsFileBundling").Return(true)
				artifactDescriptor.On("getArtifactUploadFolderStructure").Return(uploadFolderStructure)
				artifactDescriptor.On("getFilePatterns").Return(filePatterns)
				artifactDescriptor.On("getUploadFileName").Return("upload.zip")
				artifactDescriptor.On("buildUploadCommand").Return(nil, errors.New("failed to build upload command"))
				return artifactDescriptor
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "upload operation that fails executing upload command",
			configureController: func(t *testing.T) UploadController {
				t.Helper()
				fsMock := &piperMocks.FilesMock{CurrentDir: wd}
				treeEditor := newMockFileTreeEditor(t)
				treeEditor.On("TempDir", wd, "*").Return(tmpdir, nil)
				treeEditor.On("MkdirAll", uploadFolderPath, mock.Anything).Return(nil) // the tree editor should create the upload folder path incl. the root
				matcher := &mockGlobMatcher{}
				matcher.On("Glob", mock.Anything).Return(
					func(pattern string) []string {
						if pattern == "A" {
							return []string{"path/A"}
						} else if pattern == "B" {
							return []string{"path/sub/B"}
						}
						return nil
					},
					func(pattern string) error {
						return nil
					},
				)
				mover := newMockFilePatternMover(t)
				mover.On("move", foundFilePatternMatches, uploadFolderPath, mock.Anything).Return(nil) // All file matches should be moved to the target folder
				compressor := newMockArtifactCompressor(t)
				compressor.On("compressArtifact", mock.Anything, tmpdir).Return(nil) // the compressor should be called on the tempDir root
				execFactory := &blockingExecutorFactoryMock{}
				executor := &piperMocks.ExecMockRunner{}
				execFactory.On("CreateExecutor", mock.Anything, mock.Anything).Return(executor)
				cmdRunner := newDwCCommandExecutorMock(t)
				cmdRunner.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(
					func(executorFactory BlockingExecutorFactory, targetCmd dwcCommand, cliResponseTargetReference interface{}) error {
						if slices.Contains(targetCmd, uploadBaseCommand[1]) {
							return errors.New("execution of upload command failed")
						}
						return nil
					})
				orchestrator := newMockStageWatchOrchestrator(t)
				return NewUploadController(compressor, treeEditor, mover, fsMock, matcher, orchestrator, execFactory, cmdRunner.Execute)
			},
			generateControllerInput: func(t *testing.T) ArtifactDescriptor {
				t.Helper()
				artifactDescriptor := &artifactDescriptorMock{}
				artifactDescriptor.On("hasFilePatterns").Return(true)
				artifactDescriptor.On("prepareFiles").Return(nil)
				artifactDescriptor.On("needsFileBundling").Return(true)
				artifactDescriptor.On("getArtifactUploadFolderStructure").Return(uploadFolderStructure)
				artifactDescriptor.On("getFilePatterns").Return(filePatterns)
				artifactDescriptor.On("getUploadFileName").Return("upload.zip")
				artifactDescriptor.On("buildUploadCommand").Return(dwcCommand(uploadBaseCommand), nil)
				return artifactDescriptor
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "upload response stages are filtered by user input",
			configureController: func(t *testing.T) UploadController {
				t.Helper()
				fsMock := &piperMocks.FilesMock{CurrentDir: wd}
				treeEditor := newMockFileTreeEditor(t)
				treeEditor.On("TempDir", wd, "*").Return(tmpdir, nil)
				treeEditor.On("MkdirAll", uploadFolderPath, mock.Anything).Return(nil) // the tree editor should create the upload folder path incl. the root
				matcher := &mockGlobMatcher{}
				matcher.On("Glob", mock.Anything).Return(
					func(pattern string) []string {
						if pattern == "A" {
							return []string{"path/A"}
						} else if pattern == "B" {
							return []string{"path/sub/B"}
						}
						return nil
					},
					func(pattern string) error {
						return nil
					},
				)
				mover := newMockFilePatternMover(t)
				mover.On("move", foundFilePatternMatches, uploadFolderPath, mock.Anything).Return(nil) // All file matches should be moved to the target folder
				compressor := newMockArtifactCompressor(t)
				compressor.On("compressArtifact", mock.Anything, tmpdir).Return(nil) // the compressor should be called on the tempDir root
				execFactory := &blockingExecutorFactoryMock{}
				executor := &piperMocks.ExecMockRunner{}
				execFactory.On("CreateExecutor", mock.Anything, mock.Anything).Return(executor)
				cmdRunner := newDwCCommandExecutorMock(t)
				uploadResponse := ArtifactUploadResponse{
					AppName:       "test-service",
					ID:            "12345-678901234",
					CreatedVector: "01877e61-1bab-3f47-f0a3-123456789fe5",
					Type:          ArtifactTypeDocker,
					PromotionResult: []PromotionResultEntry{
						{
							Stage:    "dev/A",
							Status:   promotionResultStatusCreated,
							VectorId: "Aisah6Eo",
						},
						{
							Stage:    "dev/B",
							Status:   promotionResultStatusCreated,
							VectorId: "Komuid1g",
						},
						{
							Stage:    "dev/C",
							Status:   promotionResultStatusCreated,
							VectorId: "sgjklösdjgö3",
						},
					},
				}
				cmdRunner.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(
					func(executorFactory BlockingExecutorFactory, targetCmd dwcCommand, cliResponseTargetReference interface{}) error {
						if slices.Contains(targetCmd, uploadBaseCommand[1]) {
							uploadResp := cliResponseTargetReference.(*ArtifactUploadResponse)
							*uploadResp = uploadResponse
						}
						return nil
					})
				orchestrator := newMockStageWatchOrchestrator(t)
				orchestrator.On("WatchVectorDeployments", uploadResponse.PromotionResult[0:2], mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil) // Vector Deployments should be watched with the filtered stages to watch
				return NewUploadController(compressor, treeEditor, mover, fsMock, matcher, orchestrator, execFactory, cmdRunner.Execute)
			},
			generateControllerInput: func(t *testing.T) ArtifactDescriptor {
				t.Helper()
				artifactDescriptor := &artifactDescriptorMock{}
				artifactDescriptor.On("hasFilePatterns").Return(true)
				artifactDescriptor.On("prepareFiles").Return(nil)
				artifactDescriptor.On("needsFileBundling").Return(true)
				artifactDescriptor.On("getArtifactUploadFolderStructure").Return(uploadFolderStructure)
				artifactDescriptor.On("getFilePatterns").Return(filePatterns)
				artifactDescriptor.On("getUploadFileName").Return("upload.zip")
				artifactDescriptor.On("hasStagesToWatch").Return(true)
				artifactDescriptor.On("getStagesToWatch").Return(stagesToWatch)
				artifactDescriptor.On("buildUploadCommand").Return(dwcCommand(uploadBaseCommand), nil)
				return artifactDescriptor
			},
			wantErr: false,
			want: &ArtifactUploadResponse{
				AppName:       "test-service",
				ID:            "12345-678901234",
				CreatedVector: "01877e61-1bab-3f47-f0a3-123456789fe5",
				Type:          ArtifactTypeDocker,
				PromotionResult: []PromotionResultEntry{
					{
						Stage:    "dev/A",
						Status:   promotionResultStatusCreated,
						VectorId: "Aisah6Eo",
					},
					{
						Stage:    "dev/B",
						Status:   promotionResultStatusCreated,
						VectorId: "Komuid1g",
					},
					{
						Stage:    "dev/C",
						Status:   promotionResultStatusCreated,
						VectorId: "sgjklösdjgö3",
					},
				},
			},
		},
		{
			name: "validate all stage selector",
			configureController: func(t *testing.T) UploadController {
				t.Helper()
				fsMock := &piperMocks.FilesMock{CurrentDir: wd}
				treeEditor := newMockFileTreeEditor(t)
				treeEditor.On("TempDir", wd, "*").Return(tmpdir, nil)
				treeEditor.On("MkdirAll", uploadFolderPath, mock.Anything).Return(nil) // the tree editor should create the upload folder path incl. the root
				matcher := &mockGlobMatcher{}
				matcher.On("Glob", mock.Anything).Return(
					func(pattern string) []string {
						if pattern == "A" {
							return []string{"path/A"}
						} else if pattern == "B" {
							return []string{"path/sub/B"}
						}
						return nil
					},
					func(pattern string) error {
						return nil
					},
				)
				mover := newMockFilePatternMover(t)
				mover.On("move", foundFilePatternMatches, uploadFolderPath, mock.Anything).Return(nil) // All file matches should be moved to the target folder
				compressor := newMockArtifactCompressor(t)
				compressor.On("compressArtifact", mock.Anything, tmpdir).Return(nil) // the compressor should be called on the tempDir root
				execFactory := &blockingExecutorFactoryMock{}
				executor := &piperMocks.ExecMockRunner{}
				execFactory.On("CreateExecutor", mock.Anything, mock.Anything).Return(executor)
				cmdRunner := newDwCCommandExecutorMock(t)
				uploadResponse := ArtifactUploadResponse{
					AppName:       "test-service",
					ID:            "12345-678901234",
					CreatedVector: "01877e61-1bab-3f47-f0a3-123456789fe5",
					Type:          ArtifactTypeDocker,
					PromotionResult: []PromotionResultEntry{
						{
							Stage:    "dev/A",
							Status:   promotionResultStatusCreated,
							VectorId: "Aisah6Eo",
						},
						{
							Stage:    "dev/B",
							Status:   promotionResultStatusCreated,
							VectorId: "Komuid1g",
						},
						{
							Stage:    "dev/C",
							Status:   promotionResultStatusCreated,
							VectorId: "sgjklösdjgö3",
						},
					},
				}
				cmdRunner.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(
					func(executorFactory BlockingExecutorFactory, targetCmd dwcCommand, cliResponseTargetReference interface{}) error {
						if slices.Contains(targetCmd, uploadBaseCommand[1]) {
							uploadResp := cliResponseTargetReference.(*ArtifactUploadResponse)
							*uploadResp = uploadResponse
						}
						return nil
					})
				orchestrator := newMockStageWatchOrchestrator(t)
				orchestrator.On("WatchVectorDeployments", uploadResponse.PromotionResult, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil) // Vector Deployments should be watched with the filtered stages to watch
				return NewUploadController(compressor, treeEditor, mover, fsMock, matcher, orchestrator, execFactory, cmdRunner.Execute)
			},
			generateControllerInput: func(t *testing.T) ArtifactDescriptor {
				t.Helper()
				artifactDescriptor := &artifactDescriptorMock{}
				artifactDescriptor.On("hasFilePatterns").Return(true)
				artifactDescriptor.On("prepareFiles").Return(nil)
				artifactDescriptor.On("needsFileBundling").Return(true)
				artifactDescriptor.On("getArtifactUploadFolderStructure").Return(uploadFolderStructure)
				artifactDescriptor.On("getFilePatterns").Return(filePatterns)
				artifactDescriptor.On("getUploadFileName").Return("upload.zip")
				artifactDescriptor.On("hasStagesToWatch").Return(true)
				artifactDescriptor.On("getStagesToWatch").Return([]string{allStagesSelector})
				artifactDescriptor.On("buildUploadCommand").Return(dwcCommand(uploadBaseCommand), nil)
				return artifactDescriptor
			},
			wantErr: false,
			want: &ArtifactUploadResponse{
				AppName:       "test-service",
				ID:            "12345-678901234",
				CreatedVector: "01877e61-1bab-3f47-f0a3-123456789fe5",
				Type:          ArtifactTypeDocker,
				PromotionResult: []PromotionResultEntry{
					{
						Stage:    "dev/A",
						Status:   promotionResultStatusCreated,
						VectorId: "Aisah6Eo",
					},
					{
						Stage:    "dev/B",
						Status:   promotionResultStatusCreated,
						VectorId: "Komuid1g",
					},
					{
						Stage:    "dev/C",
						Status:   promotionResultStatusCreated,
						VectorId: "sgjklösdjgö3",
					},
				},
			},
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			controller := testCase.configureController(t)
			aDescriptor := testCase.generateControllerInput(t)
			got, err := controller.UploadArtifact(aDescriptor)
			if (err != nil) != testCase.wantErr {
				t.Fatalf("UploadArtifact() error = %v, wantErr %v", err, testCase.wantErr)
			}
			assert.Equal(t, testCase.want, got)
		})
	}
}

func TestUploadController_LoginToDwCCli(t *testing.T) {
	t.Parallel()
	const (
		wd = "/usr/local"
	)
	tests := []struct {
		name                    string
		configureController     func(t *testing.T) UploadController
		generateControllerInput func(t *testing.T) LoginDescriptor
		wantErr                 bool
	}{
		{
			name: "building and executing login command succeeds",
			configureController: func(t *testing.T) UploadController {
				fsMock := &piperMocks.FilesMock{CurrentDir: wd}
				treeEditor := newMockFileTreeEditor(t)
				matcher := &mockGlobMatcher{}
				mover := newMockFilePatternMover(t)
				compressor := newMockArtifactCompressor(t)
				execFactory := &blockingExecutorFactoryMock{}
				executor := &piperMocks.ExecMockRunner{}
				execFactory.On("CreateExecutor", mock.Anything, mock.Anything).Return(executor)
				cmdRunner := newDwCCommandExecutorMock(t)
				cmdRunner.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(
					func(executorFactory BlockingExecutorFactory, targetCmd dwcCommand, cliResponseTargetReference interface{}) error {
						return nil
					})
				orchestrator := newMockStageWatchOrchestrator(t)
				return NewUploadController(compressor, treeEditor, mover, fsMock, matcher, orchestrator, execFactory, cmdRunner.Execute)
			},
			generateControllerInput: func(t *testing.T) LoginDescriptor {
				loginDescriptor := newMockLoginDescriptor(t)
				loginDescriptor.On("buildLoginCommand").Return(dwcCommand{}, nil)
				return loginDescriptor
			},
			wantErr: false,
		},
		{
			name: "building login command fails",
			configureController: func(t *testing.T) UploadController {
				fsMock := &piperMocks.FilesMock{CurrentDir: wd}
				treeEditor := newMockFileTreeEditor(t)
				matcher := &mockGlobMatcher{}
				mover := newMockFilePatternMover(t)
				compressor := newMockArtifactCompressor(t)
				execFactory := &blockingExecutorFactoryMock{}
				executor := &piperMocks.ExecMockRunner{}
				execFactory.On("CreateExecutor", mock.Anything, mock.Anything).Return(executor)
				cmdRunner := newDwCCommandExecutorMock(t)
				orchestrator := newMockStageWatchOrchestrator(t)
				return NewUploadController(compressor, treeEditor, mover, fsMock, matcher, orchestrator, execFactory, cmdRunner.Execute)
			},
			generateControllerInput: func(t *testing.T) LoginDescriptor {
				loginDescriptor := newMockLoginDescriptor(t)
				loginDescriptor.On("buildLoginCommand").Return(nil, errors.New("failed to build login command"))
				return loginDescriptor
			},
			wantErr: true,
		},
		{
			name: "executing login command fails",
			configureController: func(t *testing.T) UploadController {
				fsMock := &piperMocks.FilesMock{CurrentDir: wd}
				treeEditor := newMockFileTreeEditor(t)
				matcher := &mockGlobMatcher{}
				mover := newMockFilePatternMover(t)
				compressor := newMockArtifactCompressor(t)
				execFactory := &blockingExecutorFactoryMock{}
				executor := &piperMocks.ExecMockRunner{}
				execFactory.On("CreateExecutor", mock.Anything, mock.Anything).Return(executor)
				cmdRunner := newDwCCommandExecutorMock(t)
				cmdRunner.On("Execute", mock.Anything, mock.Anything, mock.Anything).Return(
					func(executorFactory BlockingExecutorFactory, targetCmd dwcCommand, cliResponseTargetReference interface{}) error {
						return errors.New("login failed")
					})
				orchestrator := newMockStageWatchOrchestrator(t)
				return NewUploadController(compressor, treeEditor, mover, fsMock, matcher, orchestrator, execFactory, cmdRunner.Execute)
			},
			generateControllerInput: func(t *testing.T) LoginDescriptor {
				loginDescriptor := newMockLoginDescriptor(t)
				loginDescriptor.On("buildLoginCommand").Return(dwcCommand{}, nil)
				return loginDescriptor
			},
			wantErr: true,
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			controller := testCase.configureController(t)
			lDescriptor := testCase.generateControllerInput(t)
			err := controller.LoginToDwCCli(lDescriptor)
			if (err != nil) != testCase.wantErr {
				t.Fatalf("LoginToDwCCli() error = %v, wantErr %v", err, testCase.wantErr)
			}
		})
	}
}

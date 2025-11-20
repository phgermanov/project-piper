package dwc

import (
	"bytes"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"k8s.io/utils/strings/slices"
	"os"
	"path"
	"path/filepath"
)

type fileTreeEditor interface {
	MkdirAll(path string, perm os.FileMode) error
	TempDir(dir string, pattern string) (name string, err error)
}

type filePatternMover interface {
	move(objectsToMove []string, targetFolder string, executorFactory BlockingExecutorFactory) error
}

type wdResolver interface {
	Getwd() (string, error)
}

type globMatcher interface {
	Glob(pattern string) ([]string, error)
}

type DefaultFilePatternMover struct{}

func (mover DefaultFilePatternMover) move(objectsToMove []string, targetFolder string, executorFactory BlockingExecutorFactory) error {
	stdoutBuff := &bytes.Buffer{}
	stderrBuff := &bytes.Buffer{}
	executor := executorFactory.CreateExecutor(stdoutBuff, stderrBuff)
	args := append(append([]string{"-vn"}, objectsToMove...), targetFolder)
	if err := executor.RunExecutable("mv", args...); err != nil {
		return fmt.Errorf("execution of mv with args %v failed with exit code %d. captured stderr: %s. inner execution error: %w", args, executor.GetExitCode(), stderrBuff.String(), err)
	}
	log.Entry().Debugf("captured stdout of mv with args %v: %s", args, stdoutBuff.String())
	return nil
}

type DefaultGlobMatcher struct{}

func (matcher DefaultGlobMatcher) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

type UploadController struct {
	comp         ArtifactCompressor
	orchestrator StageWatchOrchestrator
	execFactory  BlockingExecutorFactory
	cliExecutor  CLICommandExecutor
	fileTreeEditor
	filePatternMover
	wdResolver
	globMatcher
}

func NewUploadController(comp ArtifactCompressor, fileTreeEditor fileTreeEditor, filePatternMover filePatternMover, wdResolver wdResolver, matcher globMatcher, orchestrator StageWatchOrchestrator, execFactory BlockingExecutorFactory, cliExecutor CLICommandExecutor) UploadController {
	return UploadController{comp: comp, fileTreeEditor: fileTreeEditor, filePatternMover: filePatternMover, wdResolver: wdResolver, globMatcher: matcher, orchestrator: orchestrator, execFactory: execFactory, cliExecutor: cliExecutor}
}

func (controller UploadController) LoginToDwCCli(loginDescriptor LoginDescriptor) error {
	loginCommand, err := loginDescriptor.buildLoginCommand()
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("failed to create login command: %w", err)
	}
	if err := controller.cliExecutor(controller.execFactory, loginCommand, nil); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("failed to execute login command: %w", err)
	}
	return nil
}

func (controller UploadController) UploadArtifact(artifactDescriptor ArtifactDescriptor) (*ArtifactUploadResponse, error) {
	if artifactDescriptor.hasFilePatterns() {
		log.Entry().Debugf("artifact descriptor has file patterns defined. Step is going to prepare and bundle files")
		if err := artifactDescriptor.prepareFiles(); err != nil {
			log.SetErrorCategory(log.ErrorCustom)
			return nil, fmt.Errorf("error preparing files for upload to dwc: %w", err)
		}
		if err := controller.bundleArtifactFiles(artifactDescriptor); err != nil {
			return nil, fmt.Errorf("error bundling artifact files for upload to dwc: %w", err)
		}
	} else if artifactDescriptor.needsFileBundling() {
		log.SetErrorCategory(log.ErrorConfiguration)
		return nil, fmt.Errorf("%T needs files for upload to dwc, but none where provided. You can specify files with the parameter artifactFilesToUpload", artifactDescriptor)
	}
	uploadCommand, err := artifactDescriptor.buildUploadCommand()
	if err != nil {
		log.SetErrorCategory(log.ErrorService)
		return nil, fmt.Errorf("failed to create upload command: %w", err)
	}
	var uploadResponse = &ArtifactUploadResponse{}
	if err := controller.cliExecutor(controller.execFactory, uploadCommand, uploadResponse); err != nil {
		log.SetErrorCategory(log.ErrorService)
		return nil, fmt.Errorf("failed to execute upload command: %w", err)
	}
	if err := controller.watchDeployments(uploadResponse, artifactDescriptor); err != nil {
		return nil, err
	}
	return uploadResponse, nil
}

func (controller UploadController) bundleArtifactFiles(descriptor ArtifactDescriptor) error {
	workingDir, err := controller.Getwd()
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return fmt.Errorf("failed to resolve working directory: %w", err)
	}
	log.Entry().Debugf("current working directory: %s", workingDir)

	tmpDirRoot, err := controller.TempDir(workingDir, "*")
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return fmt.Errorf("failed to create tempdir: %w", err)
	}

	uploadFolderRoot := path.Join(tmpDirRoot, descriptor.getArtifactUploadFolderStructure())
	if err := controller.MkdirAll(uploadFolderRoot, 0777); err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return fmt.Errorf("failed to create upload folder path: %w", err)
	}
	log.Entry().Debugf("created upload folder path: %s", uploadFolderRoot)
	stdoutBuff := &bytes.Buffer{}
	stderrBuff := &bytes.Buffer{}
	executor := controller.execFactory.CreateExecutor(stdoutBuff, stderrBuff)
	if err := executor.RunExecutable("ls", "-al"); err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return fmt.Errorf("ls -al execution failed with exit code %d. captured stderr: %s. inner execution error: %w", executor.GetExitCode(), stderrBuff.String(), err)
	}
	log.Entry().Debugf("%s", stdoutBuff.String())

	if err := controller.copyFilesByPattern(uploadFolderRoot, descriptor); err != nil {
		return fmt.Errorf("failed to copy files to upload folder: %w", err)
	}
	log.Entry().Debugf("copied all files matched by pattern(s) %+v successfully", descriptor.getFilePatterns())

	if err := controller.comp.compressArtifact(descriptor, tmpDirRoot); err != nil {
		return fmt.Errorf("failed to compress artifact: %w", err)
	}
	log.Entry().Debugf("compressed artifact files at %s successfully to %s", tmpDirRoot, descriptor.getUploadFileName())

	return nil
}

func (controller UploadController) copyFilesByPattern(targetFolder string, descriptor ArtifactDescriptor) error {
	var allMatches []string
	filePatterns := descriptor.getFilePatterns()
	for _, pattern := range filePatterns {
		matches, err := controller.Glob(pattern)
		if err != nil {
			log.SetErrorCategory(log.ErrorInfrastructure)
			return fmt.Errorf("failed trying to match artifact pattern: %w", err)
		}
		if len(matches) == 0 {
			log.Entry().Warnf("no matches found for file pattern %s. no files for this file pattern will uploaded", pattern)
		} else {
			allMatches = append(allMatches, matches...)
			log.Entry().Debugf("found %d matches for file pattern %s", len(matches), pattern)
			log.Entry().Debugf("matches are %v", matches)
		}
	}
	if len(allMatches) == 0 && descriptor.needsFileBundling() {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("artifact needs file bundling but no file pattern of %v matched", filePatterns)
	}
	if len(allMatches) == 0 {
		log.Entry().Warnf("no matches found for any of the specified file patterns: %v. no files will be uploaded", filePatterns)
		return nil
	}
	if err := controller.move(allMatches, targetFolder, controller.execFactory); err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return fmt.Errorf("failed to move files to upload folder: %w", err)
	}
	return nil
}

func (controller UploadController) watchDeployments(uploadResponse *ArtifactUploadResponse, artifactDescriptor ArtifactDescriptor) error {
	if !artifactDescriptor.hasStagesToWatch() {
		log.Entry().Debug("no stages to watch. Skip watching stages")
		return nil
	}
	userDefinedStagesToWatch := artifactDescriptor.getStagesToWatch()
	stagesToWatch, allStagesSelectorDetected := controller.filterForStagesToWatch(userDefinedStagesToWatch, uploadResponse.PromotionResult)
	if !allStagesSelectorDetected && len(stagesToWatch) != len(userDefinedStagesToWatch) {
		log.Entry().Warnf("User defined stages to watch are %v, but a promotion was not triggered for all of those. In fact, the following promotions where triggered %v. But we try our best and watch the vector deployments to the following stages: %+v", userDefinedStagesToWatch, uploadResponse.PromotionResult, stagesToWatch)
	}
	return controller.orchestrator.WatchVectorDeployments(stagesToWatch, log.Writer(), controller.cliExecutor, controller.execFactory, artifactDescriptor)
}

// filterForStagesToWatch extracts the predefined stages to watch from the upload response
func (controller UploadController) filterForStagesToWatch(userDefinedStagesToWatch []string, uploadResponse []PromotionResultEntry) ([]PromotionResultEntry, bool) {
	allStagesSelectorDetected := slices.Contains(userDefinedStagesToWatch, allStagesSelector)
	if allStagesSelectorDetected {
		return uploadResponse, allStagesSelectorDetected
	}
	var stagesToWatch []PromotionResultEntry
onNextUserDefinedStage:
	for _, userDefinedStageToWatch := range userDefinedStagesToWatch {
		for _, promotionResultFromUpload := range uploadResponse {
			if promotionResultFromUpload.Stage == userDefinedStageToWatch {
				stagesToWatch = append(stagesToWatch, promotionResultFromUpload)
				continue onNextUserDefinedStage
			}
		}
	}
	return stagesToWatch, allStagesSelectorDetected
}

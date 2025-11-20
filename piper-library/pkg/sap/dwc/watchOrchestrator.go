package dwc

import (
	"bytes"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"io"
	"strings"
	"sync"
	"time"
)

// StageWatchOrchestrator returns an error if the artifact descriptors watch policy has been violated.
type StageWatchOrchestrator interface {
	WatchVectorDeployments([]PromotionResultEntry, io.Writer, CLICommandExecutor, BlockingExecutorFactory, ArtifactDescriptor) error
}

const (
	// stageWatchLockMinutes defines the amount of time a temporary usage is set onto any target stage to watch.
	// it should always match the timeout specified with the watch command.
	stageWatchLockMinutes = 60
	lineBreak             = 0x0A
	dot                   = 0x2D
)

var (
	promotionResultStatusSuccessMsgTpl = fmt.Sprintf("Promotion result is in state %s. Seems like nothing changed. Skip watching stage.", promotionResultStatusSuccess)
	promotionResultStatusErrorMsgTpl   = "promotion result is in state %s. No vector deployment was triggered. The error is: %s"
	unknownPromotionResultMsgTpl       = "promotion result is in unknown state %s. Please contact the dwc team"
	orbitDeploymentMsgTpl              = "Your orbit deployment to %s was created. Watching orbit deployments while they are in-flight is currently not supported. Skip watching."
	stageWatchPrefixROITpl             = "Deployment of resource %s to stage %s with current vector %s %s."
	stageWatchPrefixTpl                = "Deployment to stage %s with current vector %s %s."
)

const (
	watchResultPrefixLineSuccess = "succeeded"
	watchResultPrefixLineFailure = "failed"
	watchResultPrefixLineUnknown = "is in unknown result state"
)

// bufferEntry represents the state of an in-flight stage watch operation
type bufferEntry struct {
	PromotionResultEntry
	failed    bool
	logBuffer *bytes.Buffer
	errWriter io.Writer
}

func (buffEntry *bufferEntry) reportError(err error) {
	buffEntry.failed = true
	_, _ = buffEntry.errWriter.Write(append([]byte(err.Error()), lineBreak)) // swallow the error as we can not do anything except panic about it
}

func (buffEntry *bufferEntry) watchVectorDeployment(dwcCommandExecutor CLICommandExecutor, executorFactory BlockingExecutorFactory, artifactDescriptor ArtifactDescriptor) {
	var waitForDeploymentCommandResponse = &WaitForDeploymentResponse{}
	if err := dwcCommandExecutor(executorFactory, newWaitForDeploymentCommand(buffEntry.Stage, buffEntry.VectorId), waitForDeploymentCommandResponse); err != nil {
		log.SetErrorCategory(log.ErrorCustom)
		buffEntry.reportError(&preStageWatchError{stageWatchError{
			stage: buffEntry.Stage,
			inner: err,
		}})
		return
	}
	landscape := waitForDeploymentCommandResponse.Landscape
	if landscape == "" {
		buffEntry.reportError(&preStageWatchError{stageWatchError{
			stage: buffEntry.Stage,
			inner: errors.New("no deployment was started for vector. Therefore, no landscape is provided"),
		}})
		return
	}
	expiry := time.Now().UTC().Add(time.Duration(stageWatchLockMinutes) * time.Minute).Format(time.RFC3339)
	if err := dwcCommandExecutor(executorFactory, newAddVectorUsageCommand(landscape, buffEntry.VectorId, expiry), nil); err != nil {
		log.SetErrorCategory(log.ErrorService)
		buffEntry.reportError(&preStageWatchError{stageWatchError{
			stage: buffEntry.Stage,
			inner: err,
		}})
		return
	}
	defer func() {
		if err := dwcCommandExecutor(executorFactory, newRemoveVectorUsageCommand(landscape, buffEntry.VectorId, expiry), nil); err != nil {
			log.SetErrorCategory(log.ErrorService)
			buffEntry.reportError(&postStageWatchError{stageWatchError{
				stage: buffEntry.Stage,
				inner: fmt.Errorf("failed to remove usage %s from stage %s in landscape %s targeting vector %s. Either delete the usage manually by installing the dwc cli https://github.tools.sap/deploy-with-confidence/cli and run %s %s -h or wait until %s for the usage to expire. Error was %s", pipelineUsageName, buffEntry.Stage, landscape, buffEntry.VectorId, deploymentVectorBaseCommand, removeVectorUsageSubcommand, expiry, err),
			}})
		}
	}()
	var liveLogs string
	if err := dwcCommandExecutor(executorFactory, newWatchVectorDeploymentCommand(landscape, buffEntry.VectorId, artifactDescriptor), &liveLogs); err != nil {
		log.SetErrorCategory(log.ErrorCustom)
		buffEntry.reportError(&stageWatchError{
			stage: buffEntry.Stage,
			inner: err,
		})
	}
	if _, err := buffEntry.logBuffer.Write([]byte(liveLogs)); err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		buffEntry.reportError(fmt.Errorf("unable to print live logs: %w", &stageWatchError{
			stage: buffEntry.Stage,
			inner: err,
		}))
	}
}

func (buffEntry *bufferEntry) watchOrbitDeployment(dwcCommandExecutor CLICommandExecutor, executorFactory BlockingExecutorFactory, artifactDescriptor ArtifactDescriptor) { // TODO: Implement. Needs CLI extension. Also DwC API Gateway needs to be inplace to allow authorized calls against Phil from the CLI.
	if _, err := buffEntry.logBuffer.Write([]byte(fmt.Sprintf(orbitDeploymentMsgTpl, buffEntry.Stage))); err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		buffEntry.reportError(fmt.Errorf("unable to print logs. Promotion result nevertheless is in state %s: %w", promotionResultStatusCreated, &stageWatchError{
			stage: buffEntry.Stage,
			inner: err,
		}))
	}
}

func (buffEntry *bufferEntry) printResult() string {
	if buffEntry.failed {
		return watchResultPrefixLineFailure
	}
	if strings.HasPrefix(buffEntry.Stage, "orbit/") && strings.ToLower(buffEntry.Status) == promotionResultStatusCreated {
		return watchResultPrefixLineUnknown
	}
	return watchResultPrefixLineSuccess
}

func (buffEntry *bufferEntry) succeeded() bool {
	return !buffEntry.failed
}

func (buffEntry *bufferEntry) getStageName() string {
	return buffEntry.Stage
}

func (buffEntry *bufferEntry) createStageWatchOutputPrefixReader(artifactDescriptor ArtifactDescriptor) io.Reader {
	var statusLine []byte
	if artifactDescriptor.watchROIOnly() {
		statusLine = []byte(fmt.Sprintf(stageWatchPrefixROITpl, artifactDescriptor.GetResourceName(), buffEntry.Stage, buffEntry.VectorId, buffEntry.printResult()))
	} else {
		statusLine = []byte(fmt.Sprintf(stageWatchPrefixTpl, buffEntry.Stage, buffEntry.VectorId, buffEntry.printResult()))
	}
	return bytes.NewReader(assembleBytesForStageWatchOutputPrefixReader(statusLine))
}

type watchBuffer struct {
	storage map[int]*bufferEntry
	rwMutex *sync.RWMutex
}

func (buffer watchBuffer) toWatchResults() []watchResult {
	entries := make([]watchResult, 0, len(buffer.storage))
	for idx := range buffer.storage {
		entry := buffer.get(idx)
		entries = append(entries, entry)
	}
	return entries
}

func (buffer watchBuffer) get(idx int) *bufferEntry {
	buffer.rwMutex.RLock()
	defer buffer.rwMutex.RUnlock()
	return buffer.storage[idx]
}

func (buffer watchBuffer) set(idx int, entry *bufferEntry) {
	buffer.rwMutex.Lock()
	defer buffer.rwMutex.Unlock()
	buffer.storage[idx] = entry
}

type DefaultStageWatchOrchestrator struct {
	wg          *sync.WaitGroup
	watchBuffer watchBuffer
}

func (orchestrator *DefaultStageWatchOrchestrator) writeResults(resultWriter io.Writer, artifactDescriptor ArtifactDescriptor) error {
	var readers []io.Reader
	for idx := range orchestrator.watchBuffer.storage {
		entry := orchestrator.watchBuffer.get(idx)
		readers = append(readers, entry.createStageWatchOutputPrefixReader(artifactDescriptor), bytes.NewReader(entry.logBuffer.Bytes()))
	}
	readers = append(readers, bytes.NewReader([]byte{lineBreak})) // last line break is needed to ensure that the buffer is flushed completely in case of github.com/sap/jenkins-library/pkg/log/writer.go
	if _, err := io.Copy(resultWriter, io.MultiReader(readers...)); err != nil {
		return fmt.Errorf("failed to write stage watch results: %w", err)
	}
	return nil
}

func (orchestrator *DefaultStageWatchOrchestrator) WatchVectorDeployments(promotionResults []PromotionResultEntry, resultWriter io.Writer, dwcCommandExecutor CLICommandExecutor, executorFactory BlockingExecutorFactory, artifactDescriptor ArtifactDescriptor) error {
	orchestrator.prepare()
	syncedDst := &syncedWriter{
		protectedDst: resultWriter,
		mux:          &sync.Mutex{},
	}
	for idx, result := range promotionResults {
		logBuff := &bytes.Buffer{}
		orchestrator.watchBuffer.set(idx, &bufferEntry{
			PromotionResultEntry: result,
			failed:               false,
			logBuffer:            logBuff,
			errWriter:            io.MultiWriter(syncedDst, logBuff), // Multiplex to stage scoped logs and live logs while stage watches are still in-flight
		})
		orchestrator.wg.Add(1)
		go func(buffIdx int) {
			orchestrator.watchStage(buffIdx, dwcCommandExecutor, executorFactory, artifactDescriptor)
		}(idx)
	}
	orchestrator.wg.Wait()
	if err := orchestrator.writeResults(syncedDst, artifactDescriptor); err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return err
	}
	watchPolicy := artifactDescriptor.getStageWatchPolicy()
	return watchPolicy(orchestrator.watchBuffer.toWatchResults())
}

func (orchestrator *DefaultStageWatchOrchestrator) prepare() {
	orchestrator.wg = &sync.WaitGroup{}
	orchestrator.watchBuffer = watchBuffer{
		storage: make(map[int]*bufferEntry),
		rwMutex: &sync.RWMutex{},
	}
}

func (orchestrator *DefaultStageWatchOrchestrator) watchStage(bufferIndex int, dwcCommandExecutor CLICommandExecutor, executorFactory BlockingExecutorFactory, artifactDescriptor ArtifactDescriptor) {
	defer orchestrator.wg.Done()
	buffEntry := orchestrator.watchBuffer.get(bufferIndex)
	switch strings.ToLower(buffEntry.Status) {
	case promotionResultStatusCreated:
		if strings.HasPrefix(buffEntry.Stage, "orbit/") {
			buffEntry.watchOrbitDeployment(dwcCommandExecutor, executorFactory, artifactDescriptor)
		} else {
			buffEntry.watchVectorDeployment(dwcCommandExecutor, executorFactory, artifactDescriptor)
		}
	case promotionResultStatusSuccess:
		if _, err := buffEntry.logBuffer.Write([]byte(promotionResultStatusSuccessMsgTpl)); err != nil {
			log.SetErrorCategory(log.ErrorInfrastructure)
			buffEntry.reportError(fmt.Errorf("unable to print logs. Promotion result nevertheless is in state %s: %w", promotionResultStatusSuccess, &stageWatchError{
				stage: buffEntry.Stage,
				inner: err,
			}))
		}
	case promotionResultStatusError:
		log.SetErrorCategory(log.ErrorCustom)
		buffEntry.reportError(&stageWatchError{
			stage: buffEntry.Stage,
			inner: fmt.Errorf(promotionResultStatusErrorMsgTpl, promotionResultStatusError, buffEntry.Error),
		})
	default:
		log.SetErrorCategory(log.ErrorService)
		buffEntry.reportError(&stageWatchError{
			stage: buffEntry.Stage,
			inner: fmt.Errorf(unknownPromotionResultMsgTpl, buffEntry.Status),
		})
	}
}

func assembleBytesForStageWatchOutputPrefixReader(statusLine []byte) []byte {
	targetBytes := []byte{lineBreak, lineBreak}
	dottedLine := getDottedLine(len(statusLine))
	targetBytes = append(targetBytes, statusLine...)
	targetBytes = append(targetBytes, lineBreak)
	targetBytes = append(targetBytes, dottedLine...)
	targetBytes = append(targetBytes, lineBreak, lineBreak)
	return targetBytes
}

func getDottedLine(numDots int) []byte {
	line := make([]byte, 0, numDots)
	for i := 0; i < numDots; i++ {
		line = append(line, dot)
	}
	return line
}

package dwc

import "fmt"

type stageWatchError struct {
	stage string
	inner error
}

func (err *stageWatchError) Error() string {
	return fmt.Sprintf("failed watching stage %s: %s", err.stage, err.inner)
}

func (err *stageWatchError) Unwrap() error { return err.inner }

type preStageWatchError struct {
	stageWatchError
}

func (err *preStageWatchError) Error() string {
	return fmt.Sprintf("failed to start watching stage %s: %s", err.stage, err.inner)
}

type postStageWatchError struct {
	stageWatchError
}

func (err *postStageWatchError) Error() string {
	return fmt.Sprintf("failed to perform cleanup tasks after watching stage %s: %s", err.stage, err.inner)
}

type stageWatchPolicyViolationError struct {
	reason string
}

func (err *stageWatchPolicyViolationError) Error() string {
	return fmt.Sprintf("sapDwcStageRelease failed after watching deployments: %s", err.reason)
}

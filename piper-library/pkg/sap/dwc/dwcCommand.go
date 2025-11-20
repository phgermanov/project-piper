package dwc

import (
	"bytes"
	"encoding/json"
	"fmt"
)

var targetBinary = "/tmp/dwc"

const (
	outputFlag       = "--output=%s"
	outputFormatJSON = "json"
)

type dwcCommand []string

type CLICommandExecutor func(executorFactory BlockingExecutorFactory, targetCmd dwcCommand, cliResponseTargetReference any) error

func DefaultCLICommandExecutor(executorFactory BlockingExecutorFactory, targetCmd dwcCommand, cliResponseTargetReference any) error {
	stdoutBuff := &bytes.Buffer{}
	stderrBuff := &bytes.Buffer{}
	executor := executorFactory.CreateExecutor(stdoutBuff, stderrBuff)
	var execError error
	if err := executor.RunExecutable(targetBinary, targetCmd...); err != nil {
		execError = fmt.Errorf("execution failed with exit code %d. captured stderr: %s. inner execution error: %w", executor.GetExitCode(), stderrBuff.String(), err)
	}
	if cliResponseTargetReference != nil {
		if json.Valid(stdoutBuff.Bytes()) {
			if err := json.Unmarshal(stdoutBuff.Bytes(), cliResponseTargetReference); err != nil {
				return fmt.Errorf("failed to unmarshal stdout: %w. captured stderr: %s, captured stdout: %s. Error during binary execution: %s", err, stderrBuff.String(), stdoutBuff.String(), execError)
			}
		} else {
			strPointer, ok := cliResponseTargetReference.(*string)
			if !ok {
				return fmt.Errorf("CLI response is not a valid json response, but also failed to interpret CLI response as string. Wanted to unmarshal result into type %T. CLI Response (stdout) to unmarshal is %s. stderr is %s. Error during binary execution: %s", cliResponseTargetReference, stdoutBuff.String(), stderrBuff.String(), execError)
			}
			*strPointer = stdoutBuff.String()
		}
	}
	return execError
}

func SetTargetBinary(path string) {
	targetBinary = path
}

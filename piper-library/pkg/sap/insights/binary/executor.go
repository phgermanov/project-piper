package binary

import (
	"github.com/SAP/jenkins-library/pkg/command"
	"io"
)

// BlockingExecutorFactory creates a BlockingExecutor scoped with a given set of stdout and stderr
type BlockingExecutorFactory interface {
	CreateExecutor(stdout, stderr io.Writer) BlockingExecutor
}

// BlockingExecutor creates a process synchronously: it blocks until completion
type BlockingExecutor interface {
	GetExitCode() int
	RunExecutable(executable string, params ...string) error
}

type DefaultExecutorFactory struct{}

func (factory DefaultExecutorFactory) CreateExecutor(stdOut io.Writer, stderr io.Writer) BlockingExecutor {
	cmd := &command.Command{}
	cmd.Stdout(stdOut)
	cmd.Stderr(stderr)
	return cmd
}

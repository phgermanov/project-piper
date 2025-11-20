package agent

import "github.com/SAP/jenkins-library/pkg/command"

type Executor interface {
	GetExitCode() int
	RunExecutable(executable string, params ...string) error
	SetEnv(env []string)
}

func DefaultExecutor() Executor {
	cmd := &command.Command{}
	return cmd
}

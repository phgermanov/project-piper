package agent

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

const (
	executableFileName = "cumulus-policy-agent"
)

type Agent interface {
	Install(version string) error
	Execute(parameter []string, environmentVariables []string) error
}

type agent struct {
	resolver   Resolver
	executor   Executor
	executable string
	files      piperutils.FileUtils
}

func (agent *agent) Install(version string) error {

	executable := buildExecutablePath(executableFileName, version)

	log.Entry().Debugf("Install cumulus policy agent to path \"%s\"...", executable)

	err := deleteOutdatedBinary(executable)
	if err != nil {
		return err
	}

	// check if exists
	exists, err := agent.files.FileExists(executable)
	if err != nil {
		return err
	}

	if !exists {
		// if not exists => download + add executable flag
		log.Entry().Debugf("Download cumulus policy agent version \"%s\"...", version)

		err = agent.resolver.Download(version, executable)
		if err != nil {
			return err
		}
		err = agent.files.Chmod(executable, 0755)

		if err != nil {
			return err
		}

	} else {
		log.Entry().Debugf("Download of cumulus policy agent skipped (already installed and up to date)")
	}
	agent.executable = executable

	// execute version command for debugging purposes
	err = agent.Execute([]string{"version", "--short"}, nil)
	if err != nil {
		return err
	}

	log.Entry().Debugf("Installation of cumulus policy agent successful!")

	return nil
}

func (agent *agent) Execute(parameter []string, environmentVariables []string) error {

	if agent.executable == "" {
		return fmt.Errorf("cumulus policy agent not installed! (run Install() first))")
	}

	agent.executor.SetEnv(environmentVariables)
	err := agent.executor.RunExecutable(agent.executable, parameter...)
	agent.executor.SetEnv(nil)

	if err != nil {
		return fmt.Errorf("execution failed with exit code %d. inner execution error: %w", agent.executor.GetExitCode(), err)
	}

	return err
}

func deleteOutdatedBinary(name string) error {
	file, err := os.Stat(name)
	if err != nil {
		return nil
	}
	modifiedtime := file.ModTime()
	if time.Since(modifiedtime).Hours() > 1 {
		return os.Remove(name)
	}
	return nil
}

func buildExecutablePath(name, version string) string {
	var executablePath string
	if version == "latest" {
		executablePath = fmt.Sprintf("./%s", name)
	} else {
		executablePath = fmt.Sprintf("./%s-%s", name, version)
	}
	if runtime.GOOS == "windows" {
		executablePath += ".exe"
	}
	return executablePath
}

func NewAgent(resolver Resolver) Agent {
	return &agent{resolver: resolver, executor: DefaultExecutor(), files: &piperutils.Files{}}
}

func NewTestAgent(resolver Resolver, executor Executor, fileUtils piperutils.FileUtils) Agent {
	return &agent{resolver: resolver, executor: executor, files: fileUtils}
}

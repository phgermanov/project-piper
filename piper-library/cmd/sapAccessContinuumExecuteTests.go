package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type sapAccessContinuumExecuteTestsUtils interface {
	command.ShellRunner
	piperutils.FileUtils
}

type sapAccessContinuumExecuteTestsUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newSapAccessContinuumExecuteTestsUtils() sapAccessContinuumExecuteTestsUtils {
	utils := sapAccessContinuumExecuteTestsUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}

	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func sapAccessContinuumExecuteTests(config sapAccessContinuumExecuteTestsOptions, telemetryData *telemetry.CustomData) {

	utils := newSapAccessContinuumExecuteTestsUtils()

	err := runSapAccessContinuumExecuteTests(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runSapAccessContinuumExecuteTests(config *sapAccessContinuumExecuteTestsOptions, telemetryData *telemetry.CustomData, utils sapAccessContinuumExecuteTestsUtils) error {
	switch config.BuildTool {
	case "npm":
		// Install command and Runcommand triggers the Access continuum to run the tests
		shellCommand := utils.RunShell("/bin/bash", fmt.Sprintf(" %s; %s;", config.InstallCommand, config.RunCommand))
		if shellCommand != nil {
			return fmt.Errorf("Error while running commands in docker %w", shellCommand)
		}
		return nil
	default:
		return fmt.Errorf("build tool should be npm")
	}
}

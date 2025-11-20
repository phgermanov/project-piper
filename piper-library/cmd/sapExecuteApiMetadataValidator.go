package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

type sapExecuteApiMetadataValidatorUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
}

type sapExecuteApiMetadataValidatorUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newSapExecuteApiMetadataValidatorUtils() sapExecuteApiMetadataValidatorUtils {
	utils := sapExecuteApiMetadataValidatorUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func sapExecuteApiMetadataValidator(config sapExecuteApiMetadataValidatorOptions, telemetryData *telemetry.CustomData) {
	utils := newSapExecuteApiMetadataValidatorUtils()

	err := runSapExecuteApiMetadataValidator(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func removeEmptyStrings(input []string) []string {
	result := []string{}

	for _, s := range input {
		if strings.TrimSpace(s) != "" {
			result = append(result, s)
		}
	}
	return result
}

func runSapExecuteApiMetadataValidator(config *sapExecuteApiMetadataValidatorOptions, telemetryData *telemetry.CustomData, utils sapExecuteApiMetadataValidatorUtils) error {
	var err error
	args := []string{}
	files := removeEmptyStrings(config.Files)

	// Always output in JSON format as it's used to upload results in Cumulus
	args = append(args, "--format", "json")

	// Always write results into a file (name must NOT change as Cumulus depends on it)
	args = append(args, "--output", "api-metadata-validator-results.json")

	if len(files) == 0 {
		return fmt.Errorf("missing mandatory parameter 'files'. This parameter is used to indicate what files need to be validated.")
	}

	if strings.TrimSpace(config.Ruleset) != "" {
		args = append(args, "--ruleset", config.Ruleset)
	}

	if strings.TrimSpace(config.FailSeverity) != "" {
		args = append(args, "--fail-severity", config.FailSeverity)
	}

	if config.Quiet {
		log.Entry().Info("quiet mode is enabled, only rules with severity 'Error' will be run and reported")
		args = append(args, "--quiet")
	}

	// Files argument MUST be always at the end
	args = append(args, files...)

	log.Entry().Info("Starting SAP API Metadata Validator")

	output := bytes.Buffer{}
	utils.Stdout(io.MultiWriter(&output, log.Writer()))

	err = utils.RunExecutable("validator", args...)

	if config.PrintToConsole {
		errPrint := utils.RunExecutable("formatter", "--format", "text", "./api-metadata-validator-results.json")
		if errPrint != nil {
			return fmt.Errorf("the execution of the formatter failed, see the log for details: %w", errPrint)
		}
	}

	if err != nil {
		return fmt.Errorf("the execution of the api-metadata-validator failed, see the log for details: %w", err)
	}

	return nil
}

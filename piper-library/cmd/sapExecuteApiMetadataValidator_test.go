//go:build unit
// +build unit

package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type sapExecuteApiMetadataValidatorMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newSapExecuteApiMetadataValidatorTestsUtils() sapExecuteApiMetadataValidatorMockUtils {
	utils := sapExecuteApiMetadataValidatorMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunSapExecuteApiMetadataValidator(t *testing.T) {
	t.Parallel()

	t.Run("Parameter `Files` is required", func(t *testing.T) {
		t.Parallel()

		config := sapExecuteApiMetadataValidatorOptions{}

		utils := newSapExecuteApiMetadataValidatorTestsUtils()

		err := runSapExecuteApiMetadataValidator(&config, nil, utils)

		assert.Error(t, err, "Missing mandatory parameter 'files'. This parameter is used to indicate what files need to be validated.")
	})

	t.Run("Files included in the binary call", func(t *testing.T) {
		t.Parallel()

		config := sapExecuteApiMetadataValidatorOptions{}
		config.Files = []string{"foo.json", "bar.yaml"}

		utils := newSapExecuteApiMetadataValidatorTestsUtils()

		err := runSapExecuteApiMetadataValidator(&config, nil, utils)

		assert.NoError(t, err, "No execution error")
		assert.Contains(t, utils.Calls[0].Params, "foo.json", "Args contain foo.json")
		assert.Contains(t, utils.Calls[0].Params, "bar.yaml", "Args contain bar.yaml")
	})

	t.Run("Empty optional parameters not included in the binary call", func(t *testing.T) {
		t.Parallel()

		config := sapExecuteApiMetadataValidatorOptions{}
		config.Files = []string{"foo.json", "bar.yaml"}
		config.Ruleset = "	   " // tabulation & spaces
		config.FailSeverity = "\n"

		utils := newSapExecuteApiMetadataValidatorTestsUtils()

		err := runSapExecuteApiMetadataValidator(&config, nil, utils)

		assert.NoError(t, err, "No execution error")
		assert.NotContains(t, utils.Calls[0].Params, "--ruleset", "Args do not contain `--ruleset`")
		assert.NotContains(t, utils.Calls[0].Params, "--fail-severity", "Args do not contain `--fail-severity`")
	})

	t.Run("Parameters should be passed to the executable", func(t *testing.T) {
		t.Parallel()

		config := sapExecuteApiMetadataValidatorOptions{}
		config.Files = []string{"foo.json", "bar.yaml"}
		config.Ruleset = "sap:core:v1"
		config.FailSeverity = "warning"
		config.Quiet = true

		m := make(map[string]string)
		m["--ruleset"] = config.Ruleset
		m["--fail-severity"] = config.FailSeverity

		utils := newSapExecuteApiMetadataValidatorTestsUtils()

		err := runSapExecuteApiMetadataValidator(&config, nil, utils)

		assert.NoError(t, err, "No execution error")

		// Expected 11 params/words: foo.json bar.yaml --format json --output api-metadata-validator-results.json --ruleset sap:core:v1 --fail-severity warning --quiet
		assert.Equal(t, 11, len(utils.Calls[0].Params), "Number of arguments is correct")

		// Check correct values for parameters without binding to their location in the arguments string
		for _, pos := range []int{4, 6} {
			argName := utils.Calls[0].Params[pos]
			assert.True(t, strings.HasPrefix(argName, "--"), "Argument name is prefixed correctly")
			assert.Equal(t, m[argName], utils.Calls[0].Params[pos+1], "Argument value is correct")
		}

		// Check correct values for boolean parameters
		assert.True(t, utils.Calls[0].Params[8] == "--quiet", "Quiet was passed correctly")

		// Files arguments must always go last
		assert.True(t, utils.Calls[0].Params[9] == "foo.json" || utils.Calls[0].Params[9] == "bar.yaml", "First argument is a file")
		assert.True(t, utils.Calls[0].Params[10] == "foo.json" || utils.Calls[0].Params[10] == "bar.yaml", "Second argument is a file")
	})

	t.Run("Error handling", func(t *testing.T) {
		t.Parallel()

		config := sapExecuteApiMetadataValidatorOptions{}
		config.Files = []string{"foo.json", "bar.yaml"}

		utils := newSapExecuteApiMetadataValidatorTestsUtils()
		utils.ShouldFailOnCommand = map[string]error{"validator": fmt.Errorf("some error")}

		err := runSapExecuteApiMetadataValidator(&config, nil, utils)

		assert.EqualError(t, err, "the execution of the api-metadata-validator failed, see the log for details: some error")
	})

	t.Run("Print to console is triggered", func(t *testing.T) {
		t.Parallel()

		config := sapExecuteApiMetadataValidatorOptions{}
		config.Files = []string{"foo.json", "bar.yaml"}
		config.PrintToConsole = true

		utils := newSapExecuteApiMetadataValidatorTestsUtils()

		err := runSapExecuteApiMetadataValidator(&config, nil, utils)

		assert.NoError(t, err, "No execution error")

		// Expected 6 params/words: foo.json bar.yaml --format json --output api-metadata-validator-results.json
		assert.Equal(t, 6, len(utils.Calls[0].Params), "Number of validator arguments is correct")

		// Expected 3 params: --format text api-metadata-validator-results.json
		assert.Equal(t, 3, len(utils.Calls[1].Params), "Number of formatter arguments is correct")
	})

	t.Run("Failing when printing to console", func(t *testing.T) {
		t.Parallel()

		config := sapExecuteApiMetadataValidatorOptions{}
		config.Files = []string{"foo.json", "bar.yaml"}
		config.PrintToConsole = true

		utils := newSapExecuteApiMetadataValidatorTestsUtils()
		utils.ShouldFailOnCommand = map[string]error{"formatter": fmt.Errorf("some error")}

		err := runSapExecuteApiMetadataValidator(&config, nil, utils)

		assert.EqualError(t, err, "the execution of the formatter failed, see the log for details: some error")
	})
}

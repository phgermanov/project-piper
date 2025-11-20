package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type sapAccessContinuumExecuteTestsMockUtils struct {
	*mock.ShellMockRunner
	*mock.FilesMock
	config sapAccessContinuumExecuteTestsOptions
}

func newSapAccessContinuumExecuteTestsTestsUtils() sapAccessContinuumExecuteTestsMockUtils {
	utils := sapAccessContinuumExecuteTestsMockUtils{
		ShellMockRunner: &mock.ShellMockRunner{},
		FilesMock:       &mock.FilesMock{},
	}
	return utils
}

func TestRunSapAccessContinuumExecuteTests(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := sapAccessContinuumExecuteTestsOptions{
			BuildTool:      "npm",
			InstallCommand: "npm install",
			RunCommand:     "npm build",
		}
		utils := newSapAccessContinuumExecuteTestsTestsUtils()

		// test
		err := runSapAccessContinuumExecuteTests(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := sapAccessContinuumExecuteTestsOptions{}

		utils := newSapAccessContinuumExecuteTestsTestsUtils()

		// test
		err := runSapAccessContinuumExecuteTests(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "build tool should be npm")
	})
}

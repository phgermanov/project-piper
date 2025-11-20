//go:build unit
// +build unit

package cmd

import (
	"os"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type sapSUPAExecuteTestsMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newSapSUPAExecuteTestsTestsUtils() sapSUPAExecuteTestsMockUtils {
	utils := sapSUPAExecuteTestsMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestSapSUPAExecuteTests(t *testing.T) {
	t.Parallel()

	t.Run("Simple run through all types", func(t *testing.T) {

		dir, err := os.MkdirTemp("", "")
		if err != nil {
			t.Fatal("failed to get temporary directory")
		}
		defer os.RemoveAll(dir) // clean up

		t.Run("New SAP SUPA Execute Tests Utils returns sapSUPAExecuteTestsUtilsBundle", func(t *testing.T) {
			t.Parallel()
			utils := newSapSUPAExecuteTestsUtils()
			assert.NotNil(t, utils)
		})

		t.Run("test env params", func(t *testing.T) {
			t.Parallel()
			// init
			const EXPECTED_S = "TEST_1=testing1 TEST_2 TEST_3=testing3= TEST_4=testing$1"
			envStr := "  TEST_1=testing1     TEST_2  TEST_3=testing3=  TEST_4=testing$1"
			// test
			cmdStr := getEnvVar(envStr)
			// assert
			assert.Equal(t, cmdStr, EXPECTED_S, "Strings shall be identical")
		})

		t.Run("test docker env params", func(t *testing.T) {
			t.Parallel()
			// init
			const EXPECTED_L = 4
			config := sapSUPAExecuteTestsOptions{}
			config.ScriptType = "default"
			envArr := []string{"Abc", "Xyz", "Mno"}
			// test
			cmdEnv := getDockerEnvParams(&config, envArr)
			val := len(cmdEnv)
			// assert
			assert.EqualValues(t, EXPECTED_L, val)
		})

		t.Run("test git params - no token", func(t *testing.T) {
			t.Parallel()
			// init
			config := sapSUPAExecuteTestsOptions{}
			config.ScriptType = "default"
			envArr := []string{"Abc", "Xyz", "Mno"}
			// test
			envs, err := getGitParams(&config, envArr, 0)
			val := len(envs)
			_ = val
			// assert
			// assert.EqualError(t, err, "PerformanceSingleUserTest error: GithubToken not defined")
			assert.NoError(t, err)
		})

		t.Run("test git params - repo but not token", func(t *testing.T) {
			t.Parallel()
			// init
			config := sapSUPAExecuteTestsOptions{}
			config.ScriptType = "default"
			envArr := []string{"Abc", "Xyz", "Mno"}
			config.TestRepository = "testRepo"
			config.GithubToken = ""
			// test
			envs, err := getGitParams(&config, envArr, 0)
			val := len(envs)
			_ = val
			// assert
			assert.EqualError(t, err, "PerformanceSingleUserTest error: repository and token not defined")
		})

		t.Run("test git params for type all - config values included in result cmd", func(t *testing.T) {
			t.Parallel()
			config := sapSUPAExecuteTestsOptions{
				ScriptType:       "all",
				TestScriptFolder: "testScriptFolder",
			}
			// defines a test case type all
			gitParams, err := getGitParams(&config, []string{}, TYPE_ALL)
			assert.NoError(t, err)
			assert.Contains(t, gitParams, "REPO_BASE_PATH=testScriptFolder")
		})

		t.Run("test error type all", func(t *testing.T) {
			t.Parallel()
			// init
			config := sapSUPAExecuteTestsOptions{}
			config.ScriptType = "all"
			utils := newSapSUPAExecuteTestsTestsUtils()
			// test
			err := runAllPerformanceSingleUserTest(&config, nil, utils)
			// assert
			assert.NoError(t, err)
		})

		t.Run("test error type Selenium", func(t *testing.T) {
			t.Parallel()
			// init
			config := sapSUPAExecuteTestsOptions{}
			config.ScriptType = "selenium"
			utils := newSapSUPAExecuteTestsTestsUtils()
			// test
			err := runSeleniumPerformanceSingleUserTest(&config, nil, utils)
			// assert
			assert.NoError(t, err)
		})

		t.Run("test error type Krypton", func(t *testing.T) {
			t.Parallel()
			// init
			config := sapSUPAExecuteTestsOptions{}
			config.ScriptType = "krypton"
			utils := newSapSUPAExecuteTestsTestsUtils()
			// test
			err := runKryptonPerformanceSingleUserTest(&config, nil, utils)
			// assert
			assert.NoError(t, err)
		})

		t.Run("test error type JMeter", func(t *testing.T) {
			t.Parallel()
			// init
			config := sapSUPAExecuteTestsOptions{}
			config.ScriptType = "jmeter"
			utils := newSapSUPAExecuteTestsTestsUtils()
			// test
			err := runJMeterPerformanceSingleUserTest(&config, nil, utils)
			// assert
			assert.NoError(t, err)
		})

		t.Run("test error UiVeri5 type", func(t *testing.T) {
			t.Parallel()
			// init
			config := sapSUPAExecuteTestsOptions{}
			config.ScriptType = "uiveri5"
			utils := newSapSUPAExecuteTestsTestsUtils()
			// test
			err := runUiveri5PerformanceSingleUserTest(&config, nil, utils)
			// assert
			assert.NoError(t, err)
		})

		t.Run("test error Wdio type", func(t *testing.T) {
			t.Parallel()
			// init
			config := sapSUPAExecuteTestsOptions{}
			config.ScriptType = "wdio"
			utils := newSapSUPAExecuteTestsTestsUtils()
			// test
			err := runWdioPerformanceSingleUserTest(&config, nil, utils)
			// assert
			assert.NoError(t, err)
		})

		t.Run("test error Qmate type", func(t *testing.T) {
			t.Parallel()
			// init
			config := sapSUPAExecuteTestsOptions{}
			config.ScriptType = "qmate"
			utils := newSapSUPAExecuteTestsTestsUtils()
			// test
			err := runQmatePerformanceSingleUserTest(&config, nil, utils)
			// assert
			assert.NoError(t, err)
		})

		t.Run("test error NPM type", func(t *testing.T) {
			t.Parallel()
			// init
			config := sapSUPAExecuteTestsOptions{}
			config.ScriptType = "npm"
			utils := newSapSUPAExecuteTestsTestsUtils()
			// test
			err := runNpmPerformanceSingleUserTest(&config, nil, utils)
			// assert
			assert.NoError(t, err)
		})

		t.Run("test error Any type", func(t *testing.T) {
			t.Parallel()
			// init
			config := sapSUPAExecuteTestsOptions{}
			config.ScriptType = "any"
			utils := newSapSUPAExecuteTestsTestsUtils()
			// test
			err := runAnyTestInContainer(&config, nil, utils)
			// assert
			assert.EqualError(t, err, "valid Any testcase not defined: missing CMD parameters for any test case")
		})

		t.Run("test error archiveTestResults", func(t *testing.T) {
			t.Skip("failing due to varying os error message")
			t.Parallel()
			// init
			supa = supaSettings{
				workingDir: "/tools/SUPA",
				resultDir:  "/home/ubuntu/supaData/results",
				logDir:     "/home/ubuntu/supaData/logs",
			}

			config := sapSUPAExecuteTestsOptions{}
			config.ArchiveSUPAResult = true
			// test
			err := archiveTestResults(&config, supa.resultDir)
			// assert
			assert.EqualError(t, err, "Error archiveTestResults - not exist resultDir: stat /home/ubuntu/supaData/results: no such file or directory")
		})

		t.Run("test error extract data", func(t *testing.T) {
			t.Skip("failing due to varying os error message")
			t.Parallel()
			// init
			supa = supaSettings{
				workingDir: "/tools/SUPA",
				resultDir:  "/home/ubuntu/supaData/results",
				logDir:     "/home/ubuntu/supaData/logs",
			}

			config := sapSUPAExecuteTestsOptions{}
			config.ScriptType = "any"
			// test
			err := extractReportDataFile(supa.resultDir)
			// assert
			assert.EqualError(t, err, "failed to extractReportDataFile, read dir: /home/ubuntu/supaData/results: open /home/ubuntu/supaData/results: no such file or directory")
		})

		t.Run("test error for extract file", func(t *testing.T) {
			t.Parallel()
			// init
			supa = supaSettings{
				workingDir: "/tools/SUPA",
				resultDir:  ".",
				logDir:     "/home/ubuntu/supaData/logs",
			}

			config := sapSUPAExecuteTestsOptions{}
			config.ScriptType = "any"
			// test
			err := extractFile(supa.resultDir, "test", supa.logDir)
			// assert
			assert.EqualError(t, err, "extractFile - failed to open zip file: test.zip")
		})

	})
}

func TestRunSapSUPAExecuteTests(t *testing.T) {
	tests := []struct {
		name        string
		scriptType  string
		setup       func(*sapSUPAExecuteTestsOptions)
		expectedErr string
	}{
		{
			name:       "All type executes successfully",
			scriptType: "all",
		},
		{
			name:       "Selenium type executes successfully",
			scriptType: "selenium",
		},
		{
			name:       "Krypton type executes successfully",
			scriptType: "krypton",
		},
		{
			name:       "JMeter type executes successfully",
			scriptType: "jmeter",
		},
		{
			name:       "UiVeri5 type executes successfully",
			scriptType: "uiveri5",
		},
		{
			name:       "Wdio type executes successfully",
			scriptType: "wdio",
		},
		{
			name:       "Qmate type executes successfully",
			scriptType: "qmate",
		},
		{
			name:       "NPM type executes successfully",
			scriptType: "npm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			config := sapSUPAExecuteTestsOptions{ScriptType: tt.scriptType}
			if tt.setup != nil {
				tt.setup(&config)
			}
			utils := newSapSUPAExecuteTestsTestsUtils()
			err := runSapSUPAExecuteTests(config, nil, utils, "")
			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetEnvVar(t *testing.T) {
	os.Setenv("MYVAR", "resolvedValue")
	defer os.Unsetenv("MYVAR")
	input := "FOO=bar BAR=$MYVAR"
	expected := "FOO=bar BAR=resolvedValue"
	result := getEnvVar(input)
	assert.Equal(t, expected, result)
}

func TestGetDockerEnvParams(t *testing.T) {
	config := &sapSUPAExecuteTestsOptions{
		SupaKeystoreKey: "key123",
		EnvVars:         []string{"FOO=bar"},
	}
	envs := getDockerEnvParams(config, []string{"INIT"})
	assert.Contains(t, envs, "SUPA_JCEKS_KEY=key123")
	assert.Contains(t, envs, "AUTOMATE_ENV=Docker_Piper")
}

func TestGetGitParams(t *testing.T) {
	config := &sapSUPAExecuteTestsOptions{
		TestRepository: "repo",
		GithubToken:    "",
	}
	envs, err := getGitParams(config, []string{}, TYPE_DEF)
	assert.Error(t, err)
	assert.Contains(t, envs, "REPOSITORY=repo")
}

func TestRunWdioPerformanceSingleUserTest(t *testing.T) {
	config := &sapSUPAExecuteTestsOptions{
		TestWdioParams: "param1",
		SupaConfig:     "cfg1",
	}
	utils := newSapSUPAExecuteTestsTestsUtils()
	err := runWdioPerformanceSingleUserTest(config, nil, utils)
	assert.NoError(t, err)
}

func TestRunQmatePerformanceSingleUserTest(t *testing.T) {
	config := &sapSUPAExecuteTestsOptions{
		TestQmateParams: "param2",
		TestQmateCfg:    "cfg2",
		SupaConfig:      "cfg3",
	}
	utils := newSapSUPAExecuteTestsTestsUtils()
	err := runQmatePerformanceSingleUserTest(config, nil, utils)
	assert.NoError(t, err)
}

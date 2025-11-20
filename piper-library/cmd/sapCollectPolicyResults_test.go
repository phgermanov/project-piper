package cmd

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/policy/agent"
)

type sapCollectPolicyResultsMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	agent.Agent
}

func (s sapCollectPolicyResultsMockUtils) WaitPeriod() time.Duration {
	return 1 * time.Millisecond
}

func newSapCollectPolicyResultsTestsUtils() sapCollectPolicyResultsMockUtils {
	filesMock := &mock.FilesMock{}
	mockRunner := &mock.ExecMockRunner{}
	utils := sapCollectPolicyResultsMockUtils{
		ExecMockRunner: mockRunner,
		FilesMock:      filesMock,
		Agent:          agent.NewTestAgent(&mockResolver{filesMock}, mockRunner, filesMock),
	}
	return utils
}

func TestRunSapCollectPolicyResults(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := sapCollectPolicyResultsOptions{
			PipelineRuns:    []string{"1;1.0", "2;1.2", "3;3.3"},
			JSONKeyFilePath: "test.json",
		}

		utils := newSapCollectPolicyResultsTestsUtils()

		// test
		err := runSapCollectPolicyResults(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("happy path - retries", func(t *testing.T) {
		t.Parallel()
		// init
		config := sapCollectPolicyResultsOptions{
			PipelineRuns:       []string{"1;1.0", "2;1.2", "3;3.3"},
			JSONKeyFilePath:    "test.json",
			ValidateCompliance: true,
			MaxWait:            2,
		}

		utils := newSapCollectPolicyResultsTestsUtils()

		// test
		err := runSapCollectPolicyResults(&config, nil, utils)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, 3, len(utils.ExecMockRunner.Calls))
		assert.NotContains(t, utils.ExecMockRunner.Calls[1].Params, "--validate-compliance")
		assert.Contains(t, utils.ExecMockRunner.Calls[1].Params, "--validate-missing")
		assert.Contains(t, utils.ExecMockRunner.Calls[2].Params, "--validate-compliance")
		assert.Contains(t, utils.ExecMockRunner.Calls[2].Params, "--validate-missing")
	})

	t.Run("failure - retries", func(t *testing.T) {
		t.Parallel()
		// init
		config := sapCollectPolicyResultsOptions{
			PipelineRuns:       []string{"1;1.0", "2;1.2", "3;3.3"},
			JSONKeyFilePath:    "test.json",
			ValidateCompliance: true,
			MaxWait:            2,
		}

		utils := newSapCollectPolicyResultsTestsUtils()
		utils.ExecMockRunner.ShouldFailOnCommand = map[string]error{"./cumulus-policy-agent- collectPolicyResults": errors.New("execution error")}

		// test
		err := runSapCollectPolicyResults(&config, nil, utils)

		// assert
		assert.ErrorContains(t, err, "execution error")
		assert.Equal(t, 9, len(utils.ExecMockRunner.Calls))
	})

	t.Run("map configuration", func(t *testing.T) {
		tests := []struct {
			config   *sapCollectPolicyResultsOptions
			expected []string
		}{
			{
				config: &sapCollectPolicyResultsOptions{
					JSONKeyFilePath: "credentials.json",
					PipelineRuns:    []string{"1,1.0", "2,1.2", "3,3.3"},
				},
				expected: []string{"collectPolicyResults", "1,1.0", "2,1.2", "3,3.3", "--credentials-file=credentials.json"},
			},
			{
				config: &sapCollectPolicyResultsOptions{
					JSONKeyFilePath:    "credentials.json",
					PipelineRuns:       []string{"1,1.0", "2,1.2", "3,3.3"},
					ValidateCompliance: false,
				},
				expected: []string{"collectPolicyResults", "1,1.0", "2,1.2", "3,3.3", "--credentials-file=credentials.json"},
			},
			{
				config: &sapCollectPolicyResultsOptions{
					JSONKeyFilePath:    "credentials.json",
					PipelineRuns:       []string{"1,1.0", "2,1.2", "3,3.3"},
					ValidateCompliance: true,
				},
				// expected: []string{"collectPolicyResults", "1,1.0", "2,1.2", "3,3.3", "--credentials-file=credentials.json", "--validate-compliance", "--validate-missing"},
				expected: []string{"collectPolicyResults", "1,1.0", "2,1.2", "3,3.3", "--credentials-file=credentials.json"},
			},
			{
				config: &sapCollectPolicyResultsOptions{
					JSONKeyFilePath:           "credentials.json",
					PipelineRuns:              []string{"1,1.0", "2,1.2", "3,3.3"},
					ValidateCompliance:        true,
					CentralPolicyKeys:         []string{"HELLO-1", "HELLO-2"},
					CustomPolicyKeys:          []string{"WORLD-1", "WORLD-2"},
					ResultFile:                "result.json",
					CumulusPolicyAgentVersion: "1.2.3",
				},
				// expected: []string{"collectPolicyResults", "1,1.0", "2,1.2", "3,3.3", "--credentials-file=credentials.json", "--central-policy-keys=HELLO-1,HELLO-2", "--custom-policy-keys=WORLD-1,WORLD-2", "--validate-compliance", "--validate-missing", "--out=result.json"},
				expected: []string{"collectPolicyResults", "1,1.0", "2,1.2", "3,3.3", "--credentials-file=credentials.json", "--central-policy-keys=HELLO-1,HELLO-2", "--custom-policy-keys=WORLD-1,WORLD-2", "--out=result.json"},
			},
		}

		for _, test := range tests {
			t.Run("map configuration", func(t *testing.T) {
				// t.Parallel()
				// test
				params, err := mapCollectPolicyResultsParams(test.config)

				// assert
				assert.NoError(t, err)
				assert.Equal(t, test.expected, params)
			})
		}
	})

	t.Run("configuration errors", func(t *testing.T) {
		tests := []struct {
			config   *sapCollectPolicyResultsOptions
			expected error
		}{
			{
				config:   &sapCollectPolicyResultsOptions{},
				expected: fmt.Errorf("configuration parameter PipelineRuns is missing"),
			},
			{
				config: &sapCollectPolicyResultsOptions{
					PipelineRuns: []string{"1,1.0", "2,1.2", "3,3.3"},
				},
				expected: fmt.Errorf("configuration parameter JSONKeyFilePath and Token is missing"),
			},
		}

		for _, test := range tests {
			t.Run("map configuration errors", func(t *testing.T) {
				// t.Parallel()
				// test
				_, err := mapCollectPolicyResultsParams(test.config)

				// assert
				assert.Error(t, err)
				assert.Equal(t, test.expected, err)
			})
		}
	})

	t.Run("can create newSapCollectPolicyResultsUtils", func(t *testing.T) {
		t.Parallel()
		// init

		params := sapCollectPolicyResultsOptions{
			GithubToken: "token",
		}

		utils, err := newSapCollectPolicyResultsUtils(params)

		// assert
		assert.NoError(t, err)
		assert.NotNil(t, utils)
	})
}

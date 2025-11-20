package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/policy"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/policy/agent"
)

type sapExecuteCentralPolicyMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	agent.Agent
}

func newSapExecuteCentralPolicyTestsUtils() sapExecuteCentralPolicyMockUtils {
	filesMock := &mock.FilesMock{}
	utils := sapExecuteCentralPolicyMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
		Agent:          agent.NewTestAgent(&mockResolver{filesMock}, &mockExecutor{}, filesMock),
	}
	return utils
}

type failingMockExecutor struct {
	executionCounter    int
	failAfterExecutions int
}

func (executor *failingMockExecutor) GetExitCode() int {
	if executor.executionCounter > executor.failAfterExecutions {
		return 1
	} else {
		return 0
	}
}

func (executor *failingMockExecutor) RunExecutable(executable string, params ...string) error {
	executor.executionCounter++
	if executor.executionCounter > executor.failAfterExecutions {
		return fmt.Errorf("failed to execute")
	} else {
		return nil
	}
}

func (executor *failingMockExecutor) SetEnv(env []string) {}

func newSapExecuteCentralPolicyTestsUtilsFailing() sapExecuteCentralPolicyMockUtils {
	filesMock := &mock.FilesMock{}
	utils := sapExecuteCentralPolicyMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
		Agent:          agent.NewTestAgent(&mockResolver{filesMock}, &failingMockExecutor{executionCounter: 0, failAfterExecutions: 2}, filesMock),
	}
	return utils
}

func TestRunSapExecuteCentralPolicy(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := sapExecuteCentralPolicyOptions{
			PolicyKey:       "POLICY-1",
			EvidenceFile:    "evidence.json",
			JSONKeyFilePath: "credentials/auth.json",
		}

		utils := newSapExecuteCentralPolicyTestsUtils()

		// test
		err := runSapExecuteCentralPolicy(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("happy path with failing execution", func(t *testing.T) {
		t.Parallel()
		// init
		config := sapExecuteCentralPolicyOptions{
			PolicyKey:       "POLICY-1",
			EvidenceFile:    "evidence.json",
			JSONKeyFilePath: "credentials/auth.json",
		}

		utils := newSapExecuteCentralPolicyTestsUtilsFailing()

		// test
		err := runSapExecuteCentralPolicy(&config, nil, utils)

		// assert
		assert.Error(t, err)
	})

	t.Run("map configuration", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			config   *sapExecuteCentralPolicyOptions
			expected []string
		}{
			{
				config: &sapExecuteCentralPolicyOptions{
					PolicyKey:             "POLICY-1",
					EvidenceFile:          "evidence.json",
					FailOnPolicyViolation: true,
				},
				expected: []string{"execute", "--type=central", "POLICY-1", "evidence.json", "--out=policy-result/POLICY-1/result.json"},
			},
			{
				config: &sapExecuteCentralPolicyOptions{
					PolicyKey:             "POLICY-2",
					EvidenceFile:          "evidence.json",
					FailOnPolicyViolation: false,
				},
				expected: []string{"execute", "--type=central", "POLICY-2", "evidence.json", "--ignore-compliance=true", "--out=policy-result/POLICY-2/result.json"},
			},
			{
				config: &sapExecuteCentralPolicyOptions{
					PolicyKey:             "POLICY-3",
					EvidenceFile:          "evidence.json",
					FailOnPolicyViolation: true,
					ResultFile:            "policy-result/<policyKey>/foo/bar.json",
				},
				expected: []string{"execute", "--type=central", "POLICY-3", "evidence.json", "--out=policy-result/POLICY-3/foo/bar.json"},
			},
			{
				config: &sapExecuteCentralPolicyOptions{
					PolicyKey:             "POLICY-3",
					EvidenceFile:          "evidence.json",
					FailOnPolicyViolation: true,
					ResultFile:            "custom-path/foo/bar.json",
				},
				expected: []string{"execute", "--type=central", "POLICY-3", "evidence.json", "--out=custom-path/foo/bar.json"},
			},
			{
				config: &sapExecuteCentralPolicyOptions{
					PolicyKey:             "POLICY-3",
					EvidenceFile:          "evidence.json",
					FailOnPolicyViolation: true,
					ResultFile:            "custom-path/<policyKey>/foo/bar.json",
				},
				expected: []string{"execute", "--type=central", "POLICY-3", "evidence.json", "--out=custom-path/POLICY-3/foo/bar.json"},
			},
			{
				config: &sapExecuteCentralPolicyOptions{
					PolicyKey:             "POLICY-4",
					EvidenceFile:          "evidence.json",
					FailOnPolicyViolation: true,
					CentralPolicyPath:     "central-policy-path",
					GenerateJunitReport:   true,
				},
				expected: []string{"execute", "--type=central", "POLICY-4", "evidence.json", "--directory=central-policy-path", "--out=policy-result/POLICY-4/result.json"},
			},
		}

		for _, test := range tests {
			t.Run("map configuration", func(t *testing.T) {
				// capture loop variable before switching to parallel execution
				test := test
				t.Parallel()
				// test
				params, err := mapExecuteCentralPolicyParams(test.config)

				// assert
				assert.NoError(t, err)
				assert.Equal(t, test.expected, params)
			})
		}
	})

	t.Run("map synchronizeCentralPolicyBundles configuration", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			config   *sapExecuteCentralPolicyOptions
			expected []string
		}{
			{
				config: &sapExecuteCentralPolicyOptions{
					JSONKeyFilePath:     "credentials/auth.json",
					CentralPolicyPath:   "central-policy-path",
					CumulusPolicyBucket: "some-other-test-bundles",
				},
				expected: []string{"synchronizePolicyBundles", "--directory=central-policy-path", "--policy-bucket=some-other-test-bundles", "--credentials-file=credentials/auth.json"},
			},
			{
				config: &sapExecuteCentralPolicyOptions{
					JSONKeyFilePath: "credentials/auth.json",
				},
				expected: []string{"synchronizePolicyBundles", "--credentials-file=credentials/auth.json"},
			},
		}

		for _, test := range tests {
			t.Run("map synchronizeCentralPolicyBundles configuration", func(t *testing.T) {
				// capture loop variable before switching to parallel execution
				test := test
				t.Parallel()
				// test
				params, err := mapSynchronizeCentralPolicyBundlesParams(test.config)

				// assert
				assert.NoError(t, err)
				assert.Equal(t, test.expected, params)
			})
		}
	})

	t.Run("configuration errors", func(t *testing.T) {
		tests := []struct {
			config   *sapExecuteCentralPolicyOptions
			expected error
		}{
			{
				config:   &sapExecuteCentralPolicyOptions{},
				expected: fmt.Errorf("configuration parameter policyKey is missing"),
			},
			{
				config: &sapExecuteCentralPolicyOptions{
					PolicyKey: "POLICY-1",
				},
				expected: fmt.Errorf("configuration parameter evidenceFile is missing"),
			},
		}

		for _, test := range tests {
			t.Run("map configuration errors", func(t *testing.T) {
				// t.Parallel()
				// test
				_, err := mapExecuteCentralPolicyParams(test.config)

				// assert
				assert.Error(t, err)
				assert.Equal(t, test.expected, err)
			})
		}
	})

	t.Run("map synchronizeCentralPolicyBundles configuration error", func(t *testing.T) {
		t.Parallel()
		test := struct {
			config   *sapExecuteCentralPolicyOptions
			expected error
		}{
			config: &sapExecuteCentralPolicyOptions{
				CentralPolicyPath:   "central-policy-path",
				CumulusPolicyBucket: "some-other-test-bundles",
			},
			expected: fmt.Errorf("configuration parameter JSONKeyFilePath and Token is missing"),
		}

		// test
		_, err := mapSynchronizeCentralPolicyBundlesParams(test.config)

		// assert
		assert.Error(t, err)
		assert.Equal(t, test.expected, err)

	})

	t.Run("can create newSapExecuteCentralPolicyUtils", func(t *testing.T) {
		t.Parallel()
		// init

		params := sapExecuteCentralPolicyOptions{
			GithubToken: "token",
		}

		utils, err := newSapExecuteCentralPolicyUtils(params)

		// assert
		assert.NoError(t, err)
		assert.NotNil(t, utils)
	})

	t.Run("can not generate junit report of central policy if result is not available", func(t *testing.T) {
		t.Parallel()
		// init

		options := &sapExecuteCentralPolicyOptions{
			PolicyKey:             "POLICY-4",
			EvidenceFile:          "evidence.json",
			FailOnPolicyViolation: true,
			CentralPolicyPath:     "central-policy-path",
			GenerateJunitReport:   true,
			ResultFile:            "does-not-exist.json",
		}

		params, err := mapExecuteCentralPolicyParams(options)

		assert.NoError(t, err)

		utils := newSapExecuteCentralPolicyTestsUtils()

		err = generateJunitReportOfCentralPolicy(params, utils)

		// assert
		assert.NotNil(t, err)
	})

	t.Run("can generate junit report of central policy if result is available", func(t *testing.T) {
		t.Parallel()
		// init

		options := &sapExecuteCentralPolicyOptions{
			PolicyKey:             "POLICY-4",
			EvidenceFile:          "evidence-exists.json",
			FailOnPolicyViolation: true,
			CentralPolicyPath:     "central-policy-path",
			GenerateJunitReport:   true,
			ResultFile:            "exists.json",
		}

		params, err := mapExecuteCentralPolicyParams(options)

		assert.NoError(t, err)

		utils := newSapExecuteCentralPolicyTestsUtils()

		utils.FilesMock.AddFile("exists.json", []byte(`{"policy": {"key": "MY-POLICY-1", "label": "This is my central policy"}, "complianceStatus": "COMPLIANT", "validationErrorMessages": []}`))

		err = generateJunitReportOfCentralPolicy(params, utils)

		// assert
		assert.Nil(t, err)

		assert.True(t, utils.FilesMock.HasWrittenFile(fmt.Sprintf("TEST-%s-policy-%s.xml", policy.Central, "MY-POLICY-1")))
	})
}

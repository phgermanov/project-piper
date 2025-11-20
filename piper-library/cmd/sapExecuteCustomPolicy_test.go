package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/policy"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/policy/agent"
)

type sapExecuteCustomPolicyMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	agent.Agent
}

type mockResolver struct {
	*mock.FilesMock
}

func (resolver *mockResolver) ResolveUrl(version string) (string, error) {
	return "", nil
}

func (resolver *mockResolver) Download(version, targetFile string) error {
	resolver.FilesMock.AddFile(targetFile, []byte("test"))
	return nil
}

type mockExecutor struct{}

func (executor *mockExecutor) GetExitCode() int {
	return 0
}

func (executor *mockExecutor) RunExecutable(executable string, params ...string) error {
	return nil
}

func (executor *mockExecutor) SetEnv(env []string) {}

func newSapExecuteCustomPolicyTestsUtils() sapExecuteCustomPolicyMockUtils {
	filesMock := &mock.FilesMock{}
	utils := sapExecuteCustomPolicyMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      filesMock,
		Agent:          agent.NewTestAgent(&mockResolver{filesMock}, &mockExecutor{}, filesMock),
	}
	return utils
}

func TestRunSapExecuteCustomPolicy(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := sapExecuteCustomPolicyOptions{
			PolicyKey:    "POLICY-1",
			EvidenceFile: "evidence.json",
		}
		utils := newSapExecuteCustomPolicyTestsUtils()

		// test
		err := runSapExecuteCustomPolicy(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("map configuration", func(t *testing.T) {
		tests := []struct {
			config   *sapExecuteCustomPolicyOptions
			expected []string
		}{
			{
				config: &sapExecuteCustomPolicyOptions{
					PolicyKey:             "POLICY-1",
					EvidenceFile:          "evidence.json",
					FailOnPolicyViolation: true,
				},
				expected: []string{"execute", "--type=custom", "POLICY-1", "evidence.json", "--out=custom-policy-result/POLICY-1/result.json"},
			},
			{
				config: &sapExecuteCustomPolicyOptions{
					PolicyKey:             "POLICY-2",
					EvidenceFile:          "evidence.json",
					FailOnPolicyViolation: false,
				},
				expected: []string{"execute", "--type=custom", "POLICY-2", "evidence.json", "--ignore-compliance=true", "--out=custom-policy-result/POLICY-2/result.json"},
			},
			{
				config: &sapExecuteCustomPolicyOptions{
					PolicyKey:             "POLICY-3",
					EvidenceFile:          "evidence.json",
					FailOnPolicyViolation: true,
					ResultFile:            "custom-policy-result/<policyKey>/foo/bar.json",
				},
				expected: []string{"execute", "--type=custom", "POLICY-3", "evidence.json", "--out=custom-policy-result/POLICY-3/foo/bar.json"},
			},
			{
				config: &sapExecuteCustomPolicyOptions{
					PolicyKey:             "POLICY-3",
					EvidenceFile:          "evidence.json",
					FailOnPolicyViolation: true,
					ResultFile:            "custom-path/foo/bar.json",
				},
				expected: []string{"execute", "--type=custom", "POLICY-3", "evidence.json", "--out=custom-path/foo/bar.json"},
			},
			{
				config: &sapExecuteCustomPolicyOptions{
					PolicyKey:             "POLICY-3",
					EvidenceFile:          "evidence.json",
					FailOnPolicyViolation: true,
					ResultFile:            "custom-path/<policyKey>/foo/bar.json",
				},
				expected: []string{"execute", "--type=custom", "POLICY-3", "evidence.json", "--out=custom-path/POLICY-3/foo/bar.json"},
			},
			{
				config: &sapExecuteCustomPolicyOptions{
					PolicyKey:             "POLICY-4",
					EvidenceFile:          "evidence.json",
					FailOnPolicyViolation: true,
					PolicyPath:            "custom-path",
					GenerateJunitReport:   true,
				},
				expected: []string{"execute", "--type=custom", "POLICY-4", "evidence.json", "--directory=custom-path", "--out=custom-policy-result/POLICY-4/result.json"},
			},
		}

		for _, test := range tests {
			t.Run("map configuration", func(t *testing.T) {
				// t.Parallel()
				// test
				params, err := mapExecuteCustomPolicyParams(test.config)

				// assert
				assert.NoError(t, err)
				assert.Equal(t, test.expected, params)
			})
		}
	})

	t.Run("configuration errors", func(t *testing.T) {
		tests := []struct {
			config   *sapExecuteCustomPolicyOptions
			expected error
		}{
			{
				config:   &sapExecuteCustomPolicyOptions{},
				expected: fmt.Errorf("configuration parameter policyKey is missing"),
			},
			{
				config: &sapExecuteCustomPolicyOptions{
					PolicyKey: "POLICY-1",
				},
				expected: fmt.Errorf("configuration parameter evidenceFile is missing"),
			},
		}

		for _, test := range tests {
			t.Run("map configuration errors", func(t *testing.T) {
				// t.Parallel()
				// test
				_, err := mapExecuteCustomPolicyParams(test.config)

				// assert
				assert.Error(t, err)
				assert.Equal(t, test.expected, err)
			})
		}
	})

	t.Run("can create newSapExecuteCustomPolicyUtils", func(t *testing.T) {
		t.Parallel()
		// init

		params := sapExecuteCustomPolicyOptions{
			GithubToken: "token",
		}

		utils, err := newSapExecuteCustomPolicyUtils(params)

		// assert
		assert.NoError(t, err)
		assert.NotNil(t, utils)
	})

	t.Run("can not generate junit report if result is not available", func(t *testing.T) {
		t.Parallel()
		// init

		options := &sapExecuteCustomPolicyOptions{
			PolicyKey:             "POLICY-4",
			EvidenceFile:          "evidence.json",
			FailOnPolicyViolation: true,
			PolicyPath:            "custom-path",
			GenerateJunitReport:   true,
			ResultFile:            "does-not-exist.json",
		}

		params, err := mapExecuteCustomPolicyParams(options)

		assert.NoError(t, err)

		utils := newSapExecuteCustomPolicyTestsUtils()

		err = generateJunitReport(params, utils)

		// assert
		assert.NotNil(t, err)
	})

	t.Run("can generate junit report if result is available", func(t *testing.T) {
		t.Parallel()
		// init

		options := &sapExecuteCustomPolicyOptions{
			PolicyKey:             "POLICY-4",
			EvidenceFile:          "evidence-exists.json",
			FailOnPolicyViolation: true,
			PolicyPath:            "custom-path",
			GenerateJunitReport:   true,
			ResultFile:            "exists.json",
		}

		params, err := mapExecuteCustomPolicyParams(options)

		assert.NoError(t, err)

		utils := newSapExecuteCustomPolicyTestsUtils()

		utils.FilesMock.AddFile("exists.json", []byte(`{"policy": {"key": "MY-POLICY-1", "label": "This is my custom policy"}, "complianceStatus": "COMPLIANT", "validationErrorMessages": []}`))

		err = generateJunitReport(params, utils)

		// assert
		assert.Nil(t, err)

		assert.True(t, utils.FilesMock.HasWrittenFile(fmt.Sprintf("TEST-%s-policy-%s.xml", policy.Custom, "MY-POLICY-1")))
	})
}

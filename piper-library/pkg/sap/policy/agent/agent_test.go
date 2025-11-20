package agent

import (
	"errors"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type mockResolver struct {
	*mock.FilesMock
}

func (resolver *mockResolver) ResolveUrl(version string) (string, error) {
	return version, nil
}

func (resolver *mockResolver) Download(version, targetFile string) error {
	resolver.FilesMock.AddFile(targetFile, []byte(version))
	return nil
}

type errorExecutor struct {
}

func (executor *errorExecutor) GetExitCode() int {
	return -1
}

func (executor *errorExecutor) RunExecutable(executable string, params ...string) error {
	return errors.New("this executer always throws an error")
}

func (executor *errorExecutor) SetEnv(env []string) {}

type dummyExecutor struct {
}

func (executor *dummyExecutor) GetExitCode() int {
	return 0
}

func (executor *dummyExecutor) RunExecutable(executable string, params ...string) error {
	// dont do anything
	return nil
}

func (executor *dummyExecutor) SetEnv(env []string) {}

func TestRunPolicyAgentExecute(t *testing.T) {
	t.Parallel()

	t.Run("Can create agent instance", func(t *testing.T) {
		t.Parallel()
		agent := NewAgent(nil)

		// assert
		assert.NotNil(t, agent)
	})

	t.Run("Execute without install", func(t *testing.T) {
		t.Parallel()
		agent := NewTestAgent(nil, nil, nil)

		// test
		err := agent.Execute([]string{}, nil)

		// assert
		assert.EqualError(t, err, "cumulus policy agent not installed! (run Install() first))")
	})

	t.Run("Execution fails with preinstalled error executor", func(t *testing.T) {
		t.Parallel()
		filesMock := &mock.FilesMock{}
		filesMock.AddFile("cumulus-policy-agent", []byte("latest"))
		agent := NewTestAgent(&mockResolver{filesMock}, &errorExecutor{}, filesMock)

		// test
		agent.Install("latest")
		err := agent.Execute([]string{}, nil)

		// assert
		assert.EqualError(t, err, "execution failed with exit code -1. inner execution error: this executer always throws an error")
	})

	t.Run("Execution fails with new installed error executor", func(t *testing.T) {
		t.Parallel()
		filesMock := &mock.FilesMock{}
		agent := NewTestAgent(&mockResolver{filesMock}, &errorExecutor{}, filesMock)

		// test
		agent.Install("not-yet-downloaded")
		err := agent.Execute([]string{}, nil)

		// assert
		assert.EqualError(t, err, "execution failed with exit code -1. inner execution error: this executer always throws an error")
	})

	t.Run("Execution with dummy executor", func(t *testing.T) {
		t.Parallel()
		filesMock := &mock.FilesMock{}
		filesMock.AddFile("cumulus-policy-agent", []byte("latest"))
		agent := NewTestAgent(&mockResolver{filesMock}, &dummyExecutor{}, filesMock)

		// test
		agent.Install("latest")
		err := agent.Execute([]string{}, nil)

		// assert
		assert.NoError(t, err)
	})

}

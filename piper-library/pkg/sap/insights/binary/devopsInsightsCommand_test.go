package binary

import (
	"fmt"
	"io"
	"strings"
	"testing"

	piperMocks "github.com/SAP/jenkins-library/pkg/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDefaultCLICommandExecutor(t *testing.T) {
	t.Parallel()
	testStr := "Write me to the buffer"
	targetBinary := "/tmp/devops-insights"
	tests := []struct {
		name              string
		generateFactory   func(t *testing.T, execMock *piperMocks.ExecMockRunner) BlockingExecutorFactory
		validateReference func(t *testing.T, execMock *piperMocks.ExecMockRunner)
		params            []string
		wantErr           bool
	}{
		{
			name: "canonical command execution",
			generateFactory: func(t *testing.T, execMock *piperMocks.ExecMockRunner) BlockingExecutorFactory {
				t.Helper()
				execMock.ExitCode = 0
				execMock.StdoutReturn = map[string]string{targetBinary: string(testStr)}
				fac := NewMockBlockingExecutorFactory(t)
				fac.On("CreateExecutor", mock.Anything, mock.Anything).Return(
					func(stdout io.Writer, stderr io.Writer) BlockingExecutor {
						execMock.Stdout(stdout)
						execMock.Stderr(stderr)
						return execMock
					})
				return fac
			},
			validateReference: func(t *testing.T, execMock *piperMocks.ExecMockRunner) {
				t.Helper()
				actualCall := append([]string{execMock.Calls[0].Exec}, execMock.Calls[0].Params...)
				assert.Equal(t, "/tmp/devops-insights", strings.Join(actualCall, " "))
			},
			wantErr: false,
		},
		{
			name: "canonical command execution with params",
			generateFactory: func(t *testing.T, execMock *piperMocks.ExecMockRunner) BlockingExecutorFactory {
				t.Helper()
				execMock.ExitCode = 0
				execMock.StdoutReturn = map[string]string{targetBinary: string(testStr)}
				fac := NewMockBlockingExecutorFactory(t)
				fac.On("CreateExecutor", mock.Anything, mock.Anything).Return(
					func(stdout io.Writer, stderr io.Writer) BlockingExecutor {
						execMock.Stdout(stdout)
						execMock.Stderr(stderr)
						return execMock
					})
				return fac
			},
			params: []string{"--config", "cfg.yaml"},
			validateReference: func(t *testing.T, execMock *piperMocks.ExecMockRunner) {
				t.Helper()
				actualCall := append([]string{execMock.Calls[0].Exec}, execMock.Calls[0].Params...)
				assert.Equal(t, "/tmp/devops-insights --config cfg.yaml", strings.Join(actualCall, " "))
			},
			wantErr: false,
		},
		{
			name: "running the executable with a non zero exit code propagates the error",
			generateFactory: func(t *testing.T, execMock *piperMocks.ExecMockRunner) BlockingExecutorFactory {
				t.Helper()
				execMock.ExitCode = 1
				execMock.ShouldFailOnCommand = map[string]error{targetBinary: errors.New("some strange error occurred")}
				execMock.StdoutReturn = map[string]string{targetBinary: string(testStr)}
				fac := NewMockBlockingExecutorFactory(t)
				fac.On("CreateExecutor", mock.Anything, mock.Anything).Return(
					func(stdout io.Writer, stderr io.Writer) BlockingExecutor {
						execMock.Stdout(stdout)
						execMock.Stderr(stderr)
						return execMock
					})
				return fac
			},
			params: []string{"--config", "err_cfg.yaml"},
			validateReference: func(t *testing.T, execMock *piperMocks.ExecMockRunner) {
				t.Helper()
				actualCall := append([]string{execMock.Calls[0].Exec}, execMock.Calls[0].Params...)
				assert.Equal(t, "/tmp/devops-insights --config err_cfg.yaml", strings.Join(actualCall, " "))
			},
			wantErr: true,
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			execMock := &piperMocks.ExecMockRunner{}
			execFactory := testCase.generateFactory(t, execMock)
			err := DefaultCLICommandExecutor(targetBinary, execFactory, testCase.params)
			fmt.Println(execMock.Calls)
			if (err != nil) != testCase.wantErr {
				t.Fatalf("DefaultCLICommandExecutor() error = %v, wantErr %v", err, testCase.wantErr)
			}
			testCase.validateReference(t, execMock)
		})
	}
}

//go:build unit
// +build unit

package dwc

import (
	"encoding/json"
	piperMocks "github.com/SAP/jenkins-library/pkg/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"testing"
)

func TestDefaultCLICommandExecutor(t *testing.T) {
	t.Parallel()
	type testResponseObject struct {
		P1 string
		P2 string
	}
	testObj := &testResponseObject{
		P1: "p1Data",
		P2: "p2Data",
	}
	testStr := "Write me to the buffer"
	tests := []struct {
		name                       string
		generateFactory            func(t *testing.T) BlockingExecutorFactory
		validateReference          func(t *testing.T, cliResponseTargetReference any)
		cliResponseTargetReference any
		wantErr                    bool
	}{
		{
			name: "canonical command execution",
			generateFactory: func(t *testing.T) BlockingExecutorFactory {
				t.Helper()
				execMock := &piperMocks.ExecMockRunner{}
				execMock.ExitCode = 0
				bytes, err := json.Marshal(testObj)
				if err != nil {
					t.Fatal(err)
				}
				execMock.StdoutReturn = map[string]string{targetBinary: string(bytes)}
				fac := newBlockingExecutorFactoryMock(t)
				fac.On("CreateExecutor", mock.Anything, mock.Anything).Return(
					func(stdout io.Writer, stderr io.Writer) BlockingExecutor {
						execMock.Stdout(stdout)
						execMock.Stderr(stderr)
						return execMock
					})
				return fac
			},
			validateReference: func(t *testing.T, cliResponseTargetReference any) {
				t.Helper()
				assert.Equal(t, testObj, cliResponseTargetReference)
			},
			cliResponseTargetReference: &testResponseObject{},
			wantErr:                    false,
		},
		{
			name: "command execution with cliResponseTargetReference as string",
			generateFactory: func(t *testing.T) BlockingExecutorFactory {
				t.Helper()
				execMock := &piperMocks.ExecMockRunner{}
				execMock.ExitCode = 0
				execMock.StdoutReturn = map[string]string{targetBinary: testStr}
				fac := newBlockingExecutorFactoryMock(t)
				fac.On("CreateExecutor", mock.Anything, mock.Anything).Return(
					func(stdout io.Writer, stderr io.Writer) BlockingExecutor {
						execMock.Stdout(stdout)
						execMock.Stderr(stderr)
						return execMock
					})
				return fac
			},
			validateReference: func(t *testing.T, cliResponseTargetReference any) {
				t.Helper()
				assert.Equal(t, &testStr, cliResponseTargetReference)
			},
			cliResponseTargetReference: new(string),
			wantErr:                    false,
		},
		{
			name: "running the executable with a non zero exit code propagates the error",
			generateFactory: func(t *testing.T) BlockingExecutorFactory {
				t.Helper()
				execMock := &piperMocks.ExecMockRunner{}
				execMock.ExitCode = 1
				execMock.ShouldFailOnCommand = map[string]error{targetBinary: errors.New("some strange error occurred")}
				bytes, err := json.Marshal(testObj)
				if err != nil {
					t.Fatal(err)
				}
				execMock.StdoutReturn = map[string]string{targetBinary: string(bytes)}
				fac := newBlockingExecutorFactoryMock(t)
				fac.On("CreateExecutor", mock.Anything, mock.Anything).Return(
					func(stdout io.Writer, stderr io.Writer) BlockingExecutor {
						execMock.Stdout(stdout)
						execMock.Stderr(stderr)
						return execMock
					})
				return fac
			},
			validateReference: func(t *testing.T, cliResponseTargetReference any) {
				t.Helper()
				assert.Equal(t, testObj, cliResponseTargetReference)
			},
			cliResponseTargetReference: &testResponseObject{},
			wantErr:                    true,
		},
		{
			name: "call with cliResponseTargetReference = nil",
			generateFactory: func(t *testing.T) BlockingExecutorFactory {
				t.Helper()
				execMock := &piperMocks.ExecMockRunner{}
				execMock.ExitCode = 0
				bytes, err := json.Marshal(testObj)
				if err != nil {
					t.Fatal(err)
				}
				execMock.StdoutReturn = map[string]string{targetBinary: string(bytes)}
				fac := newBlockingExecutorFactoryMock(t)
				fac.On("CreateExecutor", mock.Anything, mock.Anything).Return(
					func(stdout io.Writer, stderr io.Writer) BlockingExecutor {
						execMock.Stdout(stdout)
						execMock.Stderr(stderr)
						return execMock
					})
				return fac
			},
			validateReference: func(t *testing.T, cliResponseTargetReference any) {
				t.Helper()
				assert.Equal(t, nil, cliResponseTargetReference)
			},
			cliResponseTargetReference: nil,
			wantErr:                    false,
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			execFactory := testCase.generateFactory(t)
			err := DefaultCLICommandExecutor(execFactory, []string{}, testCase.cliResponseTargetReference)
			if (err != nil) != testCase.wantErr {
				t.Fatalf("DefaultCLICommandExecutor() error = %v, wantErr %v", err, testCase.wantErr)
			}
			testCase.validateReference(t, testCase.cliResponseTargetReference)
		})
	}
}

func TestSetTargetBinary(t *testing.T) {
	type test struct {
		name     string
		path     string
		expected string
	}
	tests := []test{
		{
			name:     "set target binary",
			path:     "myPath",
			expected: "myPath",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			SetTargetBinary(testCase.path)
			if targetBinary != testCase.expected {
				t.Errorf("expected %s to be set as binary path but it was set on %s", testCase.expected, targetBinary)
			}
		})
	}
}

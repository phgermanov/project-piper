//go:build unit
// +build unit

package cumulus

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
)

type gitRepositoryMock struct {
	revisionHash  *plumbing.Hash
	revisionError error
}

func (r *gitRepositoryMock) ResolveRevision(rev plumbing.Revision) (*plumbing.Hash, error) {
	return r.revisionHash, r.revisionError
}

func TestTargetPath(t *testing.T) {

	today := time.Now().Format("20060102")

	expectedOpenGitError := fmt.Errorf("expected open git error")
	expectedRevisionError := fmt.Errorf("expected revision error")

	var objectPathTests = []struct {
		c             Cumulus
		expected      string
		expectedError error
	}{
		{c: Cumulus{Version: "1.0.0", StepResultType: "junit", SubFolderPath: ""}, expected: "1.0.0/junit"},
		{c: Cumulus{Version: "1.0.0", StepResultType: "general", SubFolderPath: "backend"}, expected: "1.0.0/general/backend"},
		{c: Cumulus{Version: "1.0.0", StepResultType: "general", SubFolderPath: "backend/"}, expected: "1.0.0/general/backend"},
		{c: Cumulus{Version: "1.0.1", StepResultType: "general", SubFolderPath: "backend/backend1"}, expected: "1.0.1/general/backend/backend1"},
		{c: Cumulus{Version: "1.0.1", StepResultType: "general", SubFolderPath: "backend/backend1/"}, expected: "1.0.1/general/backend/backend1"},
		{c: Cumulus{Scheduled: true, Version: "1.0.0", StepResultType: "general", SubFolderPath: ""}, expected: fmt.Sprintf("1.0.0-%v/general", today)},
		{c: Cumulus{Scheduled: true, Version: "1.0.1-20200101_sha1", StepResultType: "general", SubFolderPath: ""}, expected: fmt.Sprintf("1.0.1-%v/general", today)},
		{c: Cumulus{Scheduled: true, Version: "1.0.2.20200101_sha1", StepResultType: "general", SubFolderPath: ""}, expected: fmt.Sprintf("1.0.2.%v/general", today)},
		{c: Cumulus{Scheduled: true, Version: "1.0.3", Revision: "sha1revision", StepResultType: "general", SubFolderPath: ""}, expected: fmt.Sprintf("1.0.3-%v-sha1revision/general", today)},
		{c: Cumulus{Scheduled: true, Version: "1.0.4-20200101", Revision: "sha1revision", StepResultType: "general", SubFolderPath: ""}, expected: fmt.Sprintf("1.0.4-%v-sha1revision/general", today)},
		{c: Cumulus{Scheduled: true, Version: "1.0", StepResultType: "general", SubFolderPath: ""}, expected: fmt.Sprintf("1.0-%v/general", today)},
		{c: Cumulus{Scheduled: true, Version: "1.0", StepResultType: "general", SubFolderPath: "", SchedulingTimestamp: "20220202"}, expected: "1.0-20220202/general"},
		{c: Cumulus{Scheduled: true, Version: "main", StepResultType: "general", SubFolderPath: ""}, expected: fmt.Sprintf("main-%v/general", today)},
		{c: Cumulus{UseCommitIDForCumulus: true, HeadCommitID: "commit_id"}, expected: "commit_id/"},
		{c: Cumulus{UseCommitIDForCumulus: true, HeadCommitID: "", OpenGitFunc: func() (GitRepository, error) {
			return nil, expectedOpenGitError
		}}, expected: "/", expectedError: expectedOpenGitError},
		{c: Cumulus{UseCommitIDForCumulus: true, HeadCommitID: "", OpenGitFunc: func() (GitRepository, error) {
			hash := plumbing.ComputeHash(plumbing.CommitObject, []byte{1, 2, 3})
			return &gitRepositoryMock{
				revisionHash: &hash,
			}, nil
		}}, expected: "428ecf70bc22df0ba3dcf194b5ce53e769abab07/"},
		{c: Cumulus{UseCommitIDForCumulus: true, HeadCommitID: "", OpenGitFunc: func() (GitRepository, error) {
			return nil, expectedRevisionError
		}}, expected: "/", expectedError: expectedRevisionError},
	}

	for key, tt := range objectPathTests {
		t.Run(fmt.Sprintf("Row %v", key+1), func(t *testing.T) {
			pipelineRunKey, err := tt.c.GetPipelineRunKey()
			actualTargetPath := tt.c.GetCumulusPath(pipelineRunKey)

			if actualTargetPath != tt.expected {
				t.Errorf("Expected '%v' was '%v'", tt.expected, actualTargetPath)
			}
			if (err != nil && tt.expectedError == nil) || (err == nil && tt.expectedError != nil) || !errors.Is(err, tt.expectedError) {
				t.Errorf("Expected '%v' was '%v'", tt.expected, err)
			} else {
				t.Logf("ERROR FOUND actual %v, expected %v, result %v", err, tt.expectedError, errors.Is(err, tt.expectedError))
			}
		})
	}
}

func TestPrepareEnv(t *testing.T) {
	os.Setenv("TESTVAR1", "test1")

	c := Cumulus{EnvVars: []EnvVar{
		{Name: "TESTVAR1", Value: "test1_new"},
		{Name: "TESTVAR2", Value: "test2_new"},
	}}

	c.PrepareEnv()

	if c.EnvVars[0].Modified {
		t.Errorf("%v - expected '%v' was '%v'", c.EnvVars[0].Name, false, c.EnvVars[0].Modified)
	}
	if !c.EnvVars[1].Modified {
		t.Errorf("%v - expected '%v' was '%v'", c.EnvVars[1].Name, true, c.EnvVars[1].Modified)
	}

	os.Setenv("TESTVAR1", "")
	os.Setenv("TESTVAR2", "")
}

func TestCleanupEnv(t *testing.T) {
	os.Setenv("TESTVAR1", "test1")
	os.Setenv("TESTVAR2", "test2")

	c := Cumulus{EnvVars: []EnvVar{
		{Name: "TESTVAR1", Modified: false},
		{Name: "TESTVAR2", Modified: true},
	}}

	c.CleanupEnv()

	if os.Getenv("TESTVAR1") != "test1" {
		t.Errorf("%v - expected '%v' was '%v'", c.EnvVars[0].Name, "test1", os.Getenv("TESTVAR1"))
	}
	if len(os.Getenv("TESTVAR2")) > 0 {
		t.Errorf("%v - expected '%v' was '%v'", c.EnvVars[1].Name, "", os.Getenv("TESTVAR2"))
	}

	os.Setenv("TESTVAR1", "")
	os.Setenv("TESTVAR2", "")
}

type testStorageWriter struct {
	content string
}

func (t *testStorageWriter) Write(content []byte) (int, error) {
	t.content = string(content)
	return len(t.content), nil
}

type testSourceReader struct{}

func (r *testSourceReader) Read(b []byte) (int, error) {
	return 0, fmt.Errorf("error on Read")
}

// func TestUploadToStorage(t *testing.T) {

// 	sourceString := "Test Input"
// 	source := strings.NewReader(sourceString)

// 	t.Run(("Success case"), func(t *testing.T) {
// 		target := &testStorageWriter{content: ""}
// 		// var c Cumulus
// 		// c.UploadToStorage(source, target)
// 		if target.content != sourceString {
// 			t.Errorf("got '%v' want '%v'", target.content, sourceString)
// 		}
// 	})

// 	t.Run(("Copy error"), func(t *testing.T) {
// 		errReader := &testSourceReader{}
// 		target := &testStorageWriter{content: ""}
// 		var c Cumulus
// 		err := c.UploadToStorage(errReader, target)
// 		if fmt.Sprintf("%v", err) != "error on Read" {
// 			t.Errorf("expected error '%v' but was '%v'", "error on Read", err)
// 		}
// 	})
// }

func TestCumulus(t *testing.T) {
	tests := []struct {
		name      string
		c         Cumulus
		wantError string
	}{
		{
			name:      "all fields valid, UseCommitIDForCumulus false",
			c:         Cumulus{PipelineID: "p", Version: "v", StepResultType: "s", UseCommitIDForCumulus: false},
			wantError: "",
		},
		{
			name:      "all fields valid, UseCommitIDForCumulus true (skip version)",
			c:         Cumulus{PipelineID: "p", StepResultType: "s", UseCommitIDForCumulus: true},
			wantError: "",
		},
		{
			name:      "missing pipelineID",
			c:         Cumulus{PipelineID: "", Version: "v", StepResultType: "s", UseCommitIDForCumulus: false},
			wantError: "pipelineID must not be empty",
		},
		{
			name:      "missing version, UseCommitIDForCumulus false",
			c:         Cumulus{PipelineID: "p", Version: "", StepResultType: "s", UseCommitIDForCumulus: false},
			wantError: "version must not be empty",
		},
		{
			name:      "missing stepResultType",
			c:         Cumulus{PipelineID: "p", Version: "v", StepResultType: "", UseCommitIDForCumulus: false},
			wantError: "stepResultType must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.c.ValidateInput()
			if tt.wantError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
			}
		})
	}
}

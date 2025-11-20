//go:build unit
// +build unit

package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bmatcuk/doublestar"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/cumulus"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/cumulus/mocks"
	"google.golang.org/api/googleapi"
)

type myFileInfo struct {
	path string
	// Name() string       // base name of the file
	// Size() int64        // length in bytes for regular files; system-dependent for others
	// Mode() FileMode     // file mode bits
	// ModTime() time.Time // modification time
	// IsDir() bool        // abbreviation for Mode().IsDir()
	// Sys() interface{}   // underlying data source (can return nil)
}

func (m myFileInfo) Name() string {
	return ""
}

func (m myFileInfo) Size() int64 {
	return 0
}

func (m myFileInfo) Mode() os.FileMode {
	return os.FileMode(0)
}

func (m myFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (m myFileInfo) IsDir() bool {
	if strings.HasSuffix(m.path, "TEST-PATHclass1.xml") {
		return true
	}
	return false
}

func (m myFileInfo) Sys() interface{} {
	return nil
}

func TestSapCumulusUpload(t *testing.T) {

	var testCases = []struct {
		testName      string
		c             cumulus.Cumulus
		expected      []cumulus.Task
		detectedFiles []string
		expectedError error
	}{
		{
			testName:      "test empty directory",
			detectedFiles: []string{},
			c:             cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml"},
			expected:      []cumulus.Task{},
		},
		{
			testName:      "test single file matching the pattern",
			detectedFiles: []string{"path/file.yml", "TEST-com.sap.class1.xml"},
			c:             cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml"},
			expected:      []cumulus.Task{{SourcePath: "TEST-com.sap.class1.xml", TargetPath: "1.0.0/junit/TEST-com.sap.class1.xml"}},
		},
		{
			testName:      "test multiple files matching the pattern",
			detectedFiles: []string{"path/file.yml", "TEST-com.sap.class1.xml", "TEST-com.sap.class2.xml"},
			c:             cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml"},
			expected: []cumulus.Task{
				{SourcePath: "TEST-com.sap.class1.xml", TargetPath: "1.0.0/junit/TEST-com.sap.class1.xml"},
				{SourcePath: "TEST-com.sap.class2.xml", TargetPath: "1.0.0/junit/TEST-com.sap.class2.xml"},
			},
		},
		{
			testName:      "test multiple files matching pattern in subfolder",
			detectedFiles: []string{"path/file.yml", "TEST-com.sap.class1.xml", "TEST-com.sap.class2.xml"},
			c:             cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml", SubFolderPath: "subfolder"},
			expected: []cumulus.Task{
				{SourcePath: "TEST-com.sap.class1.xml", TargetPath: "1.0.0/junit/subfolder/TEST-com.sap.class1.xml"},
				{SourcePath: "TEST-com.sap.class2.xml", TargetPath: "1.0.0/junit/subfolder/TEST-com.sap.class2.xml"},
			},
		},
		{
			testName:      "test multiple files matching pattern in nested subfolder",
			detectedFiles: []string{"path/file.yml", "TEST-com.sap.class1.xml", "TEST-com.sap.class2.xml"},
			c:             cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml", SubFolderPath: "subfolder/nested"},
			expected: []cumulus.Task{
				{SourcePath: "TEST-com.sap.class1.xml", TargetPath: "1.0.0/junit/subfolder/nested/TEST-com.sap.class1.xml"},
				{SourcePath: "TEST-com.sap.class2.xml", TargetPath: "1.0.0/junit/subfolder/nested/TEST-com.sap.class2.xml"},
			},
		},
		{
			testName:      "test multiple files matching pattern in nested subfolder (trailing /)",
			detectedFiles: []string{"path/file.yml", "TEST-com.sap.class1.xml", "TEST-com.sap.class2.xml"},
			c:             cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml", SubFolderPath: "subfolder/nested/"},
			expected: []cumulus.Task{
				{SourcePath: "TEST-com.sap.class1.xml", TargetPath: "1.0.0/junit/subfolder/nested/TEST-com.sap.class1.xml"},
				{SourcePath: "TEST-com.sap.class2.xml", TargetPath: "1.0.0/junit/subfolder/nested/TEST-com.sap.class2.xml"},
			},
		},
		{
			testName:      "test multiple files without leading / matching the pattern",
			detectedFiles: []string{"path/file.yml", "TEST-com.sap.class1.xml", "TEST-com.sap.class2.xml"},
			c:             cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml"},
			expected: []cumulus.Task{
				{SourcePath: "TEST-com.sap.class1.xml", TargetPath: "1.0.0/junit/TEST-com.sap.class1.xml"},
				{SourcePath: "TEST-com.sap.class2.xml", TargetPath: "1.0.0/junit/TEST-com.sap.class2.xml"},
			},
		},
		{
			testName:      "test file in subfolder matching the pattern",
			detectedFiles: []string{"path/TEST-file.xml", "TEST-com.sap.class1.xml", "TEST-com.sap.class2.xml"},
			c:             cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml", SubFolderPath: "subfolder/nested/"},
			expected: []cumulus.Task{
				{SourcePath: "TEST-com.sap.class1.xml", TargetPath: "1.0.0/junit/subfolder/nested/TEST-com.sap.class1.xml"},
				{SourcePath: "TEST-com.sap.class2.xml", TargetPath: "1.0.0/junit/subfolder/nested/TEST-com.sap.class2.xml"},
			},
		},
		{
			testName:      "test empty pipelineID should throw an error",
			detectedFiles: []string{},
			c:             cumulus.Cumulus{PipelineID: "", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml"},
			expected:      []cumulus.Task{},
			expectedError: errors.New("pipelineID must not be empty"),
		},
		{
			testName:      "test empty version should throw an error",
			detectedFiles: []string{},
			c:             cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "", StepResultType: "junit", FilePattern: "TEST-*.xml"},
			expected:      []cumulus.Task{},
			expectedError: errors.New("version must not be empty"),
		},
		{
			testName:      "test empty stepResultType should throw an error",
			detectedFiles: []string{},
			c:             cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "", FilePattern: "TEST-*.xml"},
			expected:      []cumulus.Task{},
			expectedError: errors.New("stepResultType must not be empty"),
		},
		{
			testName:      "test stepResultType root not extending targetpath",
			detectedFiles: []string{"path/file.yml", "TEST-com.sap.class1.xml", "TEST-com.sap.class2.xml"},
			c:             cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "root", FilePattern: "TEST-*.xml"},
			expected: []cumulus.Task{
				{SourcePath: "TEST-com.sap.class1.xml", TargetPath: "1.0.0/TEST-com.sap.class1.xml"},
				{SourcePath: "TEST-com.sap.class2.xml", TargetPath: "1.0.0/TEST-com.sap.class2.xml"},
			},
		},
		{
			testName:      "test recursive wildcard empty dir",
			detectedFiles: []string{},
			c:             cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "**"},
			expected:      []cumulus.Task{},
		},
		{
			testName:      "test recursive wildcard all files",
			detectedFiles: []string{"file.yaml", "path/class1.xml", "path/class2.xml", "path2/class3.xml", "path2/deeper/class4.xml"},
			c:             cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "**"},
			expected: []cumulus.Task{
				{SourcePath: "file.yaml", TargetPath: "1.0.0/junit/file.yaml"},
				{SourcePath: "path/class1.xml", TargetPath: "1.0.0/junit/path/class1.xml"},
				{SourcePath: "path/class2.xml", TargetPath: "1.0.0/junit/path/class2.xml"},
				{SourcePath: "path2/class3.xml", TargetPath: "1.0.0/junit/path2/class3.xml"},
				{SourcePath: "path2/deeper/class4.xml", TargetPath: "1.0.0/junit/path2/deeper/class4.xml"},
			},
		},
		{
			testName: "test recursive wildcard all files same name",
			detectedFiles: []string{"pom.xml", "pom.yaml",
				"proj/TEST-com.sap.class1.xml",
				"proj/pom.xml",
				"proj2/subproj/TEST-com.sap.class2.xml",
				"proj2/subproj/pom.xml"},
			c: cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "**/pom.xml"},
			expected: []cumulus.Task{
				{SourcePath: "pom.xml", TargetPath: "1.0.0/junit/pom.xml"},
				{SourcePath: "proj/pom.xml", TargetPath: "1.0.0/junit/proj/pom.xml"},
				{SourcePath: "proj2/subproj/pom.xml", TargetPath: "1.0.0/junit/proj2/subproj/pom.xml"},
			},
		},
		{
			testName: "test recursive wildcard with pattern in file name",
			detectedFiles: []string{"path/file.yaml",
				"TEST-pom.xml",
				"path/TEST-com.sap.class1.xml",
				"path/PROD-com.sap.class1.xml",
				"path/TEST-com.sap.class2.xml",
				"path2/TEST-com.sap.class3.xml",
				"path2/TEST-com.sap.class3.yaml",
				"path2/TEST-com.sap.class4.xml"},
			c: cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "**/TEST-*.xml"},
			expected: []cumulus.Task{
				{SourcePath: "TEST-pom.xml", TargetPath: "1.0.0/junit/TEST-pom.xml"},
				{SourcePath: "path/TEST-com.sap.class1.xml", TargetPath: "1.0.0/junit/path/TEST-com.sap.class1.xml"},
				{SourcePath: "path/TEST-com.sap.class2.xml", TargetPath: "1.0.0/junit/path/TEST-com.sap.class2.xml"},
				{SourcePath: "path2/TEST-com.sap.class3.xml", TargetPath: "1.0.0/junit/path2/TEST-com.sap.class3.xml"},
				{SourcePath: "path2/TEST-com.sap.class4.xml", TargetPath: "1.0.0/junit/path2/TEST-com.sap.class4.xml"},
			},
		},
		{
			testName: "test multiple FilePattern",
			detectedFiles: []string{
				"path/TEST-com.sap.class1.xml",
				"path/PROD-com.sap.class1.xml",
				"path/TEST-PATHclass1.xml", // this is meant to be a path
				"path/TEST-com.sap.class2.xml"},
			c: cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "**/TEST-*class1.xml, **/TEST-*class2.xml"},
			expected: []cumulus.Task{
				{SourcePath: "path/TEST-com.sap.class1.xml", TargetPath: "1.0.0/junit/path/TEST-com.sap.class1.xml"},
				{SourcePath: "path/TEST-com.sap.class2.xml", TargetPath: "1.0.0/junit/path/TEST-com.sap.class2.xml"},
			},
		},
		{
			testName: "test multiple FilePattern with multiple matches",
			detectedFiles: []string{
				"path/TEST-com.sap.class1.xml",
				"path/TEST-com.sap.class2.xml"},
			c: cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "**/TEST-*.xml, path/TEST-*.xml, **/*class1.xml"},
			expected: []cumulus.Task{
				{SourcePath: "path/TEST-com.sap.class1.xml", TargetPath: "1.0.0/junit/path/TEST-com.sap.class1.xml"},
				{SourcePath: "path/TEST-com.sap.class2.xml", TargetPath: "1.0.0/junit/path/TEST-com.sap.class2.xml"},
			},
		},
		{
			testName:      "test single file matching the pattern using commit id only",
			detectedFiles: []string{"path/file.yml", "TEST-com.sap.class1.xml"},
			c:             cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml", UseCommitIDForCumulus: true, HeadCommitID: "commit_id"},
			expected:      []cumulus.Task{{SourcePath: "TEST-com.sap.class1.xml", TargetPath: "commit_id/junit/TEST-com.sap.class1.xml"}},
		},
		{
			testName:      "test single file matching the pattern using commit id requested from github",
			detectedFiles: []string{"path/file.yml", "TEST-com.sap.class1.xml"},
			c: cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml", UseCommitIDForCumulus: true, HeadCommitID: "", OpenGitFunc: func() (cumulus.GitRepository, error) {
				hash := plumbing.ComputeHash(plumbing.CommitObject, []byte{1, 2, 3})
				return &gitRepositoryMock{
					revisionHash: &hash,
				}, nil
			}},
			expected: []cumulus.Task{{SourcePath: "TEST-com.sap.class1.xml", TargetPath: "428ecf70bc22df0ba3dcf194b5ce53e769abab07/junit/TEST-com.sap.class1.xml"}},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			ctx := context.Background()

			mockedClient := &mocks.GCSClient{}
			for _, expectation := range tt.expected {
				mockedClient.Mock.On("UploadFile", ctx, tt.c.PipelineID, expectation.SourcePath, expectation.TargetPath).Return(func(ctx context.Context, pipelineId string, sourcePath string, targetPath string) error { return nil }).Once()
			}

			searchFn := func(path string) ([]string, error) {
				matchedFiles := []string{}
				for _, value := range tt.detectedFiles {
					match, _ := doublestar.Match(path, value)
					if match {
						matchedFiles = append(matchedFiles, value)
					}
				}
				return matchedFiles, nil
			}

			fileInfoFn := func(name string) (os.FileInfo, error) {
				return myFileInfo{name}, nil
			}

			err := runSapCumulusUpload(ctx, &sapCumulusUploadCommonPipelineEnvironment{}, mockedClient, &tt.c, searchFn, fileInfoFn)
			if err != nil {
				if tt.expectedError != nil {
					assert.Equal(t, tt.expectedError.Error(), err.Error())
					return
				}
				t.Error(err)
				return
			} else if tt.expectedError != nil {
				t.Error("expected an error, but didn't get one")
				return
			}
			mockedClient.Mock.AssertNumberOfCalls(t, "UploadFile", len(tt.expected))
			mockedClient.Mock.AssertExpectations(t)
		})
	}
}

func TestUploadFiles(t *testing.T) {
	targetFolder := "1.0.0/junit"
	config := cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "**/TEST-*.xml, path/TEST-*.xml, **/*class1.xml"}
	uploadTargets := []cumulus.Task{
		{SourcePath: "path/TEST-com.sap.class1.xml", TargetPath: "path/TEST-com.sap.class1.xml"},
		{SourcePath: "path/TEST-com.sap.class2.xml", TargetPath: "path/TEST-com.sap.class2.xml"},
	}

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		mockedClient := &mocks.GCSClient{}
		for _, uploadTarget := range uploadTargets {
			mockedClient.Mock.On("UploadFile", ctx, config.PipelineID, uploadTarget.SourcePath, filepath.Join(targetFolder, uploadTarget.TargetPath)).Return(nil).Once()
		}

		err := uploadFiles(ctx, mockedClient, &config, &uploadTargets, targetFolder)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		ctx := context.Background()
		mockedClient := &mocks.GCSClient{}
		for _, uploadTarget := range uploadTargets {
			mockedClient.Mock.On("UploadFile", ctx, config.PipelineID, uploadTarget.SourcePath, filepath.Join(targetFolder, uploadTarget.TargetPath)).Return(fmt.Errorf("err")).Once()
		}

		err := uploadFiles(ctx, mockedClient, &config, &uploadTargets, targetFolder)
		assert.Error(t, err)
		assert.EqualError(t, err, "could not upload files to cumulus bucket: err: err")
	})

	t.Run("404 error (bucketLockedWarning true)", func(t *testing.T) {
		ctx := context.Background()
		mockedClient := &mocks.GCSClient{}
		config.BucketLockedWarning = true
		e := &googleapi.Error{Code: http.StatusNotFound, Message: "not found"}
		for _, uploadTarget := range uploadTargets {
			mockedClient.Mock.On("UploadFile", ctx, config.PipelineID, uploadTarget.SourcePath, filepath.Join(targetFolder, uploadTarget.TargetPath)).Return(e).Once()
		}

		err := uploadFiles(ctx, mockedClient, &config, &uploadTargets, targetFolder)
		assert.Error(t, err)
		assert.EqualError(t, err, "could not upload files to cumulus bucket: googleapi: Error 404: not found: googleapi: Error 404: not found")
	})

	t.Run("bucket is locked error (bucketLockedWarning false)", func(t *testing.T) {
		ctx := context.Background()
		mockedClient := &mocks.GCSClient{}
		config.BucketLockedWarning = false
		e := &googleapi.Error{Code: http.StatusForbidden}
		for _, uploadTarget := range uploadTargets {
			e.Message = fmt.Sprintf("Object '%s' is under active Temporary hold and cannot be deleted, overwritten or archived until hold is removed.", filepath.Join(targetFolder, uploadTarget.TargetPath))
			mockedClient.Mock.On("UploadFile", ctx, config.PipelineID, uploadTarget.SourcePath, filepath.Join(targetFolder, uploadTarget.TargetPath)).Return(e).Once()
		}

		err := uploadFiles(ctx, mockedClient, &config, &uploadTargets, targetFolder)
		assert.Error(t, err)
		assert.EqualError(t, err, "could not upload files to cumulus bucket: googleapi: Error 403: Object '1.0.0/junit/path/TEST-com.sap.class2.xml' is under active Temporary hold and cannot be deleted, overwritten or archived until hold is removed.: googleapi: Error 403: Object '1.0.0/junit/path/TEST-com.sap.class2.xml' is under active Temporary hold and cannot be deleted, overwritten or archived until hold is removed.")
	})

	t.Run("bucket is locked error (bucketLockedWarning true)", func(t *testing.T) {
		ctx := context.Background()
		mockedClient := &mocks.GCSClient{}
		config.BucketLockedWarning = true
		e := &googleapi.Error{Code: http.StatusForbidden}
		for _, uploadTarget := range uploadTargets {
			e.Message = fmt.Sprintf("Object '%s' is under active Temporary hold and cannot be deleted, overwritten or archived until hold is removed.", filepath.Join(targetFolder, uploadTarget.TargetPath))
			mockedClient.Mock.On("UploadFile", ctx, config.PipelineID, uploadTarget.SourcePath, filepath.Join(targetFolder, uploadTarget.TargetPath)).Return(e).Once()
		}

		err := uploadFiles(ctx, mockedClient, &config, &uploadTargets, targetFolder)
		assert.NoError(t, err)
	})
}

//go:build unit
// +build unit

package cmd

import (
	"context"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/cumulus"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/cumulus/mocks"
)

type gitRepositoryMock struct {
	revisionHash  *plumbing.Hash
	revisionError error
}

func (r *gitRepositoryMock) ResolveRevision(rev plumbing.Revision) (*plumbing.Hash, error) {
	return r.revisionHash, r.revisionError
}

func TestSapCumulusDownload(t *testing.T) {

	var testCases = []struct {
		testName      string
		c             cumulus.Cumulus
		expected      []cumulus.Task
		cumulusFiles  []string
		expectedError error
	}{
		{
			testName:     "test empty directory",
			cumulusFiles: []string{},
			c:            cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml"},
			expected:     []cumulus.Task{},
		},
		{
			testName:     "test single file matching the pattern",
			cumulusFiles: []string{"1.0.0/junit/path/file.yml", "1.0.0/junit/TEST-com.sap.class1.xml"},
			c:            cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml"},
			expected:     []cumulus.Task{{SourcePath: "1.0.0/junit/TEST-com.sap.class1.xml", TargetPath: "TEST-com.sap.class1.xml"}},
		},
		{
			testName:     "test multiple files matching the pattern",
			cumulusFiles: []string{"1.0.0/junit/path/file.yml", "1.0.0/junit/TEST-com.sap.class1.xml", "1.0.0/junit/TEST-com.sap.class2.xml"},
			c:            cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml"},
			expected: []cumulus.Task{
				{SourcePath: "1.0.0/junit/TEST-com.sap.class1.xml", TargetPath: "TEST-com.sap.class1.xml"},
				{SourcePath: "1.0.0/junit/TEST-com.sap.class2.xml", TargetPath: "TEST-com.sap.class2.xml"},
			},
		},
		{
			testName: "test multiple files matching pattern in subfolder",
			cumulusFiles: []string{
				"1.0.0/junit/subfolder/path/file.yml",
				"1.0.0/junit/subfolder/TEST-com.sap.class1.xml",
				"1.0.0/junit/subfolder/TEST-com.sap.class2.xml",
			},
			c: cumulus.Cumulus{
				PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml", SubFolderPath: "subfolder"},
			expected: []cumulus.Task{
				{SourcePath: "1.0.0/junit/subfolder/TEST-com.sap.class1.xml", TargetPath: "TEST-com.sap.class1.xml"},
				{SourcePath: "1.0.0/junit/subfolder/TEST-com.sap.class2.xml", TargetPath: "TEST-com.sap.class2.xml"},
			},
		},
		{
			testName: "test multiple files matching pattern in nested subfolder",
			cumulusFiles: []string{
				"1.0.0/junit/subfolder/nested/path/file.yml",
				"1.0.0/junit/subfolder/nested/TEST-com.sap.class1.xml",
				"1.0.0/junit/subfolder/nested/TEST-com.sap.class2.xml",
			},
			c: cumulus.Cumulus{
				PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml",
				SubFolderPath: "subfolder/nested",
			},
			expected: []cumulus.Task{
				{SourcePath: "1.0.0/junit/subfolder/nested/TEST-com.sap.class1.xml", TargetPath: "TEST-com.sap.class1.xml"},
				{SourcePath: "1.0.0/junit/subfolder/nested/TEST-com.sap.class2.xml", TargetPath: "TEST-com.sap.class2.xml"},
			},
		},
		{
			testName: "test multiple files matching pattern in nested subfolder (trailing /)",
			cumulusFiles: []string{
				"1.0.0/junit/subfolder/nested/path/file.yml",
				"1.0.0/junit/subfolder/nested/TEST-com.sap.class1.xml",
				"1.0.0/junit/subfolder/nested/TEST-com.sap.class2.xml",
			},
			c: cumulus.Cumulus{
				PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml",
				SubFolderPath: "subfolder/nested/"},
			expected: []cumulus.Task{
				{SourcePath: "1.0.0/junit/subfolder/nested/TEST-com.sap.class1.xml", TargetPath: "TEST-com.sap.class1.xml"},
				{SourcePath: "1.0.0/junit/subfolder/nested/TEST-com.sap.class2.xml", TargetPath: "TEST-com.sap.class2.xml"},
			},
		},
		{
			testName: "test multiple files without leading / matching the pattern",
			cumulusFiles: []string{
				"1.0.0/junit/path/file.yml",
				"1.0.0/junit/TEST-com.sap.class1.xml",
				"1.0.0/junit/TEST-com.sap.class2.xml",
			},
			c: cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml"},
			expected: []cumulus.Task{
				{SourcePath: "1.0.0/junit/TEST-com.sap.class1.xml", TargetPath: "TEST-com.sap.class1.xml"},
				{SourcePath: "1.0.0/junit/TEST-com.sap.class2.xml", TargetPath: "TEST-com.sap.class2.xml"},
			},
		},
		{
			testName: "test file in subfolder matching the pattern",
			cumulusFiles: []string{
				"1.0.0/junit/subfolder/nested/path/TEST-file.xml",
				"1.0.0/junit/subfolder/nested/TEST-com.sap.class1.xml",
				"1.0.0/junit/subfolder/nested/TEST-com.sap.class2.xml",
			},
			c: cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml",
				SubFolderPath: "subfolder/nested/"},
			expected: []cumulus.Task{
				{SourcePath: "1.0.0/junit/subfolder/nested/TEST-com.sap.class1.xml", TargetPath: "TEST-com.sap.class1.xml"},
				{SourcePath: "1.0.0/junit/subfolder/nested/TEST-com.sap.class2.xml", TargetPath: "TEST-com.sap.class2.xml"},
			},
		},
		{
			testName:      "test empty pipelineID should throw an error",
			cumulusFiles:  []string{},
			c:             cumulus.Cumulus{PipelineID: "", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml"},
			expected:      []cumulus.Task{},
			expectedError: errors.New("pipelineID must not be empty"),
		},
		{
			testName:      "test empty version should throw an error",
			cumulusFiles:  []string{},
			c:             cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "", StepResultType: "junit", FilePattern: "TEST-*.xml"},
			expected:      []cumulus.Task{},
			expectedError: errors.New("version must not be empty"),
		},
		{
			testName:      "test empty stepResultType should throw an error",
			cumulusFiles:  []string{},
			c:             cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "", FilePattern: "TEST-*.xml"},
			expected:      []cumulus.Task{},
			expectedError: errors.New("stepResultType must not be empty"),
		},
		{
			testName: "test stepResultType root not extending targetpath",
			cumulusFiles: []string{
				"1.0.0/path/file.yml",
				"1.0.0/TEST-com.sap.class1.xml",
				"1.0.0/TEST-com.sap.class2.xml",
			},
			c: cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "root", FilePattern: "TEST-*.xml"},
			expected: []cumulus.Task{
				{SourcePath: "1.0.0/TEST-com.sap.class1.xml", TargetPath: "TEST-com.sap.class1.xml"},
				{SourcePath: "1.0.0/TEST-com.sap.class2.xml", TargetPath: "TEST-com.sap.class2.xml"},
			},
		},
		{
			testName:     "test recursive wildcard empty dir",
			cumulusFiles: []string{},
			c:            cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "**"},
			expected:     []cumulus.Task{},
		},
		{
			testName: "test recursive wildcard all files",
			cumulusFiles: []string{
				"1.0.0/junit/file.yaml",
				"1.0.0/junit/path/class1.xml",
				"1.0.0/junit/path/class2.xml",
				"1.0.0/junit/path2/class3.xml",
				"1.0.0/junit/path2/deeper/class4.xml",
			},
			c: cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "**"},
			expected: []cumulus.Task{
				{SourcePath: "1.0.0/junit/file.yaml", TargetPath: "file.yaml"},
				{SourcePath: "1.0.0/junit/path/class1.xml", TargetPath: "path/class1.xml"},
				{SourcePath: "1.0.0/junit/path/class2.xml", TargetPath: "path/class2.xml"},
				{SourcePath: "1.0.0/junit/path2/class3.xml", TargetPath: "path2/class3.xml"},
				{SourcePath: "1.0.0/junit/path2/deeper/class4.xml", TargetPath: "path2/deeper/class4.xml"},
			},
		},
		{
			testName: "test recursive wildcard all files same name",
			cumulusFiles: []string{
				"1.0.0/junit/pom.xml", "pom.yaml",
				"1.0.0/junit/proj/TEST-com.sap.class1.xml",
				"1.0.0/junit/proj/pom.xml",
				"1.0.0/junit/proj2/subproj/TEST-com.sap.class2.xml",
				"1.0.0/junit/proj2/subproj/pom.xml"},
			c: cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "**/pom.xml"},
			expected: []cumulus.Task{
				{SourcePath: "1.0.0/junit/pom.xml", TargetPath: "pom.xml"},
				{SourcePath: "1.0.0/junit/proj/pom.xml", TargetPath: "proj/pom.xml"},
				{SourcePath: "1.0.0/junit/proj2/subproj/pom.xml", TargetPath: "proj2/subproj/pom.xml"},
			},
		},
		{
			testName: "test recursive wildcard with pattern in file name",
			cumulusFiles: []string{
				"1.0.0/junit/path/file.yaml",
				"1.0.0/junit/TEST-pom.xml",
				"1.0.0/junit/path/TEST-com.sap.class1.xml",
				"1.0.0/junit/path/PROD-com.sap.class1.xml",
				"1.0.0/junit/path/TEST-com.sap.class2.xml",
				"1.0.0/junit/path2/TEST-com.sap.class3.xml",
				"1.0.0/junit/path2/TEST-com.sap.class3.yaml",
				"1.0.0/junit/path2/TEST-com.sap.class4.xml"},
			c: cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "**/TEST-*.xml"},
			expected: []cumulus.Task{
				{SourcePath: "1.0.0/junit/TEST-pom.xml", TargetPath: "TEST-pom.xml"},
				{SourcePath: "1.0.0/junit/path/TEST-com.sap.class1.xml", TargetPath: "path/TEST-com.sap.class1.xml"},
				{SourcePath: "1.0.0/junit/path/TEST-com.sap.class2.xml", TargetPath: "path/TEST-com.sap.class2.xml"},
				{SourcePath: "1.0.0/junit/path2/TEST-com.sap.class3.xml", TargetPath: "path2/TEST-com.sap.class3.xml"},
				{SourcePath: "1.0.0/junit/path2/TEST-com.sap.class4.xml", TargetPath: "path2/TEST-com.sap.class4.xml"},
			},
		},
		{
			testName: "test multiple FilePattern",
			cumulusFiles: []string{
				"1.0.0/junit/path/TEST-com.sap.class1.xml",
				"1.0.0/junit/path/PROD-com.sap.class1.xml",
				"1.0.0/junit/path/TEST-PATHclass1.xml",
				"1.0.0/junit/path/TEST-com.sap.class2.xml"},
			c: cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit",
				FilePattern: "**/TEST-*class1.xml, **/TEST-*class2.xml"},
			expected: []cumulus.Task{
				{SourcePath: "1.0.0/junit/path/TEST-com.sap.class1.xml", TargetPath: "path/TEST-com.sap.class1.xml"},
				{SourcePath: "1.0.0/junit/path/TEST-PATHclass1.xml", TargetPath: "path/TEST-PATHclass1.xml"},
				{SourcePath: "1.0.0/junit/path/TEST-com.sap.class2.xml", TargetPath: "path/TEST-com.sap.class2.xml"},
			},
		},
		{
			testName: "test multiple FilePattern with repetition",
			cumulusFiles: []string{
				"1.0.0/junit/path/TEST-com.sap.class1.xml",
				"1.0.0/junit/path/PROD-com.sap.class1.xml",
				"1.0.0/junit/path/TEST-PATHclass1.xml",
				"1.0.0/junit/path/TEST-com.sap.class2.xml"},
			c: cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit",
				FilePattern: "**/TEST-*class1.xml, **/TEST-*class2.xml, **/TEST-*class2.xml"},
			expected: []cumulus.Task{
				{SourcePath: "1.0.0/junit/path/TEST-com.sap.class1.xml", TargetPath: "path/TEST-com.sap.class1.xml"},
				{SourcePath: "1.0.0/junit/path/TEST-PATHclass1.xml", TargetPath: "path/TEST-PATHclass1.xml"},
				{SourcePath: "1.0.0/junit/path/TEST-com.sap.class2.xml", TargetPath: "path/TEST-com.sap.class2.xml"},
			},
		},
		{
			testName: "test with targetPath",
			cumulusFiles: []string{
				"1.0.0/junit/subfolder/path/file.yml",
				"1.0.0/junit/subfolder/TEST-com.sap.class1.xml",
				"1.0.0/junit/subfolder/TEST-com.sap.class2.xml",
			},
			c: cumulus.Cumulus{
				PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml", SubFolderPath: "subfolder",
				TargetPath: "target/path"},
			expected: []cumulus.Task{
				{SourcePath: "1.0.0/junit/subfolder/TEST-com.sap.class1.xml", TargetPath: "target/path/TEST-com.sap.class1.xml"},
				{SourcePath: "1.0.0/junit/subfolder/TEST-com.sap.class2.xml", TargetPath: "target/path/TEST-com.sap.class2.xml"},
			},
		},
		{
			testName: "test with targetPath (trailing /)",
			cumulusFiles: []string{
				"1.0.0/junit/subfolder/path/file.yml",
				"1.0.0/junit/subfolder/TEST-com.sap.class1.xml",
				"1.0.0/junit/subfolder/TEST-com.sap.class2.xml",
			},
			c: cumulus.Cumulus{
				PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml", SubFolderPath: "subfolder",
				TargetPath: "target/path/"},
			expected: []cumulus.Task{
				{SourcePath: "1.0.0/junit/subfolder/TEST-com.sap.class1.xml", TargetPath: "target/path/TEST-com.sap.class1.xml"},
				{SourcePath: "1.0.0/junit/subfolder/TEST-com.sap.class2.xml", TargetPath: "target/path/TEST-com.sap.class2.xml"},
			},
		},
		{
			testName: "test several versions in cumulus",
			cumulusFiles: []string{
				"1.0.1/junit/path/file.yml",
				"1.0.0/junit/TEST-com.sap.class1.xml",
				"1.0.0/junit/TEST-com.sap.class2.xml",
				"v3/junit/TEST-com.sap.class2.xml",
				"2.0.1/junit/path/file.yml",
			},
			c: cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml"},
			expected: []cumulus.Task{
				{SourcePath: "1.0.0/junit/TEST-com.sap.class1.xml", TargetPath: "TEST-com.sap.class1.xml"},
				{SourcePath: "1.0.0/junit/TEST-com.sap.class2.xml", TargetPath: "TEST-com.sap.class2.xml"},
			},
		},
		{
			testName: "test several step result types in cumulus",
			cumulusFiles: []string{
				"1.0.0/general/path/file.yml",
				"1.0.0/junit/TEST-com.sap.class1.xml",
				"1.0.0/junit/TEST-com.sap.class2.xml",
				"1.0.0/test/TEST-com.sap.class2.xml",
				"1.0.0/test/path/file.yml",
			},
			c: cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml"},
			expected: []cumulus.Task{
				{SourcePath: "1.0.0/junit/TEST-com.sap.class1.xml", TargetPath: "TEST-com.sap.class1.xml"},
				{SourcePath: "1.0.0/junit/TEST-com.sap.class2.xml", TargetPath: "TEST-com.sap.class2.xml"},
			},
		},
		{
			testName: "test several subfolders in cumulus",
			cumulusFiles: []string{
				"1.0.0/junit/temp/path/file.yml",
				"1.0.0/junit/subfolder/TEST-com.sap.class1.xml",
				"1.0.0/junit/subfolder/TEST-com.sap.class2.xml",
				"1.0.0/junit/subfolder1/TEST-com.sap.class2.xml",
				"1.0.0/junit/sub/folder/path/file.yml",
			},
			c: cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml",
				SubFolderPath: "subfolder"},
			expected: []cumulus.Task{
				{SourcePath: "1.0.0/junit/subfolder/TEST-com.sap.class1.xml", TargetPath: "TEST-com.sap.class1.xml"},
				{SourcePath: "1.0.0/junit/subfolder/TEST-com.sap.class2.xml", TargetPath: "TEST-com.sap.class2.xml"},
			},
		},
		{
			testName: "test multiple filePattern for version, stepResultType, subfolder",
			cumulusFiles: []string{
				"1.0.0/junit/temp/path/file.yml",
				"1.0.0/junit/subfolder/TEST-com.sap.class1.xml",
				"1.0.0/junit/subfolder/TEST-com.sap.class2.xml",
				"1.0.0/junit/subfolder1/TEST-com.sap.class2.xml",
				"1.0.0/junit/sub/folder/path/file.yml",
			},
			c: cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit",
				FilePattern: "**1.0.0/**, **/junit/**, **/subfolder/**", SubFolderPath: "subfolder"},
			expected: []cumulus.Task{},
		},
		{
			testName:     "test single file matching the pattern using commit id only",
			cumulusFiles: []string{"commit_id/junit/path/file.yml", "commit_id/junit/TEST-com.sap.class1.xml"},
			c:            cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml", UseCommitIDForCumulus: true, HeadCommitID: "commit_id"},
			expected:     []cumulus.Task{{SourcePath: "commit_id/junit/TEST-com.sap.class1.xml", TargetPath: "TEST-com.sap.class1.xml"}},
		},
		{
			testName:     "test single file matching the pattern using commit id requested from github",
			cumulusFiles: []string{"428ecf70bc22df0ba3dcf194b5ce53e769abab07/junit/path/file.yml", "428ecf70bc22df0ba3dcf194b5ce53e769abab07/junit/TEST-com.sap.class1.xml"},
			c: cumulus.Cumulus{PipelineID: "test-pipeline-1", Version: "1.0.0", StepResultType: "junit", FilePattern: "TEST-*.xml", UseCommitIDForCumulus: true, HeadCommitID: "", OpenGitFunc: func() (cumulus.GitRepository, error) {
				hash := plumbing.ComputeHash(plumbing.CommitObject, []byte{1, 2, 3})
				return &gitRepositoryMock{
					revisionHash: &hash,
				}, nil
			}},
			expected: []cumulus.Task{{SourcePath: "428ecf70bc22df0ba3dcf194b5ce53e769abab07/junit/TEST-com.sap.class1.xml", TargetPath: "TEST-com.sap.class1.xml"}},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			ctx := context.Background()

			mockedClient := &mocks.GCSClient{}

			mockedClient.Mock.On("ListFiles", ctx, tt.c.PipelineID).Return(
				func(ctx context.Context, bucketID string) []string {
					return tt.cumulusFiles
				},
				func(ctx context.Context, bucketID string) error {
					return nil
				})
			pipelineRunKey, _ := tt.c.GetPipelineRunKey()
			targetPath := tt.c.GetCumulusPath(pipelineRunKey)
			tasks, err := searchCumulusFiles(ctx, mockedClient, tt.c.FilePattern, tt.c.PipelineID, tt.c.TargetPath, targetPath)
			if err != nil {
				t.Error(err)
			}

			for _, task := range tasks {
				mockedClient.Mock.On("DownloadFile", ctx, tt.c.PipelineID, task.SourcePath, task.TargetPath).Return(func(ctx context.Context, pipelineId string, sourcePath string, targetPath string) error { return nil }).Once()
			}

			err = runSapCumulusDownload(ctx, mockedClient, &tt.c)
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
			mockedClient.Mock.AssertNumberOfCalls(t, "DownloadFile", len(tt.expected))
			mockedClient.Mock.AssertExpectations(t)
			assert.Equal(t, tt.expected, tasks)
		})
	}
}

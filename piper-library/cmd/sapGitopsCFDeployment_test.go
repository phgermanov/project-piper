package cmd

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/stretchr/testify/assert"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

var existingYaml = `apiVersion: sap.crossplaneprovidermta.orchestrate.cloud.sap/v1alpha1
kind: Mta
metadata:
  name: dev-bookshop
  namespace: default
spec:
  forProvider:
    file:
      credentialsSecretRef:
        name: artifactory-credentials-secret
        namespace: default
      url: https://common.repositories.cloud.sap/deploy-releases-hyperspace-maven/com/sap/hyperspace/bookshkop/1.0.0-20250325163455+cb9c2d7dde748a3b1f1d74f9c4649f86531bc397/bookshkop-1.0.0-20250325163455+cb9c2d7dde748a3b1f1d74f9c4649f86531bc397.mtar
    namespace: default
    spaceRef:
      name: dev-space
      policy:
        resolve: Always`

type sapGitopsCFDeploymentMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	*mock.HttpClientMock
	addPaths           []string
	changedBranch      string
	commitMessage      string
	commitOpts         git.CommitOptions
	cloneUsername      string
	clonePassword      string
	cloneBranch        string
	cloneServer        string
	cloneTemp          string
	cloneCACerts       []byte
	failOnAdd          bool
	failOnChangeBranch bool
	failOnCommit       bool
	failOnClone        bool
	failOnPush         bool
	forcePush          bool
	pushCACerts        []byte
	pushUsername       string
	pushPassword       string
	skipClone          bool
}

func (s *sapGitopsCFDeploymentMockUtils) Add(path string) (plumbing.Hash, error) {
	if s.failOnAdd {
		return plumbing.Hash{}, errors.New("error on add")
	}
	s.addPaths = append(s.addPaths, path)
	return plumbing.Hash{123}, nil
}

func (s *sapGitopsCFDeploymentMockUtils) Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error) {
	if s.failOnCommit {
		return plumbing.Hash{}, errors.New("error on commit")
	}
	s.commitMessage = msg
	s.commitOpts = *opts
	return plumbing.Hash{123}, nil
}

func (s *sapGitopsCFDeploymentMockUtils) PushChangesToRepository(pushUsername string, pushPassword string, force *bool, caCerts []byte) error {
	s.pushUsername = pushUsername
	s.pushPassword = pushPassword
	s.forcePush = *force
	s.pushCACerts = caCerts

	if s.failOnPush {
		return errors.New("error on push")
	}

	return nil
}

func (s *sapGitopsCFDeploymentMockUtils) PlainCloneAndWorktree(cloneUsername, clonePassword, cloneServer, cloneBranch, directory string, caCerts []byte) error {
	if s.skipClone {
		return nil
	}
	if s.failOnClone {
		return errors.New("error on clone")
	}
	s.cloneUsername = cloneUsername
	s.clonePassword = clonePassword
	s.cloneServer = cloneServer
	s.cloneBranch = cloneBranch
	s.cloneTemp = directory
	s.cloneCACerts = caCerts

	err := s.MkdirAll(filepath.Join(directory, "dir1/dir2"), 0o755)
	if err != nil {
		return err
	}
	s.AddFile(filepath.Join(directory, "path/to/mta.yaml"), []byte(existingYaml))

	return nil
}

func (s *sapGitopsCFDeploymentMockUtils) ChangeBranch(branchName string) error {
	if s.failOnChangeBranch {
		return errors.New("error on change branch")
	}
	s.changedBranch = branchName
	return nil
}

func newSapGitopsCFDeploymentTestsUtils() sapGitopsCFDeploymentMockUtils {
	utils := sapGitopsCFDeploymentMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestNewSapGitopsCFDeploymentUtils(t *testing.T) {
	t.Parallel()

	utils := newSapGitopsCFDeploymentUtils()
	assert.NotNil(t, utils.Command)
	assert.Equal(t, utils.Files, &piperutils.Files{})

	assert.Equal(t, log.Writer(), utils.GetStdout())
	assert.Equal(t, log.Writer(), utils.GetStderr())
}

func TestRunSapGitopsCFDeployment(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		config := sapGitopsCFDeploymentOptions{
			FilePath:   "path/to/mta.yaml",
			MtarURLs:   []string{"https://the.url/to/file.mtar"},
			ServerURL:  "https://the.server.url",
			BranchName: "theBranch",
			Username:   "theUser",
			Password:   "thePassword",
		}
		utils := newSapGitopsCFDeploymentTestsUtils()

		err := runSapGitopsCFDeployment(&config, nil, &utils)
		assert.NoError(t, err)

		assert.True(t, utils.FilesMock.HasFile("/temp-test/path/to/mta.yaml"))
		assert.Equal(t, "kubectl", utils.ExecMockRunner.Calls[0].Exec)
		assert.Contains(t, utils.ExecMockRunner.Calls[0].Params, "--filename=temp-test/path/to/mta.yaml")
		assert.Contains(t, utils.ExecMockRunner.Calls[0].Params, `--patch=[{"op": "replace", "path": "/spec/forProvider/file/url", "value":"https://the.url/to/file.mtar"}]`)
		assert.Equal(t, config.Username, utils.pushUsername)
		assert.Equal(t, config.Password, utils.pushPassword)
		assert.Equal(t, config.BranchName, utils.cloneBranch)
		assert.Equal(t, config.ServerURL, utils.cloneServer)
		assert.Equal(t, config.FilePath, utils.addPaths[0])
	})

	t.Run("happy path - multiple urls", func(t *testing.T) {
		t.Parallel()

		config := sapGitopsCFDeploymentOptions{
			FilePath:   "path/to/mta.yaml",
			MtarURLs:   []string{"https://the.url/to/file.npm", "https://the.url/to/file.mtar"},
			ServerURL:  "https://the.server.url",
			BranchName: "theBranch",
			Username:   "theUser",
			Password:   "thePassword",
		}
		utils := newSapGitopsCFDeploymentTestsUtils()

		err := runSapGitopsCFDeployment(&config, nil, &utils)
		assert.NoError(t, err)

		assert.True(t, utils.FilesMock.HasFile("/temp-test/path/to/mta.yaml"))
		assert.Equal(t, "kubectl", utils.ExecMockRunner.Calls[0].Exec)
		assert.Contains(t, utils.ExecMockRunner.Calls[0].Params, "--filename=temp-test/path/to/mta.yaml")
		assert.Contains(t, utils.ExecMockRunner.Calls[0].Params, `--patch=[{"op": "replace", "path": "/spec/forProvider/file/url", "value":"https://the.url/to/file.mtar"}]`)
		assert.Equal(t, config.Username, utils.pushUsername)
		assert.Equal(t, config.Password, utils.pushPassword)
		assert.Equal(t, config.BranchName, utils.cloneBranch)
		assert.Equal(t, config.ServerURL, utils.cloneServer)
		assert.Equal(t, config.FilePath, utils.addPaths[0])
	})

	t.Run("error - clone repo", func(t *testing.T) {
		t.Parallel()

		config := sapGitopsCFDeploymentOptions{
			MtarURLs: []string{"https://the.url/to/file.mtar"},
		}
		utils := newSapGitopsCFDeploymentTestsUtils()
		utils.failOnClone = true

		err := runSapGitopsCFDeployment(&config, nil, &utils)
		assert.ErrorContains(t, err, "repository could not get prepared")
	})

	t.Run("error - no files found", func(t *testing.T) {
		t.Parallel()

		config := sapGitopsCFDeploymentOptions{
			FilePath: "wrongPath/to/mta.yaml",
			MtarURLs: []string{"https://the.url/to/file.mtar"},
		}
		utils := newSapGitopsCFDeploymentTestsUtils()

		err := runSapGitopsCFDeployment(&config, nil, &utils)
		assert.ErrorContains(t, err, "no matching files found for provided globbing pattern")
	})

	t.Run("error - no mta urls found", func(t *testing.T) {
		t.Parallel()

		config := sapGitopsCFDeploymentOptions{
			FilePath: "path/to/mta.yaml",
			MtarURLs: []string{"https://the.url/to/file.npm"},
		}
		utils := newSapGitopsCFDeploymentTestsUtils()

		err := runSapGitopsCFDeployment(&config, nil, &utils)
		assert.ErrorContains(t, err, "no mtarUrls available")
	})

	t.Run("error - kubectl", func(t *testing.T) {
		t.Parallel()

		config := sapGitopsCFDeploymentOptions{
			FilePath: "path/to/mta.yaml",
			MtarURLs: []string{"https://the.url/to/file.mtar"},
		}
		utils := newSapGitopsCFDeploymentTestsUtils()
		utils.ExecMockRunner.ShouldFailOnCommand = map[string]error{"kubectl": errors.New("kubectl error")}

		err := runSapGitopsCFDeployment(&config, nil, &utils)
		assert.ErrorContains(t, err, "error on kubectl execution")
	})

	t.Run("error - write file", func(t *testing.T) {
		t.Parallel()

		config := sapGitopsCFDeploymentOptions{
			FilePath: "path/to/mta.yaml",
			MtarURLs: []string{"https://the.url/to/file.mtar"},
		}
		utils := newSapGitopsCFDeploymentTestsUtils()
		utils.ExecMockRunner.StdoutReturn = map[string]string{"kubectl": "kubectlResult"}
		utils.FileWriteError = errors.New("failed")

		err := runSapGitopsCFDeployment(&config, nil, &utils)
		assert.ErrorContains(t, err, "failed to write file")
	})

	t.Run("error - commit", func(t *testing.T) {
		t.Parallel()

		config := sapGitopsCFDeploymentOptions{
			FilePath: "path/to/mta.yaml",
			MtarURLs: []string{"https://the.url/to/file.mtar"},
		}
		utils := newSapGitopsCFDeploymentTestsUtils()
		utils.failOnCommit = true

		err := runSapGitopsCFDeployment(&config, nil, &utils)
		assert.ErrorContains(t, err, "failed to commit and push changes")
	})
}

func TestRunKubeCtlCommand(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		utils := newSapGitopsCFDeploymentTestsUtils()
		utils.ExecMockRunner.StdoutReturn = map[string]string{"kubectl": "kubectlResult"}
		mtarUrl := "theMTAR.url"
		filePath := "theFilePath"

		expected := []string{
			"patch",
			"--local",
			"--output=yaml",
			"--patch=[{\"op\": \"replace\", \"path\": \"/spec/forProvider/file/url\", \"value\":\"theMTAR.url\"}]",
			"--filename=theFilePath",
			"--type=json",
		}

		b, err := runKubeCtlCommand(&utils, mtarUrl, filePath)
		assert.NoError(t, err)
		assert.Equal(t, "kubectl", utils.ExecMockRunner.Calls[0].Exec)
		assert.Equal(t, expected, utils.ExecMockRunner.Calls[0].Params)
		assert.Equal(t, []byte(string(utils.ExecMockRunner.StdoutReturn["kubectl"])), b)
	})

	t.Run("kubectl error", func(t *testing.T) {
		t.Parallel()

		utils := newSapGitopsCFDeploymentTestsUtils()
		utils.ExecMockRunner.ShouldFailOnCommand = map[string]error{"kubectl": errors.New("kubectl error")}

		mtarUrl := "theMTAR.url"
		filePath := "theFilePath"

		_, err := runKubeCtlCommand(&utils, mtarUrl, filePath)
		assert.ErrorContains(t, err, "failed to apply kubectl command")
	})
}

func TestDownloadCACertbunde(t *testing.T) {
	t.Parallel()

	t.Run("happy path - no certs", func(t *testing.T) {
		t.Parallel()

		customTlsCertificateLinks := []string{}
		utils := newSapGitopsCFDeploymentTestsUtils()

		b, err := downloadCACertbundle(customTlsCertificateLinks, &utils)
		assert.NoError(t, err)
		assert.Equal(t, []byte{}, b)
	})

	t.Run("happy path - certs", func(t *testing.T) {
		t.Parallel()

		customTlsCertificateLinks := []string{"https://link/to/cert"}
		utils := newSapGitopsCFDeploymentTestsUtils()
		utils.HttpClientMock = &mock.HttpClientMock{
			HTTPFileUtils: utils.FilesMock,
		}

		b, err := downloadCACertbundle(customTlsCertificateLinks, &utils)
		assert.NoError(t, err)
		assert.Equal(t, []byte(string("some content")), b)
	})
}

func TestCommitFiles(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		utils := newSapGitopsCFDeploymentTestsUtils()
		utils.AddFile("file1.txt", []byte("file1 content"))
		utils.AddFile("file2.txt", []byte("file2 content"))
		filePaths := []string{"file1.txt", "file2.txt"}
		commitMessage := "the message"
		author := "the author"

		h, err := commitFiles(&utils, filePaths, commitMessage, author)
		assert.NoError(t, err)
		assert.Equal(t, plumbing.Hash{123}, h)
		assert.Equal(t, filePaths, utils.addPaths)
		assert.Equal(t, commitMessage, utils.commitMessage)
		assert.Equal(t, true, utils.commitOpts.All)
		assert.Equal(t, author, utils.commitOpts.Author.Name)
	})

	t.Run("add error", func(t *testing.T) {
		t.Parallel()
		utils := newSapGitopsCFDeploymentTestsUtils()
		utils.AddFile("file1.txt", []byte("file1 content"))
		utils.AddFile("file2.txt", []byte("file2 content"))
		utils.failOnAdd = true
		filePaths := []string{"file1.txt", "file2.txt"}
		commitMessage := "the message"
		author := "the author"

		_, err := commitFiles(&utils, filePaths, commitMessage, author)
		assert.ErrorContains(t, err, "failed to add file to git")
	})

	t.Run("commit error", func(t *testing.T) {
		t.Parallel()
		utils := newSapGitopsCFDeploymentTestsUtils()
		utils.AddFile("file1.txt", []byte("file1 content"))
		utils.AddFile("file2.txt", []byte("file2 content"))
		utils.failOnCommit = true
		filePaths := []string{"file1.txt", "file2.txt"}
		commitMessage := "the message"
		author := "the author"

		_, err := commitFiles(&utils, filePaths, commitMessage, author)
		assert.ErrorContains(t, err, "failed to commit file")
	})
}

func TestCommitAndPushChanges(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		config := sapGitopsCFDeploymentOptions{
			CommitMessage: "the message",
			Username:      "theUser",
			Password:      "thePassword",
			ForcePush:     true,
		}

		utils := newSapGitopsCFDeploymentTestsUtils()

		utils.AddFile("file1.txt", []byte("file content"))
		filePaths := []string{"file1.txt"}

		certs := []byte{123}

		h, err := commitAndPushChanges(&config, &utils, filePaths, certs)
		assert.NoError(t, err)
		assert.Equal(t, plumbing.Hash{123}, h)

		assert.Equal(t, filePaths, utils.addPaths)
		assert.Equal(t, config.CommitMessage, utils.commitMessage)
		assert.Equal(t, config.Username, utils.commitOpts.Author.Name)

		assert.Equal(t, config.Username, utils.pushUsername)
		assert.Equal(t, config.Password, utils.pushPassword)
		assert.Equal(t, utils.forcePush, config.ForcePush)
		assert.Equal(t, certs, utils.pushCACerts)
	})

	t.Run("commit error", func(t *testing.T) {
		t.Parallel()
		config := sapGitopsCFDeploymentOptions{}

		utils := newSapGitopsCFDeploymentTestsUtils()
		utils.failOnCommit = true

		filePaths := []string{"file1.txt"}
		certs := []byte{123}

		_, err := commitAndPushChanges(&config, &utils, filePaths, certs)
		assert.ErrorContains(t, err, "committing changes failed")
	})

	t.Run("push error", func(t *testing.T) {
		t.Parallel()
		config := sapGitopsCFDeploymentOptions{}

		utils := newSapGitopsCFDeploymentTestsUtils()
		utils.failOnPush = true
		utils.AddFile("file1.txt", []byte("file content"))

		filePaths := []string{"file1.txt"}
		certs := []byte{123}

		_, err := commitAndPushChanges(&config, &utils, filePaths, certs)
		assert.ErrorContains(t, err, "pushing changes failed")
	})
}

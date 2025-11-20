package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"

	gitUtil "github.com/SAP/jenkins-library/pkg/git"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const toolKubectl = "kubectl"

type sapGitopsCFDeploymentUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)
	FileRead(path string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	Glob(pattern string) (matches []string, err error)
	RemoveAll(path string) error
	TempDir(dir string, pattern string) (name string, err error)

	DownloadFile(url string, filename string, header http.Header, cookies []*http.Cookie) error

	Add(path string) (plumbing.Hash, error)
	Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error)
	ChangeBranch(branchName string) error
	PlainCloneAndWorktree(username, password, serverURL, branchName, directory string, caCerts []byte) error
	PushChangesToRepository(username, password string, force *bool, caCerts []byte) error
}

type sapGitopsCFDeploymentUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*piperhttp.Client

	worktree   *git.Worktree
	repository *git.Repository
}

func (s *sapGitopsCFDeploymentUtilsBundle) Add(path string) (plumbing.Hash, error) {
	return s.worktree.Add(path)
}

func (s *sapGitopsCFDeploymentUtilsBundle) Commit(msg string, opts *git.CommitOptions) (plumbing.Hash, error) {
	return s.worktree.Commit(msg, opts)
}

func (s *sapGitopsCFDeploymentUtilsBundle) ChangeBranch(branchName string) error {
	return gitUtil.ChangeBranch(branchName, s.worktree)
}

func (s *sapGitopsCFDeploymentUtilsBundle) PlainCloneAndWorktree(username, password, serverURL, branchName, directory string, caCerts []byte) error {
	var err error
	s.repository, err = gitUtil.PlainClone(username, password, serverURL, branchName, directory, caCerts)
	if err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	s.worktree, err = s.repository.Worktree()
	if err != nil {
		return fmt.Errorf("failed to retrieve worktree: %w", err)
	}
	return nil
}

func (s *sapGitopsCFDeploymentUtilsBundle) PushChangesToRepository(username, password string, force *bool, caCerts []byte) error {
	return gitUtil.PushChangesToRepository(username, password, force, s.repository, caCerts)
}

func newSapGitopsCFDeploymentUtils() *sapGitopsCFDeploymentUtilsBundle {
	utils := sapGitopsCFDeploymentUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func sapGitopsCFDeployment(config sapGitopsCFDeploymentOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newSapGitopsCFDeploymentUtils()

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runSapGitopsCFDeployment(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

// func runSapGitopsCFDeployment(config *sapGitopsCFDeploymentOptions, telemetryData *telemetry.CustomData, utils sapGitopsCFDeploymentUtils) error {
func runSapGitopsCFDeployment(config *sapGitopsCFDeploymentOptions, _ *telemetry.CustomData, utils sapGitopsCFDeploymentUtils) error {
	temporaryFolder, err := utils.TempDir(".", "temp-")
	temporaryFolder = regexp.MustCompile(`^./`).ReplaceAllString(temporaryFolder, "")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}

	defer func() {
		err = utils.RemoveAll(temporaryFolder)
		if err != nil {
			log.Entry().WithError(err).Error("error during temporary directory deletion")
		}
	}()

	// download does not create error currently
	certs, _ := downloadCACertbundle(config.CustomTLSCertificateLinks, utils)

	err = utils.PlainCloneAndWorktree(config.Username, config.Password, config.ServerURL, config.BranchName, temporaryFolder, certs)
	if err != nil {
		return fmt.Errorf("repository could not get prepared: %w", err)
	}

	filePath := filepath.Join(temporaryFolder, config.FilePath)

	allFiles, err := utils.Glob(filePath)
	if err != nil {
		return fmt.Errorf("unable to expand globbing pattern: %w", err)
	} else if len(allFiles) == 0 {
		return fmt.Errorf("no matching files found for provided globbing pattern: %w", err)
	}
	utils.SetDir("./")

	// identify mtar from promoted artifacts list
	mtarURL := ""
	for _, artifactUrl := range config.MtarURLs {
		if strings.HasSuffix(artifactUrl, ".mtar") {
			mtarURL = artifactUrl
		}
	}

	if len(mtarURL) == 0 {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("no mtarUrls available")
	}

	var outputBytes []byte
	for _, currentFile := range allFiles {
		outputBytes, err = runKubeCtlCommand(utils, mtarURL, filePath)
		if err != nil {
			return fmt.Errorf("error on kubectl execution: %w", err)
		}

		if outputBytes != nil {
			err = utils.FileWrite(currentFile, outputBytes, 0o755)
			if err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}
		}
	}

	// git expects the file path relative to its root:
	for i := range allFiles {
		allFiles[i] = strings.ReplaceAll(allFiles[i], temporaryFolder+"/", "")
	}

	commit, err := commitAndPushChanges(config, utils, allFiles, certs)
	if err != nil {
		return fmt.Errorf("failed to commit and push changes: %w", err)
	}

	log.Entry().Infof("Changes committed with %s", commit.String())

	return nil
}

func runKubeCtlCommand(command sapGitopsCFDeploymentUtils, mtarUrl string, filePath string) ([]byte, error) {
	kubectlOutput := bytes.Buffer{}
	command.Stdout(&kubectlOutput)

	kubeParams := []string{
		"patch",
		"--local",
		"--output=yaml",
		"--patch=" + fmt.Sprintf(`[{"op": "replace", "path": "/spec/forProvider/file/url", "value":"%v"}]`, mtarUrl),
		"--filename=" + filePath,
		"--type=json",
	}

	err := command.RunExecutable(toolKubectl, kubeParams...)
	if err != nil {
		return nil, fmt.Errorf("failed to apply kubectl command: %w", err)
	}
	return kubectlOutput.Bytes(), nil
}

func downloadCACertbundle(customTlsCertificateLinks []string, utils sapGitopsCFDeploymentUtils) ([]byte, error) {
	certs := []byte{}
	if len(customTlsCertificateLinks) > 0 {
		for _, customTlsCertificateLink := range customTlsCertificateLinks {
			log.Entry().Infof("Downloading CA certs %s into file '%s'", customTlsCertificateLink, path.Base(customTlsCertificateLink))
			err := utils.DownloadFile(customTlsCertificateLink, path.Base(customTlsCertificateLink), nil, nil)
			if err != nil {
				return certs, nil
			}

			content, err := utils.FileRead(path.Base(customTlsCertificateLink))
			if err != nil {
				return certs, nil
			}
			log.Entry().Infof("CA certs added successfully to cert pool")

			certs = append(certs, content...)
		}
	}

	return certs, nil
}

func commitAndPushChanges(config *sapGitopsCFDeploymentOptions, utils sapGitopsCFDeploymentUtils, filePaths []string, certs []byte) (plumbing.Hash, error) {
	commitMessage := config.CommitMessage

	commit, err := commitFiles(utils, filePaths, commitMessage, config.Username)
	if err != nil {
		return [20]byte{}, fmt.Errorf("committing changes failed: %w", err)
	}

	err = utils.PushChangesToRepository(config.Username, config.Password, &config.ForcePush, certs)
	if err != nil {
		return [20]byte{}, fmt.Errorf("pushing changes failed: %w", err)
	}

	return commit, nil
}

func commitFiles(s sapGitopsCFDeploymentUtils, filePaths []string, commitMessage, author string) (plumbing.Hash, error) {
	for _, path := range filePaths {
		_, err := s.Add(path)
		if err != nil {
			return plumbing.Hash{}, fmt.Errorf("failed to add file to git: %w", err)
		}
	}

	commit, err := s.Commit(commitMessage, &git.CommitOptions{
		All:    true,
		Author: &object.Signature{Name: author, When: time.Now()},
	})
	if err != nil {
		return plumbing.Hash{}, fmt.Errorf("failed to commit file: %w", err)
	}

	return commit, nil
}

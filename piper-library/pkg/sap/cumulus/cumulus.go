package cumulus

import (
	"fmt"
	"os"
	"strings"

	gitUtils "github.com/SAP/jenkins-library/pkg/git"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
)

// Cumulus defines the information required for connections to Cumulus system
type Cumulus struct {
	EnvVars               []EnvVar
	FilePattern           string
	PipelineID            string
	Revision              string
	Scheduled             bool
	SchedulingTimestamp   string
	StepResultType        string
	SubFolderPath         string
	Version               string
	TargetPath            string
	HeadCommitID          string
	UseCommitIDForCumulus bool
	BucketLockedWarning   bool
	OpenGitFunc           func() (GitRepository, error)
	EventData             string
}

// EnvVar defines an  environment variable incl. information about a potential modification to the variable
type EnvVar struct {
	Name     string
	Value    string
	Modified bool
}

type GitRepository interface {
	ResolveRevision(plumbing.Revision) (*plumbing.Hash, error)
}

// PrepareEnv sets required environment variables in case they are not set yet
func (c *Cumulus) PrepareEnv() {
	for key, env := range c.EnvVars {
		c.EnvVars[key].Modified = setenvIfEmpty(env.Name, env.Value)
	}
}

// CleanupEnv removes environment variables set by PrepareEnv
func (c *Cumulus) CleanupEnv() {
	for _, env := range c.EnvVars {
		removeEnvIfPreviouslySet(env.Name, env.Modified)
	}
}

// ValidateInput validates that all parameters are set correctly
func (c *Cumulus) ValidateInput() error {
	if err := c.validateStringNotEmpty(c.PipelineID, "pipelineID"); err != nil {
		return err
	}
	if !c.UseCommitIDForCumulus {
		if err := c.validateStringNotEmpty(c.Version, "version"); err != nil {
			return err
		}
	}

	if err := c.validateStringNotEmpty(c.StepResultType, "stepResultType"); err != nil {
		return err
	}
	return nil
}

func (c *Cumulus) validateStringNotEmpty(value string, name string) error {
	if value == "" {
		err := errors.New(name + " must not be empty")
		return err
	}
	return nil
}

func (c *Cumulus) openGit() (GitRepository, error) {
	// used for test purposes only in sapCumulusUpload and sapCumulusDownload tests
	if c.OpenGitFunc != nil {
		return c.OpenGitFunc()
	}
	workdir, _ := os.Getwd()
	return gitUtils.PlainOpen(workdir)
}

func (c *Cumulus) GetCumulusPath(pipelineRunKey string) string {
	targetPath := fmt.Sprintf("%s/", pipelineRunKey)
	if c.StepResultType != "root" {
		targetPath += c.StepResultType
	}

	targetPath = c.handleSubFolderPath(targetPath)
	return targetPath
}

func (c *Cumulus) handleSubFolderPath(targetPath string) string {
	// check if subfolder path is used and if so, normalize it
	if len(c.SubFolderPath) > 0 {

		subFolderPath := c.SubFolderPath
		lastIndex := len(subFolderPath) - 1

		if strings.LastIndex(subFolderPath, "/") == lastIndex {
			subFolderPath = string([]rune(subFolderPath)[:lastIndex])
		}
		targetPath += fmt.Sprintf("/%v", subFolderPath)
	}
	return targetPath
}

func setenvIfEmpty(env, val string) bool {
	if len(os.Getenv(env)) == 0 {
		os.Setenv(env, val)
		return true
	}
	return false
}

func removeEnvIfPreviouslySet(env string, previouslySet bool) {
	if previouslySet {
		os.Setenv(env, "")
	}
}

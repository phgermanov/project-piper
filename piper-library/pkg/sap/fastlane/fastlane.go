package fastlane

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
)

// EnvVar defines an  environment variable incl. information about a potential modification to the variable
type EnvVar struct {
	Name     string
	Value    string
	Modified bool
}

type Utils interface {
	ConfigureGitAuthentication(username string, password string) error
	InstallDependencies() error
	ExecuteFastlaneCommand(command string) error
	AddToExecEnv(e []string)
}

type UtilsBundle struct {
	piperutils.FileUtils
	ExecRunner ExecRunner
}

type ExecRunner interface {
	AppendEnv(env []string)
	Stdout(out io.Writer)
	Stderr(out io.Writer)
	RunExecutable(executable string, params ...string) error
}

func NewUtilsBundle() UtilsBundle {
	u := UtilsBundle{
		FileUtils:  &piperutils.Files{},
		ExecRunner: &command.Command{},
	}

	u.ExecRunner.Stdout(log.Writer())
	u.ExecRunner.Stderr(log.Writer())
	return u
}

// If passwords have to be passed as parameter
func (u *UtilsBundle) disableStdout() {
	u.ExecRunner.Stdout(io.Discard)
}

func (u *UtilsBundle) enableStdout() {
	u.ExecRunner.Stdout(log.Writer())
}

// InstallDependencies uses Bundler to install the dependencies listed in your Gemfile.lock
func (u *UtilsBundle) InstallDependencies() error {
	bundlerVersion, err := u.GetBundlerVersion()

	if err == nil {
		// Try to install specific version of Bundler
		err = u.ExecRunner.RunExecutable("gem", "install", "bundler", "-v", bundlerVersion, "--force", "--no-document")
		if err != nil {
			return errors.Wrapf(err, "Failed to install the version of Bundler specified in your Gemfile.lock")
		}
	} else {
		log.Entry().Warnf("Could not find a Gemfile.lock in the current workspace. Please add it to your VCS.")
		// Fallback: check if at least one version of Bundler is already installed -> will throw a warning but most likely work as well
		_, err = ExecLookPath("bundle")
		if err != nil {
			return errors.Wrapf(err, "It seems that Bundler is not (properly) installed on this machine. Cannot install dependencies")
		}
	}

	err = u.ExecRunner.RunExecutable("bundle", "install")
	if err != nil {
		return errors.Wrapf(err, "Failed to install dependencies")
	}

	return nil
}

func (u *UtilsBundle) ConfigureGitAuthentication(username string, password string) error {
	u.disableStdout()
	auth := fmt.Sprintf("%s:%s", username, password)
	err := u.ExecRunner.RunExecutable("bundle", "config", "--local", "github.tools.sap", auth)
	if err != nil {
		u.enableStdout()
		return errors.Wrapf(err, "Failed to configure git authentication for bundler on github.tools.sap")
	}
	u.enableStdout()
	return nil
}

func (u *UtilsBundle) ExecuteFastlaneCommand(command string) error {
	// Split command and parameters
	fastlaneCommand := strings.Split(command, " ")

	parameters := append([]string{"exec", "fastlane"}, fastlaneCommand...)

	// bundle exec fastlane command
	err := u.ExecRunner.RunExecutable("bundle", parameters...)
	if err != nil {
		return errors.Wrapf(err, "Failed to execute fastlane")
	}

	return nil
}

// Exposes execRunner.AppendEnv
func (u *UtilsBundle) AddToExecEnv(e []string) {
	u.ExecRunner.AppendEnv(e)
}

// Parses the Gemfile.lock and returns the version of Bundler which has been used to generate the file
func (u *UtilsBundle) GetBundlerVersion() (string, error) {
	// Gemfile.lock should be at project-root. It contains the following to lines to specify, which version of bundler has been used:
	// BUNDLED WITH
	//    x.y.z

	gemfileLockPath := "Gemfile.lock"
	if exists, _ := u.FileExists(gemfileLockPath); !exists {
		return "", errors.New("Could not find a Gemfile.lock in the current path. It's good practice to add it to your VCS.")
	}

	gemfileLockBytes, err := u.FileRead(gemfileLockPath)
	if err != nil {
		return "", errors.Wrapf(err, "Could not open Gemfile.lock")
	}

	gemfileLockString := string(gemfileLockBytes)
	tmp := strings.Split(gemfileLockString, "\n")

	lineNumber := 0
	found := false
	for _, line := range tmp {
		if strings.Contains(line, "BUNDLED WITH") {
			found = true
			break
		}
		lineNumber++
	}

	if found {
		line := tmp[lineNumber+1]                 // bundler version is one line below "BUNDLED WITH"
		bundlerVersion := strings.TrimSpace(line) // remove leading spaces
		return bundlerVersion, nil
	} else {
		return "", errors.New("Internal error: Could not find bundler version in Gemfile.lock")
	}
}

var (
	ExecLookPath = exec.LookPath
)

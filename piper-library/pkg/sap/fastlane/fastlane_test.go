//go:build unit
// +build unit

package fastlane_test

import (
	"os/exec"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/fastlane"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/fastlane/mocks"
)

func mockExecLookPath(executable string) (string, error) {
	if shouldLookPathSucceed {
		return "/usr/local/bin/bundle", nil
	}
	return "", errors.New(executable + " is not installed")
}

func TestBundlerUtils(t *testing.T) {
	t.Parallel()
	t.Run("Get bundler version - happy path", func(t *testing.T) {
		t.Parallel()
		// init
		utils := mocks.NewFastlaneMockUtilsBundle()

		gemfileLock := `GIT
	  GEM
		remote: https://rubygems.org/
		specs:
		  CFPropertyList (3.0.3)

	  PLATFORMS
		x86_64-darwin-20

	  DEPENDENCIES
		fastlane

	  BUNDLED WITH
		 2.2.16

		`
		utils.FileUtils.FileWrite("Gemfile.lock", []byte(gemfileLock), 0666)

		// test
		bundleVersion, _ := utils.GetBundlerVersion()

		// assert
		assert.EqualValues(t, bundleVersion, "2.2.16")
	})

	t.Run("No bundler version present in Gemfile.lock", func(t *testing.T) {
		t.Parallel()
		// init
		utils := mocks.NewFastlaneMockUtilsBundle()

		gemfileLock := " \n"
		utils.FileUtils.FileWrite("Gemfile.lock", []byte(gemfileLock), 0666)

		// test
		_, err := utils.GetBundlerVersion()

		// assert
		assert.EqualError(t, err, "Internal error: Could not find bundler version in Gemfile.lock")
	})

	t.Run("No Gemfile.lock available", func(t *testing.T) {
		t.Parallel()
		// init
		utils := mocks.NewFastlaneMockUtilsBundle()

		// test
		_, err := utils.GetBundlerVersion()

		// assert
		assert.EqualError(t, err, "Could not find a Gemfile.lock in the current path. It's good practice to add it to your VCS.")
	})
}

func TestInstallDependencies(t *testing.T) {
	t.Parallel()

	t.Run("Happy Path", func(t *testing.T) {
		// init

		// Don't use mocks.NewFastlaneMockUtilsBundle() on purpose -> access execRunner later
		execRunner := mock.ExecMockRunner{}
		utils := fastlane.UtilsBundle{
			FileUtils:  &mock.FilesMock{},
			ExecRunner: &execRunner,
		}

		gemfileLock := "\n BUNDLED WITH\n    2.2.16\n"
		utils.FileUtils.FileWrite("Gemfile.lock", []byte(gemfileLock), 0666)

		// test
		err := utils.InstallDependencies()

		// assert
		assert.Nil(t, err)
		assert.Len(t, execRunner.Calls, 2)
		assert.Equal(t, mock.ExecCall{Exec: "gem", Params: []string{"install", "bundler", "-v", "2.2.16", "--force", "--no-document"}}, execRunner.Calls[0])
		assert.Equal(t, mock.ExecCall{Exec: "bundle", Params: []string{"install"}}, execRunner.Calls[1])
	})

	t.Run("Gem install failed", func(t *testing.T) {
		// init

		// Don't use mocks.NewFastlaneMockUtilsBundle() on purpose -> access execRunner later
		execRunner := mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"gem": errors.New("Haha!")}}
		utils := fastlane.UtilsBundle{
			FileUtils:  &mock.FilesMock{},
			ExecRunner: &execRunner,
		}

		gemfileLock := "\n BUNDLED WITH\n    2.2.16\n"
		utils.FileUtils.FileWrite("Gemfile.lock", []byte(gemfileLock), 0666)

		// test
		err := utils.InstallDependencies()

		// assert
		assert.Equal(t, "Failed to install the version of Bundler specified in your Gemfile.lock: Haha!", err.Error())
	})

	t.Run("Bundler not installed", func(t *testing.T) {
		// init

		utils := mocks.NewFastlaneMockUtilsBundle()
		shouldLookPathSucceed = false
		fastlane.ExecLookPath = mockExecLookPath
		defer func() { fastlane.ExecLookPath = exec.LookPath }()

		gemfileLock := " \n" // Invalid Gemfile to provoke fallback
		utils.FileUtils.FileWrite("Gemfile.lock", []byte(gemfileLock), 0666)

		// test
		err := utils.InstallDependencies()
		// assert
		assert.EqualError(t, err, "It seems that Bundler is not (properly) installed on this machine. Cannot install dependencies: bundle is not installed")
	})

	t.Run("No Gemfile.lock but bundler installed", func(t *testing.T) {
		// init

		utils := mocks.NewFastlaneMockUtilsBundle()
		shouldLookPathSucceed = true
		fastlane.ExecLookPath = mockExecLookPath
		defer func() { fastlane.ExecLookPath = exec.LookPath }()

		gemfileLock := " \n" // Invalid Gemfile to provoke fallback
		utils.FileUtils.FileWrite("Gemfile.lock", []byte(gemfileLock), 0666)

		// test
		err := utils.InstallDependencies()
		// assert
		assert.Nil(t, err)
	})
}

func TestFastlaneCalls(t *testing.T) {
	t.Parallel()
	t.Run("Assert correct command split", func(t *testing.T) {
		// init

		// Don't use mocks.NewFastlaneMockUtilsBundle() on purpose -> access execRunner later
		execRunner := mock.ExecMockRunner{}
		utils := fastlane.UtilsBundle{
			FileUtils:  &mock.FilesMock{},
			ExecRunner: &execRunner,
		}

		// test
		err := utils.ExecuteFastlaneCommand("ios enterprise --verbose true")

		// assert
		assert.Nil(t, err)
		assert.Len(t, execRunner.Calls, 1)
		assert.Equal(t, mock.ExecCall{Exec: "bundle", Params: []string{"exec", "fastlane", "ios", "enterprise", "--verbose", "true"}}, execRunner.Calls[0])
	})
}

func TestBundlerGitAuth(t *testing.T) {
	t.Parallel()
	t.Run("Check encoding", func(t *testing.T) {
		// init

		// Don't use mocks.NewFastlaneMockUtilsBundle() on purpose -> access execRunner later
		execRunner := mock.ExecMockRunner{}
		utils := fastlane.UtilsBundle{
			FileUtils:  &mock.FilesMock{},
			ExecRunner: &execRunner,
		}

		// test
		err := utils.ConfigureGitAuthentication("RickAstley", "https://www.youtube.com/watch?v=dQw4w9WgXcQ")

		// assert
		assert.Nil(t, err)
		assert.Len(t, execRunner.Calls, 1)
		assert.Equal(t, mock.ExecCall{Exec: "bundle", Params: []string{"config", "--local", "github.tools.sap", "RickAstley:https://www.youtube.com/watch?v=dQw4w9WgXcQ"}}, execRunner.Calls[0])
	})

	t.Run("git auth failed", func(t *testing.T) {
		// init

		execRunner := mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"bundle": errors.New("Haha!")}}
		utils := fastlane.UtilsBundle{
			FileUtils:  &mock.FilesMock{},
			ExecRunner: &execRunner,
		}

		// test
		err := utils.ConfigureGitAuthentication("", "")

		// assert
		assert.EqualError(t, err, "Failed to configure git authentication for bundler on github.tools.sap: Haha!")
	})
}

var (
	shouldLookPathSucceed bool
)

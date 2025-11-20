//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/fastlane"
)

// Helper method to avoid code duplication
func testMatchConfig(t *testing.T, options sapExecuteFastlaneOptions, matchPassword, matchCredentials string, shouldRunMatch bool) {
	// Don't use mocks.NewFastlaneMockUtilsBundle() on purpose -> access env in assert later
	execRunner := mock.ExecMockRunner{}
	utils := fastlane.UtilsBundle{
		FileUtils:  &mock.FilesMock{},
		ExecRunner: &execRunner,
	}

	gemfileLock := `Pogo
	BUNDLED WITH
	   x.y.z

	  `
	utils.FileUtils.FileWrite("Gemfile.lock", []byte(gemfileLock), 0666)

	// test
	err := runSapExecuteFastlane(&utils, &options)

	// assert
	assert.Nil(t, err)
	if shouldRunMatch {
		assert.Contains(t, execRunner.Env, "MATCH_PASSWORD="+matchPassword)
		assert.Contains(t, execRunner.Env, "MATCH_GIT_BASIC_AUTHORIZATION="+matchCredentials)
	} else {
		assert.NotContains(t, execRunner.Env, "MATCH_PASSWORD")
		assert.NotContains(t, execRunner.Env, "MATCH_GIT_BASIC_AUTHORIZATION")
	}
}

func TestRunSapExecuteFastlane(t *testing.T) {
	t.Parallel()

	t.Run("Happy path - With iOS Signing", func(t *testing.T) {
		t.Parallel()
		// init
		options := sapExecuteFastlaneOptions{
			LaneName:                           "enterprise",
			Platform:                           "ios",
			SetupSigning:                       true,
			FastlaneMatchPassphrase:            "FooBar42",
			FastlaneMatchRepositoryCredentials: "Umlja0FzdGxleTpodHRwczovL3d3dy55b3V0dWJlLmNvbS93YXRjaD92PWRRdzR3OVdnWGNR",
		}

		testMatchConfig(t, options, options.FastlaneMatchPassphrase, options.FastlaneMatchRepositoryCredentials, true)
	})

	t.Run("Non-iOS platform", func(t *testing.T) {
		t.Parallel()
		// init
		options := sapExecuteFastlaneOptions{
			LaneName:                           "company",
			Platform:                           "android",
			FastlaneMatchPassphrase:            "FooBar42",
			FastlaneMatchRepositoryCredentials: "Umlja0FzdGxleTpodHRwczovL3d3dy55b3V0dWJlLmNvbS93YXRjaD92PWRRdzR3OVdnWGNR",
		}

		testMatchConfig(t, options, options.FastlaneMatchPassphrase, options.FastlaneMatchRepositoryCredentials, false)
	})
}

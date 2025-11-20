package cmd

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/fastlane"
)

func sapExecuteFastlane(options sapExecuteFastlaneOptions, telemetryData *telemetry.CustomData) {
	utils := fastlane.NewUtilsBundle()

	if err := runSapExecuteFastlane(&utils, &options); err != nil {
		log.Entry().WithError(err).Fatal("Failed to execute fastlane.")
	}
}

func runSapExecuteFastlane(utils fastlane.Utils, options *sapExecuteFastlaneOptions) error {
	var err error

	if options.SetupBundlerAuthentication && options.Username != "" && options.Password != "" {
		if err = utils.ConfigureGitAuthentication(options.Username, options.Password); err != nil {
			return err
		}
	}

	// Run bundle install
	if err = utils.InstallDependencies(); err != nil {
		return err
	}

	// Check whether fastlane match is configured
	if options.Platform == "ios" || options.Platform == "mac" {
		configureMatch(utils, options)
	}

	err = utils.ExecuteFastlaneCommand(options.Platform + " " + options.LaneName)
	return err // nil or err if ExecuteFastlaneCommand failed
}

// Checks whether the match passphrase is set (in vault) for the given lane name and adds it to the environment
func configureMatch(utils fastlane.Utils, options *sapExecuteFastlaneOptions) {

	if options.SetupSigning {
		if options.FastlaneMatchPassphrase != "" {
			utils.AddToExecEnv([]string{"MATCH_PASSWORD=" + options.FastlaneMatchPassphrase})
		}
		if options.FastlaneMatchRepositoryCredentials != "" {
			utils.AddToExecEnv([]string{"MATCH_GIT_BASIC_AUTHORIZATION=" + options.FastlaneMatchRepositoryCredentials})
		}
	}
}

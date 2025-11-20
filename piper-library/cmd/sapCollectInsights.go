package cmd

import (
	"net/url"
	"os"
	"path/filepath"

	piperOsCmd "github.com/SAP/jenkins-library/cmd"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/insights/binary"
	"gopkg.in/yaml.v3"
)

const (
	LinkSnow     = "https://itsm.services.sap/sp?id=sc_cat_item&sys_id=703f22d51b3b441020c8fddacd4bcbe2&service_offering=6af220d1dbafb950034ca8ebd3961907&public=yes&description=Please%20describe%20the%20issue%20and%20provide%20a%20logfile,%20link%20to%20your%20repo%20and%20pipeline%20definition."
	LinkDoraDocu = "https://pages.github.tools.sap/hyperspace/DevOps-Insights/01_dora/overview/"
)

type sapCollectInsightsUtils interface {
	FileWrite(path string, content []byte, perm os.FileMode) error
	InstallDevOpsInsightsCli(binPath, version string) error
	ExecuteDevOpInsights(binPath string, command []string) error
	TempDir(string, string) (string, error)
	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The sapCollectInsightsUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type sapCollectInsightsUtilsBundle struct {
	*command.Command
	piperutils.FileUtils
	binary.CLIBinary
	// Embed more structs as necessary to implement methods or interfaces you add to sapCollectInsightsUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// sapCollectInsightsUtilsBundle and forward to the implementation of the dependency.
}

func newSapCollectInsightsUtils(config sapCollectInsightsOptions) (sapCollectInsightsUtils, error) {
	var err error
	var resolver binary.CliReleaseResolver
	switch orchestrator.DetectOrchestrator() {
	case orchestrator.AzureDevOps:
		resolver, err = binary.NewGithubToolsReleaseResolver(config.GithubToken)
	default:
		resolver, err = binary.NewArtifactoryReleaseResolver()
	}
	if err != nil {
		return nil, err
	}
	utils := sapCollectInsightsUtilsBundle{
		Command:   &command.Command{},
		FileUtils: &piperutils.Files{},
		CLIBinary: binary.CLIBinary{
			FileDownloader:     &piperhttp.Client{},
			PermEditor:         &piperutils.Files{},
			CLIReleaseResolver: resolver,
			CLIExecutor:        binary.DefaultCLICommandExecutor,
			ExecFactory:        binary.DefaultExecutorFactory{},
		},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils, nil
}

func sapCollectInsights(config sapCollectInsightsOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils, err := newSapCollectInsightsUtils(config)
	if err != nil {
		log.Entry().WithError(err).Fatal("unable to initialize utils")
	}

	err = runSapCollectInsights(&config, telemetryData, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func writeConfig(utils sapCollectInsightsUtils, config sapCollectInsightsOptions, fileName string) error {
	// Write the provided sapCollectInsights config to disk
	// The devops-insights binary will pick up the config and load the parameters accordingly

	if config.GitURL != "" {
		log.Entry().Warningf("Deprecated parameter GitURL, please remove it from your step. Parameter not needed.")

		// We intentionally set the GitURL to an empty string, as for teams using SSH the token is part of the URL
		// We do not want to get this data.
		config.GitURL = ""
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		log.Entry().Fatalf("marshalling config parameters to file: %v", err)
		return err
	}

	err = utils.FileWrite(fileName, data, 0755)
	if err != nil {
		log.Entry().Fatalf("error opening/creating file: %v", err)
		return err
	}
	return nil
}

func runSapCollectInsights(config *sapCollectInsightsOptions, telemetryData *telemetry.CustomData, utils sapCollectInsightsUtils) error {
	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	if config.DevOpsInsightsToken != "" || config.DoraSystemTrustToken != "" {
		tmpFolder, err := utils.TempDir(".", "temp-")

		if err != nil {
			log.Entry().WithError(err).WithField("path", tmpFolder).Debug("Creating temp directory failed")
		}
		devOpsInsightsBinName := filepath.Join(tmpFolder, "devops-insights")
		// download devops-insights binary
		if err := utils.InstallDevOpsInsightsCli(devOpsInsightsBinName, config.DevOpsInsightsVersion); err != nil {
			log.Entry().WithError(err).Fatal("unable to install DevOpsInsights CLI: %w", err)
		}

		configFileName := filepath.Join(tmpFolder, "insights.yaml")
		err = writeConfig(utils, *config, configFileName)
		if err != nil {
			log.SetErrorCategory(log.ErrorInfrastructure)
			log.Entry().WithError(err).Fatal("failed to write sapCollectInsights config to file: %w", err)
		}
		// Collecting the commands to execute
		executeCommand := []string{"--config", configFileName}

		if piperOsCmd.GeneralConfig.Verbose {
			executeCommand = append(executeCommand, "--verbose", "true")
		}
		if err := utils.ExecuteDevOpInsights(devOpsInsightsBinName, executeCommand); err != nil {
			log.SetErrorCategory(log.ErrorService)
			log.Entry().WithError(err).Fatal("failed to execute devops-insights command: %w", err)
		}

	} else {
		stringifyLinkDORACoP, err := url.QueryUnescape(LinkSnow)
		if err != nil {
			log.Entry().Errorf("Error unescaping link: %s", err)
		}
		log.Entry().Errorf("\n------------------------------------------------------------------------\n"+
			"The DevOpsInsightsToken is not provided. Please add your token properly.\n"+
			"Please check out the docs here: %s \n"+
			"In case you dont know and tried reading the logs and the docs, open an issue here: \n"+
			"%s \n"+
			"------------------------------------------------------------------------", LinkDoraDocu, stringifyLinkDORACoP)
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.New("The DevOpsInsightsToken is not provided. Please add your token properly. READ THE LOG FILE ABOVE!")
	}
	return nil
}

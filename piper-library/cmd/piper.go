package cmd

import (
	"os"

	piperOsCmd "github.com/SAP/jenkins-library/cmd"
	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "piper",
	Short: "Executes CI/CD steps from project 'Piper' ",
	Long: `
This project 'Piper' binary provides a CI/CD step library.
It contains many steps which can be used within CI/CD systems as well as directly on e.g. a developer's machine.
`,
	// ToDo: respect stageName to also come from parametersJSON -> first env.STAGE_NAME, second: parametersJSON, third: flag
}

// TODO: Remove this after full migration to single binary
func mergeMetadataResolversMaps(m1, m2 map[string]config.StepData) map[string]config.StepData {
	res := map[string]config.StepData{}
	for k, v := range m1 {
		res[k] = v
	}
	for k, v := range m2 {
		res[k] = v
	}

	return res
}

// Execute is the starting point of the piper command line tool
func Execute() {
	log.Entry().Infof("Version %s", piperOsCmd.GitCommit)
	piperOsCmd.GeneralConfig.MetaDataResolver = func() map[string]config.StepData {
		return mergeMetadataResolversMaps(piperOsCmd.GetAllStepMetadata(), GetAllStepMetadata())
	}

	// OS commands
	rootCmd.AddCommand(piperOsCmd.ArtifactPrepareVersionCommand())
	rootCmd.AddCommand(piperOsCmd.CheckStepActiveCommand())
	rootCmd.AddCommand(piperOsCmd.CheckmarxExecuteScanCommand())
	rootCmd.AddCommand(piperOsCmd.CheckmarxOneExecuteScanCommand())
	rootCmd.AddCommand(piperOsCmd.CloudFoundryDeployCommand())
	rootCmd.AddCommand(piperOsCmd.CnbBuildCommand())
	rootCmd.AddCommand(piperOsCmd.CodeqlExecuteScanCommand())
	rootCmd.AddCommand(piperOsCmd.ConfigCommand())
	rootCmd.AddCommand(piperOsCmd.DefaultsCommand())
	rootCmd.AddCommand(piperOsCmd.DetectExecuteScanCommand())
	rootCmd.AddCommand(piperOsCmd.FortifyExecuteScanCommand())
	rootCmd.AddCommand(piperOsCmd.GaugeExecuteTestsCommand())
	rootCmd.AddCommand(piperOsCmd.GcpPublishEventCommand())
	rootCmd.AddCommand(piperOsCmd.GithubPublishReleaseCommand())
	rootCmd.AddCommand(piperOsCmd.GitopsUpdateDeploymentCommand())
	rootCmd.AddCommand(piperOsCmd.GolangBuildCommand())
	rootCmd.AddCommand(piperOsCmd.GradleExecuteBuildCommand())
	rootCmd.AddCommand(piperOsCmd.HadolintExecuteCommand())
	rootCmd.AddCommand(piperOsCmd.HelmExecuteCommand())
	rootCmd.AddCommand(piperOsCmd.ImagePushToRegistryCommand())
	rootCmd.AddCommand(piperOsCmd.KanikoExecuteCommand())
	rootCmd.AddCommand(piperOsCmd.KarmaExecuteTestsCommand())
	rootCmd.AddCommand(piperOsCmd.KubernetesDeployCommand())
	rootCmd.AddCommand(piperOsCmd.MavenBuildCommand())
	rootCmd.AddCommand(piperOsCmd.MavenExecuteIntegrationCommand())
	rootCmd.AddCommand(piperOsCmd.MtaBuildCommand())
	rootCmd.AddCommand(piperOsCmd.NewmanExecuteCommand())
	rootCmd.AddCommand(piperOsCmd.NpmExecuteScriptsCommand())
	rootCmd.AddCommand(piperOsCmd.ProtecodeExecuteScanCommand())
	rootCmd.AddCommand(piperOsCmd.PythonBuildCommand())
	rootCmd.AddCommand(piperOsCmd.ReadPipelineEnv())
	rootCmd.AddCommand(piperOsCmd.SonarExecuteScanCommand())
	rootCmd.AddCommand(piperOsCmd.TerraformExecuteCommand())
	rootCmd.AddCommand(piperOsCmd.UiVeri5ExecuteTestsCommand())
	rootCmd.AddCommand(piperOsCmd.VaultRotateSecretIdCommand())
	rootCmd.AddCommand(piperOsCmd.VersionCommand())
	rootCmd.AddCommand(piperOsCmd.WhitesourceExecuteScanCommand())
	rootCmd.AddCommand(piperOsCmd.WritePipelineEnv())

	rootCmd.AddCommand(SapCallFossServiceCommand())
	rootCmd.AddCommand(SapCallStagingServiceCommand())
	rootCmd.AddCommand(SapCheckECCNComplianceCommand())
	rootCmd.AddCommand(SapCheckPPMSComplianceCommand())
	rootCmd.AddCommand(SapCreateFosstarsReportCommand())
	rootCmd.AddCommand(SapCumulusDownloadCommand())
	rootCmd.AddCommand(SapCumulusUploadCommand())
	rootCmd.AddCommand(SapDownloadArtifactCommand())
	rootCmd.AddCommand(SapExecuteFastlaneCommand())
	rootCmd.AddCommand(SapGenerateEnvironmentInfoCommand())
	rootCmd.AddCommand(SapPipelineInitCommand())
	rootCmd.AddCommand(SapURLScanCommand())
	rootCmd.AddCommand(SapSUPAExecuteTestsCommand())
	rootCmd.AddCommand(SapExecuteApiMetadataValidatorCommand())
	rootCmd.AddCommand(SapXmakeExecuteBuildCommand())
	rootCmd.AddCommand(SapReportPipelineStatusCommand())
	rootCmd.AddCommand(SapDwCStageReleaseCommand())
	rootCmd.AddCommand(SapCollectInsightsCommand())
	rootCmd.AddCommand(SapAccessContinuumExecuteTestsCommand())
	rootCmd.AddCommand(SapExecuteCustomPolicyCommand())
	rootCmd.AddCommand(SapCollectPolicyResultsCommand())
	rootCmd.AddCommand(SapOcmCreateComponentCommand())
	rootCmd.AddCommand(SapExecuteCentralPolicyCommand())
	rootCmd.AddCommand(SapDasterExecuteScanCommand())
	rootCmd.AddCommand(SapGithubSecretScanningReportCommand())
	rootCmd.AddCommand(SapGitopsCFDeploymentCommand())
	rootCmd.AddCommand(SapGithubWriteDeploymentCommand())

	// Remove?
	// rootCmd.AddCommand(piperOsCmd.AbapAddonAssemblyKitCheckCVsCommand())
	// rootCmd.AddCommand(piperOsCmd.AbapAddonAssemblyKitCheckCommand())
	// rootCmd.AddCommand(piperOsCmd.AbapAddonAssemblyKitCheckPVCommand())
	// rootCmd.AddCommand(piperOsCmd.AbapAddonAssemblyKitCreateTargetVectorCommand())
	// rootCmd.AddCommand(piperOsCmd.AbapAddonAssemblyKitPublishTargetVectorCommand())
	// rootCmd.AddCommand(piperOsCmd.AbapAddonAssemblyKitRegisterPackagesCommand())
	// rootCmd.AddCommand(piperOsCmd.AbapAddonAssemblyKitReleasePackagesCommand())
	// rootCmd.AddCommand(piperOsCmd.AbapAddonAssemblyKitReserveNextPackagesCommand())
	// rootCmd.AddCommand(piperOsCmd.AbapEnvironmentAssembleConfirmCommand())
	// rootCmd.AddCommand(piperOsCmd.AbapEnvironmentAssemblePackagesCommand())
	// rootCmd.AddCommand(piperOsCmd.AbapEnvironmentBuildCommand())
	// rootCmd.AddCommand(piperOsCmd.AbapEnvironmentCheckoutBranchCommand())
	// rootCmd.AddCommand(piperOsCmd.AbapEnvironmentCloneGitRepoCommand())
	// rootCmd.AddCommand(piperOsCmd.AbapEnvironmentCreateSystemCommand())
	// rootCmd.AddCommand(piperOsCmd.AbapEnvironmentCreateTagCommand())
	// rootCmd.AddCommand(piperOsCmd.AbapEnvironmentPullGitRepoCommand())
	// rootCmd.AddCommand(piperOsCmd.AbapEnvironmentPushATCSystemConfigCommand())
	// rootCmd.AddCommand(piperOsCmd.AbapEnvironmentRunATCCheckCommand())
	// rootCmd.AddCommand(piperOsCmd.AbapEnvironmentRunAUnitTestCommand())
	// rootCmd.AddCommand(piperOsCmd.AbapLandscapePortalUpdateAddOnProductCommand())
	// rootCmd.AddCommand(piperOsCmd.AnsSendEventCommand())
	// rootCmd.AddCommand(piperOsCmd.ApiKeyValueMapDownloadCommand())
	// rootCmd.AddCommand(piperOsCmd.ApiKeyValueMapUploadCommand())
	// rootCmd.AddCommand(piperOsCmd.ApiProviderDownloadCommand())
	// rootCmd.AddCommand(piperOsCmd.ApiProviderListCommand())
	// rootCmd.AddCommand(piperOsCmd.ApiProxyDownloadCommand())
	// rootCmd.AddCommand(piperOsCmd.ApiProxyListCommand())
	// rootCmd.AddCommand(piperOsCmd.ApiProxyUploadCommand())
	// rootCmd.AddCommand(piperOsCmd.AscAppUploadCommand())
	// rootCmd.AddCommand(piperOsCmd.AwsS3UploadCommand())
	// rootCmd.AddCommand(piperOsCmd.AzureBlobUploadCommand())
	// rootCmd.AddCommand(piperOsCmd.BatsExecuteTestsCommand())
	// rootCmd.AddCommand(piperOsCmd.CloudFoundryCreateServiceCommand())
	// rootCmd.AddCommand(piperOsCmd.CloudFoundryCreateServiceKeyCommand())
	// rootCmd.AddCommand(piperOsCmd.CloudFoundryCreateSpaceCommand())
	// rootCmd.AddCommand(piperOsCmd.CloudFoundryDeleteServiceCommand())
	// rootCmd.AddCommand(piperOsCmd.CloudFoundryDeleteSpaceCommand())
	// rootCmd.AddCommand(piperOsCmd.CommandLineCompletionCommand())
	// rootCmd.AddCommand(piperOsCmd.ContainerExecuteStructureTestsCommand())
	// rootCmd.AddCommand(piperOsCmd.ContainerSaveImageCommand())
	// rootCmd.AddCommand(piperOsCmd.ContrastExecuteScanCommand())
	// rootCmd.AddCommand(piperOsCmd.CredentialdiggerScanCommand())
	// rootCmd.AddCommand(piperOsCmd.GctsCloneRepositoryCommand())
	// rootCmd.AddCommand(piperOsCmd.GctsCreateRepositoryCommand())
	// rootCmd.AddCommand(piperOsCmd.GctsDeployCommand())
	// rootCmd.AddCommand(piperOsCmd.GctsExecuteABAPQualityChecksCommand())
	// rootCmd.AddCommand(piperOsCmd.GctsExecuteABAPUnitTestsCommand())
	// rootCmd.AddCommand(piperOsCmd.GctsRollbackCommand())
	// rootCmd.AddCommand(piperOsCmd.GithubCheckBranchProtectionCommand())
	// rootCmd.AddCommand(piperOsCmd.GithubCommentIssueCommand())
	// rootCmd.AddCommand(piperOsCmd.GithubCreateIssueCommand())
	// rootCmd.AddCommand(piperOsCmd.GithubCreatePullRequestCommand())
	// rootCmd.AddCommand(piperOsCmd.GithubSetCommitStatusCommand())
	// rootCmd.AddCommand(piperOsCmd.InfluxWriteDataCommand())
	// rootCmd.AddCommand(piperOsCmd.IntegrationArtifactDeployCommand())
	// rootCmd.AddCommand(piperOsCmd.IntegrationArtifactDownloadCommand())
	// rootCmd.AddCommand(piperOsCmd.IntegrationArtifactGetMplStatusCommand())
	// rootCmd.AddCommand(piperOsCmd.IntegrationArtifactGetServiceEndpointCommand())
	// rootCmd.AddCommand(piperOsCmd.IntegrationArtifactResourceCommand())
	// rootCmd.AddCommand(piperOsCmd.IntegrationArtifactTransportCommand())
	// rootCmd.AddCommand(piperOsCmd.IntegrationArtifactTriggerIntegrationTestCommand())
	// rootCmd.AddCommand(piperOsCmd.IntegrationArtifactUnDeployCommand())
	// rootCmd.AddCommand(piperOsCmd.IntegrationArtifactUpdateConfigurationCommand())
	// rootCmd.AddCommand(piperOsCmd.IntegrationArtifactUploadCommand())
	// rootCmd.AddCommand(piperOsCmd.IsChangeInDevelopmentCommand())
	// rootCmd.AddCommand(piperOsCmd.JsonApplyPatchCommand())
	// rootCmd.AddCommand(piperOsCmd.MalwareExecuteScanCommand())
	// rootCmd.AddCommand(piperOsCmd.MavenExecuteCommand())
	// rootCmd.AddCommand(piperOsCmd.MavenExecuteStaticCodeChecksCommand())
	// rootCmd.AddCommand(piperOsCmd.NexusUploadCommand())
	// rootCmd.AddCommand(piperOsCmd.NpmExecuteLintCommand())
	// rootCmd.AddCommand(piperOsCmd.PipelineCreateScanSummaryCommand())
	// rootCmd.AddCommand(piperOsCmd.ShellExecuteCommand())
	// rootCmd.AddCommand(piperOsCmd.TmsExportCommand())
	// rootCmd.AddCommand(piperOsCmd.TmsUploadCommand())
	// rootCmd.AddCommand(piperOsCmd.TransportRequestDocIDFromGitCommand())
	// rootCmd.AddCommand(piperOsCmd.TransportRequestReqIDFromGitCommand())
	// rootCmd.AddCommand(piperOsCmd.TransportRequestUploadCTSCommand())
	// rootCmd.AddCommand(piperOsCmd.TransportRequestUploadRFCCommand())
	// rootCmd.AddCommand(piperOsCmd.TransportRequestUploadSOLMANCommand())
	// rootCmd.AddCommand(piperOsCmd.XsDeployCommand())

	addRootFlags(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		log.Entry().Fatal(err)
	}
}

func addRootFlags(rootCmd *cobra.Command) {
	provider, err := orchestrator.GetOrchestratorConfigProvider(nil)
	if err != nil {
		log.Entry().Error(err)
	}

	rootCmd.PersistentFlags().StringVar(&piperOsCmd.GeneralConfig.CorrelationID, "correlationID", provider.BuildURL(), "ID for unique identification of a pipeline run")
	rootCmd.PersistentFlags().StringVar(&piperOsCmd.GeneralConfig.CustomConfig, "customConfig", ".pipeline/config.yml", "Path to the pipeline configuration file")
	rootCmd.PersistentFlags().StringSliceVar(&piperOsCmd.GeneralConfig.DefaultConfig, "defaultConfig", nil, "Default configurations, passed as path to yaml file")
	rootCmd.PersistentFlags().StringSliceVar(&piperOsCmd.GeneralConfig.GitHubTokens, "gitHubTokens", piperOsCmd.AccessTokensFromEnvJSON(os.Getenv("PIPER_gitHubTokens")), "List of entries in form of <hostname>:<token> to allow GitHub token authentication for downloading config / defaults")
	rootCmd.PersistentFlags().BoolVar(&piperOsCmd.GeneralConfig.IgnoreCustomDefaults, "ignoreCustomDefaults", false, "Disables evaluation of the parameter 'customDefaults' in the pipeline configuration file")
	rootCmd.PersistentFlags().StringVar(&piperOsCmd.GeneralConfig.ParametersJSON, "parametersJSON", os.Getenv("PIPER_parametersJSON"), "Parameters to be considered in JSON format")
	rootCmd.PersistentFlags().StringVar(&piperOsCmd.GeneralConfig.EnvRootPath, "envRootPath", ".pipeline", "Root path to Piper pipeline shared environments")
	rootCmd.PersistentFlags().StringVar(&piperOsCmd.GeneralConfig.StageName, "stageName", os.Getenv("STAGE_NAME"), "Name of the stage for which configuration should be included")
	rootCmd.PersistentFlags().StringVar(&piperOsCmd.GeneralConfig.StepConfigJSON, "stepConfigJSON", os.Getenv("PIPER_stepConfigJSON"), "Step configuration in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&piperOsCmd.GeneralConfig.Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&piperOsCmd.GeneralConfig.LogFormat, "logFormat", "default", "Log format to use. Options: default, timestamp, plain, full.")
	rootCmd.PersistentFlags().StringVar(&piperOsCmd.GeneralConfig.VaultServerURL, "vaultServerUrl", "", "The vault server which should be used to fetch credentials")
	rootCmd.PersistentFlags().StringVar(&piperOsCmd.GeneralConfig.VaultNamespace, "vaultNamespace", "", "The vault namespace which should be used to fetch credentials")
	rootCmd.PersistentFlags().StringVar(&piperOsCmd.GeneralConfig.VaultPath, "vaultPath", "", "The path which should be used to fetch credentials")
	rootCmd.PersistentFlags().StringVar(&piperOsCmd.GeneralConfig.GCPJsonKeyFilePath, "gcpJsonKeyFilePath", "", "File path to Google Cloud Platform JSON key file")
	rootCmd.PersistentFlags().StringVar(&piperOsCmd.GeneralConfig.GCSFolderPath, "gcsFolderPath", "", "GCS folder path. One of the components of GCS target folder")
	rootCmd.PersistentFlags().StringVar(&piperOsCmd.GeneralConfig.GCSBucketId, "gcsBucketId", "", "Bucket name for Google Cloud Storage")
	rootCmd.PersistentFlags().StringVar(&piperOsCmd.GeneralConfig.GCSSubFolder, "gcsSubFolder", "", "Used to logically separate results of the same step result type")
}

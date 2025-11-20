package dwc

type (
	GlobalConfiguration struct {
		ArtifactURLs                    []string                            `json:"artifactURLs,omitempty"`
		ArtifactVersion                 string                              `json:"artifactVersion,omitempty"`
		CliPath                         string                              `json:"cliPath,omitempty"`
		DeriveAdditionalDownloadURLs    []DeriveAdditionalDownloadURLsEntry `json:"deriveAdditionalDownloadURLs,omitempty"`
		DownloadedArchivesPath          string                              `json:"downloadedArchivesPath,omitempty"`
		GatewayCertificatePath          string                              `json:"gatewayCertificatePath,omitempty"`
		GatewayURL                      string                              `json:"gatewayURL,omitempty"`
		GithubToken                     string                              `json:"githubToken,omitempty"`
		HelmChartDirectory              string                              `json:"helmChartDirectory,omitempty"`
		HelmChartURL                    string                              `json:"helmChartUrl,omitempty"`
		MtarFilePath                    string                              `json:"mtarFilePath,omitempty"`
		MtarUIPath                      string                              `json:"mtarUIPath,omitempty"`
		ProjectName                     string                              `json:"projectName,omitempty"`
		RequiredSuccessfulStages        []string                            `json:"requiredSuccessfulStages,omitempty"`
		StagesToWatch                   []string                            `json:"stagesToWatch,omitempty"`
		StageWatchPolicy                string                              `json:"stageWatchPolicy,omitempty" validate:"possible-values=overallSuccess subsetSuccess atLeastOneSuccessfulDeployment alwaysPass"`
		ThemistoInstanceCertificatePath string                              `json:"themistoInstanceCertificatePath,omitempty"`
		ThemistoInstanceURL             string                              `json:"themistoInstanceURL,omitempty"`
		UseCertLogin                    bool                                `json:"useCertLogin,omitempty"`
		WatchResourceOfInterest         bool                                `json:"watchResourceOfInterest,omitempty"`
		VaultBasePath                   string                              `json:"vaultBasePath,omitempty"`
		VaultPipelineName               string                              `json:"vaultPipelineName,omitempty"`
		PipelineID                      string                              `json:"pipelineID,omitempty"`
		CumulusPipelineRunKey           string                              `json:"cumulusPipelineRunKey,omitempty"`
	}

	PromotionResultEntry struct {
		Stage    string `json:"branch"`
		Status   string `json:"status"`
		VectorId string `json:"vectorId"`
		Error    string `json:"error,omitempty"`
	}

	ArtifactUploadResponse struct {
		AppName         string                 `json:"appname"`
		CreatedVector   string                 `json:"createdVector"`
		ID              string                 `json:"id"`
		PromotionResult []PromotionResultEntry `json:"promotionResult"`
		Type            string                 `json:"type"`
	}

	WaitForDeploymentResponse struct {
		Landscape string `json:"landscape"`
	}

	vectorMetadataBuildEntry struct {
		Branch              string `json:"branch"`
		JobUrl              string `json:"jobUrl"`
		BuildUrl            string `json:"buildUrl"`
		BuildNumber         string `json:"buildNumber"`
		GitCommitId         string `json:"gitCommitId"`
		GithubInstance      string `json:"githubInstance"`
		GithubRepository    string `json:"githubRepository"`
		GithubRepositoryUrl string `json:"githubRepositoryUrl"`
		Orchestrator        string `json:"orchestrator"`
	}

	DeriveAdditionalDownloadURLsEntry struct {
		Key         string `json:"key"`
		FindPattern string `json:"findPattern"`
		ReplaceWith string `json:"replaceWith"`
	}
)

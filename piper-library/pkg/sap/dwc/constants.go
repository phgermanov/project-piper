package dwc

const (
	UploadTypeService                              = "service"
	UploadTypeUI                                   = "ui"
	UploadTypeOrbit                                = "orbit"
	ArtifactTypeJava                               = "java"
	ArtifactTypeMaven                              = "maven"
	ArtifactTypeMta                                = "mta"
	ArtifactTypeMavenMta                           = "maven-mta"
	ArtifactTypeDocker                             = "docker"
	ArtifactTypeDockerBuildReleaseMetadata         = "dockerbuild-releaseMetadata"
	ArtifactTypeHelm                               = "helm"
	StageWatchPolicyOverallSuccess                 = "overallSuccess"
	StageWatchPolicySubsetSuccess                  = "subsetSuccess"
	StageWatchPolicyAtLeastOneSuccessfulDeployment = "atLeastOneSuccessfulDeployment"
	StageWatchPolicyAlwaysPass                     = "alwaysPass"
	ValuesFileName                                 = "values.yaml"
	promotionResultStatusCreated                   = "created"
	promotionResultStatusError                     = "error"
	promotionResultStatusSuccess                   = "success"
	allStagesSelector                              = "*"
	DefaultMtlsGatewayUrl                          = "https://api.mtls.dwc.tools.sap"
	DefaultGatewayUrl                              = "https://api.dwc.tools.sap"
	GitHubActionsOIDCTokenRequestTokenEnvVar       = "PIPER_ACTIONS_ID_TOKEN_REQUEST_TOKEN"
	GitHubActionsOIDCTokenRequestURLEnvVar         = "PIPER_ACTIONS_ID_TOKEN_REQUEST_URL"
)

var AvailableUploadTypes = map[string]struct{}{
	UploadTypeService: {},
	UploadTypeUI:      {},
	UploadTypeOrbit:   {},
}

var AvailableArtifactTypes = map[string]struct{}{
	ArtifactTypeJava:                       {},
	ArtifactTypeMaven:                      {},
	ArtifactTypeMta:                        {},
	ArtifactTypeMavenMta:                   {},
	ArtifactTypeDocker:                     {},
	ArtifactTypeDockerBuildReleaseMetadata: {},
	ArtifactTypeHelm:                       {},
}

package dwc

import (
	"encoding/json"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"strings"
)

const (
	NoEuporieTaskCollectionMetadataKey      = "noEuporieTaskCollection"
	NoRouteAssignmentMetadataKey            = "noRouteAssignment"
	AllowStaticRoutesMetadataKey            = "allowStaticRoutes"
	SelectiveMtaModuleDeploymentMetadataKey = "selectiveModuleDeployment"
)

func resolveVectorMetadataBuildEntry(collector MetadataCollector) (string, error) {
	branch, err := collector.GetMetadataEntry(gitBranchMetadataEntry)
	if err != nil {
		return "", fmt.Errorf("failed to get gitBranch from metadata: %w", err)
	}
	jobURL, err := collector.GetMetadataEntry(jobURLMetadataEntry)
	if err != nil {
		return "", fmt.Errorf("failed to get jobURL from metadata: %w", err)
	}
	buildURL, err := collector.GetMetadataEntry(buildURLMetadataEntry)
	if err != nil {
		return "", fmt.Errorf("failed to get buildURL from metadata: %w", err)
	}
	buildNumber, err := collector.GetMetadataEntry(buildIDMetadataEntry)
	if err != nil {
		return "", fmt.Errorf("failed to get buildID from metadata: %w", err)
	}
	githubRepositoryURL, err := collector.GetMetadataEntry(githubRepoURLMetadataEntry)
	if err != nil {
		return "", fmt.Errorf("failed to get githubRepoURL from metadata: %w", err)
	}
	githubInstance, githubRepository, err := extractGithubInstanceAndRepository(githubRepositoryURL)
	if err != nil {
		return "", fmt.Errorf("failed to extract github instance and repository from githubRepoURL: %w", err)
	}
	gitCommitID, err := collector.GetMetadataEntry(commitIdMetadataEntry)
	if err != nil {
		return "", fmt.Errorf("failed to get head commitId from metadata: %w", err)
	}
	orchestrator, err := collector.GetMetadataEntry(orchestratorTypeMetadataEntry)
	if err != nil {
		return "", fmt.Errorf("failed to get orchestratorType from metadata: %w", err)
	}
	buildMetadataJSON, err := json.Marshal(vectorMetadataBuildEntry{
		Branch:              branch,
		JobUrl:              jobURL,
		BuildUrl:            buildURL,
		BuildNumber:         buildNumber,
		GitCommitId:         gitCommitID,
		GithubInstance:      githubInstance,
		GithubRepository:    githubRepository,
		GithubRepositoryUrl: githubRepositoryURL,
		Orchestrator:        orchestrator,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal build metadata model: %w", err)
	}
	log.Entry().Debugf("collected metadata: %s", buildMetadataJSON)
	return string(buildMetadataJSON), nil
}

func extractGithubInstanceAndRepository(githubRepositoryURL string) (string, string, error) {
	if !strings.HasPrefix(githubRepositoryURL, "https://") {
		return "", "", fmt.Errorf("invalid format of repository url %s. expected prefix: https://", githubRepositoryURL)
	}
	prefixTrimmedURL := strings.TrimPrefix(githubRepositoryURL, "https://")
	urlParts := strings.Split(prefixTrimmedURL, "/")
	if len(urlParts) < 3 {
		return "", "", fmt.Errorf("invalid repository url %s", githubRepositoryURL)
	}
	instance := urlParts[0]
	org := urlParts[1]
	repository := urlParts[2]
	return instance, fmt.Sprintf("%s/%s", org, repository), nil
}

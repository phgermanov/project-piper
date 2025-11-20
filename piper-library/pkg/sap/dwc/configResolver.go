package dwc

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// ConfigResolver resolves default values for the pipeline config or combines config values for further processing
type ConfigResolver struct{}

func (resolver ConfigResolver) ResolveDefaultResourceName(collector MetadataCollector) (string, error) {
	githubRepositoryURL, err := collector.GetMetadataEntry(githubRepoURLMetadataEntry)
	if err != nil {
		return "", fmt.Errorf("failed to get githubRepoURL from metadata: %w", err)
	}
	githubInstance, githubRepository, err := extractGithubInstanceAndRepository(githubRepositoryURL)
	if err != nil {
		return "", fmt.Errorf("failed to extract github instance and repository from githubRepoURL: %w", err)
	}
	branch, err := collector.GetMetadataEntry(gitBranchMetadataEntry)
	if err != nil {
		return "", fmt.Errorf("failed to get gitBranch from metadata: %w", err)
	}
	return fmt.Sprintf("%s/%s/%s", githubInstance, githubRepository, branch), nil
}

func (resolver ConfigResolver) ResolveContainerImageURL(promotedDockerImage, artifactVersion string) (string, string, error) {
	if promotedDockerImage == "" {
		return "", "", errors.New("unable to resolve container image URL. Parameter promotedDockerImage is empty")
	}
	switch parts := strings.Split(promotedDockerImage, ":"); len(parts) {
	case 1:
		return resolver.resolveUntaggedImage(promotedDockerImage, artifactVersion)
	case 2:
		if strings.Contains(parts[1], "/") {
			return resolver.resolveUntaggedImage(promotedDockerImage, artifactVersion)
		}
		return parts[0], parts[1], nil
	case 3:
		if !strings.Contains(parts[2], "/") {
			return strings.Join(parts[0:2], ":"), parts[2], nil
		}
		fallthrough
	default:
		return "", "", fmt.Errorf("unable to parse promotedDockerImage: %s. Please contact the DwC team", promotedDockerImage)
	}
}

func (resolver ConfigResolver) resolveUntaggedImage(promotedDockerImage, artifactVersion string) (string, string, error) {
	if artifactVersion == "" {
		return "", "", errors.New("unable to resolve container image URL. Parameter artifactVersion is empty, but needed when no image tag is provided with promotedDockerImage")
	}
	return promotedDockerImage, artifactVersion, nil
}

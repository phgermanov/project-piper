package agent

import (
	"fmt"
	"runtime"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"

	"github.com/SAP/jenkins-library/pkg/orchestrator"
)

const (
	binaryName = "cumulus-policy-cli-%s.%s%s"
)

type Resolver interface {
	ResolveUrl(version string) (string, error)
	Download(version, targetFile string) error
}

type baseResolver struct {
	downloader piperHttp.Downloader
}

func (resolver *baseResolver) resolveBinary() string {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	return fmt.Sprintf(binaryName, runtime.GOOS, runtime.GOARCH, ext)
}

func NewGitHubResolver(token string, downloader piperHttp.Downloader) (Resolver, error) {
	return newGitHubResolver(token, downloader, &piperGitHubReleaseClientFactory{})
}

func NewArtifactoryResolver(downloader piperHttp.Downloader) (Resolver, error) {
	return newArtifactoryResolver(downloader)
}

func NewResolverByOrchestrator(orch orchestrator.Orchestrator, githubToken string) (Resolver, error) {
	var err error
	var resolver Resolver
	switch orch {
	case orchestrator.AzureDevOps:
		resolver, err = NewGitHubResolver(githubToken, &piperHttp.Client{})
	default:
		resolver, err = NewArtifactoryResolver(&piperHttp.Client{})
	}
	return resolver, err
}

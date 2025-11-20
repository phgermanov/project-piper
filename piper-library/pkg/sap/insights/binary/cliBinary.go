package binary

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/google/go-github/v68/github"
	"github.com/pkg/errors"
)

const (
	githubAPIURL                      = "https://github.tools.sap/api/v3"
	githubOwner                       = "hyperspace"
	githubDevOpsInsightsCLIRepo       = "DevOps-Insights"
	devopsInsightsCLIReleaseAssetName = "devops-insights.%s.%s%s"
	artifactoryAssetURL               = "https://int.repositories.cloud.sap/artifactory/devops-insights-generic/bin/%s/devops-insights.%s.%s%s"
)

type filePermissionEditor interface {
	Chmod(path string, mode os.FileMode) error
}

type fileDownloader interface {
	DownloadFile(url string, filename string, header http.Header, cookies []*http.Cookie) error
}

type CliReleaseResolver interface {
	GetLatestReleaseURL() (string, error)
	GetReleaseVersionURL(version string) (string, error)
	CreateHTTPHeader() http.Header
}

type artifactoryReleaseResolver struct {
}

func getRuntimeInfo() (string, string) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	return goos, goarch
}

func getExt(goos string) string {
	if goos == "windows" {
		// in case we are running on windows we need to add the .exe extension
		return ".exe"
	}
	return ""
}

func NewArtifactoryReleaseResolver() (CliReleaseResolver, error) {
	return &artifactoryReleaseResolver{}, nil
}

// GetLatestReleaseURL returns the latest devops-insights release URL
func (resolver *artifactoryReleaseResolver) GetLatestReleaseURL() (string, error) {
	return resolver.GetReleaseVersionURL("latest")
}

// GetReleaseVersionURL returns the URL for a given version
func (resolver *artifactoryReleaseResolver) GetReleaseVersionURL(version string) (string, error) {
	goos, goarch := getRuntimeInfo()
	return resolver.getAssetURL(version, goos, goarch), nil
}

func (resolver *artifactoryReleaseResolver) getAssetURL(versionPath, goos, goarch string) string {
	ext := getExt(goos)
	return fmt.Sprintf(artifactoryAssetURL, versionPath, goos, goarch, ext)
}

func (resolver *artifactoryReleaseResolver) CreateHTTPHeader() http.Header {
	downloadBinaryHeader := make(http.Header)
	return downloadBinaryHeader
}

// GithubToolsReleaseResolver
type githubToolsReleaseResolver struct {
	ctx       context.Context
	token     string
	transport *github.Client
}

func NewGithubToolsReleaseResolver(githubToken string) (CliReleaseResolver, error) {
	return newGithubToolsReleaseResolver(githubToken, githubAPIURL)
}

func newGithubToolsReleaseResolver(githubToken, githubURL string) (*githubToolsReleaseResolver, error) {
	if githubToken == "" {
		log.SetErrorCategory(log.ErrorConfiguration)
		return nil, errors.New("GitHub token not provided")
	}
	githubCtx, transport, err := piperGithub.NewClientBuilder(githubToken, githubURL).Build()
	if err != nil {
		return nil, err
	}
	return &githubToolsReleaseResolver{token: githubToken, ctx: githubCtx, transport: transport}, nil
}

func (resolver *githubToolsReleaseResolver) CreateHTTPHeader() http.Header {
	downloadBinaryHeader := make(http.Header)
	downloadBinaryHeader.Add("Authorization", fmt.Sprintf("token %s", resolver.token))
	downloadBinaryHeader.Add("Accept", "application/octet-stream")
	return downloadBinaryHeader
}

type releaseGetter func(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error)

func (resolver *githubToolsReleaseResolver) GetReleaseVersionURL(version string) (string, error) {
	goos, goarch := getRuntimeInfo()
	getReleaseByTag := func(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error) {
		return resolver.transport.Repositories.GetReleaseByTag(ctx, owner, repo, version)
	}
	return resolver.getAssetURL(goos, goarch, getReleaseByTag)
}

func (resolver *githubToolsReleaseResolver) GetLatestReleaseURL() (string, error) {
	goos, goarch := getRuntimeInfo()
	return resolver.getAssetURL(goos, goarch, resolver.transport.Repositories.GetLatestRelease)
}

func (resolver *githubToolsReleaseResolver) getAssetURL(goos, goarch string, getRelease releaseGetter) (string, error) {
	release, _, err := getRelease(resolver.ctx, githubOwner, githubDevOpsInsightsCLIRepo)
	if err != nil {
		log.SetErrorCategory(log.ErrorService)
		return "", fmt.Errorf("error resolving release assets: %w", err)
	}
	if release == nil {
		return "", errors.New("empty release")
	}
	// Adjusting the release asset name from GitHub with runtime infos
	ext := getExt(goos)
	releaseAsset := fmt.Sprintf(devopsInsightsCLIReleaseAssetName, goos, goarch, ext)

	var assetOfInterest *github.ReleaseAsset = nil
	for _, asset := range release.Assets {
		if asset.Name != nil {
			if *asset.Name == releaseAsset {
				assetOfInterest = asset
			}
		}
	}
	if assetOfInterest == nil {
		log.SetErrorCategory(log.ErrorService)
		return "", fmt.Errorf("release asset %s could not be located", releaseAsset)
	}
	if assetOfInterest.URL == nil {
		log.SetErrorCategory(log.ErrorService)
		return "", fmt.Errorf("unable to resolve target release asset. URL not set on asset response %v", assetOfInterest)
	}
	return *assetOfInterest.URL, nil
}

type CLIBinary struct {
	FileDownloader     fileDownloader
	PermEditor         filePermissionEditor
	CLIReleaseResolver CliReleaseResolver
	CLIExecutor        cliCommandExecutor
	ExecFactory        BlockingExecutorFactory
}

func (resolver CLIBinary) ExecuteDevOpInsights(targetBinary string, command []string) error {
	return resolver.CLIExecutor(targetBinary, resolver.ExecFactory, command)
}

func (resolver CLIBinary) InstallDevOpsInsightsCli(targetBinary, version string) error {
	var cliBinaryURL string
	var err error
	if strings.TrimSpace(version) != "" {
		cliBinaryURL, err = resolver.CLIReleaseResolver.GetReleaseVersionURL(version)
		if err != nil {
			return err
		}
	} else {
		cliBinaryURL, err = resolver.CLIReleaseResolver.GetLatestReleaseURL()
		if err != nil {
			return err
		}
	}
	log.Entry().Debugf("Binary of DevOps-Insights CLI is located at %s", cliBinaryURL)
	cliDownloadHeader := resolver.CLIReleaseResolver.CreateHTTPHeader()
	if err := resolver.FileDownloader.DownloadFile(cliBinaryURL, targetBinary, cliDownloadHeader, nil); err != nil {
		log.SetErrorCategory(log.ErrorService)
		return fmt.Errorf("failed to download DevOps-Insights CLI binary from %s: %w", cliBinaryURL, err)
	}
	log.Entry().Debugf("Downloaded DevOps-Insights CLI from %s to %s successfully", cliBinaryURL, targetBinary)
	if err := resolver.PermEditor.Chmod(targetBinary, 0755); err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return fmt.Errorf("failed to update file permissions of %s to 0755: %w", targetBinary, err)
	}
	log.Entry().Debugf("Updated file permissions of %s to 0755 successfully", targetBinary)
	return nil
}

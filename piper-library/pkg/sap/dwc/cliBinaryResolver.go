package dwc

import (
	"context"
	"fmt"
	"net/http"
	"os"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/google/go-github/v68/github"
	"github.com/pkg/errors"
)

const (
	githubAPIURL           = "https://github.tools.sap/api/v3"
	githubOwnerDwC         = "deploy-with-confidence"
	githubDwCCLIRepo       = "cli"
	dwcCLIReleaseAssetName = "dwc_linux"
	dwcCLIVersion          = "2.1.0"
)

type filePermissionEditor interface {
	Chmod(path string, mode os.FileMode) error
}

type fileDownloader interface {
	DownloadFile(url string, filename string, header http.Header, cookies []*http.Cookie) error
}

type cliReleaseResolver interface {
	GetCLIReleaseURL(withToken string) (string, error)
}

// DefaultReleaseResolver is not safe for concurrent usage
type DefaultReleaseResolver struct {
	ctx       context.Context
	token     string
	transport *github.Client
}

func (resolver *DefaultReleaseResolver) GetCLIReleaseURL(withToken string) (string, error) {
	if err := resolver.init(withToken); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", fmt.Errorf("error preparing github client: %w", err)
	}
	release, _, err := resolver.transport.Repositories.GetReleaseByTag(resolver.ctx, githubOwnerDwC, githubDwCCLIRepo, dwcCLIVersion)
	if err != nil {
		log.SetErrorCategory(log.ErrorService)
		return "", fmt.Errorf("error resolving release assets: %w", err)
	}
	assetOfInterest, err := resolver.resolveTargetAsset(release)
	if err != nil {
		log.SetErrorCategory(log.ErrorService)
		return "", fmt.Errorf("unable to resolve target release asset: %w", err)
	}
	if assetOfInterest.URL == nil {
		log.SetErrorCategory(log.ErrorService)
		return "", fmt.Errorf("unable to resolve target release asset. URL not set on asset response %v", assetOfInterest)
	}
	return *assetOfInterest.URL, nil
}

func (resolver *DefaultReleaseResolver) init(withToken string) error {
	resolver.token = withToken
	githubCtx, transport, err := piperGithub.NewClientBuilder(resolver.token, githubAPIURL).Build()
	if err != nil {
		return err
	}
	resolver.ctx = githubCtx
	resolver.transport = transport
	return nil
}

func (resolver *DefaultReleaseResolver) resolveTargetAsset(release *github.RepositoryRelease) (*github.ReleaseAsset, error) {
	if release == nil {
		return nil, errors.New("empty release")
	}
	for _, asset := range release.Assets {
		if asset.Name != nil {
			if *asset.Name == dwcCLIReleaseAssetName {
				return asset, nil
			}
		}
	}
	return nil, fmt.Errorf("release asset %s could not be located", dwcCLIReleaseAssetName)
}

type CLIBinaryResolver struct {
	FileDownloader     fileDownloader
	PermEditor         filePermissionEditor
	CLIReleaseResolver cliReleaseResolver
}

func (resolver CLIBinaryResolver) InstallDwCCli(githubToken string) error {
	if githubToken == "" {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.New("github token not provided")
	}
	cliBinaryURL, err := resolver.CLIReleaseResolver.GetCLIReleaseURL(githubToken)
	if err != nil {
		return err
	}
	log.Entry().Debugf("binary of DwC CLI is located at %s", cliBinaryURL)
	cliDownloadHeader := resolver.createHTTPHeader(githubToken)
	if err := resolver.FileDownloader.DownloadFile(cliBinaryURL, targetBinary, cliDownloadHeader, nil); err != nil {
		log.SetErrorCategory(log.ErrorService)
		return fmt.Errorf("failed to download DwC CLI binary from %s: %w", cliBinaryURL, err)
	}
	log.Entry().Debugf("downloaded DwC CLI from %s to %s successfully", cliBinaryURL, targetBinary)
	if err := resolver.PermEditor.Chmod(targetBinary, 0755); err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return fmt.Errorf("failed to update file permissions of %s to 0755: %w", targetBinary, err)
	}
	log.Entry().Debugf("updated file permissions of %s to 0755 successfully", targetBinary)
	return nil
}

func (resolver CLIBinaryResolver) createHTTPHeader(githubToken string) http.Header {
	downloadBinaryHeader := make(http.Header)
	downloadBinaryHeader.Add("Authorization", fmt.Sprintf("token %s", githubToken))
	downloadBinaryHeader.Add("Accept", "application/octet-stream")
	return downloadBinaryHeader
}

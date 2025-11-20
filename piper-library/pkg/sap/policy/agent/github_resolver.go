package agent

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/google/go-github/v68/github"
)

const (
	githubAPIURL = "https://github.tools.sap/api/v3"
	githubOrg    = "P4TEAM"
	githubRepo   = "cumulus-policy-cli"
)

type piperGitHubReleaseClientFactory struct {
}

func (p *piperGitHubReleaseClientFactory) NewReleaseClientBuilder(token, baseURL string) GitHubReleaseClientBuilder {
	return &piperGitHubReleaseClientBuilder{clientBuilder: piperGithub.NewClientBuilder(token, githubAPIURL)}
}

type piperGitHubReleaseClientBuilder struct {
	clientBuilder *piperGithub.ClientBuilder
}

func (p *piperGitHubReleaseClientBuilder) Build() (context.Context, GitHubReleaseClient, error) {
	context, client, err := p.clientBuilder.Build()
	return context, client.Repositories, err
}

type GitHubReleaseClientBuilder interface {
	Build() (context.Context, GitHubReleaseClient, error)
}

type GitHubReleaseClientFactory interface {
	NewReleaseClientBuilder(token, baseURL string) GitHubReleaseClientBuilder
}

type GitHubReleaseClient interface {
	GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error)
	GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error)
}

type gitHubResolver struct {
	baseResolver
	ctx                 context.Context
	token               string
	githubReleaseClient GitHubReleaseClient
}

func newGitHubResolver(token string, downloder piperHttp.Downloader, gitHubClientFactory GitHubReleaseClientFactory) (*gitHubResolver, error) {
	if token == "" {
		return nil, errors.New("no github token provided")
	}
	githubCtx, githHubReleaseClient, err := gitHubClientFactory.NewReleaseClientBuilder(token, githubAPIURL).Build()
	if err != nil {
		return nil, err
	}
	return &gitHubResolver{
		baseResolver:        baseResolver{downloader: downloder},
		token:               token,
		ctx:                 githubCtx,
		githubReleaseClient: githHubReleaseClient,
	}, nil
}

func (resolver *gitHubResolver) ResolveUrl(version string) (string, error) {
	release, err := resolver.getReleaseByVersion(version)
	if err != nil {
		return "", err
	}
	if release == nil {
		return "", errors.New("release not found")
	}
	binary := resolver.resolveBinary()
	for _, asset := range release.Assets {
		if asset.Name != nil && *asset.Name == binary {
			if asset.URL != nil && *asset.URL != "" {
				return *asset.URL, nil
			}
		}
	}
	return "", fmt.Errorf("binary %s for release %s not found!", binary, version)
}

func (resolver *gitHubResolver) getReleaseByVersion(version string) (*github.RepositoryRelease, error) {
	var release *github.RepositoryRelease
	var err error
	if version == "latest" {
		release, _, err = resolver.githubReleaseClient.GetLatestRelease(resolver.ctx, githubOrg, githubRepo)
	} else {
		release, _, err = resolver.githubReleaseClient.GetReleaseByTag(resolver.ctx, githubOrg, githubRepo, version)
	}
	if err != nil {
		return nil, err
	}
	return release, nil
}

func (resolver *gitHubResolver) httpHeader() http.Header {
	header := make(http.Header)
	header.Add("Authorization", fmt.Sprintf("token %s", resolver.token))
	header.Add("Accept", "application/octet-stream")
	return header
}

func (resolver *gitHubResolver) Download(version, targetFile string) error {
	url, err := resolver.ResolveUrl(version)
	log.Entry().Debugf("Download cumulus policy agent from github url \"%s\"...", url)
	if err != nil {
		return err
	}
	header := resolver.httpHeader()

	err = resolver.downloader.DownloadFile(url, targetFile, header, nil)
	if err != nil {
		return err
	}
	log.Entry().Debugf("Download of cumulus policy agent successful!")

	return nil
}

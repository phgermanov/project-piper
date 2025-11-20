package agent

import (
	"context"
	"testing"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
)

type mockGitHubReleaseClientFactory struct {
	releases *[]*github.RepositoryRelease
}

func (m *mockGitHubReleaseClientFactory) NewReleaseClientBuilder(token, baseURL string) GitHubReleaseClientBuilder {
	return &mockGitHubReleaseClientBuilder{releases: m.releases}
}

type mockGitHubReleaseClientBuilder struct {
	releases *[]*github.RepositoryRelease
}

func (m *mockGitHubReleaseClientBuilder) Build() (context.Context, GitHubReleaseClient, error) {
	return nil, &mockGitHubReleaseClient{m.releases}, nil
}

type mockGitHubReleaseClient struct {
	releases *[]*github.RepositoryRelease
}

func (m *mockGitHubReleaseClient) GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error) {
	return (*m.releases)[0], nil, nil
}
func (m *mockGitHubReleaseClient) GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error) {
	return (*m.releases)[0], nil, nil
}

func TestGitHubResolver(t *testing.T) {
	t.Parallel()
	t.Run("Can resolve and download a specific version", func(t *testing.T) {
		t.Parallel()

		releases := []*github.RepositoryRelease{}

		// use a pointer to our releases slice so that we can manipulate it later on
		resolver, err := newGitHubResolver("token", &mockDownloader{}, &mockGitHubReleaseClientFactory{releases: &releases})

		assert.NoError(t, err)
		assert.NotNil(t, resolver)

		binary := resolver.resolveBinary()
		assert.NotNil(t, binary)

		// append the dummy data
		releases = append(releases, &github.RepositoryRelease{
			Assets: []*github.ReleaseAsset{{
				Name: &binary,
				URL:  &binary,
			}},
		})

		version := "1.0.0"
		url, err := resolver.ResolveUrl(version)

		assert.NotNil(t, url)

		err = resolver.Download(version, binary)
		assert.NoError(t, err)

	})

	t.Run("Can resolve and download the latest version", func(t *testing.T) {
		t.Parallel()

		releases := make([]*github.RepositoryRelease, 0)

		resolver, err := newGitHubResolver("token", &mockDownloader{}, &mockGitHubReleaseClientFactory{releases: &releases})

		assert.NoError(t, err)
		assert.NotNil(t, resolver)

		binary := resolver.resolveBinary()
		assert.NotNil(t, binary)

		releases = append(releases, &github.RepositoryRelease{
			Assets: []*github.ReleaseAsset{{
				Name: &binary,
				URL:  &binary,
			}},
		})

		version := "latest"
		url, err := resolver.ResolveUrl(version)

		assert.NotNil(t, url)

		err = resolver.Download(version, binary)
		assert.NoError(t, err)

	})
}

package agent

import (
	"net/http"

	"testing"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/stretchr/testify/assert"
)

type mockDownloader struct {
}

func (c *mockDownloader) DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error {
	return nil
}

func (c *mockDownloader) SetOptions(options piperHttp.ClientOptions) {}

func TestResolver(t *testing.T) {
	t.Parallel()
	t.Run("Can create new artifactory resolver", func(t *testing.T) {
		t.Parallel()

		resolver, err := NewArtifactoryResolver(&mockDownloader{})

		assert.NoError(t, err)
		assert.NotNil(t, resolver)
	})

	t.Run("Can create new github resolver", func(t *testing.T) {
		t.Parallel()

		resolver, err := NewGitHubResolver("token", &mockDownloader{})

		assert.NoError(t, err)
		assert.NotNil(t, resolver)
	})

	t.Run("Can create new resolver by orchestrator (Azure)", func(t *testing.T) {
		t.Parallel()

		orchestrator := orchestrator.AzureDevOps
		resolver, err := NewResolverByOrchestrator(orchestrator, "token")

		assert.NoError(t, err)
		assert.NotNil(t, resolver)
	})
	t.Run("Can create new resolver by orchestrator (default)", func(t *testing.T) {
		t.Parallel()

		orchestrator := orchestrator.Jenkins
		resolver, err := NewResolverByOrchestrator(orchestrator, "token")

		assert.NoError(t, err)
		assert.NotNil(t, resolver)
	})
}

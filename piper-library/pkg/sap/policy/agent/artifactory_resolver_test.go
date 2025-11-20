package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArtifactoryResolver(t *testing.T) {
	t.Parallel()
	t.Run("Can resolve and download a version", func(t *testing.T) {
		t.Parallel()

		resolver, err := newArtifactoryResolver(&mockDownloader{})

		assert.NoError(t, err)
		assert.NotNil(t, resolver)

		binary := resolver.resolveBinary()
		version := "1.0.0"
		url, err := resolver.ResolveUrl(version)

		assert.NoError(t, err)
		assert.Equal(t, artifactoryBaseURL+version+"/"+binary, url)

		err = resolver.Download(version, binary)
		assert.NoError(t, err)
	})
}

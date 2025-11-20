package ocm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadNameAndVersionFromUrl(t *testing.T) {
	repo, chart, version, err := ReadNameAndVersionFromUrl("https://foo.bar/stage/repository/hello-ocm-helm-0.0.1-dev.tgz")
	assert.NoError(t, err)
	assert.Equal(t, "https://foo.bar/stage/repository", repo)
	assert.Equal(t, "hello-ocm-helm", chart)
	assert.Equal(t, "0.0.1-dev", version)

	repo, chart, version, err = ReadNameAndVersionFromUrl("https://foo.bar/stage/repository/hello-ocm-helm-0.0.1-dev-20240918130750+97feb5d34cebc31a8f9c56c8cd4fd8ff8003cd1e.tgz")
	assert.NoError(t, err)
	assert.Equal(t, "https://foo.bar/stage/repository", repo)
	assert.Equal(t, "hello-ocm-helm", chart)
	assert.Equal(t, "0.0.1-dev-20240918130750+97feb5d34cebc31a8f9c56c8cd4fd8ff8003cd1e", version)

	repo, chart, version, err = ReadNameAndVersionFromUrl("https://common.repositories.cloud.sap/sok-fpa134-helm/podinfo-v6.7.1.tgz ")
	assert.NoError(t, err)
	assert.Equal(t, "https://common.repositories.cloud.sap/sok-fpa134-helm", repo)
	assert.Equal(t, "podinfo", chart)
	assert.Equal(t, "v6.7.1", version)
}

func TestFailReadNameAndVersionFromUrl(t *testing.T) {
	repo, chart, version, err := ReadNameAndVersionFromUrl("scheme://does.not.exi^st")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid character")
	assert.Empty(t, version)
	assert.Empty(t, chart)
	assert.Empty(t, repo)

	repo, chart, version, err = ReadNameAndVersionFromUrl("https://foo.bar/stage/repository/hello-ocm-helm-0.no-valid-version.0.tgz")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to extract version from URL:")
}

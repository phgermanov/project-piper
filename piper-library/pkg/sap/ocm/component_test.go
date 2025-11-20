package ocm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestNewComponents(t *testing.T) {
	cfg := NewComponents()
	assert.Len(t, cfg.Components, 1)
	assert.Equal(t, "${componentName}", cfg.Components[0].Name)
	assert.Equal(t, "${artifactVersion}", cfg.Components[0].Version)
	assert.Equal(t, "${provider}", cfg.Components[0].Provider.Name)
	assert.Len(t, cfg.Components[0].Sources, 1)
	assert.Equal(t, "src", cfg.Components[0].Sources[0].Name)
	assert.Equal(t, "filesystem", cfg.Components[0].Sources[0].Type)
	assert.Equal(t, "gitHub", cfg.Components[0].Sources[0].Access["type"])
	assert.Equal(t, "${gitURL}", cfg.Components[0].Sources[0].Access["repoUrl"])
	assert.Equal(t, "${gitCommitID}", cfg.Components[0].Sources[0].Access["commit"])
}

func TestMarshalComponentsAccess(t *testing.T) {
	cfg := NewComponents()
	cfg.AddOciResource(-1)
	cfg.AddHelmAccess(-1)
	y, err := yaml.Marshal(cfg)
	assert.NoError(t, err)
	assert.Equal(t, `components:
    - name: ${componentName}
      version: ${artifactVersion}
      provider:
        name: ${provider}
      sources:
        - name: src
          type: filesystem
          access:
            commit: ${gitCommitID}
            repoUrl: ${gitURL}
            type: gitHub
      resources:
        - name: ${OCI_NAME}
          type: ociImage
          version: ${artifactVersion}
          access:
            imageReference: ${OCI_REFERENCE}
            type: ociArtifact
        - name: ${HELM_CHART}
          type: helmChart
          version: ${artifactVersion}
          access:
            helmChart: ${HELM_CHART}
            helmRepository: ${HELM_REPOSITORY}
            type: helm
            version: ${HELM_VERSION}
`, string(y))
}

func TestMarshalComponents2(t *testing.T) {
	cfg := NewComponents()
	cfg.AddOciResource(0)
	cfg.AddOciResource(1)
	cfg.AddOciResource(2)
	y, err := yaml.Marshal(cfg)
	assert.NoError(t, err)
	assert.Equal(t, `components:
    - name: ${componentName}
      version: ${artifactVersion}
      provider:
        name: ${provider}
      sources:
        - name: src
          type: filesystem
          access:
            commit: ${gitCommitID}
            repoUrl: ${gitURL}
            type: gitHub
      resources:
        - name: ${OCI_NAME_0}
          type: ociImage
          version: ${artifactVersion}
          access:
            imageReference: ${OCI_REFERENCE_0}
            type: ociArtifact
        - name: ${OCI_NAME_1}
          type: ociImage
          version: ${artifactVersion}
          access:
            imageReference: ${OCI_REFERENCE_1}
            type: ociArtifact
        - name: ${OCI_NAME_2}
          type: ociImage
          version: ${artifactVersion}
          access:
            imageReference: ${OCI_REFERENCE_2}
            type: ociArtifact
`, string(y))
}

func TestComponents_Replace(t *testing.T) {
	cfg := NewComponents()
	assert.Equal(t, "${componentName}", cfg.Name())
	assert.Equal(t, "${artifactVersion}", cfg.Version())
	settings := ComponentInfo{
		"componentName":   "my-component",
		"artifactVersion": "1.2.3",
		"provider":        "SAP SE",
		"gitURL":          "https://github.com/example/repo",
		"gitCommitID":     "abc123",
	}
	err := cfg.Replace(settings)
	assert.NoError(t, err)
	assert.Equal(t, "my-component", cfg.Name())
	assert.Equal(t, "1.2.3", cfg.Version())
	assert.Equal(t, "SAP SE", cfg.Components[0].Provider.Name)
	assert.Equal(t, "https://github.com/example/repo", cfg.Components[0].Sources[0].Access["repoUrl"])
	assert.Equal(t, "abc123", cfg.Components[0].Sources[0].Access["commit"])
}

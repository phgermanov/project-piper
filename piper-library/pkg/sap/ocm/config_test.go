package ocm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	assert.Equal(t, "generic.config.ocm.software/v1", cfg.Type)
	assert.Len(t, cfg.Configurations, 1)
	assert.Equal(t, "credentials.config.ocm.software", cfg.Configurations[0].Type)
}

func TestMarshal(t *testing.T) {
	cfg := NewConfig()
	cfg.AddCredential(Helm, "https://staging.repositories.cloud.sap/stage/repository/371000318081-20240923-130414706-732/hello-ocm-helm-0.0.1-dev.tgz", "you", "secret")
	cfg.AddCredential(Oci, "371000848081-20240923-135845457-687.staging.repositories.cloud.sap", "you", "secret")
	cfg.AddCredential(Oci, "https://staging.repositories.cloud.sap/stage/repository/371000318081-20240923-130414706-732", "you", "secret")
	y, err := yaml.Marshal(cfg)
	assert.NoError(t, err)
	assert.Equal(t, `type: generic.config.ocm.software/v1
configurations:
    - type: credentials.config.ocm.software
      consumers:
        - identity:
            type: HelmChartRepository
            hostname: staging.repositories.cloud.sap
          credentials:
            - type: Credentials
              properties:
                username: you
                password: secret
        - identity:
            type: OCIRegistry
            hostname: 371000848081-20240923-135845457-687.staging.repositories.cloud.sap
          credentials:
            - type: Credentials
              properties:
                username: you
                password: secret
        - identity:
            type: OCIRegistry
            hostname: staging.repositories.cloud.sap
            pathprefix: stage/repository
          credentials:
            - type: Credentials
              properties:
                username: you
                password: secret
`, string(y))
}

func TestAddDockerConfig(t *testing.T) {
	cfg := NewConfig()
	cfg.AddDockerConfig("/my/path/docker/config.json")
	yamlGenerated, err := yaml.Marshal(cfg)
	assert.NoError(t, err)

	yamlExpected := []byte(`
type: generic.config.ocm.software/v1
configurations:
  - type: credentials.config.ocm.software
    consumers: []
    repositories:
      - repository:
          type: DockerConfig/v1
          dockerConfigFile: /my/path/docker/config.json
          propagateConsumerIdentity: true
`)

	var generated, expected map[string]interface{}
	assert.NoError(t, yaml.Unmarshal(yamlGenerated, &generated))
	assert.NoError(t, yaml.Unmarshal(yamlExpected, &expected))
	assert.Equal(t, expected, generated)
}

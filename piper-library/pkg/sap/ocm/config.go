package ocm

type Properties struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Email    string `yaml:"email,omitempty"`
}

type Credential struct {
	Type       string     `yaml:"type"`
	Properties Properties `yaml:"properties"`
}

type Consumer struct {
	Identity    Identity     `yaml:"identity"`
	Credentials []Credential `yaml:"credentials"`
}

type RepositoryItem struct {
	Repository Repository `yaml:"repository,omitempty"`
}

type Repository struct {
	Type                      string `yaml:"type"`
	DockerConfigFile          string `yaml:"dockerConfigFile"`
	PropagateConsumerIdentity bool   `yaml:"propagateConsumerIdentity"`
}

type Configuration struct {
	Type         string           `yaml:"type"`
	Consumers    []Consumer       `yaml:"consumers"`
	Repositories []RepositoryItem `yaml:"repositories,omitempty"`
}

/*
Config - This struct represents the OCM Config.

Simple example (~/.ocmconfig):

	---
	type: generic.config.ocm.software/v1
	configurations:
	  - type: credentials.config.ocm.software
		consumers:
		  - identity:
			  type: OCIRegistry
			  hostname: 376101718081-20250508-162504509-551.staging.repositories.cloud.sap
			credentials:
			  - type: Credentials
				properties:
				  username: ****
				  password: ****
		  - identity:
			  type: OCIRegistry
			  hostname: 376100978081-20250508-16242266-421.staging.repositories.cloud.sap
			credentials:
			  - type: Credentials
				properties:
				  username: ****
				  password: ****
		repositories:
		  - repository:
			  type: DockerConfig/v1
			  dockerConfigFile: ****
			  propagateConsumerIdentity: true
*/
type Config struct {
	Type           string          `yaml:"type"`
	Configurations []Configuration `yaml:"configurations"`
}

// NewConfig creates a new OCM Config.
func NewConfig() *Config {
	cfg := &Config{
		Type: "generic.config.ocm.software/v1",
	}
	cfg.Configurations = append(cfg.Configurations, Configuration{
		Type: "credentials.config.ocm.software",
	})
	return cfg
}

// AddCredential adds a new Consumer with Credential to the first (Credential) configuration.
func (c *Config) AddCredential(idType identityType, url, user, pass string, options ...func(*Credential)) {
	crd := &Credential{
		Type: "Credentials",
		Properties: Properties{
			Username: user,
			Password: pass,
		},
	}
	for _, option := range options {
		option(crd)
	}

	identity := NewIdentity(idType, url)
	c.Configurations[0].Consumers = append(c.Configurations[0].Consumers, Consumer{
		Identity:    identity,
		Credentials: []Credential{*crd},
	})
}

// AddDockerConfig instructs the OCM config to reuse credentials from the docker config.json file.
func (c *Config) AddDockerConfig(pathDockerConfigJSON string) {
	repo := Repository{
		Type:                      "DockerConfig/v1",
		DockerConfigFile:          pathDockerConfigJSON,
		PropagateConsumerIdentity: true,
	}

	c.Configurations[0].Repositories = append(c.Configurations[0].Repositories, RepositoryItem{
		Repository: repo,
	})
}

package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Piper          PiperConfig          `yaml:"piper"`
	SapPiper       PiperConfig          `yaml:"sapPiper"`
	ExampleProject ExampleProjectConfig `yaml:"exampleProject"`
	Act            ActConfig            `yaml:"act"`
	MockServices   MockServicesConfig   `yaml:"mockServices"`
	Secrets        map[string]string    `yaml:"secrets"`
	Variables      map[string]string    `yaml:"variables"`
	PipelineConfig map[string]any       `yaml:"pipelineConfig"`
	Logging        LoggingConfig        `yaml:"logging"`
	Test           TestConfig           `yaml:"test"`
}

type PiperConfig struct {
	BinaryPath  string `yaml:"binaryPath"`
	AutoBuild   bool   `yaml:"autoBuild"`
	SourcePath  string `yaml:"sourcePath"`
	Version     string `yaml:"version"`
}

type ExampleProjectConfig struct {
	Path         string `yaml:"path"`
	Type         string `yaml:"type"`
	Branch       string `yaml:"branch"`
	WorkflowFile string `yaml:"workflowFile"`
}

type ActConfig struct {
	BinaryPath            string            `yaml:"binaryPath"`
	Platform              string            `yaml:"platform"`
	ContainerArchitecture string            `yaml:"containerArchitecture"`
	Flags                 []string          `yaml:"flags"`
	Env                   map[string]string `yaml:"env"`
}

type MockServicesConfig struct {
	Enabled   bool               `yaml:"enabled"`
	Host      string             `yaml:"host"`
	Port      int                `yaml:"port"`
	BaseURL   string             `yaml:"baseUrl"`
	Endpoints []MockEndpoint     `yaml:"endpoints"`
}

type MockEndpoint struct {
	Path     string       `yaml:"path"`
	Method   string       `yaml:"method"`
	Response MockResponse `yaml:"response"`
}

type MockResponse struct {
	Status int            `yaml:"status"`
	Body   map[string]any `yaml:"body"`
}

type LoggingConfig struct {
	Level      string `yaml:"level"`
	OutputFile string `yaml:"outputFile"`
	Colorize   bool   `yaml:"colorize"`
}

type TestConfig struct {
	Stages   []string `yaml:"stages"`
	FailFast bool     `yaml:"failFast"`
	Cleanup  bool     `yaml:"cleanup"`
	KeepLogs bool     `yaml:"keepLogs"`
	Timeout  int      `yaml:"timeout"`
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if cfg.Piper.BinaryPath == "" {
		cfg.Piper.BinaryPath = "./bin/piper"
	}
	if cfg.SapPiper.BinaryPath == "" {
		cfg.SapPiper.BinaryPath = "./bin/sap-piper"
	}
	if cfg.MockServices.Host == "" {
		cfg.MockServices.Host = "localhost"
	}
	if cfg.MockServices.Port == 0 {
		cfg.MockServices.Port = 8888
	}
	if cfg.Act.BinaryPath == "" {
		cfg.Act.BinaryPath = "act"
	}

	return &cfg, nil
}

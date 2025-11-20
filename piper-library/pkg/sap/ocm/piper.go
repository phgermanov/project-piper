package ocm

import (
	"github.com/SAP/jenkins-library/cmd"
	"github.com/SAP/jenkins-library/pkg/log"
)

// ComponentConstructorFileName the default name of the component constructor file == component-constructor.yaml
const ComponentConstructorFileName = "component-constructor.yaml"

// SettingsFileName the default name of the settings file == settings.yaml
const SettingsFileName = "settings.yaml"

type ComponentInfo map[string]interface{}

func (c ComponentInfo) Set(key string, value interface{}) {
	if "RepoInfo" == key {
		_, ok := value.(*StagingRepoInfo)
		if !ok {
			log.Entry().Error("Error setting RepoInfo")
			return
		}
	}
	c[key] = value
}

func (c ComponentInfo) Get(key string) string {
	s, ok := c[key].(string)
	if !ok {
		log.Entry().Errorf("Error getting key %s as string! value: '%v'", key, c[key])
		return ""
	}
	return s
}

func (c ComponentInfo) RepoInfo() *StagingRepoInfo {
	if c.Has("RepoInfo") {
		return c["RepoInfo"].(*StagingRepoInfo)
	}
	return &StagingRepoInfo{}
}

func (c ComponentInfo) Has(key string) bool {
	val := c[key]
	if val == nil {
		return false
	}
	str, ok := val.(string)
	if ok {
		return IsSet(str)
	}
	return true
}

// NewComponentInfo represents the component information.
func NewComponentInfo() ComponentInfo {
	return make(map[string]interface{})
}

// /////////////////////////////////////////////////////////////////////////////
// Staging-Service Client

type StagingRepoInfo struct {
	User          string `json:"user,omitempty"`
	Password      string `json:"password,omitempty"`
	Repository    string `json:"repository,omitempty"`
	RepositoryURL string `json:"repositoryURL,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// Helper functions

type PiperStage int

// OCM relevant Piper stages
const (
	Build PiperStage = iota
	Other
)

// StageName returns the name of the current stage.
func StageName() string {
	return cmd.GeneralConfig.StageName
}

// Stage returns the stage of the pipeline.
func Stage() PiperStage {
	switch StageName() {
	case "Central Build", "Build":
		return Build
	default:
		return Other
	}
}

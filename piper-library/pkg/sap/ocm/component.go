package ocm

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"regexp"
)

// Components describes the OCM Components.
type Components struct {
	Components []Component `yaml:"components"`
}

type Component struct {
	Name     string `yaml:"name"`    // ${componentName}
	Version  string `yaml:"version"` // ${artifactVersion}
	Provider struct {
		Name string `yaml:"name"` // ${provider} - "SAP SE (github-org)"
	} `yaml:"provider"`
	Sources   []Source   `yaml:"sources,omitempty"` // gitSource()
	Resources []Resource `yaml:"resources,omitempty"`
}

// NewComponents creates new OCM Components with sources.
func NewComponents() *Components {
	c := &Component{}
	c.Name = "${componentName}"
	c.Version = "${artifactVersion}"
	c.Provider.Name = "${provider}"
	s := gitSource()
	c.Sources = []Source{*s}
	c.Resources = []Resource{}
	return &Components{
		Components: []Component{*c},
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// Source describes a source in the OCM Components.
type Source struct {
	Name   string            `yaml:"name"`   // "src"
	Type   string            `yaml:"type"`   // "filesystem"
	Access map[string]string `yaml:"access"` // gitSource()
}

func gitSource() *Source {
	s := &Source{}
	s.Name = "src"
	s.Type = "filesystem"
	s.Access = make(map[string]string)
	s.Access["type"] = "gitHub"
	s.Access["repoUrl"] = "${gitURL}"
	s.Access["commit"] = "${gitCommitID}"
	return s
}

////////////////////////////////////////////////////////////////////////////////////////////////////

// Resource describes a resource in the OCM Components.
type Resource struct {
	Name    string            `yaml:"name"`
	Type    string            `yaml:"type"` // https://github.com/open-component-model/ocm/blob/main/pkg/contexts/ocm/resourcetypes/const.go#L8
	Version string            `yaml:"version"`
	Access  map[string]string `yaml:"access,omitempty"` // dynamic, depends on type
}

func Suffix(i int) string {
	if i < 0 {
		return "}"
	}
	return fmt.Sprint("_", i, "}")
}

// helmResource adds a new Helm resource to the OCM Component.
func (c *Components) helmResource(i int) *Resource {
	r := &Resource{}
	r.Name = "${HELM_CHART" + Suffix(i)
	r.Version = "${artifactVersion}"
	r.Type = "helmChart"
	return r
}

// AddHelmAccess adds a new Helm access resource to the OCM Component.
func (c *Components) AddHelmAccess(i int) {
	r := c.helmResource(i)
	r.Access = make(map[string]string)
	r.Access["type"] = "helm"
	r.Access["helmRepository"] = "${HELM_REPOSITORY}"
	r.Access["helmChart"] = "${HELM_CHART" + Suffix(i)
	r.Access["version"] = "${HELM_VERSION" + Suffix(i)
	c.Components[0].Resources = append(c.Components[0].Resources, *r)
}

// AddOciResource adds a new oci resource to the OCM Components.
func (c *Components) AddOciResource(i int) {
	r := &Resource{}
	r.Name = "${OCI_NAME" + Suffix(i)
	r.Version = "${artifactVersion}"
	r.Type = "ociImage"
	r.Access = make(map[string]string)
	r.Access["type"] = "ociArtifact"
	r.Access["imageReference"] = "${OCI_REFERENCE" + Suffix(i)
	c.Components[0].Resources = append(c.Components[0].Resources, *r)
}

// Marshal into YAML string
func (c *Components) Marshal() ([]byte, error) {
	b, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Unmarshal from YAML string
func (c *Components) Unmarshal(data []byte) error {
	err := yaml.Unmarshal(data, c)
	if err != nil {
		return err
	}
	return nil
}

// Replace replaces the placeholders in the YAML string with the actual values.
func (c *Components) Replace(settings ComponentInfo) error {
	// get string representation of the component
	componentData, err := c.Marshal()
	if err != nil {
		return err
	}
	// replace the placeholders with the actual values
	re := regexp.MustCompile(`\$\{([^}]+)}`)
	replaced := re.ReplaceAllStringFunc(string(componentData), func(s string) string {
		key := re.FindStringSubmatch(s)[1]
		if val, ok := settings[key]; ok {
			return fmt.Sprintf("%v", val)
		}
		return s
	})
	// unmarshal the replaced string back into the component
	err = c.Unmarshal([]byte(replaced))
	if err != nil {
		return err
	}

	return nil
}

// Version returns the version of the first component in the list.
func (c *Components) Version() string {
	if len(c.Components) == 0 {
		return ""
	}
	return c.Components[0].Version
}

// Name returns the name of the first component in the list.
func (c *Components) Name() string {
	if len(c.Components) == 0 {
		return ""
	}
	return c.Components[0].Name
}

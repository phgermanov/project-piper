package ocm

import (
	"fmt"
)

// identityType represents the type of an identity.
type identityType int

const (
	// Oci == OCIRegistry represents an OCI registry.
	Oci identityType = iota
	// Helm == HelmChartRepository represents a Helm chart repository.
	Helm
)

func (t identityType) String() string {
	switch t {
	case Oci:
		return "OCIRegistry"
	case Helm:
		return "HelmChartRepository"
	}
	return "unknown"
}

func (t identityType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *identityType) UnmarshalText(text []byte) error {
	switch string(text) {
	case "OCIRegistry":
		*t = Oci
	case "HelmChartRepository":
		*t = Helm
	default:
		return fmt.Errorf("unknown identity type: %s", text)
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////

// Identity represents the identity of a consumer.
type Identity struct {
	Type       identityType `yaml:"type"`
	Hostname   string       `yaml:"hostname"`
	Scheme     string       `yaml:"scheme,omitempty"`
	Port       string       `yaml:"port,omitempty"`
	PathPrefix string       `yaml:"pathprefix,omitempty"`
}

// NewIdentity creates a new identity from the given type and url.
func NewIdentity(t identityType, url string) Identity {
	pp := PathPrefix(url)
	if t == Helm {
		// current HelmChartRepository identity does not support path prefix properly
		pp = ""
	}
	return Identity{
		Type:       t,
		Hostname:   Hostname(url),
		Scheme:     Scheme(url),
		Port:       Port(url),
		PathPrefix: pp,
	}
}

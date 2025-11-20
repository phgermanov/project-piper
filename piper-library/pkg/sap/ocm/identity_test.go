package ocm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdentityTypeString(t *testing.T) {
	tests := []struct {
		input    identityType
		expected string
	}{
		{Oci, "OCIRegistry"},
		{Helm, "HelmChartRepository"},
		{identityType(999), "unknown"},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.input.String())
	}
}

func TestIdentityTypeMarshalText(t *testing.T) {
	tests := []struct {
		input    identityType
		expected []byte
	}{
		{Oci, []byte("OCIRegistry")},
		{Helm, []byte("HelmChartRepository")},
	}

	for _, test := range tests {
		result, err := test.input.MarshalText()
		assert.NoError(t, err)
		assert.Equal(t, test.expected, result)
	}
}

func TestIdentityTypeUnmarshalText(t *testing.T) {
	tests := []struct {
		input    []byte
		expected identityType
		hasError bool
	}{
		{[]byte("OCIRegistry"), Oci, false},
		{[]byte("HelmChartRepository"), Helm, false},
		{[]byte("UnknownType"), identityType(0), true},
	}

	for _, test := range tests {
		var result identityType
		err := result.UnmarshalText(test.input)
		if test.hasError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, result)
		}
	}
}

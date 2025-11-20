package ocm

import (
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/sirupsen/logrus"
	"reflect"
	"testing"
)

const invalidUrl = "invalid url with CTL Byte \x17" // ASCII 23 in hex

func TestTrimHttpPrefix(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"http://example.com", "example.com"},
		{"https://example.com", "example.com"},
		{"ftp://example.com", "ftp://example.com"},
		{"example.com", "example.com"},
	}
	for _, c := range cases {
		result := TrimHttpPrefix(c.input)
		if result != c.expected {
			t.Errorf("TrimHttpPrefix(%q) == %q, want %q", c.input, result, c.expected)
		}
	}
}

func TestHostname(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"http://example.com:8080/path", "example.com:8080"},
		{"https://example.com", "example.com"},
		{"example.com", "example.com"},
		{invalidUrl, ""},
	}
	for _, c := range cases {
		result := Hostname(c.input)
		if result != c.expected {
			t.Errorf("Hostname(%q) == %q, want %q", c.input, result, c.expected)
		}
	}
}

func TestPathPrefix(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"http://example.com/foo/bar/file.txt", "foo/bar"},
		{"https://example.com/file.txt", ""},
		{"http://example.com/foo/bar/", "foo/bar"},
		{invalidUrl, ""},
	}
	for _, c := range cases {
		result := PathPrefix(c.input)
		if result != c.expected {
			t.Errorf("PathPrefix(%q) == %q, want %q", c.input, result, c.expected)
		}
	}
}

func TestPort(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"http://example.com:8080", "8080"},
		{"https://example.com", ""},
		{invalidUrl, ""},
	}
	for _, c := range cases {
		result := Port(c.input)
		if result != c.expected {
			t.Errorf("Port(%q) == %q, want %q", c.input, result, c.expected)
		}
	}
}

func TestIsSet(t *testing.T) {
	cases := []struct {
		input    string
		expected bool
	}{
		{"set", true},
		{"", false},
	}
	for _, c := range cases {
		result := IsSet(c.input)
		if result != c.expected {
			t.Errorf("IsSet(%q) == %v, want %v", c.input, result, c.expected)
		}
	}
}

func TestVerbose(t *testing.T) {
	cases := []struct {
		level    logrus.Level
		input    []string
		expected []string
	}{
		{logrus.DebugLevel, []string{"a", "b"}, []string{"a", "b", "--loglevel", "Debug"}},
		{logrus.InfoLevel, []string{"a", "b"}, []string{"a", "b"}},
	}
	for i, c := range cases {
		log.SetVerbose(i%2 == 0) // doesn't respect 'false' value, ...
		logrus.SetLevel(c.level) // ... but we want to test both cases

		result := Verbose(c.input)
		// test array for deep equality
		if !reflect.DeepEqual(result, c.expected) {
			t.Errorf("case %v: Verbose(%v) == %v, want %v", i, c.input, result, c.expected)
		}
	}
}

func TestScheme(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"https://example.com", ""},    // https is default, returns empty
		{"http://example.com", "http"}, // http scheme
		{"ftp://example.com", "ftp"},   // ftp scheme
		{"example.com", ""},            // no scheme, returns empty
		{invalidUrl, ""},               // invalid url, returns empty
	}
	for _, c := range cases {
		result := Scheme(c.input)
		if result != c.expected {
			t.Errorf("Scheme(%q) == %q, want %q", c.input, result, c.expected)
		}
	}
}

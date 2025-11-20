package ocm

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
)

// TrimHttpPrefix removes the http:// or https:// prefix from given string.
func TrimHttpPrefix(url string) string {
	result := strings.TrimPrefix(url, "http://")
	return strings.TrimPrefix(result, "https://")
}

// Hostname extracts the hostname from a URL.
func Hostname(rawUrl string) string {
	u, err := url.Parse(rawUrl)
	if err != nil {
		log.Entry().Warnf("parsing URL failed: %s", err.Error())
		return ""
	}
	if u.Scheme == "" {
		return Hostname("https://" + rawUrl)
	}
	return u.Host
}

func PathPrefix(rawUrl string) string {
	u, err := url.Parse(rawUrl)
	if err != nil {
		log.Entry().Warnf("parsing URL failed: %s", err.Error())
		return ""
	}
	path := filepath.ToSlash(filepath.Dir(strings.TrimPrefix(u.Path, "/")))
	if path == "." {
		return ""
	}
	return path
}

func Port(rawUrl string) string {
	u, err := url.Parse(rawUrl)
	if err != nil {
		log.Entry().Warnf("parsing URL failed: %s", err.Error())
		return ""
	}
	return u.Port()
}

func Scheme(rawUrl string) string {
	u, err := url.Parse(rawUrl)
	if err != nil {
		log.Entry().Warnf("parsing URL failed: %s", err.Error())
		return ""
	}
	if u.Scheme == "https" {
		// is the default scheme and can be omitted
		return ""
	}
	return u.Scheme
}

// IsSet checks if a string is not empty and not only whitespace.
func IsSet(s string) bool {
	return len(strings.TrimSpace(s)) > 0
}

// Verbose adds the --loglevel Debug argument to the OCM command if the log level is set to verbose.
func Verbose(ocmArgs []string) []string {
	if log.IsVerbose() {
		ocmArgs = append(ocmArgs, "--loglevel", "Debug")
	}
	return ocmArgs
}

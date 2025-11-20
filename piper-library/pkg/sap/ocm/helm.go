package ocm

import (
	"errors"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
)

// ReadNameAndVersionFromUrl reads the name and version from a Helm chart URL.
func ReadNameAndVersionFromUrl(helmChartURL string) (string, string, string, error) {
	u, err := url.Parse(helmChartURL)
	if err != nil {
		log.Entry().Errorf("Failed to parse URL: %v", err)
		return "", "", "", err
	}
	// get the last part of the URL
	fileName := path.Base(u.Path)
	// remove the file extension
	chartName := fileName[:len(fileName)-len(path.Ext(fileName))]

	// using a regular expression to extract the version from the URL
	re := regexp.MustCompile(`-(v?\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?(-[0-9]{14})?(\+[a-zA-Z0-9]{7,40})?)`)
	matches := re.FindStringSubmatch(chartName)
	if len(matches) < 2 {
		return "", "", "", errors.New("failed to extract version from URL: " + helmChartURL)
	}
	version := matches[1]

	// remove the version part from the chart name
	chartName = strings.TrimSuffix(chartName, "-"+version)
	// remove the chart name from the URL
	u.Path = path.Dir(u.Path)

	return u.String(), chartName, version, nil
}

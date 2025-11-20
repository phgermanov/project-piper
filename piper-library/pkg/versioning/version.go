package versioning

import (
	"regexp"

	"github.com/pkg/errors"
)

// Version defines a long version used for fully automated versioning
type Version struct {
	Major     string
	Minor     string
	Patch     string
	Timestamp string
	CommitID  string
}

// SplitVersion splits a long version into its parts like major, minor, patch, timestamp, commitId
func SplitVersion(fullVersion string) (Version, error) {

	//pattern for semVer part plus version additions
	semVerRe := regexp.MustCompile(`(\d+)\.?(\d+)?\.?(\d+)?(\S+)?`)
	semVer := semVerRe.FindSubmatch([]byte(fullVersion))
	if semVer == nil {
		return Version{}, errors.Errorf("invalid version: %v", fullVersion)
	}
	newVersion := Version{
		Major: string(semVer[1]),
		Minor: string(semVer[2]),
		Patch: string(semVer[3]),
	}
	if len(newVersion.Minor) == 0 {
		newVersion.Minor = "0"
	}
	if len(newVersion.Patch) == 0 {
		newVersion.Patch = "0"
	}

	if len(semVer[4]) > 0 {
		// pattern for version additions like timestamp and commitId
		extVerRe := regexp.MustCompile(`[-\.]?([\da-zA-Z-]+)[_+\.]?(\S+)?`)
		extVer := extVerRe.FindSubmatch([]byte(string(semVer[4])))
		newVersion.Timestamp = string(extVer[1])
		newVersion.CommitID = string(extVer[2])
	}

	return newVersion, nil
}

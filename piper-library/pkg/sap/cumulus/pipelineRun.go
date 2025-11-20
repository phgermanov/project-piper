package cumulus

import (
	"fmt"
	"regexp"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
)

func (c *Cumulus) GetPipelineRunKey() (pipelineRunKey string, err error) {
	if c.UseCommitIDForCumulus {
		if len(c.HeadCommitID) > 0 {
			pipelineRunKey = c.HeadCommitID
		} else {
			repository, err := c.openGit()
			if err != nil {
				return "", errors.Wrap(err, "failed to open git")
			}
			commitID, err := repository.ResolveRevision(plumbing.Revision("HEAD"))
			if err != nil {
				return "", errors.Wrap(err, "failed to retrieve git commit ID")
			}
			pipelineRunKey = commitID.String()
		}
	} else if c.Scheduled {
		re := regexp.MustCompile(`(\d+\.\d+.\d+)(.)?`)
		separator, version := c.getSeparatorAndVersion(re)
		pipelineRunKey = fmt.Sprintf("%v%v%v", version, separator, c.getSchedulingTimestamp())
		// add the revision to the target path if available
		if len(c.Revision) > 0 {
			pipelineRunKey += fmt.Sprintf("-%v", c.Revision)
		}
	} else {
		pipelineRunKey = c.Version
	}
	return pipelineRunKey, nil
}

func (c *Cumulus) getSeparatorAndVersion(re *regexp.Regexp) (string, string) {
	versionParts := re.FindStringSubmatch(c.Version)

	version := c.Version
	// get version from the first matcher group for major.minor.patch-like versions
	if len(versionParts) > 1 && len(versionParts[1]) > 0 {
		version = versionParts[1]
	}

	separator := "-"
	// get separator from the second matcher group if defined
	if len(versionParts) > 2 && len(versionParts[2]) > 0 {
		separator = versionParts[2]
	}
	return separator, version
}

func (c *Cumulus) getSchedulingTimestamp() string {
	// prepare target path with version and time
	if len(c.SchedulingTimestamp) == 0 {
		c.SchedulingTimestamp = time.Now().Format("20060102")
	}
	return c.SchedulingTimestamp
}

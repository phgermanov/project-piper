package xmake

import (
	"errors"
	"fmt"
	"strings"
)

const jobTypeStagePromote = "StagePromote"

// JobNamePatternInternal ...
const JobNamePatternInternal = "GitHub-Internal"

// JobNamePatternTools ...
const JobNamePatternTools = "GitHub-Tools"

func buildStagePromoteJobName(owner, repository, quality, shipmentType string) string {
	jobName := fmt.Sprintf("%s-%s-SP-%s-common", owner, repository, quality)
	if len(shipmentType) > 0 {
		jobName = jobName + "_" + shipmentType
	}
	return jobName
}

// GetJobName builds the job name for the given parameters.
func GetJobName(jobType, owner, repository, quality, shipmentType, jobNamePattern string) (string, error) {
	if jobType != jobTypeStagePromote {
		return "", fmt.Errorf("job type not supported: %s", jobType)
	}
	if jobNamePattern != JobNamePatternInternal && jobNamePattern != JobNamePatternTools {
		return "", fmt.Errorf("job name pattern not supported: %s", jobNamePattern)
	}
	if len(owner) == 0 {
		return "", errors.New("owner not set")
	}
	if len(repository) == 0 {
		return "", errors.New("repository not set")
	}

	switch quality {
	case "Milestone":
		quality = "MS"
		shipmentType = ""
	case "Release":
		quality = "REL"
	default:
		return "", fmt.Errorf("build quality not supported: %s", quality)
	}

	jobName := buildStagePromoteJobName(owner, repository, quality, strings.TrimSpace(shipmentType))
	if jobNamePattern == JobNamePatternTools {
		jobName = "ght-" + jobName
	}
	return jobName, nil
}

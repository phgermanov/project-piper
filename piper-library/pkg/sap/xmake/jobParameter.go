package xmake

import (
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
)

const (
	paramMode    = "MODE"
	paramTreeish = "TREEISH"
	paramStageID = "STAGING_REPO_ID"
)

const (
	modeStage   = "stage"
	modePromote = "promote"
)

func AggregateBuildParameter(isPromoteBuild bool, commitID, stagingRepositoryID string, jobParameterList []string) (result map[string]string) {
	result = map[string]string{}

	for _, parameter := range jobParameterList {
		parameter = strings.TrimSpace(parameter)
		if !strings.Contains(parameter, "=") {
			log.Entry().Errorf("Skipping job parameter '%s', could not determine key / value", parameter)
			continue
		}
		idx := strings.Index(parameter, "=")
		key := strings.TrimSpace(parameter[0:idx])
		if len(key) == 0 {
			log.Entry().Errorf("Skipping job parameter '%s', empty key", parameter)
			continue
		}
		if key == paramMode || key == paramTreeish || key == paramStageID {
			log.Entry().Errorf("Skipping job parameter '%s', not allowed", parameter)
			continue
		}
		result[key] = strings.TrimSpace(parameter[idx+1:])
	}

	result[paramMode] = modeStage
	result[paramTreeish] = commitID

	if isPromoteBuild {
		result[paramMode] = modePromote
		result[paramStageID] = stagingRepositoryID
	}

	for key, value := range result {
		log.Entry().Infof("Using job parameter '%s' with value '%s'", key, value)
	}

	return
}

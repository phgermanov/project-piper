package cumulus

var PIPELINE_RUN_MODE = struct {
	PR                  string
	STANDARD            string
	OPTIMIZED           string
	OPTIMIZED_SCHEDULED string
}{
	PR:                  "pull-request",
	STANDARD:            "standard",
	OPTIMIZED:           "optimized/commit",
	OPTIMIZED_SCHEDULED: "optimized/scheduled",
}

func GetPipelineRunMode(isPullRequest bool, useCommitID bool, isOptimized bool, isScheduled bool) string {
	if isPullRequest {
		return PIPELINE_RUN_MODE.PR
	}
	if !useCommitID {
		if isOptimized {
			if isScheduled {
				return PIPELINE_RUN_MODE.OPTIMIZED_SCHEDULED
			}
			return PIPELINE_RUN_MODE.OPTIMIZED
		}
	}
	return PIPELINE_RUN_MODE.STANDARD
}

func GetPipelineRunModeSuffix(useCommitID bool, isOptimized bool, isScheduled bool) string {
	if useCommitID && isOptimized && isScheduled {
		return "scheduled"
	}
	return ""
}

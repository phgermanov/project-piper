package dwc

import (
	"fmt"
	"k8s.io/utils/strings/slices"
)

type watchResult interface {
	succeeded() bool
	getStageName() string
}

type StageWatchPolicy func(results []watchResult) error

func OverallSuccessPolicy() StageWatchPolicy {
	return func(results []watchResult) error {
		var failures []string
		for _, wr := range results {
			if !wr.succeeded() {
				failures = append(failures, wr.getStageName())
			}
		}
		if len(failures) > 0 {
			return &stageWatchPolicyViolationError{
				reason: fmt.Sprintf("the deployment to the following stages failed: %v. But all must be successful. Have a look at the deployment logs or consider changing the stageWatchPolicy ", failures),
			}
		}
		return nil
	}
}

func AtLeastOneSuccessfulDeploymentPolicy() StageWatchPolicy {
	return func(results []watchResult) error {
		for _, wr := range results {
			if wr.succeeded() {
				return nil
			}
		}
		return &stageWatchPolicyViolationError{
			reason: "the deployment to all stages failed. But at least one must be successful. Have a look at the deployment logs or consider changing the stageWatchPolicy",
		}
	}
}

func SubsetSuccessPolicy(subsetStages []string) StageWatchPolicy {
	return func(results []watchResult) error {
		var policyViolations []string
		for _, wr := range results {
			if !wr.succeeded() {
				if slices.Contains(subsetStages, wr.getStageName()) {
					policyViolations = append(policyViolations, wr.getStageName())
				}
			}
		}
		if len(policyViolations) > 0 {
			return &stageWatchPolicyViolationError{
				reason: fmt.Sprintf("the deployment to the following stages must be successful %v, but a subset of those failed: %v. Have a look at the deployment logs or consider changing the stageWatchPolicy", subsetStages, policyViolations),
			}
		}
		return nil
	}
}

func AlwaysPassPolicy() StageWatchPolicy {
	return func(results []watchResult) error {
		return nil
	}
}

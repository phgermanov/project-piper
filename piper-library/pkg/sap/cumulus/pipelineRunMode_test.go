package cumulus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPipelineRunMode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		isPullRequest bool
		useCommitID   bool
		isOptimized   bool
		isScheduled   bool
		want          string
	}{
		{false, true, true, true, "standard"},
		{false, true, true, false, "standard"},
		{false, true, false, true, "standard"},
		{false, true, false, false, "standard"},
		{false, false, true, true, "optimized/scheduled"},
		{false, false, true, false, "optimized/commit"},
		{false, false, false, true, "standard"},
		{false, false, false, false, "standard"},
		{true, false, false, false, "pull-request"},
	}
	for _, tc := range testCases {
		assert.Equal(t, tc.want, GetPipelineRunMode(tc.isPullRequest, tc.useCommitID, tc.isOptimized, tc.isScheduled))
	}
}

func TestGetPipelineRunModeSuffix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		useCommitID bool
		isOptimized bool
		isScheduled bool
		want        string
	}{
		{true, true, true, "scheduled"},
		{true, true, false, ""},
		{true, false, true, ""},
		{true, false, false, ""},
		{false, true, true, ""},
		{false, true, false, ""},
		{false, false, true, ""},
		{false, false, false, ""},
	}
	for _, tc := range testCases {
		assert.Equal(t, tc.want, GetPipelineRunModeSuffix(tc.useCommitID, tc.isOptimized, tc.isScheduled))
	}
}

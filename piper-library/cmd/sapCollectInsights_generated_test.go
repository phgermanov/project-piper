//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapCollectInsightsCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapCollectInsightsCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapCollectInsights", testCmd.Use, "command name incorrect")

}

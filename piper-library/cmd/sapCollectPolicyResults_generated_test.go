//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapCollectPolicyResultsCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapCollectPolicyResultsCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapCollectPolicyResults", testCmd.Use, "command name incorrect")

}

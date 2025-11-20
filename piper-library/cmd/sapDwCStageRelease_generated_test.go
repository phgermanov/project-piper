//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapDwCStageReleaseCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapDwCStageReleaseCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapDwCStageRelease", testCmd.Use, "command name incorrect")

}

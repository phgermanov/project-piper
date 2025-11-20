//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapAccessContinuumExecuteTestsCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapAccessContinuumExecuteTestsCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapAccessContinuumExecuteTests", testCmd.Use, "command name incorrect")

}

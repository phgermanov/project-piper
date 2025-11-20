//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapSUPAExecuteTestsCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapSUPAExecuteTestsCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapSUPAExecuteTests", testCmd.Use, "command name incorrect")

}

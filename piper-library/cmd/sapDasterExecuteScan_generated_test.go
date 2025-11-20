//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapDasterExecuteScanCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapDasterExecuteScanCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapDasterExecuteScan", testCmd.Use, "command name incorrect")

}

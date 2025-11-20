//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapURLScanCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapURLScanCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapURLScan", testCmd.Use, "command name incorrect")

}

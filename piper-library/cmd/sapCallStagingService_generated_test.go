//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapCallStagingServiceCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapCallStagingServiceCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapCallStagingService", testCmd.Use, "command name incorrect")

}

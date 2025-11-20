//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapExecuteFastlaneCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapExecuteFastlaneCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapExecuteFastlane", testCmd.Use, "command name incorrect")

}

//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapCallFossServiceCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapCallFossServiceCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapCallFossService", testCmd.Use, "command name incorrect")

}

//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapExecuteCentralPolicyCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapExecuteCentralPolicyCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapExecuteCentralPolicy", testCmd.Use, "command name incorrect")

}

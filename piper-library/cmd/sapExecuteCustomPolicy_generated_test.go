//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapExecuteCustomPolicyCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapExecuteCustomPolicyCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapExecuteCustomPolicy", testCmd.Use, "command name incorrect")

}

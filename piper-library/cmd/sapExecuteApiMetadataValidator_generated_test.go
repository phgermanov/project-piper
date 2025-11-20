//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapExecuteApiMetadataValidatorCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapExecuteApiMetadataValidatorCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapExecuteApiMetadataValidator", testCmd.Use, "command name incorrect")

}

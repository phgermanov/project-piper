//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapPipelineInitCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapPipelineInitCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapPipelineInit", testCmd.Use, "command name incorrect")

}

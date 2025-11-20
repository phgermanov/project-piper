//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapCheckPPMSComplianceCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapCheckPPMSComplianceCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapCheckPPMSCompliance", testCmd.Use, "command name incorrect")

}

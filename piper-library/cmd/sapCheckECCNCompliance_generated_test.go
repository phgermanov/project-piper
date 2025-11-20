//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapCheckECCNComplianceCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapCheckECCNComplianceCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapCheckECCNCompliance", testCmd.Use, "command name incorrect")

}

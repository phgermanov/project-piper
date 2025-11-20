//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapGithubSecretScanningReportCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapGithubSecretScanningReportCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapGithubSecretScanningReport", testCmd.Use, "command name incorrect")

}

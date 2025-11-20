//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapCumulusDownloadCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapCumulusDownloadCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapCumulusDownload", testCmd.Use, "command name incorrect")

}

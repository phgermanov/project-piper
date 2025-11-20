//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapDownloadArtifactCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapDownloadArtifactCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapDownloadArtifact", testCmd.Use, "command name incorrect")

}

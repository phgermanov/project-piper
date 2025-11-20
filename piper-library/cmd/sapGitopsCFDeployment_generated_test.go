//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapGitopsCFDeploymentCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapGitopsCFDeploymentCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapGitopsCFDeployment", testCmd.Use, "command name incorrect")

}

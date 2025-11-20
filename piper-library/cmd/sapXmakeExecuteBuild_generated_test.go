//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapXmakeExecuteBuildCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapXmakeExecuteBuildCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapXmakeExecuteBuild", testCmd.Use, "command name incorrect")

}

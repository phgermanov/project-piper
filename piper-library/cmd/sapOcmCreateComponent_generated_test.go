//go:build unit
// +build unit

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSapOcmCreateComponentCommand(t *testing.T) {
	t.Parallel()

	testCmd := SapOcmCreateComponentCommand()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, "sapOcmCreateComponent", testCmd.Use, "command name incorrect")

}

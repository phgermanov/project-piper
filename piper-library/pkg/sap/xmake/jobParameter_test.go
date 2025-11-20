//go:build unit
// +build unit

package xmake

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAggregateJobParameters(t *testing.T) {
	commitID := "123456"
	stagingID := "ABCDEF"
	t.Run("stage", func(t *testing.T) {
		// init
		isPromote := false
		parameters := []string{}
		// test
		result := AggregateBuildParameter(isPromote, commitID, "", parameters)
		// assert
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "MODE")
		assert.Contains(t, result, "TREEISH")
		assert.Equal(t, "stage", result["MODE"])
		assert.Equal(t, commitID, result["TREEISH"])
		assert.Equal(t, 2, len(result))
	})
	t.Run("promote", func(t *testing.T) {
		// init
		isPromote := true
		parameters := []string{}
		// test
		result := AggregateBuildParameter(isPromote, commitID, stagingID, parameters)
		// assert
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "MODE")
		assert.Contains(t, result, "TREEISH")
		assert.Contains(t, result, "STAGING_REPO_ID")
		assert.Equal(t, "promote", result["MODE"])
		assert.Equal(t, commitID, result["TREEISH"])
		assert.Equal(t, stagingID, result["STAGING_REPO_ID"])
		assert.Equal(t, 3, len(result))
	})
	t.Run("error - parameters not parsable", func(t *testing.T) {
		// init
		isPromote := false
		parameters := []string{"murks"}
		// test
		result := AggregateBuildParameter(isPromote, commitID, "", parameters)
		// assert
		assert.NotEmpty(t, result)
		assert.Equal(t, 2, len(result))
	})
	t.Run("error - no parameter key", func(t *testing.T) {
		// init
		isPromote := false
		parameters := []string{"=murks"}
		// test
		result := AggregateBuildParameter(isPromote, commitID, "", parameters)
		// assert
		assert.NotEmpty(t, result)
		assert.Equal(t, 2, len(result))
	})
	t.Run("error - illegal parameter key", func(t *testing.T) {
		// init
		isPromote := false
		parameters := []string{"MODE=murks"}
		// test
		result := AggregateBuildParameter(isPromote, commitID, "", parameters)
		// assert
		assert.NotEmpty(t, result)
		assert.Equal(t, 2, len(result))
	})
	t.Run("parameters", func(t *testing.T) {
		// init
		isPromote := false
		parameters := []string{"VERSION_EXTENSION=123456"}
		// test
		result := AggregateBuildParameter(isPromote, commitID, "", parameters)
		// assert
		assert.NotEmpty(t, result)
		assert.Equal(t, 3, len(result))
		assert.Contains(t, result, "VERSION_EXTENSION")
		assert.Equal(t, "123456", result["VERSION_EXTENSION"])
	})
	t.Run("parameters - buildoptions", func(t *testing.T) {
		// init
		isPromote := false
		parameters := []string{"BUILD_OPTIONS=--buildplugin-option \"options='--build-arg oq_tests=true --build-arg unit_tests=false'\""}
		// test
		result := AggregateBuildParameter(isPromote, commitID, "", parameters)
		// assert
		assert.NotEmpty(t, result)
		assert.Equal(t, 3, len(result))
		assert.Contains(t, result, "BUILD_OPTIONS")
		assert.Equal(t, "--buildplugin-option \"options='--build-arg oq_tests=true --build-arg unit_tests=false'\"", result["BUILD_OPTIONS"])
	})
}

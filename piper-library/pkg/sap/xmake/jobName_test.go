//go:build unit
// +build unit

package xmake

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBuildStagePromoteJobName(t *testing.T) {
	// init
	owner := "Piper-Validation"
	repository := "Golang"
	quality := "MS"
	shipmentType := ""
	expectedJobName := "Piper-Validation-Golang-SP-MS-common"

	// test
	result := buildStagePromoteJobName(owner, repository, quality, shipmentType)
	// asserts
	assert.Equal(t, expectedJobName, result)
}

func TestGetJobName(t *testing.T) {
	t.Run("unsupported job type", func(t *testing.T) {
		// init
		// test
		result, err := GetJobName(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, JobNamePatternInternal)
		// asserts
		assert.Empty(t, result)
		assert.EqualError(t, err, "job type not supported: mock.Anything")
	})

	t.Run("unsupported build quality", func(t *testing.T) {
		// init
		// test
		result, err := GetJobName(jobTypeStagePromote, mock.Anything, mock.Anything, mock.Anything, mock.Anything, JobNamePatternInternal)
		// asserts
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Equal(t, "build quality not supported: mock.Anything", err.Error())
	})

	t.Run("unsupported job name pattern", func(t *testing.T) {
		// init
		quality := "Milestone"
		// test
		result, err := GetJobName(jobTypeStagePromote, mock.Anything, mock.Anything, quality, mock.Anything, mock.Anything)
		// asserts
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Equal(t, "job name pattern not supported: mock.Anything", err.Error())
	})

	t.Run("missing owner", func(t *testing.T) {
		// init
		// test
		result, err := GetJobName(jobTypeStagePromote, "", mock.Anything, mock.Anything, mock.Anything, JobNamePatternInternal)
		// asserts
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Equal(t, "owner not set", err.Error())
	})

	t.Run("missing owner", func(t *testing.T) {
		// init
		// test
		result, err := GetJobName(jobTypeStagePromote, mock.Anything, "", mock.Anything, mock.Anything, JobNamePatternInternal)
		// asserts
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Equal(t, "repository not set", err.Error())
	})

	t.Run("milestone job", func(t *testing.T) {
		// init
		owner := "Piper-Validation"
		repository := "Golang"
		quality := "Milestone"
		shipmentType := "cloud"
		expectedJobName := "Piper-Validation-Golang-SP-MS-common"
		// test
		result, err := GetJobName(jobTypeStagePromote, owner, repository, quality, shipmentType, JobNamePatternInternal)
		// asserts
		assert.NoError(t, err)
		assert.Equal(t, expectedJobName, result)
	})

	t.Run("milestone job with ghtool pattern", func(t *testing.T) {
		// init
		owner := "Piper-Validation"
		repository := "Golang"
		quality := "Milestone"
		shipmentType := "cloud"
		jobNamePattern := "GitHub-Tools"
		expectedJobName := "ght-Piper-Validation-Golang-SP-MS-common"
		// test
		result, err := GetJobName(jobTypeStagePromote, owner, repository, quality, shipmentType, jobNamePattern)
		// asserts
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Equal(t, expectedJobName, result)
	})

	t.Run("release job", func(t *testing.T) {
		// init
		owner := "Piper-Validation"
		repository := "Golang"
		quality := "Release"
		shipmentType := "cloud"
		expectedJobName := "Piper-Validation-Golang-SP-REL-common_cloud"
		// test
		result, err := GetJobName(jobTypeStagePromote, owner, repository, quality, shipmentType, JobNamePatternInternal)
		// asserts
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Equal(t, expectedJobName, result)
	})
}

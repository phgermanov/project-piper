//go:build unit
// +build unit

package cumulus

import (
	"fmt"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/stretchr/testify/assert"
)

func TestToFile(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		files := mock.FilesMock{}
		expectedFile := fmt.Sprintf("release-status-%v.json", now.Format("20060102150405"))
		s := ReleaseStatus{Status: "promoted"}

		err := s.ToFile(&files, now)
		assert.NoError(t, err)

		exists, err := files.FileExists(expectedFile)
		assert.NoError(t, err)
		assert.True(t, exists)

		content, err := files.FileRead(expectedFile)
		assert.NoError(t, err)
		assert.Equal(t, `{"releaseStatus":"promoted"}`, string(content))
	})

	t.Run("error - write file", func(t *testing.T) {
		now := time.Now()
		files := mock.FilesMock{FileWriteError: fmt.Errorf("write error")}
		expectedFile := fmt.Sprintf("release-status-%v.json", now.Format("20060102150405"))
		s := ReleaseStatus{Status: "promoted"}

		err := s.ToFile(&files, now)
		assert.EqualError(t, err, fmt.Sprintf("failed to write %v: write error", expectedFile))
	})
}

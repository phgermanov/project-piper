package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentExecutor(t *testing.T) {
	t.Parallel()
	t.Run("Can create default instance", func(t *testing.T) {
		t.Parallel()

		executor := DefaultExecutor()

		assert.NotNil(t, executor)
	})
}

package ocm

import (
	"encoding/json"
	"github.com/SAP/jenkins-library/cmd"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStagingRepoInfo(t *testing.T) {
	info := NewComponentInfo()

	// Error test: set RepoInfo with wrong type
	info.Set("RepoInfo", "not-a-repo-info")
	// Should not panic, but RepoInfo should return a new StagingRepoInfo (empty)
	assert.Equal(t, "", info.RepoInfo().RepositoryURL)

	repoJson := `{"password":"secret","repository":"369000938081-20240912-112918305-572","user":"me","repositoryURL":"369000938081-20240912-112918305-572.staging.repositories.cloud.sap"}`
	var ri StagingRepoInfo
	err := json.Unmarshal([]byte(repoJson), &ri)
	assert.NoError(t, err)
	info.Set("RepoInfo", &ri)
	assert.Equal(t, "369000938081-20240912-112918305-572.staging.repositories.cloud.sap", info.RepoInfo().RepositoryURL)
}

func TestGet(t *testing.T) {
	info := NewComponentInfo()
	// numeric values should not be returned as empty string
	info["num"] = 123
	assert.Equal(t, "", info.Get("num"))
	// string values should be returned correctly
	info["foo"] = "bar"
	assert.Equal(t, "bar", info.Get("foo"))
}

func TestHas(t *testing.T) {
	info := NewComponentInfo()
	// nothing set should return false
	assert.Equal(t, false, info.Has("num"))
	// key with empty string should return false
	info["foo"] = ""
	assert.Equal(t, false, info.Has("foo"))
}

func TestStage(t *testing.T) {
	tests := []struct {
		name      string
		stageName string
		expected  PiperStage
	}{
		{
			name:      "Returns Build for Central Build",
			stageName: "Central Build",
			expected:  Build,
		},
		{
			name:      "Returns Build for Build",
			stageName: "Build",
			expected:  Build,
		},
		{
			name:      "Returns Other for Deploy",
			stageName: "Deploy",
			expected:  Other,
		},
		{
			name:      "Returns Other for empty string",
			stageName: "",
			expected:  Other,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := cmd.GeneralConfig.StageName
			cmd.GeneralConfig.StageName = tt.stageName
			defer func() { cmd.GeneralConfig.StageName = original }()
			assert.Equal(t, tt.expected, Stage())
		})
	}
}

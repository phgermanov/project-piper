//go:build unit
// +build unit

package dwc

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResolveVectorMetadataBuildEntry(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		createCollector func() MetadataCollector
		want            any
		wantErr         bool
	}{
		{
			name: "canonical collection case",
			want: vectorMetadataBuildEntry{
				Branch:              "main",
				JobUrl:              "someJobURL",
				BuildUrl:            "someBuildURL",
				BuildNumber:         "someBuildNumber",
				GitCommitId:         "someCommitID",
				GithubInstance:      "github.tools.sap",
				GithubRepository:    "deploy-with-confidence/sponde",
				GithubRepositoryUrl: "https://github.tools.sap/deploy-with-confidence/sponde",
				Orchestrator:        "Jenkins",
			},
			wantErr: false,
			createCollector: func() MetadataCollector {
				collector := &mockMetadataCollector{}
				collector.On("GetMetadataEntry", gitBranchMetadataEntry).Return("main", nil)
				collector.On("GetMetadataEntry", jobURLMetadataEntry).Return("someJobURL", nil)
				collector.On("GetMetadataEntry", buildURLMetadataEntry).Return("someBuildURL", nil)
				collector.On("GetMetadataEntry", buildIDMetadataEntry).Return("someBuildNumber", nil)
				collector.On("GetMetadataEntry", githubRepoURLMetadataEntry).Return("https://github.tools.sap/deploy-with-confidence/sponde", nil)
				collector.On("GetMetadataEntry", commitIdMetadataEntry).Return("someCommitID", nil)
				collector.On("GetMetadataEntry", orchestratorTypeMetadataEntry).Return("Jenkins", nil)
				return collector
			},
		},
		{
			name:    "error while resolving some metadata entry leads to an error",
			want:    "",
			wantErr: true,
			createCollector: func() MetadataCollector {
				collector := &mockMetadataCollector{}
				collector.On("GetMetadataEntry", gitBranchMetadataEntry).Return("main", nil)
				collector.On("GetMetadataEntry", jobURLMetadataEntry).Return("someJobURL", nil)
				collector.On("GetMetadataEntry", buildURLMetadataEntry).Return("", errors.New("some error occurred"))
				collector.On("GetMetadataEntry", buildIDMetadataEntry).Return("someBuildNumber", nil)
				collector.On("GetMetadataEntry", githubRepoURLMetadataEntry).Return("https://github.tools.sap/deploy-with-confidence/sponde", nil)
				collector.On("GetMetadataEntry", commitIdMetadataEntry).Return("someCommitID", nil)
				collector.On("GetMetadataEntry", orchestratorTypeMetadataEntry).Return("Jenkins", nil)
				return collector
			},
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got, err := resolveVectorMetadataBuildEntry(testCase.createCollector())
			if (err != nil) != testCase.wantErr {
				t.Fatalf("resolveVectorMetadataBuildEntry() error = %v, wantErr %v", err, testCase.wantErr)
			}
			var wanted string
			if !testCase.wantErr {
				wantedBytes, err := json.Marshal(testCase.want)
				if err != nil {
					t.Fatalf("error during preparation period. Failed to marshal test data: %v", err)
				}
				wanted = string(wantedBytes)
			}
			assert.Equal(t, wanted, got)
		})
	}
}

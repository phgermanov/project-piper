//go:build unit
// +build unit

package dwc

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigResolver_ResolveDefaultResourceName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		createCollector func() MetadataCollector
		resolver        ConfigResolver
		want            string
		wantErr         bool
	}{
		{
			name: "canonical resolve",
			createCollector: func() MetadataCollector {
				collector := &mockMetadataCollector{}
				collector.On("GetMetadataEntry", gitBranchMetadataEntry).Return("main", nil)
				collector.On("GetMetadataEntry", githubRepoURLMetadataEntry).Return("https://github.tools.sap/deploy-with-confidence/sponde", nil)
				return collector
			},
			resolver: ConfigResolver{},
			want:     "github.tools.sap/deploy-with-confidence/sponde/main",
			wantErr:  false,
		},
		{
			name: "invalid githubRepoURLMetadataEntry leads to error",
			createCollector: func() MetadataCollector {
				collector := &mockMetadataCollector{}
				collector.On("GetMetadataEntry", gitBranchMetadataEntry).Return("main", nil)
				collector.On("GetMetadataEntry", githubRepoURLMetadataEntry).Return("https://github.tools.sap/deploy-with-confidence", nil)
				return collector
			},
			resolver: ConfigResolver{},
			want:     "",
			wantErr:  true,
		},
		{
			name: "failing to resolve gitBranchMetadataEntry leads to error",
			createCollector: func() MetadataCollector {
				collector := &mockMetadataCollector{}
				collector.On("GetMetadataEntry", gitBranchMetadataEntry).Return("", errors.New("some error occurred"))
				collector.On("GetMetadataEntry", githubRepoURLMetadataEntry).Return("https://github.tools.sap/deploy-with-confidence", nil)
				return collector
			},
			resolver: ConfigResolver{},
			want:     "",
			wantErr:  true,
		},
		{
			name: "failing to resolve githubRepoURLMetadataEntry leads to error",
			createCollector: func() MetadataCollector {
				collector := &mockMetadataCollector{}
				collector.On("GetMetadataEntry", gitBranchMetadataEntry).Return("main", nil)
				collector.On("GetMetadataEntry", githubRepoURLMetadataEntry).Return("", errors.New("some error occurred"))
				return collector
			},
			resolver: ConfigResolver{},
			want:     "",
			wantErr:  true,
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got, err := testCase.resolver.ResolveDefaultResourceName(testCase.createCollector())
			if (err != nil) != testCase.wantErr {
				t.Fatalf("ResolveDefaultResourceName() error = %v, wantErr %v", err, testCase.wantErr)
			}
			assert.Equal(t, testCase.want, got)
		})
	}
}

func TestConfigResolver_ResolveContainerImageURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		promotedDockerImage string
		artifactVersion     string
		resolver            ConfigResolver
		wantImage           string
		wantTag             string
		wantErr             bool
	}{
		{
			name:                "canonical image with tag",
			promotedDockerImage: "bp.common.repositories.cloud.sap/dwc-cli:latest",
			artifactVersion:     "",
			resolver:            ConfigResolver{},
			wantImage:           "bp.common.repositories.cloud.sap/dwc-cli",
			wantTag:             "latest",
			wantErr:             false,
		},
		{
			name:                "image with tag and l4 port spec",
			promotedDockerImage: "docker.wdf.sap.corp:51011/some-org/flaky-image:deprecated",
			artifactVersion:     "ihoo7ahC",
			resolver:            ConfigResolver{},
			wantImage:           "docker.wdf.sap.corp:51011/some-org/flaky-image",
			wantTag:             "deprecated",
			wantErr:             false,
		},
		{
			name:                "image without tag, but with the provided artifactVersion and l4 port spec",
			promotedDockerImage: "docker.wdf.sap.corp:51011/dwc-cli",
			artifactVersion:     "ihoo7ahC",
			resolver:            ConfigResolver{},
			wantImage:           "docker.wdf.sap.corp:51011/dwc-cli",
			wantTag:             "ihoo7ahC",
			wantErr:             false,
		},
		{
			name:                "image without tag, but with the provided artifactVersion",
			promotedDockerImage: "bp.common.repositories.cloud.sap/dwc-cli",
			artifactVersion:     "ihoo7ahC",
			resolver:            ConfigResolver{},
			wantImage:           "bp.common.repositories.cloud.sap/dwc-cli",
			wantTag:             "ihoo7ahC",
			wantErr:             false,
		},
		{
			name:                "image without tag, fails with missing artifactVersion",
			promotedDockerImage: "bp.common.repositories.cloud.sap/dwc-cli",
			artifactVersion:     "",
			resolver:            ConfigResolver{},
			wantImage:           "",
			wantTag:             "",
			wantErr:             true,
		},
		{
			name:                "no image and tag provided fails",
			promotedDockerImage: "",
			artifactVersion:     "ihoo7ahC",
			resolver:            ConfigResolver{},
			wantImage:           "",
			wantTag:             "",
			wantErr:             true,
		},
		{
			name:                "image containing slash fails",
			promotedDockerImage: "docker.wdf.sap.corp:51011/some-org/flaky-image:deprecated/stuff",
			artifactVersion:     "ihoo7ahC",
			resolver:            ConfigResolver{},
			wantImage:           "",
			wantTag:             "",
			wantErr:             true,
		},
	}
	for _, c := range tests {
		testCase := c
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			gotImage, gotTag, err := testCase.resolver.ResolveContainerImageURL(testCase.promotedDockerImage, testCase.artifactVersion)
			if (err != nil) != testCase.wantErr {
				t.Fatalf("ResolveContainerImageURL() error = %v, wantErr %v", err, testCase.wantErr)
			}
			assert.Equal(t, testCase.wantImage, gotImage)
			assert.Equal(t, testCase.wantTag, gotTag)
		})
	}
}

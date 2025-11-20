package binary

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func pString(s string) *string {
	return &s
}

type mockExecutor struct {
	mock.Mock
}

func (m *mockExecutor) do(pathToBinary string, executorFactory BlockingExecutorFactory, targetCmd devopsInsightsCommand) error {
	args := m.Called(pathToBinary, executorFactory, targetCmd)
	return args.Error(0)
}

type mockReleaseGetter struct {
	mock.Mock
}

func (m *mockReleaseGetter) do(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error) {
	args := m.Called(ctx, owner, repo)
	return (args.Get(0).(*github.RepositoryRelease)), (args.Get(1).(*github.Response)), args.Error(2)
}

func TestCLIBinary_InstallDevOpsInsightsCli(t *testing.T) {
	tests := []struct {
		name            string
		version         string
		cliReleaseError error
		downloadError   error
		chmodError      error
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name:            "succeeds without version",
			version:         "",
			cliReleaseError: nil,
			downloadError:   nil,
			chmodError:      nil,
			wantErr:         assert.NoError,
		},
		{
			name:            "succeeds without version - space in version",
			version:         "    ",
			cliReleaseError: nil,
			downloadError:   nil,
			chmodError:      nil,
			wantErr:         assert.NoError,
		},
		{
			name:            "succeeds with version",
			version:         "v1.0.0",
			cliReleaseError: nil,
			downloadError:   nil,
			chmodError:      nil,
			wantErr:         assert.NoError,
		},
		{
			name:            "fails when cli release url returns error",
			version:         "",
			cliReleaseError: fmt.Errorf("error"),
			downloadError:   nil,
			chmodError:      nil,
			wantErr:         assert.Error,
		},
		{
			name:            "fails when download fails",
			version:         "",
			cliReleaseError: nil,
			downloadError:   fmt.Errorf("error"),
			chmodError:      nil,
			wantErr:         assert.Error,
		},
		{
			name:            "fails when chmod fails",
			version:         "",
			cliReleaseError: nil,
			downloadError:   nil,
			chmodError:      fmt.Errorf("error"),
			wantErr:         assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const pathToBin = "path/to/binary"
			const url = "https://test.com"
			mockReleaseResolver := NewMockCliReleaseResolver(t)
			mockFileDownloader := newMockFileDownloader(t)
			mockFilePermissionEditor := newMockFilePermissionEditor(t)
			resolver := &CLIBinary{
				FileDownloader:     mockFileDownloader,
				PermEditor:         mockFilePermissionEditor,
				CLIReleaseResolver: mockReleaseResolver,
			}

			if strings.TrimSpace(tt.version) != "" {
				mockReleaseResolver.EXPECT().GetReleaseVersionURL(tt.version).Return(url, tt.cliReleaseError)
			} else {
				mockReleaseResolver.EXPECT().GetLatestReleaseURL().Return(url, tt.cliReleaseError)
			}

			if tt.cliReleaseError == nil {
				mockReleaseResolver.EXPECT().CreateHTTPHeader().Return(nil)
				mockFileDownloader.EXPECT().DownloadFile(url, pathToBin, mock.Anything, mock.Anything).Return(tt.downloadError)
				if tt.downloadError == nil {
					mockFilePermissionEditor.EXPECT().Chmod(pathToBin, mock.Anything).Return(tt.chmodError)
				}
			}

			err := resolver.InstallDevOpsInsightsCli(pathToBin, tt.version)
			tt.wantErr(t, err)
		})
	}
}

func TestCLIBinary_ExecuteDevOpInsights(t *testing.T) {
	type args struct {
		targetBinary string
		command      []string
	}
	tests := []struct {
		name        string
		executorErr error
		assertErr   assert.ErrorAssertionFunc
	}{
		{
			name:        "success",
			executorErr: nil,
			assertErr:   assert.NoError,
		},
		{
			name:        "failed",
			executorErr: fmt.Errorf("error"),
			assertErr:   assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const binPath = "/bin/exe"
			command := devopsInsightsCommand{""}
			m := new(mockExecutor)
			m.On("do", binPath, mock.Anything, command).Return(tt.executorErr)
			resolver := CLIBinary{CLIExecutor: m.do}
			err := resolver.ExecuteDevOpInsights(binPath, command)
			m.AssertCalled(t, "do", binPath, mock.Anything, command)
			tt.assertErr(t, err)
		})
	}
}

func TestNewArtifactoryReleaseResolver(t *testing.T) {
	tests := []struct {
		name      string
		want      CliReleaseResolver
		assertErr assert.ErrorAssertionFunc
	}{
		{
			name:      "success",
			assertErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewArtifactoryReleaseResolver()
			tt.assertErr(t, err)
		})
	}
}

func TestNewGithubToolsReleaseResolver(t *testing.T) {
	type args struct {
		githubToken string
		githubURL   string
	}
	tests := []struct {
		name      string
		args      args
		assertErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			args: args{
				githubToken: "token",
				githubURL:   "test.com",
			},
			assertErr: assert.NoError,
		},
		{
			name: "error empty token",
			args: args{
				githubToken: "",
				githubURL:   "test.com",
			},
			assertErr: assert.Error,
		},
		{
			name: "error invalid url",
			args: args{
				githubToken: "token",
				githubURL:   "%$#@!test.com",
			},
			assertErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newGithubToolsReleaseResolver(tt.args.githubToken, tt.args.githubURL)
			tt.assertErr(t, err)
			if err != nil {
				return
			}
			assert.Equal(t, tt.args.githubToken, got.token)
		})
	}
}

func Test_artifactoryReleaseResolver_CreateHTTPHeader(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "success"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &artifactoryReleaseResolver{}
			assert.NotNil(t, resolver.CreateHTTPHeader())
		})
	}
}

func Test_artifactoryReleaseResolver_getAssetURL(t *testing.T) {
	type args struct {
		path string
		goos string
		arch string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success linux",
			args: args{
				path: "latest",
				goos: "linux",
				arch: "amd64",
			},
			want: "https://int.repositories.cloud.sap/artifactory/devops-insights-generic/bin/latest/devops-insights.linux.amd64",
		},
		{
			name: "success windows",
			args: args{
				path: "latest",
				goos: "windows",
				arch: "amd64",
			},
			want: "https://int.repositories.cloud.sap/artifactory/devops-insights-generic/bin/latest/devops-insights.windows.amd64.exe",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &artifactoryReleaseResolver{}
			got := resolver.getAssetURL(tt.args.path, tt.args.goos, tt.args.arch)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_githubToolsReleaseResolver_CreateHTTPHeader(t *testing.T) {
	type fields struct {
		token string
	}
	tests := []struct {
		name  string
		token string
		want  http.Header
	}{
		{
			name:  "success",
			token: "token",
			want: http.Header{
				"Authorization": []string{"token token"},
				"Accept":        []string{"application/octet-stream"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &githubToolsReleaseResolver{
				token: tt.token,
			}
			got := resolver.CreateHTTPHeader()
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_githubToolsReleaseResolver_getAssetURL(t *testing.T) {
	asset := func(name string, url ...string) *github.ReleaseAsset {
		var pUrl *string = nil
		if len(url) > 0 {
			pUrl = pString(strings.Join(url, ""))
		}
		return &github.ReleaseAsset{
			URL:  pUrl,
			Name: pString(name),
		}
	}
	type args struct {
		goos   string
		goarch string
	}
	tests := []struct {
		name       string
		args       args
		want       string
		release    *github.RepositoryRelease
		releaseErr error
		assertErr  assert.ErrorAssertionFunc
	}{
		{
			name: "success linux",
			args: args{
				goos:   "linux",
				goarch: "amd64",
			},
			release: &github.RepositoryRelease{
				Assets: []*github.ReleaseAsset{
					asset("devops-insights.linux.amd64", "test.com/linux"),
					asset("devops-insights.windows.amd64.exe", "test.com/windows"),
				},
			},
			want:      "test.com/linux",
			assertErr: assert.NoError,
		},
		{
			name: "success windows",
			args: args{
				goos:   "windows",
				goarch: "amd64",
			},
			release: &github.RepositoryRelease{
				Assets: []*github.ReleaseAsset{
					asset("devops-insights.linux.amd64", "test.com/linux"),
					asset("devops-insights.windows.amd64.exe", "test.com/windows"),
				},
			},
			want:      "test.com/windows",
			assertErr: assert.NoError,
		},
		{
			name: "fail getting release",
			args: args{
				goos:   "linux",
				goarch: "amd64",
			},
			releaseErr: fmt.Errorf("error"),
			assertErr:  assert.Error,
		},
		{
			name: "fail release is nil",
			args: args{
				goos:   "linux",
				goarch: "amd64",
			},
			assertErr: assert.Error,
		},
		{
			name: "fail no applicable asset found",
			args: args{
				goos:   "darwin",
				goarch: "amd64",
			},
			release: &github.RepositoryRelease{
				Assets: []*github.ReleaseAsset{
					asset("devops-insights.linux.amd64", "test.com/linux"),
					asset("devops-insights.windows.amd64.exe", "test.com/windows"),
				},
			},
			assertErr: assert.Error,
		},
		{
			name: "fail no applicable asset found",
			args: args{
				goos:   "linux",
				goarch: "amd64",
			},
			release: &github.RepositoryRelease{
				Assets: []*github.ReleaseAsset{
					asset("devops-insights.linux.amd64"),
					asset("devops-insights.windows.amd64.exe", "test.com/windows"),
				},
			},
			assertErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &githubToolsReleaseResolver{
				ctx: context.Background(),
			}
			m := new(mockReleaseGetter)
			m.On("do", resolver.ctx, githubOwner, githubDevOpsInsightsCLIRepo).Return(tt.release, &github.Response{}, tt.releaseErr)
			got, err := resolver.getAssetURL(tt.args.goos, tt.args.goarch, m.do)
			tt.assertErr(t, err)
			if err != nil {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

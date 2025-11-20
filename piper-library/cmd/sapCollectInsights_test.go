//go:build unit
// +build unit

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	piperOsCmd "github.com/SAP/jenkins-library/cmd"
	"github.com/SAP/jenkins-library/pkg/telemetry"

	"github.com/stretchr/testify/assert"
)

type sapCollectInsightsUtilsMock struct {
	mockFileWrite                func(path string, content []byte, perm os.FileMode) error
	mockInstallDevOpsInsightsCli func(binPath, version string) error
	mockExecuteDevOpInsights     func(binPath string, command []string) error
	mockTempDir                  func(string, string) (string, error)
}

func (i *sapCollectInsightsUtilsMock) FileWrite(path string, content []byte, perm os.FileMode) error {
	return i.mockFileWrite(path, content, perm)
}
func (i *sapCollectInsightsUtilsMock) InstallDevOpsInsightsCli(binPath, version string) error {
	return i.mockInstallDevOpsInsightsCli(binPath, version)
}
func (i *sapCollectInsightsUtilsMock) ExecuteDevOpInsights(binPath string, command []string) error {
	return i.mockExecuteDevOpInsights(binPath, command)
}
func (i *sapCollectInsightsUtilsMock) TempDir(dir string, pattern string) (string, error) {
	return i.mockTempDir(dir, pattern)
}

type rwc struct {
	*bytes.Buffer
}

func (mwc *rwc) Close() error {
	return nil
}

func Test_writeConfig(t *testing.T) {
	type args struct {
		utils    sapCollectInsightsUtils
		config   sapCollectInsightsOptions
		fileName string
	}
	type testCase struct {
		result     *[]byte
		name       string
		args       args
		wantErr    bool
		wantConfig *string
	}
	tests := []testCase{
		func() testCase {
			result := &[]byte{}
			resultConfig := `devopsinsightstoken: token
dorasystemtrusttoken: trust_token
tokenorder: i2
devopsinsightsapi: api
githubtoken: gittoken
commitid: id
previousreleasecommitid: prev_id
branch: b
gitorganization: gitorg
gitrepository: gitrepo
gitinstance: gitin
giturl: ""
artifactversion: artv
identifier: ident
deploymenttarget: depltarget
serializeintermediateresults: false
collectchangesetproduction: true
deploymenttime: depltime
changesetretrieval: chsr
devopsinsightsversion: dev
gcsbucketid: gcsbid
`
			return testCase{
				name:   "success",
				result: result,
				args: args{
					utils: &sapCollectInsightsUtilsMock{
						mockFileWrite: func(path string, content []byte, perm os.FileMode) error {
							assert.Equal(t, "file.yaml", path)
							*result = content
							return nil
						},
					},
					config: sapCollectInsightsOptions{
						CommitID:                   "id",
						PreviousReleaseCommitID:    "prev_id",
						Branch:                     "b",
						CollectChangeSetProduction: true,
						DevOpsInsightsToken:        "token",
						DoraSystemTrustToken:       "trust_token",
						TokenOrder:                 "i2",
						DevOpsInsightsAPI:          "api",
						GithubToken:                "gittoken",
						GitOrganization:            "gitorg",
						GitRepository:              "gitrepo",
						GitInstance:                "gitin",
						ArtifactVersion:            "artv",
						Identifier:                 "ident",
						DeploymentTarget:           "depltarget",
						DeploymentTime:             "depltime",
						ChangeSetRetrieval:         "chsr",
						DevOpsInsightsVersion:      "dev",
						GcsBucketID:                "gcsbid",
					},
					fileName: "file.yaml",
				},
				wantErr:    false,
				wantConfig: &resultConfig,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := writeConfig(tt.args.utils, tt.args.config, tt.args.fileName); (err != nil) != tt.wantErr {
				t.Errorf("writeConfig() error = %v, wantErr %v", err, tt.wantErr)
			} else if err != nil {
				return
			}
			assert.Equal(t, *tt.wantConfig, string(*tt.result))
		})
	}
}

func Test_runSapCollectInsights(t *testing.T) {
	type args struct {
		config        *sapCollectInsightsOptions
		telemetryData *telemetry.CustomData
		utils         sapCollectInsightsUtils
	}
	tests := []struct {
		name      string
		args      args
		wantErr   assert.ErrorAssertionFunc
		verbosity bool
	}{
		{
			name: "standard success case",
			args: args{
				config: &sapCollectInsightsOptions{
					DevOpsInsightsToken: "token",
				},
				telemetryData: nil,
				utils: &sapCollectInsightsUtilsMock{
					mockInstallDevOpsInsightsCli: func(binPath, version string) error {
						assert.Equal(t, filepath.Join("temp-dora", "devops-insights"), binPath)
						return nil
					},
					mockExecuteDevOpInsights: func(binPath string, command []string) error {
						assert.Equal(t, filepath.Join("temp-dora", "devops-insights"), binPath)
						assert.Equal(t, []string{"--config", filepath.Join("temp-dora", "insights.yaml")}, command)
						return nil
					},
					mockTempDir: func(root string, prefix string) (string, error) {
						assert.Equal(t, ".", root)
						assert.Equal(t, "temp-", prefix)
						return root + "/" + prefix + "dora", nil
					},
					mockFileWrite: func(path string, content []byte, perm os.FileMode) error {
						assert.Equal(t, filepath.Join("temp-dora", "insights.yaml"), path)
						assert.Equal(t, string(content), `devopsinsightstoken: token
dorasystemtrusttoken: ""
tokenorder: ""
devopsinsightsapi: ""
githubtoken: ""
commitid: ""
previousreleasecommitid: ""
branch: ""
gitorganization: ""
gitrepository: ""
gitinstance: ""
giturl: ""
artifactversion: ""
identifier: ""
deploymenttarget: ""
serializeintermediateresults: false
collectchangesetproduction: false
deploymenttime: ""
changesetretrieval: ""
devopsinsightsversion: ""
gcsbucketid: ""
`)
						return nil
					},
				},
			},
			wantErr:   assert.NoError,
			verbosity: false,
		},
		{
			name: "success - with verbose mode forwarded",
			args: args{
				config: &sapCollectInsightsOptions{
					DevOpsInsightsToken: "token",
				},
				telemetryData: nil,
				utils: &sapCollectInsightsUtilsMock{
					mockInstallDevOpsInsightsCli: func(binPath, version string) error {
						assert.Equal(t, filepath.Join("temp-dora", "devops-insights"), binPath)
						return nil
					},
					mockExecuteDevOpInsights: func(binPath string, command []string) error {
						assert.Equal(t, filepath.Join("temp-dora", "devops-insights"), binPath)
						assert.Equal(t, []string{"--config", filepath.Join("temp-dora", "insights.yaml"), "--verbose", "true"}, command)
						return nil
					},
					mockTempDir: func(root string, prefix string) (string, error) {
						assert.Equal(t, ".", root)
						assert.Equal(t, "temp-", prefix)
						return root + "/" + prefix + "dora", nil
					},
					mockFileWrite: func(path string, content []byte, perm os.FileMode) error {
						assert.Equal(t, filepath.Join("temp-dora", "insights.yaml"), path)
						assert.Equal(t, string(content), `devopsinsightstoken: token
dorasystemtrusttoken: ""
tokenorder: ""
devopsinsightsapi: ""
githubtoken: ""
commitid: ""
previousreleasecommitid: ""
branch: ""
gitorganization: ""
gitrepository: ""
gitinstance: ""
giturl: ""
artifactversion: ""
identifier: ""
deploymenttarget: ""
serializeintermediateresults: false
collectchangesetproduction: false
deploymenttime: ""
changesetretrieval: ""
devopsinsightsversion: ""
gcsbucketid: ""
`)
						return nil
					},
				},
			},
			wantErr:   assert.NoError,
			verbosity: true,
		},
		{
			name: "failing early with no token available",
			args: args{
				config: &sapCollectInsightsOptions{
					ArtifactVersion: "1.0",
				},
				telemetryData: nil,
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.verbosity {
				piperOsCmd.GeneralConfig.Verbose = true
			}
			tt.wantErr(t, runSapCollectInsights(tt.args.config, tt.args.telemetryData, tt.args.utils), fmt.Sprintf("runSapCollectInsights(%v, %v, %v)", tt.args.config, tt.args.telemetryData, tt.args.utils))
		})
	}
}

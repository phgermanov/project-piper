//go:build unit
// +build unit

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	piperOsCmd "github.com/SAP/jenkins-library/cmd"
	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/SAP/jenkins-library/pkg/log"
	pipertelemetry "github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sapPipelineInitMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newSapPipelineInitTestsUtils() sapPipelineInitMockUtils {
	utils := sapPipelineInitMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func Test_getGithubDetails(t *testing.T) {
	type gitMetadata struct {
		gitBranch       string
		gitInstance     string
		gitURL          string
		gitOrganization string
		gitRepository   string
		gitCloneURL     string
		gitCommitID     string
		gitApiURL       string
	}
	tests := []struct {
		isFailure bool
		name      string
		gitInfos  gitMetadata
	}{
		{
			name: "Success Case: clone from github.wdf.sap.corp via ssh ",
			gitInfos: gitMetadata{
				gitBranch:       "master",
				gitInstance:     "github.wdf.sap.corp",
				gitURL:          "https://github.wdf.sap.corp/test-org/test-repo",
				gitOrganization: "test-org",
				gitRepository:   "test-repo",
				gitCloneURL:     "git@github.wdf.sap.corp:test-org/test-repo.git",
				gitApiURL:       "https://github.wdf.sap.corp/api/v3/",
			},
			isFailure: false,
		},
		{
			name: "Success Case: clone from github.tools.sap via https ",
			gitInfos: gitMetadata{
				gitBranch:       "master",
				gitInstance:     "github.tools.sap",
				gitURL:          "https://github.tools.sap/test-org/test-repo",
				gitOrganization: "test-org",
				gitRepository:   "test-repo",
				gitCloneURL:     "https://github.tools.sap/test-org/test-repo.git",
				gitApiURL:       "https://github.tools.sap/api/v3/",
			},
			isFailure: false,
		},
		{
			name: "Success Case: github.com",
			gitInfos: gitMetadata{
				gitBranch:       "master",
				gitInstance:     "github.com",
				gitURL:          "https://github.com/test-org/test-repo",
				gitOrganization: "test-org",
				gitRepository:   "test-repo",
				gitCloneURL:     "https://github.com/test-org/test-repo",
				gitApiURL:       "https://api.github.com/",
			},
			isFailure: false,
		},
		{
			name: "Failure Case: git repository does not exist",
			gitInfos: gitMetadata{
				gitBranch:       "",
				gitInstance:     "",
				gitURL:          "",
				gitOrganization: "",
				gitRepository:   "",
				gitCommitID:     "",
				gitApiURL:       "",
			},
			isFailure: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "githubDetailsTest")
			if err != nil {
				t.Fatal("Failed to create temporary directory")
			}
			currCWD, _ := os.Getwd()
			_ = os.Chdir(tempDir)
			if tt.isFailure == false {
				repository, err := git.PlainInit(tempDir, false)
				if err != nil {
					t.Fatal("error creating temp repo", err)
				}
				// Add a new remote, with the default fetch refspec
				_, err = repository.CreateRemote(&config.RemoteConfig{
					Name: "origin",
					URLs: []string{tt.gitInfos.gitCloneURL},
				})
				if err != nil {
					t.Fatal("error creating remote", err)
				}
				// Opens an already existing repository.
				repository, err = git.PlainOpen(".")
				if err != nil {
					t.Fatal("could not read git repo", err)
				}
				// opens a worktree
				w, err := repository.Worktree()
				if err != nil {
					t.Fatal("could not read worktree", err)
				}
				// Create a new file inside the worktree of the project
				//["echo \"hello world!\" > example-git-file"]
				filename := filepath.Join("", "example-git-file")
				err = os.WriteFile(filename, []byte("hello world!"), 0644)
				if err != nil {
					t.Fatal("error:", err)
				}
				// Adds the new file to the staging area.
				//["git add example-git-file"]
				_, err = w.Add("example-git-file")
				if err != nil {
					t.Fatal("error:", err)
				}
				// Commit the staging area to the repository, with the new file
				//["git commit -m \"example go-git commit\""]
				commit, err := w.Commit("example go-git commit", &git.CommitOptions{
					Author: &object.Signature{Name: "testcase user", Email: "piper+noreply@sap.com"},
				})
				if err != nil {
					t.Fatal("error:", err)
				}
				// Prints the current HEAD to verify that all worked well.
				//["git show -s"]
				obj, err := repository.CommitObject(commit)
				if err != nil {
					t.Fatal("error:", err)
				}
				log.Entry().Info(obj)
			}
			//call the function
			gitInfo, err := getGitDetails()
			if err != nil {
				log.Entry().WithError(err).Warning("Cannot get Github Details")
			}
			assert.Equal(t, gitInfo.gitURL, tt.gitInfos.gitURL)
			assert.Equal(t, gitInfo.gitRepository, tt.gitInfos.gitRepository)
			assert.Equal(t, gitInfo.gitOrganization, tt.gitInfos.gitOrganization)
			assert.Equal(t, gitInfo.gitBranch, tt.gitInfos.gitBranch)
			assert.Equal(t, gitInfo.gitInstance, tt.gitInfos.gitInstance)
			assert.Equal(t, gitInfo.gitApiURL, tt.gitInfos.gitApiURL)
			if tt.isFailure == false {
				assert.Regexp(t, "([a-f0-9]{40})", gitInfo.gitCommitID)
			} else {
				assert.Equal(t, tt.gitInfos.gitCommitID, gitInfo.gitCommitID)
			}
			// change to pwd
			_ = os.Chdir(currCWD)
			// clean up tmp dir
			_ = os.RemoveAll(tempDir)
		})
	}
}

func TestRunSapPipelineInit(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		pipelineConfig := sapPipelineInitOptions{}
		pipelineEnv := sapPipelineInitCommonPipelineEnvironment{}
		influx := sapPipelineInitInflux{}
		telemetry := pipertelemetry.CustomData{}

		// set MetaDataResolver
		piperOsCmd.GeneralConfig.MetaDataResolver = GetAllStepMetadata

		tempDir, err := os.MkdirTemp("", "")
		require.NoError(t, err, "could not get tempDir")

		err = os.Chdir(tempDir)
		require.NoError(t, err, "could not get change current working directory")

		err = os.Mkdir(".pipeline", 0777)
		require.NoError(t, err, "could not create .pipeline folder")

		// Start prepare dummy config.yml file
		file, err := os.Create(".pipeline/config.yml")
		require.NoError(t, err, "could not create .pipeline/config.yml file")

		// test
		err = runSapPipelineInit(&pipelineConfig, &telemetry, &pipelineEnv, &influx)
		// assert
		assert.NoError(t, err)

		err = os.RemoveAll(file.Name())
		require.NoError(t, err, "could not remove: ", file.Name())

		// Influx
		assert.Equal(t, "SUCCESS", influx.jenkins_custom_data.fields.build_result)
		assert.Equal(t, 1, influx.jenkins_custom_data.fields.build_result_key)
		assert.NotEmpty(t, influx.pipeline_data.fields.build_url)
		assert.NotEmpty(t, influx.step_data.fields.build_url)
		// PiperEnv
		assert.Equal(t, false, pipelineEnv.custom.scheduledRun)
		// Telemetry
		assert.Equal(t, false, telemetry.IsScheduled)
	})
}

func Test_runSapPipelineInit(t *testing.T) {
	type args struct {
		config        sapPipelineInitOptions
		telemetryData pipertelemetry.CustomData
		pipelineEnv   sapPipelineInitCommonPipelineEnvironment
		influx        sapPipelineInitInflux
	}
	type gitMetadata struct {
		gitBranch       string
		gitInstance     string
		gitURL          string
		gitOrganization string
		gitRepository   string
		gitCloneURL     string
		gitCommitID     string
		gitApiUrl       string
	}
	tests := []struct {
		gitInfos  gitMetadata
		name      string
		args      args
		wantErr   assert.ErrorAssertionFunc
		isFailure bool
		isCumulus bool
	}{
		{name: "success case everything available",
			gitInfos: gitMetadata{
				gitInstance:     "github.wdf.sap.corp",
				gitURL:          "https://github.wdf.sap.corp/test-org/test-repo",
				gitOrganization: "test-org",
				gitRepository:   "test-repo",
				gitCloneURL:     "git@github.wdf.sap.corp:test-org/test-repo.git",
				gitApiUrl:       "https://github.wdf.sap.corp/api/v3/",
			},
			isFailure: false,
			isCumulus: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isCumulus == false {
				tempDir, err := os.MkdirTemp("", "githubDetailsTest")
				if err != nil {
					t.Fatal("Failed to create temporary directory")
				}
				currCWD, _ := os.Getwd()
				_ = os.Chdir(tempDir)
				if tt.isFailure == false {
					repository, err := git.PlainInit(tempDir, false)
					if err != nil {
						t.Fatal("error creating temp repo", err)
					}
					// Add a new remote, with the default fetch refspec
					_, err = repository.CreateRemote(&config.RemoteConfig{
						Name: "origin",
						URLs: []string{tt.gitInfos.gitCloneURL},
					})
					if err != nil {
						t.Fatal("error creating remote", err)
					}
					// Opens an already existing repository.
					repository, err = git.PlainOpen(".")
					if err != nil {
						t.Fatal("could not read git repo", err)
					}
					// opens a worktree
					w, err := repository.Worktree()
					if err != nil {
						t.Fatal("could not read worktree", err)
					}
					// Create a new file inside the worktree of the project
					//["echo \"hello world!\" > example-git-file"]
					filename := filepath.Join("", "example-git-file")
					err = os.WriteFile(filename, []byte("hello world!"), 0644)
					if err != nil {
						t.Fatal("error:", err)
					}
					// Adds the new file to the staging area.
					//["git add example-git-file"]
					_, err = w.Add("example-git-file")
					if err != nil {
						t.Fatal("error:", err)
					}
					// Commit the staging area to the repository, with the new file
					//["git commit -m \"example go-git commit\""]
					commit, err := w.Commit("example go-git commit", &git.CommitOptions{
						Author: &object.Signature{Name: "testcase user", Email: "piper+noreply@sap.com"},
					})
					if err != nil {
						t.Fatal("error:", err)
					}
					// Prints the current HEAD to verify that all worked well.
					//["git show -s"]
					obj, err := repository.CommitObject(commit)
					if err != nil {
						t.Fatal("error:", err)
					}
					log.Entry().Info(obj)
				}

				// set MetaDataResolver
				piperOsCmd.GeneralConfig.MetaDataResolver = GetAllStepMetadata

				err = runSapPipelineInit(&tt.args.config, &tt.args.telemetryData, &tt.args.pipelineEnv, &tt.args.influx)
				if err != nil {
					log.Entry().WithError(err).Fatal("step execution failed")
				}
				assert.Equal(t, tt.args.pipelineEnv.git.url, tt.gitInfos.gitURL)
				assert.Equal(t, tt.args.pipelineEnv.git.repository, tt.gitInfos.gitRepository)
				assert.Equal(t, tt.args.pipelineEnv.git.organization, tt.gitInfos.gitOrganization)
				assert.Equal(t, tt.args.pipelineEnv.git.instance, tt.gitInfos.gitInstance)
				if tt.isFailure == false {
					assert.Regexp(t, "([a-f0-9]{40})", tt.args.pipelineEnv.git.commitID)
				} else {
					assert.Equal(t, tt.gitInfos.gitCommitID, tt.args.pipelineEnv.git.commitID)
				}
				// change to pwd
				_ = os.Chdir(currCWD)
				// clean up tmp dir
				_ = os.RemoveAll(tempDir)
			} else {

				tempDir, err := os.MkdirTemp("", "")
				require.NoError(t, err, "could not get tempDir")

				err = os.Chdir(tempDir)
				require.NoError(t, err, "could not get change current working directory")

				err = os.Mkdir(".pipeline", 0777)
				require.NoError(t, err, "could not create .pipeline folder")

				// Start prepare dummy config.yml file
				file, err := os.Create(".pipeline/config.yml")
				require.NoError(t, err, "could not create .pipeline/config.yml file")

				// set MetaDataResolver
				piperOsCmd.GeneralConfig.MetaDataResolver = GetAllStepMetadata

				// End preparation of dummy file
				err = runSapPipelineInit(&tt.args.config, &tt.args.telemetryData, &tt.args.pipelineEnv, &tt.args.influx)
				if err != nil {
					log.Entry().WithError(err).Fatal("step execution failed")
				}
				assert.Regexp(t, "([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})", tt.args.pipelineEnv.custom.cumulusPipelineID)
				assert.Regexp(t, "([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})", tt.args.pipelineEnv.custom.gcsBucketID)
				err = os.RemoveAll(file.Name())
				require.NoError(t, err, "could not remove: ", file.Name())
			}
		})
	}
}

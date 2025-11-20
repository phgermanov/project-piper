//go:build unit
// +build unit

package cmd

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/bndr/gojenkins"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/jenkins"
	"github.com/SAP/jenkins-library/pkg/jenkins/mocks"
	pipermock "github.com/SAP/jenkins-library/pkg/mock"
	StepResults "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/xmake"
)

const responseJobFinder = `{
	"_class":"com.sap.prd.jenkins.plugins.jobfinder.JobFinder",
	"job":[{
		"name":"piper-validation-golang-SP-MS-common",
		"fullName":"piper-validation/piper-validation-golang-SP-MS-common",
		"url":"https://xmake-dev.wdf.sap.corp/job/piper-validation/job/piper-validation-golang-SP-MS-common/",
		"landscape":"xmake-dev",
		"jenkinsUrl":"https://xmake-dev.wdf.sap.corp"
	}]
}`

// jenkins.Artifact Mock
type artifactMock struct {
}

func (a *artifactMock) GetData(ctx context.Context) ([]byte, error) {
	var jsonContent = "{\"JOB_FAILURE\": [\"https://nova-validation.wdf.sap.corp:443/job/xmake-landscape-validation/job/xmake-landscape-validation-docker-java-sample-SP-MS-docker_rhel-docker_rhel/17/consoleFull#L237\"],\"downstreams\": {\"job1\":\"jobUrl1\",\"job2\":\"jobUrl2\"},\"BUILDRESULTS\": [\"CFG-TCLO-0000\",\"The GAV already exists in the Nexus target repository. Provide a non-existing GAV\"]}"
	return []byte(jsonContent), nil
}
func (a *artifactMock) FileName() string {
	return "artifactMockFileName"
}
func (a *artifactMock) Save(ctx context.Context, path string) (bool, error) {
	return true, nil
}
func (a *artifactMock) SaveToDir(ctx context.Context, dir string) (bool, error) {
	return true, nil
}

func TestRunSapXmakeExecuteBuild(t *testing.T) {
	t.Run("failed - jenkins connection", func(t *testing.T) {
		//init parameters (all blank structures)
		ctx := context.Background()
		config := &sapXmakeExecuteBuildOptions{
			Owner:          "Piper-Validation",
			Repository:     "Golang",
			BuildQuality:   "Milestone",
			JobNamePattern: xmake.JobNamePatternInternal,
		}
		client := &piperhttp.Client{}
		// required to use httpmock (DefaultTransport)
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		commonPipelineEnvironment := &sapXmakeExecuteBuildCommonPipelineEnvironment{}
		// reduce waiting time
		jenkinsConnectWaitingTime = time.Nanosecond
		jobFinder := xmake.JobFinder{Client: client, Username: "user", Token: "token", RetryInterval: time.Nanosecond}
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "https://xmake-nova.wdf.sap.corp/job_finder/api/json?input=Piper-Validation-Golang-SP-MS-common",
			httpmock.NewStringResponder(http.StatusOK, responseJobFinder),
		)
		//TODO: replace with object mock
		// We mock the jenkins.Instance function to return an empty structure
		jenkinsInstanceFunc = func(ctx context.Context, client *http.Client, jenkinsURL, user, token string) (*gojenkins.Jenkins, error) {
			return &gojenkins.Jenkins{}, errors.New("mock error")
		}
		fileUtils := &pipermock.FilesMock{}
		// test
		err := runSapXmakeExecuteBuild(ctx, config, client, jobFinder, fileUtils, commonPipelineEnvironment)
		// asserts
		assert.Equal(t, 1, httpmock.GetTotalCallCount())
		assert.Contains(t, err.Error(), "Failed to connect to Jenkins with url 'https://xmake-dev.wdf.sap.corp': mock error")
	})

	t.Run("failed - FetchArtifact", func(t *testing.T) {
		//init parameters (all blank structures)
		ctx := context.Background()
		config := &sapXmakeExecuteBuildOptions{
			Owner:          "Piper-Validation",
			Repository:     "Golang",
			BuildQuality:   "Milestone",
			BuildType:      typePromote,
			JobNamePattern: xmake.JobNamePatternInternal,
		}
		client := &piperhttp.Client{}
		// required to use httpmock (DefaultTransport)
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		commonPipelineEnvironment := &sapXmakeExecuteBuildCommonPipelineEnvironment{}
		// reduce waiting time
		jenkinsConnectWaitingTime = time.Nanosecond
		jenkinsWaitForBuildToFinishWaitTime = time.Nanosecond
		fetchBuildResultArtifactWaitingTime = time.Nanosecond
		jobFinder := xmake.JobFinder{Client: client, Username: "user", Token: "token", RetryInterval: time.Nanosecond}
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "https://xmake-nova.wdf.sap.corp/job_finder/api/json?input=Piper-Validation-Golang-SP-MS-common",
			httpmock.NewStringResponder(http.StatusOK, responseJobFinder),
		)

		//TODO: replace with object mock
		// We mock the jenkins.Instance function to return an empty structure
		jenkinsInstanceFunc = func(ctx context.Context, client *http.Client, jenkinsURL, user, token string) (*gojenkins.Jenkins, error) {
			return &gojenkins.Jenkins{}, nil
		}

		//TODO: replace with object mock
		jenkinsGetJobFunc = func(ctx context.Context, _jenkins jenkins.Jenkins, jobName string) (jenkins.Job, error) {
			return &jenkins.JobImpl{}, nil
		}

		//TODO: replace with object mock
		jenkinsTriggerJobFunc = func(ctx context.Context, _jenkins jenkins.Jenkins, job jenkins.Job, parameters map[string]string) (*gojenkins.Build, error) {
			buildResponse := &gojenkins.BuildResponse{Result: "SUCCESS", URL: "dummyURL"}
			return &gojenkins.Build{Raw: buildResponse}, nil
		}

		//TODO: replace with object mock
		jenkinsWaitForBuildToFinishFunc = func(ctx context.Context, build jenkins.Build, pollInterval time.Duration) error {
			return nil
		}

		//TODO: replace with object mock
		// We mock the xmake.FetchBuildResultJSON function to return an empty Artifact structure
		xmakeFetchBuildResultJSON = func(ctx context.Context, build jenkins.Build) (jenkins.Artifact, error) {
			return nil, errors.New("mock error")
		}

		//TODO: replace with object mock
		// We mock the xmake.xmakeFetchStageJSON function to return an empty staged json struct
		xmakeFetchStageJSON = func(ctx context.Context, artifact jenkins.Artifact, _ string) (xmake.StageJSON, error) {
			return xmake.StageJSON{}, nil
		}

		//TODO: replace with object mock
		// We mock the xmake.xmakefetchPromoteJSON function to return an empty promoted json struct
		xmakeFetchPromoteJSON = func(ctx context.Context, artifact jenkins.Artifact, _ string) (xmake.PromoteJSON, error) {
			return xmake.PromoteJSON{PromoteBom: &xmake.PromoteBom{Repositories: []*xmake.Repository{}}}, nil
		}

		fileUtils := &pipermock.FilesMock{}
		// test
		err := runSapXmakeExecuteBuild(ctx, config, client, jobFinder, fileUtils, commonPipelineEnvironment)
		// asserts
		assert.NoError(t, err)
		assert.Equal(t, 1, httpmock.GetTotalCallCount())
	})
}

func TestGetXmakeBuildErrorMessages(t *testing.T) {

	t.Run("fetch from error_dictionary.json", func(t *testing.T) {

		//init parameters (all blank structures)
		ctx := context.Background()

		downstreambuildMock := &gojenkins.Build{Raw: &gojenkins.BuildResponse{Number: 1, URL: "downstreamJobURL"}, Job: &gojenkins.Job{Raw: &gojenkins.JobResponse{Name: "downstreamjobName"}}}
		xmakeGetJenkinsBuildFromUrl = func(ctx context.Context, jenkinsInstance *gojenkins.Jenkins, url string) (*gojenkins.Build, error) {
			return downstreambuildMock, nil
		}

		//return mock of error_dictionary.json artifact with BUILDRESULTS
		xmakeFetchBuildArtifact = func(ctx context.Context, build jenkins.Build, fileName string) (jenkins.Artifact, error) {
			return &artifactMock{}, nil
		}

		buildMock := &gojenkins.Build{Raw: &gojenkins.BuildResponse{Number: 1, URL: "jobURL"}, Job: &gojenkins.Job{Raw: &gojenkins.JobResponse{Name: "jobName"}}}
		consoleErrorLinks := []string{}
		consoleErrorLinks = append(consoleErrorLinks, "https://nova-validation.wdf.sap.corp:443/job/xmake-landscape-validation/job/xmake-landscape-validation-docker-java-sample-SP-MS-docker_rhel-docker_rhel/17/consoleFull#L237")

		errorMessages := getXmakeBuildErrorMessages(ctx, &gojenkins.Jenkins{}, buildMock, consoleErrorLinks)

		assert.Equal(t, len(errorMessages), 1)
		assert.Equal(t, errorMessages["downstreamjobName"], "CFG-TCLO-0000:The GAV already exists in the Nexus target repository. Provide a non-existing GAV")

	})

	t.Run("fetch rule from console log", func(t *testing.T) {

		//init parameters (all blank structures)
		ctx := context.Background()

		downstreambuildMock := &gojenkins.Build{Raw: &gojenkins.BuildResponse{Number: 1, URL: "downstreamJobURL"}, Job: &gojenkins.Job{Raw: &gojenkins.JobResponse{Name: "downstreamjobName"}}}
		xmakeGetJenkinsBuildFromUrl = func(ctx context.Context, jenkinsInstance *gojenkins.Jenkins, url string) (*gojenkins.Build, error) {
			return downstreambuildMock, nil
		}

		//return mock of no error_dictionary.json found
		xmakeFetchBuildArtifact = func(ctx context.Context, build jenkins.Build, fileName string) (jenkins.Artifact, error) {
			return nil, errors.New("artifact not found")
		}

		//Return mocked console output containing a line matching a rule regex
		xmakeGetBuildConsoleOutput = func(ctx context.Context, build *gojenkins.Build) string {
			return "Job completed. Result was FAILURE\r\nlog line"
		}

		buildMock := &gojenkins.Build{Raw: &gojenkins.BuildResponse{Number: 1, URL: "jobURL"}, Job: &gojenkins.Job{Raw: &gojenkins.JobResponse{Name: "jobName"}}}

		consoleErrorLinks := []string{}
		consoleErrorLinks = append(consoleErrorLinks, "https://nova-validation.wdf.sap.corp:443/job/xmake-landscape-validation/job/xmake-landscape-validation-docker-java-sample-SP-MS-docker_rhel-docker_rhel/17/consoleFull#L237")

		errorMessages := getXmakeBuildErrorMessages(ctx, &gojenkins.Jenkins{}, buildMock, consoleErrorLinks)

		assert.Equal(t, len(errorMessages), 1)
		assert.Equal(t, errorMessages["downstreamjobName"], "USR-JJEN-000: Please check the xmake downstream build log for details")

	})
}

func TestHandleFailedBuild(t *testing.T) {
	t.Run("TestHandleFailedBuild", func(t *testing.T) {

		//init parameters (all blank structures)
		ctx := context.Background()
		fileUtils := &pipermock.FilesMock{}

		failedBuildResponse := &gojenkins.BuildResponse{Result: "FAILURE", URL: "dummyURL"}
		failedBuildMock := &gojenkins.Build{Raw: failedBuildResponse, Job: &gojenkins.Job{Raw: &gojenkins.JobResponse{Name: "jobName"}}}

		//return mock of no error_dictionary.json found
		xmakeFetchBuildArtifact = func(ctx context.Context, build jenkins.Build, fileName string) (jenkins.Artifact, error) {
			return nil, errors.New("artifact not found")
		}

		//Return mocked console output containing a line matching a rule regex
		xmakeGetBuildConsoleOutput = func(ctx context.Context, build *gojenkins.Build) string {
			return "Job completed. Result was FAILURE\r\nlog line"
		}

		handleFailedBuild(ctx, &gojenkins.Jenkins{}, failedBuildMock, &artifactMock{}, fileUtils)

	})
}

func TestGetDownstreamBuildsFromArtifact(t *testing.T) {
	t.Run("get builds from artifact", func(t *testing.T) {

		//init parameters (all blank structures)
		ctx := context.Background()

		xmakeGetJenkinsBuildFromUrl = func(ctx context.Context, jenkinsInstance *gojenkins.Jenkins, url string) (*gojenkins.Build, error) {
			return &gojenkins.Build{}, nil
		}

		downstreamBuilds := getFailedDownstreamBuildsFromArtifact(ctx, &gojenkins.Jenkins{}, &artifactMock{})

		assert.Equal(t, len(downstreamBuilds), 2)

	})
}

func TestGetJenkinsBuildFromUrl(t *testing.T) {

	t.Run("get jenkins build from url", func(t *testing.T) {

		//init parameters (all blank structures)
		ctx := context.Background()

		xmakeGetJenkinsInstanceBuild = func(ctx context.Context, jenkinsInstance *gojenkins.Jenkins, jobName string, buildNumber int64) (*gojenkins.Build, error) {
			return &gojenkins.Build{}, nil
		}

		build, err := getJenkinsBuildFromUrl(ctx, &gojenkins.Jenkins{}, "https://nova-validation.wdf.sap.corp:443/job/xmake-landscape-validation/job/xmake-landscape-validation-piper-test-pipeline-SP-MS-linuxx86_64-linux2/91")

		assert.Nil(t, err)
		assert.NotNil(t, build)

	})
}
func TestFindErrorRule(t *testing.T) {
	t.Run("find error rule from json", func(t *testing.T) {

		consoleOutput := "Job #222 completed. Result was FAILURE"
		errorMessage := findErrorRule(consoleOutput)

		assert.Equal(t, errorMessage, "USR-JJEN-000: Please check the xmake downstream build log for details")

	})
}

func TestFetchBuildResultArtifact(t *testing.T) {
	t.Run("failure - no build artifacts", func(t *testing.T) {
		// restore
		//TODO: remove once mocking is in place everywhere
		xmakeFetchBuildResultJSON = xmake.FetchBuildResultJSON
		// init
		ctx := context.Background()
		build := &mocks.Build{}
		build.
			On("IsRunning", ctx).Return(false, nil).
			On("GetArtifacts").Return([]gojenkins.Artifact{}, nil)
		// reduce waiting time
		fetchBuildResultArtifactWaitingTime = time.Nanosecond
		// test
		result := fetchBuildResultArtifact(ctx, build, mock.Anything)
		// asserts
		build.AssertExpectations(t)
		assert.Nil(t, result)
	})
}

func TestWritePipelineEnvironment(t *testing.T) {
	t.Run("success - two repositories", func(t *testing.T) {
		// restore
		//TODO: remove once mocking is in place everywhere
		xmakeFetchPromoteJSON = xmake.FetchPromoteJSON
		// init
		ctx := context.Background()
		pipelineEnv := &sapXmakeExecuteBuildCommonPipelineEnvironment{}
		options := &sapXmakeExecuteBuildOptions{BuildType: typePromote}
		artifact := &mocks.Artifact{}
		artifact.
			On("GetData", ctx).Return([]byte(`{
				"downstreams": {
				  "anything": "nothing"
				},
				"promote-bom": {
				  "group": "group-20220503-1009090772-49",
				  "released": true,
				  "repositories": [
					{
						"repository": "9da44bb18081-20220614-173444773-293",
						"result": [
							"something"
						],
						"success": true
					},
					{
						"repository": "a1682b9a8081-20220614-173450657-903",
						"result": [
							"anything"
						],
						"success": true
					}
				  ]
				}
			  }`), nil)
		// test
		err := writePipelineEnvironment(ctx, artifact, pipelineEnv, options)
		// asserts
		artifact.AssertExpectations(t)
		assert.Contains(t, pipelineEnv.custom.promotedArtifactURLs, "something")
		assert.Contains(t, pipelineEnv.custom.promotedArtifactURLs, "anything")
		assert.Equal(t, 2, len(pipelineEnv.custom.promotedArtifactURLs))
		assert.Equal(t, `{"releaseStatus":"promoted"}`, pipelineEnv.custom.releaseStatus)
		assert.NoError(t, err)
	})
}

func TestJenkinsWaitForBuildToFinishFunc(t *testing.T) {

	t.Parallel()

	t.Run("build finished successfully", func(t *testing.T) {
		ctx := context.Background()
		jenkinsBuild := &mocks.Build{}
		jenkinsWaitForBuildToFinishWaitTime := 100 * time.Millisecond
		var jenkinsWaitForBuildToFinishFunc = jenkins.WaitForBuildToFinish
		//Test the case where the build finishes successfully
		jenkinsBuild.On("IsRunning", ctx).Return(false)
		err := jenkinsWaitForBuildToFinishFunc(ctx, jenkinsBuild, jenkinsWaitForBuildToFinishWaitTime)
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	})

	t.Run("error - fetching resilt returns EOF exception", func(t *testing.T) {
		ctx := context.Background()
		jenkinsBuild := &mocks.Build{}
		jenkinsWaitForBuildToFinishWaitTime := 1 * time.Second
		var jenkinsWaitForBuildToFinishFunc = jenkins.WaitForBuildToFinish
		//Test the case where the build has an error result
		jenkinsBuild.
			On("IsRunning", ctx).Return(true).
			On("Poll", ctx).Return(200, errors.New("EOF"))

		err := jenkinsWaitForBuildToFinishFunc(ctx, jenkinsBuild, jenkinsWaitForBuildToFinishWaitTime)
		if err == nil {
			t.Errorf("Expected an error, but got nil")
		}

		expectedErr := errors.New("Max retries (4) exceeded while waiting for build to finish. Last error: EOF")
		if err.Error() != expectedErr.Error() {
			t.Errorf("Expected error '%v', but got '%v'", expectedErr, err)
		}
	})
}

func mockWriteSBomXmlForStageBuild(stageBOM map[string]interface{}) ([][]byte, error) {
	// Define a sample data list to return
	dataList := [][]byte{[]byte("sample data")}

	return dataList, nil
}

func TestWriteSbomFiles(t *testing.T) {
	// Set up a mock config object and pipelineEnv object for testing purposes
	config := &sapXmakeExecuteBuildOptions{
		Owner:          "Piper-Validation",
		Repository:     "Golang",
		BuildQuality:   "Milestone",
		BuildType:      typeStage,
		JobNamePattern: xmake.JobNamePatternInternal,
	}
	pipelineEnv := &sapXmakeExecuteBuildCommonPipelineEnvironment{}

	// Replace the writeSBomXmlForStageBuild variable with the mock function
	writeSBomXmlForStageBuild = mockWriteSBomXmlForStageBuild

	// Create some sample StepResults.Paths to pass to the function
	reports := []StepResults.Path{
		{Target: "foo", Mandatory: true},
		{Target: "bar", Mandatory: false},
	}

	// Call the function
	newReports := writeSbomFiles(config, pipelineEnv, reports)

	// Check that the newReports slice contains the expected StepResults.Path object(s)
	expectedReport := StepResults.Path{Target: "**/sbom/**/*", Mandatory: false}
	if len(newReports) != len(reports)+1 {
		t.Errorf("Expected newReports slice length to be %d, but got %d", len(reports)+1, len(newReports))
	} else if !containsPath(newReports, expectedReport) {
		t.Errorf("Expected newReports to contain %v, but got %v", expectedReport, newReports)
	}
}

// Helper function to check if a StepResults.Path object is contained in a slice of StepResults.Path objects
func containsPath(paths []StepResults.Path, path StepResults.Path) bool {
	for _, p := range paths {
		if p.Target == path.Target && p.Mandatory == path.Mandatory {
			return true
		}
	}
	return false
}

func mockWriteSBomXmlForStageBuildError(stageBOM map[string]interface{}) ([][]byte, error) {
	return nil, errors.New("mock error")
}

func TestWriteSbomFiles_Error(t *testing.T) {
	config := &sapXmakeExecuteBuildOptions{
		BuildType: typeStage,
	}
	pipelineEnv := &sapXmakeExecuteBuildCommonPipelineEnvironment{}
	reports := []StepResults.Path{}

	// Replace the writeSBomXmlForStageBuild variable with the error mock
	orig := writeSBomXmlForStageBuild
	writeSBomXmlForStageBuild = mockWriteSBomXmlForStageBuildError
	defer func() { writeSBomXmlForStageBuild = orig }()

	// Optionally, capture log output here if you want to assert on the warning

	newReports := writeSbomFiles(config, pipelineEnv, reports)

	if len(newReports) != len(reports) {
		t.Errorf("Expected reports slice length to remain unchanged, but got %d", len(newReports))
	}
}

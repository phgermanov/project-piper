package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bndr/gojenkins"
	"github.com/pkg/errors"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/jenkins"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	StepResults "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/cumulus"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/xmake"
)

const (
	typeStage   = "xMakeStage"
	typePromote = "xMakePromote"
)

// Sender provides an interface to the piper http client for uid/pwd and token authenticated requests
type xmakeSender interface {
	xmake.Sender
	StandardClient() *http.Client
}

// Used to reduce wait time during tests
var fetchBuildResultArtifactWaitingTime = 3 * time.Second

// Used to reduce wait time during tests
var jenkinsConnectWaitingTime = 3 * time.Second

// Used to reduce wait time during tests
var jenkinsWaitForBuildToFinishWaitTime = 15 * time.Second

// vars used to mock functions
var jenkinsInstanceFunc = jenkins.Instance
var jenkinsGetJobFunc = jenkins.GetJob
var jenkinsTriggerJobFunc = jenkins.TriggerJob
var jenkinsWaitForBuildToFinishFunc = jenkins.WaitForBuildToFinish
var xmakeFetchBuildResultJSON = xmake.FetchBuildResultJSON
var xmakeFetchBuildArtifact = jenkins.FetchBuildArtifact
var xmakeFetchStageJSON = xmake.FetchStageJSON
var xmakeFetchPromoteJSON = xmake.FetchPromoteJSON
var writeSBomXmlForStageBuild = xmake.WriteSBomXmlForStageBuild
var xmakeGetJenkinsBuildFromUrl = getJenkinsBuildFromUrl
var xmakeGetBuildConsoleOutput = getBuildConsoleOutput
var xmakeFetchBuildResultArtifact = fetchBuildResultArtifact
var xmakeGetJenkinsInstanceBuild = getJenkinsInstanceBuild

func sapXmakeExecuteBuild(config sapXmakeExecuteBuildOptions, telemetry *telemetry.CustomData, commonPipelineEnvironment *sapXmakeExecuteBuildCommonPipelineEnvironment, influx *sapXmakeExecuteBuildInflux) {

	ctx := context.Background()
	client := &piperhttp.Client{}
	jobFinder := xmake.JobFinder{Client: client, Username: config.Username, Token: config.Token, RetryInterval: 3 * time.Second}

	telemetry.BuildType = config.BuildType
	telemetry.BuildQuality = config.BuildQuality
	telemetry.LegacyJobNameTemplate = config.XMakeJobNameTemplate
	telemetry.LegacyJobName = config.XMakeJobName

	config.JobNamePattern = getJobNamePatternWithRespectToLegacyParameters(config)

	influx.step_data.fields.build_quality = config.BuildQuality
	if config.BuildType == typePromote {
		influx.step_data.fields.xmakepromote = "false"
	} else {
		influx.step_data.fields.xmakestage = "false"
	}

	fileUtils := piperutils.Files{}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	if err := runSapXmakeExecuteBuild(ctx, &config, client, jobFinder, fileUtils, commonPipelineEnvironment); err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
	if config.BuildType == typePromote {
		influx.step_data.fields.xmakepromote = "true"
	} else {
		influx.step_data.fields.xmakestage = "true"
	}
}

func runSapXmakeExecuteBuild(ctx context.Context, config *sapXmakeExecuteBuildOptions, client xmakeSender, jobFinder xmake.JobFinder, fileUtils piperutils.FileUtils, pipelineEnv *sapXmakeExecuteBuildCommonPipelineEnvironment) error {

	//Set default error category - displayed in piper logs
	log.SetErrorCategory(log.ErrorConfiguration)

	// build job name
	jobName, jobNameErr := xmake.GetJobName("StagePromote", config.Owner, config.Repository, config.BuildQuality, config.ShipmentType, config.JobNamePattern)
	if jobNameErr != nil {
		return errors.Wrap(jobNameErr, "failed to construct job name")
	}
	log.Entry().Debugf("Build Job: %s", jobName)

	// find xmake job
	jobInfo, err := findJob(jobFinder, jobName)
	if err != nil {
		return err
	}

	//set empty return variables for jenkinsInstanceFunc
	jenkinsInstance, err := connectToJenkins(ctx, config, client, jobInfo)
	if err != nil {
		return err
	}

	// trigger job
	log.Entry().Infof("Triggering Jenkins Job: %s", jobInfo.URL)
	pipelineEnv.custom.xmakeJobURL = jobInfo.URL

	job, jobErr := jenkinsGetJobFunc(ctx, jenkinsInstance, jobInfo.FullName)
	if jobErr != nil {
		return errors.Wrapf(jobErr, "failed to load job '%s'", jobInfo.FullName)
	}
	// collect parameters
	jobParameters := xmake.AggregateBuildParameter(config.BuildType == typePromote, config.CommitID, config.StagingRepositoryID, config.JobParameters)
	jenkinsBuild, buildTriggerErr := jenkinsTriggerJobFunc(ctx, jenkinsInstance, job, jobParameters)

	if buildTriggerErr != nil {
		errorMessage := handleTriggerErrorMessage(buildTriggerErr, jobInfo.FullName)
		return errors.Wrap(buildTriggerErr, errorMessage)
	}

	log.Entry().Infof("Build started: %s", jenkinsBuild.GetUrl())
	pipelineEnv.custom.xmakeJobURL = jenkinsBuild.GetUrl()

	// wait for build to finish
	waitToFinishError := jenkinsWaitForBuildToFinishFunc(ctx, jenkinsBuild, jenkinsWaitForBuildToFinishWaitTime)
	if waitToFinishError != nil {
		return errors.Wrap(waitToFinishError, "Error occurred while waiting for build to finish")
	}

	// write reports and links JSON
	reportFileName := xmake.BuildResultJSONFilename
	linkTitle := "Jenkins job"
	switch config.BuildType {
	case typeStage:
		reportFileName = "xmake_stage.json"
		linkTitle = fmt.Sprintf("xmake %s (stage)", linkTitle)
	case typePromote:
		reportFileName = "xmake_promote.json"
		linkTitle = fmt.Sprintf("xmake %s (promote)", linkTitle)
	}
	reports := []StepResults.Path{{Target: reportFileName, Mandatory: true}}
	links := []StepResults.Path{{Name: linkTitle, Target: jenkinsBuild.GetUrl()}}

	// load build-results.json
	buildArtifact := xmakeFetchBuildResultArtifact(ctx, jenkinsBuild, reportFileName)

	// fail if job failed
	if jenkinsBuild.GetResult() != "SUCCESS" {

		log.SetErrorCategory(log.ErrorBuild)

		//failed build analysis
		handleFailedBuild(ctx, jenkinsInstance, jenkinsBuild, buildArtifact, fileUtils)

		return fmt.Errorf("job did not succeed: %s", jenkinsBuild.GetResult())
	}
	// write data to pipeline environment
	if err := writePipelineEnvironment(ctx, buildArtifact, pipelineEnv, config); err != nil {
		return err
	}

	// create build-type json file
	// build-type json file contains { "build-type" : "docker" }
	if err := writeBuildTypeJson(ctx, buildArtifact, pipelineEnv, config, fileUtils); err != nil {
		return err
	}

	// create and save bom.xml files
	reports = writeSbomFiles(config, pipelineEnv, reports)

	// write reports and links JSON
	buildTypeReport := StepResults.Path{Target: "build-type.json", Mandatory: false}
	reports = append(reports, buildTypeReport)

	_ = StepResults.PersistReportsAndLinks("sapXmakeExecuteBuild", "", fileUtils, reports, links)

	return nil
}

func connectToJenkins(ctx context.Context, config *sapXmakeExecuteBuildOptions, client xmakeSender, jobInfo xmake.Job) (*gojenkins.Jenkins, error) {
	jenkinsInstance := &gojenkins.Jenkins{}
	var jenkinsConnectErr error

	//Try connecting multiple times
	tryCount := 1
	for tryCount <= 5 {

		// connect to Jenkins
		log.Entry().Infof("Connecting to Jenkins: %s", jobInfo.JenkinsURL)
		jenkinsInstance, jenkinsConnectErr = jenkinsInstanceFunc(ctx, client.StandardClient(), jobInfo.JenkinsURL, config.Username, config.Token)
		// If jenkins connection fails (jenkins may be slowed down temporarily) wait and retry
		if jenkinsConnectErr != nil {
			log.Entry().Warningf("(Try %d) Failed to connect to Jenkins with url '%s' - Error: '%s'", tryCount, jobInfo.JenkinsURL, jenkinsConnectErr)
			// Wait 3 seconds before retrying
			time.Sleep(jenkinsConnectWaitingTime)
		} else {
			log.Entry().Infof("Successfully connected to Jenkins after '%d' try", tryCount)
			break
		}

		tryCount = tryCount + 1
	}

	if jenkinsConnectErr != nil {
		return nil, errors.Wrapf(jenkinsConnectErr, "Failed to connect to Jenkins with url '%s'", jobInfo.JenkinsURL)
	}
	return jenkinsInstance, nil
}

func findJob(jobFinder xmake.JobFinder, jobName string) (xmake.Job, error) {
	jobList, jobLookupErr := jobFinder.Lookup(jobName)
	if jobLookupErr != nil {
		return xmake.Job{}, errors.Wrapf(jobLookupErr, "failed to lookup job with name '%s'", jobName)
	}
	if len(jobList) == 0 {
		return xmake.Job{}, fmt.Errorf("no jobs found with name '%s'", jobName)
	}
	log.Entry().Debugf("%d job(s) found", len(jobList))
	for _, job := range jobList {
		log.Entry().Debugf("JobName: %s", job.FullName)
		log.Entry().Debugf("JobUrl: %s", job.URL)
		log.Entry().Debugf("JenkinsUrl: %s", job.JenkinsURL)
	}
	jobInfo := jobList[len(jobList)-1]
	return jobInfo, nil
}

// Parse build-results.json artifact for JOB_FAILURE links
// Fetch error-dictionnary artifact BUILDRESULTS section for xmake errors
// Parse console logs for xmake error rules regexp
// Persist an html artifact that contains found information
func handleFailedBuild(ctx context.Context, jenkinsInstance *gojenkins.Jenkins, jenkinsBuild *gojenkins.Build, buildArtifact jenkins.Artifact, fileUtils piperutils.FileUtils) {

	log.Entry().Info("Build result is " + jenkinsBuild.GetResult() + " -> Analyzing build failure")

	consoleErrorLinks := getConsoleErrorLinksFromArtifact(ctx, buildArtifact)

	xmakeErrorMessages := getXmakeBuildErrorMessages(ctx, jenkinsInstance, jenkinsBuild, consoleErrorLinks)

	//Create simple html file with failed links
	failedReportName := "Failed xmake jobs - error details.html"
	report := "<html><body><h3>Failed xmake jobs - error details</h3>"
	//Add xmake errors (from either error_dictionary.json artifact or detected error rules regexp in the log)
	if len(xmakeErrorMessages) > 0 {
		report += "<b>xmake errors:</b>"
	}
	for job, errorMessage := range xmakeErrorMessages {
		report += "<li>" + job + " : <b>" + errorMessage + "</b></li>"
	}
	//Add console log links
	if len(consoleErrorLinks) > 0 {
		report += "<b>Console error links:</b>"
	}
	for _, buildConsoleUrl := range consoleErrorLinks {
		parts := strings.Split(buildConsoleUrl, "/")
		if len(parts) > 3 {
			buildName := parts[len(parts)-3]
			buildVersion := parts[len(parts)-2]
			report += "<li><a href=" + buildConsoleUrl + ">" + buildName + " #" + buildVersion + "</a></li>"
		}
	}
	report += "</body></html>"

	//Add report to workspace
	if err := fileUtils.WriteFile(filepath.Join("", fmt.Sprint(failedReportName)), []byte(report), 0666); err != nil {
		log.Entry().Warningf("failed to write report '%s': '%s'", failedReportName, err)
	} else {
		//Persist report for sapXmakeExecuteBuild
		reports := []StepResults.Path{{Target: failedReportName}}
		_ = StepResults.PersistReportsAndLinks("sapXmakeExecuteBuild", "", fileUtils, reports, nil)
	}

}

func getConsoleErrorLinksFromArtifact(ctx context.Context, buildArtifact jenkins.Artifact) []string {
	log.Entry().Info(" -> Parsing build-results.json artifact for JOB_FAILURE links")

	consoleLinks := []string{}

	//Read build-results.json artifact to persist each JOB_FAILURE links
	buildArtifactName := buildArtifact.FileName()
	log.Entry().Info("Reading: " + buildArtifactName)
	artifactByteArray, fetchErr := buildArtifact.GetData(ctx)
	if fetchErr != nil {
		log.Entry().Warningf("failed to read build result artifact: '%s'", fetchErr)
		return consoleLinks
	}

	//Unmarshal build-results.json artifact into jsonContent structure
	var jsonContent struct {
		JOB_FAILURE []string
	}

	if err := json.Unmarshal([]byte(artifactByteArray), &jsonContent); err != nil {
		log.Entry().Warningf("Failed to unmarshal artifact content: '%s'", err)
		return consoleLinks
	}

	//Display links in console log
	for _, element := range jsonContent.JOB_FAILURE {
		log.Entry().Info("-> found xmake job failure link: " + element)
		consoleLinks = append(consoleLinks, element)
	}

	return consoleLinks
}

// Retrieves all failed builds from console error links then for each:
// Fetch a specific xmake build artifact: /.xmake/error_dictionary.json
// That artifact contains a BUILDRESULTS array section xmake with error messages
// If no artifact is found we parse the console log for any regexp from xmake_error_rules.json
func getXmakeBuildErrorMessages(ctx context.Context, jenkinsInstance *gojenkins.Jenkins, build *gojenkins.Build, consoleErrorLinks []string) map[string]string {

	errorMessages := map[string]string{}
	jobName := build.Job.GetName()
	failedBuilds := []*gojenkins.Build{}

	log.Entry().Debugf("getXmakeBuildErrorMessages for job: '%s'", jobName)

	//Get failed builds from the console error links
	//GetDownstreamBuilds method is not working as it is base on build.Job.GetDownstreamJobs which appear as Unresolved in jenkins
	failedBuilds = getFailedBuildsFromConsoleLinks(ctx, jenkinsInstance, consoleErrorLinks)

	log.Entry().Debugf("found '%d' failed build(s)", len(failedBuilds))
	for _, failedBuild := range failedBuilds {
		failedBuildErrorMessages := fetchBuildError(ctx, failedBuild)
		errorMessages = mergeMaps(errorMessages, failedBuildErrorMessages)
	}

	//Display xmake error messages in log
	for key, value := range errorMessages {
		log.Entry().Infof("Job '%s' failed with error: '%s'", key, value)
	}

	return errorMessages
}

// Fetch error_dictionary artifact then (if not found) parse console log for error rule regexp
// Return a map with job name and found error (if any)
func fetchBuildError(ctx context.Context, build *gojenkins.Build) map[string]string {

	log.Entry().Debugf("fetchBuildError for build: '%s'", build.GetUrl())

	errorMessages := map[string]string{}
	jobName := build.Job.GetName()

	//Fetch error_dictionary artifact for that build
	log.Entry().Debugf("Fetching error_dictionary.json for job: '%s' and build number #'%d'", jobName, build.GetBuildNumber())
	artifact, fetchErr := xmakeFetchBuildArtifact(ctx, build, "error_dictionary.json")
	if fetchErr != nil {
		log.Entry().Infof("Failed to fetch error_dictionary.json: '%s'", fetchErr)

		//If the error_dictionary.json does not exist, we compare the console logs with the error_rules.yml
		//This file contains regexp definition to check against each console line
		consoleOutput := xmakeGetBuildConsoleOutput(ctx, build)
		errorMessage := findErrorRule(consoleOutput)
		if errorMessage != "" {
			errorMessages[jobName] = errorMessage
		}

	} else {
		log.Entry().Debugf("error_dictionary.json artifact successfully fetched")
		artifactByteArray, fetchErr := artifact.GetData(ctx)
		if fetchErr != nil {
			log.Entry().Warningf("failed to read error_dictionary.json artifact  for job: '%s' and build number #'%d': '%s'", jobName, build.GetBuildNumber(), fetchErr)
		} else {

			errorMessages = parseErrorDictionnaryArtifact(artifactByteArray, jobName, errorMessages)

		}
	}

	return errorMessages
}

// Parse the xmake build error_dictionary.json artifact for errors
func parseErrorDictionnaryArtifact(artifactByteArray []byte, jobName string, errorMessages map[string]string) map[string]string {

	log.Entry().Debugf("parseErrorDictionnaryArtifact for job: '%s'", jobName)

	//Unmarshal error_dictionary.json artifact into jsonContent structure
	var jsonContent struct {
		BUILDRESULTS []string
	}
	if err := json.Unmarshal([]byte(artifactByteArray), &jsonContent); err != nil {
		log.Entry().Warningf("Failed to unmarshal error_dictionary.json artifact content: '%s'", err)
	} else {
		if jsonContent.BUILDRESULTS != nil {
			//Display results in console log
			for _, element := range jsonContent.BUILDRESULTS {
				log.Entry().Infof("-> found BUILDRESULTS for job: '%s': '%s'", jobName, element)
			}

			errorMessages[jobName] = strings.Join(jsonContent.BUILDRESULTS, ":")
			log.Entry().Infof("-> found xmake error for job '%s': '%s'", jobName, errorMessages[jobName])
		} else {
			log.Entry().Warningf("Can not find any BUILDRESULTS in error_dictionary.json: '%s'", jsonContent)
		}

	}

	return errorMessages

}

func getBuildConsoleOutput(ctx context.Context, build *gojenkins.Build) string {
	return build.GetConsoleOutput(ctx)
}

// Get failed jenkins.Build builds from the error console links (build urls)
func getFailedBuildsFromConsoleLinks(ctx context.Context, jenkinsInstance *gojenkins.Jenkins, consoleErrorLinks []string) []*gojenkins.Build {

	failedBuilds := []*gojenkins.Build{}

	// Iterate over all console error links to get build urls
	for i := 0; i < len(consoleErrorLinks); i++ {
		buildUrl := strings.Split(consoleErrorLinks[i], "/consoleFull")[0]

		//Get jenkins.Build instance from url
		build, err := xmakeGetJenkinsBuildFromUrl(ctx, jenkinsInstance, buildUrl)
		if err != nil {
			log.Entry().Warningf("Failed to get build instance from url: '%s': '%s'", buildUrl, err)
		} else {
			failedBuilds = append(failedBuilds, build)
		}
	}

	return failedBuilds
}

// Get downstream builds from the build-results.json artifact that contains a downstreams section
func getFailedDownstreamBuildsFromArtifact(ctx context.Context, jenkinsInstance *gojenkins.Jenkins, buildArtifact jenkins.Artifact) []*gojenkins.Build {

	downstreamBuilds := []*gojenkins.Build{}

	log.Entry().Debugf("Reading buildArtifact:'%s'", buildArtifact.FileName())
	artifactByteArray, fetchErr := buildArtifact.GetData(ctx)
	log.Entry().Debugf("artifact content: '%s'", string(artifactByteArray[:]))

	if fetchErr != nil {
		log.Entry().Warningf("failed to read build result artifact: '%s'", fetchErr)
	} else {

		log.Entry().Debug("Reading downstreams jobs from build-results.json")

		//Unmarshal build-results.json artifact into jsonContent structure
		var jsonContent struct {
			DOWNSTREAMS map[string]string
		}
		if err := json.Unmarshal([]byte(artifactByteArray), &jsonContent); err != nil {
			log.Entry().Warningf("Failed to unmarshal artifact content: '%s'", err)
		} else {
			// Iterate over all downstream builds
			for key, buildUrl := range jsonContent.DOWNSTREAMS {
				log.Entry().Debugf("parsed downstream job: '%s' with url: '%s'", key, buildUrl)

				//Get jenkins.Build instance from url
				build, err := xmakeGetJenkinsBuildFromUrl(ctx, jenkinsInstance, buildUrl)
				if err != nil {
					log.Entry().Warningf("Failed to get build instance from url: '%s': '%s'", buildUrl, err)
				} else {
					downstreamBuilds = append(downstreamBuilds, build)
				}
			}
		}
	}

	return downstreamBuilds
}

// Return a Jenkins.Build instance from a xmake build url
func getJenkinsBuildFromUrl(ctx context.Context, jenkinsInstance *gojenkins.Jenkins, buildUrl string) (*gojenkins.Build, error) {

	log.Entry().Debugf("getXmakeBuildFromUrl '%s'", buildUrl)

	splits := strings.Split(buildUrl, "/")
	if len(splits) < 5 {
		return nil, errors.New(
			"invalid xmake build url: " + buildUrl)
	}
	jobFolder := splits[len(splits)-4]
	jobName := splits[len(splits)-2]
	sBuildNumber := splits[len(splits)-1]

	job := jobFolder + "/job/" + jobName

	iBuildNumber, err := strconv.ParseInt(sBuildNumber, 10, 64)
	if err != nil {
		log.Entry().Warningf("Error converting build number from '%s': '%s'", sBuildNumber, err)
		return nil, err
	}

	build, err := xmakeGetJenkinsInstanceBuild(ctx, jenkinsInstance, job, iBuildNumber)
	if err != nil {
		log.Entry().Warningf("Error getting for job: '%s' and number: '%d': '%s'", job, iBuildNumber, err)
		return nil, err
	}

	return build, nil
}

func getJenkinsInstanceBuild(ctx context.Context, jenkinsInstance *gojenkins.Jenkins, jobName string, buildNumber int64) (*gojenkins.Build, error) {
	return jenkinsInstance.GetBuild(ctx, jobName, buildNumber)
}

// Parse a build console output content
// Compare each line with the regexps found in error_rules.yml resource file
func findErrorRule(content string) string {

	log.Entry().Debug("findErrorRule: parse console log for error rule from xmake_error_rules.json")

	var jsonErrorRules map[string][]string
	if err := json.Unmarshal([]byte(xmake.XMAKE_ERROR_RULES), &jsonErrorRules); err != nil {
		log.Entry().Warningf("Failed to unmarshal xmake_error_rules: '%s'", err)
	} else {

		lines := strings.Split(content, "\n")
		// Parse all console lines
		log.Entry().Debugf("Parsing '%d' lines", len(lines))
		for _, line := range lines {
			// Parse all error rules from xmake_error_rules
			rule := getErrorRule(line, jsonErrorRules)
			if rule != "" {
				return rule
			}
		}
	}

	log.Entry().Debugf("Found no error rule in console log")
	return ""
}

func getErrorRule(line string, jsonErrorRules map[string][]string) string {
	for rule, val := range jsonErrorRules {
		for _, regexpError := range val {
			//Check if line contains the regex
			bMatch, err := regexp.MatchString(regexpError, line)
			if err != nil {
				log.Entry().Debugf("invalid regex: '%s'", regexpError)
			} else {
				if bMatch {
					log.Entry().Debugf("Found error: '%s' in line: '%s'", rule, line)
					return rule
				}
			}
		}
	}

	return ""
}

func mergeMaps(ms ...map[string]string) map[string]string {
	res := map[string]string{}

	//We don't handle duplicate keys (just replace)
	for _, m := range ms {
		for k, v := range m {
			res[k] = v
		}
	}

	return res
}

func writeSbomFiles(config *sapXmakeExecuteBuildOptions, pipelineEnv *sapXmakeExecuteBuildCommonPipelineEnvironment, reports []StepResults.Path) []StepResults.Path {
	if config.BuildType == typeStage {
		dataList, err := writeSBomXmlForStageBuild(pipelineEnv.custom.stageBOM)
		if err != nil {
			log.Entry().Warningf("failed to write sbom file: '%s'", err)
		}
		if len(dataList) > 0 {
			sBomReport := StepResults.Path{Target: "**/sbom/**/*", Mandatory: false}
			reports = append(reports, sBomReport)
		}
	}
	return reports
}

func fetchBuildResultArtifact(ctx context.Context, build jenkins.Build, reportFilename string) jenkins.Artifact {
	var fetchErr error

	//Try connecting multiple times
	tryCount := 1
	for tryCount <= 5 {
		// fetch build-results.json
		buildResultArtifact, fetchErr := xmakeFetchBuildResultJSON(ctx, build)

		// If jenkins connection fails (jenkins may be slowed down temporarily) wait and retry
		if fetchErr != nil {
			// missing build-results.json is handled on Groovy side
			log.Entry().Warningf("failed to fetch build result artifact: '%s'....Retrying", fetchErr)
			// Wait 3 seconds before retrying
			time.Sleep(fetchBuildResultArtifactWaitingTime)
		} else {
			log.Entry().Info("Successfully fetched build result artifact")
			// save to disk
			if _, saveErr := buildResultArtifact.Save(ctx, filepath.Join(".", reportFilename)); saveErr != nil {
				// missing build-results.json is handled on Groovy side
				log.Entry().WithError(saveErr).Error("failed to save report: ", saveErr)
			}

			return buildResultArtifact
		}

		tryCount = tryCount + 1
	}

	log.Entry().WithError(fetchErr).Error("failed to fetch report: ", fetchErr)
	return nil
}

func writePipelineEnvironment(ctx context.Context, artifact jenkins.Artifact, pipelineEnv *sapXmakeExecuteBuildCommonPipelineEnvironment, config *sapXmakeExecuteBuildOptions) error {
	if config.BuildType == typeStage {
		stageJSON, err := xmakeFetchStageJSON(ctx, artifact, config.ArtifactPattern)
		if err != nil {
			return err
		}
		pipelineEnv.custom.stageBOM = stageJSON.StageBom
		log.Entry().Debugf("stageBOM: '%s'", pipelineEnv.custom.stageBOM)
		pipelineEnv.custom.xmakeDeployPackage = stageJSON.ProjectArchive
		log.Entry().Debugf("projectArchive: '%s'", pipelineEnv.custom.xmakeDeployPackage)
		pipelineEnv.custom.xmakeStagingRepositoryID = stageJSON.StagingRepoID
		log.Entry().Debugf("staging_repo_id: '%s'", pipelineEnv.custom.xmakeStagingRepositoryID)

		// relevant for Docker-related steps e.g. imagePushToRegistry
		imageNames, imageNameTags, registryCredentials, err := xmake.GetImagesAndCredentialsFromStageBOM(&stageJSON.StageBom)
		if err != nil {
			log.Entry().Debugf("no image names and registry credentials exported to the pipeline env: %s", err)
		} else {
			pipelineEnv.container.imageNames = imageNames
			log.Entry().Debugf("imageNames: '%s'", pipelineEnv.container.imageNames)
			pipelineEnv.container.imageNameTags = imageNameTags
			log.Entry().Debugf("imageNameTags: '%s'", pipelineEnv.container.imageNameTags)
			pipelineEnv.container.registryURL = registryCredentials.RepositoryURL
			log.Entry().Debugf("repositoryURL: '%s'", pipelineEnv.container.registryURL)
			pipelineEnv.container.repositoryUsername = registryCredentials.User
			log.Entry().Debugf("repositoryUsername: '%s'", pipelineEnv.container.repositoryUsername)
			pipelineEnv.container.repositoryPassword = registryCredentials.Password
			log.Entry().Debugf("repositoryPassword: '%s'", pipelineEnv.container.repositoryPassword)
		}

	} else if config.BuildType == typePromote {
		promoteJSON, err := xmakeFetchPromoteJSON(ctx, artifact, config.ArtifactPattern)
		if err != nil {
			return err
		}
		for _, repository := range promoteJSON.PromoteBom.Repositories {
			if repository.Success {
				pipelineEnv.custom.promotedArtifactURLs = append(pipelineEnv.custom.promotedArtifactURLs, repository.Result...)
			}
		}

		// once artifact is promoted write status to CPE
		status := cumulus.ReleaseStatus{Status: "promoted"}

		// ignore error since format is in our hands
		releaseStatus, _ := json.Marshal(status)
		pipelineEnv.custom.releaseStatus = string(releaseStatus)
	}

	return nil
}

const defaultJobNameTemplate = "${githubOrg}-${githubRepo}-SP-${quality}-common${shipmentType?'_'+shipmentType:''}"

func getJobNamePatternWithRespectToLegacyParameters(config sapXmakeExecuteBuildOptions) string {
	if len(config.XMakeJobName) > 0 {
		log.Entry().Warn("Parameter 'xMakeJobName' is deprecated, please use jobNamePattern")
		if strings.HasPrefix(config.XMakeJobName, "ght-") {
			return xmake.JobNamePatternTools
		}
		return xmake.JobNamePatternInternal
	}

	if len(config.XMakeJobNameTemplate) > 0 && config.XMakeJobNameTemplate != defaultJobNameTemplate {
		log.Entry().Warn("Parameter 'xMakeJobNameTemplate' is deprecated, please use jobNamePattern")
		if strings.HasPrefix(config.XMakeJobNameTemplate, "ght-") {
			return xmake.JobNamePatternTools
		}
		return xmake.JobNamePatternInternal
	}
	return config.JobNamePattern
}

func writeBuildTypeJson(ctx context.Context, buildArtifact jenkins.Artifact, pipelineEnv *sapXmakeExecuteBuildCommonPipelineEnvironment, config *sapXmakeExecuteBuildOptions, fileUtils piperutils.FileUtils) error {
	if config.BuildType == typeStage {
		_, err := xmake.WriteBuildTypeJsonForStageBuild(pipelineEnv.custom.stageBOM, fileUtils)
		if err != nil {
			log.Entry().WithError(err).Error("failed to write build-type.json file")
			return err
		}
	}
	return nil
}

func handleTriggerErrorMessage(buildTriggerErr error, jobInfoFullName string) string {
	if strings.Contains(buildTriggerErr.Error(), "Could not invoke job \"\": 404") {
		return fmt.Sprintf("failed to trigger job '%s'. The job exists but you don't have enough permission to execute this job. Please check Jenkins credentials used to trigger this job", jobInfoFullName)
	}
	return fmt.Sprintf("failed to trigger job '%s': %v", jobInfoFullName, buildTriggerErr)
}

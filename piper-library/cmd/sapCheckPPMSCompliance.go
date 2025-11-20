package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	piperversioning "github.com/SAP/jenkins-library/pkg/versioning"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/whitesource"
	"github.com/pkg/errors"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/blackduck"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/gtlc"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/ppms"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/versioning"
)

const ipWhiteSource = "WhiteSource"
const ipBlackDuck = "BlackDuck"

type fossMappingResult struct {
	failedMatches      int
	fossMappings       []fossMapping
	missingFossMapping int
	successfulMatches  int
	totalLibraries     int
}

type fossMapping struct {
	ArtifactID string
	GroupID    string
	// HasPPMSMapping indicates if a mapping in the PPMS object exists
	HasPPMSMapping bool
	// PpmsID contains the id of the FOSS object in PPMS, empty if FOSS is not yet known to PPMS
	PpmsID      string
	Version     string
	IPDetails   []gtlc.IPComplianceItem
	EccnDetails []gtlc.ExportComplianceItem
}

type changeRequestSender interface {
	SendChangeRequest(scv *ppms.SoftwareComponentVersion, params ppms.ChangeRequestParams, foss []ppms.ChangeRequestFoss) (string, error)
	WaitForInitialChangeRequestFeedback(crID string, duration time.Duration) error
}

type ppmsUtils interface {
	DirExists(path string) (bool, error)
	FileExists(filename string) (bool, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	WriteFile(path string, content []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Now() time.Time
}

type ppmsUtilsBundle struct {
	*piperhttp.Client
	*piperutils.Files
}

func (w *ppmsUtilsBundle) Now() time.Time {
	return time.Now()
}

func newPPMSUtils(config *sapCheckPPMSComplianceOptions) *ppmsUtilsBundle {
	utils := ppmsUtilsBundle{
		Client: &piperhttp.Client{},
		Files:  &piperutils.Files{},
	}
	// Configure HTTP Client
	utils.SetOptions(piperhttp.ClientOptions{TransportTimeout: time.Duration(config.Timeout) * time.Second})
	return &utils
}

type wsSystem interface {
	GetProductName(productToken string) (string, error)
	GetProjectTokens(productToken string, projectNames []string) ([]string, error)
	GetProductByName(productName string) (whitesource.Product, error)
	GetProjectsMetaInfo(productToken string) ([]whitesource.Project, error)
}

func sapCheckPPMSCompliance(config sapCheckPPMSComplianceOptions, telemetryData *telemetry.CustomData, influx *sapCheckPPMSComplianceInflux) {

	utils := newPPMSUtils(&config)

	// record telemetry data about usage
	telemetryData.ServerURL = config.ServerURL
	telemetryData.ChangeRequestUpload = strconv.FormatBool(config.UploadChangeRequest)
	telemetryData.BuildVersionCreation = strconv.FormatBool(config.CreateBuildVersion)
	telemetryData.PullRequestMode = strconv.FormatBool(config.PullRequestMode)

	mappingSystem := gtlc.MappingSystem{
		ServerURL:  "https://cmapi.gtlc.only.sap",
		HTTPClient: utils.Client,
	}

	ppmsSystem := ppms.System{
		ServerURL:  config.ServerURL,
		Username:   config.Username,
		Password:   config.Password,
		HTTPClient: utils.Client,
	}

	whitesourceSystem := whitesource.NewSystem("https://sap.whitesourcesoftware.com/api", "6971b2eec2d3420bad0caf173ec629f6a3c7d3ba63f3445ab99ffdbf1acfb1d0", config.UserToken, time.Duration(config.Timeout)*time.Second)
	blackduckClient := blackduck.NewClient(config.DetectToken, config.DetectServerURL, utils.Client)

	influx.step_data.fields.ppms = "false"
	err := runPPMSComplianceCheck(&config, telemetryData, &mappingSystem, &ppmsSystem, whitesourceSystem, &blackduckClient, utils, "", time.Second)

	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to perform PPMS compliance check.")
	}
	influx.step_data.fields.ppms = "true"
}

func runPPMSComplianceCheck(config *sapCheckPPMSComplianceOptions, telemetryData *telemetry.CustomData, mappingSystem *gtlc.MappingSystem, ppmsSystem *ppms.System, whitesourceSystem wsSystem, blackduckClient *blackduck.Client, utils ppmsUtils, workdir string, duration time.Duration) error {
	//scanType is by default Whitesource
	scanType, err := getScanType(config)
	if err != nil {
		return err
	}

	if config.PullRequestMode {
		log.Entry().Info("Running in Pull-Request mode.")
		config.CreateBuildVersion = false
		config.UploadChangeRequest = false
	}

	var fossContained []gtlc.Foss
	var wsProductName string
	var bdVersionName string
	// use BlackDuck if token is configured and no WhiteSource user token found
	if scanType == ipBlackDuck {
		bdVersionName, fossContained, err = handleBlackDuck(config, mappingSystem, blackduckClient)
		if err != nil {
			return err
		}
	} else {
		wsProductName, fossContained, err = handleWhiteSource(config, mappingSystem, whitesourceSystem)
		if err != nil {
			return err
		}
	}

	// general PPMS aspects
	buildVersionName, err := effectiveBuildVersionName(config, telemetryData)
	if err != nil {
		return errors.Wrap(err, "failed to get build version name")
	}

	ppmsScv, err := ppmsSystem.GetSCV(config.PpmsID)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve SCV '%v'", config.PpmsID)
	}

	ppmsFoss, ppmsBV, err := retrieveFoss(config, buildVersionName, &ppmsScv, ppmsSystem, duration)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve FOSS objects")
	}
	log.Entry().Debugf("Retrieved PPMS Foss: %v", ppmsFoss)

	result := compareFoss(fossContained, ppmsFoss)
	log.Entry().Debugf("Comparison result: total: %v - failed: %v", result.totalLibraries, result.failedMatches)

	scanReport := createReport(result, config, &ppmsScv, ppmsBV.Name, scanType, wsProductName, bdVersionName, utils)

	var stepResult error
	scanReport.SuccessfulScan = true
	if result.failedMatches > 0 {
		stepResult = processFailedMatches(config, wsProductName, ppmsSystem, &ppmsScv, ppmsBV.ID, &result, utils, duration)
		scanReport.SuccessfulScan = (stepResult == nil)
	}

	reportDirectoryName := strings.ToLower(scanType)
	reportFileName := fmt.Sprintf("piper_%v_ppms_report", reportDirectoryName)
	reportPaths, err := writeReports(scanReport, ".", config, reportFileName, utils)
	if err != nil {
		log.Entry().WithError(err).Error("failed to create reports")
	}

	crV1FileExists, _ := utils.FileExists(config.ChangeRequestFileName)
	if crV1FileExists {
		reportPaths = append(reportPaths, piperutils.Path{Name: "PPMS Change Request V1", Target: config.ChangeRequestFileName})
	}

	links := []piperutils.Path{{Name: config.ReportName, Target: path.Join("artifact", fmt.Sprintf("%v.html", reportFileName)), Scope: "job"}}
	_ = piperutils.PersistReportsAndLinks("sapCheckPPMSCompliance", workdir, utils, reportPaths, links)

	if stepResult != nil {
		return stepResult
	}

	log.Entry().Info("sapCheckPPMSCompliance step was successful.")
	return nil
}

func handleWhiteSource(config *sapCheckPPMSComplianceOptions, mappingSystem *gtlc.MappingSystem, whitesourceSystem wsSystem) (string, []gtlc.Foss, error) {
	// call WhiteSource system and retrieve ProjectTokens
	projectTokens, err := getWhiteSourceProjectTokens(config, whitesourceSystem)
	if err != nil {
		return "", nil, err
	}

	wsProductName, err := whitesourceSystem.GetProductName(config.WhitesourceProductToken)
	if err != nil {
		log.Entry().WithError(err).Warning("failed to retrieve product name")
		wsProductName = ""
	}

	whiteSourceMapping := gtlc.WhiteSourceMapping{
		Filter:        []string{"SAP_IP"},
		ProjectTokens: projectTokens,
	}
	fossContained, err := whiteSourceMapping.WhiteSourcePPMSMapping(mappingSystem)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to retrieve the Whitesource-PPMS mapping")
	}
	log.Entry().Debugf("Retrieved FOSS from WhiteSource mapping: %v", fossContained)
	return wsProductName, fossContained, nil
}

func handleBlackDuck(config *sapCheckPPMSComplianceOptions, mappingSystem *gtlc.MappingSystem, blackduckClient *blackduck.Client) (string, []gtlc.Foss, error) {
	// allow a customScanVersion to supersede the automatically calculated version using version and versioningModel
	bdVersionName := getBlackDuckScanVersionName(config)

	// retrieve BlackDuck project versions
	bdVersions, err := getBlackDuckScanVersions(config, blackduckClient, bdVersionName)
	if err != nil {
		return "", nil, err
	}

	blackDuckMapping := gtlc.BlackDuckMapping{
		Filter:  []string{"SAP_IP"},
		APIURLs: bdVersions,
	}

	fossContained, err := blackDuckMapping.BlackDuckPPMSMapping(mappingSystem)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to retrieve the BlackDuck-PPMS mapping")
	}
	log.Entry().Debugf("Retrieved FOSS from BlackDuck mapping: %v", fossContained)
	return bdVersionName, fossContained, nil
}

func getWhiteSourceProjectNamesFromProjectNamePattern(whiteSourceProjectNamesPattern string, whiteSourceProductToken string, whiteSourceSystem wsSystem) ([]string, error) {
	projectNamesPattern, err := regexp.Compile(whiteSourceProjectNamesPattern)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return nil, fmt.Errorf("unable to compile the provided pattern for project names '%s'", whiteSourceProjectNamesPattern)
	}

	projects, err := whiteSourceSystem.GetProjectsMetaInfo(whiteSourceProductToken)

	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return nil, fmt.Errorf("failed to resolve projects for product '%s'", whiteSourceProductToken)
	}

	var whiteSourceProjectNames []string
	for _, project := range projects {
		if projectNamesPattern.MatchString(project.Name) {
			whiteSourceProjectNames = append(whiteSourceProjectNames, project.Name)
		}
	}
	return whiteSourceProjectNames, nil
}

func getScanType(config *sapCheckPPMSComplianceOptions) (string, error) {
	//scanType is by default Whitesource
	scanType := ipWhiteSource
	log.Entry().Debugf("Run PPMS with Detect? %v", config.RunComplianceCheckWithDetect)

	if isWhiteSourceEnabled(config) {
		log.Entry().Info("Using WhiteSource as source of dependency information.")
	} else if len(config.DetectToken) > 0 {
		// use BlackDuck scan as source
		scanType = ipBlackDuck
		log.Entry().Info("Using Detect (BlackDuck) as source of dependency information.")
	} else {
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", fmt.Errorf("missing mandatory configuration for dependency information: detectToken (Detect) or userToken (WhiteSource) needs to be set")
	}
	return scanType, nil
}

func getBlackDuckScanVersions(config *sapCheckPPMSComplianceOptions, blackduckClient *blackduck.Client, bdVersionName string) ([]string, error) {
	var err error
	var apiUrls []string

	// Handle single BlackDuck project
	if len(config.BlackduckProjectNames) == 0 && len(config.BlackduckProjectName) != 0 {
		var bdVersion *blackduck.ProjectVersion
		bdVersion, err = blackduckClient.GetProjectVersion(config.BlackduckProjectName, bdVersionName)
		if err == nil && bdVersion != nil {
			apiUrls = append(apiUrls, bdVersion.Href)
		}

		// Handle multiple BlackDuck projects
	} else if len(config.BlackduckProjectNames) > 0 {
		for _, bdProjectName := range config.BlackduckProjectNames {
			var bdVersion *blackduck.ProjectVersion
			bdVersion, err = blackduckClient.GetProjectVersion(bdProjectName, bdVersionName)
			if err == nil && bdVersion != nil {
				apiUrls = append(apiUrls, bdVersion.Href)
			}
		}
	} else {
		err = fmt.Errorf("no BlackDuck project name(s) provided")
	}

	errText := fmt.Sprint(err)
	switch {
	case strings.Contains(errText, "failed to get project version(s)"):
		log.SetErrorCategory(log.ErrorConfiguration)
	case strings.Contains(errText, "not found"):
		log.SetErrorCategory(log.ErrorConfiguration)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get BlackDuck project version(s)")
	}
	return apiUrls, nil
}

func getBlackDuckScanVersionName(config *sapCheckPPMSComplianceOptions) string {
	var bdVersionName string
	if len(config.CustomScanVersion) > 0 {
		bdVersionName = config.CustomScanVersion
	} else {
		coordinates := piperversioning.Coordinates{
			Version: config.Version,
		}

		_, bdVersionName = piperversioning.DetermineProjectCoordinates("", config.VersioningModel, coordinates)
	}
	return bdVersionName
}

func getWhiteSourceProjectTokens(config *sapCheckPPMSComplianceOptions, whiteSourceSystem wsSystem) ([]string, error) {
	if !isWhiteSourceProjectNameConfigured(config) {
		log.SetErrorCategory(log.ErrorConfiguration)
		return nil, fmt.Errorf("please configure the whitesource projects which shall be considered (either with whitesourceProjectNames or whitesourceProjectNamesPattern)")
	}

	//try to resolve product token by provided name
	if isWhiteSourceProductTokenToBeResolvedFromProductName(config) {
		if wsProduct, err := whiteSourceSystem.GetProductByName(config.WhitesourceProductName); err != nil {
			log.Entry().Error(err)
		} else {
			config.WhitesourceProductToken = wsProduct.Token
		}
	}

	if isWhiteSourceTokenNotConfigured(config) {
		log.SetErrorCategory(log.ErrorConfiguration)
		return nil, fmt.Errorf("missing mandatory configuration for WhiteSource: orgToken, whitesourceProductToken/whitesourceProductName need to be set")
	}

	if isWhiteSourceProjectNameToBeResolvedFromPattern(config) {
		whiteSourceProjectNames, err := getWhiteSourceProjectNamesFromProjectNamePattern(config.WhitesourceProjectNamesPattern, config.WhitesourceProductToken, whiteSourceSystem)
		if err != nil {
			return nil, err
		}
		config.WhitesourceProjectNames = whiteSourceProjectNames
	}

	if len(config.WhitesourceProjectNames) == 0 {
		log.SetErrorCategory(log.ErrorConfiguration)
		return nil, fmt.Errorf("whitesourceProjectNames not set")
	}

	projectTokens, err := whiteSourceSystem.GetProjectTokens(config.WhitesourceProductToken, config.WhitesourceProjectNames)
	errText := fmt.Sprint(err)
	switch {
	case strings.Contains(errText, "not all project token(s) found for provided projects"):
		log.SetErrorCategory(log.ErrorConfiguration)
	case strings.Contains(errText, "no project token(s) found for provided projects"):
		log.SetErrorCategory(log.ErrorConfiguration)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get project token(s) for %v", config.WhitesourceProjectNames)
	}
	return projectTokens, nil
}

func isWhiteSourceProductTokenToBeResolvedFromProductName(config *sapCheckPPMSComplianceOptions) bool {
	return len(config.WhitesourceProductToken) == 0 && len(config.WhitesourceProductName) > 0
}

func isWhiteSourceProjectNameToBeResolvedFromPattern(config *sapCheckPPMSComplianceOptions) bool {
	return len(config.WhitesourceProjectNames) == 0 && len(config.WhitesourceProjectNamesPattern) > 0
}

func isWhiteSourceEnabled(config *sapCheckPPMSComplianceOptions) bool {
	return len(config.UserToken) > 0 && !config.RunComplianceCheckWithDetect
}

func isWhiteSourceProjectNameConfigured(config *sapCheckPPMSComplianceOptions) bool {
	return len(config.WhitesourceProjectNames) > 0 || len(config.WhitesourceProjectNamesPattern) > 0
}

func isWhiteSourceTokenNotConfigured(config *sapCheckPPMSComplianceOptions) bool {
	return len(config.WhitesourceProductToken) == 0 || len(config.OrgToken) == 0
}

func effectiveBuildVersionName(config *sapCheckPPMSComplianceOptions, telemetryData *telemetry.CustomData) (string, error) {

	if len(config.BuildVersion) == 0 && !config.CreateBuildVersion {
		return "", nil
	}

	if len(config.BuildVersion) > 0 {
		// make sure that also former groovy template style keeps working
		config.BuildVersion = convertGroovyTemplate(config.BuildVersion, telemetryData)
		log.Entry().Debugf("Using custom build version: %v", config.BuildVersion)
	} else if len(config.Version) > 0 {
		// set a default build version for the creation case in case a version is available which is then used in the resolution
		config.BuildVersion = "{{.Major}}.{{.Minor}}.{{.Patch}}"
		log.Entry().Debug("Using default build version template: {{.Major}.{{.Minor}}.{{.Patch}}")
	}

	resolvedBuildVersion := config.BuildVersion
	if len(config.Version) > 0 {
		var err error
		resolvedBuildVersion, err = resolveBuildVersion(config.BuildVersion, config.Version)
		if err != nil {
			return "", errors.Wrap(err, "failed to resolve build version from build version template")
		}
		log.Entry().Debugf("PPMS build version based on parameter 'buildVersion' is: %v", resolvedBuildVersion)
	}

	// length of PPMS build version is limited to 30 characters
	if len(resolvedBuildVersion) > 30 {
		return "", fmt.Errorf("your PPMS Build version '%v' is longer than 30 characters. Please adapt the parameter 'buildVersion' accordingly", resolvedBuildVersion)
	}
	return resolvedBuildVersion, nil
}

func retrieveFoss(config *sapCheckPPMSComplianceOptions, buildVersionName string, ppmsScv *ppms.SoftwareComponentVersion, ppmsSystem *ppms.System, duration time.Duration) ([]ppms.Foss, ppms.BuildVersion, error) {
	ppmsFoss := []ppms.Foss{}
	ppmsBV := ppms.BuildVersion{}
	ppmsBVs, err := ppmsScv.GetBVs(ppmsSystem)
	if err != nil {
		return ppmsFoss, ppmsBV, errors.Wrapf(err, "failed to retrieve build versions for SCV '%v'", ppmsScv.Name)
	}
	log.Entry().Debugf("Retrieved PPMS build versions: %v", ppmsBVs)

	if len(buildVersionName) > 0 {
		// ignore that build version might not exist, we will create it later
		ppmsBV, _ := ppmsScv.GetBuildVersionByName(ppmsSystem, buildVersionName)
		log.Entry().Infof("Using build version: '%v'", buildVersionName)

		if len(ppmsBV.Name) == 0 && !config.CreateBuildVersion {
			return ppmsFoss, ppmsBV, fmt.Errorf("build version '%v' does not exist and createBuildVersion is set to 'false'", buildVersionName)
		}

		if config.CreateBuildVersion && len(ppmsBV.Name) == 0 {
			// build version does not exist, thus create it
			ppmsBV.Name = buildVersionName
			ppmsBV.Description = ""
			ppmsBV.SoftwareComponentVersionID = ppmsScv.ID

			log.Entry().Info("Uploading change request to PPMS system ...")
			// so far we will use a fixed setting for the copy (include predecessor's foss & scv/bv list)
			crID, err := ppmsSystem.SendChangeRequestBV(&ppmsBV, config.Username, ppmsScv, true, true)
			if err != nil {
				return ppmsFoss, ppmsBV, errors.Wrap(err, "failed to upload change request to PPMS")
			}
			log.Entry().Info("Uploading change request to PPMS system finished")
			log.Entry().Info("Checking for initial feedback ...")
			err = ppmsSystem.WaitForBuildVersionChangeRequestApplied(crID, duration)
			if err != nil {
				return ppmsFoss, ppmsBV, errors.Wrap(err, "change request returned with an error")
			}
			_, err = ppmsScv.UpdateBVs(ppmsSystem)
			if err != nil {
				return ppmsFoss, ppmsBV, errors.Wrapf(err, "failed to update build versions for SCV '%v'", ppmsScv.Name)
			}

			// update the build version with the one which was just created
			// no error handling since we anyway use the cached build version from above UpdateBVs() call
			ppmsBV, _ = ppmsScv.GetBuildVersionByName(ppmsSystem, buildVersionName)
		}

		ppmsFoss, err = ppmsBV.GetFoss(ppmsSystem)
		if err != nil {
			return ppmsFoss, ppmsBV, errors.Wrapf(err, "failed to retrieve FOSS for build version '%v'", ppmsBV.Name)
		}
		return ppmsFoss, ppmsBV, nil

	} else if len(ppmsBVs) > 0 {
		// use latest BV
		log.Entry().Debug("Build versions found, using latest build version")
		ppmsBV, err := ppmsScv.GetLatestBuildVersion(ppmsSystem)
		if err != nil {
			// this error is very unlikely
			return ppmsFoss, ppmsBV, errors.Wrap(err, "failed to retrieve latest build version")
		}
		log.Entry().Infof("Using latest build version: '%v'", ppmsBV.Name)
		ppmsFoss, err = ppmsBV.GetFoss(ppmsSystem)
		if err != nil {
			return ppmsFoss, ppmsBV, errors.Wrapf(err, "failed to retrieve FOSS for build version '%v'", ppmsBV.Name)
		}
		return ppmsFoss, ppmsBV, nil
	}

	// use SCV
	ppmsFoss, err = ppmsScv.GetFoss(ppmsSystem)
	if err != nil {
		return ppmsFoss, ppmsBV, errors.Wrapf(err, "failed to retrieve FOSS for SCV '%v'", ppmsScv.Name)
	}
	return ppmsFoss, ppmsBV, nil
}

func compareFoss(current []gtlc.Foss, target []ppms.Foss) fossMappingResult {
	log.Entry().Info("Start mapping FOSS objects.")
	mappingResult := fossMappingResult{
		totalLibraries:    len(current),
		successfulMatches: 0,
		failedMatches:     0,
		fossMappings:      []fossMapping{},
	}

	for _, gtlcFoss := range current {
		fossMapping := fossMapping{
			ArtifactID:     gtlcFoss.ArtifactID,
			GroupID:        gtlcFoss.GroupID,
			Version:        gtlcFoss.Version,
			PpmsID:         gtlcFoss.FossID,
			HasPPMSMapping: false,
			EccnDetails:    gtlcFoss.ExportComplianceDetails,
			IPDetails:      gtlcFoss.IPComplianceDetails,
		}
		if len(gtlcFoss.FossID) == 0 {
			mappingResult.missingFossMapping++
			mappingResult.failedMatches++
		} else if ppmsFossContains(target, gtlcFoss.FossID) {
			fossMapping.HasPPMSMapping = true
			mappingResult.successfulMatches++
		} else {
			mappingResult.failedMatches++
		}
		mappingResult.fossMappings = append(mappingResult.fossMappings, fossMapping)
	}
	log.Entry().Info("Finished mapping FOSS objects.")
	return mappingResult
}

func processFailedMatches(config *sapCheckPPMSComplianceOptions, wsProductName string, ppmsSystem changeRequestSender, ppmsScv *ppms.SoftwareComponentVersion, buildVersionID string, result *fossMappingResult, utils ppmsUtils, duration time.Duration) error {
	sourceInfo := []string{wsProductName}
	sourceInfo = append(sourceInfo, config.WhitesourceProjectNames...)
	sourceInfo = append(sourceInfo, config.Version)
	scanSource := strings.Join(sourceInfo, "; ")

	changeRequestParams := ppms.ChangeRequestParams{
		UserID:             config.Username,
		Source:             scanSource,
		Tool:               "Piper",
		BuildVersionNumber: buildVersionID,
		AddFoss:            []ppms.ChangeRequestFoss{},
		RemoveFoss:         []ppms.ChangeRequestFoss{},
	}
	if config.UploadChangeRequest {
		fossList := []ppms.ChangeRequestFoss{}
		for _, entry := range result.fossMappings {
			if len(entry.PpmsID) > 0 {
				fossList = append(fossList, ppms.ChangeRequestFoss{PPMSFossNumber: entry.PpmsID})
			}
		}
		log.Entry().Info("Uploading change request to PPMS system ...")
		crID, err := ppmsSystem.SendChangeRequest(ppmsScv, changeRequestParams, fossList)
		if err != nil {
			return errors.Wrap(err, "failed to upload change request to PPMS")
		}
		log.Entry().Info("Uploading change request to PPMS system finished")
		log.Entry().Debugf("Change request id: %v", crID)

		log.Entry().Info("Checking for initial feedback ...")
		err = ppmsSystem.WaitForInitialChangeRequestFeedback(crID, duration)
		if err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return err
		}
	} else {
		// we do not support deletion of foss here, this is tackled with automatic change request upload
		// calculate missing foss to be included in change request document for manual upload
		addFoss := []ppms.ChangeRequestFoss{}

		for _, entry := range result.fossMappings {
			// only pick foss which is known to ppms but not yet included in the model
			if !entry.HasPPMSMapping && len(entry.PpmsID) > 0 {
				addFoss = append(addFoss, ppms.ChangeRequestFoss{PPMSFossNumber: entry.PpmsID})
			}
		}
		changeRequestParams.AddFoss = addFoss

		err := ppmsScv.WriteChangeRequestV1File(config.ChangeRequestFileName, changeRequestParams, utils.FileWrite)
		if err != nil {
			log.Entry().WithError(err).Error("failed to write PPMS change request document")
		}
	}

	// BOM in PPMS cannot be completed due to missing PPMS FOSS objects
	// According to latest rules these kind of errors can be ignored
	if result.missingFossMapping > 0 && (result.missingFossMapping == result.failedMatches) {
		return nil
	}

	// no error in case automatic upload is activated
	if !config.UploadChangeRequest {
		log.SetErrorCategory(log.ErrorCompliance)
		return fmt.Errorf("%v PPMS entries are missing for this build. A report has been generated and stored as build artifact", result.failedMatches)
	}

	return nil
}

func ppmsFossContains(ppmsFoss []ppms.Foss, fossID string) bool {
	for _, foss := range ppmsFoss {
		if fossID == foss.ID {
			return true
		}
	}
	return false
}

func convertGroovyTemplate(template string, telemetryData *telemetry.CustomData) string {
	re := regexp.MustCompile(`\${version\.(major|minor|patch|timestamp)}`)
	res := re.ReplaceAllString(template, "{{.$1}}")
	res = strings.ReplaceAll(res, ".major", ".Major")
	res = strings.ReplaceAll(res, ".minor", ".Minor")
	res = strings.ReplaceAll(res, ".patch", ".Patch")
	res = strings.ReplaceAll(res, ".timestamp", ".Timestamp")
	if res != template {
		// Telemetry & log entry in case the groovy style is still used
		telemetryData.GroovyTemplateUsed = template
		log.Entry().Warningf("Deprecated groovy build version template used:'%v'. Please switch to following golang template: '%v'", template, res)
	}
	return res
}

func resolveBuildVersion(versionTemplate, fullVersion string) (string, error) {
	currentVersion, err := versioning.SplitVersion(fullVersion)
	if err != nil {
		return "", errors.Wrapf(err, "failed to resolve build version")
	}
	tmpl, err := template.New("buildversion").Parse(versionTemplate)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create build version template: %v", versionTemplate)
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, currentVersion)
	if err != nil {
		return "", errors.Wrapf(err, "failed to execute build version template: %v", versionTemplate)
	}
	return buf.String(), nil
}

func createReport(result fossMappingResult, config *sapCheckPPMSComplianceOptions, scv *ppms.SoftwareComponentVersion, buildVersionName, scanType, wsProductName, bdVersion string, utils ppmsUtils) reporting.ScanReport {

	var matchesStyle reporting.ColumnStyle = 0
	if result.failedMatches != 0 {
		matchesStyle = reporting.Red
	}

	reportTitle := config.ReportTitle
	if len(reportTitle) == 0 {
		reportTitle = config.ReportName
	}

	scanReport := reporting.ScanReport{
		ReportTitle: reportTitle,
		Overview: []reporting.OverviewRow{
			{Description: "Total number of libraries", Details: fmt.Sprint(result.totalLibraries)},
			{Description: "Total number of successful library matches", Details: fmt.Sprint(result.successfulMatches), Style: matchesStyle},
		},
		ReportTime: utils.Now(),
	}

	var subheaders []reporting.Subheader

	if scanType == ipBlackDuck {

		if len(config.BlackduckProjectNames) == 0 && len(config.BlackduckProjectName) != 0 {

			subheaders = []reporting.Subheader{
				{Description: "BlackDuck project name", Details: config.BlackduckProjectName},
				{Description: "BlackDuck project version", Details: bdVersion},
			}

		} else {

			subheaders = []reporting.Subheader{
				{Description: "BlackDuck project names", Details: strings.Join(config.BlackduckProjectNames, ", ")},
				{Description: "BlackDuck project version", Details: bdVersion},
			}

		}

	} else {
		subheaders = []reporting.Subheader{
			{Description: "WhiteSource product name", Details: wsProductName},
			{Description: "Filtered project names", Details: strings.Join(config.WhitesourceProjectNames, ", ")},
		}
	}

	scvTextLink := fmt.Sprintf(`<a href="%v/ppmslight/#/details/cv/%v/overview" target="_blank">%v</a>`, config.ServerURL, config.PpmsID, scv.Name)
	if len(buildVersionName) > 0 {
		scvTextLink += fmt.Sprintf(" (Build version: %v)", buildVersionName)
	}
	subheaders = append(
		subheaders,
		reporting.Subheader{Description: "PPMS SCV", Details: scvTextLink},
		reporting.Subheader{Description: "PPMS go to market channel name", Details: config.ChannelID},
	)

	scanReport.Subheaders = subheaders

	changeRequestInfo := ""
	if result.failedMatches > 0 && config.UploadChangeRequest {
		changeRequestInfo = "<br />PPMS Change Request triggered automatically."
	} else if result.failedMatches > 0 {
		changeRequestInfo = `<a onclick="document.getElementById('crJson').style.display = 'inline-block';" href="#">Show</a> / <a onclick="document.getElementById('crJson').style.display = 'none';" href="#">hide</a> Change Request JSON can be found in workspace (in Jenkins also in archives)<br />
		You can send this to your PPMS entry owner in order to update the PPMS data.
		<!--<textarea id="crJson" name="changeRequest" rows="20" cols="60" style="display:none;">Change Request JSON</textarea>-->`
	}
	scanReport.FurtherInfo = changeRequestInfo

	detailTable := reporting.ScanDetailTable{
		NoRowsMessage: "No library entries found",
		Headers: []string{
			"Group-ID",
			"Artifact-ID",
			"Version",
			"Comprised in SCV",
			"PPMS FOSS license risk rating",
		},
		WithCounter:   true,
		CounterHeader: "Entry #",
	}

	for _, mapping := range result.fossMappings {
		row := reporting.ScanRow{}
		row.AddColumn(mapping.GroupID, 0)
		row.AddColumn(mapping.ArtifactID, 0)
		row.AddColumn(mapping.Version, 0)
		var ppmsStyle reporting.ColumnStyle = reporting.Green
		if !mapping.HasPPMSMapping {
			ppmsStyle = reporting.Red
		}
		row.AddColumn(comprisedTextWithLink(mapping, config), ppmsStyle)
		row.AddColumn(riskColumn(mapping, config))
		detailTable.Rows = append(detailTable.Rows, row)
	}
	scanReport.DetailTable = detailTable

	return scanReport
}

func writeReports(scanReport reporting.ScanReport, filePath string, config *sapCheckPPMSComplianceOptions, reportFileName string, utils ppmsUtils) ([]piperutils.Path, error) {
	reportPaths := []piperutils.Path{}

	// ignore templating errors since template is in our hands and issues will be detected with the automated tests
	htmlReport, _ := scanReport.ToHTML()
	htmlReportPath := filepath.Join(filePath, fmt.Sprintf("%v.html", reportFileName))
	if err := utils.FileWrite(htmlReportPath, htmlReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write html report")
	}
	reportPaths = append(reportPaths, piperutils.Path{Name: "PPMS Compliance Report", Target: htmlReportPath})

	if !scanReport.SuccessfulScan && config.CreateResultIssue && len(config.GithubToken) > 0 && len(config.GithubAPIURL) > 0 && len(config.Owner) > 0 && len(config.Repository) > 0 {
		log.Entry().Debug("Creating/updating GitHub issue with check results")
		err := ppms.UploadReportToGithub(scanReport, config.GithubToken, config.GithubAPIURL, config.Owner, config.Repository, config.Assignees)
		if err != nil {
			return reportPaths, fmt.Errorf("failed to upload scan results into GitHub: %w", err)
		}
	}

	// JSON reports are used by step pipelineCreateSummary in order to e.g. prepare an issue creation in GitHub
	// ignore JSON errors since structure is in our hands
	jsonReport, _ := scanReport.ToJSON()
	if exists, _ := utils.DirExists(reporting.StepReportDirectory); !exists {
		err := utils.MkdirAll(reporting.StepReportDirectory, 0777)
		if err != nil {
			return reportPaths, errors.Wrap(err, "failed to create reporting directory")
		}
	}
	if err := utils.FileWrite(filepath.Join(reporting.StepReportDirectory, fmt.Sprintf("%v_%v.json", reportFileName, utils.Now().Format("20060102150405"))), jsonReport, 0666); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return reportPaths, errors.Wrapf(err, "failed to write json report")
	}
	// we do not add the json report to the overall list of reports for now,
	// since it is just an intermediary report used as input for later
	// and there does not seem to be real benefit in archiving it.

	return reportPaths, nil
}

func comprisedTextWithLink(foss fossMapping, config *sapCheckPPMSComplianceOptions) string {
	if !foss.HasPPMSMapping && len(foss.PpmsID) == 0 {
		return "Mapping not found - FOSS has no PPMS ID yet"
	}

	comprisedText := "is comprised"

	if !foss.HasPPMSMapping {
		comprisedText = "is not comprised<br>"
		if config.UploadChangeRequest {
			comprisedText += `PPMS change has been triggered automatically`
		} else {
			comprisedText += fmt.Sprintf(`<a href="https://open-source.tools.sap.corp/details/%v" target="_blank">Declare FOSS Usage</a>`, foss.PpmsID)
		}
	}

	return fmt.Sprintf(`<a href="https://i7p.wdf.sap.corp/ppmslight/#/details/foss/%v/overview" target="_blank">FOSS</a> %v`, foss.PpmsID, comprisedText)
}

func riskColumn(foss fossMapping, config *sapCheckPPMSComplianceOptions) (string, reporting.ColumnStyle) {
	log.Entry().Debugf("IP details: %v", foss.IPDetails)
	log.Entry().Debugf("Channel ID: %v", config.ChannelID)
	rating, _ := gtlc.GetChannelRiskRating(foss.IPDetails, config.ChannelID)
	log.Entry().Debugf("Channel risk rating: %v", rating)
	return resolveRiskRatingView(rating.InherentRiskRating)
}

func resolveRiskRatingView(riskRating string) (riskText string, riskStyle reporting.ColumnStyle) {

	switch riskRating {
	case "GREEN":
		riskText = "Ok"
		riskStyle = reporting.Green
	case "YELLOW":
		riskText = "Medium Risk"
		riskStyle = reporting.Yellow
	case "RED":
		riskText = "High Risk"
		riskStyle = reporting.Red
	case "GREY":
		riskText = "Not yet rated"
		riskStyle = reporting.Grey
	case "BLACK":
		riskText = "Do not use!"
		riskStyle = reporting.Black
	default:
		riskText = "Unknown until FOSS is comprised in SCV"
		riskStyle = reporting.Grey
	}

	return
}

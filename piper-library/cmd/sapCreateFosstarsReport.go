package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	StepResults "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/fosstars"
)

const queryServiceUrlSuffix string = "/v2/ratings/namespaces/%s/names/%s/identifiertypes/%s/identifiers"
const modelRatingsDefinitionsEndpoint string = "/v2/ratings/definitions/%s"
const mavenBuildDescriptorFileName string = "pom.xml"
const npmBuildDescriptorFileName string = "package.json"
const mtaBuildDescriptorFileName string = "mta.yaml"

func sapCreateFosstarsReport(mySapCreateFosstarsReportOptions sapCreateFosstarsReportOptions, telemetryData *telemetry.CustomData) {
	log.Entry().WithFields(getLoggingInfo(mySapCreateFosstarsReportOptions)).Debug("Calling Fosstars Report")

	sapFosstars := getOptedSettingsForFosstrasReport(mySapCreateFosstarsReportOptions)

	previousWorkingDirectory, err := os.Getwd()
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while getting the current working directory")
	}

	checkOutPath, err := getSCMCheckoutPath(mySapCreateFosstarsReportOptions, previousWorkingDirectory, &sapFosstars)
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while cloning the repo")
	}
	// Changing the current directory to the checkout path to make sure projectStructure works fine
	err = os.Chdir(checkOutPath)
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while changing the directory")
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while getting the current working directory")
	}

	defer func() {
		if !mySapCreateFosstarsReportOptions.SkipSCMCheckOut {
			err := os.Chdir(cwd)
			if err != nil {
				log.Entry().WithError(err).Fatal("Error while changing the directory during the cleanup")
			}
			err = os.RemoveAll(checkOutPath)
			if err != nil {
				log.Entry().WithError(err).Fatal("Error while removing the directory during the cleanup")
			}
		}
	}() // clean up

	badRatingsOccuredOverAll, err := runSapCreateFosstarsReport(mySapCreateFosstarsReportOptions, &sapFosstars, previousWorkingDirectory, checkOutPath, &fosstars.InvocationClient{}, &fosstars.DependencyTreeGenerator{})
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while creating Fosstars report")
	}

	// write reports JSON
	reports := []StepResults.Path{
		{Target: "fosstars-report/*.html", Mandatory: true},
		{Target: "fosstars-report/*.json", Mandatory: true},
	}

	fileUtils := piperutils.Files{}

	StepResults.PersistReportsAndLinks("sapCreateFosstarsReport", "", &fileUtils, reports, nil)
	if badRatingsOccuredOverAll {
		log.Entry().Fatal("There were some BAD ratings for the artifacts as per the configurations. Please check the reports generated and the logs above for more information!")
	}
	log.Entry().Debug("Done with ratings report generation returning from Go")
}

func runSapCreateFosstarsReport(mySapCreateFosstarsReportOptions sapCreateFosstarsReportOptions, sapFosstars *fosstars.Fosstars, previousWorkingDirectory string, checkOutPath string, invocationClient fosstars.InvocationClientInterface, dependencyTreeGenerator fosstars.DependencyTreeGeneratorInterface) (bool, error) {
	var badRatingsOccuredOverAll bool

	if !isBuildDescriptorSupported(sapFosstars, checkOutPath) {
		return badRatingsOccuredOverAll, fmt.Errorf("unsupported build descriptor: %v", sapFosstars.BuildDescriptor)
	}

	if sapFosstars.BuildAllMavenPomFiles {
		log.Entry().Infof("Started extracting ratings for maven projects automatically")
		mavenRatingsOccured, err := extractRatingsForAllMavenProjects(sapFosstars, invocationClient, checkOutPath, previousWorkingDirectory, dependencyTreeGenerator, mySapCreateFosstarsReportOptions)
		if err != nil {
			return badRatingsOccuredOverAll, err
		}
		return mavenRatingsOccured, nil
	} else if isReportForMavenProject(sapFosstars, checkOutPath) {
		log.Entry().Infof("Started extracting ratings for maven projects")
		badRatingsOccured, err := extractRatingsForMavenProjects(sapFosstars, invocationClient, checkOutPath, previousWorkingDirectory, dependencyTreeGenerator, mySapCreateFosstarsReportOptions)
		if err != nil {
			return badRatingsOccuredOverAll, fmt.Errorf("Error while extracting ratings for the Maven project, error: %v", err)
		}
		badRatingsOccuredOverAll = badRatingsOccuredOverAll || badRatingsOccured
	}

	extractFuncs := map[bool]func(sapFosstars *fosstars.Fosstars, invocationClient fosstars.InvocationClientInterface, checkOutPath string, previousWorkingDirectory string, mySapCreateFosstarsReportOptions sapCreateFosstarsReportOptions) (bool, error){
		isReportForNpmProject(sapFosstars, checkOutPath): extractRatingsForNPMProjects,
		isReportForMtaProject(sapFosstars, checkOutPath): extractRatingsForMtaProjects,
	}

	for isConfigured, extractFn := range extractFuncs {
		if isConfigured {
			badRatingsOccured, err := extractFn(sapFosstars, invocationClient, checkOutPath, previousWorkingDirectory, mySapCreateFosstarsReportOptions)
			if err != nil {
				return badRatingsOccuredOverAll, err
			}

			badRatingsOccuredOverAll = badRatingsOccuredOverAll || badRatingsOccured
			break
		}
	}

	return badRatingsOccuredOverAll, nil
}

func isBuildDescriptorSupported(sapFosstars *fosstars.Fosstars, checkOutPath string) bool {
	return isReportForMavenProject(sapFosstars, checkOutPath) || isReportForNpmProject(sapFosstars, checkOutPath) || isReportForMtaProject(sapFosstars, checkOutPath)
}

func extractRatingsForAllMavenProjects(sapFosstars *fosstars.Fosstars, invocationClient fosstars.InvocationClientInterface, checkOutPath string, previousWorkingDirectory string, dependencyTreeGenerator fosstars.DependencyTreeGeneratorInterface, mySapCreateFosstarsReportOptions sapCreateFosstarsReportOptions) (bool, error) {
	allPomFilesToBuildDependencyTree, err := dependencyTreeGenerator.GetAllPomFilesToBuildDependencyTree(checkOutPath)
	if err != nil {
		return false, fmt.Errorf("Error while getting all pom files to build maven project automatically, error: %v", err)
	}
	log.Entry().Infof("allPomFilesToBuildDependencyTree: %v", allPomFilesToBuildDependencyTree)

	var extractedRatingsForValidPom, badRatingsOccured bool
	for _, pomFilePath := range allPomFilesToBuildDependencyTree {
		replacer := strings.NewReplacer(checkOutPath, "")
		buildDescriptorRelativePath := replacer.Replace(pomFilePath)

		sapFosstars.BuildDescriptor = buildDescriptorRelativePath
		badRatingsOccured, err = extractRatingsForMavenProjects(sapFosstars, invocationClient, checkOutPath, previousWorkingDirectory, dependencyTreeGenerator, mySapCreateFosstarsReportOptions)
		if err != nil {
			log.Entry().Warnf("Error while extracting ratings for the pom: %v error: %v", pomFilePath, err)
			continue
		}
		extractedRatingsForValidPom = true
	}

	if !extractedRatingsForValidPom {
		return badRatingsOccured, fmt.Errorf("Error while extracting ratings, none of the pom files could be built")
	}

	return badRatingsOccured, nil
}

func getSCMCheckoutPath(mySapCreateFosstarsReportOptions sapCreateFosstarsReportOptions, previousWorkingDirectory string, sapFosstars *fosstars.Fosstars) (string, error) {
	var checkOutPath string
	if !mySapCreateFosstarsReportOptions.SkipSCMCheckOut {
		checkOutPath = filepath.Join(os.TempDir(), "/fosstars")
		err := sapFosstars.ValidateInput()
		if err != nil {
			return checkOutPath, err
		}

		// Clone repository to Tempdir
		err = sapFosstars.CloneRepo(mySapCreateFosstarsReportOptions.Username, mySapCreateFosstarsReportOptions.Password, checkOutPath, time.Minute)
		if err != nil {
			return checkOutPath, err
		}
	} else {
		checkOutPath = previousWorkingDirectory
	}

	return checkOutPath, nil
}

func getLoggingInfo(mySapCreateFosstarsReportOptions sapCreateFosstarsReportOptions) map[string]interface{} {
	return map[string]interface{}{
		"Build Descriptor":               mySapCreateFosstarsReportOptions.BuildDescriptor,
		"SCM Url":                        mySapCreateFosstarsReportOptions.SCMURL,
		"Branch":                         mySapCreateFosstarsReportOptions.Branch,
		"Fosstars QueryService Base Url": mySapCreateFosstarsReportOptions.FosstarsQueryServiceBaseURL,
		"Fosstars Client suffix":         mySapCreateFosstarsReportOptions.FosstarsClientSuffix,
		"Rating Name Space":              mySapCreateFosstarsReportOptions.RatingNameSpace,
		"Rating Name":                    mySapCreateFosstarsReportOptions.RatingName,
	}
}

func getOptedSettingsForFosstrasReport(mySapCreateFosstarsReportOptions sapCreateFosstarsReportOptions) fosstars.Fosstars {
	return fosstars.Fosstars{
		BuildDescriptor:             mySapCreateFosstarsReportOptions.BuildDescriptor,
		SCMUrl:                      mySapCreateFosstarsReportOptions.SCMURL,
		Branch:                      mySapCreateFosstarsReportOptions.Branch,
		FosstarsQueryServiceBaseURL: mySapCreateFosstarsReportOptions.FosstarsQueryServiceBaseURL,
		RatingNameSpace:             mySapCreateFosstarsReportOptions.RatingNameSpace,
		RatingName:                  mySapCreateFosstarsReportOptions.RatingName,
		RatingLabelThreshold:        mySapCreateFosstarsReportOptions.RatingLabelThreshold,
		RatingValueThreshold:        mySapCreateFosstarsReportOptions.RatingValueThreshold,
		ExcludedLibraries:           mySapCreateFosstarsReportOptions.ExcludedLibraries,
		IncludeTransitiveDependency: mySapCreateFosstarsReportOptions.IncludeTransitiveDependency,
		FailOnUnclearRatings:        mySapCreateFosstarsReportOptions.FailOnUnclearRatings,
		RequestedRetryCount:         mySapCreateFosstarsReportOptions.RequestedRetryCount,
		ExcludeSAPInternalLibraries: mySapCreateFosstarsReportOptions.ExcludeSAPInternalLibraries,
		ExcludeTestDevDependencies:  mySapCreateFosstarsReportOptions.ExcludeTestDevDependencies,
		BuildAllMavenPomFiles:       mySapCreateFosstarsReportOptions.BuildAllMavenPomFiles,
	}
}

func isReportForMavenProject(sapFosstars *fosstars.Fosstars, checkOutPath string) bool {
	return isMavenBuildDescriptor(sapFosstars) || (!isBuildDescriptorProvided(sapFosstars) && hasBuildDescriptorAtRoot(checkOutPath, mavenBuildDescriptorFileName))
}

func isReportForNpmProject(sapFosstars *fosstars.Fosstars, checkOutPath string) bool {
	return isNpmBuildDescriptor(sapFosstars) || (!isBuildDescriptorProvided(sapFosstars) && hasBuildDescriptorAtRoot(checkOutPath, npmBuildDescriptorFileName))
}

func isReportForMtaProject(sapFosstars *fosstars.Fosstars, checkOutPath string) bool {
	return isMtaBuildDescriptor(sapFosstars) || (!isBuildDescriptorProvided(sapFosstars) && hasBuildDescriptorAtRoot(checkOutPath, mtaBuildDescriptorFileName))
}

func isMavenBuildDescriptor(sapFosstars *fosstars.Fosstars) bool {
	return strings.HasSuffix(sapFosstars.BuildDescriptor, mavenBuildDescriptorFileName)
}

func isNpmBuildDescriptor(sapFosstars *fosstars.Fosstars) bool {
	return strings.HasSuffix(sapFosstars.BuildDescriptor, npmBuildDescriptorFileName)
}

func isBuildDescriptorProvided(sapFosstars *fosstars.Fosstars) bool {
	return len(sapFosstars.BuildDescriptor) > 0
}

func isMtaBuildDescriptor(sapFosstars *fosstars.Fosstars) bool {
	return strings.HasSuffix(sapFosstars.BuildDescriptor, mtaBuildDescriptorFileName)
}

func hasBuildDescriptorAtRoot(checkOutPath string, buildDescriptorFileName string) bool {
	buildDescriptorfilePath := filepath.Join(checkOutPath, buildDescriptorFileName)
	isFileExist, _ := piperutils.FileExists(buildDescriptorfilePath)
	return isFileExist
}

func extractRatingsForMavenProjects(sapFosstars *fosstars.Fosstars, invocationClient fosstars.InvocationClientInterface, checkOutPath string, previousWorkingDirectory string, dependencyTreeGenerator fosstars.DependencyTreeGeneratorInterface, mySapCreateFosstarsReportOptions sapCreateFosstarsReportOptions) (bool, error) {
	dependencyTreeFilePaths, err := extractDependencyTrees(sapFosstars, dependencyTreeGenerator, checkOutPath, mySapCreateFosstarsReportOptions.GlobalSettingsFile, mySapCreateFosstarsReportOptions.BuildQuality)
	if err != nil {
		return false, fmt.Errorf("Could not generate dependency tree and hence no Fosstars report: %w", err)
	}
	var badRatingsOccuredOverAll bool
	for _, dependencyTreeFilePath := range dependencyTreeFilePaths {
		var reportName string
		artifacts, err := extractArtifacts(sapFosstars, dependencyTreeFilePath)
		if err != nil {
			log.Entry().Warnf("extractArtifacts error, skipping: %v", err)
			continue
		}
		if artifacts == nil {
			continue
		}
		fosstarsQueryServiceURL := fmt.Sprintf(sapFosstars.FosstarsQueryServiceBaseURL+queryServiceUrlSuffix, sapFosstars.RatingNameSpace, sapFosstars.RatingName, "maven")
		artifactsRatingsMap, badRatingsOccured, err := extractRatings(sapFosstars, fosstarsQueryServiceURL, invocationClient, artifacts, mySapCreateFosstarsReportOptions)
		if err != nil {
			return false, fmt.Errorf("Error while extracting Ratings.: %w", err)
		}
		if badRatingsOccured {
			badRatingsOccuredOverAll = badRatingsOccured
		}
		lines, err := fosstars.GetFileContentAsList(dependencyTreeFilePath)
		if err != nil {
			return false, fmt.Errorf("GetFileContentAsList error while generating Html Report %w", err)
		}
		if err = os.Chdir(previousWorkingDirectory); err != nil {
			return false, fmt.Errorf("Error while changing the directory %w", err)
		}
		reportName, err = fosstars.GenerateHtmlReportForMaven(artifactsRatingsMap, lines, fosstarsQueryServiceURL)
		if err != nil {
			return false, fmt.Errorf("Error while generating Html Report %w", err)
		}

		depGraphFilePath := strings.ReplaceAll(dependencyTreeFilePath, "fosstar-generated-tree.txt", "target/dependency-graph.json")
		replacer := strings.NewReplacer("fosstar-generated-tree.txt", "", checkOutPath, "", "\\", "/")
		buildDescriptorRelativePath := replacer.Replace(dependencyTreeFilePath) + mavenBuildDescriptorFileName
		pomArtifact := fosstars.GetPomArtifact(lines)
		generateMavenDependencyTreeRatingsJson(sapFosstars, artifactsRatingsMap, dependencyTreeGenerator, depGraphFilePath, reportName, pomArtifact, buildDescriptorRelativePath[1:])
		dependencyTreeGenerator = &fosstars.DependencyTreeGenerator{}
		generateJsonReportMaven(sapFosstars, invocationClient, reportName, depGraphFilePath, pomArtifact, mySapCreateFosstarsReportOptions)
	}

	return badRatingsOccuredOverAll, nil
}

func extractRatingsForNPMProjects(sapFosstars *fosstars.Fosstars, invocationClient fosstars.InvocationClientInterface, checkOutPath string, previousWorkingDirectory string, mySapCreateFosstarsReportOptions sapCreateFosstarsReportOptions) (bool, error) {
	filePathsToNpmInfoMap, err := extractNpmInfo(checkOutPath, sapFosstars)
	if err != nil {
		return false, fmt.Errorf("Error while extracting artifacts from the package.son files specified %w", err)
	}
	var badRatingsOccuredOverAll bool
	for filePath, npmInfo := range filePathsToNpmInfoMap {
		var reportName string
		artifacts := npmInfo.Artifacts
		if artifacts == nil || len(artifacts) == 0 {
			continue
		}
		log.Entry().Infof("Obtained the artifacts %v for the filepath: %v", artifacts, filePath)
		fosstarsQueryServiceURL := fmt.Sprintf(sapFosstars.FosstarsQueryServiceBaseURL+queryServiceUrlSuffix, sapFosstars.RatingNameSpace, sapFosstars.RatingName, "npm")
		artifactsRatingsMap, badRatingsOccured, err := extractRatings(sapFosstars, fosstarsQueryServiceURL, invocationClient, artifacts, mySapCreateFosstarsReportOptions)
		if err != nil {
			return false, fmt.Errorf("Error while extracting Ratings. %w", err)
		}
		if badRatingsOccured {
			badRatingsOccuredOverAll = badRatingsOccured
		}
		err = os.Chdir(previousWorkingDirectory)
		if err != nil {
			return false, fmt.Errorf("Error while changing the directory %w", err)
		}
		npmName := strings.ReplaceAll(npmInfo.Name, "/", "_")
		npmName = strings.ReplaceAll(npmName, "\\", "_")
		reportName, err = fosstars.GenerateHtmlRepotForNpm(artifactsRatingsMap, npmName+":"+npmInfo.Version, fosstarsQueryServiceURL)
		if err != nil {
			return false, fmt.Errorf("Error while generating Html Report %w", err)
		}
		replacer := strings.NewReplacer(checkOutPath, "", "\\", "/")
		buildDescriptorRelativePath := replacer.Replace(filePath)
		dependencyTreeGenerator := fosstars.DependencyTreeGenerator{}
		generateNPMDependencyTreeRatingsJson(sapFosstars, artifactsRatingsMap, &dependencyTreeGenerator, reportName, npmName, npmInfo, buildDescriptorRelativePath[1:])
		generateJsonReport(sapFosstars, invocationClient, reportName, mySapCreateFosstarsReportOptions)
	}

	return badRatingsOccuredOverAll, nil
}

func getMtaBuildDescriptor(sapFosstars *fosstars.Fosstars) string {
	if isBuildDescriptorProvided(sapFosstars) {
		return sapFosstars.BuildDescriptor
	}
	return mtaBuildDescriptorFileName
}

func extractRatingsForMtaProjects(sapFosstars *fosstars.Fosstars, invocationClient fosstars.InvocationClientInterface, checkOutPath string, previousWorkingDirectory string, mySapCreateFosstarsReportOptions sapCreateFosstarsReportOptions) (bool, error) {
	mtaPath := filepath.Join(checkOutPath, getMtaBuildDescriptor(sapFosstars))
	yamlInfo, err := fosstars.GetYamlInfo(checkOutPath, mtaPath)
	if err != nil {
		log.Entry().WithError(err).Fatal("Error while extracting ratings for the MTA project ")
	}

	var badRatingsOccuredOverAll bool
	for _, mavenModulePath := range yamlInfo.MavenModulePaths {
		badRatingsOccured, err := extractRatingsForMavenProjects(sapFosstars, invocationClient, mavenModulePath, previousWorkingDirectory, &fosstars.DependencyTreeGenerator{}, mySapCreateFosstarsReportOptions)
		if err != nil {
			log.Entry().WithError(err).Fatal("Error while extracting ratings for the Maven project ")
		}
		if badRatingsOccured {
			badRatingsOccuredOverAll = badRatingsOccured
		}
	}
	for _, npmModulePath := range yamlInfo.NpmModulePaths {
		badRatingsOccured, err := extractRatingsForNPMProjects(sapFosstars, invocationClient, npmModulePath, previousWorkingDirectory, mySapCreateFosstarsReportOptions)
		if err != nil {
			log.Entry().WithError(err).Fatal("Error while extracting ratings for the NPM project ")
		}
		if badRatingsOccured {
			badRatingsOccuredOverAll = badRatingsOccured
		}
	}
	return badRatingsOccuredOverAll, nil
}

func extractRatings(sapFosstars *fosstars.Fosstars, fosstarsQueryServiceURL string, invocationClient fosstars.InvocationClientInterface, artifacts []string, mySapCreateFosstarsReportOptions sapCreateFosstarsReportOptions) (map[string]*fosstars.Rating, bool, error) {
	artifactsRatingsMap, badRatingsOccured, err := invocationClient.GetRatings(sapFosstars, fosstarsQueryServiceURL, artifacts, time.Second, mySapCreateFosstarsReportOptions.CustomTLSCertificateLinks)
	if err != nil {
		return nil, false, fmt.Errorf("Error while getting the ratings from invocation client %w", err)
	}
	if badRatingsOccured {
		log.Entry().Errorf("Encountered BAD ratings in buildDesceriptor %v", sapFosstars.BuildDescriptor)
	}

	return artifactsRatingsMap, badRatingsOccured, nil
}

func extractDependencyTrees(sapFosstars *fosstars.Fosstars, DependencyTreeGenerator fosstars.DependencyTreeGeneratorInterface, checkOutPath string, globalSettingsFile string, buildQuality string) ([]string, error) {
	var pomFilePath string
	if len(sapFosstars.BuildDescriptor) > 0 && isMavenBuildDescriptor(sapFosstars) {
		log.Entry().Infof("As the build descriptor file: %v is mentioned, considering it.", sapFosstars.BuildDescriptor)
		pomFilePath = filepath.Join(checkOutPath, sapFosstars.BuildDescriptor)
	} else {
		log.Entry().Info("As the dependency management file is not mentioned, will try to get pom file from project root folder.")
		pomFilePath = filepath.Join(checkOutPath, mavenBuildDescriptorFileName)

		fileExist, err := piperutils.FileExists(pomFilePath)
		if !fileExist || err != nil {
			return nil, fmt.Errorf("root pom file not found at '%v'", pomFilePath)
		}
	}
	return DependencyTreeGenerator.GetDependencyTreeFiles(pomFilePath, globalSettingsFile, buildQuality)
}

func extractArtifacts(sapFosstars *fosstars.Fosstars, dependencyTreeFilePath string) ([]string, error) {

	// Modify the dependency tree file to include only direct dependnecies if opted
	if !sapFosstars.IncludeTransitiveDependency {
		err := fosstars.ExcludeTransitiveDependencies(dependencyTreeFilePath)
		if err != nil {
			return nil, fmt.Errorf("Error while filtering direct dependency %w", err)
		}
	}

	filteredLines, err := fosstars.GetFileContentAsList(dependencyTreeFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not get dependency tree file content as list: %w", err)
	}

	artifacts := []string{}
	for _, line := range filteredLines {
		artifacts = append(artifacts, fosstars.GetArtifactFromLine(line))
	}

	// Filtering dependency trees which has no entry and no dependencies
	if len(artifacts) < 2 {
		log.Entry().Infof("There are no dependencies in dependency tree file: %v", dependencyTreeFilePath)
		return nil, nil
	}

	return artifacts, nil
}

func extractNpmInfo(checkOutPath string, sapFosstars *fosstars.Fosstars) (map[string]*fosstars.NpmInfo, error) {
	filePathsToArtifactsMap := make(map[string]*fosstars.NpmInfo)
	if len(sapFosstars.BuildDescriptor) > 0 && isNpmBuildDescriptor(sapFosstars) {
		log.Entry().Infof("As the build descriptor file: %v is mentioned, considering it.", sapFosstars.BuildDescriptor)
		packageJsonFilePath := filepath.Join(checkOutPath, sapFosstars.BuildDescriptor)
		npmInfo, err := sapFosstars.GetNpmInfo(packageJsonFilePath)
		if err != nil {
			return nil, err
		}
		filePathsToArtifactsMap[packageJsonFilePath] = &npmInfo

	} else {
		log.Entry().Info("As the dependency management file is not mentioned, will try to get package.json from project root folder.")
		packageJsonFilePath := filepath.Join(checkOutPath, npmBuildDescriptorFileName)
		npmInfo, err := sapFosstars.GetNpmInfo(packageJsonFilePath)
		if err != nil {
			return nil, err
		}
		filePathsToArtifactsMap[packageJsonFilePath] = &npmInfo
	}
	return filePathsToArtifactsMap, nil
}

func generateJsonReport(sapFosstars *fosstars.Fosstars, invocationClient fosstars.InvocationClientInterface, reportName string, mySapCreateFosstarsReportOptions sapCreateFosstarsReportOptions) error {
	modelratingsDefinitionDetails, err := invocationClient.GetModelRatingDefinitionDetails(sapFosstars.FosstarsQueryServiceBaseURL+modelRatingsDefinitionsEndpoint, sapFosstars.FosstarsClientSuffix, mySapCreateFosstarsReportOptions.CustomTLSCertificateLinks)
	if err != nil {
		return fmt.Errorf("Could not get ModelRatingDefinitionDefails %w", err)
	}
	allRatingsMap := invocationClient.GetAllRatingsFromFosstars()
	for _, rating := range allRatingsMap {
		ratingsMap := rating.(map[string]interface{})
		ratingsMap["parentIdentifier"] = nil
	}
	return fosstars.CreateJsonReport(modelratingsDefinitionDetails, allRatingsMap, reportName)
}

func generateJsonReportMaven(sapFosstars *fosstars.Fosstars, invocationClient fosstars.InvocationClientInterface, reportName string, depGraphFilePath string, pomArtifact string, mySapCreateFosstarsReportOptions sapCreateFosstarsReportOptions) error {
	modelratingsDefinitionDetails, err := invocationClient.GetModelRatingDefinitionDetails(sapFosstars.FosstarsQueryServiceBaseURL+modelRatingsDefinitionsEndpoint, sapFosstars.FosstarsClientSuffix, mySapCreateFosstarsReportOptions.CustomTLSCertificateLinks)
	if err != nil {
		return fmt.Errorf("Could not get ModelRatingDefinitionDefails %w", err)
	}
	allRatingsMap := invocationClient.GetAllRatingsFromFosstars()
	childParentMap, err := fosstars.GetChildParentRelationMap(depGraphFilePath, pomArtifact)
	if err != nil {
		return fmt.Errorf("Could not get maven child parent dependency relation map %w", err)
	}

	for artifact, rating := range allRatingsMap {
		if parentDependency, found := childParentMap[artifact]; found {
			ratingsMap := rating.(map[string]interface{})
			ratingsMap["parentIdentifier"] = parentDependency
		}
	}

	delete(allRatingsMap, pomArtifact)
	return fosstars.CreateJsonReport(modelratingsDefinitionDetails, allRatingsMap, reportName)
}

func generateMavenDependencyTreeRatingsJson(sapFosstars *fosstars.Fosstars, artifactsRatingsMap map[string]*fosstars.Rating, dependencyTreeGenerator fosstars.DependencyTreeGeneratorInterface, depGraphFilePath string, reportName string, pomArtifact string, buildDescriptorRelativePath string) error {
	log.Entry().Infof("Started reading the dependency graph %v to store in dependency tree with ratings", depGraphFilePath)
	parent, err := dependencyTreeGenerator.ParseDependencyTreeFile(depGraphFilePath, artifactsRatingsMap, pomArtifact, sapFosstars.ExcludedLibraries, sapFosstars.ExcludeSAPInternalLibraries, sapFosstars.ExcludeTestDevDependencies)
	if err != nil {
		return fmt.Errorf("Could not get dependency Graph %w", err)
	}
	ratingDetailsJson, _ := json.Marshal(parent)
	ratingDetailsJsonWithFilePath, err := dependencyTreeGenerator.AddRelativeBuidDescriptorPathToJson(ratingDetailsJson, buildDescriptorRelativePath)
	if err != nil {
		return err
	}
	return fosstars.WriteJsonToFile(reportName+"_dependency_tree", ratingDetailsJsonWithFilePath)
}

func generateNPMDependencyTreeRatingsJson(sapFosstars *fosstars.Fosstars, artifactsRatingsMap map[string]*fosstars.Rating, dependencyTreeGenerator fosstars.DependencyTreeGeneratorInterface, reportName string, rootArtifact string, npmInfo *fosstars.NpmInfo, buildDescriptorRelativePath string) error {
	parent, err := dependencyTreeGenerator.GetDependencyTreeForNPM(artifactsRatingsMap, rootArtifact, sapFosstars.ExcludedLibraries, sapFosstars.ExcludeSAPInternalLibraries, sapFosstars.ExcludeTestDevDependencies, npmInfo.DevDependencies)
	if err != nil {
		return fmt.Errorf("Could not get dependency Graph %w", err)
	}
	ratingDetailsJson, _ := json.Marshal(parent)
	ratingDetailsJsonWithFilePath, err := dependencyTreeGenerator.AddRelativeBuidDescriptorPathToJson(ratingDetailsJson, buildDescriptorRelativePath)
	if err != nil {
		return err
	}
	return fosstars.WriteJsonToFile(reportName+"_dependency_tree", ratingDetailsJsonWithFilePath)
}

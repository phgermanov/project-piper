package fosstars

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/maven"
	"github.com/bmatcuk/doublestar"
	"github.com/vifraa/gopom"
)

const mavenBuildDescriptor = "pom.xml"

// Maven dependencies parent child relation info
type DependenciesInfo struct {
	From        string
	To          string
	NumericFrom int
	NumericTo   int
}

// Maven dependencies parent child relation info
type ArtifactInfo struct {
	Id        string
	NumericId int
	Scopes    []string
	Version   string
}

// Maven dependency graph info
type DependencyGraph struct {
	Dependencies []DependenciesInfo
	Artifacts    []ArtifactInfo
}

// DependencyTreeGeneratorInterface is an interface to mock DependencyTreeGenerator
type DependencyTreeGeneratorInterface interface {
	GetDependencyTreeFiles(pomFilePath string, globalSettingsFile string, buildQuality string) ([]string, error)
	ParseDependencyTreeFile(depGraphJsonFilePath string, artifactsRatingsMap map[string]*Rating, pomArtifact string, excludedLibraries []string, excludeSAPInternalLibraries bool, excludeTestDevDependencies bool) (*DependencyTree, error)
	GetDependencyTreeForNPM(artifactsRatingsMap map[string]*Rating, rootArtifact string, excludedLibraries []string, excludeSAPInternalLibraries bool, excludeTestDevDependencies bool, devDependencies []string) (*DependencyTree, error)
	AddRelativeBuidDescriptorPathToJson(ratingDetailsJson []byte, buildDescriptorRelativePath string) ([]byte, error)
	GetAllPomFilesToBuildDependencyTree(checkOutPath string) ([]string, error)
}

type DependencyTreeGenerator struct {
	artifactToDependencyTreeMap map[string]*DependencyTree
}

type DependencyTree struct {
	Artifact           string
	Version            string
	Value              float64
	Label              string
	Confidence         float32
	Created            string
	RatingDefinitionId string
	RatingId           string
	Children           map[string]*DependencyTree
	Parent             string
	Excluded           bool
}

func (dependencyTreeGenerator *DependencyTreeGenerator) ParseDependencyTreeFile(depGraphJsonFilePath string, artifactsRatingsMap map[string]*Rating, pomArtifact string, excludedLibraries []string, excludeSAPInternalLibraries bool, excludeTestDevDependencies bool) (*DependencyTree, error) {
	var parent *DependencyTree
	pContent, err := os.ReadFile(depGraphJsonFilePath)
	if err != nil {
		return parent, err
	}
	dependencyTreeGenerator.artifactToDependencyTreeMap = make(map[string]*DependencyTree)
	var dependencyGraph DependencyGraph
	json.Unmarshal(pContent, &dependencyGraph)
	devDependencies, numericIdsMap := getDevDependenciesAndNumericIdsMap(dependencyGraph.Artifacts)
	for _, dependencyInfo := range dependencyGraph.Dependencies {
		to := strings.ReplaceAll(dependencyInfo.To, ":jar", "")
		to = strings.ReplaceAll(to, ":", "/")
		from := strings.ReplaceAll(dependencyInfo.From, ":jar", "")
		from = strings.ReplaceAll(from, ":", "/")
		from = strings.TrimSuffix(from, "/pom")
		to = strings.TrimSuffix(to, "/pom")
		fromVersion := numericIdsMap[dependencyInfo.From].Version
		toVersion := numericIdsMap[dependencyInfo.To].Version
		if pomArtifact == from && parent == nil {
			parent = &DependencyTree{Artifact: from, Version: fromVersion, Children: make(map[string]*DependencyTree)}
			toArtifact := getToArtifact(dependencyTreeGenerator, to, toVersion, artifactsRatingsMap, excludedLibraries, excludeSAPInternalLibraries)
			checkDevDependencies(toArtifact, devDependencies, dependencyInfo.NumericTo, excludeTestDevDependencies)
			parent.Children[to] = toArtifact
			toArtifact.Parent = from
			dependencyTreeGenerator.artifactToDependencyTreeMap[from] = parent
			dependencyTreeGenerator.artifactToDependencyTreeMap[to] = toArtifact
		}
		if dependencyTreeGenerator.artifactToDependencyTreeMap[from] == nil {
			fromArtifact := &DependencyTree{Artifact: from, Version: fromVersion, Children: make(map[string]*DependencyTree)}
			updateArtifactRating(fromArtifact, from, artifactsRatingsMap, excludedLibraries, excludeSAPInternalLibraries)
			checkDevDependencies(fromArtifact, devDependencies, dependencyInfo.NumericFrom, excludeTestDevDependencies)
			toArtifact := getToArtifact(dependencyTreeGenerator, to, toVersion, artifactsRatingsMap, excludedLibraries, excludeSAPInternalLibraries)
			checkDevDependencies(toArtifact, devDependencies, dependencyInfo.NumericTo, excludeTestDevDependencies)
			fromArtifact.Children[to] = toArtifact
			toArtifact.Parent = from
			dependencyTreeGenerator.artifactToDependencyTreeMap[from] = fromArtifact
			dependencyTreeGenerator.artifactToDependencyTreeMap[to] = toArtifact
		} else {
			fromArtifact := dependencyTreeGenerator.artifactToDependencyTreeMap[from]
			toArtifact := getToArtifact(dependencyTreeGenerator, to, toVersion, artifactsRatingsMap, excludedLibraries, excludeSAPInternalLibraries)
			checkDevDependencies(toArtifact, devDependencies, dependencyInfo.NumericTo, excludeTestDevDependencies)
			fromArtifact.Children[to] = toArtifact
			toArtifact.Parent = from
		}
	}
	return parent, nil
}

func checkDevDependencies(artifact *DependencyTree, devDependencies []int, dependencyInfo int, excludeTestDevDependencies bool) {
	if excludeTestDevDependencies && len(devDependencies) > 0 && slices.Contains(devDependencies, dependencyInfo) {
		artifact.Excluded = true
	}
}

func (dependencyTreeGenerator *DependencyTreeGenerator) GetDependencyTreeForNPM(artifactsRatingsMap map[string]*Rating, rootArtifact string, excludedLibraries []string, excludeSAPInternalLibraries bool, excludeTestDevDependencies bool, devDependencies []string) (*DependencyTree, error) {
	var parent *DependencyTree
	dependencyTreeGenerator.artifactToDependencyTreeMap = make(map[string]*DependencyTree)
	parent = &DependencyTree{Artifact: rootArtifact, Children: make(map[string]*DependencyTree)}
	for artifact, _ := range artifactsRatingsMap {
		toArtifact := &DependencyTree{Artifact: artifact, Children: make(map[string]*DependencyTree)}
		updateArtifactRating(toArtifact, artifact, artifactsRatingsMap, excludedLibraries, excludeSAPInternalLibraries)
		if excludeTestDevDependencies && len(devDependencies) > 0 && slices.Contains(devDependencies, strings.ReplaceAll(artifact, "/", ":")) {
			toArtifact.Excluded = true
		}
		parent.Children[artifact] = toArtifact
		toArtifact.Parent = rootArtifact
		dependencyTreeGenerator.artifactToDependencyTreeMap[artifact] = toArtifact
	}
	return parent, nil
}

func (f *DependencyTreeGenerator) GetDependencyTreeFiles(pomFilePath string, globalSettingsFile string, buildQuality string) ([]string, error) {
	log.Entry().Infof("Started to get all dependencies from POM File: %v with globalSettingsFile %v", pomFilePath, globalSettingsFile)
	log.Entry().Infof("buildQuality mentioned is: %v", buildQuality)
	parentDir, err := generateDependencyTree(pomFilePath, globalSettingsFile, buildQuality)
	if err != nil {
		return nil, err
	}

	err = generateDependencyGraph(pomFilePath, globalSettingsFile, buildQuality)
	if err != nil {
		return nil, err
	}

	dependencyTreeFiles, err := getFilesInPath(parentDir, "fosstar-generated-tree.txt")
	if err != nil {
		return nil, err
	}

	log.Entry().Debugf("Generated dependency tree files: %v", dependencyTreeFiles)
	return dependencyTreeFiles, nil
}

func generateDependencyTree(pomFilePath string, globalSettingsFile string, buildQuality string) (string, error) {
	log.Entry().Infof("Started dependency tree generation for: %v, with globalSettingsFile %v", pomFilePath, globalSettingsFile)
	utils := maven.NewUtilsBundle()
	flags := getProfileFlags(buildQuality)

	executeOptions := maven.ExecuteOptions{PomPath: pomFilePath,
		Goals:              []string{"install", "org.apache.maven.plugins:maven-dependency-plugin:3.1.2:tree"},
		Defines:            []string{"-DoutputFile=fosstar-generated-tree.txt", "-Dmaven.test.skip=true"},
		GlobalSettingsFile: globalSettingsFile,
		Flags:              flags,
	}
	_, err := maven.Execute(&executeOptions, utils)
	if err != nil {
		return "", fmt.Errorf("Could not generate dependency tree: %w", err)
	}

	parentDir := filepath.Dir(pomFilePath)
	log.Entry().Infof("Dependency tree generated for: %v", pomFilePath)
	return parentDir, nil
}

func getFilesInPath(checkOutPath string, fileName string) ([]string, error) {
	matches, err := doublestar.Glob(checkOutPath + "**/**/" + fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to search for files %w", err)
	}
	return matches, nil
}

func generateDependencyGraph(pomFilePath string, globalSettingsFile string, buildQuality string) error {
	log.Entry().Infof("Started dependency graph generation for: %v", pomFilePath)
	utils := maven.NewUtilsBundle()
	flags := getProfileFlags(buildQuality)

	executeOptions := maven.ExecuteOptions{PomPath: pomFilePath,
		Goals:              []string{"com.github.ferstl:depgraph-maven-plugin:3.3.0:graph"},
		Defines:            []string{"-DgraphFormat=json"},
		GlobalSettingsFile: globalSettingsFile,
		Flags:              flags,
	}
	_, err := maven.Execute(&executeOptions, utils)
	if err != nil {
		return fmt.Errorf("Could not generate dependency graph: %w", err)
	}

	log.Entry().Infof("Dependency graph generated for: %v", pomFilePath)
	return nil
}

func GetChildParentRelationMap(depGraphJsonFilePath string, pomArtifact string) (map[string]interface{}, error) {
	childParentMap := make(map[string]interface{})

	pContent, err := os.ReadFile(depGraphJsonFilePath)
	if err != nil {
		return childParentMap, err
	}

	var dependencyGraph DependencyGraph
	json.Unmarshal(pContent, &dependencyGraph)

	for _, dependencyInfo := range dependencyGraph.Dependencies {
		to := strings.ReplaceAll(dependencyInfo.To, ":jar", "")
		to = strings.ReplaceAll(to, ":", "/")
		from := strings.ReplaceAll(dependencyInfo.From, ":jar", "")
		from = strings.ReplaceAll(from, ":", "/")
		if from == pomArtifact+"/pom" {
			childParentMap[to] = nil
		} else {
			childParentMap[to] = from
		}
	}

	log.Entry().Infof("Got parent child map for  %v", depGraphJsonFilePath)
	return childParentMap, nil
}

func GetPomArtifact(dependencyTree []string) string {
	parentArtifactLine := dependencyTree[0]
	return GetArtifactFromLine(parentArtifactLine)
}

func ExcludeTransitiveDependencies(dependencyTreeFilePath string) error {
	dependencyTreeList, _ := GetFileContentAsList(dependencyTreeFilePath)
	return WriteToFile(dependencyTreeFilePath, GetDirectDependencies(dependencyTreeList))
}

func GetDirectDependencies(dependencyTreeList []string) []string {
	directDependencies := []string{}
	for index, line := range dependencyTreeList {
		if index == 0 || strings.HasPrefix(line, "+-") {
			directDependencies = append(directDependencies, line)
		}
	}
	return directDependencies
}

func updateArtifactRating(artifactDependencyTree *DependencyTree, artifact string, artifactsRatingsMap map[string]*Rating, excludedLibraries []string, excludeSAPInternalLibraries bool) {
	if artifactsRatingsMap[artifact] != nil {
		artifactDependencyTree.Value = artifactsRatingsMap[artifact].Value
		artifactDependencyTree.Label = artifactsRatingsMap[artifact].Label
		artifactDependencyTree.Created = artifactsRatingsMap[artifact].Created
		artifactDependencyTree.Confidence = artifactsRatingsMap[artifact].Confidence
		artifactDependencyTree.RatingDefinitionId = artifactsRatingsMap[artifact].RatingDefinitionId
		artifactDependencyTree.RatingId = artifactsRatingsMap[artifact].Id
	}
	if len(excludedLibraries) > 0 && slices.Contains(excludedLibraries, strings.ReplaceAll(artifact, "/", ":")) {
		artifactDependencyTree.Excluded = true
	} else {
		artifactDependencyTree.Excluded = false
	}
	if excludeSAPInternalLibraries && len(artifactDependencyTree.Label) == 0 && (strings.HasPrefix(artifact, "com.sap") || strings.HasPrefix(artifact, "@sap")) {
		artifactDependencyTree.Excluded = true
	}
}

func getToArtifact(dependencyTreeGenerator *DependencyTreeGenerator, to string, toVersion string, artifactsRatingsMap map[string]*Rating, excludedLibraries []string, excludeSAPInternalLibraries bool) *DependencyTree {
	var toArtifact *DependencyTree
	if dependencyTreeGenerator.artifactToDependencyTreeMap[to] == nil {
		toArtifact = &DependencyTree{Artifact: to, Version: toVersion, Children: make(map[string]*DependencyTree)}
		dependencyTreeGenerator.artifactToDependencyTreeMap[to] = toArtifact
		updateArtifactRating(toArtifact, to, artifactsRatingsMap, excludedLibraries, excludeSAPInternalLibraries)
	} else {
		toArtifact = dependencyTreeGenerator.artifactToDependencyTreeMap[to]
	}
	return toArtifact
}

func getDevDependenciesAndNumericIdsMap(artifacts []ArtifactInfo) ([]int, map[string]ArtifactInfo) {
	var devDepndencies []int
	numericIdsMap := make(map[string]ArtifactInfo)
	for _, artifactInfo := range artifacts {
		if slices.Contains(artifactInfo.Scopes, "test") {
			devDepndencies = append(devDepndencies, artifactInfo.NumericId)
		}
		numericIdsMap[artifactInfo.Id] = artifactInfo
	}
	return devDepndencies, numericIdsMap

}

func (f *DependencyTreeGenerator) AddRelativeBuidDescriptorPathToJson(ratingDetailsJson []byte, buildDescriptorRelativePath string) ([]byte, error) {
	var toAddFilePath map[string]interface{}
	err := json.Unmarshal(ratingDetailsJson, &toAddFilePath)
	if err != nil {
		return nil, fmt.Errorf("Could not add file path info to root dependency tree %w", err)
	}
	toAddFilePath["filePath"] = buildDescriptorRelativePath
	ratingDetailsJsonNew, err := json.Marshal(toAddFilePath)
	if err != nil {
		return nil, fmt.Errorf("Could not marshal dependency tree json after file path addition %w", err)
	}
	return ratingDetailsJsonNew, nil
}

func (f *DependencyTreeGenerator) GetAllPomFilesToBuildDependencyTree(checkOutPath string) ([]string, error) {
	allPomFilesPath, err := getFilesInPath(checkOutPath, mavenBuildDescriptor)
	if err != nil {
		return nil, fmt.Errorf("Could not get all pom files from the checkout directory: %w", err)
	}
	allPomFilePathsInSubModules := []string{}
	for _, pomFilePath := range allPomFilesPath {
		subModulePomFilePaths, err := getSubModulePomFilePaths(pomFilePath, checkOutPath)
		if err != nil {
			return nil, err
		}
		allPomFilePathsInSubModules = append(allPomFilePathsInSubModules, subModulePomFilePaths...)
	}

	return difference(allPomFilesPath, allPomFilePathsInSubModules), nil
}

func getSubModulePomFilePaths(pomFilePath string, checkOutPath string) ([]string, error) {
	pomBytes, err := os.ReadFile(pomFilePath)
	if err != nil {
		return nil, err
	}

	var project gopom.Project
	if err := xml.Unmarshal(pomBytes, &project); err != nil {
		return nil, fmt.Errorf("Could not unmarshal pom file: %w", err)
	}

	subModulePomFilePaths := []string{}

	for _, module := range project.Modules {
		subModulePomFilePaths = append(subModulePomFilePaths, filepath.Join(checkOutPath, module, mavenBuildDescriptor))
	}

	for _, profile := range project.Profiles {
		for _, module := range profile.Modules {
			subModulePomFilePaths = append(subModulePomFilePaths, filepath.Join(checkOutPath, module, mavenBuildDescriptor))
		}
	}

	return subModulePomFilePaths, nil
}

// difference returns the elements in `a` that aren't in `b`.
func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func getProfileFlags(buildQuality string) []string {
	var flags []string
	if len(buildQuality) > 0 {
		if buildQuality == "Milestone" {
			flags = []string{"--activate-profiles", strings.Join([]string{"!snapshot.build", "milestone.build"}, ",")}
		} else {
			flags = []string{"--activate-profiles", strings.Join([]string{"!snapshot.build", "!milestone.build", "release.build"}, ",")}
		}
	}
	return flags
}

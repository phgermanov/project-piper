package fosstars

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
	piperutils "github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"gopkg.in/yaml.v3"
)

// Fosstars defines the information required for connections to Fosstars system
type Fosstars struct {
	BuildDescriptor             string
	SCMUrl                      string
	Branch                      string
	FosstarsQueryServiceBaseURL string
	FosstarsClientSuffix        string
	RatingNameSpace             string
	RatingName                  string
	RatingLabelThreshold        string
	RatingValueThreshold        int
	ExcludedLibraries           []string
	IncludeTransitiveDependency bool
	FailOnUnclearRatings        bool
	RequestedRetryCount         int
	ExcludeSAPInternalLibraries bool
	ExcludeTestDevDependencies  bool
	BuildAllMavenPomFiles       bool
}

// NpmInfo defines the information about the parsed package.json
type NpmInfo struct {
	Name            string
	Version         string
	Artifacts       []string
	Dependencies    []string
	DevDependencies []string
}

// YamlInfo defines the information about the parsed mta yaml file
type YamlInfo struct {
	MavenModulePaths []string
	NpmModulePaths   []string
}

// ValidateInput validates that all parameters are set correctly
func (f *Fosstars) ValidateInput() error {
	if err := validateStringNotEmpty(f.SCMUrl, "SCMUrl"); err != nil {
		return err
	}
	if err := validateStringNotEmpty(f.Branch, "Branch"); err != nil {
		return err
	}
	return nil
}

func validateStringNotEmpty(value string, name string) error {
	if value == "" {
		err := errors.New(name + " must not be empty")
		return err
	}
	return nil
}

func (f *Fosstars) CloneRepo(username, password, checkOutPath string, duration time.Duration) error {
	// Create a custom http(s) client with your config. This is needed to access internal sap repos
	customClient := &http.Client{
		// accept any certificate (might be useful for testing)
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},

		// 2 minutes timeout usually other wise 2 times the durations
		Timeout: 2 * duration,

		// don't follow redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Override http(s) default protocol to use our custom client
	client.InstallProtocol("https", githttp.NewClient(customClient))
	// Clones the repository into the given dir, just as a normal git clone does
	_, err := git.PlainClone(checkOutPath, false, &git.CloneOptions{
		Auth:              &githttp.BasicAuth{Username: username, Password: password},
		URL:               f.SCMUrl,
		Progress:          os.Stdout,
		ReferenceName:     plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", f.Branch)),
		SingleBranch:      true,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})

	if err != nil {
		return err
	}

	return nil
}

func GetFileContentAsList(filePath string) ([]string, error) {
	log.Entry().Infof("Started reading file: %v", filePath)

	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	allLines := []string{}
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		allLines = append(allLines, fileScanner.Text())
	}

	return allLines, nil
}

func GetArtifactFromLine(line string) string {
	// To remove extra spaces and symbols like '+', '\', '|' from maven dependency tree
	replacer := strings.NewReplacer(" ", "", "+", "", "\\", "", "|", "")
	output := replacer.Replace(line)
	split := strings.Split(output, ":")

	groupId := split[0]
	artifactId := split[1]

	//Check if the first character is '-' and remove it
	if strings.HasPrefix(groupId, "-") {
		groupId = groupId[1:]
	}

	return groupId + "/" + artifactId
}

func (f *Fosstars) GetNpmInfo(filePath string) (NpmInfo, error) {
	p := piperutils.Files{}
	artifacts := []string{}
	var result map[string]interface{}
	var npmInfo NpmInfo
	pContent, err := p.FileRead(filePath)
	if err != nil {
		return npmInfo, err
	}

	json.Unmarshal(pContent, &result)
	if result["dependencies"] != nil {
		dependencies := getNpmArtifacts(result["dependencies"])
		artifacts = append(artifacts, dependencies...)
		npmInfo.Dependencies = dependencies
	}

	if result["devDependencies"] != nil {
		devDependencies := getNpmArtifacts(result["devDependencies"])
		artifacts = append(artifacts, devDependencies...)
		npmInfo.DevDependencies = devDependencies
	}
	npmInfo.Name = result["name"].(string)
	npmInfo.Version = result["version"].(string)
	npmInfo.Artifacts = artifacts

	return npmInfo, nil
}

func getNpmArtifacts(dependencies interface{}) []string {
	artifacts := []string{}
	artifactVersionMap := dependencies.(map[string]interface{})
	for artifact, version := range artifactVersionMap {
		log.Entry().Infof("The artifact extracted is %v and with the version %v", artifact, version)
		artifacts = append(artifacts, artifact)
	}
	return artifacts
}

func WriteToFile(filePath string, lines []string) error {
	file, err := os.Create(filepath.Clean(filePath))
	if err != nil {
		return err
	}
	defer file.Close()
	for _, line := range lines {
		fmt.Fprintln(file, line) // print content to file, one per line
	}
	return nil
}

func GetYamlInfo(checkOutPath string, yamlFilePath string) (YamlInfo, error) {
	// Read MTA file
	content, err := os.ReadFile(filepath.Clean(yamlFilePath))
	if err != nil {
		return YamlInfo{}, fmt.Errorf("Error while reading MTA file %w", err)
	}

	mta := make(map[string]interface{})
	err = yaml.Unmarshal(content, &mta)
	if err != nil {
		return YamlInfo{}, fmt.Errorf("Error while extracting content from MTA file %w", err)
	}

	mavenModulePaths := []string{}
	npmModulePaths := []string{}

	if mta["modules"] != nil {
		modules, _ := mta["modules"].([]interface{})
		for _, module := range modules {
			path, _ := module.(map[string]interface{})["path"]
			if path != nil {
				modulePath := fmt.Sprintf("%v", path)
				if isMavenDescriptorAvailable(checkOutPath, modulePath) {
					mavenModulePaths = append(mavenModulePaths, filepath.Join(checkOutPath, modulePath))
				} else if isNpmDescriptorAvailable(checkOutPath, modulePath) {
					npmModulePaths = append(npmModulePaths, filepath.Join(checkOutPath, modulePath))
				}
			}
		}
	}

	return YamlInfo{MavenModulePaths: mavenModulePaths, NpmModulePaths: npmModulePaths}, nil
}

func isMavenDescriptorAvailable(checkOutPath string, modulePath string) bool {
	pomFilePath := filepath.Join(checkOutPath, modulePath, "pom.xml")
	fileExist, _ := piperutils.FileExists(pomFilePath)
	return fileExist
}

func isNpmDescriptorAvailable(checkOutPath string, modulePath string) bool {
	npmDescriptorFilePath := filepath.Join(checkOutPath, modulePath, "package.json")
	fileExist, _ := piperutils.FileExists(npmDescriptorFilePath)
	return fileExist
}

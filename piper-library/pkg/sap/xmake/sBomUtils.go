package xmake

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/SAP/jenkins-library/pkg/docker"
	"github.com/pkg/errors"
)

func fetchSBomXML(url string, username string, password string) ([]byte, error) {
	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Add basic authentication header
	auth := username + ":" + password
	b64auth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+b64auth)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		errorMessage := fmt.Sprintf("HTTP request failed with status code %d: %s", resp.StatusCode, resp.Status)
		return nil, errors.New(errorMessage)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return body, nil
}

func WriteSBomXmlForStageBuild(stageBom map[string]interface{}) ([][]byte, error) {
	var sbomList [][]byte

	jsonData, err := json.Marshal(stageBom)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse stageBom")
	}

	var stageRepositories map[string]stageRepository
	if err := json.Unmarshal(jsonData, &stageRepositories); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal stageRepositories")
	}

	for _, stageRepository := range stageRepositories {
		if stageRepository.Format == "raw" {
			artifactURLs := GetArtifactURLs(stageRepository)
			artifactNames := GetArtifactNames(stageRepository)

			// Process each SBOM file
			for i, artifactURL := range artifactURLs {
				data, err := fetchSBomXML(artifactURL, stageRepository.Credentials.User, stageRepository.Credentials.Password)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to fetch stageBom file from %s", artifactURL)
				}

				dir := filepath.Join(".", "sbom", stageRepository.Credentials.Repository)
				if err := os.MkdirAll(dir, 0750); err != nil {
					return nil, errors.Wrapf(err, "failed to create directory")
				}

				// Use corresponding name or fallback to index-based naming
				var fileName string
				if i < len(artifactNames) {
					fileName = artifactNames[i]
				} else {
					fileName = fmt.Sprintf("sbom_%d", i)
				}

				filepath := filepath.Join(dir, fileName)
				if err := os.WriteFile(filepath, data, 0600); err != nil {
					return nil, errors.Wrapf(err, "failed to write sbom file")
				}

				sbomList = append(sbomList, data)
			}
		}
	}

	return sbomList, nil
}

// GetArtifactURLs returns all artifact URLs from a stage repository
func GetArtifactURLs(stageRepository stageRepository) []string {
	var urls []string

	for _, component := range stageRepository.Components {
		for _, asset := range component.Assets {
			if asset != nil && asset.URL != "" {
				urls = append(urls, asset.URL)
			}
		}
	}

	// If no URLs found, return default
	if len(urls) == 0 {
		urls = append(urls, stageRepository.Credentials.Repository+"sbom")
	}

	return urls
}

// GetArtifactNames returns all artifact names from a stage repository
func GetArtifactNames(stageRepository stageRepository) []string {
	var names []string

	for _, component := range stageRepository.Components {
		for _, asset := range component.Assets {
			if asset != nil && asset.FileName != "" {
				names = append(names, asset.FileName)
			}
		}
	}

	// If no names found, return default
	if len(names) == 0 {
		names = append(names, "sbom")
	}

	return names
}

func GetImagesAndCredentialsFromStageBOM(data *map[string]interface{}) ([]string, []string, repositoryCredentials, error) {
	imageNames := []string{}
	imageNameTags := []string{}
	repoCredentials := repositoryCredentials{}

	sbomString, err := json.Marshal(data)
	if err != nil {
		return imageNames, imageNameTags, repoCredentials, errors.Wrapf(err, "Failed to assert stageBOM type")
	}

	sbom := map[string]*stageRepository{}
	err = json.Unmarshal(sbomString, &sbom)
	if err != nil {
		return imageNames, imageNameTags, repoCredentials, errors.Wrapf(err, "Failed to assert stageBOM type")
	}

	for _, repository := range sbom {

		if repository.Format != "docker" {
			continue
		}

		repoCredentials = *repository.Credentials
		for _, component := range repository.Components {
			// in case of a multi-arch build, there will be multiple components with an identical Artifact, but Image will be unique
			if component.Artifact != "" && !slices.Contains(imageNames, component.Artifact) {
				imageNames = append(imageNames, component.Artifact)
			}

			imageNameAndTag, err := docker.ContainerImageNameTagFromImage(component.Image)
			if err == nil {
				imageNameTags = append(imageNameTags, imageNameAndTag)
			}
		}
		break
	}

	if len(imageNames) == 0 {
		return imageNames, imageNameTags, repoCredentials, errors.New("No images found in sbom")
	}

	return imageNames, imageNameTags, repoCredentials, nil
}

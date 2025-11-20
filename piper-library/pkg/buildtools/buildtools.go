package buildtools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/versioning"
)

type BuildTool interface {
	GetBuildArtifacts(artifactsJsonString string) ([]versioning.Coordinates, error)
	GetPromotedArtifact(promotedUrl string, buildArtifacts []versioning.Coordinates) (versioning.Coordinates, error)
}

type Common struct{}

func (c Common) GetBuildArtifacts(artifactsJsonString string) ([]versioning.Coordinates, error) {
	var buildArtifacts BuildArtifacts
	err := json.Unmarshal([]byte(artifactsJsonString), &buildArtifacts)
	if err != nil {
		return nil, fmt.Errorf("unable to get build coordinates: %w", err)
	}
	return buildArtifacts.Coordinates, nil
}

type Maven struct {
	Common
}

func (m Maven) GetPromotedArtifact(promotedUrl string, buildArtifacts []versioning.Coordinates) (versioning.Coordinates, error) {
	for _, coordinate := range buildArtifacts {
		if strings.Contains(promotedUrl, coordinate.ArtifactID+"/") &&
			strings.Contains(promotedUrl, coordinate.Version) &&
			strings.Contains(promotedUrl, strings.Replace(coordinate.GroupID, ".", "/", -1)) {
			return coordinate, nil
		}
	}
	return versioning.Coordinates{}, fmt.Errorf("no build coordinate found for %v", promotedUrl)
}

type Npm struct {
	Common
}

func (n Npm) GetPromotedArtifact(promotedUrl string, buildArtifacts []versioning.Coordinates) (versioning.Coordinates, error) {
	for _, coordinate := range buildArtifacts {
		version := coordinate.Version
		if strings.Contains(version, "+") {
			version = strings.Split(version, "+")[0]
		}
		artifactName := coordinate.ArtifactID
		if strings.HasPrefix(coordinate.ArtifactID, "@") && strings.Contains(coordinate.ArtifactID, "/") { // this covers scoped packages
			artifactName = coordinate.ArtifactID[strings.Index(coordinate.ArtifactID, "/")+1:]
		}
		urlPart := artifactName + "-" + version
		if strings.Contains(promotedUrl, urlPart) {
			return coordinate, nil
		}
	}
	return versioning.Coordinates{}, fmt.Errorf("no build coordinate found for %v", promotedUrl)
}

type Mta struct {
	Common
}

func (m Mta) GetPromotedArtifact(promotedUrl string, buildArtifacts []versioning.Coordinates) (versioning.Coordinates, error) {
	for _, coordinate := range buildArtifacts {
		if strings.Contains(promotedUrl, strings.TrimSuffix(coordinate.ArtifactID, ".mtar")) &&
			strings.Contains(promotedUrl, coordinate.Version) &&
			strings.Contains(promotedUrl, coordinate.GroupID) {
			return coordinate, nil
		}
	}
	return versioning.Coordinates{}, fmt.Errorf("no build coordinate found for %v", promotedUrl)
}

type BuildArtifacts struct {
	Coordinates []versioning.Coordinates
}

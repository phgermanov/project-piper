package xmake

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/jenkins"
	"github.com/SAP/jenkins-library/pkg/log"
)

const BuildResultJSONFilename = "build-results.json"

const failedToFetchContent string = "Failed to fetch content of '%s'"
const failedToUnmarshalContent string = "Failed to unmarshal content of '%s'"

type StageJSON struct {
	StageBom       map[string]interface{} `json:"stage-bom"`
	ProjectArchive string                 `json:"projectArchive,omitempty"`
	StagingRepoID  string                 `json:"staging_repo_id,omitempty"`
}

type detailedStageJSON struct {
	StageBom       map[string]*stageRepository `json:"stage-bom,omitempty"`
	ProjectArchive string                      `json:"projectArchive,omitempty"`
	StagingRepoID  string                      `json:"staging_repo_id,omitempty"`
	// TagExtension        interface{}                `json:"TAG_EXTENSION"`
	// TagPrefix           interface{}                `json:"TAG_PREFIX"`
	// BuildOptions        []string                   `json:"build_options"`
	// Downstreams         map[string]interface{}     `json:"downstreams"`
	// ProjectArchiveFiles map[string]interface{}     `json:"projectArchiveFiles"`
	// ProjectVersion      string                     `json:"project_version"`
	// StageGitrepo        string                     `json:"stage_gitrepo"`
	// StageGittreeish     string                     `json:"stage_gittreeish"`
	// StageRepourl        string                     `json:"stage_repourl"`
	// VersionExtension    string                     `json:"version_extension"`
}

type stageRepository struct {
	Components  []*component           `json:"components,omitempty"`
	Credentials *repositoryCredentials `json:"credentials,omitempty"`
	Format      string                 `json:"format,omitempty"`
}

type repositoryCredentials struct {
	Password      string `json:"password,omitempty"`
	Repository    string `json:"repository,omitempty"`
	RepositoryURL string `json:"repositoryURL,omitempty"`
	User          string `json:"user,omitempty"`
}

type component struct {
	Artifact string   `json:"artifact,omitempty"`
	Assets   []*asset `json:"assets,omitempty"`
	Image    string   `json:"image,omitempty"`
	Group    string   `json:"group,omitempty"`
	Version  string   `json:"version,omitempty"`
}

type asset struct {
	Classifier   string `json:"classifier,omitempty"`
	Extension    string `json:"extension,omitempty"`
	FileName     string `json:"fileName,omitempty"`
	RelativePath string `json:"relativePath,omitempty"`
	URL          string `json:"url,omitempty"`
}

type PromoteJSON struct {
	// Downstreams map[string]interface{} `json:"downstreams"`
	PromoteBom *PromoteBom `json:"promote-bom,omitempty"`
}

// TODO: make private again once struct is no longer needed in sapXmakeExecuteBuild_test.go
type PromoteBom struct {
	// Group        string `json:"group"`
	// Released     bool   `json:"released"`
	Repositories []*Repository `json:"repositories,omitempty"`
}

// TODO: make private again once struct is no longer needed in sapXmakeExecuteBuild_test.go
type Repository struct {
	// Repository string   `json:"repository"`
	Result  []string `json:"result"`
	Success bool     `json:"success"`
	// Promoted   bool     `json:"promoted,omitempty"`
	// Reason     string   `json:"reason,omitempty"`
	Repository string           `json:"repository"`
	List       []DockerArtifact `json:"list"`
}

type DockerArtifact struct {
	Artifact string `json:"artifact"`
	Image    string `json:"image"`
	Success  bool   `json:"success"`
	Version  string `json:"version"`
}

// FetchBuildResultJSON fetches the "build-results.json" artifact from the given build.
func FetchBuildResultJSON(ctx context.Context, build jenkins.Build) (jenkins.Artifact, error) {
	// lookup build artifact
	return jenkins.FetchBuildArtifact(ctx, build, BuildResultJSONFilename)
}

func FetchStageJSON(ctx context.Context, artifact jenkins.Artifact, artifactPattern string) (StageJSON, error) {
	result := StageJSON{}
	helper := detailedStageJSON{}
	// fetch data of build artifact
	content, err := artifact.GetData(ctx)
	if err != nil {
		return result, errors.Wrapf(err, failedToFetchContent, artifact.FileName())
	}
	// parse JSON into defined structure and back to string first to drop NULL nodes.
	// see https://github.com/SAP/jenkins-library/blob/3d48364862bda12a647d4c4b30f2ce32f69b9191/vars/commonPipelineEnvironment.groovy#L225-L235
	// parse JSON data to map
	err = json.Unmarshal(content, &helper)
	if err != nil {
		return result, errors.Wrapf(err, "Failed to unmarshal content of '%s' to detailed structure", artifact.FileName())
	}

	if len(artifactPattern) > 0 {
		log.Entry().Debugf("filtering staged artifacts with pattern %s", artifactPattern)
		err = filterStagedAssets(&helper, artifactPattern)
		if err != nil {
			return result, err
		}
	}
	data, err := json.Marshal(helper)
	if err != nil {
		return result, errors.Wrapf(err, "Failed to marshal content of '%s' from detailed structure", artifact.FileName())
	}
	// parse JSON data to map
	err = json.Unmarshal(data, &result)
	if err != nil {
		return result, errors.Wrapf(err, failedToUnmarshalContent, artifact.FileName())
	}

	return result, err
}

func FetchPromoteJSON(ctx context.Context, artifact jenkins.Artifact, artifactPattern string) (PromoteJSON, error) {
	result := PromoteJSON{}
	// fetch data of build artifact
	content, err := artifact.GetData(ctx)
	if err != nil {
		return result, errors.Wrapf(err, failedToFetchContent, artifact.FileName())
	}
	// parse JSON data to map
	err = json.Unmarshal(content, &result)
	if err != nil {
		return result, errors.Wrapf(err, failedToUnmarshalContent, artifact.FileName())
	}
	if len(artifactPattern) > 0 {
		log.Entry().Debugf("filtering promoted artifacts with pattern %s", artifactPattern)
		err = filterPromotedAssets(&result, artifactPattern)
		if err != nil {
			return result, err
		}
	}
	return result, err
}

func filterStagedAssets(data *detailedStageJSON, artifactPattern string) error {
	for _, repository := range data.StageBom {
		for _, components := range repository.Components {
			assetList := []*asset{}
			for _, asset := range components.Assets {
				ok, err := doublestar.Match(artifactPattern, asset.FileName)
				if err != nil {
					return errors.Wrapf(err, "failed to match artifact")
				}
				if ok {
					log.Entry().Debugf("keeping %s", asset.URL)
					assetList = append(assetList, asset)
				} else {
					log.Entry().Debugf("ignoring %s", asset.URL)
				}
			}
			components.Assets = assetList
		}
	}
	return nil
}

func filterPromotedAssets(data *PromoteJSON, artifactPattern string) error {
	for _, repository := range data.PromoteBom.Repositories {
		assetList := []string{}
		// repository.Result
		for _, asset := range repository.Result {
			filename := filepath.Base(asset)
			ok, err := doublestar.Match(artifactPattern, filename)
			if err != nil {
				return errors.Wrapf(err, "failed to match artifact")
			}
			if ok {
				log.Entry().Debugf("keeping %s", asset)
				assetList = append(assetList, asset)
			} else {
				log.Entry().Debugf("ignoring %s", asset)
			}
		}
		repository.Result = assetList
	}
	return nil
}

// FetchBuildResultContent returns the content of a given artifact as a map.
// TODO refactor to use dedicated stageJSON type
func FetchBuildResultContent(ctx context.Context, artifact jenkins.Artifact) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	// fetch data of build artifact
	content, err := artifact.GetData(ctx)
	if err != nil {
		return result, errors.Wrapf(err, failedToFetchContent, artifact.FileName())
	}
	// parse JSON data to map
	err = json.Unmarshal(content, &result)
	if err != nil {
		return result, errors.Wrapf(err, failedToUnmarshalContent, artifact.FileName())
	}

	return result, err
}

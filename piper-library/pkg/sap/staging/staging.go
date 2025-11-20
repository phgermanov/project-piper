package staging

import (
	"os"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
)

type ServiceUtils interface {
	GetClient() *piperhttp.Client
	FileExists(filename string) (bool, error)
	FileRead(path string) ([]byte, error)
	FileWrite(path string, content []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
}

// PromotedArtifacts contains URLs of promoted artifacts
type PromotedArtifacts struct {
	PromotedArtifactURLs []string
	PromotedDockerImages []string
	PromotedHelmChartURL string
}

// Response from staging.getRequest is being unmarshaled to this struct
type StagingPromote struct {
	ResponseFromPromote ResponseFromPromote `json:"responseFromPromote"`
	State               string              `json:"state"`
}

type ResponseFromPromote struct {
	Repositories []Repository `json:"repositories"`
	Released     bool         `json:"released"`
	Group        string       `json:"group"`
}

type Repository struct {
	Result     []string         `json:"result"`
	Success    bool             `json:"success"`
	Repository string           `json:"repository"`
	List       []DockerArtifact `json:"list"`
}

type DockerArtifact struct {
	Artifact string `json:"artifact"`
	Image    string `json:"image"`
	Success  bool   `json:"success"`
	Version  string `json:"version"`
}

// Response from staging.GetGroupBom is being unmarshaled to this struct
type GroupBOMResponse struct {
	Repositories []GroupBOMRepository `json:"repositories"`
	Group        string               `json:"group"`
}

type GroupBOMRepository struct {
	Bom        GroupBOM `json:"BOM"`
	Repository string   `json:"repository"`
}

type GroupBOM struct {
	Components []GroupBOMComponent `json:"components"`
	Format     string              `json:"format"`
}

type GroupBOMComponent struct {
	Artifact string          `json:"artifact"`
	Assets   []GroupBOMAsset `json:"assets"`
	Version  string          `json:"version"`
	Group    string          `json:"group"`
}

type GroupBOMAsset struct {
	FileName     string `json:"fileName"`
	Extension    string `json:"extension"`
	RelativePath string `json:"relativePath"`
	URL          string `json:"url"`
}

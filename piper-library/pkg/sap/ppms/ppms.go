package ppms

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/pkg/errors"
)

//example links:
// https://i7p.wdf.sap.corp/odataint/borm/odataforosrcy/SoftwareComponentVersions('73555000100200008351')?$format=json
// https://i7p.wdf.sap.corp/odataint/borm/odataforosrcy/SoftwareComponentVersions('73555000100200008351')/FreeOpenSourceSoftwares?$format=json

// System defines access information to PPMS system
type System struct {
	ServerURL  string
	Username   string
	Password   string
	HTTPClient piperhttp.Sender
}

// SoftwareComponentVersion defines details of a PPMS Software Component Version (SCV)
type SoftwareComponentVersion struct {
	ID                         string                       `json:"Id"`
	Name                       string                       `json:"Name,omitempty"`
	TechnicalName              string                       `json:"TechnicalName,omitempty"`
	TechnicalRelease           string                       `json:"TechnicalRelease,omitempty"`
	FossLink                   map[string]map[string]string `json:"Foss,omitempty"`
	Foss                       []Foss
	BuildVersionsLink          map[string]map[string]string `json:"BuildVersions,omitempty"`
	BuildVersions              []BuildVersion
	ResponsiblesLink           map[string]map[string]string `json:"Responsibles,omitempty"`
	Responsibles               []Responsible
	ReviewModelRiskRatingsLink map[string]map[string]string `json:"ReviewModelRiskRatings,omitempty"`
	ReviewModelRiskRatings     []ReviewModelRiskRating
}

// Responsible defines a responsible person incl. role within PPMS
type Responsible struct {
	UserID          string `json:"UserId"`
	UserName        string `json:"UserName"`
	RoleID          string `json:"RoleId"`
	RoleName        string `json:"RoleName"`
	RoleDescription string `json:"RoleDescription"`
}

// BuildVersion defines a PPMS build version (BV)
type BuildVersion struct {
	ID                            string                       `json:"Id"`
	Name                          string                       `json:"Name,omitempty"`
	Description                   string                       `json:"Description"`
	SoftwareComponentVersionID    string                       `json:"SoftwareComponentVersionId"`
	SoftwareComponentVersionsName string                       `json:"SoftwareComponentVersionsName"`
	SortSequence                  string                       `json:"SortSequence"`
	FossLink                      map[string]map[string]string `json:"Foss,omitempty"`
	Foss                          []Foss
	ReviewModelRiskRatingsLink    map[string]map[string]string `json:"ReviewModelRiskRatings,omitempty"`
	ReviewModelRiskRatings        []ReviewModelRiskRating
}

// Foss defines details of a Free & Open Source Software (FOSS) component
type Foss struct {
	ID                string `json:"Id"`
	CodeCenterID      string `json:"CodeCenterId"`
	CodeCenterName    string `json:"CodeCenterName"`
	CodeCenterVersion string `json:"CodeCenterVersion"`
	LicenseID         string `json:"LicenseId"`
	License           string `json:"License"`
	Homepage          string `json:"Homepage"`
}

// ReviewModelRiskRating defines a risk rating entry in PPMS
type ReviewModelRiskRating struct {
	EntityID                   string `json:"EntityId"`
	ReviewModelID              string `json:"ReviewModelId"`
	ReviewModelName            string `json:"ReviewModelName"`
	InherentRiskRatingID       string `json:"InherentRiskRatingId"`
	InherentRiskRatingName     string `json:"InherentRiskRatingName"`
	ResidualRiskRatingID       string `json:"ResidualRiskRatingId"`
	ResidualRiskRatingName     string `json:"ResidualRiskRatingName"`
	CountTechnicalInstructions int    `json:"CountTechnicalInstructions"`
}

// defining constants for error logging
const failedToSendRequest string = "failed to send request to path %v"
const failedRetrievingBuildVersion string = "retrieving build versions failed"

// GetSCV returns details of a dedicated PPMS SCV
func (sys *System) GetSCV(id string) (SoftwareComponentVersion, error) {

	var scv SoftwareComponentVersion

	path := fmt.Sprintf("/odataint/borm/odataforosrcy/SoftwareComponentVersions('%v')?$format=json", id)

	content, err := sys.sendRequest(http.MethodGet, path, nil, nil)
	if err != nil {
		return scv, errors.Wrapf(err, failedToSendRequest, path)
	}

	var scvResponse map[string]*json.RawMessage
	err = json.Unmarshal(content, &scvResponse)
	if err != nil {
		return scv, errors.Wrap(err, "error unmarshalling SCV response")
	}

	err = json.Unmarshal(*scvResponse["d"], &scv)

	return scv, err
}

// GetFoss returns a list of Foss objects contained in a SCV
func (scv *SoftwareComponentVersion) GetFoss(sys *System) ([]Foss, error) {

	if scv.Foss != nil {
		return scv.Foss, nil
	}

	path := fmt.Sprintf("/odataint/borm/odataforosrcy/SoftwareComponentVersions('%v')/FreeOpenSourceSoftwares?$format=json", scv.ID)

	content, err := sys.sendRequest(http.MethodGet, path, nil, nil)
	if err != nil {
		return scv.Foss, errors.Wrapf(err, failedToSendRequest, path)
	}

	err = unmarshalPPMSResponse(&content, &scv.Foss)

	return scv.Foss, err
}

// GetBVs returns the list of build versions available for a SCV
func (scv *SoftwareComponentVersion) GetBVs(sys *System) ([]BuildVersion, error) {

	if scv.BuildVersions != nil {
		return scv.BuildVersions, nil
	}

	return scv.UpdateBVs(sys)
}

// UpdateBVs updates build versions available for a SCV and returns them
func (scv *SoftwareComponentVersion) UpdateBVs(sys *System) ([]BuildVersion, error) {

	path := fmt.Sprintf("/odataint/borm/odataforosrcy/SoftwareComponentVersions('%v')/BuildVersions?$format=json", scv.ID)

	content, err := sys.sendRequest(http.MethodGet, path, nil, nil)
	if err != nil {
		return scv.BuildVersions, errors.Wrapf(err, failedToSendRequest, path)
	}

	err = unmarshalPPMSResponse(&content, &scv.BuildVersions)

	return scv.BuildVersions, err
}

// GetReviewModelRiskRatings returns all ReviewModelRiskRatings for a SCV
func (scv *SoftwareComponentVersion) GetReviewModelRiskRatings(sys *System) ([]ReviewModelRiskRating, error) {

	if scv.ReviewModelRiskRatings != nil {
		return scv.ReviewModelRiskRatings, nil
	}

	path := fmt.Sprintf("/odataint/borm/odataforosrcy/SoftwareComponentVersions('%v')/ReviewModelRiskRatings?$format=json", scv.ID)

	content, err := sys.sendRequest(http.MethodGet, path, nil, nil)
	if err != nil {
		return scv.ReviewModelRiskRatings, errors.Wrapf(err, failedToSendRequest, path)
	}

	err = unmarshalPPMSResponse(&content, &scv.ReviewModelRiskRatings)

	return scv.ReviewModelRiskRatings, err
}

// GetResponsibles returns the list of build versions available for a SCV
func (scv *SoftwareComponentVersion) GetResponsibles(sys *System) ([]Responsible, error) {

	if scv.Responsibles != nil {
		return scv.Responsibles, nil
	}

	path := fmt.Sprintf("/odataint/borm/odataforosrcy/SoftwareComponentVersions('%v')/Responsibles?$format=json", scv.ID)
	content, err := sys.sendRequest(http.MethodGet, path, nil, nil)
	if err != nil {
		return scv.Responsibles, errors.Wrapf(err, failedToSendRequest, path)
	}

	err = unmarshalPPMSResponse(&content, &scv.Responsibles)

	return scv.Responsibles, err

}

// GetBuildVersionByName retrieves the build version based on the build version name provided
func (scv *SoftwareComponentVersion) GetBuildVersionByName(sys *System, name string) (BuildVersion, error) {
	buildVersions, err := scv.GetBVs(sys)
	if err != nil {
		return BuildVersion{}, errors.Wrap(err, failedRetrievingBuildVersion)
	}
	for _, bv := range buildVersions {
		if name == bv.Name {
			return bv, nil
		}
	}
	return BuildVersion{}, fmt.Errorf("build version with name '%v' not found", name)
}

// GetBuildVersionIDByName retrieves the build version id based on the build version name provided
func (scv *SoftwareComponentVersion) GetBuildVersionIDByName(sys *System, name string) (string, error) {
	buildVersions, err := scv.GetBVs(sys)
	if err != nil {
		return "", errors.Wrap(err, failedRetrievingBuildVersion)
	}
	for _, bv := range buildVersions {
		if name == bv.Name {
			return bv.ID, nil
		}
	}
	return "", fmt.Errorf("build version with name '%v' not found", name)
}

// GetLatestBuildVersion retrieves the latest build version of the SCV
func (scv *SoftwareComponentVersion) GetLatestBuildVersion(sys *System) (BuildVersion, error) {
	buildVersions, err := scv.GetBVs(sys)
	if err != nil {
		return BuildVersion{}, errors.Wrap(err, failedRetrievingBuildVersion)
	}

	if len(buildVersions) == 0 {
		return BuildVersion{}, fmt.Errorf("no build versions available")
	}

	sortedBVs := buildVersions
	sort.Sort(BySortSequence(sortedBVs))

	return sortedBVs[len(buildVersions)-1], nil
}

// BySortSequence contains a list of build versions to be used for sorting
type BySortSequence []BuildVersion

func (b BySortSequence) Len() int { return len(b) }
func (b BySortSequence) Less(i, j int) bool {
	first, _ := strconv.ParseInt(b[i].SortSequence, 10, 64)
	second, _ := strconv.ParseInt(b[j].SortSequence, 10, 64)
	return first < second
}
func (b BySortSequence) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

// GetFoss returns a list of Foss objects contained in a build version
func (bv *BuildVersion) GetFoss(sys *System) ([]Foss, error) {

	if bv.Foss != nil {
		return bv.Foss, nil
	}

	path := fmt.Sprintf("/odataint/borm/odataforosrcy/BuildVersions('%v')/FreeOpenSourceSoftwares?$format=json", bv.ID)

	content, err := sys.sendRequest(http.MethodGet, path, nil, nil)
	if err != nil {
		return bv.Foss, errors.Wrapf(err, failedToSendRequest, path)
	}

	err = unmarshalPPMSResponse(&content, &bv.Foss)

	return bv.Foss, err
}

// GetReviewModelRiskRatings returns all ReviewModelRiskRatings for a build version
func (bv *BuildVersion) GetReviewModelRiskRatings(sys *System) ([]ReviewModelRiskRating, error) {

	if bv.ReviewModelRiskRatings != nil {
		return bv.ReviewModelRiskRatings, nil
	}

	path := fmt.Sprintf("/odataint/borm/odataforosrcy/BuildVersions('%v')/ReviewModelRiskRatings?$format=json", bv.ID)

	content, err := sys.sendRequest(http.MethodGet, path, nil, nil)
	if err != nil {
		return bv.ReviewModelRiskRatings, errors.Wrapf(err, failedToSendRequest, path)
	}

	err = unmarshalPPMSResponse(&content, &bv.ReviewModelRiskRatings)

	return bv.ReviewModelRiskRatings, err
}

// GetFossIDs returns all IDs of a list of Foss components
func GetFossIDs(foss *[]Foss) []string {
	var r []string
	for _, f := range *foss {
		r = append(r, f.ID)
	}
	return r
}

func (sys *System) sendRequest(method, path string, body io.Reader, header http.Header) ([]byte, error) {

	opts := piperhttp.ClientOptions{Username: sys.Username, Password: sys.Password}
	sys.HTTPClient.SetOptions(opts)

	response, err := sys.HTTPClient.SendRequest(method, fmt.Sprintf("%v%v", sys.ServerURL, path), body, nil, nil)
	if err != nil {
		return []byte{}, errors.Wrapf(err, failedToSendRequest, path)
	}

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return []byte{}, errors.Wrap(err, "error reading response")
	}
	response.Body.Close()

	return content, nil
}

func unmarshalPPMSResponse(content *[]byte, target interface{}) error {
	var ppmsResponse map[string]map[string]*json.RawMessage

	err := json.Unmarshal(*content, &ppmsResponse)
	if err != nil {
		return errors.Wrap(err, "error unmarshalling Foss response")
	}

	return json.Unmarshal(*ppmsResponse["d"]["results"], target)
}

package ppms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const xCsrfHeader string = "x-csrf-token"
const failedToCreateChangeRequest string = "failed to create change request JSON"

const crbvAPIEndpint string = "/sap/internal/ppms/api/changerequest/v1/cvbv"

// ChangeRequestPartV1 defines the Schema for a manual PPMS change request using V1 of the Schema
type ChangeRequestPartV1 struct {
	SchemaVer         string                  `json:"schemaVer"`
	ChangeRequestData ChangeRequestDataPartV1 `json:"changeRequestData"`
}

// ChangeRequestPartV2 defines the Schema for an automatic PPMS change request using V2 of the Schema
type ChangeRequestPartV2 struct {
	SchemaVer         string                  `json:"schemaVer"`
	ChangeRequestData ChangeRequestDataPartV2 `json:"changeRequestData"`
}

// ChangeRequestBV defines the Schema for an automatic PPMS build version change request using V1 of the Schema
type ChangeRequestBV struct {
	SchemaVer         string              `json:"schemaVer"`
	ChangeRequestData ChangeRequestDataBV `json:"changeRequestData"`
}

// ChangeRequestDataPartV1 defines the data of a manual PPMS change request
type ChangeRequestDataPartV1 struct {
	ID                 string                    `json:"id,omitempty"`
	Scan               ChangeRequestScan         `json:"scan"`
	Comparison         ChangeRequestComparisonV1 `json:"comparison"`
	Target             ChangeRequestTarget       `json:"target"`
	Comment            string                    `json:"comment"`
	FossToAdd          []ChangeRequestFoss       `json:"fossToAdd,omitempty"`
	FossToRemove       []ChangeRequestFoss       `json:"fossToRemove,omitempty"`
	SendConfirmationTo []ChangeRequestUser       `json:"sendConfirmationTo,omitempty"`
}

// ChangeRequestDataPartV2 defines the data of an automated PPMS change request
type ChangeRequestDataPartV2 struct {
	ID                 string                `json:"id,omitempty"`
	Provider           ChangeRequestProvider `json:"provider,omitempty"`
	Scan               ChangeRequestScan     `json:"scan"`
	Target             ChangeRequestTarget   `json:"target"`
	Comment            string                `json:"comment,omitempty"`
	FossComprised      []ChangeRequestFoss   `json:"fossComprised"`
	SendConfirmationTo []ChangeRequestUser   `json:"sendConfirmationTo,omitempty"`
}

// ChangeRequestDataBV defines the data of an automated PPMS build version change request
type ChangeRequestDataBV struct {
	ID           string                    `json:"id,omitempty"`
	Provider     ChangeRequestProvider     `json:"provider,omitempty"`
	Target       ChangeRequestTarget       `json:"target"`
	BuildVersion ChangeRequestBuildVersion `json:"buildVersion"`
	Options      ChangeRequestOptions      `json:"options,omitempty"`
}

// ChangeRequestProvider defines the provider of a PPMS change request
type ChangeRequestProvider struct {
	Timestamp  string `json:"timeStamp"`
	Tool       string `json:"tool"`
	ReviewedBy string `json:"reviewedBy,omitempty"`
}

// ChangeRequestScan defines the underlying scan source for a PPMS change request
type ChangeRequestScan struct {
	Source string `json:"source,omitempty"`
	Tool   string `json:"tool"`
}

// ChangeRequestComparisonV1 defines the tool source for a manual PPMS change request
type ChangeRequestComparisonV1 struct {
	Tool       string `json:"tool"`
	TimeStamp  string `json:"timeStamp"`
	ReviewedBy string `json:"reviewedBy"`
}

// ChangeRequestTarget defines the target of PPMS change request
type ChangeRequestTarget struct {
	SoftwareComponentVersionNumber string `json:"softwareComponentVersionNumber"`
	SoftwareComponentVersionName   string `json:"softwareComponentVersionName,omitempty"`
	BuildVersionNumber             string `json:"buildVersionNumber,omitempty"`
	BuildVersionName               string `json:"buildVersionName,omitempty"`
}

// ChangeRequestFoss defines a Foss object within a PPMS change request
type ChangeRequestFoss struct {
	PPMSFossNumber string `json:"ppmsFossNumber"`
}

// ChangeRequestUser defines a User within a PPMS change request
type ChangeRequestUser struct {
	UserID string `json:"userId"`
	RoleID string `json:"roleId,omitempty"`
}

// ChangeRequestBuildVersion defines a build version within a PPMS change request
type ChangeRequestBuildVersion struct {
	Name                          string `json:"name"`
	Description                   string `json:"description"`
	PredecessorBuildVersionNumber string `json:"predecessorBuildVersionNumber"`
}

// ChangeRequestOptions defines options within a PPMS change request
type ChangeRequestOptions struct {
	CopyPredecessorFoss bool `json:"copyPredecessorFoss"`
	CopyPredecessorCvBv bool `json:"copyPredecessorCvBv"`
}

// ChangeRequestHeaderInformation defines the header information which can be retrieved for a change request
type ChangeRequestHeaderInformation struct {
	ID          string `json:"id"`
	CreatedAt   string `json:"createdAt"`
	CreatedBy   string `json:"createdBy"`
	Status      string `json:"status"`
	StatusText  string `json:"statusText"`
	ProcessedAt string `json:"processedAt"`
	ProcessedBy string `json:"processedBy"`
}

type ChangeRequestParams struct {
	UserID             string
	Source             string
	Tool               string
	BuildVersionNumber string
	AddFoss            []ChangeRequestFoss
	RemoveFoss         []ChangeRequestFoss
}

// GetChangeRequestV1 returns a PPMS change request object which can be sent to a PPMS entry owner in order to apply it
func (scv *SoftwareComponentVersion) GetChangeRequestV1(userID, source, tool, buildVersionNumber string, addFoss []ChangeRequestFoss, removeFoss []ChangeRequestFoss) ChangeRequestPartV1 {
	var cr ChangeRequestPartV1

	cr.SchemaVer = "1-0-0"

	cr.ChangeRequestData.ID = uuid.New().String()
	cr.ChangeRequestData.Scan.Source = source
	cr.ChangeRequestData.Scan.Tool = tool

	cr.ChangeRequestData.Comparison.Tool = "Piper"
	// reference timestamp Mon Jan 2 15:04:05 -0700 MST 2006
	cr.ChangeRequestData.Comparison.TimeStamp = time.Now().Format("20060102150405")
	cr.ChangeRequestData.Comparison.ReviewedBy = userID

	cr.ChangeRequestData.Target.SoftwareComponentVersionNumber = scv.ID
	cr.ChangeRequestData.Target.SoftwareComponentVersionName = scv.Name

	cr.ChangeRequestData.Comment = "Auto-generated by Piper - Please add to PPMS model."
	cr.ChangeRequestData.SendConfirmationTo = []ChangeRequestUser{{UserID: userID}}

	cr.ChangeRequestData.FossToAdd = addFoss
	cr.ChangeRequestData.FossToRemove = removeFoss

	if len(buildVersionNumber) > 0 {
		cr.ChangeRequestData.Target.BuildVersionNumber = buildVersionNumber
	}

	return cr
}

// WriteChangeRequestV1File creates a change request document which can be manually uploaded to PPMS
func (scv *SoftwareComponentVersion) WriteChangeRequestV1File(fileName string, params ChangeRequestParams, writeFile func(filename string, data []byte, perm os.FileMode) error) error {
	cr := scv.GetChangeRequestV1(params.UserID, params.Source, params.Tool, params.BuildVersionNumber, params.AddFoss, params.RemoveFoss)
	crJSON, err := json.Marshal(cr)
	if err != nil {
		return errors.Wrap(err, failedToCreateChangeRequest)
	}

	err = writeFile(fileName, crJSON, 0755)
	if err != nil {
		return errors.Wrap(err, "failed to write change request V1 file")
	}
	return nil
}

// GetChangeRequestV2 returns a PPMS change request object which can be automatically applied to PPMS
func (scv *SoftwareComponentVersion) GetChangeRequestV2(userID, source, tool, buildVersionNumber string, foss []ChangeRequestFoss) ChangeRequestPartV2 {
	var cr ChangeRequestPartV2

	cr.SchemaVer = "2-0-0"

	cr.ChangeRequestData.ID = uuid.New().String()
	cr.ChangeRequestData.Scan.Source = source
	cr.ChangeRequestData.Scan.Tool = tool

	cr.ChangeRequestData.Provider.Tool = "Piper"
	// Mon Jan 2 15:04:05 -0700 MST 2006
	cr.ChangeRequestData.Provider.Timestamp = time.Now().Format("20060102150405")
	cr.ChangeRequestData.Provider.ReviewedBy = userID

	cr.ChangeRequestData.Target.SoftwareComponentVersionNumber = scv.ID
	cr.ChangeRequestData.Target.SoftwareComponentVersionName = scv.Name

	cr.ChangeRequestData.Comment = "Auto-generated by Piper."
	cr.ChangeRequestData.SendConfirmationTo = []ChangeRequestUser{{UserID: userID}}

	cr.ChangeRequestData.FossComprised = foss

	if len(buildVersionNumber) > 0 {
		cr.ChangeRequestData.Target.BuildVersionNumber = buildVersionNumber
	}

	return cr
}

// SendChangeRequest sends a Change Request to the PPMS system
func (sys *System) SendChangeRequest(scv *SoftwareComponentVersion, params ChangeRequestParams, foss []ChangeRequestFoss) (string, error) {
	cr := scv.GetChangeRequestV2(params.UserID, params.Source, params.Tool, params.BuildVersionNumber, foss)
	crJSON, err := json.Marshal(cr)
	if err != nil {
		return "", errors.Wrap(err, failedToCreateChangeRequest)
	}

	return sys.SendChangeRequestDoc(bytes.NewBuffer(crJSON), "/sap/internal/ppms/api/changerequest/v1/cvpart")
}

// GetChangeRequestBV returns a PPMS change request object for creating a build version which can be automatically applied to PPMS
func (bv *BuildVersion) GetChangeRequestBV(sys *System, userID string, scv *SoftwareComponentVersion, copyPredecessorFoss, copyPredecessorCvBv bool) ChangeRequestBV {
	var cr ChangeRequestBV

	cr.SchemaVer = "1-0-0"
	cr.ChangeRequestData.ID = uuid.New().String()

	cr.ChangeRequestData.Provider.Tool = "Piper"
	cr.ChangeRequestData.Provider.Timestamp = time.Now().Format("20060102150405")
	cr.ChangeRequestData.Provider.ReviewedBy = userID

	cr.ChangeRequestData.Target.SoftwareComponentVersionNumber = bv.SoftwareComponentVersionID

	cr.ChangeRequestData.BuildVersion.Name = bv.Name
	cr.ChangeRequestData.BuildVersion.Description = bv.Description

	latestBuildVersion, err := scv.GetLatestBuildVersion(sys)
	if err == nil {
		cr.ChangeRequestData.BuildVersion.PredecessorBuildVersionNumber = latestBuildVersion.ID
	}

	cr.ChangeRequestData.Options.CopyPredecessorFoss = copyPredecessorFoss
	cr.ChangeRequestData.Options.CopyPredecessorCvBv = copyPredecessorCvBv

	return cr
}

// SendChangeRequestBV sends a build version change request to the PPMS system
func (sys *System) SendChangeRequestBV(bv *BuildVersion, userID string, scv *SoftwareComponentVersion, copyPredecessorFoss, copyPredecessorCvBv bool) (string, error) {
	cr := bv.GetChangeRequestBV(sys, userID, scv, copyPredecessorFoss, copyPredecessorCvBv)
	crJSON, err := json.Marshal(cr)
	if err != nil {
		return "", errors.Wrap(err, failedToCreateChangeRequest)
	}

	return sys.SendChangeRequestDoc(bytes.NewBuffer(crJSON), crbvAPIEndpint)
}

// SendChangeRequestDoc sends a change request document to the PPMS system
func (sys *System) SendChangeRequestDoc(cr io.Reader, endpoint string) (string, error) {

	headers := http.Header{}
	headers.Add(xCsrfHeader, "Fetch")
	response, err := sys.HTTPClient.SendRequest("HEAD", fmt.Sprintf("%v%v", sys.ServerURL, endpoint), nil, headers, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve PPMS CSRF token")
	}
	response.Body.Close()
	token := response.Header.Get(xCsrfHeader)
	cookies := response.Cookies()

	headers.Set(xCsrfHeader, token)

	response, err = sys.HTTPClient.SendRequest("POST", fmt.Sprintf("%v%v", sys.ServerURL, endpoint), cr, headers, cookies)
	if err != nil {
		return "", errors.Wrap(err, "failed to send change request")
	}
	content, err := io.ReadAll(response.Body)
	if err != nil {
		return "", errors.Wrap(err, "error reading response")
	}
	response.Body.Close()

	var responseDoc map[string]string
	err = json.Unmarshal(content, &responseDoc)
	if err != nil {
		return "", errors.Wrap(err, "error unmarshalling change request response")
	}

	return responseDoc["crId"], nil
}

// GetChangeRequestHeaderInfo returns change request header which contains status information about the change request
func (sys *System) GetChangeRequestHeaderInfo(crType, crID string) (ChangeRequestHeaderInformation, error) {
	var headerInfo ChangeRequestHeaderInformation
	content, err := sys.sendRequest("GET", fmt.Sprintf("/sap/internal/ppms/api/changerequest/v1/%v/%v", crType, crID), nil, nil)
	if err != nil {
		return headerInfo, errors.Wrap(err, "failed to retrieve change request header info")
	}
	err = json.Unmarshal(content, &headerInfo)
	if err != nil {
		return headerInfo, errors.Wrap(err, "error unmarshalling change request header info")
	}

	return headerInfo, nil
}

// WaitForInitialChangeRequestFeedback waits for initial Feedback of a change request
func (sys *System) WaitForInitialChangeRequestFeedback(crID string, duration time.Duration) error {
	logger := log.Entry().WithField("package", "ContinuousDelivery/piper-library/pkg/ppms")
	for i := 0; i < 12; i++ {
		time.Sleep(5 * duration)
		crInfo, err := sys.GetChangeRequestHeaderInfo("cvpart", crID)
		if err != nil {
			logger.WithError(err).Error("failed to retrieve status of PPMS change request")
		}
		logger.Debugf("Status of PPMS change request: %v (%v)", crInfo.Status, crInfo.StatusText)
		if crInfo.Status != "PENDING" && crInfo.Status != "APPLIED" {
			return fmt.Errorf("PPMS upload failed. Status of change request %v is '%v'", crID, crInfo.Status)
		}
		if crInfo.Status == "APPLIED" {
			break
		}
	}
	return nil
}

// WaitForBuildVersionChangeRequestApplied waits for a build version change request to be applied
func (sys *System) WaitForBuildVersionChangeRequestApplied(crID string, duration time.Duration) error {
	logger := log.Entry().WithField("package", "ContinuousDelivery/piper-library/pkg/ppms")
	crApplied := false
	for i := 0; i < 120; i++ {
		time.Sleep(5 * duration)
		crInfo, err := sys.GetChangeRequestHeaderInfo("cvbv", crID)
		if err != nil {
			logger.WithError(err).Error("failed to retrieve status of PPMS change request")
		}
		logger.Debugf("Status of PPMS change request: %v (%v)", crInfo.Status, crInfo.StatusText)
		if crInfo.Status != "PENDING" && crInfo.Status != "APPLIED" {
			return fmt.Errorf("PPMS build version creation failed. Status of change request %v is '%v'", crID, crInfo.Status)
		}
		if crInfo.Status == "APPLIED" {
			crApplied = true
			break
		}
	}
	if !crApplied {
		return fmt.Errorf("PPMS build version creation timed out. Status of change request %v is still 'PENDING'", crID)
	}
	return nil
}

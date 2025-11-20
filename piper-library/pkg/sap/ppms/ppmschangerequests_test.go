//go:build unit
// +build unit

package ppms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

type ppmsCRMockClient struct {
	httpMethod  []string
	urlsCalled  []string
	requestBody []string
}

func (c *ppmsCRMockClient) SetOptions(opts piperhttp.ClientOptions) {}

var crStatusApplied bool = false

func (c *ppmsCRMockClient) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	c.httpMethod = append(c.httpMethod, method)
	c.urlsCalled = append(c.urlsCalled, url)
	if body != nil {
		bodyContent, _ := io.ReadAll(body)
		c.requestBody = append(c.requestBody, string(bodyContent))
	} else {
		c.requestBody = append(c.requestBody, "")
	}
	if method == "HEAD" {
		csrf := header.Get("x-csrf-token")
		respHeader := http.Header{}
		if csrf == "Fetch" {
			respHeader.Add("x-csrf-token", "testToken")
		}
		return &http.Response{StatusCode: 200, Header: respHeader, Body: io.NopCloser(bytes.NewReader([]byte("")))}, nil
	}

	if method == "POST" && header.Get("x-csrf-token") == "testToken" {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(`{"crId":"1000"}`)))}, nil
	}

	if method == "GET" && url == "https://my.test.server/sap/internal/ppms/api/changerequest/v1/cvpart/1000" {
		respContent := `{"id": "1000","createdBy": "testUser","status": "PENDING","statusText": "Not yet applied","processedAt": null,"processedBy": null}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(respContent)))}, nil
	}

	if method == "GET" && url == "https://my.test.server/sap/internal/ppms/api/changerequest/v1/cvpart/1001" {
		respContent := ""
		if crStatusApplied {
			respContent = `{"id": "1001","createdBy": "testUser","status": "APPLIED","statusText": "Applied","processedAt": null,"processedBy": null}`
		} else {
			respContent = `{"id": "1001","createdBy": "testUser","status": "PENDING","statusText": "Not yet applied","processedAt": null,"processedBy": null}`
			crStatusApplied = true
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(respContent)))}, nil
	}

	if method == "GET" && url == "https://my.test.server/sap/internal/ppms/api/changerequest/v1/cvpart/1002" {
		respContent := `{"id": "1002","createdBy": "testUser","status": "ERROR","statusText": "Error occurred","processedAt": null,"processedBy": null}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(respContent)))}, nil
	}

	if method == "GET" && url == "https://my.test.server/sap/internal/ppms/api/changerequest/v1/cvbv/1003" {
		respContent := ""
		if crStatusApplied {
			respContent = `{"id": "1003","createdBy": "testUser","status": "APPLIED","statusText": "Applied","processedAt": null,"processedBy": null}`
		} else {
			respContent = `{"id": "1003","createdBy": "testUser","status": "PENDING","statusText": "Not yet applied","processedAt": null,"processedBy": null}`
			crStatusApplied = true
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(respContent)))}, nil
	}

	if method == "GET" && url == "https://my.test.server/sap/internal/ppms/api/changerequest/v1/cvbv/1004" {
		respContent := `{"id": "1004","createdBy": "testUser","status": "ERROR","statusText": "Error occurred","processedAt": null,"processedBy": null}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(respContent)))}, nil
	}

	if method == "GET" && url == "https://my.test.server/sap/internal/ppms/api/changerequest/v1/cvbv/1005" {
		respContent := `{"id": "1005","createdBy": "testUser","status": "PENDING","statusText": "Not yet applied","processedAt": null,"processedBy": null}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte(respContent)))}, nil
	}

	return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte("")))}, nil
}

var files map[string][]byte

func writeFileMock(filename string, data []byte, perm os.FileMode) error {
	if files == nil {
		files = make(map[string][]byte)
	}
	files[filename] = data
	return nil
}

func TestGetChangeRequestV1(t *testing.T) {

	t.Run("With BV", func(t *testing.T) {
		scv := SoftwareComponentVersion{ID: "1", Name: "TestSCV"}
		addFoss := []ChangeRequestFoss{{PPMSFossNumber: "fossId1"}, {PPMSFossNumber: "fossId2"}}
		removeFoss := []ChangeRequestFoss{}
		cr := scv.GetChangeRequestV1("testUser", "testSource", "testTool", "bvNumber", addFoss, removeFoss)

		assert.Equal(t, "1-0-0", cr.SchemaVer)
		assert.True(t, len(cr.ChangeRequestData.ID) > 0, "no change request id available")
		assert.Equal(t, "testSource", cr.ChangeRequestData.Scan.Source)
		assert.Equal(t, "testTool", cr.ChangeRequestData.Scan.Tool)
		assert.Equal(t, "Piper", cr.ChangeRequestData.Comparison.Tool)
		assert.True(t, len(cr.ChangeRequestData.Comparison.TimeStamp) > 0, "no change reqtimestamp available")
		assert.Equal(t, "testUser", cr.ChangeRequestData.Comparison.ReviewedBy)
		assert.Equal(t, scv.ID, cr.ChangeRequestData.Target.SoftwareComponentVersionNumber)
		assert.Equal(t, scv.Name, cr.ChangeRequestData.Target.SoftwareComponentVersionName)
		assert.Equal(t, []ChangeRequestUser{{UserID: "testUser"}}, cr.ChangeRequestData.SendConfirmationTo)
		assert.Equal(t, "bvNumber", cr.ChangeRequestData.Target.BuildVersionNumber)
		assert.Equal(t, []ChangeRequestFoss{{PPMSFossNumber: "fossId1"}, {PPMSFossNumber: "fossId2"}}, cr.ChangeRequestData.FossToAdd)
	})

	t.Run("Without BV", func(t *testing.T) {
		scv := SoftwareComponentVersion{ID: "1", Name: "TestSCV"}
		addFoss := []ChangeRequestFoss{}
		removeFoss := []ChangeRequestFoss{}
		cr := scv.GetChangeRequestV1("testUser", "testSource", "testTool", "", addFoss, removeFoss)

		assert.Equal(t, "", cr.ChangeRequestData.Target.BuildVersionNumber)
	})
}

func TestWriteChangeRequestV1File(t *testing.T) {
	scv := SoftwareComponentVersion{ID: "1", Name: "TestSCV"}
	addFoss := []ChangeRequestFoss{}
	removeFoss := []ChangeRequestFoss{}
	chageRequestParams := ChangeRequestParams{
		UserID:             "testUser",
		Source:             "testSource",
		Tool:               "testTool",
		BuildVersionNumber: "bvNumber",
		AddFoss:            addFoss,
		RemoveFoss:         removeFoss,
	}
	err := scv.WriteChangeRequestV1File("testFile.json", chageRequestParams, writeFileMock)

	assert.NoError(t, err)
	assert.Greater(t, len(files["testFile.json"]), 0)
}

func TestGetChangeRequestV2(t *testing.T) {

	t.Run("With BV", func(t *testing.T) {
		scv := SoftwareComponentVersion{ID: "1", Name: "TestSCV"}
		foss := []ChangeRequestFoss{{PPMSFossNumber: "fossId1"}, {PPMSFossNumber: "fossId2"}}
		cr := scv.GetChangeRequestV2("testUser", "testSource", "testTool", "bvNumber", foss)

		assert.Equal(t, "2-0-0", cr.SchemaVer)
		assert.True(t, len(cr.ChangeRequestData.ID) > 0, "no change request id available")
		assert.Equal(t, "testSource", cr.ChangeRequestData.Scan.Source)
		assert.Equal(t, "testTool", cr.ChangeRequestData.Scan.Tool)
		assert.Equal(t, "Piper", cr.ChangeRequestData.Provider.Tool)
		assert.True(t, len(cr.ChangeRequestData.Provider.Timestamp) > 0, "no timestamp available")
		assert.Equal(t, "testUser", cr.ChangeRequestData.Provider.ReviewedBy)
		assert.Equal(t, scv.ID, cr.ChangeRequestData.Target.SoftwareComponentVersionNumber)
		assert.Equal(t, scv.Name, cr.ChangeRequestData.Target.SoftwareComponentVersionName)
		assert.Equal(t, []ChangeRequestUser{{UserID: "testUser"}}, cr.ChangeRequestData.SendConfirmationTo)
		assert.Equal(t, "bvNumber", cr.ChangeRequestData.Target.BuildVersionNumber)
		assert.Equal(t, []ChangeRequestFoss{{PPMSFossNumber: "fossId1"}, {PPMSFossNumber: "fossId2"}}, cr.ChangeRequestData.FossComprised)
	})

	t.Run("Without BV", func(t *testing.T) {
		scv := SoftwareComponentVersion{ID: "1", Name: "TestSCV"}
		foss := []ChangeRequestFoss{}
		cr := scv.GetChangeRequestV2("testUser", "testSource", "testTool", "", foss)

		assert.Equal(t, "", cr.ChangeRequestData.Target.BuildVersionNumber)
	})
}

func TestSendChangeRequest(t *testing.T) {
	scv := SoftwareComponentVersion{ID: "1", Name: "TestSCV"}
	foss := []ChangeRequestFoss{}
	myTestClient := ppmsCRMockClient{}
	sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}
	addFoss := []ChangeRequestFoss{}
	removeFoss := []ChangeRequestFoss{}
	chageRequestParams := ChangeRequestParams{
		UserID:             "testUser",
		Source:             "testSource",
		Tool:               "testTool",
		BuildVersionNumber: "bvNumber",
		AddFoss:            addFoss,
		RemoveFoss:         removeFoss,
	}

	crID, err := sys.SendChangeRequest(&scv, chageRequestParams, foss)

	assert.NoError(t, err, "Error occurred but none expected")

	assert.Equal(t, []string{"HEAD", "POST"}, myTestClient.httpMethod)
	assert.Equal(t, []string{"https://my.test.server/sap/internal/ppms/api/changerequest/v1/cvpart", "https://my.test.server/sap/internal/ppms/api/changerequest/v1/cvpart"}, myTestClient.urlsCalled)

	assert.Contains(t, myTestClient.requestBody[1], `"fossComprised":[]`)
	assert.Equal(t, "1000", crID)
}

func TestSendChangeRequestBV(t *testing.T) {
	scv := SoftwareComponentVersion{ID: "1", Name: "TestSCV"}
	bv := BuildVersion{ID: "11", Name: "TestBV"}
	myTestClient := ppmsCRMockClient{}
	sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}

	crID, err := sys.SendChangeRequestBV(&bv, "testUser", &scv, false, false)

	assert.NoError(t, err, "Error occurred but none expected")
	assert.Equal(t, "1000", crID)
}

func TestSendChangeRequestDoc(t *testing.T) {
	myTestClient := ppmsCRMockClient{}
	sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}

	cr := ChangeRequestPartV2{}

	doc, _ := json.Marshal(cr)

	crID, err := sys.SendChangeRequestDoc(bytes.NewReader(doc), "/sap/internal/ppms/api/changerequest/v1/cvpart")

	assert.NoError(t, err, "Error occurred but none expected")

	assert.Equal(t, []string{"HEAD", "POST"}, myTestClient.httpMethod)
	assert.Equal(t, []string{"https://my.test.server/sap/internal/ppms/api/changerequest/v1/cvpart", "https://my.test.server/sap/internal/ppms/api/changerequest/v1/cvpart"}, myTestClient.urlsCalled)

	assert.Equal(t, "1000", crID)
}

func TestGetChangeRequestHeaderInfo(t *testing.T) {
	myTestClient := ppmsCRMockClient{}
	sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}

	crInfo, err := sys.GetChangeRequestHeaderInfo("cvpart", "1000")

	assert.NoError(t, err, "Error occurred but none expected")
	assert.Equal(t, "https://my.test.server/sap/internal/ppms/api/changerequest/v1/cvpart/1000", myTestClient.urlsCalled[0])
	assert.Equal(t, "1000", crInfo.ID)
	assert.Equal(t, "testUser", crInfo.CreatedBy)
	assert.Equal(t, "PENDING", crInfo.Status)
	assert.Equal(t, "Not yet applied", crInfo.StatusText)
	assert.Empty(t, crInfo.ProcessedAt)
	assert.Empty(t, crInfo.ProcessedBy)
}

func TestWaitForInitialChangeRequestFeedback(t *testing.T) {

	myTestClient := ppmsCRMockClient{}

	t.Run("Success case", func(t *testing.T) {
		sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}
		err := sys.WaitForInitialChangeRequestFeedback("1001", time.Microsecond)

		assert.NoError(t, err)
	})

	t.Run("Error case", func(t *testing.T) {
		sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}
		err := sys.WaitForInitialChangeRequestFeedback("1002", time.Microsecond)

		assert.Contains(t, fmt.Sprint(err), "PPMS upload failed. Status of change request 1002 is 'ERROR'")
	})
}

func TestWaitForBuildVersionChangeRequestApplied(t *testing.T) {

	myTestClient := ppmsCRMockClient{}

	t.Run("Success case", func(t *testing.T) {
		sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}
		err := sys.WaitForBuildVersionChangeRequestApplied("1003", time.Microsecond)

		assert.NoError(t, err)
	})

	t.Run("Error case", func(t *testing.T) {
		sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}
		err := sys.WaitForBuildVersionChangeRequestApplied("1004", time.Microsecond)

		assert.Contains(t, fmt.Sprint(err), "PPMS build version creation failed. Status of change request 1004 is 'ERROR'")
	})

	t.Run("Error case timeout", func(t *testing.T) {
		sys := System{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}
		err := sys.WaitForBuildVersionChangeRequestApplied("1005", time.Microsecond)

		assert.Contains(t, fmt.Sprint(err), "PPMS build version creation timed out. Status of change request 1005 is still 'PENDING'")
	})
}

//go:build unit
// +build unit

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/reporting"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/whitesource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/blackduck"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/gtlc"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/ppms"
)

type ppmsUtilsMock struct {
	*mock.FilesMock
}

func (p *ppmsUtilsMock) Now() time.Time {
	now, _ := time.Parse("2006-01-02 15:04:05", "2010-05-10 00:15:42")
	return now
}

func newPPMSUtilsMock() *ppmsUtilsMock {
	return &ppmsUtilsMock{
		FilesMock: &mock.FilesMock{},
	}
}

type changeRequestSenderMock struct {
	buildVersionNumber string
	crID               string
	returnCrID         string
	foss               []ppms.ChangeRequestFoss
	scv                *ppms.SoftwareComponentVersion
	source             string
	tool               string
	userID             string
	sendError          error
	waitError          error
}

func (c *changeRequestSenderMock) SendChangeRequest(scv *ppms.SoftwareComponentVersion, params ppms.ChangeRequestParams, foss []ppms.ChangeRequestFoss) (string, error) {
	c.foss = foss
	c.scv = scv
	c.userID = params.UserID
	c.source = params.Source
	c.tool = params.Tool
	c.buildVersionNumber = params.BuildVersionNumber
	return c.returnCrID, c.sendError
}

func (c *changeRequestSenderMock) WaitForInitialChangeRequestFeedback(crID string, duration time.Duration) error {
	c.crID = crID
	return c.waitError
}

type httpErrorClient struct{}

func (c *httpErrorClient) SetOptions(opts piperhttp.ClientOptions) {}
func (c *httpErrorClient) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	return &http.Response{}, fmt.Errorf("Send error")
}

type httpMockClient struct {
	responseBodyForURL map[string]string
	errorMessageForURL map[string]string
}

func (c *httpMockClient) SetOptions(opts piperhttp.ClientOptions) {}
func (c *httpMockClient) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	response := http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(""))),
	}

	if c.errorMessageForURL[url] != "" {
		response.StatusCode = 400
		return &response, errors.New(c.errorMessageForURL[url])
	}

	if c.responseBodyForURL[url] != "" {
		response.Body = io.NopCloser(bytes.NewReader([]byte(c.responseBodyForURL[url])))
		return &response, nil
	}

	return &response, nil
}

type whitesourceMock struct {
	productName              string
	product                  whitesource.Product
	projectTokens            []string
	projects                 []whitesource.Project
	getProductByNameError    error
	getProductError          error
	getProjectTokensError    error
	getProjectsMetaInfoError error
}

func (w *whitesourceMock) GetProductName(productToken string) (string, error) {
	if w.getProductError != nil {
		return "", w.getProductError
	}
	return w.productName, nil
}

func (w *whitesourceMock) GetProjectTokens(productToken string, projectNames []string) ([]string, error) {
	if w.getProjectTokensError != nil {
		return []string{}, w.getProjectTokensError
	}
	return w.projectTokens, nil
}

func (w *whitesourceMock) GetProductByName(productName string) (whitesource.Product, error) {
	if w.getProductByNameError != nil {
		return whitesource.Product{}, w.getProductByNameError
	}
	return w.product, nil
}

func (w *whitesourceMock) GetProjectsMetaInfo(productToken string) ([]whitesource.Project, error) {
	if w.getProjectsMetaInfoError != nil {
		return []whitesource.Project{}, w.getProjectsMetaInfoError
	}
	return w.projects, nil
}

func TestRunPPMSComplianceCheck(t *testing.T) {

	t.Run("Simple run through", func(t *testing.T) {
		myTestClient := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.whitesource.system": `{
					"projectVitals":[
						{
							"pluginName":"test-plugin",
							"name": "Test Project",
							"token": "test_project_token",
							"uploadedBy": "test_upload_user",
							"creationDate": "2020-01-01 00:00:00",
							"lastUpdatedDate": "2020-01-01 01:00:00"
						},
						{
							"pluginName": "test-plugin",
							"name": "Component B - master",
							"token": "component_b_master_token",
							"uploadedBy": "test_upload_user",
							"creationDate": "2021-01-01 00:00:00",
							"lastUpdatedDate": "2021-01-01 01:00:00"
						}
					],
					"productTags":[
						{
							"name": "WhiteSourceTestProduct"
						}
					]
				}`,
				"https://my.blackduck.system/api/tokens/authenticate": `{
					"bearerToken":"bearerTestToken",
					"expiresInMilliseconds":7199997
				}`,
				"https://my.blackduck.system/api/projects?q=name%3ASHC-PiperTest": `{
					"totalCount": 1,
					"items": [
						{
							"name": "SHC-PiperTest",
							"_meta": {
								"href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf",
								"links": [
									{
										"rel": "versions",
										"href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions"
									}
								]
							}
						}
					]
				}`,
				"https://my.blackduck.system/api/projects?q=name%3ASHC-PiperCustom": `{
					"totalCount": 1,
					"items": [
						{
							"name": "SHC-PiperCustom",
							"_meta": {
								"href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbg",
								"links": [
									{
										"rel": "versions",
										"href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbg/versions"
									}
								]
							}
						}
					]
				}`,
				"https://my.blackduck.system/api/projects?q=name%3ASHC-PiperCustom1": `{
					"totalCount": 1,
					"items": [
						{
							"name": "SHC-PiperCustom1",
							"_meta": {
								"href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbe",
								"links": [
									{
										"rel": "versions",
										"href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbg/versions"
									}
								]
							}
						}
					]
				}`,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions?limit=100&offset=0": `{
					"totalCount": 1,
					"items": [
						{
							"versionName": "1.0",
							"_meta": {
								"href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36",
								"links": []
							}
						}
					]
				}`,
				"https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbg/versions?limit=100&offset=0": `{
					"totalCount": 1,
					"items": [
						{
							"versionName": "customVersion",
							"_meta": {
								"href": "https://my.blackduck.system/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36",
								"links": []
							}
						}
					]
				}`,
				"https://my.ppms.system/odataint/borm/odataforosrcy/SoftwareComponentVersions('1001')?$format=json":                         `{"d":{"Id":"1001"}}`,
				"https://my.ppms.system/odataint/borm/odataforosrcy/SoftwareComponentVersions('1001')/BuildVersions?$format=json":           `{"d":{"results":[]}}`,
				"https://my.ppms.system/odataint/borm/odataforosrcy/SoftwareComponentVersions('1001')/FreeOpenSourceSoftwares?$format=json": `{"d":{"results":[]}}`,
			},
		}

		telemetryCustomData := telemetry.CustomData{}

		mappingSystem := gtlc.MappingSystem{
			HTTPClient: &myTestClient,
			ServerURL:  "https://my.gtlc.system",
		}
		ppmsSystem := ppms.System{
			HTTPClient: &myTestClient,
			ServerURL:  "https://my.ppms.system",
		}

		whitesourceSystem := whitesourceMock{
			productName:   "WhiteSourceTestProduct",
			projectTokens: []string{"test_project_token"},
		}

		blackDuckClient := blackduck.NewClient("bdTestToken", "https://my.blackduck.system", &myTestClient)

		dir := t.TempDir()

		t.Run("master mode - WhiteSource", func(t *testing.T) {
			config := sapCheckPPMSComplianceOptions{
				OrgToken:                "OrgToken",
				PpmsID:                  "1001",
				ServerURL:               "https://my.server.url",
				WhitesourceProductToken: "productToken",
				UserToken:               "userToken",
				WhitesourceProjectNames: []string{"Test Project"},
			}
			utilsMock := newPPMSUtilsMock()
			err := runPPMSComplianceCheck(&config, &telemetryCustomData, &mappingSystem, &ppmsSystem, &whitesourceSystem, &blackDuckClient, utilsMock, dir, time.Microsecond)
			assert.NoError(t, err)
			fileContent, err := utilsMock.FileRead("piper_whitesource_ppms_report.html")
			assert.NoError(t, err, "file not existing")
			assert.Contains(t, string(fileContent), "WhiteSource product name: WhiteSourceTestProduct")
			assert.Contains(t, string(fileContent), "Filtered project names: Test Project")

			fileContent, err = utilsMock.FileRead(filepath.Join(reporting.StepReportDirectory, "piper_whitesource_ppms_report_20100510001542.json"))
			assert.NoError(t, err, "file not existing")
			assert.Contains(t, string(fileContent), `"successfulScan":true`)

			reportFile, err := utilsMock.FileRead(filepath.Join(dir, "sapCheckPPMSCompliance_reports.json"))
			assert.NoError(t, err)
			assert.Contains(t, string(reportFile), fmt.Sprintf(`"target":"%v"`, "piper_whitesource_ppms_report.html"))
		})

		t.Run("master mode - WhiteSource (with project name pattern)", func(t *testing.T) {
			wsSystem := whitesourceMock{
				productName:   "WhiteSourceTestProduct",
				projectTokens: []string{"component_b_master_token"},
				projects: []whitesource.Project{
					{Name: "WhiteSourceTestProduct"},
					{Name: "Component B - master"},
				},
			}

			config := sapCheckPPMSComplianceOptions{
				OrgToken:                "OrgToken",
				PpmsID:                  "1001",
				ReportFileName:          "testReport",
				ServerURL:               "https://my.server.url",
				WhitesourceProductToken: "productToken",
				UserToken:               "userToken",
				//WhitesourceProjectNames: []string{"Component B - master"},
				WhitesourceProjectNamesPattern: "^.*- master$",
			}
			utilsMock := newPPMSUtilsMock()
			err := runPPMSComplianceCheck(&config, &telemetryCustomData, &mappingSystem, &ppmsSystem, &wsSystem, &blackDuckClient, utilsMock, dir, time.Microsecond)
			assert.NoError(t, err)
			fileContent, err := utilsMock.FileRead("piper_whitesource_ppms_report.html")
			assert.NoError(t, err, "file not existing")
			assert.Contains(t, string(fileContent), "WhiteSource product name: WhiteSourceTestProduct")
			assert.Contains(t, string(fileContent), "Filtered project names: Component B - master")

			fileContent, err = utilsMock.FileRead(filepath.Join(reporting.StepReportDirectory, "piper_whitesource_ppms_report_20100510001542.json"))
			assert.NoError(t, err, "file not existing")
			assert.Contains(t, string(fileContent), `"successfulScan":true`)

			reportFile, err := utilsMock.FileRead(filepath.Join(dir, "sapCheckPPMSCompliance_reports.json"))
			assert.NoError(t, err)
			assert.Contains(t, string(reportFile), fmt.Sprintf(`"target":"%v"`, "piper_whitesource_ppms_report.html"))
		})

		t.Run("master mode - BlackDuck", func(t *testing.T) {
			config := sapCheckPPMSComplianceOptions{
				DetectToken:          "bdTestToken",
				BlackduckProjectName: "SHC-PiperTest",
				Version:              "1.0.0",
				VersioningModel:      "major-minor",
				PpmsID:               "1001",
				ServerURL:            "https://my.server.url",
			}
			utilsMock := newPPMSUtilsMock()
			err := runPPMSComplianceCheck(&config, &telemetryCustomData, &mappingSystem, &ppmsSystem, &whitesourceSystem, &blackDuckClient, utilsMock, dir, time.Microsecond)
			assert.NoError(t, err)
			fileContent, err := utilsMock.FileRead("piper_blackduck_ppms_report.html")
			assert.NoError(t, err, "file not existing")
			assert.Contains(t, string(fileContent), "BlackDuck project name: SHC-PiperTest")
			assert.Contains(t, string(fileContent), "BlackDuck project version: 1.0")
		})

		t.Run("master mode - BlackDuck with custom version", func(t *testing.T) {
			config := sapCheckPPMSComplianceOptions{
				DetectToken:          "bdTestToken",
				BlackduckProjectName: "SHC-PiperCustom",
				CustomScanVersion:    "customVersion",
				Version:              "1.0.0",
				VersioningModel:      "major-minor",
				PpmsID:               "1001",
				ServerURL:            "https://my.server.url",
			}
			utilsMock := newPPMSUtilsMock()
			err := runPPMSComplianceCheck(&config, &telemetryCustomData, &mappingSystem, &ppmsSystem, &whitesourceSystem, &blackDuckClient, utilsMock, dir, time.Microsecond)
			assert.NoError(t, err)
			fileContent, err := utilsMock.FileRead("piper_blackduck_ppms_report.html")
			assert.NoError(t, err, "file not existing")
			assert.Contains(t, string(fileContent), "BlackDuck project name: SHC-PiperCustom")
			assert.Contains(t, string(fileContent), "BlackDuck project version: customVersion")
		})

		t.Run("BlackDuck multiple project names", func(t *testing.T) {
			config := sapCheckPPMSComplianceOptions{
				DetectToken:           "bdTestToken",
				BlackduckProjectNames: []string{"SHC-PiperCustom", "SHC-PiperCustom1"},
				Version:               "1.0.0",
				CustomScanVersion:     "customVersion",
				VersioningModel:       "major-minor",
				PpmsID:                "1001",
				ServerURL:             "https://my.server.url",
			}
			versions, err := getBlackDuckScanVersions(&config, &blackDuckClient, "customVersion")
			assert.NoError(t, err)
			fmt.Println(err)
			fmt.Println(versions)
			assert.Contains(t, versions[0], "5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36")
			assert.Contains(t, versions[1], "5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36")
		})

		t.Run("master mode - project token failure", func(t *testing.T) {
			wsSystem := whitesourceMock{
				productName:           "WhiteSourceTestProduct",
				getProjectTokensError: fmt.Errorf("token error"),
			}
			config := sapCheckPPMSComplianceOptions{
				OrgToken:                "OrgToken",
				PpmsID:                  "1001",
				WhitesourceProductToken: "productToken",
				UserToken:               "userToken",
				WhitesourceProjectNames: []string{"Test Project"},
			}
			utilsMock := newPPMSUtilsMock()
			err := runPPMSComplianceCheck(&config, &telemetryCustomData, &mappingSystem, &ppmsSystem, &wsSystem, &blackDuckClient, utilsMock, dir, time.Microsecond)
			assert.Contains(t, fmt.Sprint(err), "failed to get project token(s) for [Test Project]")
		})

		t.Run("master mode - failed matches", func(t *testing.T) {
			config := sapCheckPPMSComplianceOptions{
				OrgToken:                "OrgToken",
				PpmsID:                  "1001",
				ServerURL:               "https://my.server.url",
				WhitesourceProductToken: "productToken",
				UserToken:               "userToken",
				WhitesourceProjectNames: []string{"Test Project"},
			}
			mappingSys := gtlc.MappingSystem{
				HTTPClient: &httpMockClient{
					responseBodyForURL: map[string]string{
						"https://my.gtlc.system/api/mapping/whitesource/projects/test_project_token/bom?expand=SAP_IP": `[{"standardFoss":{"complianceInfo":{"ipCompliance":{"fossId": "9001"}}}}]`,
					},
				},
				ServerURL: "https://my.gtlc.system",
			}
			utilsMock := newPPMSUtilsMock()
			err := runPPMSComplianceCheck(&config, &telemetryCustomData, &mappingSys, &ppmsSystem, &whitesourceSystem, &blackDuckClient, utilsMock, dir, time.Microsecond)
			assert.Contains(t, fmt.Sprint(err), "1 PPMS entries are missing for this build")
			fileContent, err := utilsMock.FileRead("piper_whitesource_ppms_report.html")
			assert.NoError(t, err, "file not existing")
			assert.Contains(t, string(fileContent), "Total number of libraries: 1")
			assert.Contains(t, string(fileContent), `Total number of successful library matches: 0`)

			fileContent, err = utilsMock.FileRead(filepath.Join(reporting.StepReportDirectory, "piper_whitesource_ppms_report_20100510001542.json"))
			assert.NoError(t, err, "file not existing")
			assert.Contains(t, string(fileContent), `"successfulScan":false`)
		})

		t.Run("master mode - missing whitesourceUserToken", func(t *testing.T) {
			config := sapCheckPPMSComplianceOptions{
				OrgToken:                "OrgToken",
				PpmsID:                  "1001",
				ServerURL:               "https://my.server.url",
				WhitesourceProjectNames: []string{"Test Project"},
				WhitesourceProductToken: "productToken",
			}
			utilsMock := newPPMSUtilsMock()
			err := runPPMSComplianceCheck(&config, &telemetryCustomData, &mappingSystem, &ppmsSystem, &whitesourceSystem, &blackDuckClient, utilsMock, dir, time.Microsecond)
			assert.EqualError(t, err, "missing mandatory configuration for dependency information: detectToken (Detect) or userToken (WhiteSource) needs to be set")
		})

		t.Run("master mode - missing whitesourceProjectNames", func(t *testing.T) {
			config := sapCheckPPMSComplianceOptions{
				OrgToken:                "OrgToken",
				PpmsID:                  "1001",
				ServerURL:               "https://my.server.url",
				WhitesourceProductToken: "productToken",
				UserToken:               "userToken",
			}
			utilsMock := newPPMSUtilsMock()
			err := runPPMSComplianceCheck(&config, &telemetryCustomData, &mappingSystem, &ppmsSystem, &whitesourceSystem, &blackDuckClient, utilsMock, dir, time.Microsecond)
			assert.EqualError(t, err, "please configure the whitesource projects which shall be considered (either with whitesourceProjectNames or whitesourceProjectNamesPattern)")
		})

		t.Run("master mode - missing whitesourceProjectNames", func(t *testing.T) {
			config := sapCheckPPMSComplianceOptions{
				OrgToken:                "OrgToken",
				PpmsID:                  "1001",
				ServerURL:               "https://my.server.url",
				WhitesourceProductToken: "productToken",
				UserToken:               "userToken",
			}
			utilsMock := newPPMSUtilsMock()
			err := runPPMSComplianceCheck(&config, &telemetryCustomData, &mappingSystem, &ppmsSystem, &whitesourceSystem, &blackDuckClient, utilsMock, dir, time.Microsecond)
			assert.EqualError(t, err, "please configure the whitesource projects which shall be considered (either with whitesourceProjectNames or whitesourceProjectNamesPattern)")
		})

		t.Run("master mode - missing orgToken", func(t *testing.T) {
			config := sapCheckPPMSComplianceOptions{
				PpmsID:                  "1001",
				ReportFileName:          "testReport",
				ServerURL:               "https://my.server.url",
				WhitesourceProjectNames: []string{"Test Project"},
				WhitesourceProductToken: "productToken",
				UserToken:               "userToken",
			}
			utilsMock := newPPMSUtilsMock()
			err := runPPMSComplianceCheck(&config, &telemetryCustomData, &mappingSystem, &ppmsSystem, &whitesourceSystem, &blackDuckClient, utilsMock, dir, time.Microsecond)
			assert.EqualError(t, err, "missing mandatory configuration for WhiteSource: orgToken, whitesourceProductToken/whitesourceProductName need to be set")
		})

		t.Run("master mode - missing whitesourceProductToken", func(t *testing.T) {
			config := sapCheckPPMSComplianceOptions{
				OrgToken:                "OrgToken",
				PpmsID:                  "1001",
				ReportFileName:          "testReport",
				ServerURL:               "https://my.server.url",
				WhitesourceProjectNames: []string{"Test Project"},
				UserToken:               "userToken",
			}
			utilsMock := newPPMSUtilsMock()
			err := runPPMSComplianceCheck(&config, &telemetryCustomData, &mappingSystem, &ppmsSystem, &whitesourceSystem, &blackDuckClient, utilsMock, dir, time.Microsecond)
			assert.EqualError(t, err, "missing mandatory configuration for WhiteSource: orgToken, whitesourceProductToken/whitesourceProductName need to be set")
		})

		t.Run("master mode - invalid pattern", func(t *testing.T) {
			config := sapCheckPPMSComplianceOptions{
				OrgToken:                       "OrgToken",
				PpmsID:                         "1001",
				ReportFileName:                 "testReport",
				ServerURL:                      "https://my.server.url",
				WhitesourceProjectNamesPattern: "\\K",
				WhitesourceProductToken:        "productToken",
				UserToken:                      "userToken",
			}
			utilsMock := newPPMSUtilsMock()
			err := runPPMSComplianceCheck(&config, &telemetryCustomData, &mappingSystem, &ppmsSystem, &whitesourceSystem, &blackDuckClient, utilsMock, dir, time.Microsecond)
			assert.EqualError(t, err, "unable to compile the provided pattern for project names '\\K'")
		})

		t.Run("master mode - projects resolution error", func(t *testing.T) {
			wsSystem := whitesourceMock{
				productName:              "WhiteSourceTestProduct",
				getProjectsMetaInfoError: fmt.Errorf("token error"),
			}

			config := sapCheckPPMSComplianceOptions{
				OrgToken:                       "OrgToken",
				PpmsID:                         "1001",
				ReportFileName:                 "testReport",
				ServerURL:                      "https://my.server.url",
				WhitesourceProjectNamesPattern: ".*",
				WhitesourceProductToken:        "productToken",
				UserToken:                      "userToken",
			}
			utilsMock := newPPMSUtilsMock()
			err := runPPMSComplianceCheck(&config, &telemetryCustomData, &mappingSystem, &ppmsSystem, &wsSystem, &blackDuckClient, utilsMock, dir, time.Microsecond)
			assert.EqualError(t, err, "failed to resolve projects for product 'productToken'")
		})

		t.Run("PR mode", func(t *testing.T) {
			config := sapCheckPPMSComplianceOptions{
				OrgToken:                "OrgToken",
				CreateBuildVersion:      true,
				PpmsID:                  "1001",
				PullRequestMode:         true,
				ServerURL:               "https://my.server.url",
				UploadChangeRequest:     true,
				WhitesourceProjectNames: []string{"Test Project"},
				WhitesourceProductToken: "productToken",
				UserToken:               "userToken",
			}
			utilsMock := newPPMSUtilsMock()
			err := runPPMSComplianceCheck(&config, &telemetryCustomData, &mappingSystem, &ppmsSystem, &whitesourceSystem, &blackDuckClient, utilsMock, dir, time.Microsecond)
			assert.NoError(t, err)
			assert.Equal(t, false, config.CreateBuildVersion)
			assert.Equal(t, false, config.UploadChangeRequest)
		})

		t.Run("Run Whitesource but both Whitesource and Blackduck tokens present", func(*testing.T) {
			config := sapCheckPPMSComplianceOptions{
				OrgToken:                "OrgToken",
				PpmsID:                  "1001",
				ServerURL:               "https://my.server.url",
				WhitesourceProductToken: "productToken",
				UserToken:               "userToken",
				WhitesourceProjectNames: []string{"Test Project"},
				DetectToken:             "detect",
			}
			assert.NoError(
				t,
				runPPMSComplianceCheck(
					&config,
					&telemetryCustomData,
					&mappingSystem,
					&ppmsSystem,
					&whitesourceSystem,
					&blackDuckClient,
					newPPMSUtilsMock(),
					dir,
					time.Microsecond,
				),
			)
		})

		t.Run("Run Blackduck but both Whitesource and Blackduck tokens present", func(*testing.T) {
			config := sapCheckPPMSComplianceOptions{
				PpmsID:                       "1001",
				ReportFileName:               "testReport",
				ServerURL:                    "https://my.server.url",
				BlackduckProjectName:         "SHC-PiperTest",
				DetectToken:                  "detect",
				Version:                      "1.0.0",
				VersioningModel:              "major-minor",
				UserToken:                    "token",
				RunComplianceCheckWithDetect: true,
			}
			assert.NoError(
				t,
				runPPMSComplianceCheck(
					&config,
					&telemetryCustomData,
					&mappingSystem,
					&ppmsSystem,
					&whitesourceSystem,
					&blackDuckClient,
					newPPMSUtilsMock(),
					dir,
					time.Microsecond,
				),
			)
		})
	})
}

func TestEffectiveBuildVersionName(t *testing.T) {

	t.Run("Success cases", func(t *testing.T) {
		tt := []struct {
			config   sapCheckPPMSComplianceOptions
			expected string
		}{
			{config: sapCheckPPMSComplianceOptions{}, expected: ""},
			{config: sapCheckPPMSComplianceOptions{BuildVersion: "testBuildVersion"}, expected: "testBuildVersion"},
			{config: sapCheckPPMSComplianceOptions{Version: "1.2.3"}, expected: ""},
			{config: sapCheckPPMSComplianceOptions{Version: "1.2.3", CreateBuildVersion: true}, expected: "1.2.3"},
		}

		for _, test := range tt {
			result, err := effectiveBuildVersionName(&test.config, &telemetry.CustomData{})
			assert.NoError(t, err)
			assert.Equal(t, test.expected, result)
		}
	})

	t.Run("Error cases", func(t *testing.T) {
		tt := []struct {
			config      sapCheckPPMSComplianceOptions
			expectedErr string
		}{
			{
				config:      sapCheckPPMSComplianceOptions{BuildVersion: "thisBuildVersionIsLongerThan30c"},
				expectedErr: `your PPMS Build version 'thisBuildVersionIsLongerThan30c' is longer than 30 characters. Please adapt the parameter 'buildVersion' accordingly`,
			},
			{
				config:      sapCheckPPMSComplianceOptions{BuildVersion: "invalid {{eq 7}}", Version: "1.2.3"},
				expectedErr: "invalid {{eq 7}}",
			},
		}

		for _, test := range tt {
			_, err := effectiveBuildVersionName(&test.config, &telemetry.CustomData{})
			assert.Contains(t, fmt.Sprint(err), test.expectedErr)
		}
	})
}

func TestRetrieveFoss(t *testing.T) {
	t.Run("Failure - retrieve build versions", func(t *testing.T) {
		ppmsSystem := ppms.System{HTTPClient: &httpErrorClient{}}
		_, _, err := retrieveFoss(&sapCheckPPMSComplianceOptions{}, "", &ppms.SoftwareComponentVersion{Name: "mySCV"}, &ppmsSystem, time.Microsecond)
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve build versions for SCV 'mySCV'")
	})

	t.Run("Success - build version exists", func(t *testing.T) {
		config := sapCheckPPMSComplianceOptions{CreateBuildVersion: false}
		buildVersionName := "myBuildVersion"
		expectedFoss := []ppms.Foss{{ID: "1001"}, {ID: "1002"}}
		expectedBV := ppms.BuildVersion{ID: "1", Name: buildVersionName, Foss: expectedFoss}

		ppmsScv := ppms.SoftwareComponentVersion{ID: "0001", Name: "mySCV", BuildVersions: []ppms.BuildVersion{expectedBV}}
		ppmsSystem := ppms.System{}

		fossList, bv, err := retrieveFoss(&config, buildVersionName, &ppmsScv, &ppmsSystem, time.Microsecond)
		assert.NoError(t, err)
		assert.Equal(t, expectedBV, bv)
		assert.Equal(t, expectedFoss, fossList)
	})

	t.Run("Failure - build version exists", func(t *testing.T) {
		config := sapCheckPPMSComplianceOptions{CreateBuildVersion: false}
		buildVersionName := "myBuildVersion"
		expectedBV := ppms.BuildVersion{ID: "1", Name: buildVersionName}

		ppmsScv := ppms.SoftwareComponentVersion{ID: "0001", Name: "mySCV", BuildVersions: []ppms.BuildVersion{expectedBV}}
		ppmsSystem := ppms.System{HTTPClient: &httpErrorClient{}}

		_, _, err := retrieveFoss(&config, buildVersionName, &ppmsScv, &ppmsSystem, time.Microsecond)
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve FOSS for build version 'myBuildVersion'")
	})

	t.Run("Failure - build version does not exist", func(t *testing.T) {
		config := sapCheckPPMSComplianceOptions{CreateBuildVersion: false}
		buildVersionName := "myBuildVersion"

		ppmsScv := ppms.SoftwareComponentVersion{ID: "0001", Name: "mySCV", BuildVersions: []ppms.BuildVersion{}}
		ppmsSystem := ppms.System{}

		_, _, err := retrieveFoss(&config, buildVersionName, &ppmsScv, &ppmsSystem, time.Microsecond)
		assert.Contains(t, fmt.Sprint(err), "build version 'myBuildVersion' does not exist and createBuildVersion is set to 'false'")
	})

	t.Run("Success - create build version", func(t *testing.T) {
		config := sapCheckPPMSComplianceOptions{CreateBuildVersion: true, Username: "myPPMSUser"}

		buildVersionName := "myBuildVersion"
		expectedFoss := []ppms.Foss{{ID: "1001"}, {ID: "1002"}}
		expectedBV := ppms.BuildVersion{ID: "1", Name: buildVersionName, Foss: expectedFoss}

		ppmsScv := ppms.SoftwareComponentVersion{ID: "0001", Name: "mySCV", BuildVersions: []ppms.BuildVersion{}}
		client := httpMockClient{
			responseBodyForURL: map[string]string{
				"https://my.ppms.system/sap/internal/ppms/api/changerequest/v1/cvbv":                                              `{"crId":"cr1"}`,
				"https://my.ppms.system/sap/internal/ppms/api/changerequest/v1/cvbv/cr1":                                          `{"crId":"cr1","status":"APPLIED"}`,
				"https://my.ppms.system/odataint/borm/odataforosrcy/SoftwareComponentVersions('0001')/BuildVersions?$format=json": `{"d":{"results":[{"Id":"1","Name":"myBuildVersion"}]}}`,
				"https://my.ppms.system/odataint/borm/odataforosrcy/BuildVersions('1')/FreeOpenSourceSoftwares?$format=json":      `{"d":{"results":[{"Id":"1001"},{"Id":"1002"}]}}`,
			},
		}
		ppmsSystem := ppms.System{HTTPClient: &client, ServerURL: "https://my.ppms.system"}

		fossList, bv, err := retrieveFoss(&config, buildVersionName, &ppmsScv, &ppmsSystem, time.Microsecond)
		assert.NoError(t, err)
		assert.Equal(t, expectedBV, bv)
		assert.Equal(t, expectedFoss, fossList)
	})

	t.Run("Change request upload failure - create build version", func(t *testing.T) {
		config := sapCheckPPMSComplianceOptions{CreateBuildVersion: true, Username: "myPPMSUser"}

		buildVersionName := "myBuildVersion"

		ppmsScv := ppms.SoftwareComponentVersion{ID: "0001", Name: "mySCV", BuildVersions: []ppms.BuildVersion{}}
		client := httpMockClient{
			errorMessageForURL: map[string]string{
				"https://my.ppms.system/sap/internal/ppms/api/changerequest/v1/cvbv": "upload error",
			},
		}
		ppmsSystem := ppms.System{HTTPClient: &client, ServerURL: "https://my.ppms.system"}

		_, _, err := retrieveFoss(&config, buildVersionName, &ppmsScv, &ppmsSystem, time.Microsecond)
		assert.Contains(t, fmt.Sprint(err), "failed to upload change request to PPMS")
	})

	t.Run("Change request status failure - create build version", func(t *testing.T) {
		config := sapCheckPPMSComplianceOptions{CreateBuildVersion: true, Username: "myPPMSUser"}

		buildVersionName := "myBuildVersion"

		ppmsScv := ppms.SoftwareComponentVersion{ID: "0001", Name: "mySCV", BuildVersions: []ppms.BuildVersion{}}
		client := httpMockClient{
			errorMessageForURL: map[string]string{
				"https://my.ppms.system/sap/internal/ppms/api/changerequest/v1/cvbv/cr1": "upload error",
			},
			responseBodyForURL: map[string]string{
				"https://my.ppms.system/sap/internal/ppms/api/changerequest/v1/cvbv": `{"crId":"cr1"}`,
			},
		}
		ppmsSystem := ppms.System{HTTPClient: &client, ServerURL: "https://my.ppms.system"}

		_, _, err := retrieveFoss(&config, buildVersionName, &ppmsScv, &ppmsSystem, time.Microsecond)
		assert.Contains(t, fmt.Sprint(err), "change request returned with an error")
	})

	t.Run("Update failure - create build version", func(t *testing.T) {
		config := sapCheckPPMSComplianceOptions{CreateBuildVersion: true, Username: "myPPMSUser"}

		buildVersionName := "myBuildVersion"

		ppmsScv := ppms.SoftwareComponentVersion{ID: "0001", Name: "mySCV", BuildVersions: []ppms.BuildVersion{}}
		client := httpMockClient{
			errorMessageForURL: map[string]string{
				"https://my.ppms.system/odataint/borm/odataforosrcy/SoftwareComponentVersions('0001')/BuildVersions?$format=json": "update error",
			},
			responseBodyForURL: map[string]string{
				"https://my.ppms.system/sap/internal/ppms/api/changerequest/v1/cvbv":     `{"crId":"cr1"}`,
				"https://my.ppms.system/sap/internal/ppms/api/changerequest/v1/cvbv/cr1": `{"crId":"cr1","status":"APPLIED"}`,
			},
		}
		ppmsSystem := ppms.System{HTTPClient: &client, ServerURL: "https://my.ppms.system"}

		_, _, err := retrieveFoss(&config, buildVersionName, &ppmsScv, &ppmsSystem, time.Microsecond)
		assert.Contains(t, fmt.Sprint(err), "failed to update build versions for SCV 'mySCV'")
	})

	t.Run("Success - latest build version", func(t *testing.T) {
		config := sapCheckPPMSComplianceOptions{CreateBuildVersion: false}
		expectedFoss := []ppms.Foss{{ID: "1001"}, {ID: "1002"}}
		bv0 := ppms.BuildVersion{ID: "1", Name: "BV0", Foss: []ppms.Foss{}, SortSequence: "0"}
		expectedBV := ppms.BuildVersion{ID: "1", Name: "BV1", Foss: expectedFoss, SortSequence: "1"}

		ppmsScv := ppms.SoftwareComponentVersion{ID: "0001", Name: "mySCV", BuildVersions: []ppms.BuildVersion{bv0, expectedBV}}
		ppmsSystem := ppms.System{}

		fossList, bv, err := retrieveFoss(&config, "", &ppmsScv, &ppmsSystem, time.Microsecond)
		assert.NoError(t, err)
		assert.Equal(t, expectedBV, bv)
		assert.Equal(t, expectedFoss, fossList)
	})

	t.Run("FOSS retrieve failure - latest build version", func(t *testing.T) {
		config := sapCheckPPMSComplianceOptions{CreateBuildVersion: false}
		bv0 := ppms.BuildVersion{ID: "1", Name: "BV0", SortSequence: "0"}
		expectedBV := ppms.BuildVersion{ID: "1", Name: "BV1", SortSequence: "1"}

		ppmsScv := ppms.SoftwareComponentVersion{ID: "0001", Name: "mySCV", BuildVersions: []ppms.BuildVersion{bv0, expectedBV}}
		ppmsSystem := ppms.System{HTTPClient: &httpErrorClient{}}

		_, _, err := retrieveFoss(&config, "", &ppmsScv, &ppmsSystem, time.Microsecond)
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve FOSS for build version 'BV1'")
	})

	t.Run("Success - SCV", func(t *testing.T) {
		config := sapCheckPPMSComplianceOptions{CreateBuildVersion: false}
		expectedFoss := []ppms.Foss{{ID: "1001"}, {ID: "1002"}}

		ppmsScv := ppms.SoftwareComponentVersion{ID: "0001", Name: "mySCV", Foss: expectedFoss, BuildVersions: []ppms.BuildVersion{}}
		ppmsSystem := ppms.System{}

		fossList, bv, err := retrieveFoss(&config, "", &ppmsScv, &ppmsSystem, time.Microsecond)
		assert.NoError(t, err)
		assert.Equal(t, ppms.BuildVersion{}, bv)
		assert.Equal(t, expectedFoss, fossList)
	})

	t.Run("Failure - SCV", func(t *testing.T) {
		config := sapCheckPPMSComplianceOptions{CreateBuildVersion: false}

		ppmsScv := ppms.SoftwareComponentVersion{ID: "0001", Name: "mySCV", BuildVersions: []ppms.BuildVersion{}}
		ppmsSystem := ppms.System{HTTPClient: &httpErrorClient{}}

		_, _, err := retrieveFoss(&config, "", &ppmsScv, &ppmsSystem, time.Microsecond)
		assert.Contains(t, fmt.Sprint(err), "failed to retrieve FOSS for SCV 'mySCV'")
	})

}

func TestCompareFoss(t *testing.T) {
	current := []gtlc.Foss{
		{
			GroupID:                 "com.piper",
			ArtifactID:              "test1",
			Version:                 "1.0.0",
			FossID:                  "1001",
			IPComplianceDetails:     []gtlc.IPComplianceItem{{ReviewModel: "A"}},
			ExportComplianceDetails: []gtlc.ExportComplianceItem{{Progress: "ECCNASSIGNED1"}},
		},
		{
			GroupID:                 "com.piper",
			ArtifactID:              "test2",
			Version:                 "2.0.0",
			FossID:                  "1002",
			IPComplianceDetails:     []gtlc.IPComplianceItem{{ReviewModel: "B"}},
			ExportComplianceDetails: []gtlc.ExportComplianceItem{{Progress: "ECCNASSIGNED2"}},
		},
		{
			GroupID:                 "com.piper",
			ArtifactID:              "test3",
			Version:                 "3.0.0",
			IPComplianceDetails:     []gtlc.IPComplianceItem{{ReviewModel: "C"}},
			ExportComplianceDetails: []gtlc.ExportComplianceItem{{Progress: "ECCNASSIGNED3"}},
		},
	}
	target := []ppms.Foss{
		{
			ID: "1001",
		},
	}
	expected := fossMappingResult{
		failedMatches:      2,
		missingFossMapping: 1,
		successfulMatches:  1,
		totalLibraries:     3,
		fossMappings: []fossMapping{
			{
				GroupID:        "com.piper",
				ArtifactID:     "test1",
				Version:        "1.0.0",
				HasPPMSMapping: true,
				PpmsID:         "1001",
				IPDetails:      []gtlc.IPComplianceItem{{ReviewModel: "A"}},
				EccnDetails:    []gtlc.ExportComplianceItem{{Progress: "ECCNASSIGNED1"}},
			},
			{
				GroupID:        "com.piper",
				ArtifactID:     "test2",
				Version:        "2.0.0",
				HasPPMSMapping: false,
				PpmsID:         "1002",
				IPDetails:      []gtlc.IPComplianceItem{{ReviewModel: "B"}},
				EccnDetails:    []gtlc.ExportComplianceItem{{Progress: "ECCNASSIGNED2"}},
			},
			{
				GroupID:        "com.piper",
				ArtifactID:     "test3",
				Version:        "3.0.0",
				HasPPMSMapping: false,
				IPDetails:      []gtlc.IPComplianceItem{{ReviewModel: "C"}},
				EccnDetails:    []gtlc.ExportComplianceItem{{Progress: "ECCNASSIGNED3"}},
			},
		},
	}

	assert.Equal(t, expected, compareFoss(current, target))
}

func TestProcessFailedMatches(t *testing.T) {

	user := "testUser"
	wsProduct := "myWSProduct"
	buildVersionID := "myBuildVersionId"
	config := sapCheckPPMSComplianceOptions{Username: user, UploadChangeRequest: true, Version: "1.2.3"}
	ppmsScv := ppms.SoftwareComponentVersion{ID: "0001"}

	t.Run("Upload success", func(t *testing.T) {
		crMock := changeRequestSenderMock{returnCrID: "1"}
		utilsMock := newPPMSUtilsMock()
		result := fossMappingResult{missingFossMapping: 0, fossMappings: []fossMapping{{PpmsID: "fossId1"}, {PpmsID: "fossId2"}}}

		err := processFailedMatches(&config, wsProduct, &crMock, &ppmsScv, buildVersionID, &result, utilsMock, time.Microsecond)
		assert.NoError(t, err)
		assert.Equal(t, user, crMock.userID)
		assert.Equal(t, fmt.Sprintf("%v; %v", wsProduct, config.Version), crMock.source)
		assert.Equal(t, ppmsScv, *crMock.scv)
		assert.Equal(t, crMock.returnCrID, crMock.crID)
		assert.Equal(t, []ppms.ChangeRequestFoss{{PPMSFossNumber: "fossId1"}, {PPMSFossNumber: "fossId2"}}, crMock.foss)
	})

	t.Run("Upload send failure", func(t *testing.T) {
		crMock := changeRequestSenderMock{returnCrID: "1", sendError: fmt.Errorf("send error")}
		utilsMock := newPPMSUtilsMock()
		result := fossMappingResult{missingFossMapping: 0}

		err := processFailedMatches(&config, wsProduct, &crMock, &ppmsScv, buildVersionID, &result, utilsMock, time.Microsecond)
		assert.EqualError(t, err, "failed to upload change request to PPMS: send error")
	})

	t.Run("Upload wait failure", func(t *testing.T) {
		crMock := changeRequestSenderMock{returnCrID: "1", waitError: fmt.Errorf("wait error")}
		utilsMock := newPPMSUtilsMock()
		result := fossMappingResult{missingFossMapping: 0}

		err := processFailedMatches(&config, wsProduct, &crMock, &ppmsScv, buildVersionID, &result, utilsMock, time.Microsecond)
		assert.EqualError(t, err, "wait error")
	})

	t.Run("Missing mapping - ignored", func(t *testing.T) {
		crMock := changeRequestSenderMock{returnCrID: "1"}
		utilsMock := newPPMSUtilsMock()
		result := fossMappingResult{missingFossMapping: 1}

		err := processFailedMatches(&config, wsProduct, &crMock, &ppmsScv, buildVersionID, &result, utilsMock, time.Microsecond)
		assert.NoError(t, err)
	})

	t.Run("Failed matches", func(t *testing.T) {
		crMock := changeRequestSenderMock{returnCrID: "1"}
		utilsMock := newPPMSUtilsMock()

		result := fossMappingResult{
			failedMatches: 1,
			fossMappings: []fossMapping{
				{PpmsID: "fossId1", HasPPMSMapping: true},
				{PpmsID: "fossId2", HasPPMSMapping: false},
				{HasPPMSMapping: false},
			}}
		config := sapCheckPPMSComplianceOptions{UploadChangeRequest: false, ChangeRequestFileName: "myCrFile"}

		err := processFailedMatches(&config, wsProduct, &crMock, &ppmsScv, buildVersionID, &result, utilsMock, time.Microsecond)
		assert.EqualError(t, err, "1 PPMS entries are missing for this build. A report has been generated and stored as build artifact")
		fileContent, err := utilsMock.FileRead(config.ChangeRequestFileName)
		assert.NoError(t, err, "file not existing")
		assert.Contains(t, string(fileContent), `"ppmsFossNumber":"fossId2"`)
		assert.NotContains(t, string(fileContent), `"ppmsFossNumber":"fossId1"`)
	})

}

func TestConvertGroovyTemplate(t *testing.T) {
	telemetryCustomData := telemetry.CustomData{}
	tt := []struct {
		template string
		expected string
	}{
		{template: "", expected: ""},
		{template: "test", expected: "test"},
		{template: "${whatever}", expected: "${whatever}"},
		{template: "${version.major}", expected: "{{.Major}}"},
		{template: "${version.major}.${version.minor}.${version.patch}.${version.timestamp}", expected: "{{.Major}}.{{.Minor}}.{{.Patch}}.{{.Timestamp}}"},
		{template: "{{.Major}}", expected: "{{.Major}}"},
		{template: "{{.major}}.{{.minor}}.{{.patch}}.{{.timestamp}}", expected: "{{.Major}}.{{.Minor}}.{{.Patch}}.{{.Timestamp}}"},
	}

	for _, test := range tt {
		assert.Equal(t, test.expected, convertGroovyTemplate(test.template, &telemetryCustomData))
	}
}

func TestResolveBuildVersion(t *testing.T) {
	version, err := resolveBuildVersion("{{.Major}}.{{.Minor}}.{{.Patch}}.{{.Timestamp}}", "1.2.3-20200202+gitSha")
	assert.NoError(t, err)
	assert.Equal(t, "1.2.3.20200202", version)
}

func TestCreateReport(t *testing.T) {

	tt := []struct {
		result           fossMappingResult
		config           sapCheckPPMSComplianceOptions
		scv              ppms.SoftwareComponentVersion
		buildVersionName string
		ipScantype       string
		wsProductName    string
		bdVersion        string
		contains         []string
		notContains      []string
	}{
		// no entries & WhiteSource details
		{
			result:        fossMappingResult{totalLibraries: 0},
			config:        sapCheckPPMSComplianceOptions{ReportTitle: "myReportTitle", WhitesourceProjectNames: []string{"myP1", "myP2"}},
			scv:           ppms.SoftwareComponentVersion{Name: "mySCV"},
			wsProductName: "myWSProductName",
			ipScantype:    ipWhiteSource,
			contains:      []string{"myReportTitle", "WhiteSource product name:", "myWSProductName", "myP1, myP2", "mySCV", "Total number of libraries: 0", "No library entries found"},
			notContains:   []string{"Build version:"},
		},
		// no misses & WhiteSource details
		{
			result:           fossMappingResult{totalLibraries: 2, fossMappings: []fossMapping{{ArtifactID: "A1"}, {ArtifactID: "A2"}}},
			config:           sapCheckPPMSComplianceOptions{BuildVersion: "myBuildVersion"},
			scv:              ppms.SoftwareComponentVersion{},
			buildVersionName: "myBuildVersion",
			contains:         []string{"WhiteSource product name:", "Build version: myBuildVersion", "Total number of libraries: 2", "A1", "A2"},
			notContains:      []string{"No library entries found"},
		},
		// misses & no upload
		{
			result:      fossMappingResult{totalLibraries: 2, failedMatches: 1, fossMappings: []fossMapping{{ArtifactID: "A1"}, {ArtifactID: "A2"}}},
			config:      sapCheckPPMSComplianceOptions{},
			scv:         ppms.SoftwareComponentVersion{},
			contains:    []string{"You can send this to your PPMS entry owner"},
			notContains: []string{},
		},
		// misses & upload
		{
			result:      fossMappingResult{totalLibraries: 2, failedMatches: 1, fossMappings: []fossMapping{{ArtifactID: "A1"}, {ArtifactID: "A2"}}},
			config:      sapCheckPPMSComplianceOptions{UploadChangeRequest: true},
			scv:         ppms.SoftwareComponentVersion{},
			contains:    []string{"PPMS Change Request triggered automatically."},
			notContains: []string{},
		},
		// BlackDuck version & project name
		{
			result:        fossMappingResult{totalLibraries: 0},
			config:        sapCheckPPMSComplianceOptions{ReportTitle: "myReportTitle", DetectToken: "bdToken", BlackduckProjectName: "myBDProjectName"},
			scv:           ppms.SoftwareComponentVersion{Name: "mySCV"},
			ipScantype:    ipBlackDuck,
			wsProductName: "myWSProductName",
			bdVersion:     "1.2.3",
			contains:      []string{"1.2.3", "BlackDuck project name: myBDProjectName"},
		},

		// BlackDuck version & multiple project names
		{
			result:        fossMappingResult{totalLibraries: 0},
			config:        sapCheckPPMSComplianceOptions{ReportTitle: "myReportTitle", DetectToken: "bdToken", BlackduckProjectNames: []string{"myBDProjectName1", "myBDProjectName2"}},
			scv:           ppms.SoftwareComponentVersion{Name: "mySCV"},
			ipScantype:    ipBlackDuck,
			wsProductName: "myWSProductName",
			bdVersion:     "1.2.3",
			contains:      []string{"1.2.3", "BlackDuck project names: myBDProjectName1, myBDProjectName2"},
		},
	}

	for _, test := range tt {
		utilsMock := newPPMSUtilsMock()
		report := createReport(test.result, &test.config, &test.scv, test.buildVersionName, test.ipScantype, test.wsProductName, test.bdVersion, utilsMock)
		byteResult, err := report.ToHTML()
		assert.NoError(t, err)
		resultString := string(byteResult)
		for _, c := range test.contains {
			assert.Contains(t, resultString, c)
		}
		for _, c := range test.notContains {
			assert.NotContains(t, resultString, c)
		}
	}
}

func TestComprisedTextWithLink(t *testing.T) {
	tt := []struct {
		foss     fossMapping
		config   sapCheckPPMSComplianceOptions
		expected string
	}{
		{foss: fossMapping{HasPPMSMapping: true, PpmsID: "0001"}, config: sapCheckPPMSComplianceOptions{}, expected: "is comprised"},
		{foss: fossMapping{HasPPMSMapping: false, PpmsID: "0001"}, config: sapCheckPPMSComplianceOptions{UploadChangeRequest: false}, expected: "Declare FOSS Usage"},
		{foss: fossMapping{HasPPMSMapping: false, PpmsID: "0001"}, config: sapCheckPPMSComplianceOptions{UploadChangeRequest: true}, expected: "PPMS change has been triggered automatically"},
		{foss: fossMapping{HasPPMSMapping: false, PpmsID: ""}, config: sapCheckPPMSComplianceOptions{}, expected: "Mapping not found - FOSS has no PPMS ID yet"},
	}

	for _, test := range tt {
		assert.Contains(t, comprisedTextWithLink(test.foss, &test.config), test.expected)
	}
}

func TestResolveRiskRatingView(t *testing.T) {
	tt := []struct {
		riskRating    string
		expectedText  string
		expectedStyle reporting.ColumnStyle
	}{
		{riskRating: "ANY", expectedText: "Unknown until FOSS is comprised in SCV", expectedStyle: reporting.Grey},
		{riskRating: "", expectedText: "Unknown until FOSS is comprised in SCV", expectedStyle: reporting.Grey},
		{riskRating: "GREEN", expectedText: "Ok", expectedStyle: reporting.Green},
		{riskRating: "YELLOW", expectedText: "Medium Risk", expectedStyle: reporting.Yellow},
		{riskRating: "RED", expectedText: "High Risk", expectedStyle: reporting.Red},
		{riskRating: "GREY", expectedText: "Not yet rated", expectedStyle: reporting.Grey},
		{riskRating: "BLACK", expectedText: "Do not use!", expectedStyle: reporting.Black},
	}

	for _, test := range tt {
		text, style := resolveRiskRatingView(test.riskRating)
		assert.Equal(t, test.expectedText, text)
		assert.Equal(t, test.expectedStyle, style)
	}
}

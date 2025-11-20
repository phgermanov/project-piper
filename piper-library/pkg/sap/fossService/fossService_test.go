//go:build unit
// +build unit

package fossService

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	http2 "github.com/SAP/jenkins-library/pkg/http"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type senderMock struct{}

func (s *senderMock) SendRequest(_, _ string, _ io.Reader, _ http.Header, _ []*http.Cookie) (*http.Response, error) {
	return nil, nil
}

func (s *senderMock) SetOptions(_ http2.ClientOptions) {

}

func TestInstance_Call(t *testing.T) {
	var instanceMock = NewInstance(MustURL("https://foss-service.wdf.sap.corp"))
	instanceMock.sender = &senderMock{}
}

func TestReadResponseBody(t *testing.T) {
	var resp *http.Response
	var body JsonObjectString
	var err error

	resp = &http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"foo":"bar"}`)))}
	body, err = ReadResponseBody(resp, errors.New("http response error"))
	assert.EqualError(t, err, "http response error")
	assert.EqualValues(t, JsonObjectString(`{"foo":"bar"}`), body)

	resp = &http.Response{Body: io.NopCloser(bytes.NewReader([]byte{}))}
	body, err = ReadResponseBody(resp, nil)
	assert.NoError(t, err)
	assert.EqualValues(t, EmptyBody, body)

	resp = &http.Response{}
	body, err = ReadResponseBody(resp, nil)
	assert.NoError(t, err)
	assert.EqualValues(t, EmptyBody, body)

	//to provoke an error in io.ReadAll seams to be hard
}

// Just tests a view important (general) cases
func TestReadAsAnalysisInput(t *testing.T) {
	var ai *AnalysisInput
	var err error

	ai, err = JsonObjectString(`{}`).ReadAsAnalysisInput()
	assert.NoError(t, err)
	assert.EqualValues(t, AnalysisInput{}, *ai)

	ai, err = JsonObjectString(`{
        "meta" : {
            "scmIdentification" : {
                "scmUrl"   : "http://foo/bar",
                "revision" : "master"
            },
            "pipelineIdentification" : {
                "globalIdentifier" : "70db0462-e33e-4da5-8ff8-e3acb2b44dc9",
                "backUrl" : "http://foo/back/bar"
            },
            "predecessorVia" : "scmIdentification",
            "tags" : ["foo","bar"]
        },
        "source" : [{
            "vendor" : "WS",
            "product" : "abc",
            "project" : "api - v1"
        }, {
            "vendor" : "WS",
            "product" : "abc",
            "project" : "impl - v1"
        }],
        "target" : [{
            "vendor" : "PPMS",
            "scv" : "fooSCV",
            "buildVersion" : "v1"
        }],
        "options" : {
            "skipCache" : true,
            "checkCompliance" : true,
            "goToMarketChannels" : ["GTMC_1", "GTMC_3"]
        }
    }`).ReadAsAnalysisInput()
	assert.NoError(t, err)
	assert.EqualValues(t, AnalysisInput{
		ClientInformation: ClientInformation{
			ScmIdentification: ScmIdentification{
				ScmUrl:   "http://foo/bar",
				Revision: "master",
			},
			PipelineIdentification: PipelineIdentification{
				GlobalIdentifier: uuid.MustParse("70db0462-e33e-4da5-8ff8-e3acb2b44dc9"),
				BackUrl:          "http://foo/back/bar",
			},
			PredecessorVia: "scmIdentification",
			Tags:           []string{"foo", "bar"},
		},
		Source: []BomIdentifier{{
			Vendor:         "WS",
			WsProduct:      "abc",
			WsOrHubProject: "api - v1",
		}, {
			Vendor:         "WS",
			WsProduct:      "abc",
			WsOrHubProject: "impl - v1",
		}},
		Target: []BomIdentifier{{
			Vendor:           "PPMS",
			PpmsScv:          "fooSCV",
			PpmsBuildVersion: "v1",
		}},
		Options: AnalysisOptions{
			SkipCache:          true,
			CheckCompliance:    true,
			GoToMarketChannels: []string{"GTMC_1", "GTMC_3"},
		},
	}, *ai)

	ai, err = JsonObjectString(`{wrong Json}`).ReadAsAnalysisInput()
	assert.Error(t, err)
	assert.Nil(t, ai)
}

func TestGetApi(t *testing.T) {
	var api *FossRestEndpointDescriptor
	var err error

	api, err = RestApiRegistry.GetApi("analyze")
	assert.NoError(t, err)
	assert.Equal(t, "analyze", api.MethodName)
	assert.Equal(t, "POST", api.HttpMethod)
	assert.Equal(t, "application/json", api.httpContent)
	assert.Equal(t, "application/json", api.httpAccept)
	assert.Equal(t, 1, api.NumOfParams)

	api, err = RestApiRegistry.GetApi("compare")
	assert.NoError(t, err)
	assert.Equal(t, "compare", api.MethodName)
	assert.Equal(t, "GET", api.HttpMethod)
	assert.Equal(t, "*/*", api.httpContent)
	assert.Equal(t, "application/json", api.httpAccept)
	assert.Equal(t, 2, api.NumOfParams)

	api, err = RestApiRegistry.GetApi("getResultCertificate")
	assert.NoError(t, err)
	assert.Equal(t, "getResultCertificate", api.MethodName)
	assert.Equal(t, "GET", api.HttpMethod)
	assert.Equal(t, "*/*", api.httpContent)
	assert.Equal(t, "application/json", api.httpAccept)
	assert.Equal(t, 1, api.NumOfParams)

	api, err = RestApiRegistry.GetApi("sendPpmsChangeRequest")
	assert.NoError(t, err)
	assert.Equal(t, "sendPpmsChangeRequest", api.MethodName)
	assert.Equal(t, "POST", api.HttpMethod)
	assert.Equal(t, "application/json", api.httpContent)
	assert.Equal(t, "application/json", api.httpAccept)
	assert.Equal(t, 1, api.NumOfParams)

	api, err = RestApiRegistry.GetApi("createPpmsBuildVersion")
	assert.NoError(t, err)
	assert.Equal(t, "createPpmsBuildVersion", api.MethodName)
	assert.Equal(t, "POST", api.HttpMethod)
	assert.Equal(t, "application/json", api.httpContent)
	assert.Equal(t, "application/json", api.httpAccept)
	assert.Equal(t, 1, api.NumOfParams)

	assert.Equal(t, 5, len(RestApiRegistry.methods), "(All) 5 API methods need to be registered.")

	api, err = RestApiRegistry.GetApi("foo")
	assert.Error(t, err)
	assert.Nil(t, api)
	assert.Equal(t, "Foss API method 'foo' is not known", err.Error())
}

// I first tried some simple testing mechanisms to get a feeling for the language.
// In this case, a SimpleFossRequest should be invalid (if validated via Validate()), iff there is no
// pipelineGUID given
func TestValidatePipelineIdentification(t *testing.T) {
	p := PipelineIdentification{}
	validate, problems := p.Validate()
	fmt.Println(p.GlobalIdentifier)

	assert.False(t, validate, "Result must be invalid because a global pipeline Identifier is missing")
	assert.True(t, len(problems) > 0, "The problem should be reported, but was missing")
	assert.Contains(t, problems[0].Error(), "pipelineGUID", "Error message must hint on the missing pipelineGUID")
}

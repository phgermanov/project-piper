//go:build unit
// +build unit

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	fs "github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/fossService"
)

var analyzeRequestBody = fs.JsonObjectString(fmt.Sprintf(
	`{
        "source": [
         {
           "vendor": "ws",
           "product": "SHC-DLM RELEASE CHECK OD 1.0",
           "project": "FnfDlmUtils-api - 0.10.0"
         },
         {
           "vendor": "ws",
           "product": "SHC-DLM RELEASE CHECK OD 1.0",
           "project": "FnfDlmUtils-impl - 0.10.0"
         }
        ],
       "target": [
         {
           "vendor": "ppms",
           "scv": "RELEASE CHECK CORE OD 1.0",
           "buildVersion": "piperTest - %v"
         }
       ],
       "options": {
         "skipCache": true,
         "checkCompliance": true
       }
    }`, uuid.New().String()[0:8]) /* Should not be too long - we speak about PPMS */)

type FossApiMockInstance struct {
	fs.Instance
	ExpectedResults []string
}

func (fossServer *FossApiMockInstance) GoCreatePpmsBuildVersion(_ *fs.ChangeRequestCreateBuildVersionInput) (fs.JsonObjectString, error) {
	return `{"GoCreatePpmsBuildVersion":"called"}`, nil
}

func (fossServer *FossApiMockInstance) GoAnalyze(_ *fs.AnalysisInput) (fs.JsonObjectString, error) {
	return `{"GoAnalyze":"called", "resultId" : "70db0462-e33e-4da5-8ff8-e3acb2b44dc9"}`, nil
}

func (fossServer *FossApiMockInstance) GoSendPpmsChangeRequest(_ *fs.ChangeRequestInput) (fs.JsonObjectString, error) {
	return `{"GoAnalyze":"called"}`, nil
}

func TestExecuteDefaultBehavior(t *testing.T) {
	var config *sapCallFossServiceConfig
	var compResult fs.JsonObjectString
	var err error

	//call PPMS update
	mock := &FossApiMockInstance{}
	config = &sapCallFossServiceConfig{
		options: &sapCallFossServiceOptions{Username: "usr", Password: "pwd"},
		analyzeInput: &fs.AnalysisInput{
			Target: []fs.BomIdentifier{{Vendor: fs.PPMS}},
		},
		fossApi: mock,
	}
	compResult, err = executeDefaultFossBehavior(config)
	assert.NoError(t, err)
	assert.Equal(t, `{"analyseResult":{"GoAnalyze":"called","resultId":"70db0462-e33e-4da5-8ff8-e3acb2b44dc9"},"createBuildVersionResult":{},"errors":{},"ppmsUpdateResult":{"GoAnalyze":"called"}}`, compResult.String())

	//leave out PPMS update
	config = &sapCallFossServiceConfig{
		options: &sapCallFossServiceOptions{},
		fossApi: mock,
	}
	compResult, err = executeDefaultFossBehavior(config)
	assert.NoError(t, err)
	assert.Equal(t, `{"analyseResult":{"GoAnalyze":"called","resultId":"70db0462-e33e-4da5-8ff8-e3acb2b44dc9"},"createBuildVersionResult":{},"errors":{},"ppmsUpdateResult":{}}`, compResult.String())
}

func TestReadFossServiceIntegrationOptions(t *testing.T) {
	var err error
	var options *sapCallFossServiceOptions
	var config *sapCallFossServiceConfig

	uuid1, uuid2 := uuid.New(), uuid.New()
	options = &sapCallFossServiceOptions{
		PipelineIDentification: uuid1.String() + ";" + uuid2.String(),
		CallAPI:                `analyze;{"foo", "bar"}`,
		FossBaseURL:            "http://my-service.url",
	}
	config, err = readFossServiceIntegrationOptions(options)
	assert.NoError(t, err)
	assert.Equal(t, fs.PipelineIdentification{
		GlobalIdentifier:   uuid1,
		InstanceIdentifier: uuid2,
	}, *config.pipelineIdentification)

	//error in pipelineIdentification must be send back
	options.PipelineIDentification = uuid1.String() + "," + uuid2.String() //wrong separator"
	config, err = readFossServiceIntegrationOptions(options)
	assert.Error(t, err)
	assert.Nil(t, config)

	//scmIdentification
	options = &sapCallFossServiceOptions{
		ScmIDentification: "http://foo/bar;master",
		CallAPI:           `analyze;{"foo", "bar"}`,
		FossBaseURL:       "http://my-service.url",
	}
	config, err = readFossServiceIntegrationOptions(options)
	assert.NoError(t, err)
	assert.EqualValues(t, fs.ScmIdentification{
		ScmUrl:   "http://foo/bar",
		Revision: "master",
	}, *config.scmIdentification)

	//error in scmIdentification must be send back
	options.ScmIDentification = "http://foo/bar,master" //wrong separator"
	config, err = readFossServiceIntegrationOptions(options)
	assert.Error(t, err)
	assert.Nil(t, config)

	options = &sapCallFossServiceOptions{
		AnalyzeRequestBody: analyzeRequestBody.String(),
		CallAPI:            "stepBehavior",
		FossBaseURL:        "http://my-service.url",
	}
	config, err = readFossServiceIntegrationOptions(options)
	assert.NoError(t, err)
	assert.Nil(t, config.apiCall) //do not set apiCall executing default behavior

	options = &sapCallFossServiceOptions{CallAPI: "stepBehavior"}
	config, err = readFossServiceIntegrationOptions(options)
	assert.Nil(t, config)
	assert.Contains(t, msgOf(err),
		"default step behavior is executed, the analyze request body must be given")

	//analyze request body errors must be transmitted back
	options = &sapCallFossServiceOptions{
		AnalyzeRequestBody: `{foo:bar}`,
		CallAPI:            "stepBehavior",
	}
	config, err = readFossServiceIntegrationOptions(options)
	assert.Nil(t, config)
	assert.Error(t, err)

	options = &sapCallFossServiceOptions{CallAPI: `analyze;{"foo", "bar"}`}
	config, err = readFossServiceIntegrationOptions(options)
	assert.Error(t, err)
	assert.Nil(t, config)

	options = &sapCallFossServiceOptions{
		CallAPI:     `analyze;{"foo", "bar"}`,
		FossBaseURL: "http://my-service.url",
	}
	config, err = readFossServiceIntegrationOptions(options)
	assert.NoError(t, err)
	instance := fs.NewInstance(fs.MustURL("http://my-service.url"))
	assert.Equal(t, &instance, config.fossApi) //default instance should be prod

	options.FossBaseURL = "justScheme:/"
	config, err = readFossServiceIntegrationOptions(options)
	assert.Nil(t, config)
	assert.Contains(t, msgOf(err), "no valid FOSS service base URL")

	options.FossBaseURL = "justHost.url"
	config, err = readFossServiceIntegrationOptions(options)
	assert.Nil(t, config)
	assert.Contains(t, msgOf(err), "no valid FOSS service base URL")

}

func TestFossStepDefaultResultAsJsonString(t *testing.T) {
	var defaultResult FossDefaultStepResult
	var json fs.JsonObjectString
	var err error

	defaultResult = FossDefaultStepResult{}
	json, err = defaultResult.asJsonString()
	assert.Equal(t, `{"analyseResult":{},"createBuildVersionResult":{},"errors":{},"ppmsUpdateResult":{}}`, string(json))
	assert.Contains(t, msgOf(err), "create build version problems")
	assert.Contains(t, msgOf(err), "analyse problems")
	assert.Contains(t, msgOf(err), "PPMS update problems")
	assert.Contains(t, msgOf(err), "CreateBuildVersionResult was not set")
	assert.Contains(t, msgOf(err), "AnalyseResult was not set")
	assert.Contains(t, msgOf(err), "PpmsUpdateResult was not set")

	defaultResult = FossDefaultStepResult{
		CreateBuildVersionResult: fs.EmptyObject,
		PpmsUpdateResult:         fs.EmptyObject,
		AnalyseResult:            fs.EmptyObject}
	json, err = defaultResult.asJsonString()
	assert.NoError(t, err)
	assert.Equal(t, `{"analyseResult":{},"createBuildVersionResult":{},"errors":{},"ppmsUpdateResult":{}}`, json.String())

	defaultResult.CreateBuildVersionError = errors.New("Err1")
	_, err = defaultResult.asJsonString()
	assert.NotContains(t, msgOf(err), "PPMS update problems")
	assert.NotContains(t, msgOf(err), "analyse problems")
	assert.Contains(t, msgOf(err), "create build version problems")
	assert.Contains(t, msgOf(err), "Err1")

	defaultResult.PpmsUpdateError = errors.New("Err2")
	_, err = defaultResult.asJsonString()
	assert.Contains(t, msgOf(err), "create build version problems")
	assert.Contains(t, msgOf(err), "Err1")
	assert.Contains(t, msgOf(err), "PPMS update problems")
	assert.Contains(t, msgOf(err), "Err2")
	assert.NotContains(t, msgOf(err), "analyse problems")

	defaultResult.AnalyseError = errors.New("Err3")
	_, err = defaultResult.asJsonString()
	assert.Contains(t, msgOf(err), "create build version problems")
	assert.Contains(t, msgOf(err), "Err1")
	assert.Contains(t, msgOf(err), "PPMS update problems")
	assert.Contains(t, msgOf(err), "Err2")
	assert.Contains(t, msgOf(err), "analyse problems")
	assert.Contains(t, msgOf(err), "Err3")

	defaultResult = FossDefaultStepResult{
		CreateBuildVersionResult: `{"key" : notValidValue }`,
		PpmsUpdateResult:         `{ ivalidKey : "validValue" }`,
		AnalyseResult:            `{"key" : "value" ,  }`}
	json, err = defaultResult.asJsonString()
	assert.Equal(t, `{"analyseResult":{},"createBuildVersionResult":{},"errors":{},"ppmsUpdateResult":{}}`, string(json))
	assert.Contains(t, msgOf(err), `invalid character 'o' in literal null (expecting 'u') Response was: {"key" : notValidValue }`)
	assert.Contains(t, msgOf(err), `invalid character 'i' looking for beginning of object key string Response was: { ivalidKey : "validValue" }`)
	assert.Contains(t, msgOf(err), `invalid character '}' looking for beginning of object key string Response was: {"key" : "value" ,  }`)
}

func TestCanUpdatePpms(t *testing.T) {

	config := &sapCallFossServiceConfig{
		options: &sapCallFossServiceOptions{
			Username: "usr",
			Password: "pwd"},
		analyzeInput: &fs.AnalysisInput{
			Target: []fs.BomIdentifier{{Vendor: fs.PPMS}}}}

	assert.True(t, canUpdatePpms(config))

	config.options.Username, config.options.Password = "", "pwd"
	assert.False(t, canUpdatePpms(config))

	config.options.Username, config.options.Password = "usr", ""
	assert.False(t, canUpdatePpms(config))

	config.options.Username, config.options.Password = "", ""
	assert.False(t, canUpdatePpms(config))

	config.analyzeInput = &fs.AnalysisInput{}
	assert.False(t, canUpdatePpms(config))

	config.analyzeInput = &fs.AnalysisInput{Target: []fs.BomIdentifier{{Vendor: fs.WS}}}
	assert.False(t, canUpdatePpms(config))

	config.analyzeInput = &fs.AnalysisInput{Target: []fs.BomIdentifier{{Vendor: fs.PPMS}, {Vendor: fs.PPMS}}}
	assert.False(t, canUpdatePpms(config))
}

func (fossServer *FossApiMockInstance) GetResultCertificate(source uuid.UUID) (fs.JsonObjectString, error) {
	fossServer.ExpectedResults = []string{"getResultCertificate", source.String()}
	return "good", nil
}

func (fossServer *FossApiMockInstance) Compare(source uuid.UUID, target uuid.UUID) (fs.JsonObjectString, error) {
	fossServer.ExpectedResults = []string{"compare", source.String(), target.String()}
	return "good", nil
}

func (fossServer *FossApiMockInstance) Call(api *fs.FossRestEndpointDescriptor, jsonBody fs.JsonObjectString, queryParams map[string]string) (*http.Response, error) {
	log.Entry().Debug(queryParams)
	fossServer.ExpectedResults = []string{api.MethodName, jsonBody.String()}
	return &http.Response{Body: io.NopCloser(bytes.NewReader([]byte(`{"foo":"bar"}`)))}, nil
}

func TestExecuteApiCall(t *testing.T) {

	var instance = &FossApiMockInstance{}
	var fsiError error
	var res fs.JsonObjectString
	resultId1, resultId2 := uuid.New().String(), uuid.New().String()

	res, fsiError = executeApiCall(instance, &ApiCall{method: "analyze", params: []string{`{"some","Json"}`}})
	assert.NoError(t, fsiError)
	assert.Equal(t, []string{"analyze", `{"some","Json"}`}, instance.ExpectedResults)
	assert.Equal(t, `{"foo":"bar"}`, res.String())

	res, fsiError = executeApiCall(instance, &ApiCall{method: "sendPpmsChangeRequest", params: []string{`{"ppms","Change"}`}})
	assert.NoError(t, fsiError)
	assert.Equal(t, []string{"sendPpmsChangeRequest", `{"ppms","Change"}`}, instance.ExpectedResults)
	assert.Equal(t, `{"foo":"bar"}`, res.String())

	res, fsiError = executeApiCall(instance, &ApiCall{method: "createPpmsBuildVersion", params: []string{`{"ppms","Version"}`}})
	assert.NoError(t, fsiError)
	assert.Equal(t, []string{"createPpmsBuildVersion", `{"ppms","Version"}`}, instance.ExpectedResults)
	assert.Equal(t, `{"foo":"bar"}`, res.String())

	res, fsiError = executeApiCall(instance, &ApiCall{method: "compare", params: []string{resultId1, resultId2}})
	assert.NoError(t, fsiError)
	assert.Equal(t, []string{"compare", resultId1, resultId2}, instance.ExpectedResults)
	assert.Equal(t, "good", res.String())

	res, fsiError = executeApiCall(instance, &ApiCall{method: "getResultCertificate", params: []string{resultId1}})
	assert.NoError(t, fsiError)
	assert.Equal(t, []string{"getResultCertificate", resultId1}, instance.ExpectedResults)
	assert.Equal(t, "good", res.String())

	res, fsiError = executeApiCall(instance, &ApiCall{method: "forSureUnknown", params: []string{resultId1}})
	assert.Equal(t, fs.EmptyBody, res)
	assert.Contains(t, msgOf(fsiError), "could not successfully execute forSureUnknown, Reason: Could not found method forSureUnknown")
}

func TestReadPipelineIdentification(t *testing.T) {
	var options *sapCallFossServiceOptions
	var pipelineIdentification *fs.PipelineIdentification
	var err error

	//happyPath
	globalId := uuid.New()
	options = &sapCallFossServiceOptions{PipelineIDentification: globalId.String()}
	pipelineIdentification, err = options.readPipelineIdentification()
	assert.NoError(t, err)
	assert.Equal(t, fs.PipelineIdentification{GlobalIdentifier: globalId}, *pipelineIdentification)

	instanceId := uuid.New()
	options = &sapCallFossServiceOptions{PipelineIDentification: globalId.String() + ";" + instanceId.String()}
	pipelineIdentification, err = options.readPipelineIdentification()
	assert.NoError(t, err)
	assert.Equal(t, fs.PipelineIdentification{
		GlobalIdentifier: globalId, InstanceIdentifier: instanceId}, *pipelineIdentification)

	backUrl := "https://github.wdf.sap.corp/DLM-FOSS/Foss"
	options = &sapCallFossServiceOptions{PipelineIDentification: globalId.String() + ";" + instanceId.String() + ";" + backUrl}
	pipelineIdentification, err = options.readPipelineIdentification()
	assert.NoError(t, err)
	assert.Equal(t, fs.PipelineIdentification{
		GlobalIdentifier: globalId, InstanceIdentifier: instanceId, BackUrl: backUrl}, *pipelineIdentification)

	//not given
	options = &sapCallFossServiceOptions{}
	pipelineIdentification, err = options.readPipelineIdentification()
	assert.NoError(t, err)
	assert.Nil(t, pipelineIdentification)

	//empty
	options = &sapCallFossServiceOptions{PipelineIDentification: ""}
	pipelineIdentification, err = options.readPipelineIdentification()
	assert.NoError(t, err)
	assert.Nil(t, pipelineIdentification)

	//no uuid
	options = &sapCallFossServiceOptions{PipelineIDentification: "wrong1"}
	pipelineIdentification, err = options.readPipelineIdentification()
	assert.Nil(t, pipelineIdentification)
	assert.Contains(t, msgOf(err), "first parameter of pipeline identification must be a valid uuid")

	options = &sapCallFossServiceOptions{PipelineIDentification: globalId.String() + ";wrong2"}
	pipelineIdentification, err = options.readPipelineIdentification()
	assert.Nil(t, pipelineIdentification)
	assert.Contains(t, msgOf(err), "second parameter of pipeline identification must be a valid uuid")

	//wrong separator
	options = &sapCallFossServiceOptions{PipelineIDentification: globalId.String() + "," + instanceId.String()}
	pipelineIdentification, err = options.readPipelineIdentification()
	assert.Nil(t, pipelineIdentification)
	assert.Contains(t, msgOf(err), "first parameter of pipeline identification must be a valid uuid")

	invalidBackUrl1 := "git@github.wdf.sap.corp:DLM-CoDePipeS/CoDePipeS.git"
	options = &sapCallFossServiceOptions{PipelineIDentification: globalId.String() + ";" + instanceId.String() + ";" + invalidBackUrl1}
	pipelineIdentification, err = options.readPipelineIdentification()
	assert.Nil(t, pipelineIdentification)
	assert.Contains(t, msgOf(err), "third parameter of pipeline identification must be a valid URL")

	invalidBackUrl2 := "foo"
	options = &sapCallFossServiceOptions{PipelineIDentification: globalId.String() + ";" + instanceId.String() + ";" + invalidBackUrl2}
	pipelineIdentification, err = options.readPipelineIdentification()
	assert.Nil(t, pipelineIdentification)
	assert.Contains(t, msgOf(err), "third parameter of pipeline identification must be a valid URL")

	//even path must be given
	invalidBackUrl3 := "http://foo"
	options = &sapCallFossServiceOptions{PipelineIDentification: globalId.String() + ";" + instanceId.String() + ";" + invalidBackUrl3}
	pipelineIdentification, err = options.readPipelineIdentification()
	assert.Nil(t, pipelineIdentification)
	assert.Contains(t, msgOf(err), "third parameter of pipeline identification must be a valid URL")

	//scheme must be given
	invalidBackUrl4 := "/foo/bar"
	options = &sapCallFossServiceOptions{PipelineIDentification: globalId.String() + ";" + instanceId.String() + ";" + invalidBackUrl4}
	pipelineIdentification, err = options.readPipelineIdentification()
	assert.Nil(t, pipelineIdentification)
	assert.Contains(t, msgOf(err), "third parameter of pipeline identification must be a valid URL")
}

func TestReadScmIdentification(t *testing.T) {
	var options *sapCallFossServiceOptions
	var scmIdentification *fs.ScmIdentification
	var err error
	//happy path
	options = &sapCallFossServiceOptions{ScmIDentification: `https://github.wdf.sap.corp/DLM-FOSS/Foss;master`}
	scmIdentification, err = options.readScmIdentification()
	assert.NoError(t, err)
	assert.Equal(t, fs.ScmIdentification{ScmUrl: "https://github.wdf.sap.corp/DLM-FOSS/Foss",
		Revision: "master"}, *scmIdentification)

	//check failure handling
	//do not support checkout via ssh currently
	options = &sapCallFossServiceOptions{ScmIDentification: `git@github.wdf.sap.corp:DLM-CoDePipeS/CoDePipeS.git;master`}
	scmIdentification, err = options.readScmIdentification()
	assert.Nil(t, scmIdentification)
	assert.Contains(t, msgOf(err), "first path segment in URL cannot contain colon")

	//option not given
	options = &sapCallFossServiceOptions{}
	scmIdentification, err = options.readScmIdentification()
	assert.NoError(t, err)
	assert.Nil(t, scmIdentification)

	options = &sapCallFossServiceOptions{ScmIDentification: `fooServer;master`}
	scmIdentification, err = options.readScmIdentification()
	assert.Nil(t, scmIdentification)
	assert.Contains(t, msgOf(err), "At least host and scheme must be included")

	options = &sapCallFossServiceOptions{ScmIDentification: `http:/;master`}
	scmIdentification, err = options.readScmIdentification()
	assert.Nil(t, scmIdentification)
	assert.Contains(t, msgOf(err), "At least host and scheme must be included")

	//wrong separator
	options = &sapCallFossServiceOptions{ScmIDentification: `https://github.wdf.sap.corp/DLM-FOSS/Foss,master`}
	scmIdentification, err = options.readScmIdentification()
	assert.Nil(t, scmIdentification)
	assert.EqualError(t, err, "scmIdentification need to have two parameters separated by semicolon. Given: 1")

	options = &sapCallFossServiceOptions{ScmIDentification: `https://github.wdf.sap.corp/DLM-FOSS/Foss;master;0x1`}
	scmIdentification, err = options.readScmIdentification()
	assert.Nil(t, scmIdentification)
	assert.EqualError(t, err, "scmIdentification need to have two parameters separated by semicolon. Given: 3")
}

func TestReadCallApi(t *testing.T) {
	var options *sapCallFossServiceOptions
	var apiCall *ApiCall
	var err error
	//happy path
	options = &sapCallFossServiceOptions{CallAPI: `analyze;{"some","Json"}`}
	apiCall, err = options.readCallApi()
	assert.NoError(t, err)
	assert.Equal(t, ApiCall{method: "analyze", params: []string{`{"some","Json"}`}}, *apiCall)

	//check failure handling
	options = &sapCallFossServiceOptions{}
	apiCall, err = options.readCallApi()
	assert.Nil(t, apiCall)
	assert.Equal(t, msgOf(err), "Foss API method '' is not known")

	options = &sapCallFossServiceOptions{CallAPI: "foo;{}"}
	apiCall, err = options.readCallApi()
	assert.Nil(t, apiCall)
	assert.Equal(t, msgOf(err), "Foss API method 'foo' is not known")

	options = &sapCallFossServiceOptions{CallAPI: "analyze;foo;bar"}
	apiCall, err = options.readCallApi()
	assert.Nil(t, apiCall)
	assert.Equal(t, msgOf(err),
		"For the method 'analyze' you have to specify 1 parameter(s). Found: 2 (after method name)")

	//read also wrong parameters (format is checked later)
	options = &sapCallFossServiceOptions{CallAPI: "compare;noUUID;noUUID"}
	apiCall, err = options.readCallApi()
	assert.NoError(t, err)
	assert.Equal(t, ApiCall{method: "compare", params: []string{"noUUID", "noUUID"}}, *apiCall)
}

func TestTokenizeString(t *testing.T) {
	res, err := TokenizeString("a,b,c", ',', '\\')
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, res)

	res, err = TokenizeString("a;b;c\\,d", ';', '\\')
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c,d"}, res)

	res, err = TokenizeString("a;b;c\\;d", ';', '\\')
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c;d"}, res)

	res, err = TokenizeString("a;b;c\\;;d", ';', '\\')
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c;", "d"}, res)

	res, err = TokenizeString("a;b;c\\;;;d", ';', '\\')
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c;", "", "d"}, res)

	res, err = TokenizeString("a;b;c+;d", ';', '+')
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c;d"}, res)
}

func msgOf(err error) string {
	if err == nil {
		return ""
	} else {
		return err.Error()
	}
}

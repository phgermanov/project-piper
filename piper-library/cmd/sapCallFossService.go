package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	. "github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/fossService"
)

func sapCallFossService(options sapCallFossServiceOptions, _ *telemetry.CustomData, commonPipelineEnvironment *sapCallFossServiceCommonPipelineEnvironment) {
	// Read and check inputs more closely

	var err error
	var readConfig *sapCallFossServiceConfig
	var result JsonObjectString

	readConfig, err = readFossServiceIntegrationOptions(&options)
	if err == nil {
		result, err = runSapCallFossService(readConfig)
		if err == nil {
			commonPipelineEnvironment.foss.resultJSON = result.String()
		}
	}

	if err != nil {
		log.Entry().Fatal(err)
	}

}

func runSapCallFossService(
	config *sapCallFossServiceConfig) (JsonObjectString, error) {

	var result JsonObjectString
	var err error
	if config.apiCall == nil {
		result, err = executeDefaultFossBehavior(config)
	} else {
		result, err = executeApiCall(config.fossApi, config.apiCall)
	}

	return result, err
}

func executeDefaultFossBehavior(
	config *sapCallFossServiceConfig) (JsonObjectString, error) {

	result := FossDefaultStepResult{PpmsUpdateResult: EmptyObject, CreateBuildVersionResult: EmptyObject}
	analyzeInput := config.analyzeInput
	pipelineIdentification := config.pipelineIdentification
	scmIdentification := config.scmIdentification
	fossApi := config.fossApi
	ppmsUsername := config.options.Username
	ppmsPassword := config.options.Password

	var buildUrl = ""
	if pipelineIdentification != nil {
		buildUrl = pipelineIdentification.BackUrl
		analyzeInput.ClientInformation.PipelineIdentification = *pipelineIdentification
	}

	if scmIdentification != nil {
		analyzeInput.ClientInformation.ScmIdentification = *scmIdentification
	}

	var updatePpms = canUpdatePpms(config)
	if updatePpms {
		// ensure that the build version is available, create it, if it is not existing.

		// build version to use is ... because canUpdatePpms ensures target PPMS with one entry
		ppmsBv := analyzeInput.Target[0].PpmsBuildVersion
		ppmsScv := analyzeInput.Target[0].PpmsScv
		if len(ppmsBv) > 0 {
			result.CreateBuildVersionResult, result.CreateBuildVersionError = fossApi.GoCreatePpmsBuildVersion(
				NewChangeRequestCreateBuildVersionInput(ppmsScv, ppmsBv, buildUrl, ppmsUsername, ppmsPassword))
		} /* else: all good - no build version given - update on scv only. */
	}

	result.AnalyseResult, result.AnalyseError = fossApi.GoAnalyze(analyzeInput)

	if updatePpms && result.AnalyseError == nil {
		analyzeMap, err := UnmarshalJsonObjStr(&result.AnalyseResult)
		if err == nil {
			result.PpmsUpdateResult, result.PpmsUpdateError = fossApi.GoSendPpmsChangeRequest(&ChangeRequestInput{
				ResultId: uuid.MustParse(fmt.Sprint(analyzeMap["resultId"])), User: ppmsUsername, Password: ppmsPassword})
		} else {
			result.PpmsUpdateError = errors.Errorf("Could not unmarshal analyse result: %v", err)
		}
	}

	return result.asJsonString()
}

// if callApi is given in config, the values are parsed to this struct
type ApiCall struct {
	method string
	params []string
}

type sapCallFossServiceConfig struct {
	options                *sapCallFossServiceOptions
	pipelineIdentification *PipelineIdentification
	scmIdentification      *ScmIdentification
	analyzeInput           *AnalysisInput
	apiCall                *ApiCall
	fossApi                FossApi
}

func readFossServiceIntegrationOptions(
	config *sapCallFossServiceOptions) (*sapCallFossServiceConfig, error) {

	var err error
	var apiCall *ApiCall
	var analyzeInput *AnalysisInput
	// in all cases: readPipelineIdentification
	var pipelineIdentification *PipelineIdentification
	var scmIdentification *ScmIdentification
	var instance Instance

	pipelineIdentification, err = config.readPipelineIdentification()
	if err != nil {
		return nil, err
	}

	scmIdentification, err = config.readScmIdentification()
	if err != nil {
		return nil, err
	}

	if config.CallAPI == "stepBehavior" {
		// In default, check if analyse request is given
		if strings.TrimSpace(config.AnalyzeRequestBody) == "" {
			return nil, errors.New(
				"if the default step behavior is executed, the analyze request body must be given")
		}

		//check analyze request
		analyzeInput, err = JsonObjectString(config.AnalyzeRequestBody).ReadAsAnalysisInput()
		if err != nil {
			return nil, err
		}

	} else {
		//No step behavior parse 'callApi' string
		apiCall, err = config.readCallApi()
		if err != nil {
			return nil, err
		}
	}

	// read optional Foss Service Base URL and construct.
	if len(config.FossBaseURL) > 0 {
		baseUrl, err := url.Parse(config.FossBaseURL)
		if err != nil || strings.TrimSpace(baseUrl.Host) == "" || strings.TrimSpace(baseUrl.Scheme) == "" {
			return nil, fmt.Errorf("no valid FOSS service base URL given. "+
				"At least host and scheme must be included. Details: %v", err)
		}

		instance = NewInstance(baseUrl)
	} else {
		return nil, errors.New("FOSS service base URL is missing.")
	}

	return &sapCallFossServiceConfig{
		options:                config,
		pipelineIdentification: pipelineIdentification,
		scmIdentification:      scmIdentification,
		analyzeInput:           analyzeInput,
		apiCall:                apiCall,
		fossApi:                &instance,
	}, nil
}

type FossDefaultStepResult struct {
	CreateBuildVersionResult JsonObjectString
	CreateBuildVersionError  error
	AnalyseResult            JsonObjectString
	AnalyseError             error
	PpmsUpdateResult         JsonObjectString
	PpmsUpdateError          error
}

type defaultStepResultMarshalContext struct {
	errorMessage    string
	needToSendError bool
	resultMap       map[string]interface{}
	errorMap        map[string]interface{}
}

func (r *FossDefaultStepResult) asJsonString() (JsonObjectString, error) {
	var ctxt = defaultStepResultMarshalContext{
		errorMessage:    "error in FOSS default behavior:",
		needToSendError: false,
		resultMap:       make(map[string]interface{}),
		errorMap:        make(map[string]interface{}),
	}

	r.handleCreateBuildVersion(&ctxt)
	r.handleAnalyse(&ctxt)
	r.handlePpmsUpdate(&ctxt)

	ctxt.resultMap["errors"] = ctxt.errorMap
	byteResult, err := json.Marshal(ctxt.resultMap)
	if err != nil {
		ctxt.needToSendError = true
		ctxt.errorMessage += " *** result marshaling error:"
		ctxt.errorMessage += " * " + err.Error()
	}

	var compoundError error = nil
	if ctxt.needToSendError {
		compoundError = errors.New(ctxt.errorMessage)
	}

	return JsonObjectString(byteResult), compoundError
}

func (r *FossDefaultStepResult) handleCreateBuildVersion(ctxt *defaultStepResultMarshalContext) {
	handleReqResultAndError(ctxt, r.CreateBuildVersionResult, r.CreateBuildVersionError,
		"createBuildVersion", "create build version")
}

func (r *FossDefaultStepResult) handleAnalyse(ctxt *defaultStepResultMarshalContext) {
	handleReqResultAndError(ctxt, r.AnalyseResult, r.AnalyseError,
		"analyse", "analyse")
}

func (r *FossDefaultStepResult) handlePpmsUpdate(ctxt *defaultStepResultMarshalContext) {
	handleReqResultAndError(ctxt, r.PpmsUpdateResult, r.PpmsUpdateError,
		"ppmsUpdate", "PPMS update")
}

func handleReqResultAndError(ctxt *defaultStepResultMarshalContext, jsonObjStr JsonObjectString, err error,
	resultName string, problemsName string) {

	ppmsUpdateMap, ppmsErr := UnmarshalJsonObjStr(&jsonObjStr)
	ctxt.resultMap[resultName+"Result"] = ppmsUpdateMap
	if err != nil || ppmsErr != nil {
		ctxt.needToSendError = true
		ctxt.errorMessage += " *** " + problemsName + " problems:"
		if err != nil {
			ctxt.errorMap[resultName] = err.Error()
			ctxt.errorMessage += " * " + err.Error()
		}
		if ppmsErr != nil {
			if jsonObjStr.Empty() {
				ctxt.errorMessage += " * " + ppmsErr.Error() + " because " + strings.Title(resultName) + "Result was not set."
			} else {
				ctxt.errorMessage += " * " + ppmsErr.Error() + " Response was: " + jsonObjStr.String()
			}
		}
	}
}

func canUpdatePpms(config *sapCallFossServiceConfig) bool {
	if strings.TrimSpace(config.options.Username) != "" &&
		strings.TrimSpace(config.options.Password) != "" &&
		len(config.analyzeInput.Target) == 1 {
		bomIdentifier := config.analyzeInput.Target[0]
		if bomIdentifier.Vendor == PPMS {
			return true
		}
	}

	return false
}

func executeApiCall(
	fossApi FossApi,
	apiCall *ApiCall) (JsonObjectString, error) {

	var api *FossRestEndpointDescriptor
	var result = EmptyBody
	var err error

	api, err = RestApiRegistry.GetApi(apiCall.method)
	if api != nil {
		if api.HttpMethod == "POST" {
			result, err = ReadResponseBody(fossApi.Call(api, JsonObjectString(apiCall.params[0]), EmptyQuery))
		} else {
			if api.MethodName == "getResultCertificate" {
				result, err = fossApi.GetResultCertificate(uuid.MustParse(apiCall.params[0]))
			}
			if api.MethodName == "compare" {
				result, err = fossApi.Compare(uuid.MustParse(apiCall.params[0]), uuid.MustParse(apiCall.params[1]))
			}
		}
	} else {
		err = errors.New("Could not found method " + apiCall.method)
	}

	if err != nil {
		return EmptyBody, errors.Errorf("could not successfully execute %v, Reason: %v", apiCall.method, err)
	} else {
		return result, nil
	}
}

func (config *sapCallFossServiceOptions) readPipelineIdentification() (*PipelineIdentification, error) {
	if strings.TrimSpace(config.PipelineIDentification) == "" {
		return nil, nil /*option not given*/
	}

	parts, err := TokenizeString(config.PipelineIDentification, ';', '\\')
	if err != nil {
		return nil, err
	}

	if len(parts) > 3 {
		return nil, errors.New("you can only give 3 parameters separated by semicolons for " +
			"pipeline identification (escape via backslash '\\')")
	}

	var uuid1 uuid.UUID
	uuid1, err = uuid.Parse(parts[0])
	if err != nil {
		return nil, errors.New("first parameter of pipeline identification must be a valid uuid")
	}

	var uuid2 uuid.UUID
	if len(parts) > 1 {
		uuid2, err = uuid.Parse(parts[1])
		if err != nil {
			return nil, errors.New("second parameter of pipeline identification must be a valid uuid")
		}
	}

	var backUrlString = ""
	if len(parts) > 2 {
		var backUrl *url.URL
		backUrl, err = url.Parse(parts[2])
		if err != nil || len(backUrl.Host) == 0 || len(backUrl.Scheme) == 0 || len(backUrl.Path) == 0 {
			return nil, errors.New(
				"third parameter of pipeline identification must be a valid URL (Host, Scheme " +
					"and Path must be present)")
		} else {
			backUrlString = backUrl.String()
		}
	}

	identification := PipelineIdentification{
		GlobalIdentifier: uuid1, InstanceIdentifier: uuid2, BackUrl: backUrlString}

	return &identification, nil
}

func (config *sapCallFossServiceOptions) readScmIdentification() (*ScmIdentification, error) {
	if strings.TrimSpace(config.ScmIDentification) == "" {
		return nil, nil /*option not given*/
	}

	parts, err := TokenizeString(config.ScmIDentification, ';', '\\')
	if err != nil {
		return nil, err
	}

	if len(parts) == 2 {
		scmUrl, err := url.Parse(parts[0])
		if err != nil || strings.TrimSpace(scmUrl.Host) == "" || strings.TrimSpace(scmUrl.Scheme) == "" {
			return nil, fmt.Errorf("no valid SCM URL given. "+
				"At least host and scheme must be included. Details: %v", err)
		}

		scmIdentification := ScmIdentification{ScmUrl: scmUrl.String(), Revision: parts[1]}

		return &scmIdentification, nil
	} else {
		return nil, fmt.Errorf(
			"scmIdentification need to have two parameters separated by semicolon. Given: %v", len(parts))
	}
}

// config.CallAPI must be set to stepBehavior if no other api should be called.
func (config *sapCallFossServiceOptions) readCallApi() (*ApiCall, error) {
	parts, err := TokenizeString(config.CallAPI, ';', '\\')
	if err != nil {
		return nil, err
	}

	apiCall := ApiCall{parts[0], parts[1:]}
	var endpoint *FossRestEndpointDescriptor
	endpoint, err = RestApiRegistry.GetApi(apiCall.method)
	if err != nil {
		return nil, err
	}

	if endpoint.NumOfParams != len(apiCall.params) {
		return nil, errors.New("For the method '" + endpoint.MethodName +
			"' you have to specify " + fmt.Sprintf("%v", endpoint.NumOfParams) +
			" parameter(s). Found: " + fmt.Sprintf("%v", len(apiCall.params)) + " (after method name)")
	}

	return &apiCall, nil
}

// Safely splits a string by a separator that is allowed to be escaped.
// Best way to do it in go (simple negative lookbehind is not supported in Go regex) which is
// (?<!a)b matches a “b” that is not preceded by an “a”
func TokenizeString(s string, sep, escape rune) (tokens []string, err error) {
	var runes []rune
	inEscape := false
	for _, r := range s {
		switch {
		case inEscape:
			inEscape = false
			runes = append(runes, r)
		case r == escape:
			inEscape = true
		case r == sep:
			tokens = append(tokens, string(runes))
			runes = runes[:0]
		default:
			runes = append(runes, r)
		}
	}
	tokens = append(tokens, string(runes))
	if inEscape {
		err = errors.New("invalid terminal escape")
	}
	return tokens, err
}

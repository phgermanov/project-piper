package fossService

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/ppms"
)

// A FOSS service (server) instance
type Instance struct {
	FossBaseUrl *url.URL
	ApiPath     string
	UiPath      string
	sender      piperhttp.Sender
}

// NewSomething create new instance of Something
func NewInstance(fossBaseUrl *url.URL) Instance {
	inst := Instance{}
	inst.FossBaseUrl = fossBaseUrl
	inst.ApiPath = "/api"
	inst.UiPath = "/ui"
	inst.sender = &piperhttp.Client{}
	inst.sender.SetOptions(piperhttp.ClientOptions{TransportTimeout: 3 * time.Minute})
	return inst
}

// The raw API uses raw json input and output as expected from the REST API interface
type RawApi interface {
	Analyze(request JsonObjectString) (JsonObjectString, error)
	Compare(source uuid.UUID, target uuid.UUID) (JsonObjectString, error)
	GetResultCertificate(resultId uuid.UUID) (JsonObjectString, error)
	SendPpmsChangeRequest(request JsonObjectString) (JsonObjectString, error)
	CreatePpmsBuildVersion(request JsonObjectString) (JsonObjectString, error)
}

func (fossServer *Instance) Analyze(request JsonObjectString) (JsonObjectString, error) {
	return ReadResponseBody(fossServer.Call(RestApiRegistry.analyze, request, EmptyQuery))
}

func (fossServer *Instance) Compare(source uuid.UUID, target uuid.UUID) (JsonObjectString, error) {
	return ReadResponseBody(fossServer.Call(RestApiRegistry.compare, EmptyBody,
		map[string]string{"source": source.String(), "target": target.String()}))
}

func (fossServer *Instance) GetResultCertificate(resultId uuid.UUID) (JsonObjectString, error) {
	return ReadResponseBody(fossServer.Call(RestApiRegistry.getResultCertificate, EmptyBody,
		map[string]string{"resultId": resultId.String()}))
}

func (fossServer *Instance) SendPpmsChangeRequest(request JsonObjectString) (JsonObjectString, error) {
	result, err := ReadResponseBody(fossServer.Call(RestApiRegistry.sendPpmsChangeRequest, request, EmptyQuery))
	if err != nil && strings.Contains(err.Error(), "returned with HTTP Code 428") {
		return "{\"info\" : \"All good, no update needed.\" }", nil
	} else {
		return result, err
	}
}

// implement this here as blocking behavior
func (fossServer *Instance) CreatePpmsBuildVersion(request JsonObjectString) (JsonObjectString, error) {
	return ReadResponseBody(fossServer.Call(RestApiRegistry.createPpmsBuildVersion, request, EmptyQuery))
}

// The full api currently supported in go
type FossApi interface {
	Analyze(request JsonObjectString) (JsonObjectString, error)
	Compare(source uuid.UUID, target uuid.UUID) (JsonObjectString, error)
	GetResultCertificate(resultId uuid.UUID) (JsonObjectString, error)
	SendPpmsChangeRequest(request JsonObjectString) (JsonObjectString, error)
	CreatePpmsBuildVersion(request JsonObjectString) (JsonObjectString, error)
	GoAnalyze(request *AnalysisInput) (JsonObjectString, error)
	GoSendPpmsChangeRequest(request *ChangeRequestInput) (JsonObjectString, error)
	GoCreatePpmsBuildVersion(request *ChangeRequestCreateBuildVersionInput) (JsonObjectString, error)
	Call(api *FossRestEndpointDescriptor, jsonBody JsonObjectString, queryParams map[string]string) (*http.Response, error)
}

func (fossServer *Instance) GoAnalyze(request *AnalysisInput) (JsonObjectString, error) {
	jsonString, err := MarshalToJsonObjStr(request)
	if err != nil {
		return EmptyBody, err
	}
	return fossServer.Analyze(jsonString)
}

func (fossServer *Instance) GoSendPpmsChangeRequest(request *ChangeRequestInput) (JsonObjectString, error) {
	jsonString, err := MarshalToJsonObjStr(request)
	if err != nil {
		return EmptyBody, err
	}
	return fossServer.SendPpmsChangeRequest(jsonString)
}

func (fossServer *Instance) GoCreatePpmsBuildVersion(request *ChangeRequestCreateBuildVersionInput) (JsonObjectString, error) {
	jsonString, err := MarshalToJsonObjStr(request)
	if err != nil {
		return EmptyBody, err
	}

	createBVResponse, err := fossServer.CreatePpmsBuildVersion(jsonString)
	if err != nil {
		return EmptyBody, err
	}

	return waitForPpmsBuildVersionApply(request, createBVResponse)
}

var EmptyQuery = map[string]string{}

const EmptyBody = JsonObjectString("")
const EmptyObject = JsonObjectString("{}")

func (fossServer *Instance) Call(api *FossRestEndpointDescriptor, jsonBody JsonObjectString,
	queryParams map[string]string) (*http.Response, error) {

	header := make(map[string][]string)
	header["Content-Type"] = []string{api.httpContent}
	header["Accept"] = []string{api.httpAccept}

	var requestUrl = *fossServer.FossBaseUrl
	requestUrl.Path = fossServer.ApiPath + "/" + api.MethodName
	query := requestUrl.Query()
	for k, v := range queryParams {
		query.Add(k, v)
	}
	requestUrl.RawQuery = query.Encode()

	fmt.Println(api.MethodName + ": " + jsonBody.String())

	var resp, err = fossServer.sender.SendRequest(api.HttpMethod, requestUrl.String(),
		bytes.NewBufferString(jsonBody.String()), header, nil)

	return resp, err
}

func ReadResponseBody(resp *http.Response, respError error) (JsonObjectString, error) {
	var bodyString = EmptyBody
	var bodyBytes []byte
	var err error
	if resp.Body != nil {
		bodyBytes, err = io.ReadAll(resp.Body)
		if err == nil {
			_ = resp.Body.Close()
			bodyString = JsonObjectString(string(bodyBytes))
			err = respError
		} else if respError != nil {
			err = fmt.Errorf("%v \n\tcaused by: %v", err, respError.Error())
		}
	}

	return bodyString, err
}

// The pipeline identification allows the Foss Service to identify previous results automatically.
// A pipeline identification is mandatory and at least the PipelineGUID must be given.
type PipelineIdentification struct {
	// mandatory
	// Identifies a pipeline (not its instance). Thus, multiple pipeline runs sends the same
	// global identifier.
	GlobalIdentifier uuid.UUID `json:"globalIdentifier"`
	// Identifies a pipeline instance (:= pipeline run). Typically this is always a new UUID.
	// The client can use this number to identify the calling run (usually this uuid is associated with a build number).
	InstanceIdentifier uuid.UUID `json:"instanceIdentifier"`
	// The client can use this URL to link to the calling foss run (Build URL).
	BackUrl string `json:"backUrl"`
}

func (r *PipelineIdentification) Validate() (bool, []error) {
	var problems []error
	if r.GlobalIdentifier == uuid.Nil {
		problems = append(problems, errors.New("A valid pipeline global identifier (pipelineGUID) must be given"))
	}

	if len(problems) > 0 {
		return false, problems
	} else {
		return true, nil
	}
}

type ScmIdentification struct {
	ScmUrl   string `json:"scmUrl"`
	Revision string `json:"revision"`
}

func (r *ScmIdentification) Validate() (bool, []error) {
	return true, nil
}

type AnalysisInput struct {
	ClientInformation ClientInformation `json:"meta"`
	Source            []BomIdentifier   `json:"source"`
	Target            []BomIdentifier   `json:"target"`
	Options           AnalysisOptions   `json:"options"`
}

type ClientInformation struct {
	PipelineIdentification PipelineIdentification `json:"pipelineIdentification"`
	ScmIdentification      ScmIdentification      `json:"scmIdentification"`
	PredecessorVia         string                 `json:"predecessorVia"`
	Tags                   []string               `json:"tags"`
}

type BomIdentifier struct {
	Vendor           vendor `json:"vendor"`
	WsProduct        string `json:"product"`
	WsOrHubProject   string `json:"project"`
	HubVersion       string `json:"version"`
	PpmsScv          string `json:"scv"`
	PpmsBuildVersion string `json:"buildVersion"`
}

type vendor string

const (
	WS   vendor = "ws"
	HUB  vendor = "hub"
	PPMS vendor = "ppms"
)

type AnalysisOptions struct {
	SkipCache          bool     `json:"skipCache"`
	CheckCompliance    bool     `json:"checkCompliance"`
	GoToMarketChannels []string `json:"goToMarketChannels"`
}

type ChangeRequestCreateBuildVersionInput struct {
	User          string                    `json:"user"`
	Password      string                    `json:"pwd"`
	ScvTarget     *NamedId                  `json:"scvTarget"`
	BuildVersion  *PpmsEntity               `json:"buildVersion"`
	PredecessorBv *NamedId                  `json:"predecessorBv"`
	Options       ppms.ChangeRequestOptions `json:"options"`
}

func NewChangeRequestCreateBuildVersionInput(
	ppmsScv string, ppmsBv string, buildUrl string,
	ppmsUsername string, ppmsPassword string) *ChangeRequestCreateBuildVersionInput {

	nCR := ChangeRequestCreateBuildVersionInput{}

	nCR.User = ppmsUsername
	nCR.Password = ppmsPassword
	nCR.ScvTarget = &NamedId{Name: ppmsScv}
	nCR.BuildVersion = &PpmsEntity{
		Name:        ppmsBv,
		Description: buildUrl,
	}
	nCR.Options = ppms.ChangeRequestOptions{
		CopyPredecessorFoss: true,
		CopyPredecessorCvBv: true,
	}

	return &nCR
}

type PpmsEntity struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type NamedId struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type ChangeRequestInput struct {
	ResultId uuid.UUID `json:"resultId,omitempty"`
	User     string    `json:"user,omitempty"`
	Password string    `json:"pwd,omitempty"`
}

// A string formatted as JSON starting as JSON object
type JsonObjectString string

func (js JsonObjectString) String() string {
	return string(js)
}

func (js JsonObjectString) Empty() bool {
	return strings.TrimSpace(js.String()) == ""
}

func UnmarshalJsonObjStr(js *JsonObjectString) (map[string]interface{}, error) {
	jMap := map[string]interface{}{}
	var err error
	if js != nil {
		err = json.Unmarshal([]byte(js.String()), &jMap)
	}
	return jMap, err
}

func MarshalToJsonObjStr(input interface{}) (JsonObjectString, error) {
	b, err := json.Marshal(input)
	return JsonObjectString(string(b)), err
}

func (js JsonObjectString) ReadAsAnalysisInput() (*AnalysisInput, error) {
	data := AnalysisInput{}
	err := json.Unmarshal([]byte(js.String()), &data)
	if err != nil {
		return nil, err
	} else {
		return &data, nil
	}
}

// Description for the existing RestApiRegistry with their most significant http connection properties.
type FossRestEndpointDescriptor struct {
	MethodName  string
	HttpMethod  string
	httpContent string
	httpAccept  string
	NumOfParams int
}

var RestApiRegistry = newFossRestApiRegistry()

func newFossRestApiRegistry() *fossRestApiRegistry {

	const contentTypeJson = "application/json"
	const httpAcceptJson = contentTypeJson
	const httpGet, httpPost = "GET", "POST"

	analyze := &FossRestEndpointDescriptor{
		"analyze", httpPost, contentTypeJson, httpAcceptJson, 1}
	compare := &FossRestEndpointDescriptor{
		"compare", httpGet, "*/*", httpAcceptJson, 2}
	getResultCertificate := &FossRestEndpointDescriptor{
		"getResultCertificate", httpGet, "*/*", httpAcceptJson, 1}
	sendPpmsChangeRequest := &FossRestEndpointDescriptor{
		"sendPpmsChangeRequest", httpPost, "application/json", httpAcceptJson, 1}
	createPpmsBuildVersion := &FossRestEndpointDescriptor{
		"createPpmsBuildVersion", httpPost, contentTypeJson, httpAcceptJson, 1}
	methods := []FossRestEndpointDescriptor{*analyze, *compare, *getResultCertificate, *sendPpmsChangeRequest, *createPpmsBuildVersion}

	return &fossRestApiRegistry{
		analyze:                analyze,
		compare:                compare,
		getResultCertificate:   getResultCertificate,
		sendPpmsChangeRequest:  sendPpmsChangeRequest,
		createPpmsBuildVersion: createPpmsBuildVersion,
		methods:                methods}
}

func (reg *fossRestApiRegistry) GetApi(methodName string) (*FossRestEndpointDescriptor, error) {
	for _, m := range reg.methods {
		if m.MethodName == methodName {
			return &m, nil
		}
	}

	return nil, errors.New("Foss API method '" + methodName + "' is not known")
}

type fossRestApiRegistry struct {
	analyze                *FossRestEndpointDescriptor
	compare                *FossRestEndpointDescriptor
	getResultCertificate   *FossRestEndpointDescriptor
	sendPpmsChangeRequest  *FossRestEndpointDescriptor
	createPpmsBuildVersion *FossRestEndpointDescriptor
	methods                []FossRestEndpointDescriptor
}

const xCsrfHeader = "x-csrf-token"

func waitForPpmsBuildVersionApply(request *ChangeRequestCreateBuildVersionInput, createBVResponse JsonObjectString) (JsonObjectString, error) {

	crl, err := getChangeRequestUrl(createBVResponse)
	if err != nil {
		return EmptyBody, err
	}

	sender := &piperhttp.Client{}
	sender.SetOptions(piperhttp.ClientOptions{TransportTimeout: 10 * time.Second, Username: request.User, Password: request.Password})

	header, cookies, err := requestCsrfTokenAndCookies(err, sender, crl)
	if err != nil {
		return EmptyBody, err
	}

	var numOfTries, checkErrors = 0, make([]error, 0)
	for {
		time.Sleep(calcWaitDuration(numOfTries)) //waits 5s, 10s, 15s, 20s, 5s, 10s 15s, 20s and then gives up
		numOfTries += 1

		checkResp, crErr := ReadResponseBody(sender.SendRequest("GET", crl.String(), nil, header, cookies))
		if crErr != nil {
			checkErrors = append(checkErrors, errors.Wrap(crErr, "Attempt: "+fmt.Sprint(numOfTries)+
				" Problem calling "+crl.String()+" to check status of build version creation"))
		} else {
			checkMap, unmErr := UnmarshalJsonObjStr(&checkResp)
			if unmErr != nil {
				checkErrors = append(checkErrors, errors.Wrap(crErr, "Attempt: "+fmt.Sprint(numOfTries)+
					" Problem unmarshal response of calling"+crl.String()+" to check status of build version creation."+
					" Response was: "+checkResp.String()))
			}

			if getBuildVersionApplied(checkMap) {
				return createBVResponse, nil
			}
		}

		err := checkForDeterminationAndCreateError(checkErrors, numOfTries)
		if err != nil {
			return EmptyBody, err
		}
	}
}

func checkForDeterminationAndCreateError(checkErrors []error, numOfTries int) error {
	if len(checkErrors) > 3 {
		return errors.Errorf("creation of build version failed. Details: [%v]", checkErrors)
	}

	if numOfTries > 8 {
		return errors.Errorf("creation of build takes too long (Timeout): Sub-errors: [%v]", checkErrors)
	}

	return nil
}

func getBuildVersionApplied(checkMap map[string]interface{}) bool {
	status := fmt.Sprint(checkMap["status"])
	return status == "APPLIED"
}

func calcWaitDuration(numOfTries int) time.Duration {
	return time.Duration(5*((numOfTries%4)+1)) * time.Second
}

func requestCsrfTokenAndCookies(err error, sender *piperhttp.Client, crl *url.URL) (http.Header, []*http.Cookie, error) {
	//request token
	header := http.Header{}
	header.Add(xCsrfHeader, "Fetch")
	response, err := sender.SendRequest("HEAD",
		crl.Scheme+"://"+crl.Host+"/sap/internal/ppms/api/changerequest/v1/cvpart", nil, header, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to retrieve PPMS CSRF token")
	}
	_ = response.Body.Close()
	token := response.Header.Get(xCsrfHeader)
	cookies := response.Cookies()
	header.Set(xCsrfHeader, token)

	return header, cookies, nil
}

func getChangeRequestUrl(createBVResponse JsonObjectString) (*url.URL, error) {
	createBVMap, unmErr := UnmarshalJsonObjStr(&createBVResponse)
	if unmErr != nil {
		return nil, unmErr
	} else {
		return MustURL(fmt.Sprint(createBVMap["changeRequestLocation"])), nil
	}
}

// Ignores the url creation parsing error
// Should only be used in fixed and known to work declarations
func MustURL(rawUrl string) *url.URL {
	parsedUrl, _ := url.Parse(rawUrl)
	return parsedUrl
}

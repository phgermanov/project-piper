//go:build unit
// +build unit

package eccn

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

type eccnMockClient struct {
	username       string
	password       string
	httpMethod     string
	httpStatusCode int
	urlCalled      string
	requestBody    io.Reader
	responseBody   string
}

func (c *eccnMockClient) SetOptions(opts piperhttp.ClientOptions) {
	c.username = opts.Username
	c.password = opts.Password
}

func (c *eccnMockClient) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	c.httpMethod = method
	c.urlCalled = url
	return &http.Response{StatusCode: c.httpStatusCode, Body: io.NopCloser(bytes.NewReader([]byte(c.responseBody)))}, nil
}

func TestGetECCNData(t *testing.T) {

	t.Run("Message status S and details true", func(t *testing.T) {
		myTestClient := eccnMockClient{responseBody: `{"DATA":{"ONR":"testonr","ONAME":"testoname","OTYPE":"testPV","MESSAGE":"Questionnaire for PV - HANA AS A SERVICE 1.0 already exists with EU classification 'Checked: no classif.'!","MESSAGE_STATUS":"S","QUESTIONNIARE_LINK":"https://ifp.bss.net.sap/sap/bc/webdynpro/sap/zec_main?sap-language=EN&iv_onr=73555000100900002516&iv_process_id=","ALL_QUESTIONNIARES_LINK":"https://ifp.bss.net.sap/sap/bc/webdynpro/sap/zec_main?sap-language=EN&iv_onr=73555000100900002516&iv_process_id=&iv_reporting_type=AGGREGATION&iv_otype=PV"}}`}
		add_details := []additionaldetails{{Type: "S", Text: "Total number of comprised components: 1"}, {Type: "S", Text: "Total number of comprised components: 1"}}
		sys := System{ServerURL: "https://ifp.bss.net.sap", HTTPClient: &myTestClient}
		expected := eccndata{Onr: "testonr", Oname: "testoname", Otype: "testPV", Message: "Questionnaire for PV - HANA AS A SERVICE 1.0 already exists with EU classification 'Checked: no classif.'!", MessageStatus: "S", QuestionnaireLink: "https://ifp.bss.net.sap/sap/bc/webdynpro/sap/zec_main?sap-language=EN&iv_onr=73555000100900002516&iv_process_id=", AllQuestionnairesLink: "https://ifp.bss.net.sap/sap/bc/webdynpro/sap/zec_main?sap-language=EN&iv_onr=73555000100900002516&iv_process_id=&iv_reporting_type=AGGREGATION&iv_otype=PV", Additionaldetails: add_details}

		eccn, err := sys.GetECCNData("73555000100900002516", true)

		assert.NoError(t, err, "Error occurred")

		assert.Equal(t, "https://ifp.bss.net.sap/eccn_piper/onr/73555000100900002516?%24format=json&details=true", myTestClient.urlCalled, "Called url incorrect")

		assert.Equal(t, expected.Onr, eccn.Data.Onr, "Retrieved  object number incorrect")
		assert.Equal(t, expected.Otype, eccn.Data.Otype, "Retrieved object type classification incorrect")
		assert.Equal(t, expected.Oname, eccn.Data.Oname, "Retrieved object name classification incorrect")
		assert.Equal(t, expected.Message, eccn.Data.Message, "Retrieved message  incorrect")
		assert.Equal(t, expected.MessageStatus, eccn.Data.MessageStatus, "Retrieved message status  incorrect")
		assert.Equal(t, expected.QuestionnaireLink, eccn.Data.QuestionnaireLink, "Retrieved questionnire link  incorrect")
		assert.Equal(t, expected.AllQuestionnairesLink, eccn.Data.AllQuestionnairesLink, "Retrieved all questionnire link   incorrect")

	})
	t.Run("Message status S", func(t *testing.T) {
		myTestClient := eccnMockClient{responseBody: `{"DATA":{"ONR":"testonr","ONAME":"testoname","OTYPE":"testPV","MESSAGE":"Questionnaire for PV - HANA AS A SERVICE 1.0 already exists with EU classification 'Checked: no classif.'!","MESSAGE_STATUS":"S","QUESTIONNIARE_LINK":"https://ifp.bss.net.sap/sap/bc/webdynpro/sap/zec_main?sap-language=EN&iv_onr=73555000100900002516&iv_process_id=","ALL_QUESTIONNIARES_LINK":"https://ifp.bss.net.sap/sap/bc/webdynpro/sap/zec_main?sap-language=EN&iv_onr=73555000100900002516&iv_process_id=&iv_reporting_type=AGGREGATION&iv_otype=PV"}}`}

		sys := System{ServerURL: "https://ifp.bss.net.sap", HTTPClient: &myTestClient}
		expected := eccndata{Onr: "testonr", Oname: "testoname", Otype: "testPV", Message: "Questionnaire for PV - HANA AS A SERVICE 1.0 already exists with EU classification 'Checked: no classif.'!", MessageStatus: "S", QuestionnaireLink: "https://ifp.bss.net.sap/sap/bc/webdynpro/sap/zec_main?sap-language=EN&iv_onr=73555000100900002516&iv_process_id=", AllQuestionnairesLink: "https://ifp.bss.net.sap/sap/bc/webdynpro/sap/zec_main?sap-language=EN&iv_onr=73555000100900002516&iv_process_id=&iv_reporting_type=AGGREGATION&iv_otype=PV"}
		eccn, err := sys.GetECCNData("73555000100900002516", true)

		assert.NoError(t, err, "Error occurred")

		assert.Equal(t, "https://ifp.bss.net.sap/eccn_piper/onr/73555000100900002516?%24format=json&details=true", myTestClient.urlCalled, "Called url incorrect")

		assert.Equal(t, expected.Onr, eccn.Data.Onr, "Retrieved  object number incorrect")
		assert.Equal(t, expected.Otype, eccn.Data.Otype, "Retrieved object type classification incorrect")
		assert.Equal(t, expected.Oname, eccn.Data.Oname, "Retrieved object name classification incorrect")
		assert.Equal(t, expected.Message, eccn.Data.Message, "Retrieved message  incorrect")
		assert.Equal(t, expected.MessageStatus, eccn.Data.MessageStatus, "Retrieved message status  incorrect")
		assert.Equal(t, expected.QuestionnaireLink, eccn.Data.QuestionnaireLink, "Retrieved questionnire link  incorrect")
		assert.Equal(t, expected.AllQuestionnairesLink, eccn.Data.AllQuestionnairesLink, "Retrieved all questionnire link   incorrect")

	})
	t.Run("Message status W", func(t *testing.T) {
		myTestClient := eccnMockClient{responseBody: `{"DATA":{"ONR":"73554900100200013923","ONAME":"AI BUS DOCINFO EXTR OD 1.0","OTYPE":"CV","MESSAGE":"Questionnaire for CV - AI BUS DOCINFO EXTR OD 1.0 already exists with EU classification 'Checked: no classif.' in ISP, but the questionnaire is in reclassification!","MESSAGE_STATUS":"W","QUESTIONNIARE_LINK":"https://ifp.bss.net.sap/sap/bc/webdynpro/sap/zec_main?sap-language=EN&iv_onr=73554900100200013923&iv_process_id=","ALL_QUESTIONNIARES_LINK":"https://ifp.bss.net.sap/sap/bc/webdynpro/sap/zec_main?sap-language=EN&iv_onr=73554900100200013923&iv_process_id=&iv_reporting_type=AGGREGATION&iv_otype=CV"}}`}
		//	myTestClient := eccnMockClient{responseBody: `{"d":{"Id":"1","Name":"TestSCV","TechnicalName":"TechSCV","TechnicalRelease":"1","ReviewModelRiskRatings":{"__deferred":{"uri":"https://my.Uri"}},"BuildVersions":{"__deferred":{"uri":"https://my.Uri"}},"FreeOpenSourceSoftwares":{"__deferred":{"uri":"https://my.Uri"}},"Responsibles":{"__deferred":{"uri":"https://my.Uri"}}}}`}

		sys := System{ServerURL: "https://ifp.bss.net.sap", HTTPClient: &myTestClient}
		expected := eccndata{Onr: "73554900100200013923", Oname: "AI BUS DOCINFO EXTR OD 1.0", Otype: "CV", Message: "Questionnaire for CV - AI BUS DOCINFO EXTR OD 1.0 already exists with EU classification 'Checked: no classif.' in ISP, but the questionnaire is in reclassification!", MessageStatus: "W", QuestionnaireLink: "https://ifp.bss.net.sap/sap/bc/webdynpro/sap/zec_main?sap-language=EN&iv_onr=73554900100200013923&iv_process_id=", AllQuestionnairesLink: "https://ifp.bss.net.sap/sap/bc/webdynpro/sap/zec_main?sap-language=EN&iv_onr=73554900100200013923&iv_process_id=&iv_reporting_type=AGGREGATION&iv_otype=CV"}
		eccn, err := sys.GetECCNData("73554900100200013923", true)

		assert.NoError(t, err, "Error occurred")
		assert.Equal(t, "https://ifp.bss.net.sap/eccn_piper/onr/73554900100200013923?%24format=json&details=true", myTestClient.urlCalled, "Called url incorrect")

		assert.Equal(t, expected.Onr, eccn.Data.Onr, "Retrieved  object number incorrect")
		assert.Equal(t, expected.Otype, eccn.Data.Otype, "Retrieved object type classification incorrect")
		assert.Equal(t, expected.Oname, eccn.Data.Oname, "Retrieved object name classification incorrect")
		assert.Equal(t, expected.Message, eccn.Data.Message, "Retrieved message  incorrect")
		assert.Equal(t, expected.MessageStatus, eccn.Data.MessageStatus, "Retrieved message status  incorrect")
		assert.Equal(t, expected.QuestionnaireLink, eccn.Data.QuestionnaireLink, "Retrieved questionnire link  incorrect")
		assert.Equal(t, expected.AllQuestionnairesLink, eccn.Data.AllQuestionnairesLink, "Retrieved all questionnire link   incorrect")

	})

	t.Run("Message status E", func(t *testing.T) {
		myTestClient := eccnMockClient{responseBody: `{"DATA":{"ONR":"73554900100200013883","ONAME":"CTNR ELM CONNECTOR 1.0","OTYPE":"CV","MESSAGE":"Questionnaire for CV - CTNR ELM CONNECTOR 1.0 is not yet classified","MESSAGE_STATUS":"E","QUESTIONNIARE_LINK":"https://ifp.bss.net.sap/sap/bc/webdynpro/sap/zec_main?sap-language=EN&iv_onr=73554900100200013883&iv_process_id=","ALL_QUESTIONNIARES_LINK":"https://ifp.bss.net.sap/sap/bc/webdynpro/sap/zec_main?sap-language=EN&iv_onr=73554900100200013883&iv_process_id=&iv_reporting_type=AGGREGATION&iv_otype=CV"}}`}
		//	myTestClient := eccnMockClient{responseBody: `{"d":{"Id":"1","Name":"TestSCV","TechnicalName":"TechSCV","TechnicalRelease":"1","ReviewModelRiskRatings":{"__deferred":{"uri":"https://my.Uri"}},"BuildVersions":{"__deferred":{"uri":"https://my.Uri"}},"FreeOpenSourceSoftwares":{"__deferred":{"uri":"https://my.Uri"}},"Responsibles":{"__deferred":{"uri":"https://my.Uri"}}}}`}

		sys := System{ServerURL: "https://ifp.bss.net.sap", HTTPClient: &myTestClient}
		expected := eccndata{Onr: "73554900100200013883", Oname: "CTNR ELM CONNECTOR 1.0", Otype: "CV", Message: "Questionnaire for CV - CTNR ELM CONNECTOR 1.0 is not yet classified", MessageStatus: "E", QuestionnaireLink: "https://ifp.bss.net.sap/sap/bc/webdynpro/sap/zec_main?sap-language=EN&iv_onr=73554900100200013883&iv_process_id=", AllQuestionnairesLink: "https://ifp.bss.net.sap/sap/bc/webdynpro/sap/zec_main?sap-language=EN&iv_onr=73554900100200013883&iv_process_id=&iv_reporting_type=AGGREGATION&iv_otype=CV"}
		eccn, err := sys.GetECCNData("73554900100200013883", true)

		assert.EqualError(t, err, "An error occurred: "+eccn.Data.Message)

		assert.Equal(t, "https://ifp.bss.net.sap/eccn_piper/onr/73554900100200013883?%24format=json&details=true", myTestClient.urlCalled, "Called url incorrect")

		assert.Equal(t, expected.Onr, eccn.Data.Onr, "Retrieved  object number incorrect")
		assert.Equal(t, expected.Otype, eccn.Data.Otype, "Retrieved object type classification incorrect")
		assert.Equal(t, expected.Oname, eccn.Data.Oname, "Retrieved object name classification incorrect")
		assert.Equal(t, expected.Message, eccn.Data.Message, "Retrieved message  incorrect")
		assert.Equal(t, expected.MessageStatus, eccn.Data.MessageStatus, "Retrieved message status  incorrect")
		assert.Equal(t, expected.QuestionnaireLink, eccn.Data.QuestionnaireLink, "Retrieved questionnire link  incorrect")
		assert.Equal(t, expected.AllQuestionnairesLink, eccn.Data.AllQuestionnairesLink, "Retrieved all questionnire link   incorrect")

	})

	t.Run("Message status empty", func(t *testing.T) {
		myTestClient := eccnMockClient{responseBody: `{"DATA":{"ONR":"","ONAME":"","OTYPE":"","MESSAGE":"","MESSAGE_STATUS":"","QUESTIONNIARE_LINK":"","ALL_QUESTIONNIARES_LINK":""}}`}
		//	myTestClient := eccnMockClient{responseBody: `{"d":{"Id":"1","Name":"TestSCV","TechnicalName":"TechSCV","TechnicalRelease":"1","ReviewModelRiskRatings":{"__deferred":{"uri":"https://my.Uri"}},"BuildVersions":{"__deferred":{"uri":"https://my.Uri"}},"FreeOpenSourceSoftwares":{"__deferred":{"uri":"https://my.Uri"}},"Responsibles":{"__deferred":{"uri":"https://my.Uri"}}}}`}

		sys := System{ServerURL: "https://ifp.bss.net.sap", HTTPClient: &myTestClient}
		expected := eccndata{Onr: "", Oname: "", Otype: "", Message: "", MessageStatus: "", QuestionnaireLink: "", AllQuestionnairesLink: ""}
		eccn, err := sys.GetECCNData("01200615320900001212", true)

		assert.EqualError(t, err, "An empty message status was returned.")

		assert.Equal(t, "https://ifp.bss.net.sap/eccn_piper/onr/01200615320900001212?%24format=json&details=true", myTestClient.urlCalled, "Called url incorrect")

		assert.Equal(t, expected.Onr, eccn.Data.Onr, "Retrieved  object number incorrect")
		assert.Equal(t, expected.Otype, eccn.Data.Otype, "Retrieved object type classification incorrect")
		assert.Equal(t, expected.Oname, eccn.Data.Oname, "Retrieved object name classification incorrect")
		assert.Equal(t, expected.Message, eccn.Data.Message, "Retrieved message  incorrect")
		assert.Equal(t, expected.MessageStatus, eccn.Data.MessageStatus, "Retrieved message status  incorrect")
		assert.Equal(t, expected.QuestionnaireLink, eccn.Data.QuestionnaireLink, "Retrieved questionnire link  incorrect")
		assert.Equal(t, expected.AllQuestionnairesLink, eccn.Data.AllQuestionnairesLink, "Retrieved all questionnire link   incorrect")

	})

}

func TestSendRequest(t *testing.T) {
	myTestClient := eccnMockClient{responseBody: "OK"}
	sys := System{ServerURL: "https://ifp.bss.net.sap", Username: "test", Password: "test", HTTPClient: &myTestClient}

	response, err := sys.sendRequest("GET", "73555000100900002516", true, bytes.NewReader([]byte("Test")), nil)

	assert.NoError(t, err, "Error occurred but none expected")

	assert.Equal(t, sys.Username, myTestClient.username, "Username not passed correctly")
	assert.Equal(t, sys.Password, myTestClient.password, "Password not passed correctly")
	assert.Equal(t, "GET", myTestClient.httpMethod, "Wrong HTTP method used")
	assert.Equal(t, "https://ifp.bss.net.sap/eccn_piper/onr/73555000100900002516?%24format=json&details=true", myTestClient.urlCalled, "Called url incorrect")

	assert.Equal(t, []byte("OK"), response, "Response incorrect")

}

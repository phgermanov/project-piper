package eccn

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
)

// System defines access information to ifp system
type System struct {
	ServerURL  string
	Username   string
	Password   string
	HTTPClient piperhttp.Sender
}

type eccndata struct {
	Onr                   string              `json:"ONR,omitempty"`
	Oname                 string              `json:"ONAME,omitempty"`
	Otype                 string              `json:"OTYPE,omitempty"`
	Message               string              `json:"MESSAGE,omitempty"`
	MessageStatus         string              `json:"MESSAGE_STATUS,omitempty"`
	QuestionnaireLink     string              `json:"QUESTIONNIARE_LINK"`
	AllQuestionnairesLink string              `json:"ALL_QUESTIONNIARES_LINK"`
	Additionaldetails     []additionaldetails `json:"ADDITIONAL_DETAILS"`
}

type data struct {
	Data eccndata `json:"DATA"`
}

type additionaldetails struct {
	Type string `json:"TYPE,omitempty"`
	Text string `json:"TEXT,omitempty"`
}

const ifpEndpoint = "/eccn_piper/onr/"

// GetECCNData returns details of a dedicated ECCN system
func (sys *System) GetECCNData(ppmsID string, eccnDetails bool) (data, error) {

	var eccnResponseObject data

	content, err := sys.sendRequest(http.MethodGet, ppmsID, eccnDetails, nil, nil)
	if err != nil {
		return eccnResponseObject, err
	}

	if len(content) > 0 {
		err = json.Unmarshal(content, &eccnResponseObject)
		if err != nil {
			return eccnResponseObject, errors.Wrap(err, "unmarshalling of eccn response object failed")
		}
		switch eccnResponseObject.Data.MessageStatus {
		case "S":
			// success
			fallthrough
		case "W":
			// warning
			fmt.Println("---------------------------Begin of ECCN related messages--------------------")
			log.Entry().Info(eccnResponseObject.Data.Message)
			log.Entry().Info("Questionnare link:", eccnResponseObject.Data.QuestionnaireLink)
			if eccnDetails == true {
				log.Entry().Info("Additional Information:")
				fmt.Println("     |     ", "-------------------------------Details on Comprised components: ------------------------------- ")
				for i := range eccnResponseObject.Data.Additionaldetails {
					fmt.Println("     |     ", eccnResponseObject.Data.Additionaldetails[i].Text)
				}
				fmt.Println("---------------------------End of ECCN related messages--------------------")
			}

		case "E":
			// error
			return eccnResponseObject, errors.Errorf("An error occurred: %v", eccnResponseObject.Data.Message)

		case "":
			// empty
			return eccnResponseObject, errors.Errorf("An empty message status was returned.")
		default:
			// unknown
			return eccnResponseObject, errors.Errorf("An unknown message status was returned: %v", eccnResponseObject.Data.MessageStatus)
		}
	} else {
		return eccnResponseObject, errors.Errorf("An empty message was returned from the backend system.")
	}
	return eccnResponseObject, err

}

func (sys *System) sendRequest(method, ppmsID string, eccnDetails bool, body io.Reader, header http.Header) ([]byte, error) {
	opts := piperhttp.ClientOptions{
		Username: sys.Username,
		Password: sys.Password,
		Logger:   log.Entry(),
	}
	sys.HTTPClient.SetOptions(opts)

	sysURL, err := url.Parse(sys.ServerURL)
	if err != nil {
		log.Entry().Warnf("failed to parse serverUrl as http string")
	}
	sysURL.Path = path.Join(sysURL.Path, ifpEndpoint, ppmsID)
	sysURL.ForceQuery = true

	requestParameter := url.Values{}
	if eccnDetails == true {
		requestParameter.Set("details", "true")
	}
	requestParameter.Set("$format", "json")

	httpPath := sysURL.String() + requestParameter.Encode()

	response, err := sys.HTTPClient.SendRequest(method, httpPath, body, nil, nil)
	if err != nil {
		return []byte{}, errors.Wrapf(err, "failed to send ECCN request to %v", httpPath)
	}
	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return []byte{}, errors.Wrap(err, "error reading response from ECCN request")
	}
	return content, nil
}

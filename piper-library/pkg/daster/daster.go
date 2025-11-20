package daster

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
	"github.com/SAP/jenkins-library/pkg/log"
)

type Daster interface {
	FetchOAuthToken(oAuthServiceURL, oAuthGrantType,
		oAuthSource, clientID, clientSecret string, verbose bool) (string, error)
	TriggerScan(request map[string]interface{}) (string, error)
	GetScan(scanId string) (*Scan, error)
	DeleteScan(scanId string) error
}

type Scan struct {
	Terminated bool
	Results    string
	Summary    *Violations
	ExitCode   int
	Reason     string
}

type Violations struct {
	High   int
	Medium int
	Low    int
	Info   int
	All    int
}

type OAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
}

var RetryCodes = map[int]bool{
	100: true, 101: true, 102: true, 103: true, 404: true, 408: true, 425: true,
	500: true, 502: true, 503: true, 504: true, // not really common but a DASTer specific issue
}

func fetchOAuthToken(client piperhttp.Sender, oAuthServiceUrl, oAuthGrandType, oAuthSource, clientId, clientSecret string,
	verbose bool, retryDelay time.Duration) (string, error) {
	body := strings.NewReader(url.Values{
		"grant_type":    []string{oAuthGrandType},
		"scope":         []string{oAuthSource},
		"client_id":     []string{clientId},
		"client_secret": []string{clientSecret},
	}.Encode())
	header := http.Header{}
	header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := callAPI(client, oAuthServiceUrl, http.MethodPost, body, header, verbose, 1, retryDelay)
	if err != nil {
		return "", err
	}

	var result OAuthTokenResponse
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return "", err
	}
	return result.AccessToken, nil
}

func callScanAPI(client piperhttp.Sender, url, mode string, requestBody interface{}, verbose bool, maxRetries int, retryDelay time.Duration) ([]byte, error) {
	body, err := getRequestBody(requestBody)
	if err != nil {
		return nil, err
	}
	header := http.Header{}
	header.Set("Content-Type", "application/json")
	header.Set("Accept", "application/json")

	resp, err := callAPI(client, url, mode, bytes.NewBuffer(body), header, verbose, maxRetries, retryDelay)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func callAPI(client piperhttp.Sender, url, mode string, body io.Reader, header http.Header, verbose bool, maxRetries int, retryDelay time.Duration) ([]byte, error) {
	for attempts := 0; attempts < maxRetries; attempts++ {
		if verbose {
			log.Entry().Infof("Sending HTTP request: attempt %d/%d", attempts+1, maxRetries)
		}
		response, err := client.SendRequest(mode, url, body, header, nil)
		if err != nil {
			return nil, err
		}

		respBody, err := readResponse(response, verbose)
		if err != nil {
			return nil, err
		}

		if response.StatusCode == http.StatusOK {
			return respBody, nil
		}

		if RetryCodes[response.StatusCode] {
			log.Entry().Warnf("received retry code %d: %s", response.StatusCode, string(respBody))
			time.Sleep(retryDelay)
			continue
		}

		return nil, fmt.Errorf("error %d from %s %s: %s", response.StatusCode, mode, url, string(respBody))
	}
	return nil, fmt.Errorf("timeout error occurred")
}

func getRequestBody(body interface{}) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return requestBody, nil
}

func readResponse(response *http.Response, verbose bool) ([]byte, error) {
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Entry().WithError(err).Error("failed reading response")
		return nil, err
	}
	if verbose {
		log.Entry().Infof("received response with code %d: %s", response.StatusCode, string(body))
	}
	return body, nil
}

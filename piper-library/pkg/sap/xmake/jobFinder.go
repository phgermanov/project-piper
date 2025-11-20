package xmake

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/SAP/jenkins-library/pkg/log"
)

const serviceURL = "https://xmake-nova.wdf.sap.corp/"
const serviceEndpoint = "job_finder/api/json"

// JobFinder ...
type JobFinder struct {
	Client        Sender
	Username      string
	Token         string
	RetryInterval time.Duration
}

// Sender provides an interface to the piper http client for uid/pwd and token authenticated requests
type Sender interface {
	Send(*http.Request) (*http.Response, error)
}

// JobMatches ...
type JobMatches struct {
	Class   string `json:"_class"`
	JobList []Job  `json:"job"`
}

// Job ...
type Job struct {
	Name       string `json:"name"`
	FullName   string `json:"fullName"`
	URL        string `json:"url"`
	Landscape  string `json:"landscape"`
	JenkinsURL string `json:"jenkinsUrl"`
	Branch     string `json:"branch"`
}

func (finder *JobFinder) Lookup(jobName string) ([]Job, error) {

	log.Entry().Infof("Looking up job in JobFinder: '%s'", jobName)
	requestURL := finder.getRequestURL(jobName)

	var lastError error
	for tryCount := 1; tryCount <= 5; tryCount++ {
		response, err := finder.Send(requestURL)
		if err != nil {
			logConnectError(tryCount, requestURL, err)
			lastError = handleFailedResponse(response, err)
			time.Sleep(finder.RetryInterval)
		} else {
			log.Entry().Infof("Successfully retrieved xmake job URL from JobFinder: '%s'", requestURL)
			content, respErr := handleSuccessfulResponse(response)
			return content, respErr
		}
	}
	return []Job{}, logFinalError(lastError)
}

func logConnectError(tryCount int, requestURL string, err error) {
	log.Entry().Warningf("(Try %d) Unable to retrieve xmake job URL from JobFinder: '%s' - Error: '%s'", tryCount, requestURL, err)
}

func handleSuccessfulResponse(response *http.Response) ([]Job, error) {
	contentJSON, err := io.ReadAll(response.Body)
	if err != nil {
		return []Job{}, errors.Wrap(err, "error reading response body")
	}

	var content JobMatches
	if err := json.Unmarshal(contentJSON, &content); err != nil {
		return []Job{}, errors.Wrap(err, "error unmarshalling response")
	}

	return content.JobList, nil
}

func handleFailedResponse(response *http.Response, connectionError error) error {
	var errMsg = ""
	if response == nil {
		errMsg = fmt.Sprintf("No response received from the server or connection was dropped. Error: %v", connectionError)
		log.Entry().Error(errMsg)
		return errors.New(errMsg)
	}
	statusCode := response.StatusCode

	switch {
	case statusCode >= 500:
		errMsg = fmt.Sprintf("Jenkins server error: service unavailable. HTTP Status: %d. Error: %v", statusCode, connectionError)
	case statusCode >= 400:
		errMsg = fmt.Sprintf("Jenkins client error: possible incorrect credentials or bad request. HTTP Status: %d. Error: %v", statusCode, connectionError)
	}

	return errors.New(errMsg)
}

func logFinalError(lastError error) error {
	errMsg := fmt.Sprintf("JobFinder finally failed after all retries: %v", lastError)
	log.Entry().Error(errMsg)
	return errors.New(errMsg)
}

func (finder *JobFinder) getRequestURL(jobName string) string {
	request, _ := url.Parse(serviceURL)
	request.Path = serviceEndpoint
	request.RawQuery = url.Values{"input": {jobName}}.Encode()
	return request.String()
}

func (finder *JobFinder) Send(requestURL string) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodGet, requestURL, nil)
	request.SetBasicAuth(finder.Username, finder.Token)
	log.Entry().Infof("Sending request to JobFinder with user '%s'", finder.Username)
	if err != nil {
		return nil, errors.Wrap(err, "error creating request")
	}
	return finder.Client.Send(request)
}

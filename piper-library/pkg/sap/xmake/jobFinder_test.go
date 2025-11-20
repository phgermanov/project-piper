package xmake

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
)

func TestGetRequestURL(t *testing.T) {
	//TODO: test encoding
	// init
	jobName := "Piper-Validation-Golang-SP-MS-common"
	jobFinder := JobFinder{}
	// test
	result := jobFinder.getRequestURL(jobName)
	// asserts
	assert.Equal(t, "https://xmake-nova.wdf.sap.corp/job_finder/api/json?input="+jobName, result)
}

func TestLookup(t *testing.T) {
	log.SetVerbose(true)
	jobName := "Piper-Validation-Golang-SP-MS-common"
	targetURL := "https://xmake-nova.wdf.sap.corp/job_finder/api/json?input=" + jobName

	t.Run("success", func(t *testing.T) {
		// init
		client := &piperhttp.Client{}
		// required to use httpmock (DefaultTransport)
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		jobFinder := JobFinder{Client: client, RetryInterval: time.Nanosecond}
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, targetURL,
			httpmock.NewStringResponder(http.StatusOK, `{
		        "_class":"com.sap.prd.jenkins.plugins.jobfinder.JobFinder",
		        "job":[{
		            "name": "mock.Anything",
		            "fullName":"piper-validation/piper-validation-golang-SP-MS-common",
		            "url": "mock.Anything",
		            "landscape":"xmake-dev",
		            "jenkinsUrl":"https://xmake-dev.wdf.sap.corp"
		        }]
		    }`),
		)
		// test
		jobList, err := jobFinder.Lookup(jobName)
		// asserts
		assert.Equal(t, 1, httpmock.GetTotalCallCount())
		assert.NoError(t, err)
		assert.NotNil(t, jobList)
		assert.NotEmpty(t, jobList)
		assert.Equal(t, 1, len(jobList))
		for _, job := range jobList {
			assert.Equal(t, mock.Anything, job.Name)
			assert.Equal(t, mock.Anything, job.URL)
		}
	})

	t.Run("success - empty result", func(t *testing.T) {
		// init
		client := &piperhttp.Client{}
		// required to use httpmock (DefaultTransport)
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		jobFinder := JobFinder{Client: client, RetryInterval: time.Nanosecond}
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, targetURL,
			httpmock.NewStringResponder(http.StatusOK, `{}`),
		)
		// test
		jobList, err := jobFinder.Lookup(jobName)
		// asserts
		assert.NoError(t, err)
		assert.Empty(t, jobList)
	})

	t.Run("success - empty job list", func(t *testing.T) {
		// init
		client := &piperhttp.Client{}
		// required to use httpmock (DefaultTransport)
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		jobFinder := JobFinder{Client: client, RetryInterval: time.Nanosecond}
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, targetURL,
			httpmock.NewStringResponder(http.StatusOK, `{
				"_class": "com.sap.prd.jenkins.plugins.jobfinder.JobFinder",
				"job": []
			}`),
		)
		// test
		jobList, err := jobFinder.Lookup(jobName)
		// asserts
		assert.NoError(t, err)
		assert.Empty(t, jobList)
	})

	t.Run("error - invalid JSON", func(t *testing.T) {
		// init
		client := &piperhttp.Client{}
		// required to use httpmock (DefaultTransport)
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		jobFinder := JobFinder{Client: client, RetryInterval: time.Nanosecond}
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, targetURL,
			httpmock.NewStringResponder(http.StatusOK, ``),
		)
		// test
		jobList, err := jobFinder.Lookup(jobName)
		// asserts
		assert.EqualError(t, err, "error unmarshalling response: unexpected end of JSON input")
		assert.Empty(t, jobList)
	})

	t.Run("error - jenkins connection retries", func(t *testing.T) {
		// init
		client := &piperhttp.Client{}
		// required to use httpmock (DefaultTransport)
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
		jobFinder := JobFinder{Client: client, RetryInterval: time.Nanosecond}
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, targetURL,
			httpmock.NewErrorResponder(errors.New("EOF")),
		)
		// test
		jobList, err := jobFinder.Lookup(jobName)
		// asserts
		assert.Contains(t, err.Error(), "No response received from the server or connection was dropped.")
		assert.Equal(t, jobList, []Job{})
	})
}

func TestLookupHTTPErrorHandling(t *testing.T) {
	jobName := "Piper-Validation-Golang-SP-MS-common"
	targetURL := "https://xmake-nova.wdf.sap.corp/job_finder/api/json?input=" + jobName
	client := &piperhttp.Client{}
	client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})
	jobFinder := JobFinder{Client: client, RetryInterval: time.Nanosecond}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	t.Run("error - server error (HTTP 500)", func(t *testing.T) {
		httpmock.RegisterResponder(http.MethodGet, targetURL,
			httpmock.NewStringResponder(http.StatusInternalServerError, ""),
		)
		jobList, err := jobFinder.Lookup(jobName)
		assert.NotNil(t, jobList)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Jenkins server error: service unavailable. HTTP Status: 500")
	})

	t.Run("error - client error (HTTP 403)", func(t *testing.T) {
		httpmock.RegisterResponder(http.MethodGet, targetURL,
			httpmock.NewStringResponder(http.StatusForbidden, ""),
		)
		jobList, err := jobFinder.Lookup(jobName)
		assert.NotNil(t, jobList)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Jenkins client error: possible incorrect credentials or bad request. HTTP Status: 403")
	})
}

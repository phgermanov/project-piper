package daster

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewFioriDASTScan(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		url := "http://localhost:8080"
		verbose := false
		maxRetries := 1
		retryDelay := time.Second
		daster := NewFioriDASTScan(url, verbose, maxRetries, retryDelay)
		assert.NotNil(t, daster)
		assert.NotEmpty(t, daster)
		assert.Equal(t, "http://localhost:8080/fioriDASTScan", daster.url)
	})
}

func TestPopulateScan(t *testing.T) {
	t.Parallel()
	t.Run("Scan terminated, has risk summary", func(t *testing.T) {
		scanResponse := GetFioriDASTScanResponse{
			Results: "results",
			State: struct {
				Terminated *ScanState `json:"terminated"`
			}{
				Terminated: &ScanState{
					ExitCode: 1,
					Reason:   "Completed",
				},
			},
			RiskSummary: &RiskSummary{
				High:   5,
				Medium: 10,
				Low:    20,
			},
		}

		scan := populateScan(scanResponse)
		assert.NotNil(t, scan)
		assert.True(t, scan.Terminated)
		assert.Equal(t, 1, scan.ExitCode)
		assert.Equal(t, "Completed", scan.Reason)
		assert.Equal(t, "results", scan.Results)
		assert.NotNil(t, scan.Summary)
		assert.Equal(t, 5, scan.Summary.High)
		assert.Equal(t, 10, scan.Summary.Medium)
		assert.Equal(t, 20, scan.Summary.Low)
	})

	t.Run("Scan terminated with error", func(t *testing.T) {
		scanResponse := GetFioriDASTScanResponse{
			Results: "results",
			State: struct {
				Terminated *ScanState `json:"terminated"`
			}{
				Terminated: &ScanState{
					ExitCode: 1,
					Reason:   "Error",
				},
			},
		}

		scan := populateScan(scanResponse)
		assert.NotNil(t, scan)
		assert.True(t, scan.Terminated)
		assert.Equal(t, 1, scan.ExitCode)
		assert.Equal(t, "Error", scan.Reason)
		assert.Equal(t, "results", scan.Results)
		assert.Nil(t, scan.Summary)
	})

	t.Run("Scan is not terminated", func(t *testing.T) {
		scanResponse := GetFioriDASTScanResponse{
			Results: "",
			State: struct {
				Terminated *ScanState `json:"terminated"`
			}{},
		}

		scan := populateScan(scanResponse)
		assert.NotNil(t, scan)
		assert.False(t, scan.Terminated)
		assert.Equal(t, 0, scan.ExitCode)
		assert.Equal(t, "", scan.Reason)
		assert.Equal(t, "", scan.Results)
		assert.Nil(t, scan.Summary)
	})
}

func TestFioriFetchOAuthToken(t *testing.T) {
	url := "http://localhost:8080"
	verbose := false
	maxRetries := 1
	retryDelay := time.Second
	daster := NewFioriDASTScan(url, verbose, maxRetries, retryDelay)
	t.Run("ok", func(t *testing.T) {
		client := newClientMock(0, false, `{"access_token": "ok"}`)
		daster.client = client
		token, err := daster.FetchOAuthToken("http://oauth.service", "client_credentials", "resource", "client_id", "secret", false)
		assert.NoError(t, err)
		assert.Equal(t, "ok", token)
	})
	t.Run("api error", func(t *testing.T) {
		client := newClientMock(0, true, "")
		daster.client = client
		_, err := daster.FetchOAuthToken("http://oauth.service", "client_credentials", "resource", "client_id", "secret", false)
		assert.Error(t, err)
	})
	t.Run("unmarshalling error", func(t *testing.T) {
		client := newClientMock(0, true, "token")
		daster.client = client
		_, err := daster.FetchOAuthToken("http://oauth.service", "client_credentials", "resource", "client_id", "secret", false)
		assert.Error(t, err)
	})
}

func TestFioriTriggerScan(t *testing.T) {
	url := "http://localhost:8080"
	verbose := false
	maxRetries := 1
	retryDelay := time.Second
	daster := NewFioriDASTScan(url, verbose, maxRetries, retryDelay)
	t.Run("ok", func(t *testing.T) {
		client := newClientMock(0, false, `{"scanId": "id"}`)
		daster.client = client
		scanId, err := daster.TriggerScan(map[string]interface{}{"token": "token"})
		assert.NoError(t, err)
		assert.Equal(t, "id", scanId)
	})
	t.Run("api error", func(t *testing.T) {
		client := newClientMock(0, true, ``)
		daster.client = client
		_, err := daster.TriggerScan(map[string]interface{}{"token": "token"})
		assert.Error(t, err)
	})
	t.Run("unmarshalling error", func(t *testing.T) {
		client := newClientMock(0, false, `id`)
		daster.client = client
		_, err := daster.TriggerScan(map[string]interface{}{"token": "token"})
		assert.Error(t, err)
	})
}

func TestFioriGetScan(t *testing.T) {
	url := "http://localhost:8080"
	verbose := false
	maxRetries := 1
	retryDelay := time.Second
	daster := NewFioriDASTScan(url, verbose, maxRetries, retryDelay)
	t.Run("ok, scan is terminated", func(t *testing.T) {
		client := newClientMock(0, false, `{
			"results": "results",
			"riskSummary": {
				"high": 5,
				"medium": 10
			},
			"state": {
				"terminated": {
					"containerID": "container-id",
					"exitCode": 0,
					"reason": "Success"
				}
			}
		}`)
		daster.client = client
		scan, err := daster.GetScan("id")
		assert.NoError(t, err)
		assert.NotNil(t, scan)
		assert.True(t, scan.Terminated)
		assert.Equal(t, 0, scan.ExitCode)
		assert.Equal(t, "Success", scan.Reason)
		assert.NotNil(t, scan.Summary)
		assert.Equal(t, 5, scan.Summary.High)
		assert.Equal(t, 10, scan.Summary.Medium)
		assert.Equal(t, 0, scan.Summary.Low)
		assert.Equal(t, 0, scan.Summary.Info)
	})
	t.Run("ok, scan is pending", func(t *testing.T) {
		client := newClientMock(0, false, `{
			"results": "results",
			"state": {
				"pending": {
					"containerID": "container-id"
				}
			}
		}`)
		daster.client = client
		scan, err := daster.GetScan("id")
		assert.NoError(t, err)
		assert.NotNil(t, scan)
		assert.False(t, scan.Terminated)
		assert.Equal(t, 0, scan.ExitCode)
		assert.Equal(t, "", scan.Reason)
		assert.Nil(t, scan.Summary)
	})
	t.Run("scan is terminated with error", func(t *testing.T) {
		client := newClientMock(0, false, `{
			"results": "results",
			"state": {
				"terminated": {
					"containerID": "container-id",
					"exitCode": 1,
					"reason": "Error"
				}
			}
		}`)
		daster.client = client
		scan, err := daster.GetScan("id")
		assert.NoError(t, err)
		assert.NotNil(t, scan)
		assert.True(t, scan.Terminated)
		assert.Equal(t, 1, scan.ExitCode)
		assert.Equal(t, "Error", scan.Reason)
		assert.Nil(t, scan.Summary)
	})
	t.Run("api error", func(t *testing.T) {
		client := newClientMock(0, true, ``)
		daster.client = client
		scan, err := daster.GetScan("id")
		assert.Error(t, err)
		assert.Nil(t, scan)
	})
	t.Run("unmarshalling error", func(t *testing.T) {
		client := newClientMock(0, false, `text`)
		daster.client = client
		scan, err := daster.GetScan("id")
		assert.Error(t, err)
		assert.Nil(t, scan)
	})
}

func TestFioriDeleteScan(t *testing.T) {
	url := "http://localhost:8080"
	verbose := false
	maxRetries := 1
	retryDelay := time.Second
	daster := NewFioriDASTScan(url, verbose, maxRetries, retryDelay)
	t.Run("ok", func(t *testing.T) {
		client := newClientMock(0, false, `ok`)
		daster.client = client
		err := daster.DeleteScan("id")
		assert.NoError(t, err)
	})
	t.Run("api error", func(t *testing.T) {
		client := newClientMock(0, true, ``)
		daster.client = client
		err := daster.DeleteScan("id")
		assert.Error(t, err)
	})
}

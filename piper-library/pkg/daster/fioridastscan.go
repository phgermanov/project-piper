package daster

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
)

var FioriDASTScanType = "fioriDASTScan"

type FioriDASTScan struct {
	client     piperhttp.Sender
	url        string
	verbose    bool
	maxRetries int
	retryDelay time.Duration
}

type NewFioriDASTScanResponse struct {
	ScanId string `json:"scanId"`
}

type GetFioriDASTScanResponse struct {
	Results     string       `json:"results"`
	RiskSummary *RiskSummary `json:"riskSummary"`
	State       struct {
		Terminated *ScanState `json:"terminated"`
	} `json:"state"`
}

type ScanState struct {
	ContainerID string    `json:"containerID"`
	ExitCode    int       `json:"exitCode"`
	FinishedAt  time.Time `json:"finishedAt"`
	Reason      string    `json:"reason"`
	StartedAt   time.Time `json:"startedAt"`
}

type RiskSummary struct {
	High          int `json:"high"`
	Medium        int `json:"medium"`
	Low           int `json:"low"`
	Informational int `json:"informational"`
}

func NewFioriDASTScan(url string, verbose bool, maxRetries int, retryDelay time.Duration) *FioriDASTScan {
	return &FioriDASTScan{
		verbose:    verbose,
		maxRetries: maxRetries,
		url:        fmt.Sprintf("%s/%s", strings.TrimSuffix(url, "/"), FioriDASTScanType),
		client:     &piperhttp.Client{},
		retryDelay: retryDelay,
	}
}

func (d *FioriDASTScan) FetchOAuthToken(oAuthServiceURL, oAuthGrantType,
	oAuthSource, clientID, clientSecret string, verbose bool) (string, error) {
	return fetchOAuthToken(d.client, oAuthServiceURL, oAuthGrantType, oAuthSource,
		clientID, clientSecret, verbose, d.retryDelay)
}

func (d *FioriDASTScan) TriggerScan(request map[string]interface{}) (string, error) {
	resp, err := callScanAPI(d.client, d.url, http.MethodPost, request, d.verbose, d.maxRetries, d.retryDelay)
	if err != nil {
		return "", err
	}

	var scan NewFioriDASTScanResponse
	err = json.Unmarshal(resp, &scan)
	if err != nil {
		return "", err
	}
	return scan.ScanId, nil
}

func (d *FioriDASTScan) GetScan(scanId string) (*Scan, error) {
	resp, err := callScanAPI(d.client, fmt.Sprintf("%s/%s", d.url, scanId), http.MethodGet, nil, d.verbose, d.maxRetries, d.retryDelay)
	if err != nil {
		return nil, err
	}

	var scanResponse GetFioriDASTScanResponse
	err = json.Unmarshal(resp, &scanResponse)
	if err != nil {
		return nil, err
	}

	return populateScan(scanResponse), nil
}

func populateScan(scanResponse GetFioriDASTScanResponse) *Scan {
	scan := &Scan{
		Results:    scanResponse.Results,
		Terminated: scanResponse.State.Terminated != nil,
	}
	if scan.Terminated {
		scan.ExitCode = scanResponse.State.Terminated.ExitCode
		scan.Reason = scanResponse.State.Terminated.Reason
	}
	if scanResponse.RiskSummary != nil {
		scan.Summary = &Violations{
			High:   scanResponse.RiskSummary.High,
			Medium: scanResponse.RiskSummary.Medium,
			Low:    scanResponse.RiskSummary.Low,
		}
	}
	return scan
}

func (d *FioriDASTScan) DeleteScan(scanId string) error {
	_, err := callScanAPI(d.client, fmt.Sprintf("%s/%s", d.url, scanId), http.MethodDelete, nil, d.verbose, d.maxRetries, d.retryDelay)
	if err != nil {
		return err
	}
	return nil
}

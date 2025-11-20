package reportpipelinestatus

import (
	"bufio"
	"encoding/json"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
)

// ParseLogFile parses error Messages and stepDetails out  of a logFile
func ParseLogFile(logFile *[]byte) ([]map[string]interface{}, string, []map[string]interface{}) {
	if logFile == nil || len(*logFile) == 0 {
		log.Entry().Debug("received empty logFile, can not parse it. Returning with empty error and step list.")
		return []map[string]interface{}{}, "FAILURE", []map[string]interface{}{}
	}
	stringify := string(*logFile)

	maxCapacity := len(*logFile) + 1 // len in bytes
	buf := make([]byte, maxCapacity)
	scanner := bufio.NewScanner(strings.NewReader(stringify))
	scanner.Buffer(buf, maxCapacity)

	var (
		stepDetailsMap  []map[string]interface{}
		errorDetailsMap []map[string]interface{}
		stepDetails     []StepDetails
		errorDetails    []ErrorDetail
	)

	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, "Step telemetry data:") {
			matchStepInfo(text, &stepDetails)
		}
	}

	// Prepare step details in a map string interface
	marshal, err := json.Marshal(stepDetails)
	if err != nil {
		log.Entry().WithError(err).Error("could not marshal determined Steps")
	}
	err = json.Unmarshal(marshal, &stepDetailsMap)
	if err != nil {
		log.Entry().WithError(err).Error("could not unmarshal stepDetailsMap")
	}

	// Get error details in dedicated map string interface and cleanup errorDetails where none exist
	pipelineStatus := orchestrator.BuildStatusSuccess
	for _, stepDetail := range stepDetailsMap {
		if stepDetail["ErrorCode"] == "1" {
			// indicates that at least one error is present
			pipelineStatus = orchestrator.BuildStatusFailure
		}
		if stepDetail["ErrorCode"] == "0" {
			// No error we set ErrorDetail to null
			stepDetail["ErrorDetail"] = nil
		}
	}
	// Get error details in dedicated map string interface and cleanup errorDetails where none exist
	for _, stepDetail := range stepDetails {
		if stepDetail.ErrorDetail.Message != "" && stepDetail.ErrorDetail.Error != "" {
			errorDetails = append(errorDetails, stepDetail.ErrorDetail)
		}
	}

	// Prepare errorDetails inf a map string interface
	marshal, err = json.Marshal(errorDetails)
	if err != nil {
		log.Entry().WithError(err).Error("could not marshal errorDetails")
	}
	err = json.Unmarshal(marshal, &errorDetailsMap)
	if err != nil {
		log.Entry().WithError(err).Error("could not unmarshal errorDetailsMap")
	}
	return errorDetailsMap, pipelineStatus, stepDetailsMap
}

// SendLogs determines if logs should be sent based on the errorCategories present
func SendLogs(sendErrorLogs []string, details []map[string]interface{}) bool {
	// Figure out if there has been a failure
	if len(details) > 0 {
		log.Entry().Debugf("Found the following errorDetails: %v", details)
	} else {
		log.Entry().Debugf("No errors found, setting sendLogs=false.")
		return false
	}

	var errorCategories []string
	for _, errorDetail := range details {
		errorCategories = append(errorCategories, errorDetail["Category"].(string))
	}

	sendLogs := false
	if CategoryInList(errorCategories, sendErrorLogs) {
		sendLogs = true
		log.Entry().Debugf("Error category %s is in list of categories to send logs. Sending logs for these categories: %v. Setting sendLogs=true.", errorCategories, sendErrorLogs)
	} else {
		log.Entry().Debugf("Error category %s is not in list of categories to send logs. Sending logs for these categories: %v . Setting sendLogs=false.", errorCategories, sendErrorLogs)
		sendLogs = false
	}
	return sendLogs
}

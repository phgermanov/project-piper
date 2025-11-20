package reportpipelinestatus

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
)

// const regexFatalError = `(?is)fatal error: errorDetails\{correlationId:\"(?P<correlationId>.*?)\",stepName:\"(?P<stepName>.*?)\",category:\"(?P<category>.*?)\",error:\"(?P<error>.*?)\",result:\"(?P<result>.*?)\",message:\"(?P<message>.*?)\"`
// regexFatalError defines a regex which matches any key value pair with the following format
// "Key":"Value"
const regexFatalError = `(?i)\"(\w+)":\"(.*?)\"`

// ErrorDetail struct holds information about errors of the step
type ErrorDetail struct {
	Message       string
	Error         string
	Category      string
	Result        string
	CorrelationId string
	StepName      string
	Time          string
}

func matchError(line string, errorDetails *[]ErrorDetail) {
	if strings.Contains(line, "fatal error: errorDetails") {
		res := getMatches(line, regexFatalError)
		if res != nil && len(res) != 0 {
			errorDetail := ErrorDetail{}
			resJson, err := json.Marshal(res)
			if err != nil {
				log.Entry().WithError(err).Error("could not marshal errorDetail matches")
			}
			err = json.Unmarshal(resJson, &errorDetail)
			if err != nil {
				log.Entry().WithError(err).Error("could not unmarshal errorDetail matches")
			}
			log.Entry().Debugf("found an error from the logs: %v", errorDetail)
			*errorDetails = append(*errorDetails, errorDetail)
		}
	}
}

// getErrorsJson reads errorDetails.json files from the CPE and returns an ErrorDetail struct.
func getErrorsJson() ([]ErrorDetail, error) {
	fileName := "errorDetails.json"
	path, err := os.Getwd()
	if err != nil {
		log.Entry().Error("can not get current working dir")
		return []ErrorDetail{}, err
	}
	// TODO: read cpe using cpe pkg
	pathCPE := path + "/.pipeline/commonPipelineEnvironment"
	matches, err := filepath.Glob(pathCPE + "/*" + fileName)
	if err != nil {
		log.Entry().Error("could not search filepath for *errorDetails.json files")
		return []ErrorDetail{}, err
	}
	if len(matches) == 0 {
		log.Entry().Debug("no errors in CPE found, returning empty errorDetails")
		return []ErrorDetail{}, nil
	}
	log.Entry().Debugf("found the following errorDetails files: %v", matches)

	var errorDetails []ErrorDetail
	log.Entry().Debugf("Found %v files", matches)

	for _, v := range matches {
		errorDetail, err := readErrorJson(v)
		if err != nil {
			log.Entry().Errorf("could not read error details for file %v", v)
			errorDetail = ErrorDetail{}
		}
		errorDetails = append(errorDetails, errorDetail)

	}
	return errorDetails, nil
}

func readErrorJson(filePath string) (ErrorDetail, error) {
	errorDetails := ErrorDetail{}
	jsonFile, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		log.Entry().Errorf("could not read file from path: %v", filePath)
		return ErrorDetail{}, err
	}
	err = json.Unmarshal(jsonFile, &errorDetails)
	if err != nil {
		log.Entry().Error("could not unmarshal error details")
		return ErrorDetail{}, err
	}
	return errorDetails, nil
}

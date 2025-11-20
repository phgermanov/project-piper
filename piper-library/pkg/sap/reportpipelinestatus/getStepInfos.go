package reportpipelinestatus

import (
	"encoding/json"
	"github.com/SAP/jenkins-library/pkg/log"
	"reflect"
	"regexp"
)

// StepDetails struct holds information about errors of the step
type StepDetails struct {
	StepStartTime   string      `json:"StepStartTime,omitempty"`
	PipelineUrlHash string      `json:"PipelineUrlHash"`
	BuildUrlHash    string      `json:"BuildUrlHash"`
	StageName       string      `json:"StageName"`
	StepName        string      `json:"StepName"`
	ErrorCode       string      `json:"ErrorCode"`
	Duration        string      `json:"Duration,omitempty"`     // omitempty to have backwards compatibility, remove in future releases
	StepDuration    string      `json:"StepDuration,omitempty"` // omitempty to have backwards compatibility, remove in future releases
	ErrorCategory   string      `json:"ErrorCategory"`
	ErrorDetail     ErrorDetail `json:"ErrorDetail"`
	CorrelationID   string      `json:"CorrelationID"`
	CommitHash      string      `json:"CommitHash,omitempty"`
	PiperCommitHash string      `json:"PiperCommitHash,omitempty"` // omitempty to have backwards compatibility, remove in future releases
	Branch          string      `json:"Branch,omitempty"`          // omitempty to have backwards compatibility, remove in future releases
	GitOwner        string      `json:"GitOwner,omitempty"`        // omitempty to have backwards compatibility, remove in future releases
	GitRepository   string      `json:"GitRepository,omitempty"`   // omitempty to have backwards compatibility, remove in future releases
}

func matchStepInfo(line string, stepDetails *[]StepDetails) {
	re := regexp.MustCompile(`(?m){.*}`) // regex to match {} encapsulated step data
	slice := re.Find([]byte(line))
	stepDetail := StepDetails{}
	err := json.Unmarshal(slice, &stepDetail)
	if err != nil {
		log.Entry().WithError(err).Error("could not unmarshal stepDetail")
	}
	for _, v := range *stepDetails {
		//check to remove duplicate steps as currently the logfile creates has duplicate log entries
		if reflect.DeepEqual(v, stepDetail) {
			log.Entry().Debugf("duplicate step entry found for %s", stepDetail.StepName)
			return
		}
	}

	*stepDetails = append(*stepDetails, stepDetail)
}

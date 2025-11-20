package reportpipelinestatus

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
)

// CategoryInList evaluates if elements of categories are in the list
func CategoryInList(categories []string, list []string) bool {
	for _, i := range categories {
		for _, j := range list {
			if i == j {
				return true
			}
		}
	}
	return false
}

func getMatches(text string, regex string) map[string]interface{} {
	re := regexp.MustCompile(regex)
	matches := re.FindAllStringSubmatch(text, -1)
	if matches == nil {
		return map[string]interface{}{}
	}
	result := map[string]interface{}{}
	for _, match := range matches {
		result[match[1]] = match[2]
	}
	return result
}

// WriteLogFile writes a byte object to disk
func WriteLogFile(logFile *[]byte, fileName string) {
	// persist logfile in workspace for further consumption e.g. cumulus upload
	// if we encounter an error we log it, but it does not have an effect on the step.
	f, err := os.Create(filepath.Clean(fileName))
	if err != nil {
		log.Entry().Errorf(err.Error(), "could not create file %v", fileName)
	}
	_, err = f.Write(*logFile)
	if err != nil {
		log.Entry().Errorf(err.Error(), "could not write data to logFile %v", fileName)
	}
	log.Entry().Debugf("Successfully persisted logfile to %v", fileName)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Entry().WithError(err).Error("could not close written log file")
		}
	}(f)
}

// PipelineTelemetry object to store pipeline related telemetry information
type PipelineTelemetry struct {
	CorrelationId           string                   `json:"CorrelationId"`       // CorrelationId
	Duration                string                   `json:"PipelineDuration"`    // Duration of the pipeline in milliseconds
	Orchestrator            string                   `json:"Orchestrator"`        // Orchestrator, e.g. Jenkins or Azure
	OrchestratorVersion     string                   `json:"OrchestratorVersion"` // OrchestratorVersion
	PipelineStartTime       string                   `json:"PipelineStartTime"`   // PipelineStartTime Pipeline start time
	BuildId                 string                   `json:"BuildId"`             // BuildId of the pipeline run
	JobName                 string                   `json:"JobName"`
	PipelineStatus          string                   `json:"PipelineStatus"`
	GitInstance             string                   `json:"GitInstance"`
	JobURL                  string                   `json:"JobURL"`
	CommitHash              string                   `json:"CommitHash"`
	ChangeSet               []orchestrator.ChangeSet `json:"ChangeSet"` // ChangeSet list of Commit hashes and timestamps
	Branch                  string                   `json:"Branch"`
	GitOrganization         string                   `json:"GitOrganization"`
	GitRepository           string                   `json:"GitRepository"`
	GitUrl                  string                   `json:"GitURL"`
	StepData                []map[string]interface{} `json:"StepData"`
	ArtifactVersion         string                   `json:"ArtifactVersion"`
	CumulusPipelineId       string                   `json:"CumulusPipelineId"`
	BuildReason             string                   `json:"BuildReason"`
	ScheduledBuild          bool                     `json:"ScheduledBuild"`
	IsScheduledAndOptimized bool                     `json:"IsScheduledAndOptimized"`
	IsProductiveBranch      bool                     `json:"IsProductiveBranch"`
	OrchestratorCredentials bool                     `json:"OrchestratorCredentials"`
}

// CalcDuration Calculates the time from X to now and returns it as a string in milliseconds
func CalcDuration(pipelineTime time.Time) string {
	timeSince := time.Since(pipelineTime)
	timeSinceMilliSeconds := timeSince.Milliseconds()
	stringify := strconv.FormatInt(timeSinceMilliSeconds, 10)
	return stringify
}

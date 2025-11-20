package cmd

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"context"

	"github.com/SAP/jenkins-library/pkg/gcs"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	piperversioning "github.com/SAP/jenkins-library/pkg/versioning"
	"github.com/bmatcuk/doublestar"
	"github.com/pkg/errors"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/events"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/cumulus"
	"google.golang.org/api/googleapi"
)

const bucketLockedErrString = `Object '.*?' is under active Temporary hold and cannot be deleted, overwritten or archived until hold is removed.`

func sapCumulusUpload(config sapCumulusUploadOptions, _telemetryData *telemetry.CustomData, pipelineEnv *sapCumulusUploadCommonPipelineEnvironment) {
	coordinates := piperversioning.Coordinates{
		Version: config.Version,
	}
	_, projectVersion := piperversioning.DetermineProjectCoordinates("", config.VersioningModel, coordinates)

	sapCumulus := cumulus.Cumulus{
		EnvVars: []cumulus.EnvVar{
			{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: config.JSONKeyFilePath, Modified: false},
		},
		FilePattern:           config.FilePattern,
		PipelineID:            config.PipelineID,
		Revision:              config.Revision,
		Scheduled:             config.Scheduled,
		SchedulingTimestamp:   config.SchedulingTimestamp,
		StepResultType:        config.StepResultType,
		SubFolderPath:         config.SubFolderPath,
		Version:               projectVersion,
		HeadCommitID:          config.GitHeadCommitID,
		UseCommitIDForCumulus: config.UseCommitIDForCumulus,
		BucketLockedWarning:   config.BucketLockedWarning,
		EventData:             config.EventData,
	}

	ctx := context.Background()
	client, err := gcs.NewClient(config.JSONKeyFilePath, config.Token)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to initialize GCS client")
		return
	}

	if err = runSapCumulusUpload(ctx, pipelineEnv, client, &sapCumulus, doublestar.Glob, os.Stat); err != nil {
		log.Entry().WithError(err).Fatal("Failed to perform Cumulus upload.")
		return
	}
	pipelineEnv.custom.schedulingTimestamp = sapCumulus.SchedulingTimestamp
}

func runSapCumulusUpload(ctx context.Context, pipelineEnv *sapCumulusUploadCommonPipelineEnvironment, gcsClient gcs.Client, sapCumulus *cumulus.Cumulus, searchFn func(string) (matches []string, err error), fileInfo func(string) (os.FileInfo, error)) error {
	if err := sapCumulus.ValidateInput(); err != nil {
		return err
	}

	sapCumulus.PrepareEnv()
	defer sapCumulus.CleanupEnv()

	pipelineRunKey, err := sapCumulus.GetPipelineRunKey()
	if err != nil {
		return errors.Wrap(err, "failed to determine pipeline run key")
	}
	pipelineEnv.custom.cumulusPipelineRunKey = pipelineRunKey
	if err = exportPipelineRunKey(pipelineEnv, sapCumulus); err != nil {
		log.Entry().Debugf("pipeline run key couldn't be exported: %s", err)
	}

	targetPath := sapCumulus.GetCumulusPath(pipelineRunKey)

	log.Entry().Debugf("target path is %v", targetPath)

	path, err := os.Getwd()
	if err != nil {
		log.Entry().Infof("error in getting current working directory")
	}
	log.Entry().Infof("looking for %v in %v", sapCumulus.FilePattern, path)
	matches, err := searchFiles(sapCumulus.FilePattern, searchFn, fileInfo)
	if err != nil {
		return errors.Wrapf(err, "open source failed: %v", err)
	}
	log.Entry().Infof("uploading %v file(s) into %v", len(matches), targetPath)

	if err = uploadFiles(ctx, gcsClient, sapCumulus, &matches, targetPath); err != nil {
		return errors.Wrapf(err, "upload failed: %v", err)
	}

	log.Entry().Infoln("cumulus upload was successful")
	return nil
}

func searchFiles(patternStr string, searchFn func(string) (matches []string, err error), fileInfo func(string) (os.FileInfo, error)) ([]cumulus.Task, error) {
	keys := make(map[cumulus.Task]bool)
	matches := []cumulus.Task{}

	filePatterns := strings.Split(strings.ReplaceAll(patternStr, " ", ""), ",")
	for _, pattern := range filePatterns {
		filePaths, err := searchFn(pattern)
		if err != nil {
			return nil, err
		}
		for _, value := range filePaths {
			info, err := fileInfo(value)
			if err != nil {
				log.Entry().WithError(err).Warnf("failed to get file info, skipping '%s'", value)
				continue
			}
			if info.IsDir() {
				continue
			}
			task := cumulus.Task{SourcePath: value, TargetPath: value}
			if !keys[task] {
				keys[task] = true
				matches = append(matches, task)
			}
		}
	}
	return matches, nil
}

func uploadFiles(ctx context.Context, gcsClient gcs.Client, config *cumulus.Cumulus, uploadTargets *[]cumulus.Task, targetFolder string) error {
	for _, element := range *uploadTargets {
		err := gcsClient.UploadFile(ctx, config.PipelineID, element.SourcePath, filepath.Join(targetFolder, element.TargetPath))
		if err != nil && config.BucketLockedWarning {
			var e *googleapi.Error
			if errors.As(err, &e) {
				// determine if err is 'cumulus bucket closed' error
				re := regexp.MustCompile(bucketLockedErrString)
				if e.Code == http.StatusForbidden && re.MatchString(e.Message) {
					log.Entry().Warningf("skipping uploading %v since file already exists in %v", element.SourcePath, targetFolder)
					continue
				}
			}
		}
		if err != nil {
			return errors.Wrapf(err, "could not upload files to cumulus bucket: %v", err)
		}
	}
	return nil
}

func exportPipelineRunKey(pipelineEnv *sapCumulusUploadCommonPipelineEnvironment, sapCumulus *cumulus.Cumulus) error {
	log.Entry().Infof("Exporting pipeline run key: %s", pipelineEnv.custom.cumulusPipelineRunKey)
	log.Entry().Infof("eventData: %s", sapCumulus.EventData)
	if len(sapCumulus.EventData) > 0 {
		eventData, err := events.FromJSON([]byte(sapCumulus.EventData))
		if err != nil {
			return errors.Wrapf(err, "failed to read event data")
		}
		eventData.CumulusInformation.RunId = pipelineEnv.custom.cumulusPipelineRunKey
		eventBytes, err := eventData.ToJSON()
		if err != nil {
			return errors.Wrapf(err, "failed to write event data")
		}
		pipelineEnv.custom.eventData = string(eventBytes)
	}
	return nil
}

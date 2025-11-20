package cmd

import (
	"context"
	"github.com/SAP/jenkins-library/pkg/gcs"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/bmatcuk/doublestar"
	"github.com/pkg/errors"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/cumulus"
)

func sapCumulusDownload(config sapCumulusDownloadOptions, _ *telemetry.CustomData) {
	sapCumulus := cumulus.Cumulus{
		EnvVars: []cumulus.EnvVar{
			{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: config.JSONKeyFilePath, Modified: false},
		},
		FilePattern:           config.FilePattern,
		PipelineID:            config.PipelineID,
		StepResultType:        config.StepResultType,
		SubFolderPath:         config.SubFolderPath,
		Version:               config.Version,
		TargetPath:            config.TargetPath,
		HeadCommitID:          config.GitHeadCommitID,
		UseCommitIDForCumulus: config.UseCommitIDForCumulus,
	}

	ctx := context.Background()
	client, err := gcs.NewClient(config.JSONKeyFilePath, config.Token)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to initialize GCS client")
		return
	}

	if err = runSapCumulusDownload(ctx, client, &sapCumulus); err != nil {
		log.Entry().WithError(err).Fatal("Failed to perform Cumulus download.")
	}
}

// searchFiles searches files in cumulus using file patterns for certain sourcePath
// The sourcePath represents string "<version>/<stepResultType>/subFolderPath"
// e.g. "v1/general/sub/folder" (see parameters)
func searchCumulusFiles(ctx context.Context, gcsClient gcs.Client, patternStr string, bucketID string, targetPath string, sourcePath string) ([]cumulus.Task, error) {
	keys := make(map[cumulus.Task]bool)
	matches := []cumulus.Task{}

	targetPath = appendSuffix(targetPath)

	filePatterns := strings.Split(strings.ReplaceAll(patternStr, " ", ""), ",")
	fileNames, err := gcsClient.ListFiles(ctx, bucketID)
	if err != nil {
		return nil, errors.Wrapf(err, "list bucket files failed: %v", err)
	}
	for _, pattern := range filePatterns {
		for _, name := range fileNames {
			if !strings.HasPrefix(name, sourcePath) {
				continue
			}
			nameWithoutSourcePath := strings.TrimLeft(strings.TrimPrefix(name, sourcePath), "/")
			ok, err := doublestar.Match(pattern, nameWithoutSourcePath)
			if err != nil {
				return nil, errors.Wrapf(err, "pattern match failed: %v", err)
			}
			task := cumulus.Task{SourcePath: name, TargetPath: targetPath + nameWithoutSourcePath}
			if ok && !keys[task] {
				keys[task] = true
				matches = append(matches, task)
			}
		}
	}

	log.Entry().Debugf("Filtered filenames: %v", matches)
	return matches, nil
}

func appendSuffix(path string) string {
	if path != "" && !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	return path
}

func runSapCumulusDownload(ctx context.Context, gcsClient gcs.Client, sapCumulus *cumulus.Cumulus) error {
	if err := sapCumulus.ValidateInput(); err != nil {
		return err
	}

	sapCumulus.PrepareEnv()
	defer sapCumulus.CleanupEnv()

	pipelineRunKey, err := sapCumulus.GetPipelineRunKey()
	if err != nil {
		return errors.Wrap(err, "failed to determine pipeline run key")
	}

	sourcePath := sapCumulus.GetCumulusPath(pipelineRunKey)

	log.Entry().Infof("looking for %v in cumulus bucket %v", sapCumulus.FilePattern, sapCumulus.PipelineID)
	matches, err := searchCumulusFiles(ctx, gcsClient, sapCumulus.FilePattern, sapCumulus.PipelineID, sapCumulus.TargetPath, sourcePath)
	if err != nil {
		return errors.Wrapf(err, "search files failed: %v", err)
	}
	log.Entry().Infof("downloading %v file(s) from %v", len(matches), sourcePath)

	if err = downloadFiles(ctx, gcsClient, sapCumulus.PipelineID, &matches); err != nil {
		return errors.Wrapf(err, "download failed: %v", err)
	}

	log.Entry().Infoln("cumulus download was successful")
	return nil
}

func downloadFiles(ctx context.Context, gcsClient gcs.Client, bucketID string, downloadTargets *[]cumulus.Task) error {
	for _, element := range *downloadTargets {
		err := gcsClient.DownloadFile(ctx, bucketID, element.SourcePath, element.TargetPath)
		if err != nil {
			return errors.Wrapf(err, "could not download files from cumulus bucket: %v", err)
		}
	}
	return nil
}

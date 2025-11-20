package xmake

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/SAP/jenkins-library/pkg/log"

	"github.com/SAP/jenkins-library/pkg/piperutils"
)

const (
	typeStage   = "xMakeStage"
	typePromote = "xMakePromote"
)

type BuildType struct {
	BuildType string `json:"build-type,omitempty"`
}

type fileWriter interface {
	FileWrite(filename string, data []byte, perm os.FileMode) error
}

func (report *BuildType) ToBuildTypeFile(files fileWriter) error {
	fileName := "build-type.json"

	// ignore error since format is in our hands
	data, _ := json.Marshal(report)

	if err := files.FileWrite(fileName, data, 0o666); err != nil {
		return fmt.Errorf("failed to write %v: %w", fileName, err)
	}

	return nil
}

func WriteBuildTypeJsonForStageBuild(stageBom map[string]interface{}, fileUtils piperutils.FileUtils) (string, error) {
	var format BuildType
	var buildType string

	// Search if type docker build
	for _, stageBOM := range stageBom {
		sb := stageBOM.(map[string]interface{})
		buildType = fmt.Sprint(sb["format"])
		if buildType == "docker" {
			log.Entry().Info("staged docker build")
			break
		}
	}

	if buildType != "docker" {
		buildType = "bin"
	}

	format = BuildType{BuildType: buildType}
	if err := WriteBuildTypeToFile(format, fileUtils); err != nil {
		return "", err
	}
	return buildType, nil
}

func WriteBuildTypeToFile(format BuildType, fileUtils piperutils.FileUtils) error {
	if err := format.ToBuildTypeFile(fileUtils); err != nil {
		log.Entry().WithError(err).Error("failed to write build-type file")
		return err
	}
	return nil
}

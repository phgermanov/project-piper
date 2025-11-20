package dwc

import (
	"encoding/json"
	"fmt"
)

// Flag collection for dwc upload commands

const (
	genericErrorFormat                  = "%s: %w"
	failedToAppendUploadMetadata string = "failed to append upload metadata"
)

var uploadBaseCommand = dwcCommand{"artifact", "upload"}

const (
	uiUploadSubcommand         = "ui"
	javaUploadSubcommand       = "java"
	mtaUploadSubcommand        = "mta"
	orbitUploadSubcommand      = "orbit"
	dockerUploadSubcommand     = "docker"
	helmUploadSubcommand       = "helm"
	waitForPromotionFlag       = "--wait-for-promotion"
	metadataFlag               = "--metadata=%s=%s"
	buildMetadataKey           = "build"
	resourceFlag               = "--resource=%s"
	filesFlag                  = "--files=%s"
	jarURLFlag                 = "--jar-url=%s"
	mtaURLFlag                 = "--mta-url=%s"
	appNameFlag                = "--appname=%s"
	dockerImageFlag            = "--docker-image=%s"
	containerImageFlag         = "--container-image=%s"
	additionalDownloadUrlsFlag = "--additional-download-urls=%s"
)

func newUIUploadCommand(artifact *UIArtifact) (dwcCommand, error) {
	args := append(uploadBaseCommand, []string{uiUploadSubcommand, waitForPromotionFlag, fmt.Sprintf(metadataFlag, buildMetadataKey, artifact.buildMetadata), fmt.Sprintf(resourceFlag, artifact.ResourceName), fmt.Sprintf(filesFlag, artifact.getUploadFileName())}...)
	args = appendAppNameArgs(args, artifact.DescriptorBase)
	args, err := appendUploadMetadata(args, artifact.UploadMetadata, artifact.DescriptorBase.Apps)
	if err != nil {
		return nil, fmt.Errorf(genericErrorFormat, failedToAppendUploadMetadata, err)
	}
	args = appendOutputFormat(args, outputFormatJSON)
	return args, nil
}

func newJavaUploadCommand(artifact *JavaArtifact) (dwcCommand, error) {
	args := append(uploadBaseCommand, []string{javaUploadSubcommand, waitForPromotionFlag, fmt.Sprintf(resourceFlag, artifact.ResourceName), fmt.Sprintf(jarURLFlag, artifact.ArtifactURL), fmt.Sprintf(metadataFlag, buildMetadataKey, artifact.buildMetadata), fmt.Sprintf(filesFlag, artifact.getUploadFileName())}...)
	args = appendAppNameArgs(args, artifact.DescriptorBase)
	args, err := appendUploadMetadata(args, artifact.UploadMetadata, artifact.DescriptorBase.Apps)
	if err != nil {
		return nil, fmt.Errorf(genericErrorFormat, failedToAppendUploadMetadata, err)
	}
	args = appendAdditionalDownloadUrls(args, artifact.AdditionalDownloadURLs)
	args = appendOutputFormat(args, outputFormatJSON)
	return args, nil
}

func newMTAUploadCommand(artifact *MTAArtifact) (dwcCommand, error) {
	args := append(uploadBaseCommand, []string{mtaUploadSubcommand, waitForPromotionFlag, fmt.Sprintf(resourceFlag, artifact.ResourceName), fmt.Sprintf(mtaURLFlag, artifact.ArtifactURL), fmt.Sprintf(metadataFlag, buildMetadataKey, artifact.buildMetadata)}...)
	args = appendAppNameArgs(args, artifact.DescriptorBase)
	if artifact.hasFilePatterns() {
		args = append(args, fmt.Sprintf(filesFlag, artifact.getUploadFileName()))
	}
	args, err := appendUploadMetadata(args, artifact.UploadMetadata, artifact.DescriptorBase.Apps)
	if err != nil {
		return nil, fmt.Errorf(genericErrorFormat, failedToAppendUploadMetadata, err)
	}
	args = appendAdditionalDownloadUrls(args, artifact.AdditionalDownloadURLs)
	args = appendOutputFormat(args, outputFormatJSON)
	return args, nil
}

func newDockerUploadCommand(artifact *DockerArtifact) (dwcCommand, error) {
	args := append(uploadBaseCommand, []string{dockerUploadSubcommand, waitForPromotionFlag, fmt.Sprintf(dockerImageFlag, artifact.ContainerImageLocator), fmt.Sprintf(resourceFlag, artifact.ResourceName), fmt.Sprintf(metadataFlag, buildMetadataKey, artifact.buildMetadata), fmt.Sprintf(filesFlag, artifact.getUploadFileName())}...)
	args = appendAppNameArgs(args, artifact.DescriptorBase)
	args, err := appendUploadMetadata(args, artifact.UploadMetadata, artifact.DescriptorBase.Apps)
	if err != nil {
		return nil, fmt.Errorf(genericErrorFormat, failedToAppendUploadMetadata, err)
	}
	args = appendAdditionalDownloadUrls(args, artifact.AdditionalDownloadURLs)
	args = appendOutputFormat(args, outputFormatJSON)
	return args, nil
}

func newOrbitUploadCommand(artifact *OrbitArtifact) (dwcCommand, error) {
	args := append(uploadBaseCommand, []string{orbitUploadSubcommand, waitForPromotionFlag, fmt.Sprintf(containerImageFlag, artifact.ContainerImageLocator), fmt.Sprintf(resourceFlag, artifact.ResourceName), fmt.Sprintf(metadataFlag, buildMetadataKey, artifact.buildMetadata), fmt.Sprintf(filesFlag, artifact.getUploadFileName())}...)
	args = appendAppNameArgs(args, artifact.DescriptorBase)
	args, err := appendUploadMetadata(args, artifact.UploadMetadata, artifact.DescriptorBase.Apps)
	if err != nil {
		return nil, fmt.Errorf(genericErrorFormat, failedToAppendUploadMetadata, err)
	}
	args = appendAdditionalDownloadUrls(args, artifact.AdditionalDownloadURLs)
	args = appendOutputFormat(args, outputFormatJSON)
	return args, nil
}

func newHelmUploadCommand(artifact *HelmArtifact) (dwcCommand, error) {
	args := append(uploadBaseCommand, []string{helmUploadSubcommand, waitForPromotionFlag, fmt.Sprintf(resourceFlag, artifact.ResourceName), fmt.Sprintf(filesFlag, artifact.getUploadFileName()), fmt.Sprintf(metadataFlag, buildMetadataKey, artifact.buildMetadata)}...)
	args = appendAppNameArgs(args, artifact.DescriptorBase)
	args, err := appendUploadMetadata(args, artifact.UploadMetadata, artifact.DescriptorBase.Apps)
	if err != nil {
		return nil, fmt.Errorf(genericErrorFormat, failedToAppendUploadMetadata, err)
	}
	args = appendOutputFormat(args, outputFormatJSON)
	return args, nil
}

func appendAppNameArgs(args dwcCommand, descriptorBase *DescriptorBase) dwcCommand {
	if descriptorBase.AppName != "" {
		args = append(args, fmt.Sprintf(appNameFlag, descriptorBase.AppName))
		return args
	}
	for _, app := range descriptorBase.Apps {
		args = append(args, fmt.Sprintf(appNameFlag, app.Name))
	}
	return args
}

func getEuporieTaskCollectionOptOutMetadataValue(apps []App) []string {
	optOuts := make([]string, 0)
	for _, app := range apps {
		if app.NoEuporieTaskCollection {
			optOuts = append(optOuts, app.Name)
		}
	}
	return optOuts
}

func getRouteAssignmentOptOutMetadataValue(apps []App) []string {
	optOuts := make([]string, 0)
	for _, app := range apps {
		if app.NoRouteAssignment {
			optOuts = append(optOuts, app.Name)
		}
	}
	return optOuts
}

func getStaticRoutesOptInMetadataValue(apps []App) []string {
	optIns := make([]string, 0) // codespell:ignore optins
	for _, app := range apps {
		if app.AllowStaticRoutes {
			optIns = append(optIns, app.Name) // codespell:ignore optins
		}
	}
	return optIns // codespell:ignore optins
}

func getJsonArrayStringFromSlice(slice []string) (string, error) {
	jsonArray, err := json.Marshal(slice)
	if err != nil {
		return "", err
	}
	return string(jsonArray), nil
}

func appendUploadMetadata(args dwcCommand, uploadMetadata map[string]string, apps []App) (dwcCommand, error) {
	for key, value := range uploadMetadata {
		args = append(args, []string{fmt.Sprintf(metadataFlag, key, value)}...)
	}
	euporieTaskCollectionOptOuts := getEuporieTaskCollectionOptOutMetadataValue(apps)
	if len(euporieTaskCollectionOptOuts) > 0 {
		jsonArrayString, err := getJsonArrayStringFromSlice(euporieTaskCollectionOptOuts)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal euporie task collection opt outs as json: %w", err)
		}
		args = append(args, fmt.Sprintf(metadataFlag, NoEuporieTaskCollectionMetadataKey, jsonArrayString))
	}
	routeAssignmentOptOuts := getRouteAssignmentOptOutMetadataValue(apps)
	if len(routeAssignmentOptOuts) > 0 {
		jsonArrayString, err := getJsonArrayStringFromSlice(routeAssignmentOptOuts)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal route assignments opt outs as json: %w", err)
		}
		args = append(args, fmt.Sprintf(metadataFlag, NoRouteAssignmentMetadataKey, jsonArrayString))
	}
	staticRoutesOptIns := getStaticRoutesOptInMetadataValue(apps)
	if len(staticRoutesOptIns) > 0 {
		jsonArrayString, err := getJsonArrayStringFromSlice(staticRoutesOptIns)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal static routes opt ins as json: %w", err)
		}
		args = append(args, fmt.Sprintf(metadataFlag, AllowStaticRoutesMetadataKey, jsonArrayString))
	}
	return args, nil
}

func appendAdditionalDownloadUrls(args dwcCommand, additionalDownloadUrls map[string]string) dwcCommand {
	for key, additionalDownloadUrl := range additionalDownloadUrls {
		keyToUrl := fmt.Sprintf("%s=%s", key, additionalDownloadUrl)
		args = append(args, fmt.Sprintf(additionalDownloadUrlsFlag, keyToUrl))
	}
	return args
}

func appendOutputFormat(args dwcCommand, outputFormat string) dwcCommand {
	return append(args, fmt.Sprintf(outputFlag, outputFormat))
}

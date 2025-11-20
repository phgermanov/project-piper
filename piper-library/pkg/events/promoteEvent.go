package events

import (
	"fmt"
	"strings"

	piperOsCmd "github.com/SAP/jenkins-library/cmd"
	"github.com/SAP/jenkins-library/pkg/config"
	piperOsEvents "github.com/SAP/jenkins-library/pkg/events"
	"github.com/SAP/jenkins-library/pkg/gcp"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/versioning"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/buildtools"
)

type StagingServiceData struct {
	HeadCommitID         string
	CommitID             string
	GitURL               string
	MavenBuildArtifacts  string
	NpmBuildArtifacts    string
	MtaBuildArtifacts    string
	PromotedArtifactURLs []string
}

type PromotedArtifactsEventData struct {
	CommitId          string                   `json:"commitId"`
	TagCommitId       string                   `json:"tagCommitId,omitempty"`
	RepositoryUrl     string                   `json:"repositoryUrl"`
	PromotedArtifacts []versioning.Coordinates `json:"promotedArtifacts"`
}

func SendPromotedArtifactsEvent(data StagingServiceData) {
	eventData, err := getArtifactPromotionEvent(data)
	if err != nil {
		log.Entry().Warnf("Error getting promoted artifacts event data: %v", err)
		return
	}

	eventBytes, err := eventData.ToBytes()
	if err != nil {
		log.Entry().Warnf("failed to create event data: %v", err)
		return
	}

	if len(eventBytes) == 0 {
		log.Entry().Warn("No event data to send")
		return
	}

	log.Entry().Debugf("Sending promoted artifacts event as json: %v", string(eventBytes))
	publishEvent(eventBytes)
}

func getArtifactPromotionEvent(data StagingServiceData) (piperOsEvents.Event, error) {
	promotedArtifacts, err := getPromotedArtifacts(data)
	if err != nil {
		return piperOsEvents.Event{}, err
	}

	log.Entry().Infof("Promoted Artifacts: %+v", promotedArtifacts)

	tagCommitId := ""
	if data.CommitID != "" && data.CommitID != data.HeadCommitID {
		tagCommitId = data.CommitID
	}

	promotedArtifactsEventData := PromotedArtifactsEventData{
		CommitId:          data.HeadCommitID,
		TagCommitId:       tagCommitId,
		RepositoryUrl:     data.GitURL,
		PromotedArtifacts: promotedArtifacts,
	}

	event := piperOsEvents.NewEvent("sap.hyperspace.artifactsPromoted", "/default/sap.hyperspace.piper", data.HeadCommitID).Create(promotedArtifactsEventData)
	return event, nil
}

func publishEvent(eventData []byte) {
	vaultClient := config.GlobalVaultClient()
	err := gcp.NewGcpPubsubClient(
		vaultClient,
		piperOsCmd.GeneralConfig.HookConfig.GCPPubSubConfig.ProjectNumber,
		piperOsCmd.GeneralConfig.HookConfig.GCPPubSubConfig.IdentityPool,
		piperOsCmd.GeneralConfig.HookConfig.GCPPubSubConfig.IdentityProvider,
		piperOsCmd.GeneralConfig.CorrelationID,
		piperOsCmd.GeneralConfig.HookConfig.OIDCConfig.RoleID,
	).Publish("hyperspace-artifacts-promoted", eventData)
	if err != nil {
		log.Entry().WithError(err).Warn("event publish failed")
	}
}

func getPromotedArtifacts(data StagingServiceData) ([]versioning.Coordinates, error) {
	var buildTool buildtools.BuildTool
	var artifactsJsonString string

	switch {
	case data.MavenBuildArtifacts != "":
		buildTool = buildtools.Maven{}
		artifactsJsonString = data.MavenBuildArtifacts
	case data.NpmBuildArtifacts != "":
		buildTool = buildtools.Npm{}
		artifactsJsonString = data.NpmBuildArtifacts
	case data.MtaBuildArtifacts != "":
		buildTool = buildtools.Mta{}
		artifactsJsonString = data.MtaBuildArtifacts
	default:
		return nil, fmt.Errorf("no build artifacts found for maven, npm or mta")
	}

	buildArtifacts, err := buildTool.GetBuildArtifacts(artifactsJsonString)
	if err != nil {
		return nil, err
	}

	var promotedArtifacts []versioning.Coordinates
	for _, promotedUrl := range data.PromotedArtifactURLs {
		buildCoordinate, err := buildTool.GetPromotedArtifact(promotedUrl, buildArtifacts)
		if err != nil {
			log.Entry().Infof("no build coordinate found for %v", promotedUrl)
			continue
		}

		buildCoordinate.URL = promotedUrl
		if strings.HasSuffix(promotedUrl, "pom") {
			buildCoordinate.Packaging = "pom"
		}

		promotedArtifacts = append(promotedArtifacts, buildCoordinate)
	}

	if len(promotedArtifacts) == 0 {
		return nil, fmt.Errorf("no promoted artifacts event data: %w", err)
	}

	return promotedArtifacts, nil
}

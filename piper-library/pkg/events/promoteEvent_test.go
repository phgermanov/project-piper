package events

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"

	piperOsEvents "github.com/SAP/jenkins-library/pkg/events"
	"github.com/stretchr/testify/assert"
)

func TestGetArtifactPromotionEvent(t *testing.T) {
	t.Parallel()

	t.Run("error case with no build coordinates", func(t *testing.T) {
		promotedArtifactURLs := []string{"test/artifact/1.0/artifact-1.0.pom", "test/artifact/1.0/artifact-1.0.jar"}
		stagingServiceData := StagingServiceData{
			PromotedArtifactURLs: promotedArtifactURLs,
		}
		_, err := getArtifactPromotionEvent(stagingServiceData)
		assert.Error(t, err)
	})

	t.Run("success case - maven module", func(t *testing.T) {
		MavenBuildArtifacts := `{"coordinates":[{"GroupID":"test","ArtifactID":"artifact","Version":"1.0","Packaging":"pom","BuildPath":"/tmp","URL":"artifact/1.0/artifact-1.0.pom"},
{"GroupID":"test","ArtifactID":"artifact","Version":"1.0","Packaging":"pom","BuildPath":"/tmp","URL":"artifact/1.0/artifact-1.0.jar"}]}`

		promotedArtifactURLs := []string{"test/artifact/1.0/artifact-1.0.pom", "test/artifact/1.0/artifact-1.0.jar"}

		stagingServiceData := StagingServiceData{
			MavenBuildArtifacts:  MavenBuildArtifacts,
			PromotedArtifactURLs: promotedArtifactURLs,
		}

		artifactPromotionEvent, err := getArtifactPromotionEvent(stagingServiceData)
		assert.NoError(t, err)

		artifactPromotionEventData, err := extractEventData(artifactPromotionEvent)
		assert.NoError(t, err)

		assert.Equal(t, len(promotedArtifactURLs), 2)
		assert.Equal(t, artifactPromotionEventData.PromotedArtifacts[0].BuildPath, "/tmp")
	})

	t.Run("success case - npm", func(t *testing.T) {
		NpmBuildArtifacts := `{"Coordinates": [{"GroupID":"","ArtifactID":"artifact","Version":"1.0","BuildPath":"/tmp","URL":"artifact-1.0.tgz"}]}`
		promotedArtifactURLs := []string{"https://staging.repositories.cloud.sap/stage/repository/artifact/1.0/artifact-1.0.tgz"}

		stagingServiceData := StagingServiceData{
			NpmBuildArtifacts:    NpmBuildArtifacts,
			PromotedArtifactURLs: promotedArtifactURLs,
		}

		artifactPromotionEvent, err := getArtifactPromotionEvent(stagingServiceData)
		assert.NoError(t, err)

		artifactPromotionEventData, err := extractEventData(artifactPromotionEvent)
		assert.NoError(t, err)

		assert.Equal(t, len(artifactPromotionEventData.PromotedArtifacts), 1)
		assert.Equal(t, artifactPromotionEventData.PromotedArtifacts[0].BuildPath, "/tmp")
		assert.Equal(t, artifactPromotionEventData.PromotedArtifacts[0].URL, "https://staging.repositories.cloud.sap/stage/repository/artifact/1.0/artifact-1.0.tgz")
	})

	t.Run("success case - mta", func(t *testing.T) {
		MtaBuildArtifacts := `{"Coordinates":[{"groupId":"test-group","artifactId":"test-name.mtar","version":"1.0","packaging":"mtar","buildPath":".","url":"https://staging.repositories.cloud.sap/stage/repository/test-group/test-name/1.0/test-name-1.0.mtar","purl":"pkg:mta/azure-demo-cf-mta@1.0.0-id"}]}`
		promotedArtifactURLs := []string{"https://staging.repositories.cloud.sap/stage/repository/test-group/test-name/1.0/test-name-1.0.mtar"}

		stagingServiceData := StagingServiceData{
			MtaBuildArtifacts:    MtaBuildArtifacts,
			PromotedArtifactURLs: promotedArtifactURLs,
		}

		artifactPromotionEvent, err := getArtifactPromotionEvent(stagingServiceData)
		assert.NoError(t, err)

		artifactPromotionEventData, err := extractEventData(artifactPromotionEvent)
		assert.NoError(t, err)

		assert.Equal(t, len(artifactPromotionEventData.PromotedArtifacts), 1)
		assert.Equal(t, artifactPromotionEventData.PromotedArtifacts[0].URL, promotedArtifactURLs[0])
		assert.Equal(t, artifactPromotionEventData.PromotedArtifacts[0].BuildPath, ".")
		assert.Equal(t, artifactPromotionEventData.PromotedArtifacts[0].PURL, "pkg:mta/azure-demo-cf-mta@1.0.0-id")
		assert.Equal(t, artifactPromotionEventData.PromotedArtifacts[0].Packaging, "mtar")
		assert.Equal(t, artifactPromotionEventData.PromotedArtifacts[0].Version, "1.0")
		assert.Equal(t, artifactPromotionEventData.PromotedArtifacts[0].GroupID, "test-group")
		assert.Equal(t, artifactPromotionEventData.PromotedArtifacts[0].ArtifactID, "test-name.mtar")
	})
}

func TestSendPromotedArtifactsEvent(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	data := StagingServiceData{
		HeadCommitID:         "testHeadCommitID",
		CommitID:             "012345",
		GitURL:               "url",
		MavenBuildArtifacts:  `{"coordinates":[{"GroupID":"test.group","ArtifactID":"artifact","Version":"1.0","Packaging":"pom","BuildPath":"/tmp","URL":"test/group/artifact/1.0/artifact-1.0.pom"}]}`,
		PromotedArtifactURLs: []string{"test/group/artifact/1.0/artifact-1.0.pom"},
	}

	SendPromotedArtifactsEvent(data)

	logged := buf.String()
	assert.Empty(t, logged)
}

type paEventData struct {
	Data PromotedArtifactsEventData `json:"data"`
}

func extractEventData(event piperOsEvents.Event) (PromotedArtifactsEventData, error) {
	data, err := event.ToBytes()
	if err != nil {
		return PromotedArtifactsEventData{}, err
	}
	var paData paEventData
	err = json.Unmarshal(data, &paData)
	if err != nil {
		return PromotedArtifactsEventData{}, err
	}
	return paData.Data, nil
}

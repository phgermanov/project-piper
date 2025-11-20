package buildtools

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/versioning"
	"github.com/stretchr/testify/assert"
)

func TestGetPromotedArtifactForMaven(t *testing.T) {
	t.Parallel()
	t.Run("coordinates found", func(t *testing.T) {
		coordinates := []versioning.Coordinates{{GroupID: "com.test", ArtifactID: "artifact", Version: "1.0", BuildPath: "/tmp"}}
		mavenTool := Maven{}
		coordinate, err := mavenTool.GetPromotedArtifact("com/test/artifact/1.0/artifact-1.0.jar", coordinates)
		assert.NoError(t, err)
		assert.Equal(t, "com.test", coordinate.GroupID)
	})
	t.Run("coordinates not found", func(t *testing.T) {
		mavenTool := Maven{}
		coordinate, err := mavenTool.GetPromotedArtifact("com/test/artifact/2.0/artifact-2.0.jar", []versioning.Coordinates{})
		assert.Error(t, err)
		assert.Empty(t, coordinate.GroupID)
	})
}

func TestGetPromotedArtifactForNpm(t *testing.T) {
	t.Parallel()
	t.Run("coordinates found", func(t *testing.T) {
		coordinates := []versioning.Coordinates{{GroupID: "test", ArtifactID: "artifact", Version: "1.0", BuildPath: "/tmp"}}
		npmTool := Npm{}
		coordinate, err := npmTool.GetPromotedArtifact("artifact-1.0", coordinates)
		assert.NoError(t, err)
		assert.Equal(t, "test", coordinate.GroupID)
	})
	t.Run("coordinates found for scoped packages", func(t *testing.T) {
		npmTool := Npm{}
		coordinates := []versioning.Coordinates{
			{
				GroupID:    "",
				ArtifactID: "@sap-ppms/sbom-gateway",
				Version:    "0.0.1-20250624030611+f3a2bcface39b728991d8dc39a49c83e23851696",
				BuildPath:  "./", URL: "https://staging.repositories.cloud.sap/stage/repository/376100898081-20250624-030617368-209/",
				PURL:      "pkg:npm/%40sap-ppms/sbom-gateway@0.0.1-20250624030611+f3a2bcface39b728991d8dc39a49c83e23851696",
				Packaging: "tgz",
			},
		}
		coordinate, err := npmTool.GetPromotedArtifact("https://common.repositories.cloud.sap/api/npm/deploy-releases-****-npm/@sap-ppms/sbom-gateway/-/sbom-gateway-0.0.1-20250624030611.tgz", coordinates)
		assert.NoError(t, err)
		assert.Equal(t, "@sap-ppms/sbom-gateway", coordinate.ArtifactID)
	})
	t.Run("coordinates not found", func(t *testing.T) {
		npmTool := Npm{}
		coordinate, err := npmTool.GetPromotedArtifact("artifact-2.0", []versioning.Coordinates{})
		assert.Error(t, err)
		assert.Empty(t, coordinate.GroupID)
	})
}

func TestGetPromotedArtifactForMta(t *testing.T) {
	t.Parallel()
	t.Run("coordinates found", func(t *testing.T) {
		coordinates := []versioning.Coordinates{{GroupID: "com.test", ArtifactID: "artifact.mtar", Version: "1.0", BuildPath: "/tmp"}}
		mtaTool := Mta{}
		coordinate, err := mtaTool.GetPromotedArtifact("https://staging.repositories.cloud.sap/stage/repository/com.test/artifact/1.0/artifact-1.0.mtar", coordinates)
		assert.NoError(t, err)
		assert.Equal(t, "com.test", coordinate.GroupID)
	})
	t.Run("coordinates not found", func(t *testing.T) {
		mtaTool := Mta{}
		coordinate, err := mtaTool.GetPromotedArtifact("https://staging.repositories.cloud.sap/stage/repository/com.test/artifact/2.0/artifact-2.0.mtar", []versioning.Coordinates{})
		assert.Error(t, err)
		assert.Empty(t, coordinate.GroupID)
	})
}

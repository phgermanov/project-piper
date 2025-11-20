//go:build unit
// +build unit

package fosstars

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const (
	REPORT_PATH string = "./fosstars-report"
)

func TestGenerateHtmlReport(t *testing.T) {
	defer os.RemoveAll(REPORT_PATH) // clean up
	artifactsRatingsMap :=
		make(map[string]*Rating)
	artifactsRatingsMap["com.sap.test:test-common"] = &Rating{Id: "123445", RatingDefinitionId: "32435645765", Created: "2020-04-14T122344", Value: 0.23444, Confidence: 0.3, Label: "BAD", NameSpace: "", Name: "", RepositoryType: "", CoordinateValue: ""}
	artifactsRatingsMap["com.fasterxml.jackson.core:jackson-databind"] = &Rating{Id: "125687", RatingDefinitionId: "123545", Created: "2020-04-14T122344", Value: 0.43444, Confidence: 0.3, Label: "BAD", NameSpace: "", Name: "", RepositoryType: "", CoordinateValue: ""}

	dependencyTree := []string{}
	dependencyTree = append(dependencyTree, "com.sap.test:test-common:pom:1.0")
	dependencyTree = append(dependencyTree, "+- com.fasterxml.jackson.core:jackson-databind:jar:2.9.10")

	GenerateHtmlReportForMaven(artifactsRatingsMap, dependencyTree, "https://fosstars-qs.cfapps.sap.hana.ondemand.com/v1/ratings/namespaces/Phosphor/names/Security/repos/maven/identifiers")
	if _, err := os.Stat(REPORT_PATH + "/fosstars-test-common-fosstars.html"); os.IsNotExist(err) {
		t.Error("Expected Fosstar Report but its not created")
	}
}

func TestGenerateHtmlRepotForNpm(t *testing.T) {
	defer os.RemoveAll(REPORT_PATH) // clean up
	artifactsRatingsMap :=
		make(map[string]*Rating)
	artifactsRatingsMap["axios"] = &Rating{Id: "123445", RatingDefinitionId: "32435645765", Created: "2020-04-14T122344", Value: 0.23444, Confidence: 0.3, Label: "BAD", NameSpace: "", Name: "", RepositoryType: "", CoordinateValue: ""}
	artifactsRatingsMap["commander"] = &Rating{Id: "125687", RatingDefinitionId: "123545", Created: "2020-04-14T122344", Value: 0.43444, Confidence: 0.3, Label: "BAD", NameSpace: "", Name: "", RepositoryType: "", CoordinateValue: ""}

	GenerateHtmlRepotForNpm(artifactsRatingsMap, "test-common-npm:test", "https://fosstars-qs.cfapps.sap.hana.ondemand.com/v1/ratings/namespaces/Phosphor/names/Security/repos/maven/identifiers")
	if _, err := os.Stat(REPORT_PATH + "/fosstars-test-common-npm-fosstars.html"); os.IsNotExist(err) {
		t.Error("Expected Fosstars NPM Report but its not created")
	}

}

func TestCreateJsonReport(t *testing.T) {
	if _, err := os.Stat(REPORT_PATH); os.IsNotExist(err) {
		os.Mkdir(REPORT_PATH, 0777)
	}
	defer os.RemoveAll(REPORT_PATH) // clean up
	artifactsRatingsMap :=
		make(map[string]interface{})
	artifactsRatingsMap["axios"] = Rating{Id: "123445", RatingDefinitionId: "32435645765", Created: "2020-04-14T122344", Value: 0.23444, Confidence: 0.3, Label: "BAD", NameSpace: "", Name: "", RepositoryType: "", CoordinateValue: ""}
	artifactsRatingsMap["commander"] = Rating{Id: "125687", RatingDefinitionId: "123545", Created: "2020-04-14T122344", Value: 0.43444, Confidence: 0.3, Label: "BAD", NameSpace: "", Name: "", RepositoryType: "", CoordinateValue: ""}
	modelratingsDefinitionDetails := make(map[string]string)
	modelratingsDefinitionDetails["32435645765"] = "testModelRatingDefinition"
	CreateJsonReport(modelratingsDefinitionDetails, artifactsRatingsMap, "test-Json-Report")
	if _, err := os.Stat(REPORT_PATH + "/fosstars-test-Json-Report.json"); os.IsNotExist(err) {
		t.Error("Expected Fosstars test-Json-Report but its not created")
	}

}

func TestGetRatingStatistics(t *testing.T) {
	ratingDetails := []RatingData{}
	ratingDetails = append(ratingDetails, RatingData{Artifact: "one", RatingId: "", RatingLabel: "GOOD", Created: ""})
	ratingDetails = append(ratingDetails, RatingData{Artifact: "two", RatingId: "", RatingLabel: "BAD", Created: ""})
	ratingDetails = append(ratingDetails, RatingData{Artifact: "three", RatingId: "", RatingLabel: "MODERATE", Created: ""})
	ratingDetails = append(ratingDetails, RatingData{Artifact: "four", RatingId: "", RatingLabel: "UNKNOWN", Created: ""})
	ratingDetails = append(ratingDetails, RatingData{Artifact: "five", RatingId: "", RatingLabel: "GOOD", Created: ""})
	ratingStatistics := GetRatingStatistics(ratingDetails)
	assert.Equal(t, ratingStatistics.TotalDepedenciesCount, 5, "Expecting the proper TotalDepedenciesCount count")
	assert.Equal(t, ratingStatistics.TotalBadRatingsCount, 1, "Expecting the proper TotalBadRatingsCount count")
	assert.Equal(t, ratingStatistics.TotalModerateRatingsCount, 1, "Expecting the proper TotalModerateRatingsCount count")
	assert.Equal(t, ratingStatistics.TotalGoodRatingsCount, 2, "Expecting the proper TotalGoodRatingsCount count")
	assert.Equal(t, ratingStatistics.TotalUnknownRatingsCount, 1, "Expecting the proper count")
}

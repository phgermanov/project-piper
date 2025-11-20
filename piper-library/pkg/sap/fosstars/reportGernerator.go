package fosstars

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
)

const htmlTemplate = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Fosstars Report</title>
	</head>
	<style>
		body {font-family: Arial, Verdana;}
		table {border-collapse: collapse;}
		div.code {font-family: "Courier New", "Lucida Console";}
		th {padding: 10px; border: 1px solid #ddd;}
		td {padding: 5px;border-bottom: 1px solid #ddd;border-right: 1px solid #ddd;}
		tr:nth-child(even) {background-color: #f2f2f2;}
		.bold {font-weight: bold;}
		pre {margin: 0px;}
	</style>
	<body>
	<h1>{{ .ReportTitle }}</h1>
	<p>
	Generated for: {{ .BuildDescriptorName }} <br />
	- {{ .BuildDescriptorIdentifier }} <br />
	</p>
	<i id="totalDepedenciesCount">Total number of dependencies: {{ .RatingStatisticsDetails.TotalDepedenciesCount }}</i><br />
	<i id="totalBadRatingsCount">  - {{ .RatingStatisticsDetails.TotalBadRatingsCount }} have BAD ratings</i><br />
	<i id="totalModerateRatingsCount">  - {{ .RatingStatisticsDetails.TotalModerateRatingsCount }} have MODERATE ratings</i><br />
	<i id="totalGoodRatingsCount">  - {{ .RatingStatisticsDetails.TotalGoodRatingsCount }} have GOOD ratings</i><br />
	<i id="totalUnknownRatingsCount">  - Couldn't calculate rating for {{ .RatingStatisticsDetails.TotalUnknownRatingsCount }} dependencies</i><br />
	<p>Report generated at {{ .CreatedOn.Format "Jan 02, 2006 15:04:05 GMT" }} <br />
	<a href="https://go.sap.corp/phosphor_fosstars_feedback">Give feedback via Fosstars feedback channel</a></p>

		<table>
    		<tr>
      			<th>Entry #</th>
				<th>Artifact</th>
      			<th>Rating</th>
      			<th>Created</th>
			</tr>
		{{range $key, $value := .RatingDetails}}
		<tr>
		<td>{{inc $key}}</td>
		<td><pre> {{ $value.Artifact }} </pre></td>
		{{if $value.RatingId }}
		<td><a href="https://{{ $.FosstarsQueryServiceHost }}/v2/ratings//{{ $value.RatingId }}">{{ $value.RatingLabel }}</td>
		{{else}}
		<td>UNKNOWN</td>
		{{end}}
		<td>{{ parseDate $value.Created }}</td>
		</tr>
		{{end}}
		</table>
	</body>
</html>`

const (
	htmlFileSuffix       = "-fosstars.html"
	reportFolder         = "fosstars-report"
	fosstarsReportPrefix = "fosstars-"
)

type HtmlReportData struct {
	CreatedOn                 time.Time
	ReportTitle               string
	FosstarsQueryServiceHost  string
	BuildDescriptorName       string
	BuildDescriptorIdentifier string
	RatingStatisticsDetails   RatingStatistics
	RatingDetails             []RatingData
}

type RatingStatistics struct {
	TotalDepedenciesCount, TotalBadRatingsCount, TotalModerateRatingsCount, TotalGoodRatingsCount, TotalUnknownRatingsCount int
}

type RatingData struct {
	Artifact, RatingId, RatingLabel, Created string
}

type ArtifactData struct {
	GroupId, ArtifactId, Version string
}

type RatingDetails struct {
	RatingdefinitionDetails map[string]string      `json:"ratingdefinitionDetails"`
	Components              map[string]interface{} `json:"components"`
}

var templateFunctions = template.FuncMap{
	// The name "inc" is what the function will be called in the template text.
	"inc": func(i int) int {
		return i + 1
	},

	"parseDate": func(dateAndTime string) string {
		parts := strings.Split(dateAndTime, "T")
		return parts[0]
	},
}

func GenerateHtmlRepotForNpm(artifactsRatingsMap map[string]*Rating, buildDescriptorIdentifier string, fosstarsQueryServiceURL string) (string, error) {
	buildDescriptorName := getBuildDescriptorNameNpm(buildDescriptorIdentifier)
	fosstarsHost, err := getFosstarsHost(fosstarsQueryServiceURL)
	if err != nil {
		return "", err
	}

	ratingData := getRatingDetailsNpm(artifactsRatingsMap)

	htmlReportData := HtmlReportData{
		CreatedOn:                 time.Now(),
		ReportTitle:               "Fosstars Report: Security Rating",
		FosstarsQueryServiceHost:  fosstarsHost,
		BuildDescriptorName:       buildDescriptorName,
		BuildDescriptorIdentifier: buildDescriptorIdentifier,
		RatingStatisticsDetails:   GetRatingStatistics(ratingData),
		RatingDetails:             ratingData}
	reportName := getReportNameNpm(buildDescriptorName)
	return reportName, createHtmlFile(htmlReportData, reportName+htmlFileSuffix)
}

func GenerateHtmlReportForMaven(artifactsRatingsMap map[string]*Rating, dependencyTree []string, fosstarsQueryServiceURL string) (string, error) {
	artifactData := getRootArtifact(dependencyTree)
	fosstarsHost, err := getFosstarsHost(fosstarsQueryServiceURL)
	if err != nil {
		return "", err
	}

	ratingData := getRatingDetailsMaven(artifactsRatingsMap, dependencyTree)

	htmlReportData := HtmlReportData{
		CreatedOn:                 time.Now(),
		ReportTitle:               "Fosstars Report: Security Rating",
		FosstarsQueryServiceHost:  fosstarsHost,
		BuildDescriptorName:       artifactData.ArtifactId,
		BuildDescriptorIdentifier: getBuildDescriptorIdentifierMaven(artifactData),
		RatingStatisticsDetails:   GetRatingStatistics(ratingData),
		RatingDetails:             ratingData}
	reportName := getReportNameMaven(artifactData)
	return reportName, createHtmlFile(htmlReportData, reportName+htmlFileSuffix)
}

func GetRatingStatistics(ratingData []RatingData) RatingStatistics {
	ratingStatistics := RatingStatistics{}
	for _, ratingDetails := range ratingData {
		switch ratingDetails.RatingLabel {
		case "BAD":
			ratingStatistics.TotalBadRatingsCount++
		case "GOOD":
			ratingStatistics.TotalGoodRatingsCount++
		case "MODERATE":
			ratingStatistics.TotalModerateRatingsCount++
		case "UNKNOWN":
			ratingStatistics.TotalUnknownRatingsCount++
		}
	}
	ratingStatistics.TotalDepedenciesCount = ratingStatistics.TotalBadRatingsCount + ratingStatistics.TotalGoodRatingsCount + ratingStatistics.TotalModerateRatingsCount + ratingStatistics.TotalUnknownRatingsCount
	return ratingStatistics
}

func getFosstarsHost(fosstarsQueryServiceURL string) (string, error) {
	fosstarsQueryServiceUrl, err := url.Parse(fosstarsQueryServiceURL)

	if err != nil {
		log.Entry().Debugf("Could not get Fosstars Host from the URL %v", fosstarsQueryServiceURL)
		return "", err
	}

	return fosstarsQueryServiceUrl.Host, nil
}

func getRatingDetailsMaven(artifactsRatingsMap map[string]*Rating, dependencyTree []string) []RatingData {
	ratingDetails := []RatingData{}

	for index, line := range dependencyTree {

		if index == 0 {
			continue
		}

		arifact := GetArtifactFromLine(line)
		rating, found := artifactsRatingsMap[arifact]
		split := strings.Split(line, ":")
		artifactWithTreeInfo := split[0] + ":" + split[1] + ":" + split[3]
		if found && rating.Label != "" {
			ratingDetails = append(ratingDetails, RatingData{Artifact: artifactWithTreeInfo, RatingId: rating.Id, RatingLabel: rating.Label, Created: rating.Created})
		} else {
			ratingDetails = append(ratingDetails, RatingData{Artifact: artifactWithTreeInfo, RatingId: "", RatingLabel: "UNKNOWN", Created: ""})
		}
	}

	return ratingDetails
}

func createHtmlFile(htmlReportData HtmlReportData, reportName string) error {

	t, err := template.New("t").Funcs(templateFunctions).Parse(htmlTemplate)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(reportFolder, 0777); err != nil /* #nosec G301 */ {
		return err
	}

	filePath := filepath.Clean(filepath.Join(reportFolder, fosstarsReportPrefix+reportName))
	htmlReportFile, err := os.Create(filePath)
	if err != nil {
		return err
	}

	err = t.Execute(htmlReportFile, htmlReportData)
	if err != nil {
		return err
	}
	log.Entry().Debugf("HTML report file created: %s", filePath)

	return nil
}

func CreateJsonReport(modelratingsDefinitionDetails map[string]string, allRatingsMap map[string]interface{}, reportName string) error {
	ratingDetails := new(RatingDetails)
	ratingDetails.RatingdefinitionDetails = modelratingsDefinitionDetails
	ratingDetails.Components = allRatingsMap
	ratingDetailsJson, err := json.Marshal(ratingDetails)
	if err != nil {
		return fmt.Errorf("Error while creating ratingDetailsJson %w", err)
	}
	return WriteJsonToFile(reportName, ratingDetailsJson)
}

func WriteJsonToFile(reportName string, ratingDetailsJson []byte) error {
	filePath := filepath.Join(reportFolder, fosstarsReportPrefix+reportName+".json")
	log.Entry().Debugf("Write JSON to file: %s", filePath)

	if err := os.WriteFile(filePath, ratingDetailsJson, 0644); err != nil /* #nosec G306 */ {
		return fmt.Errorf("error while writing JSON to file: %w", err)
	}
	return nil
}

func getRootArtifact(dependencyTree []string) ArtifactData {
	parentArtifactLine := dependencyTree[0]
	parentArtifactWithoutVersion := GetArtifactFromLine(parentArtifactLine)
	forArtifact := strings.Split(parentArtifactWithoutVersion, "/")
	forVersion := strings.Split(parentArtifactLine, ":")
	return ArtifactData{GroupId: forArtifact[0], ArtifactId: forArtifact[1], Version: forVersion[3]}
}

func getBuildDescriptorNameNpm(buildDescriptorIdentifier string) string {
	split := strings.Split(buildDescriptorIdentifier, ":")
	return split[0]
}

func getBuildDescriptorIdentifierMaven(artifactData ArtifactData) string {
	return artifactData.GroupId + ":" + artifactData.ArtifactId + ":" + artifactData.Version
}

func getReportNameMaven(artifactData ArtifactData) string {
	return artifactData.ArtifactId
}

func getReportNameNpm(buildDescriptorName string) string {
	return buildDescriptorName
}

func getRatingDetailsNpm(artifactsRatingsMap map[string]*Rating) []RatingData {
	ratingDetails := []RatingData{}

	for artifact, rating := range artifactsRatingsMap {
		ratingDetails = append(ratingDetails, RatingData{Artifact: artifact, RatingId: rating.Id, RatingLabel: rating.Label, Created: rating.Created})
	}

	return ratingDetails
}

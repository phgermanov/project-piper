package gtlc

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

type mappingMockClient struct {
	httpMethod     string
	httpStatusCode int
	urlCalled      string
	requestBody    io.Reader
	responseBody   string
}

func (c *mappingMockClient) SetOptions(opts piperhttp.ClientOptions) {}

func (c *mappingMockClient) SendRequest(method, url string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	c.httpMethod = method
	c.urlCalled = url
	if c.httpStatusCode != 200 {
		return &http.Response{StatusCode: c.httpStatusCode}, fmt.Errorf("HTTP Error")
	}
	return &http.Response{StatusCode: c.httpStatusCode, Body: io.NopCloser(bytes.NewReader([]byte(c.responseBody)))}, nil
}

func TestFossFromMapping(t *testing.T) {
	t.Parallel()

	t.Run("empty mapping", func(t *testing.T) {
		t.Parallel()
		fm := []APIFossMappingPair{}
		foss := FossFromMapping(fm)
		assert.Equal(t, 0, len(foss))
	})

	t.Run("with various mappings", func(t *testing.T) {
		t.Parallel()
		fm := []APIFossMappingPair{
			{StandardFoss: FossRelationshipTO{ComplianceInfo: ComplianceInfo{
				IPCompliance: IPCompliance{
					FossID:  "foss1",
					Details: []IPComplianceItem{{ReviewModel: "model1"}},
				},
				ExportControl: ExportCompliance{
					Details: []ExportComplianceItem{{Progress: "progress1"}},
				},
			}}},
			{VendorFoss: FossRelationshipTO{FossTO: FossTO{FossName: "foss2"}}},
			{VendorFoss: FossRelationshipTO{PmiTOs: []PmiTO{{PmiValue: "foss3"}}}},
			{VendorFoss: FossRelationshipTO{PmiTOs: []PmiTO{{PmiValue: "foss4:1.0"}}}},
			{VendorFoss: FossRelationshipTO{PmiTOs: []PmiTO{{PmiValue: "group:foss5:2.0"}}}},
		}
		expected := []Foss{
			{FossID: "foss1", ArtifactID: "artifact ID unknown", IPComplianceDetails: []IPComplianceItem{{ReviewModel: "model1"}}, ExportComplianceDetails: []ExportComplianceItem{{Progress: "progress1"}}},
			{ArtifactID: "foss2"},
			{ArtifactID: "foss3"},
			{ArtifactID: "foss4", Version: "1.0"},
			{ArtifactID: "foss5", Version: "2.0", GroupID: "group"},
		}
		foss := FossFromMapping(fm)
		assert.Equal(t, expected, foss)
	})
}

func TestRemoveFossMappingPairDuplicates(t *testing.T) {
	fossMappings := []APIFossMappingPair{
		{
			StandardFoss: FossRelationshipTO{
				ComplianceInfo: ComplianceInfo{
					IPCompliance: IPCompliance{FossID: "fossID1"},
					ExportControl: ExportCompliance{
						Details: []ExportComplianceItem{
							{Progress: "ECCNASSIGNED", ECCNEU: "EU", ECCNUS: "US", Crypto: false, ClassificationCompleted: true},
						},
					},
				},
			},
			VendorFoss: FossRelationshipTO{
				PmiTOs: []PmiTO{
					{PmiValue: "group2:artifact1:version1"},
				},
			},
		},
		{
			StandardFoss: FossRelationshipTO{
				ComplianceInfo: ComplianceInfo{
					IPCompliance: IPCompliance{FossID: "fossID2"},
					ExportControl: ExportCompliance{
						Details: []ExportComplianceItem{
							{Progress: "ECCNASSIGNED", ECCNEU: "EU", ECCNUS: "US", Crypto: false, ClassificationCompleted: true},
						},
					},
				},
			},
			VendorFoss: FossRelationshipTO{
				PmiTOs: []PmiTO{
					{PmiValue: "group2:artifact2:version2"},
				},
			},
		},
		{
			StandardFoss: FossRelationshipTO{
				ComplianceInfo: ComplianceInfo{
					IPCompliance: IPCompliance{FossID: "fossID1"},
					ExportControl: ExportCompliance{
						Details: []ExportComplianceItem{
							{Progress: "ECCNASSIGNED", ECCNEU: "EU", ECCNUS: "US", Crypto: false, ClassificationCompleted: true},
						},
					},
				},
			},
			VendorFoss: FossRelationshipTO{
				PmiTOs: []PmiTO{
					{PmiValue: "group2:artifact1:version1"},
				},
			},
		},
	}

	expected := []APIFossMappingPair{
		{
			StandardFoss: FossRelationshipTO{
				ComplianceInfo: ComplianceInfo{
					IPCompliance: IPCompliance{FossID: "fossID1"},
					ExportControl: ExportCompliance{
						Details: []ExportComplianceItem{
							{Progress: "ECCNASSIGNED", ECCNEU: "EU", ECCNUS: "US", Crypto: false, ClassificationCompleted: true},
						},
					},
				},
			},
			VendorFoss: FossRelationshipTO{
				PmiTOs: []PmiTO{
					{PmiValue: "group2:artifact1:version1"},
				},
			},
		},
		{
			StandardFoss: FossRelationshipTO{
				ComplianceInfo: ComplianceInfo{
					IPCompliance: IPCompliance{FossID: "fossID2"},
					ExportControl: ExportCompliance{
						Details: []ExportComplianceItem{
							{Progress: "ECCNASSIGNED", ECCNEU: "EU", ECCNUS: "US", Crypto: false, ClassificationCompleted: true},
						},
					},
				},
			},
			VendorFoss: FossRelationshipTO{
				PmiTOs: []PmiTO{
					{PmiValue: "group2:artifact2:version2"},
				},
			},
		},
	}

	assert.Equal(t, expected, removeFossMappingPairDuplicates(fossMappings))
}

func TestGetChannelRiskRating(t *testing.T) {
	tt := []struct {
		ipComplianceItems []IPComplianceItem
		channel           string
		expected          RiskRating
		expectedError     string
	}{
		{ipComplianceItems: []IPComplianceItem{}, channel: "A", expected: RiskRating{}, expectedError: "invalid GTMC channel provided: 'A'"},
		{ipComplianceItems: []IPComplianceItem{}, channel: "GTMC_09", expected: RiskRating{}, expectedError: "GTMC channel 'GTMC_09' not yet supported"},
		{ipComplianceItems: []IPComplianceItem{}, channel: "GTMC_10", expected: RiskRating{}, expectedError: "no risk rating found for channel 'GTMC_10'"},
		{
			ipComplianceItems: []IPComplianceItem{
				{InherentRiskRating: "inhRisk", ResidualRiskRating: "resRisk", ReviewModel: "E"},
			},
			channel:  "GTMC_10",
			expected: RiskRating{InherentRiskRating: "inhRisk", ResidualRiskRating: "resRisk"},
		},
	}

	for run, test := range tt {
		t.Run(fmt.Sprintf("Run %v", run), func(t *testing.T) {
			res, err := GetChannelRiskRating(test.ipComplianceItems, test.channel)
			if test.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, res)
			} else {
				assert.EqualError(t, err, test.expectedError)
			}
		})
	}
}

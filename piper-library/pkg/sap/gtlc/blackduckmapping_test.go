package gtlc

import (
	"fmt"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

func TestBlackDuckPPMSMapping(t *testing.T) {

	t.Run("Prefilled BlackDuckMapping", func(t *testing.T) {

		myTestClient := mappingMockClient{}
		sys := MappingSystem{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}

		t.Run("No mapping available", func(t *testing.T) {
			w := BlackDuckMapping{
				filled:       true,
				FossMappings: []APIFossMappingPair{},
			}
			got, err := w.BlackDuckPPMSMapping(&sys)
			assert.NoError(t, err, "Error occurred but none expected")
			t.Run("Get empty mapping", func(t *testing.T) {
				if len(got) > 0 {
					t.Errorf("Expected size 0, got size %v", len(got))
				}
			})
		})

		t.Run("Mapping available", func(t *testing.T) {
			w := BlackDuckMapping{
				filled: true,
				FossMappings: []APIFossMappingPair{
					{
						StandardFoss: FossRelationshipTO{},
						VendorFoss: FossRelationshipTO{
							PmiTOs: []PmiTO{
								{PmiValue: "artifact0"},
							},
						},
					},
					{
						StandardFoss: FossRelationshipTO{
							ComplianceInfo: ComplianceInfo{
								IPCompliance: IPCompliance{
									FossID: "fossID1",
									Details: []IPComplianceItem{
										{ReviewModel: "", InherentRiskRating: "", ResidualRiskRating: "", TechnicalInstructionsCount: 1},
									},
								},
							},
						},
						VendorFoss: FossRelationshipTO{
							PmiTOs: []PmiTO{
								{PmiValue: "artifact1:version1"},
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
				},
			}
			got, err := w.BlackDuckPPMSMapping(&sys)
			assert.NoError(t, err, "Error occurred but none expected")
			expected := []Foss{
				{ArtifactID: "artifact0"},
				{
					FossID:     "fossID1",
					GroupID:    "",
					ArtifactID: "artifact1",
					Version:    "version1",
					IPComplianceDetails: []IPComplianceItem{
						{ReviewModel: "", InherentRiskRating: "", ResidualRiskRating: "", TechnicalInstructionsCount: 1},
					},
				},
				{
					FossID:     "fossID2",
					GroupID:    "group2",
					ArtifactID: "artifact2",
					Version:    "version2",
					ExportComplianceDetails: []ExportComplianceItem{
						{Progress: "ECCNASSIGNED", ECCNEU: "EU", ECCNUS: "US", Crypto: false, ClassificationCompleted: true},
					},
				},
			}

			for k, v := range expected {
				t.Run(fmt.Sprintf("Run %v", k), func(t *testing.T) {
					assert.Equal(t, v, got[k])
				})
			}

		})
	})

	t.Run("retrieve BlackDuckMapping", func(t *testing.T) {

		mappingResponse := `[
	{"standardFoss": {"complianceInfo": {"ipCompliance": {"fossId": "foss1"}}}},
	{"standardFoss": {"complianceInfo": {"ipCompliance": {"fossId": "foss2"}}}}
]`
		t.Run("from project", func(t *testing.T) {
			myTestClient := mappingMockClient{httpStatusCode: 200, responseBody: mappingResponse}
			sys := MappingSystem{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}

			w := BlackDuckMapping{
				APIURLs: []string{"https://test.url/api/projects"},
				Filter:  []string{"SAP_IP", "SAP_ECCN"},
			}

			got, err := w.BlackDuckPPMSMapping(&sys)

			assert.NoError(t, err, "Error occurred but none expected")
			assert.Equal(t, "https://my.test.server/api/mapping/hub/project/version/bom?expand=SAP_IP,SAP_ECCN", myTestClient.urlCalled, "incorrect url called")
			assert.Equal(t, "foss1", got[0].FossID)
			assert.Equal(t, "foss2", got[1].FossID)
		})

		t.Run("no project urls", func(t *testing.T) {
			myTestClient := mappingMockClient{httpStatusCode: 200, responseBody: mappingResponse}
			sys := MappingSystem{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}
			w := BlackDuckMapping{
				Filter: []string{"SAP_IP", "SAP_ECCN"},
			}

			_, err := w.BlackDuckPPMSMapping(&sys)
			assert.EqualError(t, err, "no blackduck project urls available")
		})

		t.Run("http error", func(t *testing.T) {
			myTestClient := mappingMockClient{httpStatusCode: 500}
			sys := MappingSystem{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}
			w := BlackDuckMapping{
				APIURLs: []string{"https://test.url/api/projects"},
				Filter:  []string{"SAP_IP", "SAP_ECCN"},
			}

			_, err := w.BlackDuckPPMSMapping(&sys)
			assert.Error(t, err, "Error expected but none occurred")
			assert.Contains(t, fmt.Sprint(err), "HTTP Error")
		})

	})
}

func TestBlackDuckPPMSMappingIntegration(t *testing.T) {
	t.Skip()
	sys := MappingSystem{
		ServerURL:  "https://cmapi.gtlc.only.sap",
		HTTPClient: &piperhttp.Client{},
	}
	w := BlackDuckMapping{
		APIURLs: []string{"https://sap.blackducksoftware.com/api/projects/5ca86e11-1983-4e7b-97d4-eb1a0aeffbbf/versions/a6c94786-0ee6-414f-9054-90d549c69c36"},
		Filter:  []string{"SAP_IP", "SAP_ECCN"},
	}
	got, err := w.BlackDuckPPMSMapping(&sys)
	if err != nil {
		t.Log(got)
	}
	assert.NoError(t, err)
	assert.True(t, len(got) > 0)
}

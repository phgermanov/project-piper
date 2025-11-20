package gtlc

import (
	"fmt"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

func TestWhiteSourcePPMSMapping(t *testing.T) {

	t.Run("Prefilled WhiteSourceMapping", func(t *testing.T) {

		myTestClient := mappingMockClient{}
		sys := MappingSystem{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}

		t.Run("No mapping available", func(t *testing.T) {
			w := WhiteSourceMapping{
				filled:       true,
				FossMappings: []APIFossMappingPair{},
			}
			got, err := w.WhiteSourcePPMSMapping(&sys)
			assert.NoError(t, err, "Error occurred but none expected")
			t.Run("Get empty mapping", func(t *testing.T) {
				if len(got) > 0 {
					t.Errorf("Expected size 0, got size %v", len(got))
				}
			})
		})

		t.Run("Mapping available", func(t *testing.T) {
			w := WhiteSourceMapping{
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
			got, err := w.WhiteSourcePPMSMapping(&sys)
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

	t.Run("retrieve WhiteSourceMapping", func(t *testing.T) {

		mappingResponse := `[
{"standardFoss": {"complianceInfo": {"ipCompliance": {"fossId": "foss1"}}}},
{"standardFoss": {"complianceInfo": {"ipCompliance": {"fossId": "foss2"}}}}
]`

		t.Run("from project", func(t *testing.T) {
			myTestClient := mappingMockClient{httpStatusCode: 200, responseBody: mappingResponse}
			sys := MappingSystem{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}

			w := WhiteSourceMapping{
				ProjectTokens: []string{"testProjectToken"},
				Filter:        []string{"SAP_IP", "SAP_ECCN"},
			}

			got, err := w.WhiteSourcePPMSMapping(&sys)

			assert.NoError(t, err, "Error occurred but none expected")
			assert.Equal(t, "https://my.test.server/api/mapping/whitesource/projects/testProjectToken/bom?expand=SAP_IP,SAP_ECCN", myTestClient.urlCalled, "incorrect url called")
			assert.Equal(t, "foss1", got[0].FossID)
			assert.Equal(t, "foss2", got[1].FossID)
		})

		t.Run("from product", func(t *testing.T) {
			myTestClient := mappingMockClient{httpStatusCode: 200, responseBody: mappingResponse}
			sys := MappingSystem{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}

			w := WhiteSourceMapping{
				ProductToken: "testProductToken",
				Filter:       []string{"SAP_IP", "SAP_ECCN"},
			}

			got, err := w.WhiteSourcePPMSMapping(&sys)

			assert.NoError(t, err, "Error occurred but none expected")
			assert.Equal(t, "https://my.test.server/api/mapping/whitesource/products/testProductToken/bom?expand=SAP_IP,SAP_ECCN", myTestClient.urlCalled, "incorrect url called")
			assert.Equal(t, "foss1", got[0].FossID)
			assert.Equal(t, "foss2", got[1].FossID)
		})

		t.Run("no token", func(t *testing.T) {
			myTestClient := mappingMockClient{httpStatusCode: 200, responseBody: mappingResponse}
			sys := MappingSystem{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}
			w := WhiteSourceMapping{
				Filter: []string{"SAP_IP", "SAP_ECCN"},
			}

			_, err := w.WhiteSourcePPMSMapping(&sys)
			assert.EqualError(t, err, "Error when connecting to Mapping system: Neither project token nor product token available")
		})

		t.Run("http error", func(t *testing.T) {
			myTestClient := mappingMockClient{httpStatusCode: 500}
			sys := MappingSystem{ServerURL: "https://my.test.server", HTTPClient: &myTestClient}
			w := WhiteSourceMapping{
				ProjectTokens: []string{"testProjectToken"},
				Filter:        []string{"SAP_IP", "SAP_ECCN"},
			}

			_, err := w.WhiteSourcePPMSMapping(&sys)
			assert.Error(t, err, "Error expected but none occurred")
			assert.NotEqual(t, "Neither project token nor product token available", err.Error())
		})

	})
}

func TestWhiteSourcePPMSMappingIntegration(t *testing.T) {
	t.Skip()
	sys := MappingSystem{
		ServerURL:  "https://cmapi.gtlc.only.sap",
		HTTPClient: &piperhttp.Client{},
	}
	w := WhiteSourceMapping{
		ProjectTokens: []string{"a8f90c3dee2a434e8636fcfe7746a589d411cf4406124b4c8069ca035dac8a58"},
		Filter:        []string{"SAP_IP", "SAP_ECCN"},
	}
	got, err := w.WhiteSourcePPMSMapping(&sys)
	if err != nil {
		t.Log(got)
	}
	assert.NoError(t, err)
	assert.True(t, len(got) > 0)
}

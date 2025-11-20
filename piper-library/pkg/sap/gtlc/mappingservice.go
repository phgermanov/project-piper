package gtlc

import (
	"fmt"
	"io"
	"strings"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

// Test tokens:
// Java: 91bbf11eb5be43d28a5f32e9d9968f191dd257d66794445b82b42fc5702104db

// MappingSystem defines access information to the system running the mapping service
// Detailed information can be found here
// * https://github.wdf.sap.corp/pages/GTLCDEV-COMPONENTMAPPING/ComponentMapping/
// * https://cmapi.gtlc.only.sap/swagger-ui.html
type MappingSystem struct {
	ServerURL  string
	Username   string
	Password   string
	HTTPClient piperhttp.Sender
}

// APIFossMappingPair defines a mapping pair bringing source information and SAP information together
type APIFossMappingPair struct {
	StandardFoss FossRelationshipTO `json:"standardFoss"`
	VendorFoss   FossRelationshipTO `json:"vendorFoss"`
}

// FossRelationshipTO defines compliance info and mappings for a Foss object
type FossRelationshipTO struct {
	ComplianceInfo ComplianceInfo `json:"complianceInfo"`
	FossTO         FossTO         `json:"fossTo"`
	PmiTOs         []PmiTO        `json:"pmiTOs"`
}

// ComplianceInfo defines compliance information of a Foss object
type ComplianceInfo struct {
	ExportControl ExportCompliance   `json:"exportControl"`
	IPCompliance  IPCompliance       `json:"ipCompliance"`
	SecurityVuln  AbstractCompliance `json:"securityVuln"`
}

// FossTO defines a Foss object
type FossTO struct {
	AdditionalVendorInfo AdditionalVendorInfo `json:"additionalVendorInfo"`
	DeclaredLicense      License              `json:"declaredLicense"`
	FossName             string               `json:"fossName"`
	FossVersion          string               `json:"fossVersion"`
	Homepage             string               `json:"homepage"`
	Licenses             []License            `json:"licenses"`
	VendorIdentifier     VendorIdentifier     `json:"vendorIdentifier"`
}

// PmiTO defines package management information of a component
type PmiTO struct {
	PmiType        string `json:"pmiType"`
	PmiValue       string `json:"pmiValue"`
	VendorPmiValue string `json:"vendorPmiValue"`
}

// AbstractCompliance defines a general purpose compliance information
type AbstractCompliance struct {
	Details  []map[string]interface{} `json:"details"`
	FossID   string                   `json:"fossId"`
	Provider string                   `json:"provider"`
}

// IPCompliance defines IP compliance information
type IPCompliance struct {
	Details  []IPComplianceItem `json:"details"`
	FossID   string             `json:"fossId"`
	Provider string             `json:"provider"`
}

// IPComplianceItem defines a concrete IP compliance item
type IPComplianceItem struct {
	ReviewModel                string `json:"reviewModel"`
	InherentRiskRating         string `json:"inherentRiskRating"`
	ResidualRiskRating         string `json:"residualRiskRating"`
	TechnicalInstructionsCount int    `json:"technicalInstructionsCount"`
}

// RiskRating contains the different aspects of risk ratings (inherent, residual)
type RiskRating struct {
	InherentRiskRating string `json:"inherentRiskRating"`
	ResidualRiskRating string `json:"residualRiskRating"`
}

// ExportCompliance defines export compliance information
type ExportCompliance struct {
	Details  []ExportComplianceItem `json:"details"`
	FossID   string                 `json:"fossId"`
	Provider string                 `json:"provider"`
}

// ExportComplianceItem defines a concrete export compliance item
type ExportComplianceItem struct {
	Progress                string `json:"progress"`
	ECCNEU                  string `json:"eccnEU"`
	ECCNUS                  string `json:"eccnUS"`
	Crypto                  bool   `json:"crypto"`
	ClassificationCompleted bool   `json:"classificationCompleted"`
}

// AdditionalVendorInfo defines additional vendor information for an object
type AdditionalVendorInfo struct {
	Description string   `json:"description"`
	Homepages   []string `json:"homepages"`
}

// License defines a concrete license
type License struct {
	Customized bool   `json:"customized"`
	LicenseID  string `json:"licenseId"`
	Name       string `json:"name"`
}

// VendorIdentifier definces the source of the information
type VendorIdentifier struct {
	FossVendor           string   `json:"fossVendor"` //possible values: [CODECENTER, HUB, WHITESOURCE, SCA]
	PersistentAttributes []string `json:"persistentAttributes"`
}

// Foss defines FOSS objects with their artifact references, IP compliance details and export control details
type Foss struct {
	FossID                  string
	GroupID                 string
	ArtifactID              string
	Version                 string
	IPComplianceDetails     []IPComplianceItem
	ExportComplianceDetails []ExportComplianceItem
}

// FossFromMapping returns all FOSS dependencies including further details if available.
func FossFromMapping(fm []APIFossMappingPair) []Foss {

	var fs []Foss

	for _, m := range fm {
		var f Foss
		if len(m.StandardFoss.ComplianceInfo.IPCompliance.FossID) > 0 {
			f.FossID = m.StandardFoss.ComplianceInfo.IPCompliance.FossID
		}

		if len(m.VendorFoss.PmiTOs) > 0 {
			pp := strings.Split(m.VendorFoss.PmiTOs[0].PmiValue, ":")
			switch len(pp) {
			case 2:
				f.ArtifactID = pp[0]
				f.Version = pp[1]
			case 3:
				f.GroupID = pp[0]
				f.ArtifactID = pp[1]
				f.Version = pp[2]
			default:
				f.ArtifactID = m.VendorFoss.PmiTOs[0].PmiValue
			}
		} else if len(m.VendorFoss.FossTO.FossName) > 0 {
			f.ArtifactID = m.VendorFoss.FossTO.FossName
		} else {
			f.ArtifactID = "artifact ID unknown"
		}

		if len(m.StandardFoss.ComplianceInfo.IPCompliance.Details) > 0 {
			f.IPComplianceDetails = m.StandardFoss.ComplianceInfo.IPCompliance.Details
		}
		if len(m.StandardFoss.ComplianceInfo.ExportControl.Details) > 0 {
			f.ExportComplianceDetails = m.StandardFoss.ComplianceInfo.ExportControl.Details
		}
		fs = append(fs, f)
	}
	return fs
}

func removeFossMappingPairDuplicates(mappingPairs []APIFossMappingPair) []APIFossMappingPair {
	res := []APIFossMappingPair{}
	for _, pair := range mappingPairs {
		if !containsFossMappingPair(&res, pair) {
			res = append(res, pair)
		}
	}
	return res
}

func containsFossMappingPair(mappingPairs *[]APIFossMappingPair, contains APIFossMappingPair) bool {
	for _, pair := range *mappingPairs {
		if cmp.Equal(contains, pair) {
			return true
		}
	}
	return false
}

func (sys *MappingSystem) sendRequest(method, path, body string) ([]byte, error) {

	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = strings.NewReader(body)
	}
	response, err := sys.HTTPClient.SendRequest(method, fmt.Sprintf("%v%v", sys.ServerURL, path), bodyReader, nil, nil)
	if err != nil {
		return []byte{}, errors.Wrapf(err, "failed to send request to path %v", path)
	}

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return []byte{}, errors.Wrap(err, "error reading response")
	}
	response.Body.Close()

	return content, nil
}

//Mapping of review models to go-to-market channels can be found here: https://wiki.wdf.sap.corp/wiki/display/ppmscont/FOSS+-+Review+Model
//as of July 15th, 2019:
//
//
// | GTMC key |GTMC short name        |GTMC long name                                                      | Review models|
// |----------|-----------------------|--------------------------------------------------------------------|--------------|
// | GTMC_01  | Direct OnPrem         | Direct On-premise                                                  | A            |
// | GTMC_02  | VAR OnPrem            | Value Added Reseller (VAR)                                         | A            |
// | GTMC_03  | Softw Subscr          | Software Subscription                                              | A            |
// | GTMC_04  | ASRunt OnPrem         | Application Specific Runtime                                       | A            |
// | GTMC_05  | oOEM OnPrem           | OOEM On-premise Embedded                                           | A            |
// | GTMC_06  | Public PMCloud        | OOEM / PMC [public cloud] (Partner On-premise, Customer On-demand) | A + B        |
// | GTMC_07  | Chann SaaS            | Channel Cloud SaaS                                                 | E            |
// | GTMC_08  | Chann P/IaaS          | Channel Cloud PaaS & IaaS                                          | E            |
// | GTMC_09  | Privat PMCloud        | PMC/ BPO [private cloud] (Partner On-premise, Customer On-demand)  | A + B        |
// | GTMC_10  | Direct Cloud          | Cloud Direct (SaaS, PaaS & IaaS)                                   | E            |
// | GTMC_11  | Mobile Pure Licensing | Mobile Pure Licensing                                              | C + D        |
// | GTMC_12  | OOEM Mobile Store     | OOEM Mobile Store                                                  | C + D        |

// GetChannelRiskRating returns the risk rating of a FOSS for a dedicated go-to-market channel
func GetChannelRiskRating(ipComplianceItems []IPComplianceItem, channelKey string) (RiskRating, error) {
	res := RiskRating{}

	if !strings.HasPrefix(channelKey, "GTMC_") {
		return res, fmt.Errorf("invalid GTMC channel provided: '%v'", channelKey)
	}

	// currently only Direct Cloud channel (GTMC_10) supported.
	if !sliceContains([]string{"GTMC_07", "GTMC_08", "GTMC_10"}, channelKey) {
		return res, fmt.Errorf("GTMC channel '%v' not yet supported", channelKey)
	}

	for _, detail := range ipComplianceItems {
		if detail.ReviewModel == "E" {
			res.InherentRiskRating = detail.InherentRiskRating
			res.ResidualRiskRating = detail.ResidualRiskRating
			return res, nil
		}
	}

	return res, fmt.Errorf("no risk rating found for channel '%v'", channelKey)
}

func sliceContains(slice []string, find string) bool {
	for _, elem := range slice {
		if elem == find {
			return true
		}
	}
	return false
}

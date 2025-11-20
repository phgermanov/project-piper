package gtlc

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// BlackDuckMapping defines mapping information for BlackDuck
type BlackDuckMapping struct {
	// APIURL is the full path for BlackDuck HUB to find the project details for a specific version.
	// e.g. https://sap.blackducksoftware.com/api/projects/d8d8830e-50b4-4f99-a05c-6dbf9a3304c7/versions/f00b1bcd-78c1-4f28-b3ea-0edae0c0ab02
	APIURLs      []string //
	Filter       []string //possible values: [Vendor_Policy, Vendor_SecVuln, SAP_IP, SAP_ECCN, SAP_SecVuln]
	filled       bool
	FossMappings []APIFossMappingPair
}

// BlackDuckPPMSMapping retrieves all FOSS dependencies identified by BlackDuck including further details if available.
func (b *BlackDuckMapping) BlackDuckPPMSMapping(sys *MappingSystem) ([]Foss, error) {
	if !b.filled {
		if len(b.APIURLs) == 0 {
			return nil, fmt.Errorf("no blackduck project urls available")
		}
		allMappings := []APIFossMappingPair{}
		for _, apiURL := range b.APIURLs {
			currentMappings := []APIFossMappingPair{}
			content, err := sys.sendRequest("POST", fmt.Sprintf("/api/mapping/hub/project/version/bom?expand=%v", strings.Join(b.Filter, ",")), fmt.Sprintf(`{"apiUrl":"%v"}`, apiURL))
			if err != nil {
				return nil, errors.Wrapf(err, "failed to retrieve project mapping for project %v", apiURL)
			}
			json.Unmarshal(content, &currentMappings)
			allMappings = append(allMappings, currentMappings...)
		}

		b.FossMappings = removeFossMappingPairDuplicates(allMappings)
		b.filled = true
	}
	return FossFromMapping(b.FossMappings), nil
}

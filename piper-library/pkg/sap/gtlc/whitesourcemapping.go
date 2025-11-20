package gtlc

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// WhiteSourceMapping defines mapping information for WhiteSource
type WhiteSourceMapping struct {
	Filter        []string //possible values: [Vendor_Policy, Vendor_SecVuln, SAP_IP, SAP_ECCN, SAP_SecVuln]
	ProjectTokens []string
	ProductToken  string
	filled        bool
	FossMappings  []APIFossMappingPair
}

// WhiteSourcePPMSMapping retrieves all FOSS dependencies identified by WhiteSource including further details if available.
func (w *WhiteSourceMapping) WhiteSourcePPMSMapping(sys *MappingSystem) ([]Foss, error) {
	if !w.filled {
		var err error
		if len(w.ProjectTokens) > 0 {
			err = w.whitesourceMappingByProjects(sys)
		} else if len(w.ProductToken) > 0 {
			err = w.whitesourceMappingByProduct(sys)
		} else {
			err = fmt.Errorf("Neither project token nor product token available")
		}
		if err != nil {
			return nil, errors.Wrap(err, "Error when connecting to Mapping system")
		}
	}
	return FossFromMapping(w.FossMappings), nil
}

func (w *WhiteSourceMapping) whitesourceMappingByProjects(sys *MappingSystem) error {

	allMappings := []APIFossMappingPair{}
	for _, projectToken := range w.ProjectTokens {
		currentMappings := []APIFossMappingPair{}
		content, err := sys.sendRequest("GET", fmt.Sprintf("/api/mapping/whitesource/projects/%v/bom?expand=%v", projectToken, strings.Join(w.Filter, ",")), "")
		if err != nil {
			return errors.Wrap(err, "failed to retrieve project mapping")
		}
		json.Unmarshal(content, &currentMappings)
		allMappings = append(allMappings, currentMappings...)
	}

	w.FossMappings = removeFossMappingPairDuplicates(allMappings)

	w.filled = true
	return nil
}

func (w *WhiteSourceMapping) whitesourceMappingByProduct(sys *MappingSystem) error {
	content, err := sys.sendRequest("GET", fmt.Sprintf("/api/mapping/whitesource/products/%v/bom?expand=%v", w.ProductToken, strings.Join(w.Filter, ",")), "")
	if err != nil {
		return errors.Wrap(err, "failed to retrieve project mapping")
	}
	json.Unmarshal(content, &w.FossMappings)
	w.filled = true
	return nil
}

package report

import (
	"encoding/xml"
	"os"

	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/policy"
)

type PolicyReporterIO interface {
	FileExists(filename string) (bool, error)
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, content []byte, perm os.FileMode) error
}

type junitTestSuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Name      string          `xml:"name,attr"`
	Time      float32         `xml:"time,attr"`
	Tests     int             `xml:"tests,attr"`
	Errors    int             `xml:"errors,attr"`
	Skipped   int             `xml:"skipped,attr"`
	Failures  int             `xml:"failures,attr"`
	TestCases []junitTestCase `xml:"testcase"`
}

type junitTestCase struct {
	XMLName   xml.Name `xml:"testcase"`
	Name      string   `xml:"name,attr"`
	ClassName string   `xml:"classname,attr"`
	Time      float32  `xml:"time,attr"`
	Failure   string   `xml:"failure,omitempty"`
}

type policyMetadata struct {
	Key   string `json:"key" validate:"required"`
	Label string `json:"label" validate:"required"`
}

type policyResult struct {
	Policy                  policyMetadata                `json:"policy" validate:"required"`
	ComplianceStatus        policy.PolicyComplianceStatus `json:"complianceStatus" validate:"required"`
	ValidationErrorMessages []string                      `json:"validationErrorMessages" validate:"required"`
}

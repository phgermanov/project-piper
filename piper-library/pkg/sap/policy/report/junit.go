package report

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/policy"
)

type PolicyJunitReporter interface {
	ReportPolicyResult() error
}

type policyJunitReporter struct {
	policyType           policy.PolicyType
	policyResultFilePath string
	io                   PolicyReporterIO
}

func NewPolicyJunitReporter(policyType policy.PolicyType, policyResultFilePath string, io PolicyReporterIO) PolicyJunitReporter {
	return &policyJunitReporter{policyType: policyType, policyResultFilePath: policyResultFilePath, io: io}
}

func (p *policyJunitReporter) ReportPolicyResult() error {
	result, err := p.getPolicyResult(p.policyResultFilePath)
	if err != nil {
		return err
	}

	xml, err := p.createPolicyJunitXML(result)
	if err != nil {
		return err
	}
	return p.io.WriteFile(fmt.Sprintf("TEST-%s-policy-%s.xml", p.policyType, result.Policy.Key), xml, 0644)
}

func (p *policyJunitReporter) getPolicyResult(resultFilePath string) (policyResult, error) {
	result := policyResult{}
	exists, err := p.io.FileExists(resultFilePath)
	if err != nil {
		return result, err
	}
	if !exists {
		return result, fmt.Errorf("policy result file %s not found", resultFilePath)
	}
	bytes, err := p.io.ReadFile(resultFilePath)
	if err != nil {
		return result, nil
	}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (p *policyJunitReporter) createPolicyJunitXML(result policyResult) ([]byte, error) {

	testsuitename := fmt.Sprintf("%v Policy", strings.ToUpper(string(p.policyType)[:1])+string(p.policyType)[1:])

	testsuite := junitTestSuite{Name: testsuitename, Tests: 0, Time: 0.0, Errors: 0, Skipped: 0, Failures: 0}
	testcase := junitTestCase{Name: result.Policy.Key, ClassName: result.Policy.Label}
	if result.ComplianceStatus == policy.NotCompliant {
		testcase.Failure = strings.Join(result.ValidationErrorMessages, "\n")
		testsuite.Failures = testsuite.Failures + 1
	}

	testsuite.TestCases = append(testsuite.TestCases, testcase)
	testsuite.Tests = testsuite.Tests + 1

	return xml.MarshalIndent(testsuite, "", "  ")
}

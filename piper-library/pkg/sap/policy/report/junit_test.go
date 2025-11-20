package report

import (
	"fmt"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/sap/policy"
)

func TestCreateJunitXML(t *testing.T) {
	t.Parallel()
	t.Run("Can convert a non compliant policy result to XML", func(t *testing.T) {
		t.Parallel()

		reporter := policyJunitReporter{policyType: policy.Custom, policyResultFilePath: "test.json", io: &mock.FilesMock{}}

		policy := policyMetadata{Key: "MY-POLICY-1", Label: "This is my custom policy"}
		result := policyResult{Policy: policy, ComplianceStatus: "NOT_COMPLIANT", ValidationErrorMessages: []string{"This is a validation error"}}

		xml, err := reporter.createPolicyJunitXML(result)

		assert.NotNil(t, xml)
		assert.NoError(t, err)

		expected :=
			`<testsuite name="Custom Policy" time="0" tests="1" errors="0" skipped="0" failures="1">
				<testcase name="MY-POLICY-1" classname="This is my custom policy" time="0">
					<failure>This is a validation error</failure>
				</testcase>
			</testsuite>`

		assert.Equal(t, normalize(expected), normalize(string(xml)))

	})
	t.Run("Can convert a compliant policy result to XML", func(t *testing.T) {
		t.Parallel()

		reporter := policyJunitReporter{policyType: policy.Custom, policyResultFilePath: "test.json", io: &mock.FilesMock{}}

		policy := policyMetadata{Key: "MY-POLICY-1", Label: "This is my custom policy"}
		result := policyResult{Policy: policy, ComplianceStatus: "COMPLIANT"}

		xml, err := reporter.createPolicyJunitXML(result)

		assert.NotNil(t, xml)
		assert.NoError(t, err)

		expected :=
			`<testsuite name="Custom Policy" time="0" tests="1" errors="0" skipped="0" failures="0">
				<testcase name="MY-POLICY-1" classname="This is my custom policy" time="0">
				</testcase>
			</testsuite>`

		assert.Equal(t, normalize(expected), normalize(string(xml)))

	})

	t.Run("Can not report if the result file does not exist", func(t *testing.T) {
		t.Parallel()

		reporter := policyJunitReporter{policyType: policy.Custom, policyResultFilePath: "test.json", io: &mock.FilesMock{}}

		err := reporter.ReportPolicyResult()

		assert.Error(t, err)
	})

	t.Run("Can report if the result file exist", func(t *testing.T) {
		t.Parallel()

		io := &mock.FilesMock{}

		io.AddFile("test.json", []byte(`{"policy": {"key": "MY-POLICY-1", "label": "This is my custom policy"}, "complianceStatus": "COMPLIANT", "validationErrorMessages": []}`))

		reporter := policyJunitReporter{policyType: policy.Custom, policyResultFilePath: "test.json", io: io}

		err := reporter.ReportPolicyResult()

		assert.NoError(t, err)

		assert.True(t, io.HasWrittenFile(fmt.Sprintf("TEST-%s-policy-%s.xml", policy.Custom, "MY-POLICY-1")))

	})

	t.Run("Can not report if the result file is invalid json", func(t *testing.T) {
		t.Parallel()

		io := &mock.FilesMock{}

		io.AddFile("test.json", []byte(`{"policy": {"key": `))

		reporter := policyJunitReporter{policyType: policy.Custom, policyResultFilePath: "test.json", io: io}

		err := reporter.ReportPolicyResult()

		assert.Error(t, err)

	})

	t.Run("Can create new policy junit reporter", func(t *testing.T) {
		t.Parallel()

		reporter := NewPolicyJunitReporter(policy.Custom, "test.json", &mock.FilesMock{})

		assert.NotNil(t, reporter)

	})

}

func normalize(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), ""))
}

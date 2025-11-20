package policy

type (
	PolicyComplianceStatus string
	PolicyType             string
)

const (
	Custom       PolicyType             = "custom"
	Central      PolicyType             = "central"
	Compliant    PolicyComplianceStatus = "COMPLIANT"
	NotCompliant PolicyComplianceStatus = "NOT_COMPLIANT"
)

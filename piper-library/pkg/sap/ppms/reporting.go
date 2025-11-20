package ppms

import (
	"fmt"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	"github.com/SAP/jenkins-library/pkg/reporting"
)

// UploadReportToGithub creates/updates a GitHub issue with PPMS check results
// ToDo: functionality needs to move to reporting package in Piper Open Source repository as reuse functionality
func UploadReportToGithub(scanReport reporting.ScanReport, token, APIURL, owner, repository string, assignees []string) error {
	// JSON reports are used by step pipelineCreateSummary in order to e.g. prepare an issue creation in GitHub
	// ignore JSON errors since structure is in our hands
	markdownReport, _ := scanReport.ToMarkdown()
	options := piperGithub.CreateIssueOptions{
		Token:          token,
		APIURL:         APIURL,
		Owner:          owner,
		Repository:     repository,
		Title:          "SAP PPMS Check Results",
		Body:           markdownReport,
		Assignees:      assignees,
		UpdateExisting: true,
	}
	_, err := piperGithub.CreateIssue(&options)
	if err != nil {
		return fmt.Errorf("failed to upload PPMS check results into GitHub issue: %w", err)
	}
	return nil
}

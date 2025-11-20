package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/google/go-github/v68/github"
)

type githubClientWrapper struct {
	client *github.Client
}

func (gcw *githubClientWrapper) GetRepo(
	ctx context.Context, owner, repo string,
) (*github.Repository, *github.Response, error) {
	return gcw.client.Repositories.Get(ctx, owner, repo)
}

func (gcw *githubClientWrapper) ListAlertsForRepo(
	ctx context.Context,
	owner, repo string,
	opts *github.SecretScanningAlertListOptions,
) ([]*github.SecretScanningAlert, *github.Response, error) {
	return gcw.client.SecretScanning.ListAlertsForRepo(ctx, owner, repo, opts)
}

type sapGithubSecretScanningReportUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)
	Create(name string) (io.ReadWriteCloser, error)

	GetRepo(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error)
	ListAlertsForRepo(
		ctx context.Context,
		owner, repo string,
		opts *github.SecretScanningAlertListOptions,
	) ([]*github.SecretScanningAlert, *github.Response, error)
}

type sapGithubSecretScanningReportUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*githubClientWrapper
	// Embed more structs as necessary to implement methods or interfaces you add to sapGithubSecretScanningReportUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// sapGithubSecretScanningReportUtilsBundle and forward to the implementation of the dependency.
}

func newSapGithubSecretScanningReportUtils(
	ghClient *github.Client,
) sapGithubSecretScanningReportUtils {
	utils := sapGithubSecretScanningReportUtilsBundle{
		Command:             &command.Command{},
		Files:               &piperutils.Files{},
		githubClientWrapper: &githubClientWrapper{ghClient},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func sapGithubSecretScanningReport(
	config sapGithubSecretScanningReportOptions,
	telemetryData *telemetry.CustomData,
) {
	apiUrl := ""
	if len(config.GithubServerURL) > 0 && len(config.GithubAPIVersion) > 0 {
		if strings.HasPrefix(config.GithubServerURL, "https") {
			// in github actions the cpe value already contains https but not in jenkins
			apiUrl = fmt.Sprintf("%s/api/%s", config.GithubServerURL, config.GithubAPIVersion)
		} else {
			apiUrl = fmt.Sprintf("https://%s/api/%s", config.GithubServerURL, config.GithubAPIVersion)
		}
	} else {
		log.Entry().Fatal("Failed to get github server url / github api version")
	}
	ctx, ghClient, err := piperGithub.NewClientBuilder(config.Token, apiUrl).Build()
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to get github client")
	}

	utils := newSapGithubSecretScanningReportUtils(ghClient)

	err = runSapGithubSecretScanningReport(ctx, &config, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runSapGithubSecretScanningReport(
	ctx context.Context,
	config *sapGithubSecretScanningReportOptions,
	utils sapGithubSecretScanningReportUtils,
) error {
	report, err := generateGithubSecretScanningReport(ctx, config, utils)
	if err != nil {
		return fmt.Errorf("could not generate the secret scanning report :%w", err)
	}

	reportFile, err := utils.Create("github-secretscanning-report.json")
	if err != nil {
		return fmt.Errorf("could not create json report for secret scanning : %w", err)
	}
	defer reportFile.Close()

	if err = json.NewEncoder(reportFile).Encode(report); err != nil {
		return fmt.Errorf("couldn't save the github secret scanning report: %w", err)
	}

	if report.Findings[0].Audited < report.Findings[0].Total {
		return fmt.Errorf(
			"COMPLIANCE VIOLATION : there are unaudited secret alerts identified in the repo. please evaluate alerts . visit https://docs.github.com/en/enterprise-server@3.13/code-security/secret-scanning/managing-alerts-from-secret-scanning/evaluating-alerts  for more details ",
		)
	}

	return nil
}

// generateGithubSecretScanningReport generates a secret scanning report for a specified GitHub repository.
// It retrieves all open secret scanning alerts, compiles their details including secret type, state,
// and locations, and returns a structured report. If any error occurs during the retrieval process,
// the function returns an error.
func generateGithubSecretScanningReport(
	ctx context.Context,
	config *sapGithubSecretScanningReportOptions,
	utils sapGithubSecretScanningReportUtils,
) (*githubSecretScanningReportType, error) {
	repo, _, err := utils.GetRepo(ctx, config.Owner, config.Repository)
	if err != nil {
		log.Entry().Debugf("Unable to get github repo for %s:%s", config.Owner, config.Repository)
		return nil, err
	}

	optHigConfidenceAlerts := &github.SecretScanningAlertListOptions{
		ListOptions: github.ListOptions{PerPage: 30, Page: 1},
	}

	var allHighConfidenceAlerts []*github.SecretScanningAlert

	for {
		// query github for high confidence alerts
		secretHighConfidenceAlerts, resp, err := utils.ListAlertsForRepo(
			ctx,
			config.Owner,
			config.Repository,
			optHigConfidenceAlerts,
		)
		if err != nil {
			log.Entry().
				Debugf("Unable to list high confidence alerts for github repo %s:%s", config.Owner, config.Repository)
			return nil, err
		}

		allHighConfidenceAlerts = append(allHighConfidenceAlerts, secretHighConfidenceAlerts...)
		log.Entry().Infof(
			"Fetched page %d of high confidence alerts (total so alerts far: %d)\n",
			optHigConfidenceAlerts.ListOptions.Page,
			len(allHighConfidenceAlerts),
		)

		if resp.NextPage == 0 {
			break
		}
		optHigConfidenceAlerts.ListOptions.Page = resp.NextPage
	}

	optOtherAlerts := &github.SecretScanningAlertListOptions{
		ListOptions: github.ListOptions{PerPage: 30, Page: 1},
		SecretType:  "http_basic_authentication_header,http_bearer_authentication_header,mongodb_connection_string,mysql_connection_string,openssh_private_key,pgp_private_key,postgres_connection_string,rsa_private_key",
	}

	var allOtherAlerts []*github.SecretScanningAlert

	for {
		// query github for other secret alerts
		secretOtherAlerts, resp, err := utils.ListAlertsForRepo(
			ctx,
			config.Owner,
			config.Repository,
			optOtherAlerts,
		)
		if err != nil {
			log.Entry().
				Debugf("Unable to list other secret alerts for github repo %s:%s", config.Owner, config.Repository)
			return nil, err
		}

		allOtherAlerts = append(allOtherAlerts, secretOtherAlerts...)
		log.Entry().Infof(
			"Fetched page %d of other alerts (total alerts so far: %d)\n",
			optOtherAlerts.ListOptions.Page,
			len(allOtherAlerts),
		)

		if resp.NextPage == 0 {
			break
		}
		optOtherAlerts.ListOptions.Page = resp.NextPage

	}

	secretScanningAlerts := append(allHighConfidenceAlerts, allOtherAlerts...)

	alertsTotal := len(secretScanningAlerts)
	alertsAudited := 0

	// query actual finding locations
	for _, alert := range secretScanningAlerts {
		if alert.State != nil && *alert.State == "resolved" {
			alertsAudited = alertsAudited + 1
		}
	}

	report := &githubSecretScanningReportType{
		ToolName: "GitHubSecretScanning",
		Findings: []githubSecretScanningFinding{
			{
				ClassificationName: "Audit All",
				Total:              alertsTotal,
				Audited:            alertsAudited,
			},
		},
	}

	if repo.HTMLURL != nil {
		report.RepositoryURL = *repo.HTMLURL
		report.SecretScanningURL = fmt.Sprintf("%s/security/secret-scanning", *repo.HTMLURL)
	}

	return report, nil
}

type githubSecretScanningReportType struct {
	ToolName          string                        `json:"toolName"`
	RepositoryURL     string                        `json:"repositoryUrl,omitempty"`
	SecretScanningURL string                        `json:"secretScanningUrl,omitempty"`
	Findings          []githubSecretScanningFinding `json:"findings"`
}

type githubSecretScanningFinding struct {
	ClassificationName string `json:"classificationName"`
	Total              int    `json:"total"`
	Audited            int    `json:"audited"`
}

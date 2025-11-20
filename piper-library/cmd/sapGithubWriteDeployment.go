package cmd

import (
	"context"
	"strconv"
	"time"

	"errors"

	piperGithub "github.com/SAP/jenkins-library/pkg/github"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/google/go-github/v68/github"

	gh "github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/github"
)

func sapGithubWriteDeployment(config sapGithubWriteDeploymentOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *sapGithubWriteDeploymentCommonPipelineEnvironment) {
	ctx, client, err := piperGithub.NewClientBuilder(config.Token, config.APIURL).
		WithTimeout(time.Duration(config.GithubAPITimeout) * time.Second).
		Build()
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to get GitHub client.")
	}

	err = runSapGithubWriteDeployment(ctx, &config, commonPipelineEnvironment, client.Repositories)
	if err != nil {
		log.Entry().WithError(err).Fatal("Failed to write deployment information.")
	}
}

var (
	ErrNoEnvironment          = errors.New("no github environment provided")
	ErrNoComitish             = errors.New("no committish provided")
	ErrCreateDeployment       = errors.New("unable to create deployment")
	ErrNoDeploymentID         = errors.New("no deploymentID provided")
	ErrParseDeploymentID      = errors.New("unable to parse deployment id")
	ErrCreateDeploymentStatus = errors.New("unable to create deployment status")
	ErrInvalidStatus          = errors.New("invalid deployment status")
)

var validDeploymentStatuses = map[string]bool{
	"error":       true,
	"failure":     true,
	"inactive":    true,
	"in_progress": true,
	"queued":      true,
	"pending":     true,
	"success":     true,
}

func validateDeploymentStatus(status string) error {
	if !validDeploymentStatuses[status] {
		return ErrInvalidStatus
	}
	return nil
}

func runSapGithubWriteDeployment(
	ctx context.Context,
	config *sapGithubWriteDeploymentOptions,
	commonPipelineEnvironment *sapGithubWriteDeploymentCommonPipelineEnvironment,
	repoClient gh.GithubRepositoryClient,
) error {
	deploymentIDStr := config.DeploymentID
	if config.CreateNew {
		if len(config.Environment) == 0 {
			return ErrNoEnvironment
		}

		if len(config.Comitish) == 0 {
			return ErrNoComitish
		}

		req := &github.DeploymentRequest{
			Ref:                   &config.Comitish,
			Environment:           &config.Environment,
			ProductionEnvironment: &config.IsProduction,
			AutoMerge:             github.Ptr(false),
			RequiredContexts:      github.Ptr([]string{}),
		}

		deployment, _, err := repoClient.CreateDeployment(ctx, config.Owner, config.Repository, req)
		if err != nil {
			log.SetErrorCategory(log.ErrorService)
			return errors.Join(ErrCreateDeployment, err)
		}

		deploymentIDStr = strconv.Itoa(int(deployment.GetID()))
		log.Entry().Infof("Deployment %v created on %v/%v", deploymentIDStr, config.Owner, config.Repository)
		commonPipelineEnvironment.git.github_deploymentID = deploymentIDStr
	}

	if len(deploymentIDStr) == 0 && config.Status == "" {
		log.Entry().Warn("Make sure to either use the createNew flag to create a new deployment, or provide deploymentID together with status to update the status of an existing deployment. Otherwise, nothing will happen.")
	}

	if config.Status != "" {
		if len(deploymentIDStr) == 0 {
			log.SetErrorCategory(log.ErrorConfiguration)
			return ErrNoDeploymentID
		}

		if err := validateDeploymentStatus(config.Status); err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return err
		}

		req := &github.DeploymentStatusRequest{
			State: &config.Status,
		}

		deploymentIDInt, err := strconv.ParseInt(deploymentIDStr, 10, 64)
		if err != nil {
			log.SetErrorCategory(log.ErrorCustom)
			return errors.Join(ErrParseDeploymentID, err)
		}

		_, _, err = repoClient.CreateDeploymentStatus(ctx, config.Owner, config.Repository, deploymentIDInt, req)
		if err != nil {
			log.SetErrorCategory(log.ErrorService)
			return errors.Join(ErrCreateDeploymentStatus, err)
		}

		log.Entry().Infof("Deployment %v on %v/%v was updated with status %v", deploymentIDStr, config.Owner, config.Repository, config.Status)
	}

	return nil
}

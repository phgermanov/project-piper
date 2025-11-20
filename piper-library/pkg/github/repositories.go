package github

import (
	"context"

	"github.com/google/go-github/v68/github"
)

type GithubRepositoryClient interface {
	CreateDeployment(ctx context.Context, owner, repo string, request *github.DeploymentRequest) (*github.Deployment, *github.Response, error)
	CreateDeploymentStatus(ctx context.Context, owner, repo string, deployment int64, request *github.DeploymentStatusRequest) (*github.DeploymentStatus, *github.Response, error)
}

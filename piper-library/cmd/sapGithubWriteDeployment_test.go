package cmd

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.wdf.sap.corp/ContinuousDelivery/piper-library/pkg/github/mocks"
)

func TestRunSapGithubWriteDeployment(t *testing.T) {
	tests := []struct {
		name             string
		config           sapGithubWriteDeploymentOptions
		setupMock        func(m *mocks.GithubRepositoryClient)
		wantErr          error
		wantDeploymentID string
	}{
		{
			name: "success - create new deployment",
			config: sapGithubWriteDeploymentOptions{
				CreateNew:   true,
				Environment: "dev",
				Comitish:    "abc123",
				Owner:       "org",
				Repository:  "repo",
			},
			setupMock: func(m *mocks.GithubRepositoryClient) {
				m.On("CreateDeployment",
					mock.Anything,
					"org",
					"repo",
					mock.MatchedBy(func(req *github.DeploymentRequest) bool {
						return *req.Ref == "abc123" && *req.Environment == "dev"
					}),
				).Return(&github.Deployment{ID: github.Ptr[int64](42)}, nil, nil)
			},
			wantErr:          nil,
			wantDeploymentID: "42",
		},
		{
			name: "error - missing environment",
			config: sapGithubWriteDeploymentOptions{
				CreateNew: true,
				Comitish:  "abc123",
			},
			setupMock: func(m *mocks.GithubRepositoryClient) {},
			wantErr:   ErrNoEnvironment,
		},
		{
			name: "error - missing comitish",
			config: sapGithubWriteDeploymentOptions{
				CreateNew:   true,
				Environment: "dev",
			},
			setupMock: func(m *mocks.GithubRepositoryClient) {},
			wantErr:   ErrNoComitish,
		},
		{
			name: "error - CreateDeployment fails",
			config: sapGithubWriteDeploymentOptions{
				CreateNew:   true,
				Environment: "dev",
				Comitish:    "abc123",
				Owner:       "org",
				Repository:  "repo",
			},
			setupMock: func(m *mocks.GithubRepositoryClient) {
				m.On("CreateDeployment", mock.Anything, "org", "repo", mock.Anything).
					Return(nil, nil, errors.New("boom"))
			},
			wantErr: ErrCreateDeployment,
		},
		{
			name: "success - update status on existing deployment",
			config: sapGithubWriteDeploymentOptions{
				DeploymentID: "123",
				Status:       "success",
				Owner:        "org",
				Repository:   "repo",
			},
			setupMock: func(m *mocks.GithubRepositoryClient) {
				m.On("CreateDeploymentStatus",
					mock.Anything,
					"org",
					"repo",
					int64(123),
					mock.MatchedBy(func(req *github.DeploymentStatusRequest) bool {
						return *req.State == "success"
					}),
				).Return(nil, nil, nil)
			},
			wantErr:          nil,
			wantDeploymentID: "",
		},
		{
			name: "error - missing deploymentID",
			config: sapGithubWriteDeploymentOptions{
				Owner:      "org",
				Repository: "repo",
				Status:     "success",
			},
			setupMock: func(m *mocks.GithubRepositoryClient) {},
			wantErr:   ErrNoDeploymentID,
		},
		{
			name: "error - invalid deployment ID parse",
			config: sapGithubWriteDeploymentOptions{
				DeploymentID: "notanumber",
				Status:       "failure",
				Owner:        "org",
				Repository:   "repo",
			},
			setupMock: func(m *mocks.GithubRepositoryClient) {},
			wantErr:   ErrParseDeploymentID,
		},
		{
			name: "error - CreateDeploymentStatus fails",
			config: sapGithubWriteDeploymentOptions{
				DeploymentID: strconv.Itoa(123),
				Status:       "failure",
				Owner:        "org",
				Repository:   "repo",
			},
			setupMock: func(m *mocks.GithubRepositoryClient) {
				m.On("CreateDeploymentStatus", mock.Anything, "org", "repo", int64(123), mock.Anything).
					Return(nil, nil, errors.New("status error"))
			},
			wantErr: ErrCreateDeploymentStatus,
		},
		{
			name: "error - invalid deployment status",
			config: sapGithubWriteDeploymentOptions{
				DeploymentID: "123",
				Status:       "invalid_status",
				Owner:        "org",
				Repository:   "repo",
			},
			setupMock: func(m *mocks.GithubRepositoryClient) {},
			wantErr:   ErrInvalidStatus,
		},
		{
			name: "success - create new deployment and update its status",
			config: sapGithubWriteDeploymentOptions{
				CreateNew:   true,
				Environment: "prod",
				Comitish:    "main",
				Owner:       "org",
				Repository:  "repo",
				Status:      "success",
			},
			setupMock: func(m *mocks.GithubRepositoryClient) {
				m.On("CreateDeployment",
					mock.Anything,
					"org",
					"repo",
					mock.MatchedBy(func(req *github.DeploymentRequest) bool {
						return *req.Ref == "main" && *req.Environment == "prod"
					}),
				).Return(&github.Deployment{ID: github.Ptr[int64](101)}, nil, nil)

				m.On("CreateDeploymentStatus",
					mock.Anything,
					"org",
					"repo",
					int64(101),
					mock.MatchedBy(func(req *github.DeploymentStatusRequest) bool {
						return *req.State == "success"
					}),
				).Return(nil, nil, nil)
			},
			wantErr:          nil,
			wantDeploymentID: "101",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.GithubRepositoryClient)
			tt.setupMock(mockRepo)

			cpe := &sapGithubWriteDeploymentCommonPipelineEnvironment{}

			err := runSapGithubWriteDeployment(
				context.Background(),
				&tt.config,
				cpe,
				mockRepo,
			)

			if tt.wantErr != nil {
				assert.ErrorContains(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.wantDeploymentID != "" {
				assert.Equal(t, tt.wantDeploymentID, cpe.git.github_deploymentID)
			} else {
				assert.Empty(t, cpe.git.github_deploymentID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

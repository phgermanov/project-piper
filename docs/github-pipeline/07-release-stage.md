# Release Stage

## Overview

The Release stage deploys the successfully tested and promoted application to the production environment. This stage executes in an unattended mode or triggers a release process after all quality gates have passed and promotion has been approved.

## Stage Purpose

The Release stage performs the following key functions:

- Downloads promoted artifacts from production repository
- Provisions production infrastructure (if needed)
- Deploys application to production environment
- Executes deployment to Cloud Foundry or Kubernetes
- Updates GitOps repositories for production
- Publishes GitHub releases
- Collects deployment insights and metrics
- Uploads to Deploy with Confidence (DwC)
- Creates production deployment records

## When the Stage Runs

The Release stage runs when:

- The Init stage activates it based on configuration
- `fromJSON(inputs.active-stages-map).Release == true`
- **Only on productive branch**: `inputs.on-productive-branch == 'true'`
- Promote stage completed successfully
- Manual approval granted (via GitHub environment)

## Manual Approval

The Release stage uses GitHub's deployment environment feature:

- Default environment name: `Piper Release`
- Configurable via `environment` input parameter
- Requires reviewers to approve production deployment
- Supports wait timers and protection rules

## Steps in the Stage

### 1. Checkout repository
- **Action**: `actions/checkout@v4.3.0`
- **Purpose**: Checks out source code and deployment configurations
- **Configuration**: Supports submodules and LFS

### 2. Setup Go for development builds
- **Action**: `actions/setup-go@v6.0.0`
- **Condition**: Only for development Piper versions
- **Go Version**: 1.24

### 3. Checkout global extension
- **Action**: `actions/checkout@v4`
- **Condition**: When extensibility is enabled
- **Path**: `.pipeline/tmp/global_extensions`

### 4. Retrieve System Trust session token
- **Action**: `project-piper/system-trust-composite-action`
- **Permissions Required**: `id-token: write`
- **Error Handling**: Continues on error

### 5. preRelease
- **Condition**: When extensibility is enabled
- **Purpose**: Custom pre-release actions
- **Use Cases**: Backup production, notify stakeholders, prepare monitoring

### Infrastructure and Deployment Steps

#### 6. terraformExecute
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `terraformExecute`
- **Condition**: Explicitly activated (inactive by default)
- **Purpose**: Provisions or updates production infrastructure

#### 7. gitopsUpdateDeployment
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `gitopsUpdateDeployment`
- **Flags**: `--username github-actions --password ${{ github.token }}`
- **Condition**: Explicitly activated (inactive by default)
- **Purpose**: Updates production deployment manifest in Git repository

#### 8. sapDownloadArtifact
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapDownloadArtifact`
- **Flags**: `--fromStaging=false`
- **Condition**: Active for native builds or when Helm chart URL available
- **Purpose**: Downloads promoted artifacts from production repository
- **Note**: Uses production repository, not staging

#### 9. cloudFoundryDeploy
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `cloudFoundryDeploy`
- **Condition**: `cfSpace` or `cloudFoundry/space` is configured
- **Purpose**: Deploys application to Cloud Foundry production
- **Strategy**: Typically uses blue-green deployment

#### 10. kubernetesDeploy
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `kubernetesDeploy`
- **Flags**: `--githubToken ${{ github.token }}`
- **Condition**: `deployTool` is helm, helm3, or kubectl
- **Purpose**: Deploys application to Kubernetes production cluster

### Release Management Steps

#### 11. sapDwCStageRelease
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapDwCStageRelease`
- **Condition**: DwC URL and project name configured
- **Purpose**: Uploads release information to Deploy with Confidence
- **Use Case**: SAP internal release management

#### 12. sapCollectInsights
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapCollectInsights`
- **Flags**: `--githubToken ${{ github.token }}`
- **Condition**: Identifier and organization configured
- **Purpose**: Collects DORA metrics from pipeline
- **Error Handling**: Continues on error
- **Metrics**: Deployment frequency, lead time, change failure rate

#### 13. githubPublishRelease
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `githubPublishRelease`
- **Flags**: `--token ${{ github.token }}`
- **Condition**: GitHub token configured
- **Purpose**: Creates GitHub release with artifacts and release notes
- **Includes**: Version tag, changelog, release artifacts

### 14. postRelease
- **Condition**: When extensibility is enabled
- **Purpose**: Custom post-release actions
- **Use Cases**: Verify deployment, run smoke tests, notify users

### 15. Export pipeline environment
- **Purpose**: Exports CPE for downstream stages
- **Condition**: Always runs
- **Output**: `pipelineEnv`

## Configuration Options

### Production Deployment Configuration

Configure in `.pipeline/config.yml`:

```yaml
stages:
  Release:
    # Explicitly activate stage
    active: true
```

### Cloud Foundry Production Deployment

```yaml
steps:
  cloudFoundryDeploy:
    # Production Cloud Foundry API
    apiEndpoint: 'https://api.cf.production.example.com'

    # Production organization
    org: 'production-org'

    # Production space
    space: 'production'

    # Deployment type
    deployType: 'blue-green'

    # Manifest file
    manifest: 'manifest-production.yml'

    # Application name
    appName: 'my-app'

    # Credentials from Vault
    credentialsId: 'cf-production-credentials'

    # Keep old app on rollback
    keepOldInstance: true

    # Smoke test script
    smokeTestScript: './scripts/smoke-test.sh'
```

### Kubernetes Production Deployment

```yaml
steps:
  kubernetesDeploy:
    # Deploy tool
    deployTool: 'helm3'

    # Production namespace
    namespace: 'production'

    # Helm chart
    helmChartPath: './helm/my-app'

    # Production values
    helmValues:
      - 'helm/values-production.yaml'

    # Container image (from promoted artifacts)
    image: 'production-registry.com/my-app:${version}'

    # Production kubeconfig
    kubeConfig: 'kubeconfig-production'

    # Ingress configuration
    ingressHosts:
      - 'my-app.example.com'
      - 'www.my-app.example.com'

    # Additional parameters
    additionalParameters:
      - '--atomic'  # Rollback on failure
      - '--wait'
      - '--timeout 10m'
```

### GitOps Production Update

```yaml
steps:
  gitopsUpdateDeployment:
    # Production manifests repository
    gitRepository: 'my-org/production-k8s'

    # Production branch
    branchName: 'main'

    # Commit message
    commitMessage: 'Deploy ${version} to production'

    # File to update
    filePath: 'production/my-app/deployment.yaml'

    # Container configuration
    containerName: 'my-app'
    containerImage: 'production-registry.com/my-app:${version}'
```

### Terraform Production Infrastructure

```yaml
steps:
  terraformExecute:
    command: 'apply'

    # Production workspace
    terraformWorkspace: 'production'

    # Production variables
    terraformVariables:
      - 'terraform-production.tfvars'

    # Require approval (already handled by GitHub environment)
    autoApprove: false
```

### GitHub Release Configuration

```yaml
steps:
  githubPublishRelease:
    # Release body template
    releaseBodyTemplate: '.github/release-template.md'

    # Add changelog
    addDeltaToLastRelease: true

    # Asset patterns to upload
    assetPaths:
      - 'dist/*.zip'
      - 'target/*.jar'

    # Pre-release flag
    preRelease: false

    # Release name pattern
    releaseNamePattern: 'Release ${version}'
```

### Deploy with Confidence Configuration

```yaml
steps:
  sapDwCStageRelease:
    # DwC instance URL
    themistoInstanceURL: 'https://dwc.example.com'

    # Project name
    dwcProjectName: 'my-project'

    # Version
    version: '${version}'

    # Additional metadata
    deploymentTarget: 'production'
```

### DORA Metrics Collection

```yaml
steps:
  sapCollectInsights:
    # Organization identifier
    gitOrganization: 'my-org'

    # Project identifier
    identifier: 'my-project'

    # Deployment target
    deploymentTarget: 'production'

    # Repository info
    gitRepository: 'my-app'
```

## Example Usage

### Basic Release Stage

```yaml
jobs:
  release:
    needs: [init, build, promote]
    uses: project-piper/piper-pipeline-github/.github/workflows/release.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.promote.outputs.pipeline-env }}
    secrets: inherit
```

### Release with Custom Environment

```yaml
jobs:
  release:
    needs: [init, build, promote]
    uses: project-piper/piper-pipeline-github/.github/workflows/release.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.promote.outputs.pipeline-env }}
      environment: 'Production Deployment'
      extensibility-enabled: true
      global-extensions-repository: 'my-org/pipeline-extensions'
    secrets: inherit
```

### Complete Production Pipeline

```yaml
name: Production Pipeline

on:
  push:
    branches: [main]

permissions:
  id-token: write
  contents: write
  packages: write

jobs:
  init:
    uses: project-piper/piper-pipeline-github/.github/workflows/init.yml@main

  build:
    needs: init
    uses: project-piper/piper-pipeline-github/.github/workflows/build.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.init.outputs.pipeline-env }}
    secrets: inherit

  integration:
    needs: [init, build]
    uses: project-piper/piper-pipeline-github/.github/workflows/integration.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.build.outputs.pipeline-env }}
    secrets: inherit

  acceptance:
    needs: [init, build, integration]
    uses: project-piper/piper-pipeline-github/.github/workflows/acceptance.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.integration.outputs.pipeline-env }}
    secrets: inherit

  promote:
    needs: [init, build, acceptance]
    uses: project-piper/piper-pipeline-github/.github/workflows/promote.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.acceptance.outputs.pipeline-env }}
      environment: 'Piper Promote'
    secrets: inherit

  release:
    needs: [init, build, promote]
    uses: project-piper/piper-pipeline-github/.github/workflows/release.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.promote.outputs.pipeline-env }}
      environment: 'Piper Release'
    secrets: inherit
```

## Extension Examples

### preRelease Extension

Create `.pipeline/extensions/preRelease.sh`:

```bash
#!/bin/bash
set -e

echo "Preparing for production release..."

# Backup current production
kubectl -n production get deployment my-app -o yaml > backup/deployment-$(date +%Y%m%d-%H%M%S).yaml

# Notify stakeholders
curl -X POST https://slack.com/api/chat.postMessage \
  -H "Authorization: Bearer $SLACK_TOKEN" \
  -d "channel=#releases" \
  -d "text=:rocket: Starting production deployment of version ${VERSION}"

# Verify production readiness
./scripts/production-readiness-check.sh

echo "Pre-release preparation complete"
```

### postRelease Extension

Create `.pipeline/extensions/postRelease.sh`:

```bash
#!/bin/bash
set -e

echo "Post-release verification..."

# Run smoke tests
./scripts/production-smoke-tests.sh

# Verify health
HEALTH_URL="https://my-app.example.com/health"
HEALTH_STATUS=$(curl -s -o /dev/null -w "%{http_code}" $HEALTH_URL)

if [ "$HEALTH_STATUS" != "200" ]; then
    echo "Health check failed! Status: $HEALTH_STATUS"
    exit 1
fi

# Update status page
curl -X POST https://status.example.com/api/incidents \
  -d "status=resolved" \
  -d "message=Deployment completed successfully"

# Notify success
curl -X POST https://slack.com/api/chat.postMessage \
  -H "Authorization: Bearer $SLACK_TOKEN" \
  -d "channel=#releases" \
  -d "text=:white_check_mark: Production deployment successful! Version ${VERSION} is live."

echo "Post-release verification complete"
```

## GitHub Environment Configuration

### Creating Production Environment

1. **Settings** > **Environments** > **New environment**
2. Name: `Piper Release`
3. Configure protection rules

### Protection Rules

#### Required Reviewers
- Add production deployment approvers
- Consider requiring 2+ reviewers
- Include on-call engineer

#### Deployment Branches
- Restrict to `main` branch only
- Ensures only productive branch releases

#### Wait Timer
- Optional cooling-off period
- Recommended: 5-15 minutes
- Allows time to cancel if issues detected

## Deployment Strategies

### Blue-Green Deployment (Cloud Foundry)

```yaml
steps:
  cloudFoundryDeploy:
    deployType: 'blue-green'
    keepOldInstance: true
    smokeTestScript: './scripts/smoke-test.sh'
```

Process:
1. Deploy new version (green)
2. Run smoke tests
3. Switch traffic to green
4. Keep blue for rollback

### Rolling Update (Kubernetes)

```yaml
steps:
  kubernetesDeploy:
    deployTool: 'helm3'
    additionalParameters:
      - '--set strategy.type=RollingUpdate'
      - '--set strategy.rollingUpdate.maxUnavailable=0'
      - '--set strategy.rollingUpdate.maxSurge=1'
```

### Canary Deployment (Kubernetes)

Use preRelease and postRelease extensions:

```bash
# preRelease: Deploy canary
kubectl apply -f k8s/canary-deployment.yaml

# Monitor metrics
# If successful, continue to full rollout

# postRelease: Remove canary
kubectl delete -f k8s/canary-deployment.yaml
```

## Rollback Procedures

### Automated Rollback

For Kubernetes with `--atomic` flag:
- Automatic rollback on failure
- Reverts to previous version

### Manual Rollback

Cloud Foundry:
```bash
# List apps
cf apps

# Rollback to blue
cf delete my-app-venerable
cf rename my-app my-app-venerable
cf rename my-app-blue my-app
```

Kubernetes:
```bash
# Rollback to previous revision
kubectl rollout undo deployment/my-app -n production

# Rollback to specific revision
kubectl rollout undo deployment/my-app -n production --to-revision=2
```

## Troubleshooting

### Deployment Failures

1. **Timeout errors**:
   - Increase deployment timeout
   - Check resource availability
   - Verify image pull credentials

2. **Health check failures**:
   - Verify health endpoint
   - Check application logs
   - Review resource limits

3. **Permission errors**:
   - Verify credentials in Vault
   - Check namespace permissions
   - Validate service account

### Release Creation Issues

1. **GitHub release failed**:
   - Verify GitHub token permissions
   - Check tag doesn't already exist
   - Validate release notes template

2. **Asset upload failed**:
   - Verify asset paths exist
   - Check file sizes (< 2GB limit)
   - Validate file permissions

### GitOps Update Failures

1. **Repository access**:
   - Verify GitHub token has write access
   - Check repository exists
   - Validate branch name

2. **Merge conflicts**:
   - Resolve conflicts manually
   - Update deployment manifest
   - Retry pipeline

## Best Practices

1. **Monitoring**: Set up comprehensive monitoring before release
2. **Rollback Plan**: Always have rollback procedure ready
3. **Smoke Tests**: Run automated smoke tests post-deployment
4. **Communication**: Notify stakeholders before, during, and after
5. **Change Windows**: Deploy during planned maintenance windows
6. **Database Migrations**: Handle separately from code deployment
7. **Feature Flags**: Use for gradual feature rollout
8. **Health Checks**: Implement robust health and readiness probes
9. **Metrics**: Collect and monitor deployment metrics
10. **Documentation**: Document what each release includes

## DORA Metrics

The Release stage contributes to DORA metrics:

- **Deployment Frequency**: Tracked per release
- **Lead Time for Changes**: From commit to production
- **Change Failure Rate**: Failed vs. successful deployments
- **Mean Time to Recovery**: Time to recover from failures

## Security Considerations

1. **Credentials**: Use Vault for all production credentials
2. **Approval**: Require manual approval for production
3. **Audit Trail**: Maintain complete deployment history
4. **Access Control**: Limit who can approve releases
5. **Secrets Management**: Rotate secrets regularly
6. **Compliance**: Ensure regulatory compliance

## Related Stages

- **Promote Stage**: Must succeed before release
- **Post Stage**: Runs after all stages including release
- **Build Stage**: Provides artifacts for release

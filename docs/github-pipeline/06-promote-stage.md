# Promote Stage

## Overview

The Promote stage promotes a successfully tested build artifact from staging to production-ready status. This stage only runs after all previous test stages (Build, Integration, Acceptance, Performance) have completed successfully, implementing a "quality gate" pattern.

## Stage Purpose

The Promote stage performs the following key functions:

- Validates that all quality gates have passed
- Promotes build artifacts from staging to production repository
- Creates a locked pipeline run marker (optional)
- Marks artifacts as production-ready in the staging service
- Ensures only validated artifacts proceed to release
- Implements manual approval workflow via GitHub environments

## When the Stage Runs

The Promote stage runs when:

- The Init stage activates it based on configuration
- `fromJSON(inputs.active-stages-map).Promote == true`
- **Only on productive branch**: `inputs.on-productive-branch == 'true'`
- All previous stages (Build, Integration, Acceptance, Performance) completed successfully
- Manual approval is granted (via GitHub environment)

## Manual Approval

The Promote stage uses GitHub's **deployment environment** feature for manual approval:

- Default environment name: `Piper Promote`
- Configurable via `environment` input parameter
- Requires reviewers to approve before promoting
- Supports wait timers and other protection rules

## Steps in the Stage

### 1. Checkout repository
- **Action**: `actions/checkout@v4.3.0`
- **Purpose**: Checks out source code for configuration access
- **Configuration**: Supports submodules and LFS

### 2. Retrieve System Trust session token
- **Action**: `project-piper/system-trust-composite-action`
- **Purpose**: Obtains session token for System Trust integration
- **Permissions Required**: `id-token: write`
- **Error Handling**: Continues on error

### 3. Setup Go for development builds
- **Action**: `actions/setup-go@v6.0.0`
- **Condition**: Only for development Piper versions
- **Go Version**: 1.24

### 4. sapCallStagingService
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapCallStagingService`
- **Flags**: `--action promote`
- **Purpose**: Promotes artifacts from staging to production repository
- **When**: When step is active in configuration
- **Effect**: Moves artifacts to production-ready state

### 5. Read stage configuration
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `getConfig`
- **Flags**: `--stageConfig --outputFile stage-config.json`
- **Purpose**: Reads pipeline configuration including lock settings

### 6. Set lock pipeline run configuration
- **Purpose**: Determines if pipeline run should be locked
- **Source**: `lockPipelineRun` from stage configuration
- **Output**: `lockPipelineRun` environment variable

### 7. Create lock file for productive branch
- **Condition**: `lockPipelineRun == 'true'`
- **Purpose**: Creates marker file to lock this pipeline run
- **File**: `lock-run.json`
- **Use Case**: Prevents multiple simultaneous releases

### 8. sapCumulusUpload (lock-run.json)
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapCumulusUpload`
- **Flags**: `--filePattern **/lock-run.json --stepResultType root`
- **Condition**: `lockPipelineRun == 'true'`
- **Purpose**: Uploads lock file to Cumulus

### 9. Export pipeline environment
- **Action**: `SAP/project-piper-action@v1.22`
- **Purpose**: Exports CPE for downstream stages
- **Condition**: Always runs
- **Output**: `pipelineEnv`

## Configuration Options

### Stage Configuration

Configure in `.pipeline/config.yml`:

```yaml
stages:
  Promote:
    # Explicitly activate stage (usually auto-activated)
    active: true
```

### Staging Service Configuration

```yaml
steps:
  sapCallStagingService:
    # Staging service URL (from Vault or config)
    stagingServiceUrl: 'https://staging.example.com'

    # Credentials for staging service (from Vault)
    credentialsId: 'staging-credentials'

    # Group ID
    groupId: 'com.example.myapp'

    # Artifact ID
    artifactId: 'my-application'

    # Version (from artifactPrepareVersion)
    version: '${version}'
```

### Pipeline Lock Configuration

```yaml
general:
  # Lock pipeline run after promotion
  lockPipelineRun: true
```

When `lockPipelineRun: true`:
- Prevents concurrent promotions/releases
- Creates lock marker in Cumulus
- Useful for coordinated releases

### Cumulus Configuration

```yaml
steps:
  sapCumulusUpload:
    # Pipeline ID
    pipelineId: 'your-pipeline-id'

    # Server URL (from Vault or config)
    serverUrl: 'https://cumulus.example.com'
```

## Example Usage

### Basic Promote Stage

```yaml
jobs:
  promote:
    needs: [init, build, acceptance]
    uses: project-piper/piper-pipeline-github/.github/workflows/promote.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.acceptance.outputs.pipeline-env }}
    secrets: inherit
```

### Promote with Custom Environment Name

```yaml
jobs:
  promote:
    needs: [init, build, acceptance]
    uses: project-piper/piper-pipeline-github/.github/workflows/promote.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.acceptance.outputs.pipeline-env }}
      environment: 'Production Approval'  # Custom environment name
    secrets: inherit
```

### Complete Pipeline with Promote

```yaml
name: Production Pipeline

on:
  push:
    branches: [main]

permissions:
  id-token: write
  contents: read

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

  performance:
    needs: [init, build, acceptance]
    uses: project-piper/piper-pipeline-github/.github/workflows/performance.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.acceptance.outputs.pipeline-env }}
    secrets: inherit

  promote:
    needs: [init, build, integration, acceptance, performance]
    uses: project-piper/piper-pipeline-github/.github/workflows/promote.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.performance.outputs.pipeline-env }}
      environment: 'Piper Promote'
    secrets: inherit
```

## GitHub Environment Configuration

### Creating the Environment

1. Navigate to **Repository Settings** > **Environments**
2. Click **New environment**
3. Name it `Piper Promote` (or your custom name)
4. Click **Configure environment**

### Configuring Protection Rules

#### Required Reviewers

1. Under **Deployment protection rules**, check **Required reviewers**
2. Add team members or teams who can approve promotions
3. Recommendations:
   - Add release managers
   - Add product owners
   - Consider requiring 2+ reviewers for production

#### Wait Timer

1. Check **Wait timer**
2. Set delay in minutes (e.g., 5 minutes)
3. Use case: Cooling-off period to catch issues

#### Deployment Branches

1. Under **Deployment branches**, select:
   - **Selected branches** to restrict to specific branches
   - Add rule for your productive branch (e.g., `main`)

### Environment Secrets

Configure environment-specific secrets if needed:

- `STAGING_SERVICE_TOKEN`
- `PRODUCTION_CREDENTIALS`

## Manual Approval Workflow

### Approval Process

1. **Pipeline reaches Promote stage**
   - GitHub Actions pauses execution
   - Creates deployment request
   - Notifies configured reviewers

2. **Reviewers receive notification**
   - Email notification
   - GitHub UI notification
   - Can view pipeline logs and test results

3. **Review and approve/reject**
   - Navigate to Actions > Workflow run
   - View deployment request
   - Review artifacts and test results
   - Approve or Reject

4. **Pipeline continues or stops**
   - If approved: Promote stage executes
   - If rejected: Pipeline fails, no promotion

### Approval Best Practices

1. **Review Test Results**: Check all previous stage results
2. **Verify Artifacts**: Confirm correct version being promoted
3. **Check Cumulus**: Review uploaded evidence and reports
4. **Communication**: Use GitHub comments to document decision
5. **Timing**: Promote during maintenance windows when possible

## Staging Service Promote Action

The promote action in staging service:

1. **Validates** the artifact exists in staging
2. **Checks** all quality gates passed
3. **Moves** artifact from staging to production repository
4. **Updates** metadata to mark as production-ready
5. **Creates** audit trail entry

## Lock Pipeline Run Feature

### Purpose

- Prevents concurrent releases
- Ensures only one version is being released at a time
- Useful for coordinated multi-component releases

### How It Works

1. Promote stage creates `lock-run.json` file
2. File uploaded to Cumulus
3. Subsequent pipeline runs can check for lock
4. Lock can be released after successful release

### Configuration Example

```yaml
general:
  lockPipelineRun: true

steps:
  sapCumulusUpload:
    pipelineId: 'my-pipeline'
```

### Checking for Locks

In custom extensions, you can check for locks:

```bash
# Check if pipeline is locked
LOCK_EXISTS=$(curl -s "https://cumulus.example.com/api/runs/latest/lock")

if [ "$LOCK_EXISTS" == "true" ]; then
    echo "Pipeline is locked, waiting for release..."
    exit 1
fi
```

## Troubleshooting

### Promote Stage Not Running

1. **Check productive branch**: Verify running on productive branch
2. **Stage activation**: Confirm stage is in `active-stages-map`
3. **Previous stages**: Ensure all previous stages succeeded
4. **Configuration**: Verify stage configuration is correct

### Staging Service Errors

1. **Connection issues**:
   - Verify staging service URL
   - Check network connectivity
   - Validate credentials in Vault

2. **Promotion failures**:
   - Ensure artifact exists in staging
   - Verify artifact version matches
   - Check quality gates passed

3. **Permission denied**:
   - Verify credentials have promote permissions
   - Check user/service account roles

### Approval Issues

1. **No reviewers notified**:
   - Check environment is configured correctly
   - Verify reviewers are added
   - Check GitHub notification settings

2. **Cannot approve**:
   - Verify user is in reviewers list
   - Check permissions on repository
   - Ensure environment protection rules allow

### Lock File Issues

1. **Lock not created**:
   - Verify `lockPipelineRun: true` in config
   - Check Cumulus upload succeeded
   - Review step logs

2. **Lock not released**:
   - Implement lock release in Release stage
   - Check for orphaned locks
   - Manual cleanup may be needed

## Best Practices

1. **Quality Gates**: Ensure all tests pass before promotion
2. **Manual Review**: Always require manual approval for production
3. **Audit Trail**: Document approval decisions
4. **Timing**: Promote during planned deployment windows
5. **Communication**: Notify stakeholders of pending promotions
6. **Rollback Plan**: Have rollback procedure ready
7. **Version Tracking**: Maintain clear version history
8. **Lock Management**: Use locks for coordinated releases
9. **Monitoring**: Monitor promotion success rates
10. **Documentation**: Document what each version contains

## Integration with Other Systems

### Notification Integration

Add notifications in `.pipeline/extensions/postPromote.sh`:

```bash
#!/bin/bash

# Notify via Slack
curl -X POST https://slack.com/api/chat.postMessage \
  -H "Authorization: Bearer $SLACK_TOKEN" \
  -d "channel=#releases" \
  -d "text=Artifact promoted to production: version ${VERSION}"

# Create Jira release
curl -X POST https://jira.example.com/rest/api/2/version \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"${VERSION}\", \"released\": false}"
```

### Change Management Integration

```bash
# Create change request in ServiceNow
curl -X POST https://servicenow.example.com/api/now/table/change_request \
  -H "Content-Type: application/json" \
  -d "{
    \"short_description\": \"Release version ${VERSION}\",
    \"description\": \"Automated release from pipeline\",
    \"type\": \"Standard\"
  }"
```

## Security Considerations

1. **Approval Authority**: Limit who can approve promotions
2. **Credentials**: Store staging credentials securely in Vault
3. **Audit Logging**: Maintain complete audit trail
4. **Separation of Duties**: Different approvers for different environments
5. **Time-based Controls**: Restrict promotions to specific time windows

## Related Stages

- **Build, Integration, Acceptance, Performance**: Must all succeed before promotion
- **Release Stage**: Executes after successful promotion
- **Post Stage**: Runs after all stages complete

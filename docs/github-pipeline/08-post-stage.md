# Post Stage

## Overview

The Post stage is the final stage of the GitHub Actions General Purpose Pipeline. It executes post-pipeline actions such as reporting, cleanup, monitoring, and publishing pipeline completion events. This stage **always runs**, regardless of the success or failure of previous stages.

## Stage Purpose

The Post stage performs the following key functions:

- Reports pipeline status for monitoring and availability tracking
- Uploads final result files and logs to Cumulus
- Rotates Vault secrets (secretId rotation)
- Publishes pipeline completion events to GCP
- Generates release status information
- Executes custom post-pipeline cleanup
- Sends notifications (Jenkins only)
- Provides pipeline execution summary

## When the Stage Runs

The Post stage **always runs**:

- After all other stages complete (success or failure)
- For both productive and non-productive branches
- For all trigger types (push, pull request, schedule)
- Independent of previous stage outcomes

## Steps in the Stage

### 1. Checkout repository
- **Action**: `actions/checkout@v4.3.0`
- **Purpose**: Checks out source code for configuration access
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

### 5. sapReportPipelineStatus
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapReportPipelineStatus`
- **Condition**: Always runs (success or failure)
- **Flags**: `--gitHubToken ${{ github.token }}`
- **Purpose**: Reports pipeline execution status for monitoring
- **Environment**: `GITHUB_URL: 'https://github.tools.sap/'`
- **Use Case**: Availability monitoring, pipeline analytics

### 6. vaultRotateSecretId
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `vaultRotateSecretId`
- **Condition**: Active when Vault namespace is configured
- **Purpose**: Rotates Vault AppRole secret ID for security
- **Flags**:
  ```
  --secretStore github
  --owner ${{ github.repository_owner }}
  --repository ${{ github.event.repository.name }}
  --githubApiUrl ${{ github.api_url }}
  --vaultAppRoleSecretTokenCredentialsId PIPER_vaultAppRoleSecretID
  ```
- **Note**:
  - For Azure/Jenkins: Runs when configured
  - For GitHub Actions: Inactive by default (handled by automaticd for Portal pipelines)
  - Only runs for pipelines registered with freestyle operator

### 7. sapCumulusUpload (cumulus-configuration)
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapCumulusUpload`
- **Condition**: On productive branch
- **Flags**: `--filePattern **/cumulus-configuration.json --stepResultType root`
- **Purpose**: Uploads Cumulus configuration metadata

### 8. sapCumulusUpload (pipelineLog)
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapCumulusUpload`
- **Condition**: On productive branch
- **Flags**: `--filePattern ./pipelineLog*.log --stepResultType log`
- **Purpose**: Uploads complete pipeline execution logs

### 9. gcpPublishEvent - pipelineTaskRunFinished
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `gcpPublishEvent`
- **Condition**: On productive branch with configured Vault namespace
- **Flags**: `--eventType sap.hyperspace.pipelineRunFinished --topic hyperspace-pipelinerun-finished`
- **Purpose**: Publishes pipeline completion event to Google Cloud Platform
- **Use Case**: Pipeline monitoring, analytics, downstream triggers

### 10. sapGenerateEnvironmentInfo
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapGenerateEnvironmentInfo`
- **Condition**: Always runs on productive branch
- **Flags**: `--generateFiles releaseStatus`
- **Purpose**: Generates release status file with pipeline results
- **Note**: Must run after gcpPublishEvent event

### 11. sapCumulusUpload (release status info)
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapCumulusUpload`
- **Condition**: On productive branch
- **Flags**: `--filePattern **/release-status-*.json --stepResultType .status-log/release`
- **Purpose**: Uploads release status information
- **Error Handling**: Continues on error

### 12. postPost
- **Condition**: When extensibility is enabled
- **Purpose**: Custom post-stage extensions
- **Use Cases**: Final cleanup, summary reports, notifications

### 13. Export pipeline environment
- **Action**: `SAP/project-piper-action@v1.22`
- **Condition**: Always runs
- **Purpose**: Exports final CPE state
- **Output**: `pipelineEnv`

## Configuration Options

### Pipeline Status Reporting

```yaml
steps:
  sapReportPipelineStatus:
    # Enable/disable status reporting
    # (Usually always active for monitoring)
    enabled: true

    # GitHub token (from Vault or input)
    gitHubToken: ${{ github.token }}
```

### Vault Secret Rotation

```yaml
general:
  # Vault namespace (must be configured for rotation)
  vaultNamespace: 'ies/hyperspace/pipelines'

steps:
  vaultRotateSecretId:
    # Secret store type
    secretStore: 'github'

    # Rotation interval (handled by Piper)
    # Automatically rotates when needed
```

### Cumulus Upload Configuration

```yaml
steps:
  sapCumulusUpload:
    # Pipeline ID (required for Cumulus)
    pipelineId: 'your-pipeline-id'

    # Server URL (from Vault)
    serverUrl: 'https://cumulus.example.com'

    # Upload logs and configuration
    uploadLogs: true
```

### GCP Event Publishing

```yaml
steps:
  gcpPublishEvent:
    # Vault namespace (required for OIDC token)
    vaultNamespace: 'ies/hyperspace/pipelines'

    # Event topic
    topic: 'hyperspace-pipelinerun-finished'

    # Event type
    eventType: 'sap.hyperspace.pipelineRunFinished'
```

### Notification Configuration (Jenkins)

For Jenkins GPP:

```yaml
steps:
  slackSendNotification:
    # Slack channel
    channel: '#deployments'

    # Webhook URL (from Vault)
    webhookUrl: '${vault:slack/webhook}'

    # Message template
    message: 'Pipeline completed: ${currentBuild.result}'
```

## Example Usage

### Basic Post Stage

```yaml
jobs:
  post:
    needs: [init, build, release]
    if: always()  # Always run, even if previous jobs failed
    uses: project-piper/piper-pipeline-github/.github/workflows/post.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.release.outputs.pipeline-env || needs.build.outputs.pipeline-env }}
    secrets: inherit
```

### Post with Extensions

```yaml
jobs:
  post:
    needs: [init, build, release]
    if: always()
    uses: project-piper/piper-pipeline-github/.github/workflows/post.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.release.outputs.pipeline-env || needs.build.outputs.pipeline-env }}
      extensibility-enabled: true
      global-extensions-repository: 'my-org/pipeline-extensions'
    secrets: inherit
```

### Complete Pipeline with Post Stage

```yaml
name: Production Pipeline

on:
  push:
    branches: [main]

permissions:
  id-token: write
  contents: write

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

  acceptance:
    needs: [init, build]
    uses: project-piper/piper-pipeline-github/.github/workflows/acceptance.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.build.outputs.pipeline-env }}
    secrets: inherit

  promote:
    needs: [init, build, acceptance]
    uses: project-piper/piper-pipeline-github/.github/workflows/promote.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.acceptance.outputs.pipeline-env }}
    secrets: inherit

  release:
    needs: [init, build, promote]
    uses: project-piper/piper-pipeline-github/.github/workflows/release.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.promote.outputs.pipeline-env }}
    secrets: inherit

  post:
    needs: [init, build, acceptance, promote, release]
    if: always()  # Critical: Always run post stage
    uses: project-piper/piper-pipeline-github/.github/workflows/post.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.release.outputs.pipeline-env || needs.promote.outputs.pipeline-env || needs.build.outputs.pipeline-env }}
    secrets: inherit
```

## Extension Examples

### postPost Extension

Create `.pipeline/extensions/postPost.sh`:

```bash
#!/bin/bash
set -e

echo "Executing post-pipeline cleanup..."

# Generate pipeline summary
cat > pipeline-summary.txt <<EOF
Pipeline Execution Summary
==========================
Repository: ${GITHUB_REPOSITORY}
Branch: ${GITHUB_REF_NAME}
Commit: ${GITHUB_SHA}
Workflow: ${GITHUB_WORKFLOW}
Run ID: ${GITHUB_RUN_ID}
Run Number: ${GITHUB_RUN_NUMBER}
EOF

# Upload summary to artifact storage
if [ -f "pipeline-summary.txt" ]; then
    echo "Pipeline summary generated"
fi

# Send custom notifications
if [ "${PIPELINE_STATUS}" == "success" ]; then
    MESSAGE="✅ Pipeline completed successfully"
else
    MESSAGE="❌ Pipeline failed - review logs"
fi

# Slack notification
curl -X POST "${SLACK_WEBHOOK}" \
    -H 'Content-Type: application/json' \
    -d "{\"text\": \"${MESSAGE}\"}"

# Email notification (example)
echo "${MESSAGE}" | mail -s "Pipeline ${GITHUB_RUN_ID}" team@example.com

# Cleanup temporary resources
echo "Cleaning up temporary resources..."
docker system prune -f
rm -rf /tmp/pipeline-*

echo "Post-pipeline cleanup complete"
```

### Advanced Monitoring Integration

```bash
#!/bin/bash
set -e

# Collect pipeline metrics
DURATION=$(($(date +%s) - ${PIPELINE_START_TIME}))

# Send metrics to monitoring system
curl -X POST "https://metrics.example.com/api/pipeline-metrics" \
    -H "Content-Type: application/json" \
    -d "{
        \"pipeline\": \"${GITHUB_WORKFLOW}\",
        \"duration\": ${DURATION},
        \"status\": \"${PIPELINE_STATUS}\",
        \"branch\": \"${GITHUB_REF_NAME}\",
        \"timestamp\": $(date +%s)
    }"

# Update dashboard
curl -X POST "https://dashboard.example.com/api/update" \
    -d "pipeline=${GITHUB_WORKFLOW}&status=${PIPELINE_STATUS}"
```

## Pipeline Status Tracking

The Post stage provides comprehensive status tracking:

### Status Information

- **Overall Result**: Success, failure, or partial success
- **Stage Results**: Individual stage outcomes
- **Duration**: Total pipeline execution time
- **Artifact Information**: Generated artifacts and versions
- **Test Results**: Summary of test executions
- **Quality Gates**: Pass/fail status of quality checks

### Release Status File

Format of `release-status-*.json`:

```json
{
  "pipelineId": "my-pipeline",
  "runId": "12345",
  "version": "1.2.3",
  "status": "success",
  "startTime": "2025-01-01T10:00:00Z",
  "endTime": "2025-01-01T10:30:00Z",
  "duration": 1800,
  "stages": {
    "build": "success",
    "integration": "success",
    "acceptance": "success",
    "promote": "success",
    "release": "success"
  },
  "artifacts": [
    {
      "name": "my-app.jar",
      "version": "1.2.3",
      "repository": "production"
    }
  ]
}
```

## Vault Secret Rotation

### Why Rotate Secrets?

- **Security Best Practice**: Regular rotation reduces exposure risk
- **Compliance**: Meets security compliance requirements
- **Automated**: Piper handles rotation automatically
- **Zero Downtime**: Rotation happens without pipeline interruption

### Rotation Process

1. **Check Rotation Need**: Piper determines if rotation is needed
2. **Generate New Secret**: Creates new secret ID in Vault
3. **Update GitHub Secret**: Updates repository secret
4. **Verify New Secret**: Validates new secret works
5. **Revoke Old Secret**: Removes old secret from Vault

### Configuration

Vault namespace must be configured:

```yaml
general:
  vaultNamespace: 'ies/hyperspace/pipelines'
  vaultBasePath: 'your-base-path'
  vaultPipelineName: 'your-pipeline'
```

## Cumulus Upload Details

### Files Uploaded in Post Stage

1. **Cumulus Configuration**: `cumulus-configuration.json`
   - Pipeline configuration metadata
   - Cumulus settings

2. **Pipeline Logs**: `pipelineLog*.log`
   - Complete execution logs
   - All stage outputs
   - Error messages and stack traces

3. **Release Status**: `release-status-*.json`
   - Pipeline execution summary
   - Stage results
   - Artifact information

### Upload Conditions

- Only on productive branch
- After all other operations complete
- Includes files from all stages

## GCP Event Publishing

### Event Schema

```json
{
  "eventType": "sap.hyperspace.pipelineRunFinished",
  "source": "github-actions",
  "data": {
    "pipelineId": "my-pipeline",
    "runId": "12345",
    "status": "success",
    "duration": 1800,
    "branch": "main",
    "commit": "abc123",
    "timestamp": "2025-01-01T10:30:00Z"
  }
}
```

### Use Cases

- **Monitoring**: Track pipeline execution metrics
- **Analytics**: Analyze pipeline trends
- **Downstream Triggers**: Trigger dependent pipelines
- **Alerting**: Send alerts on failures

## Troubleshooting

### Status Reporting Failures

1. **GitHub API errors**:
   - Verify GitHub token has required permissions
   - Check API rate limits
   - Validate GitHub URL configuration

2. **Monitoring system unreachable**:
   - Check network connectivity
   - Verify monitoring endpoints
   - Review firewall rules

### Vault Rotation Issues

1. **Rotation failed**:
   - Verify Vault namespace configuration
   - Check Vault permissions
   - Validate GitHub secret access

2. **Secret not updated**:
   - Check GitHub secret permissions
   - Verify repository settings
   - Review rotation logs

### Cumulus Upload Failures

1. **Connection errors**:
   - Verify Cumulus URL
   - Check credentials in Vault
   - Test network connectivity

2. **File not found**:
   - Verify log file generation
   - Check file patterns
   - Review previous stage outputs

### GCP Event Publishing Failures

1. **Authentication errors**:
   - Verify Vault namespace configuration
   - Check OIDC token generation
   - Validate GCP permissions

2. **Topic not found**:
   - Verify topic name
   - Check GCP project configuration
   - Review permissions

## Best Practices

1. **Always Run**: Use `if: always()` to ensure Post stage runs
2. **Pipeline Environment**: Pass latest available pipeline-env
3. **Log Collection**: Always upload logs to Cumulus
4. **Monitoring**: Implement comprehensive status reporting
5. **Notifications**: Configure appropriate notification channels
6. **Cleanup**: Clean up temporary resources
7. **Metrics**: Collect and analyze pipeline metrics
8. **Documentation**: Document pipeline outcomes
9. **Error Handling**: Continue on error for non-critical steps
10. **Security**: Rotate secrets regularly

## Monitoring and Analytics

The Post stage enables:

- **Pipeline Health**: Track success/failure rates
- **Performance**: Monitor execution duration
- **Trends**: Analyze patterns over time
- **Alerts**: Configure alerting on failures
- **Dashboards**: Build pipeline dashboards
- **Reporting**: Generate executive reports

## Related Stages

- **All Stages**: Post runs after all stages complete
- **Init Stage**: Publishes start event (Post publishes finish event)
- **Build Stage**: Post uploads final logs including build logs
- **Release Stage**: Post reports final deployment status

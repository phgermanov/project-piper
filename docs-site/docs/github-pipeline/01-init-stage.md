# Init Stage

## Overview

The Init stage is the first stage of the GitHub Actions General Purpose Pipeline (GPP). It initializes the pipeline run and prepares the environment for further execution.

## Stage Purpose

The Init stage performs the following key functions:

- Checks out the source code repository
- Reads and validates pipeline configuration
- Determines which stages and steps should be executed
- Identifies the productive branch and sets environment flags
- Performs automatic versioning for productive branches (Jenkins only)
- Creates pipeline environment context
- Publishes pipeline start events for monitoring

## When the Stage Runs

The Init stage **always runs** as it is required to initialize the pipeline and determine which subsequent stages should execute. It runs for:

- All branch types (productive and non-productive)
- All trigger types (push, pull request, schedule, manual)
- Both standard and optimized pipeline runs

## Steps in the Stage

### 1. Checkout repository
- **Action**: `actions/checkout@v4.3.0`
- **Purpose**: Checks out the source code from the repository
- **Configuration**:
  - Supports submodules (optional)
  - Supports Git LFS (optional)

### 2. Setup Go for development builds
- **Action**: `actions/setup-go@v6.0.0`
- **Purpose**: Sets up Go environment for development builds of Piper
- **Condition**: Only when using development versions of Piper
- **Go Version**: 1.24

### 3. version
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `version`
- **Purpose**: Displays current Piper version information

### 4. Read stage configuration
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `getConfig`
- **Purpose**: Reads stage configuration from `.pipeline/config.yml`
- **Output**: `stage-config.json`
- **Flags**: `--stageConfig --outputFile stage-config.json`

### 5. Set productive branch environment variable
- **Purpose**: Extracts the productive branch name from configuration
- **Source**: `stage-config.json`

### 6. Check if on productive branch
- **Purpose**: Determines if current branch matches the productive branch
- **Logic**:
  - Compares `github.ref_name` with `productiveBranch` configuration
  - Excludes merge queue branches (prefixed with `gh-readonly-queue/`)
- **Output**: `onProductiveBranch` environment variable (true/false)

### 7. Check if scheduled run
- **Purpose**: Determines if the pipeline was triggered by a schedule
- **Logic**: Checks if `github.event_name == "schedule"`
- **Output**: `scheduledRun` environment variable (true/false)

### 8. Set pipeline optimization flag
- **Purpose**: Checks if pipeline optimization is enabled
- **Source**: `pipelineOptimization` from stage configuration
- **Output**: `pipelineOptimization` environment variable (true/false)

### 9. Check if optimized and scheduled
- **Purpose**: Determines if this is an optimized scheduled run
- **Logic**: `pipelineOptimization == true && event_name == "schedule"`
- **Output**: `isOptimizedAndScheduled` environment variable (true/false)

### 10. Set global extensions repository
- **Purpose**: Extracts global extensions repository configuration
- **Source**: `globalExtensionsRepository` from stage configuration

### 11. Set global extensions reference
- **Purpose**: Extracts global extensions reference (branch/tag)
- **Source**: `globalExtensionsRef` from stage configuration

### 12. Sanitize global extensions repository URL
- **Purpose**: Validates and sanitizes the extensions repository URL
- **Logic**: Filters out null values and non-standard URLs

### 13. Set vault base path
- **Purpose**: Extracts Vault base path for secrets management
- **Source**: `vaultBasePath` from stage configuration

### 14. Set vault pipeline name
- **Purpose**: Extracts Vault pipeline name for secrets management
- **Source**: `vaultPipelineName` from stage configuration

### 15. Prepare outputs
- **Purpose**: Sets job outputs for use by downstream stages
- **Outputs**:
  - `onProductiveBranch`
  - `pipelineOptimization`
  - `isOptimizedAndScheduled`
  - `globalExtensionsRepository`
  - `vaultBasePath`
  - `vaultPipelineName`
  - `globalExtensionsRef`

### 16. sapPipelineInit
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapPipelineInit`
- **Purpose**: Initializes the SAP Piper pipeline environment
- **Flags**: `--isScheduled=${{ env.scheduledRun }} --githubToken ${{ github.token }}`
- **Note**: GitHub token will be overridden by Vault secret if configured

### 17. checkIfStepActive
- **Action**: `SAP/project-piper-action@v1.22`
- **Purpose**: Determines which stages and steps should be active
- **Output Files**:
  - `.pipeline/stage_out.json` - Active stages map
  - `.pipeline/step_out.json` - Active steps map

### 18. Prepare outputs (checkIfStepActive)
- **Purpose**: Exports active stages and steps as job outputs
- **Condition**: Always runs
- **Outputs**:
  - `activeStagesMap` - JSON map of active stages
  - `activeStepsMap` - JSON map of active steps

### 19. Write System Trust URL to CPE
- **Purpose**: Writes System Trust URL to Common Pipeline Environment
- **Source**: `hooks.systemtrust.serverURL` from stage configuration
- **Destination**: `.pipeline/commonPipelineEnvironment/custom/systemTrustURL`
- **Error Handling**: Continues on error

### 20. gcpPublishEvent - pipelineRunStarted
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `gcpPublishEvent`
- **Purpose**: Publishes pipeline start event to Google Cloud Platform
- **Conditions**:
  - Step is active in configuration
  - Running on productive branch
  - Vault namespace is configured (ies/hyperspace/pipelines or hpp/portal)
- **Flags**: `--eventType sap.hyperspace.pipelineRunStarted --topic hyperspace-pipelinerun-started`
- **Error Handling**: Continues on error

### 21. Export pipeline environment
- **Action**: `SAP/project-piper-action@v1.22`
- **Purpose**: Exports Common Pipeline Environment for downstream stages
- **Condition**: Always runs
- **Output**: `pipelineEnv` job output

## Stage Outputs

The Init stage provides the following outputs to downstream stages:

| Output | Description |
|--------|-------------|
| `on-productive-branch` | Whether the current branch is the productive branch |
| `is-optimized-and-scheduled` | Whether this is an optimized scheduled run |
| `active-steps-map` | JSON map of all active steps across all stages |
| `active-stages-map` | JSON map of all active stages |
| `pipeline-env` | Common Pipeline Environment data |
| `pipeline-optimization` | Whether pipeline optimization is enabled |
| `global-extensions-repository` | Repository containing global extensions |
| `vault-base-path` | Vault base path for secrets |
| `vault-pipeline-name` | Vault pipeline name for secrets |
| `global-extensions-ref` | Git reference for global extensions |

## Configuration Options

### Input Parameters

Configure these in the workflow that calls the Init stage:

```yaml
inputs:
  piper-version:
    description: Version of Piper CLI to use
    default: 'latest'

  sap-piper-version:
    description: Version of SAP Piper library to use
    default: 'latest'

  runs-on:
    description: Runner labels as JSON array
    default: '[ "self-hosted" ]'

  custom-defaults-paths:
    description: Custom defaults file paths
    default: ''

  custom-stage-conditions-path:
    description: Path to custom stage conditions file
    default: ''

  checkout-submodules:
    description: Whether to checkout submodules
    default: 'false'

  checkout-lfs:
    description: Whether to checkout Git LFS files
    default: 'false'
```

### Pipeline Configuration

Configure in `.pipeline/config.yml`:

```yaml
general:
  # Define the productive branch name
  productiveBranch: 'main'

  # Enable pipeline optimization
  pipelineOptimization: false

  # Vault configuration
  vaultBasePath: 'your/vault/path'
  vaultPipelineName: 'your-pipeline-name'
  vaultNamespace: 'ies/hyperspace/pipelines'

  # Global extensions
  globalExtensionsRepository: 'owner/repo'
  globalExtensionsRef: 'main'

  # System Trust integration
  hooks:
    systemtrust:
      serverURL: 'https://system-trust.example.com'
```

## Example Usage

### Basic Init Stage Call

```yaml
jobs:
  init:
    name: Initialize Pipeline
    uses: project-piper/piper-pipeline-github/.github/workflows/init.yml@main
    with:
      piper-version: 'latest'
      runs-on: '[ "self-hosted" ]'
```

### Init with Custom Configuration

```yaml
jobs:
  init:
    name: Initialize Pipeline
    uses: project-piper/piper-pipeline-github/.github/workflows/init.yml@main
    with:
      piper-version: 'v1.150.0'
      runs-on: '[ "self-hosted", "linux" ]'
      checkout-submodules: 'true'
      custom-defaults-paths: '.pipeline/defaults.yml'
    secrets: inherit
```

### Using Init Outputs in Subsequent Jobs

```yaml
jobs:
  init:
    uses: project-piper/piper-pipeline-github/.github/workflows/init.yml@main
    with:
      piper-version: 'latest'

  build:
    needs: init
    if: fromJSON(needs.init.outputs.active-stages-map).Build == true
    uses: project-piper/piper-pipeline-github/.github/workflows/build.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.init.outputs.pipeline-env }}
```

## Required Secrets

The Init stage requires the following secrets (configure in repository/organization settings):

| Secret | Description |
|--------|-------------|
| `PIPER_VAULTAPPROLEID` | Vault AppRole ID for authentication |
| `PIPER_VAULTAPPROLESECRETID` | Vault AppRole Secret ID for authentication |
| `PIPER_ENTERPRISE_SERVER_URL` | Enterprise server URL (optional) |
| `PIPER_WDF_GITHUB_TOKEN` | WDF GitHub token (optional) |

## Environment Variables

The Init stage sets the following environment variables:

| Variable | Description |
|----------|-------------|
| `PIPER_ACTION_PIPER_VERSION` | Piper CLI version |
| `PIPER_ACTION_SAP_PIPER_VERSION` | SAP Piper library version |
| `PIPER_ACTION_GITHUB_ENTERPRISE_TOKEN` | GitHub token |
| `PIPER_ACTION_CUSTOM_DEFAULTS_PATHS` | Custom defaults paths |
| `PIPER_ACTION_CUSTOM_STAGE_CONDITIONS_PATH` | Custom stage conditions path |
| `PIPER_PIPELINE_TEMPLATE_NAME` | Template name for telemetry |
| `PIPER_PIPELINE_STAGE_TEMPLATE_NAME` | Stage template name (hyperspace-piper-init) |

## Troubleshooting

### Stage Configuration Not Found

If the Init stage fails to find the configuration file:

1. Ensure `.pipeline/config.yml` exists in your repository
2. Check that the repository checkout is successful
3. Verify file permissions

### Vault Authentication Failures

If Vault authentication fails:

1. Verify `PIPER_VAULTAPPROLEID` and `PIPER_VAULTAPPROLESECRETID` secrets are set
2. Check that `vaultBasePath` and `vaultPipelineName` are configured correctly
3. Ensure your Vault namespace is properly configured

### No Stages Activated

If no stages are activated after Init:

1. Check `.pipeline/config.yml` for stage activation rules
2. Review file patterns that trigger stages
3. Verify build tool configuration

## Best Practices

1. **Version Pinning**: Use specific Piper versions in production pipelines instead of 'latest'
2. **Vault Integration**: Always use Vault for secrets management in productive branches
3. **Branch Protection**: Configure productive branch in alignment with Git branch protection rules
4. **Monitoring**: Enable GCP event publishing for pipeline monitoring and analytics
5. **Extensions**: Use global extensions for shared pipeline logic across multiple repositories

# Piper Azure DevOps Task Extension

## Overview

The Piper Azure DevOps Task (`piper@1`) enables execution of Project Piper steps within Azure pipelines, handling binary management, configuration loading, and step execution.

## Installation

### Prerequisites

- Azure DevOps organization admin access
- GitHub Enterprise service connection (github.tools.sap)
- Vault credentials for authentication

### Installing

1. Navigate to organization settings → Extensions → Browse marketplace
2. Search for "Project Piper" or "piper-azure-task"
3. Install to your organization

**Marketplace**: [ProjectPiper.piper-azure-task](https://marketplace.visualstudio.com/items?itemName=ProjectPiper.piper-azure-task-dev)

**Note**: Extension requires installation permissions from Piper team.

## Basic Usage

```yaml
# Simple execution
- task: piper@1
  inputs:
    stepName: 'help'

# With parameters
- task: piper@1
  inputs:
    stepName: 'mavenBuild'
    flags: '--pomPath pom.xml'
    piperVersion: 'v1.150.0'
```

## Task Parameters

### Core Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `stepName` | string | Piper step to execute (required) | `help` |
| `flags` | string | Command-line flags | `""` |
| `piperVersion` | string | OS Piper version (exact) | `""` |
| `sapPiperVersion` | string | SAP Piper version (exact) | `""` |
| `customConfigLocation` | string | Config file path | `.pipeline/config.yml` |
| `customDefaults` | string | Custom defaults (multi-line) | `""` |

### Advanced Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `restorePipelineDefaults` | string | Base-64 encoded config |
| `exportPipelineEnv` | boolean | Export pipeline environment |
| `preserveDefaultConfig` | boolean | Export default config |
| `createCheckIfStepActiveMaps` | boolean | Create activation maps |

### Docker Parameters

| Parameter | Description |
|-----------|-------------|
| `dockerImage` | Docker image for step |
| `dockerOptions` | Docker run options |
| `dockerEnvVars` | Environment variables (JSON) |
| `sidecarImage` | Sidecar container image |
| `sidecarOptions` | Sidecar options |
| `sidecarEnvVars` | Sidecar environment (JSON) |

### Service Connections

| Parameter | Description |
|-----------|-------------|
| `gitHubConnection` | GitHub Enterprise connection |
| `dockerRegistryConnection` | Docker registry connection |
| `fetchPiperBinaryVersionTag` | Fetch binary version |

## Examples

### Custom Defaults

```yaml
inputs:
  customDefaults: |
    https://raw.githubusercontent.com/org/defaults/java.yml
    https://raw.githubusercontent.com/org/defaults/security.yml
```

### Docker Configuration

```yaml
inputs:
  stepName: 'mavenBuild'
  dockerImage: 'maven:3.8-openjdk-11'
  dockerOptions: '--memory=4g'
  dockerEnvVars: '{"MAVEN_OPTS": "-Xmx2048m"}'
```

### Sidecar Containers

```yaml
inputs:
  stepName: 'integrationTest'
  sidecarImage: 'postgres:13'
  sidecarEnvVars: '{"POSTGRES_PASSWORD": "test"}'
```

### Export Pipeline Environment

```yaml
- task: piper@1
  name: build
  inputs:
    stepName: 'mavenBuild'
    exportPipelineEnv: true

# Use in next stage
variables:
  pipelineEnvironment_b64: $[ stageDependencies.Build.job.outputs['build.PipelineEnv'] ]
```

## Configuration

### Piper Config File

Default: `.pipeline/config.yml`

```yaml
general:
  productiveBranch: 'main'
  vault:
    serverUrl: 'https://vault.example.com'

stages:
  Build:
    mavenBuild: true

steps:
  mavenBuild:
    goals: ['clean', 'install']
    pomPath: 'pom.xml'
```

### Pipeline Variables

**Required** (SAP-internal):
- `hyperspace.vault.roleId`
- `hyperspace.vault.secretId`

**Optional**:
- `hyperspace.piper.version`
- `hyperspace.sappiper.version`
- `pipelineEnvironment_b64`

## Usage Patterns

### Single Steps

```yaml
steps:
  - task: piper@1
    displayName: 'Version artifact'
    inputs:
      stepName: 'artifactPrepareVersion'

  - task: piper@1
    displayName: 'Build'
    inputs:
      stepName: 'mavenBuild'
```

### Multi-Stage Pipeline

```yaml
stages:
  - stage: Init
    jobs:
      - job: setup
        steps:
          - task: piper@1
            name: defaults
            inputs:
              stepName: 'version'
              preserveDefaultConfig: true

  - stage: Build
    dependsOn: Init
    variables:
      sapDefaults_b64: $[ stageDependencies.Init.setup.outputs['defaults.DefaultConfig'] ]
    jobs:
      - job: build
        steps:
          - task: piper@1
            inputs:
              stepName: 'mavenBuild'
              restorePipelineDefaults: $(sapDefaults_b64)
```

### Conditional Execution

```yaml
- task: piper@1
  condition: eq(variables['Build.SourceBranchName'], 'main')
  inputs:
    stepName: 'artifactPrepareVersion'
    flags: '--versioningType cloud'

- task: piper@1
  condition: ne(variables['Build.SourceBranchName'], 'main')
  inputs:
    stepName: 'artifactPrepareVersion'
    flags: '--versioningType cloud_noTag'
```

## Binary Management

### Version Selection

1. Task parameters (`piperVersion`, `sapPiperVersion`)
2. Pipeline variables
3. Built-in defaults

### Sources

- **Open-source**: GitHub releases (SAP/jenkins-library)
- **SAP Piper**: GitHub Enterprise (project-piper/sap-piper)

### Caching

Binaries cached in `$(Agent.ToolsDirectory)`.

### Development Versions

```yaml
variables:
  - name: hyperspace.piper.version
    value: 'feature-branch'
  - name: hyperspace.piper.isBranch
    value: true

steps:
  - task: GoTool@0
    inputs:
      version: '1.21'
  - task: piper@1
    inputs:
      stepName: 'mavenBuild'
```

## Best Practices

1. **Pin versions**: `piperVersion: 'v1.150.0'`
2. **Organization defaults**: Use `customDefaults`
3. **Descriptive names**: Clear `displayName` values
4. **Service connections**: No hardcoded credentials
5. **Appropriate images**: Match Docker image to needs
6. **Binary caching**: Implement in Init stage

## Troubleshooting

### Task Not Found
**Error**: `Task 'piper' not found`
**Solution**: Install extension in organization

### Binary Download Fails
**Solutions**:
- Check network connectivity
- Verify service connection
- Confirm version exists

### Config Not Found
**Solutions**:
- Ensure file exists
- Check `customConfigLocation`
- Verify not in `.gitignore`

### Docker Errors
**Solutions**:
- Ensure Docker running
- Check service connection
- Verify permissions

### Step Fails
**Solutions**:
- Add `flags: "--verbose"`
- Check `.pipeline/config.yml`
- Verify parameters
- Review Piper docs

## Quick Examples

### Node.js

```yaml
- task: piper@1
  inputs:
    stepName: 'npmExecuteScripts'
    flags: '--runScripts ci-build,ci-test'
```

### Java Maven

```yaml
- task: piper@1
  inputs:
    stepName: 'mavenBuild'
    dockerImage: 'maven:3.8-openjdk-11'

- task: piper@1
  inputs:
    stepName: 'sonarExecuteScan'
```

### Container Build

```yaml
- task: piper@1
  inputs:
    stepName: 'kanikoExecute'
    flags: '--createBOM --containerImageTag $(Build.BuildId)'
    dockerRegistryConnection: 'my-registry'
```

## Reference

### Task Metadata
- **ID**: `13d0fb42-a9b4-4bf6-90f2-6ae85059b638`
- **Name**: `piper`
- **Category**: `Utility`
- **Min Agent**: `3.232.1`
- **Execution**: `Node20_1`

### Output Variables
- `PipelineEnv` (when `exportPipelineEnv: true`)
- `DefaultConfig` (when `preserveDefaultConfig: true`)
- `binaryVersion`, `osBinaryVersion` (when `fetchPiperBinaryVersionTag: true`)

### Links
- [Task Repository](https://github.tools.sap/project-piper/piper-azure-task)
- [Piper Docs](https://www.project-piper.io)
- [Azure Tasks](https://docs.microsoft.com/en-us/azure/devops/pipelines/process/tasks)

# Configuration System Overview

Project Piper provides a flexible, hierarchical configuration system that allows you to configure CI/CD pipelines across different execution environments including Jenkins, Azure DevOps, and GitHub Actions.

## Table of Contents

- [Introduction](#introduction)
  - [Configuration Files](#configuration-files)
    - [Primary Configuration File](#primary-configuration-file)
    - [Default Configuration](#default-configuration)
    - [Custom Defaults](#custom-defaults)
  - [Configuration Structure](#configuration-structure)
    - [General Section](#general-section)
    - [Steps Section](#steps-section)
    - [Stages Section](#stages-section)
    - [Hooks Section](#hooks-section)
  - [Configuration Formats](#configuration-formats)
  - [Platform Support](#platform-support)
    - [Jenkins](#jenkins)
    - [Azure DevOps](#azure-devops)
    - [GitHub Actions](#github-actions)
  - [Key Concepts](#key-concepts)
    - [Configuration Inheritance](#configuration-inheritance)
    - [Parameter Overriding](#parameter-overriding)
    - [Context-Aware Execution](#context-aware-execution)
  - [Getting Started](#getting-started)
    - [Minimal Configuration](#minimal-configuration)
    - [Adding Custom Defaults](#adding-custom-defaults)
  - [Best Practices](#best-practices)
  - [Common Use Cases](#common-use-cases)
    - [Multi-Repository Configuration](#multi-repository-configuration)
    - [Environment-Specific Settings](#environment-specific-settings)
    - [Shared Team Configuration](#shared-team-configuration)
  - [Next Steps](#next-steps)

## Introduction

The Project Piper configuration system is designed to provide:

- **Hierarchical configuration**: Multiple layers of configuration that merge intelligently
- **Flexibility**: Configure at global, stage, or step level
- **Reusability**: Share configurations across multiple projects
- **Security**: Integration with HashiCorp Vault and credential management
- **Platform agnostic**: Works across Jenkins, Azure DevOps, and GitHub Actions

## Configuration Files

### Primary Configuration File

Your project's main configuration file is located at:

```
.pipeline/config.yml
```

This file should be committed to your repository's **master/main branch** and contains project-specific settings.

**Example:**
```yaml
general:
  buildTool: 'maven'
  productiveBranch: 'main'
  gitSshKeyCredentialsId: 'github-ssh-key'

steps:
  mavenExecute:
    goals: 'clean install'

stages:
  Build:
    mavenExecuteStaticCodeChecks: true
```

### Default Configuration

Project Piper ships with comprehensive defaults located at:

```
resources/default_pipeline_environment.yml
```

**Online reference:**
https://github.com/SAP/jenkins-library/blob/master/resources/default_pipeline_environment.yml

These defaults cover:
- Common step parameters
- Docker images for various build tools
- Standard tool configurations
- Default timeouts and thresholds

### Custom Defaults

Organizations can provide shared configuration defaults for multiple projects:

```yaml
customDefaults:
  - 'https://github.company.com/raw/org/defaults/backend-services.yml'
  - 'https://github.company.com/raw/org/defaults/compliance.yml'

general:
  buildTool: 'npm'
```

Custom defaults are:
- Loaded from URLs or file paths
- Applied in order (later items have higher precedence)
- Merged before project-specific configuration
- Useful for organization-wide policies

## Configuration Structure

### General Section

The `general` section contains parameters available across all steps and stages:

```yaml
general:
  buildTool: 'maven'                    # Build tool: maven, npm, docker, etc.
  productiveBranch: 'master'            # Main branch name
  gitSshKeyCredentialsId: 'git-key'    # SSH credentials for Git operations
  verbose: true                         # Enable detailed logging
  collectTelemetryData: false          # Disable telemetry

  # Vault configuration
  vaultServerUrl: 'https://vault.company.com'
  vaultNamespace: 'team/project'
  vaultPath: 'piper/my-project'

  # Change management
  changeManagement:
    type: 'NONE'                        # SOLMAN, CTS, NONE
```

### Steps Section

The `steps` section configures individual pipeline steps:

```yaml
steps:
  cloudFoundryDeploy:
    deployTool: 'cf_native'
    cloudFoundry:
      org: 'my-org'
      space: 'production'
      apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'
      credentialsId: 'CF_CREDENTIALS'

  mavenExecute:
    goals: 'clean verify'
    defines: '-Dmaven.test.skip=false'

  dockerExecute:
    dockerImage: 'maven:3.8-openjdk-11'
```

### Stages Section

The `stages` section provides stage-specific configuration:

```yaml
stages:
  Build:
    mavenExecuteStaticCodeChecks: true
    mavenExecuteIntegration: false

  Acceptance:
    cloudFoundryDeploy:
      space: 'acceptance'

  Release:
    cloudFoundryDeploy:
      space: 'production'
```

### Hooks Section

The `hooks` section configures integrations like monitoring and alerting:

```yaml
hooks:
  splunk:
    dsn: 'https://splunk.company.com:8088'
    token: 'splunk-hec-token'
    index: 'ci-cd'
    sendLogs: true

  ans:
    serviceKeyCredentialsId: 'ans-service-key'
```

## Configuration Formats

Project Piper supports two primary formats:

**YAML (Recommended):**
```yaml
general:
  buildTool: 'npm'
  verbose: true
```

**JSON (Supported):**
```json
{
  "general": {
    "buildTool": "npm",
    "verbose": true
  }
}
```

## Platform Support

### Jenkins

**Jenkinsfile:**
```groovy
@Library('piper-lib-os') _

piperPipeline script: this
```

**Features:**
- Full pipeline orchestration
- Native Jenkins credentials integration
- Kubernetes pod execution
- Extensive plugin support

### Azure DevOps

**Azure Pipeline YAML:**
```yaml
steps:
  - task: Piper@1
    inputs:
      piperCommand: 'version'
      flags: '--noTelemetry'
```

**Features:**
- Azure-native task execution
- Azure Key Vault integration
- Azure Container Registry support

### GitHub Actions

**Workflow YAML:**
```yaml
- uses: SAP/project-piper-action@main
  with:
    piper-version: 'latest'
    command: 'version'
```

**Features:**
- GitHub-native secret management
- GitHub Container Registry support
- GitHub API integration

## Key Concepts

### Configuration Inheritance

Configuration flows from general to specific:

```
Default Config → Custom Defaults → General → Steps/Stages → Direct Parameters
```

### Parameter Overriding

Later configuration sources override earlier ones:

```yaml
# In defaults
steps:
  mavenExecute:
    goals: 'clean install'

# In .pipeline/config.yml (overrides)
steps:
  mavenExecute:
    goals: 'clean verify deploy'
```

### Context-Aware Execution

Piper automatically activates stages based on:
- Available configuration
- File patterns in repository
- Build tool detection
- Branch context

## Getting Started

### Minimal Configuration

For pull request validation:

```yaml
general:
  buildTool: 'npm'
```

For full pipeline:

```yaml
general:
  buildTool: 'maven'
  productiveBranch: 'main'
  gitSshKeyCredentialsId: 'github-key'
```

### Adding Custom Defaults

**Jenkinsfile:**
```groovy
@Library('piper-lib-os') _

piperPipeline script: this, customDefaults: ['org-defaults.yml']
```

**config.yml:**
```yaml
customDefaults:
  - 'https://github.company.com/raw/org/defaults/java-backend.yml'

general:
  buildTool: 'maven'
```

## Best Practices

1. **Start Small**: Begin with minimal configuration and add as needed
2. **Use Custom Defaults**: Share common settings across projects
3. **Document Overrides**: Comment why you override defaults
4. **Secure Secrets**: Use Vault or platform credential stores
5. **Version Control**: Keep configuration in Git
6. **Test Changes**: Use feature branches to test configuration changes
7. **Validate YAML**: Use YAML linters before committing
8. **Follow Conventions**: Use standard parameter names
9. **Enable Verbose Mode**: During initial setup for debugging
10. **Review Defaults**: Understand what's inherited from defaults

## Common Use Cases

### Multi-Repository Configuration

Organization defaults file (`org-defaults.yml`):
```yaml
general:
  vaultServerUrl: 'https://vault.company.com'
  vaultBasePath: 'piper'
  collectTelemetryData: false

steps:
  checksPublishResults:
    pmd:
      active: true
    checkstyle:
      active: true
```

### Environment-Specific Settings

```yaml
stages:
  Acceptance:
    cloudFoundryDeploy:
      space: 'dev'

  Release:
    cloudFoundryDeploy:
      space: 'production'
      smokeTest: true
```

### Shared Team Configuration

Custom defaults for microservices team:
```yaml
general:
  buildTool: 'maven'
  dockerPullImage: true

steps:
  mavenExecute:
    dockerImage: 'maven:3.8-openjdk-17'

  dockerExecute:
    dockerOptions:
      - '--memory=4g'
      - '--cpus=2'
```

## Next Steps

- [Configuration Hierarchy](01-configuration-hierarchy.md) - Understand the 7-level precedence system
- [Default Settings](02-default-settings.md) - Explore built-in defaults
- [Platform Deviations](03-platform-deviations.md) - Learn platform-specific differences
- [Stage Configuration](04-stage-configuration.md) - Configure pipeline stages
- [Step Configuration](05-step-configuration.md) - Configure individual steps
- [Credentials Management](06-credentials-management.md) - Secure secrets handling

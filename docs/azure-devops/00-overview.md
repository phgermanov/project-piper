# Azure DevOps Integration Overview

## Introduction

Project Piper provides comprehensive integration with Azure DevOps through two main components:

1. **piper-azure-task**: An Azure DevOps task extension that allows you to execute individual Piper steps
2. **piper-pipeline-azure**: A complete general purpose pipeline template with pre-configured stages

This integration enables you to leverage Piper's powerful CI/CD capabilities within Azure DevOps, providing standardized and automated build, test, security scanning, and deployment workflows.

## Key Features

### Piper Azure Task Extension

- Execute any Piper step directly in your Azure pipeline
- Support for both open-source and SAP-internal Piper binaries
- Docker container support for isolated execution
- Configurable step parameters and environment variables
- Built-in caching for improved performance
- Service connection integration (GitHub, Docker registries)

### General Purpose Pipeline Templates

- Pre-configured multi-stage pipeline (Init, Build, Security, Integration, Acceptance, Performance, Release, etc.)
- Automatic stage activation based on project configuration
- Extensible through pre/post step hooks
- Support for multiple build tools (Maven, Gradle, npm, Go, Python, MTA, etc.)
- Integrated security scanning (SAST, OSS, IP scan)
- Manual confirmation gates for production deployments
- Optimized for both Microsoft-hosted and self-hosted agents

## Architecture

### Component Relationship

```
┌─────────────────────────────────────────────────────────────┐
│                    Azure DevOps Pipeline                     │
│                                                               │
│  ┌────────────────────────────────────────────────────────┐ │
│  │         Piper General Purpose Pipeline                  │ │
│  │         (piper-pipeline-azure)                          │ │
│  │                                                          │ │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐             │ │
│  │  │   Init   │→ │   Build  │→ │ Security │→ ...        │ │
│  │  └────┬─────┘  └────┬─────┘  └────┬─────┘             │ │
│  │       │             │             │                     │ │
│  │       v             v             v                     │ │
│  │  ┌─────────────────────────────────────────────────┐   │ │
│  │  │         Piper Azure Task (piper@1.x)            │   │ │
│  │  └─────────────────────┬───────────────────────────┘   │ │
│  └────────────────────────┼───────────────────────────────┘ │
│                           │                                  │
│                           v                                  │
│                  ┌─────────────────┐                        │
│                  │  Piper Binary   │                        │
│                  └─────────────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

### Execution Flow

1. **Pipeline Definition**: You define an `azure-pipelines.yml` that references the Piper pipeline templates
2. **Template Loading**: Azure DevOps loads the pipeline templates from the `piper-pipeline-azure` repository
3. **Stage Execution**: Each stage uses the `piper` task to execute specific Piper steps
4. **Binary Management**: The task downloads and caches the appropriate Piper binary version
5. **Step Execution**: Piper executes the configured step with your project configuration

## Supported CI/CD Scenarios

### Build Scenarios

- **Java/JVM Projects**: Maven and Gradle builds with dependency management
- **Node.js Projects**: npm-based builds with script execution
- **Go Projects**: Native Go compilation with module support
- **Python Projects**: Python builds with pip/poetry support
- **Multi-Target Applications (MTA)**: SAP MTA builds for Cloud Foundry/XSA
- **Container Images**: Docker image building with Kaniko or Cloud Native Buildpacks

### Testing Scenarios

- Unit testing with various frameworks
- Integration testing with service containers
- Acceptance testing in dedicated stage
- Performance testing capabilities
- Code coverage reporting (JaCoCo, Cobertura)

### Security Scenarios

- Static Application Security Testing (SAST) with SonarQube
- Open Source Software (OSS) vulnerability scanning
- IP scan for license compliance
- PPMS integration for SAP compliance
- Container image scanning

### Deployment Scenarios

- Cloud Foundry deployments
- Kubernetes deployments with Helm
- Container registry publishing
- Artifact repository uploads
- SAP-specific deployment targets

## Prerequisites

### Azure DevOps Organization Setup

1. **Extension Installation**: The Piper Azure Task extension must be installed in your Azure DevOps organization
2. **Service Connections**: Configure required service connections:
   - GitHub Enterprise connection for template repository access
   - Docker registry connections (if using private registries)
   - Additional service connections as needed by your project

### Project Requirements

1. **Piper Configuration**: Create a `.pipeline/config.yml` in your repository root
2. **Vault Credentials**: Set up required vault credentials as pipeline variables:
   - `hyperspace.vault.roleId`
   - `hyperspace.vault.secretId`

### Access Requirements

- Access to the `piper-pipeline-azure` repository on GitHub Enterprise
- Appropriate permissions in your Azure DevOps organization
- Access to required service endpoints and registries

## Quick Start

### Minimal Setup

Create an `azure-pipelines.yml` in your repository root:

```yaml
# Using Piper general purpose pipeline for Azure

trigger:
- main

resources:
  repositories:
    - repository: piper-templates
      endpoint: github.tools.sap
      type: githubenterprise
      name: project-piper/piper-pipeline-azure
      ref: refs/tags/v1.2.3  # Use specific version

extends:
  template: sap-piper-pipeline.yml@piper-templates
```

### Basic Configuration

Create `.pipeline/config.yml`:

```yaml
general:
  productiveBranch: 'main'

stages:
  Build:
    npmExecuteScripts: true

steps:
  npmExecuteScripts:
    runScripts:
      - 'build'
      - 'test'
```

### Advanced Setup with Customization

```yaml
trigger:
- main
- develop

resources:
  repositories:
    - repository: piper-templates
      endpoint: github.tools.sap
      type: githubenterprise
      name: project-piper/piper-pipeline-azure
      ref: refs/tags/v1.2.3

extends:
  template: sap-piper-pipeline.yml@piper-templates
  parameters:
    # Version control
    piperVersion: 'v1.150.0'
    sapPiperVersion: 'v2.45.0'

    # Custom defaults
    customDefaults: |
      https://raw.githubusercontent.com/my-org/my-defaults.yml

    # Checkout options
    checkoutSubmodule: true
    checkoutLFS: false

    # Agent pools
    msHostedPool: 'Azure Pipelines'
    poolVmImage: 'ubuntu-latest'
    selfHostedPool: 'Self-hosted'

    # Build stage customization
    buildPreSteps:
      - script: echo "Running pre-build steps"
        displayName: Pre-build preparation

    buildPostSteps:
      - script: echo "Running post-build steps"
        displayName: Post-build tasks

    # Service containers for integration testing
    integrationServiceContainers:
      postgres:
        image: postgres:13
        env:
          POSTGRES_PASSWORD: test
        ports:
          - 5432:5432

    # Skip optional stages
    skipAcceptanceStage: false
    skipPerformanceStage: true
    skipReleaseStage: false

    # Manual confirmation
    manualConfirmationTimeoutMinutes: 43200  # 30 days
```

## Configuration Management

### Configuration Hierarchy

Piper uses a hierarchical configuration system:

1. **Pipeline Defaults**: Built-in defaults from Piper
2. **Custom Defaults**: Organization/team-specific defaults
3. **Project Config**: Project-specific `.pipeline/config.yml`
4. **Step Parameters**: Inline parameters in task calls

Configuration is merged with later sources overriding earlier ones.

### Environment Variables

Key environment variables set by the pipeline:

- `hyperspace.piper.version`: Open-source Piper version
- `hyperspace.sappiper.version`: SAP Piper version
- `piper.pipeline.template.name`: Pipeline template identifier
- `PIPER_PIPELINE_TEMPLATE_NAME`: Template name for telemetry

### Secrets Management

Sensitive data should be stored in:

1. **Azure Pipeline Variables**: For pipeline-level secrets
2. **Vault Integration**: For application secrets (via vault credentials)
3. **Service Connections**: For external system authentication

## Best Practices

### Version Pinning

Always pin specific versions for production pipelines:

```yaml
resources:
  repositories:
    - repository: piper-templates
      ref: refs/tags/v1.2.3  # Don't use 'main'

extends:
  template: sap-piper-pipeline.yml@piper-templates
  parameters:
    piperVersion: 'v1.150.0'      # Exact version
    sapPiperVersion: 'v2.45.0'    # Not 'latest'
```

### Caching Strategy

- Enable binary caching for faster pipeline execution
- Cache is automatically handled by the Init stage
- On self-hosted agents, caching can be disabled if binaries are pre-installed

### Branch Protection

Configure different behavior for productive vs. development branches:

```yaml
general:
  productiveBranch: 'main|release/.*'
```

Certain steps (like artifact uploads, deployments) only run on productive branches.

### Custom Stage Extensions

Use pre/post steps to extend stages without modifying templates:

```yaml
parameters:
  buildPreSteps:
    - task: SomeCustomTask@1
      inputs:
        setting: value
```

### Service Container Usage

Define service containers for integration tests:

```yaml
parameters:
  integrationServiceContainers:
    redis:
      image: redis:6-alpine
      ports:
        - 6379:6379
    postgres:
      image: postgres:13
      env:
        POSTGRES_PASSWORD: test
      ports:
        - 5432:5432
```

### Error Handling

- Use `continueOnError: true` for optional steps
- Configure retry logic for flaky operations
- Set appropriate timeouts for long-running operations

## Comparison with Other CI/CD Platforms

### vs. Jenkins

| Feature | Jenkins (Piper) | Azure DevOps (Piper) |
|---------|-----------------|---------------------|
| Pipeline Definition | Groovy-based Jenkinsfile | YAML-based azure-pipelines.yml |
| Template System | Shared Libraries | Template Repositories |
| Agent Management | Jenkins Agents | Azure Agent Pools |
| UI | Jenkins Blue Ocean | Azure DevOps UI |
| Artifact Storage | Jenkins/Artifactory | Azure Artifacts |

### vs. GitHub Actions

| Feature | GitHub Actions (Piper) | Azure DevOps (Piper) |
|---------|----------------------|---------------------|
| Workflow Definition | YAML workflows | YAML pipelines |
| Marketplace | GitHub Marketplace Actions | Azure DevOps Extensions |
| Execution Environment | GitHub-hosted/self-hosted runners | Microsoft-hosted/self-hosted agents |
| Integration | Native GitHub integration | Service connection required |
| Matrix Builds | Native support | Manual configuration |

## Migration Guide

### From Jenkins to Azure DevOps

1. **Convert Jenkinsfile to azure-pipelines.yml**
2. **Map Jenkins stages to Azure stages**
3. **Configure service connections** (equivalent to Jenkins credentials)
4. **Migrate environment variables** to Azure pipeline variables
5. **Update Piper configuration** if needed (mostly compatible)

### From Custom Azure Pipelines

1. **Identify existing stages/jobs**
2. **Map to Piper stage structure**
3. **Configure stage activation** in `.pipeline/config.yml`
4. **Use pre/post steps** for custom logic
5. **Test incrementally** stage by stage

## Troubleshooting

### Common Issues

**Issue**: Extension not found
- **Solution**: Ensure the Piper Azure Task extension is installed in your organization

**Issue**: Binary download fails
- **Solution**: Check network connectivity and GitHub service connection

**Issue**: Stage skipped unexpectedly
- **Solution**: Verify stage conditions and `checkIfStepActive` configuration

**Issue**: Docker permission errors
- **Solution**: Configure Docker registry service connections

### Debug Mode

Enable verbose logging:

```yaml
- task: piper@1
  inputs:
    stepName: yourStep
    flags: "--verbose"
```

### Getting Help

- **Documentation**: [Project Piper Documentation](https://www.project-piper.io)
- **GitHub Issues**: Report issues in respective repositories
- **ServiceNow**: For SAP-internal support

## Next Steps

- [Learn about the Piper Azure Task](01-azure-task.md)
- [Explore Pipeline Templates in detail](02-pipeline-templates.md)
- Review example projects and advanced configurations
- Set up your first Piper pipeline on Azure DevOps

## Additional Resources

- [Azure DevOps YAML Schema](https://docs.microsoft.com/en-us/azure/devops/pipelines/yaml-schema)
- [Azure DevOps Templates](https://docs.microsoft.com/en-us/azure/devops/pipelines/process/templates)
- [Project Piper Configuration](https://www.project-piper.io/configuration/)
- [Piper Steps Reference](https://www.project-piper.io/steps/)

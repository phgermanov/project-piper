# Project Piper Action - Overview

## Introduction

The Project Piper Action is a GitHub Action that enables running [Project "Piper"](https://www.project-piper.io/) steps in GitHub Actions workflows. Project "Piper" is a comprehensive CI/CD tooling suite specifically designed for the SAP Ecosystem, providing ready-to-use, opinionated building blocks for building continuous delivery pipelines.

## What is Project Piper Action?

The Project Piper Action serves as a bridge between GitHub Actions and the Piper CLI toolset, allowing you to:

- Execute any Piper step directly in your GitHub Actions workflows
- Leverage SAP-specific CI/CD capabilities in a cloud-native environment
- Benefit from consistent tooling across different CI/CD platforms (Jenkins, Azure DevOps, GitHub Actions)
- Build production-ready pipelines for SAP applications with minimal configuration

## Core Concepts

### Piper Steps

Piper steps are individual, self-contained units of functionality that perform specific CI/CD tasks. Examples include:

- **Build Steps**: `mavenBuild`, `npmExecuteScripts`, `golangBuild`
- **Test Steps**: `npmExecuteTests`, `mavenExecuteTest`, `batsExecuteTests`
- **Security Steps**: `detectExecuteScan`, `checkmarxExecuteScan`, `fortifyExecuteScan`
- **Deployment Steps**: `cloudFoundryDeploy`, `kubernetesDeploy`, `helmExecute`
- **Quality Steps**: `sonarExecuteScan`, `whitesourceExecuteScan`

Each step is designed to handle a specific aspect of the CI/CD process with sensible defaults and extensive configuration options.

### Action Architecture

The Project Piper Action follows a modular architecture:

```
┌─────────────────────────────────────────────────┐
│         GitHub Actions Workflow                 │
└──────────────────┬──────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────┐
│        Project Piper Action (Node.js)           │
│  ┌─────────────────────────────────────────┐   │
│  │  Configuration Management               │   │
│  │  - Action inputs                        │   │
│  │  - Environment variables                │   │
│  │  - Custom defaults                      │   │
│  └─────────────────────────────────────────┘   │
│                                                  │
│  ┌─────────────────────────────────────────┐   │
│  │  Binary Management                      │   │
│  │  - Download/cache Piper binary          │   │
│  │  - Version selection                    │   │
│  │  - Build from source (devel mode)       │   │
│  └─────────────────────────────────────────┘   │
│                                                  │
│  ┌─────────────────────────────────────────┐   │
│  │  Container Management                   │   │
│  │  - Docker container execution           │   │
│  │  - Sidecar containers                   │   │
│  │  - Network configuration                │   │
│  └─────────────────────────────────────────┘   │
│                                                  │
│  ┌─────────────────────────────────────────┐   │
│  │  Pipeline Environment (CPE)             │   │
│  │  - Load/export pipeline state           │   │
│  │  - Share data between steps/jobs        │   │
│  └─────────────────────────────────────────┘   │
└──────────────────┬──────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────┐
│          Piper Binary Execution                 │
│  (Open Source or SAP Internal)                  │
└─────────────────────────────────────────────────┘
```

## Key Components

### 1. Piper Binary Management

The action automatically handles downloading and caching of the Piper binary:

- **Version Selection**: Supports specific versions (e.g., `v1.300.0`), `latest`, or `master`
- **Automatic Caching**: Downloaded binaries are cached in the workflow workspace to avoid repeated downloads
- **Development Mode**: Can build Piper from source for testing unreleased features
- **Dual Binaries**: Supports both open-source (`piper`) and SAP-internal (`sap-piper`) binaries

### 2. Configuration Management

Piper's configuration system is highly flexible:

- **Configuration Files**: `.pipeline/config.yml` in your repository
- **Custom Defaults**: Load shared configurations from other repositories
- **Step Parameters**: Pass additional flags via the action's `flags` input
- **Environment Variables**: Configure steps using `PIPER_*` environment variables
- **Configuration Precedence**: Step parameters > Environment variables > Custom defaults > Default configuration

### 3. Docker Integration

The action provides seamless Docker container support:

- **Containerized Execution**: Run Piper steps inside Docker containers for consistent environments
- **Sidecar Containers**: Start additional containers (e.g., databases, services) for testing
- **Network Management**: Automatic network creation for container communication
- **Volume Mounting**: Workspace and Piper binary are automatically mounted
- **Environment Propagation**: GitHub Actions variables and secrets are passed to containers

### 4. Common Pipeline Environment (CPE)

The CPE is Piper's mechanism for sharing state across steps and jobs:

- **State Management**: Store and retrieve values between different steps
- **Cross-Job Communication**: Export pipeline environment from one job and import in another
- **Automatic Persistence**: Environment is stored in `.pipeline/commonPipelineEnvironment/`
- **Artifact Support**: Can be uploaded/downloaded as GitHub Actions artifacts

## Why Use Project Piper Action?

### Benefits

1. **SAP-Specific Expertise**: Built-in support for SAP technologies (ABAP, CAP, UI5, etc.)
2. **Reduced Boilerplate**: Predefined steps eliminate the need for custom scripts
3. **Best Practices**: Steps follow SAP and community best practices
4. **Consistent Experience**: Same steps work across Jenkins, Azure DevOps, and GitHub Actions
5. **Active Maintenance**: Regular updates and security patches from SAP
6. **Extensive Documentation**: Comprehensive docs at [project-piper.io](https://www.project-piper.io/)

### Use Cases

- Building and deploying SAP Cloud Application Programming (CAP) applications
- ABAP Environment (Steampunk) continuous integration
- UI5/Fiori application deployment
- Integration with SAP services (Cloud Foundry, Kubernetes)
- Security scanning and compliance checks for SAP applications
- Multi-technology projects combining Java, Node.js, and SAP technologies

## Comparison with Direct CLI Usage

Instead of writing custom scripts:

```yaml
# Without Piper Action - Custom Script
- name: Build Maven Project
  run: |
    mvn clean install -DskipTests
    mvn package
    # ... more custom commands
```

You can use the action:

```yaml
# With Piper Action - One Line
- uses: SAP/project-piper-action@main
  with:
    step-name: mavenBuild
    flags: '--publish --createBOM'
```

## Repository Structure

The Project Piper Action repository contains:

```
project-piper-action/
├── action.yml              # Action metadata and input definitions
├── src/                    # TypeScript source code
│   ├── main.ts            # Entry point
│   ├── piper.ts           # Core execution logic
│   ├── config.ts          # Configuration management
│   ├── docker.ts          # Container management
│   ├── download.ts        # Binary download
│   ├── pipelineEnv.ts     # CPE handling
│   └── ...
├── dist/                   # Compiled JavaScript
├── test/                   # Unit tests
└── docs/                   # Documentation
```

## Getting Started

To start using the Project Piper Action in your workflow:

1. Add the action to your workflow file
2. Specify the Piper step to execute
3. Optionally configure the step using flags or configuration files
4. Run your workflow

Basic example:

```yaml
name: CI
on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: SAP/project-piper-action@main
        with:
          step-name: version
```

This will download the Piper binary and display its version, confirming the action is working correctly.

## Version Management

The action supports multiple versioning strategies:

### For the Action Itself

- **`@main`**: Latest development version (may have breaking changes)
- **`@v1`**: Latest stable v1.x.x release (recommended for production)
- **`@v1.2.3`**: Specific version for maximum stability

### For the Piper Binary

- **`latest`**: Most recent released version (recommended)
- **`master`**: Latest build from master branch (deprecated)
- **`v1.300.0`**: Specific release version
- **`devel:org:repo:commit`**: Build from source at specific commit (advanced)

## System Requirements

- **GitHub Actions Runner**: Ubuntu, macOS, or Windows (Ubuntu recommended)
- **Node.js**: Node 20 (bundled in action)
- **Docker**: Required if using `docker-image` input
- **Disk Space**: Varies by step (typically 500MB-2GB for Piper binary and dependencies)

## Support and Resources

- **Official Documentation**: [https://www.project-piper.io/](https://www.project-piper.io/)
- **GitHub Repository**: [https://github.com/SAP/project-piper-action](https://github.com/SAP/project-piper-action)
- **Jenkins Library**: [https://github.com/SAP/jenkins-library](https://github.com/SAP/jenkins-library) (contains Piper CLI source)
- **Community**: GitHub Issues and Discussions
- **SAP Community**: Questions tagged with `project-piper`

## License

The Project Piper Action is released under the Apache 2.0 license.

## Next Steps

- **[Features and Capabilities](01-features.md)**: Learn about all features in detail
- **[Usage Guide](02-usage-guide.md)**: Step-by-step instructions and examples
- **[Piper Steps](https://www.project-piper.io/steps/)**: Browse available Piper steps

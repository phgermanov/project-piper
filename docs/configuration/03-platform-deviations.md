# Platform Deviations

Project Piper supports multiple CI/CD platforms with platform-specific configurations and behavior differences. This guide covers Jenkins, Azure DevOps, and GitHub Actions.

## Table of Contents

- [Platform Deviations](#platform-deviations)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Platform Comparison](#platform-comparison)
  - [Jenkins Configuration](#jenkins-configuration)
    - [Basic Setup](#basic-setup)
    - [Jenkins-Specific Settings](#jenkins-specific-settings)
    - [Credentials Management](#credentials-management)
    - [Kubernetes Execution](#kubernetes-execution)
    - [Extension Points](#extension-points)
    - [Environment Variables](#environment-variables)
  - [Azure DevOps Configuration](#azure-devops-configuration)
    - [Basic Setup](#basic-setup-1)
    - [Task Configuration](#task-configuration)
    - [Azure-Specific Settings](#azure-specific-settings)
    - [Credentials Management](#credentials-management-1)
    - [Service Connections](#service-connections)
    - [Environment Variables](#environment-variables-1)
  - [GitHub Actions Configuration](#github-actions-configuration)
    - [Basic Setup](#basic-setup-2)
    - [Action Configuration](#action-configuration)
    - [GitHub-Specific Settings](#github-specific-settings)
    - [Secrets Management](#secrets-management)
    - [GitHub Services](#github-services)
    - [Environment Variables](#environment-variables-2)
  - [Credential Management Comparison](#credential-management-comparison)
    - [Jenkins Credentials](#jenkins-credentials)
    - [Azure Key Vault](#azure-key-vault)
    - [GitHub Secrets](#github-secrets)
  - [Vault Integration Differences](#vault-integration-differences)
    - [Jenkins Vault Integration](#jenkins-vault-integration)
    - [Azure DevOps Vault Integration](#azure-devops-vault-integration)
    - [GitHub Actions Vault Integration](#github-actions-vault-integration)
  - [Docker Execution Differences](#docker-execution-differences)
    - [Jenkins](#jenkins)
    - [Azure DevOps](#azure-devops)
    - [GitHub Actions](#github-actions)
  - [Common Configuration](#common-configuration)
  - [Platform Migration Guide](#platform-migration-guide)
    - [Jenkins to Azure DevOps](#jenkins-to-azure-devops)
    - [Jenkins to GitHub Actions](#jenkins-to-github-actions)
  - [Best Practices by Platform](#best-practices-by-platform)
    - [Jenkins Best Practices](#jenkins-best-practices)
    - [Azure DevOps Best Practices](#azure-devops-best-practices)
    - [GitHub Actions Best Practices](#github-actions-best-practices)

## Overview

Project Piper provides a **unified configuration model** across platforms while accommodating platform-specific features and constraints.

**Core Principle**: Write configuration once in `.pipeline/config.yml`, use across platforms with minimal platform-specific adjustments.

**Platform Support Matrix**:

| Feature | Jenkins | Azure DevOps | GitHub Actions |
|---------|---------|--------------|----------------|
| Full Pipeline | Yes | Limited | Limited |
| Individual Steps | Yes | Yes | Yes |
| Vault Integration | Yes | Yes | Yes |
| Kubernetes Execution | Yes | Via AKS | Via Runners |
| Custom Defaults | Yes | Yes | Yes |
| Stage Configuration | Yes | Limited | Limited |
| Extension Mechanism | Yes | Limited | Limited |

## Platform Comparison

**Jenkins** (Primary Platform):
- Full-featured pipeline orchestration
- Native Groovy-based pipelines
- Complete stage support
- Rich extension mechanism
- Mature credential management

**Azure DevOps**:
- Task-based execution
- YAML pipelines
- Limited stage support
- Azure-native integrations
- Azure Key Vault integration

**GitHub Actions**:
- Workflow-based execution
- YAML workflows
- Limited stage support
- GitHub-native integrations
- GitHub Secrets integration

## Jenkins Configuration

### Basic Setup

**Jenkinsfile** (declarative):
```groovy
@Library('piper-lib-os') _

piperPipeline script: this
```

**Jenkinsfile** (with custom defaults):
```groovy
@Library(['piper-lib-os', 'company-library']) _

piperPipeline script: this, customDefaults: ['org-defaults.yml']
```

**Configuration** (`.pipeline/config.yml`):
```yaml
general:
  buildTool: 'maven'
  productiveBranch: 'main'
  gitSshKeyCredentialsId: 'github-ssh-key'

steps:
  mavenExecute:
    goals: 'clean verify'
```

### Jenkins-Specific Settings

**Kubernetes Execution**:
```yaml
general:
  jenkinsKubernetes:
    jnlpAgent: 'jenkins/inbound-agent:jdk17'
    securityContext:
      # runAsUser: 1000
      # fsGroup: 1000

steps:
  dockerExecuteOnKubernetes:
    containerCommand: '/kaniko/executor'
    containerShell: '/busybox/sh'
```

**Node Labels**:
```yaml
steps:
  piperStageWrapper:
    nodeLabel: 'docker'  # Run on nodes with 'docker' label

stages:
  Build:
    nodeLabel: 'maven-builder'
```

**Shared Libraries**:
```groovy
// Load multiple libraries
@Library(['piper-lib-os@v1.220.0', 'company-shared-lib@main']) _
```

### Credentials Management

**Jenkins Credential Types**:

```yaml
# SSH Key
general:
  gitSshKeyCredentialsId: 'github-ssh-key'

# Username/Password
steps:
  cloudFoundryDeploy:
    cloudFoundry:
      credentialsId: 'cf-credentials'  # Username+Password

# Secret Text
steps:
  sonarExecuteScan:
    sonarTokenCredentialsId: 'sonar-token'

# Secret File
steps:
  gcpPublish:
    serviceAccountKeyFileCredentialsId: 'gcp-key-file'
```

**Creating Jenkins Credentials**:
```groovy
// In Jenkins UI: Manage Jenkins > Credentials
// Or via Groovy script:
import com.cloudbees.plugins.credentials.*
import com.cloudbees.plugins.credentials.impl.*
import com.cloudbees.plugins.credentials.domains.*

def credentials = new UsernamePasswordCredentialsImpl(
  CredentialsScope.GLOBAL,
  "my-cred-id",
  "Description",
  "username",
  "password"
)

SystemCredentialsProvider.getInstance().getStore().addCredentials(
  Domain.global(),
  credentials
)
```

### Kubernetes Execution

**Pod Template Configuration**:
```yaml
steps:
  dockerExecuteOnKubernetes:
    dockerImage: 'maven:3.8-openjdk-17'
    containerCommand: ''
    containerShell: '/bin/bash'

    # Resource limits
    resources:
      limits:
        cpu: '2'
        memory: '4Gi'
      requests:
        cpu: '1'
        memory: '2Gi'

    # Volume mounts
    volumes:
      - name: 'docker-socket'
        hostPath:
          path: '/var/run/docker.sock'
```

### Extension Points

**Stage Extensions**:
```yaml
steps:
  piperStageWrapper:
    projectExtensionsDirectory: '.pipeline/extensions/'
```

**Extension File** (`.pipeline/extensions/Build.groovy`):
```groovy
void call(Map parameters) {
    echo "Custom Build extension"

    // Pre-build actions
    sh 'echo "Running custom checks"'

    // Continue with standard stage
    piperStageWrapper(parameters)

    // Post-build actions
    sh 'echo "Build completed"'
}
```

### Environment Variables

**Declarative Pipeline**:
```groovy
pipeline {
    agent any

    environment {
        PIPER_verbose = 'true'
        PIPER_buildTool = 'maven'
    }

    stages {
        stage('Build') {
            steps {
                script {
                    @Library('piper-lib-os') _
                    piperPipeline script: this
                }
            }
        }
    }
}
```

## Azure DevOps Configuration

### Basic Setup

**azure-pipelines.yml**:
```yaml
trigger:
  branches:
    include:
      - main
      - develop

pool:
  vmImage: 'ubuntu-latest'

steps:
  - task: Piper@1
    displayName: 'Execute Piper'
    inputs:
      piperCommand: 'version'
      flags: '--noTelemetry'
```

**With Configuration File**:
```yaml
steps:
  - task: Piper@1
    displayName: 'Build with Maven'
    inputs:
      piperCommand: 'mavenBuild'
      flags: '--verbose'
      dockerImage: 'maven:3.8-openjdk-17'
```

### Task Configuration

**Piper Task Inputs**:

```yaml
steps:
  - task: Piper@1
    inputs:
      # Required
      piperCommand: 'mavenBuild'  # Piper command to execute

      # Optional
      flags: '--verbose'           # Additional flags
      dockerImage: 'maven:3.8-openjdk-17'  # Docker image
      dockerOptions: '--memory=4g'  # Docker options
      stepConfigJSON: '$(stepConfigJSON)'  # JSON config
```

**Multi-Step Pipeline**:
```yaml
stages:
  - stage: Build
    jobs:
      - job: BuildJob
        steps:
          - task: Piper@1
            displayName: 'Maven Build'
            inputs:
              piperCommand: 'mavenBuild'
              dockerImage: 'maven:3.8-openjdk-17'

  - stage: Deploy
    jobs:
      - job: DeployJob
        steps:
          - task: Piper@1
            displayName: 'Cloud Foundry Deploy'
            inputs:
              piperCommand: 'cloudFoundryDeploy'
```

### Azure-Specific Settings

**Configuration** (`.pipeline/config.yml`):
```yaml
general:
  buildTool: 'maven'
  verbose: true

steps:
  mavenBuild:
    goals: 'clean verify'

  cloudFoundryDeploy:
    cloudFoundry:
      org: 'my-org'
      space: 'production'
      apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'
```

**Variable Groups**:
```yaml
# Reference Azure variable groups
variables:
  - group: 'piper-config-variables'
  - group: 'cloud-foundry-credentials'

steps:
  - task: Piper@1
    inputs:
      piperCommand: 'cloudFoundryDeploy'
    env:
      CF_USERNAME: $(cf-username)
      CF_PASSWORD: $(cf-password)
```

### Credentials Management

**Azure Pipeline Variables**:
```yaml
variables:
  # Plain variables
  BUILD_TOOL: 'maven'

  # Secret variables (from variable group)
  CF_PASSWORD: $(cf-password)

steps:
  - task: Piper@1
    env:
      PIPER_cloudFoundry_password: $(CF_PASSWORD)
```

**Inline Variable Definition**:
```yaml
variables:
  buildConfiguration: 'Release'

steps:
  - task: Piper@1
    inputs:
      piperCommand: 'mavenBuild'
    env:
      PIPER_defines: '-Dconfiguration=$(buildConfiguration)'
```

### Service Connections

**Docker Registry Connection**:
```yaml
resources:
  containers:
    - container: maven
      image: maven:3.8-openjdk-17
      endpoint: companyDockerRegistry

steps:
  - task: Piper@1
    container: maven
```

**Cloud Foundry Connection**:
```yaml
# Service connection configured in Azure DevOps
# Used via environment variables

steps:
  - task: Piper@1
    inputs:
      piperCommand: 'cloudFoundryDeploy'
    env:
      CF_USERNAME: $(CF_SERVICE_CONNECTION_USER)
      CF_PASSWORD: $(CF_SERVICE_CONNECTION_PASSWORD)
```

### Environment Variables

```yaml
variables:
  PIPER_verbose: 'true'
  PIPER_buildTool: 'maven'

steps:
  - task: Piper@1
    inputs:
      piperCommand: 'version'
```

## GitHub Actions Configuration

### Basic Setup

**Workflow File** (`.github/workflows/build.yml`):
```yaml
name: Piper Build

on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v3

      - name: Execute Piper
        uses: SAP/project-piper-action@main
        with:
          piper-version: 'latest'
          command: 'version'
```

### Action Configuration

**Maven Build Example**:
```yaml
jobs:
  maven-build:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Maven Build
        uses: SAP/project-piper-action@main
        with:
          piper-version: 'v1.220.0'
          command: 'mavenBuild'
          flags: '--verbose'
```

**Multi-Job Workflow**:
```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build
        uses: SAP/project-piper-action@main
        with:
          command: 'mavenBuild'

  deploy:
    runs-on: ubuntu-latest
    needs: build
    if: github.ref == 'refs/heads/main'

    steps:
      - uses: actions/checkout@v3

      - name: Deploy
        uses: SAP/project-piper-action@main
        with:
          command: 'cloudFoundryDeploy'
        env:
          CF_USERNAME: ${{ secrets.CF_USERNAME }}
          CF_PASSWORD: ${{ secrets.CF_PASSWORD }}
```

### GitHub-Specific Settings

**Configuration** (`.pipeline/config.yml`):
```yaml
general:
  buildTool: 'maven'
  githubApiUrl: 'https://api.github.com'
  githubServerUrl: 'https://github.com'

steps:
  githubPublishRelease:
    addClosedIssues: true
    addDeltaToLastRelease: true
    owner: 'SAP'
    repository: 'jenkins-library'
```

**GitHub Token Usage**:
```yaml
steps:
  - name: Create Release
    uses: SAP/project-piper-action@main
    with:
      command: 'githubPublishRelease'
    env:
      GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Secrets Management

**GitHub Secrets**:
```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest

    steps:
      - uses: SAP/project-piper-action@main
        with:
          command: 'cloudFoundryDeploy'
        env:
          # Access GitHub secrets
          PIPER_cloudFoundry_username: ${{ secrets.CF_USERNAME }}
          PIPER_cloudFoundry_password: ${{ secrets.CF_PASSWORD }}
          PIPER_cloudFoundry_org: ${{ secrets.CF_ORG }}
```

**Environment Secrets**:
```yaml
jobs:
  deploy-production:
    runs-on: ubuntu-latest
    environment: production  # Use environment-specific secrets

    steps:
      - uses: SAP/project-piper-action@main
        with:
          command: 'cloudFoundryDeploy'
        env:
          CF_PASSWORD: ${{ secrets.CF_PROD_PASSWORD }}
```

### GitHub Services

**GitHub API Integration**:
```yaml
general:
  githubApiUrl: 'https://api.github.com'

steps:
  githubPublishRelease:
    token: '$(githubToken)'  # From secrets
    owner: 'company'
    repository: 'my-repo'
```

**GitHub Container Registry**:
```yaml
steps:
  - name: Login to GHCR
    uses: docker/login-action@v2
    with:
      registry: ghcr.io
      username: ${{ github.actor }}
      password: ${{ secrets.GITHUB_TOKEN }}

  - name: Build and Push
    uses: SAP/project-piper-action@main
    with:
      command: 'dockerBuild'
```

### Environment Variables

```yaml
env:
  PIPER_verbose: 'true'
  PIPER_buildTool: 'maven'

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
      - uses: SAP/project-piper-action@main
        with:
          command: 'mavenBuild'
```

## Credential Management Comparison

### Jenkins Credentials

```yaml
# Configuration references Jenkins credential IDs
steps:
  cloudFoundryDeploy:
    cloudFoundry:
      credentialsId: 'CF_CREDENTIALS'  # Jenkins credential ID

  sonarExecuteScan:
    sonarTokenCredentialsId: 'SONAR_TOKEN'
```

### Azure Key Vault

```yaml
# Azure Pipeline with Key Vault
variables:
  - group: 'key-vault-secrets'  # Variable group linked to Key Vault

steps:
  - task: Piper@1
    env:
      PIPER_cloudFoundry_password: $(cf-password)  # From Key Vault
```

### GitHub Secrets

```yaml
# GitHub Actions with Secrets
env:
  PIPER_cloudFoundry_password: ${{ secrets.CF_PASSWORD }}
```

## Vault Integration Differences

### Jenkins Vault Integration

```yaml
# .pipeline/config.yml
general:
  vaultServerUrl: 'https://vault.company.com'
  vaultNamespace: 'team/project'
  vaultPath: 'piper/my-project'

# Provide via Jenkins credentials
# vaultAppRoleTokenCredentialsId: 'vault-app-role-id'
# vaultAppRoleSecretTokenCredentialsId: 'vault-app-role-secret'
```

**Custom defaults** (not in project config):
```yaml
# In custom defaults file
hooks:
  ans:
    serviceKeyCredentialsId: 'ans-service-key'
```

### Azure DevOps Vault Integration

```yaml
# Provide via environment variables
variables:
  PIPER_vaultAppRoleID: $(vault-app-role-id)
  PIPER_vaultAppRoleSecretID: $(vault-app-role-secret)

steps:
  - task: Piper@1
    env:
      PIPER_vaultServerUrl: 'https://vault.company.com'
      PIPER_vaultPath: 'piper/my-project'
```

### GitHub Actions Vault Integration

```yaml
steps:
  - uses: SAP/project-piper-action@main
    env:
      PIPER_vaultServerUrl: 'https://vault.company.com'
      PIPER_vaultAppRoleID: ${{ secrets.VAULT_APP_ROLE_ID }}
      PIPER_vaultAppRoleSecretID: ${{ secrets.VAULT_APP_ROLE_SECRET }}
      PIPER_vaultPath: 'piper/my-project'
```

## Docker Execution Differences

### Jenkins

**Native Docker Plugin**:
```groovy
dockerExecute(
    script: this,
    dockerImage: 'maven:3.8-openjdk-17'
) {
    sh 'mvn clean verify'
}
```

**Kubernetes**:
```groovy
dockerExecuteOnKubernetes(
    script: this,
    dockerImage: 'maven:3.8-openjdk-17'
) {
    sh 'mvn clean verify'
}
```

### Azure DevOps

**Container Job**:
```yaml
jobs:
  - job: Build
    container: maven:3.8-openjdk-17
    steps:
      - task: Piper@1
        inputs:
          piperCommand: 'mavenBuild'
```

### GitHub Actions

**Container Job**:
```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    container: maven:3.8-openjdk-17

    steps:
      - uses: SAP/project-piper-action@main
        with:
          command: 'mavenBuild'
```

## Common Configuration

**Shared `.pipeline/config.yml`** works across all platforms:

```yaml
general:
  buildTool: 'maven'
  productiveBranch: 'main'
  verbose: true

steps:
  mavenExecute:
    goals: 'clean verify'
    dockerImage: 'maven:3.8-openjdk-17'

  cloudFoundryDeploy:
    deployType: 'blue-green'
    cloudFoundry:
      org: 'my-org'
      apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'
```

**Platform-specific credentials** provided via platform mechanisms:
- Jenkins: Credential IDs in config
- Azure: Variable groups in pipeline
- GitHub: Secrets in workflow

## Platform Migration Guide

### Jenkins to Azure DevOps

1. **Convert Jenkinsfile to azure-pipelines.yml**:

**Before** (Jenkinsfile):
```groovy
@Library('piper-lib-os') _
piperPipeline script: this
```

**After** (azure-pipelines.yml):
```yaml
steps:
  - task: Piper@1
    inputs:
      piperCommand: 'mavenBuild'
```

2. **Migrate Credentials**:
   - Jenkins credentials → Azure variable groups
   - Add variables to Azure Pipeline settings

3. **Keep `.pipeline/config.yml` unchanged**

### Jenkins to GitHub Actions

1. **Convert Jenkinsfile to workflow**:

**Before** (Jenkinsfile):
```groovy
@Library('piper-lib-os') _
piperPipeline script: this
```

**After** (.github/workflows/build.yml):
```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: SAP/project-piper-action@main
        with:
          command: 'mavenBuild'
```

2. **Migrate Secrets**:
   - Jenkins credentials → GitHub secrets
   - Add secrets in repository settings

3. **Keep `.pipeline/config.yml` unchanged**

## Best Practices by Platform

### Jenkins Best Practices

1. Use Shared Libraries for reusability
2. Store credentials in Jenkins credential store
3. Use Kubernetes for dynamic agents
4. Implement stage extensions for customization
5. Use declarative pipelines when possible

### Azure DevOps Best Practices

1. Use variable groups for secrets
2. Link variable groups to Azure Key Vault
3. Use container jobs for consistency
4. Implement gates for approvals
5. Use service connections for external services

### GitHub Actions Best Practices

1. Use environment secrets for production
2. Implement protection rules
3. Use reusable workflows
4. Cache dependencies
5. Use GitHub Container Registry

---

**Next**: [Stage Configuration](04-stage-configuration.md) - Configuring pipeline stages

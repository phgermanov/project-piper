# Project Piper Action - Usage Guide

## Table of Contents

1. [Installation and Setup](#installation-and-setup)
2. [Basic Usage](#basic-usage)
3. [Action Inputs Reference](#action-inputs-reference)
4. [Action Outputs Reference](#action-outputs-reference)
5. [Common Patterns](#common-patterns)
6. [Step-by-Step Examples](#step-by-step-examples)
7. [CPE Management](#cpe-management)
8. [Binary Caching Strategies](#binary-caching-strategies)
9. [Docker Usage Patterns](#docker-usage-patterns)
10. [Troubleshooting](#troubleshooting)
11. [Best Practices](#best-practices)

---

## Installation and Setup

### Prerequisites

1. A GitHub repository with Actions enabled
2. Basic understanding of GitHub Actions workflows
3. (Optional) Docker for containerized steps

### Minimal Workflow

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Run Piper
        uses: SAP/project-piper-action@v1
        with:
          step-name: version
```

This minimal workflow:
- Checks out your code
- Downloads the Piper binary
- Executes the `version` step to verify setup

---

## Basic Usage

### Simple Step Execution

```yaml
- uses: SAP/project-piper-action@v1
  with:
    step-name: mavenBuild
```

### Step with Flags

```yaml
- uses: SAP/project-piper-action@v1
  with:
    step-name: mavenBuild
    flags: '--publish --createBOM --logSuccessfulMavenTransfers'
```

### Step with Configuration

```yaml
- uses: SAP/project-piper-action@v1
  with:
    step-name: npmExecuteScripts
    flags: 'ci build test'
```

### Multiple Steps in One Job

```yaml
- name: Build
  uses: SAP/project-piper-action@v1
  with:
    step-name: mavenBuild
    flags: '--publish'

- name: Test
  uses: SAP/project-piper-action@v1
  with:
    step-name: mavenExecuteTest

- name: Security Scan
  uses: SAP/project-piper-action@v1
  with:
    step-name: detectExecuteScan
```

---

## Action Inputs Reference

### Core Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `step-name` | Name of Piper step to execute | No | - |
| `flags` | Flags/arguments for the step | No | - |

### Binary Management

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `piper-version` | Piper binary version (latest, master, or version tag) | No | latest |
| `piper-owner` | Owner of Piper repository | No | SAP |
| `piper-repository` | Piper repository name | No | jenkins-library |
| `github-token` | Token to access GitHub API | No | - |

### SAP Internal Binary

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `sap-piper-version` | SAP Piper binary version | No | - |
| `sap-piper-owner` | Owner of SAP Piper repository | No | - |
| `sap-piper-repository` | SAP Piper repository name | No | - |
| `github-enterprise-token` | Token for GitHub Enterprise | No | - |

### Docker Configuration

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `docker-image` | Docker image to run Piper in | No | - |
| `docker-options` | Docker options for container | No | - |
| `docker-env-vars` | Environment variables for Docker (JSON) | No | - |

### Sidecar Configuration

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `sidecar-image` | Docker image for sidecar container | No | - |
| `sidecar-options` | Docker options for sidecar | No | - |
| `sidecar-env-vars` | Environment variables for sidecar | No | - |

### Advanced Configuration

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `custom-defaults-paths` | Comma-separated custom defaults paths | No | - |
| `custom-stage-conditions-path` | Path to custom stage conditions YAML | No | - |
| `export-pipeline-environment` | Export CPE as output variable | No | false |
| `create-check-if-step-active-maps` | Create step/stage activation maps | No | false |

---

## Action Outputs Reference

| Output | Description | Available When |
|--------|-------------|----------------|
| `pipelineEnv` | Exported Common Pipeline Environment (JSON) | `export-pipeline-environment: true` |

### Using Outputs

```yaml
- name: Build
  id: build-step
  uses: SAP/project-piper-action@v1
  with:
    step-name: artifactPrepareVersion
    export-pipeline-environment: true

- name: Print Version
  run: |
    echo "Pipeline Environment: ${{ steps.build-step.outputs.pipelineEnv }}"
```

---

## Common Patterns

### Pattern 1: Build and Test

```yaml
name: Build and Test

on: [push, pull_request]

jobs:
  maven-build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build
        uses: SAP/project-piper-action@v1
        with:
          step-name: mavenBuild
          flags: '--publish'

      - name: Test
        uses: SAP/project-piper-action@v1
        with:
          step-name: mavenExecuteTest
```

### Pattern 2: Multi-Stage Pipeline

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: SAP/project-piper-action@v1
        id: build
        with:
          step-name: mavenBuild
          export-pipeline-environment: true

      - uses: actions/upload-artifact@v4
        with:
          name: build-artifacts
          path: |
            target/*.jar
            .pipeline/

  test:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/download-artifact@v4
        with:
          name: build-artifacts

      - uses: SAP/project-piper-action@v1
        with:
          step-name: mavenExecuteTest

  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4

      - uses: actions/download-artifact@v4
        with:
          name: build-artifacts

      - uses: SAP/project-piper-action@v1
        with:
          step-name: cloudFoundryDeploy
        env:
          PIPER_cfUsername: ${{ secrets.CF_USERNAME }}
          PIPER_cfPassword: ${{ secrets.CF_PASSWORD }}
```

### Pattern 3: Matrix Strategy

```yaml
name: Multi-Version Build

on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        node-version: [16, 18, 20]
    steps:
      - uses: actions/checkout@v4

      - uses: SAP/project-piper-action@v1
        with:
          step-name: npmExecuteScripts
          docker-image: node:${{ matrix.node-version }}
          flags: 'ci test'
```

### Pattern 4: Conditional Execution

```yaml
- name: Security Scan
  if: github.event_name == 'push' && github.ref == 'refs/heads/main'
  uses: SAP/project-piper-action@v1
  with:
    step-name: detectExecuteScan

- name: Deploy to Dev
  if: github.ref == 'refs/heads/develop'
  uses: SAP/project-piper-action@v1
  with:
    step-name: cloudFoundryDeploy
  env:
    PIPER_cloudFoundry_org: dev-org
    PIPER_cloudFoundry_space: dev-space
```

---

## Step-by-Step Examples

### Example 1: Maven Build

**Workflow:**

```yaml
name: Maven Build

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Maven Build
        uses: SAP/project-piper-action@v1
        with:
          step-name: mavenBuild
          flags: '--publish --createBOM'
```

**Configuration (`.pipeline/config.yml`):**

```yaml
steps:
  mavenBuild:
    pomPath: pom.xml
    goals: clean install
    defines: -DskipTests=false
    publish: true
    createBOM: true
```

### Example 2: NPM with Docker

**Workflow:**

```yaml
name: NPM Build

on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: NPM Build
        uses: SAP/project-piper-action@v1
        with:
          step-name: npmExecuteScripts
          docker-image: node:18-alpine
          flags: 'ci build test'
          docker-env-vars: '{"NPM_TOKEN": "${{ secrets.NPM_TOKEN }}"}'
```

**Configuration:**

```yaml
steps:
  npmExecuteScripts:
    install: true
    runScripts:
      - ci
      - build
      - test
```

### Example 3: Cloud Foundry Deployment

**Workflow:**

```yaml
name: Deploy to Cloud Foundry

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4

      - name: Build
        uses: SAP/project-piper-action@v1
        with:
          step-name: mtaBuild

      - name: Deploy
        uses: SAP/project-piper-action@v1
        with:
          step-name: cloudFoundryDeploy
        env:
          PIPER_cfApiEndpoint: https://api.cf.us10.hana.ondemand.com
          PIPER_cfOrg: my-org
          PIPER_cfSpace: production
          PIPER_cfUsername: ${{ secrets.CF_USERNAME }}
          PIPER_cfPassword: ${{ secrets.CF_PASSWORD }}
```

**Configuration:**

```yaml
steps:
  cloudFoundryDeploy:
    mtaDeployParameters: --version-rule ALL
    deployType: standard
```

### Example 4: Security Scanning

**Workflow:**

```yaml
name: Security Scan

on:
  schedule:
    - cron: '0 2 * * 1'  # Weekly on Monday at 2 AM
  push:
    branches: [main]

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Detect Scan
        uses: SAP/project-piper-action@v1
        with:
          step-name: detectExecuteScan
        env:
          PIPER_detectTokenCredentialsId: ${{ secrets.BLACKDUCK_TOKEN }}

      - name: SonarQube Scan
        uses: SAP/project-piper-action@v1
        with:
          step-name: sonarExecuteScan
        env:
          PIPER_sonarToken: ${{ secrets.SONAR_TOKEN }}
```

**Configuration:**

```yaml
steps:
  detectExecuteScan:
    scanners:
      - signature
    scanPaths:
      - '.'
    failOn:
      - BLOCKER

  sonarExecuteScan:
    serverUrl: https://sonarcloud.io
    organization: my-org
    projectKey: my-project
```

### Example 5: Multi-Technology Build

**Workflow:**

```yaml
name: Multi-Tech Build

on: [push]

jobs:
  backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build Java Backend
        uses: SAP/project-piper-action@v1
        with:
          step-name: mavenBuild
          flags: '--pomPath backend/pom.xml'

  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build UI5 Frontend
        uses: SAP/project-piper-action@v1
        with:
          step-name: npmExecuteScripts
          docker-image: node:18
          flags: 'ci build'
          docker-options: '--workdir=/github/workspace/frontend'

  integration:
    needs: [backend, frontend]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build MTA
        uses: SAP/project-piper-action@v1
        with:
          step-name: mtaBuild
```

### Example 6: Integration Tests with Database

**Workflow:**

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  integration-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run Integration Tests
        uses: SAP/project-piper-action@v1
        with:
          step-name: npmExecuteScripts
          flags: 'test:integration'
          docker-image: node:18
          sidecar-image: postgres:15
          sidecar-env-vars: '{"POSTGRES_DB": "testdb", "POSTGRES_USER": "test", "POSTGRES_PASSWORD": "test123"}'
          docker-env-vars: '{"DATABASE_URL": "postgresql://test:test123@postgres:5432/testdb"}'
```

**Configuration:**

```yaml
steps:
  npmExecuteScripts:
    dockerName: app
    sidecarName: postgres
    dockerEnvVars:
      DATABASE_URL: postgresql://test:test123@postgres:5432/testdb
```

---

## CPE Management

### Understanding CPE Workflow

Common Pipeline Environment (CPE) allows sharing data between steps and jobs.

### Pattern: Share Version Between Jobs

**Job 1: Prepare Version**

```yaml
prepare-version:
  runs-on: ubuntu-latest
  outputs:
    pipeline-env: ${{ steps.version.outputs.pipelineEnv }}
  steps:
    - uses: actions/checkout@v4

    - name: Prepare Version
      id: version
      uses: SAP/project-piper-action@v1
      with:
        step-name: artifactPrepareVersion
        export-pipeline-environment: true

    - name: Upload CPE
      uses: actions/upload-artifact@v4
      with:
        name: pipeline-env
        path: .pipeline/
```

**Job 2: Use Version**

```yaml
build:
  needs: prepare-version
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4

    - name: Download CPE
      uses: actions/download-artifact@v4
      with:
        name: pipeline-env
        path: .pipeline/

    - name: Build with Version
      uses: SAP/project-piper-action@v1
      with:
        step-name: mavenBuild
      env:
        PIPER_ACTION_PIPELINE_ENV: ${{ needs.prepare-version.outputs.pipeline-env }}
```

### Pattern: Share Build Information

```yaml
build:
  runs-on: ubuntu-latest
  outputs:
    build-info: ${{ steps.build.outputs.pipelineEnv }}
  steps:
    - uses: actions/checkout@v4

    - name: Build
      id: build
      uses: SAP/project-piper-action@v1
      with:
        step-name: mtaBuild
        export-pipeline-environment: true

    - name: Upload CPE and Artifacts
      uses: actions/upload-artifact@v4
      with:
        name: build-output
        path: |
          .pipeline/
          *.mtar

deploy:
  needs: build
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4

    - name: Download Build Output
      uses: actions/download-artifact@v4
      with:
        name: build-output

    - name: Deploy
      uses: SAP/project-piper-action@v1
      with:
        step-name: cloudFoundryDeploy
      env:
        PIPER_ACTION_PIPELINE_ENV: ${{ needs.build.outputs.build-info }}
```

### Manual CPE Management

If you need to manually manipulate CPE:

```yaml
- name: Create CPE Directory
  run: mkdir -p .pipeline/commonPipelineEnvironment/custom

- name: Write Custom Data
  run: |
    echo '{"myKey": "myValue"}' > .pipeline/commonPipelineEnvironment/custom/mydata.json

- name: Use in Piper Step
  uses: SAP/project-piper-action@v1
  with:
    step-name: myCustomStep
```

---

## Binary Caching Strategies

### Default Caching (Automatic)

The action automatically caches binaries within the same job:

```yaml
- uses: SAP/project-piper-action@v1
  with:
    step-name: version
    # Binary downloaded to ./v1_300_0/piper

- uses: SAP/project-piper-action@v1
  with:
    step-name: mavenBuild
    # Binary reused from ./v1_300_0/piper
```

### Cross-Job Caching

Use GitHub Actions cache:

```yaml
jobs:
  setup:
    runs-on: ubuntu-latest
    steps:
      - name: Cache Piper Binary
        uses: actions/cache@v3
        with:
          path: ./piper-bin
          key: piper-${{ runner.os }}-latest

      - name: Download Binary
        uses: SAP/project-piper-action@v1
        with:
          step-name: version
          piper-version: latest

  build:
    needs: setup
    runs-on: ubuntu-latest
    steps:
      - name: Restore Piper Binary
        uses: actions/cache@v3
        with:
          path: ./piper-bin
          key: piper-${{ runner.os }}-latest

      - uses: SAP/project-piper-action@v1
        with:
          step-name: mavenBuild
```

### Version-Specific Caching

```yaml
- name: Cache Piper Binary
  uses: actions/cache@v3
  with:
    path: |
      ./v1_300_0
      ./latest
    key: piper-${{ runner.os }}-${{ hashFiles('.github/workflows/*.yml') }}
    restore-keys: |
      piper-${{ runner.os }}-
```

### Optimal Strategy

For most use cases, rely on automatic in-job caching:

```yaml
jobs:
  ci:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      # First use downloads binary
      - uses: SAP/project-piper-action@v1
        with:
          step-name: artifactPrepareVersion

      # Subsequent uses reuse binary
      - uses: SAP/project-piper-action@v1
        with:
          step-name: mavenBuild

      - uses: SAP/project-piper-action@v1
        with:
          step-name: mavenExecuteTest
```

---

## Docker Usage Patterns

### Pattern 1: Consistent Build Environment

```yaml
- uses: SAP/project-piper-action@v1
  with:
    step-name: mavenBuild
    docker-image: maven:3.9-eclipse-temurin-17
    docker-options: '--memory=4g --cpus=2'
```

### Pattern 2: Node.js with Custom Registry

```yaml
- uses: SAP/project-piper-action@v1
  with:
    step-name: npmExecuteScripts
    docker-image: node:18-alpine
    docker-env-vars: '{"NPM_CONFIG_REGISTRY": "https://registry.company.com"}'
    flags: 'ci build'
```

### Pattern 3: Private Docker Registry

```yaml
- name: Login to Docker Registry
  run: docker login registry.company.com -u ${{ secrets.DOCKER_USER }} -p ${{ secrets.DOCKER_PASSWORD }}

- uses: SAP/project-piper-action@v1
  with:
    step-name: mavenBuild
    docker-image: registry.company.com/maven-custom:latest
```

### Pattern 4: Multi-Container Setup

```yaml
- uses: SAP/project-piper-action@v1
  with:
    step-name: npmExecuteScripts
    docker-image: node:18
    sidecar-image: redis:7-alpine
    sidecar-env-vars: '{"REDIS_PASSWORD": "test123"}'
    docker-env-vars: '{"REDIS_HOST": "redis", "REDIS_PORT": "6379"}'
    flags: 'test:e2e'
```

### Pattern 5: Docker with Mounted Volumes

```yaml
- name: Create Config Directory
  run: mkdir -p ${{ github.workspace }}/config

- uses: SAP/project-piper-action@v1
  with:
    step-name: mavenBuild
    docker-image: maven:3.9-jdk-17
    docker-options: '--volume ${{ github.workspace }}/config:/config:ro'
```

### Pattern 6: Docker Build with Kaniko

```yaml
- uses: SAP/project-piper-action@v1
  with:
    step-name: kanikoExecute
    flags: |
      --containerImageName myapp
      --containerImageTag ${{ github.sha }}
      --containerRegistryUrl registry.company.com
  env:
    PIPER_containerRegistryUser: ${{ secrets.REGISTRY_USER }}
    PIPER_containerRegistryPassword: ${{ secrets.REGISTRY_PASSWORD }}
```

---

## Troubleshooting

### Common Issues and Solutions

#### Issue 1: Binary Not Found

**Error:**
```
Error: Piper binary path is empty
```

**Solution:**
```yaml
# Ensure piper-version is specified
- uses: SAP/project-piper-action@v1
  with:
    step-name: mavenBuild
    piper-version: latest
```

#### Issue 2: Docker Permission Denied

**Error:**
```
docker: permission denied while trying to connect to the Docker daemon socket
```

**Solution:**
```yaml
# Ensure Docker is available on runner
# GitHub-hosted runners have Docker pre-installed
# For self-hosted runners, install Docker or use non-containerized steps
```

#### Issue 3: Step Configuration Not Found

**Error:**
```
No configuration found for step: myStep
```

**Solution:**
1. Check step name spelling
2. Verify `.pipeline/config.yml` exists and is valid
3. Ensure step configuration is under correct key

```yaml
# Correct structure
steps:
  mavenBuild:  # Step name must match exactly
    pomPath: pom.xml
```

#### Issue 4: Authentication Failed

**Error:**
```
401 Unauthorized
```

**Solution:**
```yaml
# Provide required tokens
- uses: SAP/project-piper-action@v1
  with:
    step-name: detectExecuteScan
    github-token: ${{ secrets.GITHUB_TOKEN }}
  env:
    PIPER_detectToken: ${{ secrets.BLACKDUCK_TOKEN }}
```

#### Issue 5: Container Network Issues

**Error:**
```
Cannot connect to database at localhost
```

**Solution:**
```yaml
# Use service name instead of localhost
docker-env-vars: '{"DB_HOST": "postgres"}'  # Not "localhost"
sidecar-image: postgres:15
```

### Debug Mode

Enable verbose logging:

```yaml
- uses: SAP/project-piper-action@v1
  with:
    step-name: mavenBuild
    flags: '--verbose'
  env:
    ACTIONS_STEP_DEBUG: true  # GitHub Actions debug mode
```

### Step-Specific Help

Get help for a specific step:

```yaml
- uses: SAP/project-piper-action@v1
  with:
    step-name: mavenBuild
    flags: '--help'
```

---

## Best Practices

### 1. Version Pinning

**Good:**
```yaml
# Pin action to major version
- uses: SAP/project-piper-action@v1
  with:
    piper-version: v1.300.0  # Specific Piper version
```

**Avoid:**
```yaml
# Unpredictable behavior
- uses: SAP/project-piper-action@main
  with:
    piper-version: master
```

### 2. Configuration Management

**Good:**
```yaml
# Use configuration file
# .pipeline/config.yml
steps:
  mavenBuild:
    publish: true
    createBOM: true
    pomPath: pom.xml
```

**Avoid:**
```yaml
# Long flags string is hard to maintain
flags: '--publish --createBOM --pomPath pom.xml --goals "clean install" --defines "-DskipTests=false"'
```

### 3. Secrets Handling

**Good:**
```yaml
env:
  PIPER_cfPassword: ${{ secrets.CF_PASSWORD }}
  PIPER_vaultAppRoleSecretID: ${{ secrets.VAULT_SECRET }}
```

**Avoid:**
```yaml
# Never hardcode secrets
flags: '--password mySecretPassword123'
```

### 4. Resource Management

**Good:**
```yaml
# Limit container resources
docker-options: '--memory=4g --cpus=2 --memory-swap=4g'
```

**Good:**
```yaml
# Clean up after jobs
- uses: SAP/project-piper-action@v1
  with:
    step-name: mavenBuild

- name: Cleanup
  if: always()
  run: docker system prune -f
```

### 5. Error Handling

**Good:**
```yaml
- name: Build
  id: build
  uses: SAP/project-piper-action@v1
  with:
    step-name: mavenBuild
  continue-on-error: false

- name: Notify on Failure
  if: failure()
  run: echo "Build failed, sending notification..."
```

### 6. Caching Strategy

**Good:**
```yaml
# Cache Maven dependencies
- name: Cache Maven Repo
  uses: actions/cache@v3
  with:
    path: ~/.m2/repository
    key: maven-${{ hashFiles('**/pom.xml') }}

- uses: SAP/project-piper-action@v1
  with:
    step-name: mavenBuild
```

### 7. Workflow Organization

**Good:**
```yaml
# Separate concerns into jobs
jobs:
  build:
    # Build artifacts

  test:
    needs: build
    # Run tests

  security:
    needs: build
    # Security scans

  deploy:
    needs: [test, security]
    # Deploy to environment
```

### 8. Documentation

**Good:**
```yaml
- name: Build Application
  uses: SAP/project-piper-action@v1
  with:
    step-name: mavenBuild
    flags: '--publish --createBOM'
  # This step builds the Maven project, publishes artifacts to Nexus,
  # and creates a Bill of Materials (BOM) for supply chain security
```

### 9. Environment-Specific Configuration

**Good:**
```yaml
deploy-dev:
  if: github.ref == 'refs/heads/develop'
  environment: development
  steps:
    - uses: SAP/project-piper-action@v1
      with:
        step-name: cloudFoundryDeploy
      env:
        PIPER_cfOrg: dev-org
        PIPER_cfSpace: dev-space

deploy-prod:
  if: github.ref == 'refs/heads/main'
  environment: production
  steps:
    - uses: SAP/project-piper-action@v1
      with:
        step-name: cloudFoundryDeploy
      env:
        PIPER_cfOrg: prod-org
        PIPER_cfSpace: prod-space
```

### 10. Testing Changes

**Good:**
```yaml
# Test in feature branch first
on:
  push:
    branches:
      - main
      - develop
      - 'feature/**'  # Test on feature branches
  pull_request:
    # Validate on PRs before merging
```

---

## Quick Reference

### Most Common Steps

| Step | Use Case | Example |
|------|----------|---------|
| `mavenBuild` | Build Java/Maven projects | `step-name: mavenBuild` |
| `npmExecuteScripts` | Build Node.js projects | `step-name: npmExecuteScripts` |
| `mtaBuild` | Build MTA applications | `step-name: mtaBuild` |
| `cloudFoundryDeploy` | Deploy to CF | `step-name: cloudFoundryDeploy` |
| `detectExecuteScan` | Security scanning | `step-name: detectExecuteScan` |
| `sonarExecuteScan` | Code quality analysis | `step-name: sonarExecuteScan` |

### Essential Flags

| Flag | Purpose | Example |
|------|---------|---------|
| `--verbose` | Detailed logging | `flags: '--verbose'` |
| `--noTelemetry` | Disable telemetry | `flags: '--noTelemetry'` |
| `--help` | Show step help | `flags: '--help'` |

### Configuration Files

| File | Purpose |
|------|---------|
| `.pipeline/config.yml` | Main Piper configuration |
| `.pipeline/defaults.yml` | Custom defaults |
| `.pipeline/commonPipelineEnvironment/` | CPE storage |

---

## Additional Resources

- **Piper Documentation**: [https://www.project-piper.io/](https://www.project-piper.io/)
- **Step Reference**: [https://www.project-piper.io/steps/](https://www.project-piper.io/steps/)
- **Configuration**: [https://www.project-piper.io/configuration/](https://www.project-piper.io/configuration/)
- **GitHub Actions Docs**: [https://docs.github.com/actions](https://docs.github.com/actions)
- **Action Repository**: [https://github.com/SAP/project-piper-action](https://github.com/SAP/project-piper-action)

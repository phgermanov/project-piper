# GitHub Piper Pipeline (GPP) - Overview

## Table of Contents

1. [What is GPP?](#what-is-gpp)
2. [Project Structure](#project-structure)
3. [Available Stages](#available-stages)
4. [How to Use GPP](#how-to-use-gpp)
5. [Key Features](#key-features)
6. [Quick Start Guide](#quick-start-guide)
7. [Additional Resources](#additional-resources)

## What is GPP?

**GitHub Piper Pipeline (GPP)** is Piper's general purpose pipeline for GitHub Actions, providing a standardized, enterprise-ready CI/CD solution for SAP development projects. Based on [SAP's Project Piper Action](https://github.com/SAP/project-piper-action), it implements the same stages as Jenkins and Azure DevOps Piper pipelines.

**Key Benefits**:
- Pre-configured CI/CD stages for build, test, security, and deployment
- Integration with SAP services (System Trust, Cumulus, Vault)
- Extensibility through composite actions
- Automated compliance and policy enforcement
- Multi-technology support (Maven, npm, Gradle, Golang, Python, MTA)

## Project Structure

```
piper-pipeline-github/
├── .github/workflows/
│   ├── sap-piper-workflow.yml     # Main GPP workflow
│   ├── sap-oss-ppms-workflow.yml  # OSS compliance workflow
│   ├── init.yml                   # Init stage
│   ├── build.yml                  # Build stage
│   ├── integration.yml            # Integration testing
│   ├── acceptance.yml             # Acceptance testing
│   ├── performance.yml            # Performance testing
│   ├── promote.yml                # Promotion with approval
│   ├── release.yml                # Release/deployment
│   ├── post.yml                   # Post-pipeline actions
│   ├── oss.yml                    # Security/OSS scanning
│   └── ppms.yml                   # IPScan and PPMS
├── docs/
│   ├── extensibility.md           # Extensibility guide
│   └── adr/                       # Architecture decisions
└── examples/custom_workflow/      # Usage examples
```

## Available Stages

### 1. Init Stage
**Purpose**: Pipeline initialization and configuration

**Key Actions**:
- Determine productive branch
- Initialize Common Pipeline Environment (CPE)
- Configure vault integration and System Trust
- Generate active stages/steps maps
- Send pipeline start telemetry

**Outputs**: `on-productive-branch`, `active-stages-map`, `active-steps-map`, `pipeline-env`

### 2. Build Stage
**Purpose**: Build artifacts, run tests, perform code analysis

**Build Steps** (auto-detected based on project):
- `mtaBuild` - Multi-Target Applications
- `mavenBuild` - Java/Maven projects
- `npmExecuteScripts` - Node.js projects
- `gradleExecuteBuild` - Gradle projects
- `pythonBuild` - Python projects
- `golangBuild` - Go projects

**Container Operations**:
- `kanikoExecute` - Docker image builds
- `cnbBuild` - Cloud Native Buildpacks
- `hadolintExecute` - Dockerfile linting
- `helmExecute` - Helm chart operations
- `imagePushToRegistry` - Push to registry

**Quality & Testing**:
- `karmaExecuteTests` - JavaScript tests
- `sonarExecuteScan` - Code quality analysis

**Evidence Collection**: Test results, code coverage, SBOM, policy evidence

**Extensibility**: `preBuild`, `postBuild`

### 3-5. Testing Stages (Parallel Execution)

#### Integration Stage
Execute integration tests in deployed environment

#### Acceptance Stage
Run acceptance tests and UAT validation

#### Performance Stage
Execute performance and load testing

**All run in parallel after Build completes**

**Extensibility**: `pre<Stage>`, `post<Stage>` for each

### 6. Promote Stage
**Purpose**: Manual approval gate for production

**Key Features**:
- Uses GitHub Environment (`Piper Promote`)
- Requires manual approval via environment protection rules
- Only runs on productive branches
- Generates lock-run.json for tracking

**Note**: Piper's `manualConfirmation` parameter is NOT used (differs from Jenkins/Azure)

### 7. Release Stage
**Purpose**: Deploy to production

Executes after Promote approval. Handles production deployment and validation.

**Extensibility**: `preRelease`, `postRelease`

### 8. Post Stage
**Purpose**: Finalization and cleanup

- Generate release status
- Upload final evidence to Cumulus
- Send notifications
- Cleanup resources

**Extensibility**: `prePost`, `postPost`

### 9-10. Security/Compliance Stages

#### Security/OSS Stage
Open source security and license scanning:
- `whitesourceExecuteScan` - Mend/WhiteSource
- `detectExecuteScan` - BlackDuck
- `protecodeExecuteScan` - Container scanning

#### IPScan and PPMS Stage
Intellectual property compliance and PPMS integration

**Extensibility**: `preOSS`, `postOSS`, `prePPMS`, `postPPMS`

### Stage Flow

```
Init → Build → Integration/Acceptance/Performance → Promote → Release → Post
                    (parallel execution)          (manual)
```

## How to Use GPP

### Prerequisites

- GitHub Actions enabled
- SUGAR Service registration for runners
- Required secrets:
  - `PIPER_VAULTAPPROLEID`
  - `PIPER_VAULTAPPROLESECRETID`

### Basic Usage

Create `.github/workflows/piper.yml`:

```yaml
name: Piper workflow

on:
  push:
    branches: [main]
  workflow_dispatch:

jobs:
  piper:
    uses: project-piper/piper-pipeline-github/.github/workflows/sap-piper-workflow.yml@v1
    secrets: inherit
```

**Recommendation**: Use version tags (`@v1`, `@v1.2.3`) instead of `@main` for stability and to enable automated updates via ospo-renovate.

### Custom Runner Tags

```yaml
jobs:
  piper:
    uses: project-piper/piper-pipeline-github/.github/workflows/sap-piper-workflow.yml@v1
    secrets: inherit
    with:
      runs-on: '[ "self-hosted", "my-custom-runner-tag" ]'
```

### OSS Compliance Workflow

```yaml
name: OSS Compliance

on:
  schedule:
    - cron: '0 2 * * 1'  # Weekly

jobs:
  oss-compliance:
    uses: project-piper/piper-pipeline-github/.github/workflows/sap-oss-ppms-workflow.yml@v1
    secrets: inherit
```

### Pipeline Configuration

Create `.pipeline/config.yml`:

```yaml
general:
  productiveBranch: 'main'
  globalExtensionsRepository: 'your-org/pipeline-extensions'  # Optional
  vaultBasePath: 'team/project'
  vaultPipelineName: 'my-pipeline'

stages:
  Build:
    mavenBuild: true
    kanikoExecute: true
  Acceptance:
    cloudFoundryDeploy: true
  Release:
    cloudFoundryDeploy: true

steps:
  mavenBuild:
    goals: ['clean', 'install']
  kanikoExecute:
    containerImageName: 'my-app'
    containerImageTag: '${VERSION}'
```

## Key Features

### 1. Extensibility

**Local Extensions**: Place in `.pipeline/extensions/`

```yaml
# .pipeline/extensions/preBuild/action.yml
name: 'PreBuild'
runs:
  using: "composite"
  steps:
    - name: Custom Step
      uses: SAP/project-piper-action@v1
      with:
        step-name: shellExecute
        flags: --sources .pipeline/extensions/preBuild/script.sh
      shell: bash
```

**Enable in workflow**:

```yaml
jobs:
  piper:
    uses: project-piper/piper-pipeline-github/.github/workflows/sap-piper-workflow.yml@v1
    secrets: inherit
    with:
      extensibility-enabled: true
```

**Global Extensions**: Configure in `.pipeline/config.yml`:

```yaml
general:
  globalExtensionsRepository: 'org-name/global-extensions'
  globalExtensionsRef: 'main'  # Optional
```

**Extensible Stages**: Build, Acceptance, Integration, Performance, Release, Post, OSS, PPMS

### 2. System Trust Integration

Automatic integration with SAP's System Trust service for secure session tokens.

**Configuration**:

```yaml
general:
  hooks:
    systemtrust:
      serverURL: 'https://system-trust.example.com'
```

**Features**:
- OIDC-based authentication
- Time-limited session tokens
- Automatic token retrieval in all stages
- Requires `id-token: write` permissions

### 3. Cumulus Integration

Automatic evidence upload to Cumulus for compliance tracking.

**Evidence Types**:
- Build artifacts (config, settings, env info)
- Test results (JUnit, code coverage)
- Security scans (SonarQube, WhiteSource, BlackDuck)
- Compliance evidence (SBOM, policy evidence)
- Deployment logs (release status, lock-run)

**Policy Evidence**: SLC-25, SLC-29-PI, SLC-29-PNB, FC-1

### 4. Custom Piper Versions

Test development versions:

```yaml
jobs:
  piper:
    uses: project-piper/piper-pipeline-github/.github/workflows/sap-piper-workflow.yml@v1
    secrets: inherit
    with:
      piper-version: 'devel:SAP:jenkins-library:<SHA>'
      sap-piper-version: 'devel:ContinuousDelivery:piper-library:<SHA>'
```

**Required secrets**:
- `PIPER_ENTERPRISE_SERVER_URL`: `https://github.wdf.sap.corp`
- `PIPER_WDF_GITHUB_TOKEN`: WDF access token

### 5. Git Configuration

```yaml
jobs:
  piper:
    uses: project-piper/piper-pipeline-github/.github/workflows/sap-piper-workflow.yml@v1
    secrets: inherit
    with:
      checkout-submodules: 'recursive'  # or 'true'
      checkout-lfs: 'true'
      fetch-depth: '0'  # Full history
```

### 6. Custom Defaults

```yaml
jobs:
  piper:
    uses: project-piper/piper-pipeline-github/.github/workflows/sap-piper-workflow.yml@v1
    secrets: inherit
    with:
      custom-defaults-paths: '.pipeline/custom-defaults.yml,.pipeline/team-defaults.yml'
```

## Quick Start Guide

### Step 1: Prerequisites

- [ ] GitHub Actions enabled
- [ ] SUGAR Service registration
- [ ] Vault secrets configured
- [ ] Self-hosted runners available

### Step 2: Create Configuration

`.pipeline/config.yml`:

```yaml
general:
  productiveBranch: 'main'

stages:
  Build:
    mavenBuild: true
  Release:
    cloudFoundryDeploy: true

steps:
  mavenBuild:
    goals: ['clean', 'verify']
```

### Step 3: Create Workflow

`.github/workflows/piper.yml`:

```yaml
name: Piper CI/CD

on:
  push:
    branches: [main, 'feature/**']
  pull_request:
  workflow_dispatch:

jobs:
  piper:
    uses: project-piper/piper-pipeline-github/.github/workflows/sap-piper-workflow.yml@v1
    secrets: inherit
```

### Step 4: Configure Manual Approval (Optional)

1. **Settings** → **Environments**
2. Create `Piper Promote` environment
3. Add **Required reviewers**
4. Save protection rules

### Step 5: Push and Monitor

```bash
git add .pipeline/config.yml .github/workflows/piper.yml
git commit -m "Add Piper pipeline"
git push
```

View execution in **Actions** tab.

## Additional Resources

### Documentation

- [Extensibility Guide](../piper-pipeline-github/docs/extensibility.md)
- [Architecture Decisions](../piper-pipeline-github/docs/adr/)
- [Project Piper Action](https://github.com/SAP/project-piper-action)
- [Piper Steps Reference](https://www.project-piper.io/steps/)

### Stage-Specific Documentation

- [01-init-stage.md](01-init-stage.md) - Init configuration
- [02-build-stage.md](02-build-stage.md) - Build steps and technologies
- [03-test-stages.md](03-test-stages.md) - Testing stages
- [04-promote-release-stages.md](04-promote-release-stages.md) - Promotion and release
- [05-security-compliance.md](05-security-compliance.md) - Security and compliance
- [06-extensibility.md](06-extensibility.md) - Advanced extensibility
- [07-troubleshooting.md](07-troubleshooting.md) - Common issues

### Examples

- [gha-demo-k8s-node](../gha-demo-k8s-node/) - Node.js + Kubernetes example
- [Custom Workflows](../piper-pipeline-github/examples/) - Customization patterns

### Key Differences

**vs. Jenkins**:
- Approval: GitHub Environments (not `manualConfirmation`)
- Runners: Self-hosted/SUGAR (not Jenkins agents)
- Extensibility: Composite Actions (not Shared Libraries)

**vs. Azure DevOps**:
- Similar workflow YAML structure
- GitHub Environments vs. Azure Environments
- GitHub Artifacts vs. Azure Artifacts

### Version Requirements

- GitHub Enterprise Server: 3.4+
- Actions Runner: 2.300+
- Piper Action: v1.22+
- Go (for devel builds): 1.24+

### Known Limitations

1. Promote stage uses GitHub Environments (not `manualConfirmation` parameter)
2. No interactive commands (e.g., `git rebase -i`)
3. GitHub Artifacts size limitations apply
4. Concurrency limited by available runners

---

**Last Updated**: 2025-11-14
**GPP Version**: v1.x
**Maintained By**: SAP Piper Team

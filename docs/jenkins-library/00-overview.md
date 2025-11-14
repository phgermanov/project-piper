# Jenkins Library (piper-os) - Overview

## Introduction

The **jenkins-library** (also known as **piper-os**) is the core component of Project Piper. It provides a comprehensive Jenkins Shared Library written in Groovy, backed by a powerful Go CLI tool that implements over 200 CI/CD pipeline steps.

**Repository**: `jenkins-library/`
**Technologies**: Go 1.24.0, Groovy
**Status**: No longer accepting contributions (archived), but available for use

## What is jenkins-library?

jenkins-library is a **dual-purpose library**:

1. **Jenkins Shared Library**: 188 Groovy pipeline steps that integrate natively with Jenkins
2. **Standalone CLI Tool**: Go-based `piper` binary that can run independently on any platform

This dual nature allows Piper to:
- Provide rich Jenkins integration (credentials, stashing, UI features)
- Execute consistently across different CI/CD platforms
- Run steps locally for testing and development

## Key Statistics

- **188 Groovy Pipeline Steps** (`vars/`)
- **200+ Go CLI Commands** (`cmd/`)
- **70+ Go Packages** (`pkg/`)
- **153+ Step Metadata Files** (`resources/metadata/`)
- **59 Groovy Utility Classes** (`src/com/sap/piper/`)
- **Comprehensive Documentation** (`documentation/`)

## Core Components

### 1. Go CLI (`cmd/` and `pkg/`)

The Piper binary is the execution engine:

```bash
# Build the binary
make build

# Execute a step
./piper mavenBuild --pomPath=pom.xml --publish

# Get help
./piper --help
./piper mavenBuild --help
```

**Key Features**:
- Self-contained binary (no runtime dependencies)
- Cross-platform (Linux, macOS, Windows)
- Container-aware execution
- Rich parameter validation
- Vault integration for secrets

### 2. Jenkins Shared Library (`vars/` and `src/`)

Groovy wrappers that integrate with Jenkins:

```groovy
// Jenkinsfile
@Library('piper-lib-os') _

piperPipeline script: this
```

**Key Features**:
- Native Jenkins integration
- Credential management
- Workspace and node handling
- Build result publishing
- Stage orchestration

### 3. Metadata System (`resources/metadata/`)

YAML files that define each step:

```yaml
# Example: mavenBuild.yaml
metadata:
  name: mavenBuild
  description: Executes a Maven build
  longDescription: |
    This step executes a Maven build including...

spec:
  inputs:
    params:
      - name: pomPath
        type: string
        scope: [PARAMETERS, STAGES, STEPS]
        default: pom.xml
```

**Benefits**:
- Single source of truth
- Automatic code generation
- Type-safe configuration
- Consistent documentation

## Feature Categories

### Build Tools (22 steps)
- Maven, Gradle, npm, MTA
- Golang, Python, Docker, Kaniko
- Cloud Native Buildpacks (CNB)
- Generic build execution

### Security Scanning (12 steps)
- Checkmarx, Checkmarx One
- Fortify, WhiteSource/Mend
- Black Duck Detect, Snyk
- CodeQL, Protecode
- Contrast Security
- Credential scanning
- Malware scanning

### Testing Frameworks (7 steps)
- BATS, Gatling, Gauge
- Karma, Newman
- Selenium, UIVeri5

### SAP Integration (50+ steps)
- ABAP Environment (30+ steps)
- Integration Suite (10 steps)
- API Management (6 steps)
- gCTS (6 steps)
- Transport Management
- Cloud Foundry deployment

### Deployment (10+ steps)
- Cloud Foundry, Kubernetes
- SAP Neo, SAP XS
- Helm, Terraform
- Multi-cloud deployment
- GitOps

### Container Operations (8 steps)
- Docker build and execution
- Kaniko build
- CNB (Cloud Native Buildpacks)
- Container push to registries
- Structure tests

### Version Control (8 steps)
- GitHub integration (6 steps)
- Git operations
- Artifact versioning

### Utilities (30+ steps)
- Notifications (Mail, Slack, ANS)
- Test result publishing
- Report aggregation
- Debug reporting
- Health checks

## Pipeline Stages

### General Purpose Pipeline

The main pipeline (`piperPipeline`) includes these stages:

1. **Init**: Repository checkout, configuration, versioning
2. **Pull-Request Voting**: PR validation (Jenkins only)
3. **Build**: Compilation, unit tests, code quality
4. **Additional Unit Tests**: Extended testing (Karma, etc.)
5. **Integration**: Integration tests
6. **Acceptance**: End-to-end acceptance testing
7. **Security**: Security scans (SAST, dependency scanning)
8. **Performance**: Load and performance testing
9. **Compliance**: SonarQube, licensing
10. **Confirm**: Manual approval gate
11. **Promote**: Artifact promotion
12. **Release**: Production deployment
13. **Post**: Cleanup, notifications, reporting

### ABAP Environment Pipeline

Specialized pipeline for SAP ABAP development:

1. **Prepare System**: ABAP system setup
2. **Clone Repositories**: Git repository cloning
3. **Initial Checks**: Syntax and consistency checks
4. **Build**: Package assembly
5. **Test**: ABAP Unit tests
6. **Integration Test**: ATC checks
7. **Confirm**: Manual approval
8. **Publish**: Add-on publication
9. **Post**: Cleanup

## Architecture

### Component Interaction

```
Jenkinsfile
    ↓
Groovy Step (vars/mavenBuild.groovy)
    ↓
piperExecuteBin
    ↓
Piper Binary (cmd/mavenBuild.go)
    ↓
Technology Package (pkg/maven/)
    ↓
Tool Execution (mvn)
    ↓
Results → Common Pipeline Environment
```

### Configuration Flow

```
1. Built-in Defaults (resources/metadata/)
   ↓
2. Stage Defaults (resources/com.sap.piper/pipeline/)
   ↓
3. Custom Defaults (.pipeline/defaults.yml)
   ↓
4. Project Config (.pipeline/config.yml)
   ↓
5. Step Parameters (in pipeline code)
   ↓
6. Environment Variables (PIPER_*)
   ↓
7. Command-line Flags (highest priority)
```

## Usage Patterns

### 1. Complete Pipeline

```groovy
@Library('piper-lib-os') _

piperPipeline script: this
```

### 2. Individual Steps

```groovy
@Library('piper-lib-os') _

pipeline {
    agent any
    stages {
        stage('Build') {
            steps {
                script {
                    mavenBuild script: this
                }
            }
        }
        stage('Security') {
            steps {
                script {
                    checkmarxExecuteScan script: this
                }
            }
        }
    }
}
```

### 3. Custom Stages with Piper Steps

```groovy
@Library('piper-lib-os') _

pipeline {
    agent any
    stages {
        stage('Custom Build') {
            steps {
                script {
                    setupCommonPipelineEnvironment script: this

                    mavenBuild(
                        script: this,
                        pomPath: 'custom/pom.xml',
                        goals: ['clean', 'install'],
                        publish: true
                    )
                }
            }
        }
    }
}
```

## Configuration

### Project Configuration

Create `.pipeline/config.yml` in your project:

```yaml
general:
  productiveBranch: 'main'
  collectTelemetryData: false

stages:
  Build:
    mavenBuild: true
  Security:
    checkmarxExecuteScan: true

steps:
  mavenBuild:
    pomPath: 'pom.xml'
    goals: ['clean', 'install']
    publish: true

  checkmarxExecuteScan:
    projectName: 'MyProject'
    vulnerabilityThresholdHigh: 0
```

### Stage Configuration

Configure entire stages:

```yaml
stages:
  Build:
    karmaExecuteTests: false
    npmExecuteScripts: true

  Acceptance:
    cloudFoundryDeploy: true
    newmanExecute: true
```

### Step Configuration

Configure individual steps:

```yaml
steps:
  cloudFoundryDeploy:
    cloudFoundry:
      org: 'my-org'
      space: 'dev'
      credentialsId: 'cf-credentials'
    deployType: 'blue-green'
    keepOldInstance: true
```

## Extensibility

### Stage Extensions

Extend or override stages:

```groovy
// .pipeline/extensions/Build.groovy
void call(Map params) {
    echo "Before build stage"

    // Run original stage
    params.originalStage()

    echo "After build stage"
    // Add custom logic here
}
```

### Global Extensions

Load global extensions across all pipelines:

```groovy
// Shared library
piperLoadGlobalExtensions(
    script: this,
    globalExtensionsDirectory: 'globalExtensions'
)
```

## Key Features

### 1. Multi-Tool Support

Piper automatically detects and supports:
- Maven (pom.xml)
- npm (package.json)
- MTA (mta.yaml)
- Gradle (build.gradle)
- Go (go.mod)
- Python (setup.py)
- Docker (Dockerfile)

### 2. Security-First

12 integrated security scanners:
- SAST (Static Application Security Testing)
- SCA (Software Composition Analysis)
- Container scanning
- Secret detection
- License compliance

### 3. SAP Ecosystem Integration

Deep integration with SAP technologies:
- SAP BTP (Cloud Foundry, Neo, XS)
- ABAP Development
- SAP Integration Suite
- SAP API Management
- SAP Transport Management

### 4. Cloud Native

Built for cloud and containers:
- Docker and Kubernetes native
- Helm chart support
- Cloud Native Buildpacks
- Multi-cloud deployment (AWS, Azure, GCP)

### 5. Enterprise Ready

Features for enterprise environments:
- HashiCorp Vault integration
- Change management integration
- Audit trail and reporting
- Compliance checking

## Documentation Structure

```
jenkins-library/documentation/
├── docs/
│   ├── index.md                    # Main documentation
│   ├── guidedtour.md              # Getting started
│   ├── configuration.md            # Configuration guide
│   ├── extensibility.md            # Extensibility guide
│   ├── steps/                      # Step documentation (153 files)
│   ├── stages/                     # Stage documentation
│   ├── scenarios/                  # Use case scenarios
│   └── infrastructure/             # Infrastructure setup
└── mkdocs.yml                      # MkDocs configuration
```

## Getting Started

### Prerequisites

1. Jenkins with Pipeline plugin
2. Docker (for container execution)
3. Git

### Quick Start

1. **Add Piper library to Jenkins**:
   - Manage Jenkins → Configure System → Global Pipeline Libraries
   - Name: `piper-lib-os`
   - Default version: `master`
   - Retrieval method: Modern SCM → Git
   - Project Repository: `https://github.com/SAP/jenkins-library.git`

2. **Create Jenkinsfile**:
```groovy
@Library('piper-lib-os') _

piperPipeline script: this
```

3. **Create configuration** (`.pipeline/config.yml`):
```yaml
general:
  productiveBranch: 'main'

stages:
  Build:
    mavenBuild: true
```

4. **Run pipeline**

## Next Steps

- [Build Tools Documentation](./01-build-tools.md)
- [Security Scanning Documentation](./02-security-scanning.md)
- [Testing Frameworks Documentation](./03-testing-frameworks.md)
- [Deployment Documentation](./04-deployment.md)
- [SAP Integration Documentation](./05-sap-integration.md)
- [ABAP Development Documentation](./06-abap-development.md)
- [Container Operations Documentation](./07-container-operations.md)
- [Version Control Documentation](./08-version-control.md)
- [Utilities Documentation](./09-utilities.md)

## Resources

- **GitHub**: https://github.com/SAP/jenkins-library
- **Documentation**: https://www.project-piper.io/
- **Step Reference**: Browse `documentation/docs/steps/` for individual step documentation

---

*jenkins-library is the foundation of Project Piper, providing comprehensive CI/CD capabilities for the SAP ecosystem.*

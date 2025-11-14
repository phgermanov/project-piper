# Piper CI/CD Architecture

This document describes the overall architecture of the Project Piper CI/CD tool and how its components work together.

## High-Level Architecture

Piper follows a layered architecture that separates platform-specific concerns from core CI/CD logic:

```
┌─────────────────────────────────────────────────────────┐
│         Platform Layer (Jenkins/GitHub/Azure)           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   Jenkins    │  │   GitHub     │  │    Azure     │  │
│  │   Groovy     │  │   Action     │  │   Pipeline   │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────┐
│              Stage Orchestration Layer                   │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐         │
│  │ Init │→│ Build│→│ Test │→│ Scan │→│Deploy│         │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘         │
└─────────────────────────────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────┐
│                Core Piper Binary (Go)                    │
│  ┌─────────────────────────────────────────────────┐   │
│  │  Configuration System                            │   │
│  │  • Merge from multiple sources                   │   │
│  │  • Vault integration                             │   │
│  │  • Validation                                    │   │
│  └─────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────┐   │
│  │  Step Execution Engine                           │   │
│  │  • 200+ Generated Steps                          │   │
│  │  • Container Management                          │   │
│  │  • Error Handling                                │   │
│  └─────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────┐   │
│  │  Common Pipeline Environment (CPE)               │   │
│  │  • Filesystem-based state store                  │   │
│  │  • JSON serialization                            │   │
│  │  • Cross-step data sharing                       │   │
│  └─────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────┐
│            Technology-Specific Packages                  │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐         │
│  │Maven │ │ NPM  │ │Docker│ │ K8s  │ │ etc. │         │
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘         │
└─────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Piper Binary (Go)

**Location**: `jenkins-library/cmd/`

The heart of Piper is a standalone Go binary that implements all pipeline steps:

- **Entry Point**: `main.go` → `cmd/piper.go`
- **CLI Framework**: Uses Cobra for command routing
- **Step Implementation**: ~200+ commands implemented in Go
- **Metadata-Driven**: Steps generated from YAML metadata files

**Execution Flow**:
```
1. User invokes: piper <stepName> [flags]
2. Cobra routes to step command
3. Configuration loaded and merged from multiple sources
4. Step validation and execution
5. Results written to Common Pipeline Environment
6. Exit with status code
```

### 2. Configuration System

**Location**: `jenkins-library/pkg/config/`

Multi-source configuration management with precedence hierarchy:

```
Highest Priority
    ↓
1. Command-line flags
2. Environment variables (PIPER_*)
3. Parameters JSON
4. Custom config file (.pipeline/config.yml)
5. Custom defaults
6. Stage-specific config
7. Built-in defaults
    ↓
Lowest Priority
```

**Key Features**:
- Type-safe parameter handling
- Secret resolution from Vault
- Validation against metadata schemas
- Platform-specific deviations
- Scope-based parameter visibility (GENERAL, STAGES, STEPS, PARAMETERS)

### 3. Common Pipeline Environment (CPE)

**Location**: `jenkins-library/pkg/piperenv/`

Filesystem-based state management for cross-step communication:

**Directory Structure**:
```
.pipeline/
  commonPipelineEnvironment/
    git/
      commitId              # Plain text
      branch                # Plain text
    custom/
      buildArtifacts.json   # JSON
      mavenSettings         # String
    artifact/
      version               # String
```

**Implementation**:
- Key-value store using filesystem
- JSON serialization for complex types
- String storage for simple values
- Hierarchical directory structure

### 4. Platform Adapters

#### Jenkins Adapter
**Location**: `jenkins-library/vars/`

Groovy shared library that wraps Piper binary:

```groovy
// vars/piperPipelineStageBuild.groovy
void call(Map parameters = [:]) {
    piperStageWrapper (script: script, stageName: stageName) {
        // Execute Piper binary steps
        buildExecute script: script

        // Jenkins-specific features
        pipelineStashFilesAfterBuild script: script
    }
}
```

**Features**:
- Jenkins-native stashing/unstashing
- Credentials management
- Node and workspace handling
- Build metadata publishing

#### GitHub Actions Adapter
**Location**: `project-piper-action/`

TypeScript action that downloads and executes Piper:

```typescript
// src/piper.ts
export async function run() {
    await preparePiperBinary(actionCfg)
    await loadPipelineEnv()      // Load CPE from artifacts
    await executePiper(stepName, flags)
    await exportPipelineEnv()    // Save CPE to artifacts
}
```

**Features**:
- Binary caching between runs
- CPE artifact upload/download
- Container execution
- GitHub API integration

#### Azure DevOps Adapter
**Location**: `piper-azure-task/` and `piper-pipeline-azure/`

Azure DevOps extension and YAML templates:

```yaml
# stages/build.yml
steps:
  - template: ../steps/piper-step.yml
    parameters:
      stepName: 'mavenBuild'
      stepParams: '--publish --createBOM'
```

## Design Patterns

### 1. Strategy Pattern
Different implementations for common interfaces:
- Build tools (Maven, NPM, Gradle, MTA)
- Deployment targets (Cloud Foundry, Kubernetes, Neo)
- Version management (per build tool)

### 2. Template Method Pattern
Pipeline stages define structure, steps provide implementation:
```
Stage (Abstract)
  ├── preStage() - Extension point
  ├── executeSteps() - Implemented by concrete stage
  └── postStage() - Extension point
```

### 3. Command Pattern
Each step is a discrete, executable command:
```go
// cmd/mavenBuild.go
func addMavenBuildFlags(cmd *cobra.Command) {
    // Flag definitions
}

func mavenBuild(config mavenBuildOptions, telemetryData *telemetry.CustomData) {
    // Step implementation
}
```

### 4. Repository Pattern
Configuration abstraction with multiple sources:
```go
type ConfigProvider interface {
    GetConfig() (Config, error)
}

// Implementations:
// - FileConfigProvider
// - VaultConfigProvider
// - EnvConfigProvider
```

### 5. Builder Pattern
Configuration assembly:
```go
config := ConfigurationHelper{
    config: defaultConfig,
    filters: defaultFilters,
}.merge(customConfig).
  merge(stepConfig).
  validate().
  build()
```

### 6. Factory Pattern
Step generation from metadata:
```go
// Generated from metadata YAML
func NewMavenBuildCommand() *cobra.Command {
    metadata := getStepMetadata("mavenBuild")
    return createCommandFromMetadata(metadata)
}
```

## Integration Points

### Key Integration Architecture

```
┌──────────────────────┐
│   Vault Server       │←─────── Secret retrieval
└──────────────────────┘
         ▲
         │
┌──────────────────────┐
│  Piper Binary        │
└──────────────────────┘
         │
         ├──────────→ Docker Daemon (container execution)
         │
         ├──────────→ File System
         │            • .pipeline/ (config)
         │            • CPE files (state)
         │            • Stash files (artifacts)
         │
         ├──────────→ CI/CD Platform
         │            • Jenkins
         │            • GitHub Actions
         │            • Azure DevOps
         │
         └──────────→ External Services
                      • SonarQube
                      • Checkmarx
                      • Cloud providers
                      • Artifact repositories
```

### Service Integration Points

1. **Source Control**: Git operations, branch detection, tagging
2. **Artifact Repositories**: Nexus, Artifactory, npm registry, Maven Central
3. **Security Scanners**: Checkmarx, Fortify, WhiteSource, Snyk, etc.
4. **Quality Tools**: SonarQube, linters, test frameworks
5. **Cloud Platforms**: SAP BTP, AWS, Azure, GCP, Kubernetes
6. **SAP Services**: ABAP systems, Integration Suite, API Management
7. **Secret Management**: HashiCorp Vault, CI/CD platform secrets
8. **Notification Services**: Email, Slack, SAP Alert Notification

## Metadata-Driven Architecture

### Metadata Schema
**Location**: `jenkins-library/resources/metadata/*.yaml`

Each step is defined by a YAML metadata file:

```yaml
metadata:
  name: mavenBuild
  description: This step will build the Maven project
  longDescription: |
    Executes a Maven build including...

spec:
  inputs:
    secrets:
      - name: altDeploymentRepositoryPasswordId
        description: Jenkins credential ID
        type: jenkins

    params:
      - name: pomPath
        type: string
        description: Path to pom.xml
        scope:
          - PARAMETERS
          - STAGES
          - STEPS
        mandatory: false
        default: pom.xml

      - name: goals
        type: "[]string"
        description: Maven goals to execute
        scope:
          - PARAMETERS
          - STAGES
          - STEPS
        mandatory: false
        default: ['install']

    resources:
      - name: buildDescriptor
        type: stash

  containers:
    - name: maven
      image: maven:3.6-jdk-8

  outputs:
    resources:
      - name: commonPipelineEnvironment
        type: piperEnvironment
        params:
          - name: custom/buildSettingsInfo
          - name: custom/mavenExecute
```

### Code Generation

Steps are generated from metadata:

```
1. Metadata YAML files (resources/metadata/)
   ↓
2. Generator reads metadata (pkg/generator/)
   ↓
3. Go code generated (cmd/<stepName>_generated.go)
   ↓
4. Custom implementation (cmd/<stepName>.go)
   ↓
5. Compiled into piper binary
```

**Benefits**:
- Consistency across steps
- Automatic flag generation
- Type-safe configuration
- Documentation generation
- Validation rules

## Container Execution Model

### Docker Integration

Steps can run in containers for consistency:

```yaml
# Metadata specifies container
containers:
  - name: maven
    image: maven:3.6-jdk-8
    env:
      - name: MAVEN_OPTS
        value: "-Xmx1024m"
    options:
      - name: -u
        value: "1000"
```

**Container Execution Flow**:
```
1. Check if containerized execution required
2. Pull/verify Docker image
3. Mount workspace and CPE directories
4. Set environment variables
5. Execute command in container
6. Capture output and exit code
7. Extract artifacts from container
```

### Kubernetes Execution

For Jenkins, steps can execute on Kubernetes pods:

```groovy
dockerExecuteOnKubernetes(
    script: script,
    dockerImage: 'maven:3.6-jdk-8'
) {
    mavenBuild script: script
}
```

## Extensibility Architecture

### Extension Points

1. **Stage Extensions**
```groovy
// .pipeline/extensions/Build.groovy
void call(Map params) {
    echo "Before build"
    params.originalStage()  // Run original stage
    echo "After build"
}
```

2. **Pre/Post Steps** (GitHub/Azure)
```yaml
parameters:
  buildPreSteps:
    - script: echo "Before"
  buildPostSteps:
    - script: echo "After"
```

3. **Custom Defaults**
```yaml
# .pipeline/config.yml
customDefaults: ['custom-defaults.yml']
```

4. **Step Override**
```yaml
# .pipeline/config.yml
steps:
  mavenBuild:
    dockerImage: 'custom-maven:latest'
```

## Security Architecture

### Secret Management

1. **Vault Integration**:
```yaml
steps:
  mavenBuild:
    altDeploymentRepositoryPassword:
      vaultPath: 'maven'
      vaultKey: 'password'
```

2. **CI/CD Platform Secrets**:
```groovy
// Jenkins
withCredentials([string(credentialsId: 'id', variable: 'PWD')]) {
    mavenBuild script: script
}
```

3. **Environment Variables**:
```bash
export PIPER_altDeploymentRepositoryPassword=secret
piper mavenBuild
```

### Credential Rotation

Vault secret ID rotation:
```yaml
steps:
  vaultRotateSecretId:
    vaultServerUrl: https://vault.example.com
```

## Performance Optimizations

1. **Binary Caching**: Reuse Piper binary across runs
2. **Configuration Caching**: Cache merged configuration
3. **Download Caching**: Cache downloaded tools and dependencies
4. **Parallel Execution**: Independent steps run concurrently
5. **Conditional Execution**: Skip inactive stages/steps
6. **Stashing**: Efficient artifact transfer between stages

## Error Handling

### Error Propagation

```
Step Error
  ↓
Captured by step handler
  ↓
Written to CPE (error state)
  ↓
Stage marked as failed
  ↓
Pipeline status updated
  ↓
Notifications sent
  ↓
Post stage executes (even on failure)
```

### Retry Mechanisms

Configurable retry for network operations:
```yaml
steps:
  mavenBuild:
    retry: 3
    retryDelay: 10s
```

## Telemetry and Monitoring

### Data Collection

1. **Execution Metrics**: Duration, success/failure
2. **Configuration**: Active steps, parameters used
3. **Environment**: OS, platform, container info
4. **Errors**: Error messages, stack traces

### Reporting Integration

- InfluxDB for time-series metrics
- Splunk for log aggregation
- Custom reporting hooks
- Pipeline status dashboards

## Summary

Piper's architecture demonstrates:

- **Separation of Concerns**: Clear boundaries between platform, orchestration, and execution
- **Extensibility**: Multiple extension points at every layer
- **Reusability**: Shared core binary across platforms
- **Flexibility**: Multi-source configuration with precedence rules
- **Maintainability**: Metadata-driven code generation
- **Scalability**: Containerized execution and parallel processing
- **Security**: Multiple secret management options
- **Observability**: Comprehensive telemetry and monitoring

This architecture enables Piper to serve as a comprehensive CI/CD solution for diverse SAP ecosystem requirements while maintaining consistency and best practices across platforms.

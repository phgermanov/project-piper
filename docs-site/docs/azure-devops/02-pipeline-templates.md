# Piper Pipeline Templates for Azure DevOps

## Overview

The `piper-pipeline-azure` repository provides a complete general purpose pipeline (GPP) template implementing a multi-stage CI/CD workflow for Azure DevOps.

## Quick Start

```yaml
# azure-pipelines.yml
trigger: [main]

resources:
  repositories:
    - repository: piper-templates
      endpoint: github.tools.sap
      type: githubenterprise
      name: project-piper/piper-pipeline-azure
      ref: refs/tags/v1.2.3

extends:
  template: sap-piper-pipeline.yml@piper-templates
```

```yaml
# .pipeline/config.yml
general:
  productiveBranch: 'main'

stages:
  Build:
    npmExecuteScripts: true

steps:
  npmExecuteScripts:
    runScripts: ['ci-build', 'ci-test']
```

## Pipeline Stages

**Stage Flow**: Init → Build → Security → Integration → Acceptance → Performance → IP Scan/PPMS → Confirm → Promote → Release → Post

### Init
Initializes environment, fetches defaults, determines stage activation.

### Build
Builds application, runs tests, creates artifacts. Supports Maven, Gradle, npm, Go, Python, MTA, Docker, Helm.

### Security
Runs SAST (SonarQube) and OSS vulnerability scanning.

### Integration/Acceptance/Performance
Optional testing stages (disabled by default).

### IP Scan/PPMS
License compliance and SAP PPMS integration.

### Confirm
Manual approval gate (requires `manualConfirmation: true` in config).

### Promote/Release
Promotes artifacts and deploys to production.

### Post
Cleanup and notifications.

## Common Parameters

```yaml
extends:
  template: sap-piper-pipeline.yml@piper-templates
  parameters:
    # Versions
    piperVersion: 'v1.150.0'
    sapPiperVersion: 'v2.45.0'

    # Configuration
    customDefaults: 'https://raw.githubusercontent.com/org/defaults.yml'

    # Repository
    checkoutSubmodule: true
    checkoutLFS: false

    # Agent Pools
    msHostedPool: 'Azure Pipelines'
    poolVmImage: 'ubuntu-latest'
    selfHostedPool: 'Self-hosted'

    # Service Connections
    gitHubToolsConnectionName: 'github.tools.sap'
    dockerRegistryConnection: 'my-registry'

    # Caching
    disableCachingOnMSHostedAgents: false
    disableCachingOnSelfHostedAgents: false

    # Stage Customization
    buildPreSteps: []
    buildPostSteps: []
    buildServiceContainers: {}

    # Skip Stages
    skipIntegrationStage: false
    skipAcceptanceStage: true
    skipPerformanceStage: true
    skipReleaseStage: false

    # Manual Confirmation
    manualConfirmationTimeoutMinutes: 43200  # 30 days
```

## Stage Customization

Each stage supports:
- `<stage>PreSteps`: Steps before stage
- `<stage>PostSteps`: Steps after stage
- `<stage>OverwriteSteps`: Replace all steps
- `<stage>ServiceContainers`: Service containers
- `skip<Stage>Stage`: Skip stage

**Example**:

```yaml
parameters:
  buildPreSteps:
    - script: echo "Pre-build"
  buildServiceContainers:
    postgres:
      image: postgres:13
      env:
        POSTGRES_PASSWORD: test
  integrationOverwriteSteps:
    - script: npm run integration-tests
```

## Configuration Patterns

### Java Maven

```yaml
# .pipeline/config.yml
general:
  productiveBranch: 'main|release/.*'

stages:
  Build:
    mavenBuild: true
    kanikoExecute: true
    sonarExecuteScan: true

steps:
  mavenBuild:
    pomPath: 'pom.xml'
    goals: ['clean', 'install']
  kanikoExecute:
    containerImageNameAndTag: 'myapp:${version}'
  sonarExecuteScan:
    serverUrl: 'https://sonar.example.com'
```

### Node.js

```yaml
# .pipeline/config.yml
stages:
  Build:
    npmExecuteScripts: true
    cnbBuild: true

steps:
  npmExecuteScripts:
    runScripts: ['ci-build', 'ci-test']
  cnbBuild:
    containerImageName: 'myapp'
```

### Go

```yaml
# .pipeline/config.yml
stages:
  Build:
    golangBuild: true
    kanikoExecute: true

steps:
  golangBuild:
    createBOM: true
    runTests: true
  kanikoExecute:
    containerImageNameAndTag: 'mygoapp:${version}'
```

### Python

```yaml
# .pipeline/config.yml
stages:
  Build:
    pythonBuild: true

steps:
  pythonBuild:
    publish: true
    targetRepositoryURL: 'https://pypi.example.com'
```

## Advanced Customization

### Custom Steps

```yaml
parameters:
  buildPreSteps:
    - task: DownloadSecureFile@1
      name: key
      inputs:
        secureFile: 'deploy_key'
    - script: cp $(key.secureFilePath) ~/.ssh/id_rsa

  buildPostSteps:
    - script: ./scripts/post-build.sh
```

### Service Containers

```yaml
parameters:
  integrationServiceContainers:
    redis:
      image: redis:6-alpine
      ports: ['6379:6379']
    postgres:
      image: postgres:13
      env:
        POSTGRES_PASSWORD: test
      ports: ['5432:5432']
```

### Multi-Stage Variables

```yaml
# Export in one stage
- bash: echo "##vso[task.setvariable variable=ver;isOutput=true]1.0.0"
  name: exportVer

# Use in another stage
- bash: echo "$(stageDependencies.Build.job.outputs['exportVer.ver'])"
```

## Best Practices

### 1. Version Pinning

```yaml
resources:
  repositories:
    - repository: piper-templates
      ref: refs/tags/v1.2.3  # Not 'main'
```

### 2. Branch Strategy

```yaml
general:
  productiveBranch: 'main|release/.*|hotfix/.*'
```

### 3. Minimal Pipeline File

Keep `azure-pipelines.yml` minimal, use `.pipeline/config.yml` for configuration.

### 4. Organization Defaults

```yaml
parameters:
  customDefaults: 'https://github.company.com/devops/defaults.yml'
```

### 5. Manual Confirmation

```yaml
general:
  manualConfirmation: true
```

### 6. Secure Secrets

```yaml
variables:
  - group: 'production-secrets'
```

## Troubleshooting

### Stage Not Running

Check Init stage outputs for activation:

```yaml
- bash: |
    echo "Active Build: $(checkIfStepActive.activeBuild)"
    echo "Active steps: $(activeStepsBuild)"
```

### Step Skipped

Check step activation:

```yaml
- bash: cat .pipeline/step_out.json
```

### Configuration Issues

Debug configuration:

```yaml
- task: piper@1
  inputs:
    stepName: 'getConfig'
    flags: '--outputFile config.json'
- bash: cat config.json
```

## Complete Examples

### Minimal Node.js

```yaml
# azure-pipelines.yml
trigger: [main]

resources:
  repositories:
    - repository: piper-templates
      endpoint: github.tools.sap
      type: githubenterprise
      name: project-piper/piper-pipeline-azure

extends:
  template: sap-piper-pipeline.yml@piper-templates
```

```yaml
# .pipeline/config.yml
stages:
  Build:
    npmExecuteScripts: true

steps:
  npmExecuteScripts:
    runScripts: ['ci-build', 'ci-test']
```

### Full-Featured Java

```yaml
# azure-pipelines.yml
trigger: [main, develop]

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
    checkoutSubmodule: true
    skipPerformanceStage: true
    buildServiceContainers:
      postgres:
        image: postgres:13
        env:
          POSTGRES_PASSWORD: test
    manualConfirmationTimeoutMinutes: 1440
```

```yaml
# .pipeline/config.yml
general:
  productiveBranch: 'main'
  manualConfirmation: true

stages:
  Build:
    mavenBuild: true
    kanikoExecute: true
    sonarExecuteScan: true
  Security:
    whitesourceExecuteScan: true

steps:
  mavenBuild:
    pomPath: 'pom.xml'
    goals: ['clean', 'install']
  kanikoExecute:
    dockerConfigJsonCredentialsId: 'docker-credentials'
    containerImageNameAndTag: 'myapp:${version}'
  sonarExecuteScan:
    serverUrl: 'https://sonarqube.company.com'
    projectKey: 'my-project'
```

## Reference

### Template Structure

```
piper-pipeline-azure/
├── sap-piper-pipeline.yml    # Main template
├── stages/                    # Stage definitions
│   ├── init.yml
│   ├── build.yml
│   ├── security.yml
│   ├── integration.yml
IP Scan/PPMS
  ↓
Confirm (manual)
  ↓
Promote
  ↓
Release (optional)
  ↓
Post (always)
```

### Key Variables

Set by Init stage:
- `sapDefaults_b64`: Base-64 encoded config
- `onProductiveBranch`: Branch indicator
- `binaryVersion`, `osBinaryVersion`: Binary versions
- `active<Stage>`: Stage activation
- `activeSteps<Stage>`: Step activation

### Links

- [Piper Azure Task](01-azure-task.md)
- [Azure DevOps Overview](00-overview.md)
- [Project Piper Docs](https://www.project-piper.io)
- [Repository](https://github.tools.sap/project-piper/piper-pipeline-azure)

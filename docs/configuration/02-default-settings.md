# Default Settings

Project Piper ships with comprehensive default configuration covering common CI/CD scenarios. This document explores the built-in defaults and explains how to effectively use and override them.

## Table of Contents

- [Default Settings](#default-settings)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Default Configuration File](#default-configuration-file)
  - [General Defaults](#general-defaults)
    - [Core Settings](#core-settings)
    - [Git Configuration](#git-configuration)
    - [Docker Configuration](#docker-configuration)
    - [Change Management](#change-management)
  - [Build Tool Defaults](#build-tool-defaults)
    - [Maven](#maven)
    - [NPM](#npm)
    - [Gradle](#gradle)
    - [Docker](#docker)
    - [Golang](#golang)
  - [Step-Specific Defaults](#step-specific-defaults)
    - [Version Management](#version-management)
    - [Cloud Foundry Deployment](#cloud-foundry-deployment)
    - [Testing Steps](#testing-steps)
    - [Quality Checks](#quality-checks)
    - [Security Scanning](#security-scanning)
  - [Stage Wrapper Defaults](#stage-wrapper-defaults)
  - [Docker Image Defaults](#docker-image-defaults)
  - [Stash Configuration Defaults](#stash-configuration-defaults)
    - [Before Build Stashing](#before-build-stashing)
    - [After Build Stashing](#after-build-stashing)
  - [Notification Defaults](#notification-defaults)
  - [Working with Defaults](#working-with-defaults)
    - [Viewing Default Values](#viewing-default-values)
    - [Understanding Default Behavior](#understanding-default-behavior)
    - [When to Override Defaults](#when-to-override-defaults)
  - [Best Practices](#best-practices)

## Overview

Default configuration in Project Piper:

- **Location**: `resources/default_pipeline_environment.yml`
- **Purpose**: Provide sensible defaults for all supported scenarios
- **Scope**: Covers 100+ pipeline steps
- **Maintenance**: Updated with each library release
- **Customization**: Can be viewed but not modified directly

**Key Principle**: Defaults are designed to work for common cases while allowing easy customization for specific needs.

## Default Configuration File

The complete default configuration is available at:

```
https://github.com/SAP/jenkins-library/blob/master/resources/default_pipeline_environment.yml
```

**Structure**:
```yaml
general:
  # Global settings

steps:
  # Step-specific defaults
  stepName:
    parameter: value

  # Build tool-specific nested config
  anotherStep:
    maven:
      setting: value
    npm:
      setting: value
```

## General Defaults

### Core Settings

```yaml
general:
  # Telemetry
  collectTelemetryData: false      # Disabled by default

  # Logging
  logFormat: 'plain'                # Format: plain, json

  # Branch configuration
  productiveBranch: 'master'        # Main branch for releases

  # Docker
  dockerPullImage: true             # Pull images before use
  sidecarPullImage: true           # Pull sidecar images

  # Extensions
  globalExtensionsDirectory: '.pipeline/tmp/global_extensions/'
```

**Usage example**:
```yaml
# Override in your config.yml
general:
  productiveBranch: 'main'         # Use 'main' instead of 'master'
  collectTelemetryData: false      # Keep telemetry disabled
  logFormat: 'json'                # Switch to JSON logging
```

### Git Configuration

```yaml
general:
  gitSshKeyCredentialsId: ''       # SSH key for Git operations
  githubApiUrl: 'https://api.github.com'
  githubServerUrl: 'https://github.com'
```

**Common overrides**:
```yaml
general:
  # GitHub Enterprise
  githubApiUrl: 'https://github.company.com/api/v3'
  githubServerUrl: 'https://github.company.com'
  gitSshKeyCredentialsId: 'github-enterprise-ssh'
```

### Docker Configuration

```yaml
general:
  dockerPullImage: true
  sidecarPullImage: true

  jenkinsKubernetes:
    jnlpAgent: 'jenkins/inbound-agent:jdk17'
```

### Change Management

```yaml
general:
  changeManagement:
    type: 'NONE'                    # Options: SOLMAN, CTS, NONE
    transportRequestLabel: 'TransportRequest\s?:'
    changeDocumentLabel: 'ChangeDocument\s?:'
    clientOpts: ''
    credentialsId: 'CM'

    git:
      from: 'origin/master'
      to: 'HEAD'
      format: '%b'

    solman:
      docker:
        image: 'ppiper/cm-client:2.0.1.0'
        pullImage: true

    cts:
      osDeployUser: 'node'
      deployToolDependencies:
        - '@ui5/cli'
        - '@sap/ux-ui5-tooling'
      deployConfigFile: 'ui5-deploy.yaml'
```

## Build Tool Defaults

### Maven

```yaml
steps:
  mavenExecute:
    dockerImage: 'maven:3.5-jdk-7'
    logSuccessfulMavenTransfers: false

  buildExecute:
    npmInstall: true
    npmRunScripts: []
```

**Recommended override** (use newer Java version):
```yaml
steps:
  mavenExecute:
    dockerImage: 'maven:3.8-openjdk-17'
```

### NPM

```yaml
steps:
  npmExecute:
    dockerImage: 'node:lts-bookworm'

  npmExecuteScripts:
    install: true

  npmExecuteEndToEndTests:
    runScript: 'ci-e2e'
```

**Example customization**:
```yaml
steps:
  npmExecute:
    dockerImage: 'node:18-alpine'

  npmExecuteScripts:
    install: true
    runScripts:
      - 'build'
      - 'test'
```

### Gradle

Gradle uses similar defaults to Maven but with Gradle-specific images.

### Docker

```yaml
steps:
  dockerExecute:
    stashContent: []

  dockerExecuteOnKubernetes:
    stashContent: []
    stashIncludes:
      workspace: '**/*'
    stashExcludes:
      workspace: 'nohup.out'
      stashBack: '**/node_modules/**,nohup.out,.git/**'
```

### Golang

```yaml
steps:
  artifactSetVersion:
    golang:
      filePath: 'VERSION'
      versioningTemplate: '${version}-${timestamp}${commitId?"+"+commitId:""}'
```

## Step-Specific Defaults

### Version Management

```yaml
steps:
  artifactSetVersion:
    timestampTemplate: '%Y%m%d%H%M%S'
    tagPrefix: 'build_'
    commitVersion: true
    gitPushMode: 'SSH'
    verbose: false
    gitHttpsCredentialsId: 'git'
    gitDisableSslVerification: false

    # Tool-specific templates
    maven:
      filePath: 'pom.xml'
      versioningTemplate: '${version}-${timestamp}${commitId?"_"+commitId:""}'

    npm:
      filePath: 'package.json'
      versioningTemplate: '${version}-${timestamp}${commitId?"+"+commitId:""}'

    docker:
      filePath: 'Dockerfile'
      versioningTemplate: '${version}-${timestamp}${commitId?"_"+commitId:""}'

    mta:
      filePath: 'mta.yaml'
      versioningTemplate: '${version}-${timestamp}${commitId?"+"+commitId:""}'

    golang:
      filePath: 'VERSION'
      versioningTemplate: '${version}-${timestamp}${commitId?"+"+commitId:""}'

    pip:
      filePath: 'version.txt'
      versioningTemplate: '${version}.${timestamp}${commitId?"."+commitId:""}'
```

**Customization example**:
```yaml
steps:
  artifactSetVersion:
    commitVersion: true
    tagPrefix: 'release_'
    maven:
      versioningTemplate: '${version}'  # No timestamp in version
```

### Cloud Foundry Deployment

```yaml
steps:
  cloudFoundryDeploy:
    cloudFoundry:
      apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'
    apiParameters: ''
    loginParameters: ''
    deployType: 'standard'            # Options: standard, blue-green
    keepOldInstance: false
    cfNativeDeployParameters: ''
    mtaDeployParameters: '-f'
    mtaExtensionDescriptor: ''
    mtaPath: ''
    smokeTestScript: 'blueGreenCheckScript.sh'
    smokeTestStatusCode: 200

    stashContent:
      - 'deployDescriptor'
      - 'pipelineConfigAndTests'

    cf_native:
      dockerImage: 'ppiper/cf-cli:latest'
      dockerWorkspace: '/home/piper'

    mtaDeployPlugin:
      dockerImage: 'ppiper/cf-cli:latest'
      dockerWorkspace: '/home/piper'

    # Build tool-specific deploy tools
    mta:
      deployTool: 'mtaDeployPlugin'
    maven:
      deployTool: 'cf_native'
    npm:
      deployTool: 'cf_native'
```

**Override example**:
```yaml
steps:
  cloudFoundryDeploy:
    deployType: 'blue-green'
    keepOldInstance: true
    cloudFoundry:
      org: 'my-org'
      space: 'production'
```

### Testing Steps

**Karma (Unit Tests)**:
```yaml
steps:
  karmaExecuteTests:
    containerPortMappings:
      'node:lts-bookworm':
        - containerPort: 9876
          hostPort: 9876

    dockerEnvVars:
      NO_PROXY: 'localhost,selenium,$NO_PROXY'
      no_proxy: 'localhost,selenium,$no_proxy'

    dockerImage: 'node:lts-bookworm'
    dockerName: 'karma'
    dockerWorkspace: '/home/node'
    installCommand: 'npm install --quiet'
    modules:
      - '.'
    runCommand: 'npm run karma'

    stashContent:
      - buildDescriptor
      - tests
```

**Selenium (E2E Tests)**:
```yaml
steps:
  seleniumExecuteTests:
    buildTool: 'npm'
    containerPortMappings:
      'selenium/standalone-chrome':
        - containerPort: 4444
          hostPort: 4444

    failOnError: true
    sidecarImage: 'selenium/standalone-chrome'
    sidecarName: 'selenium'
    sidecarVolumeBind:
      '/dev/shm': '/dev/shm'

    stashContent:
      - 'tests'

    npm:
      dockerImage: 'node:lts-bookworm'
      dockerName: 'npm'
      dockerWorkspace: '/home/node'

    maven:
      dockerImage: 'maven:3.5-jdk-8'
      dockerName: 'maven'
      dockerWorkspace: ''
```

**Gatling (Performance Tests)**:
```yaml
steps:
  gatlingExecuteTests:
    stashContent:
      - 'source'
```

### Quality Checks

```yaml
steps:
  checksPublishResults:
    failOnError: false

    tasks:
      pattern: '**/*.java'
      low: ''
      normal: 'TODO,REVISE,XXX'
      high: 'FIXME'
      archive: true
      active: false
      qualityGates:
        - threshold: 1
          type: 'TOTAL_HIGH'
          unstable: false

    pmd:
      pattern: '**/target/pmd.xml'
      archive: true
      active: false
      qualityGates:
        - threshold: 1
          type: 'TOTAL_HIGH'
          unstable: false

    checkstyle:
      pattern: '**/target/checkstyle-result.xml'
      archive: true
      active: false
      qualityGates:
        - threshold: 1
          type: 'TOTAL_HIGH'
          unstable: false

    findbugs:
      pattern: '**/target/findbugsXml.xml, **/target/findbugs.xml'
      archive: true
      active: false

    eslint:
      pattern: '**/eslint.xml'
      archive: true
      active: false
```

**Enable quality checks**:
```yaml
steps:
  checksPublishResults:
    pmd:
      active: true
    checkstyle:
      active: true
    eslint:
      active: true
```

### Security Scanning

```yaml
steps:
  whitesourceExecuteScan:
    vulnerabilityReportFileName: 'piper_whitesource_vulnerability_report'
    stashExcludes:
      stashBack: '.pipeline/** whitesourceExecuteScan_*.* whitesource-reports/**'

  snykExecute:
    buildDescriptorFile: './package.json'
    dockerImage: 'node:lts-bookworm'
    exclude: []
    monitor: true
    scanType: 'npm'
    stashContent:
      - 'buildDescriptor'
      - 'opensourceConfiguration'
    toJson: false
    toHtml: false
```

## Stage Wrapper Defaults

```yaml
steps:
  piperStageWrapper:
    projectExtensionsDirectory: '.pipeline/extensions/'
    stageLocking: true
    nodeLabel: ''
    stashContent:
      - 'pipelineConfigAndTests'
```

## Docker Image Defaults

**By Build Tool**:

| Build Tool | Default Image | Purpose |
|------------|---------------|---------|
| Maven | `maven:3.5-jdk-7` | Java builds (consider upgrading) |
| NPM | `node:lts-bookworm` | Node.js builds |
| Golang | (varies by step) | Go builds |
| Docker | (user-defined) | Container builds |
| Dub | `dlang2/dmd-ubuntu:latest` | D language builds |
| SBT | (varies) | Scala builds |

**By Step Type**:

| Step | Default Image | Purpose |
|------|---------------|---------|
| Cloud Foundry | `ppiper/cf-cli:latest` | CF deployment |
| Neo | `ppiper/neo-cli` | SAP Neo platform |
| Karma | `node:lts-bookworm` | JavaScript testing |
| Selenium | `selenium/standalone-chrome` | Browser testing |
| Hadolint | (varies) | Dockerfile linting |

## Stash Configuration Defaults

### Before Build Stashing

```yaml
steps:
  pipelineStashFilesBeforeBuild:
    stashIncludes:
      buildDescriptor: >
        **/pom.xml, **/.mvn/**, **/assembly.xml,
        **/package.json, **/requirements.txt,
        **/mta*.y*ml, **/.npmrc, **/Dockerfile,
        **/VERSION, **/Gopkg.*, **/build.sbt

      deployDescriptor: >
        **/manifest*.y*ml, **/*.mtaext.y*ml,
        **/xs-app.json, helm/**

      git: '.git/**'

      opensourceConfiguration: >
        **/srcclr.yml, **/vulas-custom.properties,
        **/.nsprc, **/.retireignore, **/.snyk,
        **/wss-unified-agent.config

      pipelineConfigAndTests: '.pipeline/**'

      securityDescriptor: '**/xs-security.json'

      tests: >
        **/pom.xml, **/*.json, **/*.xml,
        **/src/**, **/node_modules/**,
        **/specs/**, **/tests/**

    stashExcludes:
      buildDescriptor: '**/node_modules/**/package.json'
      git: ''
      opensourceConfiguration: ''

    noDefaultExludes:
      - 'git'
```

### After Build Stashing

```yaml
steps:
  pipelineStashFilesAfterBuild:
    stashIncludes:
      buildResult: >
        **/target/*.war, **/target/*.jar,
        **/*.mtar, **/dist/**

      classFiles: >
        **/target/classes/**/*.class,
        **/target/test-classes/**/*.class

      sonar: >
        **/jacoco*.exec,
        **/sonar-project.properties

    stashExcludes:
      buildResult: ''
      classFiles: ''
      sonar: ''

    noDefaultExludes: []
```

## Notification Defaults

**Email**:
```yaml
steps:
  mailSendNotification:
    notificationAttachment: true
    notifyCulprits: true
    numLogLinesInBody: 100
    wrapInNode: false
```

**Slack**:
```yaml
steps:
  slackSendNotification:
    color: "${['SUCCESS': '#8cc04f', 'FAILURE': '#d54c53', 'ABORTED': '#949393', 'UNSTABLE': '#f6b44b', 'PAUSED': '#24b0d5', 'UNKNOWN': '#d54cc4'].get(buildStatus, '#d54cc4')}"
    defaultMessage: "${buildStatus}: Job ${env.JOB_NAME} <${env.BUILD_URL}|#${env.BUILD_NUMBER}>"
```

## Working with Defaults

### Viewing Default Values

**Method 1: GitHub**
```
https://github.com/SAP/jenkins-library/blob/master/resources/default_pipeline_environment.yml
```

**Method 2: Local Copy**
```bash
# Clone the library
git clone https://github.com/SAP/jenkins-library.git

# View defaults
cat jenkins-library/resources/default_pipeline_environment.yml
```

**Method 3: During Pipeline Execution**
```groovy
// In Jenkinsfile
def defaults = DefaultValueCache.getInstance()?.getDefaultValues()
echo "Defaults: ${defaults}"
```

### Understanding Default Behavior

**Defaults are merged** with your configuration:

```yaml
# Default
steps:
  cloudFoundryDeploy:
    deployType: 'standard'
    keepOldInstance: false
    smokeTestStatusCode: 200

# Your config
steps:
  cloudFoundryDeploy:
    deployType: 'blue-green'

# Result (merged)
cloudFoundryDeploy:
  deployType: 'blue-green'          # Your override
  keepOldInstance: false             # From default
  smokeTestStatusCode: 200          # From default
```

### When to Override Defaults

**Always override**:
- Organization-specific endpoints
- Team-specific credentials
- Project-specific paths

**Consider overriding**:
- Outdated Docker images (e.g., old JDK versions)
- Quality gate thresholds
- Test patterns
- Resource limits

**Usually keep defaults**:
- Docker workspace paths
- Standard port mappings
- Common file patterns
- Tool-specific conventions

## Best Practices

1. **Review Defaults First**: Before adding configuration, check if defaults suit your needs
   ```yaml
   # Instead of duplicating defaults
   steps:
     mavenExecute:
       # Check if you really need to override
   ```

2. **Override Minimally**: Only override what's necessary
   ```yaml
   # Good - minimal override
   steps:
     mavenExecute:
       dockerImage: 'maven:3.8-openjdk-17'

   # Avoid - unnecessarily verbose
   steps:
     mavenExecute:
       dockerImage: 'maven:3.8-openjdk-17'
       logSuccessfulMavenTransfers: false  # This is already the default
       verbose: false                       # This is already the default
   ```

3. **Document Deviations**: Explain why you deviate from defaults
   ```yaml
   steps:
     mavenExecute:
       # Using JDK 17 for Spring Boot 3.0 compatibility
       dockerImage: 'maven:3.8-openjdk-17'
   ```

4. **Use Custom Defaults for Organization Policies**: Put company-wide overrides in custom defaults
   ```yaml
   # In org-defaults.yml
   general:
     collectTelemetryData: false  # Company policy

   steps:
     mavenExecute:
       dockerImage: 'maven:3.8-openjdk-17'  # Standardized version
   ```

5. **Stay Updated**: Review defaults when upgrading Piper
   ```bash
   # Check what changed
   git diff v1.100.0 v1.101.0 resources/default_pipeline_environment.yml
   ```

6. **Test Default Behavior**: Test pipelines with minimal config first
   ```yaml
   # Start with minimal config
   general:
     buildTool: 'maven'

   # Add overrides only as needed
   ```

---

**Next**: [Platform Deviations](03-platform-deviations.md) - Platform-specific configuration differences

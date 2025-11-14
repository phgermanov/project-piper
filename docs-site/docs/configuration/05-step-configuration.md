# Step Configuration

This guide provides detailed examples of configuring individual pipeline steps in Project Piper, covering common scenarios and advanced use cases.

## Table of Contents

- [Overview](#overview)
  - [Build Steps](#build-steps)
    - [Maven Build](#maven-build)
    - [NPM Build](#npm-build)
    - [Gradle Build](#gradle-build)
    - [Docker Build](#docker-build)
    - [Go Build](#go-build)
  - [Version Management](#version-management)
    - [Artifact Prepare Version](#artifact-prepare-version)
    - [Artifact Set Version](#artifact-set-version)
  - [Testing Steps](#testing-steps)
    - [Karma Tests (JavaScript)](#karma-tests-javascript)
    - [Selenium Tests](#selenium-tests)
    - [Gatling Performance Tests](#gatling-performance-tests)
    - [Gauge Tests](#gauge-tests)
  - [Deployment Steps](#deployment-steps)
    - [Cloud Foundry Deploy](#cloud-foundry-deploy)
    - [Kubernetes Deploy](#kubernetes-deploy)
    - [Neo Deploy](#neo-deploy)
  - [Security and Compliance](#security-and-compliance)
    - [Checkmarx Scan](#checkmarx-scan)
    - [WhiteSource Scan](#whitesource-scan)
    - [SonarQube Analysis](#sonarqube-analysis)
    - [Fortify Scan](#fortify-scan)
  - [Container Operations](#container-operations)
    - [Kaniko Build](#kaniko-build)
    - [Container Push](#container-push)
    - [CNB Build](#cnb-build)
  - [GitHub Integration](#github-integration)
    - [GitHub Publish Release](#github-publish-release)
    - [GitHub Create Pull Request](#github-create-pull-request)
  - [Quality and Reporting](#quality-and-reporting)
    - [Checks Publish Results](#checks-publish-results)
    - [Tests Publish Results](#tests-publish-results)
  - [Notification Steps](#notification-steps)
    - [Mail Notifications](#mail-notifications)
    - [Slack Notifications](#slack-notifications)
  - [Advanced Patterns](#advanced-patterns)
    - [Multi-Module Projects](#multi-module-projects)
    - [Monorepo Configuration](#monorepo-configuration)
    - [Matrix Builds](#matrix-builds)
  - [Best Practices](#best-practices)

## Overview

**Steps** are individual executable units in your pipeline. Each step:
- Performs a specific task (build, test, deploy, scan)
- Has its own set of parameters
- Can be configured at multiple levels (general, step, stage)
- Inherits from defaults and custom configuration

**Configuration Syntax**:
```yaml
steps:
  stepName:
    parameter: value
    nestedParameter:
      key: value
```

## Build Steps

### Maven Build

**Basic Maven build configuration**:

```yaml
steps:
  mavenExecute:
    goals: 'clean install'
    defines: '-DskipTests=false'
    dockerImage: 'maven:3.8-openjdk-17'
    pomPath: 'pom.xml'
```

**Advanced Maven configuration**:

```yaml
steps:
  mavenExecute:
    goals: 'clean deploy'

    # Maven options
    defines: >
      -DskipTests=false
      -Dmaven.javadoc.skip=true
      -B

    # Maven settings
    projectSettingsFile: 'settings.xml'
    globalSettingsFile: '.mvn/global-settings.xml'

    # Build profiles
    profiles:
      - 'production'
      - 'optimize'

    # Docker execution
    dockerImage: 'maven:3.8-openjdk-17'
    dockerOptions:
      - '--memory=4g'
      - '--cpus=2'

    # Logging
    logSuccessfulMavenTransfers: false
    verbose: true
```

**Maven with custom repository**:

```yaml
steps:
  mavenExecute:
    goals: 'clean install'
    defines: >
      -DaltDeploymentRepository=nexus::default::https://nexus.company.com/repository/maven-releases/
```

**Multi-module Maven project**:

```yaml
steps:
  mavenExecute:
    goals: 'clean install'
    pomPath: 'pom.xml'  # Parent POM

  mavenExecuteStaticCodeChecks: true

  mavenExecuteIntegration:
    goals: 'verify'
    defines: '-DskipIntegrationTests=false'
```

### NPM Build

**Basic NPM configuration**:

```yaml
steps:
  npmExecute:
    runScripts:
      - 'build'
      - 'test'
    dockerImage: 'node:18'
```

**Advanced NPM configuration**:

```yaml
steps:
  npmExecuteScripts:
    install: true  # Run npm install first

    # Scripts to execute
    runScripts:
      - 'lint'
      - 'test:unit'
      - 'test:coverage'
      - 'build'

    # NPM options
    defaultNpmRegistry: 'https://registry.npmjs.org'
    sapNpmRegistry: 'https://npm.sap.com'

    # Docker execution
    dockerImage: 'node:18-alpine'
    dockerWorkspace: '/home/node'
    dockerOptions:
      - '--memory=2g'

    # Environment variables
    dockerEnvVars:
      NODE_ENV: 'production'
      NPM_CONFIG_LOGLEVEL: 'warn'
```

**NPM with private registry**:

```yaml
steps:
  npmExecuteScripts:
    install: true
    defaultNpmRegistry: 'https://npm.company.com'

    # Provide auth via Vault or credentials
    npmRegistryCredentialsId: 'NPM_REGISTRY_TOKEN'

    runScripts:
      - 'build'
```

**Workspace configuration**:

```yaml
steps:
  npmExecuteScripts:
    # For NPM workspaces/monorepo
    install: true
    runScripts:
      - 'build'

    # Run in specific workspace
    scriptOptions:
      - '--workspace=packages/app'
```

### Gradle Build

**Basic Gradle configuration**:

```yaml
steps:
  gradleExecute:
    task: 'build'
    dockerImage: 'gradle:7-jdk17'
```

**Advanced Gradle configuration**:

```yaml
steps:
  gradleExecute:
    task: 'build test jacocoTestReport'

    # Gradle options
    gradleOptions: >
      --no-daemon
      --stacktrace
      --build-cache

    # Build file
    buildFile: 'build.gradle.kts'

    # Docker execution
    dockerImage: 'gradle:7-jdk17'
    dockerOptions:
      - '--memory=4g'

    # Exclude tests
    excludeTests: false
```

### Docker Build

**Basic Docker build**:

```yaml
steps:
  dockerExecute:
    dockerImage: 'docker:20'
    containerCommand: 'docker build -t my-app:${version} .'
```

**Multi-stage Docker build**:

```yaml
steps:
  dockerExecute:
    dockerImage: 'docker:20'
    dockerOptions:
      - '--privileged'

    containerCommand: >
      docker build
      --target production
      --build-arg VERSION=${version}
      --tag my-app:${version}
      --tag my-app:latest
      .
```

**Build and push**:

```yaml
steps:
  dockerExecute:
    dockerImage: 'docker:20'
    dockerOptions:
      - '--privileged'

    # Build
    containerCommand: 'docker build -t registry.company.com/my-app:${version} .'

  # Push to registry
  containerPushToRegistry:
    dockerRegistryUrl: 'https://registry.company.com'
    dockerCredentialsId: 'DOCKER_REGISTRY'
    dockerImageName: 'my-app'
    dockerImageTags:
      - '${version}'
      - 'latest'
```

### Go Build

**Basic Go build**:

```yaml
general:
  buildTool: 'golang'

steps:
  buildExecute:
    buildOptions: 'build -v ./...'
```

**Advanced Go configuration**:

```yaml
steps:
  buildExecute:
    buildTool: 'golang'

    # Build options
    buildOptions: >
      build
      -v
      -ldflags="-X main.version=${version} -X main.commit=${commitId}"
      -o bin/app
      ./cmd/app

    # Docker execution
    dockerImage: 'golang:1.21'
    dockerOptions:
      - '--memory=2g'

    # Environment
    dockerEnvVars:
      CGO_ENABLED: '0'
      GOOS: 'linux'
      GOARCH: 'amd64'
```

**Go with testing**:

```yaml
steps:
  buildExecute:
    buildOptions: 'build -v ./...'

  # Run tests
  executeTest:
    testCommand: 'go test -v -cover ./...'
    testReportFilePath: 'coverage.out'
```

## Version Management

### Artifact Prepare Version

**Automatic version from Git**:

```yaml
steps:
  artifactPrepareVersion:
    versioningType: 'cloud'  # Options: library, cloud, cloud_noTag

    # Git configuration
    gitSshKeyCredentialsId: 'github-ssh-key'
    gitHttpsCredentialsId: 'github-token'

    # Tag prefix
    tagPrefix: 'v'

    # Timestamp format
    timestampTemplate: '%Y%m%d%H%M%S'
```

**Version with build number**:

```yaml
steps:
  artifactPrepareVersion:
    versioningType: 'cloud'
    includeCommitId: true

    # Results in: 1.0.0-20231115120000+abc123
    versioningTemplate: '${version}-${timestamp}+${commitId}'
```

### Artifact Set Version

**Update version in files**:

```yaml
steps:
  artifactSetVersion:
    buildTool: 'maven'  # Detected automatically

    # Git commit
    commitVersion: true
    gitSshKeyCredentialsId: 'github-ssh-key'

    # Tag creation
    tagPrefix: 'build_'

    # Build tool-specific
    maven:
      filePath: 'pom.xml'
      versioningTemplate: '${version}'

    npm:
      filePath: 'package.json'
      versioningTemplate: '${version}'
```

**Custom versioning template**:

```yaml
steps:
  artifactSetVersion:
    commitVersion: true

    maven:
      versioningTemplate: '${version}-${timestamp}'

    npm:
      versioningTemplate: '${version}+${commitId}'
```

## Testing Steps

### Karma Tests (JavaScript)

**Basic Karma configuration**:

```yaml
steps:
  karmaExecuteTests:
    modules:
      - '.'
    installCommand: 'npm install'
    runCommand: 'npm run karma'
```

**Advanced Karma configuration**:

```yaml
steps:
  karmaExecuteTests:
    # Modules to test
    modules:
      - 'packages/app'
      - 'packages/lib'

    # Installation
    installCommand: 'npm ci --prefer-offline'

    # Execution
    runCommand: 'npm run karma'

    # Docker configuration
    dockerImage: 'node:18'
    dockerWorkspace: '/home/node'

    # Port mappings
    containerPortMappings:
      'node:18':
        - containerPort: 9876
          hostPort: 9876

    # Environment variables
    dockerEnvVars:
      CHROME_BIN: '/usr/bin/chromium-browser'
      NO_PROXY: 'localhost,selenium'

    # Stash results
    stashContent:
      - 'buildDescriptor'
      - 'tests'
```

### Selenium Tests

**Basic Selenium configuration**:

```yaml
steps:
  seleniumExecuteTests:
    buildTool: 'npm'
    failOnError: true
```

**Advanced Selenium configuration**:

```yaml
steps:
  seleniumExecuteTests:
    buildTool: 'maven'

    # Selenium configuration
    sidecarImage: 'selenium/standalone-chrome:latest'
    sidecarName: 'selenium'

    # Volume for shared memory
    sidecarVolumeBind:
      '/dev/shm': '/dev/shm'

    # Port mapping
    containerPortMappings:
      'selenium/standalone-chrome':
        - containerPort: 4444
          hostPort: 4444

    # Test execution
    maven:
      dockerImage: 'maven:3.8-openjdk-17'
      dockerName: 'maven'
      testRepository: 'https://github.com/company/e2e-tests.git'

    # Environment
    dockerEnvVars:
      SELENIUM_HOST: 'selenium'
      SELENIUM_PORT: '4444'

    # Error handling
    failOnError: true
    stashContent:
      - 'tests'
```

### Gatling Performance Tests

**Basic Gatling configuration**:

```yaml
steps:
  gatlingExecuteTests:
    pomPath: 'pom.xml'
```

**Advanced Gatling configuration**:

```yaml
steps:
  gatlingExecuteTests:
    # Maven configuration
    pomPath: 'performance-tests/pom.xml'
    goals: 'gatling:test'

    # Docker execution
    dockerImage: 'maven:3.8-openjdk-17'
    dockerOptions:
      - '--memory=4g'

    # Environment
    dockerEnvVars:
      JAVA_OPTS: '-Xmx3g'
      TARGET_URL: 'https://staging.company.com'
      USERS: '100'
      DURATION: '300'

    # Results
    stashContent:
      - 'source'
```

### Gauge Tests

**NPM-based Gauge tests**:

```yaml
steps:
  gaugeExecuteTests:
    buildTool: 'npm'

    # Docker configuration
    npm:
      dockerImage: 'node:18'
      dockerName: 'npm'
      dockerWorkspace: '/home/node'
      languageRunner: 'js'
      runCommand: 'gauge run'
      testOptions: 'specs'

    # Installation
    installCommand: 'npm install'

    # Test content
    stashContent:
      - 'buildDescriptor'
      - 'tests'

    # Error handling
    failOnError: false
```

## Deployment Steps

### Cloud Foundry Deploy

**Basic CF deployment**:

```yaml
steps:
  cloudFoundryDeploy:
    deployTool: 'cf_native'
    cloudFoundry:
      org: 'my-org'
      space: 'production'
      apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'
      credentialsId: 'CF_CREDENTIALS'
    manifest: 'manifest.yml'
```

**Blue-green deployment**:

```yaml
steps:
  cloudFoundryDeploy:
    deployTool: 'cf_native'
    deployType: 'blue-green'

    cloudFoundry:
      org: 'my-org'
      space: 'production'
      apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'
      credentialsId: 'CF_PROD_CREDENTIALS'

    # Manifest
    manifest: 'manifest-prod.yml'

    # Keep old instance
    keepOldInstance: true

    # Smoke test
    smokeTest: true
    smokeTestScript: 'smoke-test.sh'
    smokeTestStatusCode: 200

    # Docker image
    cf_native:
      dockerImage: 'ppiper/cf-cli:latest'
      dockerWorkspace: '/home/piper'
```

**MTA deployment**:

```yaml
steps:
  cloudFoundryDeploy:
    deployTool: 'mtaDeployPlugin'

    cloudFoundry:
      org: 'my-org'
      space: 'production'
      apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'
      credentialsId: 'CF_CREDENTIALS'

    # MTA configuration
    mtaPath: 'mta_archives/my-app_1.0.0.mtar'
    mtaDeployParameters: '-f --version-rule ALL'
    mtaExtensionDescriptor: 'extension-prod.mtaext'

    # Docker image
    mtaDeployPlugin:
      dockerImage: 'ppiper/cf-cli:latest'
```

**Multi-target deployment**:

```yaml
steps:
  multicloudDeploy:
    cfTargets:
      - org: 'my-org'
        space: 'eu-production'
        apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'
        credentialsId: 'CF_EU_CREDENTIALS'
        manifest: 'manifest-eu.yml'

      - org: 'my-org'
        space: 'us-production'
        apiEndpoint: 'https://api.cf.us10.hana.ondemand.com'
        credentialsId: 'CF_US_CREDENTIALS'
        manifest: 'manifest-us.yml'

    # Deployment options
    enableZeroDowntimeDeployment: true
    parallelExecution: false
```

### Kubernetes Deploy

**Basic Kubernetes deployment**:

```yaml
steps:
  kubernetesDeploy:
    kubeConfig: 'kubeconfig'
    namespace: 'production'
    deploymentFiles:
      - 'k8s/deployment.yaml'
      - 'k8s/service.yaml'
```

**Helm deployment**:

```yaml
steps:
  helmExecute:
    helmCommand: 'upgrade'

    # Helm options
    chartPath: 'helm/my-app'
    releaseName: 'my-app'
    namespace: 'production'

    # Values
    values:
      - 'values-prod.yaml'

    # Set values
    helmSetValues:
      - 'image.tag=${version}'
      - 'replicas=3'

    # Kubernetes config
    kubeConfig: 'kubeconfig'
    kubeContext: 'prod-cluster'
```

### Neo Deploy

**SAP Neo deployment**:

```yaml
steps:
  neoDeploy:
    deployMode: 'mta'

    # Neo configuration
    neo:
      host: 'hana.ondemand.com'
      account: 'my-account'
      credentialsId: 'NEO_CREDENTIALS'
      size: 'lite'

    # MTA path
    source: 'mta_archives/my-app.mtar'

    # Docker image
    dockerImage: 'ppiper/neo-cli'
```

## Security and Compliance

### Checkmarx Scan

**Basic Checkmarx scan**:

```yaml
steps:
  checkmarxExecuteScan:
    serverUrl: 'https://checkmarx.company.com'
    projectName: 'my-app'
    teamName: '/CxServer/Team'
    credentialsId: 'CHECKMARX_CREDENTIALS'
```

**Advanced Checkmarx configuration**:

```yaml
steps:
  checkmarxExecuteScan:
    # Server configuration
    serverUrl: 'https://checkmarx.company.com'
    credentialsId: 'CHECKMARX_CREDENTIALS'

    # Project configuration
    projectName: 'my-app-${env.BRANCH_NAME}'
    teamName: '/CxServer/Platform/Team'
    preset: 'SAP_JS_Default'

    # Scan configuration
    incremental: true
    fullScansScheduled: true
    fullScanCycle: 10

    # Source configuration
    sourceEncoding: 'UTF-8'
    filterPattern: >
      !**/*.spec.js,
      !**/node_modules/**,
      !**/test/**

    # Vulnerability thresholds
    vulnerabilityThresholdHigh: 0
    vulnerabilityThresholdMedium: 10
    vulnerabilityThresholdLow: 100

    # Error handling
    vulnerabilityThresholdEnabled: true
    failOnError: true
```

### WhiteSource Scan

**Basic WhiteSource scan**:

```yaml
steps:
  whitesourceExecuteScan:
    productName: 'my-product'
    projectName: 'my-app'
    serviceUrl: 'https://saas.whitesourcesoftware.com/api'
    orgToken: '$(whitesourceOrgToken)'
```

**Advanced WhiteSource configuration**:

```yaml
steps:
  whitesourceExecuteScan:
    # Organization/Product
    serviceUrl: 'https://saas.whitesourcesoftware.com/api'
    orgToken: '$(whitesourceOrgToken)'
    productName: 'my-product'
    projectName: 'my-app'

    # Scan configuration
    buildDescriptorFile: 'pom.xml'
    scanType: 'maven'

    # Reporting
    reporting: true
    vulnerabilityReportFileName: 'whitesource-vulnerability-report'

    # Vulnerability assessment
    cvssSeverityLimit: 7
    failOnSevereVulnerabilities: true

    # Exclusions
    excludes:
      - '**/test/**'
      - '**/node_modules/**'
```

### SonarQube Analysis

**Basic SonarQube scan**:

```yaml
steps:
  sonarExecuteScan:
    serverUrl: 'https://sonarqube.company.com'
    projectKey: 'my-app'
    sonarTokenCredentialsId: 'SONAR_TOKEN'
```

**Advanced SonarQube configuration**:

```yaml
steps:
  sonarExecuteScan:
    # Server configuration
    serverUrl: 'https://sonarqube.company.com'
    sonarTokenCredentialsId: 'SONAR_TOKEN'

    # Project configuration
    projectKey: 'com.company:my-app'
    projectName: 'My Application'
    projectVersion: '${version}'

    # Source configuration
    sources: 'src/main'
    tests: 'src/test'
    sourceEncoding: 'UTF-8'

    # Coverage
    coverageReportPaths: 'target/site/jacoco/jacoco.xml'

    # Quality gates
    qualityGates:
      - metric: 'coverage'
        threshold: 80
        operator: 'GREATER_THAN'

      - metric: 'security_rating'
        threshold: 'A'
        operator: 'EQUALS'

      - metric: 'bugs'
        threshold: 0
        operator: 'EQUALS'

    # Options
    options:
      - '-Dsonar.exclusions=**/*Test.java'
      - '-Dsonar.issue.ignore.multicriteria=e1'
```

### Fortify Scan

**Basic Fortify scan**:

```yaml
steps:
  fortifyExecuteScan:
    serverUrl: 'https://fortify.company.com'
    projectName: 'my-app'
    credentialsId: 'FORTIFY_CREDENTIALS'
```

## Container Operations

### Kaniko Build

**Basic Kaniko build**:

```yaml
steps:
  kanikoExecute:
    containerImageName: 'my-app'
    containerImageTag: '${version}'
    containerRegistry: 'docker.io/company'
```

**Advanced Kaniko configuration**:

```yaml
steps:
  kanikoExecute:
    # Image configuration
    containerImageName: 'my-app'
    containerImageTags:
      - '${version}'
      - 'latest'
      - '${env.BRANCH_NAME}'

    # Registry
    containerRegistry: 'registry.company.com'
    containerRegistryCredentialsId: 'DOCKER_REGISTRY'

    # Build configuration
    dockerfilePath: 'Dockerfile'
    dockerConfigJsonCredentialsId: 'DOCKER_CONFIG'

    # Build args
    buildArgs:
      - 'VERSION=${version}'
      - 'BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")'

    # Multi-stage build
    target: 'production'

    # Cache
    containerBuildOptions:
      - '--cache=true'
      - '--cache-ttl=24h'
```

### Container Push

**Push to registry**:

```yaml
steps:
  containerPushToRegistry:
    dockerRegistryUrl: 'https://registry.company.com'
    dockerCredentialsId: 'DOCKER_REGISTRY'
    dockerImageName: 'my-app'
    dockerImageTags:
      - '${version}'
      - 'latest'
      - 'stable'
```

### CNB Build

**Cloud Native Buildpacks**:

```yaml
steps:
  cnbBuild:
    # Container configuration
    containerImageName: 'my-app'
    containerImageTag: '${version}'
    containerRegistry: 'registry.company.com'

    # Builder
    dockerImage: 'paketobuildpacks/builder:full'

    # Buildpacks
    buildpacks:
      - 'gcr.io/paketo-buildpacks/java'

    # Environment
    buildEnvVars:
      BP_JVM_VERSION: '17'
      BP_MAVEN_BUILT_ARTIFACT: 'target/*.jar'
```

## GitHub Integration

### GitHub Publish Release

**Basic release creation**:

```yaml
steps:
  githubPublishRelease:
    owner: 'company'
    repository: 'my-app'
    token: '$(githubToken)'
    addClosedIssues: true
```

**Advanced release configuration**:

```yaml
steps:
  githubPublishRelease:
    # Repository
    githubApiUrl: 'https://api.github.com'
    owner: 'company'
    repository: 'my-app'
    token: '$(githubToken)'

    # Release configuration
    version: '${version}'
    commitish: '${env.GIT_COMMIT}'
    releaseBodyHeader: '## Release ${version}'

    # Issues
    addClosedIssues: true
    addDeltaToLastRelease: true

    # Labels to exclude
    excludeLabels:
      - 'duplicate'
      - 'invalid'
      - 'wontfix'
      - 'question'

    # Assets
    assetPath: 'target/*.jar'
```

### GitHub Create Pull Request

**Create PR from pipeline**:

```yaml
steps:
  githubCreatePullRequest:
    owner: 'company'
    repository: 'my-app'
    token: '$(githubToken)'

    # PR configuration
    title: 'Automated update: ${version}'
    body: 'This PR was automatically created by the pipeline'
    head: 'feature/update-${version}'
    base: 'main'

    # Labels and assignees
    labels:
      - 'automated'
      - 'dependencies'

    assignees:
      - 'team-lead'

    reviewers:
      - 'developer1'
      - 'developer2'
```

## Quality and Reporting

### Checks Publish Results

**Publish quality checks**:

```yaml
steps:
  checksPublishResults:
    failOnError: false

    # PMD
    pmd:
      pattern: '**/target/pmd.xml'
      archive: true
      active: true
      qualityGates:
        - threshold: 1
          type: 'TOTAL_HIGH'
          unstable: false

    # Checkstyle
    checkstyle:
      pattern: '**/target/checkstyle-result.xml'
      archive: true
      active: true

    # ESLint
    eslint:
      pattern: '**/eslint.xml'
      archive: true
      active: true
```

### Tests Publish Results

**Publish test results**:

```yaml
steps:
  testsPublishResults:
    failOnError: false

    # JUnit
    junit:
      pattern: '**/TEST-*.xml'
      updateResults: false
      allowEmptyResults: true
      archive: true
      active: true

    # JaCoCo coverage
    jacoco:
      pattern: '**/target/*.exec'
      allowEmptyResults: true
      archive: true
      active: true

    # Cobertura coverage
    cobertura:
      pattern: '**/target/coverage/**/cobertura-coverage.xml'
      onlyStableBuilds: true
      archive: true
      active: true
```

## Notification Steps

### Mail Notifications

**Email notifications**:

```yaml
steps:
  mailSendNotification:
    # Recipients
    recipients: 'team@company.com'
    notifyCulprits: true

    # Content
    subject: 'Pipeline ${currentBuild.result}: ${env.JOB_NAME}'
    body: 'Build ${env.BUILD_NUMBER} ${currentBuild.result}'

    # Attachments
    notificationAttachment: true
    numLogLinesInBody: 100

    # Conditions
    notifyOnSuccess: false
    notifyOnFailure: true
    notifyOnUnstable: true
```

### Slack Notifications

**Slack notifications**:

```yaml
steps:
  slackSendNotification:
    # Channel
    channel: '#deployments'
    webhookUrl: '$(slackWebhook)'

    # Message
    defaultMessage: '${buildStatus}: Job ${env.JOB_NAME} <${env.BUILD_URL}|#${env.BUILD_NUMBER}>'

    # Color based on status
    color: "${['SUCCESS': '#8cc04f', 'FAILURE': '#d54c53'].get(buildStatus, '#949393')}"

    # Conditions
    notifyOnSuccess: true
    notifyOnFailure: true
```

## Advanced Patterns

### Multi-Module Projects

**Maven multi-module**:

```yaml
steps:
  mavenExecute:
    pomPath: 'pom.xml'
    goals: 'clean install'

    # Build reactor
    defines: >
      -DskipTests=false
      -pl module1,module2
      -am

  # Test specific modules
  mavenExecuteIntegration:
    pomPath: 'integration-tests/pom.xml'
    goals: 'verify'
```

### Monorepo Configuration

**NPM workspaces**:

```yaml
steps:
  npmExecuteScripts:
    install: true

    # Build all workspaces
    runScripts:
      - 'build'

    scriptOptions:
      - '--workspaces'

  # Test specific workspace
  npmExecuteEndToEndTests:
    runScript: 'test:e2e'
    scriptOptions:
      - '--workspace=packages/frontend'
```

### Matrix Builds

**Test multiple versions**:

```yaml
# Via stage extensions
stages:
  Build:
    # Define matrix in extension
    # .pipeline/extensions/Build.groovy
```

**.pipeline/extensions/Build.groovy**:
```groovy
void call(Map parameters) {
    def nodeVersions = ['14', '16', '18']

    nodeVersions.each { version ->
        stage("Build Node ${version}") {
            npmExecute(
                script: this,
                dockerImage: "node:${version}",
                runScripts: ['test', 'build']
            )
        }
    }
}
```

## Best Practices

1. **Start with Defaults**: Only override what's necessary

2. **Use Descriptive Names**: Clear parameter names and values
   ```yaml
   steps:
     cloudFoundryDeploy:
       # Clear and descriptive
       deployType: 'blue-green'
       keepOldInstance: true
   ```

3. **Document Complex Configuration**: Add comments
   ```yaml
   steps:
     checkmarxExecuteScan:
       # Zero tolerance for high vulnerabilities in production
       vulnerabilityThresholdHigh: 0
   ```

4. **Group Related Configuration**: Keep related settings together
   ```yaml
   steps:
     cloudFoundryDeploy:
       cloudFoundry:
         org: 'my-org'
         space: 'production'
         apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'
   ```

5. **Use Variables**: Reference shared values
   ```yaml
   general:
     dockerImage: 'maven:3.8-openjdk-17'

   steps:
     mavenExecute:
       # Inherits from general
   ```

6. **Test Configuration**: Validate in non-production first

7. **Keep Secrets Secure**: Use Vault or credential stores
   ```yaml
   steps:
     cloudFoundryDeploy:
       cloudFoundryPasswordVaultSecretName: 'cf-password'
   ```

8. **Enable Verbose Mode**: For debugging
   ```yaml
   steps:
     mavenExecute:
       verbose: true
   ```

---

**Next**: [Credentials Management](06-credentials-management.md) - Vault and secrets handling

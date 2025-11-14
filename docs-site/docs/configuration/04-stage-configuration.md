# Stage Configuration

Pipeline stages in Project Piper represent major phases of your CI/CD workflow. This guide covers how to configure and customize stages for your specific needs.

## Table of Contents

- [Stage Configuration](#stage-configuration)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Available Stages](#available-stages)
  - [Stage Configuration Basics](#stage-configuration-basics)
    - [Basic Structure](#basic-structure)
    - [Stage Activation](#stage-activation)
    - [Stage-Specific Parameters](#stage-specific-parameters)
  - [Stage Descriptions](#stage-descriptions)
    - [Init Stage](#init-stage)
    - [Pull-Request Voting Stage](#pull-request-voting-stage)
    - [Build Stage](#build-stage)
    - [Additional Unit Tests Stage](#additional-unit-tests-stage)
    - [Integration Stage](#integration-stage)
    - [Acceptance Stage](#acceptance-stage)
    - [Security Stage](#security-stage)
    - [Performance Stage](#performance-stage)
    - [Compliance Stage](#compliance-stage)
    - [Confirm Stage](#confirm-stage)
    - [Promote Stage](#promote-stage)
    - [Release Stage](#release-stage)
  - [Configuration Examples](#configuration-examples)
    - [Example 1: Simple Build and Deploy](#example-1-simple-build-and-deploy)
    - [Example 2: Full Pipeline with Testing](#example-2-full-pipeline-with-testing)
    - [Example 3: Multi-Environment Deployment](#example-3-multi-environment-deployment)
    - [Example 4: Security and Compliance Focus](#example-4-security-and-compliance-focus)
  - [Advanced Stage Configuration](#advanced-stage-configuration)
    - [Conditional Stage Execution](#conditional-stage-execution)
    - [Stage Extensions](#stage-extensions)
    - [Stage Locking](#stage-locking)
    - [Custom Stage Parameters](#custom-stage-parameters)
  - [Stage Sequence Control](#stage-sequence-control)
  - [Common Patterns](#common-patterns)
    - [Pattern 1: Feature Branch Pipeline](#pattern-1-feature-branch-pipeline)
    - [Pattern 2: Main Branch Pipeline](#pattern-2-main-branch-pipeline)
    - [Pattern 3: Release Pipeline](#pattern-3-release-pipeline)
  - [Troubleshooting](#troubleshooting)
    - [Stage Not Executing](#stage-not-executing)
    - [Stage Failing Unexpectedly](#stage-failing-unexpectedly)
    - [Configuration Not Taking Effect](#configuration-not-taking-effect)
  - [Best Practices](#best-practices)

## Overview

**Stages** are high-level phases in your pipeline that:
- Group related steps together
- Provide logical separation of concerns
- Enable environment-specific configuration
- Support conditional execution
- Allow manual approval gates

**Stage vs Step**:
- **Stage**: High-level phase (e.g., "Build", "Deploy")
- **Step**: Individual action (e.g., "mavenExecute", "cloudFoundryDeploy")

**Configuration Priority**: Stage configuration overrides step and general configuration for steps executed within that stage.

## Available Stages

Project Piper provides these standard stages:

| Stage | Purpose | When Executed |
|-------|---------|---------------|
| Init | Pipeline initialization | Always |
| Pull-Request Voting | PR validation | On pull requests |
| Build | Artifact creation | Always (main branches) |
| Additional Unit Tests | Extra test execution | If configured |
| Integration | Integration testing | If configured |
| Acceptance | End-to-end testing | If configured |
| Security | Security scans | If configured |
| Performance | Performance testing | If configured |
| Compliance | Code quality checks | If configured |
| Confirm | Manual approval | If configured |
| Promote | Artifact promotion | On productive branch |
| Release | Production deployment | On productive branch |

## Stage Configuration Basics

### Basic Structure

**Configuration syntax**:
```yaml
stages:
  StageName:
    stepName:
      parameter: value

    anotherStep:
      parameter: value
```

**Example**:
```yaml
stages:
  Build:
    mavenExecute:
      goals: 'clean verify'

  Release:
    cloudFoundryDeploy:
      cloudFoundry:
        space: 'production'
```

### Stage Activation

Stages are activated based on:

1. **Configuration presence**: Stage with configuration is activated
2. **File patterns**: Detected files trigger certain stages
3. **Branch context**: Some stages only run on specific branches
4. **Explicit activation**: Direct stage execution requests

**Automatic activation example**:
```yaml
# Build stage activates automatically if buildTool is set
general:
  buildTool: 'maven'

# Acceptance stage activates with this config
stages:
  Acceptance:
    cloudFoundryDeploy:
      cloudFoundry:
        space: 'acceptance'
```

### Stage-Specific Parameters

Override step configuration for specific stages:

```yaml
steps:
  # Default for all stages
  cloudFoundryDeploy:
    deployType: 'standard'
    cloudFoundry:
      org: 'my-org'

stages:
  # Override for Acceptance stage
  Acceptance:
    cloudFoundryDeploy:
      cloudFoundry:
        space: 'acceptance'
      smokeTestScript: 'acceptance-tests.sh'

  # Override for Release stage
  Release:
    cloudFoundryDeploy:
      deployType: 'blue-green'
      cloudFoundry:
        space: 'production'
      smokeTestScript: 'production-tests.sh'
```

## Stage Descriptions

### Init Stage

**Purpose**: Initialize pipeline environment

**Activities**:
- Checkout repository
- Load configuration
- Determine active stages
- Perform version management
- Setup common pipeline environment

**Configuration**:
```yaml
stages:
  Init:
    # Usually no configuration needed
```

**Related Steps**:
- `setupCommonPipelineEnvironment`
- `artifactPrepareVersion`
- `checkoutGit`

### Pull-Request Voting Stage

**Purpose**: Validate pull requests before merge

**Activities**:
- Build verification
- Unit test execution
- Static code analysis
- PR checks

**Configuration**:
```yaml
general:
  buildTool: 'npm'  # Activates PR voting

stages:
  'Pull-Request Voting':
    npmExecuteScripts:
      runScripts:
        - 'lint'
        - 'test'
        - 'build'
```

**Activation**: Automatically on pull requests

### Build Stage

**Purpose**: Create deployable artifacts

**Activities**:
- Execute build tool (Maven, NPM, etc.)
- Run unit tests
- Create artifacts
- Publish test results
- Execute static checks

**Configuration**:
```yaml
stages:
  Build:
    mavenExecute:
      goals: 'clean install'
      defines: '-DskipTests=false'

    mavenExecuteStaticCodeChecks: true

    mavenExecuteIntegration: true
```

**Common Parameters**:
```yaml
stages:
  Build:
    # Enable/disable integration tests
    mavenExecuteIntegration: false

    # Static code checks
    mavenExecuteStaticCodeChecks: true

    # NPM-specific
    npmExecuteLint: true
```

### Additional Unit Tests Stage

**Purpose**: Run additional test frameworks

**Activities**:
- Karma tests (JavaScript)
- OPA5 tests (SAPUI5)
- Additional test suites

**Configuration**:
```yaml
stages:
  'Additional Unit Tests':
    karmaExecuteTests:
      modules:
        - '.'
      installCommand: 'npm install'
      runCommand: 'npm run karma'
```

### Integration Stage

**Purpose**: Execute integration tests

**Activities**:
- Service integration tests
- API tests
- Database integration tests

**Configuration**:
```yaml
stages:
  Integration:
    npmExecuteScripts:
      runScripts:
        - 'integration-test'

    # Or Maven integration tests
    mavenExecute:
      goals: 'integration-test'
      defines: '-DskipIntegrationTests=false'
```

### Acceptance Stage

**Purpose**: End-to-end testing in staging environment

**Activities**:
- Deploy to acceptance environment
- Run E2E tests
- Execute smoke tests
- Validate functionality

**Configuration**:
```yaml
stages:
  Acceptance:
    cloudFoundryDeploy:
      cloudFoundry:
        org: 'my-org'
        space: 'acceptance'
      smokeTest: true
      smokeTestScript: 'acceptance-smoke-test.sh'

    # Optional: Run additional tests after deployment
    healthExecuteCheck:
      healthEndpoint: 'https://app-acceptance.company.com/health'

    uiVeri5ExecuteTests:
      testServerUrl: 'https://app-acceptance.company.com'
```

### Security Stage

**Purpose**: Execute security scans

**Activities**:
- SAST scanning (Checkmarx, Fortify)
- SCA scanning (WhiteSource, Black Duck)
- Container scanning
- Vulnerability assessment

**Configuration**:
```yaml
stages:
  Security:
    checkmarxExecuteScan:
      projectName: 'my-project'
      teamName: 'my-team'
      serverUrl: 'https://checkmarx.company.com'
      credentialsId: 'CHECKMARX_CREDENTIALS'
      vulnerabilityThresholdMedium: 10
      vulnerabilityThresholdHigh: 0

    whitesourceExecuteScan:
      productName: 'my-product'
      vulnerabilityReportFileName: 'whitesource-report'
```

### Performance Stage

**Purpose**: Execute performance tests

**Activities**:
- Load testing
- Stress testing
- Performance benchmarking

**Configuration**:
```yaml
stages:
  Performance:
    gatlingExecuteTests:
      pomPath: 'performance-tests/pom.xml'
```

### Compliance Stage

**Purpose**: Code quality and compliance checks

**Activities**:
- SonarQube analysis
- Code coverage checks
- License compliance
- Quality gate validation

**Configuration**:
```yaml
stages:
  Compliance:
    sonarExecuteScan:
      serverUrl: 'https://sonarqube.company.com'
      projectKey: 'my-project'
      sonarTokenCredentialsId: 'SONAR_TOKEN'

      # Quality gates
      qualityGates:
        - metric: 'coverage'
          threshold: 80
          operator: 'GREATER_THAN'
```

### Confirm Stage

**Purpose**: Manual approval gate

**Activities**:
- Pause pipeline execution
- Request manual approval
- Allow review before production

**Configuration**:
```yaml
general:
  manualConfirmation: true
  manualConfirmationMessage: 'Proceed to production deployment?'
  manualConfirmationTimeout: 720  # 12 hours in minutes

stages:
  Confirm:
    # Usually no additional configuration
```

### Promote Stage

**Purpose**: Promote artifacts to production registry

**Activities**:
- Push to production artifact repository
- Tag containers for production
- Update artifact metadata

**Configuration**:
```yaml
stages:
  Promote:
    containerPushToRegistry:
      dockerRegistryUrl: 'https://registry.company.com'
      dockerCredentialsId: 'DOCKER_REGISTRY_PROD'
      dockerImageTags:
        - '${version}'
        - 'latest'
```

### Release Stage

**Purpose**: Deploy to production

**Activities**:
- Production deployment
- Smoke tests in production
- Post-deployment verification
- Release notifications

**Configuration**:
```yaml
stages:
  Release:
    cloudFoundryDeploy:
      deployType: 'blue-green'
      cloudFoundry:
        org: 'my-org'
        space: 'production'
        credentialsId: 'CF_PROD_CREDENTIALS'
      smokeTest: true
      smokeTestScript: 'production-smoke-test.sh'
      keepOldInstance: true

    healthExecuteCheck:
      healthEndpoint: 'https://api.company.com/health'

    # Optional: Create GitHub release
    githubPublishRelease:
      addClosedIssues: true
      addDeltaToLastRelease: true
```

## Configuration Examples

### Example 1: Simple Build and Deploy

**Minimal configuration for basic CI/CD**:

```yaml
general:
  buildTool: 'maven'
  productiveBranch: 'main'
  gitSshKeyCredentialsId: 'github-ssh-key'

steps:
  cloudFoundryDeploy:
    cloudFoundry:
      org: 'my-org'
      credentialsId: 'CF_CREDENTIALS'
      apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'

stages:
  Build:
    mavenExecute:
      goals: 'clean package'

  Release:
    cloudFoundryDeploy:
      cloudFoundry:
        space: 'production'
```

### Example 2: Full Pipeline with Testing

**Comprehensive testing pipeline**:

```yaml
general:
  buildTool: 'npm'
  productiveBranch: 'main'

stages:
  Build:
    npmExecuteScripts:
      install: true
      runScripts:
        - 'lint'
        - 'build'
        - 'test'

  'Additional Unit Tests':
    karmaExecuteTests:
      modules:
        - '.'
      installCommand: 'npm install'
      runCommand: 'npm run karma'

  Integration:
    npmExecuteScripts:
      runScripts:
        - 'test:integration'

  Acceptance:
    cloudFoundryDeploy:
      cloudFoundry:
        space: 'acceptance'
      smokeTest: true

    uiVeri5ExecuteTests:
      testServerUrl: 'https://app-acceptance.company.com'

  Security:
    whitesourceExecuteScan:
      productName: 'my-app'

  Compliance:
    sonarExecuteScan:
      projectKey: 'my-app'

  Release:
    cloudFoundryDeploy:
      cloudFoundry:
        space: 'production'
      deployType: 'blue-green'
```

### Example 3: Multi-Environment Deployment

**Progressive deployment through environments**:

```yaml
general:
  buildTool: 'maven'

steps:
  cloudFoundryDeploy:
    cloudFoundry:
      org: 'my-org'
      apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'
      credentialsId: 'CF_CREDENTIALS'

stages:
  Build:
    mavenExecute:
      goals: 'clean package'

  Acceptance:
    cloudFoundryDeploy:
      cloudFoundry:
        space: 'dev'
      manifest: 'manifest-dev.yml'
      smokeTestScript: 'smoke-test.sh'

  'Staging':
    cloudFoundryDeploy:
      cloudFoundry:
        space: 'staging'
      manifest: 'manifest-staging.yml'
      smokeTestScript: 'smoke-test.sh'

  Confirm:
    # Manual approval before production

  Release:
    cloudFoundryDeploy:
      deployType: 'blue-green'
      cloudFoundry:
        space: 'production'
      manifest: 'manifest-prod.yml'
      smokeTestScript: 'smoke-test.sh'
      keepOldInstance: true
```

### Example 4: Security and Compliance Focus

**Security-first pipeline**:

```yaml
general:
  buildTool: 'maven'

stages:
  Build:
    mavenExecute:
      goals: 'clean package'
    mavenExecuteStaticCodeChecks: true

  Security:
    # SAST scanning
    checkmarxExecuteScan:
      projectName: 'secure-app'
      vulnerabilityThresholdHigh: 0
      vulnerabilityThresholdMedium: 5

    # SCA scanning
    whitesourceExecuteScan:
      productName: 'secure-app'
      failOnSevereVulnerabilities: true

    # Container scanning
    executeDockerScan:
      dockerImage: 'my-app:${version}'
      severityThreshold: 'high'

  Compliance:
    sonarExecuteScan:
      projectKey: 'secure-app'
      qualityGates:
        - metric: 'security_rating'
          threshold: 'A'
          operator: 'EQUALS'
        - metric: 'coverage'
          threshold: 80
          operator: 'GREATER_THAN'

  Release:
    # Only after all checks pass
    cloudFoundryDeploy:
      cloudFoundry:
        space: 'production'
```

## Advanced Stage Configuration

### Conditional Stage Execution

**Based on branch**:
```groovy
// In pipeline extension
if (env.BRANCH_NAME == 'main') {
    stage('Release') {
        cloudFoundryDeploy script: this
    }
}
```

**Based on configuration**:
```yaml
stages:
  Performance:
    # Only executes if this configuration exists
    gatlingExecuteTests:
      pomPath: 'performance/pom.xml'
```

### Stage Extensions

**Extend stages with custom logic**:

`.pipeline/extensions/Build.groovy`:
```groovy
void call(Map parameters) {
    // Pre-stage actions
    echo "Running custom pre-build checks"
    sh 'scripts/pre-build.sh'

    // Execute standard stage
    // (configuration from config.yml is applied)
    piperStageWrapper(parameters)

    // Post-stage actions
    echo "Running custom post-build tasks"
    sh 'scripts/post-build.sh'
}
```

**Enable extensions**:
```yaml
steps:
  piperStageWrapper:
    projectExtensionsDirectory: '.pipeline/extensions/'
```

### Stage Locking

**Prevent concurrent stage execution**:

```yaml
steps:
  piperStageWrapper:
    stageLocking: true  # Default is true
```

**Configure lock behavior**:
```groovy
// In extension
lock(resource: 'production-deployment', quantity: 1) {
    cloudFoundryDeploy script: this
}
```

### Custom Stage Parameters

**Add custom parameters**:

```yaml
stages:
  Release:
    # Standard Piper parameters
    cloudFoundryDeploy:
      cloudFoundry:
        space: 'production'

    # Custom parameters (for extensions)
    customNotification: true
    slackChannel: '#deployments'
```

**Use in extension**:
```groovy
void call(Map parameters) {
    if (parameters.customNotification) {
        slackSend(
            channel: parameters.slackChannel,
            message: "Starting Release stage"
        )
    }

    piperStageWrapper(parameters)
}
```

## Stage Sequence Control

**Default sequence**: Stages execute in predefined order

**Custom sequences** via extensions:

```groovy
// Custom Jenkinsfile
piperPipeline script: this, customDefaults: ['custom-defaults.yml']

// Or explicit stage calls
stage('Build') {
    piperStageBuild script: this
}

stage('Custom Stage') {
    // Your custom logic
}

stage('Release') {
    piperStageRelease script: this
}
```

## Common Patterns

### Pattern 1: Feature Branch Pipeline

**PR validation only**:

```yaml
general:
  buildTool: 'npm'
  # No productiveBranch set

stages:
  'Pull-Request Voting':
    npmExecuteScripts:
      runScripts:
        - 'lint'
        - 'test'
        - 'build'
```

### Pattern 2: Main Branch Pipeline

**Build, test, deploy to staging**:

```yaml
general:
  buildTool: 'maven'
  productiveBranch: 'main'

stages:
  Build:
    mavenExecute:
      goals: 'clean verify'

  Integration:
    mavenExecute:
      goals: 'integration-test'

  Acceptance:
    cloudFoundryDeploy:
      cloudFoundry:
        space: 'staging'
```

### Pattern 3: Release Pipeline

**Full production deployment**:

```yaml
general:
  buildTool: 'maven'
  productiveBranch: 'main'
  manualConfirmation: true

stages:
  Build:
    mavenExecute:
      goals: 'clean package'

  Security:
    checkmarxExecuteScan:
      projectName: 'prod-app'

  Compliance:
    sonarExecuteScan:
      projectKey: 'prod-app'

  Acceptance:
    cloudFoundryDeploy:
      cloudFoundry:
        space: 'staging'

  Confirm:
    # Manual approval

  Promote:
    nexusUpload:
      repository: 'releases'

  Release:
    cloudFoundryDeploy:
      deployType: 'blue-green'
      cloudFoundry:
        space: 'production'
```

## Troubleshooting

### Stage Not Executing

**Check**:
1. Stage has configuration in `.pipeline/config.yml`
2. Branch context matches requirements
3. Required files/patterns exist
4. No errors in previous stages

**Enable debug**:
```yaml
general:
  verbose: true
```

### Stage Failing Unexpectedly

**Check**:
1. Step configuration is correct
2. Credentials are available
3. Network connectivity
4. Resource availability

**Review logs**:
```groovy
// In Jenkins
// Check Console Output
// Look for error messages
```

### Configuration Not Taking Effect

**Check precedence**:
1. Verify configuration is in `stages:` section
2. Check for typos in stage names
3. Ensure step names match exactly
4. Review configuration hierarchy

**Verify merged config**:
```yaml
general:
  verbose: true  # Shows effective configuration
```

## Best Practices

1. **Use Meaningful Stage Names**: Follow conventions (Build, Test, Deploy)

2. **Configure Progressively**: Start simple, add complexity as needed
   ```yaml
   # Start here
   stages:
     Build:
       mavenExecute:
         goals: 'clean package'

   # Then add
   stages:
     Security:
       checkmarxExecuteScan:
         projectName: 'my-app'
   ```

3. **Environment-Specific Configuration**: Use stage config for differences
   ```yaml
   stages:
     Acceptance:
       cloudFoundryDeploy:
         cloudFoundry:
           space: 'acceptance'

     Release:
       cloudFoundryDeploy:
         cloudFoundry:
           space: 'production'
   ```

4. **Document Custom Behavior**: Comment non-obvious configuration
   ```yaml
   stages:
     Security:
       checkmarxExecuteScan:
         # Stricter threshold for production releases
         vulnerabilityThresholdHigh: 0
   ```

5. **Test Stage Configuration**: Use feature branches to test changes

6. **Use Extensions for Custom Logic**: Keep config declarative, logic in extensions

7. **Enable Appropriate Gates**: Use Confirm stage for production

8. **Monitor Stage Execution**: Review stage duration and success rates

---

**Next**: [Step Configuration](05-step-configuration.md) - Detailed step configuration examples

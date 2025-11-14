# Jenkins Library - Utility Steps

This document covers utility steps in jenkins-library for notifications, result publishing, pipeline management, monitoring, and debugging.

## Overview

Utility steps provide essential cross-cutting functionality for CI/CD pipelines:

- **Notifications**: Email, Slack, SAP Alert Notification Service
- **Result Publishing**: Test results, code analysis, warnings
- **Pipeline Utilities**: State management, file stashing, environment variables
- **Monitoring**: Health checks, duration measurement, metrics collection
- **Debugging**: Debug report archiving
- **Other Utilities**: Shell execution, Vault rotation, Spinnaker integration

---

## Notification Steps

### Overview

Send alerts about pipeline events through email, Slack, or SAP Alert Notification Service.

### mailSendNotification

Send email notifications about build status.

```groovy
post { always { mailSendNotification script: this } }
```

**Config:** `recipients`, `notifyCulprits`, `notificationAttachment`

### slackSendNotification

Send notifications to Slack channels. Requires Slack plugin and credentials.

```groovy
post { failure { slackSendNotification(script: this, channel: '#critical-alerts') } }
```

**Config:** `channel`, `credentialsId`, `color`

### ansSendEvent

Send events to SAP Alert Notification Service.

**Location**: `vars/ansSendEvent.groovy`, `cmd/ansSendEvent.go`

```groovy
ansSendEvent(script: this, ansServiceKeyCredentialsId: "myANSCredential",
    eventType: "errorEvent", severity: "ERROR", category: "EXCEPTION",
    subject: "Build Failed", body: "Details...", priority: 3,
    tags: [pipeline: "Production"])
```

**Key Parameters:** `severity` (INFO/NOTICE/WARNING/ERROR/FATAL), `category` (NOTIFICATION/ALERT/EXCEPTION), `priority` (1-1000)

### Best Practices

- Place notifications in pipeline `post` sections
- Use different channels for different severities
- Include context (build URL, branch, error details)
- Avoid notification spam on development branches

---

## Result Publishing Steps

### Overview

Publish test results, code analysis, and quality metrics to Jenkins.

### checksPublishResults

Publishes static code analysis results from PMD, Checkstyle, FindBugs, ESLint, etc.

**Prerequisites:** Static check result files, [warnings-ng plugin](https://plugins.jenkins.io/warnings-ng/)

**Supported Tools:** tasks, pmd, cpd, findbugs, checkstyle, eslint, pylint

```groovy
// Java
checksPublishResults(archive: true, pmd: true, cpd: true, findbugs: true, checkstyle: true)
// JavaScript
checksPublishResults(archive: true, eslint: [pattern: '**/eslint-results.xml'])
// With quality gates
checksPublishResults(pmd: [pattern: '**/pmd.xml',
    qualityGates: [[threshold: 1, type: 'TOTAL_HIGH', unstable: false]]])
```

### testsPublishResults

Publishes test results and code coverage (JUnit, JaCoCo, Cobertura, JMeter).

**Prerequisites:** junit, jacoco, cobertura, performance plugins

```groovy
testsPublishResults(
    junit: [pattern: '**/TEST-*.xml', updateResults: true, archive: true],
    jacoco: [pattern: '**/target/*.exec', archive: true])
```

**Options:** JUnit (`pattern`, `archive`, `updateResults`), JaCoCo (`pattern`, `include`, `exclude`), Cobertura, JMeter

### piperPublishWarnings

Publishes warnings from build logs and reports.

**Basic Usage:**
```groovy
piperPublishWarnings script: this
```

### Best Practices

- Always set `archive: true`
- Define quality gates for build quality
- Standardize file naming patterns
- Monitor quality trends over time

---

## Pipeline Utility Steps

### Overview

Manage pipeline state, file stashing, and environment variables across stages.

### pipelineStashFiles

Stashes files before/after build for use across nodes.

**Location**: `vars/pipelineStashFiles.groovy`

**Default Stashes:**

| Name | Pattern | Description |
|------|---------|-------------|
| buildDescriptor | `**/pom.xml, **/package.json, **/mta*.y*ml` | Build files |
| deployDescriptor | `**/manifest*.y*ml, helm/**` | Deploy files |
| tests | `**/src/**, **/*.xml` | Test files |
| classFiles | `**/target/classes/**/*.class` | Compiled classes |
| sonar | `**/jacoco*.exec` | SonarQube files |

```groovy
pipelineStashFiles(script: this) { mavenBuild script: this }
// Custom stashes
pipelineStashFiles(script: this,
    stashIncludes: [custom: '**/custom/**'],
    stashExcludes: [tests: '**/ignored/**']) { /* build */ }
```

**Related:** `pipelineStashFilesBeforeBuild`, `pipelineStashFilesAfterBuild`

### setupCommonPipelineEnvironment

Initializes the common pipeline environment.

**Location**: `vars/setupCommonPipelineEnvironment.groovy`

```groovy
stage('Init') { steps { setupCommonPipelineEnvironment script: this } }
```

**Options:** `configFile`, `customDefaults`, `inferBuildTool`

### readPipelineEnv / writePipelineEnv

```groovy
writePipelineEnv(script: this, key: 'buildVersion', value: '1.2.3')
def version = readPipelineEnv(script: this, key: 'buildVersion')
```

### Best Practices

- Initialize `setupCommonPipelineEnvironment` first
- Minimize stashed files to reduce overhead
- Use standard stashes when possible
- Share state via environment variables

---

## Monitoring Steps

### Overview

Track pipeline health, measure performance, and collect metrics.

### healthExecuteCheck

Execute health checks on deployed applications.

**Location**: `vars/healthExecuteCheck.groovy`, `cmd/healthExecuteCheck.go`

Requires unauthenticated health endpoint (e.g., Spring Boot `/health`).

```groovy
healthExecuteCheck testServerUrl: 'https://myapp.example.com'
```

### durationMeasure

Measures execution duration of stages or steps.

**Location**: `vars/durationMeasure.groovy`

```groovy
durationMeasure(script: this, measurementName: 'build_duration') { mavenBuild script: this }
durationMeasure(script: this, measurementName: 'scan_duration') { /* security scan */ }
```

### influxWriteData

Writes pipeline metrics to InfluxDB.

**Location**: `vars/influxWriteData.groovy`, `cmd/influxWriteData.go`

**Prerequisites:** InfluxDB instance, credentials

```groovy
post { always { influxWriteData script: this } }
```

**Config (1.8):** `serverUrl`, `authToken` (username:password), `bucket`
**Config (2.0):** Add `organization`

**Setup:** `docker run -d -p 8086:8086 influxdb:1.8`

**Collected Metrics:**

| Measurement | Description |
|-------------|-------------|
| jenkins_data | Build result, time, test counts |
| pipeline_data | Duration measurements from `durationMeasure` |
| jenkins_custom_data | Custom data, errors |
| step_data | Individual step results |
| jacoco_data / cobertura_data | Coverage metrics |

**Custom Metrics:**
```groovy
commonPipelineEnvironment.setInfluxCustomDataProperty('deploymentCount', 42)
```

### Best Practices

- Measure key stages (build, test, deploy)
- Set up Grafana dashboards for visualization
- Track trends over time
- Alert on duration anomalies

---

## Debugging Steps

### debugReportArchive

Archives debug information for troubleshooting.

**Location**: `vars/debugReportArchive.groovy`

```groovy
post { failure { debugReportArchive script: this } }
```

Automatically collects logs, configuration, and environment information.

---

## Other Utility Steps

### shellExecute

Execute shell scripts with remote download support.

**Location**: `vars/shellExecute.groovy`, `cmd/shellExecute.go`

```groovy
// Local
shellExecute(script: this, sources: ['.pipeline/scripts/build.sh'])
// Multiple with arguments
shellExecute(script: this, sources: ['.pipeline/deploy.sh'],
    scriptArguments: ['production', 'us-east-1'])
// Remote (GitHub)
shellExecute(script: this,
    sources: ['https://github.com/api/v3/repos/org/repo/contents/script.sh'],
    githubToken: 'token')
```

### vaultRotateSecretId

Rotate Vault AppRole Secret ID before expiry.

**Location**: `vars/vaultRotateSecretId.groovy`, `cmd/vaultRotateSecretId.go`

```groovy
vaultRotateSecretId(script: this, vaultServerUrl: 'https://vault.example.com',
    vaultAppRoleSecretTokenCredentialsId: 'vault-secret-id',
    secretStore: 'jenkins', daysBeforeExpiry: 15)
```

**Secret Stores:** `jenkins` (+ jenkinsUrl/jenkinsUsername), `ado` (+ adoOrganization/adoProject/adoPipelineId), `github` (+ owner/repository)

### spinnakerTriggerPipeline

Trigger Spinnaker pipelines from Jenkins.

**Location**: `vars/spinnakerTriggerPipeline.groovy`

```groovy
spinnakerTriggerPipeline(script: this,
    spinnakerPipeline: 'Deploy to Production', spinnakerApplication: 'myapp')
```

---

## Summary

jenkins-library provides comprehensive utility steps:

- **3 Notification channels**: Email, Slack, SAP ANS
- **3 Result publishers**: Code checks, tests, warnings
- **6 Pipeline utilities**: State management, file stashing, environment
- **3 Monitoring tools**: Health checks, duration tracking, InfluxDB metrics
- **1 Debugging tool**: Debug report archiving
- **3 Other utilities**: Shell execution, Vault rotation, Spinnaker integration

## Best Practices Summary

1. **Notifications**: Use post blocks, avoid spam, target channels appropriately
2. **Result Publishing**: Always archive, set quality gates, monitor trends
3. **Pipeline Utilities**: Initialize early, minimize stashes, share state carefully
4. **Monitoring**: Measure key stages, visualize with Grafana, alert on anomalies
5. **Debugging**: Archive on failure, include context, clean up regularly
6. **Security**: Rotate secrets, use Vault for sensitive data

---

## Next Steps

- [Build Tools Documentation](./01-build-tools.md)
- [Security Scanning Documentation](./02-security-scanning.md)
- [Testing Frameworks Documentation](./03-testing-frameworks.md)
- [Deployment Documentation](./04-deployment.md)

## Resources

- **GitHub**: https://github.com/SAP/jenkins-library
- **Documentation**: https://www.project-piper.io/
- **Step Reference**: `jenkins-library/documentation/docs/steps/`

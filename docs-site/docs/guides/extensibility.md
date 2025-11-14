# Pipeline Extensibility Guide

Guide for extending and customizing Project Piper pipelines.

## Prerequisites

- Understanding of CI/CD platform (Jenkins/GitHub/Azure)
- Basic Groovy (Jenkins) or YAML knowledge
- Familiarity with Piper configuration
- Project build process knowledge

## Extension Methods

Ordered by recommended usage:

1. **Configuration-Based**: Modify through config without code
2. **Stage Extensions**: Extend individual stages
3. **Step Hooks**: Custom logic before/after steps
4. **Custom Steps**: Create new steps
5. **Modified Pipeline**: Copy and modify entire pipeline

## Stage Extensions (Jenkins)

### Creating Extensions

Create `.pipeline/extensions/<StageName>.groovy`:

```groovy
void call(Map params) {
  // Pre-stage logic
  echo "Stage: ${params.stageName}"
  echo "Config: ${params.config}"

  // Execute original stage
  params.originalStage()

  // Post-stage logic
  echo "Branch: ${params.script.commonPipelineEnvironment.gitBranch}"
}

return this
```

### Parameters

| Parameter | Description |
|-----------|-------------|
| `script` | Global script and commonPipelineEnvironment |
| `originalStage` | Execute original stage |
| `stageName` | Current stage name |
| `config` | Stage and general configuration |

### Available Stages

Build, Additional Unit Tests, Integration, Acceptance, Security, Performance, Compliance, Promote, Release

**Note**: Init stage cannot be extended

### Example: Pre/Post Actions

```groovy
void call(Map params) {
  echo "Pre-build validation"

  // Execute build
  params.originalStage()

  // Archive additional artifacts
  archiveArtifacts artifacts: 'custom-reports/**/*'

  // Send notification
  if (params.config.customNotifications?.enabled) {
    emailext(
      subject: "Build completed: ${env.JOB_NAME}",
      body: "Build #${env.BUILD_NUMBER} completed",
      to: params.config.customNotifications.recipients
    )
  }
}

return this
```

### Example: Replace Stage

```groovy
void call(Map params) {
  echo "Custom security implementation"

  node() {
    stage('SAST Scan') {
      dockerExecute(
        script: params.script,
        dockerImage: 'custom-sast:latest'
      ) {
        sh 'run-sast-scan.sh'
      }
    }

    stage('License Check') {
      sh 'check-licenses.sh'
    }
  }

  // originalStage() NOT called - completely replaced
}

return this
```

### Example: Conditional Logic

```groovy
void call(Map params) {
  def branch = params.script.commonPipelineEnvironment.gitBranch

  if (branch == 'main') {
    timeout(time: 24, unit: 'HOURS') {
      input message: 'Deploy to production?'
    }
    params.originalStage()
    sh 'npm run smoke-tests:production'
  } else if (branch == 'develop') {
    params.originalStage()
  } else {
    echo "Feature branch - skipping deployment"
  }
}

return this
```

### Example: Additional Linters

```groovy
void call(Map params) {
  params.originalStage()

  // Add Checkstyle
  mavenExecute(
    script: params.script,
    goals: ['checkstyle:checkstyle']
  )

  recordIssues(
    enabledForFailure: true,
    tool: checkStyle()
  )
}

return this
```

### Disable Extensions

```groovy
environment {
  PIPER_DISABLE_EXTENSIONS = 'true'
}
```

Or:
```yaml
general:
  disableExtensions: true
```

## Step Hooks

### GitHub Actions

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Pre-build Hook
        run: npm run setup-test-data

      - name: Build
        uses: SAP/project-piper-action@main
        with:
          step-name: 'npmExecuteScripts'

      - name: Post-build Hook
        run: npm run validate-build
```

### Azure DevOps

```yaml
extends:
  template: sap-piper-pipeline.yml@piper-pipeline-azure
  parameters:
    buildPreSteps:
      - script: npm run pre-build
        displayName: 'Pre-build Hook'

    buildPostSteps:
      - script: npm run post-build
        displayName: 'Post-build Hook'
      - task: PublishTestResults@2
```

## Custom Steps

### Create Custom Piper Step

**vars/customSecurityScan.groovy:**
```groovy
import com.sap.piper.ConfigurationHelper

void call(Map parameters = [:]) {
  def script = parameters.script ?: this

  Map config = ConfigurationHelper.newInstance(this)
    .loadStepDefaults()
    .mixinGeneralConfig(script.commonPipelineEnvironment)
    .mixinStepConfig(script.commonPipelineEnvironment, parameters)
    .use()

  dockerExecute(
    script: script,
    dockerImage: config.scannerImage
  ) {
    sh "custom-scanner --target ${config.scanTarget}"
  }
}
```

**resources/defaults/customSecurityScan.yml:**
```yaml
customSecurityScan:
  scannerImage: 'my-org/scanner:latest'
  scanTarget: '.'
  failOnHigh: true
```

**Use in Jenkinsfile:**
```groovy
@Library(['piper-lib-os', 'my-piper-steps']) _

node() {
  stage('Build') {
    mavenBuild script: this
  }
  stage('Custom Scan') {
    customSecurityScan script: this
  }
}
```

## Platform-Specific Extensions

### Jenkins: Modified Pipeline

```groovy
@Library('piper-lib-os') _

call script: this

void call(Map parameters) {
  node() {
    stage('Init') {
      checkout scm
      setupCommonPipelineEnvironment script: this
    }

    stage('Custom Pre-Build') {
      sh 'npm run custom-prebuild'
    }

    stage('Build') {
      piperStageWrapper(script: this, stageName: 'Build') {
        mavenBuild script: this
      }
    }

    parallel(
      'Unit Tests': {
        mavenExecute script: this, goals: ['test']
      },
      'Integration': {
        mavenExecute script: this, goals: ['integration-test']
      }
    )

    stage('Deploy') {
      cloudFoundryDeploy script: this
    }
  }
}
```

### GitHub: Custom Reusable Workflow

**.github/workflows/custom-build.yml:**
```yaml
name: Custom Build

on:
  workflow_call:
    inputs:
      build-tool:
        required: true
        type: string

jobs:
  custom-build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Pre-build
        run: npm run prebuild

      - name: Build
        uses: SAP/project-piper-action@main
        with:
          step-name: '${{ inputs.build-tool }}Build'

      - name: Post-build
        run: npm run postbuild
```

**Use:**
```yaml
jobs:
  build:
    uses: ./.github/workflows/custom-build.yml
    with:
      build-tool: 'maven'
```

### Azure: Custom Template

**templates/custom-pipeline.yml:**
```yaml
parameters:
  - name: stages
    type: object

stages:
  - stage: Init
    jobs:
      - job: Initialize
        steps:
          - checkout: self
          - script: echo "Init"

  - ${{ each stage in parameters.stages }}:
      - stage: ${{ stage.name }}
        jobs:
          - job: ${{ stage.name }}Job
            steps:
              - ${{ if stage.preSteps }}:
                  - ${{ stage.preSteps }}

              - task: piper@1
                inputs:
                  stepName: ${{ stage.piperStep }}

              - ${{ if stage.postSteps }}:
                  - ${{ stage.postSteps }}
```

**Use:**
```yaml
extends:
  template: templates/custom-pipeline.yml
  parameters:
    stages:
      - name: Build
        piperStep: 'mavenBuild'
        preSteps:
          - script: echo "Pre-build"
```

## Examples

### Custom Notifications

```groovy
void call(Map params) {
  try {
    params.originalStage()
    slackSend(
      channel: params.config.notifications.channel,
      color: 'good',
      message: "Build successful: ${env.JOB_NAME}"
    )
  } catch (Exception e) {
    slackSend(
      channel: params.config.notifications.channel,
      color: 'danger',
      message: "Build failed: ${env.JOB_NAME}"
    )
    throw e
  }
}

return this
```

### Database Migration (GitHub)

```yaml
jobs:
  build:
    services:
      postgres:
        image: postgres:13
    steps:
      - uses: actions/checkout@v3
      - name: Migrations
        run: npm run db:migrate
      - uses: SAP/project-piper-action@main
      - name: Seed Data
        run: npm run db:seed
```

### Performance Testing (Azure)

```yaml
extends:
  template: sap-piper-pipeline.yml@piper-pipeline-azure
  parameters:
    performancePreSteps:
      - script: kubectl apply -f k8s/test.yml
        displayName: 'Setup Environment'

    performancePostSteps:
      - task: PublishLoadTestResults@1
      - script: kubectl delete -f k8s/test.yml
        displayName: 'Cleanup'
```

### Multi-Cloud Deployment

```groovy
void call(Map params) {
  parallel(
    'Cloud Foundry': {
      cloudFoundryDeploy(
        script: params.script,
        cloudFoundry: params.config.cloudFoundry
      )
    },
    'Kubernetes': {
      kubernetesDeploy(
        script: params.script,
        containerRegistryUrl: params.config.kubernetes.registry
      )
    },
    'AWS ECS': {
      sh """
        aws ecs update-service \
          --cluster ${params.config.aws.cluster} \
          --service ${params.config.aws.service}
      """
    }
  )
}

return this
```

## Troubleshooting

### Extension Not Loaded

**Solutions:**
1. Check file naming: `.pipeline/extensions/<StageName>.groovy`
2. Verify `return this` at end
3. Check `PIPER_DISABLE_EXTENSIONS` not set

### Parameters Null

```groovy
void call(Map params) {
  if (!params.script) {
    error "Script parameter required"
  }

  def config = params.config ?: [:]
}

return this
```

### Original Stage Not Executing

```groovy
void call(Map params) {
  if (params.originalStage) {
    params.originalStage()
  } else {
    echo "Warning: originalStage not available"
  }
}

return this
```

### GitHub Workflow Order

```yaml
jobs:
  pre-build:
    steps:
      - run: echo "Pre"

  build:
    needs: pre-build
    steps:
      - uses: SAP/project-piper-action@main

  post-build:
    needs: build
    steps:
      - run: echo "Post"
```

### Azure Parameter Not Applied

```yaml
extends:
  template: sap-piper-pipeline.yml@piper-pipeline-azure
  parameters:
    buildPreSteps:  # Correct name
      - script: echo "test"
```

## Best Practices

1. **Minimal Extensions**: Only extend what's necessary
2. **Documentation**: Document purpose and usage
3. **Testing**: Test in non-production first
4. **Version Control**: Track changes
5. **Error Handling**: Add proper error handling
6. **Configuration**: Make extensions configurable
7. **Reusability**: Create reusable patterns
8. **Maintenance**: Keep aligned with Piper updates
9. **Performance**: Optimize execution time
10. **Security**: Follow security best practices

## Additional Resources

- [Jenkins Shared Libraries](https://www.jenkins.io/doc/book/pipeline/shared-libraries/)
- [GitHub Reusable Workflows](https://docs.github.com/en/actions/using-workflows/reusing-workflows)
- [Azure Templates](https://docs.microsoft.com/en-us/azure/devops/pipelines/process/templates)
- [Piper Configuration](https://www.project-piper.io/configuration/)

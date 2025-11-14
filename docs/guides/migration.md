# Platform Migration Guide

Guide for migrating Project Piper pipelines between Jenkins, GitHub Actions, and Azure DevOps.

## Migration Overview

### Platform Differences

| Feature | Jenkins | GitHub Actions | Azure DevOps |
|---------|---------|----------------|--------------|
| **Language** | Groovy | YAML | YAML |
| **Execution** | Agents/Nodes | Runners | Agents |
| **Extensibility** | Groovy Scripts | Workflow Composition | Template Parameters |
| **Secrets** | Credentials Plugin | Repository Secrets | Variable Groups |
| **Artifacts** | Archive Plugin | Actions Artifacts | Pipeline Artifacts |
| **Caching** | Manual/Plugins | Built-in | Built-in |

### What Transfers Easily

- Piper Configuration (.pipeline/config.yml)
- Build logic and steps
- Docker images
- Vault integration

### What Needs Migration

- Pipeline definition syntax
- Stage extensions
- Secret management
- Conditional logic
- Parallel execution

## Jenkins to GitHub Actions

### Migration Steps

#### 1. Analyze Jenkins Pipeline

**Jenkinsfile:**
```groovy
@Library('piper-lib-os') _
piperPipeline script: this
```

#### 2. Create GitHub Workflow

**.github/workflows/pipeline.yml:**
```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: SAP/project-piper-action@main
        with:
          step-name: 'mavenBuild'
        env:
          PIPER_vaultAppRoleID: ${{ secrets.PIPER_VAULTAPPROLEID }}
          PIPER_vaultAppRoleSecretID: ${{ secrets.PIPER_VAULTAPPROLESECRETID }}
```

#### 3. Migrate Credentials

**Jenkins:** Uses Jenkins Credentials Store
**GitHub:** Uses Repository Secrets

```yaml
env:
  PIPER_cloudFoundry_username: ${{ secrets.CF_USERNAME }}
  PIPER_cloudFoundry_password: ${{ secrets.CF_PASSWORD }}
```

#### 4. Migrate Extensions

**Jenkins (.pipeline/extensions/Build.groovy):**
```groovy
void call(Map params) {
  echo "Pre-build"
  params.originalStage()
  echo "Post-build"
}
return this
```

**GitHub (workflow composition):**
```yaml
steps:
  - name: Pre-build
    run: echo "Pre-build"
  - uses: SAP/project-piper-action@main
  - name: Post-build
    run: echo "Post-build"
```

#### 5. Migrate Multi-Stage

**Jenkins:**
```groovy
node() {
  stage('Build') { mavenBuild script: this }
  stage('Test') { mavenExecute script: this, goals: ['test'] }
  stage('Deploy') { cloudFoundryDeploy script: this }
}
```

**GitHub:**
```yaml
jobs:
  build:
    steps:
      - uses: actions/checkout@v3
      - uses: SAP/project-piper-action@main
        with:
          step-name: 'mavenBuild'

  test:
    needs: build
    steps:
      - uses: actions/checkout@v3
      - uses: SAP/project-piper-action@main
        with:
          step-name: 'mavenExecute'
          flags: '--goals test'

  deploy:
    needs: test
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v3
      - uses: SAP/project-piper-action@main
        with:
          step-name: 'cloudFoundryDeploy'
```

#### 6. Migrate Parallel Execution

**Jenkins:**
```groovy
parallel(
  'Unit Tests': {
    mavenExecute script: this, goals: ['test']
  },
  'Integration Tests': {
    npmExecuteScripts script: this, runScripts: ['test:integration']
  }
)
```

**GitHub (parallel jobs):**
```yaml
jobs:
  unit-tests:
    steps:
      - uses: SAP/project-piper-action@main
        with:
          step-name: 'mavenExecute'
          flags: '--goals test'

  integration-tests:
    steps:
      - uses: SAP/project-piper-action@main
        with:
          step-name: 'npmExecuteScripts'
          flags: '--runScripts=test:integration'
```

## Jenkins to Azure DevOps

### Migration Steps

#### 1. Install Piper Task

1. Navigate to Azure Marketplace
2. Install "Project Piper Azure Task"
3. Approve for organization

#### 2. Create Service Connections

```
Project Settings → Service connections
- GitHub Enterprise Server: github-tools
- Docker Registry: docker-registry
```

#### 3. Create Variable Groups

```
Pipelines → Library → Variable groups
Name: piper-config
Variables:
  - hyperspace.vault.roleId
  - hyperspace.vault.secretId (secret)
```

#### 4. Create Pipeline

**azure-pipelines.yml:**
```yaml
trigger:
  branches:
    include: [main, develop]

pool:
  vmImage: 'ubuntu-latest'

variables:
  - group: piper-config

stages:
  - stage: Build
    jobs:
      - job: BuildJob
        steps:
          - task: piper@1
            inputs:
              stepName: 'mavenBuild'
```

#### 5. Migrate Extensions

**Jenkins:**
```groovy
void call(Map params) {
  echo "Pre-build"
  params.originalStage()
  echo "Post-build"
}
return this
```

**Azure (template parameters):**
```yaml
extends:
  template: sap-piper-pipeline.yml@piper-pipeline-azure
  parameters:
    buildPreSteps:
      - script: echo "Pre-build"
    buildPostSteps:
      - script: echo "Post-build"
```

#### 6. Migrate Multi-Stage

**azure-pipelines.yml:**
```yaml
stages:
  - stage: Build
    jobs:
      - job: BuildJob
        steps:
          - task: piper@1
            inputs:
              stepName: 'mavenBuild'

  - stage: Test
    dependsOn: Build
    jobs:
      - job: TestJob
        steps:
          - task: piper@1
            inputs:
              stepName: 'mavenExecute'
              flags: '--goals test'

  - stage: Deploy
    dependsOn: Test
    condition: eq(variables['Build.SourceBranch'], 'refs/heads/main')
    jobs:
      - deployment: DeployProd
        environment: 'production'
        strategy:
          runOnce:
            deploy:
              steps:
                - task: piper@1
                  inputs:
                    stepName: 'cloudFoundryDeploy'
```

## GitHub Actions to Azure DevOps

### Migration Steps

#### 1. Identify Workflow

**GitHub (.github/workflows/pipeline.yml):**
```yaml
on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: SAP/project-piper-action@main
        with:
          step-name: 'mavenBuild'
```

#### 2. Convert to Azure

**azure-pipelines.yml:**
```yaml
trigger:
  branches:
    include: [main]

pool:
  vmImage: 'ubuntu-latest'

stages:
  - stage: Build
    jobs:
      - job: BuildJob
        steps:
          - checkout: self
          - task: piper@1
            inputs:
              stepName: 'mavenBuild'
```

#### 3. Migrate Secrets

**GitHub:**
```yaml
env:
  PIPER_vaultAppRoleID: ${{ secrets.PIPER_VAULTAPPROLEID }}
```

**Azure:**
```yaml
variables:
  - group: piper-config
# Variables automatically available
```

#### 4. Migrate Matrix

**GitHub:**
```yaml
strategy:
  matrix:
    os: [ubuntu-latest, windows-latest]
    node: [16, 18]
runs-on: ${{ matrix.os }}
```

**Azure:**
```yaml
strategy:
  matrix:
    ubuntu_node16:
      imageName: 'ubuntu-latest'
      nodeVersion: '16'
    ubuntu_node18:
      imageName: 'ubuntu-latest'
      nodeVersion: '18'
pool:
  vmImage: $(imageName)
```

## Azure DevOps to GitHub Actions

### Migration Steps

**Azure:**
```yaml
trigger:
  branches:
    include: [main]

pool:
  vmImage: 'ubuntu-latest'

variables:
  - group: piper-config

stages:
  - stage: Build
    jobs:
      - job: BuildJob
        steps:
          - task: piper@1
            inputs:
              stepName: 'mavenBuild'
```

**GitHub:**
```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: SAP/project-piper-action@main
        with:
          step-name: 'mavenBuild'
        env:
          PIPER_vaultAppRoleID: ${{ secrets.PIPER_VAULTAPPROLEID }}
          PIPER_vaultAppRoleSecretID: ${{ secrets.PIPER_VAULTAPPROLESECRETID }}
```

## Configuration Migration

### Piper Configuration

The `.pipeline/config.yml` is **platform-agnostic**:

```yaml
general:
  buildTool: 'maven'
  productiveBranch: 'main'

steps:
  mavenBuild:
    goals: ['clean', 'install']
  cloudFoundryDeploy:
    cloudFoundry:
      apiEndpoint: 'https://api.cf.example.com'
      org: 'my-org'
      space: 'dev'
```

### Credential Configuration

**Jenkins:**
```yaml
steps:
  cloudFoundryDeploy:
    cloudFoundry:
      credentialsId: 'cf-credentials'
```

**GitHub Actions:**
```yaml
# Pass via environment variables in workflow
```

**Azure DevOps:**
```yaml
# Use Vault via variable groups
```

## Common Patterns

### Conditional Deployment

**Jenkins:**
```groovy
if (env.BRANCH_NAME == 'main') {
  cloudFoundryDeploy script: this
}
```

**GitHub:**
```yaml
if: github.ref == 'refs/heads/main'
```

**Azure:**
```yaml
condition: eq(variables['Build.SourceBranch'], 'refs/heads/main')
```

### Artifact Handling

**Jenkins:**
```groovy
archiveArtifacts artifacts: '**/*.jar'
```

**GitHub:**
```yaml
- uses: actions/upload-artifact@v3
  with:
    name: build-artifacts
    path: '**/*.jar'
```

**Azure:**
```yaml
- task: PublishPipelineArtifact@1
  inputs:
    targetPath: '$(Build.SourcesDirectory)'
    artifact: 'build-artifacts'
```

### Environment Variables

**Jenkins:**
```groovy
def env = env.BRANCH_NAME == 'main' ? 'prod' : 'dev'
```

**GitHub:**
```yaml
- run: |
    if [ "${{ github.ref }}" == "refs/heads/main" ]; then
      echo "CF_SPACE=prod" >> $GITHUB_ENV
    fi
```

**Azure:**
```yaml
- script: |
    if [ "$(Build.SourceBranch)" == "refs/heads/main" ]; then
      echo "##vso[task.setvariable variable=CF_SPACE]prod"
    fi
```

## Troubleshooting

### Different Behaviors

**Solution:**
Enable verbose logging:
```yaml
general:
  verbose: true
```

### Credential Access

**Solution:**
1. Verify credentials exist in new platform
2. Update configuration for credential format
3. Test with minimal pipeline first

### Extensions Don't Transfer

**Solution:**
Convert to platform-specific approach:
- **Jenkins**: .groovy files
- **GitHub**: Workflow composition
- **Azure**: Template parameters

### Parallel Execution Differences

**Note:**
- Jenkins: Parallel stages share workspace
- GitHub/Azure: Parallel jobs are isolated

Adjust artifact sharing accordingly.

### Path Differences

**Solution:**
Use relative paths or configure working directory:
- **Jenkins**: Workspace root
- **GitHub**: `/home/runner/work/<repo>/<repo>`
- **Azure**: `$(Build.SourcesDirectory)`

## Migration Checklist

- [ ] Identify all stages and steps
- [ ] Export credentials/secrets
- [ ] Create service connections
- [ ] Migrate .pipeline/config.yml
- [ ] Convert pipeline syntax
- [ ] Migrate extensions
- [ ] Update credential references
- [ ] Configure runners/agents
- [ ] Test in non-production
- [ ] Update documentation
- [ ] Train team
- [ ] Set up monitoring
- [ ] Plan rollback

## Best Practices

1. **Incremental**: Migrate one pipeline at a time
2. **Parallel Running**: Run both during transition
3. **Test Thoroughly**: All stages in non-prod
4. **Document**: Keep migration notes
5. **Version Control**: Clear commit messages
6. **Backup**: Keep working pipeline
7. **Monitor**: Set up alerting
8. **Training**: Ensure team understands new platform

## Additional Resources

- [Jenkins Setup Guide](./jenkins-setup.md)
- [GitHub Actions Setup Guide](./github-setup.md)
- [Azure DevOps Setup Guide](./azure-setup.md)
- [Extensibility Guide](./extensibility.md)
- [Project Piper Documentation](https://www.project-piper.io/)

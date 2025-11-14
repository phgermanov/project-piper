# Azure DevOps Setup Guide

Complete guide for setting up Project Piper on Azure DevOps.

## Prerequisites

### Required
- **Azure DevOps Organization**: With pipelines enabled
- **Azure DevOps Project**: Repository location
- **Piper Azure Task**: From marketplace
- **Service Connections**: GitHub, Docker, etc.

### Agent Requirements

**Microsoft-Hosted**: Windows, Ubuntu, macOS with standard tools

**Self-Hosted**: Windows Server 2016+/Ubuntu 18.04+, 2+ cores, 8GB RAM, Docker 20.10+, .NET Core 3.1+

## Installation

### Step 1: Install Piper Task

1. Navigate to Azure Marketplace
2. Search "Project Piper Azure Task"
3. Install to organization
4. Approve (contact team if needed)

### Step 2: Create Service Connections

**GitHub Connection:**
```
Project Settings → Service connections → New
Type: GitHub Enterprise Server
Name: github-tools
Server: https://github.com
Token: <your-pat>
```

**Docker Registry:**
```
Type: Docker Registry
Registry: https://docker.example.com
Username/Password: <credentials>
Name: docker-registry
```

### Step 3: Configure Variable Groups

```
Pipelines → Library → + Variable group
Name: piper-config
Variables:
  - hyperspace.vault.roleId: <vault-role-id>
  - hyperspace.vault.secretId: <secret> (mark secret)
```

### Step 4: Set Up Self-Hosted Agent (Optional)

```bash
# Linux
mkdir myagent && cd myagent
tar zxvf ~/Downloads/vsts-agent-linux-x64-*.tar.gz
./config.sh
sudo ./svc.sh install
sudo ./svc.sh start
```

## Configuration

### Project Structure
```
my-project/
├── .pipeline/
│   └── config.yml
├── azure-pipelines.yml
└── src/
```

### Basic Pipeline

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
              flags: '--publish --createBOM'
```

### Using SAP Piper Pipeline

```yaml
trigger:
  branches:
    include: [main]

resources:
  repositories:
    - repository: piper-pipeline-azure
      type: github
      name: project-piper/piper-pipeline-azure
      ref: main
      endpoint: github-tools

extends:
  template: sap-piper-pipeline.yml@piper-pipeline-azure
  parameters:
    piperVersion: 'latest'
```

### Piper Configuration

```yaml
general:
  buildTool: 'maven'
  productiveBranch: 'main'

steps:
  mavenBuild:
    goals: ['clean', 'install']
  mavenExecuteStaticCodeChecks:
    spotBugs: true
    pmd: true
```

### Task Inputs

| Input | Description | Default |
|-------|-------------|---------|
| `stepName` | Piper step | help |
| `flags` | Command flags | '' |
| `piperVersion` | OS Piper version | latest |
| `customConfigLocation` | Config path | .pipeline/config.yml |
| `dockerImage` | Docker image | '' |
| `gitHubConnection` | GitHub connection | '' |

## Common Patterns

### Multi-Stage Pipeline

```yaml
stages:
  - stage: Build
    jobs:
      - job: Build
        steps:
          - task: piper@1
            inputs:
              stepName: 'mavenBuild'

  - stage: Test
    dependsOn: Build
    jobs:
      - job: UnitTest
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

### Parallel Jobs

```yaml
stages:
  - stage: Test
    jobs:
      - job: UnitTests
        steps:
          - task: piper@1
            inputs:
              stepName: 'npmExecuteScripts'
              flags: '--runScripts=test:unit'

      - job: IntegrationTests
        steps:
          - task: piper@1
            inputs:
              stepName: 'npmExecuteScripts'
              flags: '--runScripts=test:integration'
```

### Template-Based Pipeline

**templates/build-template.yml:**
```yaml
parameters:
  - name: buildTool
    type: string

steps:
  - task: piper@1
    displayName: 'Build with ${{ parameters.buildTool }}'
    inputs:
      stepName: '${{ parameters.buildTool }}Build'
```

**azure-pipelines.yml:**
```yaml
stages:
  - stage: Build
    jobs:
      - job: BuildJob
        steps:
          - template: templates/build-template.yml
            parameters:
              buildTool: 'maven'
```

### Stage Extensions

```yaml
extends:
  template: sap-piper-pipeline.yml@piper-pipeline-azure
  parameters:
    buildPreSteps:
      - script: echo "Pre-build"
    buildPostSteps:
      - task: PublishTestResults@2
        inputs:
          testResultsFormat: 'JUnit'
    buildServiceContainers:
      postgres:
        image: postgres:13
```

### Conditional Deployment

```yaml
- stage: Deploy
  condition: |
    or(
      eq(variables['Build.SourceBranch'], 'refs/heads/main'),
      eq(variables['Build.SourceBranch'], 'refs/heads/develop')
    )
  jobs:
    - deployment: DeployToDev
      condition: eq(variables['Build.SourceBranch'], 'refs/heads/develop')
      environment: 'development'

    - deployment: DeployToProd
      condition: eq(variables['Build.SourceBranch'], 'refs/heads/main')
      environment: 'production'
```

## Examples

### Node.js Application

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
      - job: BuildAndTest
        steps:
          - task: piper@1
            displayName: 'Lint'
            inputs:
              stepName: 'npmExecuteLint'

          - task: piper@1
            displayName: 'Build'
            inputs:
              stepName: 'npmExecuteScripts'
              flags: '--runScripts=build'

          - task: PublishTestResults@2
            inputs:
              testResultsFormat: 'JUnit'
              testResultsFiles: '**/junit.xml'
```

### Maven with Caching

```yaml
variables:
  - group: piper-config
  - name: MAVEN_CACHE_FOLDER
    value: $(Pipeline.Workspace)/.m2/repository

stages:
  - stage: Build
    jobs:
      - job: MavenBuild
        steps:
          - task: Cache@2
            inputs:
              key: 'maven | "$(Agent.OS)" | **/pom.xml'
              path: $(MAVEN_CACHE_FOLDER)

          - task: piper@1
            inputs:
              stepName: 'mavenBuild'
```

### MTA Deployment

```yaml
stages:
  - stage: Build
    jobs:
      - job: MTABuild
        steps:
          - task: piper@1
            inputs:
              stepName: 'mtaBuild'
              flags: '--buildTarget CF'

          - task: PublishPipelineArtifact@1
            inputs:
              targetPath: '$(Build.SourcesDirectory)'
              artifact: 'mta-archive'

  - stage: Deploy
    jobs:
      - deployment: DeployToCloudFoundry
        environment: 'cf-production'
        strategy:
          runOnce:
            deploy:
              steps:
                - task: DownloadPipelineArtifact@2
                  inputs:
                    artifact: 'mta-archive'
                - task: piper@1
                  inputs:
                    stepName: 'cloudFoundryDeploy'
```

### Docker Build

```yaml
stages:
  - stage: Build
    jobs:
      - job: DockerBuild
        steps:
          - task: piper@1
            inputs:
              stepName: 'mavenBuild'
              dockerImage: 'maven:3.8-jdk-11'
              dockerOptions: '--volume /tmp:/tmp'
              gitHubConnection: 'github-tools'
              dockerRegistryConnection: 'docker-registry'
```

## Troubleshooting

### Extension Not Installed

**Error:** `No task could be found with identifier 'piper@1'`

**Solution:**
- Install from marketplace
- Verify organization has extension
- Check project permissions

### Service Connection Not Found

**Solution:**
1. Create in **Project Settings** → **Service connections**
2. Match name in pipeline
3. Check permissions

### Vault Credentials

**Solution:**
```yaml
variables:
  - group: piper-config  # Must be linked

# Variables automatically available to task
```

### Pipeline Environment

```yaml
stages:
  - stage: Build
    variables:
      pipelineEnvironment_b64: ''
    jobs:
      - job: BuildJob
        steps:
          - task: piper@1
            inputs:
              exportPipelineEnv: true

  - stage: Deploy
    variables:
      pipelineEnvironment_b64: $[ stageDependencies.Build.BuildJob.outputs['piper.pipelineEnvironment_b64'] ]
```

### Agent Disk Space

```bash
# Self-hosted agents
docker system prune -af
cd /agent/_work && find . -name "_temp" -exec rm -rf {} +
```

### Docker Permission

```bash
sudo usermod -aG docker $(whoami)
sudo systemctl restart vsts.agent.service
```

### Enable Diagnostics

```yaml
variables:
  system.debug: true
```

### Inspect Agent

```yaml
steps:
  - script: |
      echo "Agent: $(Agent.Name)"
      echo "OS: $(Agent.OS)"
      df -h
      docker --version
    displayName: 'Debug Info'
```

## Best Practices

1. **Variable Groups**: Centralize secrets
2. **Caching**: Cache dependencies
3. **Environments**: Use for approvals
4. **Artifacts**: Version build outputs
5. **Templates**: Reuse common patterns
6. **Self-Hosted Agents**: Better control
7. **Service Connections**: Secure credentials
8. **Conditions**: Optimize runs
9. **Monitor**: Track usage
10. **Documentation**: Document customizations

## Additional Resources

- [Azure Pipelines Docs](https://docs.microsoft.com/en-us/azure/devops/pipelines/)
- [Piper Azure Task](https://github.com/SAP/piper-azure-task)
- [Project Piper](https://www.project-piper.io/)
- [YAML Schema](https://docs.microsoft.com/en-us/azure/devops/pipelines/yaml-schema)

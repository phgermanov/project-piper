# Jenkins Setup Guide

Complete guide for setting up Project Piper on Jenkins.

## Prerequisites

### Required
- **Jenkins**: Version 2.289.1+
- **Java**: JDK 8+
- **Git**: Version 2.20+
- **Docker**: For containerized execution (recommended)

### Required Jenkins Plugins
```text
- Pipeline (workflow-aggregator)
- Pipeline: Shared Groovy Libraries
- Git, Docker Pipeline
- Kubernetes (if using K8s agents)
- Credentials Binding
```

## Installation

### Step 1: Configure Shared Library

Navigate to **Manage Jenkins** → **Configure System** → **Global Pipeline Libraries**:

```groovy
Name: piper-lib-os
Default version: master (or v1.120.0)
Load implicitly: [unchecked]
Allow default version override: [checked]

Retrieval: Modern SCM → Git
Repository: https://github.com/SAP/jenkins-library.git
```

Or in Jenkinsfile:
```groovy
@Library('piper-lib-os@master') _
// Pin version: @Library('piper-lib-os@v1.120.0') _
```

### Step 2: Configure Docker (Optional)

```bash
# Ubuntu/Debian
sudo apt-get install docker.io
sudo usermod -aG docker jenkins
sudo systemctl restart jenkins
```

### Step 3: Configure Vault (Optional)

Create Jenkins credentials for Vault AppRole:
- RoleId: Secret text credential
- SecretId: Secret text credential

## Configuration

### Project Structure
```
my-project/
├── .pipeline/
│   ├── config.yml
│   └── extensions/
│       └── Build.groovy
├── Jenkinsfile
└── src/
```

### Basic Jenkinsfile
```groovy
@Library('piper-lib-os') _

piperPipeline script: this
```

### Configuration File (.pipeline/config.yml)
```yaml
general:
  buildTool: 'maven'  # or 'npm', 'mta', 'golang'
  productiveBranch: 'main'
  verbose: false

stages:
  Build:
    npmExecuteLint: true
  Security:
    whitesourceScan: true
  Performance:
    activate: false

steps:
  mavenBuild:
    goals: ['clean', 'install']
  cloudFoundryDeploy:
    cloudFoundry:
      credentialsId: 'cf-credentials'
      apiEndpoint: 'https://api.cf.example.com'
      org: 'my-org'
      space: 'dev'
```

## Common Patterns

### Maven Build
```groovy
@Library('piper-lib-os') _
piperPipeline script: this
```

```yaml
# .pipeline/config.yml
general:
  buildTool: 'maven'
steps:
  mavenBuild:
    goals: ['clean', 'install']
```

### Node.js Build
```yaml
general:
  buildTool: 'npm'
steps:
  npmExecuteScripts:
    runScripts: ['build', 'test']
```

### Multi-Cloud Deployment
```yaml
stages:
  Release:
    cloudFoundryDeploy: true
    kubernetesDeploy: true

steps:
  cloudFoundryDeploy:
    cloudFoundry:
      credentialsId: 'cf-prod'
  kubernetesDeploy:
    containerRegistryUrl: 'https://docker.example.com'
    namespace: 'production'
```

### Docker-Based Build
```yaml
steps:
  mavenBuild:
    dockerImage: 'maven:3.8-jdk-11'
    dockerOptions: '--volume /tmp:/tmp'
```

### Parallel Testing
```groovy
@Library('piper-lib-os') _

node() {
  stage('Init') {
    checkout scm
    setupCommonPipelineEnvironment script: this
  }

  stage('Build') {
    mavenBuild script: this
  }

  stage('Test') {
    parallel(
      'Unit Tests': {
        mavenExecute script: this, goals: ['test']
      },
      'Integration Tests': {
        mavenExecute script: this, goals: ['integration-test']
      }
    )
  }
}
```

### Vault Secrets
```yaml
general:
  vaultServerUrl: 'https://vault.example.com'
  vaultNamespace: 'piper'
  vaultBasePath: 'piper/my-project'
```

```groovy
withCredentials([
  string(credentialsId: 'vault-role-id', variable: 'VAULT_ROLE_ID'),
  string(credentialsId: 'vault-secret-id', variable: 'VAULT_SECRET_ID')
]) {
  cloudFoundryDeploy script: this
}
```

## Examples

### MTA Application
```yaml
general:
  buildTool: 'mta'
steps:
  mtaBuild:
    buildTarget: 'CF'
  cloudFoundryDeploy:
    mtaDeployParameters: '-f'
    cloudFoundry:
      credentialsId: 'cf-deploy'
      apiEndpoint: 'https://api.cf.us10.hana.ondemand.com'
```

### Custom Pipeline
```groovy
@Library('piper-lib-os') _

node() {
  stage('Checkout') {
    checkout scm
    setupCommonPipelineEnvironment script: this
  }

  stage('Build') {
    mavenBuild script: this
  }

  stage('Custom Step') {
    sh 'npm run custom-validation'
  }

  stage('Deploy') {
    cloudFoundryDeploy script: this
  }
}
```

## Troubleshooting

### Library Not Found
**Error:** `Library piper-lib-os not found`

**Solution:**
- Verify library in **Manage Jenkins** → **Configure System**
- Check GitHub connectivity
- Verify branch/tag exists

### Docker Permission Denied
**Error:** `permission denied while connecting to Docker daemon`

**Solution:**
```bash
sudo usermod -aG docker jenkins
sudo systemctl restart jenkins
```

### Credentials Not Found
**Error:** `Could not find credentials entry with ID 'xxx'`

**Solution:**
1. Add credential in **Jenkins** → **Credentials**
2. Match exact ID in config
3. Verify credential type

### Stage Not Executing
**Solution:**
```yaml
stages:
  Security:
    activate: true  # Explicit activation
general:
  productiveBranch: 'main'  # Check branch
```

### Docker Pull Failures
**Solution:**
```yaml
steps:
  dockerExecute:
    dockerImage: 'my-registry.com/image:tag'
    dockerPullImage: true
    dockerRegistryCredentialsId: 'docker-registry'
```

### Out of Memory
**Solution:**
```yaml
steps:
  mavenBuild:
    dockerOptions: '-m 4g'
    defines: ['-Xmx3072m']
```

### Vault Connection Issues
**Solution:**
```yaml
general:
  vaultServerUrl: 'https://vault.example.com:8200'
  vaultNamespace: 'my-namespace'
  vaultPath: 'secret/path'
```

### Debug Configuration
```groovy
node() {
  stage('Debug') {
    checkout scm
    setupCommonPipelineEnvironment script: this
    echo "Config: ${script.commonPipelineEnvironment.configuration}"
  }
}
```

Enable verbose:
```yaml
general:
  verbose: true
```

## Best Practices

1. **Version Pinning**: Pin library versions for production
2. **Caching**: Use Maven repository caching
3. **Resource Management**: Clean workspace and containers
4. **Credentials**: Use Jenkins credentials or Vault
5. **Documentation**: Document custom extensions
6. **Testing**: Test changes in feature branches
7. **Monitoring**: Set up build notifications

## Additional Resources

- [Project Piper Documentation](https://www.project-piper.io/)
- [Jenkins Pipeline Docs](https://www.jenkins.io/doc/book/pipeline/)
- [Piper Steps](https://www.project-piper.io/steps/)
- [Pipeline Stages](https://www.project-piper.io/stages/)

# Troubleshooting Guide

Common issues and solutions for Project Piper.

## Table of Contents

- [General](#general)
- [Configuration](#configuration)
- [Build](#build)
- [Deployment](#deployment)
- [Container & Docker](#container--docker)
- [Security](#security)
- [Authentication](#authentication)

---

## General

### Library not found
**Error**: `@Library('piper-lib-os')` cannot be resolved

**Fix**:
- Verify library in Jenkins: Manage Jenkins → Configure System → Global Pipeline Libraries
- Repository: `https://github.com/SAP/jenkins-library.git`
- Use version: `@Library('piper-lib-os@master') _`

### Step not found
**Error**: `No such DSL method 'stepName'`

**Fix**:
- Check spelling (case-sensitive)
- Use: `@Library('piper-lib-os') _` (note underscore)
- Verify step in [step-reference.md](step-reference.md)

### Missing script parameter
**Error**: `The 'script' parameter is required`

**Fix**: Always include `script: this`
```groovy
mavenBuild script: this
```

### Config file not found
**Error**: `Could not find .pipeline/config.yml`

**Fix**:
- Path must be: `.pipeline/config.yml` at repo root
- Verify committed: `git status`
- Ensure workspace checked out

---

## Configuration

### Invalid YAML
**Error**: Error parsing config.yml

**Fix**:
- Validate: `python -c "import yaml; yaml.safe_load(open('.pipeline/config.yml'))"`
- Use spaces (not tabs)
- Consistent indentation (2 spaces)
```yaml
steps:
  mavenBuild:
    goals: 'install'
```

### Config not applied
**Problem**: Step ignores config.yml

**Fix**:
- Check hierarchy: Step params > Stage > Step config > General
- Verify proper YAML nesting

### Vault connection fails
**Error**: Cannot fetch from Vault

**Fix**:
```yaml
general:
  vaultServerUrl: 'https://vault.example.com:8200'
  vaultBasePath: 'piper/my-project'
```
- Test: `curl -k https://vault.example.com:8200/v1/sys/health`
- Verify AppRole credentials

---

## Build

### Maven fails
**Solutions**:
```yaml
steps:
  mavenBuild:
    dockerImage: 'maven:3.8-jdk-11'  # Match Java version
    projectSettingsFile: 'settings.xml'  # Custom settings
    dockerOptions: ['-e', 'MAVEN_OPTS=-Xmx2048m']  # More memory
```

### npm fails
**Solutions**:
```yaml
steps:
  npmExecuteScripts:
    dockerImage: 'node:18'
    defaultNpmRegistry: 'https://registry.npmjs.org/'
```
- Clean: `rm -rf node_modules package-lock.json`

### MTA fails
**Solutions**:
- Validate mta.yaml locally
- Increase memory:
```yaml
steps:
  mtaBuild:
    dockerOptions: ['-m', '4g']
```

### Gradle fails
**Solutions**:
```yaml
steps:
  gradleExecuteBuild:
    dockerImage: 'gradle:7.6-jdk11'
```
```groovy
gradleExecuteBuild(
    script: this,
    options: '--no-daemon --stacktrace'
)
```

---

## Deployment

### Cloud Foundry fails
**Solutions**:
```groovy
cloudFoundryDeploy(
    script: this,
    cloudFoundry: [
        apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com',
        credentialsId: 'cf-credentials',
        org: 'my-org',
        space: 'dev'
    ]
)
```
- Check manifest.yml syntax
- Verify quota: `cf space <space-name>`
- Use `random-route: true` for conflicts

### Kubernetes fails
**Solutions**:
```yaml
steps:
  kubernetesDeploy:
    kubeConfig: 'kubeconfig'
    namespace: 'my-namespace'
```
- Create namespace: `kubectl create namespace my-namespace`
- Check image pull secrets

### Blue-green issues
**CF Standard**:
```groovy
cloudFoundryDeploy(
    script: this,
    deployType: 'blue-green'
)
```

**MTA**:
```groovy
cloudFoundryDeploy(
    script: this,
    deployTool: 'mtaDeployPlugin',
    deployType: 'bg-deploy'
)
```

### CF CLI not found
**Fix**: Use Piper container:
```yaml
steps:
  cloudFoundryDeploy:
    dockerImage: 'ppiper/cf-cli:latest'
```

### Service creation fails
**Solutions**:
- Check: `cf marketplace`
- Wait for async:
```groovy
cloudFoundryCreateService(
    script: this,
    timeout: 30
)
```

---

## Container & Docker

### Docker daemon unavailable
**Error**: Cannot connect to Docker daemon

**Fix**: Use Kaniko (no daemon needed):
```groovy
kanikoExecute(
    script: this,
    containerImageName: 'my-app',
    containerImageTag: '1.0.0'
)
```
- Or add Jenkins user to docker group: `usermod -aG docker jenkins`

### Image pull fails
**Solutions**:
```yaml
steps:
  containerPushToRegistry:
    dockerRegistryUrl: 'https://my-registry.com'
    dockerCredentialsId: 'registry-creds'
general:
  httpsProxy: 'http://proxy.example.com:8080'
```
- Use authentication to avoid rate limits

### Build fails
**Solutions**:
- Test locally: `docker build -t test .`
- Add .dockerignore:
```
node_modules
.git
*.log
```
- Pass args:
```groovy
kanikoExecute(
    script: this,
    buildOptions: ['--build-arg', 'VERSION=1.0.0']
)
```

### Hadolint errors
**Fix**:
```yaml
steps:
  hadolintExecute:
    ignoreRules: ['DL3008', 'DL3018']
```
- Specify package versions
- Use COPY not ADD
- Don't use latest tag

---

## Security

### SonarQube fails
**Solutions**:
```yaml
steps:
  sonarExecuteScan:
    serverUrl: 'https://sonarqube.example.com'
    sonarTokenCredentialsId: 'sonar-token'
    projectKey: 'my-project'
    options: ['-Dsonar.coverage.jacoco.xmlReportPaths=target/site/jacoco/jacoco.xml']
    dockerOptions: ['-e', 'SONAR_SCANNER_OPTS=-Xmx2048m']
```

### Checkmarx fails
**Solutions**:
```yaml
steps:
  checkmarxExecuteScan:
    serverUrl: 'https://checkmarx.example.com'
    credentialsId: 'checkmarx-creds'
    scanPollingTimeout: 30
    fileExclude: '*.spec.js,test/**,node_modules/**'
```

### Fortify fails
**Fix**:
```yaml
steps:
  fortifyExecuteScan:
    serverUrl: 'https://fortify.example.com/ssc'
    authToken: '${FORTIFY_TOKEN}'
    buildId: 'my-project-build'
```

### WhiteSource/Mend fails
**Fix**:
```yaml
steps:
  whitesourceExecuteScan:
    orgToken: '${WHITESOURCE_ORG_TOKEN}'
    userTokenCredentialsId: 'ws-user-token'
    productName: 'My Product'
    timeout: 60
```

---

## Authentication

### Credentials not found
**Error**: `Credentials 'xxx' not found`

**Fix**:
- Verify ID in Jenkins (case-sensitive)
- Check scope (System/Global)
- Correct type:
  - Username/Password → `usernamePassword()`
  - Secret Text → `string()`
  - Secret File → `file()`

### Vault auth fails
**Fix**:
```yaml
general:
  vaultAppRoleID: 'my-role-id'
  vaultAppRoleSecretTokenCredentialsId: 'vault-secret-id'
```
- Or token:
```yaml
general:
  vaultTokenCredentialsId: 'vault-token'
```
- Verify Vault policies

### SSH key issues
**Fix**:
- Use HTTPS instead
- Add SSH key: Credentials → SSH Username with private key
- Accept host: `ssh-keyscan github.com >> ~/.ssh/known_hosts`

---

## Performance

### Slow pipeline
**Solutions**:
- Use parallel:
```groovy
parallel {
    stage('Tests') { steps { /* ... */ } }
    stage('Security') { steps { /* ... */ } }
}
```
- Use specific tags (not `latest`)
- Use alpine images
- Enable caching:
```yaml
steps:
  npmExecuteScripts:
    cache: true
```

### Slow Docker pulls
**Fix**:
```yaml
general:
  dockerRegistryUrl: 'https://mirror.gcr.io'
```
- Pre-pull on agents
- Use authenticated pulls

### Cache not working
**Maven**:
```yaml
steps:
  mavenBuild:
    m2Path: '/home/jenkins/.m2'
```

**npm**:
```yaml
steps:
  npmExecuteScripts:
    cache: true
```

**Kaniko**:
```groovy
kanikoExecute(
    script: this,
    buildOptions: ['--cache=true', '--cache-ttl=24h']
)
```

---

## ABAP

### Connection fails
**Fix**:
```yaml
steps:
  abapEnvironmentRunAUnitTest:
    host: 'https://my-system.abap.eu10.hana.ondemand.com'
    credentialsId: 'abap-credentials'
```
- Check connectivity
- Verify authorizations

### ATC fails
**Fix**:
```yaml
steps:
  abapEnvironmentRunATCCheck:
    failOnSeverity: 'error'
```

### gCTS fails
**Fix**:
```yaml
steps:
  gctsDeploy:
    host: 'https://my-system.example.com'
    client: '001'
    repository: 'my-repo'
```
- Check branch config
- Verify transports

---

## Getting Help

- **Docs**: [https://sap.github.io/jenkins-library/](https://sap.github.io/jenkins-library/)
- **Issues**: [https://github.com/SAP/jenkins-library/issues](https://github.com/SAP/jenkins-library/issues)
- **Group**: [https://groups.google.com/forum/#!forum/project-piper](https://groups.google.com/forum/#!forum/project-piper)
- **Debug**:
```groovy
debugReportArchive script: this
```

---

*Last Updated: 2025-11-14*

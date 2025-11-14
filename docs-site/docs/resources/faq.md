# Frequently Asked Questions (FAQ)

Common questions about Project Piper and the Jenkins library.

## Table of Contents

- [General](#general)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Build & Deployment](#build--deployment)
- [Security](#security)
- [Cloud Platforms](#cloud-platforms)
- [ABAP](#abap)

---

## General

### What is Project Piper?

Project Piper provides ready-made CI/CD pipelines for SAP technologies:
- Pre-configured Jenkins pipelines
- Reusable pipeline steps library
- Best practices for SAP systems
- Integration with SAP BTP and cloud providers

### Is Project Piper still maintained?

**Important**: Project Piper is no longer accepting new contributions. Existing functionality remains available. Check the GitHub repository for latest status.

### What technologies are supported?

- **Languages**: Java, JavaScript/Node.js, Python, Golang, ABAP
- **Build Tools**: Maven, npm, Gradle, MTA, pip
- **Platforms**: Cloud Foundry, Kubernetes, SAP BTP, Neo
- **Security**: Checkmarx, Fortify, SonarQube, Snyk, Mend, BDBA
- **Containers**: Docker, Kaniko, Cloud Native Buildpacks

### Where is the documentation?

- **Main Docs**: [https://sap.github.io/jenkins-library/](https://sap.github.io/jenkins-library/)
- **GitHub**: [https://github.com/SAP/jenkins-library](https://github.com/SAP/jenkins-library)
- **Step Reference**: [step-reference.md](step-reference.md)
- **Google Group**: [https://groups.google.com/forum/#!forum/project-piper](https://groups.google.com/forum/#!forum/project-piper)

---

## Getting Started

### How do I install Piper?

**Jenkins**:
```groovy
@Library('piper-lib-os') _
```

**CLI**:
```bash
curl -L https://github.com/SAP/jenkins-library/releases/latest/download/piper_linux_amd64 -o piper
chmod +x piper
```

**GitHub Actions**:
```yaml
- uses: SAP/project-piper-action@master
```

### What are the prerequisites?

- **Jenkins**: 2.60+ with Pipeline plugin
- **CLI/Actions**: Docker for containerized steps
- **All**: Appropriate credentials configured

### How do I create a basic pipeline?

```groovy
@Library('piper-lib-os') _

pipeline {
    agent any
    stages {
        stage('Build') {
            steps {
                mavenBuild script: this
            }
        }
        stage('Test') {
            steps {
                karmaExecuteTests script: this
            }
        }
        stage('Deploy') {
            steps {
                cloudFoundryDeploy script: this
            }
        }
    }
}
```

### Can I use Piper outside Jenkins?

Yes! Piper provides:
- **CLI**: Run steps from command line
- **GitHub Actions**: Use in workflows
- **Azure Pipelines**: Use Azure extension

---

## Configuration

### Where do I put configuration?

Use `.pipeline/config.yml` at repository root:
```yaml
general:
  productiveBranch: 'main'

steps:
  mavenBuild:
    dockerImage: 'maven:3.8-jdk-11'

stages:
  Build:
    mavenBuild:
      goals: 'clean install'
```

### What is the configuration hierarchy?

Precedence (highest to lowest):
1. Step parameters (in Jenkinsfile)
2. Stage configuration
3. Step configuration
4. General configuration
5. Default values

### How do I manage secrets?

Piper supports:
- **Jenkins Credentials**: Use `credentialsId` parameters
- **HashiCorp Vault**: Configure in config.yml
- **Cloud Secrets**: Azure Key Vault, AWS Secrets Manager
- **Environment Variables**: For CLI/GitHub Actions

Vault example:
```yaml
general:
  vaultBasePath: 'piper/my-project'
  vaultServerUrl: 'https://vault.example.com'
```

### Can I override Docker images?

Yes:
```yaml
steps:
  mavenBuild:
    dockerImage: 'maven:3.9-jdk-17'
  npmExecute:
    dockerImage: 'node:18-alpine'
```

---

## Build & Deployment

### Which build tools are supported?

- **Maven**: `mavenBuild`, `mavenExecute`
- **npm/yarn/pnpm**: `npmExecuteScripts`
- **Gradle**: `gradleExecuteBuild`
- **Golang**: `golangBuild`
- **Python**: `pythonBuild`
- **MTA**: `mtaBuild`
- **CNB**: `cnbBuild`

### How do I build an MTA?

```groovy
mtaBuild(
    script: this,
    buildTarget: 'CF'
)
```

### How do I deploy to Cloud Foundry?

```groovy
cloudFoundryDeploy(
    script: this,
    cloudFoundry: [
        org: 'my-org',
        space: 'dev',
        credentialsId: 'cf-creds'
    ]
)
```

### How do I do blue-green deployments?

```groovy
cloudFoundryDeploy(
    script: this,
    deployType: 'blue-green'
)
```

### How do I version artifacts?

```groovy
artifactPrepareVersion(
    script: this,
    buildTool: 'maven',
    versioningType: 'cloud'
)
```

---

## Security

### Which security tools are supported?

- **SAST**: Checkmarx, Fortify, SonarQube, CodeQL
- **SCA**: Mend (WhiteSource), Snyk, BlackDuck
- **Container**: BDBA (Protecode)
- **Credentials**: Credential Digger
- **DAST**: Contrast
- **Malware**: SAP Malware Scanning Service

### How do I run SonarQube scans?

```groovy
sonarExecuteScan(
    script: this,
    serverUrl: 'https://sonarqube.example.com',
    projectKey: 'my-project'
)
```

### How do I handle scan failures?

Configure thresholds:
```yaml
steps:
  checkmarxExecuteScan:
    vulnerabilityThresholdHigh: 10
    vulnerabilityThresholdMedium: 50
```

### Can I run multiple scans?

Yes, use parallel stages:
```groovy
stage('Security') {
    parallel {
        stage('SAST') {
            steps { sonarExecuteScan script: this }
        }
        stage('SCA') {
            steps { whitesourceExecuteScan script: this }
        }
    }
}
```

---

## Cloud Platforms

### How do I use manifest variables in CF?

```groovy
cloudFoundryDeploy(
    script: this,
    manifest: 'manifest.yml',
    manifestVariables: ['key1=value1', 'key2=value2'],
    manifestVariablesFiles: ['vars-dev.yml']
)
```

### How do I create CF services?

```groovy
cloudFoundryCreateService(
    script: this,
    cloudFoundry: [...],
    serviceManifest: 'services.yml'
)
```

### Can I deploy Docker images to CF?

Yes:
```groovy
cloudFoundryDeploy(
    script: this,
    deployDockerImage: 'my-registry/app:1.0.0',
    dockerUsername: 'user',
    dockerPassword: credentials('docker-pass')
)
```

### How do I build Docker images?

**Kaniko (no daemon)**:
```groovy
kanikoExecute(
    script: this,
    containerImageName: 'my-app',
    containerImageTag: '1.0.0'
)
```

**Cloud Native Buildpacks**:
```groovy
cnbBuild(
    script: this,
    containerImageName: 'my-app'
)
```

### How do I deploy to Kubernetes?

```groovy
kubernetesDeploy(
    script: this,
    kubeConfig: 'kubeconfig',
    namespace: 'production',
    deploymentManifest: 'deployment.yaml'
)
```

### How do I use Helm?

```groovy
helmExecute(
    script: this,
    helmCommand: 'upgrade',
    chartPath: './chart',
    namespace: 'production'
)
```

---

## ABAP

### What ABAP scenarios are supported?

- **gCTS**: Git-enabled Change and Transport System
- **ABAP Environment**: SAP BTP ABAP Environment (Steampunk)
- **Add-ons**: AAKaaS integration
- **Quality**: ATC and ABAP Unit tests

### How do I run ABAP Unit tests?

```groovy
abapEnvironmentRunAUnitTest(
    script: this,
    host: 'https://my-system.abap.eu10.hana.ondemand.com',
    credentialsId: 'abap-creds'
)
```

### How do I deploy via gCTS?

```groovy
gctsDeploy(
    script: this,
    host: 'https://my-system.example.com',
    client: '001',
    repository: 'my-repo'
)
```

### How do I run ATC checks?

```groovy
abapEnvironmentRunATCCheck(
    script: this,
    host: 'https://my-system.abap.eu10.hana.ondemand.com'
)
```

---

## Troubleshooting

### Why is my step failing?

1. Check error message in Jenkins console
2. Enable verbose logging
3. Archive debug reports:
```groovy
debugReportArchive script: this
```

### Config file not found?

- Must be at: `.pipeline/config.yml`
- Check YAML syntax
- Ensure file is committed
- Verify workspace checkout

### Credentials not working?

- Verify ID matches exactly (case-sensitive)
- Check credential type matches requirement
- Ensure proper scope (System/Global)
- Verify user has access permissions

### Docker not available?

Use Kaniko for Docker-less builds:
```groovy
kanikoExecute script: this
```

### How do I debug?

- Review Jenkins console output
- Check step-specific logs
- Use verbose mode
- Archive debug reports
- Search GitHub issues

---

## Additional Resources

- **Troubleshooting**: [troubleshooting.md](troubleshooting.md)
- **Glossary**: [glossary.md](glossary.md)
- **Step Reference**: [step-reference.md](step-reference.md)
- **Main Docs**: [https://sap.github.io/jenkins-library/](https://sap.github.io/jenkins-library/)

---

*Last Updated: 2025-11-14*

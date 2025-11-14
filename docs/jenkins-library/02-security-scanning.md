# Jenkins Library - Security Scanning

This document covers all security scanning integrations available in jenkins-library.

## Overview

jenkins-library integrates with 12 major security scanning tools, providing comprehensive security coverage:

- **SAST** (Static Application Security Testing): Checkmarx, Fortify, CodeQL
- **SCA** (Software Composition Analysis): WhiteSource, Black Duck, Snyk
- **Container Security**: Protecode, Aqua, Snyk Container
- **Secret Detection**: Credentialdigger
- **Malware Scanning**: Malware scan
- **Runtime Security**: Contrast Security

---

## Security Scanning Steps

| Step | Tool | Type | Focus |
|------|------|------|-------|
| `checkmarxExecuteScan` | Checkmarx | SAST | Source code vulnerabilities |
| `checkmarxOneExecuteScan` | Checkmarx One | SAST | Next-gen Checkmarx |
| `fortifyExecuteScan` | Micro Focus Fortify | SAST | Enterprise SAST |
| `codeqlExecuteScan` | GitHub CodeQL | SAST | Semantic code analysis |
| `whitesourceExecuteScan` | WhiteSource/Mend | SCA | Open source dependencies |
| `detectExecuteScan` | Synopsys Detect (Black Duck) | SCA | OSS & container scanning |
| `snykExecute` | Snyk | SCA | Developer-first security |
| `protecodeExecuteScan` | Protecode/BDBA | Container | Container image scanning |
| `contrastExecuteScan` | Contrast Security | IAST/RASP | Runtime analysis |
| `credentialdiggerScan` | Credentialdigger | Secret | Hardcoded secrets |
| `malwareExecuteScan` | Malware Scanner | Malware | Malware detection |
| `sonarExecuteScan` | SonarQube | Quality/Security | Code quality + security |

---

## Checkmarx (SAST)

### Step: `checkmarxExecuteScan`

**Location**: `vars/checkmarxExecuteScan.groovy`, `cmd/checkmarxExecuteScan.go`

Performs static application security testing using Checkmarx CxSAST.

### Basic Usage

```groovy
checkmarxExecuteScan script: this
```

### Configuration

```yaml
steps:
  checkmarxExecuteScan:
    # Server configuration
    serverUrl: 'https://checkmarx.example.com'
    username: 'checkmarx-user'
    password: 'vault:checkmarx:password'

    # Project configuration
    projectName: 'MyProject'
    teamName: '/CxServer/MyTeam'
    preset: 'Checkmarx Default'

    # Scan configuration
    sourceEncoding: 'UTF-8'
    filterPattern: '!**/*.spec.js, !**/node_modules/**'

    # Vulnerability thresholds
    vulnerabilityThresholdHigh: 0
    vulnerabilityThresholdMedium: 10
    vulnerabilityThresholdLow: 100

    # Per-query thresholds
    vulnerabilityThresholdLowPerQuery: true
    vulnerabilityThresholdLowPerQueryMax: 5

    # Reporting
    generatePdfReport: true
    checkMarxProjectPreviousVersioning: true

    # Incremental scan
    incremental: true
    fullScansScheduled: true
    fullScanCycle: 7  # Full scan every 7 days
```

### Key Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `serverUrl` | string | Checkmarx server URL | - |
| `projectName` | string | Project name in Checkmarx | Auto-detected |
| `teamName` | string | Team path | `/CxServer` |
| `preset` | string | Scan preset | `Checkmarx Default` |
| `incremental` | bool | Incremental scan | `true` |
| `vulnerabilityThresholdHigh` | int | Max high vulnerabilities | `100` |
| `generatePdfReport` | bool | Generate PDF report | `true` |

### Features

- **Incremental scans**: Fast incremental + periodic full scans
- **Custom presets**: Use organization-specific presets
- **File filtering**: Exclude test files, dependencies
- **Per-query thresholds**: Fine-grained vulnerability control
- **PDF reports**: Executive-friendly reports
- **Project versioning**: Track scans over time

---

## Checkmarx One (Next-Gen SAST)

### Step: `checkmarxOneExecuteScan`

**Location**: `vars/checkmarxOneExecuteScan.groovy`, `cmd/checkmarxOneExecuteScan.go`

Uses Checkmarx One (next-generation platform).

### Configuration

```yaml
steps:
  checkmarxOneExecuteScan:
    serverUrl: 'https://checkmarxone.example.com'
    clientId: 'client-id'
    clientSecret: 'vault:checkmarxone:secret'

    projectName: 'MyProject'
    preset: 'ASA Premium'

    # Scan types
    scanTypes: ['sast', 'sca', 'iac']

    # Thresholds
    vulnerabilityThresholdHigh: 0
    vulnerabilityThresholdMedium: 5
```

### Features

- **Multi-scan**: SAST, SCA, IaC in one scan
- **Cloud-native**: Modern cloud platform
- **API-first**: RESTful API integration

---

## Fortify (Enterprise SAST)

### Step: `fortifyExecuteScan`

**Location**: `vars/fortifyExecuteScan.groovy`, `cmd/fortifyExecuteScan.go`

Enterprise-grade SAST using Micro Focus Fortify.

### Configuration

```yaml
steps:
  fortifyExecuteScan:
    # Server configuration
    serverUrl: 'https://fortify.example.com'
    authToken: 'vault:fortify:token'

    # Project configuration
    projectName: 'MyProject'
    projectVersionId: '12345'

    # Build configuration
    buildId: 'MyBuild'
    buildProject: 'MyProject'

    # Python-specific
    pythonRequirementsFile: 'requirements.txt'
    pythonInstallCommand: 'pip install -r requirements.txt'
    pythonAdditionalPath: ['src', 'lib']

    # Reporting
    generatePdfReport: true
    reportType: 'PDF'

    # Quality gate
    spotCheckMinimum: 90
    mustAuditIssueGroups: 'Corporate Security Requirements'
```

### Key Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `serverUrl` | string | Fortify SSC URL | - |
| `projectName` | string | Project name | - |
| `buildId` | string | Build identifier | Auto-generated |
| `spotCheckMinimum` | int | Minimum spot check % | `50` |
| `generatePdfReport` | bool | Generate PDF | `true` |

### Features

- **SSC integration**: Fortify Software Security Center
- **Python support**: Enhanced Python analysis
- **Audit workbench**: Integration with audit requirements
- **Compliance**: Corporate security policy enforcement
- **PDF reports**: Executive reporting

---

## GitHub CodeQL (Semantic Analysis)

### Step: `codeqlExecuteScan`

**Location**: `vars/codeqlExecuteScan.groovy`, `cmd/codeqlExecuteScan.go`

Semantic code analysis using GitHub CodeQL.

### Configuration

```yaml
steps:
  codeqlExecuteScan:
    # Language
    language: 'java'  # or 'javascript', 'python', 'go', etc.

    # Database
    buildCommand: 'mvn clean install'

    # Queries
    querySuite: 'security-extended'

    # GitHub configuration
    githubToken: 'vault:github:token'
    repository: 'myorg/myrepo'

    # Upload to GitHub
    uploadResults: true

    # Vulnerability threshold
    vulnerabilityThreshold: 0
```

### Supported Languages

- Java/Kotlin
- JavaScript/TypeScript
- Python
- C/C++
- C#
- Go
- Ruby

### Features

- **Semantic analysis**: Deep code understanding
- **Custom queries**: Write organization-specific queries
- **GitHub integration**: Native GitHub Security tab
- **SARIF output**: Standard security format
- **Multi-language**: Support for 7+ languages

---

## WhiteSource/Mend (SCA)

### Step: `whitesourceExecuteScan`

**Location**: `vars/whitesourceExecuteScan.groovy`, `cmd/whitesourceExecuteScan.go`

Software Composition Analysis for open source dependencies.

### Configuration

```yaml
steps:
  whitesourceExecuteScan:
    # Server configuration
    serverUrl: 'https://saas.whitesourcesoftware.com/api'
    orgToken: 'vault:whitesource:org-token'
    userToken: 'vault:whitesource:user-token'

    # Project configuration
    productName: 'MyProduct'
    projectName: 'MyProject'

    # Vulnerability policy
    cvssSeverityLimit: 7.0
    vulnerabilityRiskScoreThreshold: 50

    # License policy
    licensingVulnerabilities: true

    # Reporting
    reporting: true
    vulnerabilityReportFileName: 'whitesource-report'

    # Build tool specific
    buildTool: 'maven'  # Auto-detected
    buildDescriptorFile: 'pom.xml'
```

### Key Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `serverUrl` | string | WhiteSource API URL | WhiteSource SaaS |
| `productName` | string | Product name | - |
| `cvssSeverityLimit` | float | Min CVSS score | `7.0` |
| `vulnerabilityRiskScoreThreshold` | int | Risk score threshold | `50` |
| `licensingVulnerabilities` | bool | Check license violations | `true` |

### Features

- **OSS inventory**: Complete dependency inventory
- **Vulnerability detection**: Known CVEs in dependencies
- **License compliance**: License policy enforcement
- **Auto-remediation**: Suggested fixes
- **Supply chain**: Transitive dependency analysis

---

## Synopsys Detect / Black Duck (SCA)

### Step: `detectExecuteScan`

**Location**: `vars/detectExecuteScan.groovy`, `cmd/detectExecuteScan.go`

Comprehensive SCA using Synopsys Detect and Black Duck.

### Configuration

```yaml
steps:
  detectExecuteScan:
    # Server configuration
    serverUrl: 'https://blackduck.example.com'
    apiToken: 'vault:blackduck:token'

    # Project configuration
    projectName: 'MyProject'
    projectVersion: '1.0.0'

    # Scan configuration
    scanners: ['SIGNATURE']
    scanPaths: ['.']
    includedPackageManagers: ['MAVEN', 'NPM']

    # Vulnerability policy
    failOn: ['BLOCKER', 'CRITICAL']

    # Code transparency (CTP)
    codeTransparencyMode: false
```

### Scan Types

- **SIGNATURE**: Binary/source signature scanning
- **DETECTOR**: Package manager detection
- **IAC**: Infrastructure as Code scanning

### Features

- **Multi-package manager**: 30+ package managers
- **Container scanning**: Docker image analysis
- **Snippet matching**: Partial code matches
- **CTP mode**: Code Transparency Protocol
- **Policy management**: Custom security policies

---

## Snyk (Developer-First Security)

### Step: `snykExecute`

**Location**: `vars/snykExecute.groovy`, `cmd/snykExecute.go`

Developer-friendly security scanning.

### Configuration

```yaml
steps:
  snykExecute:
    # Authentication
    snykAuthToken: 'vault:snyk:token'

    # Organization
    snykOrg: 'my-organization'

    # Scan type
    scanType: 'test'  # 'test' or 'monitor'

    # Severity threshold
    severityThreshold: 'high'  # 'low', 'medium', 'high', 'critical'

    # Build tool
    buildTool: 'maven'  # Auto-detected

    # Additional flags
    additionalArguments: ['--all-projects']
```

### Scan Types

- **test**: Test and report vulnerabilities
- **monitor**: Continuous monitoring

### Features

- **Fast scans**: Quick developer feedback
- **CLI integration**: Local development testing
- **Fix suggestions**: Automated PR generation
- **Container scanning**: Docker image vulnerabilities
- **IaC scanning**: Terraform, Kubernetes manifests

---

## Protecode / BDBA (Container Security)

### Step: `protecodeExecuteScan`

**Location**: `vars/protecodeExecuteScan.groovy`, `cmd/protecodeExecuteScan.go`

Container image scanning using Protecode (Black Duck Binary Analysis).

### Configuration

```yaml
steps:
  protecodeExecuteScan:
    # Server configuration
    protecodeServerUrl: 'https://protecode.example.com'
    credentialsId: 'protecode-credentials'

    # Scan configuration
    scanImage: 'myregistry.com/myapp:1.0.0'
    dockerRegistryUrl: 'https://myregistry.com'

    # Policy
    failOnSevereVulnerabilities: true
    vulnerabilitySeverityThreshold: 7.0

    # Timeout
    timeoutMinutes: 60
```

### Features

- **Container scanning**: Docker, OCI images
- **Layer analysis**: Per-layer vulnerability detection
- **Binary analysis**: Without source code
- **License compliance**: Container license detection

---

## Contrast Security (IAST/RASP)

### Step: `contrastExecuteScan`

**Location**: `vars/contrastExecuteScan.groovy`, `cmd/contrastExecuteScan.go`

Interactive/Runtime Application Security Testing.

### Configuration

```yaml
steps:
  contrastExecuteScan:
    # Server configuration
    serverUrl: 'https://app.contrastsecurity.com/Contrast'
    apiKey: 'vault:contrast:api-key'
    serviceKey: 'vault:contrast:service-key'
    username: 'contrast-user'

    # Application
    applicationId: 'app-id-from-contrast'
    applicationName: 'MyApplication'

    # Vulnerability threshold
    vulnerabilityThreshold: 0
```

### Features

- **Runtime analysis**: Tests running application
- **IAST**: Interactive application security testing
- **RASP**: Runtime application self-protection
- **Real vulnerabilities**: No false positives
- **API security**: REST API vulnerability detection

---

## Credentialdigger (Secret Detection)

### Step: `credentialdiggerScan`

**Location**: `vars/credentialdiggerScan.groovy`, `cmd/credentialdiggerScan.go`

Detects hardcoded credentials in source code.

### Configuration

```yaml
steps:
  credentialdiggerScan:
    # Repository
    repository: 'https://github.com/myorg/myrepo'

    # Scan configuration
    snapshot: true

    # Export
    export: 'credentials-report.csv'
```

### Features

- **Git history**: Scans entire Git history
- **Pattern matching**: Known credential patterns
- **False positive reduction**: Machine learning
- **Export**: CSV/JSON reports

---

## Malware Scanning

### Step: `malwareExecuteScan`

**Location**: `vars/malwareExecuteScan.groovy`, `cmd/malwareExecuteScan.go`

Scans artifacts for malware.

### Configuration

```yaml
steps:
  malwareExecuteScan:
    # Target
    scanPath: 'target/'

    # Service
    malwareScannerUrl: 'https://malware-scanner.example.com'
```

---

## SonarQube (Quality + Security)

### Step: `sonarExecuteScan`

**Location**: `vars/sonarExecuteScan.groovy`, `cmd/sonarExecuteScan.go`

Code quality and security analysis.

### Configuration

```yaml
steps:
  sonarExecuteScan:
    # Server configuration
    serverUrl: 'https://sonarqube.example.com'
    token: 'vault:sonar:token'

    # Project configuration
    projectKey: 'my-project'
    projectName: 'My Project'

    # Branch analysis
    branchName: 'main'
    pullRequestProvider: 'GitHub'

    # Coverage
    coverageExclusions: ['**/test/**', '**/generated/**']

    # Quality gate
    waitForQualityGate: true
```

### Features

- **Code quality**: Code smells, bugs, technical debt
- **Security**: Security hotspots and vulnerabilities
- **Coverage**: Test coverage tracking
- **Quality gates**: Automated quality enforcement
- **PR decoration**: Comments on pull requests

---

## Security Stage Configuration

### Complete Security Stage

```yaml
stages:
  Security:
    # SAST
    checkmarxExecuteScan: true
    fortifyExecuteScan: false
    codeqlExecuteScan: true

    # SCA
    whitesourceExecuteScan: true
    detectExecuteScan: false
    snykExecute: false

    # Container
    protecodeExecuteScan: true

    # Secrets
    credentialdiggerScan: true

    # Quality
    sonarExecuteScan: true

steps:
  # Checkmarx configuration
  checkmarxExecuteScan:
    vulnerabilityThresholdHigh: 0
    generatePdfReport: true

  # WhiteSource configuration
  whitesourceExecuteScan:
    cvssSeverityLimit: 7.0

  # CodeQL configuration
  codeqlExecuteScan:
    language: 'java'
    querySuite: 'security-extended'

  # SonarQube configuration
  sonarExecuteScan:
    waitForQualityGate: true
```

---

## Best Practices

### 1. Multi-Layer Security

Use multiple scan types:

```yaml
stages:
  Security:
    checkmarxExecuteScan: true      # SAST
    whitesourceExecuteScan: true    # SCA
    protecodeExecuteScan: true      # Container
    credentialdiggerScan: true      # Secrets
```

### 2. Threshold Management

Set appropriate thresholds:

```yaml
steps:
  checkmarxExecuteScan:
    vulnerabilityThresholdHigh: 0      # No high vulns
    vulnerabilityThresholdMedium: 5    # Max 5 medium
    vulnerabilityThresholdLow: 20      # Max 20 low
```

### 3. Incremental Scanning

Use incremental scans for speed:

```yaml
steps:
  checkmarxExecuteScan:
    incremental: true
    fullScansScheduled: true
    fullScanCycle: 7  # Weekly full scan
```

### 4. Branch-Specific Scans

Different rules for different branches:

```yaml
general:
  productiveBranch: 'main'

steps:
  checkmarxExecuteScan:
    # Only on main branch
    preset: 'Checkmarx Default'

  sonarExecuteScan:
    # PR decoration
    pullRequestProvider: 'GitHub'
```

### 5. Secret Management

Never hardcode credentials:

```yaml
steps:
  checkmarxExecuteScan:
    serverUrl: 'https://checkmarx.example.com'
    username: 'checkmarx-user'
    password: 'vault:checkmarx:password'  # From Vault
```

---

## Security Scanning Comparison

| Tool | Type | Speed | Accuracy | False Positives | Container | License |
|------|------|-------|----------|-----------------|-----------|---------|
| Checkmarx | SAST | Medium | High | Medium | No | Commercial |
| Fortify | SAST | Slow | Very High | Low | No | Commercial |
| CodeQL | SAST | Fast | High | Low | No | Free (OSS) |
| WhiteSource | SCA | Fast | High | Low | No | Commercial |
| Black Duck | SCA | Medium | High | Low | Yes | Commercial |
| Snyk | SCA | Fast | High | Low | Yes | Freemium |
| Protecode | Container | Fast | High | Low | Yes | Commercial |
| Contrast | IAST | Real-time | Very High | Very Low | No | Commercial |

---

## Troubleshooting

### "Authentication failed"

**Solution**: Verify credentials in Vault:

```yaml
steps:
  checkmarxExecuteScan:
    password: 'vault:checkmarx:password'
```

### "Threshold exceeded"

**Solution**: Review and fix vulnerabilities, or adjust thresholds:

```yaml
steps:
  checkmarxExecuteScan:
    vulnerabilityThresholdHigh: 5  # Increase temporarily
```

### "Scan timeout"

**Solution**: Increase timeout or use incremental scans:

```yaml
steps:
  checkmarxExecuteScan:
    incremental: true
    vulnerabilityThresholdTimeout: 3600  # 1 hour
```

---

## Summary

jenkins-library provides comprehensive security scanning with:

- **12 integrated security tools**
- **Multi-layer security**: SAST, SCA, Container, Secrets, IAST
- **Flexible thresholds**: Per-tool, per-severity configuration
- **Quality gates**: Automated security enforcement
- **Reporting**: PDF, SARIF, JSON formats
- **CI/CD integration**: Native pipeline integration

Build secure applications with confidence using Piper's security scanning capabilities!

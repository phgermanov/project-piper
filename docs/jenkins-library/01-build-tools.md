# Jenkins Library - Build Tools

This document covers all build tool integrations available in jenkins-library.

## Overview

jenkins-library supports a wide range of build tools with automatic detection and configuration. The library can detect your project type based on descriptor files and execute the appropriate build commands.

## Build Tool Detection

Piper automatically detects build tools based on these files:

| File | Build Tool | Step |
|------|-----------|------|
| `pom.xml` | Maven | `mavenBuild` |
| `package.json` | npm | `npmExecute` |
| `mta.yaml` | MTA | `mtaBuild` |
| `build.gradle` | Gradle | `gradleExecuteBuild` |
| `go.mod` | Go | `golangBuild` |
| `setup.py` | Python/Pip | `pythonBuild` |
| `Dockerfile` | Docker | `kanikoExecute`, `dockerExecute` |
| `build.sbt` | Scala/SBT | (via `buildExecute`) |
| `dub.json` | D Language | `dubExecute` |

---

## Maven Build

### Step: `mavenBuild`

**Location**: `vars/mavenBuild.groovy`, `cmd/mavenBuild.go`

Executes Maven builds with comprehensive configuration options.

### Basic Usage

```groovy
mavenBuild script: this
```

### Advanced Configuration

```yaml
# .pipeline/config.yml
steps:
  mavenBuild:
    pomPath: 'pom.xml'
    goals: ['clean', 'install']
    profiles: ['production']

    # Publishing
    publish: true
    createBOM: true

    # Maven settings
    projectSettingsFile: '.maven/settings.xml'
    globalSettingsFile: '.maven/global-settings.xml'
    m2Path: '.m2'

    # Deployment
    altDeploymentRepositoryUrl: 'https://nexus.example.com/repository/maven-releases'
    altDeploymentRepositoryID: 'nexus'
    altDeploymentRepositoryUser: 'deploy-user'
    altDeploymentRepositoryPassword: 'secret'

    # Build options
    flatten: true
    verify: false
    defines: ['-DskipTests=false']

    # TLS certificates
    customTlsCertificateLinks: ['https://certs.example.com/ca.crt']
```

### Key Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `pomPath` | string | Path to pom.xml | `pom.xml` |
| `goals` | []string | Maven goals | `['install']` |
| `profiles` | []string | Maven profiles to activate | `[]` |
| `publish` | bool | Publish artifacts | `false` |
| `createBOM` | bool | Create bill of materials | `false` |
| `flatten` | bool | Flatten POMs for CI-friendly versions | `true` |
| `verify` | bool | Run verify instead of install | `false` |
| `logSuccessfulMavenTransfers` | bool | Log successful downloads | `false` |

### Features

- **CI-friendly versioning**: Automatic version flattening
- **Custom repositories**: Deploy to custom Maven repositories
- **Settings inheritance**: Project and global settings
- **Certificate handling**: Custom TLS certificates for corporate proxies
- **Build metadata**: Automatic build info generation
- **SBOM generation**: Create Software Bill of Materials

### Example: Multi-Module Maven Project

```yaml
steps:
  mavenBuild:
    pomPath: 'parent/pom.xml'
    goals: ['clean', 'install', 'deploy']
    profiles: ['production', 'docker']
    publish: true
    createBOM: true
    defines:
      - '-DskipITs=false'
      - '-Ddocker.registry=myregistry.com'
```

---

## npm Build

### Step: `npmExecute`

**Location**: `vars/npmExecute.groovy`, `cmd/npmExecute.go`

Executes npm commands with registry configuration and caching.

### Basic Usage

```groovy
npmExecute script: this, npmCommand: 'run build'
```

### Advanced Configuration

```yaml
steps:
  npmExecute:
    npmCommand: 'run build'
    defaultNpmRegistry: 'https://registry.npmjs.org'
    sapNpmRegistry: 'https://npm.sap.com'

    # Build options
    production: true
    createBOM: true

    # Custom certificates
    customTlsCertificateLinks:
      - 'https://certs.example.com/ca.crt'
```

### Related Steps

#### `npmExecuteScripts`
Execute multiple npm scripts:

```yaml
steps:
  npmExecuteScripts:
    runScripts: ['clean', 'build', 'test']
    install: true
    production: false
```

#### `npmExecuteLint`
Run npm linting:

```yaml
steps:
  npmExecuteLint:
    failOnError: true
    defaultNpmRegistry: 'https://registry.npmjs.org'
```

#### `npmExecuteTests`
Run npm tests:

```yaml
steps:
  npmExecuteTests:
    runScripts: ['test']
    failOnError: true
```

#### `npmExecuteEndToEndTests`
Run E2E tests:

```yaml
steps:
  npmExecuteEndToEndTests:
    runScript: 'e2e'
    baseUrl: 'http://localhost:8080'
    credentialsId: 'e2e-credentials'
```

### Features

- **Registry management**: Support for multiple npm registries
- **Certificate handling**: Custom TLS certificates
- **Caching**: Automatic npm cache management
- **SBOM generation**: Create package bill of materials
- **Script execution**: Run any npm script

---

## Gradle Build

### Step: `gradleExecuteBuild`

**Location**: `vars/gradleExecuteBuild.groovy`, `cmd/gradleExecuteBuild.go`

Executes Gradle builds with wrapper support.

### Basic Usage

```groovy
gradleExecuteBuild script: this
```

### Configuration

```yaml
steps:
  gradleExecuteBuild:
    path: 'build.gradle'
    task: 'build'
    useWrapper: true

    # Publishing
    publish: true
    createBOM: true

    # Gradle options
    excludes: ['test']
```

### Key Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `path` | string | Path to build.gradle | `build.gradle` |
| `task` | string | Gradle task to execute | `build` |
| `useWrapper` | bool | Use Gradle wrapper | `true` |
| `publish` | bool | Publish artifacts | `false` |
| `createBOM` | bool | Create BOM | `false` |

---

## Multi-Target Application (MTA) Build

### Step: `mtaBuild`

**Location**: `vars/mtaBuild.groovy`, `cmd/mtaBuild.go`

Builds Multi-Target Applications for SAP BTP.

### Basic Usage

```groovy
mtaBuild script: this
```

### Configuration

```yaml
steps:
  mtaBuild:
    buildTarget: 'CF'
    mtaJarLocation: '/opt/sap/mta/lib/mta.jar'

    # Extension descriptors
    extension: 'extension_prod.mtaext'

    # Build options
    platform: 'CF'
    applicationName: 'my-app'

    # Publishing
    publish: true
    createBOM: true

    # MTA-specific
    mtaBuildTool: 'cloudMbt'
    source: './'
    target: 'target'
```

### Key Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `buildTarget` | string | Build target (CF, NEO, XSA) | `CF` |
| `mtaJarLocation` | string | Path to MTA jar | Auto-detected |
| `extension` | string | MTA extension descriptor | `''` |
| `platform` | string | Target platform | `CF` |
| `mtaBuildTool` | string | MTA build tool | `cloudMbt` |

### Features

- **Multi-platform**: Support for CF, Neo, XSA
- **Extension descriptors**: Multiple environment support
- **Module building**: Automatic module detection and building
- **Resource optimization**: Efficient build caching

---

## Go Build

### Step: `golangBuild`

**Location**: `vars/golangBuild.groovy`, `cmd/golangBuild.go`

Builds Go applications with module support.

### Basic Usage

```groovy
golangBuild script: this
```

### Configuration

```yaml
steps:
  golangBuild:
    buildFlags: ['-v']
    cgoEnabled: false
    coverageFormat: 'html'
    createBOM: true

    # Build options
    packages: ['./...']
    output: 'binary-name'

    # Testing
    runTests: true
    runIntegrationTests: false

    # Linting
    runLint: true
    lintTool: 'golangci-lint'

    # Environment
    targetArchitectures: ['linux,amd64', 'darwin,amd64', 'windows,amd64']
```

### Key Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `buildFlags` | []string | Go build flags | `[]` |
| `cgoEnabled` | bool | Enable CGO | `false` |
| `packages` | []string | Packages to build | `['./...']` |
| `output` | string | Output binary name | Module name |
| `runTests` | bool | Run tests | `true` |
| `runLint` | bool | Run linting | `false` |

### Features

- **Multi-arch builds**: Build for multiple platforms
- **Module support**: Go modules (go.mod)
- **Testing integration**: Unit and integration tests
- **Linting**: golangci-lint integration
- **Coverage reports**: HTML and Cobertura formats
- **SBOM generation**: Dependency bill of materials

---

## Python Build

### Step: `pythonBuild`

**Location**: `vars/pythonBuild.groovy`, `cmd/pythonBuild.go`

Builds Python projects with pip and setuptools.

### Basic Usage

```groovy
pythonBuild script: this
```

### Configuration

```yaml
steps:
  pythonBuild:
    buildFlags: ['--verbose']
    createBOM: true
    publish: true

    # Build type
    buildType: 'wheel'  # or 'sdist'

    # Installation
    installRequirements: true
    requirementsFilePath: 'requirements.txt'

    # Testing
    runTests: true
    testResultPath: 'test-results'

    # PyPI configuration
    targetRepositoryURL: 'https://upload.pypi.org/legacy/'
    targetRepositoryUser: 'user'
    targetRepositoryPassword: 'secret'
```

### Key Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `buildType` | string | Build type (wheel, sdist) | `wheel` |
| `installRequirements` | bool | Install requirements.txt | `true` |
| `publish` | bool | Publish to PyPI | `false` |
| `createBOM` | bool | Create BOM | `false` |

---

## Generic Build

### Step: `buildExecute`

**Location**: `vars/buildExecute.groovy`, `cmd/buildExecute.go`

Generic build execution that auto-detects build tool.

### Basic Usage

```groovy
buildExecute script: this
```

### Features

- **Auto-detection**: Automatically detects build tool
- **Fallback**: Manual build tool specification
- **All build tools**: Supports all Piper build tools

### Configuration

```yaml
steps:
  buildExecute:
    buildTool: 'maven'  # Force specific build tool
```

---

## Docker Build

### Step: `dockerExecute`

**Location**: `vars/dockerExecute.groovy`, `cmd/dockerExecute.go`

Execute commands in Docker containers.

### Basic Usage

```groovy
dockerExecute(script: this, dockerImage: 'maven:3.6-jdk-8') {
    sh 'mvn clean install'
}
```

### Configuration

```yaml
steps:
  dockerExecute:
    dockerImage: 'maven:3.6-jdk-8'
    dockerOptions: ['-u', '1000:1000']
    dockerEnvVars:
      MAVEN_OPTS: '-Xmx1024m'
    dockerVolumeBind:
      '/home/jenkins/.m2': '/root/.m2'
```

---

## Kaniko Build

### Step: `kanikoExecute`

**Location**: `vars/kanikoExecute.groovy`, `cmd/kanikoExecute.go`

Build container images using Kaniko (without Docker daemon).

### Basic Usage

```groovy
kanikoExecute script: this
```

### Configuration

```yaml
steps:
  kanikoExecute:
    containerImageName: 'myapp'
    containerImageTag: '1.0.0'
    containerRegistryUrl: 'https://myregistry.com'

    # Multi-image build
    containerMultiImageBuild: true
    containerMultiImageBuildExcludes: ['**/node_modules/**']
    containerMultiImageBuildTrimDir: 'src/'

    # Dockerfile
    dockerfilePath: 'Dockerfile'

    # Build options
    buildOptions: ['--skip-tls-verify', '--cache=true']

    # SBOM
    createBOM: true
    syftDownloadUrl: 'https://github.com/anchore/syft/releases'
```

### Features

- **Daemon-less**: No Docker daemon required
- **Multi-arch**: Build for multiple architectures
- **Caching**: Layer caching support
- **Multi-image**: Build multiple images from one repository
- **SBOM**: Syft integration for SBOM generation

---

## Cloud Native Buildpacks (CNB)

### Step: `cnbBuild`

**Location**: `vars/cnbBuild.groovy`, `cmd/cnbBuild.go`

Build container images using Cloud Native Buildpacks.

### Basic Usage

```groovy
cnbBuild script: this
```

### Configuration

```yaml
steps:
  cnbBuild:
    containerImageName: 'myapp'
    containerImageTag: 'latest'

    # Builder
    cnbBuilder: 'paketobuildpacks/builder:base'
    cnbBuildpacks: ['paketo-buildpacks/java']

    # Build options
    cnbBuildEnv:
      BP_JVM_VERSION: '11'
      BP_MAVEN_BUILT_ARTIFACT: 'target/*.jar'

    # Registry
    containerRegistryUrl: 'https://myregistry.com'
    createBOM: true
```

### Features

- **Auto-detection**: Automatic buildpack selection
- **No Dockerfile**: Build without Dockerfile
- **Best practices**: Secure, efficient images
- **Caching**: Smart layer caching
- **SBOM**: Built-in SBOM generation

---

## D Language Build

### Step: `dubExecute`

**Location**: `vars/dubExecute.groovy`, `cmd/dubExecute.go`

Build D language projects using DUB.

### Basic Usage

```groovy
dubExecute script: this
```

### Configuration

```yaml
steps:
  dubExecute:
    dubCommand: 'build'
    dockerImage: 'dlang2/dmd-ubuntu:latest'
```

---

## Build Result Publishing

### Artifact Preparation

#### Step: `artifactPrepareVersion`

**Location**: `vars/artifactPrepareVersion.groovy`, `cmd/artifactPrepareVersion.go`

Prepares artifact version based on Git information.

### Configuration

```yaml
steps:
  artifactPrepareVersion:
    versioningType: 'cloud'
    timestampTemplate: '%Y%m%d%H%M%S'
    tagPrefix: 'build_'
    commitVersion: true

    # Build tool specific
    buildTool: 'maven'  # Auto-detected
```

### Versioning Types

- **cloud**: Semantic version + timestamp (1.0.0-20240115120000)
- **cloud_noTag**: No Git tagging
- **library**: Semantic versioning only

### Set Version

#### Step: `artifactSetVersion`

Sets version in build descriptors.

```yaml
steps:
  artifactSetVersion:
    buildTool: 'maven'
    version: '1.2.3'
```

---

## Build Comparison Matrix

| Feature | Maven | npm | Gradle | MTA | Go | Python | Docker | Kaniko | CNB |
|---------|-------|-----|--------|-----|----|----|--------|--------|-----|
| Auto-detect | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| SBOM | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Publishing | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ |
| Multi-arch | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ | ✅ | ✅ |
| Container | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Caching | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

---

## Best Practices

### 1. Version Management

Use `artifactPrepareVersion` at the start of build stage:

```yaml
stages:
  Build:
    artifactPrepareVersion: true
    mavenBuild: true
```

### 2. SBOM Generation

Always create Software Bill of Materials:

```yaml
steps:
  mavenBuild:
    createBOM: true
```

### 3. Container Builds

Prefer CNB for standard applications, Kaniko for custom Dockerfiles:

```yaml
# Standard Java app - use CNB
cnbBuild:
  cnbBuilder: 'paketobuildpacks/builder:base'

# Custom Dockerfile - use Kaniko
kanikoExecute:
  dockerfilePath: 'Dockerfile'
```

### 4. Publishing

Only publish from productive branches:

```yaml
steps:
  mavenBuild:
    publish: true

general:
  productiveBranch: 'main'
```

---

## Troubleshooting

### Maven: "Could not resolve dependencies"

**Solution**: Configure custom Maven settings:

```yaml
steps:
  mavenBuild:
    projectSettingsFile: '.maven/settings.xml'
    globalSettingsFile: '.maven/global-settings.xml'
```

### npm: "Unable to authenticate"

**Solution**: Configure npm registry with credentials:

```yaml
steps:
  npmExecute:
    defaultNpmRegistry: 'https://registry.npmjs.org'
```

### Kaniko: "Error pushing image"

**Solution**: Ensure registry credentials are configured:

```yaml
steps:
  kanikoExecute:
    containerRegistryUrl: 'https://myregistry.com'
    # Credentials via Jenkins credentialsId or Vault
```

---

## Summary

jenkins-library provides comprehensive build tool support with:

- **9 build tools** supported out of the box
- **Automatic detection** based on project files
- **Consistent interface** across all build tools
- **SBOM generation** for supply chain security
- **Container support** for all build tools
- **Publishing** to artifact repositories
- **Version management** with Git integration

Choose the right build tool for your project, and Piper handles the rest!

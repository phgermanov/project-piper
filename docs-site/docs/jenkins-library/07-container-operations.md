# Jenkins Library - Container Operations

This document covers all container-related operations available in jenkins-library.

## Overview

jenkins-library provides comprehensive container operations covering the entire container lifecycle:

**Build**: Docker, Kaniko, Cloud Native Buildpacks | **Test**: Structure tests, Dockerfile linting | **Push**: Registry deployment | **Execute**: Docker/Kubernetes execution

### Container Operations Matrix

| Step | Purpose | Docker Daemon | Kubernetes | Multi-Arch |
|------|---------|---------------|------------|------------|
| `dockerExecute` | Execute in containers | Required | Optional | No |
| `dockerExecuteOnKubernetes` | Execute on K8s | No | Required | No |
| `kanikoExecute` | Build without daemon | No | Yes | Yes |
| `cnbBuild` | Build with Buildpacks | No | Yes | Yes |
| `containerExecuteStructureTests` | Test structure | Optional | Yes | No |
| `hadolintExecute` | Lint Dockerfiles | No | Yes | No |
| `containerPushToRegistry` | Push images | Preferred | Yes | No |
| `imagePushToRegistry` | Copy images | No | Yes | Yes |
| `containerSaveImage` | Save as tar | No | Yes | No |

---

## dockerExecute

**Location**: `vars/dockerExecute.groovy`

Executes commands inside Docker containers. Auto-detects environment (Docker/Kubernetes).

### Basic Usage

```groovy
dockerExecute(script: this, dockerImage: 'maven:3.8-jdk-11') {
    sh 'mvn clean install'
}
```

### Configuration

```yaml
steps:
  dockerExecute:
    dockerImage: 'maven:3.8-jdk-11'
    dockerPullImage: true
    dockerRegistryUrl: 'https://my.registry.com'
    dockerRegistryCredentialsId: 'registry-creds'

    # Environment
    dockerEnvVars:
      MAVEN_OPTS: '-Xmx2048m'

    # Docker only
    dockerVolumeBind:
      '/home/jenkins/.m2': '/root/.m2'
    dockerOptions: ['-u 1000:1000', '--memory=4g']

    # Kubernetes only
    containerCommand: '/usr/bin/tail -f /dev/null'
    dockerWorkspace: '/workspace'
```

### Key Features

- Auto-detects Docker daemon or Kubernetes
- Proxy environment inheritance
- Sidecar containers for testing
- Volume mounting and credential management

---

## dockerExecuteOnKubernetes

**Location**: `vars/dockerExecuteOnKubernetes.groovy`

Executes in Kubernetes pods with full pod configuration control.

### Basic Usage

```groovy
dockerExecuteOnKubernetes(script: this, dockerImage: 'node:16') {
    sh 'npm install && npm test'
}
```

### Advanced Configuration

```yaml
steps:
  dockerExecuteOnKubernetes:
    dockerImage: 'maven:3.8-jdk-11'
    containerName: 'maven'

    # Resources
    resources:
      maven:
        requests: {memory: '2Gi', cpu: '1'}
        limits: {memory: '4Gi', cpu: '2'}

    # Security
    securityContext:
      runAsUser: 1000
      fsGroup: 1000

    # Pod properties
    additionalPodProperties:
      imagePullSecrets: [{name: 'registry-secret'}]
      nodeSelector: {disktype: 'ssd'}
```

### Key Features

- Multi-container pods
- Resource limits (CPU/memory)
- Security contexts
- Init containers
- Custom pod specifications

---

## kanikoExecute

**Location**: `vars/kanikoExecute.groovy`, `cmd/kanikoExecute.go`

Builds container images using Kaniko (no Docker daemon required).

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
    containerRegistryUrl: 'https://gcr.io'
    dockerfilePath: 'Dockerfile'

    buildOptions:
      - '--cache=true'
      - '--cache-repo=gcr.io/project/cache'
      - '--build-arg=VERSION=1.0.0'

    # Authentication
    containerRegistryUser: 'oauth2accesstoken'
    containerRegistryPassword: 'vault:gcr:token'

    # SBOM
    createBOM: true
```

### Multi-Image Build

```yaml
# Single root with multiple Dockerfiles
containerMultiImageBuild: true
# Or multiple contexts
multipleImages:
  - {containerImageName: 'myorg/api', contextSubPath: 'services/api'}
  - {containerImageName: 'myorg/frontend', contextSubPath: 'services/frontend'}
```

### Key Features

- No Docker daemon needed
- Layer caching for speed
- Multi-arch support
- Multi-image builds
- SBOM generation with Syft

---

## cnbBuild

**Location**: `vars/cnbBuild.groovy`, `cmd/cnbBuild.go`

Builds images using Cloud Native Buildpacks (no Dockerfile required).

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
    containerRegistryUrl: 'docker.io'

    # Buildpacks
    buildpacks: ['paketo-buildpacks/java']

    # Environment
    buildEnvVars:
      BP_JVM_VERSION: '17'
      BP_MAVEN_BUILT_ARTIFACT: 'target/*.jar'

    additionalTags: ['latest', 'v1.0']
    createBOM: true
```

### Key Features

- No Dockerfile needed
- Auto-detects language/framework
- Best practices built-in
- Layer rebase capability
- Built-in SBOM generation

---

## containerExecuteStructureTests

**Location**: `vars/containerExecuteStructureTests.groovy`, `cmd/containerExecuteStructureTests.go`

Validates container images with structure tests.

### Basic Usage

```groovy
containerExecuteStructureTests(
    script: this,
    testImage: 'myapp:latest',
    testConfiguration: 'tests/structure-test.yaml'
)
```

### Test Configuration

```yaml
# structure-test.yaml
schemaVersion: '2.0.0'
fileExistenceTests:
  - {name: 'App exists', path: '/app/server', shouldExist: true}
fileContentTests:
  - {name: 'No secrets', path: '/app/.env', excludedContents: ['PASSWORD.*']}
metadataTest:
  exposedPorts: ['8080']
  workdir: '/app'
  user: 'app'
commandTests:
  - {name: 'Health check', command: 'curl', args: ['http://localhost:8080/health'], expectedOutput: ['OK']}

steps:
  containerExecuteStructureTests:
    testImage: 'myapp:1.0.0'
    testConfiguration: 'tests/*.yaml'
    testDriver: 'docker'
```

---

## hadolintExecute

**Location**: `vars/hadolintExecute.groovy`, `cmd/hadolintExecute.go`

Lints Dockerfiles for best practices.

### Basic Usage

```groovy
hadolintExecute script: this
```

### Configuration

```yaml
steps:
  hadolintExecute:
    dockerFile: './Dockerfile'
    configurationFile: '.hadolint.yaml'
    qualityGates:
      - threshold: 1
        type: 'TOTAL_ERROR'
```

### Hadolint Config

```yaml
# .hadolint.yaml
ignored: [DL3008, DL3009]
trustedRegistries: [docker.io, gcr.io]
override:
  error: [DL3001]
```

---

## containerPushToRegistry

**Location**: `vars/containerPushToRegistry.groovy`

Pushes images to registries (Docker daemon or Skopeo).

### Basic Usage

```groovy
containerPushToRegistry(
    script: this,
    dockerImage: 'myapp:1.0.0',
    dockerRegistryUrl: 'https://docker.io',
    dockerCredentialsId: 'dockerhub-creds'
)
```

### Configuration

```yaml
steps:
  containerPushToRegistry:
    dockerImage: 'myorg/myapp:1.0.0'
    dockerRegistryUrl: 'https://docker.io'
    dockerCredentialsId: 'dockerhub-creds'
    tagLatest: true
    tagArtifactVersion: true

    # Move from another registry
    sourceImage: 'gcr.io/project/myapp:1.0.0'
    sourceRegistryUrl: 'https://gcr.io'
```

---

## imagePushToRegistry

**Location**: `vars/imagePushToRegistry.groovy`, `cmd/imagePushToRegistry.go`

Copies images between registries using Crane (no daemon).

### Basic Usage

```groovy
imagePushToRegistry script: this
```

### Configuration

```yaml
steps:
  imagePushToRegistry:
    # Source
    sourceImages: ['myapp', 'sidecar']
    sourceImageTag: '1.0.0'
    sourceRegistryUrl: 'https://gcr.io'
    sourceRegistryUser: 'oauth2accesstoken'
    sourceRegistryPassword: 'vault:gcr:token'

    # Target
    targetRegistryUrl: 'https://docker.io'
    targetRegistryUser: 'username'
    targetRegistryPassword: 'vault:dockerhub:token'

    # Mapping
    targetImages:
      myapp: 'myorg/application'
      sidecar: 'myorg/sidecar'

    tagLatest: true
```

### Key Features

- No Docker daemon
- Multi-platform support
- Efficient layer deduplication
- Flexible image renaming

---

## containerSaveImage

**Location**: `vars/containerSaveImage.groovy`, `cmd/containerSaveImage.go`

Saves container images as tar archives.

### Basic Usage

```groovy
containerSaveImage(
    script: this,
    containerImage: 'myapp:1.0.0',
    containerRegistryUrl: 'https://docker.io'
)
```

### Configuration

```yaml
steps:
  containerSaveImage:
    containerImage: 'myapp:1.0.0'
    containerRegistryUrl: 'https://docker.io'
    filePath: 'myapp-image.tar'
    imageFormat: 'oci'  # legacy, tarball, oci
    containerRegistryUser: 'username'
    containerRegistryPassword: 'vault:registry:password'
```

---

## Best Practices

### 1. Choose the Right Build Tool

```yaml
# Standard app -> CNB (no Dockerfile)
cnbBuild: {containerImageName: 'myapp'}

# Custom requirements -> Kaniko
kanikoExecute: {dockerfilePath: 'Dockerfile'}
```

### 2. Always Generate SBOMs

```yaml
kanikoExecute: {createBOM: true}
cnbBuild: {createBOM: true}
```

### 3. Implement Quality Gates

```yaml
hadolintExecute:
  qualityGates: [{threshold: 0, type: 'TOTAL_ERROR'}]

containerExecuteStructureTests:
  testConfiguration: 'tests/*.yaml'
```

### 4. Secure Credentials

```yaml
# Use Vault
containerRegistryPassword: 'vault:registry:password'

# Or Jenkins credentials
dockerConfigJsonCredentialsId: 'docker-config'
```

---

## Troubleshooting

### No Docker Daemon

**Solution**: Use Kaniko or CNB
```yaml
kanikoExecute: {containerImageName: 'myapp'}
```

### Out of Memory

**Solution**: Configure resources
```yaml
dockerExecuteOnKubernetes:
  resources:
    DEFAULT:
      limits: {memory: '8Gi'}
```

### Registry Auth Failed

**Solution**: Verify credentials
```yaml
kanikoExecute:
  dockerConfigJSON: '.docker/config.json'
  customTlsCertificateLinks: ['https://ca.example.com/cert.pem']
```

---

## Summary

jenkins-library provides complete container operations:

- **3 build methods**: Docker, Kaniko, Cloud Native Buildpacks
- **2 execution contexts**: Docker daemon, Kubernetes pods
- **Testing**: Structure tests, Hadolint linting
- **Registry operations**: Push, copy, save
- **Multi-arch support**: Build for multiple platforms
- **SBOM generation**: Security compliance built-in
- **No daemon required**: Works in restricted environments

Choose the right tool for your environment!

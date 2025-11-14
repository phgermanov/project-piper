# Piper Simple Go - Docker Build Example

This document explains how to use Docker as the build tool instead of native Go builds for the piper-simple-go project.

## Overview

The piper-simple-go project can be built in two ways:

1. **Native Go Build** (default `config.yml`): Uses `golangBuild` step
2. **Docker Build** (this example): Uses `kanikoExecute` or `dockerExecute` step

## Why Use Docker Build?

### Advantages:
- ✅ **Consistent builds** across all environments
- ✅ **Production-ready container** as build artifact
- ✅ **No local Go installation** needed
- ✅ **Multi-stage builds** for smaller images
- ✅ **Security scanning** of container images
- ✅ **Easy deployment** to Kubernetes/Cloud platforms

### Disadvantages:
- ❌ Slower builds (unless using cache)
- ❌ Larger artifacts (container images vs binaries)
- ❌ Requires container registry

## Files

### Dockerfile
Multi-stage Dockerfile that:
- **Stage 1 (builder)**: Compiles Go application with `golang:1.24-alpine`
- **Stage 2 (runtime)**: Minimal Alpine Linux with compiled binary only
- **Result**: ~15MB image instead of ~300MB

### .dockerignore
Optimizes Docker build by excluding:
- Git files and history
- Pipeline configuration
- Documentation
- IDE files
- Test files

### .pipeline/config-docker.yml
Complete Piper configuration using Docker build tool with:
- Kaniko for daemon-less builds
- Container structure tests
- Dockerfile linting with Hadolint
- Security scanning with Snyk
- SBOM generation

## Quick Start

### Local Docker Build

```bash
# Build the image
docker build -t piper-simple-go:local .

# Run the container
docker run -p 8080:8080 piper-simple-go:local

# Test the application
curl http://localhost:8080
# Output: Hello, Go!
```

### Using Piper Pipeline

#### Option 1: Use the Docker config

```bash
# Copy the Docker config to be your main config
cp .pipeline/config-docker.yml .pipeline/config.yml

# Or reference it in your pipeline setup
```

#### Option 2: Switch buildTool in existing config

```yaml
# In .pipeline/config.yml
general:
  buildTool: docker  # Changed from 'golang'

steps:
  kanikoExecute:
    containerImageName: piper-simple-go
    containerImageTag: ${VERSION}
```

## Configuration Comparison

### Native Go Build (config.yml)

```yaml
general:
  buildTool: golang
  nativeBuild: true

steps:
  golangBuild:
    packages: ['./...']
    output: piper-simple-go
    runTests: true
    createBOM: true
```

**Produces**: Binary file `piper-simple-go`

### Docker Build (config-docker.yml)

```yaml
general:
  buildTool: docker

steps:
  kanikoExecute:
    containerImageName: piper-simple-go
    containerImageTag: latest
    dockerfilePath: Dockerfile
    createBOM: true
```

**Produces**: Container image `piper-simple-go:latest`

## Complete Pipeline Stages

When using Docker build, the pipeline executes:

### 1. Init Stage
- Checkout code
- Prepare version
- Initialize pipeline environment

### 2. Build Stage
```yaml
# Hadolint - Lint Dockerfile
hadolintExecute

# Kaniko - Build container image
kanikoExecute

# Structure Tests - Validate image
containerExecuteStructureTests
```

### 3. Security Stage
```yaml
# Snyk - Scan container for vulnerabilities
snykExecute:
  scanType: docker
```

### 4. Promote Stage
```yaml
# Push image to registry
containerPushToRegistry
```

## Advanced Configuration

### Multi-Architecture Builds

Build for multiple platforms:

```yaml
steps:
  kanikoExecute:
    targetArchitectures:
      - linux/amd64
      - linux/arm64
      - linux/arm/v7
```

### Custom Build Arguments

Pass build-time variables:

```yaml
steps:
  kanikoExecute:
    buildOptions:
      - --build-arg=VERSION=${VERSION}
      - --build-arg=BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
```

Then in Dockerfile:
```dockerfile
ARG VERSION
ARG BUILD_DATE
LABEL version="${VERSION}"
LABEL build-date="${BUILD_DATE}"
```

### Registry Authentication

Using Vault:
```yaml
steps:
  kanikoExecute:
    containerRegistryUrl: https://myregistry.com
    containerRegistryUser: deploy-user
    # Password from Vault:
    # vaultPath: piper/credentials/docker-registry
    # vaultKey: password
```

Using Jenkins credentials:
```yaml
steps:
  kanikoExecute:
    dockerConfigJSON: docker-credentials-id
```

### Container Structure Tests

Validate the built image:

```yaml
steps:
  containerExecuteStructureTests:
    testConfiguration: |
      schemaVersion: 2.0.0

      commandTests:
        - name: "App runs"
          command: "/root/piper-simple-go"
          args: ["--version"]

      fileExistenceTests:
        - name: "CA certificates exist"
          path: "/etc/ssl/certs/ca-certificates.crt"
          shouldExist: true

      metadataTest:
        exposedPorts: ["8080"]
        workdir: "/root"
        cmd: ["./piper-simple-go"]
```

## Optimizing Docker Builds

### 1. Layer Caching

Enable Kaniko cache:
```yaml
kanikoExecute:
  buildOptions:
    - --cache=true
    - --cache-ttl=24h
    - --cache-repo=myregistry.com/cache
```

### 2. Multi-Stage Build Best Practices

```dockerfile
# Use specific versions
FROM golang:1.24.0-alpine AS builder

# Copy dependencies first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Then copy source code
COPY . .
RUN go build -o app .

# Minimal runtime image
FROM alpine:3.19
COPY --from=builder /app/app .
```

### 3. Reduce Image Size

```dockerfile
# Use distroless for even smaller images
FROM gcr.io/distroless/static:nonroot

# Or use scratch for minimal size
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/app .
```

## Deployment

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: piper-simple-go
spec:
  replicas: 3
  selector:
    matchLabels:
      app: piper-simple-go
  template:
    metadata:
      labels:
        app: piper-simple-go
    spec:
      containers:
      - name: app
        image: your-registry.com/piper-simple-go:1.0.0
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "200m"
```

### Cloud Foundry

```bash
# Push as Docker image
cf push piper-simple-go --docker-image your-registry.com/piper-simple-go:1.0.0
```

### Docker Compose

```yaml
version: '3.8'
services:
  app:
    image: piper-simple-go:latest
    ports:
      - "8080:8080"
    restart: unless-stopped
```

## Troubleshooting

### Build fails with "permission denied"

**Solution**: Add buildOptions:
```yaml
kanikoExecute:
  buildOptions:
    - --skip-tls-verify-registry=myregistry.com
```

### Image too large

**Solution**: Use multi-stage build and distroless base:
```dockerfile
FROM gcr.io/distroless/static:nonroot
```

### Cannot push to registry

**Solution**: Check credentials configuration:
```yaml
# Ensure Vault path is correct
vault:
  containerRegistryPassword:
    vaultPath: piper/credentials/docker-registry
    vaultKey: password
```

## References

- [Piper Docker Build Documentation](../../docs/jenkins-library/build-tools.md#docker-build)
- [Kaniko Documentation](../../docs/jenkins-library/build-tools.md#kaniko-build)
- [Container Operations](../../docs/jenkins-library/container-operations.md)
- [Dockerfile Best Practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)

## Comparison Summary

| Feature | Native Go Build | Docker Build |
|---------|----------------|--------------|
| **Build Tool** | `golang` | `docker` |
| **Primary Step** | `golangBuild` | `kanikoExecute` |
| **Artifact** | Binary file | Container image |
| **Size** | ~8MB | ~15MB (multi-stage) |
| **Deploy To** | VMs, bare metal | Containers, K8s, Cloud |
| **Build Speed** | Faster | Slower (but cacheable) |
| **Consistency** | Platform-dependent | Platform-independent |
| **Testing** | Unit tests only | Structure + security tests |
| **Best For** | Traditional deployment | Cloud-native deployment |

## Next Steps

1. **Choose your approach**: Native Go or Docker build
2. **Configure registry**: Set up container registry credentials
3. **Run pipeline**: Test the build in your CI/CD environment
4. **Deploy**: Use container image in Kubernetes or Cloud Foundry

For questions or issues, refer to the main [Piper documentation](../../docs/README.md).

# Jenkins Library - Deployment Features

Comprehensive guide to deployment integrations for Cloud Foundry, Kubernetes, SAP platforms, and GitOps workflows.

## Overview

| Step | Platform | Tools | Strategies |
|------|----------|-------|------------|
| `cloudFoundryDeploy` | Cloud Foundry | cf CLI, MTA Plugin | Standard, Blue-Green, Rolling |
| `kubernetesDeploy` | Kubernetes | kubectl, Helm 2/3 | Rolling updates |
| `neoDeploy` | SAP Neo | Neo CLI | Standard, Blue-Green |
| `xsDeploy` | SAP XS | xs CLI | Standard, Blue-Green |
| `multicloudDeploy` | Multi-Cloud | Multiple | Parallel execution |
| `helmExecute` | Kubernetes | Helm 3 | Package management |
| `terraformExecute` | Infrastructure | Terraform | IaC deployment |
| `gitopsUpdateDeployment` | GitOps | kubectl/Helm/Kustomize | Git-based CD |

---

## cloudFoundryDeploy

### Overview
Deploys applications to SAP Cloud Foundry with support for MTA and non-MTA applications.

### Basic Usage
```groovy
cloudFoundryDeploy script: this

cloudFoundryDeploy(script: this, org: 'my-org', space: 'production')
```

### Configuration
```yaml
steps:
  cloudFoundryDeploy:
    apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'
    org: 'my-organization'
    space: 'production'
    credentialsId: 'cf-credentials'
    appName: 'my-app'
    manifest: 'manifest.yml'
    manifestVariables: ['instances=3', 'memory=1024M']

    # Deployment strategy
    deployType: 'standard'
    cfNativeDeployParameters: '--strategy rolling'

    # MTA deployment
    mtaPath: 'my-app.mtar'
    mtaDeployParameters: '-f --strategy blue-green'
```

### Key Parameters
| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `apiEndpoint` | string | CF API endpoint | Required |
| `org` | string | Target organization | Required |
| `space` | string | Target space | Required |
| `deployType` | string | standard/blue-green | `standard` |
| `manifest` | string | Manifest file | `manifest.yml` |

### Deployment Strategies

**Standard**: Uses `cf push`
**Rolling** (recommended): `cfNativeDeployParameters: '--strategy rolling'`
**Blue-Green (MTA)**: `mtaDeployParameters: '--strategy blue-green'`

### Best Practices
- Use rolling deployment for zero-downtime
- Store credentials in Jenkins Credentials or Vault
- Define health checks in manifest
- Always specify resource limits (memory, disk)

---

## kubernetesDeploy

### Overview
Deploys applications to Kubernetes using kubectl or Helm.

### Basic Usage
```groovy
kubernetesDeploy script: this

kubernetesDeploy(
    script: this,
    deployTool: 'helm3',
    chartPath: 'helm-chart',
    deploymentName: 'my-release'
)
```

### Configuration
```yaml
steps:
  kubernetesDeploy:
    namespace: 'production'
    deployTool: 'helm3'

    containerRegistryUrl: 'https://docker.io'
    containerImageName: 'my-app'
    containerImageTag: '1.0.0'

    chartPath: './helm-chart'
    deploymentName: 'my-app-release'
    helmValues: ['values.yaml', 'values-prod.yaml']
    helmDeployWaitSeconds: 300

    forceUpdates: true
    runHelmTests: true
```

### Key Parameters
| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `deployTool` | string | kubectl/helm/helm3 | `kubectl` |
| `namespace` | string | Target namespace | `default` |
| `chartPath` | string | Helm chart path | Required for Helm |
| `deploymentName` | string | Helm release name | Required for Helm |

### Best Practices
- Use Helm 3 for complex deployments
- Enable `--atomic` for automatic rollback
- Define readiness/liveness probes
- Set resource limits and requests
- Use specific image tags, never `latest`

---

## neoDeploy

### Overview
Deploys applications to SAP BTP Neo environment.

### Configuration
```yaml
steps:
  neoDeploy:
    account: 'myaccount'
    host: 'hana.ondemand.com'
    application: 'my-app'
    runtime: 'neo-javaee7-wp'
    runtimeVersion: '3'
    credentialsId: 'neo-credentials'
    size: 'lite'
    vmArguments: '-Xmx512m'
```

---

## xsDeploy

### Overview
Deploys MTA applications to SAP XS Advanced.

### Configuration
```yaml
steps:
  xsDeploy:
    apiUrl: 'https://xs.example.com:30030'
    org: 'my-org'
    space: 'production'
    credentialsId: 'xs-credentials'
    mtaPath: 'my-mta.mtar'
    mode: 'DEPLOY'  # or 'BG_DEPLOY'
```

**Modes**: `DEPLOY` (standard), `BG_DEPLOY` (blue-green with resume/abort)

---

## multicloudDeploy

### Overview
Deploys to multiple Cloud Foundry and Neo instances in parallel.

### Configuration
```yaml
steps:
  multicloudDeploy:
    parallelExecution: true
    enableZeroDowntimeDeployment: true

    cfTargets:
      - apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'
        credentialsId: 'cf-eu10-creds'
        org: 'my-org'
        space: 'production'

      - apiEndpoint: 'https://api.cf.us10.hana.ondemand.com'
        credentialsId: 'cf-us10-creds'
        org: 'my-org'
        space: 'production'

    neoTargets:
      - account: 'account1'
        host: 'eu1.hana.ondemand.com'
        application: 'my-app'
```

### Best Practices
- Enable parallel execution
- Use region-specific manifests
- Create services before deployment

---

## helmExecute

### Overview
Executes Helm 3 operations for Kubernetes package management.

### Configuration
```yaml
steps:
  helmExecute:
    helmCommand: 'upgrade'  # install, test, lint, publish, dependency
    chartPath: './charts/my-app'
    namespace: 'production'
    image: 'my-app:1.0.0'
    helmValues: ['values.yaml', 'values-prod.yaml']

    # Publishing
    publish: true
    version: '1.0.0'
    targetRepositoryURL: 'https://charts.example.com'
    targetRepositoryUser: 'user'
    targetRepositoryPassword: 'vault:helm:password'

    additionalParameters: ['--atomic', '--create-namespace']
```

### Available Commands
| Command | Description |
|---------|-------------|
| `upgrade` | Install or upgrade release |
| `lint` | Check chart for issues |
| `test` | Run tests for release |
| `dependency` | Manage dependencies |
| `publish` | Package and publish chart |

### Best Practices
- Always specify chart versions
- Update dependencies before packaging
- Lint charts before publishing
- Secure credentials in Vault

---

## terraformExecute

### Overview
Executes Terraform for Infrastructure as Code deployments.

### Configuration
```yaml
steps:
  terraformExecute:
    command: 'plan'  # or apply, destroy, validate
    init: true
    workspace: 'production'

    globalOptions: ['-chdir=terraform']
    additionalArgs:
      - '-var=environment=prod'
      - '-var-file=prod.tfvars'

    cliConfigFile: 'vault:terraform:cli-config'
```

### Key Parameters
| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `command` | string | Terraform command | `plan` |
| `init` | bool | Run terraform init | `false` |
| `workspace` | string | Workspace name | - |

### Deployment Pattern
```yaml
# Plan
- terraformExecute:
    command: 'plan'
    init: true
    additionalArgs: ['-out=tfplan']

# Apply
- terraformExecute:
    command: 'apply'
    additionalArgs: ['tfplan']
```

### Best Practices
- Use remote state backends (S3, Azure)
- Separate workspaces per environment
- Always review plans before applying
- Pin Terraform and provider versions
- Store secrets in Vault

---

## gitopsUpdateDeployment

### Overview
Updates Kubernetes deployment manifests in Git repositories for GitOps workflows.

### Configuration
```yaml
steps:
  gitopsUpdateDeployment:
    serverUrl: 'https://github.com'
    username: 'git-user'
    password: 'vault:git:token'
    branchName: 'main'

    tool: 'kubectl'  # or helm, kustomize
    filePath: 'k8s/deployment.yaml'

    containerRegistryUrl: 'https://docker.io'
    containerImageNameTag: 'my-app:1.2.3'
    containerName: 'app'

    commitMessage: 'Update app to version 1.2.3'
```

### Key Parameters
| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `tool` | string | kubectl/helm/kustomize | `kubectl` |
| `filePath` | string | Deployment file path | Required |
| `branchName` | string | Git branch | `master` |
| `containerImageNameTag` | string | Image:tag | Required |

### Tool-Specific Usage

**kubectl**: Updates container image in deployment YAML
**Helm**: Renders full template and commits
**Kustomize**: Updates image in kustomization.yaml

### Best Practices
- Use PR workflows instead of direct commits
- Always use specific image tags
- Keep app code and manifests separate
- Use tokens with minimal permissions
- Validate manifests before committing

---

## Deployment Strategy Comparison

| Strategy | Downtime | Complexity | Resources | Rollback | Use Case |
|----------|----------|------------|-----------|----------|----------|
| **Standard** | Yes | Low | 1x | Fast | Development |
| **Rolling** | No | Medium | 1-2x | Medium | Production |
| **Blue-Green** | No | High | 2x | Instant | Critical apps |
| **GitOps** | No | Medium | 1x | Medium | Declarative |

---

## Common Patterns

### Multi-Region Cloud Foundry
```yaml
multicloudDeploy:
  parallelExecution: true
  cfTargets:
    - apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'
      org: 'org'
      space: 'prod'
    - apiEndpoint: 'https://api.cf.us10.hana.ondemand.com'
      org: 'org'
      space: 'prod'
```

### Helm + GitOps
```yaml
# Package
- helmExecute:
    helmCommand: 'package'
    chartPath: './chart'

# Update GitOps
- gitopsUpdateDeployment:
    tool: 'helm'
    filePath: 'releases/prod/app.yaml'
```

### Terraform + Kubernetes
```yaml
# Provision
- terraformExecute:
    command: 'apply'
    init: true

# Deploy
- kubernetesDeploy:
    deployTool: 'helm3'
    chartPath: './chart'
```

---

## Troubleshooting

### Cloud Foundry
```yaml
# Deployment timeout
cfNativeDeployParameters: '--health-check-timeout 300'

# Memory issues
manifestVariables: ['memory=2048M']
```

### Kubernetes
```yaml
# Image pull errors
createDockerRegistrySecret: true
containerRegistryPassword: 'vault:registry:password'

# Helm timeout
helmDeployWaitSeconds: 600
additionalParameters: ['--timeout=10m']
```

### GitOps
```yaml
# Authentication
username: 'gitops-bot'
password: 'vault:github:token'
```

---

## Security Best Practices

1. **Credentials**: Use Vault or Jenkins Credentials Manager
2. **Image Scanning**: Scan images before deployment
3. **RBAC**: Apply least-privilege principles
4. **Secrets**: Never commit secrets to Git
5. **TLS**: Enable TLS for all connections
6. **Network Policies**: Implement network segmentation
7. **Audit Logs**: Enable deployment logging

---

## Example Scenarios

### Cloud Foundry Zero-Downtime Deployment
```yaml
steps:
  cloudFoundryDeploy:
    org: 'production'
    space: 'prod'
    deployType: 'standard'
    cfNativeDeployParameters: '--strategy rolling'
    manifestVariables:
      - 'instances=5'
      - 'memory=2048M'
```

### Kubernetes Production Deployment
```yaml
steps:
  kubernetesDeploy:
    deployTool: 'helm3'
    namespace: 'production'
    chartPath: './helm/my-app'
    deploymentName: 'my-app-prod'
    helmValues: ['values.yaml', 'values-production.yaml']
    additionalParameters:
      - '--atomic'
      - '--wait'
      - '--timeout=600s'
    runHelmTests: true
```

### GitOps Workflow
```yaml
# Build and push image
- kanikoExecute

# Update manifest
- gitopsUpdateDeployment:
    tool: 'kustomize'
    serverUrl: 'https://github.com'
    branchName: 'main'
    filePath: 'apps/production/kustomization.yaml'
    containerImageNameTag: 'my-app:${BUILD_NUMBER}'
    commitMessage: 'chore: update app to ${BUILD_NUMBER}'
```

### Multi-Region Deployment
```yaml
steps:
  multicloudDeploy:
    parallelExecution: true
    cfTargets:
      - apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'
        org: 'org'
        space: 'prod'
        manifest: 'manifest-eu.yml'
      - apiEndpoint: 'https://api.cf.us10.hana.ondemand.com'
        org: 'org'
        space: 'prod'
        manifest: 'manifest-us.yml'
      - apiEndpoint: 'https://api.cf.ap10.hana.ondemand.com'
        org: 'org'
        space: 'prod'
        manifest: 'manifest-ap.yml'
```

### Infrastructure + Application Pipeline
```yaml
stages:
  - name: Provision
    steps:
      - terraformExecute:
          command: 'apply'
          init: true
          workspace: 'production'
          additionalArgs: ['-auto-approve']

  - name: Deploy
    steps:
      - kubernetesDeploy:
          deployTool: 'helm3'
          chartPath: './charts/app'
          namespace: 'production'
```

---

## Additional Resources

- [Cloud Foundry Docs](https://docs.cloudfoundry.org/)
- [Kubernetes Docs](https://kubernetes.io/docs/)
- [Helm Docs](https://helm.sh/docs/)
- [Terraform Docs](https://www.terraform.io/docs/)
- [GitOps Guide](https://www.gitops.tech/)
- [SAP BTP Documentation](https://help.sap.com/btp)
- [Project Piper Documentation](https://www.project-piper.io/)

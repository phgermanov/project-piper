# Acceptance Stage

## Overview

The Acceptance stage deploys the application to a test environment and executes automated end-to-end acceptance tests to ensure new functionality works correctly and there are no regressions in existing functionality. This stage is crucial for SAP's functional correctness corporate requirement FC-2.

## Stage Purpose

The Acceptance stage performs the following key functions:

- Deploys application to acceptance/test environment
- Executes infrastructure provisioning (Terraform, GitOps)
- Downloads artifacts from staging service
- Deploys to Cloud Foundry or Kubernetes
- Runs end-to-end acceptance tests (Newman, UIVeri5, Gauge)
- Uploads test results and mappings to Cumulus
- Validates functional correctness

## When the Stage Runs

The Acceptance stage runs when:

- The Init stage activates it based on configuration
- `fromJSON(inputs.active-stages-map).Acceptance == true`
- **Only on productive branch**: `inputs.on-productive-branch == 'true'`
- After successful Build and Integration stages

## Steps in the Stage

### 1. Checkout repository
- **Action**: `actions/checkout@v4.3.0`
- **Purpose**: Checks out source code and test definitions
- **Configuration**: Supports submodules and LFS

### 2. Setup Go for development builds
- **Action**: `actions/setup-go@v6.0.0`
- **Condition**: Only for development Piper versions
- **Go Version**: 1.24

### 3. Checkout global extension
- **Action**: `actions/checkout@v4`
- **Condition**: When extensibility is enabled
- **Path**: `.pipeline/tmp/global_extensions`

### 4. Retrieve System Trust session token
- **Action**: `project-piper/system-trust-composite-action`
- **Permissions Required**: `id-token: write`
- **Error Handling**: Continues on error

### 5. preAcceptance
- **Condition**: When extensibility is enabled
- **Purpose**: Custom pre-acceptance test setup
- **Use Cases**: Configure test environment, setup test data, start monitoring

### Infrastructure and Deployment Steps

#### 6. terraformExecute
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `terraformExecute`
- **Condition**: Explicitly activated (inactive by default)
- **Purpose**: Provisions infrastructure using Terraform
- **Use Case**: Create cloud resources for test environment

#### 7. gitopsUpdateDeployment
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `gitopsUpdateDeployment`
- **Condition**: Explicitly activated (inactive by default)
- **Purpose**: Updates Kubernetes deployment manifest in Git repository
- **Flags**: `--username github-actions --password ${{ github.token }}`
- **Pattern**: GitOps deployment workflow

#### 8. sapDownloadArtifact
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapDownloadArtifact`
- **Condition**: Active for native builds or when Helm chart URL is available
- **Purpose**: Downloads build artifacts from staging service
- **Artifacts**: JAR, WAR, Docker images, Helm charts

#### 9. cloudFoundryDeploy
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `cloudFoundryDeploy`
- **Condition**: `cfSpace` or `cloudFoundry/space` is configured
- **Purpose**: Deploys application to Cloud Foundry
- **Supports**: Blue-green deployment, service bindings

#### 10. kubernetesDeploy
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `kubernetesDeploy`
- **Condition**: `deployTool` is helm, helm3, or kubectl
- **Flags**: `--githubToken ${{ github.token }}`
- **Purpose**: Deploys application to Kubernetes cluster
- **Supports**: Helm charts, kubectl manifests

### Testing Steps

#### 11. newmanExecute
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `newmanExecute`
- **Condition**: Newman collection file configured
- **Purpose**: Executes API tests using Newman (Postman)
- **Format**: Postman collection JSON

#### 12. uiVeri5ExecuteTests
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `uiVeri5ExecuteTests`
- **Condition**: `conf.js` file exists or test repository configured
- **Purpose**: Executes end-to-end UI tests using UIVeri5
- **Use Case**: SAP Fiori/UI5 application testing

#### 13. gaugeExecuteTests
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `gaugeExecuteTests`
- **Condition**: `*.spec` files exist or test repository configured
- **Purpose**: Executes behavior-driven tests using Gauge
- **Format**: Gauge specifications

### Result Upload Steps

#### 14. sapCumulusUpload (acceptance test results)
- **Flags**: `--filePattern **/TEST-*.xml --stepResultType acceptance-test`
- **Purpose**: Uploads JUnit XML test results

#### 15. sapCumulusUpload (requirement mapping)
- **Flags**: `--filePattern **/requirement.mapping --stepResultType requirement-mapping`
- **Purpose**: Uploads requirement traceability

#### 16. sapCumulusUpload (delivery mapping)
- **Flags**: `--filePattern **/delivery.mapping --stepResultType delivery-mapping`
- **Purpose**: Uploads delivery traceability

### 17. postAcceptance
- **Condition**: When extensibility is enabled
- **Purpose**: Custom post-acceptance test cleanup
- **Use Cases**: Cleanup test environment, generate reports, undeploy test services

### 18. Export pipeline environment
- **Purpose**: Exports CPE for downstream stages
- **Output**: `pipelineEnv`

## Configuration Options

### Stage Activation

```yaml
stages:
  Acceptance:
    # Explicitly activate stage (auto-activated based on conditions)
    active: true
```

### Cloud Foundry Deployment

```yaml
steps:
  cloudFoundryDeploy:
    # Cloud Foundry API endpoint
    apiEndpoint: 'https://api.cf.example.com'

    # Organization
    org: 'my-org'

    # Space
    space: 'acceptance'

    # Deployment type
    deployType: 'blue-green'  # or 'standard'

    # Manifest file
    manifest: 'manifest-acceptance.yml'

    # Application name
    appName: 'my-app-acceptance'

    # Credentials (from Vault)
    credentialsId: 'cf-credentials'
```

### Kubernetes Deployment

```yaml
steps:
  kubernetesDeploy:
    # Deploy tool
    deployTool: 'helm3'  # Options: helm, helm3, kubectl

    # Namespace
    namespace: 'acceptance'

    # Helm chart path or URL
    helmChartPath: './helm/my-app'

    # Values file
    helmValues:
      - 'helm/values-acceptance.yaml'

    # Container image
    image: 'my-registry.com/my-app:1.0.0'

    # Kubeconfig (from Vault)
    kubeConfig: 'kubeconfig-acceptance'

    # Ingress hostname
    ingressHosts:
      - 'my-app-acceptance.example.com'
```

### GitOps Deployment

```yaml
steps:
  gitopsUpdateDeployment:
    # Git repository for manifests
    gitRepository: 'my-org/k8s-manifests'

    # Branch to update
    branchName: 'acceptance'

    # Server URL (optional)
    serverUrl: 'https://github.com'

    # File path to update
    filePath: 'apps/my-app/deployment.yaml'

    # Container image
    containerName: 'my-app'
    containerImage: 'my-registry.com/my-app:1.0.0'
```

### Newman API Tests

```yaml
steps:
  newmanExecute:
    # Postman collection
    newmanCollection: 'tests/api/postman-collection.json'

    # Environment file
    newmanEnvironment: 'tests/api/acceptance-environment.json'

    # Global variables
    newmanGlobals: 'tests/api/globals.json'

    # Newman run options
    newmanRunOptions:
      - '--folder acceptance-tests'
      - '--timeout-request 30000'

    # Fail on error
    failOnError: true
```

### UIVeri5 UI Tests

```yaml
steps:
  uiVeri5ExecuteTests:
    # Test repository (optional)
    testRepository: 'https://github.com/my-org/ui-tests.git'

    # Configuration file
    confFilePath: 'conf.js'

    # Test files
    testPath: 'tests/acceptance'

    # Base URL
    baseUrl: 'https://my-app-acceptance.example.com'

    # Browser
    browsers:
      - 'chrome'

    # Screenshots on failure
    takeScreenshots: true
```

### Gauge Tests

```yaml
steps:
  gaugeExecuteTests:
    # Test repository (optional)
    testRepository: 'https://github.com/my-org/gauge-tests.git'

    # Specification path
    specPath: 'specs/acceptance'

    # Language
    language: 'java'

    # Environment
    environment: 'acceptance'

    # Tags
    tags: 'acceptance,smoke'
```

### Terraform Infrastructure

```yaml
steps:
  terraformExecute:
    # Terraform command
    command: 'apply'

    # Workspace
    terraformWorkspace: 'acceptance'

    # Variables file
    terraformVariables:
      - 'terraform-acceptance.tfvars'

    # Backend configuration
    terraformBackendConfig:
      bucket: 'terraform-state-bucket'
      key: 'acceptance/terraform.tfstate'
```

## Example Usage

### Basic Acceptance Stage

```yaml
jobs:
  acceptance:
    needs: [init, build]
    uses: project-piper/piper-pipeline-github/.github/workflows/acceptance.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.build.outputs.pipeline-env }}
    secrets: inherit
```

### Acceptance with Custom Environment

```yaml
jobs:
  acceptance:
    needs: [init, build]
    uses: project-piper/piper-pipeline-github/.github/workflows/acceptance.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.build.outputs.pipeline-env }}
      environment: 'Acceptance Test'  # GitHub environment with protection rules
    secrets: inherit
```

### Complete Pipeline with Acceptance

```yaml
name: Production Pipeline

on:
  push:
    branches: [main]

jobs:
  init:
    uses: project-piper/piper-pipeline-github/.github/workflows/init.yml@main

  build:
    needs: init
    uses: project-piper/piper-pipeline-github/.github/workflows/build.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.init.outputs.pipeline-env }}
    secrets: inherit

  integration:
    needs: [init, build]
    uses: project-piper/piper-pipeline-github/.github/workflows/integration.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.build.outputs.pipeline-env }}
    secrets: inherit

  acceptance:
    needs: [init, build, integration]
    uses: project-piper/piper-pipeline-github/.github/workflows/acceptance.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.integration.outputs.pipeline-env }}
      environment: 'Piper Acceptance'
    secrets: inherit
```

## Test Structure Examples

### Newman/Postman Collection Structure

```
tests/
└── api/
    ├── postman-collection.json       # API test collection
    ├── acceptance-environment.json   # Environment variables
    ├── globals.json                  # Global variables
    └── requirement.mapping           # Test-to-requirement mapping
```

### UIVeri5 Test Structure

```
tests/
└── ui/
    ├── conf.js                       # UIVeri5 configuration
    ├── specs/
    │   ├── LoginPage.spec.js
    │   ├── DashboardPage.spec.js
    │   └── CheckoutFlow.spec.js
    ├── pages/
    │   ├── LoginPage.js
    │   └── DashboardPage.js
    └── requirement.mapping
```

### Gauge Test Structure

```
tests/
└── gauge/
    ├── specs/
    │   ├── login.spec
    │   └── checkout.spec
    ├── src/
    │   └── test/
    │       └── java/
    │           └── StepImplementation.java
    ├── env/
    │   └── acceptance/
    │       └── acceptance.properties
    └── requirement.mapping
```

## Troubleshooting

### Deployment Failures

**Cloud Foundry**:
1. Verify CF API endpoint and credentials
2. Check organization and space exist
3. Review manifest file for errors
4. Check service bindings

**Kubernetes**:
1. Verify kubeconfig credentials
2. Check namespace exists
3. Validate Helm chart syntax
4. Review pod logs for errors

### Test Execution Issues

**Newman/API Tests**:
1. Verify collection and environment files exist
2. Check base URL accessibility
3. Review authentication configuration
4. Validate request/response formats

**UIVeri5/UI Tests**:
1. Ensure application is deployed and accessible
2. Check browser driver compatibility
3. Verify selectors and page objects
4. Review screenshot logs

**Gauge Tests**:
1. Verify specification syntax
2. Check step implementations
3. Validate test data
4. Review execution logs

### Environment Access Issues

1. **Network connectivity**: Verify pipeline can reach test environment
2. **Credentials**: Check Vault credentials are correct
3. **Firewall rules**: Ensure pipeline runner has access
4. **DNS resolution**: Verify hostnames resolve correctly

## Best Practices

1. **Environment Isolation**: Use dedicated acceptance environment
2. **Data Management**: Use test data that doesn't affect other environments
3. **Test Independence**: Ensure tests can run independently
4. **Deployment Strategy**: Use blue-green or canary deployments
5. **Test Coverage**: Focus on critical user journeys
6. **Performance**: Keep acceptance tests < 30 minutes
7. **Flaky Tests**: Fix or quarantine unreliable tests
8. **Monitoring**: Collect metrics during test execution
9. **Cleanup**: Always cleanup resources after testing
10. **Documentation**: Document test scenarios and expected outcomes

## GitHub Environment Protection

Configure GitHub environment for manual approval:

1. Go to repository Settings > Environments
2. Create environment named "Piper Acceptance"
3. Configure protection rules:
   - Required reviewers
   - Wait timer
   - Deployment branches

## Compliance

The Acceptance stage supports:

- **FC-2 Compliance**: Functional correctness validation
- **Requirement Traceability**: Test-to-requirement mapping
- **Delivery Mapping**: Feature-to-test coverage
- **Audit Trail**: All test executions logged in Cumulus
- **Policy Evidence**: Test results for policy evaluation

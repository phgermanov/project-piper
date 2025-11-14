# Performance Stage

## Overview

The Performance stage executes automated performance tests to validate that the application meets performance requirements and to detect performance regressions. This stage can include various types of performance testing from component-level to full load testing.

## Stage Purpose

The Performance stage performs the following key functions:

- Deploys application to performance test environment
- Executes infrastructure provisioning for performance testing
- Downloads artifacts from staging service
- Deploys to Cloud Foundry or Kubernetes
- Runs performance tests (single-user, load, stress)
- Executes SUPA (Single User Performance Analysis) tests
- Uploads performance test results to Cumulus
- Validates performance requirements and SLAs

## When the Stage Runs

The Performance stage runs when:

- The Init stage activates it based on configuration
- `fromJSON(inputs.active-stages-map).Performance == true`
- **Only on productive branch**: `inputs.on-productive-branch == 'true'`
- After successful Build, Integration, and Acceptance stages
- Performance tests are configured in the project

## Performance Test Categorization

1. **Unit Performance Tests**: Class/function level tests (run in Build stage)
2. **Component Performance Tests**: Service-level tests (e.g., REST APIs)
3. **Single User Performance Tests**: End-to-end including UI (SUPA, Fiori/UI5)
4. **Load Tests**: Simulate real production scenarios and loads

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

### 5. prePerformance
- **Condition**: When extensibility is enabled
- **Purpose**: Custom pre-performance test setup
- **Use Cases**: Configure load generators, prepare test data, start monitoring

### Infrastructure and Deployment Steps

#### 6. terraformExecute
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `terraformExecute`
- **Condition**: Explicitly activated (inactive by default)
- **Purpose**: Provisions infrastructure using Terraform
- **Use Case**: Create cloud resources optimized for performance testing

#### 7. gitopsUpdateDeployment
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `gitopsUpdateDeployment`
- **Condition**: Explicitly activated (inactive by default)
- **Flags**: `--username github-actions --password ${{ github.token }}`
- **Purpose**: Updates Kubernetes deployment manifest for GitOps workflow

#### 8. sapDownloadArtifact
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapDownloadArtifact`
- **Condition**: Active for native builds or when Helm chart URL is available
- **Purpose**: Downloads build artifacts from staging service
- **Note**: Downloads the same version tested in Acceptance stage

#### 9. cloudFoundryDeploy
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `cloudFoundryDeploy`
- **Condition**: `cfSpace` or `cloudFoundry/space` is configured
- **Purpose**: Deploys application to Cloud Foundry performance environment

#### 10. kubernetesDeploy
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `kubernetesDeploy`
- **Condition**: `deployTool` is helm, helm3, or kubectl
- **Flags**: `--githubToken ${{ github.token }}`
- **Purpose**: Deploys application to Kubernetes cluster

### Performance Testing Steps

#### 11. sapSUPAExecuteTests
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapSUPAExecuteTests`
- **Condition**: Explicitly activated (inactive by default)
- **Purpose**: Executes Single User Performance Analysis tests
- **Use Case**: SAP Fiori/UI5 application performance testing
- **Metrics**: Page load times, navigation performance, resource usage

#### 12. sapCumulusUpload
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapCumulusUpload`
- **Condition**: Active when configured
- **Purpose**: Uploads performance test results to Cumulus
- **Results**: Performance metrics, test reports, trend data

### 13. postPerformance
- **Condition**: When extensibility is enabled
- **Purpose**: Custom post-performance test cleanup
- **Use Cases**: Cleanup test environment, generate reports, stop monitoring

### 14. Export pipeline environment
- **Purpose**: Exports CPE for downstream stages
- **Output**: `pipelineEnv`

## Configuration Options

### Stage Activation

```yaml
stages:
  Performance:
    # Explicitly activate stage
    active: true
```

### SUPA Configuration

```yaml
steps:
  sapSUPAExecuteTests:
    # SUPA server URL
    supaServerUrl: 'https://supa.example.com'

    # Test suite
    testSuite: 'performance-tests'

    # Application URL
    applicationUrl: 'https://my-app-performance.example.com'

    # User credentials (from Vault)
    credentialsId: 'supa-credentials'

    # Test configuration
    testConfiguration:
      iterations: 5
      thinkTime: 2000
      timeout: 300000

    # Performance thresholds
    thresholds:
      pageLoadTime: 3000
      navigationTime: 1000
```

### Gatling Configuration (Jenkins)

For Jenkins GPP, Gatling is available:

```yaml
steps:
  gatlingExecuteTests:
    # POM path
    pomPath: 'performance-tests/pom.xml'

    # Simulation class
    simulation: 'com.example.MySimulation'

    # Gatling options
    options:
      - '-rf results'
      - '-sf simulations'
```

### Cloud Foundry Performance Environment

```yaml
steps:
  cloudFoundryDeploy:
    # API endpoint
    apiEndpoint: 'https://api.cf.example.com'

    # Organization and space
    org: 'my-org'
    space: 'performance'

    # Manifest with performance tuning
    manifest: 'manifest-performance.yml'

    # Instances and memory for load testing
    instances: 3
    memory: '2G'
```

### Kubernetes Performance Environment

```yaml
steps:
  kubernetesDeploy:
    deployTool: 'helm3'
    namespace: 'performance'
    helmChartPath: './helm/my-app'

    # Values file with performance settings
    helmValues:
      - 'helm/values-performance.yaml'

    # Resources for performance testing
    additionalParameters:
      - '--set replicaCount=3'
      - '--set resources.requests.memory=2Gi'
      - '--set resources.requests.cpu=1000m'
```

### Terraform Infrastructure for Performance

```yaml
steps:
  terraformExecute:
    command: 'apply'
    terraformWorkspace: 'performance'

    # Variables for performance environment
    terraformVariables:
      - 'terraform-performance.tfvars'

    # Example: Larger instances for load testing
    # In terraform-performance.tfvars:
    # instance_type = "m5.xlarge"
    # min_instances = 3
    # max_instances = 10
```

### Cumulus Upload Configuration

```yaml
steps:
  sapCumulusUpload:
    pipelineId: 'your-pipeline-id'

    # Performance test results patterns
    filePattern: '**/performance-results.json,**/gatling/**/*.html'
    stepResultType: 'performance-test'
```

## Example Usage

### Basic Performance Stage

```yaml
jobs:
  performance:
    needs: [init, build, acceptance]
    uses: project-piper/piper-pipeline-github/.github/workflows/performance.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.acceptance.outputs.pipeline-env }}
    secrets: inherit
```

### Performance with Custom Environment

```yaml
jobs:
  performance:
    needs: [init, build, acceptance]
    uses: project-piper/piper-pipeline-github/.github/workflows/performance.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.acceptance.outputs.pipeline-env }}
      environment: 'Piper Performance'
      extensibility-enabled: true
      global-extensions-repository: 'my-org/pipeline-extensions'
    secrets: inherit
```

### Complete Pipeline with Performance

```yaml
name: Production Pipeline

on:
  push:
    branches: [main]

permissions:
  id-token: write
  contents: read

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

  acceptance:
    needs: [init, build]
    uses: project-piper/piper-pipeline-github/.github/workflows/acceptance.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.build.outputs.pipeline-env }}
    secrets: inherit

  performance:
    needs: [init, build, acceptance]
    uses: project-piper/piper-pipeline-github/.github/workflows/performance.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.acceptance.outputs.pipeline-env }}
      environment: 'Piper Performance'
    secrets: inherit
```

## Performance Test Examples

### Extension: JMeter Performance Test

Create `.pipeline/extensions/prePerformance.sh`:

```bash
#!/bin/bash
set -e

echo "Starting JMeter performance tests..."

# Download JMeter if not present
if [ ! -d "apache-jmeter-5.5" ]; then
    wget https://archive.apache.org/dist/jmeter/binaries/apache-jmeter-5.5.tgz
    tar -xzf apache-jmeter-5.5.tgz
fi

# Run JMeter tests
./apache-jmeter-5.5/bin/jmeter -n \
    -t tests/performance/LoadTest.jmx \
    -l results/jmeter-results.jtl \
    -e -o results/jmeter-report \
    -Jthreads=50 \
    -Jrampup=30 \
    -Jduration=300

echo "JMeter tests completed"
```

### Extension: k6 Load Testing

Create `.pipeline/extensions/prePerformance.sh`:

```bash
#!/bin/bash
set -e

echo "Running k6 load tests..."

# Install k6
if ! command -v k6 &> /dev/null; then
    wget https://github.com/grafana/k6/releases/download/v0.45.0/k6-v0.45.0-linux-amd64.tar.gz
    tar -xzf k6-v0.45.0-linux-amd64.tar.gz
    mv k6-v0.45.0-linux-amd64/k6 /usr/local/bin/
fi

# Run k6 tests
k6 run \
    --vus 100 \
    --duration 5m \
    --out json=results/k6-results.json \
    tests/performance/load-test.js

echo "k6 tests completed"
```

### Custom Performance Metrics Collection

Create `.pipeline/extensions/postPerformance.sh`:

```bash
#!/bin/bash
set -e

echo "Collecting performance metrics..."

# Collect application metrics
curl -s http://my-app-performance.example.com/metrics > results/app-metrics.txt

# Collect infrastructure metrics (example with kubectl)
kubectl top pods -n performance > results/pod-metrics.txt
kubectl top nodes > results/node-metrics.txt

# Generate summary report
python3 scripts/generate-performance-report.py \
    --results results/ \
    --output results/performance-summary.json

echo "Performance metrics collected"
```

## Performance Test Structure

```
performance-tests/
├── simulations/              # Gatling simulations (Java/Scala)
│   └── MyLoadSimulation.scala
├── jmeter/                   # JMeter test plans
│   ├── LoadTest.jmx
│   └── StressTest.jmx
├── k6/                       # k6 scripts
│   ├── load-test.js
│   └── spike-test.js
├── supa/                     # SUPA test configurations
│   └── ui5-app-test.json
└── results/                  # Test results directory
    ├── jmeter-results.jtl
    ├── k6-results.json
    └── performance-summary.json
```

## Troubleshooting

### Performance Test Failures

1. **Timeout issues**:
   - Increase test timeout values
   - Check application and infrastructure resources
   - Verify network latency

2. **Resource constraints**:
   - Scale up performance environment
   - Check CPU/memory limits
   - Review application configuration

3. **Test flakiness**:
   - Add proper warm-up periods
   - Implement retry logic
   - Check for external dependencies

### Deployment Issues

1. **Environment not ready**:
   - Add health checks before testing
   - Implement wait conditions
   - Verify all services are running

2. **Resource allocation**:
   - Ensure sufficient resources for load
   - Configure auto-scaling if needed
   - Monitor resource usage

### Data and State Issues

1. **Test data management**:
   - Reset test data between runs
   - Use data generation scripts
   - Implement data cleanup

2. **State interference**:
   - Isolate performance environment
   - Clear caches before tests
   - Reset application state

## Best Practices

1. **Environment Sizing**: Size performance environment to match production
2. **Test Data**: Use realistic, production-like test data
3. **Warm-up**: Include warm-up period before measurements
4. **Duration**: Run tests long enough to reach steady state
5. **Monitoring**: Collect comprehensive metrics during tests
6. **Baseline**: Establish performance baselines for comparison
7. **Thresholds**: Define and enforce performance SLAs
8. **Trends**: Track performance trends over time
9. **Isolation**: Isolate performance tests from other activities
10. **Documentation**: Document test scenarios and expected results
11. **Parallel Execution**: Consider running different test types in parallel
12. **Custom Pipelines**: For long-running tests, use separate scheduled pipelines

## Performance Metrics to Collect

- **Response Times**: Min, max, average, percentiles (P50, P90, P95, P99)
- **Throughput**: Requests per second, transactions per second
- **Error Rates**: Percentage of failed requests
- **Resource Usage**: CPU, memory, disk I/O, network
- **Application Metrics**: Custom business metrics
- **Database Performance**: Query times, connection pool usage
- **External Services**: API call latencies

## Integration with Monitoring Tools

Configure monitoring integration in prePerformance extension:

```bash
# Example: Start Prometheus/Grafana monitoring
docker-compose -f monitoring/docker-compose.yml up -d

# Wait for monitoring to be ready
sleep 30

# Configure scrape targets
curl -X POST http://prometheus:9090/api/v1/admin/tsdb/snapshot
```

## GitHub Environment Protection

For performance testing with approval:

1. Create GitHub environment "Piper Performance"
2. Configure required reviewers
3. Set deployment protection rules
4. Configure environment-specific secrets

## Related Stages

- **Build Stage**: May include unit performance tests
- **Acceptance Stage**: Validates functionality before performance testing
- **Release Stage**: Benefits from performance validation

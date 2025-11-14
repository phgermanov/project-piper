# Integration Stage

## Overview

The Integration stage executes project-specific integration tests to validate that different components of the application work together correctly. This stage runs after the Build stage and before Acceptance testing.

## Stage Purpose

The Integration stage performs the following key functions:

- Executes backend integration tests
- Validates component interactions
- Tests API integrations
- Verifies database interactions
- Uploads test results to Cumulus for reporting and compliance
- Uploads requirement and delivery mappings for traceability

## When the Stage Runs

The Integration stage runs when:

- The Init stage activates it based on configuration
- `fromJSON(inputs.active-stages-map).Integration == true`
- **Only on productive branch**: `inputs.on-productive-branch == 'true'`
- Integration test files are detected (e.g., `integration-tests/pom.xml`)

## Steps in the Stage

### 1. Checkout repository
- **Action**: `actions/checkout@v4.3.0`
- **Purpose**: Checks out source code
- **Configuration**: Supports submodules and LFS

### 2. Setup Go for development builds
- **Action**: `actions/setup-go@v6.0.0`
- **Condition**: Only for development Piper versions
- **Go Version**: 1.24

### 3. Checkout global extension
- **Action**: `actions/checkout@v4`
- **Condition**: When extensibility is enabled
- **Path**: `.pipeline/tmp/global_extensions`
- **Purpose**: Checks out global pipeline extensions

### 4. Retrieve System Trust session token
- **Action**: `project-piper/system-trust-composite-action`
- **Purpose**: Obtains session token for System Trust integration
- **Permissions Required**: `id-token: write`
- **Error Handling**: Continues on error

### 5. preIntegration
- **Condition**: When extensibility is enabled
- **Purpose**: Executes custom pre-integration test extensions
- **Use Case**: Setup test databases, deploy test services, configure test environments

### 6. mavenExecuteIntegration
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `mavenExecuteIntegration`
- **Condition**: Active in step configuration
- **Purpose**: Executes Maven integration tests
- **Default Module**: `integration-tests/pom.xml`

### 7. sapCumulusUpload (integration test results)
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapCumulusUpload`
- **Flags**: `--filePattern **/TEST-*.xml --stepResultType integration-test`
- **Purpose**: Uploads JUnit test results from integration tests
- **Format**: JUnit XML format

### 8. sapCumulusUpload (requirement mapping)
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapCumulusUpload`
- **Flags**: `--filePattern **/requirement.mapping --stepResultType requirement-mapping`
- **Purpose**: Uploads requirement traceability mapping
- **Use Case**: Links test cases to requirements for compliance

### 9. sapCumulusUpload (delivery mapping)
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapCumulusUpload`
- **Flags**: `--filePattern **/delivery.mapping --stepResultType delivery-mapping`
- **Purpose**: Uploads delivery traceability mapping
- **Use Case**: Links deliverables to test results

### 10. postIntegration
- **Condition**: When extensibility is enabled
- **Purpose**: Executes custom post-integration test extensions
- **Use Case**: Cleanup test resources, generate reports, notify stakeholders

### 11. Export pipeline environment
- **Action**: `SAP/project-piper-action@v1.22`
- **Purpose**: Exports Common Pipeline Environment for downstream stages
- **Output**: `pipelineEnv` job output

## Configuration Options

### Stage Activation

Configure in `.pipeline/config.yml`:

```yaml
stages:
  Integration:
    # Activate stage explicitly (optional, auto-detected by file pattern)
    active: true
```

### Maven Integration Tests

```yaml
steps:
  mavenExecuteIntegration:
    # Path to integration tests POM
    pomPath: 'integration-tests/pom.xml'

    # Maven goals to execute
    goals: 'verify'

    # Additional Maven defines
    defines: '-Dskip.unit.tests=true'

    # Maven options
    mavenOptions: '-B -Dmaven.test.failure.ignore=true'

    # Test retry configuration
    retry: 1

    # Profile to activate
    profiles:
      - 'integration-tests'
```

### File Pattern Detection

The Integration stage is automatically activated when:

```yaml
# File pattern that triggers the stage
filePattern: 'integration-tests/pom.xml'
```

### Cumulus Upload Configuration

```yaml
steps:
  sapCumulusUpload:
    # Pipeline ID for Cumulus (required)
    pipelineId: 'your-pipeline-id'

    # Server URL (from Vault or config)
    serverUrl: 'https://cumulus.example.com'
```

### Requirement Mapping Format

Create `requirement.mapping` file in your test directory:

```
TEST_ID=REQ_ID
IntegrationTest1=REQ-123
IntegrationTest2=REQ-124,REQ-125
```

### Delivery Mapping Format

Create `delivery.mapping` file:

```
DELIVERY_ID=TEST_ID
DELIVERABLE-001=IntegrationTest1,IntegrationTest2
DELIVERABLE-002=IntegrationTest3
```

## Example Usage

### Basic Integration Stage

```yaml
jobs:
  integration:
    needs: [init, build]
    uses: project-piper/piper-pipeline-github/.github/workflows/integration.yml@main
    with:
      piper-version: 'latest'
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.build.outputs.pipeline-env }}
    secrets: inherit
```

### Integration with Extensions

```yaml
jobs:
  integration:
    needs: [init, build]
    uses: project-piper/piper-pipeline-github/.github/workflows/integration.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.build.outputs.pipeline-env }}
      extensibility-enabled: true
      global-extensions-repository: 'my-org/pipeline-extensions'
      global-extensions-ref: 'main'
    secrets: inherit
```

### Complete Pipeline Example

```yaml
name: CI/CD Pipeline

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
```

## Integration Test Patterns

### Maven Integration Tests Structure

```
project-root/
├── src/
│   └── main/
├── integration-tests/
│   ├── pom.xml                    # Integration tests POM
│   ├── src/
│   │   └── test/
│   │       └── java/
│   │           └── com/example/
│   │               └── integration/
│   │                   ├── ApiIntegrationTest.java
│   │                   └── DatabaseIntegrationTest.java
│   └── requirement.mapping        # Test-to-requirement mapping
└── pom.xml                        # Main project POM
```

### Integration Tests POM Example

```xml
<project>
    <modelVersion>4.0.0</modelVersion>
    <artifactId>integration-tests</artifactId>
    <packaging>jar</packaging>

    <dependencies>
        <!-- Main project -->
        <dependency>
            <groupId>com.example</groupId>
            <artifactId>main-app</artifactId>
            <version>${project.version}</version>
        </dependency>

        <!-- Test frameworks -->
        <dependency>
            <groupId>org.junit.jupiter</groupId>
            <artifactId>junit-jupiter</artifactId>
            <scope>test</scope>
        </dependency>

        <!-- REST Assured for API testing -->
        <dependency>
            <groupId>io.rest-assured</groupId>
            <artifactId>rest-assured</artifactId>
            <scope>test</scope>
        </dependency>
    </dependencies>

    <build>
        <plugins>
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-surefire-plugin</artifactId>
                <configuration>
                    <includes>
                        <include>**/*IntegrationTest.java</include>
                        <include>**/*IT.java</include>
                    </includes>
                </configuration>
            </plugin>
        </plugins>
    </build>
</project>
```

## Extension Points

### preIntegration Extension

Create `.pipeline/extensions/preIntegration.sh`:

```bash
#!/bin/bash
set -e

echo "Setting up integration test environment..."

# Start test database
docker-compose -f docker-compose.integration.yml up -d

# Wait for services to be ready
./scripts/wait-for-services.sh

# Seed test data
./scripts/seed-test-data.sh

echo "Integration test environment ready"
```

### postIntegration Extension

Create `.pipeline/extensions/postIntegration.sh`:

```bash
#!/bin/bash
set -e

echo "Cleaning up integration test environment..."

# Collect additional logs
docker-compose -f docker-compose.integration.yml logs > integration-logs.txt

# Stop and remove test containers
docker-compose -f docker-compose.integration.yml down -v

echo "Integration test environment cleaned up"
```

## Troubleshooting

### Integration Tests Not Running

1. **Check file pattern**: Ensure `integration-tests/pom.xml` exists
2. **Verify stage activation**: Check that Integration stage is in `active-stages-map`
3. **Branch requirement**: Integration stage only runs on productive branch
4. **Check step activation**: Verify `mavenExecuteIntegration` is active in configuration

### Test Failures

1. **Review test logs**: Check Maven output for specific test failures
2. **Environment setup**: Verify preIntegration extension completed successfully
3. **Service availability**: Ensure dependent services are running and accessible
4. **Test data**: Verify test data is properly seeded

### Cumulus Upload Failures

1. **Verify pipeline ID**: Ensure `pipelineId` is configured correctly
2. **Check file patterns**: Verify test result files match the patterns
3. **Network connectivity**: Check connectivity to Cumulus server
4. **Credentials**: Verify Cumulus credentials in Vault

### Extension Issues

1. **Script permissions**: Ensure extension scripts are executable
2. **Extension path**: Verify global extensions repository and path
3. **Error handling**: Check extension scripts for proper error handling

## Best Practices

1. **Test Isolation**: Ensure integration tests are independent and can run in any order
2. **Environment Management**: Use Docker Compose for consistent test environments
3. **Data Management**: Use test data fixtures or generate test data programmatically
4. **Test Coverage**: Focus integration tests on critical paths and component interactions
5. **Resource Cleanup**: Always clean up test resources in postIntegration
6. **Requirement Traceability**: Maintain requirement.mapping for compliance
7. **Test Performance**: Keep integration tests fast (target < 10 minutes)
8. **Retry Logic**: Configure retry for flaky tests, but fix root causes
9. **Logging**: Collect comprehensive logs for troubleshooting
10. **Parallel Execution**: Consider test parallelization for faster execution

## Compliance and Reporting

The Integration stage supports various compliance requirements:

- **Test Results**: JUnit XML format uploaded to Cumulus
- **Requirement Mapping**: Links tests to functional requirements
- **Delivery Mapping**: Links deliverables to test coverage
- **Policy Evidence**: Results available for policy evaluation
- **Audit Trail**: All test executions tracked in Cumulus

## Related Stages

- **Build Stage**: Builds the application before integration testing
- **Acceptance Stage**: Runs end-to-end tests after integration tests
- **Performance Stage**: May include integration performance tests

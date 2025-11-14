# Project Piper Action - Features and Capabilities

## Table of Contents

1. [Step Execution](#step-execution)
2. [Binary Management and Caching](#binary-management-and-caching)
3. [Configuration Management](#configuration-management)
4. [Docker Container Support](#docker-container-support)
5. [Sidecar Containers](#sidecar-containers)
6. [Common Pipeline Environment (CPE)](#common-pipeline-environment-cpe)
7. [Custom Defaults](#custom-defaults)
8. [Secrets Management](#secrets-management)
9. [Enterprise Features](#enterprise-features)
10. [Stage and Step Activation](#stage-and-step-activation)

---

## Step Execution

### Overview

The core functionality of the action is executing Piper steps. Each step is a self-contained unit that performs a specific CI/CD task.

### Features

- **200+ Available Steps**: Execute any step from the Piper library
- **Flexible Parameters**: Pass flags and arguments to customize step behavior
- **Exit Code Handling**: Automatic failure detection and reporting
- **Verbose Output**: Optional detailed logging with `--verbose` flag
- **Telemetry Control**: Disable telemetry with `--noTelemetry` flag

### Example Steps by Category

#### Build Steps
- `mavenBuild` - Build Maven projects
- `npmExecuteScripts` - Execute npm scripts
- `golangBuild` - Build Go applications
- `mtaBuild` - Build Multi-Target Applications
- `kanikoExecute` - Build Docker images with Kaniko

#### Test Steps
- `npmExecuteTests` - Run npm tests
- `mavenExecuteTest` - Run Maven tests
- `batsExecuteTests` - Execute Bash Automated Testing
- `gaugeExecuteTests` - Run Gauge tests

#### Security & Compliance
- `detectExecuteScan` - Run BlackDuck/Synopsis Detect scans
- `checkmarxExecuteScan` - Execute Checkmarx security scans
- `checkmarxOneExecuteScan` - Execute Checkmarx One scans
- `fortifyExecuteScan` - Run Fortify scans
- `whitesourceExecuteScan` - Execute WhiteSource scans
- `sonarExecuteScan` - Run SonarQube analysis

#### Deployment Steps
- `cloudFoundryDeploy` - Deploy to Cloud Foundry
- `kubernetesDeploy` - Deploy to Kubernetes
- `helmExecute` - Execute Helm charts
- `azureBlobUpload` - Upload to Azure Blob Storage
- `awsS3Upload` - Upload to AWS S3

#### Artifact Management
- `artifactPrepareVersion` - Prepare artifact version
- `nexusUpload` - Upload to Nexus repository
- `mavenPublish` - Publish Maven artifacts

---

## Binary Management and Caching

### Automatic Binary Download

The action automatically downloads the appropriate Piper binary based on your configuration:

```yaml
- uses: SAP/project-piper-action@main
  with:
    step-name: version
    piper-version: latest  # or specific version like v1.300.0
```

### Caching Mechanism

**How It Works:**

1. Binary is downloaded to `{workspace}/{version}/piper` or `{workspace}/{version}/sap-piper`
2. Version is normalized: `v1.300.0` becomes `v1_300_0`
3. Subsequent steps in the same job reuse the cached binary
4. Different versions are cached separately

**Benefits:**

- Faster workflow execution (no repeated downloads)
- Offline capability (after initial download)
- Predictable file locations
- Reduced network usage

### Version Selection Strategies

#### Latest Release (Recommended)
```yaml
piper-version: latest
```
- Uses the most recent stable release
- Balances features and stability
- Recommended for most use cases

#### Specific Version (Maximum Stability)
```yaml
piper-version: v1.300.0
```
- Pins to exact version
- Predictable behavior across runs
- Best for production pipelines

#### Master Branch (Bleeding Edge)
```yaml
piper-version: master
```
- Latest build from master branch
- May include unreleased features
- Less stable, use for testing only

#### Development Mode (Advanced)
```yaml
piper-version: devel:SAP:jenkins-library:ff8df33b8ab17c19e9f4c48472828ed809d4496a
```
- Build from specific commit
- Format: `devel:owner:repo:commit-sha`
- Requires GitHub token
- Used for testing patches or features

### Binary Sources

**Open Source Binary:**
- Repository: `SAP/jenkins-library`
- Binary name: `piper`
- Default for most steps

**SAP Internal Binary:**
- Repository: Configurable via `sap-piper-repository`
- Binary name: `sap-piper`
- Used for enterprise/internal steps
- Requires authentication

---

## Configuration Management

### Configuration Hierarchy

Piper uses a layered configuration system (highest to lowest precedence):

1. **Command-line flags** (via `flags` input)
2. **Environment variables** (`PIPER_*`)
3. **Custom defaults** (from other repos)
4. **Project configuration** (`.pipeline/config.yml`)
5. **Step defaults** (built into Piper)

### Configuration File

Create `.pipeline/config.yml` in your repository:

```yaml
general:
  gitUrl: https://github.com/myorg/myrepo
  gitBranch: main

stages:
  build:
    mavenBuild:
      publish: true
      createBOM: true

steps:
  mavenBuild:
    goals: clean install
    defines: -DskipTests=false
```

### Context-Aware Configuration

The action automatically reads context-specific configuration:

- **Stage Name**: Derived from `GITHUB_JOB`
- **Step Name**: From action input
- **Context Config**: Retrieved via `getConfig` step

Example:
```yaml
jobs:
  build:  # This becomes the stage name
    runs-on: ubuntu-latest
    steps:
      - uses: SAP/project-piper-action@main
        with:
          step-name: mavenBuild  # Step name
```

Configuration lookup: `stages.build.mavenBuild` → `steps.mavenBuild` → defaults

### Environment Variable Configuration

Configure steps using environment variables:

```yaml
- uses: SAP/project-piper-action@main
  with:
    step-name: mavenBuild
  env:
    PIPER_mavenExecute_defines: "-DskipTests=true"
    PIPER_mavenExecute_pomPath: "app/pom.xml"
```

---

## Docker Container Support

### Overview

Run Piper steps inside Docker containers for consistent, isolated environments.

### Basic Docker Execution

```yaml
- uses: SAP/project-piper-action@main
  with:
    step-name: npmExecuteScripts
    docker-image: node:18
    flags: 'install build'
```

### Docker Configuration

#### From Action Input
```yaml
docker-image: maven:3.9-jdk-17
docker-options: '--memory=4g --cpus=2'
```

#### From Step Configuration
```yaml
# .pipeline/config.yml
steps:
  mavenBuild:
    dockerImage: maven:3.9-jdk-17
    dockerOptions:
      - '--memory=4g'
      - '--cpus=2'
```

### Container Features

**Automatic Workspace Mounting:**
- Current directory mounted at same path in container
- Piper binary mounted to `/piper`
- Working directory preserved

**Environment Variable Propagation:**
- GitHub Actions variables (`GITHUB_*`)
- Vault credentials (`PIPER_vault*`)
- Proxy settings (`http_proxy`, `https_proxy`, `no_proxy`)
- Custom variables (via `docker-env-vars`)

**Network Management:**
- Automatic network creation when using sidecars
- Network aliases for service discovery
- Cleanup after job completion

### Docker Environment Variables

Pass custom environment variables as JSON:

```yaml
docker-env-vars: '{"DATABASE_URL": "postgres://localhost:5432/testdb", "API_KEY": "test-key"}'
```

Or from step configuration:
```yaml
# .pipeline/config.yml
steps:
  npmExecuteScripts:
    dockerEnvVars:
      DATABASE_URL: postgres://localhost:5432/testdb
      API_KEY: test-key
```

### Container Lifecycle

1. **Start**: Container started with `docker run --detach`
2. **Execute**: Step runs inside container via `docker exec`
3. **Cleanup**: Container automatically stopped with `--time=1`

---

## Sidecar Containers

### Overview

Start additional containers alongside your main execution environment (e.g., databases, message queues, mock services).

### Basic Usage

```yaml
- uses: SAP/project-piper-action@main
  with:
    step-name: npmExecuteScripts
    docker-image: node:18
    sidecar-image: postgres:15
    sidecar-env-vars: 'POSTGRES_PASSWORD=test123'
```

### Network Configuration

When a sidecar is specified:

1. A Docker network is automatically created
2. Both containers join the network
3. Main container can access sidecar by name or alias
4. Network is removed after job completion

### Example: Integration Tests with Database

```yaml
- uses: SAP/project-piper-action@main
  with:
    step-name: npmExecuteScripts
    flags: 'test:integration'
    docker-image: node:18
    sidecar-image: postgres:15
    sidecar-options: '--health-cmd="pg_isready" --health-interval=10s'
    sidecar-env-vars: '{"POSTGRES_DB": "testdb", "POSTGRES_PASSWORD": "testpass"}'
    docker-env-vars: '{"DATABASE_URL": "postgresql://postgres:testpass@postgres:5432/testdb"}'
```

### Supported Sidecar Scenarios

- **Databases**: PostgreSQL, MySQL, MongoDB, Redis
- **Message Queues**: RabbitMQ, Kafka
- **Mock Services**: WireMock, Mockserver
- **Tools**: Selenium, LocalStack

---

## Common Pipeline Environment (CPE)

### Overview

The CPE is Piper's mechanism for persisting and sharing state across steps and jobs.

### How It Works

```
Step 1 (Job A)         Step 2 (Job A)         Step 3 (Job B)
     │                       │                       │
     ▼                       ▼                       ▼
┌─────────┐            ┌─────────┐            ┌─────────┐
│ Write   │────────────│  Read   │            │  Read   │
│ to CPE  │            │ from CPE│            │ from CPE│
└─────────┘            └─────────┘            └─────────┘
     │                       │                       │
     ▼                       ▼                       ▼
[.pipeline/commonPipelineEnvironment/]        [Artifact]
```

### Storage Location

Data is stored in `.pipeline/commonPipelineEnvironment/` with subdirectories:

- `git/` - Git information (commit, branch, etc.)
- `mtaBuild/` - MTA build artifacts
- `artifactVersion/` - Version information
- `custom/` - Custom data from steps

### Exporting Pipeline Environment

Export CPE from one job:

```yaml
job-a:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4

    - uses: SAP/project-piper-action@main
      id: build
      with:
        step-name: artifactPrepareVersion
        export-pipeline-environment: true

    - name: Upload CPE
      uses: actions/upload-artifact@v4
      with:
        name: pipeline-env
        path: .pipeline/
```

### Importing Pipeline Environment

Import CPE in another job:

```yaml
job-b:
  needs: job-a
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4

    - name: Download CPE
      uses: actions/download-artifact@v4
      with:
        name: pipeline-env
        path: .pipeline/

    - uses: SAP/project-piper-action@main
      with:
        step-name: cloudFoundryDeploy
      env:
        PIPER_ACTION_PIPELINE_ENV: ${{ needs.job-a.outputs.pipelineEnv }}
```

### Use Cases

1. **Version Management**: Share artifact version across build and deploy jobs
2. **Build Artifacts**: Pass information about built artifacts to deployment
3. **Git Information**: Share commit SHA, branch name across jobs
4. **Custom Data**: Store and retrieve custom values between steps

---

## Custom Defaults

### Overview

Share common configuration across multiple repositories by loading defaults from external sources.

### Single Custom Defaults File

```yaml
- uses: SAP/project-piper-action@main
  with:
    step-name: mavenBuild
    custom-defaults-paths: 'config/piper-defaults.yml'
```

### Multiple Custom Defaults Files

```yaml
custom-defaults-paths: 'config/company-defaults.yml,config/team-defaults.yml'
```

Files are loaded in order, later files override earlier ones.

### Loading from Other Repositories

```yaml
custom-defaults-paths: 'myorg/shared-config/defaults.yml@v1.0.0,myorg/team-config/overrides.yml@main'
```

Format: `org/repo/path/to/file.yml@ref`

- `org/repo`: GitHub organization and repository
- `path/to/file.yml`: Path within repository
- `@ref`: Git reference (branch, tag, or SHA)

### HTTP URLs

```yaml
custom-defaults-paths: 'https://raw.githubusercontent.com/myorg/config/main/defaults.yml'
```

### Example Custom Defaults File

```yaml
# company-defaults.yml
general:
  productionBranch: main
  verbose: true

steps:
  mavenBuild:
    publish: true
    createBOM: true

  npmExecuteScripts:
    dockerImage: node:18
    install: true

  sonarExecuteScan:
    serverUrl: https://sonarqube.company.com
    organization: myorg
```

### Configuration Precedence with Custom Defaults

1. Step flags
2. Environment variables
3. Custom defaults (in order specified)
4. Project configuration
5. Built-in defaults

---

## Secrets Management

### Vault Integration

Piper can load secrets from HashiCorp Vault:

```yaml
- uses: SAP/project-piper-action@main
  with:
    step-name: mavenBuild
  env:
    PIPER_vaultAppRoleID: ${{ secrets.VAULT_ROLE_ID }}
    PIPER_vaultAppRoleSecretID: ${{ secrets.VAULT_SECRET_ID }}
```

### Vault Configuration

In `.pipeline/config.yml`:

```yaml
general:
  vaultBasePath: secret/piper
  vaultPath: secret
  vaultServerUrl: https://vault.company.com

steps:
  mavenBuild:
    vaultCredentialPath: maven
    vaultCredentialKeys: ['username', 'password']
```

### GitHub Secrets

Pass GitHub secrets to steps:

```yaml
- uses: SAP/project-piper-action@main
  with:
    step-name: cloudFoundryDeploy
    flags: '--username $CF_USER --password $CF_PASSWORD'
  env:
    CF_USER: ${{ secrets.CF_USERNAME }}
    CF_PASSWORD: ${{ secrets.CF_PASSWORD }}
```

### Docker Environment Secrets

Pass secrets to Docker containers:

```yaml
docker-env-vars: '{"NPM_TOKEN": "${{ secrets.NPM_TOKEN }}", "MAVEN_PASSWORD": "${{ secrets.MAVEN_PWD }}"}'
```

---

## Enterprise Features

### SAP Internal Steps

Some steps require the SAP-internal `sap-piper` binary:

```yaml
- uses: SAP/project-piper-action@main
  with:
    step-name: sapInternalStep
    sap-piper-version: latest
    sap-piper-owner: SAP-internal
    sap-piper-repository: jenkins-library-internal
    github-enterprise-token: ${{ secrets.GH_ENTERPRISE_TOKEN }}
```

### Enterprise Configuration

Load enterprise defaults automatically:

```yaml
- uses: SAP/project-piper-action@main
  with:
    step-name: getDefaults
    sap-piper-version: latest
    github-enterprise-token: ${{ secrets.GH_ENTERPRISE_TOKEN }}
```

This downloads and caches default configurations from the enterprise repository.

### GitHub Enterprise Support

Connect to GitHub Enterprise:

```yaml
env:
  GITHUB_SERVER_URL: https://github.company.com
  GITHUB_API_URL: https://github.company.com/api/v3
```

The action automatically detects GitHub Enterprise environment.

---

## Stage and Step Activation

### Overview

Conditionally activate stages and steps based on configuration and project state.

### Creating Activation Maps

```yaml
- uses: SAP/project-piper-action@main
  with:
    step-name: getDefaults
    create-check-if-step-active-maps: true
```

This creates:
- `.pipeline/stage_out.json` - Active stages map
- `.pipeline/step_out.json` - Active steps map

### Using Activation Maps

```yaml
- name: Check if step is active
  id: check
  run: |
    ACTIVE=$(jq -r '.mavenBuild' .pipeline/step_out.json)
    echo "active=$ACTIVE" >> $GITHUB_OUTPUT

- name: Run step if active
  if: steps.check.outputs.active == 'true'
  uses: SAP/project-piper-action@main
  with:
    step-name: mavenBuild
```

### Custom Stage Conditions

Provide custom stage condition configuration:

```yaml
custom-stage-conditions-path: 'config/stage-conditions.yml'
```

Or from another repository:

```yaml
custom-stage-conditions-path: 'myorg/config/stage-conditions.yml@v1.0.0'
```

---

## Best Practices

### Performance

1. **Pin Action Version**: Use `@v1` instead of `@main` in production
2. **Cache Dependencies**: Leverage workflow-level caching for node_modules, .m2, etc.
3. **Specific Binary Version**: Use specific `piper-version` for faster startup
4. **Minimize CPE**: Only export pipeline environment when necessary

### Security

1. **Use Secrets**: Never hardcode credentials in workflow files
2. **Vault Integration**: Prefer Vault over GitHub Secrets for sensitive data
3. **Minimal Permissions**: Use least privilege for tokens and credentials
4. **Regular Updates**: Keep action and Piper versions up to date

### Reliability

1. **Error Handling**: Monitor step exit codes and failures
2. **Verbose Logging**: Enable `--verbose` for troubleshooting
3. **Version Pinning**: Use specific versions in production pipelines
4. **Test Changes**: Test configuration changes in feature branches first

### Maintainability

1. **Custom Defaults**: Centralize common configuration
2. **Clear Comments**: Document non-obvious configuration choices
3. **Consistent Naming**: Use consistent job and step names
4. **Configuration Files**: Prefer `.pipeline/config.yml` over inline flags

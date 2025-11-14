# Build Stage

## Overview

The Build stage is responsible for building the application artifacts, running tests, and performing code quality checks. It supports multiple build tools and technologies including Maven, Gradle, npm, MTA, Docker, Python, and Golang.

## Stage Purpose

The Build stage performs the following key functions:

- Prepares artifact versioning for productive branches
- Executes build for configured build tool (Maven, Gradle, npm, MTA, Docker, etc.)
- Runs unit tests and code quality checks
- Builds and publishes Docker containers
- Packages and publishes Helm charts
- Performs SonarQube code analysis
- Uploads build artifacts and results to staging service and Cumulus
- Generates and uploads software bill of materials (SBOM)

## When the Stage Runs

The Build stage runs when:

- The Init stage activates it based on configuration
- `fromJSON(inputs.active-stages-map).Build == true`
- For all branches (productive and non-productive)
- Different behavior for productive vs. non-productive branches

## Steps in the Stage

### 1. Checkout repository
- **Action**: `actions/checkout@v4.3.0`
- **Purpose**: Checks out source code
- **Configuration**: Supports submodules, LFS, and custom fetch depth

### 2. Setup Go for development builds
- **Action**: `actions/setup-go@v6.0.0`
- **Condition**: Only for development Piper versions
- **Go Version**: 1.24

### 3. Read stage configuration
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `getConfig`
- **Output**: `stage-config.json`

### 4. Checkout global extension
- **Action**: `actions/checkout@v4`
- **Condition**: When extensibility is enabled
- **Purpose**: Checks out global pipeline extensions

### 5. Retrieve System Trust session token
- **Action**: `project-piper/system-trust-composite-action`
- **Purpose**: Obtains session token for System Trust integration
- **Permissions Required**: `id-token: write`

### 6. preBuild
- **Condition**: When extensibility is enabled
- **Purpose**: Executes custom pre-build extensions

### 7. artifactPrepareVersion (productive branch)
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `artifactPrepareVersion`
- **Condition**: On productive branch only
- **Flags**: `--username github-actions --password ${{ github.token }}`
- **Purpose**: Prepares and increments version for production artifacts
- **Note**: Username and password overridden by Vault secrets if configured

### 8. artifactPrepareVersion (non-productive branch)
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `artifactPrepareVersion`
- **Condition**: Not on productive branch
- **Flags**: `--versioningType cloud_noTag`
- **Purpose**: Prepares version without tagging for non-production builds

### 9. sapCumulusUpload (generate pipeline run key)
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapCumulusUpload`
- **Condition**: On productive branch
- **Flags**: `--filePattern "" --stepResultType ""`
- **Purpose**: Generates pipeline run key without uploading files

### 10. gcpPublishEvent - pipelineTaskRunFinished
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `gcpPublishEvent`
- **Condition**: On productive branch with configured Vault namespace
- **Purpose**: Publishes event indicating artifactPrepareVersion completion

### 11. sapCallStagingService (createGroup)
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapCallStagingService`
- **Flags**: `--action createGroup`
- **Purpose**: Creates staging group for build artifacts

### 12. sapCallStagingService (createRepositories)
- **Action**: `SAP/project-piper-action@v1.22`
- **Step**: `sapCallStagingService`
- **Flags**: `--action createRepositories`
- **Purpose**: Creates repositories in staging service

### Build Tool Steps

The following steps execute based on the configured `buildTool`:

#### 13. mtaBuild
- **Condition**: `buildTool: 'mta'`
- **Purpose**: Builds Multi-Target Application (MTA) archives
- **Supported Formats**: JAR, WAR, ZIP, MTAR

#### 14. gradleExecuteBuild
- **Condition**: `buildTool: 'gradle'`
- **Purpose**: Executes Gradle build
- **Includes**: Unit tests, code analysis

#### 15. npmExecuteScripts
- **Condition**: `buildTool: 'npm'` or `buildTool: 'CAP'`
- **Purpose**: Executes npm build scripts
- **Default Scripts**: `ci-build`, `ci-test`

#### 16. mavenBuild
- **Condition**: `buildTool: 'maven'` or `buildTool: 'CAP'`
- **Purpose**: Executes Maven build
- **Includes**: Unit tests, static analysis

#### 17. pythonBuild
- **Condition**: `buildTool: 'pip'`
- **Purpose**: Builds Python packages
- **Supports**: pip, poetry, pipenv

#### 18. golangBuild
- **Condition**: `buildTool: 'golang'`
- **Purpose**: Builds Go applications
- **Includes**: Unit tests, binary packaging

#### 19. kanikoExecute
- **Condition**: `buildTool: 'docker'`
- **Purpose**: Builds Docker images using Kaniko
- **Advantages**: Rootless, secure container builds

#### 20. cnbBuild
- **Condition**: Explicitly activated
- **Purpose**: Cloud Native Buildpacks build
- **Status**: Inactive by default

### Additional Build Steps

#### 21. hadolintExecute
- **Condition**: `Dockerfile` exists
- **Purpose**: Lints Dockerfile for best practices
- **Skipped When**: Optimized and scheduled run

#### 22. helmExecute
- **Condition**: Explicitly activated
- **Purpose**: Packages and publishes Helm charts
- **Status**: Inactive by default

#### 23. checkIfStepActive (post helmExecute)
- **Condition**: After helmExecute runs
- **Purpose**: Re-evaluates step activation based on CPE values from Helm

#### 24. sapOcmCreateComponent
- **Condition**: Explicitly activated and on productive branch
- **Purpose**: Creates Open Component Model component version
- **Status**: Inactive by default

#### 25. imagePushToRegistry
- **Condition**: Explicitly activated and on productive branch
- **Purpose**: Pushes Docker images to container registry
- **Status**: Inactive by default

### Result Upload Steps

#### 26. sapCumulusUpload (URL log upload)
- **Purpose**: Uploads URL access logs
- **Flags**: `--filePattern **/url-log.json --stepResultType access-log`

#### 27. sapGenerateEnvironmentInfo
- **Condition**: On productive branch
- **Flags**: `--buildEnvironment Hyperspace_GitHubActions_native_BuildStep`
- **Purpose**: Generates environment information

#### 28-31. Upload build artifacts
- **Actions**: `actions/upload-artifact@v3`
- **Artifacts**:
  - `build-settings.json`
  - `env*.json`

#### 32-38. sapCumulusUpload (multiple result types)
- **Upload Types**:
  - piper-config.yaml (config)
  - build-settings.json (settings, SLC-29-PNB policy)
  - env*.json (root, SLC-25 policy, SLC-29-PI policy)
  - hs-assessments.yaml (assessment)
  - bom-*.xml (SBOM)

#### 39. sapCallStagingService (close)
- **Condition**: Always runs
- **Flags**: `--action close`
- **Purpose**: Closes staging repository

### Testing and Quality Steps

#### 40. karmaExecuteTests
- **Condition**: `karma.conf.js` exists
- **Purpose**: Executes Karma/Jasmine tests (UI5, QUnit)

#### 41-43. sapCumulusUpload (test results)
- **Upload Types**:
  - JUnit results (TEST-*.xml)
  - JaCoCo coverage (jacoco.xml)
  - Cobertura coverage (cobertura-coverage.xml)
  - Requirement mapping

#### 44. sonarExecuteScan
- **Condition**: `sonar-project.properties` exists
- **Purpose**: Performs SonarQube analysis

#### 45-46. sapCumulusUpload (SonarQube results)
- **Upload Types**:
  - sonarscan-result.json, sonarscan.json (sonarqube)
  - Same files for FC-1 policy evidence

#### 47. postBuild
- **Condition**: Extensibility enabled
- **Purpose**: Executes custom post-build extensions

#### 48-49. Export outputs
- **Purpose**: Exports active stages/steps maps and pipeline environment
- **Condition**: Always runs

## Configuration Options

### Build Tool Configuration

Configure in `.pipeline/config.yml`:

```yaml
general:
  buildTool: 'maven'  # Options: maven, gradle, npm, mta, docker, pip, golang, CAP

  # For native build (default: true)
  nativeBuild: true

  # Docker registry for container builds
  dockerRegistryUrl: 'https://your-registry.example.com'
```

### Version Management

```yaml
steps:
  artifactPrepareVersion:
    versioningType: 'cloud'  # Options: cloud, cloud_noTag, library
    gitSshKeyCredentialsId: 'git-ssh-key'
    commitVersion: true
```

### Build Tool Specific Configuration

#### Maven Build

```yaml
steps:
  mavenBuild:
    pomPath: 'pom.xml'
    goals: 'clean install'
    defines: '-DskipTests=false'
    projectSettingsFile: 'settings.xml'
```

#### Gradle Build

```yaml
steps:
  gradleExecuteBuild:
    path: '.'
    task: 'build'
    buildFlags: '--no-daemon'
```

#### npm Build

```yaml
steps:
  npmExecuteScripts:
    runScripts:
      - 'ci-build'
      - 'ci-test'
    defaultNpmRegistry: 'https://registry.npmjs.org'
```

#### Docker Build (Kaniko)

```yaml
steps:
  kanikoExecute:
    dockerfilePath: 'Dockerfile'
    containerRegistryUrl: 'https://gcr.io'
    containerImageName: 'my-app'
    containerImageTag: '1.0.0'
```

### SonarQube Configuration

```yaml
steps:
  sonarExecuteScan:
    serverUrl: 'https://sonarqube.example.com'
    instance: 'SonarQube'
    projectKey: 'my-project'
    organization: 'my-org'
```

### Cumulus Upload Configuration

```yaml
steps:
  sapCumulusUpload:
    pipelineId: 'your-pipeline-id'
    # Cumulus server URL from Vault or configuration
```

## Example Usage

### Basic Build Stage

```yaml
jobs:
  build:
    needs: init
    uses: project-piper/piper-pipeline-github/.github/workflows/build.yml@main
    with:
      piper-version: 'latest'
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.init.outputs.pipeline-env }}
    secrets: inherit
```

### Build with Extensions

```yaml
jobs:
  build:
    needs: init
    uses: project-piper/piper-pipeline-github/.github/workflows/build.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.init.outputs.pipeline-env }}
      extensibility-enabled: true
      global-extensions-repository: 'my-org/pipeline-extensions'
      global-extensions-ref: 'main'
    secrets: inherit
```

### Build with Custom Runner and Fetch Depth

```yaml
jobs:
  build:
    needs: init
    uses: project-piper/piper-pipeline-github/.github/workflows/build.yml@main
    with:
      on-productive-branch: ${{ needs.init.outputs.on-productive-branch }}
      active-stages-map: ${{ needs.init.outputs.active-stages-map }}
      active-steps-map: ${{ needs.init.outputs.active-steps-map }}
      pipeline-env: ${{ needs.init.outputs.pipeline-env }}
      runs-on: '[ "self-hosted", "linux", "x64" ]'
      fetch-depth: '0'  # Full history for SonarQube
    secrets: inherit
```

## Build Tool Support Matrix

| Build Tool | Step Name | Primary Use Case |
|------------|-----------|------------------|
| Maven | `mavenBuild` | Java applications, Spring Boot |
| Gradle | `gradleExecuteBuild` | Java/Kotlin applications, Android |
| npm | `npmExecuteScripts` | Node.js applications, UI5 apps |
| MTA | `mtaBuild` | SAP Multi-Target Applications |
| Docker | `kanikoExecute` | Container images |
| Python | `pythonBuild` | Python packages |
| Golang | `golangBuild` | Go applications |
| CAP | `mavenBuild` or `npmExecuteScripts` | SAP Cloud Application Programming |

## Troubleshooting

### Build Failures

1. **Version conflict**: Check that `artifactPrepareVersion` completes successfully
2. **Missing dependencies**: Verify package manager configurations (pom.xml, package.json, etc.)
3. **Test failures**: Review test logs in the build output

### Staging Service Issues

1. Verify staging service credentials in Vault
2. Check network connectivity to staging service
3. Review staging group and repository creation logs

### Container Build Issues

1. Ensure Dockerfile is present and valid
2. Verify container registry credentials
3. Check Kaniko executor logs for detailed error messages

### SonarQube Scan Failures

1. Verify `sonar-project.properties` configuration
2. Check SonarQube server connectivity
3. Ensure quality gate is properly configured

## Best Practices

1. **Test Coverage**: Include unit tests in build process
2. **Code Quality**: Enable SonarQube for code analysis
3. **Version Control**: Use semantic versioning for artifacts
4. **Artifact Management**: Always upload to staging service on productive branch
5. **Security Scanning**: Enable Dockerfile linting with Hadolint
6. **SBOM Generation**: Ensure bill of materials is generated and uploaded
7. **Build Caching**: Use appropriate fetch depth and caching strategies
8. **Parallel Execution**: Configure test parallelization for faster builds

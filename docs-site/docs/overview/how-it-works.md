# How Piper CI/CD Works

This document explains the execution flow and internal workings of the Project Piper CI/CD tool.

## Overview

Piper operates as a **configuration-driven, metadata-based CI/CD execution engine** that can run natively on multiple CI/CD platforms while maintaining consistent behavior and configuration.

## Execution Flow

### High-Level Flow

```
1. Pipeline Triggered
   ↓
2. Init Stage: Load Configuration
   ↓
3. Determine Active Stages/Steps
   ↓
4. Execute Stages Sequentially
   ↓
5. Each Stage Executes Steps
   ↓
6. State Persisted in CPE
   ↓
7. Post Stage: Reporting & Cleanup
   ↓
8. Pipeline Complete
```

### Detailed Execution Flow

```
┌─────────────────────────────────────┐
│  1. Pipeline Initialization         │
│  • Checkout repository              │
│  • Load .pipeline/config.yml        │
│  • Merge with defaults              │
│  • Initialize CPE                   │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│  2. Configuration Resolution         │
│  • Read metadata for each step      │
│  • Merge configurations             │
│  • Resolve secrets from Vault       │
│  • Validate parameters              │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│  3. Stage Activation                 │
│  • Check stage conditions           │
│  • Evaluate step activation rules   │
│  • Build execution plan             │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│  4. Step Execution                   │
│  • Load step configuration          │
│  • Prepare execution environment    │
│  • Execute step logic               │
│  • Capture output                   │
│  • Update CPE                       │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│  5. State Persistence                │
│  • Write CPE to filesystem          │
│  • Upload artifacts (GitHub/Azure)  │
│  • Stash files (Jenkins)            │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│  6. Reporting & Cleanup              │
│  • Publish test results             │
│  • Generate reports                 │
│  • Send notifications               │
│  • Clean up temporary files         │
└─────────────────────────────────────┘
```

---

## Platform-Specific Execution

### Jenkins Execution

```groovy
// Jenkinsfile
@Library('piper-lib-os') _
piperPipeline script: this
```

**Flow**:
1. Jenkins loads shared library from Git
2. `piperPipeline.groovy` orchestrates stages
3. Each stage calls step wrappers in `vars/`
4. Step wrappers execute Piper binary
5. Results captured and published using Jenkins APIs

**Example**: `mavenBuild` execution
```groovy
// vars/mavenBuild.groovy
void call(Map parameters = [:]) {
    handlePipelineStepErrors(...) {
        // 1. Merge configuration
        def config = ConfigurationHelper.newInstance()
            .loadStepDefaults()
            .mixinStepConfig(script, parameters)
            .build()

        // 2. Execute Piper binary
        piperExecuteBin(
            parameters,
            stepName,
            metadataFile,
            []
        )

        // 3. Process results
        testsPublishResults(...)
    }
}
```

### GitHub Actions Execution

```yaml
# workflow.yml
- uses: SAP/project-piper-action@main
  with:
    step-name: mavenBuild
    flags: '--publish --createBOM'
```

**Flow**:
1. Action downloads/caches Piper binary
2. Loads CPE from previous step artifacts
3. Executes: `piper mavenBuild --publish --createBOM`
4. Saves CPE to artifacts for next steps
5. Action completes with exit code from Piper

**TypeScript Implementation**:
```typescript
// src/main.ts
async function run() {
    // 1. Setup
    const actionConfig = parseActionConfig()
    await preparePiperBinary(actionConfig)

    // 2. Load state
    await loadPipelineEnv()

    // 3. Execute
    const exitCode = await executePiper(
        actionConfig.stepName,
        actionConfig.flags
    )

    // 4. Persist state
    await exportPipelineEnv()

    // 5. Complete
    if (exitCode !== 0) {
        core.setFailed(`Step failed with exit code ${exitCode}`)
    }
}
```

### Azure DevOps Execution

```yaml
# azure-pipelines.yml
resources:
  repositories:
    - repository: piper
      type: github
      name: SAP/piper-pipeline-azure

stages:
  - template: sap-piper-pipeline.yml@piper
```

**Flow**:
1. Azure loads pipeline template from repository
2. Template defines stage structure
3. Each stage uses piper task extension
4. Task downloads and executes Piper binary
5. Results published using Azure APIs

---

## Step Execution Internals

### 1. Command Routing

```go
// cmd/piper.go
func Execute() {
    rootCmd := &cobra.Command{
        Use:   "piper",
        Short: "Executes CI/CD steps",
    }

    // Register all step commands
    rootCmd.AddCommand(MavenBuildCommand())
    rootCmd.AddCommand(KanikoExecuteCommand())
    // ... 200+ more steps

    rootCmd.Execute()
}
```

### 2. Configuration Loading

```go
// Step command setup
func MavenBuildCommand() *cobra.Command {
    metadata := getStepMetadata("mavenBuild")

    cmd := &cobra.Command{
        Use:   "mavenBuild",
        Short: metadata.Metadata.Description,
        Run: func(cmd *cobra.Command, args []string) {
            // 1. Load configuration
            config := prepareConfig(cmd, metadata)

            // 2. Execute step
            mavenBuild(config, telemetry)

            // 3. Handle errors
            if err != nil {
                log.SetErrorCategory(log.ErrorBuild)
                log.Entry().WithError(err).Fatal("step execution failed")
            }
        },
    }

    // Add flags from metadata
    addMavenBuildFlags(cmd, metadata)

    return cmd
}
```

### 3. Configuration Merging

Configuration sources are merged in this order (highest to lowest priority):

```go
// Simplified configuration resolution
func prepareConfig(cmd *cobra.Command, metadata *config.StepData) Config {
    config := Config{}

    // 7. Load built-in defaults (lowest priority)
    config.merge(loadDefaultsFromMetadata(metadata))

    // 6. Load stage defaults
    config.merge(loadStageDefaults())

    // 5. Load custom defaults
    if customDefaults := getCustomDefaults(); customDefaults != "" {
        config.merge(loadYAML(customDefaults))
    }

    // 4. Load project config (.pipeline/config.yml)
    config.merge(loadProjectConfig())

    // 3. Load parameters JSON (if provided)
    if paramsJSON := cmd.Flag("parameters").Value; paramsJSON != "" {
        config.merge(loadJSON(paramsJSON))
    }

    // 2. Load environment variables (PIPER_*)
    config.merge(loadEnvVars(metadata))

    // 1. Load command-line flags (highest priority)
    config.merge(loadFlags(cmd, metadata))

    // Validate final configuration
    config.validate(metadata)

    return config
}
```

### 4. Secret Resolution

```go
// Resolve secrets from Vault
func resolveSecrets(config *Config, metadata *config.StepData) {
    for _, secret := range metadata.Spec.Inputs.Secrets {
        if config.hasVaultPath(secret.Name) {
            vaultClient := getVaultClient()

            secretValue := vaultClient.GetSecret(
                config.getVaultPath(secret.Name),
                config.getVaultKey(secret.Name),
            )

            config.set(secret.Name, secretValue)
        }
    }
}
```

### 5. Step Execution

```go
// Example: mavenBuild implementation
func mavenBuild(config mavenBuildOptions, telemetry *telemetry.CustomData) {
    // 1. Detect build descriptor
    utils := maven.NewUtilsBundle()
    pom := config.PomPath

    // 2. Define Maven goals
    goals := config.Goals
    if len(goals) == 0 {
        goals = []string{"install"}
    }

    // 3. Build Maven command
    mavenOptions := maven.ExecuteOptions{
        Goals:                   goals,
        Defines:                 config.Defines,
        Profiles:                config.Profiles,
        ProjectSettingsFile:     config.ProjectSettingsFile,
        GlobalSettingsFile:      config.GlobalSettingsFile,
        M2Path:                  config.M2Path,
        ReturnStdout:            false,
    }

    // 4. Execute Maven
    _, err := maven.Execute(&mavenOptions, utils)
    if err != nil {
        return err
    }

    // 5. Publish artifacts (if configured)
    if config.Publish {
        publishArtifacts(config)
    }

    // 6. Create BOM (if configured)
    if config.CreateBOM {
        createBillOfMaterials(config)
    }

    // 7. Write to CPE
    writeToCPE("maven/buildArtifacts", artifacts)
    writeToCPE("maven/goals", goals)
}
```

---

## Common Pipeline Environment (CPE)

### Purpose
The CPE is a **filesystem-based state store** that enables data sharing between pipeline steps.

### Structure

```
.pipeline/
  commonPipelineEnvironment/
    git/
      commitId                    # String: abc123def
      branch                      # String: main
      headCommitId               # String: def456ghi

    artifact/
      version                     # String: 1.0.0-20240115120000

    custom/
      buildArtifacts.json         # JSON: {"files": ["target/app.jar"]}
      mavenExecute               # String: install
      buildSettingsInfo          # String: {"maven": "3.8.1"}

    abap/
      repositories.json           # JSON: [{"name": "repo1", ...}]
```

### Reading from CPE

```go
// pkg/piperenv/CPEClient
func (c *CPEClient) GetString(path string) string {
    filePath := filepath.Join(c.baseDir, path)
    content, _ := ioutil.ReadFile(filePath)
    return string(content)
}

func (c *CPEClient) GetJSON(path string, target interface{}) {
    filePath := filepath.Join(c.baseDir, path)
    content, _ := ioutil.ReadFile(filePath)
    json.Unmarshal(content, target)
}
```

### Writing to CPE

```go
func (c *CPEClient) WriteString(path string, value string) {
    filePath := filepath.Join(c.baseDir, path)
    os.MkdirAll(filepath.Dir(filePath), 0755)
    ioutil.WriteFile(filePath, []byte(value), 0644)
}

func (c *CPEClient) WriteJSON(path string, value interface{}) {
    content, _ := json.Marshal(value)
    c.WriteString(path, string(content))
}
```

### CPE Persistence

**Jenkins**: Uses stashing/unstashing
```groovy
// Stash CPE files
dir('.pipeline') {
    stash name: 'pipelineEnv', includes: 'commonPipelineEnvironment/**'
}

// Unstash in next stage
unstash 'pipelineEnv'
```

**GitHub Actions**: Uses artifacts
```yaml
- name: Upload CPE
  uses: actions/upload-artifact@v3
  with:
    name: piper-cpe
    path: .pipeline/commonPipelineEnvironment/

- name: Download CPE
  uses: actions/download-artifact@v3
  with:
    name: piper-cpe
    path: .pipeline/commonPipelineEnvironment/
```

**Azure DevOps**: Uses pipeline artifacts
```yaml
- task: PublishPipelineArtifact@1
  inputs:
    artifactName: 'piper-cpe'
    targetPath: '.pipeline/commonPipelineEnvironment'
```

---

## Container Execution

### When to Use Containers

Steps use containers when:
1. Metadata specifies container image
2. Consistent environment needed
3. Tool not available on host
4. Isolation required

### Container Execution Flow

```go
func executeInContainer(config ContainerConfig, command []string) {
    // 1. Pull image
    dockerClient.PullImage(config.Image)

    // 2. Create container
    container := dockerClient.CreateContainer(ContainerConfig{
        Image:      config.Image,
        Cmd:        command,
        WorkingDir: "/workspace",
        Volumes: map[string]string{
            workspaceDir: "/workspace",
            cpeDir:       "/.pipeline/commonPipelineEnvironment",
        },
        Env: config.Env,
    })

    // 3. Start container
    dockerClient.StartContainer(container.ID)

    // 4. Stream output
    dockerClient.ContainerLogs(container.ID, os.Stdout, os.Stderr)

    // 5. Wait for completion
    exitCode := dockerClient.WaitContainer(container.ID)

    // 6. Cleanup
    dockerClient.RemoveContainer(container.ID)

    return exitCode
}
```

### Sidecar Containers

For testing, steps can use sidecar containers:

```yaml
# Metadata example
spec:
  sidecars:
    - name: selenium
      image: selenium/standalone-chrome:latest
      options:
        - name: --shm-size
          value: "2g"
```

---

## Stage Orchestration

### General Purpose Pipeline

```groovy
// vars/piperPipeline.groovy
void call(Map parameters = [:]) {
    // 1. Init
    piperPipelineStageInit script: this

    // 2. PR Voting (Jenkins only)
    if (env.BRANCH_NAME && isPullRequest()) {
        piperPipelineStagePRVoting script: this
        return // Exit after PR voting
    }

    // 3. Build
    piperPipelineStageBuild script: this

    // 4. Additional Unit Tests
    piperPipelineStageAdditionalUnitTests script: this

    // 5. Integration
    piperPipelineStageIntegration script: this

    // 6. Acceptance
    piperPipelineStageAcceptance script: this

    // 7. Security
    piperPipelineStageSecurity script: this

    // 8. Performance
    piperPipelineStagePerformance script: this

    // 9. Compliance
    piperPipelineStageCompliance script: this

    // 10. Confirm
    piperPipelineStageConfirm script: this

    // 11. Promote
    piperPipelineStagePromote script: this

    // 12. Release
    piperPipelineStageRelease script: this

    // Always execute Post stage
    piperPipelineStagePost script: this
}
```

### Stage Wrapper

Each stage is wrapped with:
- Configuration loading
- Conditional execution
- Extension support
- Error handling

```groovy
// vars/piperStageWrapper.groovy
void call(Map parameters = [:], body) {
    def stageName = parameters.stageName
    def config = loadStageConfig(stageName)

    // Check if stage is active
    if (!isStageActive(stageName, config)) {
        echo "Stage ${stageName} is not active, skipping"
        return
    }

    // Run pre-stage extension
    if (hasExtension("${stageName}Pre")) {
        runExtension("${stageName}Pre", config)
    }

    try {
        // Execute stage body
        body()
    } catch (Exception e) {
        // Handle error
        handleStageError(stageName, e)
        throw e
    } finally {
        // Run post-stage extension
        if (hasExtension("${stageName}Post")) {
            runExtension("${stageName}Post", config)
        }
    }
}
```

---

## Conditional Execution

### Stage Activation

Stages execute based on:

```yaml
# .pipeline/config.yml
stages:
  Build:
    active: true

  Acceptance:
    # Active if acceptance tests configured

  Performance:
    # Active if performance tests configured

  Security:
    # Active if security scans configured
```

### Step Activation

Steps execute based on:

```yaml
# Built into metadata
spec:
  inputs:
    params:
      - name: myParam
        conditions:
          - conditionRef: strings-equal
            params:
              - name: buildTool
                value: maven
```

---

## Error Handling

### Error Categories

```go
const (
    ErrorCompliance  = "compliance"
    ErrorBuild       = "build"
    ErrorTest        = "test"
    ErrorSecurity    = "security"
    ErrorConfiguration = "configuration"
)
```

### Error Flow

```
Step Error Occurs
    ↓
Error category set
    ↓
Error logged with context
    ↓
CPE updated with error state
    ↓
Exit with non-zero code
    ↓
Platform handles failure
    ↓
Post stage still executes
    ↓
Notifications sent
```

---

## Telemetry and Reporting

### Telemetry Collection

```go
telemetryData := &telemetry.CustomData{
    StepName: "mavenBuild",
    Duration: duration,
    ErrorCode: errorCode,
    ErrorCategory: errorCategory,
}

if collectTelemetryData {
    telemetry.Send(telemetryData)
}
```

### Reporting

1. **Test Results**: JUnit XML, Jacoco, Cobertura
2. **Security Scans**: SARIF, custom JSON
3. **Quality Metrics**: SonarQube, linters
4. **Build Artifacts**: Artifact lists, BOMs
5. **Pipeline Metrics**: Duration, success rate

---

## Summary

Piper works by:

1. **Unifying** CI/CD across platforms with a shared core
2. **Configuring** through hierarchical YAML configuration
3. **Executing** containerized steps with consistent behavior
4. **Persisting** state through CPE filesystem store
5. **Orchestrating** stages with conditional activation
6. **Reporting** comprehensive results and metrics
7. **Extending** through multiple customization points

This architecture enables teams to write CI/CD configuration once and run it consistently across Jenkins, GitHub Actions, and Azure DevOps.

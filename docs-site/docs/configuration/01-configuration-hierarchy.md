# Configuration Hierarchy

Project Piper uses a sophisticated 7-level hierarchical configuration system that merges configuration from multiple sources with clearly defined precedence rules.

## Table of Contents

- [Configuration Hierarchy](#configuration-hierarchy)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [The 7 Configuration Levels](#the-7-configuration-levels)
    - [Level 1: Default Configuration (Lowest Precedence)](#level-1-default-configuration-lowest-precedence)
    - [Level 2: Custom Default Configuration](#level-2-custom-default-configuration)
    - [Level 3: General Configuration](#level-3-general-configuration)
    - [Level 4: Step Configuration](#level-4-step-configuration)
    - [Level 5: Stage Configuration](#level-5-stage-configuration)
    - [Level 6: Environment Variables](#level-6-environment-variables)
    - [Level 7: Direct Step Parameters (Highest Precedence)](#level-7-direct-step-parameters-highest-precedence)
  - [Configuration Merging Process](#configuration-merging-process)
    - [Merge Rules](#merge-rules)
    - [Deep Merging Example](#deep-merging-example)
  - [Practical Examples](#practical-examples)
    - [Example 1: Simple Override](#example-1-simple-override)
    - [Example 2: Multi-Level Configuration](#example-2-multi-level-configuration)
    - [Example 3: Build Tool-Specific Configuration](#example-3-build-tool-specific-configuration)
  - [Special Configuration Features](#special-configuration-features)
    - [Parameter Aliases](#parameter-aliases)
    - [Conditional Defaults](#conditional-defaults)
    - [Configuration Filters](#configuration-filters)
  - [Configuration Loading Order](#configuration-loading-order)
  - [Advanced Topics](#advanced-topics)
    - [Custom Defaults Chain](#custom-defaults-chain)
    - [Environment-Based Configuration](#environment-based-configuration)
    - [Vault Parameter Injection](#vault-parameter-injection)
  - [Debugging Configuration](#debugging-configuration)
    - [Enable Verbose Logging](#enable-verbose-logging)
    - [View Effective Configuration](#view-effective-configuration)
    - [Common Issues](#common-issues)
  - [Best Practices](#best-practices)

## Overview

The configuration hierarchy determines which value takes precedence when the same parameter is defined in multiple places. Understanding this hierarchy is crucial for effective pipeline configuration.

**Precedence Order (Highest to Lowest):**

```
7. Direct Step Parameters (highest)
   ↓
6. Environment Variables (PIPER_*)
   ↓
5. Stage Configuration
   ↓
4. Step Configuration
   ↓
3. General Configuration
   ↓
2. Custom Default Configuration
   ↓
1. Default Configuration (lowest)
```

## The 7 Configuration Levels

### Level 1: Default Configuration (Lowest Precedence)

**Source**: `resources/default_pipeline_environment.yml` in the jenkins-library

**Purpose**: Provides sensible defaults for all steps

**Example**:
```yaml
# From default_pipeline_environment.yml
steps:
  mavenExecute:
    dockerImage: 'maven:3.5-jdk-7'
    logSuccessfulMavenTransfers: false

  npmExecute:
    dockerImage: 'node:lts-bookworm'

  cloudFoundryDeploy:
    deployType: 'standard'
    keepOldInstance: false
```

**Characteristics**:
- Maintained by Project Piper team
- Updated with library releases
- Provides baseline configuration
- Cannot be modified by users

### Level 2: Custom Default Configuration

**Source**: Files referenced in `customDefaults` parameter

**Purpose**: Organization or team-wide defaults

**Configuration**:
```yaml
# In .pipeline/config.yml
customDefaults:
  - 'https://github.company.com/raw/org/ci-cd/java-defaults.yml'
  - 'https://github.company.com/raw/org/ci-cd/security-defaults.yml'
```

**Example custom defaults file**:
```yaml
# java-defaults.yml
general:
  dockerPullImage: true

steps:
  mavenExecute:
    dockerImage: 'maven:3.8-openjdk-17'
    defines: '-Dmaven.test.failure.ignore=false'

  checksPublishResults:
    pmd:
      active: true
      pattern: '**/target/pmd.xml'
    checkstyle:
      active: true
      pattern: '**/target/checkstyle-result.xml'
```

**Characteristics**:
- Loaded from URLs or file paths
- Multiple files processed in order (later wins)
- Merged before project config
- Requires anonymous read access (for URLs)

### Level 3: General Configuration

**Source**: `general` section in `.pipeline/config.yml`

**Purpose**: Project-wide settings available to all steps

**Example**:
```yaml
general:
  buildTool: 'maven'
  productiveBranch: 'main'
  gitSshKeyCredentialsId: 'github-ssh-key'
  verbose: true

  # Available to all steps
  dockerImage: 'maven:3.8-openjdk-11'

  # Cloud Foundry settings
  cloudFoundry:
    org: 'my-organization'
    credentialsId: 'CF_CREDENTIALS'
```

**Available to**:
- All pipeline steps
- All pipeline stages
- Can be overridden by step/stage config

### Level 4: Step Configuration

**Source**: `steps` section in `.pipeline/config.yml`

**Purpose**: Configure specific step behavior

**Example**:
```yaml
steps:
  mavenExecute:
    goals: 'clean verify'
    defines: '-DskipTests=false'
    pomPath: 'pom.xml'

  cloudFoundryDeploy:
    deployTool: 'cf_native'
    cloudFoundry:
      space: 'development'
      manifest: 'manifest.yml'

  sonarExecuteScan:
    serverUrl: 'https://sonarqube.company.com'
    projectKey: 'my-project'
```

**Characteristics**:
- Applies to named step everywhere it's used
- Overrides general configuration
- Can be further overridden by stage config

### Level 5: Stage Configuration

**Source**: `stages` section in `.pipeline/config.yml`

**Purpose**: Stage-specific step parameters

**Example**:
```yaml
stages:
  Build:
    mavenExecuteStaticCodeChecks: true
    mavenExecuteIntegration: true

  Acceptance:
    cloudFoundryDeploy:
      cloudFoundry:
        space: 'acceptance'
      smokeTest: true
      smokeTestScript: 'acceptance-test.sh'

  Release:
    cloudFoundryDeploy:
      cloudFoundry:
        space: 'production'
      smokeTest: true
      smokeTestStatusCode: 200
```

**Characteristics**:
- Highest priority for configuration files
- Overrides both step and general config
- Stage-specific execution context

### Level 6: Environment Variables

**Source**: Environment variables with `PIPER_` prefix

**Purpose**: Runtime parameter injection

**Example**:
```bash
# In shell or CI environment
export PIPER_verbose=true
export PIPER_dockerImage='maven:3.8-openjdk-17'
export PIPER_cloudFoundry_space='staging'
```

**In Jenkins**:
```groovy
environment {
    PIPER_verbose = 'true'
    PIPER_buildTool = 'maven'
}
```

**Characteristics**:
- Dynamic runtime configuration
- Useful for CI/CD platform integration
- Overrides file-based configuration
- Use underscore for nested parameters

### Level 7: Direct Step Parameters (Highest Precedence)

**Source**: Parameters passed directly to step calls

**Purpose**: Override all other configuration

**Example in Groovy (Jenkins)**:
```groovy
// Direct parameters override everything
mavenExecute(
    script: this,
    goals: 'clean deploy',
    defines: '-Dproduction=true',
    verbose: true
)

cloudFoundryDeploy(
    script: this,
    cloudFoundry: [
        space: 'production',
        apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'
    ]
)
```

**Example in CLI**:
```bash
# Command-line flags have highest precedence
piper mavenExecute --goals="clean verify" --verbose=true
```

**Characteristics**:
- Highest precedence
- Explicit override mechanism
- Useful for testing or special cases
- Not recommended for regular use

## Configuration Merging Process

### Merge Rules

1. **Primitive values**: Later value completely replaces earlier value
2. **Arrays/Lists**: Later value completely replaces earlier value
3. **Maps/Objects**: Deep merge - nested keys are merged recursively

### Deep Merging Example

**Default configuration**:
```yaml
steps:
  cloudFoundryDeploy:
    cloudFoundry:
      org: 'default-org'
      space: 'dev'
      apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'
    deployType: 'standard'
    keepOldInstance: false
```

**Project configuration**:
```yaml
steps:
  cloudFoundryDeploy:
    cloudFoundry:
      org: 'my-org'
      space: 'production'
    keepOldInstance: true
```

**Merged result**:
```yaml
steps:
  cloudFoundryDeploy:
    cloudFoundry:
      org: 'my-org'              # Overridden
      space: 'production'         # Overridden
      apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'  # Inherited
    deployType: 'standard'        # Inherited
    keepOldInstance: true         # Overridden
```

## Practical Examples

### Example 1: Simple Override

**Scenario**: Override default Docker image for Maven builds

**Default** (level 1):
```yaml
steps:
  mavenExecute:
    dockerImage: 'maven:3.5-jdk-7'
```

**Project config** (level 4):
```yaml
steps:
  mavenExecute:
    dockerImage: 'maven:3.8-openjdk-17'
```

**Result**: Uses `maven:3.8-openjdk-17`

### Example 2: Multi-Level Configuration

**Scenario**: Configure Cloud Foundry deployment across multiple levels

**Custom defaults** (level 2):
```yaml
general:
  cloudFoundry:
    org: 'company-org'
    apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'

steps:
  cloudFoundryDeploy:
    deployType: 'blue-green'
    smokeTest: true
```

**Project config** (level 3, 4, 5):
```yaml
general:
  cloudFoundry:
    credentialsId: 'CF_PROD_CREDS'

steps:
  cloudFoundryDeploy:
    manifest: 'manifest-prod.yml'

stages:
  Acceptance:
    cloudFoundryDeploy:
      cloudFoundry:
        space: 'acceptance'
      smokeTestScript: 'smoke-test-acceptance.sh'

  Release:
    cloudFoundryDeploy:
      cloudFoundry:
        space: 'production'
      smokeTestScript: 'smoke-test-production.sh'
```

**Effective configuration for Release stage**:
```yaml
cloudFoundryDeploy:
  cloudFoundry:
    org: 'company-org'                           # From custom defaults
    apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'  # From custom defaults
    credentialsId: 'CF_PROD_CREDS'              # From general
    space: 'production'                          # From stage config
  deployType: 'blue-green'                       # From custom defaults
  smokeTest: true                                # From custom defaults
  manifest: 'manifest-prod.yml'                  # From step config
  smokeTestScript: 'smoke-test-production.sh'   # From stage config
```

### Example 3: Build Tool-Specific Configuration

**Scenario**: Different configurations for different build tools

**Default** (level 1):
```yaml
steps:
  artifactSetVersion:
    maven:
      filePath: 'pom.xml'
      versioningTemplate: '${version}-${timestamp}${commitId?"_"+commitId:""}'
    npm:
      filePath: 'package.json'
      versioningTemplate: '${version}-${timestamp}${commitId?"+"+commitId:""}'
```

**Project config** (level 3, 4):
```yaml
general:
  buildTool: 'maven'

steps:
  artifactSetVersion:
    commitVersion: true
    maven:
      versioningTemplate: '${version}-${timestamp}'
```

**Result**: Uses Maven-specific settings with custom template

## Special Configuration Features

### Parameter Aliases

Some parameters have aliases for backward compatibility or convenience:

```yaml
steps:
  someStep:
    # These are equivalent
    dockerImage: 'node:14'
    # vs
    containerImage: 'node:14'
```

**Warning message** (if using deprecated alias):
```
[WARNING] The parameter 'dockerImage' is DEPRECATED, use 'containerImage' instead. (jenkins-library/someStep)
```

### Conditional Defaults

Step parameters can have conditional defaults based on other parameters:

```yaml
# Simplified example from step metadata
parameters:
  - name: deployTool
    default: 'cf_native'

  - name: manifest
    default: 'manifest.yml'
    conditions:
      - params:
          - name: deployTool
            value: 'cf_native'
```

### Configuration Filters

Steps define which parameters they accept from each configuration level:

```yaml
# In step metadata
filters:
  general:
    - buildTool
    - verbose
    - dockerImage
  steps:
    - goals
    - defines
    - pomPath
  stages:
    - goals
    - defines
```

## Configuration Loading Order

Detailed sequence during pipeline execution:

1. **Initialization Phase**
   - Load default configuration from library
   - Parse `.pipeline/config.yml`
   - Download and parse custom defaults (in order)
   - Merge custom defaults with defaults

2. **Step Preparation Phase**
   - Extract general configuration
   - Extract step configuration
   - Extract stage configuration (if in stage context)
   - Apply parameter aliases
   - Merge in precedence order

3. **Runtime Phase**
   - Read environment variables (PIPER_*)
   - Merge environment variables
   - Apply direct parameters (if any)
   - Resolve Vault references
   - Apply conditional defaults

4. **Validation Phase**
   - Check mandatory parameters
   - Validate parameter types
   - Verify parameter values

## Advanced Topics

### Custom Defaults Chain

Multiple custom defaults are processed in order:

```yaml
customDefaults:
  - 'https://company.com/org-defaults.yml'      # Loaded first
  - 'https://company.com/team-defaults.yml'     # Overrides org defaults
  - 'https://company.com/project-type.yml'      # Overrides team defaults

general:
  buildTool: 'maven'  # Overrides all custom defaults
```

### Environment-Based Configuration

Combine environment variables with file config:

```yaml
# .pipeline/config.yml
general:
  cloudFoundry:
    org: 'my-org'
    apiEndpoint: 'https://api.cf.eu10.hana.ondemand.com'

stages:
  Release:
    cloudFoundryDeploy:
      cloudFoundry:
        space: 'production'
```

```bash
# In CI/CD, set space dynamically
export PIPER_cloudFoundry_space="${ENVIRONMENT_NAME}"
```

### Vault Parameter Injection

Vault parameters are injected after file configuration merging:

```yaml
general:
  vaultServerUrl: 'https://vault.company.com'
  vaultPath: 'piper/my-project'

steps:
  cloudFoundryDeploy:
    cloudFoundryPasswordVaultSecretName: 'cf-password'
    # Password retrieved from: <vaultPath>/cf-password
```

## Debugging Configuration

### Enable Verbose Logging

```yaml
general:
  verbose: true

steps:
  mavenExecute:
    verbose: true
```

Output shows:
```
[mavenExecute] Configuration: {
  goals: 'clean verify',
  dockerImage: 'maven:3.8-openjdk-17',
  ...
}
```

### View Effective Configuration

**In Jenkins pipeline**:
```groovy
def config = ConfigurationHelper.newInstance(this)
    .loadStepDefaults()
    .mixinGeneralConfig(commonPipelineEnvironment)
    .mixinStepConfig(commonPipelineEnvironment)
    .use()

echo "Effective config: ${config}"
```

**Using CLI**:
```bash
piper getConfig --stepName=mavenExecute
```

### Common Issues

**Issue**: Configuration not taking effect

**Solution**: Check precedence - higher levels override lower levels

---

**Issue**: Unexpected parameter value

**Solution**: Enable verbose mode to see merged configuration

---

**Issue**: Custom defaults not loading

**Solution**: Verify URL is accessible and returns valid YAML

---

**Issue**: Environment variable not working

**Solution**: Ensure `PIPER_` prefix and correct naming (use underscore for nesting)

## Best Practices

1. **Document Overrides**: Add comments explaining why defaults are overridden
   ```yaml
   steps:
     mavenExecute:
       # Using JDK 17 for Spring Boot 3 compatibility
       dockerImage: 'maven:3.8-openjdk-17'
   ```

2. **Use Custom Defaults Wisely**: Put organization-wide policies in custom defaults
   ```yaml
   # org-defaults.yml
   general:
     collectTelemetryData: false  # Company policy
   ```

3. **Minimize Direct Parameters**: Prefer configuration files over direct parameters
   ```groovy
   // Avoid
   mavenExecute(script: this, goals: 'clean verify', dockerImage: '...')

   // Prefer
   // Put in config.yml and just call:
   mavenExecute script: this
   ```

4. **Stage-Specific Overrides**: Use stage configuration for environment differences
   ```yaml
   stages:
     Acceptance:
       cloudFoundryDeploy:
         cloudFoundry:
           space: 'acceptance'
     Release:
       cloudFoundryDeploy:
         cloudFoundry:
           space: 'production'
   ```

5. **Test Configuration Changes**: Use feature branches to test configuration changes
   ```bash
   git checkout -b test-config-changes
   # Modify .pipeline/config.yml
   # Trigger pipeline on feature branch
   # Verify results before merging to main
   ```

6. **Keep Secrets Out**: Never put secrets in configuration files
   ```yaml
   # WRONG - Don't do this
   general:
     cloudFoundry:
       password: 'my-secret-password'

   # RIGHT - Use Vault or credentials
   general:
     vaultPath: 'piper/my-project'
   steps:
     cloudFoundryDeploy:
       cloudFoundryPasswordVaultSecretName: 'cf-password'
   ```

---

**Next**: [Default Settings](02-default-settings.md) - Explore built-in defaults in detail

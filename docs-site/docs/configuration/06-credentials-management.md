# Credentials Management

Secure handling of credentials and secrets is critical for CI/CD pipelines. This guide covers Project Piper's credential management features including HashiCorp Vault integration, Jenkins credentials, and platform-specific secret stores.

## Table of Contents

- [Overview](#overview)
  - [HashiCorp Vault Integration](#hashicorp-vault-integration)
    - [Vault Setup](#vault-setup)
    - [Authentication Methods](#authentication-methods)
    - [Basic Configuration](#basic-configuration)
    - [Vault Path Hierarchy](#vault-path-hierarchy)
    - [Automatic Parameter Resolution](#automatic-parameter-resolution)
  - [Vault Configuration Examples](#vault-configuration-examples)
    - [Example 1: Basic Vault Setup](#example-1-basic-vault-setup)
    - [Example 2: Cloud Foundry Credentials](#example-2-cloud-foundry-credentials)
    - [Example 3: Multiple Vault Paths](#example-3-multiple-vault-paths)
    - [Example 4: Namespace Configuration](#example-4-namespace-configuration)
  - [General Purpose Credentials](#general-purpose-credentials)
    - [Single Credential Path](#single-credential-path)
    - [Multiple Credential Paths](#multiple-credential-paths)
    - [Custom Environment Variable Prefix](#custom-environment-variable-prefix)
  - [Test Credentials (Deprecated)](#test-credentials-deprecated)
  - [Jenkins Credentials](#jenkins-credentials)
    - [Credential Types](#credential-types)
    - [Referencing Jenkins Credentials](#referencing-jenkins-credentials)
    - [Creating Jenkins Credentials](#creating-jenkins-credentials)
  - [Platform-Specific Credentials](#platform-specific-credentials)
    - [Azure DevOps](#azure-devops)
    - [GitHub Actions](#github-actions)
  - [Credential Resolution Process](#credential-resolution-process)
  - [Advanced Vault Features](#advanced-vault-features)
    - [Disabling Vault Overwrite](#disabling-vault-overwrite)
    - [Skipping Vault Lookup](#skipping-vault-lookup)
    - [Vault Secret Files](#vault-secret-files)
  - [Security Best Practices](#security-best-practices)
  - [Troubleshooting](#troubleshooting)
    - [Vault Authentication Failures](#vault-authentication-failures)
    - [Secret Not Found](#secret-not-found)
    - [Permission Denied](#permission-denied)
    - [Token Expiration](#token-expiration)
  - [Migration Guide](#migration-guide)
    - [From Jenkins Credentials to Vault](#from-jenkins-credentials-to-vault)
    - [From Hardcoded to Vault](#from-hardcoded-to-vault)
  - [Common Patterns](#common-patterns)
    - [Pattern 1: Multi-Environment Secrets](#pattern-1-multi-environment-secrets)
    - [Pattern 2: Shared Team Secrets](#pattern-2-shared-team-secrets)
    - [Pattern 3: Application-Specific Secrets](#pattern-3-application-specific-secrets)

## Overview

Project Piper supports multiple credential management approaches:

1. **HashiCorp Vault** (Recommended): Centralized secret management
2. **Jenkins Credentials**: Native Jenkins credential store
3. **Platform Secrets**: Azure Key Vault, GitHub Secrets, etc.
4. **Environment Variables**: Direct injection via PIPER_* variables

**Key Principles**:
- Never hardcode secrets in configuration files
- Use Vault for centralized secret management
- Rotate credentials regularly
- Apply least-privilege access

## HashiCorp Vault Integration

### Vault Setup

**Prerequisites**:
1. Vault server accessible from pipeline
2. KV secrets engine enabled (v1 or v2)
3. AppRole or Token authentication configured
4. Policies granting read access to secrets

**Enable KV engine**:
```bash
# KV v2 (recommended)
vault secrets enable -path=piper kv-v2

# KV v1
vault secrets enable -path=piper kv
```

**Create policy**:
```hcl
# policy.hcl
path "piper/data/my-project/*" {
  capabilities = ["read"]
}

path "piper/data/GROUP-SECRETS/*" {
  capabilities = ["read"]
}
```

**Apply policy**:
```bash
vault policy write piper-policy policy.hcl
```

### Authentication Methods

**AppRole Authentication (Recommended)**:

```bash
# Enable AppRole
vault auth enable approle

# Create AppRole
vault write auth/approle/role/piper-role \
  token_policies="piper-policy" \
  token_ttl=1h \
  token_max_ttl=4h

# Get Role ID
vault read auth/approle/role/piper-role/role-id

# Generate Secret ID
vault write -f auth/approle/role/piper-role/secret-id
```

**Token Authentication**:

```bash
# Create token
vault token create -policy=piper-policy -ttl=1h
```

### Basic Configuration

**AppRole authentication**:

```yaml
# .pipeline/config.yml
general:
  vaultServerUrl: 'https://vault.company.com'
  vaultNamespace: 'team/project'  # Optional, for Vault Enterprise
  vaultPath: 'piper/my-project'
```

**Provide credentials** (Jenkins):
```yaml
# In custom defaults (not in project config)
hooks:
  vault:
    appRoleTokenCredentialsId: 'vault-app-role-id'
    appRoleSecretTokenCredentialsId: 'vault-app-role-secret'
```

**Provide credentials** (Environment):
```bash
export PIPER_vaultAppRoleID='role-id'
export PIPER_vaultAppRoleSecretID='secret-id'

# Or with Token
export PIPER_vaultToken='vault-token'
```

### Vault Path Hierarchy

Piper searches for secrets in this order:

1. **`<vaultPath>/<secretName>`**
   - Project-specific secrets
   - Highest priority

2. **`<vaultBasePath>/<vaultPipelineName>/<secretName>`**
   - Pipeline-specific secrets
   - Shared across environments

3. **`<vaultBasePath>/GROUP-SECRETS/<secretName>`**
   - Team/group secrets
   - Shared across projects

**Example configuration**:
```yaml
general:
  vaultServerUrl: 'https://vault.company.com'
  vaultBasePath: 'piper'
  vaultPipelineName: 'PIPELINE-12345'
  vaultPath: 'piper/my-project'
```

**Search order for secret `cf-password`**:
1. `piper/my-project/cf-password`
2. `piper/PIPELINE-12345/cf-password`
3. `piper/GROUP-SECRETS/cf-password`

### Automatic Parameter Resolution

Parameters with `VaultSecretName` suffix are automatically resolved:

**Configuration**:
```yaml
steps:
  cloudFoundryDeploy:
    cloudFoundry:
      org: 'my-org'
      space: 'production'
      username: 'cf-user@company.com'

    # Reference to Vault secret
    cloudFoundryPasswordVaultSecretName: 'cf-password'
```

**Vault secret path**: `piper/my-project/cf-password`

**Vault secret value**:
```json
{
  "password": "actual-password-value"
}
```

**Result**: Password is automatically injected into `cloudFoundryPassword` parameter.

## Vault Configuration Examples

### Example 1: Basic Vault Setup

**Vault structure**:
```
piper/
├── my-project/
│   ├── cf-password
│   ├── sonar-token
│   └── npm-token
```

**Configuration**:
```yaml
general:
  vaultServerUrl: 'https://vault.company.com'
  vaultPath: 'piper/my-project'

steps:
  cloudFoundryDeploy:
    cloudFoundryPasswordVaultSecretName: 'cf-password'

  sonarExecuteScan:
    sonarTokenVaultSecretName: 'sonar-token'

  npmExecute:
    npmTokenVaultSecretName: 'npm-token'
```

### Example 2: Cloud Foundry Credentials

**Vault secrets**:

```bash
# Create CF credentials in Vault
vault kv put piper/my-project/cf-credentials \
  username='cf-user@company.com' \
  password='secure-password'
```

**Configuration**:
```yaml
general:
  vaultServerUrl: 'https://vault.company.com'
  vaultPath: 'piper/my-project'

steps:
  cloudFoundryDeploy:
    cloudFoundry:
      org: 'my-org'
      space: 'production'

    # Vault will inject both username and password
    cloudFoundryCredentialsVaultSecretName: 'cf-credentials'
```

### Example 3: Multiple Vault Paths

**Use case**: Different secrets for different environments

**Vault structure**:
```
piper/
├── my-project-dev/
│   └── cf-password
├── my-project-staging/
│   └── cf-password
└── my-project-prod/
    └── cf-password
```

**Configuration** (dynamic path):
```yaml
general:
  vaultServerUrl: 'https://vault.company.com'
  vaultPath: 'piper/my-project-${env.ENVIRONMENT}'

steps:
  cloudFoundryDeploy:
    cloudFoundryPasswordVaultSecretName: 'cf-password'
```

**Usage**:
```bash
export ENVIRONMENT='prod'
# Uses: piper/my-project-prod/cf-password
```

### Example 4: Namespace Configuration

**Vault Enterprise with namespaces**:

```yaml
general:
  vaultServerUrl: 'https://vault.company.com'
  vaultNamespace: 'engineering/platform'
  vaultPath: 'piper/my-project'
```

## General Purpose Credentials

Vault general purpose credentials allow fetching any credentials for use in custom extensions or tests.

### Single Credential Path

**Configuration**:
```yaml
general:
  vaultServerUrl: 'https://vault.company.com'
  vaultPath: 'piper/my-project'

steps:
  someStep:
    vaultCredentialPath: 'api-credentials'
    vaultCredentialKeys:
      - 'apiKey'
      - 'apiSecret'
```

**Vault secret** (`piper/my-project/api-credentials`):
```json
{
  "apiKey": "key-123",
  "apiSecret": "secret-xyz"
}
```

**Environment variables exposed**:
```bash
PIPER_VAULTCREDENTIAL_APIKEY='key-123'
PIPER_VAULTCREDENTIAL_APISECRET='secret-xyz'

# Base64 encoded
PIPER_VAULTCREDENTIAL_APIKEY_BASE64='a2V5LTEyMw=='
PIPER_VAULTCREDENTIAL_APISECRET_BASE64='c2VjcmV0LXh5eg=='
```

### Multiple Credential Paths

**Configuration**:
```yaml
steps:
  someStep:
    vaultCredentialPath:
      - 'db-credentials'
      - 'api-credentials'

    vaultCredentialKeys:
      - ['dbHost', 'dbPassword']
      - ['apiKey', 'apiSecret']
```

**Environment variables**:
```bash
# From db-credentials
PIPER_VAULTCREDENTIAL_DBHOST='db.company.com'
PIPER_VAULTCREDENTIAL_DBPASSWORD='db-password'

# From api-credentials
PIPER_VAULTCREDENTIAL_APIKEY='key-123'
PIPER_VAULTCREDENTIAL_APISECRET='secret-xyz'
```

### Custom Environment Variable Prefix

**Single prefix**:
```yaml
steps:
  someStep:
    vaultCredentialPath: 'api-credentials'
    vaultCredentialKeys:
      - 'apiKey'
    vaultCredentialEnvPrefix: 'MY_APP_'
```

**Result**:
```bash
MY_APP_APIKEY='key-123'
```

**Multiple prefixes**:
```yaml
steps:
  someStep:
    vaultCredentialPath:
      - 'db-credentials'
      - 'api-credentials'

    vaultCredentialKeys:
      - ['dbPassword']
      - ['apiKey']

    vaultCredentialEnvPrefix:
      - 'DB_'
      - 'API_'
```

**Result**:
```bash
DB_DBPASSWORD='db-password'
API_APIKEY='key-123'
```

## Test Credentials (Deprecated)

**Note**: Use general purpose credentials instead.

**Configuration**:
```yaml
steps:
  someStep:
    vaultTestCredentialPath: 'test-credentials'
    vaultTestCredentialKeys:
      - 'testUser'
      - 'testPassword'
```

**Environment variables**:
```bash
PIPER_TESTCREDENTIAL_TESTUSER='test-user'
PIPER_TESTCREDENTIAL_TESTPASSWORD='test-password'
```

## Jenkins Credentials

### Credential Types

**1. Secret Text**:
```yaml
steps:
  sonarExecuteScan:
    sonarTokenCredentialsId: 'SONAR_TOKEN'
```

**2. Username with Password**:
```yaml
steps:
  cloudFoundryDeploy:
    cloudFoundry:
      credentialsId: 'CF_CREDENTIALS'
```

**3. SSH Username with Private Key**:
```yaml
general:
  gitSshKeyCredentialsId: 'GITHUB_SSH_KEY'
```

**4. Secret File**:
```yaml
steps:
  gcpPublish:
    serviceAccountKeyFileCredentialsId: 'GCP_KEY_FILE'
```

### Referencing Jenkins Credentials

**By ID**:
```yaml
general:
  gitSshKeyCredentialsId: 'github-ssh-key'

steps:
  cloudFoundryDeploy:
    cloudFoundry:
      credentialsId: 'CF_PROD_CREDENTIALS'

  sonarExecuteScan:
    sonarTokenCredentialsId: 'SONAR_TOKEN'
```

### Creating Jenkins Credentials

**Via Jenkins UI**:
1. Manage Jenkins → Credentials
2. Select domain (usually Global)
3. Add Credentials
4. Choose credential type
5. Enter ID and credential data

**Via Groovy script**:
```groovy
import com.cloudbees.plugins.credentials.*
import com.cloudbees.plugins.credentials.impl.*
import com.cloudbees.plugins.credentials.domains.*
import org.jenkinsci.plugins.plaincredentials.impl.*

// Secret Text
def secretText = new StringCredentialsImpl(
    CredentialsScope.GLOBAL,
    'SONAR_TOKEN',
    'SonarQube Token',
    hudson.util.Secret.fromString('token-value')
)

// Username with Password
def userPass = new UsernamePasswordCredentialsImpl(
    CredentialsScope.GLOBAL,
    'CF_CREDENTIALS',
    'Cloud Foundry Credentials',
    'username',
    'password'
)

// Add to store
def store = SystemCredentialsProvider.getInstance().getStore()
store.addCredentials(Domain.global(), secretText)
store.addCredentials(Domain.global(), userPass)
```

## Platform-Specific Credentials

### Azure DevOps

**Variable Groups**:
```yaml
# azure-pipelines.yml
variables:
  - group: 'piper-secrets'
  - group: 'cloud-foundry-credentials'

steps:
  - task: Piper@1
    inputs:
      piperCommand: 'cloudFoundryDeploy'
    env:
      PIPER_cloudFoundry_username: $(cf-username)
      PIPER_cloudFoundry_password: $(cf-password)
```

**Azure Key Vault**:
```yaml
# Link variable group to Key Vault
variables:
  - group: 'key-vault-secrets'  # Linked to Azure Key Vault

steps:
  - task: Piper@1
    env:
      PIPER_cloudFoundry_password: $(cf-password)
```

### GitHub Actions

**Repository Secrets**:
```yaml
# .github/workflows/deploy.yml
jobs:
  deploy:
    runs-on: ubuntu-latest

    steps:
      - uses: SAP/project-piper-action@main
        with:
          command: 'cloudFoundryDeploy'
        env:
          PIPER_cloudFoundry_username: ${{ secrets.CF_USERNAME }}
          PIPER_cloudFoundry_password: ${{ secrets.CF_PASSWORD }}
```

**Environment Secrets**:
```yaml
jobs:
  deploy-production:
    runs-on: ubuntu-latest
    environment: production  # Uses environment-specific secrets

    steps:
      - uses: SAP/project-piper-action@main
        with:
          command: 'cloudFoundryDeploy'
        env:
          PIPER_cloudFoundry_password: ${{ secrets.CF_PROD_PASSWORD }}
```

## Credential Resolution Process

**Order of precedence** (highest to lowest):

1. **Direct parameters** (in pipeline code)
2. **Environment variables** (PIPER_*)
3. **Stage configuration** (in stages section)
4. **Step configuration** (in steps section)
5. **Vault secrets** (if configured)
6. **General configuration** (in general section)
7. **Jenkins credentials** (if credential ID provided)

**Example**:
```yaml
# Configuration
steps:
  cloudFoundryDeploy:
    cloudFoundry:
      username: 'default-user'
    cloudFoundryPasswordVaultSecretName: 'cf-password'

# If Vault is available: uses Vault secret
# If Vault not available: parameter remains unset (or uses Jenkins credential)
```

## Advanced Vault Features

### Disabling Vault Overwrite

**Prevent Vault from overwriting configured values**:

```yaml
# Global
general:
  vaultDisableOverwrite: true

# Stage-specific
stages:
  Release:
    vaultDisableOverwrite: true

# Step-specific
steps:
  cloudFoundryDeploy:
    vaultDisableOverwrite: true
```

**Use case**: When you want to use configuration values instead of Vault for certain parameters.

### Skipping Vault Lookup

**Skip Vault entirely for certain steps**:

```yaml
# Global
general:
  skipVault: true

# Step-specific
steps:
  npmExecute:
    skipVault: true  # Don't lookup Vault secrets for this step
```

### Vault Secret Files

**Store file content in Vault**:

**Vault secret** (`piper/my-project/service-key`):
```json
{
  "credentials": "{\"client_id\":\"...\",\"client_secret\":\"...\"}"
}
```

**Configuration**:
```yaml
steps:
  someStep:
    serviceKeyVaultSecretName: 'service-key'
```

**Result**: File is created in temporary directory and path is provided to step.

## Security Best Practices

1. **Use Vault for All Secrets**
   ```yaml
   # Good
   steps:
     cloudFoundryDeploy:
       cloudFoundryPasswordVaultSecretName: 'cf-password'

   # Bad - never hardcode
   steps:
     cloudFoundryDeploy:
       cloudFoundry:
         password: 'hardcoded-password'  # NEVER DO THIS
   ```

2. **Rotate Credentials Regularly**
   ```bash
   # Update Vault secret
   vault kv put piper/my-project/cf-password password='new-password'
   ```

3. **Use AppRole Instead of Tokens**
   - AppRole supports secret rotation
   - More secure than long-lived tokens

4. **Apply Least Privilege**
   ```hcl
   # Grant only read access
   path "piper/data/my-project/*" {
     capabilities = ["read"]
   }
   ```

5. **Use Separate Secrets per Environment**
   ```
   piper/
   ├── my-project-dev/
   ├── my-project-staging/
   └── my-project-prod/
   ```

6. **Never Commit Secrets**
   ```gitignore
   # .gitignore
   **/*secret*
   **/*password*
   **/*credentials*
   .env
   ```

7. **Enable Vault Audit Logging**
   ```bash
   vault audit enable file file_path=/var/log/vault-audit.log
   ```

8. **Use Short-Lived Tokens**
   ```bash
   vault write auth/approle/role/piper-role \
     token_ttl=1h \
     token_max_ttl=4h
   ```

## Troubleshooting

### Vault Authentication Failures

**Error**: `Vault authentication failed`

**Check**:
```yaml
# Verify Vault URL
general:
  vaultServerUrl: 'https://vault.company.com'  # Correct URL?

# Verify credentials are provided
# In Jenkins: Check credential IDs exist
# In CLI: Check environment variables
```

**Debug**:
```yaml
general:
  verbose: true  # Enable detailed logging
```

### Secret Not Found

**Error**: `Secret not found in Vault`

**Check search paths**:
```bash
# Check if secret exists in any of these paths
vault kv get piper/my-project/secret-name
vault kv get piper/PIPELINE-12345/secret-name
vault kv get piper/GROUP-SECRETS/secret-name
```

**Create secret**:
```bash
vault kv put piper/my-project/secret-name \
  password='secret-value'
```

### Permission Denied

**Error**: `Permission denied`

**Check policy**:
```bash
# View current token capabilities
vault token capabilities piper/data/my-project/secret-name

# Should show: ["read"]
```

**Update policy**:
```hcl
path "piper/data/my-project/*" {
  capabilities = ["read"]
}
```

### Token Expiration

**Error**: `Token expired`

**Solutions**:
1. Use AppRole (supports renewal)
2. Increase token TTL
3. Use shorter pipeline execution times

**Check token info**:
```bash
vault token lookup
```

## Migration Guide

### From Jenkins Credentials to Vault

**Before** (Jenkins credentials):
```yaml
steps:
  cloudFoundryDeploy:
    cloudFoundry:
      credentialsId: 'CF_CREDENTIALS'
```

**After** (Vault):

1. **Create secret in Vault**:
```bash
vault kv put piper/my-project/cf-credentials \
  username='cf-user@company.com' \
  password='secure-password'
```

2. **Update configuration**:
```yaml
general:
  vaultServerUrl: 'https://vault.company.com'
  vaultPath: 'piper/my-project'

steps:
  cloudFoundryDeploy:
    cloudFoundryCredentialsVaultSecretName: 'cf-credentials'
```

### From Hardcoded to Vault

**Before** (hardcoded):
```yaml
steps:
  sonarExecuteScan:
    serverUrl: 'https://sonarqube.company.com'
    token: 'hardcoded-token'  # Bad practice
```

**After** (Vault):

1. **Store secret**:
```bash
vault kv put piper/my-project/sonar-token \
  token='actual-token-value'
```

2. **Update configuration**:
```yaml
steps:
  sonarExecuteScan:
    serverUrl: 'https://sonarqube.company.com'
    sonarTokenVaultSecretName: 'sonar-token'
```

## Common Patterns

### Pattern 1: Multi-Environment Secrets

**Vault structure**:
```
piper/
├── my-app-dev/
│   ├── cf-password
│   └── db-password
├── my-app-staging/
│   ├── cf-password
│   └── db-password
└── my-app-prod/
    ├── cf-password
    └── db-password
```

**Configuration**:
```yaml
general:
  vaultPath: 'piper/my-app-${env.ENVIRONMENT}'

steps:
  cloudFoundryDeploy:
    cloudFoundryPasswordVaultSecretName: 'cf-password'
```

### Pattern 2: Shared Team Secrets

**Vault structure**:
```
piper/
├── GROUP-SECRETS/
│   ├── npm-token      # Shared across team
│   ├── docker-token   # Shared across team
│   └── sonar-token    # Shared across team
└── my-project/
    └── cf-password    # Project-specific
```

**Configuration**:
```yaml
general:
  vaultBasePath: 'piper'
  vaultPath: 'piper/my-project'

steps:
  npmExecute:
    npmTokenVaultSecretName: 'npm-token'  # Found in GROUP-SECRETS

  cloudFoundryDeploy:
    cloudFoundryPasswordVaultSecretName: 'cf-password'  # Found in my-project
```

### Pattern 3: Application-Specific Secrets

**For custom extensions**:

```yaml
steps:
  myCustomStep:
    vaultCredentialPath: 'app-secrets'
    vaultCredentialKeys:
      - 'apiKey'
      - 'apiSecret'
      - 'webhookUrl'
    vaultCredentialEnvPrefix: 'APP_'
```

**In custom extension** (`.pipeline/extensions/MyCustomStep.groovy`):
```groovy
void call(Map parameters) {
    sh '''
        curl -H "Authorization: Bearer ${APP_APIKEY}" \
             -X POST ${APP_WEBHOOKURL} \
             -d "secret=${APP_APISECRET}"
    '''
}
```

---

**Congratulations!** You've completed the Project Piper configuration documentation series. You should now have a comprehensive understanding of how to configure and secure your CI/CD pipelines using Project Piper.

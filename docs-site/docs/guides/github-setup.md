# GitHub Actions Setup Guide

Complete guide for setting up Project Piper on GitHub Actions.

## Prerequisites

### Required
- **GitHub Repository**: With Actions enabled
- **Secrets**: Configured in repository settings
- **Runner**: GitHub-hosted or self-hosted

### Runner Requirements

**GitHub-Hosted**: Ubuntu, Windows, macOS with Docker pre-installed

**Self-Hosted**: Linux (Ubuntu 20.04+), 2+ CPU cores, 8GB+ RAM, Docker 20.10+, Git 2.20+

## Installation

### Step 1: Create Workflow File

Create `.github/workflows/piper.yml`:

```yaml
name: Piper CI/CD

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Run Piper Build
        uses: SAP/project-piper-action@main
        with:
          step-name: 'mavenBuild'
          flags: '--publish --createBOM'
```

### Step 2: Configure Secrets

Navigate to **Settings** → **Secrets and variables** → **Actions**:

```text
PIPER_VAULTAPPROLEID: <vault-role-id>
PIPER_VAULTAPPROLESECRETID: <vault-secret-id>
GH_TOKEN: <github-token>
```

### Step 3: Create Piper Configuration

Create `.pipeline/config.yml`:

```yaml
general:
  buildTool: 'maven'
  productiveBranch: 'main'

steps:
  mavenBuild:
    goals: ['clean', 'install']
```

## Configuration

### Basic Workflow Structure

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main]
  pull_request:

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3

      - name: Piper Build
        uses: SAP/project-piper-action@main
        with:
          step-name: 'mavenBuild'
          flags: '--flag1 --flag2'
        env:
          PIPER_vaultAppRoleID: ${{ secrets.PIPER_VAULTAPPROLEID }}
          PIPER_vaultAppRoleSecretID: ${{ secrets.PIPER_VAULTAPPROLESECRETID }}
```

### Action Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `step-name` | Piper step to execute | Yes | - |
| `flags` | Command flags | No | '' |
| `piper-version` | Piper version | No | latest |
| `custom-defaults-paths` | Custom configs | No | '' |

### Configuration File

```yaml
general:
  buildTool: 'npm'
  productiveBranch: 'main'
  vaultServerUrl: 'https://vault.example.com'
  vaultNamespace: 'piper'

steps:
  npmExecuteScripts:
    runScripts: ['build', 'test']
  mavenBuild:
    goals: ['clean', 'verify']
```

## Common Patterns

### Multi-Step Build

```yaml
name: Build and Test

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Version
        uses: SAP/project-piper-action@main
        with:
          step-name: 'artifactPrepareVersion'

      - name: Build
        uses: SAP/project-piper-action@main
        with:
          step-name: 'mavenBuild'

      - name: Test
        uses: SAP/project-piper-action@main
        with:
          step-name: 'mavenExecute'
          flags: '--goals test'
```

### Parallel Jobs

```yaml
name: Parallel Pipeline

on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: SAP/project-piper-action@main
        with:
          step-name: 'npmExecuteScripts'
          flags: '--runScripts=build'

  test-unit:
    runs-on: ubuntu-latest
    needs: build
    steps:
      - uses: actions/checkout@v3
      - uses: SAP/project-piper-action@main
        with:
          step-name: 'npmExecuteScripts'
          flags: '--runScripts=test:unit'

  test-integration:
    runs-on: ubuntu-latest
    needs: build
    steps:
      - uses: actions/checkout@v3
      - uses: SAP/project-piper-action@main
        with:
          step-name: 'npmExecuteScripts'
          flags: '--runScripts=test:integration'
```

### Matrix Builds

```yaml
name: Matrix Build

on: [push]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest]
        node-version: [16, 18, 20]

    runs-on: ${{ matrix.os }}

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: ${{ matrix.node-version }}
      - uses: SAP/project-piper-action@main
        with:
          step-name: 'npmExecuteScripts'
          flags: '--runScripts=test'
```

### Conditional Deployment

```yaml
name: Build and Deploy

on:
  push:
    branches: [main, develop]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: SAP/project-piper-action@main
        with:
          step-name: 'mavenBuild'

  deploy-dev:
    needs: build
    if: github.ref == 'refs/heads/develop'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: SAP/project-piper-action@main
        with:
          step-name: 'cloudFoundryDeploy'
        env:
          CF_SPACE: 'dev'

  deploy-prod:
    needs: build
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v3
      - uses: SAP/project-piper-action@main
        with:
          step-name: 'cloudFoundryDeploy'
        env:
          CF_SPACE: 'prod'
```

### Custom Defaults from Remote

```yaml
name: Build with Custom Defaults

on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: SAP/project-piper-action@main
        with:
          step-name: 'mavenBuild'
          custom-defaults-paths: 'myorg/configs/defaults.yml@v1.0.0'
        env:
          PIPER_ACTION_GITHUB_TOOLS_TOKEN: ${{ secrets.GH_TOKEN }}
```

## Examples

### Node.js Application

```yaml
name: Node.js CI/CD

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: SAP/project-piper-action@main
        with:
          step-name: 'npmExecuteLint'

      - uses: SAP/project-piper-action@main
        with:
          step-name: 'npmExecuteScripts'
          flags: '--runScripts=build,test'
```

### Maven with Security Scan

```yaml
name: Maven Build and Scan

on: [push]

jobs:
  build-and-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: SAP/project-piper-action@main
        with:
          step-name: 'mavenBuild'

      - uses: SAP/project-piper-action@main
        with:
          step-name: 'whitesourceExecuteScan'
        env:
          PIPER_whitesourceUserKey: ${{ secrets.WS_USER_KEY }}
```

### MTA Deployment

```yaml
name: MTA Build and Deploy

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: SAP/project-piper-action@main
        with:
          step-name: 'mtaBuild'
          flags: '--buildTarget CF'

      - uses: actions/upload-artifact@v3
        with:
          name: mta-archive
          path: '**/*.mtar'

  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/download-artifact@v3
        with:
          name: mta-archive

      - uses: SAP/project-piper-action@main
        with:
          step-name: 'cloudFoundryDeploy'
        env:
          PIPER_vaultAppRoleID: ${{ secrets.PIPER_VAULTAPPROLEID }}
```

### Reusable Workflows

**Create** `.github/workflows/piper-build.yml`:
```yaml
name: Reusable Build

on:
  workflow_call:
    inputs:
      step-name:
        required: true
        type: string

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: SAP/project-piper-action@main
        with:
          step-name: ${{ inputs.step-name }}
```

**Use** `.github/workflows/main.yml`:
```yaml
name: Main Pipeline

on: [push]

jobs:
  build:
    uses: ./.github/workflows/piper-build.yml
    with:
      step-name: 'mavenBuild'
```

## Troubleshooting

### Action Not Found

**Error:** `Unable to resolve action SAP/project-piper-action@main`

**Solution:**
- Verify repository name
- Use specific version: `@v1.0.0`
- Check repository accessibility

### Credentials Not Available

**Error:** `Could not resolve credentials`

**Solution:**
1. Check secrets in repository settings
2. Verify secret names match:
   ```yaml
   env:
     PIPER_vaultAppRoleID: ${{ secrets.PIPER_VAULTAPPROLEID }}
   ```

### Step Execution Fails

**Solution:**
Enable verbose:
```yaml
with:
  flags: '--verbose'
```

### Docker Permission Issues (Self-hosted)

```bash
sudo usermod -aG docker $USER
sudo systemctl restart docker
```

### Out of Disk Space (Self-hosted)

```bash
docker system prune -af
cd /actions-runner/_work/_temp && rm -rf *
```

### Custom Defaults Not Found

**Solution:**
```yaml
custom-defaults-paths: 'org/repo/path/file.yml@ref'
env:
  PIPER_ACTION_GITHUB_TOOLS_TOKEN: ${{ secrets.GH_TOKEN }}
```

### Debug Logging

Add secret: `ACTIONS_STEP_DEBUG` = `true`

```yaml
steps:
  - name: Debug
    run: |
      echo "OS: $RUNNER_OS"
      echo "Workspace: $GITHUB_WORKSPACE"
      df -h
      docker --version
```

## Best Practices

1. **Pin Versions**: Use specific tags for production
2. **Use Environments**: For deployment approvals
3. **Cache Dependencies**: Speed up builds
4. **Matrix Builds**: Test multiple versions
5. **Artifact Management**: Store build outputs
6. **Secret Rotation**: Regular credential updates
7. **Self-Hosted Runners**: Better performance
8. **Resource Cleanup**: Clean containers

## Additional Resources

- [GitHub Actions Docs](https://docs.github.com/en/actions)
- [Piper Action](https://github.com/SAP/project-piper-action)
- [Project Piper](https://www.project-piper.io/)
- [Workflow Syntax](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions)

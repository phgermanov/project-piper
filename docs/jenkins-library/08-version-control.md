# Jenkins Library - Version Control and GitHub Integration

Comprehensive guide for version management and GitHub automation in jenkins-library.

## Overview

**Version Control (2 steps)**: Automatic and manual version management
**GitHub Integration (6 steps)**: Issues, PRs, releases, status checks, branch protection

---

## artifactPrepareVersion

Automatic version generation with Git tagging for continuous delivery.

### Versioning Types

- **`cloud`**: Auto-version `<major>.<minor>.<patch>-<timestamp>-<commitId>` + Git tag (default)
- **`cloud_noTag`**: Auto-version without Git tag (no traceability)
- **`library`**: Pure semantic `<major>.<minor>.<patch>` (team-controlled)

### Usage

```groovy
artifactPrepareVersion script: this
```

```yaml
steps:
  artifactPrepareVersion:
    versioningType: 'cloud'          # cloud, cloud_noTag, library
    buildTool: 'maven'               # maven, npm, gradle, mta, golang, pip, helm, docker, custom

    # Version format
    includeCommitId: true
    shortCommitId: false             # 7-char hash
    timestampTemplate: '%Y%m%d%H%M%S'
    tagPrefix: 'build_'

    # Multi-tool versioning
    additionalTargetTools: ['helm', 'docker']
    additionalTargetDescriptors: ['**/Chart.yaml']

    # Authentication
    gitSshKeyCredentialsId: 'git-ssh-key'
```

### Supported Build Tools

| Tool | File | Version Location |
|------|------|------------------|
| maven | pom.xml | `<version>` |
| npm | package.json | `version` |
| gradle | build.gradle | `version` |
| mta | mta.yaml | `version` |
| helm | Chart.yaml | `version` |
| custom | Any | Configurable |

### Custom Build Tool

```yaml
# Text file: VERSION.txt
artifactPrepareVersion:
  buildTool: 'custom'
  filePath: 'VERSION.txt'

# JSON: metadata.json
artifactPrepareVersion:
  buildTool: 'custom'
  filePath: 'metadata.json'
  customVersionField: 'appVersion'
```

---

## artifactSetVersion

Manual version setting in build descriptors (alias of `artifactPrepareVersion`).

```groovy
artifactSetVersion(script: this, buildTool: 'maven', version: '1.2.3')
```

---

## GitHub Integration

**Prerequisites**: Create GitHub personal access token with `repo`, `write:discussion`, `admin:repo_hook` scopes. Store in Jenkins credentials as "Secret text".

### githubCheckBranchProtection

Verify branch protection rules compliance.

```groovy
githubCheckBranchProtection(
    script: this, owner: 'my-org', repository: 'my-repo', branch: 'main',
    requiredChecks: ['ci/jenkins'], requireEnforceAdmins: true,
    requiredApprovingReviewCount: 2, githubTokenCredentialsId: 'github-token'
)
```

**Configuration**:
```yaml
steps:
  githubCheckBranchProtection:
    owner: 'my-org'
    repository: 'my-repo'
    branch: 'main'
    requiredChecks: ['ci/jenkins', 'security/scan']
    requireEnforceAdmins: true
    requiredApprovingReviewCount: 2
    githubTokenCredentialsId: 'github-token'
```

### githubCommentIssue

Add comments to issues or pull requests (supports Markdown).

```groovy
githubCommentIssue(
    script: this, owner: 'my-org', repository: 'my-repo',
    number: 42, body: '## Build Status\n\n✅ Build successful!',
    githubTokenCredentialsId: 'github-token'
)
```

**Example - PR Notification**:
```groovy
if (env.CHANGE_ID) {
    githubCommentIssue(
        script: this, number: env.CHANGE_ID.toInteger(),
        body: "## Build #${env.BUILD_NUMBER}\n**Status**: ${currentBuild.result}\n[View](${env.BUILD_URL})",
        githubTokenCredentialsId: 'github-token'
    )
}
```

### githubCreateIssue

Create GitHub issues for automated reporting.

```groovy
githubCreateIssue(
    script: this, owner: 'my-org', repository: 'my-repo',
    title: 'Security vulnerability detected', body: 'Details...',
    assignees: ['security-team'], updateExisting: true,
    githubTokenCredentialsId: 'github-token'
)
```

**Configuration**:
```yaml
steps:
  githubCreateIssue:
    owner: 'my-org'
    repository: 'my-repo'
    title: 'Automated Issue'
    body: 'Description'                    # Or use bodyFilePath
    assignees: ['user1']
    updateExisting: false                  # Update existing instead of creating new
    githubTokenCredentialsId: 'github-token'
```

### githubCreatePullRequest

Create pull requests for GitOps workflows.

```groovy
githubCreatePullRequest(
    script: this, owner: 'my-org', repository: 'my-repo',
    head: 'feature-branch', base: 'main',
    title: 'Add feature', body: 'Description',
    labels: ['enhancement'], githubTokenCredentialsId: 'github-token'
)
```

**Configuration**:
```yaml
steps:
  githubCreatePullRequest:
    owner: 'my-org'
    repository: 'my-repo'
    head: 'feature-branch'                 # Source
    base: 'main'                           # Target
    title: 'Update dependencies'
    body: '## Changes\n- Updated deps'
    assignees: ['reviewer1']
    labels: ['dependencies', 'automated']
    githubTokenCredentialsId: 'github-token'
```

**Example - Automated Update**:
```groovy
sh 'npm update && git checkout -b deps-${BUILD_NUMBER} && git add . && git commit -m "deps" && git push'
githubCreatePullRequest(script: this, head: "deps-${BUILD_NUMBER}", base: 'main',
    title: "chore: Update dependencies", githubTokenCredentialsId: 'github-token')
```

### githubPublishRelease

Create GitHub releases with tags, notes, and artifacts.

```groovy
githubPublishRelease(
    script: this, owner: 'my-org', repository: 'my-repo',
    version: '1.2.3', addClosedIssues: true, addDeltaToLastRelease: true,
    assetPathList: ['dist/*.jar'], githubTokenCredentialsId: 'github-token'
)
```

**Configuration**:
```yaml
steps:
  githubPublishRelease:
    owner: 'my-org'
    repository: 'my-repo'
    version: '1.2.3'                       # From artifactPrepareVersion
    tagPrefix: 'v'                         # Creates 'v1.2.3'
    commitish: 'main'
    releaseBodyHeader: '## Release Notes'
    addClosedIssues: true                  # Include closed issues/PRs
    addDeltaToLastRelease: true            # Commit comparison link
    labels: ['bug', 'enhancement']         # Filter issues
    excludeLabels: ['wontfix']
    preRelease: false
    assetPathList: ['dist/*.jar', 'checksums.txt']
    githubTokenCredentialsId: 'github-token'
```

**Example - Complete Release**:
```groovy
stage('Release') {
    when { branch 'main' }
    steps {
        script {
            artifactPrepareVersion script: this
            mavenBuild script: this
            githubPublishRelease(script: this, addClosedIssues: true,
                assetPathList: ['target/*.jar'], githubTokenCredentialsId: 'github-token')
        }
    }
}
```

### githubSetCommitStatus

Set commit status checks (enforced by branch protection).

```groovy
githubSetCommitStatus(
    script: this, owner: 'my-org', repository: 'my-repo',
    commitId: env.GIT_COMMIT, context: 'ci/jenkins', status: 'success',
    description: 'Build passed', targetUrl: env.BUILD_URL,
    githubTokenCredentialsId: 'github-token'
)
```

**Status Values**: `pending`, `success`, `failure`

**Configuration**:
```yaml
steps:
  githubSetCommitStatus:
    owner: 'my-org'
    repository: 'my-repo'
    commitId: 'abc123'                     # From commonPipelineEnvironment
    context: 'ci/jenkins'                  # Label in PR
    status: 'success'                      # pending/success/failure
    description: 'Build passed'
    targetUrl: 'https://jenkins.example.com/job/123'
    githubTokenCredentialsId: 'github-token'
```

**Example - Status Tracking**:
```groovy
githubSetCommitStatus(script: this, context: 'ci/build', status: 'pending', description: 'Building...')
try {
    mavenBuild script: this
    githubSetCommitStatus(script: this, context: 'ci/build', status: 'success', targetUrl: env.BUILD_URL)
} catch (e) {
    githubSetCommitStatus(script: this, context: 'ci/build', status: 'failure', targetUrl: env.BUILD_URL)
    throw e
}
```

---

## Common Workflows

### Workflow 1: Automated Versioning and Release

```yaml
# .pipeline/config.yml - Complete release workflow
stages:
  Init:
    artifactPrepareVersion: true
  Build:
    mavenBuild: true
  Release:
    githubPublishRelease: true

steps:
  artifactPrepareVersion:
    versioningType: 'cloud'
    tagPrefix: 'v'
  githubPublishRelease:
    addClosedIssues: true
    addDeltaToLastRelease: true
```

### Workflow 2: Pull Request Automation

```groovy
when { changeRequest() }
githubSetCommitStatus(script: this, context: 'pr/checks', status: 'pending')
mavenBuild script: this
checkmarxExecuteScan script: this
githubCommentIssue(script: this, number: env.CHANGE_ID.toInteger(), body: "✅ Checks passed!")
githubSetCommitStatus(script: this, context: 'pr/checks', status: 'success')
```

### Workflow 3: GitOps Configuration Updates

```groovy
sh """
    git checkout -b config-${BUILD_NUMBER}
    sed -i 's/version: .*/version: ${VERSION}/' k8s/deployment.yaml
    git add k8s/deployment.yaml && git commit -m 'chore: update' && git push
"""
githubCreatePullRequest(script: this, head: "config-${BUILD_NUMBER}", base: 'main',
    title: "Update to ${VERSION}", githubTokenCredentialsId: 'github-token')
```

### Workflow 4: Security Issue Reporting

```groovy
def findings = checkmarxExecuteScan(script: this)
if (findings > 0) {
    githubCreateIssue(script: this, title: "Security: ${findings} vulnerabilities",
        bodyFilePath: 'report.md', assignees: ['security-team'], updateExisting: true)
}
```

### Workflow 5: Multi-Status Pipeline

```groovy
def setStatus(ctx, st, desc) {
    githubSetCommitStatus(script: this, context: ctx, status: st, description: desc,
        targetUrl: env.BUILD_URL, githubTokenCredentialsId: 'github-token')
}

stage('Build') {
    setStatus('ci/build', 'pending', 'Building...')
    try {
        mavenBuild script: this
        setStatus('ci/build', 'success', 'Passed')
    } catch (e) {
        setStatus('ci/build', 'failure', 'Failed')
        throw e
    }
}
```

---

## Best Practices

### Version Management

**Cloud deployments - use automatic versioning:**
```yaml
artifactPrepareVersion:
  versioningType: 'cloud'        # Unique versions with timestamp
  includeCommitId: true          # Full traceability
  fetchCoordinates: true         # For downstream steps
```

**Libraries - use semantic versioning:**
```yaml
artifactPrepareVersion:
  versioningType: 'library'      # Team controls major.minor.patch
```

**Multi-tool projects:**
```yaml
artifactPrepareVersion:
  buildTool: 'maven'
  additionalTargetTools: ['helm', 'docker']
```

### GitHub Authentication

- **Always use Jenkins credentials**: `githubTokenCredentialsId: 'github-token'`
- **Never hardcode tokens**: Avoid `token: 'ghp_xxx'` in configuration
- **Use Vault for enterprise**: Configure `githubVaultSecretName`
- **Minimum scopes**: `repo`, `write:discussion`, `admin:repo_hook`

### Status Checks

**Proper status lifecycle:**
```groovy
githubSetCommitStatus(context: 'ci/test', status: 'pending', description: 'Running...')
try {
    runTests()
    githubSetCommitStatus(context: 'ci/test', status: 'success', targetUrl: env.BUILD_URL)
} catch (e) {
    githubSetCommitStatus(context: 'ci/test', status: 'failure', targetUrl: env.BUILD_URL)
    throw e
}
```

**Use descriptive contexts:**
- `ci/build`, `ci/test`, `ci/integration` - Build/test steps
- `security/scan`, `security/dependencies` - Security checks
- `quality/sonar`, `quality/coverage` - Code quality

### Pull Requests

```yaml
githubCreatePullRequest:
  title: 'feat: Add authentication'        # Use conventional commits
  body: '## Changes\n- JWT\n## Testing\n- [x] Tests\nCloses #123'
  labels: ['enhancement', 'security']
```

### Releases

```yaml
githubPublishRelease:
  tagPrefix: 'v'                           # v1.2.3
  addClosedIssues: true                    # Auto-changelog
  addDeltaToLastRelease: true              # Commit comparison
  labels: ['bug', 'enhancement']           # Filter issues
  excludeLabels: ['wontfix']
  assetPathList: ['dist/*.jar']
```

**Semantic versioning**: Major (1.0.0) = breaking, Minor (0.1.0) = features, Patch (0.0.1) = fixes

### Error Handling

**Non-critical**: Catch and continue
```groovy
try { githubCreatePullRequest(/*...*/) }
catch (e) { echo "Warning: ${e.message}" }
```

**Critical**: Let fail
```groovy
githubPublishRelease(/*...*/)  // Fail pipeline if release cannot be created
```

### Branch Protection

```groovy
githubCheckBranchProtection(script: this, branch: 'main',
    requiredChecks: ['ci/build', 'ci/test'], requireEnforceAdmins: true,
    requiredApprovingReviewCount: 2)
```

---

## Configuration Reference

**Common GitHub Parameters**: `owner`/`githubOrg`, `repository`/`githubRepo`, `githubTokenCredentialsId`, `apiUrl`, `serverUrl`

**Version Control Parameters**: `versioningType` (cloud/cloud_noTag/library), `buildTool` (auto-detected), `tagPrefix`, `includeCommitId`, `timestampTemplate`

---

## Summary

jenkins-library provides comprehensive version control and GitHub integration:

**Version Control**:
- Automatic versioning with multiple patterns
- Support for 10+ build tools (auto-detection)
- Git tagging and authentication
- Custom build tool support

**GitHub Integration**:
- Complete PR/issue automation
- Commit status checks
- Release management with changelogs
- Branch protection verification

**Key Features**:
- Security-first (credentials, Vault)
- GitOps workflows
- Multi-tool versioning
- Professional CI/CD automation

**Use Cases**:
- Continuous deployment with unique versions
- Library releases with semantic versioning
- Automated dependency updates
- Security vulnerability reporting
- Release automation with changelogs

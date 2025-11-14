---
date: 2025-02-05
title: 'Using GitHub-provided automatic token for Piper GPP on GitHub Actions'
authors:
  - ashly
  - gulomjon
  - anil
categories:
  - General Purpose Pipeline
  - GitHub Actions
---

Piper is providing the possibility to use the [GitHub automatic token](https://docs.github.com/en/actions/security-for-github-actions/security-guides/automatic-token-authentication) for Piper steps which need GitHub personal access token (PAT) on **GitHub Actions**.
This means that it is not necessary to maintain a PAT in Vault for Piper steps which need one on **Piper GitHubActions**

<!-- more -->
This will hugely benefit users of Piper steps to reduce the overload of maintaining a PAT in Vault.

## 📢 Do I need to do something?

### For new Pipelines via Hyperspace Portal

Nothing to be done. The automatic GitHub token will be used if the PAT is not found in Vault.

### For existing Pipelines using Piper General Purpose Pipeline

If your PAT is maintained in Vault , no immediate action is needed if you want to **not** use this functionality.
However, if you wish to use the GitHub Automatic token instead of configuring your PAT in vault , please follow below steps

- Remove secret in Vault, if any at the pipeline or pipeline group level
- Following write permissions to be set in your GitHub Actions workflow which calls Piper GPP workflow (i.e.,sap-piper-workflow.yml,see example below )

  ```yaml
  name: Piper
  on:
    workflow_dispatch:
  jobs:
    piper:
      uses: project-piper/piper-pipeline-github/.github/workflows/sap-piper-workflow.yml@v1.17.0
      secrets: inherit
      permissions:
        contents: write  # write required for piper to create tags and releases
        id-token: write  # for connections to system-to-system trust
  ```

### For existing Pipelines using extensions

If you want to pass the GitHub Automatic token in your extension to a piper step, please provide the token as a step flag. For eg., your composite action.yml can be something as below to directly call the sapPipelineInit & artifactPrepareVersion step via a local extension. Note that the GitHub Automatic Token needs to be passed as a flag to the correct parameter of respective step

Location: .pipeline/extensions/preAcceptance
File: action.yml

```yaml

name: PreAcceptance
runs:
  using: composite
  steps:
    - name: sapPipelineInit
      uses: SAP/project-piper-action@main
      with:
        step-name: sapPipelineInit
        flags: --githubToken {{ github.token }}
    - name: artifactPrepareVersion
        uses: SAP/project-piper-action@main
        with:
          step-name: artifactPrepareVersion
          docker-image: ${{ env.dockerImage }}
          flags: --username github-actions --password ${{ github.token }}
```

**Note**
The automatic token currently doesn't apply for `vaultRotateSecretId` step as GitHub repository secrets cannot be updated by the token.
For users of Hyperspace Portal, secrets [are automatically refreshed](https://pages.github.tools.sap/hyperspace/cicd-setup-documentation/managed-services/build/piper.html)

## 💡 What do I need to know?

Vault secret is provided precedence over the automatic token.
This ensures that this functionality is not a breaking change for current users of Piper GitHubActions including users not yet onboarded to Hyperspace Portal.
Hence in order to use GitHub Automatic token, make sure Vault secret is not found at pipeline or pipeline Group level.

Following table provides more details on steps and respective stages where new token is enabled:

| Piper step   | Piper stage  | Automatic GitHub token enabled?  | Additional Info |
|--------------|--------------|----------------------------------|----------------|
| sapPipelineInit | Init | Yes |    |
| artifactPrepareVersion| Build | Yes |   |
| sonarExecuteScan| Build | No | GitHub token not necessary for latest SonarQube servers|
| gitopsUpdateDeployment | Acceptance, Performance, Release | Yes |   |
| kubernetesDeploy | Acceptance, Performance, Release | Yes |   |
| githubPublishRelease | Release | Yes |    |
| sapCollectInsights | Release | Yes |    |
| sapReportPipelineStatus | Release | Yes |   |
| vaultRotateSecretId| Post| No | See note above |

The stage workflows(`init.yml`, `build.yml`, etc.) in Piper are modified to pass the new automatic token.

## 📖 Learn more

- [Official documentation on GitHub Automatic Token](https://docs.github.com/en/actions/security-for-github-actions/security-guides/automatic-token-authentication)

- [PR introducing the change](https://github.tools.sap/project-piper/piper-pipeline-github/pull/237)

- [Permissions on the GitHub Automatic Token](https://docs.github.com/en/actions/security-for-github-actions/security-guides/automatic-token-authentication#permissions-for-the-github_token)

- [News on Hyperspace System Trust for Sonar](../../2024/10/trustengineSonarIntegration.md)

- [News on Hyperspace System Trust for Cumulus](../../2024/11/trustEngineCumulusIntegration.md)

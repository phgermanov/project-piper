# Piper General Purpose Pipeline for GitHub Actions

[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-%23FE5196?logo=conventionalcommits&logoColor=white)](https://conventionalcommits.org)
[![Log4brains ADRs](https://pages.github.tools.sap/project-piper/piper-pipeline-github/badge.svg)](https://pages.github.tools.sap/project-piper/piper-pipeline-github)

This repository contains [Piper's *general purpose pipeline*](https://go.sap.corp/piper/stages/) for GitHub Actions and is based on [Piper's GitHub Action](https://github.com/SAP/project-piper-action/tree/main).

## Prerequisites

- The repository needs to have [GitHub Actions enabled](https://pages.github.tools.sap/github/features-and-usecases/features/actions/start).
- The repository needs to be registered to the [SUGAR Service](https://pages.github.tools.sap/github/features-and-usecases/features/actions/runners/#solinas-and-sugar-service---sharedenterprise-runners).

### Secrets

The workflow needs the following secrets [to be defined](https://docs.github.com/en/enterprise-server/actions/security-guides/encrypted-secrets) in your repository/organization:

- `PIPER_VAULTAPPROLEID`
- `PIPER_VAULTAPPROLESECRETID`

## Usage

The *general purpose pipeline* is defined in [`.github/workflows/sap-piper-workflow.yml`](.github/workflows/sap-piper-workflow.yml). To use it, create a workflow file in the `.github/workflows` folder of your repository as follows:

```yml
name: Piper workflow

on:
  workflow_dispatch:
jobs:
  piper:
    uses: project-piper/piper-pipeline-github/.github/workflows/sap-piper-workflow.yml@v1
    secrets: inherit
#    with:
#      runs-on: '[ "self-hosted", "my-custom-runner-tag" ]'
```

**Note:** The Piper team strongly recommends referring version tags (such as `v1`, `v1.2`, or `v1.2.3`) instead of the `main` branch. By referencing Piper GPP via version tags, you can also configure the **ospo-renovate** bot to automatically maintain and update Piper GPP versions. For more information on how to enable ospo-renovate, please [check this page](https://github.tools.sap/OSPO/renovate-controller?tab=readme-ov-file#renovate-controller).

## Limitations

- deviation of confirm stage compared to Jenkins and Azure DevOps in form of usage of GitHub environments (`Piper Promote`) for manual confirmation via [environment protection rules](https://docs.github.com/en/enterprise-server/actions/deployment/targeting-different-environments/using-environments-for-deployment#required-reviewers). :warning: **The Piper config parameter `manualConfirmation` is not considered at all!**

## Extensibility

It's possible to extend existing stages with pre and post steps, or the general purpose workflow as a whole.
You can find the documentation on it [here](docs/extensibility.md).

## Individual General Purpose Stages

The individual workflows for the stages are located in the [`.github/workflows`](.github/workflows) folder. Using these enables you to build a customised pipeline workflow as shown in [this example](https://github.tools.sap/project-piper/piper-pipeline-github/blob/main/examples/custom_workflow/custom_pipeline.yml).

:warning: The stages have (optional and mandatory) inputs that are listed and described inside the specific workflow file.

## Using Custom Piper Versions with GPP

Developers who want to use custom Piper versions can specify the `piper-version` or `sap-piper-version` in the YAML configuration. The format for these versions is as follows:

- For `piper-version`: `devel:SAP:jenkins-library:customSHA`
- For `sap-piper-version`: `devel:ContinuousDelivery:piper-library:customSHA`

Additionally, using custom Piper versions requires setting the WDF GitHub token in the testing repository's settings, as well as the Enterprise Server URL.
To do this, open your repository on GitHub, navigate to `Settings > Secrets and variables`, and add both as separate secrets by clicking `New repository secret`.
Set `PIPER_ENTERPRISE_SERVER_URL` to `https://github.wdf.sap.corp`, and `PIPER_WDF_GITHUB_TOKEN` to your access token.

### Example configuration

```yml
name: Piper workflow

on:
  workflow_dispatch:

jobs:
  piper:
    # if you want to pin specific version use for example @v1.2.3 instead of @main
    uses: project-piper/piper-pipeline-github/.github/workflows/sap-piper-workflow.yml@main
    secrets: inherit
    with:
      # will compile piper with the given SHA
      piper-version: 'devel:SAP:jenkins-library:eef61447b9ff4aafe5dcd4e0bbf5d482be7e7871'
      sap-piper-version: 'devel:ContinuousDelivery:piper-library:eef61447b9ff4aafe5dcd4e0bbf5d482be7e7871'
```

# Piper Azure DevOps Task

[![Quality Gate Status](https://sonar.tools.sap/api/project_badges/measure?project=piper-azure-task&metric=alert_status&token=sqb_01f331374dedca51cde3d23917ad14c405f1f881)](https://sonar.tools.sap/dashboard?id=piper-azure-task)
[![Azure Marketplace release](https://img.shields.io/badge/Azure_Marketplace-RELEASE-green)](https://marketplace.visualstudio.com/items?itemName=ProjectPiper.piper-azure-task-dev)
[![Azure Marketplace test](https://img.shields.io/badge/Azure_Marketplace-TEST-green)](https://marketplace.visualstudio.com/items?itemName=ProjectPiperDev.piper-azure-task-dev)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-%23FE5196?logo=conventionalcommits&logoColor=white)](https://conventionalcommits.org)
[![Log4brains ADRs](https://pages.github.tools.sap/project-piper/piper-azure-task/badge.svg)](https://pages.github.tools.sap/project-piper/piper-azure-task)

This Azure DevOps task allows running [Piper](https://go.sap.corp/piper) on Azure DevOps. It is used in [Piper's *general purpose pipeline* for GitHub Actions](https://github.tools.sap/project-piper/piper-pipeline-azure).

## Prerequisites

The task is [published](https://marketplace.visualstudio.com/items?itemName=ProjectPiper.piper-azure-task-dev) on the marketplace and needs to be [installed](https://docs.microsoft.com/en-us/azure/devops/marketplace/install-extension?view=azure-devops&tabs=browser) into your Azure organization, but is currently not publicly available. Feel free to contact one of the code owners to get install permissions in case you would like to use it in your own organisation.

### Secrets

The extension requires Vault credentials and a [service connection](https://learn.microsoft.com/en-us/azure/devops/pipelines/library/service-endpoints?view=azure-devops&tabs=yaml) to GitHub Tools available.
The Vault credentials should be provided to the pipeline as `hyperspace.vault.roleId` and `hyperspace.vault.secretId` [variables](https://learn.microsoft.com/en-gb/azure/devops/pipelines/process/variables?view=azure-devops&tabs=yaml%2Cbatch).

## Usage

[[YML_SNIPPET]]

### Arguments

[[ARGUMENTS_TABLE]]

If you plan to use this Piper Task in your custom pipelines, and your steps use **commonPipelineEnv**, you need to ensure that you set `pipelineEnvironment_b64` in the stage-level variables. For example:

```yaml
stages:
  variables:
    pipelineEnvironment_b64: 'thereMustBeSomeBase64EncodedJSONString=='
```

Internally, the value of `pipelineEnvironment_b64` is stored in the `.pipeline/commonPipelineEnvironment` file as JSON. If one of your steps modifies this file, please ensure that you set the `exportPipelineEnv` task argument to `true` so that your subsequent stages can access the modified values. You may refer to Piper's [general purpose pipeline](https://github.tools.sap/project-piper/piper-pipeline-azure) for example usage.

## Development

See [docs/development.md](docs/development.md) for development guidelines and [docs/release-drafter.md](docs/release-drafter.md) for release process documentation.

## Known Issues

See [open issues](https://github.tools.sap/project-piper/piper-github-action/issues).

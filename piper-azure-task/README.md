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

```yml
# Project Piper
# Execute a step from the SAP Piper library
- task: piper@1
  # or if you would like to pin specific version of the task, use 'piper@1.2.3'
  inputs:
    stepName: 'help'
    flags: '' # Optional
    piperVersion: '' # Optional
    sapPiperVersion: '' # Optional
    customConfigLocation: '.pipeline/config.yml' # Optional
    customDefaults: '' # Optional
    restorePipelineDefaults: '' # Optional
    restorePipelineStageConditions: '' # Optional
    exportPipelineEnv: false # Optional
    preserveDefaultConfig: false # Optional
    preserveStageConditions: false # Optional
    createCheckIfStepActiveMaps: false # Optional
    dockerImage: '' # Optional
    dockerOptions: '' # Optional
    dockerEnvVars: '' # Optional
    sidecarImage: '' # Optional
    sidecarOptions: '' # Optional
    sidecarEnvVars: '' # Optional
    gitHubConnection: '' # Optional
    gitHubComConnection: '' # Optional
    dockerRegistryConnection: '' # Optional
    getSapPiperLatestVersion: false # Optional
    fetchPiperBinaryVersionTag: false # Optional
```

### Arguments

| Argument | Description |
| -------- | ----------- |
| `stepName` </br> Step Name | (Required) The name of a Piper step to execute. </br> Default value: `help` |
| `flags` </br> command options | (Optional) Option flags for Piper step |
| `piperVersion` </br> Piper Version | (Optional) Specify the exact version of the OS Piper binary to use, overriding the default version. 'latest' is not allowed; only exact versions are permitted. |
| `sapPiperVersion` </br> SAP Piper Version | (Optional) Specify the exact version of the SAP Piper binary to use, overriding the default version. 'latest' is not allowed; only exact versions are permitted. |
| `customConfigLocation` </br> Custom Config | (Optional) Path to configuration file for Piper. If not present, path `.pipeline/config.yml will be used` </br> Default value: `.pipeline/config.yml` |
| `customDefaults` </br> Custom defaults | (Optional) Paths to custom default configurations for Piper. Can also take multiple files by passing a multi-line string (using literal style with a '\|' ), in which the files are in ascending order of importance. |
| `restorePipelineDefaults` </br> Pipeline default config | (Optional) Imports the pipeline default config from an given base-64 encoded YAML string. |
| `restorePipelineStageConditions` </br> Pipeline stage conditions | (Optional) DEPRECATED. Refer to <https://github.tools.sap/project-piper/piper-pipeline-azure/pull/304> and input createCheckIfStepActiveMaps instead. |
| `exportPipelineEnv` </br> Export Pipeline Environment | (Optional) Exports the pipeline environment to an output variable for later use. </br> Default value: `false` |
| `preserveDefaultConfig` </br> Preserve Default Config | (Optional) Exports the default configuration to an output variable for later use. </br> Default value: `false` |
| `preserveStageConditions` </br> Preserve Stage Conditions | (Optional) DEPRECATED. Refer to <https://github.tools.sap/project-piper/piper-pipeline-azure/pull/304> and input createCheckIfStepActiveMaps instead. </br> Default value: `false` |
| `createCheckIfStepActiveMaps` </br> Create checkIfStepActive maps | (Optional) Create active steps and stages maps in the .pipeline folder </br> Default value: `false` |
| `dockerImage` </br> Docker Image | (Optional) The Docker image to run the step in, use 'none' to run without docker. |
| `dockerOptions` </br> Docker Options | (Optional) The options to pass to the docker run command. |
| `dockerEnvVars` </br> Docker Environment Variables | (Optional) The environment variables to pass to the docker run command, as JSON string (e.g. '{"testKey": "testValue"}'). |
| `sidecarImage` </br> Sidecar Docker Image | (Optional) The Docker image to run as a sidecar container |
| `sidecarOptions` </br> Sidecar Docker Options | (Optional) The Docker options for the sidecar container |
| `sidecarEnvVars` </br> Sidecar Docker Environment Variables | (Optional) The environment variables to pass to the docker run command for the sidecar container. |
| `gitHubConnection` </br> GitHub Enterprise connection (OAuth or PAT) | (Optional) Specify the name of the GitHub service connection to use to connect to the GitHubEnterprise repository. The connection must be based on a GitHub user's GitHub personal access token. |
| `gitHubComConnection` </br> GitHub connection (OAuth or PAT) | (Optional) DEPRECATED. Setting this parameter doesn't affect anything. Will be removed in future releases. |
| `dockerRegistryConnection` </br> Docker registry connection (PAT) | (Optional) Specify the name of the Docker registry service connection to use to pull an image that is used to run Piper with. The connection must be have a username and PAT as password. |
| `getSapPiperLatestVersion` </br> Get SAP Piper Latest Version | (Optional) DEPRECATED. Setting this parameter doesn't affect anything. Will be removed in future releases. </br> Default value: `false` |
| `fetchPiperBinaryVersionTag` </br> Fetch version tag of Piper binary | (Optional) Fetch version tag of Piper binary from releases and output variable for later use, such as for the Cache task </br> Default value: `false` |

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

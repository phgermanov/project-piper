# Piper General Purpose Pipeline for Azure DevOps

[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-%23FE5196?logo=conventionalcommits&logoColor=white)](https://conventionalcommits.org)

This repository contains [Piper's *general purpose pipeline*](https://go.sap.corp/piper/stages/) for Azure DevOps and is based on [Piper's Azure Task](https://github.tools.sap/project-piper/piper-azure-task).

## Prerequisites

The pipeline requires [Piper's Azure Task](https://github.tools.sap/project-piper/piper-azure-task) to be [installed](https://github.tools.sap/project-piper/piper-azure-task/blob/main/README.md#prerequisites) in the projects Azure organization

## Usage

The *general purpose pipeline* is defined in [`sap-piper-pipeline.yml`](./sap-piper-pipeline.yml). To use it, create an `azure-pipelines.yml` in the root folder of your repository as follows:

```yaml
# Using Piper general purpose pipeline for Azure

trigger:
- main

resources:
  repositories:
    - repository: piper-templates
      endpoint: <name-of-gh-endpoint>
      type: githubenterprise
      name: project-piper/piper-pipeline-azure
      # By default, Azure will load GPP from 'refs/heads/main'
      # but if you want the specific version of GPP then uncomment the line below:
      # ref: refs/tags/<release_tag>
      # and replace <release_tag> to one from here: https://github.tools.sap/project-piper/piper-pipeline-azure/releases

extends:
  template: sap-piper-pipeline.yml@piper-templates
```

❗The endpoint name needs to be looked up in you projects service connection settings on [dev.azure.com](https://dev.azure.com/).

## Playground on Azure DevOps

<https://dev.azure.com/hyperspace-pipelines/project-piper>

## Example projects

- <https://github.tools.sap/project-piper/piper-ado-minimal-docker>

## Microsoft Azure DevOps documentation

[Templates](https://docs.microsoft.com/en-us/azure/devops/pipelines/process/templates?view=azure-devops)

[Tasks](https://docs.microsoft.com/en-us/azure/devops/pipelines/process/tasks?view=azure-devops&tabs=yaml)

[Task groups](https://docs.microsoft.com/en-us/azure/devops/pipelines/library/task-groups?view=azure-devops)

[YAML Schema](https://docs.microsoft.com/en-us/azure/devops/pipelines/yaml-schema?view=azure-devops&tabs=schema%2Cparameter-schema)

[Stages](https://docs.microsoft.com/en-us/azure/devops/pipelines/process/stages?view=azure-devops&tabs=yaml)

[Dependencies, e.g. for variables](https://docs.microsoft.com/en-us/azure/devops/pipelines/process/expressions?view=azure-devops#stage-to-stage-dependencies)

[Conditions](https://docs.microsoft.com/en-us/azure/devops/pipelines/process/conditions?view=azure-devops&tabs=yaml)

[Pre-defined variables](https://docs.microsoft.com/en-us/azure/devops/pipelines/build/variables?view=azure-devops&tabs=yaml)

## Additional Documentation

[Passing variables between jobs](https://gaunacode.com/passing-variables-between-jobs-for-azure-devops-pipelines)

## SAP ADO documentation

[Service documentation](https://pages.github.tools.sap/azure-pipelines/)

[Service repository](https://github.tools.sap/azure-pipelines/templates)

## Further tips & tricks

### Using variables

Variables can be very helpful for example when using in conditions.

Here an example how variables can be used in such a way:

```yml
    - bash: |
        ./piper getConfig --stageConfig > stage-config.json
        echo "##vso[task.setvariable variable=productiveBranch]$(jq -j .productiveBranch stage-config.json)"
        echo "##vso[task.setvariable variable=lockPipelineRun]$(jq -j .lockPipelineRun stage-config.json)"
      name: promoteconfig
    - bash: |
        echo "Writing lock-run.json for productive branch $(productiveBranch)"
        touch lock-run.json
      condition: and(succeeded(), eq(variables.productiveBranch, variables['Build.SourceBranchName']), eq(variables.lockPipelineRun, 'true'))
```

### Different types of variables

Variables can have a job scope or also be "exported" and then be used outside of the job.

**Important note:** Two variables with the same name and different scope appear as two different variables and need to be handled individually!

```yml
    - bash: |
        ./piper getConfig --stageConfig > stage-config.json
        echo "##vso[task.setvariable variable=productiveBranch]$(jq -j .productiveBranch stage-config.json)"
        echo "##vso[task.setvariable variable=productiveBranch;isOutput=true]$(jq -j .productiveBranch stage-config.json)"
      name: config
    - bash: |
        echo productiveBranchStandardVar: $(productiveBranch)
        echo productiveBranchOutputVar: $(config.productiveBranch)
```

### Using variables in stage conditions

Following way of referencing output variables in stage conditions is possible:

```yml
  condition: eq(dependencies.Init.outputs['setup.optimization.isOptimizedAndScheduled'], 'false')
```

**Format using `stageDependencies` like `stageDependencies.Init.setup.outputs['piper_defaults.StageConditions']` does not work for stage conditions!**

### Passing custom location of Piper config

Piper has a [customConfig](https://pages.github.tools.sap/project-piper/steps/getConfig/) flag, with which an alternative location of the Piper config file (`.pipeline/config.yml`) can be set.
This is currently only possible to do by creating a custom pipeline, and passing the [`customConfigLocation` parameter](https://github.tools.sap/project-piper/piper-azure-task#arguments) to the Piper Azure Task in the custom pipeline template.

## Using a development version of Piper

### Using a branch of open source Piper

You can specify a branch name for development purposes. This only works for open source Piper. It can be done by changing the `hyperspace.piper.version` variable in `sap-piper-pipeline.yml` to the name of the branch, and adding an additional variable:

```yaml
- name: hyperspace.piper.isBranch
  value: true
```

You may also need to add to ensure that [the correct version of Go](https://github.com/SAP/jenkins-library/blob/master/go.mod#L3) is installed.

```yaml
- task: GoTool@0
  inputs:
    version: '1.17'
```

if you don't keep the version 1.17 you may see the error : ``info  golangBuild - pkg/piperutils/FileUtils.go:11:2: package io/fs is not in GOROOT (/opt/hostedtoolcache/go/1.15.15/x64/src/io/fs)``

In case you are caching the binary with the below steps only add the above task to ``Init.yaml`` and if you are not caching then add it to every ``stage`` after the ``checkout`` step

#### Preventing Piper being built during every stage (binary caching)

When using a Piper binary version, the Azure task has to retrieve it again in every stage since it is not cached. A development version has to be built after fetching it, and this will add up during a pipeline run. To make it more efficient, you can add a [pipeline caching](https://docs.microsoft.com/en-us/azure/devops/pipelines/release/caching?view=azure-devops) step in every stage's job, e.g.:

```yaml
steps:
  - checkout: self
    submodules: ${{ parameters.checkoutSubmodule }}
  - task: Cache@2
    inputs:
      key: piper-branch-$(hyperspace.piper.branchCacheVersion)
      path: $(Pipeline.Workspace)/jenkins-library/binaries
    displayName: Piper binary cache
```

Note that there is a reference to `hyperspace.piper.branchCacheVersion`. You can create add variable in `sap-piper-pipeline.yml` as :

```yaml
- name: hyperspace.piper.branchCacheVersion
  value: v1
```

the value `v1` is any string you wish to keep . Since there is native caching of the binary, when you push a change to your branch the caching will prevent a rebuild of the code and your old cached binary will take effect. In such case just change the version of the value to e.g `v2` which will then prompt the azure task to rebuild the binary for you.

### Using a self-built inner source Piper binary

The Piper Azure task retrieves the inner source binary from the releases of the [sap-piper repository](https://github.tools.sap/project-piper/sap-piper/).
To test your changes to inner source Piper, you can build it yourself, and upload it as asset to a release in your own repository (for example a fork of the aforementioned repository) on `github.tools.sap`.
The latest release will always be used.
Usually Linux is run on the Azure DevOps agents, so you could build the binary with for example
`GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build`.

You can then point to your repository by changing the following variables in a custom pipeline:

- `hyperspace.sappiper.repository`
- `hyperspace.sappiper.owner`

Note that binary must be named `sap-piper` in the release assets.

### Forcing a specific Piper arch

The Azure Task [can't detect yet](https://github.tools.sap/project-piper/piper-azure-task/pull/178) if it's running on Apple silicon, so there is the possibility to force the usage of a specific arch build of the Piper binaries with the `hyperspace.piper.enforcedOSArch` variable.

## Help and support

Piper team has decided to close this repository for user issues as the team monitors multiple repositories.
If you would like to request an enhancement please use [piper-library](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues)
If you would like to raise an incident, please use [Service now](https://itsm.services.sap/sp?id=sc_cat_item&sys_id=703f22d51b3b441020c8fddacd4bcbe2&sysparm_category=e15706fc0a0a0aa7007fc21e1ab70c2f&catalog_id=e0d08b13c3330100c8b837659bba8fb4) with offering ISV-DL-CICD-HYPERSPACE.

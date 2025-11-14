# setupPipelineEnvironment

## Description

This step defines the [`globalPipelineEnvironment`](../objects/globalPipelineEnvironment.md) which will be used throughout the complete pipeline.

It will read a configuration file with properties (default location: `.pipeline/config.yml`). The property values are used as default values for many pipeline steps as you can see in the descriptions of the pipeline steps.

In addition other environment settings are initialized like:

* git / GitHub information like (git url, repository, ...)
* initializing data Maps used for staring information which should be written to Influ using step [influxWriteData](influxWriteData.md)
* optional: retrieving GitHub statistics about current commit

!!! tip
    This step needs to run at the beginning of a pipeline right after the SCM checkout.
    Then the subsequent pipeline steps can just consume the information from the `globalPipelineEnvironment` and the information does not need to be passed to the pipeline steps again and again.

## Prerequisites for pipeline configuration

Configuration is available via [config.yml file](../configuration.md).

## Example

Usage of pipeline step:

```groovy
setupPipelineEnvironment script: this
```

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|buildDiscarder|no|`[daysToKeep:-1, numToKeep:10, artifactDaysToKeep:-1, artifactNumToKeep:-1]`||
|configYmlFile|no|`.pipeline/config.yml`||
|gitBranch|no|||
|gitCommitId|no|||
|gitHttpsUrl|no|`https://github.wdf.sap.corp/${githubOrg}/${githubRepo}.git`||
|gitSshUrl|no|`git@github.wdf.sap.corp:${githubOrg}/${githubRepo}.git`||
|githubApiUrl|no|`https://github.wdf.sap.corp/api/v3`||
|githubOrg|yes|||
|githubRepo|yes|||
|nightlySchedule|no|`'H(0-59) H(18-23) * * *'`||
|relatedLibraries|no|<ul><li>`piper-lib-os`</li></ul>||
|runNightly|no|`false`||
|storeGithubStatistics|no|`false`||

### Details

* `script` defines the global script environment of the Jenkinsfile run. `this` needs to be passed in order for the initialization to take effect.
* By using the parameter `configFile` you can point to a custom property file somewhere within your source code repository. This is then read and the properties are stored in the `globalPipelineEnvironment`.
* `storeGithubStatistics` defines if you want to read the GitHub statistics for the current commit and store it in the `globalPipelineEnvironment`.
* All git / GitHub related properties allow you to overwrite the default behavior of identifying e.g. GitHub organization, GitHub repository. Default behavior is that a job structure is expected as you get it from the [GitHub Organization Folder Plugin](https://wiki.jenkins-ci.org/display/JENKINS/GitHub+Organization+Folder+Plugin)
* Scheduling pipeline runs can be accomplished by setting `runNightly: true`. To individualize the schedule being applied please check parameter `nightlySchedule` and related [Jenkins documentation on scheduling syntax](https://jenkins.io/doc/book/pipeline/syntax/#cron-syntax). Please also have a look at Piper scheduled mode [in this page](../stages/optimizations.md).

!!! tip
    If your job is not inside a __GitHub Organization Folder__ you can customize all git / GitHub related information using dedicated parameters. This is for example useful if you have a repository within Git and not within GitHub.

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|buildDiscarder||X|X|
|configFile||X|X|
|configYmlFile||X|X|
|gitBranch||X|X|
|gitCommitId||X|X|
|gitHttpsUrl||X|X|
|gitSshUrl||X|X|
|githubApiUrl|X|X|X|
|runNightly||X|X|
|githubRepo||X|X|
|nightlySchedule||X|X|
|relatedLibraries||X|X|
|githubRepo||X|X|
|storeGithubStatistics||X|X|

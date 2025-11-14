# triggerXMakeRemoteJob (DEPRECATED)

## Description

This step triggers a job in the xMake environment.
For xMake details please see the documentation of [Central xmake Build Service](https://github.wdf.sap.corp/pages/xmake-ci/User-Guide/).

## Prerequisites

The central xMake build needs to be configured and activated as [described here](../build/xMake.md#build-service).

The [Parameterized Remote Trigger Plugin](https://github.wdf.sap.corp/sap-jenkins-plugins/Parameterized-Remote-Trigger-plugin) (SAP version) needs to be available in your Jenkins instance.

 You need to create specific Jenkins Credentials with a pair of ***user-id*** and ***Jenkins API token*** on your Jenkins instance to successfully trigger your job.
 Configuration details can be found in the description of the [xMake build service](../build/xMake.md#parameterized-remote-trigger-plugin).

## Example

Example usage of pipeline step:

```groovy
triggerXMakeRemoteJob (
    script: script,
    xMakeJobName: xMakeJobName,
    xMakeJobParameters: 'MODE=promote\nTREEISH=' + gitCommitId + '\nSTAGING_REPO_ID=' + xMakeStagingRepoId
)
```

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|xMakeDevCredentialsId|no|`xmakeDev`||
|xMakeJobName|no|||
|xMakeJobParameters|no|||
|xMakeNovaCredentialsId|no|`xmakeNova`||
|xMakeServer|no|`xmake-dev`||

!!! tip "Find xmake jobs"
    You can find your xmake jobs using [Job Finder](https://xmake-nova.wdf.sap.corp/job_finder).

!!! tip "Required job parameters"
    You can find the [parameters required for the xmake StagePromote job here](https://wiki.wdf.sap.corp/wiki/pages/viewpage.action?pageId=1855680260#OnDemandStage&PromoteBuild(ODSP)-OnDemandStageandPromote(SP)).

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|xMakeDevCredentialsId|X|X|X|
|xMakeJobName|X|X|X|
|xMakeJobParameters|X|X|X|
|xMakeNovaCredentialsId|X|X|X|
|xMakeServer|X|X|X|

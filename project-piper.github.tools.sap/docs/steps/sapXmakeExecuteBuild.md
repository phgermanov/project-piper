# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

- The central xMake build needs to be configured and activated as [described here](../build/xMake.md#build-service).

- You need to create specific Jenkins Credentials with a pair of **userID** and **Jenkins API token** on your Jenkins instance to successfully trigger your job.
  Configuration details can be found in the description of the [xMake build service](../build/xMake.md#parameterized-remote-trigger-plugin).

!!! tip "Required job parameters"
    You can find the [parameters required for the xmake StagePromote job here](https://wiki.wdf.sap.corp/wiki/pages/viewpage.action?pageId=1855680260#OnDemandStage&PromoteBuild(ODSP)-OnDemandStageandPromote(SP)).

## ${docGenParameters}

## ${docGenConfiguration}

## Migrating from triggerXmakeRemoteJob

When migrating from the `triggerXmakeRemoteJob` step, you need to mind the following:

- the parameters `xMakeServer` and `xMakeJobNameTemplate` for the `executeBuild` step are no longer supported
- the parameters `xMakeServer` and `xMakeJobName` for the `triggerXmakeRemoteJob` step are no longer supported

The step only supports the xmake *stage-promote* pattern!

It is no longer possible to define a custom *job name* or a custom *job name pattern* and with it trigger other jobs than a *stage-promote` job.

You now can choose between two patterns for `jobNamePattern`:

- `GitHub-Internal` (default) -> `<owner>-<repository>-SP-<quality>-common[_<shipmentType>]`
- `GitHub-Tools` -> `ght-<owner>-<repository>-SP-<quality>-common[_<shipmentType>]`

### Error: Could not invoke job "": 401

When you see an error like this, you probably have a xmake job that runs on `xmake-dev` an have no credentials defined in your config.

```text
[sapXmakeExecuteBuild] Step execution failed (category: undefined). Error: Failed to trigger job '<job name>': Could not invoke job "": 401
```

In this case the default credentialId is used which is `xmakeNova`. Make sure to define a credential for xmakeDev in your config to fix this behavior:

```yml
general:
  xMakeCredentialsId: 'xmakeDev'
```

## Access Your Build errors

If the xmake build failed, you can find in the file *build-results.json* links pointing to the error line in your xmake build logs.

For more details, see in [xmake user-guide](https://github.wdf.sap.corp/pages/xmake-ci/User-Guide/Setting_up_a_Build/AccessBuildResultsLogsWS/#description-of-build-resultsjson-file)

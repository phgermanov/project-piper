# Deprecated Steps

In the course of the migration to [`piper-lib-os`](https://github.com/SAP/jenkins-library/) the following steps were deprecated:

## deployToCloudFoundry

Please use [cloudFoundryDeploy](../steps/cloudFoundryDeploy.md) instead.

## deployToKubernetes

Please use [kubernetesDeploy](../steps/kubernetesDeploy.md) instead.

## executeBatsTests

Please use [batsExecuteTests](../steps/batsExecuteTests.md) instead.

## executeCheckmarxScan

Please use [checkmarxExecuteScan](../steps/checkmarxExecuteScan.md) instead.

## executeDocker

Please use [dockerExecute](../steps/dockerExecute.md) instead.

## executeDockerOnKubernetes

Please use [dockerExecuteOnKubernetes](../steps/dockerExecuteOnKubernetes.md) instead.

## executeFortifyScan

Please use [fortifyExecuteScan](../steps/fortifyExecuteScan.md) instead.

## executeFortifyAuditStatusCheck

Please use [fortifyExecuteScan](../steps/fortifyExecuteScan.md) instead.

## executeHealthCheck

Please use [healthExecuteCheck](../steps/healthExecuteCheck.md) instead.

## executeHubDetectScan

Please use [detectExecuteScan](../steps/detectExecuteScan.md) instead.

## executeNewmanTests

Please use [newmanExecute](../steps/newmanExecute.md) instead.

## executeNspScan

Use native NPM functionality.

## executeOnePageAcceptanceTests

Please use [karmaExecuteTests](../steps/karmaExecuteTests.md) instead.

## executeOpenSourceDependencyScan

Please use specific scan steps instead.

## executePerformanceJMeterTests

Please use [sapSUPAExecuteTests](../steps/sapSUPAExecuteTests.md) instead.

## executePerformanceSingleUserTests

Please use [sapSUPAExecuteTests](../steps/sapSUPAExecuteTests.md) instead.

## executePerformanceUnitTests

Please use [sapSUPAExecuteTests](../steps/sapSUPAExecuteTests.md) instead.

## executeProtecodeScan

Please use [protecodeExecuteScan](../steps/protecodeExecuteScan.md) instead.

## executePPMSComplianceCheck

Please use [sapCheckPPMSCompliance](../steps/sapCheckPPMSCompliance.md) instead.

## executePPMSWhitesourceComplianceCheck

Please use [sapCheckPPMSCompliance](../steps/sapCheckPPMSCompliance.md) instead.

## executeSonarScan

Please use [sonarExecuteScan](../steps/sonarExecuteScan.md) instead.

## executeSourceclearScan

Please use [whitesourceExecuteScan](../steps/whitesourceExecuteScan.md) instead.

## executeVulasScan

Vulas is deprecated, use other scanning tools.

## executeWhitesourceScan

Please use [whitesourceExecuteScan](../steps/whitesourceExecuteScan.md) instead.

## handleStepErrors

Please use [handlePipelineStepErrors](../steps/handlePipelineStepErrors.md) instead.

## karmaExecuteTests

Please use [npmExecuteTests](../steps/npmExecuteTests.md) instead.

## measureDuration

Please use [durationMeasure](../steps/durationMeasure.md) instead.

## npmExecute

Please use [npmExecuteScript](../steps/npmExecuteScripts.md) instead.

## npmExecuteEndToEndTests

Please use [npmExecuteTests](../steps/npmExecuteTests.md) instead.

## publishCheckResults

Please use [checksPublishResults](../steps/checksPublishResults.md) instead.

## publishGithubRelease

Please use [githubPublishRelease](../steps/githubPublishRelease.md) instead.

## publishTestResults

Please use [testsPublishResults](../steps/testsPublishResults.md) instead.

## pushToDockerRegistry

Please use [containerPushToRegistry](../steps/containerPushToRegistry.md) instead.

## restartableSteps

Please use [pipelineRestartSteps](../steps/pipelineRestartSteps.md) instead.

## sapPiperPublishNotifications

Please use [piperPublishWarnings](../steps/piperPublishWarnings.md) instead.

## sendNotificationMail

Please use [mailSendNotification](../steps/mailSendNotification.md) instead.

## sendNotificationSlack

Please use [slackSendNotification](../steps/slackSendNotification.md) instead.

## setVersion

Please use [artifactPrepareVersion](../steps/artifactPrepareVersion.md) instead.

## stashFiles

Please use [pipelineStashFiles](../steps/pipelineStashFiles.md) instead.

## stashFilesAfterBuild

Please use [pipelineStashFilesAfterBuild](../steps/pipelineStashFilesAfterBuild.md) instead.

## stashFilesBeforeBuild

Please use [pipelineStashFilesBeforeBuild](../steps/pipelineStashFilesBeforeBuild.md) instead.

## triggerXMakeRemoteJob

Please use [sapXmakeExecuteBuild](../steps/sapXmakeExecuteBuild.md) instead.

## uiVeri5ExecuteTests

Please use [npmExecuteTests](../steps/npmExecuteTests.md) instead.

## writeInflux

Please use [influxWriteData](../steps/influxWriteData.md) instead.

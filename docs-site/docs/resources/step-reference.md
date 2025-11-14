# Piper Step Reference

Complete quick reference table of all available Piper steps with brief descriptions.

## Table of Contents

- [ABAP Steps](#abap-steps)
- [API Management Steps](#api-management-steps)
- [Artifact Management Steps](#artifact-management-steps)
- [Build Steps](#build-steps)
- [Cloud Deployment Steps](#cloud-deployment-steps)
- [Container & Kubernetes Steps](#container--kubernetes-steps)
- [Integration & Testing Steps](#integration--testing-steps)
- [Notification & Events](#notification--events)
- [Pipeline Management Steps](#pipeline-management-steps)
- [Security & Compliance Steps](#security--compliance-steps)
- [Source Control Steps](#source-control-steps)
- [Transport Management Steps](#transport-management-steps)
- [Utility Steps](#utility-steps)

---

## ABAP Steps

| Step Name | Description |
|-----------|-------------|
| `abapAddonAssemblyKitCheck` | Checks validity of ABAP Addon Product Modelling via AAKaaS |
| `abapAddonAssemblyKitCheckCVs` | Validates ABAP Software Component Versions |
| `abapAddonAssemblyKitCheckPV` | Validates Addon Product Version |
| `abapAddonAssemblyKitCreateTargetVector` | Creates Target Vector for software lifecycle operations |
| `abapAddonAssemblyKitPublishTargetVector` | Publishes Target Vector according to specified scope |
| `abapAddonAssemblyKitRegisterPackages` | Uploads SAR archives and creates physical Delivery Packages |
| `abapAddonAssemblyKitReleasePackages` | Releases physical Delivery Packages |
| `abapAddonAssemblyKitReserveNextPackages` | Determines ABAP delivery packages for Software Component Versions |
| `abapEnvironmentAssembleConfirm` | Confirms delivery of assembly for installation in SAP BTP ABAP Environment |
| `abapEnvironmentAssemblePackages` | Assembles installation, support package or patch in SAP BTP ABAP Environment |
| `abapEnvironmentBuild` | Executes builds as defined with the build framework |
| `abapEnvironmentCheckoutBranch` | Switches between branches of a git repository on SAP BTP ABAP Environment |
| `abapEnvironmentCloneGitRepo` | Clones a git repository to SAP BTP ABAP Environment system |
| `abapEnvironmentCreateSystem` | Creates a SAP BTP ABAP Environment system (Steampunk) |
| `abapEnvironmentCreateTag` | Creates a tag for a git repository on SAP BTP ABAP Environment |
| `abapEnvironmentPullGitRepo` | Pulls a git repository to SAP BTP ABAP Environment system |
| `abapEnvironmentPushATCSystemConfig` | Creates/Updates ATC System Configuration |
| `abapEnvironmentRunATCCheck` | Runs ATC (ABAP Test Cockpit) checks |
| `abapEnvironmentRunAUnitTest` | Runs ABAP Unit tests |
| `abapLandscapePortalUpdateAddOnProduct` | Updates AddOn product in SAP BTP ABAP Environment Landscape Portal |
| `checkChangeInDevelopment` | Checks if a change is in 'in development' status |
| `gctsCloneRepository` | Clones a Git repository for gCTS |
| `gctsCreateRepository` | Creates a Git repository on an ABAP system |
| `gctsDeploy` | Deploys Git repository to local repository and ABAP system |
| `gctsExecuteABAPQualityChecks` | Runs ABAP unit tests and ATC checks for specified scope |
| `gctsExecuteABAPUnitTests` | Runs ABAP unit tests and ATC checks for specified scope |
| `gctsRollback` | Performs rollback of one or several commits |
| `isChangeInDevelopment` | Checks if a certain change is in 'in development' status |

---

## API Management Steps

| Step Name | Description |
|-----------|-------------|
| `apiKeyValueMapDownload` | Downloads specific Key Value Map from API Portal |
| `apiKeyValueMapUpload` | Creates API key value map artifact in API Portal |
| `apiProviderDownload` | Downloads specific API Provider from API Portal |
| `apiProviderList` | Gets full list of all API providers from API Portal |
| `apiProviderUpload` | Creates API provider artifact in API Portal |
| `apiProxyDownload` | Downloads specific API Proxy from API Portal |
| `apiProxyList` | Gets list of API Proxies from API Portal |
| `apiProxyUpload` | Uploads API proxy artifact to API Portal |

---

## Artifact Management Steps

| Step Name | Description |
|-----------|-------------|
| `artifactPrepareVersion` | Prepares and updates artifact version before building |
| `artifactSetVersion` | Sets the version of an artifact |
| `awsS3Upload` | Uploads files or directories to AWS S3 Bucket |
| `azureBlobUpload` | Uploads files or directories to Azure Blob Storage |
| `nexusUpload` | Uploads artifacts to Nexus Repository Manager |

---

## Build Steps

| Step Name | Description |
|-----------|-------------|
| `buildExecute` | Executes build process for various build tools |
| `cnbBuild` | Executes Cloud Native Buildpacks |
| `golangBuild` | Executes golang build |
| `gradleExecuteBuild` | Runs gradle build with specified parameters |
| `mavenBuild` | Installs maven project into local maven repository |
| `mavenExecute` | Runs maven commands |
| `mavenExecuteIntegration` | Executes backend integration tests via Jacoco |
| `mavenExecuteStaticCodeChecks` | Executes SpotBugs and PMD static code checks |
| `mtaBuild` | Performs MTA (Multi-Target Application) build |
| `npmExecute` | Executes npm commands |
| `npmExecuteLint` | Executes linting scripts on npm packages |
| `npmExecuteScripts` | Handles JavaScript dependencies and npm commands |
| `pythonBuild` | Builds Python project |

---

## Cloud Deployment Steps

| Step Name | Description |
|-----------|-------------|
| `cfManifestSubstituteVariables` | Substitutes variables in Cloud Foundry manifest |
| `cloudFoundryCreateService` | Creates one or multiple services in Cloud Foundry |
| `cloudFoundryCreateServiceKey` | Creates service key in Cloud Foundry |
| `cloudFoundryCreateSpace` | Creates user-defined space in Cloud Foundry |
| `cloudFoundryDeleteService` | Deletes Cloud Foundry service |
| `cloudFoundryDeleteSpace` | Deletes space in Cloud Foundry |
| `cloudFoundryDeploy` | Deploys application to Cloud Foundry |
| `helmExecute` | Executes helm3 as Kubernetes package manager |
| `kubernetesDeploy` | Deploys to Kubernetes test or production namespace |
| `multicloudDeploy` | Deploys application to multiple cloud platforms |
| `neoDeploy` | Deploys application to SAP Cloud Platform Neo |
| `xsDeploy` | Performs XS deployment |

---

## Container & Kubernetes Steps

| Step Name | Description |
|-----------|-------------|
| `containerExecuteStructureTests` | Executes Container Structure Tests |
| `containerPushToRegistry` | Pushes container to registry |
| `containerSaveImage` | Saves container image as tar file |
| `dockerExecute` | Executes commands inside a Docker container |
| `dockerExecuteOnKubernetes` | Executes Docker commands on Kubernetes cluster |
| `hadolintExecute` | Executes Hadolint Dockerfile linter |
| `imagePushToRegistry` | Copies Docker image between container registries |
| `kanikoExecute` | Executes Kaniko build for creating Docker containers |

---

## Integration & Testing Steps

| Step Name | Description |
|-----------|-------------|
| `batsExecuteTests` | Executes tests using Bash Automated Testing System (bats-core) |
| `dubExecute` | Executes dub commands for D language projects |
| `gatlingExecuteTests` | Executes Gatling performance tests |
| `gaugeExecuteTests` | Installs gauge and executes specified tests |
| `healthExecuteCheck` | Executes health checks on deployed applications |
| `integrationArtifactDeploy` | Deploys CPI integration flow |
| `integrationArtifactDownload` | Downloads integration flow runtime artifact |
| `integrationArtifactGetMplStatus` | Gets MPL status of integration flow |
| `integrationArtifactGetServiceEndpoint` | Gets deployed CPI integration flow service endpoint |
| `integrationArtifactResource` | Adds, deletes, or updates integration flow resource file |
| `integrationArtifactTransport` | Transports Integration Package using SAP Content Agent Service |
| `integrationArtifactTriggerIntegrationTest` | Tests service endpoint of iFlow |
| `integrationArtifactUnDeploy` | Undeploys integration flow |
| `integrationArtifactUpdateConfiguration` | Updates integration flow configuration parameters |
| `integrationArtifactUpload` | Uploads or updates integration flow designtime artifact |
| `karmaExecuteTests` | Executes Karma test runner |
| `newmanExecute` | Installs newman and executes specified collections |
| `npmExecuteEndToEndTests` | Executes end-to-end tests using npm |
| `npmExecuteTests` | Executes tests using npm |
| `seleniumExecuteTests` | Executes Selenium tests |
| `testsPublishResults` | Publishes test results |
| `uiVeri5ExecuteTests` | Executes UI5 e2e tests using uiVeri5 |

---

## Notification & Events

| Step Name | Description |
|-----------|-------------|
| `ansSendEvent` | Sends event to SAP Alert Notification Service |
| `gcpPublishEvent` | Publishes event to GCP using OIDC authentication (beta) |
| `mailSendNotification` | Sends email notifications |
| `slackSendNotification` | Sends Slack notifications |

---

## Pipeline Management Steps

| Step Name | Description |
|-----------|-------------|
| `checksPublishResults` | Publishes check results |
| `commonPipelineEnvironment` | Manages common pipeline environment variables |
| `debugReportArchive` | Archives debug reports |
| `durationMeasure` | Measures and records duration of pipeline steps |
| `handlePipelineStepErrors` | Handles errors in pipeline steps |
| `influxWriteData` | Writes metrics to InfluxDB |
| `jenkinsMaterializeLog` | Materializes Jenkins logs |
| `pipelineCreateScanSummary` | Collects scan results and creates summary report |
| `pipelineExecute` | Executes pipeline based on configuration |
| `pipelineRestartSteps` | Restarts specific pipeline steps |
| `pipelineStashFiles` | Stashes files for use in other pipeline stages |
| `pipelineStashFilesAfterBuild` | Stashes files after build stage |
| `pipelineStashFilesBeforeBuild` | Stashes files before build stage |
| `piperLoadGlobalExtensions` | Loads global Piper extensions |
| `piperPublishWarnings` | Publishes warnings from Piper steps |
| `prepareDefaultValues` | Prepares default configuration values |
| `setupCommonPipelineEnvironment` | Sets up common pipeline environment |

---

## Security & Compliance Steps

| Step Name | Description |
|-----------|-------------|
| `ascAppUpload` | Uploads app to ASC (Application Security Center) |
| `checkmarxExecuteScan` | Executes Checkmarx security scan |
| `checkmarxOneExecuteScan` | Executes CheckmarxOne security scan |
| `codeqlExecuteScan` | Executes CodeQL scan for static code analysis |
| `contrastExecuteScan` | Evaluates Contrast Assess audit requirements |
| `credentialdiggerScan` | Scans GitHub repository with Credential Digger |
| `detectExecuteScan` | Executes BlackDuck Detect scan |
| `fortifyExecuteScan` | Executes Fortify static code analysis scan |
| `malwareExecuteScan` | Performs malware scan using SAP Malware Scanning Service |
| `protecodeExecuteScan` | Executes Black Duck Binary Analysis (BDBA/Protecode) scan |
| `snykExecute` | Executes Snyk security scan |
| `sonarExecuteScan` | Executes SonarQube scanner |
| `whitesourceExecuteScan` | Executes Mend (WhiteSource) scan |

---

## Source Control Steps

| Step Name | Description |
|-----------|-------------|
| `githubCheckBranchProtection` | Checks branch protection of GitHub branch |
| `githubCommentIssue` | Comments on GitHub issues and pull requests |
| `githubCreateIssue` | Creates new GitHub issue |
| `githubCreatePullRequest` | Creates pull request on GitHub |
| `githubPublishRelease` | Publishes release in GitHub |
| `githubSetCommitStatus` | Sets status of certain commit |
| `gitopsUpdateDeployment` | Updates Kubernetes Deployment Manifest in Infrastructure Git Repository |

---

## Transport Management Steps

| Step Name | Description |
|-----------|-------------|
| `tmsExport` | Exports MTA file to TMS landscape |
| `tmsUpload` | Uploads MTA file to TMS landscape |
| `transportRequestCreate` | Creates transport request |
| `transportRequestDocIDFromGit` | Retrieves change document ID from Git repository |
| `transportRequestRelease` | Releases transport request |
| `transportRequestReqIDFromGit` | Retrieves transport request ID from Git repository |
| `transportRequestUploadCTS` | Uploads UI5 application to SAPUI5 ABAP repository |
| `transportRequestUploadFile` | Uploads file to transport request |
| `transportRequestUploadRFC` | Uploads UI5 application via RFC connections |
| `transportRequestUploadSOLMAN` | Uploads file to transport via Solution Manager |

---

## Utility Steps

| Step Name | Description |
|-----------|-------------|
| `jsonApplyPatch` | Patches JSON with patch file |
| `shellExecute` | Executes defined shell script |
| `spinnakerTriggerPipeline` | Triggers Spinnaker pipeline |
| `terraformExecute` | Executes Terraform |
| `vaultRotateSecretId` | Rotates Vault AppRole Secret ID |
| `writeTemporaryCredentials` | Writes temporary credentials |

---

## Additional Information

- **Total Steps**: 153+
- **Documentation**: [https://sap.github.io/jenkins-library/](https://sap.github.io/jenkins-library/)
- **GitHub Repository**: [https://github.com/SAP/jenkins-library](https://github.com/SAP/jenkins-library)

## Usage Example

```groovy
@Library('piper-lib-os') _

pipeline {
    agent any
    stages {
        stage('Build') {
            steps {
                mavenBuild script: this
            }
        }
        stage('Security Scan') {
            steps {
                sonarExecuteScan script: this
            }
        }
        stage('Deploy') {
            steps {
                cloudFoundryDeploy script: this
            }
        }
    }
}
```

## Finding More Information

Each step has detailed documentation including:
- Prerequisites
- Parameters and configuration options
- Examples
- Pipeline configuration
- Best practices

Visit the individual step documentation at:
`https://sap.github.io/jenkins-library/steps/<step-name>/`

---

*Last Updated: 2025-11-14*

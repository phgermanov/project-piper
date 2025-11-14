# FAQs

## Platform support

**Q:** What platforms are supported?

**A:** Currently Piper is only built for Linux(amd64) and MacOS(amd64 and arm64)

## Debug logging in Piper Github Actions pipelines

**Q:** How to enable debug logging?

**A:** Option 1: Rerun failed workflow with `Enable debug logging` checkbox selected:

![image](https://github.tools.sap/project-piper/project-piper.github.tools.sap/assets/95136/5a915311-6d18-4cc8-8814-c60170efc7a6)

Option 2: Enable debug logging by adding a new repository variable in **Settings → Secrets and variables → Variables** with key `ACTIONS_STEP_DEBUG` and value `true`.

![image](https://github.tools.sap/project-piper/project-piper.github.tools.sap/assets/95136/3a37b95d-1cde-4c21-adc5-16dba2e8a19d)

**Additionally**, set `verbose: true` in `general` section of your **.pipeline/config.yml** to enable verbose logging of Piper.

## Builds in Jenkins GPP being superseded by later builds triggered in parallel

**Q:** I have a build in Jenkins GPP that is being superseded by a later build triggered in parallel. How can I prevent this?

**A:** This issue occurs when Jenkins **Pipeline: Milestone Step** version is outdated and is incompatible with the Jenkins version installed in JaaS infrastructure.
Make sure to update the plugin to version [138.v78ca_76831a_43](https://plugins.jenkins.io/pipeline-milestone-step/releases/#version_138.v78ca_76831a_43) or above.
To update the plugin, go to **Manage Jenkins** > **Plugins** > **Updates** tab, select the **Pipeline: Milestone Step** plugin and click **Update**. You must have admin privileges to perform plugin updates.

## Artifactory API Key deprecation

**Q:** Would downloadArtifactsFromNexus step (still being used by Xmake users ) also be impacted because of artifactory api key deprecation?

```text
Artifactory API key has been replaced with artifactory reference token.
```

**A:** Xmake will always promote the artifact to [internal artifactory](https://int.repositories.cloud.sap/) and also to [common repo / internet facing artifactory](https://common.repositories.cloud.sap/), in case of double promote, more info can be found [here](https://github.wdf.sap.corp/pages/xmake-ci/User-Guide/Setting_up_a_Build/About_Build_Properties/Add_builds_properties/#private-cloud-shipment)) . But since the artifact is always in internal artifactory , we look for the internal artifactory url in the step downloadArtifactsFromNexus and use this url for download. Anonymous access is still allowed for internal artifactory and the credentials are anyways not needed. So downloadArtifactsFromNexus should not be affected and can be used without any disruptions.

## Github.com quota

**Q:** My pipeline runs ages when loading a library from a repository due to a quota limit on github.com:

```text
GitHub API Usage: Current quota has x remaining (x over budget). Next quota of 60 in x min. Sleeping until reset.
```

**A:** The library is not configured correctly. The Github.com API has a query limit that is reached if the retrieval method *Github* is used. Instead use the *Git* retrieval method listed under *legacy SCM*. The process is also described in detail [here](lib/setupLibrary.md#set-up).
The library can also be configured with a [Groovy script](https://github.wdf.sap.corp/ContinuousDelivery/piper-docker-jenkins/blob/master/scripts/libraries.groovy) via the Script Console.

## Custom Piper Pipelines on Azure DevOps

**Q:** How to identify whether i have a General purpose Piper pipeline or a custom Piper pipeline?

**A:** For Azure DevOps:

1. Check the pipeline definition in `azure-pipelines.yml` file in the root directory. It should have one single resource definition like below:

    ```yml
    trigger:
    - main

    resources:
      repositories:
      - repository: piper-templates
        endpoint: <name-of-gh-endpoint>
        type: githubenterprise
        name: project-piper/piper-pipeline-azure

    extends:
      template: sap-piper-pipeline.yml@piper-templates
    ```

> anything other than this is considered as a custom configuration. For example, **name** parameter of the resource is not `project-piper/piper-pipeline-azure`

1. Check if there are any extensions in the pipeline, for now build, acceptance, integration, performance and release stages can be extended. Check the following:

   After the GPP definition in the azure-pipelines.yml file, there will be configuration to extend the pipeline with parameters like <stage>PreSteps and <stage>PostSteps (where <stage> can be any stage like `build`, `acceptance` and etc), for example:

```yml

extends:
  template: sap-piper-pipeline.yml@piper-templates
  parameters:
    buildPreSteps:
    - ...
```

## Github.com API rate limit in Azure

**Q:** My Azure pipeline shows an API rate limit error:

```text
HttpError: API rate limit exceeded
```

**A:** This error should not be occurring anymore since [this PR](https://github.tools.sap/project-piper/piper-azure-task/pull/199). Please [report it](forum.md) as a bug if it does.

## A stage runs unnecessarily in my Azure pipeline (e.g. integration, acceptance)

**Q:** A stage runs without any of its steps being applicable to my Azure pipeline, how can I prevent this?

**A:** It is possible to set a parameter in your Azure DevOps pipeline definition to skip the acceptance, performance, integration, and release stages. This looks for example like this:

```yml
extends:
  template: sap-piper-pipeline.yml@piper-pipeline-azure
  parameters:
    skipAcceptanceStage: true
```

## No such step 'setupCommonPipelineEnvironment'/'handlePipelineStepErrors'/

**Q:** My job fails e.g. with the following message:

```text
java.lang.NoSuchMethodError: No such DSL method 'setupCommonPipelineEnvironment' found among steps [....
```

or

```text
java.lang.NoSuchMethodError: No such DSL method 'handlePipelineStepErrors' found among steps
```

**A:** The Piper library implicitly loads the [open-source](https://github.com/SAP/jenkins-library) version of Piper which needs to be defined in your Jenkins as `piper-lib-os`. Please follow the setup steps [here](others/jenkins_instance/initial_jenkins_config.md#define-global-pipeline-libraries).
If you are not using the ready-made [Piper Templates](stages/introduction.md),
but rather the Pipeline library: you must [explicitly load](lib/setupLibrary.md#set-up)
the `piper-lib-os` at the beginning of your Jenkinsfile. In this case, the start of your Jenkinsfile should
say `@Library(['piper-lib', 'piper-lib-os']) _`.

## Pipeline Optimizations

### Security vulnerabilities with CVSS score >=7

**Q:** What about security vulnerabilities with CVSS score >=7

**A:** Latest information can be found here: <https://wiki.wdf.sap.corp/wiki/display/PSSEC>

As per the time of writing (March 16th, 2021):
> In the following case an exception request does not need to be requested any more:
>
> If a Software already contains vulnerabilities with CVSS base score > 7.0 and a new release does not fix the issue an exception approval is not needed in case:
>
> * the vulnerailitiy is already released and not introduced newly by the new (feature) release
> * the vulnerability will be timely fixed (update May 2020: 30 days for very high, 30 days for high severe vulnerabilities)

Here, an important aspect is put on SLAs for fixing issues. It is accepted that at any point in time a new vulnerability can appear which should not stop the process in case it is handled correctly (SLAs, ...).

In addition it is important that the term "deployment" is distinguished from the term "release".
With the new pipeline optimization behavior we aim to address the "deployment" aspect NOT the "release" aspect.

!!! note Deployment vs. Release
    Deployment (to production) is the activity responsible for movement of new or changed artifact, hardware, software, documentation, process etc. to the (productive) environment. If the deployment does not introduce new or changed features it is not considered as a "release".
    It is best practice in cloud delivery scenarios to separate deployment from releases via feature toggles or other means to hide the new (not yet released) functionality.

Considering the separation of deployment and release combined with the aspects in the product standard security guidance it is considered acceptable to run Open Source Security scanning on the "main" branch only once per day.

In order for a developer to get earlier insights into possible vulnerabilities before even merging changes into the "main" branch, a developer can use e.g. language built-in features (npm, ...), WhiteSource IDE plugin or Piper's advanced pull-request voting.

### ECCN classification and PPMS update

**Q:** What about ECCN classification and PPMS update?

**A:** ECCN classification of Open Source components is done by a central team and NOT in the responsibility of the individual development teams.

It is the responsibility of the team though to adapt its product's ECCN questionnaires. This can for example be the case if a new feature is implemented using a newly introduced dependency. Another option could be that en existing dependency is used in a new way, affecting the ECCN classification.

The situation that the BOM is not fully up to date can already occur without switching to the new optimization feature of Piper.
Already previously, it did not provide a showstopper:
There is documentation available for this situation [https://wiki.wdf.sap.corp/wiki/x/az4siQ](https://wiki.wdf.sap.corp/wiki/x/az4siQ)

> An incomplete list of FOSS objects in your SCV’s comprised component list - because of missing FOSS mappings - does not violate a product standard.

### Check the licenses of all FOSS components

**Q:** We have to check the licenses of all FOSS components. How would we make sure to not deliver anything which has a "bad license"?

**A:** According to the latest guidance *"the product owner for each SAP solution is empowered to accept (or not) the risks associated with executing the IP Scans on a scheduled time"*.

Please see [IP Scan Wiki page](https://wiki.wdf.sap.corp/wiki/x/KTu0Yw) for details.

extract of version as per November 2021:
> This is particularly applicable to teams who scan their cloud solutions on a pre-defined timing schedule (e.g. once every 12 or 24 hours). In such cases, rerunning a dedicated scan directly prior to any type of “patch/deployment/release/etc.” may not be deemed a critical necessity by the responsible product owner, and as indicated, an affirmative acceptance of risks is sufficient to proceed without a further exceptional approval. (As a concise example, if a Dev team runs on a CI pipeline designed only to execute a scan every 24 hours, the pipeline’s documentation will make clear the IP risks assumed, and as such, simply operating on such a pipeline will automatically be assumed to be an affirmative acceptance of risks.)

Using the "shift-left" possibilities like e.g. Piper's advanced pull-request voting allows developers to get early insights and even further reduces the likelihood of introducing a component with a "bad license".

### Custom stage & step activation (Jenkins only)

**Q:** How can I activate a step or a stage with different conditions than piper defaulted way in Jenkins?

**A:** Piper uses the [piper-stage-config.yaml](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/blob/master/resources/piper-stage-config.yml) & [stageDefaults.yml](https://github.com/SAP/jenkins-library/blob/master/resources/com.sap.piper/pipeline/stageDefaults.yml) file to activate /deactivate a step in a stage in Jenkins.

However, to specify your own activation/deactivation conditions ,a resource file should be passed as parameter *stageConfigResource* for sapPiperStageInit or piperPipelineStageInit scripts.

Example (Jenkins):

```sh
sapPiperStageInit script: parameters.script, stageConfigResource: 'my-stage-conditions.yml'
```

Example (Azure DevOps):

```yml
extends:
  template: sap-piper-pipeline.yml@piper-pipeline-azure
  parameters:
    customStageConditions: my-stage-conditions.yml
    # customStageConditions could also point to release asset API URL or github raw content URL
```

Example (GitHub Actions):

```yml
jobs:
  piper:
    uses: project-piper/piper-pipeline-github/.github/workflows/sap-piper-workflow.yml@main
    with:
      custom-stage-conditions-path: my-stage-conditions.yml
      # customStageConditions could also point to release asset API URL or github raw content URL
```

**Q:** What should be the format of my custom stage conditions file?

**A:** Stage conditions in Piper are in CRD style / V1 format. This format provides possibilities for documenting a ready-made pipeline in an agnostic manner.
Example:

```yml
spec:
  stages:
  - name: build
    displayName: 'Build'
    description: |
      This is the description
      of the build stage
      which still needs to be updated.
    steps:
    - name: hadolintExecute
      description: Executes Haskell Docker Linter in docker based scenario to analyse structural issues in the Dockerfile.
      conditions:
      - filePattern: 'Dockerfile'
      orchestrators:
      - Jenkins
    - name: kanikoExecute
      description: Executes a Kaniko build which creates and publishes a container image.
      conditions:
      - config:
          buildTool:
          - 'docker'
```

!!! Attention

 Stage conditions file in Jenkins are now completely migrated to CRD style format for [piper-library releases](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/releases) >1.180.0 and [jenkins-library releases](https://github.com/SAP/jenkins-library/releases) > v1.241.0 . If you are using custom *stageConfigResource* please adapt your files to this format.

As reference, you could use  [Internal stage conditions](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/blob/master/resources/piper-stage-config.yml), [Open source stage conditions](https://github.com/SAP/jenkins-library/blob/master/resources/com.sap.piper/pipeline/stageDefaults.yml) for Jenkins or [Azure stage conditions](https://github.tools.sap/project-piper/sap-piper/blob/master/resources/piper-stage-config.yml)

## Sonar Code coverage

**Q:** How does SonarQube calculate my code coverage?

**A:**

```text
SonarQube itself does not calculate coverage. To include coverage results in your analysis, you must set up a third-party coverage tool and configure SonarQube to import the results produced by that tool.
```

Please check more detailed explanation [here](https://docs.sonarqube.org/9.9/analyzing-source-code/test-coverage/overview/).
For example: For a [java project](https://docs.sonarqube.org/9.9/analyzing-source-code/test-coverage/java-test-coverage/) SonarQube directly supports the JaCoCo coverage tool and does so by using the jacoco-maven-plugin. All you will have to do it is to have the plugin and goals in your pom.xml.

If you are using Piper native build(maven), this is already [taken care of](steps/mavenBuild.md#description).

## Pulling a private Docker image for a step

**Q**: How do I run a step with an image from a private Docker registry?

**A**:
See [this page](others/private_images.md) for more information.

## Parallel execution of Piper steps in custom pipelines

**Q**: What to do in case you have custom pipelines and are trying to execute steps in parallel, but encounter a `text file busy` error?

**A**: This error is related to file corruption. In the Piper General Purpose Pipeline, a new workspace is created for each stage, which helps to prevent file corruption issues. To avoid such issues during the parallel execution of Piper steps in custom pipelines, it is recommended to run each stage on a new Jenkins node.

### Example

Jenkinsfile

```Jenkinsfile
pipeline {
    agent none
    stages {
        stage('Run Tests') {
            parallel {
                stage('Test On Windows') {
                    agent {
                        label "windows"
                    }
                    steps {
                        bat "run-tests.bat"
                    }
                }
                stage('Test On Linux') {
                    agent {
                        label "linux"
                    }
                    steps {
                        sh "run-tests.sh"
                    }
                }
            }
        }
    }
}
```

## Does Piper support git submodules?

**A**:
Yes.

Jenkins: look over [here](others/jenkins_instance/create_jenkins_job.md#using-git-submodules).

Azure DevOps: set [this variable](https://github.tools.sap/project-piper/piper-pipeline-azure/blob/main/sap-piper-pipeline.yml#L24) to true in your `azure-pipelines.yml`.

GitHub Actions: see [this example](https://github.tools.sap/project-piper/piper-pipeline-github/blob/main/examples/custom_workflow/submodules.yml).

## Extensibility

### (Jenkins) Post stage extension is execute twice

**Q**: Why is my Jenkins Post stage extension executed twice?

**A**: This is [a bug](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues/3109).
There is a [workaround](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues/3109) that you can use in your extension to prevent this from happening.
Make sure not to wrap the execution of the original stage (`params.originalStage()`) in the if-block, but only your custom logic, otherwise some steps (such as `vaultRotateSecretId`) will be skipped.

## Stages being skipped in PRs

**Q:** Why does Piper not run  the acceptance/performance/integration test stages in PR pipelines?

**A:** These stages are only active in the main build pipeline, not for PRs . As part of PRs supporting trunk based development , we only provide running Build stage for PRs. Running other stages for every PR can cause interference between (e.g., deployments overriding each other) and significantly increase server/system load when multiple PRs are triggered simultaneously. This approach ensures pipeline stability and efficient resource usage. For more details, see [Piper PR Voting](https://pages.github.tools.sap/project-piper/build/native/#pull-request-voting)

## Rerunning Pipelines

**Q**: Is it possible to rerun a pipeline(failed/successful/cancelled)?

**A**: Rerunning pipelines is not supported in Piper, because handling is orchestrator-specific and the impact of reruns on Piper is unpredictable.
Please run every single time a new pipeline.

## Generating and Uploading Release Status File

**Q:** How can I generate and upload a release status?

**A:** According to cumulus documentation, it is recommended to generate a release status file with status "released" in the Release stage extension. To achieve this, you can create an [extension](extensibility.md). Note that after the release extension, the standard piper post stage should run which creates the file and also uploads the file to Cumulus.

For details on the json content structure and business logic please refer to the [documentation](https://wiki.one.int.sap/wiki/x/tbvpy#AutomaticallysetPipelineRunReleaseStatusbyFileUpload-release-information).

`writePipelineEnv` step now accepts a `value` flag that sets the status to a file. Example:

!!! tip ""

    === "Github Actions"

        ```yaml
        name: 'PostRelease'
        runs:
          using: "composite"
          steps:
            - name: Create releaseStatus
              uses: SAP/project-piper-action@d39f6ccbe7cbffe9690443bc7831681c1efeb4d1 # v1.16.0
                with:
                  step-name: writePipelineEnv
                  flags: --value custom/releaseStatus={"releaseStatus":"released","selectionType":"latest-delivery"}
        ```

    === "Jenkins"

        ```groovy
        void call(Map params) {
            def jsonValue = /{"releaseStatus":"released","selectionType":"latest-delivery"}/
            writePipelineEnv(
                script: params.script,
                flags: ["--value", "custom/releaseStatus=${jsonValue}"]
            )
        }
        return this
        ```

    === "Azure"

        ```yaml
        steps:
          - task: piper@1
            name: 'writePipelineEnv'
            inputs:
              stepName: "writePipelineEnv"
              flags: '--value "custom/releaseStatus={\"releaseStatus\":\"released\",\"selectionType\":\"latest-delivery\"}"'
        ```

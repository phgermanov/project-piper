# Step execution

## Executing Piper library steps

!!! tip ""

    === "Jenkins"

        Jenkins integration is based on the [Jenkins shared libraries](https://jenkins.io/doc/book/pipeline/shared-libraries/) concept.

        Every Piper step is wrapped as a groovy-based step which can directly be used inside a Jenkins pipeline:

        ```groovy
        @Library(['piper-lib', 'piper-lib-os']) _
        // piper-lib refers to the SAP-specific steps
        // piper-lib-os refers to the Open Source steps

        pipeline {
            stages {
                stage('Versioning') {
                    steps {
                        // Passing script:this is essential to preserve the context of the root script and allow exchange of information between steps.
                        // Typically, further configuration is passed via the file .pipeline/config.yml to have a clear separation of flow and config
                        // Having this clear separation makes it simple to reuse pipeline scripts across multiple repositories.
                        artifactPrepareVersion script: this
                    }
                }
            }
        }

        ```

    === "Azure DevOps"

        Azure DevOps integration is based on the concept of [tasks](https://docs.microsoft.com/en-us/azure/devops/pipelines/process/tasks?view=azure-devops&tabs=yaml).

        There is one universal [Piper task](https://github.tools.sap/project-piper/piper-azure-task) available to execute Piper. The name of the step which should be executed is passed as a parameter:

        ```yml
        trigger:
        - main

        pool:
          name: "$(hyperspace.azure.pool.name.mshosted)"
          vmImage: "$(hyperspace.azure.pool.vmImage.ubuntu)"

        steps:
          - task: piper@1
            name: versioning
            inputs:
              stepName: artifactPrepareVersion
              # optionally, flags can be provided
              # typically, further configuration is passed via the file .pipeline/config.yml though to have a clear separation of flow and config
              #flags:
        ```

    === "GitHub Actions"

        GitHub Actions integration is based on the concept of [actions](https://docs.github.com/en/enterprise-server@3.9/actions).

        There is one universal [Piper action](https://github.com/SAP/project-piper-action/tree/main) available to execute Piper. The name of the step which should be executed is passed as a parameter:

        ```yml
        name: "CI"

        on:
          push:
            branches:
              - main
          pull_request:
            branches:
              - main

        jobs:
          build:
            runs-on: self-hosted
            steps:
              - uses: actions/checkout@v3
              - uses: SAP/project-piper-action@main
                with:
                  step-name: mavenBuild
                  flags: '--publish --createBOM --logSuccessfulMavenTransfers'
        ```

    === "Command Line"

        The Piper step library can be used on a command line as well.

        You can download the binaries from:

        * Non-SAP-specific steps in [Open Source binary](https://github.com/SAP/jenkins-library/releases/latest)
        * SAP-specific steps in [Inner Source binary](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/releases/latest)

        Once the binaries are available in your `$PATH`, they are ready to use:

        ```sh
        $ # get details about the usage of the Open Source binary
        $ piper --help
        $ ...
        $ #
        $ # get details about the usage of the SAP InnerSource version
        $ sap-piper --help
        $ ...
        ```

## Execution environment

Piper library steps which do not have dedicated runtime pre-requisites (e.g. pure API calls) are executed on the respective executor of the orchestrator (e.g. Jenkins node, Azure runner).

Piper library steps with runtime pre-requisites (e.g. `maven`, `npm`) are executed inside a Docker container.
Two flavors are possible here:

1. Execution via Docker

    A Docker image will be started, the current orchestrator workspace will be mounted into the container and the Piper library step will be executed inside the container.

2. Execution inside Kubernetes (e.g. Jenkins as a Service - JaaS)

    A new Kubernetes pod will be started inside the cluster which contains the required runtime container. **This container will run with a new dedicated workspace which is empty by default!**

    If `stashContent` is provided (either as parameter or as a default value of the Piper library step) files from the [stash](https://www.jenkins.io/doc/pipeline/steps/workflow-basic-steps/#stash-stash-some-files-to-be-used-later-in-the-build) are copied into the pod's workspace.

    Only in case the `stashContent` parameter is empty the content of the current workspace is stashed and copied into the pod's workspace.

??? note "Jenkins: Default stashing behavior on Kubernetes (e.g. JaaS)"

    In order to prevent negative side-effects the stashing excludes a list of files according to the [Ant default excludes](http://ant.apache.org/manual/dirtasks.html#defaultexcludes). This for example excludes also the folder `.git`.

    In case a different behavior is required, please use the parameter `stashNoDefaultExcludes` from step [`dockerExecute`](../steps/dockerExecute.md).

!!! note "Changing the execution environment"

    The execution environment can be changed by adapting the Piper configuration, like

    ```yml
    # .pipeline/config.yml

    ...

    steps:
      whitesourceExecuteScan:
        dockerImage: my.registry.com/path/to/myImage
        # further parameters can be provided as indicated in the respective step documentation
        #dockerOptions: ...  // provides options to docker run

    ```

## Providing credentials

We recommend to store secrets in [SAP's Vault system](https://vault.tools.sap/) which will be set up as part of [Hyperspace Onboarding](https://hyperspace.tools.sap/home).
The SAP defaults contain already the correct system url.

In addition the config file requires following entries:

```yaml
# .pipeline/config.yml

vaultBasePath: "<your-pipeline-group>" # see Hyperspace Onboarding
vaultNamespace: "<your-namespace>" # see Hyperspace Onboarding
vaultPipelineName: "<your-pipeline-name>" # see Hyperspace Onboarding

# Defines the location of the Vault server if SAP defaults are not used
# vaultServerUrl
```

Using credentials from Vault works as follows:

!!! tip ""

    === "Jenkins"

        For Jenkins you need to maintain the credentials inside Jenkins according to [using credentials](https://www.jenkins.io/doc/book/using/using-credentials/).

        Then you can refer to the credential ids inside the configuration via:

        ```yml
        # .pipeline/config.yml

        general:
          vaultBasePath: "<your-pipeline-group>" # see Hyperspace Onboarding
          vaultNamespace: "<your-namespace>" # see Hyperspace Onboarding
          vaultPipelineName: "<your-pipeline-name>" # see Hyperspace Onboarding
          vaultAppRoleTokenCredentialsId: myJenkinsVaultTokenCredentialsId
          vaultAppRoleSecretTokenCredentialsId: myJenkinsVaultSecretCredentialsId
        ```

    === "Azure DevOps"

        The Azure task considers following variables defined in Azure. Both are set via the [Hyperspace Onboarding](https://hyperspace.tools.sap/home):

        * `hyperspace.vault.roleId`
        * `hyperspace.vault.secretId`

    === "GitHub Actions"

        The GitHub action considers following environment variables which should be defined as [repository or organization secrets](https://docs.github.com/en/codespaces/managing-codespaces-for-your-organization/managing-encrypted-secrets-for-your-repository-and-organization-for-github-codespaces).

        ```yml
        env:
          PIPER_vaultAppRoleID: ${{ secrets.PIPER_VAULTAPPROLEID }}
          PIPER_vaultAppRoleSecretID: ${{ secrets.PIPER_VAULTAPPROLESECRETID }}
        ```

    === "Command Line"

        Provide the connection secrets to Vault as environment variables:

        ```sh
        $ export PIPER_vaultAppRoleID=appRole
        $ export PIPER_vaultAppRoleSecretID=appRoleSecret
        $ git clone ...
        $ piper artifactPrepareVersion
        ```

Further information is available in the [Vault chapter](../infrastructure/vault.md).

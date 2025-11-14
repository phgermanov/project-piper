# Development of the Piper task

## Installation and usage of the development task

There is a [development version](https://marketplace.visualstudio.com/items?itemName=ProjectPiperDEV.piper-azure-task-dev&targetId=933b982b-2375-4823-88c1-ce425a0c6aab&utm_source=vstsproduct&utm_medium=ExtHubManageList) of the Piper task.
To use it, the development task needs to be installed in your ADO organization.
It's already installed in the [`hyperspace-pipelines-test`](https://dev.azure.com/hyperspace-pipelines-test/) organization, and there are [pipelines](https://dev.azure.com/hyperspace-pipelines-test/project-piper-test/_build) there that are ready to use.

To use it, the steps in the pipeline template need to use the `piper-dev` task instead of `piper`:

`- task: piper-dev@1`

## How to test changes

A new version of the development task is built and published automatically on every commit to a PR, which can be seen under the PR checks.

In the Azure GPP repository, there is an [automatically](https://github.tools.sap/project-piper/piper-pipeline-azure/blob/main/.github/workflows/dev_azure_task.yml) maintained [branch](https://github.tools.sap/project-piper/piper-pipeline-azure/tree/devAzureTask) that uses the development task.
You can simply run a test pipeline with the GPP from this branch. By default, `piper-dev@1` is being used, but if you want to pin specific dev version of the task, you can trigger workflow manually by inputing the version:
![image](https://media.github.tools.sap/user/95136/files/7e3cc889-dd36-42c9-979b-a7096a6808c9)

There is a slight delay between the publishing of the new version, and it actually being available on ADO, so keep this in mind when you are running test pipelines - pipelines might still use the previous version for a bit.
It's best to wait a few minutes before triggering a test pipeline, after publishing a new version.
In doubt, you can confirm if your test pipeline is using the correct (newly built) dev version, by checking if the [published version](https://github.tools.sap/project-piper/piper-azure-task/actions/workflows/publish_task.yml) matches with the executed version:

![This screenshot comes from the [Piper publish workflow](https://github.tools.sap/project-piper/piper-azure-task/actions/workflows/pr_publish.yml)](https://media.github.tools.sap/user/95136/files/84814aec-5830-4903-9d57-3c3a893ed06d)

![And this one shows a GPP pipeline run, with the version of the used task](https://media.github.tools.sap/user/95136/files/4236a86c-7ff7-4ead-9334-19030367fd90)

### Building the development task manually

You can trigger the [Piper task CI/CD pipeline](https://github.tools.sap/project-piper/piper-azure-task/actions/workflows/pr_publish.yml) manually on your branch, to build a new version of the development task with your changes.

## Limitations

As far as I know, it's only possible to use the latest version of the development task.
Since every commit to a PR will build the development task, simultaneous development on different branches can cause conflicts in testing.

## Permissions

Feel free to ask the [code owners](.github/CODEOWNERS) in case you need permissions to e.g. install the development task, or get access to the ADO test organization that we use.

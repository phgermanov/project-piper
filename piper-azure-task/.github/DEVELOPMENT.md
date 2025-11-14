# Developer's documentation

[Create a custom pipelines task extension](https://docs.microsoft.com/en-us/azure/devops/extend/develop/add-build-task?view=azure-devops#step-1-create-a-custom-task)

## Release

The task is built using a commit-based [Azure DevOps pipeline](https://dev.azure.com/hyperspace-pipelines/project-piper/_build?definitionId=277) and then published to the [VisualStudio Marketplace](https://marketplace.visualstudio.com/items?itemName=ProjectPiper.piper-azure-task-dev) using a [publisher (`projectpiper`)](https://marketplace.visualstudio.com/manage/publishers/projectpiper).
The pipeline increases the `patch` number automatically.

To publish a task there is an [Azure token](https://dev.azure.com/hyperspace-pipelines/_usersSettings/tokens) needed. The token need to have `Marketplace (publish)` scope and no organization assigned to it.

### Production Releases

Task: [SAP Piper library](https://marketplace.visualstudio.com/items?itemName=ProjectPiper.piper-azure-task-dev)
Publisher: [Project Piper](https://marketplace.visualstudio.com/manage/publishers/projectpiper)

### Evaluation Releases

Task: [SAP Piper library](https://marketplace.visualstudio.com/items?itemName=ProjectPiperDev.piper-azure-task-dev)
Publisher: [Project Piper (DEV)](https://marketplace.visualstudio.com/manage/publishers/projectpiperdev)

Next to the general release there is an evaluation release available which is published using a [different publisher (`projectpiperdev`)](https://marketplace.visualstudio.com/manage/publishers/projectpiperdev). The evaluation release is created during the commit-based pipeline for Pull-Request (last one wins!).

To use the evaluation release one need to remove the regular task from the Azure organization and install [this task](https://marketplace.visualstudio.com/items?itemName=ProjectPiperDev.piper-azure-task-dev) instead. So it can't be used in the  `hyperspace-pipelines` organisation, since it would clash with the productive version of the task.

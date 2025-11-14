---
date: 2025-03-27
title: Docker Hub pull rate limits reduction mitigation
authors:
  - gulomjon
  - amandeep
categories:
  - Azure DevOps
  - System Trust
  - Docker
---

<!-- markdownlint-disable MD013 -->

Docker Hub will reduce the rate limits for downloads (docker image pull) from docker.io by 40% from April 1st, 2025.
We aim to enable mitigation actions to avoid pipelines execution disruption.

<!-- more -->

## 📢 Do I need to do something?

New rate limits will be applied on April 1st. So if after that date your Azure pipelines will start failing with `TOOMANYREQUESTS` error coming from hub.docker.com (or docker.io), you can try the following based on the category your pipeline fall under -

- Case 1. Piper GPP with or without extensions, created using Hyperspace. This category of pipelines shouldn't take any actions, as the issue is mitigated on the Piper side.
- Case 2. Piper GPP with custom stages (Piper parts are being consumed to create custom stages). If you created a pipeline using Hyperspace, you can try the below recommendation (for manually created pipelines this will not work!):
    1) Make sure that Piper's [Init stage](https://github.tools.sap/project-piper/piper-pipeline-azure/blob/main/stages/init.yml) is executed in your custom pipeline. Otherwise, this recommendation will not work.
    2) In the beginning of your custom stage, add a stage level variable called `systemTrustToken`:

       ```yaml
        stages:
          - stage: MyCustomStage
            variables:
              systemTrustToken: $[ stageDependencies.Init.setup.outputs['getSystemTrustToken.systemTrustSessionToken'] ]
       ```

       After that all Piper steps, that execute inside a docker container will start pulling their docker images from [Common repository](https://common.repositories.cloud.sap/) instead of docker hub.

       Also note that this will work only if your stage is running in a Microsoft hosted agents. If you configured your stage to run in a self-hosted agent, we'd advice you to configure the docker daemon in the agent instead \[If you face an issue with configuring docker daemon then contact your agent maintainers].

- Case 3. Your pipelines are using forked and modified version of [our GPP](https://github.tools.sap/project-piper/piper-pipeline-azure). In this case please check the following commit and sync with [those changes](https://github.tools.sap/project-piper/piper-pipeline-azure/commit/09d3afcf54af66a72a1e8ebbfbc211718d7c2320) where necessary.

- Case 4. Customized Piper Pipelines (Piper steps are cherry picked to write own template)
  Can only be tackled on case to case basis as we can't imagine all the ways in which the customizations would have been done for your specific scenario.
  Please try the approach described below or make use of the official support channel [ISV-DL-CICD-PIPELINE-PIPER](https://itsm.services.sap/sp?id=sc_cat_item&sys_id=703f22d51b3b441020c8fddacd4bcbe2&service_offering=455f43371b487410341e11739b4bcb67) and we shall guide you.

<!-- markdownlint-disable-next-line MD001 -->
#### Configuring your Piper steps to pull from Common repo instead of hub.docker.com

1) Create an Identity Token in Common repo. How to do it you can [find here](https://jfrog.com/help/r/jfrog-platform-administration-documentation/generate-identity-token). It's done via [Profile page](https://common.repositories.cloud.sap/ui/user_profile) in Common repo.

2) You need to configure a service connection in your Azure project. How to do it, you can [find here](https://learn.microsoft.com/en-us/azure/devops/pipelines/library/service-endpoints?view=azure-devops#create-a-service-connection). Type of service connection should be Docker registry.

3) Set Piper's `dockerRegistryConnection` input. The value must be the same as you configured in step 2 for field **Service Connection Name**. Example:

```yaml
  - task: piper@1
    inputs:
      stepName: yourStepName
      dockerRegistryConnection: myCommonRepoConnectionName
      dockerImage: docker-hub.common.repositories.cloud.sap/maven:latest
```

<!-- markdownlint-disable-next-line MD029 -->
4) Prepend `docker-hub.common.repositories.cloud.sap/` registry name to the docker image that your step is using. You can do it via `dockerImage` parameter of the step in config.yml or directly passing as input to the step call as mentioned above.

## 💡 What do I need to know?

1) If you are using Piper GPP(General Purpose Pipeline Template) with Hyperspace supported Orchestrators (Jenkins as a Service or GitHub Actions) then you don't need to take any actions.
Because of the following -

- JAAS already has a caching mechanism in place which significantly reduces the risk for running into rate limit issues.
- GitHub Actions uses SUGAR runners which also has an in-built mirroring for docker images. Hence we don't see any foreseeable risk here as well.

!!! note

    System Trust integration with Azure has enabled us to mitigate the risk with reduced rate limits from DockerHub.
    Currently, artifactory authentication via System Trust is available for all pipelines created in Hyperspace Portal and Onboarding.

2) The logs indicate that System Trust is being used as follows: `System trust token retrieved successfully`.
3) Common repository uses the caching mechanism to fetch the docker images from Dockerhub.

## Troubleshooting Guide

- System to System Trust integration with Azure Pipelines might throw an error - Failed to retrieve system trust token. But this shouldn't fail the pipelines as we have fallback mechanism in place to retrieve the secrets from vault and to pull the docker images from DockerHub.
- If the pipeline fails and you don't see a clear error message then switch on verbose logging for Azure Pipelines as Piper verbose wouldn't be sufficient. More details on How to enable it can be found [here](https://learn.microsoft.com/en-us/azure/devops/pipelines/troubleshooting/review-logs?view=azure-devops&tabs=windows-agent#configure-verbose-logs).

## 📖 Learn more

[System Trust Documentation](https://pages.github.tools.sap/system-trust/)

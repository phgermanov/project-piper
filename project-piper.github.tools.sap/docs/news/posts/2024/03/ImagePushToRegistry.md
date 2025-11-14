---
date: 2024-03-22
title: ImagePushToRegistry to strengthen the GitOps deployment via Piper
authors:
  - amandeep
categories:
  - General Purpose Pipeline
  - Docker
---

There is new Piper step ``ImagePushToRegistry`` which is now a part of the Piper GPP: Build stage that allows to push the build docker image from step ``kanikoBuild``/``cnbBuild`` into a target docker registry that a user can bring in.

<!-- more -->

## ❓ Why is this functionality needed ?

Piper GPP allowed to build and push docker images to staging services repo . The staging services always created a new docker registry on the fly with a new repository url, repository password and username every time the pipeline ran.

deploying the above built image in pre-promote stage(s) like Acceptance and Performance was only possible via the ``kubernetesDeploy`` step as this step when configured will first create the secret in the cluster and then pull the the build image into the cluster during deployment.

deploying the above built image via ``gitOpsUpdateDeployment`` step was not possible in pre-promote stage(s) like Acceptance and Performance since only updating the image name/tag in a github repo would not have created the needed secret to pull the image in the k8s cluster

## 📢 Do I need to do something?

if you wish to move the staging repository docker image to a fixed registry of your choice you would need to set the following parameters for the step [imagePushToRegistry](https://www.project-piper.io/steps/imagePushToRegistry/#imagepushtoregistry)

- [targetregistryurl](https://www.project-piper.io/steps/imagePushToRegistry/#targetregistryurl), [targetregistrypassword](https://www.project-piper.io/steps/imagePushToRegistry/#targetregistrypassword)
and [targetRegistryUser](https://www.project-piper.io/steps/imagePushToRegistry/#targetregistryuser)

OR

- [dockerconfigjson](https://www.project-piper.io/steps/imagePushToRegistry/#dockerconfigjson) with the correct target registry credentials

the source image registry credentials and metadata will be passed on to this step via ``kanikoBuild``/``cnbBuild``

## 💡 What do I need to know?

[imagePushToRegistry](https://www.project-piper.io/steps/imagePushToRegistry/#imagepushtoregistry) is already a part of the piper provided ``Build`` stage and can be activated via ``imagePushToRegistry: true`` in the stage.

[GitOpsUpdateDeployment](https://www.project-piper.io/steps/gitopsUpdateDeployment/) is now part of the piper provided ``Acceptance`` and ``Performance`` stage and can be activated via ``gitOpsUpdateDeployment: true`` in the stage(s). [GitOpsUpdateDeployment](https://www.project-piper.io/steps/gitopsUpdateDeployment/) can be configured to fix the target the registry url for the deployment. Please make sure that the target registry credentials are already created in the cluster for GitOps deployment to work.

To create your own registry in SAP maintained artifactory instance(s) use the [Self Service](https://pages.github.tools.sap/Common-Repository/Artifactory-Internet-Facing/how_to/how_to_request_new_repo/#self-service)

### Sample Piper config

Scenario: ``kanikoBuild/cnbBuild`` is used to build docker image and the image needs to be deployed using ``gitOpsUpdateDeployment`` step to three different cluster in Acceptance, Performance and Release stage

``` yaml
stages:
    Build:
        imagePushToRegistry: true
        kanikoExecute: true
        # cnbBuild: true # the docker image can be built with kaniko or cnbBuild

    Performance:
        filePath: manifest-performance-stage.yaml # filePath is a parameter of step gitOpsUpdateDeployment that is overwritten in Performance stage

    Acceptance:
        filePath: manifest-acceptance-stage.yaml # filePath is a parameter of step gitOpsUpdateDeployment that is overwritten in Acceptance stage

    Release:
        filePath: manifest-release-stage.yaml # filePath is a parameter of step gitOpsUpdateDeployment that is overwritten in Release stage

steps:
    imagePushToRegistry:
        targetRegistryUrl: https://piper-test.common.repositories.cloud.sap

    gitOpsUpdateDeployment: # the step is only configured once where all parameters are common for deployment to all stages, any parameter specific to the stage should be maintained at the stage level
        forcePush: true
        chartPath: helm/azure-demo-k8s-go
        deploymentName: azure-demo-k8s-go
        serverUrl: https://github.tools.sap/project-piper/azure-demo-k8s-go.git
        tool: helm

```

[targetregistryuser](https://www.project-piper.io/steps/imagePushToRegistry/#targetregistryuser) and [targetregistrypassword](https://www.project-piper.io/steps/imagePushToRegistry/#targetregistrypassword)
are stored in vault

[Sample github repo for reference](https://github.tools.sap/project-piper/azure-demo-k8s-go/tree/imagePushToRegistry-gitopsUpdateDeployment-example)

## ➡️ What's next?

[imagePushToRegistry](https://www.project-piper.io/steps/imagePushToRegistry/#imagepushtoregistry) is not a replacement for [containerPushToRegistry](https://www.project-piper.io/steps/containerPushToRegistry/) currently. for the future we plan to harmonize the step and only keep [imagePushToRegistry](https://www.project-piper.io/steps/imagePushToRegistry/#imagepushtoregistry)

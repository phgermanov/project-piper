# pushToDockerRegistry

## Description

This step allows you to push a docker image into a docker registry.

There are various docker registries such as e.g. [dockerhub][05876491] , [SAP Internal Artifactory][1d7d9be1]
 , [SAP DMZ Artifacory][0a61afde] , etc..

  [05876491]: https://hub.docker.com/ "dockerhub"
  [1d7d9be1]: https://docker.wdf.sap.corp:10443 "SAP Internal Artifactory"
  [0a61afde]: https://docker.repositories.sap.ondemand.com "SAP DMZ Artifacory"

By default the local docker image will be pushed. In case you want to pull an existing image from a different docker registry, a source image and source docker registry needs to be specified.
For example you can pull a docker image from internal docker registry and push to an external docker registry.

## Prerequisites

You need to have a valid user with read/write permissions in the docker registry. Credentials for the target docker registry have been configured in Jenkins with a dedicated Id.

## Example

Usage of pipeline step:

**OPTION A:** To pull a docker image from an existing docker registry and push to a different docker registry:

```groovy
pushToDockerRegistry script: this,
                     sourceRegistryUrl: 'mysourceRegistry',
                     sourceImage: 'mySourceImageNameWithTag',
                     dockerRegistryUrl: 'https://my.target.docker.registry:50000'
```

**OPTION B:** To push a locally build docker image into the target registry:

```groovy
pushToDockerRegistry script: this,
                     dockerRegistryUrl: 'https://my.target.docker.registry:50000'
```

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|dockerBuildImage|yes|||
|dockerCredentialsId|yes|||
|dockerImage|no|||
|dockerRegistryUrl|yes|||
|skopeoImage|no|`piper.int.repositories.cloud.sap/piper/skopeo`||
|sourceImage|no|||
|sourceRegistryUrl|no|||
|tagLatest|no|`false`|`true`, `false`|

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|dockerBuildImage|X|X|X|
|dockerCredentialsId|X|X|X|
|dockerImage|X|X|X|
|dockerRegistryUrl|X|X|X|
|skopeoImage|X|X|X|
|sourceImage|X|X|X|
|sourceRegistryUrl|X|X|X|
|tagLatest|X|X|X|

### Available parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| skopeoImage | no | `piper.int.repositories.cloud.sap/piper/skopeo` ||
| sourceImage | no | - | If not set, a local dockerImage will be used |
| sourceImageName | no | - | If not set, a local dockerImage will be used |
| sourceImageTag | no | - | If not set, a local dockerImage will be used |
| dockerRegistryUrl | no | `globalPipelineEnvironment.getConfigProperty('dockerRegistry')` | |
| tagLatest | no | `false` | `true`\|`false` |

Details:

* `script` defines the global script environment of the Jenkinsfile run. Typically this is passed to this parameter. This allows the function to access the [globalPipelineEnvironment](../objects/globalPipelineEnvironment.md) for retrieving e.g. configuration parameters.
* `skopeoImage` only used if no Docker daemon available on your Jenkins image: Docker image to be used for [Skopeo](https://github.com/containers/skopeo) calls
* `sourceRegistryUrl` defines the full url of the docker registry of the source Image. If not set, a local dockerImage will be used
* `sourceImage` defines the name of the source image including tag. If not set, a local dockerImage will be used
* `dockerRegistryUrl` defines the full url of the target docker registry (`<protocol>://<dockerRegistry>:<port>`)
* `tagLatest`: if set to true, the docker image will additionally be pushed with tag 'latest'

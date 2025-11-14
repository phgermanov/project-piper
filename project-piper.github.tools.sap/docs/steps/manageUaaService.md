# manageUaaService

## Description

Creates or updates an instance of UAA backing service in Cloud Foundry. Uses xs-security.json file to configure the service instance.
The property `xsappname` will be enhanced with `-[SPACENAME]` to assure unique xsappnames across all spaces.

Variables supported in xs-security.json:

- `${subaccountId}`-> will be replaced with value of parameter `subaccountId`
- `${space}` -> will be replaced by value of parameter `cfSpace` (in lower case)

## Prerequisites

- Cloud Foundry organization, space and deployment user are available
- Credentials for deployment have been configured in Jenkins with a dedicated Id

## Example

Usage of pipeline step:

```groovy
manageUaaService script: this, instanceName: 'my-uaa'
manageUaaService script: this, instanceName: 'my-uaa', xsSecurityFile: 'security/xs-security.json'
```

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|cfApiEndpoint|no|`https://api.cf.sap.hana.ondemand.com`||
|cfCredentialsId|yes|||
|cfOrg|yes|||
|cfServiceInstance|yes|||
|cfServicePlan|no|`broker`||
|cfSpace|yes|||
|cfSubaccountId|no|||
|dockerImage|no|`docker.wdf.sap.corp:50000/piper/cf-cli`||
|dockerWorkspace|no|`/home/piper`||
|stashContent|no|<ul><li>`securityDescriptor`</li></ul>||
|xsSecurityFile|no|`xs-security.json`||

### Details

- `script` - global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`globalPipelineEnvironment`](../objects/globalPipelineEnvironment.md) for retrieving e.g. configuration parameters.
- `cfCredentialsId` - credentials used for Cloud Foundry deployment (`CF_USER`). Either this information needs to be maintained or the CAM information (camCredentialsId, deployUser, camSystemRole).
- `cfApiEndpoint` - Cloud Foundry API endpoint.
- `cfOrg` - Cloud Foundry organization where the application will be deployed to.
- `cfSpace` - Cloud Foundry space where the application will be deployed to.
- `cfServiceInstance` - name of the UAA service that should be created/updated
- `cfServicePlan` - plan of the UAA service (either `broker` or `application`)
- `xsSecurityFile`- xs-security.json that should be used to configure the service
- `cfSubaccountId`- every usage of `${subaccountId}` in the xs-security.json will be replaced by this value

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|cfApiEndpoint|X|X|X|
|cfCredentialsId|X|X|X|
|cfOrg|X|X|X|
|cfServiceInstance|X|X|X|
|cfServicePlan|X|X|X|
|cfSpace|X|X|X|
|cfSubaccountId|X|X|X|
|dockerImage|X|X|X|
|dockerWorkspace|X|X|X|
|stashContent|X|X|X|
|xsSecurityFile|X|X|X|

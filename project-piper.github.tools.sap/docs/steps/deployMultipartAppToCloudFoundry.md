# deployMultipartAppToCloudFoundry

## Description

A lightweight deployment of multiple parts (e.g. Java application + UI) of one app to Cloud Foundry by only using standard Cloud Foundry operations.

Features:

* Blue-green-deployment as described at <https://docs.cloudfoundry.org/devguide/deploy-apps/blue-green.html>
* Registering Service Brokers (cf. <https://github.wdf.sap.corp/xs2/node-sbf>)  
* Setting environment variables for application parts (for varying variables; stable ones should go into manifest.yml)
* Replacement of URL references to other app parts in manifest.yml (e.g. `${myAppName-url}`)

## Prerequisites

* Cloud Foundry organization, space and deployment user are available
* Credentials for deployment have been configured in Jenkins with a dedicated Id

## Example

Usage of pipeline step:

```groovy
def modules = [
        [cfAppName: 'my-app', cfManifestPath: 'app/manifest.yml', cfEnvVariables: ['my-env-variable' : 'my-value']],
        [cfAppName: 'my-ui', cfManifestPath: 'ui/manifest.yml'],
        [cfAppName: 'my-broker', cfManifestPath: 'broker/manifest.yml', registerAsServiceBroker : true, serviceBrokerHasSpaceScope: true]
]
deployMultipartAppToCloudFoundry modules: modules, script: this
```

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|cfApiEndpoint|no|`https://api.cf.sap.hana.ondemand.com`||
|cfCredentialsId|yes|||
|cfDomain|no|`cfapps.sap.hana.ondemand.com`||
|cfOrg|yes|||
|cfSpace|yes|||
|deployType|no|`standard`||
|dockerImage|no|`docker.wdf.sap.corp:50000/piper/cf-cli`||
|dockerWorkspace|no|`/home/piper`||
| modules | yes | `[]` | |
| module: cfAppName | yes | | |
| module: cfManifestPath | yes | | |
| module: cfHostname | no | ${cfAppName}-${cfSpace} | |
| module: cfDomain | no | value of global `cfDomain` | |
| module: cfEnvVariables | no | `[]` | |
| module: deployType | no | value of global `deployType` | `blue-green` or `standard` |
| module: registerAsServiceBroker | no | `false`  | `true` or `false` |
| module: serviceBrokerHasSpaceScope | no | `true`  | `true` or `false` |

### Details

* `script` - global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`globalPipelineEnvironment`](../objects/globalPipelineEnvironment.md) for retrieving e.g. configuration parameters.
* `cfCredentialsId` - credentials used for Cloud Foundry deployment (`CF_USER`).
* `cfApiEndpoint` - Cloud Foundry API endpoint.
* `cfOrg` - Cloud Foundry organization where the application will be deployed to.
* `cfSpace` - Cloud Foundry space where the application will be deployed to.
* `cfDomain` - Cloud Foundry domain
* `deployType` - Sets deployment type for all modules (unless specified otherwise on module level)
* `modules` - An array of module definitions that should be deployed together.
  * `cfAppName`- Cloud Foundry application name
  * `cfManifestPath` - Path to manifest file that should be used for deployment
  * `cfHostname` - Hostname of the application. If not specified it will be cfAppName-cfSpaceName (space name formatted to lower case)
  * `cfDomain` - Sets Cloud Foundry domain for this module (overwrites global parameter)
  * `cfEnvVariables` - Array of environment variables that should be set during deployment
  * `deployType` - Sets deployment type for this module (overwrites global parameter)
  * `registerAsServiceBroker` - Should this app be registered as service broker? (cf. <https://github.wdf.sap.corp/xs2/node-sbf>)
  * `serviceBrokerHasSpaceScope` - If `registerAsServiceBroker` is true, this parameter controls if the service broker should be registered with space-scope only.

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|cfApiEndpoint|X|X|X|
|cfCredentialsId|X|X|X|
|cfDomain|X|X|X|
|cfOrg|X|X|X|
|cfSpace|X|X|X|
|deployType|X|X|X|
|dockerImage|X|X|X|
|dockerWorkspace|X|X|X|
|modules|X|X|X|

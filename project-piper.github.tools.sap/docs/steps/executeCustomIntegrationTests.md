# executeCustomIntegrationTests

## Deprecation notice

please use [stage extensibility capabilities](../extensibility.md) instead.

## Description

This step allows you to execute custom integration tests.

!!! note "Custom Exit"
    A custom exit is available which allows you to plug in any kind of integration test script into the pipeline.

## Prerequisites

You have to create a custom script in a separate file and put this into your source code repository, e.g. `.pipeline/integration.groovy`.
The file containing the script has to contain a function in the form:

```groovy
def call(script) {
  echo "Integration Test"
  echo "CF ORG: ${script.globalPipelineEnvironment.getConfigProperty('cfOrg')}"
}
return this;
```

The function has one parameter `script`.
This parameter allows you to consume all global variables of the main Jenkinsfile script.

**Remark**: Within the script you cannot make use of the `stage {}` syntax since the script will be included into the stage *Integration*.

## Example

Pipeline step:

```shell
executeCustomIntegrationTests script: this
```

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|extensionIntegrationTestScript|yes||.pipeline/integration.groovy|

Details:

* `script` defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`globalPipelineEnvironment`](../objects/globalPipelineEnvironment.md) for retrieving e.g. configuration parameters.
* Using `extensionIntegrationTestScript` a custom script can directly be defined and path will not be retrieved from `globalPipelineEnvironment`

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|extensionIntegrationTestScript||X|X|

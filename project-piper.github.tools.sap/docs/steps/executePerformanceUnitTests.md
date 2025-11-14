# executePerformanceUnitTests

## Description

Execute JUnit tests enhanced with [ContiPerf](http://databene.org/contiperf.html) and compare performance measurement against application defined threshold

## Prerequisites

Unit Performance Test: annotated unit test with ContiPerf

## Example

Usage of pipeline step:

```groovy
executePerformanceUnitTests script: this
publishTestResults contiperf: [archive: true], allowUnstableBuilds: false
```

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|dockerCommand|no|`mvn test`||
|dockerImage|no|`docker.wdf.sap.corp:50000/piper/maven`||
|dockerWorkspace|no|`/home/piper`||
|failOnError|no|`false`||
|stashContent|no|<ul><li>`buildDescriptor`</li><li>`tests`</li></ul>||

### Details

* actual test will be executed within executePerformanceUnitTests(), and result will be published via publishTestResults
* With `failOnError` you can define the behavior, in case tests fail. When this is set to `true` test results cannot be recorded using the `publishTestResults` step afterwards.

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|dockerCommand|X|X|X|
|dockerImage|X|X|X|
|dockerWorkspace|X|X|X|
|failOnError|X|X|X|
|stashContent|X|X|X|

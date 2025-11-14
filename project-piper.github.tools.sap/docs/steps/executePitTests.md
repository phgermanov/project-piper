# executePitTests

## Description

This step allows for an easy execution of [PIT mutation tests](http://pitest.org/) for Java projects.

Please be aware that mutation tests take a long time. Thus, it is recommended to not execute this step in each and every pipeline run. Set parameter 'runOnlyScheduled' to 'true' and the step will only be executed in runs started by a timer (e.g. nightly runs triggered from Vulas step).

## Prerequisites

Simply include the PIT maven plugin in section build/plugins of your pom.xml. You can follow the description at <http://pitest.org/quickstart/maven/> or copy and adapt the example below.

```xml
            <plugin>
                <groupId>org.pitest</groupId>
                <artifactId>pitest-maven</artifactId>
                <version>1.4.0</version>
                <configuration>
                    <threads>4</threads>
                    <excludedClasses>
                        <param>com.sap.my-project.generated.*</param>
                    </excludedClasses>
                    <outputFormats>HTML,XML</outputFormats>
                </configuration>
            </plugin>
```

Please do not change the report output directory as the step is currently only working with default folders.

## Example

```groovy
executePitTests script: this, coverageThreshold: 80.0, mutationThreshold: 80.0, buildDescriptorFile: 'app/pom.xml', runOnlyScheduled: false
```

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|buildDescriptorFile|no|`./pom.xml`||
|dockerImage|no|`docker.wdf.sap.corp:50000/piper/maven`||
|dockerWorkspace|no|`/home/piper`||
|coverageThreshold|no|`50`|between 0 and 100|
|mutationThreshold|no|`50`|between 0 and 100|
|runOnlyScheduled|no|`false`|true or false|
|stashContent|no|<ul><li>`buildDescriptor`</li><li>`tests`</li></ul>||

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|buildDescriptorFile|X|X|X|
|dockerImage|X|X|X|
|dockerWorkspace|X|X|X|
|coverageThreshold|X|X|X|
|mutationThreshold|X|X|X|
|runOnlyScheduled|X|X|X|
|stashContent|X|X|X|

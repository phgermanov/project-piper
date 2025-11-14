# executePerformanceSingleUserTests

## Description

Execute selenium based test with docker to simulate single user in browser.  It is possible either use standard browser API or SUPA to retrieve performance metrics. Details can be found [Single User Performance Test](https://wiki.wdf.sap.corp/wiki/display/IndCldHCP/Performance+Test+Tools#PerformanceTestTools-SingleUserPerfTestTool).

## Prerequisites

Selenium script integrated with Browser API or SUPA need to be developed

## Example

Usage of pipeline step:

```groovy
executePerformanceSingleUserTests script: this
archiveArtifacts 'target/supa/**'
archiveArtifacts '**/*.csv'
publishTestResults supa: [archive: true], allowUnstableBuilds: false
```

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|dockerCommand|no|testTool=`supa`: `/bin/bash /opt/selenium/start_mvn test -Dtest=${config.testCase} -DargLine='-DPerfTestUrl=${config.testServerUrl}'`<br />testTool=`browserApi`: `/bin/bash /opt/selenium/start_mvn test -Dtest=${config.testCase} -DargLine='-DCSVFile=${config.testResultFile} -DPerfTestUrl=${config.testServerUrl}'`<br />||
|dockerEnvVars|no|`[:]`||
|dockerImage|no|`docker.wdf.sap.corp:50000/piper/performance`||
|dockerVolumeBind|no|`[/dev/shm:/dev/shm]`||
|dockerWorkspace|no|||
|failOnError|no|`false`||
|gitBranch|no|||
|gitSshKeyCredentialsId|no|``||
|stashContent|no|`[]`||
|testCase|yes|||
|testRepository|yes|||
|testResultFile|no|`browserApiResult.csv`||
|testServerUrl|yes|||
|testTool|no|`supa`|<ul><li>`'supa'`: measuring performance with SUPA</li><li>`'browserApi'`: measuring performance with standard BrowserAPI</li></ul> |

### Details

* Set runPerformanceSUT to true to enable Single User Test
* Configure all perfSUT* parameters per your scenario. e.g.:

```properties
perfSUTRepository=git@github.wdf.sap.corp:IndustryCloudFoundation/SUPT_Selenium.git
perfSUTestUrl=https://icf-performance-currency-exchange-rates-web.cfapps.sap.hana.ondemand.com/launchpage/index.html
perfSUTResultCSV=testCurrencyConversion.csv
perfSUTTestCase=WDTest#testCurrencyConversion
```

* Actual test will be executed within executePerformanceSingleUserTests()
* With `failOnError` you can define the behavior, in case tests fail. When this is set to `true` test results cannot be recorded using the `publishTestResults` step afterwards.
* In case the test implementation is stored in a different repository than the code itself, you can define the repository containing the tests using parameter `testRepository` and if required `gitBranch` (for a different branch than master) and `gitSshKeyCredentialsId` (for protected repositories). For protected repositories the `testRepository` needs to contain the ssh git url.

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|dockerCommand|X|X|X|
|dockerEnvVars|X|X|X|
|dockerImage|X|X|X|
|dockerVolumeBind|X|X|X|
|dockerWorkspace|X|X|X|
|failOnError|X|X|X|
|gitBranch|X|X|X|
|gitSshKeyCredentialsId|X|X|X|
|stashContent|X|X|X|
|testCase|X|X|X|
|testRepository|X|X|X|
|testResultFile|X|X|X|
|testServerUrl|X|X|X|
|testTool|X|X|X|

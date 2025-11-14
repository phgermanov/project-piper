# globalPipelineEnvironment

## Description

This object is used to store pipeline-related information which can be accessed throughout the whole pipeline.

Following information is stored in the object using setter methods and can be retrieved with getter methods:

* Configuration information which can be read from a configuration properties file (see [setupPipelineEnvironment](../steps/setupPipelineEnvironment.md) for details.
* Git-related information like branch, sshUrl, HttpsUrl, ...
* GitHub-related information like organization, repository, ...
* Properties retrieved from xMake build
* Nexus information
* Artifact version
* CommitId
* Influx data
* ...

In addition it contains some convenience methods around the pipeline environment like:

* `setGithubStatistics()`: retrieve statistics from GitHub and store in the `globalPipelineEnvironment`
* `addError()`: set the build status to error

## Usage

Within the Jenkinsfile it can just be used like:

```groovy
globalPipelineEnvironment.setGithubOrg('MyOrg')
```

If you want to use it within a pipeline step the Jenkinsfile context (`this`) has to be passed to the step.
This is for example done in the following manner:

```groovy
sendNotificationMail script: this
```

# sapJiraWriteIssueList

## Description

This task will assist you to persist a list of Jira issues (according to some filter criteria) into an html file.

!!! hint
    One usage example for this step is to select all open prio 1/2 bugs related to a dedicated delivery/version.

    With this step a bug report can be created - ideally showing that there are no open bugs. This report can then be uploaded automatically to Sirius using the step [siriusUploadDocument](siriusUploadDocument.md).

## Prerequisites

You need to have a user for accessing Jira.

Its username and password have to be maintained in the Jenkins credentials store.

**Recommendation is to use a technical user instead of your SAP GLOBAL account:**

In order to get a technical user please create an issue in the [JIRAADMIN project](https://sapjira.wdf.sap.corp/secure/IssueHierarchyOverview!default.jspa?projectKey=JIRAADMIN&displayMode=hierarchy).
Please use component _technical user_ and take one of the previous requests as sample.

## Example

Usage of pipeline step:

```groovy
//Example 1: Using Jira filter id
sapJiraWriteIssueList script: this,
                        credentialsId: 'myJiraCredentialsId',
                        filterId: 86133,
                        reportTitle: 'This is my report title',
                        fileName: 'example1.html'

//Example 2: Using Jira jql
sapJiraWriteIssueList script: this,
                        credentialsId: 'myJiraCredentialsId',
                        jql: 'project = SPENS1SHIPMENT AND issuetype = "User Story" AND fixVersion = "D_SAP Sports One 1.0 - 1708"',
                        reportTitle: 'This is my report title',
                        fileName: 'example2.html'
```

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|fileName|no|`issueList.html`||
|jiraApiUrl|no|`https://sapjira.wdf.sap.corp/rest/api/2`||
|jiraCredentialsId|yes|||
|jiraFilterId|no|||
|jql|yes|||
|reportTitle|no|`Jira Issue List`||
|style|no|||

### Details

* `script` defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`globalPipelineEnvironment`](../objects/globalPipelineEnvironment.md) for retrieving e.g. configuration parameters.
* Jira credentials id in Jenkins is defined with `jiraCredentialsId`.
* With the parameter `jql` a standard Jira query can be passed. Example: `project = SPENS1SHIPMENT AND issuetype = "User Story" AND fixVersion = "D_SAP Sports One 1.0 - 1708"`.
* Alternatively to `jql` you could also define the `jiraFilterId` of a filter previously created in the Jira system.
* The Jira API url is defined with `jiraApiUrl`. You may want to use a different Jira system than the SAP default one.
* With `reportTitle` you define the title of the generated html document as well as the headline of the document
* The generated html document is stored with the name `fileName`
* For css styling of the document a custom style can be used. The custom css is passed as text via the `style` parameter. The default style can be found in the file [piper.css](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/blob/master/resources/piper.css)

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|fileName|X|X|X|
|jiraApiUrl|X|X|X|
|jiraCredentialsId|X|X|X|
|jiraFilterId|X|X|X|
|jql|X|X|X|
|reportTitle|X|X|X|
|style|X|X|X|

## Example output for a bug report using following query

```text
project = SPENS1SHIPMENT AND issuetype = bug AND status not in (completed, inactive) AND (priority = High OR priority = "Very High") AND (labels != not_rel_ship OR labels is EMPTY)
```

![IssueListExample](../images/issueList.png)

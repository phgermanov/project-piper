# Traceability

## Description

In order to fulfill requirements [FC-2](https://wiki.wdf.sap.corp/wiki/display/FunctionalCorrectness/FC-2) (**Corporate**) and [FC-3](https://wiki.wdf.sap.corp/wiki/display/FunctionalCorrectness/FC-3) you're requested to show,
which new or existing functionalities have been tested with which test cases and that these test cases have been executed
successfully (in this context we use the term “traceability").

The task [sapCreateTraceabilityReport](../steps/sapCreateTraceabilityReport.md)
can create a traceability report based on automated test results which are available for the current Jenkins pipeline
run.

!!! warning "Please note:"
    Manual test cases are not covered in the generated test reports

## Alternatives

For projects not using Piper, following alternatives exist in order to report automated tests on [FC-2](https://wiki.wdf.sap.corp/wiki/display/FunctionalCorrectness/FC-2) and [FC-3](https://wiki.wdf.sap.corp/wiki/display/FunctionalCorrectness/FC-3):

- [Continuous Traceability Monitor (CTM)](https://github.com/SAP/quality-continuous-traceability-monitor)
  An open source tool, which is platform and CI independent (written in [go](https://golang.org/))
  which works on the same input (Requirement-, Deliveryfile) and output files (HTML, JSON reports) as the task [sapCreateTraceabilityReport](../steps/sapCreateTraceabilityReport.md)

  !!! warning "Please note:"
      Currently the tool is not supported by any team in SAP, as the Cloud Quality Coaching team that created it no longer exists.

----

## Requirement file (_requirement.mapping_)

List all your **existing functionalities** in this ```json``` file in order to create [FC-3 (Traceability)](https://wiki.wdf.sap.corp/wiki/display/FunctionalCorrectness/FC-3)
reports (HTML and JSON) which will show that regression tests have been executed.
The created reports should be uploaded into the respective Sirius delivery e.g. using the [siriusUploadDocument](../steps/siriusUploadDocument.md)
step.

!!! hint "Cumulus usage"
    Thanks to the [Piper-Cumulus integration](../stages/cumulus-integration.md), once you have connected your pipelines to your Sirius deliveries, you no longer need to manually upload the _requirement.mapping_ file in Sirius.

!!! hint "Automatic generation of requirement file"
    The _requirement.mapping_ file can automatically be created. There are parsers available outside the Piper-scope:

    * [Jira mapping using annotations in Java test code](https://github.tools.sap/P4TEAM/cumulus-jiralinking#java-annotations)
    * [Jira mapping based on comments for Jasmine, OPA5, qUnit](https://github.tools.sap/P4TEAM/cumulus-jiralinking#comment-parser)

### Requirement file structure

- The full automated test class or method name. If a single test method should be referred to, add
brackets ```()``` at the end.

```json
"source_reference": "com.sap.suite.cloud.foundation.currencyconversion.CurrencyConversionTest.convert()"
```

- A list of new backlog items maintained in Jira which map to the given test class or method

```json
"jira_keys": [
    "JENKINSBCKLG-4"
]
```

- A list of new backlog items maintained in GitHub which map to the given test class or method

```json
"github_keys": [
    "org/repo#3",
    "org/repo#1"
]
```

- **Optional** A (to the repository root) relative path  to the file in which the test class or method is located.
  Could be used to add a source code link in the created reports

```json
"filepath": "src/test/com/sap/bulletinboard/ads/resources/CurrencyConversionTest.java"
```

### Example of _requirement.mapping_

```json
[
  {
    "source_reference": "com.sap.suite.cloud.foundation.currencyconversion.CurrencyConversionTest.convert()",
    "jira_keys": [
      "JENKINSBCKLG-4"
    ],
    "github_keys": [
      "org/repo#5"
    ],
    "filepath": "src/test/com/sap/bulletinboard/ads/resources/CurrencyConversionTest.java"
  },
  {
    "source_reference": "com.sap.suite.cloud.foundation.currencyconversion.CurrencyConversionUnitTest",
    "jira_keys": [
      "JENKINSBCKLG-4",
      "JENKINSBCKLG-3"
    ],
    "github_keys": [
      "org/repo#3",
      "org/repo#1"
    ]
  }
]
```

## Delivery file (_delivery.mapping_)

List all your **new functionalities**, which are introduced with the given delivery,
in this ```json``` file in order to create [FC-2 (Traceability)](https://wiki.wdf.sap.corp/wiki/display/FunctionalCorrectness/FC-2)
reports (HTML and JSON).
The created reports should be uploaded into the respective Sirius delivery e.g. using the [siriusUploadDocument](../steps/siriusUploadDocument.md)
step.

!!! hint "Cumulus usage"
    If you are using [Cumulus for Traceability](https://wiki.wdf.sap.corp/wiki/x/1psyhQ), you don't need to create a _delivery.mapping_ file.

### Delivery file structure

- The name of the sirius program

```json
"sirius_program": "My sirius program name"
```

- The name of the sirius delivery

```json
"sirius_delivery": "My sirius delivery name"
```

- A list of new backlog items maintained in Jira

```json
  "jira_keys": [
    "MYJIRAPROJECT-1",
    "MYJIRAPROJECT-2",
    "MYJIRAPROJECT-3"
  ]
```

- A list of new backlog items maintained in GitHub

```json
  "github_keys": [
    "myOrg/mySourcecodeRepo#42",
    "myOrg/mySourcecodeRepo#43"
    "myOrg/mySourcecodeRepo#44"
  ]
```

### Example of _delivery.mapping_

```json
{
  "sirius_program": "My sirius program name",
  "sirius_delivery": "My sirius delivery name",
  "jira_keys": [
    "MYJIRAPROJECT-3"
  ],
  "github_keys": [
    "myOrg/mySourcecodeRepo#42",
    "myOrg/mySourcecodeRepo#43"
  ]
}
```

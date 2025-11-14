# ${docGenStepName} ![Jenkins only](https://img.shields.io/badge/-Jenkins%20only-yellowgreen)

## Description

This task will create a traceability report based on the automated test results which are available for the current Jenkins pipeline run.

It will assist you in fulfilling [Functional Correctness product standards FC-2](https://wiki.wdf.sap.corp/wiki/display/FunctionalCorrectness/FC-2) and [Functional Correctness product standard FC-3](https://wiki.wdf.sap.corp/wiki/display/FunctionalCorrectness/FC-3).

As per FC-2 description:

> The implementation of new functionality in a product shall be verified successfully in accordance with the chosen test strategy.
> It shall be possible to show, which new functionality was tested with which test cases and that these test cases have been executed successfully (in this context we use the term “traceability")

The step creates following documents:

1. _piper_traceability_delivery.html_

    This is an HTML report which shows the test status of requirements relevant for a certain delivery.
    The delivery mapping is retrieved from the `delivery.mapping` json file. Further details can be found [here](../others/traceability.md).

    It is therefore the **relevant file for proving tracability** (see [Functional Correctness product standards FC-2](https://wiki.wdf.sap.corp/wiki/display/FunctionalCorrectness/FC-2)) and the report should be uploaded to the respective Sirius delivery, e.g. via the step [siriusUploadDocument](siriusUploadDocument.md)

    !!! hint "Cumulus usage" If you are using [Cumulus for Traceability](https://wiki.wdf.sap.corp/wiki/x/1psyhQ), you get this functionality out-of-the-box.

2. _piper_traceability_delivery.json_

    This is a file in machine-readable json format containing the same information as described in 1.

3. _piper_traceability_all.html_

    This is an HTML report which shows the test status all mapped requirements according to the `requirement.mapping` json file. Further details can be found [here](../others/traceability.md).

    It will help to prove that regression tests have been conducted successfully.

    It therefore assists in automatically documenting compliance to [Functional Correctness product standard FC-3](https://wiki.wdf.sap.corp/wiki/display/FunctionalCorrectness/FC-3).

4. _piper_traceability_all.json_

    This is a file in machine-readable json format containing the same information as described in 3.

## Prerequisites

Following mapping files need to exist in your source code repository

1. `requirement.mapping` json file: see [here](../others/traceability.md) for further details.
2. `delivery.mapping` json file: see [here](../others/traceability.md) for further details.

## Example

Usage of pipeline step:

```groovy hl_lines="3"
junit allowEmptyResults: true, testResults: '**/target/surefire-reports/*.xml'

sapCreateTraceabilityReport script:this, requirementMappingFile: 'jira.mapping'
```

## ${docGenParameters}

## ${docGenConfiguration}

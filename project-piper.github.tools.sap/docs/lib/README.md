# Piper step library

The Piper step library offers you full flexibility by providing a set of general purpose steps.

It is written in an agnostic way so that it can be used on various orchestrators (e.g. Azure DevOps, Jenkins) as well as locally on a command line.
The CLI is built using the go programming language and thus is distributed in a binary file.

The library consists of two parts:

* Universal and core functionality via [Open Source available on github.com](https://github.com/SAP/jenkins-library) using `piper` binary
* SAP-specific parts via [Inner Source on github.wdf.sap.corp](https://github.wdf.sap.corp/ContinuousDelivery/piper-library) using `sap-piper` binary

!!! note "Step execution"

    Find more information in the [step execution documentation](step_execution.md).

# Guidance on how to contribute

All contributions to this project will be released SAP internally AS IS WITHOUT WARRANTY OF ANY KIND.
Potentially we aim also to release parts according to the [MIT license terms](https://github.wdf.sap.corp/ContinuousDelivery/jenkins-pipelines/blob/master/LICENSE.md) to customers/partners.

By submitting a pull request or filing a bug, issue, or feature request, you are agreeing to comply with the waiver.

There are two primary ways to help:
* Using our forum,
* Using the issue tracker, and
* Changing the code-base.

## Using our forum [CURRENTLY OBSOLETE AND NOT ACTIVELY MONITORED]
We have a [Jam group](https://workzone.one.int.sap/site#workzone-home&/groups/fhHMfvtxIARCiFoTLNY9iN/overview_page/ImtWKpg1uhYyutTNzogaXa) including a forum where ideas, solutions and also questions can be raised.
** Please use issue tracker instead **

## Using the issue tracker

Use the issue tracker to suggest feature requests, report bugs, and ask questions. This is also a great way to connect with the developers of the project as well as others who are interested in this solution.

Use the issue tracker to find ways to contribute. Find a bug or a feature, mention in the issue that you will take on that effort, then follow the Changing the code-base guidance below.

## Changing the code-base

Generally speaking, you should fork this repository, make changes in your own fork, and then submit a pull request. All new code should be thoroughly tested to validate implemented features and ensure the presence or absence of defects, accompanied by adequate documentation.

All pipeline library code MUST include automated tests as well as sufficient documentation.

Additionally, your code should adhere to any stylistic and architectural guidelines defined by the project. In the absence of such guidelines, mimic the styles and patterns already present in the existing codebase.

When creating a pull request from a fork, a reviewer from the piper-core team will review the submission. If the pull request appears ready, the reviewer will issue the /mirror command (by commenting on the pull request), creating a mirrored copy of the PR with appropriate permissions, enabling the necessary checks to run.

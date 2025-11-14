# Prerequisites

## Source code management

Host your project on [github.tools.sap](https://github.tools.sap/) (internet-facing) or [github.wdf.sap.corp](https://github.wdf.sap.corp/) (not internet-facing.

You can find a more [in-depth comparison of the GitHub instances here](https://pages.github.tools.sap/github/contact-and-support/instances-comparison/).

!!! note

    An internet-facing GitHub instance like [github.tools.sap](https://github.tools.sap/) is required by some tools like [Azure Pipelines](https://pages.github.tools.sap/azure-pipelines/) and [Deploy with Confidence (DwC)](https://pages.github.tools.sap/deploy-with-confidence/solar-system/).

## Build dependencies - Azure only

SAP Corporate internal common repositories are not accessible during the Build stage execution of the pipeline. But the SAP internet-facing repositories in DMZ are accessible.
If you need particular dependency artifacts built out from other SAP internal build processes, please ensure than those are made available in [SAP repositories in DMZ](https://common.repositories.cloud.sap/).

## Set up orchestrator

!!! tip ""

    === "Jenkins"

         1. [Set up a Jenkins instance](setup_jenkins.md).
         2. Use [Hyperspace Onboarding](https://hyperspace.tools.sap/pipelines)/[Hyperspace Portal](https://pages.github.tools.sap/hyperspace/cicd-setup-documentation/getting-started.html)(see note below) to easily set up a pipeline selecting one of the "Jenkins" templates or build an extensible pipeline. Alternatively, you can manually create a Jenkins build job for your project. For more information, check out the [Jenkins documentation](https://jenkins.io/doc/book/pipeline/multibranch/).

    === "Azure DevOps"

        1. To setup a new pipeline: select one of the "Azure Pipelines" templates, either "Piper Starter" or build an extensible pipeline in [Hyperspace Onboarding](https://hyperspace.tools.sap/pipelines).


    === "GitHub Actions"

        1. With GitHub Actions , one can build, test, and deploy your code right from GitHub, using actions combined with the SUGAR service and Piper General Purpose Pipeline (Piper GPP).[More here](https://pages.github.tools.sap/github/features-and-usecases/features/actions/status/).   Please follow the [Getting Started](https://pages.github.tools.sap/hyperspace/cicd-setup-documentation/getting-started.html) in the **Hyperspace Portal** documentation.

!!! info

    Information on the orchestrator technologies can be found [here](orchestrator_technologies.md).

!!! note

    The easiest way to provision your CI/CD setup is via

    - [Hyperspace Portal](https://pages.github.tools.sap/hyperspace/cicd-setup-documentation/getting-started.html) (new way covering partially CI/CD setup via Jenkins and GitHub Actions)
    - [Hyperspace Onboarding](https://hyperspace.tools.sap/pipelines) (old way covering template-based and extensible CI/CD setup via Jenkins and Azure DevOps)

    In some cases you can transfer your CI/CD setup from Hyperspace Onboarding to Hyperspace Portal. Refer to our [step-by-step guide](https://github.tools.sap/hyper-pipe/portal/blob/master/features_and_use_cases/how_tos/transfer-pipeline-to-portal) to find out which cases are supported.

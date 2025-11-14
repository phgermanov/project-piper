# Set up pipelines

## General Purpose Pipeline

At the moment, the setup of a Piper General Purpose Pipeline depends on the orchestrator you choose. We are working on unifying this process.

!!! tip ""

    === "Jenkins"

         1. [Set up a Jenkins instance](setup_jenkins.md).
         2. For ongoing development projects, go to **Hyperspace Onboarding** and select one of the Jenkins templates or the "Piper Starter". [Learn more](https://hyperspace.tools.sap/docs/features_and_use_cases/use_cases/pipeline-template/#create-pipeline-with-template).
         3. If you are starting a new project, go to the **Hyperspace Portal** and follow the ["Build your code" documentation](https://pages.github.tools.sap/hyperspace/cicd-setup-documentation/use-cases/build-your-code.html).
         3. Alternatively, you can manually create a Jenkins build job for your project. For more information, check out the [Jenkins documentation](https://jenkins.io/doc/book/pipeline/multibranch/).

    === "Azure Pipelines"

        Go to **Hyperspace Onboarding** and select one of the Azure Pipelines templates or the "Piper Starter". [Learn more](https://hyperspace.tools.sap/docs/features_and_use_cases/use_cases/pipeline-template/#create-pipeline-with-template)

    === "GitHub Actions"

        Go to the **Hyperspace Portal** and follow the ["Build your code" documentation](https://pages.github.tools.sap/hyperspace/cicd-setup-documentation/use-cases/fresh-ci-cd-setup/fresh-build.html#set-up-your-build-pipeline).

## Security and Compliance pipeline (only Github Actions)

- For Github actions the Piper GPP will not contain the Security and Compliance stages as a part of the GPP. The optimized mode of the Pipelines does not apply for Github actions. Github Actions GPP will only contain the following stages:
  - [Init](https://github.tools.sap/project-piper/piper-pipeline-github/blob/main/.github/workflows/init.yml)
  - [Build](https://github.tools.sap/project-piper/piper-pipeline-github/blob/main/.github/workflows/build.yml)
  - [Acceptance](https://github.tools.sap/project-piper/piper-pipeline-github/blob/main/.github/workflows/acceptance.yml)
  - [Performance](https://github.tools.sap/project-piper/piper-pipeline-github/blob/main/.github/workflows/performance.yml)
  - [Integration](https://github.tools.sap/project-piper/piper-pipeline-github/blob/main/.github/workflows/integration.yml)
  - [Promote](https://github.tools.sap/project-piper/piper-pipeline-github/blob/main/.github/workflows/promote.yml)
  - [Release](https://github.tools.sap/project-piper/piper-pipeline-github/blob/main/.github/workflows/release.yml)
  - [Post](https://github.tools.sap/project-piper/piper-pipeline-github/blob/main/.github/workflows/post.yml)

- How must a development team run security and compliance stages in github actions ?
  - Open source security (OSS) scans and PPMS compliance stage
    - Although not a part of the GPP, the OSS and PPMS compliance is available as a [dedicated workflow](https://github.tools.sap/project-piper/piper-pipeline-github/blob/main/.github/workflows/sap-oss-ppms-workflow.yml) . To include the above workflow in your pipeline definition create a new yaml file ``piper-oss-ppms.yaml`` in your github repository at path ``.github/workflows``

        ```yaml
        name: Piper OSS workflow
        on:
            push:
                branches:
                    - 'main'
                    - 'releases/**'
            jobs:
                piper:
                    uses: project-piper/piper-pipeline-github/.github/workflows/sap-oss-ppms-workflow.yml@main
                    secrets: inherit
        ```

  - Can i run the workflow on a schedule as the earlier optimized mode allowed or Can i run the workflow on every pull request ?
      Yes : Github workflows can be triggered on events like below:
      Running the workflow on a schedule

      ```yaml
        name: Piper OSS workflow
        on:
            schedule:
            # * is a special character in YAML so you have to quote this string
            - cron:  '0 0 * * *'
      ```

  - Running the workflow on every pull request

      ```yaml
        name: Piper OSS workflow
        on:
            pull_request:
                types: [opened, reopened]
      ```

      Please refer to [events that trigger workflows](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#about-events-that-trigger-workflows) for all possible triggers.

  - Static Application Security Testing (SAST)
    - Although not a part of the GPP, reusable [CodeQL workflows using GitHub Actions](https://github.wdf.sap.corp/pages/Security-Testing/doc/ghas/producing/#example-2-codeql-piper-github-action-with-cumulus-upload) and [CheckmarxOne workflows using Github Actions](https://github.wdf.sap.corp/pages/Security-Testing/doc/cxone/CICD/#saps-piper-github-action) workflows can be incorporated in your repository
    - Can I run the workflow on a schedule as the earlier optimized mode allowed or Can I run the workflow on every pull request?

       Yes:  Please refer to [events that trigger workflows](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#about-events-that-trigger-workflows) for all possible triggers

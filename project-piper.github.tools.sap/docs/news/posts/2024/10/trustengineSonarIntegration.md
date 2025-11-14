---
date: 2024-10-17
title: 'Credentialless usage of Sonar in pipeline with Hyperspace System Trust'
authors:
  - jordi
categories:
  - General Purpose Pipeline
  - GitHub Actions
  - System Trust
  - Sonar
---

As a first step in making the [*Hyperspace System Trust*](https://jira.tools.sap/browse/SP-21267) concept publicly available, the Vault, Sonar and Piper teams have been working on the integration into Piper on **GitHub Actions** for the **Sonar** step.
This means that it won't be necessary anymore to maintain Sonar credentials in Vault.
The functionality is still being tested and will be generally available soon.

<!-- more -->

Up to now, it is necessary to maintain a Sonar token in Vault that provides access to a pipeline's Sonar project.
This token would expire eventually, resulting in errors in pipelines and manual effort to get them up and running again.

The *Hyperspace System Trust* removes this burden by dynamically generating a token for your Sonar project during the pipeline run.
Piper will request a Sonar token from the *Hyperspace System Trust* **when it is not available** in Vault.
The functionality will only available on GitHub Actions for now.

## 📢 Do I need to do something?

No, the existing way of authenticating to Sonar will work as before.

However, there are a few requirements for using the new functionality:

- Your Sonar project needs to be connected to your pipeline on Hyperspace Onboarding (extensible pipeline only) or Portal (work is in progress)
- Using Piper's general purpose pipeline (GPP) or Build stage
- Or alternatively add the changes to your custom workflow (see the Piper GPP PR under "Learn more")
- The Sonar token must not be available in Vault (on any level - project or group)
- The workflow from which you call Piper's reusable workflow(s), needs to give them the permissions as shown in the example underneath

```yaml
name: Piper
on:
  workflow_dispatch:
jobs:
  piper:
    uses: project-piper/piper-pipeline-github/.github/workflows/sap-piper-workflow.yml@v1.17.0
    secrets: inherit
    permissions:
      contents: read
      id-token: write
```

## 💡 What do I need to know?

!!! note "Beta phase"

    The *Hyperspace System Trust* is still in **testing phase**

Existing pipelines will not be affected.
You can still use Vault to store your own Sonar token, as it is still taken from there by default.
New pipelines that are created in Hyperspace Portal will eventually rely on the *Hyperspace System Trust* by default.

## ➡️ What's next?

There are more Piper steps for which the *Hyperspace System Trust* will be enabled in the future, such as the Cumulus and Checkmarx One steps, further reducing the need to manually maintain credentials!
Other orchestrators may also be supported in the future.

## 📖 Learn more

- [PR for the integration into the Piper GPP](https://github.tools.sap/project-piper/piper-pipeline-github/pull/195)
- [sonarExecuteScan token documentation - see the resource references](../../../../steps/sonarExecuteScan.md#token)

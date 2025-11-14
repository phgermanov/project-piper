---
date: 2025-04-02
title: System Trust integration now available on Azure DevOps
authors:
  - max
categories:
  - General Purpose Pipeline
  - Azure DevOps
  - System Trust
---

In October, we brought [news about](../../2024/10/trustengineSonarIntegration.md) the [System Trust](https://pages.github.tools.sap/system-trust/) becoming available on GitHub Actions, for credentialless access to tools that are used in Piper (i.e. Artifactory, Cumulus and Sonar).
It is now also available for Piper pipelines that run on Azure DevOps, thanks to collaboration between the Trust and Piper teams.

<!-- more -->

## 📢 Do I need to do something?

System Trust only works with pipelines that have been configured on [Hyperspace Portal](https://portal.hyperspace.tools.sap/) or [Hyperspace Onboarding](https://hyperspace.tools.sap/pipelines) (as extensible pipeline).
Template based pipelines from Hyperspace Onboarding are supported for Artifactory only.
If that is not the case for your pipeline, then you can simply add your existing repository on there (preferably on Portal!).

## 💡 What do I need to know?

In case if you are using GPP, you are on the safe side.
The feature is available in the next version of `piper-library` (1.357.0) version.
Additionally, the next version of `piper-azure-task` (piper@1.30.3) is needed.

System Trust currently works with the sapCumulusUpload, sonarExecuteScan and sapCollectInsights steps.
Azure DevOps is now able to get System Trust token for Artifactory and perform `docker login` with it to download images.
In the case of Sonar, you currently have to remove the token from Vault first or by setting `skipVault: true` in the step configuration..

The logs indicate that System Trust is being used as follows: `Getting 'token' from System Trust` (depending on the exact parameter name).

## ➡️ What's next?

The System Trust will also become available for other Piper steps in the near future.

## 📖 Learn more

- [System Trust documentation](https://pages.github.tools.sap/system-trust/)

---
date: 2025-02-11
title: System Trust integration now available on JaaS
authors:
  - jordi
categories:
  - General Purpose Pipeline
  - Jenkins
  - JaaS
  - System Trust
---

In October, we brought [news about](../../2024/10/trustengineSonarIntegration.md) the [System Trust](https://pages.github.tools.sap/system-trust/) becoming available on GitHub Actions, for credentialless access to tools that are used in Piper (i.e. Cumulus and Sonar).
It is now also available for Piper pipelines that run on Jenkins as a Service, thanks to collaboration between the Trust and Piper teams.

<!-- more -->

## 📢 Do I need to do something?

System Trust only works with pipelines that have been configured on [Hyperspace Portal](https://portal.hyperspace.tools.sap/) or [Hyperspace Onboarding](https://hyperspace.tools.sap/pipelines) (as extensible pipeline).
If that is not the case for your pipeline, then you can simply add your existing repository on there (preferably on Portal!).

## 💡 What do I need to know?

The feature is available in the next version of `piper-library` (1.357.0), or the `master` version.
Additionally, for Sonar, the next version of `jenkins-library` (v1.426.0, or`master`) is needed.

System Trust currently works with the sapCumulusUpload, sonarExecuteScan and sapCollectInsights steps.
In the case of Sonar, you currently have to remove the token from Vault first.

The logs indicate that System Trust is being used as follows: `Getting 'token' from System Trust` (depending on the exact parameter name).

## ➡️ What's next?

The System Trust will also become available for Piper on Azure DevOps in the near future.

## 📖 Learn more

- [System Trust documentation](https://pages.github.tools.sap/system-trust/)

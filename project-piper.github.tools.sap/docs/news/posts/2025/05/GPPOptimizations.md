---
date: 2025-05-05
title: GPP Optimizations with OSC and SAST
authors:
  - oliver
categories:
  - General Purpose Pipeline
#tags:
#  - ...
---

With the advent of support for GitHub Actions (GHA) from within Hyperspace, and the provision of the [Hyperspace Open-Source Compliance service (OSC)](https://pages.github.tools.sap/hyperspace/academy/services/osc/) we took the next step in simplifying the [Piper general purpose pipelines (GPP)](../../../../stages/README.md#general-purpose-pipeline), to make them faster, and more robust and focused.

<!-- more -->

Simplification will address the areas Security, IPScan and PPMS.

## 📢 Do I need to do something?

No action required from you when setting up new pipelines via the Hyperspace Portal:

Open-source checks using OSC and the execution of Static Application Security Testing (SAST) scans no longer run inside the GPP, making the need both to configure tools configurations as well as PPMS Free and open-source software (FOSS) updates obsolete.

## 💡 What do I need to know?

The new way of setting up OSC is already available in the [Hyperspace Portal](https://portal.hyperspace.tools.sap/home). And we expect to enable the SAST checks in the Hyperspace Portal in Q2/2025:

- [Open-Source Compliance Service](https://pages.github.tools.sap/hyperspace/cicd-setup-documentation/managed-services/validate/osc.html)
- [GitHub Advanced Security (SAST)](https://pages.github.tools.sap/hyperspace/cicd-setup-documentation/managed-services/validate/ghas.html)
- [Checkmarx ONE (SAST)](https://pages.github.tools.sap/hyperspace/cicd-setup-documentation/managed-services/validate/cxone.html)

Running dedicated GHA workflows is also our recommendation for any custom automation going forward.

We are aware that this new optimization may mean your team has to deal with two systems in parallel if your build/GPP pipeline runs on Jenkins or Azure whereas SAST leverages GHA. However, we strongly believe that the benefits will outweigh the possible inconvenience in cases where you use native GHA capabilities for SAST: you get direct feedback from the system you are interacting with while coding, especially when working with Pull-Requests.

## ➡️ What's next?

This new way of optimizing pipeline runtimes and development flow will be the default for any new automation we will provide going forward:
Our primary choice will be to provide services similar to OSC while the backup choice will be to provide GHA workflows which are separated from the GPP.

## 📖 Learn more

- [Pipeline optimizations documentation](https://pages.github.tools.sap/project-piper/stages/optimizations/#optimizations-of-the-general-purpose-pipeline-with-respect-to-open-source-compliance-and-security)

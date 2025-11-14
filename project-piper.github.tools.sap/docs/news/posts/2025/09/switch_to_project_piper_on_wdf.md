---
date: 2025-09-04
title: GitHub GPP Repository Migration to project-piper Organization
author: Valentin Uchkunev
categories:
  - Github Actions
  - General Purpose Pipeline
tags:
 - GPP
 - GHA
 - GithubActions
---

In order to reduce maintenance overhead and discrepancies between Github Tools and Github WDF we decided to use the same structure for our organization and repository.

Current state of GH GPP:

- Tools: [https://github.tools.sap/project-piper/piper-pipeline-github](https://github.tools.sap/project-piper/piper-pipeline-github)
- WDF: [https://github.wdf.sap.corp/ContinuousDelivery/piper-pipeline-github](https://github.wdf.sap.corp/ContinuousDelivery/piper-pipeline-github)

Future state:

- Tools: [https://github.tools.sap/project-piper/piper-pipeline-github](https://github.tools.sap/project-piper/piper-pipeline-github)
- WDF: [https://github.wdf.sap.corp/project-piper/piper-pipeline-github](https://github.wdf.sap.corp/project-piper/piper-pipeline-github)

<!-- more -->

## 📢 Do I need to do something?

In the coming days Hyperspace Portal will open a PR in your repository for the necessary changes which you have to merge.

If it doesn't, you will manually need to change your `.github/workflows/piper.yml`
and change the line

`uses: "ContinuousDelivery/piper-pipeline-github/.github/workflows/sap-piper-workflow.yml@main"`

to

`uses: "project-piper/piper-pipeline-github/.github/workflows/sap-piper-workflow.yml@main"`

## 💡 What do I need to know?

ContinuousDelivery organisation will be archived and it will stop receiving updates.

This only affects users of WDF Github instance ([https://github.wdf.sap.corp](https://github.wdf.sap.corp))

## ➡️ What's next?

Wait for PR from Hyperspace Portal or manually update your `.github/workflows/piper.yml`

If there are any issues with your pipelines after the change please raise a [SNOW ticket](https://itsm.services.sap/sp?id=sc_cat_item&sys_id=703f22d51b3b441020c8fddacd4bcbe2&service_offering=455f43371b487410341e11739b4bcb67)

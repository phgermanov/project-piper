---
date: 2025-02-14
title: Open source Piper sunset for Azure pipelines and migration steps
authors:
- gulomjon
categories:
  - Custom pipelines
  - Azure DevOps
---

!!! warning "The changes described below are no longer a breaking change"
    Our implementation has been reworked and merged, and we do not anticipate any breaking changes.
    This means that currently no actions are required from user. If you encounter an issue, please reach out to us via [ISV-DL-CICD-PIPELINE-PIPER](https://itsm.services.sap/sp?id=sc_cat_item&sys_id=703f22d51b3b441020c8fddacd4bcbe2&service_offering=455f43371b487410341e11739b4bcb67) channel.

Our team is moving forward with the sunset of the open-source binary and starting this process from Azure DevOps pipelines. Starting with version [v1.21.0](https://github.tools.sap/project-piper/piper-pipeline-azure/releases) of **Azure GPP**, all Piper steps will be executed in a single binary. Starting with version [1.26.0](https://github.tools.sap/project-piper/piper-azure-task/releases) of **Piper Azure task**, there will be some changes that you should pay attention to if you are calling Piper task in your custom pipelines.

<!-- more -->

## 💡 What do I need to know?

Our changes are mostly internal and you shouldn't worry about them unless you have custom Azure pipeline where you call [Piper Azure task](https://github.tools.sap/project-piper/piper-azure-task#usage). If you have such pipeline, please read the next section.

## 📢 Do I need to do something?

**If you are using GPP without stage extensions (pre and post steps)**, you don't need to take any actions - our team has already migrated to new changes in the GPP template and Azure task.

**If you are using GPP with stage extensions**, after the release of the mentioned Azure task version above, make sure to remove the following parameters from the [Piper Azure task call](https://github.tools.sap/project-piper/piper-azure-task#usage), if you have them: *getSapPiperLatestVersion, piperVersion,* and *sapPiperVersion*.

**If you have a custom pipeline and you reuse Piper's [Init stage](https://github.tools.sap/project-piper/piper-pipeline-azure/blob/main/stages/init.yml)** in the beginning of your pipeline, do the following.
First of all, to avoid any breakages when our release is published, pin down the exact version of the Piper Azure task and GPP:

```yaml

# For Piper GPP's stages reused in your custom pipeline, like Init, Build and etc.
resources:
  repositories:
    - repository: piper-pipeline-azure
      ...
      name: project-piper/piper-pipeline-azure
      ref: refs/tags/v1.18.0  (or the latest available version before March 3rd)

# For Piper Azure task calls in your custom stages
- task: piper@1.25.0  (or the latest available version before March 3rd)

```

After our releases are published, adapt your custom pipelines by following these steps:

- Switch Piper Azure task and GPP back to the latest version:

```yaml

# GPP
ref: refs/tags/v1.21.0  # or remove this parameter to use 'main' branch (not recommended)
# Azure task
- task: piper@1.26.0  # or piper@1 to always use latest major version

```

- If you've set the Azure variable `hyperspace.sappiper.binaryname: sap-piper` in your template, please remove it, as the binary name is handled internally and this variable is now intended to be used only for testing.

- Make sure that further stages (which call Piper's Azure task) after our `Init` stage read cached binary version from proper source. Your stage must have `binaryVersion` and `osBinaryVersion` variable, which values come from Init stage's outputs `fetch_version.binaryVersion` and `fetch_version.osBinaryVersion` respectively.  See the example [here](https://github.tools.sap/project-piper/piper-pipeline-azure/blob/71211ea959db5f8255b29d51c2a3ff8fc53520d3/stages/build.yml#L52-L53).
`binaryVersion` and `osBinaryVersion` are replacement for `piperCacheVersion` and `sapPiperCacheVersion` variables are used by Azure task to fetch binaries from tool cache.

- If you have **Cache@1** Azure task for caching Piper binary, adjust the keys and variables pointing to `binaryVersion` and `osBinaryVersion` variables. Please refer to [this](https://github.tools.sap/project-piper/piper-pipeline-azure/blob/71211ea959db5f8255b29d51c2a3ff8fc53520d3/stages/build.yml#L69-L86) example.

- Test the changes.

**If you have a custom pipeline and you do not reuse Piper's [Init stage](https://github.tools.sap/project-piper/piper-pipeline-azure/blob/main/stages/init.yml)** at the beginning of your pipeline, it means that you may have different binary caching logic or a modified version of Piper's Init stage. First, please refer to the steps above and check if there are things that are suitable for your pipeline. Then, refer to our Init stage template and [this PR](https://github.tools.sap/project-piper/piper-pipeline-azure/pull/359/files) for examples and inspiration to adapt your pipelines to the upcoming changes.

## ➡️ What's next?

We are planning to publish the mentioned changes **on March 3, 2025**. We kindly ask affected users to pin down versions or adapt your pipelines accordingly before that date.

## 📖 Learn more

[Azure GPP PR](https://github.tools.sap/project-piper/piper-pipeline-azure)
[Piper's Azure task PR](https://github.tools.sap/project-piper/piper-azure-task)

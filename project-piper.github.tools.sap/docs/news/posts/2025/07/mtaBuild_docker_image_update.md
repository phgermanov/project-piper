---
date: 2025-07-30
title: Docker image update for mtaBuild
authors:
  - petko
categories:
  - updates
tags:
  - docker image, mtaBuild
---

We will be changing the default docker image for the mtaBuild step in a future piper release (18.08.2025). The old image ('devxci/mbtci-java11-node14') will be replaced by the new one 'devxci/mbtci-java21-node22'. The reasoning behind the change is that older versions of Java and Node are causing security concerns of using such images in the default config.
<!-- more -->

## 📢 Do I need to do something?

If your workflow is using the mtaBuild step, please check your code for Java 21 and Node 22 compatibility. If you are using the default docker image from the step and you still want to use the current one, you need to add dockerImage: 'devxci/mbtci-java11-node14' in your mtaBuild step config.

Example (in the configuration file found at '.pipeline/config.yml'):

```yaml
#.pipeline/config.yml
steps:
  mtaBuild:
    dockerImage: 'devxci/mbtci-java11-node14'
```

## 💡 What do I need to know?

Please note that the update could bring breaking changes to your pipeline flow and action could be needed from the user's side.

## ➡️ What's next?

The update will be available on 18.08.2025. Please make sure to test your pipelines before the update. If another MTA docker image is better suited for your use case, you can specify it in the mtaBuild step configuration.

Other available MTA [docker images](https://hub.docker.com/u/devxci).

## 📖 Learn more

If you are interested in checking which default docker image is used for the mtaBuild step, this information is easily accessible via the official [documentation](https://pages.github.tools.sap/project-piper/steps/mtaBuild/#dockerimage).

For issues or support regarding the mta build tool or its docker image, please refer to the [official repository](https://github.com/SAP/cloud-mta-build-tool).

Issues can be reported directly in the repository's issue tracker.

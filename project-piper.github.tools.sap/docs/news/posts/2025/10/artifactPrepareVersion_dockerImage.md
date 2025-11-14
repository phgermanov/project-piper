---
date: 2025-10-14
title: Allow overriding dockerImage in artifactPrepareVersion step
authors:
 - philip
categories:
 - General Purpose Pipeline
tags:
 - GPP
 - artifactPrepareVersion
---

The `artifactPrepareVersion` step in GitHub Actions GPP always used a fixed Docker image and ignored the `dockerImage` parameter. Following Piper release on 14.10.2025,it is possible for GitHub Actions users to override the default Docker image by specifying a custom image through the `dockerImage` parameter. This enhancement provides greater flexibility, allowing users to utilize different Docker images as needed for their specific use cases.

<!-- more -->

## 📢 Do I need to do something?

If you haven't already set any parameters for the `artifactPrepareVersion` step in your '.pipeline/config.yml' file, no action is required. The step will continue to use the default Docker image (`maven:3.8.6-jdk-8` for Maven and CAP build tools).

If you have provided a custom Docker image via the `dockerImage` parameter, please ensure that the specified image meets the following requirements:

- The image must contain the necessary build tools for your project (e.g., Maven, npm, gradle, etc.)
- For Maven-based projects (including CAP), the image should have Maven and a compatible JDK installed
- The image should have git installed if you're using versioning types that interact with the repository
- Any additional tools or dependencies required by your build process must be available in the image
- The image should be accessible from your CI/CD environment (proper registry authentication if using private registries)

## 💡 What do I need to know?

The `dockerImage` parameter allows you to specify a custom Docker image for the `artifactPrepareVersion` step. This can be particularly useful if you need to use a specific version.

## ➡️ What's next?

We recommend reviewing your pipeline configurations to determine if you need to specify a custom Docker image for the `artifactPrepareVersion` step. If you decide to use a different image, update the `dockerImage` parameter accordingly.

## 📖 Learn more

- [artifactPrepareVersion step documentation](../../../../steps/artifactPrepareVersion.md)
- [Jenkins library PR](https://github.com/SAP/jenkins-library/pull/5499)
- [resources PR](https://github.tools.sap/project-piper/resources/pull/122)
- [GPP PR](https://github.tools.sap/project-piper/piper-pipeline-github/pull/386)

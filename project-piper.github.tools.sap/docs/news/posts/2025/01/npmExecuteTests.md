---
date: 2025-01-03
title: Introduction of npmExecuteTests Step
authors:
  - philip
categories:
  - Steps
tags:
  - npmExecuteTests
  - uiVeri5ExecuteTests
  - karmaExecuteTests
  - wdi5
---

We are excited to announce the creation of a new step, `npmExecuteTests` which aims to replace both `uiVeri5ExecuteTests` and `karmaExecuteTests`. <!-- more -->
The step is currently standalone and it is not part of Piper General Purpose Pipeline. It is designed to be more flexible when choosing a testing tool. Its documentation is available at [npmExecuteTests Documentation](../../../../steps/npmExecuteTests.md).

## 📢 Do I need to do something?

If you are interested in trying out the new `npmExecuteTests` step, you can start integrating it into your pipelines. Please refer to the documentation for guidance on how to implement the new step.

## 💡 What do I need to know?

The reason for creating this step is:

1. uiVeri5 is [deprecated](https://github.com/SAP/ui5-uiveri5) and replaced by [wdi5](https://github.com/ui5-community/wdi5?tab=readme-ov-file)
2. Karma has been deprecated since April 2022 and it's recommended to migrate to modern test runners like wdi5

See this blog post for [more details](https://community.sap.com/t5/technology-blogs-by-sap/use-wdio-and-wdi5-for-testing-ui5-apps/ba-p/13515863) on wdi5 (addon) / wdio(framework).

The `npmExecuteTests` step is in beta and may undergo changes based on user feedback. It uses `wdi5` by default but could easily be replaced by your tool of choice. Your feedback is valuable to help us improve the step.

The reason this step was created is that `npmExecuteEndToEndTests` step was written in Groovy and it couldn't be extended to all orchestrators without breaking changes.

If you're migrating from uiVeri5, you need to first migrate your project to use wdi5. Here's an example on how to use the new step with wdi5(default):

```yaml
stages:
  - name: Test
    steps:
      - name: npmExecuteTests
        type: npmExecuteTests
        params:
          workingDirectory: "src/test/wdi5"
          runCommand: "./node_modules/.bin/wdio run wdio.conf.js --headless"
```

The above is a very minimal example and you should configure the step based on your own setup by following the [documentation](../../../../steps/npmExecuteTests.md)

If you're using a different test runner like Qmate, you can specify the runCommand and dockerImage like so:

```yaml
stages:
  - name: Test
    steps:
      - name: npmExecuteTests
        type: npmExecuteTests
        params:
          workingDirectory: "tests"
          runCommand: "qmate test.config.js"
          dockerImage: "qmate.int.repositories.cloud.sap/qmate-executor:latest"
```

## ➡️ What's next?

Begin experimenting with the `npmExecuteTests` step in your pipelines. Provide feedback to help us refine and improve the step. Monitor the [documentation](../../../../steps/npmExecuteTests.md) for any updates or additional information.

## 📖 Learn more

- [npmExecuteTests Documentation](../../../../steps/npmExecuteTests.md)
- [Project Piper Documentation](../../../../index.md)

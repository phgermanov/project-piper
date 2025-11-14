---
title: "Piper's release strategy for GitHub Actions as an orchestrator"
date: 2025-10-30
author: "gulomjon"
categories:
  - GitHub Actions
  - General Purpose Pipeline
---

The Piper team is excited to announce updates to the versioning and releasing in the [General Purpose Pipeline](https://github.tools.sap/project-piper/piper-pipeline-github) (GPP) in GitHub Actions and [Piper Action](https://github.com/SAP/project-piper-action). These changes are designed to improve compatibility and consistency of Piper pipelines.

<!-- more -->

## 🏷️ Dynamic version tags

Piper GPP and Piper Action now support dynamic semantic versioning tags such as `v1`, `v1.2`, and `v1.2.3`. We highly recommend using these version tags rather than relying on the `main` branch, as this approach ensures more consistent and stable pipelines; please note that the `main` branch may occasionally include breaking changes that have not yet been released.

You can choose the level of version control that best fits your needs by specifying a major (`v1`), minor (`v1.2`), or full (`v1.2.3`) version tag when consuming GPP or Piper Action.

Example of using the tags in GPP Piper workflow:

```yaml
jobs:
  piper:
    uses: project-piper/piper-pipeline-github/.github/workflows/sap-piper-workflow.yml@v1  # or v1.2 or v1.2.3
```

The same applies if you reuse particular Piper stages in your workflows:

```yaml
  piper-build:
    uses: project-piper/piper-pipeline-github/.github/workflows/build.yml@v1  # or v1.2 or v1.2.3
```

and Piper Action alone:

```yaml
    steps:
      - uses: SAP/project-piper-action@v1  # or v1.2 or v1.2.3
        with:
          step-name: mavenBuild
          ...
```

### Renovate bot for tracking Piper GPP and Piper Action update

If your repository uses the **ospo-renovate GitHub App**, updates to GPP and Piper Action will be automatically detected and pull requests will be created for new releases. This automation works only when you reference GPP or Piper Action using version tags mentioned above, not the `main` branch. If ospo-renovate is not enabled in your repository, please visit [this page](https://github.tools.sap/OSPO/renovate-controller).

### Notice for pipelines registered/created via Hyperspace Portal

In the coming days, Hyperspace Portal will create a PR in your git repository to update the GPP reference from `main` to `v1`. We strongly encourage you to merge this change.

⚠️ From November 17, 2025 onwards 'main' branch might not be stable

## 🛠️ Deprecation of Piper binary version selection parameter

To ensure full compatibility between the Piper binary and the GPP workflow, selecting the Piper binary version using the `piper-version` or `sap-piper-version` inputs will no longer be supported after GPP release on **November 17, 2025**. The Piper binary version will be pinned in the GPP, so you should use the GPP versioning as described above.

### Do I need to do something?

If you currently call `sap-piper-pipeline.yml` workflow or any of the reusable stages (for example, `build.yml`) without `piper-version` and `sap-piper-version` inputs, you can skip reading this section as nothing will change for you.
If you currently use any of `piper-version` and `sap-piper-version` inputs, please remove them. These inputs will remain available **by the end of 2025** to prevent breaking existing pipelines; however, they will no longer affect the actual binary version used.

If you have a valid reason to continue using these inputs — for instance, if you are using an older version of the Piper binary — you should pin your GPP version to the currently published release or the one immediately preceding the release scheduled for **November 17, 2025** (see the examples above for instructions on how to pin the GPP version). These inputs will still be supported in those versions. However, please note that you will eventually need to migrate to the latest version of GPP to benefit from new features and bug fixes.

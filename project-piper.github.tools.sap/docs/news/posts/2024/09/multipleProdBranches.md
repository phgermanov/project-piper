---
date: 2024-09-02
title: Multiple Productive Branches
authors:
  - vyacheslav
categories:
  - General Purpose Pipeline
  - Azure DevOps
  - GitHub Actions
  - Jenkins
---

Piper supports multiple productive branches across all orchestrators. You can specify productive branches using regular expressions, such as `main|dev`, `rel-[0-9]+|develop`, and others. Piper will treat any branch that matches the specified regex as a productive branch.

<!-- more -->

## ❓ Why is this functionality needed?

Piper runs the entire GPP (General Purpose Pipeline) only against productive branches. We’ve introduced this feature in response to cases where end-users want to run the GPP against development branches as well. This allows code to be scanned, tested, and deployed to a preproduction environment before it is merged into the main branch and deployed to production.

## 📢 Do I need to do something?

The feature is available in Jenkins, Azure DevOps, and GitHub Actions. To enable multiple productive branches, specify the `productiveBranch` parameter as a regular expression.

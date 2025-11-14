# Use .github/workflows directory to publish reusable workflows

- Status: accepted

## Context and Problem Statement

Where can we store the reusable workflows files?

## Considered Options

- root directory
- `.github/workflows`

## Decision Outcome

Currently it's [required by GitHub Actions](https://docs.github.com/en/enterprise-server@3.7/actions/using-workflows/reusing-workflows#creating-a-reusable-workflow) that reusable workflows as other workflows are located in the `.github/workflows/` directory.

> ... locate reusable workflows in the .github/workflows directory of a repository. Subdirectories of the workflows directory are not supported.

### Negative Consequences <!-- optional -->

- This can lead to confusion as reusable workflows (content of this repository) and local workflows (workflows to verify the content) are mixed in the same directory.

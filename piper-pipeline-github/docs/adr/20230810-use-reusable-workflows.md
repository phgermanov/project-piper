# use reusable workflows

- Status: accepted
- Deciders: @I557621
- Date: 2022-12-16
- Tags:

Technical Story:

## Context and Problem Statement

How should a general purpose pipeline on GitHub Actions be implemented?

## Decision Drivers <!-- optional -->

- @I557621

## Considered Options

- [Composite Action](https://docs.github.com/en/actions/creating-actions/creating-a-composite-action)
- [Reusable Workflow](https://docs.github.com/en/actions/using-workflows/reusing-workflows#creating-a-reusable-workflow)
- [Starter Workflow](https://docs.github.com/en/enterprise-server@3.7/actions/using-workflows/creating-starter-workflows-for-your-organization)

## Decision Outcome

We haven chosen to use a reusable workflows.

## Pros and Cons of the Options <!-- optional -->

### Composite Action

- Good, because we could nest more levels of composite actions, but that is only the case up to GitHub Enterprise 3.6. You could probably do that by calling a composite action for every stage from a workflow, but I don't see the need to.
- Good, because [argument b]
- Bad, because Composite Actions unlike reusable workflows are not able to run parallel actions
- Bad, because it does not looks as good in the Actions UI because it would show every stage as a step of a job, instead of showing every stage as a job.
- Bad, because the syntax of reusable workflows is nicer.

### Reusable Workflow

[example | description | pointer to more information | …] <!-- optional -->

- Good, because can be referred to from another org/repo.
- Good, because can inherit secrets.
- Bad, because reusable workflows need to be placed within the `.github/workflows` folder, and can not be reference from the root of a repository.
- bad, because it is not possible with the current version of GitHub Enterprise (3.6) to nest workflows on more than 1 level, but with version 3.7, it will be possible to use [4 levels](https://docs.github.com/en/enterprise-server@3.7/actions/using-workflows/reusing-workflows#nesting-reusable-workflows). This means that all stages currently need to be contained within [one workflow](.github/workflows/sap-piper-workflow.yml).
- Good, because soon we can refer to reusable stages from this file, which makes for a neater structure, and also gives users the possibility to use individual stages, instead of the whole general purpose pipeline.

### Starter Workflow

When creating a new workflow in the Actions tab, there is a number of starter workflows that a user can pick. These are defined in [actions/starter-workflows](https://github.tools.sap/actions/starter-workflows).  Maybe it's possible to create a YAML there which refers to our GPP workflow file, but then you nest one more level, and the Actions UI would show jobs as e.g. `piper/piper/Init`.

- Bad, because we do not know who maintains it, but it will surely be difficult to maintain our general purpose workflow from there.
- Bad, because we could create a YAML there which refers to our GPP workflow file, but then you nest one more level, and the Actions UI would show jobs as e.g. `piper/piper/Init`.

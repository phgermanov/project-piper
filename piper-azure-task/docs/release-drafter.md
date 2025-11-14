# Release Drafter Implementation

This document describes the implementation of GitHub's Release Drafter in our project to automate and standardize our release process.

## Overview

Release Drafter automates our release process by:

- Creating and updating release drafts when changes are merged into the main branch
- Categorizing changes based on PR labels
- Managing version numbers according to semantic versioning
- Maintaining a consistent changelog format

## Configuration

The release drafter configuration is defined in [.github/release-drafter.yml](../.github/release-drafter.yml) and includes version management rules, change categories, and auto-labeling patterns. Refer to the configuration file for detailed rules and patterns.

## Integration with Workflows

### Release Creation Process

1. The release-drafter workflow (`release-drafter.yml`) runs on:
   - Push to main branch
   - Pull request events (opened, reopened, synchronize)

2. For each PR:
   - Auto-labels the PR based on configured rules

3. When release is triggered:
   - The make_release workflow picks up the draft release
   - Publishes Piper's Azure task extension to Azure marketplace
   - Updates task versions and documentation

### Version Management Changes

Previously, version management was handled by a manual script. The new approach:

- Uses Release Drafter's version resolution system
- Determines version numbers by PR labels
- For development versions:
  - Uses the same major.minor as the latest release
  - Uses GitHub Actions run number as patch version
- Generates consistent version numbers across:
  - GitHub releases
  - Azure DevOps task versions
  - Package versions

## Usage Guidelines

### Conventional Commits

We follow the [Conventional Commits specification](https://www.conventionalcommits.org) for commit messages. This helps in automatic version management and changelog generation.

### Pull Request Guidelines

1. PR Creation:
   - Use conventional commit messages as described above
   - Labels will be automatically applied based on your branch name, PR title, and changed files
   - For breaking changes, include "BREAKING CHANGE:" in the PR body or ! in your commit (feat!: some change)

2. Before you merge:
   - Verify that the auto-applied labels are correct for your changes
   - Ensure at least one category label is present (enhancement, bug, documentation, dependencies)
   - For breaking changes, double-check that the 'breaking' label is applied

3. Release Process:
   - Draft releases are automatically created and updated
   - Release workflow will handle publishing and version updates automatically

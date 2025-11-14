---
date: 2025-01-29
title: 'Action required (GitHub Actions): disabling of Vault secret ID rotation'
authors:
  - jordi
categories:
  - General Purpose Pipeline
  - GitHub Actions
  - Vault
---

Piper fetches credentials from Vault via [AppRole](https://developer.hashicorp.com/vault/docs/auth/approle) authentication using a [RoleID](https://developer.hashicorp.com/vault/docs/auth/approle#roleid) and a [SecretID](https://developer.hashicorp.com/vault/docs/auth/approle#secretid).
A SecretID expires after 90 days, and has to be rotated within that time limit in order to keep pipelines running.
Up to now, in the general purpose pipeline, this has been handled automatically by the [vaultRotateSecretId](../../../../steps/vaultRotateSecretId.md) step, which would check if the pipeline's SecretID expires within 15 days, and in that case creates a new one in Vault, and update it in the used orchestrator.

However, with GitHub Actions coming to Hyperspace Portal soon, the backend of the Portal will automatically take care of the SecretID rotation for pipelines that are set up from there.
This means that the vaultRotateSecretId step is not needed for those pipelines, and therefore it is now disabled by default in the general purpose pipeline.

<!-- more -->

## 📢 Do I need to do something?

Yes, your SecretID is going to expire if you don't take action, as current GitHub Actions pipelines have not been created using Hyperspace Portal.
This can take somewhere between 90 days (if your SecretID got rotated by Piper today) and 15 days (because Piper would have rotated it if it were valid for <15 days today).

You can either:

- Set up the CI/CD for your GitHub repository on GitHub Actions via the Hyperspace Portal as [described here](https://pages.github.tools.sap/hyperspace/cicd-setup-documentation/use-cases/build-your-code.html). --> Coming shortly!
- Enable the vaultRotateSecretId step in your Piper config (`.pipeline/config.yml`), by adding

        ```
        stages:
          Post:
            vaultRotateSecretId: true
        ```
- Create your own rotation logic, for example through a pipeline [extension](https://github.tools.sap/project-piper/piper-pipeline-github/blob/main/docs/extensibility.md). The previous two options are recommended instead, because these are supported by Hyperspace.

## 💡 I saw this post too late, and now my SecretID expired

You can unblock your pipeline by doing the CI/CD Setup for your repository on Hyperspace Portal.
Alternatively, if you are using a Vault group that was set up in Hyperspace Onboarding, you can find a button there to refresh the SecretID, copy the new one from the `appRoleCredentials` entry in Vault, and update the `PIPER_VAULTAPPROLESECRETID` secret in your GitHub repository secrets.

## 📖 Learn more

Information on how Hyperspace Portal takes care of the SecretID rotation, can be found [over here](https://pages.github.tools.sap/hyperspace/cicd-setup-documentation/managed-services/build/piper.html#auto-rotating-piper-credentials-in-vault).

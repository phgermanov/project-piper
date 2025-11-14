---
date: 2025-09-15
title: Rotating Vault AppRole credentials in pipelines from Hyperspace Portal and Onboarding
authors:
  - gulomjon
categories:
  - General Purpose Pipeline
  - Vault
  - Hyperspace Portal
  - Hyperspace Onboarding
---

Piper retrieves credentials from Vault using [AppRole](https://developer.hashicorp.com/vault/docs/auth/approle) authentication, which relies on a [RoleID](https://developer.hashicorp.com/vault/docs/auth/approle#roleid) and a [SecretID](https://developer.hashicorp.com/vault/docs/auth/approle#secretid).

The SecretID expires after 90 days and must be rotated within that time frame to ensure the pipelines continue to function. In the general-purpose pipeline, this process is handled automatically by the [vaultRotateSecretId](../../../../steps/vaultRotateSecretId.md) step. This step checks whether the pipeline's SecretID is set to expire within 15 days and, if so, creates a new SecretID in Vault and updates it in the orchestrator being used.

Traditionally, in Hyperspace Onboarding template pipelines, Piper handled the rotation of Vault AppRole secrets using the 'vaultRotateSecretId' step.

However, with the introduction of extensible pipelines through Hyperspace Onboarding and the ability to create pipelines via the Hyperspace Portal, secret rotation can now be automated by the backend engine known as AutomatiCD.

Given the shift-down approach in Hyperspace, this automated method is the recommended way forward.

That said, the Piper step is still required in certain scenarios to accommodate existing users. As a result, the responsibility for secret rotation varies depending on how the pipeline is configured (Template, Extensible, or Independent pipeline) and where it resides (Hyperspace Onboarding, Portal, or elsewhere).

<!-- more -->

## 📢 Do I need to do something?

If you receive a PR from the Hyperspace Portal with the `vaultRotateSecretId` step disabled, please proceed to merge it. If no PR has been created, it indicates that Piper is managing the AppRole secret rotation, and no further action is required on your part.

## 🛠️ Troubleshooting

* If your pipeline is managed in the Portal and you rely on it for rotating the SecretID, but the pipeline suddenly fails with an `invalid role or secret ID` error, it likely indicates an issue with SecretID rotation in the Portal. In such cases, please reach out to the Hyperspace support team or create a SNOW ticket for the `ISV-DL-CICD-PROVISIONING` service offering.
* If you rely on the `vaultRotateSecretId` step for rotating the SecretID, and the step fails, please check if Jenkins/Azure credentials are still valid and exist in Vault by configured path. For documentation about the step, please see [vaultRotateSecretId](../../../../steps/vaultRotateSecretId.md).

## 📖 Learn more

[Related Piper news article](https://pages.github.tools.sap/project-piper/news/2025/01/29/action-required-github-actions-disabling-of-vault-secret-id-rotation/)

---
date: 2024-11-14
title: Credentialless usage of Cumulus in pipeline with Hyperspace System Trust
authors:
  - gulomjon
  - philip
categories:
  - General Purpose Pipeline
  - GitHub Actions
  - System Trust
  - Cumulus
---

The `sapCumulusUpload` and `sapCumulusDownload` steps now support a token provided by the *Hyperspace System Trust* for uploading files to Cumulus. This eliminates the need for a JSON key stored in Vault. No action is needed if the JSON key is already configured.

<!-- more -->

## 📢 Do I need to do something?

No immediate action is required if your pipeline already uses a JSON key stored in Vault. The new token will be used automatically if the JSON key is not found.

## 💡 What do I need to know?

Existing pipelines will not be affected.
You can still use Vault to store your own Cumulus JSON key, as it is still taken from there by default.
New pipelines that are created in Hyperspace Portal will eventually rely on the *Hyperspace System Trust* by default.

## ➡️ What's next?

Future updates will extend the *Hyperspace System Trust* to other Piper steps, further reducing the need for manual credential management.

## 📖 Learn more

- [PR for the integration into the Piper GPP](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/5384)
- [sapCumulusUpload token documentation - see the resource references](../../../../steps/sapCumulusUpload.md#token)
- [sapCumulusDownload token documentation - see the resource references](../../../../steps/sapCumulusDownload.md#token)
- [credentialless usage of Sonar in pipeline with Hyperspace System Trust](../10/trustengineSonarIntegration.md)

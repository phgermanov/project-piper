<!-- markdownlint-disable-next-line first-line-h1 -->
## SAP-Specifics

!!! info

   Pipelines created via Hyperspace Portal or Extensible pipelines created via Hyperspace Onboarding don't have to configure this step, since SecretID rotation is handled automatically by the Portal's backend. For such cases Hyperspace Portal creates PR where `vaultRotateSecretId` step is disabled and Piper team recommends to keep it disabled.
   Pipelines not set up using Portal or Extensible pipeline from Onboarding, have to configure credentials (`adoPersonalAccessToken`, `githubToken` or `jenkinsUsername`+`jenkinsUsername`+`jenkinsUrl`) in order to rotate the secret ID automatically.

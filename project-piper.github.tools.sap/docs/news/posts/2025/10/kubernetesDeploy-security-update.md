---
date: 2025-10-15
title: Enhanced Security in kubernetesDeploy Step – CA Certificate Support and Secure Defaults
authors:
  - petko
categories:
  - General Purpose Pipeline
  - updates
tags:
  - kubernetesDeploy, Security, k8s, kubernetes
---

The `kubernetesDeploy` step in the Piper has been significantly enhanced to improve security and flexibility for Kubernetes deployments. The update will go live in the next Piper release. The latest update introduces support for custom CA certificates and makes secure connections the default, reducing the risk of insecure API interactions.
<!-- more -->

## 📢 Do I need to do something?

If you are using corporate k8s clusters (clusters provided by SAP) - no need to make changes.
In case if you are using local k8s cluster you need to update your pipeline configurations to reference the correct CA certificate as needed [(link to docs)](https://pages.github.tools.sap/project-piper/steps/kubernetesDeploy/).
Review the new [demo repo example](https://github.tools.sap/piper-demo/actions-demo-k8s-go/blob/c80afe930d223ee409fbd9b3911fceba041014c6/.pipeline/config.yml#L108) for a simple demo of the new flags included in the step.

## 💡 What do I need to know?

- The new `insecureSkipTLSVerify` property is now set by default to `false`, ensuring that TLS certificate verification is enforced for all Kubernetes API connections. This uses the `--insecure-skip-tls-verify` flag with `kubectl` with the given boolean value from the config.
  This change helps prevent accidental connections to clusters with untrusted or self-signed certificates unless explicitly configured.
- Setting `insecureSkipTLSVerify` to `true` disables certificate validation, which exposes your cluster to man-in-the-middle attacks and allows connections to potentially malicious or untrusted endpoints. This should only be used for testing or in trusted, isolated environments, and is strongly discouraged for production workloads.
- A new `CACertificate` parameter allows users to specify the path to a custom Kubernetes CA certificate which is stored in the Vault.
When provided, the step uses the `--certificate-authority` flag with `kubectl`, enabling secure connections to clusters with self-signed or enterprise CAs.
This is especially useful for custom k8s user clusters and environments requiring custom trust chains.
- Existing pipelines will benefit from improved security with minimal changes required.
Users with custom or self-signed CAs must update their pipeline configuration to use the new `CACertificate` parameter.
The change is backward compatible - if no CA certificate is provided, the step can fall back to the previous behavior with explicit configuration of the `insecureSkipTLSVerify` to `true`.
This will change `--insecure-skip-tls-verify` to `true`, but this is strongly discouraged.

## 📖 Learn more

- [Piper docs](https://pages.github.tools.sap/project-piper/steps/kubernetesDeploy/)
- [Demo repo example](https://github.tools.sap/piper-demo/actions-demo-k8s-go/blob/c80afe930d223ee409fbd9b3911fceba041014c6/.pipeline/config.yml#L108)
- [Kubernetes: Organize Cluster Access Using kubeconfig](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/)

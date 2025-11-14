# Example Configurations

This page shows you some pipeline configuration examples.

As `Jenkinsfile` only following code is required:

```groovy
@Library('piper-lib') _

sapPiperPipeline script: this
```

## Pure Pull-Request Voting

.pipeline/config.yml:

```yml
general:
  buildTool: 'npm'
```

## Using custom defaults

It is possible to use custom defaults as indicated on the section about [Configuration](../configuration.md).

In order to use a custom defaults only a simple extension to the `Jenkinsfile` is required:

```groovy
@Library(['piper-lib', 'myCustomLibrary']) _

sapPiperPipeline script: this, customDefaults: ['myCustomDefaults.yml']
```

## Example Projects

### GitHub Actions
  
- [Npm sample repository](https://github.tools.sap/piper-demo/actions-demo-k8s-node/)  

### Azure Pipelines

- [Gradle sample repository](https://github.tools.sap/piper-demo/azure-demo-k8s-gradle)
- [Go sample repository](https://github.tools.sap/piper-demo/azure-demo-k8s-go)
- [MTA (java + node.js) sample repository](https://github.tools.sap/piper-demo/azure-demo-cf-mta)

You can find more examples in the [piper-demo](https://github.tools.sap/piper-demo/) GitHub organization.

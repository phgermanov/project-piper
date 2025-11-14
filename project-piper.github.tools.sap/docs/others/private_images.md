# Private Docker images

## Description

Many Piper steps run inside Docker containers.
By default, these images are pulled from Docker Hub.
In order to run a step with an image from a private Docker registry (e.g. the Common Repository), in many cases, credentials to that registry must be provided. How this works varies from orchestrator to orchestrator.

## JaaS

JaaS pipelines run in a Kubernetes-based environment and use the [Kubernetes Plugin](https://pages.github.tools.sap/hyperspace/jaas-documentation/features_and_use_cases/plugin_usage/kubernetes_plugin/), which creates a temporary pod in the cluster.
Also, the dockerExecute step internally calls dockerExecuteOnKubernetes if you are running in a Kubernetes environment that does not have the dockerRegistryCredentialsId parameter.

Hence, it is necessary to get help from the JaaS team to create a Kubernetes secret directly in the cluster in your namespace to pull a private Docker image.

**Please follow the steps below**:

1. Raise a ServiceNow incident to the [JaaS team](https://itsm.services.sap/sp?id=sc_cat_item&sys_id=703f22d51b3b441020c8fddacd4bcbe2&sysparm_category=e15706fc0a0a0aa7007fc21e1ab70c2f&service_offering=c55f03371b487410341e11739b4bcbf1) asking them to create a Kubernetes secret in your namespace.

   **Mention the following in the JaaS ticket**:

   1. Your JaaS instance URL

   2. Username, token, and your private registry URL (without https) required to create the secret in the cluster

   3. Request the JaaS team to create the secret as a Docker config JSON and mention that it is needed to pull a private image.

   4. (Optional) Your preferred secret name.

  **Note**:You can pass the username, token, and other confidential information to the JaaS team in any preferred way. If you already have a dockerConfigJSON, then you can send that to the JaaS team directly and ask for the secret. Please check this with the colleagues in the ServiceNow ticket.

  [Official Kubernetes documentation reference](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/#create-a-secret-by-providing-credentials-on-the-command-line) for the creation of a secret if needed.
2. Once the JaaS team creates the secret and you know the secret name,, please pass it to `dockerExecute` or `dockerExecuteOnKubernetes` step directly as a parameter, as follows:

```text
additionalPodProperties: [
                    imagePullSecrets: ['secret-name']
                ]

```

Please also make sure 'ON_K8S' environment variable is set, as mentioned [here](../steps/dockerExecuteOnKubernetes.md#prerequisites).

Assume the name of the secret created was named `docker-secret`:

```text
dockerExecuteOnKubernetes(
                script: this,
                verbose: true,
                dockerImage: 'deploy-releases-hyperspace-docker.common.repositories.cloud.sap/golang:1.0.0-20211123154630_dc0998ed17e94b0e5800d12b0e5273eb28373654',
                dockerRegistryUrl: 'https://deploy-releases-hyperspace-docker.common.repositories.cloud.sap',
                additionalPodProperties: [
                    imagePullSecrets: ['docker-secret']
                ]
                ) {
                    sh 'go version'
                }
```

## Jenkins/Non K8s environment

Please use the `dockerExecute` [step](../steps/dockerExecute.md#description) from Piper and pass the `dockerRegistryCredentialsId` parameter via step/stage in your `.pipeline/config.yaml`, or directly as a parameter.

Let's assume that 'ArtifactoryCredentials' is the name of the credential which is needed to pull docker image and it's also present in Jenkins. In this case since we are pulling from Artifactory, it would be technical username and API key.

```text
dockerExecute(script: this,
                dockerImage: 'deploy-releases-hyperspace-docker.common.repositories.cloud.sap/golang:1.0.0-20211123154630_dc0998ed17e94b0e5800d12b0e5273eb28373654',
                dockerRegistryUrl: 'https://deploy-releases-hyperspace-docker.common.repositories.cloud.sap',
                dockerRegistryCredentialsId: 'ArtifactoryCredentials'
                ) {
                    sh 'echo "Hola!"'
                }
```

## Azure DevOps

In the general purpose pipeline, only the cnbBuild and sonarExecute steps have the possibility to run with a private image.
Generally, the image should come from the Common Repository, as not all stages run on ADO agents that have access to SAP's internal network.
It requires adding a [service connection](https://learn.microsoft.com/en-us/azure/devops/pipelines/library/service-endpoints?view=azure-devops&tabs=yaml) for the Docker registry in the ADO user interface.
The name of the added service connection needs to be passed as parameter to the general purpose pipeline in the `azure-pipelines.yml` file as follows:

```text
extends:
  template: sap-piper-pipeline.yml@piper-pipeline-azure
  parameters:
    dockerRegistryConnection: '<name of added service connection>'
```

### cnbBuild

Besides the above mentioned steps, for the cnbBuild step the Docker registry credentials need to be set in Vault for `dockerConfigJSON` as well.
Open the Hyperspace Vault UI (see [details](https://hyperspace.tools.sap/docs/features_and_use_cases/connected_tools/vault.html#connecting-your-pipeline-to-vault)) and add a new secret named `docker-config` with key `dockerConfigJSON` and value `<your-docker-config>`. Ensure that the credentials provided in the `auth` property (username and password) are concatenated using a colon and encoded as base64: `echo -n "<usename>:<password> | base64`.

```json
{
  "dockerConfigJSON": "{\"auths\": {\"$server\": {\"auth\": \"base64($username + ':' + $password)\"}}"
}
```

!!! note
    The value **must** be provided as a string and not as a JSON object. When pasting unescaped content, Vault will automatically escape the necessary parts.

Example:

```json
{
  "dockerConfigJSON": "{\"auths\": {\"common.repositories.cloud.sap\": {\"auth\": \"dXNlcm5hbWU6cGFzc3dvcmQ=\"}}"
}
```

### Other steps

For other steps, to make it work in the way explained above, you could fork the general purpose pipeline, and add the following line to the step:

`dockerRegistryConnection: ${{ parameters.dockerRegistryConnection }}`

## GitHub Actions

The Solinas runners are able to pull images from the internal Artifactory without additional configuration.

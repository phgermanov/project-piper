# Set-up Jenkins instance

On this page you find information on how to set up your Jenkins instance.

!!! info

    Information on the orchestrator technology can be found [here](orchestrator_technologies.md).

To execute Piper ready-made pipelines, a Jenkins server is required. You can choose between the following options:

## 1. Jenkins as a Service (preferred)

By using [Jenkins as a Service (JaaS)](https://pages.github.tools.sap/hyperspace/jaas-documentation/), you do not need to take care about the system infrastructure for your Jenkins.

Just select the latest version of the [**Jenkins_GKE_x.x.x_PIPER_xxx** template in JaaS](https://jenx.int.sap.eu2.hana.ondemand.com/api/#/imageOverview) and create your instance based on it. This image contains a pre-configured Jenkins and is designed for running all flavors of Piper ready-made pipelines.

!!! note
    Make sure the environment variable `ON_K8S` is set to `true`. This can be done via `Jenkins -> Manage Jenkins -> Configure System -> Global properties -> Environment variables`.

!!! caution
    If you do not choose the APOLLO image but another JaaS image, you will not be able to use Piper as expected, as it relies for example on dedicated plugins and certain pre-configured settings.

## 2. Pre-configured Docker Image

If you have specific requirements towards the infrastructure of your CD server (e.g. CPU/RAM), you might want to run it on an own machine, for example, on Converged Cloud. All you need, is a Linux environment with a recent version of Docker.

### Piper Jenkins Docker Image with cx-server

You can easily instantiate an own server by using the `cx-server`.
For more information on using the `cx-server` tool for running your own Jenkins, consult the [Jenkins operations guide](https://github.com/SAP/devops-docker-cx-server/blob/master/docs/operations/cx-server-operations-guide.md).

Use the following settings in your `server.cfg` for using the internal image:

```ini
### Address of the used docker registry
docker_registry=docker.wdf.sap.corp:50000

### Name of the used docker image
docker_image="piper/jenkins"
```

### Manually managed Docker container

!!! caution
    This approach is not recommended since you have to make sure yourself that your Jenkins instance fully complies to the Piper requirements.

Alternatively, you can manually run and manage your own container with the [JaaS base image](https://github.tools.sap/hyperspace/jaas-image).<br />

<!---
!!! tip "Set up pipeline"
    To set up your pipeline, go to the documentation part ("Set up pipeline")[].
--->

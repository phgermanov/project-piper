# What you get

Piper offers you the following:

* [Piper ready-made pipelines](stages/README.md) (e.g. general purpose pipeline with main focus on cloud applications and services)
* [Piper step library](lib/README.md)
* [Piper common configuration layer](configuration.md)

You can take the Piper ready-made pipeline out of the box or use it as starting point for your [custom extensions](extensibility.md).

The Piper step library helps you to get your pipeline into a shape as you need it - and this with very low effort.

!!! note "Ready for Cloud"

    The Piper ready-made pipelines **focus on "microservice"-like use-cases**: where you have one source code repository which results in one artifact which can be autonomously deployed. The Piper ready-made pipelines are optimized for speedy deployments (in the best case if required in a  _Continuous Deployment_ mode) and are tailored to independent agile teams.

!!! note "Building on SAP's Cloud Curriculum concepts"

    The Piper ready-made pipelines using the Piper step library follow the pipeline principles which are trained in SAP's Cloud Curriculum course "[Continuous Delivery and DevOps](https://github.wdf.sap.corp/cc-devops-course/coursematerial)".
    The ready-made pipelines and the step library help you bring the trained concepts into your daily life.

!!! tip ""

    === "Jenkins"

        **Jenkins 2.0 Pipelines as Code**

        The Piper ready-made pipelines are based on the general concepts of [Jenkins 2.0 Pipelines as Code](https://jenkins.io/doc/book/pipeline-as-code/).

        In case you need to go beyond the ready-made pipeline, you have the power of our internal as well as the external Jenkins community at hand to enhance your pipelines.

    === "Azure DevOps"

        **Azure Templates**

        The Piper ready-made pipelines are based on the general concepts of [Azure DevOps Pipeline templates](https://docs.microsoft.com/en-us/azure/devops/pipelines/process/templates?view=azure-devops).

        In case you need to go beyond the ready-made pipeline, you have the power of Azure DevOps Pipelines at hand to enhance your pipelines.

    === "GitHub Actions"

        **Workflow Templates**

        The Piper ready-made pipelines are based on the concepts of [GitHub Actions reusable workflows](https://docs.github.com/en/enterprise-server@3.9/actions/using-workflows/reusing-workflows).

        In case you need to go beyond the ready-made pipeline, you have the power of GitHub Actions workflows at hand to enhance your pipelines.

## The best-practice way: Piper ready-made pipelines

Do you want to efficiently deliver your project with a best-practice pipeline? Then don't reinvent the wheel and start with one of our  [ready-made pipelines](stages/README.md) (e.g. general purpose pipeline) offered in [Hyperspace Onboarding](https://hyperspace.tools.sap/pipelines).

For most use cases, you will not need to write a single line of code. Instead, your delivery pipeline is fully controlled by a sophisticated configuration file. However, don't worry - if your project setting requires modifications to the pipeline, you can adapt and extend its behavior with our unique extension mechanism.

Our pipeline implementations are continuously improved and you will benefit from pipeline qualities and capabilities that are contributed by a growing community of Continuous Delivery experts around Piper.

## The do-it-yourself way: Use the Piper step library

Do you have specific requirements in your project that aren't matched by one of our ready-made pipelines? Then don't despair. Piper offers also a rich [shared library](lib/README.md) of codified steps, enabling you to customize our ready-made pipelines, or to even build your own customized pipeline.

In some cases, it's not desirable to include specialized features in the ready-made pipelines. However, you can still benefit from their qualities, if you address your requirements through an [extension](extensibility.md).

!!! info "Need some background information?"
    Get information on what a Continuous Delivery pipeline is [here](background.md).

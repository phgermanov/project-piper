# Background information

## What's a Continuous Delivery pipeline?

Piper pipelines can be built in various flavours. Their structure is composed of stages which represent meaningful units of work, for example, performing the application build. Stages comprise a number of individual steps which perform steps towards the stage goal, for example, compiling the application. The following image illustrates a schematic overview of a typical Continuous Delivery pipeline:

![Pipeline Overview](https://github.wdf.sap.corp/cc-devops-course/coursematerial/raw/master/ContinuousDelivery/images/Overview.png)  
\[Ref: [Cloud Curriculum - DevOps Training](https://github.wdf.sap.corp/cc-devops-course/coursematerial)\]

The pipeline starts with every commit, executes a central build (xMake) in order to comply with SAP's corporate requirement [_SLC-29_](https://wiki.wdf.sap.corp/wiki/display/pssl/SLC-29), runs a number of tests with respect to quality, security, deploys the artifact to different landscapes/spaces in the Cloud where further tests are executed.
After all tests have passed successfully this exact build artifact is taken and promoted, e.g. to the Release Repository (Nexus) and finally deployed to the production landscape.

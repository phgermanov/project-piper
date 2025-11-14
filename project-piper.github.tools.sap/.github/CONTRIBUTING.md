# Guidance on how to contribute

All contributions to this project will be released SAP internally AS IS WITHOUT WARRANTY OF ANY KIND.
Potentially we aim also to release parts according to the [MIT license terms](https://github.wdf.sap.corp/ContinuousDelivery/jenkins-pipelines/blob/master/LICENSE.md) to customers/partners.

By submitting a pull request or filing a bug, issue, or feature request, you are agreeing to comply with the waiver.

There are two primary ways to help:
* Using our forum,
* Using the issue tracker, and
* Changing the code-base.

## Using our forum [OBSOLETE]
We have a [Jam group](https://jam4.sapjam.com/groups/about_page/fhHMfvtxIARCiFoTLNY9iN) including a forum where ideas, solutions and also questions can be raised.

## Using the issue tracker

Use the issue tracker to suggest feature requests, report bugs, and ask questions. This is also a great way to connect with the developers of the project as well as others who are interested in this solution.

Use the issue tracker to find ways to contribute. Find a bug or a feature, mention in the issue that you will take on that effort, then follow the Changing the code-base guidance below.

## Changing the code-base

Generally speaking, you should fork this repository, make changes in your own fork, and then submit a pull-request. All new code should have thoroughly tested to validate implemented features and the presence or lack of defects and it should come with an adequate documentation. 

All pipeline library coding MUST come with an automated test as well as adequate documentation.

Additionally, the code should follow any stylistic and architectural guidelines prescribed by the project. In the absence of such guidelines, mimic the styles and patterns in the existing code-base.

### Piper stage documentation

Each Piper Stage documentation is from [piper-stage-config.yml](https://github.tools.sap/project-piper/sap-piper/blob/master/resources/piper-stage-config.yml). 
Eg., [Init Stage](https://github.wdf.sap.corp/pages/ContinuousDelivery/piper-doc/stages/gpp/init/) documentation [here](https://github.tools.sap/project-piper/sap-piper/blob/5bb445079f7baee3f7ce5017673b4889830cf8cd/resources/piper-stage-config.yml#L57)

### Piper step documentation
Step documentation comes from step_name.yaml.
Eg., [sapCumulusUpload](https://github.wdf.sap.corp/pages/ContinuousDelivery/piper-doc/steps/sapCumulusUpload/) documentation [sapCumulusUpload.yaml](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/blob/0521962a7cdda1825c8e0c01e381eff49cb989f9/resources/metadata/sapCumulusUpload.yaml#L2)

Some steps might need some additional information/ SAP specific information/Preferred support channel. In those cases documentation for this section comes from the specific md file.
Few examples here:
[hadolintExecute SAP specific section](https://github.wdf.sap.corp/pages/ContinuousDelivery/piper-doc/steps/hadolintExecute/#sap-specifics) documentation [here](../steps/__hadolintExecute_sap.md)

[Support channel specified in DwC step](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/blob/454788eabe4e7fad11c8c2e856d9ea934a3bac96/resources/metadata/sapDwCStageRelease.yaml#L9)

[Codeql Go step SAP specific section with some useful links and support info](../docs/steps/__codeqlExecuteScan_sap.md)

Please submit a Pull request with changes.
You can test it locally if needed , as mentioned in [README.md](../README.md#build-and-run-documentation-locally)

Once changes are merged , we need to run [piper-doc](https://jenkins.piper.c.eu-de-2.cloud.sap/job/ContinuousDelivery/job/piper-doc/job/master/) pipeline job to update the docs either manually to see the changes immediately or wait for the scheduled job to run at 12.30 pm CET to reflect your changes.

You can also see link to same job in toolbar of this Github repo as well 

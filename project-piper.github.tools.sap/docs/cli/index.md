# Project "Piper" CLI

The CLI is built using the go programming language and thus is distributed in a single binary file for Linux.

There are two binaries available:

* `sap-piper` for SAP-internal steps (Piper InnerSource project)
* `piper` for universal steps (Piper Open Source project)

The latest released version can be downloaded via

* `wget https://github.wdf.sap.corp/ContinuousDelivery/piper-library/releases/latest/download/sap-piper`
* `wget https://github.com/SAP/jenkins-library/releases/latest/download/piper`

Specific versions can be downloaded from the [github.wdf.sap.corp releases](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/releases) and [github.com releases](https://github.com/SAP/jenkins-library/releases) pages.

Once available in `$PATH`, it is ready to use.

To verify the version you got, run `piper version` or `sap-piper version`.

To read the online help, run `piper help` or `sap-piper help`.

If you're interested in using the Open Source version with GitHub Actions, see [the Project "Piper" Action](https://github.com/SAP/project-piper-action) which makes the tool more convenient to use.

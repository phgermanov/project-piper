# For Piper Innersource Community

This document covers the basics to make a contribution to Piper library via a piper step.

## General Info

- Piper supports 3 orchestrators Jenkins, GitHub actions and Azure Pipelines.
- Piper requires all new steps to be implemented in GoLang to be able to run those step in all orchestrators
- Even if the step is written in GoLang, there is a need to have a tiny Groovy wrapper around it for Jenkins.
- **Important:** GoLang program is a program which doesn't start a Docker Container out of the service Image, but instead is executed inside the Image provided for the step as a base. In case of the API Metadata Validator, the GoLang binary will be triggered inside our [image](Dockerfile_Library) we provide.

### Creating a new Piper step

This procedure was initially done for the validator step, but can be useful if we decide to create additional steps:

1. Piper is divided in two parts. One part of Piper code is [open source](https://github.com/SAP/jenkins-library/) and contains Piper code that is not SAP specific. Another part is SAP internal and runs as an [inner source project](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/) which contains SAP specific code. At the moment, we are no longer accepting new steps in the open source part. Please make sure to read the [general documentation](https://github.com/SAP/jenkins-library/blob/master/DEVELOPMENT.md), and make sure to apply it in the inner source Piper repository. If you wish to be a collaborator in the [inner source Piper code base](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/), then please see section [Collaborations and teams](#collaborations-and-teams)

2. Fork and checkout [Piper library](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/) repository.

3. **GoLang** compiler is required on the machine. Mac users can use `brew install go` to install it.
   **Note:** Make sure the **Go** version is the same as required by the Piper library. The current version can be found in the [`go.mod`](go.mod) file (normally at the top of the file).

4. Create `yaml` file with the step metadata and run `go generate` as described [here](https://github.com/SAP/jenkins-library/blob/master/documentation/DEVELOPMENT.md#best-practices-for-writing-piper-go-steps).
   - This should create some `.go` files in the project with some boilerplate code.
   - DO NOT MODIFY files which have postfix `_generated`.
   - Every time `yaml` file changes, `go generate` should be called to update `_generated` files, otherwise automatic tests will fail.

5. Once your contribution PR is ready, please send an email to `DL_6287AE4DEC3CA802990E86E5@global.corp.sap`, to notify Piper team. Please also clarify the urgency and any useful background information which can help us know about your use case

### Editing a step | *An example step is sapExecuteApiMetadataValidator*

The validator footprint/assets in the Piper repository:

- [./cmd/sapExecuteApiMetadataValidator.go](cmd/sapExecuteApiMetadataValidator.go) — Main logic is here
- [./cmd/sapExecuteApiMetadataValidator_test.go](cmd/sapExecuteApiMetadataValidator_test.go) - Tests are mandatory for the main logic
- [./cmd/sapExecuteApiMetadataValidator_generated.go](cmd/sapExecuteApiMetadataValidator_generated.go) - DO NOT MODIFY
- [./cmd/sapExecuteApiMetadataValidator_generated_test.go](cmd/sapExecuteApiMetadataValidator_generated_test.go) - DO NOT MODIFY, in case of its failures try to run `go generate` first to regenerate it.
- [./vars/sapExecuteApiMetadataValidator.groovy](vars/sapExecuteApiMetadataValidator.groovy) - Groovy wrapper for Jenkins
- [./vars/sapPiperStageCentralBuild.groovy](vars/sapPiperStageCentralBuild.groovy) - This is a part of "General Purpose Pipeline" (GPP), the validator step is called there, also Cumulus upload happens there (search for the `sapExecuteApiMetadataValidator` in this file).
- [./resources/metadata/sapExecuteApiMetadataValidator.yaml](resources/metadata/sapExecuteApiMetadataValidator.yaml) - metadata file. It's also used to generate the [step documentation](https://github.wdf.sap.corp/pages/ContinuousDelivery/piper-doc/steps/sapExecuteApiMetadataValidator/). Always run `go generate` after changing this file.
And for the documentation to be updated a separate [documentation job](https://jenkins.piper.c.eu-de-2.cloud.sap/job/ContinuousDelivery/job/piper-doc/job/master/) needs to run. This can be run manually or it runs everyday on schedule at 12.30 pm CET

#### Run tests locally

```sh
go test ./cmd/sapExecuteApiMetadataValidator.go ./cmd/sapExecuteApiMetadataValidator_generated.go ./cmd/sapExecuteApiMetadataValidator_test.go
```

#### Run Groovy wrapper locally

```sh
go build
./piper-library sapExecuteApiMetadataValidator --ruleset=sap:core:v2
./piper-library sapExecuteApiMetadataValidator --files='./**/*.json','./*.yaml' --ruleset sap:core:v2
```

## Default configuration and stage conditions

The default configuration (and stage config) are now maintained in a [separate repository](https://github.tools.sap/project-piper/resources).
It's described [here](https://github.tools.sap/project-piper/resources/blob/main/README.md#how-to-make-updates-to-the-defaults) how to do it.

This means that the [`resources/piper-defaults.yml`](resources/piper-defaults.yml) and [`resources/piper-stage-config.yml`](resources/piper-stage-config.yml) files shouldn't be altered directly anymore. These are referenced to by Jenkins, so changes would come into affect, but these files will be overwritten during Piper's release.

If you wish to edit the default(s) / stage conditions then you could fork the repo and create the pull request with your changes.

If you wish to be a contributor with write access to this repository then please create a [SNOW](https://itsm.services.sap/sp?id=sc_cat_item&sys_id=703f22d51b3b441020c8fddacd4bcbe2&service_offering=455f43371b487410341e11739b4bcb67) incident with the below message:

```text
      Please add my <D/I/C number>, my <sap email address> to the team https://github.tools.sap/orgs/project-piper/teams/toolowners

      Business Justification for the request: < please provide a valid justification for the request including a brief description of your team and your role in piper maintenance >
```

## Collaborations and teams

Piper has always run on the principle of inner/open source project and provides full support to other development teams to maintain parts of the Piper code/steps so that teams are empowered to improve the Piper code base.

- <i>Contribution to inner source piper code at <https://github.wdf.sap.corp/ContinuousDelivery/piper-library></i>.

  - To enable better collaboration piper has enabled [CODEOWNERS](.github/CODEOWNERS) which contains teams and the respective code
   that they maintain

      If you are interested in maintaining only part of the Piper code (e.g. as a team who maintains one Piper step) we recommend to first create a github team under the github org ``ContinuousDelivery``. to get a team created please create a [SNOW](https://itsm.services.sap/sp?id=sc_cat_item&sys_id=703f22d51b3b441020c8fddacd4bcbe2&service_offering=455f43371b487410341e11739b4bcb67) incident and include the below text in the SNOW message:

      ```yaml
      Please create a github team : <name of your team> under github org ContinuousDelivery . Please include the below users in the team
      - D/I/C number, emailId of the team member 1
      - D/I/C number, emailId of the team member 2

      Business Justification for the need of the team: < please provide a valid justification for the request including a brief description of your team and your teams role in piper maintenance >

      Team maintainer: < please nominate one or more D/I/C number, emailId of the team member who would be maintainers of the team going forward >
      ```

      Once the above SNOW incident is serviced, create a pull request (from your fork) for [CODEOWNERS](.github/CODEOWNERS). in [CODEOWNERS](.github/CODEOWNERS) please make sure to include the correct file pattern for the part of the code that your newly created team wishes to maintain.

      Once the PR for [CODEOWNERS](.github/CODEOWNERS) is merged then you will become a collaborator for the code base you wish to maintain and can create and merge PR(s) within your team.

- <i>Contributing without owning a specific step</i>.
  - To contribute Piper by not owning any step can be done via fork by following the below steps:
    - Fork the [Piper library](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/) repository.
    - Create a branch for your changes.
    - Make your changes and commit them.
    - Create a pull request to the master branch of the Piper library repository.
    - The Piper team will review your changes and merge them if they are appropriate by running few checks manually.
    - The checks can be run by any collaborator to the repository by commenting "/mirror" in the pull request.
    - Inform us via `DL_6287AE4DEC3CA802990E86E5@global.corp.sap`, if a contribution is urgent and blocking your use case or has a business criticality.

## Contribution to Piper General Purpose Pipeline (GPP) definition and Orchestrator Enabler(s)

GPP is the Piper defined stages with steps that could be reused by development teams at SAP. Contributing to the Piper GPP generally means adding/modifying stages/steps in the GPP (for example adding a new step to the Release stage) which will effect all development teams using the entire GPP as is or who use stage(s) from the GPP.

If you are interested in contributing to the Piper GPP definition then please create a fork and create the respective PR:

| Orchestrator    | Piper GPP Repo |
| --------        | ------- |
| Jenkins         | <https://github.wdf.sap.corp/ContinuousDelivery/piper-library>   |
| Azure           | <https://github.tools.sap/project-piper/piper-pipeline-azure>     |
| Github actions on wdf      | <https://github.wdf.sap.corp/ContinuousDelivery/piper-pipeline-github>    |
| Github actions on tools    | <https://github.tools.sap/project-piper/piper-pipeline-github>    |

:bulb: Why are there two repositories ``Github actions on wdf`` and ``Github actions on tools`` for the Piper GPP on github actions?
Github actions running on [github.wdf](https://github.wdf.sap.corp/) instance cannot access github workflows from [github.tools](https://github.tools.sap/) and vice versa. The two repositories are mirrors and contain the same Piper GPP definition so that development teams having code base on on [github.wdf](https://github.wdf.sap.corp/) or on [github.tools](https://github.tools.sap/) can use github actions.

:bulb: the Piper GPP must provide a seamless user experience irrespective of the orchestrator and hence changes must be consistent across all orchestrator. If you plan to change a stage in the Jenkins since that is most relevant in your use-case please feel free to create a pull request only for Jenkins. However as we have always believed in the inner source mentality to help the wider developer community at SAP, please also add a line in the PR to make sure that the changes are carried forward to Azure and Github Actions. The Piper team will be work with you to make sure the changes are consistent across all orchestrators.

To run Piper on Azure and Github Action, there are two more repositories that contain an custom [Azure Task](https://learn.microsoft.com/en-us/azure/devops/pipelines/process/tasks?view=azure-devops&tabs=yaml) and[Github Action](https://docs.github.com/en/actions) for Piper. If you are interested in contributing to the Piper Azure task or the Piper github action then please create a fork and create the respective PR:

| Orchestrator Enabler    | Repo |
| --------                | ------- |
| Piper Azure Task              | <https://github.tools.sap/project-piper/piper-azure-task>   |
| Piper Github Actions          | <https://github.com/SAP/project-piper-action>     |

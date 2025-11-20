# Development

This project builds on the [Open Source project "Piper"](https://www.project-piper.io/).

Sources of the Open source project can be found on [SAP/jenkins-library](https://github.com/SAP/jenkins-library).

We aim for a clear separation for the steps:

- SAP-specific steps (e.g. with connections of SAP proprietary systems) are part of this InnerSource project repository.
- Steps connecting to publicly available software/systems (e.g. Fortify, Sonar, ...) are part of the Open Source repository. They are possibly applicable to customers, partners and other interested parties.

Due to the string dependency to the Open Source project,
please see [DEVELOPMENT.md](https://github.com/SAP/jenkins-library/blob/master/DEVELOPMENT.md) for details about how to get started with Piper step development.

## Documentation

Most of the [documentation](go.sap.corp/piper) is generated via the yaml file of the respective piper step.
If you are updating the yaml file for documentations , please run [this](https://jenkins.piper.c.eu-de-2.cloud.sap/job/ContinuousDelivery/job/piper-doc/job/master/) job manually .
If you dont have access, make sure to inform one of the piper team members to run it for you.
This is necessary for the changes to be reflected in the documentation.

## Release

*Project Piper* consists of several consumable artifacts which are releases:

- GOlang binary (in this repository)
- Jenkins Groovy library containing the Jenkins *general purpose pipeline* (in this repository)
- Azure task (in [project-piper/piper-azure-task](https://github.tools.sap/project-piper/piper-azure-task) repository)
- Azure *general purpose pipeline* (in [project-piper/piper-pipeline-azure](https://github.tools.sap/project-piper/piper-pipeline-azure) repository)
- GitHub Action (in [SAP/project-piper-action](https://github.com/SAP/project-piper-action) repository)
- GitHub tools *general purpose pipeline* (in [project-piper/piper-pipeline-github](https://github.tools.sap/project-piper/piper-pipeline-github) repository)
- GitHub wdf *general purpose pipeline* (in [project-piper/piper-pipeline-github](https://github.wdf.sap.corp/project-piper/piper-pipeline-github) repository)

Here the release of the GOlang binary and Jenkins library is described, all other release processes are described in the respictive repository (`DEVELOPMENT.md`).

There is a commit-based pipeline running on the GitHub actions. The pipeline is defined in this repository in [piper-golang-gpp](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/blob/master/.github/workflows/piper-golang-gpp.yml), and it runs for each commit to the main branch (via PR only) and all pull-requests that target this branch.

The release pipeline is scheduled to run every Monday 10am GMT to validate the repository content and upon successful validation publishes the GOlang binary as `sap-piper` to the [most recent release](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/releases/latest) on the GitHub repository on `github.wdf.sap.corp`. The GOlang binaries are also released to the [most recent release on a separate repository](https://github.tools.sap/project-piper/sap-piper/releases/latest) on `github.tools.sap`. Besides that, the binaries are also published as a Docker image `https://docker.wdf.sap.corp:51131/piper/library:latest` to artifactory based on the [Dockerfile_Library](Dockerfile_Library).

NOTE: Changes to the Jenkins library (non-GOlang) **are effective immediately** upon commit to the main branch!

The scheduled release (`sap-piper`) is publishing the binary built for Linux, Windows and Mac (Darwin) to GH Tools and GH WDF.

## Testing a Piper branch

### Jenkins

The binaries can be built on the fly. This can for example by adding the following to the Jenkinsfile:

```groovy
@Library(['piper-lib@testInnerSourceBranch', 'piper-lib-os@testOpenSourceBranch']) _

node {
    stage('Build open source Piper') {
        deleteDir()
        git url:'https://github.com/SAP/jenkins-library.git', branch: 'testOpenSourceBranch'
        dockerExecute(script: this,  dockerImage: 'docker.wdf.sap.corp:50000/golang:1.21', dockerOptions: '-u 0') {
                        sh 'GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o piper . && chmod 777 piper'
        }
        sh "./piper version"
        stash name: 'piper-bin', includes: 'piper'
    }
}

node {
    stage('Build inner source Piper') {
        deleteDir()
        git url:'https://github.wdf.sap.corp/ContinuousDelivery/piper-library.git', branch: 'testInnerSourceBranch', credentialsId: 'Piper_GitHub'
        dockerExecute(script: this,  dockerImage: 'docker.wdf.sap.corp:50000/golang:1.21', dockerOptions: '-u 0') {
                        sh 'GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -o sap-piper . && chmod 777 sap-piper'
        }
        sh "./sap-piper version"
        stash name: 'sap-piper-bin', includes: 'sap-piper'
    }
}

sapPiperPipeline script: this
```

- Your branch name needs to be specified for the library import in the first line, and in the `git url` line.
- The Golang version for the `dockerExecute` step needs to match with Piper's [Golang version](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/blob/master/go.mod#L5).
- For building inner source Piper, `credentialsId` needs to contain a personal access token for GitHub WDF.

### ADO & GHA

It's documented in the GPP repositories:
[https://github.tools.sap/project-piper/piper-pipeline-azure](https://github.tools.sap/project-piper/piper-pipeline-azure)
[https://github.tools.sap/project-piper/piper-pipeline-github](https://github.tools.sap/project-piper/piper-pipeline-github)

## Enabling System Trust in a step

The Hyperspace System Trust can be used in a step, by adding it as resource reference to a parameter in a step's YAML definition.
This is done in the same way as is done with Vault secrets.
[Example](https://github.com/SAP/jenkins-library/blob/master/resources/metadata/sonarExecuteScan.yaml#L44) from `sonarExecuteScan.yaml`:

```YAML
- name: token
type: string
description: "Token used to authenticate with the Sonar Server."
scope:
    - PARAMETERS
secret: true
resourceRef:
    - type: vaultSecret
    name: sonarVaultSecretName
    default: sonar
    - name: sonarTokenCredentialsId
    type: secret
    - type: systemTrustSecret
    name: sonarSystemtrustSecretName
    default: sonar
```

The last resource reference in the parameter refers to System Trust (which used to be named Trust Engine).

- `type: systemTrustSecret` indicates that it's a System Trust secret.
- `name: sonarSystemtrustSecretName` will be the name of the secret in Piper's core code.
- `default: sonar` is the system for which the secret will be queried from the System Trust API. At the time of writing, the accessed System Trust endpoint would look like this: `https://api.trust.tools.sap/tokens?systems=${default}`

If a parameter has both a Vault and System Trust resource reference, then Piper will first attempt to retrieve the secret from Vault, and if it's not there, it will access System Trust.

Piper's step documentation generator will [automatically](https://github.com/SAP/jenkins-library/blob/aa1e67547a4fdfeb0f829d1dfeee6578c15ec7d4/pkg/documentation/generator/parameters.go#L132) handle the documentation of your step.

## Step definitions

### Parameter scopes

Parameters can be passed to a Piper step in two ways: directly as flag when the binary is called, or through the Piper configuration file (`.pipeline/config.yml`).
It is defined in steps' YAML files which ways are possible for a parameter, which can look like this for instance:

```YAML
- name: pipelineId
  description: Id of the Cumulus pipeline.
  type: string
  mandatory: true
  scope:
    - GENERAL
    - PARAMETERS
    - STAGES
    - STEPS
```

The following scopes are possible (with the example of the `pipelineId` parameter of the `sapCumulusUpload` step):

- `PARAMETERS` - enables to pass it as a flag

```sh
./piper --pipelineId 123
```

- `GENERAL` - enables adding it to the `general:` part of the `config.yml`

```YAML
general:
  pipelineId: 123
```

- `STAGES` - enables adding it to the `stages:` part of the `config.yml`

```YAML
stages:
  Release:
    pipelineId: 123
```

- `STEPS` - enables adding it to the `steps:` part of the `config.yml`

```YAML
steps:
  sapCumulusUpload:
    pipelineId: 123
```

Scopes do not affect resource references. Even without any scope, resource refs work.

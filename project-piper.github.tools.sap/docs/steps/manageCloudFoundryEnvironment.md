# manageCloudFoundryEnvironment

## Description

The step `manageCloudFoundryEnvironment` is utilizing the 'EnvironmentManager' (native Groovy module/ part of piper-library) to setup and manage Cloud Foundry spaces.
This currently includes creation/deletion of spaces, creation/deletion of backing services, as well as user provided services and assigning/unassigning users, incl. roles within a given space.
You can use this piper-step via single commands (e.g. create-space) and passing parameters, similar to CF CLI, or you can use the 'setup-environment' or 'setup-all-environments' commands, passing one/ multiple 'environment-yml' files, which specify the setup of a space completely.

## Prerequisites

- Cloud Foundry organization and deployment user are available
- Credentials for deployment have been configured in Jenkins with a dedicated Id
- on the Jenkins either docker is installed or groovy and cf cli is installed

## Example

### Generic usage of pipeline step

```groovy
manageCloudFoundryEnvironment script:this, command:<EnvMan-command [EnvMan-command-options]>
```

### Sample usage of pipeline step

#### Complete environment

```groovy
manageCloudFoundryEnvironment script:this, command:"setup-environment -y <your-environment.yml-file>"
manageCloudFoundryEnvironment script:this, command:"setup-all-environments -e <your-folder-with-multiple-environment.yml-files>",dockerImage:'someImageUrl',piperLibraryName:"piper-library",cfCredentialsId:"myCredentials"
```

#### Single commands

```groovy
manageCloudFoundryEnvironment script:this, command:"delete-space -a $cfApiEndpoint -o $cfOrg --spacename $spacename",dockerImage:'',piperLibraryName:"piper-library",cfCredentialsId:"myCredentials"
manageCloudFoundryEnvironment script:this, command:"create-service -a $cfApiEndpoint -o $cfOrg -s $spacename --cs_options 'name plan instance' --jsonString '{"key":"value"}' "
```

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|cfCredentialsId|yes|`CF_CREDENTIAL`|`any String`|
|mhCredentialsId|no||`any String`|
|command|no|`setup-environment -y ${config.environmentDescriptorFile}`|see [commands](../others/EnvironmentManager/readme.md)|
|dockerImage|no|`docker.wdf.sap.corp:50000/piper/cf-cli`||
|dockerWorkspace|no|`/home/piper`||
|envManPath|no|||
|environmentDescriptorFile|no|`environment.yml`||
|piperLibraryName|yes|||

### Details

#### Commands

A list and details of the commands can be found at the [EnvironmentManager documentation](../others/EnvironmentManager/readme.md).

#### cfCredentialsId

The name of the credential ID for Cloud Foundry is configured in Jenkins. This ID is used to log into CF and has to have the rights to create spaces. You can provide an ID directly with the step parameters. Alternatively the step will look for an ID name in the config.ymlies file. At last if no such property exists it will use 'CF_CREDENTIAL' as the ID name.

#### mhCredentialsId

The name of the credential ID of the User used for security setup with EnvMan. For more on the see [EnvironmentManager security docu](../others/EnvironmentManager/security.md#credentials).

#### piperLibraryName

Provides the name of the Library in Jenkins. By default name is the name of the first library listed in the Jenkinsfile. Used to access the EnvironmentManager script in the resources folder of the Library.

#### dockerImage

By default the EnvironmentManager groovy script will be executed within an piper-cf-cli docker image (docker.wdf.sap.corp:50000/piper/cf-cli).  If the Jenkins VM does not support docker but groovy and cf cli directly you can deactivate the use of docker by passing an empty string `''`. If you provide an dockerImage URL it will be used instead of the default one.

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|cfCredentialsId|X|X|X|
|mhCredentialsId|X|X|X|
|command|X|X|X|
|dockerImage|X|X|X|
|dockerWorkspace|X|X|X|
|envManPath|X|X|X|
|environmentDescriptorFile|X|X|X|
|piperLibraryName|X|X|X|

# EnvironmentManager module of piper library

This is the documentation of the standalone groovy script. The corresponding pipeline step can be found [here](../../steps/manageCloudFoundryEnvironment.md). Sometimes in this documentation the EnvironmentManager will also be called EnvMan.
The EnvironmentManager is used to setup spaces on Cloud Foundry. It wraps cf cli calls into single commands to make creating spaces and starting up services simple.
You can use it with a piperstep in your pipeline, or directly on shell/bash. Therefore also direct shell calls during your classic Jenkins jobs are possible.
For the best overview what EnvMan can do see the [yml description](yml.md) of all possible sections.

## Prerequirements

To use the EnvironmentManager directly the following steps are needed:

- install groovy
- clone piper-library
- add `[piper-library]/resources/EnvironmentManager` to your `PATH` variable to get access to em.sh/em.bat from everywhere.

## General use

The general use of EnvironmentManager is:

- On Unix (bash): `em.sh commandString`
- On Windows: `em.bat commandString`
- piper-step: `manageCloudFoundryEnvironment script:this,command:commandString`

The `commandString` consists of: `command [options]`

## Commands

The EnvironmentManager knows the following commands:

- [create-space](create-space.md)
- [delete-space](delete-space.md)
- [create-service](create-service.md)
- [create-user-provided-service](create-user-provided-service.md)
- [setup-security (new)](setup-security.md)
- [setup-environment](setup-environment.md)
- [setup-all-environments](setup-all-environments.md)
- [delete-spaces-older-than](delete-spaces-older-than.md)
- [list-spaces-older-than](list-spaces-older-than.md)
- [list-services](list-services.md)
- [clean-org-users](clean-org-users.md)

## Command options

Not all options are used for every command. In general a given option on the command line overrides the value stored in an environment variable. Both have higher priority than values given in a YAML file. Exception to this is the [`setup-all-environments`](setup-all-environments.md) command. All commands need at least username & password.

!!! note "Username and Password for the command are already set, if the EnvironmentManager is used by the [piper-step](../../steps/manageCloudFoundryEnvironment.md) (with the given cfCredentialID)."

For individual commands navigate to their pages linked in the list above. Some options are also possible to be passed as environment variables.

### Overview list

| short opt. | long opt. | EnvironmentVariable | short description |
| --- | --- | --- | --- |
| -u | username | CF_USERNAME | CF username |
| -p | password | CF_PASSWORD | Password for Username |
| -a | api_url | CF_API_ENDPOINT | API URL for CF |
| -o | org | CF_ORGANISATION | CF organization |
| -s | space | CF_SPACE | CF space |
|    | spacename |  | same as -s/space |
|  | cs_options |  | Name, plan and instanceName for a new service (one(!) String with 3 words)|
|  | ups_name |  | user provided service name |
| -j | jsonPath |  | path to a json file for create-service or create-user-provided-service |
|  | jsonString |  | String containing a json for create-service or create-user-provided-service |
|  | recreate |  | true: (user provided) service will be delete before creation if it already exists |
| -y | yml | CF_ENVIRONMENT_CONFIGURATION | path to a yml configuration file |
| -e | environmentFiles |  | path to folder with yml files |
|  | days |  | maximum age of spaces for delete old spaces |
|  | pattern |  | pattern for filter of list/delete old spaces |
|  | debug |  | true: more output for debugging |

## Yml-file

This is a file containing all information needed to set up a space with (user-provided) services and roles: [Link](yml.md)

# setup-all-environments

This is the command used to setup many spaces with services and assigned users on Cloud Foundry.

## Mandatory options

The following information is needed for the execution of the `setup-all-environments` command.
It can be provided through commandline options or environment variables.

- username
- password
- environmentFiles

## EnvironmentFiles

This is a path to a folder containing yml files. Setup-environment will be used with all those files setting up each specified environment.

## Optional options

The following parameters in the yml files can be overridden by commandline options or environment variables.

- api_url
- org

## Examples

All examples assume CF_USERNAME and CF_PASSWORD is set.

### Unix (bash)

command: `em.sh setup-environment -y environment.yml -s differentSpacename`

### Windows (powershell)

command: `em.bat setup-environment -y environment.yml -o someOtherOrg`

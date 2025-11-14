# setup-security

This is the command used to setup the authorization model of a business application on Cloud Foundry. Please see the [EnvironmentManager security section](security.md) before using it.

## Mandatory options

The following information is needed for the execution of the `setup-security` command.
It can be provided through command line options or environment variables.

- username
- password
- yml

## yml

Only space, org, api, and security information in the given [yml file](yml.md) is used. See the link on the format of the information.

## Optional options

The following parameters in the yml files can be overridden by command line options or environment variables.

- api_url
- org
- space

## Examples

All examples assume CF_USERNAME and CF_PASSWORD is set.

### Unix (bash)

command: `em.sh setup-security -y security.yml`

### Windows (powershell)

command: `em.bat setup-security -y security.yml`

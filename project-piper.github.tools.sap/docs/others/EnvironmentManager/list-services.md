# list-services

This is the command used to list services on Cloud Foundry.

## Mandatory options

The following information is needed for the execution of the `list-services` command.
It can be provided through commandline options or environment variables.

- api_url
- org
- space
- username
- password

## yml

The use of a yml file to provide api_url, org and space is possible

## Examples

All examples assume CF_USERNAME and CF_PASSWORD is set.

### Unix (bash)

command: `em.sh list-services -a https://api.cf.sap.hana.ondemand.com -o CCXCourseTesting -s testspace`

### Windows (powershell)

command: `em.bat list-services -a https://api.cf.sap.hana.ondemand.com -o CCXCourseTesting -s testspace`

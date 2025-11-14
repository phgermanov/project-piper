# create-space

This is the command used to create a new space on Cloud Foundry.

## Mandatory options

The following information is needed for the execution of the `create-space` command.
It can be provided through commandline options or environment variables.

- api_url
- org
- space
- username
- password

## Examples

All examples assume CF_USERNAME and CF_PASSWORD is set.

### Unix (bash)

command: `em.sh create-space -a https://api.cf.sap.hana.ondemand.com -o CCXCourseTesting -s testspace`

### Windows (powershell)

command: `em.bat create-space -a https://api.cf.sap.hana.ondemand.com -o CCXCourseTesting -s testspace`

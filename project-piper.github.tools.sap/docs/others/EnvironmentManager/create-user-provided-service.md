# create-user-provided-service

This is the command used to create a user provided service in a space on Cloud Foundry.

## Mandatory options

The following information is needed for the execution of the `create-user-provided-service` command.
It can be provided through commandline options or environment variables.

- api_url
- org
- space
- username
- password
- ups_name
- jsonString or jsonPath

## yml

  If a yml file is given all user provided services in that file will be created.
  If also `ups_name` & `jsonString/jsonPath` are used that service will be created too.

## Examples

All examples assume CF_USERNAME and CF_PASSWORD is set.

### Unix (bash)

command: `em.sh create-user-provided-service -a https://api.cf.sap.hana.ondemand.com -o CCXCourseTesting -s testspace --ups_name 'hiddenCredentials' --jsonString '{"username":"c5261232","password":"P4SSW0RD"}'`

### Windows (powershell)

command: `em.bat create-user-provided-service -a https://api.cf.sap.hana.ondemand.com -o CCXCourseTesting -s testspace --ups_name 'hiddenCredentials' --jsonString '{\"username\":\"c5261232\",\"password\":\"P4SSW0RD\"}'`

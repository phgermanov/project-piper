# delete-space

This is the command used to delete a existing space from Cloud Foundry. The EnvironmentManager will wait for the deletion of all service instances and periodically try to delete the space until it either succeeded or tried 100 times.

## Mandatory options

The following information is needed for the execution of the `delete-space` command.
It can be provided through commandline options or environment variables.

- api_url
- org
- space
- username
- password

## Examples

All examples assume CF_USERNAME and CF_PASSWORD is set.

### Unix (bash)

command: `em.sh delete-space -a https://api.cf.sap.hana.ondemand.com -o CCXCourseTesting -s testspace`

### Windows (powershell)

command: `em.bat delete-space -a https://api.cf.sap.hana.ondemand.com -o CCXCourseTesting -s testspace`

# create-service

This is the command used to create a service in a space on Cloud Foundry.

## Mandatory options

The following information is needed for the execution of the `create-service` command.
It can be provided through commandline options or environment variables or a yml file.

- api_url
- org
- space
- username
- password
- cs_options (one String with 3 words for name, plan and instanceName)

## optional options

The following information is optional for this command.

- To pass a json to configure the service use either
  - `--jsonString '{one String containing the json}'` or
  - `--jsonPath path_to_a_file`.
  If both are given the jsonString is used.
- To recreate a Service if it already exists use the `--recreate true` option. On default this is set to false. If set true, an already existing service will be deleted and created afterwards. All bindings will be lost.
- To turn on waiting on service status after service creation use `--waitOnServiceStatus true`. If this option is added to the command the Environment Manager will check every 30 seconds for the last_operation of the service to reach succeed status before it continuous. This is false by default.

## yml

If a [yml](yml.md) file is given all services in that file will be created. If also `cs_options` is used that service will be created too. Using a [yml](yml.md) file also gives the option of adding tags, see [Link](yml.md).

## Examples

All examples assume CF_USERNAME and CF_PASSWORD is set.

### Unix (bash)

command: `em.sh create-service -a https://api.cf.sap.hana.ondemand.com -o CCXCourseTesting -s testspace --cs_options 'xsuaa application uaa-bulletinboard' --jsonString '{"xsappname":"bulletinboard-c5261232"}'`

### Windows (powershell)

command: `em.bat create-service -a https://api.cf.sap.hana.ondemand.com -o CCXCourseTesting -s testspace --cs_options 'xsuaa application uaa-bulletinboard' --jsonString '{\"xsappname\":\"bulletinboard-c5261232\"}'`

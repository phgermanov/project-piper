# setup-environment

This is the command used to setup a space with services and assigned users on Cloud Foundry.

## Mandatory options

The following information is needed for the execution of the `setup-environment` command.
It can be provided through command line options or environment variables.

- username
- password
- yml

## yml

All information in the given [yml file](yml.md) is used. See the link on the format of the information. If your YAML file contains a security section please be aware of the [further requirements](security.md) of this.

## Optional options

The following parameters in the yml files can be overridden by command line options or environment variables.

- api_url
- org

A (user provided) service can be added by the use of command line option. See the corresponding commands for the options needed.

## Examples

All examples assume CF_USERNAME and CF_PASSWORD is set.

### Unix (bash)

command: `em.sh setup-environment -y environment.yml -s differentSpacename`

### Windows (powershell)

command: `em.bat setup-environment -y environment.yml -o someOtherOrg`

## On the order of working through the Yml

The information given in the [yml file](yml.md) will be read in and the entities within setup in the following order. It is currently not possible to change this order:

(since version 0.8.1)

1. setup Space
2. setup Services
3. setup User Provided Services
4. setup Service Keys
5. setup Bindings
6. setup Roles
7. restage Apps
8. setup Security

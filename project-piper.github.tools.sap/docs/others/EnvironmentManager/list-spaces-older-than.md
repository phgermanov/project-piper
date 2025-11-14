# list-spaces-older-than

This is the command used to list spaces older than X days on Cloud Foundry.
There is a default filter pattern ,that only spaces starting with dev|acceptance|integration|production|performance followed by -C|D|I number gets listed. The pattern gets matched against the _lower-cased_ version of the spacename!

## Mandatory options

The following information is needed for the execution of the `list-spaces-older-than` command.
It can be provided through commandline options or environment variables.

- api_url
- org
- username
- password

## Optional options

By default days will be set to 365. By default the filter pattern is `(dev|acceptance|integration|production|performance)(-)([cdi])([0-9]{6,7})`.
To list all spaces use `--days 0 --pattern '()'`

- days
- pattern

## yml

If yml file is given, org and api are used. Rest gets ignored.

## Examples

All examples assume CF_USERNAME and CF_PASSWORD is set.

### Unix (bash)

command: `em.sh list-spaces-older-than -a https://api.cf.sap.hana.ondemand.com -o CCXCourseTesting --days 28`

### Windows (powershell)

command: `em.bat list-spaces-older-than -a https://api.cf.sap.hana.ondemand.com -o CCXCourseTesting --days 28`

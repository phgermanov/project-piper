# delete-spaces-older-than

This is the command used to delete spaces older than X days on Cloud Foundry.
There is a default filter pattern, that only spaces starting with dev|acceptance|integration|production|performance followed by -C|D|I number gets deleted. The pattern gets matched against the _lower-cased_ version of the spacename!
We suggest that you use the command [`list-spaces-older-than`](list-spaces-older-than.md) before to get a list of the spaces which will be deleted.

## Mandatory options

The following information is needed for the execution of the `delete-spaces-older-than` command.
It can be provided through commandline options or environment variables.

- api_url
- org
- username
- password

## Optional options

By default days will be set to 365. By default the filter pattern is `(dev|acceptance|integration|production|performance)(-)([cdi])([0-9]{6,7})`.
To list all spaces in an organisation use `--days 0 --pattern '()'`

- days
- pattern

## yml

If yml file is given, org and api are used. The rest is ignored.

## Examples

All examples assume CF_USERNAME and CF_PASSWORD is set.

### Unix (bash)

command: `em.sh delete-spaces-older-than -a https://api.cf.sap.hana.ondemand.com -o CCXCourseTesting --days 28 --pattern '()'`

### Windows (powershell)

command: `em.bat delete-spaces-older-than -a https://api.cf.sap.hana.ondemand.com -o CCXCourseTesting --days 28`

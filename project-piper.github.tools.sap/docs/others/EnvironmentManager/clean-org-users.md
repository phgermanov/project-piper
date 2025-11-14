# clean-org-users

This is the command removes all users from an org who only hold the orgUsers role in that org. This happens if you remove a user from all spaces (or delete all spaces of a user) and remove him from OrgManager, OrgAuditor and OrgBillingManager roles. There is no cf cli command to remove that role, but cf org-users lists users with this role. The command has an option to only list the users that would be removed.

## Mandatory options

The following information is needed for the execution of the `clean-org-users` command.
It can be provided through commandline options or environment variables.

- api_url
- org
- username
- password

## Optional options

There is the argument: `--list_only true` to disable the actual removal of the orgUser role from users.

## yml

The use of a yml file to provide api_url, org and space is possible

## Examples

All examples assume CF_USERNAME and CF_PASSWORD is set.

### Unix (bash)

command: `em.sh clean-org-users -a https://api.cf.sap.hana.ondemand.com -o CCXCourseTesting --list_only true`

### Windows (powershell)

command: `em.bat clean-org-users -a https://api.cf.sap.hana.ondemand.com -o CCXCourseTesting --list_only true`

### via piper step

step call: `manageCloudFoundryEnvironment(script:this,cfCredentialsId:'CF_CREDENTIAL', command:"clean-org-users -o cccourse -a https://api.cf.sap.hana.ondemand.com --list_only true")`

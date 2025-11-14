<!-- markdownlint-disable-next-line MD041 -->
## environment.yml

Here your find an example of an environment.yml file used to setup the independent _getMeMyToken_-space for the `setup-security` command of EnvMan.

```yml
---
cf_api_endpoint: https://api.cf.sap.hana.ondemand.com/
cf_organization: <org>
cf_space: getMeMyToken
required-services:
- instance_name: uaa-getMeMyToken
  service: xsuaa
  plan: application
  jsonPath: 'xs-security.json'
  recreate: update
user-roles:
  remove-other-users: true
  developer:
    - pXXXXXXXXXX # P-User of the CF_Credentials
```

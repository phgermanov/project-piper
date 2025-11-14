# Examples of environment YAML files with security section

## A basic setup

The following file was used during the DevX Day EnvMan Demo.
Before it is read in by EnvMan $space is substituted with the actual spacename.

Command:
`em.sh setup-environment -y security.yml`

environment.yml:

```yml
---
space: "$space"
org: <org>
api: <api_url>
required-services:
- instance_name: uaa-bulletinboard
  service: xsuaa
  plan: application
  jsonPath: 'xs-security.json'
  recreate: update
[...]
security:
  appSpace: getMeMyToken
  xsuaaName: "uaa-getMeMyToken"
  xsAppName: "bulletinboard-$space"
  createUsersIfNotFound: true
  roles:
    - name: "ViewerRole-$space"
      description: "Viewer only in $space"
      templateName: "Viewer"
    - name: "AdvertiserRole-$space"
      description: "A real Advertiser in $space"
      templateName: "Advertiser"
  roleCollections:
    - name: "ViewerRC-$space"
      description: "roleCollection with Viewer only in $space"
      roles:
        - "ViewerRole-$space"
    - name: "MasterRC-$space"
      description: "roleCollection with all Roles in $space"
      roles:
        - "ViewerRole-$space"
        - "AdvertiserRole-$space"
  userMapping:
    - user: "kai-martin.dittkrist@sap.com"
      rcName: "MasterRC-$space"
    - user: "ralf.schmitt-roquette@sap.com"
      rcName: "MasterRC-$space"
```

The file xs-security.json included xsAppName _bulletinboard-$space_ and the two RoleTemplates _Advertiser_ and _Viewer_.

## Only samlMappings

The following file is used to setup the FakeIdP user groups. Before it is read in by EnvMan $space is substituted with the actual spacename.
Here for the roleCollections only the names are given, which is enough to identify them in the samlMapping section.

Command:
`em.sh setup-security -y muenchhausen.yml -s getMeMyToken -o <org> -a <api_url>`

muenchhausen.yml:

```yml
---
security:
  xsuaaName: "uaa-getMeMyToken"
  xsAppName: "bulletinboard-$space"
  roleCollections:
    - name: "ViewerRC-$space"
    - name: "MasterRC-$space"
  samlMapping:
    - name: "Advertiser"
      rcName: "MasterRC-$space"
    - name: "Viewer"
      rcName: "ViewerRC-$space"
```

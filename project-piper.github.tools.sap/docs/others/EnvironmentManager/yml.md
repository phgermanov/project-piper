# Yml File for EnvironmentManager

Description of the structure of a yml environment file for the EnvironmentManager.

## root

| name | example value | mandatory | description |
| --- | --- |:---:| --- |
| cf_api_endpoint | <https://api.cf.sap.hana.ondemand.com/> | yes | api endpoint url |
| cf_organization | CCXCourseTesting | yes | org name |
| cf_space | emtest-c5261232 | yes | spacename |
| cf_api_get_request_timeout | 30000 | no | default 5000 (in milliseconds), rather pragmatic value is 30000 |
| required-services | [List of this](#required-services) | no | list of services |
| user-provided-service | [List of this](#user-provided-service) | no | list of user provided services |
| service-keys | [List of this](#service-keys) | no | list of service keys |
| bind-services | [Map of this](#bind-services) | no | list of service bindings |
| security | [Map of this](#security) | no | map for role collections etc. See also [security.md](security.md) for details |
| user-roles | [Map of this](#user-roles) | no | list of users for roles in space |
| activateRestageApps | true | no | default value: False; toggles possible *cf restage* of an app after a binding was made to it. |
| waitOnServiceStatus | true | no | default value: False; globally activates waitOnServiceStatus for service creation/update/deletzion |

EnvMan remembers all the apps that got a new binding during an execution and Restaging happens at the end of an execution, so each app will only be restaged once.

## required-services

(version > 0.1.0; rebinding + service key recreation on recreate true  > 0.8.1; statusCheck > 0.10.0)
List with items containing:

| name | example value | mandatory | description |
| --- | --- |:---:| --- |
| instance_name | postgres-bulletinboard-ads | yes | name of instance created |
| service | postgresql | yes | name of service |
| plan | v9.4-dev | yes | plan of service used |
| recreate | true | no | default: *false*, toggles possible deletion of existing service, possible also *update*: does a update-service if service already exists |
| jsonString | {"key":"value"} | no | string with json |
| jsonPath | jsonFile.json | no | path to a json file (relative to location of this yml)|
| tag | some tag | no | string as tag for service (-t) |
| checkStatus | false | no | default Value: `unset`; turn on or off the wait on status check for operations on this instances |
| timeout | 3600 | no | default Value: 7200; sets the timeout for the wait on status check (in seconds) |

!!! info
    On recreate:
    Default value is *false*, meaning the service will not be created if it already exists.
    If set to *update* a `cf update-service` call will be made instead of `cf create-service`. Meaning all bindings to apps will be kept.
    If recreate is set to *true*:
    A service that already exists it will be deleted and created again.
    Furthermore if any applications is bound to the service EnvMan will unbind the service from the applications and after recreation of the service instance rebind it again to the applications. For this rebinding *no* bind specific information can be used. Therefore be aware that possible bind specific information (a Json given during the original bind) will be lost. For this we recommend to also specify the binding in the [bindings section](#bind-services) with the Json in the same YML (with `recreate: true`).
    Associated is global Boolean [activateRestageApps](#root) to turn restaging of Apps on/off (a bind typically needs a restage of the app).
    If any Service Key for this instance exist, it will be deleted and recreated with the same name after the service is recreated. If the Service Key was created with a *Json*, this will Json will be *ignored*. In this case we recommend to add the service Key with the Json to the [service-keys section](#service-keys) of the yaml with parameter `recreate: true` so a desired version gets created.

!!! info
    On checkStatus and the global waitOnServiceStatus:
    By default no check of the status of the last operation of a service instance is done. You can turn this check on per instance, or for all with a global flag. In the second case you can turn it again of for individual instances.
    A status check on an instance is done when `checkStatus` is true, or `waitOnStatusCheck`is true and `checkStatus` is not false.

## user-provided-service

List with items containing:

| name | example value | mandatory | description |
| --- | --- |:---:| --- |
| ups_name | hana-aws | yes | name of the user provided service |
| jsonString | {"key":"value"} | or jsonPath | string with json |
| jsonPath | jsonFile.json | or jsonString | path to a json file (relative to location of this yml) |
| recreate | true | no | default: false, toggles possible deletion of existing service |

## service-keys

(version >0.8.1)
Contains the following items:

| name | example value | mandatory | description |
| --- | --- |:---:| --- |
| instanceName | postgres-bulletinboard-ads | yes | instance Name you want to create the service key from |
| keyName | postgres-key | yes | name to give the service key |
| jsonString | {"key":"value"} | or jsonPath | string with json |
| jsonPath | jsonFile.json | or jsonString | path to a json file (relative to location of this yml) |
| recreate | true | no | default: false, toggles possible deletion of existing service-key |

## bind-services

(version >0.8.1)
Contains the following items:

| name | example value | mandatory | description |
| --- | --- |:---:| --- |
| instanceName | postgres-bulletinboard-ads | yes | instance Name you want to create the service key from |
| appName | dummyApp | yes | name of App that should be binded to the service |
| jsonString | {"key":"value"} | or jsonPath | string with json |
| jsonPath | jsonFile.json | or jsonString | path to a json file (relative to location of this yml) |
| recreate | true | no | default: false, toggles possible deletion of existing service binding |

Associated is global Boolean [activateRestageApps](#root) to turn restaging of Apps on/off (a bind typically needs a restage of the app).

## user-roles

Contains the following items:

| name | example value | mandatory | description |
| --- | --- |:---:| --- |
| remove-other-users | true | no | flag if other users are removed, default is *false* |
| manager | List of Strings | no | list of users to be assigned role spaceManager |
| developer | List of Strings | no | list of users to be assigned role spaceDeveloper |
| auditor | List of Strings | no | list of users to be assigned role spaceAuditor |

If remove-other-users is true, EnvMan will remove all CF users of a space not specified in the yml and add those who are not yet in the space. If remove-other-users is false, it will just add all specified users & roles who are not already assigned to the space.

## security

(version >0.7.0)
Contains the following items:
!!! note
    Please see the [EnvironmentManager security section](security.md) before using it.

| name | example value | mandatory | description |
| --- | --- |:---:| --- |
| xsAppName | bulletinboard-integration | yes | value of attribute xsAppName in xs-security.json of xsuaa instance of which you want create roles for |
| xsuaaName | uaa-getMeMyToken | either this or appName | name of xsuaa instance with API scopes, has priority over appName |
| appName | getMeMyToken | either this or xsuaaName | name of App bound to xsuaa with API scopes |
| appSpace | getMeMyToken | no | name of space containing xsuaa instance with API scopes, can be used if the root argument *space* points to a different space |
| samlIdpName | customIdP1 | no | name of custom IdP for samlMapping ,default value is `xsuaa-monitoring-idp` |
| createUsersIfNotFound | true | no | default: false, Boolean value specifying if user not known in the identityzone are created or if EnvMan stops |
| roles | List of roles | no | information about roles and their *RoleTemplates*, details see below |
| roleCollections | List of RoleCollections | no | information about RoleCollections and their roles, details see below |
| samlMapping | List of samlMappings | no | mapping of RoleCollections to user groups, details see below |
| userMapping | List of users and their RCs | no | mapping of RoleCollections to users, details see below |

### roles

Contains a List of the following items:

| name | example value | mandatory | description |
| --- | --- |:---:| --- |
| name | ViewerRole | yes | name of the new role |
| description | Viewer only | no | description of the new role |
| templateName | Viewer | yes | name of RoleTemplate (defined in xs-security.json) on which the role is based |

### roleCollections

Contains a List of the following items:

| name | example value | mandatory | description |
| --- | --- |:---:| --- |
| name | MasterRC | yes | name of the new RoleCollection |
| description | Collection of all Roles | no | description of the new RoleCollection |
| roles | List of role names | no | names of a subset of all roles defined under roles (roles have to be specified in a roles section before) |

### samlMapping

Contains a List of the following items:

| name | example value | mandatory | description |
| --- | --- |:---:| --- |
| name | Advertiser | yes | name of a user group |
| rcName | MasterRC | yes | name of the RC that should be added to the user group (rcName has to be specified under roleCollections before) |

### userMapping

Contains a List of the following items:

| name | example value | mandatory | description |
| --- | --- |:---:| --- |
| user | <dummy.user@sap.com> | yes | user mail address used during login |
| rcName | MasterRC | yes | name of the RC that should be added to the user (rcName has to be specified under roleCollections before) |

## Example

[ExampleFile](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/blob/master/test/resources/EnvironmentManagerTest.yml)
[ExampleFile with Security](example-security-yml.md)
[ExampleFile used for EnvManTesting 1](https://github.wdf.sap.corp/cc-devops-envman/envman-samples-and-tests/blob/master/newSpace.yml)
[ExampleFile used for EnvManTesting 2](https://github.wdf.sap.corp/cc-devops-envman/envman-samples-and-tests/blob/master/oldSpace.yml)

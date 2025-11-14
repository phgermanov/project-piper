<!-- markdownlint-disable-next-line MD041 -->
## xs-security.json

Here your find an example of a xs-security.json file used to setup the xsuaa instance needed for the `setup-security` command of EnvMan.

```yml
{
    "xsappname"     : "getMeMyToken",
    "description"   : "get Me My Access Token for security RoleCollection Management",
    "tenant-mode"   : "shared",
    "scopes"        : [
                        { "name"                 : "xs_authorization.read",
                          "description"          : "Audit Authorizations"
                        },
                        { "name"                 : "xs_authorization.write",
                          "description"          : "Administrate Authorizations"
                        },
                        { "name"                 : "xs_user.read",
                          "description"          : "Audit Authorizations User Read"
                        },
                        { "name"                 : "xs_user.write",
                          "description"          : "Administrate Authorizations User Write"
                        }
                      ],
    "role-templates": [
                        { "name"                 : "AuthAdmin",
                          "description"          : "Administrate Authorizations",
                          "scope-references"     : [
                                                     "xs_authorization.read",
                                                     "xs_authorization.write",
                                                    "xs_user.read",
                                                    "xs_user.write"
                                                   ]
                        }
    ]
}
```

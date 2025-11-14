<!-- markdownlint-disable-next-line first-line-h1 -->
## SAP-Specifics

SAP-specific prerequisites:

- Application project available in Fortify 360 server with name `yourMavenGroupId-yourMavenArtifactId` and version `yourMavenMajorVersion`. Details can be found in the [Fortify Build Wiki page](https://wiki.wdf.sap.corp/wiki/display/CI/Fortify+Build).

- Generate an access token to enable the audit status checking.

  To generate a Fortify SSC token, you have the following options:
  - Use the [Fortify SSC UI](https://fortify.tools.sap/ssc/html/ssc) (difficult for technical users)
  - fortifyclient utility (part of the SCA distribution and also installed in the [Piper SCA docker image](https://github.wdf.sap.corp/ContinuousDelivery/piper-docker-fortify))
  - Use the [auth-token-controller API](https://fortify.tools.sap/ssc/html/docs/api-reference/index.jsp#/auth-token-controller/createAuthToken) (click the authorize button first)

- Please do not forget to grant the named user that has been used for generating the token access to your Fortify project(s)!

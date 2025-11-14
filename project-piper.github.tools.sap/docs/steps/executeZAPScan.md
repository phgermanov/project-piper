# executeZAPScan (Deprecated)

## Description

The [OWASP Zed Attack Proxy (ZAP)](https://github.com/zaproxy/zap-core-help/wiki) is a penetration testing tool, which executes some baseline attacks on your running web application. ZAP is not as easy and simple to use as you might have thought but setting up individual security checks with ZAP allows you to *automatically* apply a security baseline and ensure that commonly known vulnerabilities like SQL injection, XSS, XSRF or XXE are not applicable to your solution.

ZAP can be used to test web frontends of any technology as well as backend services providing a REST API.

!!! danger
    Do not attack any web applications that you do not own and do not attack any applications deployed into a production landscape.
    If you take down a service or platform you will have to compensate for the damage caused and take full responsibility.

This step can be executed against a deployed application instance which is reachable by the Jenkins or Jenkins as a Service Host.
ZAP is executed inside a Docker container and publishes the scan results as an HTML Report.

## Prerequisites

Deployment of an application instance into a non-production landscape

## Pipeline configuration

Default ZAP Configuration

```yml
   executeZAPScan:
    dockerImage: 'docker.wdf.sap.corp:50000/owasp/zap2docker-stable'
    dockerWorkspace: '/zap/'
    zapPort: '8090'
    stashContent:
      - zapFiles
    addonInstallList:
      - ascanrules
      - pscanrules
      - groovy
    context:
      name: ''
      user:
        name: ''
        id:
        credentialsId: ''
        userParameterName: ''
        passwordParameterName: ''
        additionalCreationParameters: {}
      forcedUserMode: false
      authentication:
        method: ''
        methodConfiguration: {}
        loggedInIndicator: ''
        loggedOutIndicator: ''
      sessionManagement:
        method: ''
        methodConfiguration: {}
      authorization:
        headerRegex: ''
        bodyRegex: ''
        statusCode: ''
        logicalOperator: 'AND'
    scanner:
      scanMode: 'scan'
      inScope: 'true'
      recurse: 'true'
      additionalParameters: ''
      activeScan:
        enabled: true
        method: 'GET'
        header: ''
        postData: ''
      ajaxSpiderScan:
        enabled: true
      spiderScan:
        enabled: true
      passiveScan:
        enabled: true
        scanners: []
    alertThreshold: -1
    suppressedIssues: []
    verbose: false
```

These settings are required for running the docker container and ZAP inside of it. For more details on the parameters themselves please refer to the table below.

!!! info
    ZAP supports ZEST scripts as well as Rhino/Nashorn Javascripts to be used for different purpose i.e. supporting more complex authentication scenarios like SAML or oAuth or for customizing the HTTPfuzzer. In addition to that ZAP may also require a context file that i.e. binds the user into the authentication script. You can place such `scripts` and `context` files into your project's root directory within a folder named `zap`.

```text
       zap
       |
       |\--scripts
       |   |
       |   |\--authentication
       |   |   |
       |   |   \--myCustomAuth.js
       |   |
       |   \--httpsender
       |      |
       |      \--mySender.js
       \--context
          |
          \--Default.context
```

ZAP follows a specific purpose based subfolder structure for scripts i.e. `zap/scripts/authentication` would be the place to put you authentication scripts while `zap/scripts/httpsender` is the place for HTTPSender scripts. Please use the ZAP Script Editor and available documentation for more details.

## Explanation of pipeline step

The targetUrls list is a mandatory parameter, which contains all starting points of the target application. All other parameters will be used as provided in the default Piper configuration and can be overwritten on different levels as provided by the configuration framework.

Usage of pipeline step:

```groovy
def urls = [
             "https://url.one",
             "https://url.two",
             "https://url.three"
           ]
executeZAPScan targetUrls: urls, verbose: true
```

Available parameters:

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| script | no | empty `globalPipelineEnvironment` | |
| dockerImage | no | 'docker.wdf.sap.corp:50000/owasp/zap2docker-stable' | |
| dockerWorkspace | no | '/zap/' | |
| zapPort | no | `8090` | Any unassigned port |
| targetUrls| yes | | [`https://myvulnerableoffering.cfapps.sap.hana.ondemand.com`] |
| stashContent | no | `- zapFiles` | List of stashes, if empty whole workspace is used |
| addonInstallList | no | - ascanrules<br />- pscanrules<br />- groovy | List of addons to install before scanning |
| context.name | no | '' | `'Default'` - The name of the context as defined in the context file if supplied, if no context file is supplied a context will be created on the fly |
| context.user.name | no | `''` | `'technical user'` - The name of the user that has been defined in the ZAP context or should be created on the fly |
| context.user.id | no | `null` | `0` or `1`... - The ID of the user that has been defined in the ZAP context, to be left empty in case user shall be created on the fly |
| context.user.credentialsId | no | `''` | Jenkins credentials ID (Username/Password) for the user to be created on the fly |
| context.user.userParameterName | no | `''` | Name of the parameter the user name value is bound to in the authentication script i.e. `'client_id'` or `'user'`. Only relevant for on the fly context creation. |
| context.user.passwordParameterName | no | `''` | Name of the parameter the password value is bound to in the authentication script i.e. `'client_secret'` or `'password'`. Only relevant for on the fly context creation. |
| context.user.additionalCreationParameters | no | `{}` | Highly depends on the authentication method. Only relevant for on the fly context creation. |
| context.forcedUserMode | no | `false` | `true`, `false` - Whether to enforce an authenticated user on every request or not. Also relevant for contexts loaded from a file. |
| context.authentication.method | no | `''` | Usually one of `'formBasedAuthentication'`, `'scriptBasedAuthentication'`, `'httpAuthentication'`, `'manualAuthentication'`. Only relevant for on the fly context creation. |
| context.authentication.methodConfiguration | no | `{}` | Highly depends on the authentication method you choose i.e. `['scriptName': 'oAuthClientCredentialsFlow.js', 'API URL': 'https://some-account.authentication.sap.hana.ondemand.com/oauth/token']` for `'scriptBasedAuthentication'` as it is part of the samples. Only relevant for on the fly context creation. |
| context.authentication.loggedInIndicator | no | `''` | A regular expression that matches the page/response in case the user is logged in i.e. `'\\QLog out\\E'`. Only relevant for on the fly context creation. |
| context.authentication.loggedOutIndicator | no | `''` | A regular expression that matches the page/response in case the user is logged out i.e. `'\\QWelcome guest, please log into your\\E'`. Only relevant for on the fly context creation. |
| context.authorization.headerRegex | no | `''` | A regular expression used to identify a server response following an unauthorized request based on the response headers. Only relevant for on the fly context creation. |
| context.authorization.bodyRegex | no | `''` | A regular expression used to identify a server response following an unauthorized request based on the response body. Only relevant for on the fly context creation. |
| context.authorization.statusCode | no | `''` | HTTP status code value used to identify a server response following an unauthorized request. Only relevant for on the fly context creation. |
| context.authorization.logicalOperator | no | `'AND'` | `'OR'` or `'AND'` - Locical operator used for above context.authorization conditions. Only relevant for on the fly context creation. |
| scanner.scanMode | no | `'scan'` | `'scan'`, `'scanAsUser'` - Unauthenticated / authenticated scenario based on a user context |
| scanner.inScope | no | `true` | `true`, `false` - Whether to just scan URLs relative to the startiung URLs or beyond |
| scanner.recurse | no | `true` | `true`, `false` - Whether to recurse into deeper structures or just to scan the top level URLs |
| scanner.additionalParameters | no | `''` | `'&whateverZapScanParameter=value&anotherScanParameter=value'` - Additional parameters to be supplied to the URL when invoking the scan action |
| scanner.activeScan.enabled | no | `true` | `true`, `false` |
| scanner.activeScan.method | no | `'GET'` | `'GET'`, `'POST'` - The HTTP verb being used for any of the active scan requests |
| scanner.activeScan.header | no | `''` | \|-<br />User-Agent: Mozilla/5.0 (Windows NT 6.3; WOW64; rv:39.0) Gecko/20100101 Firefox/39.0<br />Pragma: no-cache<br />Cache-Control: no-cache<br />Content-Type: application/xml |
| scanner.activeScan.postData | no | `''` | \|- `&lt;?xml version="1.0" encoding="UTF-8"?&gt;<br /><msg:RequestMessage xmlns:msg="http://iec.ch/TC57/2011/schema/message"><msg:Header><msg:Verb>create</msg:Verb><msg:Noun>MeterConfig</msg:Noun></msg:Header><msg:Payload><m:MeterConfig xmlns:m="http://iec.ch/TC57/CIM-c4e#"><m:Meter><m:amrSystem>METER_AMR_SYSTEM</m:amrSystem><m:isVirtual>false</m:isVirtual><m:serialNumber>Serial-10137616</m:serialNumber><m:timeZone>UTC</m:timeZone><m:timeZoneOffset>0</m:timeZoneOffset><m:ConfigurationEvents><m:effectiveDateTime>2018-08-02T09:35:59.442Z</m:effectiveDateTime></m:ConfigurationEvents><m:EndDeviceInfo><m:AssetModel><m:modelNumber>Model-DC_AMI_LL</m:modelNumber><m:Manufacturer><m:name>o3-Telefonica.test</m:name></m:Manufacturer></m:AssetModel></m:EndDeviceInfo></m:Meter></m:MeterConfig></msg:Payload></msg:RequestMessage>` |
| scanner.ajaxSpiderScan.enabled | no | `true` | `true`, `false` |
| scanner.spiderScan.enabled | no | `true` | `true`, `false` |
| scanner.passiveScan.enabled | no | `true` | `true`, `false` |
| scanner.passiveScan.scanners | no | `[]` | A list of specific scanners |
| alertThreshold | no | `-1` | Any positive `Integer` |
| suppressedIssues | no | `[]` | - cwe: '200'<br />&nbsp;&nbsp;url: '<https://eds-dev.cloudforenergy.cfapps.sap.hana.ondemand.com/api/v1/core>'<br />&nbsp;&nbsp;reasoning: 'false positive' |
| verbose | no | `false` | `true`, `false` |

Details:

* Docker container is being started and configured
* Spider Scan and AJAX Spider Scan should be used for UI scenarios whereas an Active Scan can usually be used to scan the backend API (also via POST requests)
* Each provided target URL will be spidered and attacked by default, in case of a web UI also URLs referenced by the initial web pages at attacked
* After attacking the application a report is generated providing additional information about detected issues and potential mitigation options
* The log will also provide details on the requests that have been issued as well as the responses received
* Parameters `'threshold'`, and `'suppressedIssues'` can be used to apply a limit that causes the build to fail and to suppress false positives to avoid them being considered for threshold analysis.
* Parameters `'scanner.activeScan.method'`, `'scanner.activeScan.header'`, and `'scanner.activeScan.postData'` can be used to modify the contents of the attack request being sent when actively scanning a backend API
* To scan in an authenticated scenario switch `'scanner.scanMode'` to `'scanAsUser'` and supply related scripts and context files to allow ZAP to properly authenticate the provided user before scanning your application's UI / API
* You can use `'context.user.credentialsId'`, `'context.user.name'`,`'context.user.userParameterName'`, `'context.user.passwordParameterName'`, `'context.authentication.method'`, `'context.authentication.methodConfiguration'`, `'context.authentication.loggedInIndicator'`, and `'context.authentication.loggedOutIndicator'` to create and populate a context on the fly

!!! caution "Protecting user credentials being used"
    If losing the credentials being used in your test is a major risk for your scenario, please go for the secondary option and use the Jenkins credentials store and let the context and the user be created on the fly. Storage of credentials within a context file cannot be regarded as secure since they are not encrypted.

Example:

Assumed the following HTML document is available at the URL
`www.target.de` and specified in the targetUrls array.

```html
<body>
    <a href="https://www.target.de/subresource/134599.html" title="Go to subpage"/>
    <a href="https://www.google.de/" title="Go to Google"/>
</body>
```

The executeZAPScan step, downloads and scans the referenced documents in the same scope and would search also for containing href tags.
So the first referenced html document will be downloaded and analyzed, the second referenced document not, because the root target URL does not match.

!!! caution "Add not spiderable urls to the targetUrls array"
    If your applications html documents that do not contain any hrefs to specific URLs you have to add them manually
    e.g. API URLs which are accessed via AJAX-calls.

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|addonInstallList||X|X|
|alertThreshold||X|X|
|context||X|X|
|dockerImage||X|X|
|dockerWorkspace||X|X|
|scanner||X|X|
|stashContent||X|X|
|suppressedIssues||X|X|
|targetUrls||X|X|
|verbose||X|X|
|zapPort||X|X|

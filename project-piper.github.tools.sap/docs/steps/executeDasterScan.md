# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

The tokens have to be created via the [DASTer Web UI](https://app.daster.tools.sap/ui5/#/tokens) and have to be stored in the Jenkins credentials store as a Secure text/Username password credential depending on their type. The related ID is to be supplied as value to a configuration parameter named like the original settings parameter with the suffix `CredentialsId`. For more details on storing the credential please refer to [executeBuild step documentation](../build/xMake.md#parameterized-remote-trigger-plugin).<br /><br />
Since DAST tools impose a certain risk to DoS the application under test or platform below when generating a DASTer token your consent to the tool's terms of use is recorded and stored.

## Example

Usage of pipeline step (FioriDAST scan):

```groovy
executeDasterScan(
    script: this,
    scanType: 'fioriDASTScan',
    settings: [
       dasterTokenCredentialsId: 'daster_fiorigoat',
       targetUrl: 'http://myhost:8000/ui?sap-client=001&sap-language=EN#FioriGoat-open?returnUrl=/ui&configUrl=/sap/bc/ui5_ui5/sap/zfiorigoat/config/config.js',
       userCredentialsCredentialsId: 'daster_fiotigoat_usercredentials',
       recipients: "mymail@sap.com"
    ]
)
```

## ${docGenParameters}

## ${docGenConfiguration}

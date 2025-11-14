# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

Create token via the [DASTer Web UI](https://app.daster.tools.sap/ui5/#/tokens) and store it in the Jenkins credentials store as a Secret text. The related ID is to be supplied as value to a configuration parameter named `dasterTokenCredentialsId`.

Since DAST tools impose a certain risk to DoS the application under test or platform below when generating a DASTer token your consent to the tool's terms of use is recorded and stored.

## Example

To run the DASTer step, developers can follow the example below. This format can be given in the Jenkinsfile.

Example:

```groovy
sapDasterExecuteScan(
    script: this,
    scanType: 'fioriDASTScan',
    dasterTokenCredentialsId: 'daster_fiorigoat',
    targetUrl: 'http://myhost:8000',
    userCredentials: 'daster_fiorigoat_usercredentials',
    settings: [
       recipients: "mymail@sap.com"
    ]
)
```

For further reference, see the example of the Piper general purpose Jenkins pipeline setup at: [Github repository](https://github.tools.sap/Security-Testing/daster-piper-pipeline-ref)

## ${docGenParameters}

## ${docGenConfiguration}

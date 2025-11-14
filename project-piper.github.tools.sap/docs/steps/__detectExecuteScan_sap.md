<!-- markdownlint-disable-next-line first-line-h1 -->
## SAP-Specifics

Further SAP-specific documentation:

- Please see [*How-to-configure-Black-Duck-for-scan*](https://github.wdf.sap.corp/SynopsysHUBSupport/GeneralSupport/wiki/Black-Duck-Scan-Configuration#How-to-configure-Black-Duck-for-scan) for details about how to get the API token.
- You find background information about the scan in the Jam group [Black Duck Support Center](https://jam4.sapjam.com/groups/IkjN0Yvx8okwEHVmsixFMc/overview_page/ajslrilwXb3YWKBoZpWH6r)
- If you are setting up a custom pipeline and configuring `sapCheckPPMSCompliance` step, please make sure to use [detectServerUrl parameter](sapCheckPPMSCompliance.md#detectserverurl) of that step *OR* use `detectServerUrl` alias from `detectExecuteScan` step [parameter](detectExecuteScan.md#serverurl)

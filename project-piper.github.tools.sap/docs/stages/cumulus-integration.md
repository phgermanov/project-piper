# Cumulus integration

[Cumulus](https://wiki.wdf.sap.corp/wiki/x/nCu1g) is integrated in Piper's general purpose pipeline. If you configure the [sapCumulusUpload](../steps/sapCumulusUpload.md) step properly, your pipeline will upload test reports, configuration files and more to Cumulus to cover compliance standards automatically.
Depending on the [Release Status](https://wiki.one.int.sap/wiki/x/rrvpy), Cumulus will lock specific files of the pipeline run to prevent them from being overwritten.
The [Release Status](https://wiki.one.int.sap/wiki/x/rrvpy) `Promoted` is automatically set using the piper [Promote](./gpp/promote.md) stage.
Other `Release Status` can be set [automatically](https://wiki.one.int.sap/wiki/x/tbvpy), by uploading a the `release-status-<timestamp>.json` file i.e. in a [Piper Extension](../extensibility.md)

## Configuration of sapCumulusUpload

There are multiple ways to enable the Cumulus upload for the general purpose pipeline or to use Cumulus in a custom pipeline

You can **set up a fresh Piper general purpose pipeline**

- via [Hyperspace Onboarding](https://hyperspace.tools.sap/docs/features_and_use_cases/use_cases/), or
- via the [Hyperspace Portal](https://pages.github.tools.sap/hyperspace/cicd-setup-documentation/getting-started.html) (Jenkins pipelines only).

In this case, the Cumulus credentials are already stored in Vault and the `pipelineId` is prefilled in the `.pipeline/config.yml`

You can **get access to use Cumulus in your custom pipeline**

- by [creating an extensible pipeline](https://hyperspace.tools.sap/docs/features_and_use_cases/use_cases/pipeline-extensible/) in Hyperspace Onboarding and adding the Cumulus Storage service to it. The [Store compliance artifacts](https://hyperspace.tools.sap/docs/features_and_use_cases/use_cases/pipeline-extensible/custom-build.html#store-compliance-artifacts) section describes how to store credentials in Vault and how to further extend you pipeline with more Hyperspace services.
- (Coming soon!) by [creating a Component](https://pages.github.tools.sap/hyperspace/cicd-setup-documentation/how-tos/store-compliance-artifacts.html) in your Hyperspace Portal Project.

Further information:

- Piper step: [sapCumulusUpload](../steps/sapCumulusUpload.md)

## Usage

When using the general purpose pipeline no additional configuration is needed.

By default, the general purpose pipeline uses a commit ID ([useCommitIdForCumulus](../steps/sapCumulusUpload.md#usecommitidforcumulus)) as a location for uploading files. If you don’t want to use commit ID as a location, please set the useCommitIdForCumulus parameter to false. If the parameter is set to false, the sapCumulusUpload step relies on the [version](../steps/sapCumulusUpload.md#version) parameter for determining the uploading location.

There is the [bucketLockedWarning](../steps/sapCumulusUpload.md#bucketlockedwarning) parameter which, by default, is set to true for the general purpose pipeline. If the sapCumulusUpload step is trying to rewrite files that are locked in the Cumulus bucket, the step logs a warning message, instead of returning an error and breaking pipeline execution. If you want to have an error and break pipeline execution in such cases, please set the bucketLockedWarning parameter to false.

If extensions are used where files are generated, that are not uploaded automatically (either because the related step regarding the file pattern is not used or the pattern does not match), the `sapCumulusUpload` can be called manually in your extension. Please refer to the Cumulus [Features & Use Cases](https://wiki.wdf.sap.corp/wiki/x/toR8hw)  documentation to find out how.

The step is resilient, so no error will be thrown if the specified pattern does not match any files. Besides, the following features are supported:

- **doublestar syntax:** Support of recursive file pattern matching using the "doublestar syntax", e.g. `**/TEST-*.xml` for all XML files with "TEST-" prefix in all subfolders

- **multiple patterns:** Multiple valid file patterns can be passed comma-separated in the same call, e.g. `'**/TEST-A-*.xml, **/TEST-C-*.xml'`

## Overview of automated uploads using the general purpose pipeline

The following table shows the current implementation of the Cumulus upload integration in the general purpose pipeline. The folder in Cumulus for the files to be stored is represented by the `stepResultType` parameter.

| Stage | Step | File pattern | stepResultType |
|-------|------|--------------|----------------|
| **Build** | executeBuild | `**/target/surefire-reports/*.xml` | junit |
| | | `**/jacoco.xml` | jacoco-coverage |
| | | `**/target/coverage/**/cobertura-coverage.xml` | cobertura-coverage |
| | | `**/requirement.mapping` | requirement-mapping |
| | | `**/xmake_stage.json` | xmake |
| |executeSonarScan|`sonarscan-result.json`|sonarqube|
| |sapExecuteApiMetadataValidator|`api-metadata-validator-results.json`|api-metadata-validator|
| **Additional Unit Tests** | karmaExecuteTests | `**/TEST-*.xml`, `**/target/coverage/**/cobertura-coverage.xml` | karma |
| | npmExecuteScripts | `**/TEST-*.xml` | junit |
| |  | `**/target/coverage/**/cobertura-coverage.xml` | cobertura-coverage |
| |  | `**/e2e/*.json` | cucumber |
| **Integration** |mavenExecuteIntegration|`**/requirement.mapping`|requirement-mapping|
| **Acceptance** |newman/uiVeri5/gauge|`**/TEST-*.xml`|acceptance-test|
| |sapCreateTraceabilityReport|`**/piper_traceability*`|traceability-report|
| | |`**/requirement.mapping`|requirement-mapping|
| | |`**/delivery.mapping`|delivery-mapping|
| **Security** |executeFortifyScan|`**/*.fpr`, `**/fortify-scan.*`, `**/*.PDF,**/toolrun_fortify_*.json,**/piper_fortify_report.json,**/fortify/*.sarif,**/fortify/*.sarif.gz`|fortify|
| |executeCheckmarxScan|`**/CxSASTReport_*.pdf`, `**/*CxSAST*.html`, `**/ScanReport.*,**/toolrun_checkmarx_*.json,**/piper_checkmarx_report.json,**/checkmarx/*.sarif`|checkmarx|
| |executeOpenSourceDependencyScan|`**/*BlackDuck_RiskReport.pdf, **/detectExecuteScan_policy_*.json, **/piper_detect_vulnerability_report.html, **/toolrun_detectExecute_*.json, **/piper_detect_vulnerability.sarif, **/piper_hub_detect_sbom.xml`|blackduck-security|
| | |`**/whitesource-riskReport.pdf, **/toolrun_whitesource_*.json, **/piper_whitesource_vulnerability_report.html, **/piper_whitesource_vulnerability.sarif, **/piper_whitesource_sbom.xml`|whitesource-security|
| | |`protecode_report.pdf`|protecode|
| |executeZAPScan|`zap_report.html`|zap|
| **IP Scan & PPMS** |whitesourceExecuteScan|`**/whitesource-ip.json, **/whitesource-riskReport.pdf, **/toolrun_whitesource_*.json`|whitesource-ip|
| |detectExecuteScan|`**/*BlackDuck_RiskReport.pdf, **/blackduck-ip.json, **/toolrun_detectExecute_*.json, **/piper_detect_policy_violation_report.html`|blackduck-ip|
| |sapCheckPPMSCompliance|`**/piper_whitesource_ppms_report.*`|whitesource-ip|
| | |`**/piper_blackduck_ppms_report.*`|blackduck-ip|
| **Post Actions** |sapPiperPipelineStagePost|`**/cumulus-configuration.json`, `env.json`, `lock-run.json`|*root*|
| | |`jenkins-log.txt` (Jenkins only)|log|

## Further Information

- [Hyperspace Onboarding documentation for Cumulus](https://hyperspace.tools.sap/docs/features_and_use_cases/connected_tools/cumulus.html)

- [Set up Cumulus Wiki](https://wiki.wdf.sap.corp/wiki/x/3z5mh)

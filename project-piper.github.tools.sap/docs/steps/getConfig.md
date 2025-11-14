# getConfig

## Description

This step allows to resolve the piper configuration.
It is only possible to be used via the command line ,via Azure or Gihub Actions

## Example

!!! tip ""

    === "Jenkins"

        *coming soon*

    === "Azure DevOps"

        ```yml
        steps:
        - task: piper@1
          name: getConfig
          inputs:
            stepName: getConfig
            flags: "--outputFile theConfig.json"
        ```

    === "GitHub Actions"

        [Piper's Github Action](https://github.com/SAP/project-piper-action/tree/main) can be used to call this step.
        Please note that using this step as a standalone will only work if the proper environment is set up. Following example demonstrates how the step can be called as a standalone by setting the environment and using publicly available Github action "checkout".


        ```yml
        jobs:
          config:
            runs-on: self-hosted
            env:
              PIPER_ACTION_PIPER_VERSION: 'latest'
              PIPER_ACTION_SAP_PIPER_VERSION: 'latest'
              PIPER_ACTION_SAP_PIPER_OWNER: project-piper
              PIPER_ACTION_SAP_PIPER_REPOSITORY: sap-piper
              PIPER_ACTION_GITHUB_ENTERPRISE_TOKEN: ${{ secrets.GITHUB_TOKEN }}
            steps:
              - name: Checkout to current repository
                uses: actions/checkout@eef61447b9ff4aafe5dcd4e0bbf5d482be7e7871 # v4.2.1
              - name: Read stage configuration
                id: stage_config
                uses: SAP/project-piper-action@f8d534f183edbc21e42dc403abe8eb63592e4afd # v1.9.0
                with:
                  step-name: getConfig
                  flags: --stageConfig --outputFile stageConfig.json
        ```

    === "Command Line"

        ```sh
        piper getConfig --help
        ```

## Usage

```text
Usage:
  piper getConfig [flags]

Flags:
      --contextConfig                           Defines if step context configuration should be loaded instead of step config
  -h, --help                                    help for getConfig
      --output string                           Defines the output format (default "json")
      --outputFile string                       Defines a file path. f set, the output will be written to the defines file
      --stageConfig                             Defines if step stage configuration should be loaded and no step-specific config
      --stageConfigAcceptedParams stringArray   Defines the parameters used for filtering stage/general configuration when accessing stage config
      --stepMetadata string                     Step metadata, passed as path to yaml
      --stepName string                         Step name, used to get step metadata if yaml path is not set

Global Flags:
      --correlationID string    ID for unique identification of a pipeline run
      --customConfig string     Path to the pipeline configuration file (default ".pipeline/config.yml")
      --defaultConfig strings   Default configurations, passed as path to yaml file (default [.pipeline/defaults.yaml])
      --envRootPath string      Root path to Piper pipeline shared environments (default ".pipeline")
      --gitHubTokens strings    List of entries in form of <hostname>:<token> to allow GitHub token authentication for downloading config / defaults
      --ignoreCustomDefaults    Disables evaluation of the parameter 'customDefaults' in the pipeline configuration file
      --logFormat string        Log format to use. Options: default, timestamp, plain, full. (default "default")
      --noTelemetry             Deprecated flag. Has no effect. Please don't use it.
      --parametersJSON string   Parameters to be considered in JSON format
      --stageName string        Name of the stage for which configuration should be included
      --stepConfigJSON string   Step configuration in JSON format
      --vaultNamespace string   The vault namespace which should be used to fetch credentials
      --vaultPath string        The path which should be used to fetch credentials
      --vaultServerUrl string   The vault server which should be used to fetch credentials
  -v, --verbose                 verbose output
```

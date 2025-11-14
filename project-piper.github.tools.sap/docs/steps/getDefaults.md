# getDefaults

## Description

This step allows to fetch Piper (default) configuration files.
As opposed to [getConfig](getConfig.md), the files are not resolved into one configuration, but are kept raw.
It is only possible to be used via the command line via Azure and Github Actions.

## Example

!!! tip ""

    === "Jenkins"

        *coming soon*

    === "Azure DevOps"

        ```yml
        steps:
        - task: piper@1
          name: getDefaults
          inputs:
            stepName: getDefaults
            flags: "--outputFile theConfig.json"
        ```

    === "GitHub Actions"

        ```yml
        steps:
          - name: Read default stage configuration
            id: default_stage_config
            uses: SAP/project-piper-action@f8d534f183edbc21e42dc403abe8eb63592e4afd # v1.9.0
            with:
              step-name: getDefaults
              flags: --defaultsFile <file-location/>piper-defaults.yml --outputFile default-config.json

        ```

    === "Command Line"

        ```sh
        piper getDefaults --help
        ```

## Usage

```text
Usage:
  piper getDefaults [flags]

Flags:
      --defaultsFile stringArray   Defines the input defaults file(s)
  -h, --help                       help for getDefaults
      --output string              Defines the format of the configs embedded into a JSON object (default "yaml")
      --outputFile string          Defines the output filename
      --useV1                      Input files are CRD-style stage configuration

Global Flags:
      --correlationID string        ID for unique identification of a pipeline run (default "n/a")
      --customConfig string         Path to the pipeline configuration file (default ".pipeline/config.yml")
      --defaultConfig strings       Default configurations, passed as path to yaml file (default [.pipeline/defaults.yaml])
      --envRootPath string          Root path to Piper pipeline shared environments (default ".pipeline")
      --gcpJsonKeyFilePath string   File path to Google Cloud Platform JSON key file
      --gcsBucketId string          Bucket name for Google Cloud Storage
      --gcsFolderPath string        GCS folder path. One of the components of GCS target folder
      --gcsSubFolder string         Used to logically separate results of the same step result type
      --gitHubTokens strings        List of entries in form of <hostname>:<token> to allow GitHub token authentication for downloading config / defaults
      --ignoreCustomDefaults        Disables evaluation of the parameter 'customDefaults' in the pipeline configuration file
      --logFormat string            Log format to use. Options: default, timestamp, plain, full. (default "default")
      --noTelemetry                 Deprecated flag. Has no effect. Please don't use it.
      --parametersJSON string       Parameters to be considered in JSON format
      --stageName string            Name of the stage for which configuration should be included
      --stepConfigJSON string       Step configuration in JSON format
      --vaultNamespace string       The Vault namespace which should be used to fetch credentials
      --vaultPath string            The path which should be used to fetch credentials
      --vaultServerUrl string       The Vault server which should be used to fetch credentials
  -v, --verbose                     verbose output
```

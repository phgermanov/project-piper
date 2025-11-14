# Extensibility

It is possible to extend Piper workflows in different ways.

## Reuse of Predefined Stages

It is possible to use the stages that we offer together with your own custom stages as shown in [this example](https://github.tools.sap/project-piper/piper-pipeline-github/blob/main/examples/custom_workflow/extending_pipeline.yml).

## Extensibility of Existing Stages

Also, Piper\'s general purpose pipeline stages can be extended via [composite actions](https://docs.github.com/en/actions/creating-actions/creating-a-composite-action). The stages that support extensibility include:
`Build`, `Acceptance`, `Integration`, `Performance`, `Release`, `Post`, `OSS` and `PPMS`.

To enable extensions in Piper's general purpose pipeline, set the `extensibility-enabled` input to `true` in your project's workflow file.

Example configuration in your project workflow file:

```yaml
jobs:
  piper:
    uses: project-piper/piper-pipeline-github/.github/workflows/sap-piper-workflow.yml@main
    with:
      extensibility-enabled: true
    secrets: inherit
```

An extension consists of a folder containing a composite action (`action.yml`) and, if needed, some additional helper files. The folder name should follow the format `[pre|post]<stage name>`, where `pre` or `post` indicating whether it is a pre-stage or post-stage extension. `<stage name>` should correspond to the name of the stage you intend to extend. Since extensibility is implemented via composite actions, the name of the composite action file must be `action.yml` (or `action.yaml`).

```text
postBuild
├── action.yml
└── script.sh
preAcceptance
└── action.yml
```

Extensions can be *local* or *global* (remote), with local extensions taking precedence over global ones.

### Local Extensions

Local extensions should be placed in the `.pipeline/extensions` folder. The structure would look like this:

```text
.
└── .pipeline
    └── extensions
        ├── postBuild
        │   ├── action.yml
        │   └── script.sh
        └── preAcceptance
            └── action.yml
```

### Global extensions (Central repository with extensions)

Global extensions should be located in a dedicated repository. This dedicated repository should be hosted on the same GitHub instance as the one where the pipeline runs.

To configure Piper to look for global extensions, set the `globalExtensionsRepository` parameter in the general section of your repository's `.pipeline/config.yml` file. This parameter expects a value in the `{owner}/{repository}` format. Optionally, you may specify a branch of your global extension repository.

```yaml
general:
  globalExtensionsRepository: your-org-name/my-global-extensions
  globalExtensionsRef: yourBranchName # Optional, if not specified the default branch will be used
```

### Example

This extension is designed to execute a shell script using the shellExecute Piper step before the Build stage begins.

- Location: `.pipeline/extensions/preBuild`
- Folder Name: `preBuild`
- Main Components:
  - **action.yml**: The composite action file defining the execution of the shell script.
  - **script.sh**: The shell script to be executed.

Ensure the extension is structured as follows:

```txt
.
└── .pipeline
    └── extensions
        └── preBuild
            ├── action.yml
            └── script.sh
```

action.yml:

```yaml
name: 'PreBuild'
runs:
  using: "composite"
  steps:
    - name: shellExecute
      uses: SAP/project-piper-action@v1
      with:
        step-name: shellExecute
        flags: --sources .pipeline/extensions/preBuild/script.sh
      shell: bash
```

## Common issues

It is not needed to add a checkout step (`actions/checkout@v4`) when extending existing stages, as it's already present in the workflow definition of the stage, prior to the extensions.
Adding this step to your extension can break your workflow.

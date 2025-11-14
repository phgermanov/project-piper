# Testing changes

There is a workflow that runs a few test pipelines (amongst which the reference pipelines) [here](.github/workflows/sample_pipeline_tests.yml), which can be triggered by commenting `/test` in a PR.
Additional pipelines can easily be added in there by adding them to the `PIPELINE_NAMES` and `PIPELINE_ORGS` environment variables.

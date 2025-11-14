<!-- markdownlint-disable-next-line first-line-h1 -->
## SAP-Specifics

Further SAP-specific documentation:

- SAP internal [wiki](https://go.sap.corp/bdba) is available for Black Duck Binary Analysis (Protecode) for more information.
- Please note: Black Duck Binary Analysis (Protecode) is a different tool from Black Duck and scans are performed on the final binary or executable generated from a build. Both tools are 3rd party tools provided by the same vendor and we request you to treat these two tools as separate scanning tools.
- New BDBA group or new technical user creation can be done via [self-service](https://go.sap.corp/oss_shp_docu_protecode).
- In case of `scanImage` or `dockerImage` and the step still tries to pull and save it via docker daemon, please make sure your JaaS environment has the variable `ON_K8S` declared and set to `true`.

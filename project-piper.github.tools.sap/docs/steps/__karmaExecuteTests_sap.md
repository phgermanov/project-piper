<!-- markdownlint-disable-next-line first-line-h1 -->
## SAP-Specifics

Further SAP-specific documentation:

Due to the inkompatability of Karma (deprecated since early 2023) and Selenium it is recommend to use [a combined image](https://github.com/SAP/devops-docker-node-browsers) (without Selenium) instead of the side-by-side images of Node and Selenium as configured by default.

To achieve this,

- remove the dependency [`@sap/piper-karma-config`](https://github.tools.sap/project-piper/piper-karma-config) from the node dependencies and remove `require("@sap/piper-karma-config")(...` from the `karma.conf.js`
- and remove and `customLaunchers` definition using `base: "WebDriver"`
- and set the `dockerImage` in your Piper config to a tag of [ppiper/node-browsers](https://hub.docker.com/r/ppiper/node-browsers/tags) and set the sidecarImage in Piper to `''`:

```yml
steps:
  karmeExecuteTests:
    dockerImage: 'ppiper/node-browsers:latest'
    sidecarImage: ''
```

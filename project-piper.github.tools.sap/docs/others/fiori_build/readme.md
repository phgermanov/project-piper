# Building Fiori apps for deployment on SCP Cloud Foundry

If you are building Fiori applications, this documentation might be interesting for you.
Some parts may be relevant to the Piper pipeline, others are just Build-specific and will work with any build pipeline setup.

## Disclaimer

The documentation below refers to *stable and robust parts*, but some parts are known to be *in flux and instable*, requiring temporary work-around.
Eventually the *WebIDE best practice build* aka *Evo build* is expected to provide relief for the work-arounds, once available.

This documentation is considered a *cooking recipe*, providing links to more exhaustive documentation, rather than repeating details here redundantly.

It tries to be a comprehensive collection of best-practices, tips & tricks & pitfalls, helpful side tracks and detours.
And it gives you templates along with explanations on how to use and how to adapt.

## Prerequisites, Tools, Knowledge

Recommended know-how to understand and adapt the steps described below is

- node.js [basics](https://www.w3schools.com/nodejs/default.asp)
- [Grunt](https://gruntjs.com/getting-started)
- [Maven](https://maven.apache.org/guides/getting-started/maven-in-five-minutes.html)
- [MTA Build Tool](https://wiki.wdf.sap.corp/wiki/display/CXP/MTA+Build+Tool)
- [Karma](https://karma-runner.github.io/2.0/index.html)

## Building

Examples

- [Grunt Build Module](https://github.wdf.sap.corp/DevX/devx-grunt)
- [Fiori build with Grunt](https://github.wdf.sap.corp/ReuseModel-Test/CountriesApp/blob/master/CountryUIV2/Gruntfile.js)
- [Excise Duty](https://github.wdf.sap.corp/ICDCloudArchitcture/excise_duty/tree/feature/adoptAppRepoProjectRoomChanges)
  - [npm install](https://github.wdf.sap.corp/ICDCloudArchitcture/excise_duty/blob/feature/adoptAppRepo/site-app-conf-exciseduty/package.json#L10)
  - [Grunt task](https://github.wdf.sap.corp/ICDCloudArchitcture/excise_duty/blob/feature/adoptAppRepo/site-app-conf-exciseduty/Gruntfile.js)

## Testing

For testing see

- [How to enable QUnit](testing-qunit.md)
- [How to enable OPA5](testing-opa5.md)
- [How to enable UIVeri5](testing-uiveri5.md)
- [How to enable UI Test Code Coverage](testing-code-coverage.md)

For code checks see

- [How to enable ESLint (S/4 rules, SAP-static-code-checks)](checking-eslint.md)

## Packaging

### How to enable L0 change merges

TODO - does this only become available with **One FLP** roadmap?

### How to enable "Changes Bundle" (Fiori Elements UI Adaptation)

TODO - does this only become available with **One FLP** roadmap?

### How to enable Minification

Use the [grunt-openui5](https://github.com/SAP/grunt-openui5) plugin.

### How to enable Obfuscation

TODO

### How to enable Preload Packaging

Use the [grunt-openui5](https://github.com/SAP/grunt-openui5) plugin.

A [grunt-sapui5-bestpractice-build](https://github.wdf.sap.corp/DevX/devx-grunt) plugin is available to perform this task for a standard sapui5 application structure.

## Security & IP Compliance

### How to enable WhiteSource

See documentation of executeWhiteSourceScan

### How to enable CheckMarx

See documentation of executeCheckmarxScan)

### How to enable SourceClear

See documentation of executeSourceclearScan and [SourceClear](https://wiki.wdf.sap.corp/wiki/display/osssec/SourceClear)

## Deployment

### How to enable Cachebuster?

[CacheBuster](https://sapui5.hana.ondemand.com/sdk/#/topic/ff7aceda0bd24039beb9bca8e882825d) is meant to automatically invalidate Fiori apps cached on users' browsers (you don't really want to tell your customers to manually clear their browser cache every time you publish a new Fiori app version). This Apollo Requirement [SAPCPCFSE-205](https://jtrack/browse/SAPCPCFSE-205) is designed to take more responsibility off the products and provide cache busting as capability with the upcoming *HTML5 App Registry*. Currently planned for 1805.

*Workaround:* [SAP Newton](https://github.wdf.sap.corp/Newton/launchpad) implemented the cache busting capability (quite nicely) on their own in a proprietary way (invasive change into the FLP logic).

## Further Resources

TODO

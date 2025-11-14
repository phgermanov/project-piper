# downloadArtifactsFromNexus

## Description

This step downloads an artifact from nexus depending on the build descriptor file (package.json, pom.xml, mta.yaml).

The step can fetch artifacts from staging repositories or Milestone/Release repository. Depending on the `artifactType` (npm, java, mta, dockerbuild-releaseMetadata) and the `buildTool` (maven, mta) the url is constructed differently.

## Prerequisites

this step either requires

- a build descriptor to be stashed with the name `buildDescriptor`
- or alternatively a build descriptor to be available in the current workspace

## Nexus Migration to Artifactory

`downloadArtifactsFromNexus` is capable of downloading artifacts from internal artifactory (int.repositories.cloud.sap/) for a seamless nexus migration . There are 3 cases of how the migration from nexus is handled . Please make sure you consider the correct case as per your pipeline

- Case 1 : Custom Pipeline : when `downloadArtifactsFromNexus` runs after xMake build (step `sapXmakeExecuteBuild` with `buildType` : `xMakeStage` or `buildType` : `xMakePromote` :
  - No changes are needed in your pipeline since the required metadata will be automatically passed from `sapXmakeExecuteBuild` to `downloadArtifactsFromNexus`. When xMake switches the upload of build artifact to internal artifactory , `downloadArtifactsFromNexus` can then start downloading the artifact from internal artifactory without any change in the pipeline.

  - Please make sure the step `downloadArtifactsFromNexus` runs with config `fromStaging: false` when its running after xMake promote, so that the step does not download the artifact from staging repository after xMake promotes the artifact to nexus / artifactory.

- Case 2 : General Purpose Pipeline (GPP) :
  - GPP follows the above execution sequence hence no change require in the pipeline.

- Case 3 : Custom Pipeline: when `downloadArtifactsFromNexus` runs independent of xMake build and the step is used to download an artifact independent of xMake build :
  - you can make the step to point to artifactory (build quality : release) using the following configuration . if the build quality is milestone then keep the parameter `promoteRepository: 'build-milestones/'`

    ```groovy
    downloadArtifactsFromNexus  script: this,
                                nexusUrl: 'https://int.repositories.cloud.sap/',
                                promoteEndpoint : 'artifactory/',
                                promoteRepository: 'build-releases/'

    ```

## Pipeline configuration

none

## Example

Usage of pipeline step:

```groovy
downloadArtifactsFromNexus script: this, artifactType: 'mta', buildTool: 'mta', fromStaging: true
```

## Parameters

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
|script|yes|||
|artifactId|no|||
|artifactType|no|`java`| `java`, `mta`, `npm`, `dockerbuild-releaseMetadata`, `zip`|
|artifactVersion|no|||
|assemblyPath|no|`assembly`||
|buildDescriptorFile|no|artifactType=`dub`: `dub.json`<br />artifactType=`java`: `pom.xml`<br />artifactType=`maven-mta`: `${assemblyPath}/pom.xml`<br />artifactType=`mta`: `mta.yaml`<br />artifactType=`npm`: `package.json`<br />artifactType=`sbt`: `sbtDescriptor.json`<br />artifactType=`zip`: `pom.xml`<br />||
|buildTool|no|`maven`| only applicable for mta <br /> `maven`, `mta`|
|classifier|no|artifactType=`dockerbuild-releaseMetadata`: `releaseMetadata`<br />||
|extractPackage|no|artifactType=`npm`: `true`<br />artifactType=`python`: `true`<br />artifactType=`zip`: `true`<br />||
|fromStaging|no|`false`| `true`, `false`|
|group|no|artifactType=`dub`: `com.sap.dlang`<br />artifactType=`python`: `com.sap.pypi`<br />||
|helpEvaluateVersion|no|||
|nexusStageFilePath|no|||
|nexusUrl|no|`http://nexus.wdf.sap.corp:8081/`||
|packaging|no|artifactType=`dub`: `tar.gz`<br />artifactType=`dockerbuild-releaseMetadata`: `zip`<br />artifactType=`maven-mta`: `mtar`<br />artifactType=`mta`: `mtar`<br />artifactType=`npm`: `tgz`<br />artifactType=`python`: `tar.gz`<br />artifactType=`zip`: `zip`<br />||
|stageRepository|no|||
|versionExtension|no|||
|xMakeBuildQuality|no|`Milestone`|`Milestone`, `Release`|

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections the configuration is possible:

| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
|script||||
|artifactId||X|X|
|artifactType||X|X|
|artifactVersion||X|X|
|assemblyPath||X|X|
|buildDescriptorFile||X|X|
|buildTool||X|X|
|classifier||X|X|
|extractPackage||X|X|
|fromStaging||X|X|
|group||X|X|
|helpEveluateVersion||x|x|
|nexusStageFilePath||X|X|
|nexusUrl||X|X|
|packaging||X|X|
|stageRepository||X|X|
|versionExtension||X|X|
|xMakeBuildQuality|X|X|X|

### Details

In general the nexus url is constructed like this:

`nexusUrl/repository/group/artifactId/artifactVersion/artifactId-artifactVersion.packaging`

But there are some characteristics depending on the `artifactType` and `fromStaging`

#### repositories

For Milestone/Release artifacts the repository is the corresponding one of `deploy.milestones` or `deploy.releases`. Artifacts from type `npm` are loaded from `deploy.milestones.npm` or `deploy.releases.npm`.

For staging artifacts the repository is determined by the `staging_repo_id` from the xMakeProperties on the globalPipelineEnvironment. Or the url is completely provided by the `nexusStageFilePath` property.

#### npm

The `artifactId` and `artifactVersion` are determined by the `package.json` on root level.
The packaging for npm is `tgz`.
For staging artifact the `group` is always `com.sap.npm`, for Milestone/Release it is empty.
For Milestone/Release artifacts the `artifactVersion` has no ´commitId´ (version is truncated on the '+' sign). Also the `artifactVersion` folder in the url is replaced with `/-/`

The artifact is loaded to `artifactId.packaging`.

#### dub

The `group` is always `com.sap.dlang`.
The packaging for dub is `tar.gz`.

The artifact is loaded to `artifactId.packaging`.

#### java

The `group`, `artifactId`, `artifactVersion` and `packaging` are determined by the `pom.xml`.

By default the `pom.xml` on root level is taken. This can be overwritten through parameter `buildDescriptorFile`.

The artifact is loaded to `target/artifactId.packaging`.

#### mta

The `artifactId` and `artifactVersion` are determined by the `mta.yaml`. The `packaging` is always `mtar`. As the mta.yaml does not contain any groupId, this [must be specified](https://github.wdf.sap.corp/dtxmake/xmake-mta-plugin#configuration-xmakecfg) in the `.xmake.cfg`.

!!! warning
    This step expects the groupId to be defined in the `.xmake.cfg`.

The artifact is loaded to `target/artifactId.packaging`.

#### mta-maven

The `group`, `artifactId` and `artifactVersion` are determined by the `pom.xml` in the assembly module (default `assembly/` but can be specified by `assemblyPath` parameter). The `packaging` is always `mtar`.

The artifact is loaded to `target/artifactId.packaging`.

#### zip

The `group`, `artifactId`, `artifactVersion` and `packaging` are determined by the `pom.xml`. By default the `pom.xml` on root level is taken. Alternatively, parameter `buildDescriptorFile` offers the possibility to specify another `pom.xml`.

The artifact is loaded to `target/artifactId.packaging`. Additionally, the zip is extracted in the directory of `pom.xml.

### Further additional options available

#### classifier

A dash and the given classifier is added to the base name of the artifact to be downloaded from nexus, i.e. `artifactId-version-classifier.packaging`. The dash is inserted automatically if the classifier is non-empty.

The artifact is loaded to `target/artifactId.packaging`.

#### Direct download of dedicated archive

It is possible to directly load a specific archive. Passing following optional parameters to the step will overwrite the default retrieval method:

- `group`
- `artifactId`
- `artifactVersion`
- `packaging`

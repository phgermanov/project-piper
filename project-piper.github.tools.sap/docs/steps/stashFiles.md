# stashFiles

## Description

This step stashes files that are needed in other build steps (on other nodes).

## Prerequisites

none

## Pipeline configuration

none

## Explanation of pipeline step

Usage of pipeline step:

```groovy
stashFiles (script: this) {
  executeBuild (script: this, buildType: 'xMakeStage')
}
```

## Parameters

This step is separated into stashing before and after execution of the Closure (like `executeBuild ...`)

Available parameters:

| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
| script | no | empty `globalPipelineEnvironment` |  |
| stashIncludes | no | not set |  |
| stashExcludes | no | not set |  |

Details:

The step is stashing files before and after the build. This is due to the fact, that some of the code that needs to be stashed, is generated during the build (TypeScript for NPM).

| stash name | mandatory | prerequisite | pattern |
|---|---|---|---|
|buildDescriptor|yes| |`stashIncludes: ''**/pom.xml, **/.mvn/**, **/assembly.xml, **/.swagger-codegen-ignore, **/package.json, **/dub.json, **/requirements.txt, **/setup.py, **/whitesource_config.py, **/mta*.y*ml, **/.npmrc, **/whitesource.*.json, **/whitesource-fs-agent.config, .xmake.cfg, Dockerfile, **/VERSION, **/version.txt, **/build.sbt, **/sbtDescriptor.json, **/project/*'`|
|deployDescriptor|no| |`stashIncludes: '**/manifest*.y*ml, **/*.mtaext.y*ml, **/*.mtaext, **/xs-app.json, helm/**, *.y*ml'`|
|git|no| |`stashIncludes: '**/gitmetadata/**'`|
|opa5|yes|if OPA5 is enabled|`stashIncludes: '**/*.*'`|
|opensource configuration|no| |`stashIncludes: '**/srcclr.yml, **/vulas-custom.properties, **/.nsprc, **/'`|
|pipelineConfigAndTests|no| |`stashIncludes: '.pipeline/*.*'`|
|securityDescriptor|no| |`stashIncludes: '**/xs-security.json'`|
|tests|no| |`stashIncludes: '**/pom.xml, **/*.json, **/*.xml, **/src/**, **/node_modules/**, **/specs/**, **/env/**, **/*.js, **/*.cds'`|
|traceabilityMapping|no| | `'**/*.mapping'`|

!!! note "Overwriting default stashing behavior"
    It is possible to overwrite the default behavior all stashes using the parameters `stashIncludes` and `stashExcludes` , e.g.

    * `stashIncludes: [buildDescriptor: '**/mybuild.yml]`
    * `stashExcludes: [tests: '**/NOTRELEVANT.*]`    

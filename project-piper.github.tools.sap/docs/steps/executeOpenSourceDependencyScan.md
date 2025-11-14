# ${docGenStepName}

## ${docGenDescription}

## Prerequisites & Pipeline configuration

Please see the individual tool sections.

!!! note
    For MTA projects, you can exclude modules from the scanning.
    Therefore you can set the config property `openSourceDependencyScanExclude` with a comma separated list of `pom.xml`, `package.json`, `setup.py`, and `build.sbt` to exclude from the scan. This property is overridden by the `exclude` parameter.

## Examples

Usage of pipeline step:

For Maven projects:

```groovy
executeOpenSourceDependencyScan script: this, buildTool: 'maven'
```

For Node projects:

```groovy
executeOpenSourceDependencyScan script: this, buildTool: 'npm'
```

For MTA projects:

```groovy
executeOpenSourceDependencyScan script: this, buildTool: 'mta'
```

For Python projects:

```groovy
executeOpenSourceDependencyScan script: this, buildTool: 'pip'
```

For golang projects:

```groovy
executeOpenSourceDependencyScan script: this, buildTool: 'golang', filePath: 'someDockerImage.tar'
```

For D projects:

```groovy
executeOpenSourceDependencyScan script: this, buildTool: 'dub', filePath: 'someArchive.tar'
```

## ${docGenParameters}

## ${docGenConfiguration}

## Tool Details

### Whitesource

[Whitesource](https://www.whitesourcesoftware.com) is an SAP external tool that helps to identify known vulnerabilities in your D, node.js, Java and Scala projects. It is the successor of SourceClear and is already in use for license compliance scanning.

For more details please refer to the executeWhitesourceScan specific documentation.

### Node Security Project (nsp)

[nsp](https://github.com/nodesecurity/nsp) was an SAP external tool that helped to identify known vulnerabilities in your node.js projects. Since version 6 nsp moved into the node package manager (npm) and became part of its core functionality via the [`npm audit` command](https://docs.npmjs.com/getting-started/running-a-security-audit). Though the recommended scanning solution for node.js is Whitesource we keep this step for backward compatibility and to move it into [Piper Open Source](https://github.com/SAP/jenkins-library).

!!! warning "nsp step won't fail your build in case of findings"
    With moving nsp into the core of npm they unfortunately killed an important feature which is being able to distinguish between production and development dependencies. `npm audit` command is not yet capable to do so and therefore failing the build would also happen on any vulnerable dev dependency. Since the consequences for our consumers would be dramatic we decided - for the time being - to catch any exceptions thrown by the execution of the `npm audit` command in the sake of stability for our consumers. If you rely on this step - while you should be using Whitesource ideally - please regularly check your logs for any findings.

### VULAS Open Source Vulnerability Scan

VULAS is an analysis tool that aims to identify vulnerabilities in the open-source dependencies of Java applications and helps you in assessing and mitigating them.

For any details please refer to the executeVulasScan specific documentation.

### Protecode

Protecode is a product of [Synopsys](https://www.synopsys.com/) and therefore an SAP external tool that helps to identify known vulnerabilities in your D and golang projects or docker images. It is the only tool available at SAP which is capable of analysing binaries for publicly known vulnerabilities and can therefore also be used to analyse artifacts created by using the C language family.

For more details please refer to the executeProtecodeScan specific documentation.

import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Deprecate
import com.sap.piper.internal.MapUtils
import com.sap.piper.internal.Notify
import groovy.transform.Field
import groovy.text.GStringTemplateEngine

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'downloadArtifactsFromNexus'
@Field Set GENERAL_CONFIG_KEYS = [
    'artifactType',
    'buildTool',
    'xMakeBuildQuality'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'artifactId',
    'artifactVersion',
    'assemblyPath',
    'buildDescriptorFile',
    'classifier',
    'extractPackage',
     /**
     * Defines if the artifact will be downloaded from a nexus staging repository.
     * @possibleValues true, false
     */
    'fromStaging',
    'group',
    'nexusStageFilePath', // legacy parameter
    /**
     * Defines the host and port of the nexus repository, e.g. 'http://nexus.wdf.sap.corp:8081'.
     */
    'nexusUrl',
    'packaging',
    'promoteEndpoint',
    'promoteRepository',
    /**
     * Defines the path to the nexus staging endpoint, e.g. 'nexus/content/repositories'.
     */
    'stageEndpoint',
    /**
     * Defines the path to the nexus staging repository, e.g. 'deploy.snapshots'.
     * @mandatory for `fromStaging`
     */
    'stageRepository',
    'versionExtension',
    'disableLegacyNaming',
    'helpEvaluateVersion'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

@Field stagingCredentialsMap = [:]

void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()

        // handle legacy MTAs
        def mtaLegacy = false
        if(parameters.artifactType == 'mta' && parameters.buildTool != 'mta') {
            parameters.artifactType = 'maven-mta'
            mtaLegacy = true
        }

        // handle deprecated parameters
        Deprecate.parameter(this, parameters, 'extractNpmPackage', 'extractPackage')

        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(
                stageRepository: script.globalPipelineEnvironment.getXMakeProperty('staging_repo_id') ?: script.commonPipelineEnvironment?.getValue('xmakeStagingRepositoryId'),
                artifactVersion: script.globalPipelineEnvironment.getArtifactVersion()
            )
            .mixin(parameters, PARAMETER_KEYS)
            .use()


        //default handling based on buildTool in case artifactType is not set
        if (config.buildTool && config.artifactType == 'maven') {
            switch (config.buildTool) {
                case ['dub', 'mta', 'npm', 'sbt']:
                    config.artifactType = config.buildTool
                    break
                case 'pip':
                    config.artifactType = 'python'
                    break
            }
        }
        if (config.artifactType == 'npm' && config.classifier == 'bundle' && !config.packaging) {
            config.packaging = 'tar.gz'
        }

        config = new ConfigurationHelper(config)
            .dependingOn('artifactType').mixin('buildDescriptorFile')
            .dependingOn('artifactType').mixin('extractPackage')
            .dependingOn('artifactType').mixin('group')
            .dependingOn('artifactType').mixin('packaging')
            .dependingOn('artifactType').mixin('classifier')
            .use()

        if(config.fromStaging) {
            if(!script.commonPipelineEnvironment.getValue('stageBOM')) {
                // Stage URL (stage_repourl) could be one of this:
                // - http://nexus.wdf.sap.corp:8081/nexus/content/repositories/xmakedeploymilestonesprofile-246724
                // - http://nexus.wdf.sap.corp:8081/stage/repository/8efef2648092-milestones-20180506-233957735-422
                if(script.globalPipelineEnvironment.getXMakeProperty('stage_repourl')?.contains(config.promoteEndpoint)){
                    config.stageEndpoint = config.promoteEndpoint
                }
            }
        }
        def nexusRepositoryUrl = config.fromStaging && script.commonPipelineEnvironment.getValue('stageBOM') && !config.nexusStageFilePath?null:getRepositoryUrl(config)

        echo "Nexus repository: ${nexusRepositoryUrl}"
        utils.unstash 'buildDescriptor'

        try {
            if ((config.group != null) && config.artifactId && config.artifactVersion && config.packaging) {
                // Note: This does NOT take the version into account (for backwards compatibility) when naming the target file.
                //       This if-branch, however, should hardly be in use in the future, as we now have automatic versioning.
                handleArtifactDownloadForUserDefinedGAV(config, nexusRepositoryUrl, script)
            } else {
                switch(config.artifactType) {
                    case 'python':
                        handlePythonArtifactDownload(utils, config, nexusRepositoryUrl, script)
                        break
                    case 'dub':
                        handleDubArtifactDownload(utils, config, nexusRepositoryUrl, script)
                        break
                    case 'zip':
                        handleZipArtifactDownload(utils, config, nexusRepositoryUrl, script)
                        break
                    case 'java':
                    case 'maven':
                        handleMavenArtifactDownload(utils, config, nexusRepositoryUrl, script)
                        break
                    case 'maven-mta':
                        handleMavenMtaArtifactDownload(utils, config, nexusRepositoryUrl, script)
                        break
                    case 'mta':
                        handleMtaArtifactDownload(utils, config, nexusRepositoryUrl, script)
                        break
                    case 'sbt':
                        handleSbtArtifactDownload(utils, config, nexusRepositoryUrl, script)
                        break
                    case 'dockerbuild-releaseMetadata':
                        handleDockerBuildReleaseMetadataArtifactDownload(utils, config, nexusRepositoryUrl, script)
                        break
                    case 'npm':
                        handleNpmArtifactDownload(utils, config, nexusRepositoryUrl, script)
                        break
                    case 'golang':
                        handleGolangArtifactDownload(utils, config, nexusRepositoryUrl, script)
                        break
                }
            }
        } catch (err) {
            Notify.error(this, "Failed to download artifact:  ${err}, please see log for details.", STEP_NAME)
        }
    }
}

private void handleArtifactDownloadForUserDefinedGAV(config, nexusRepositoryUrl, script) {
    def targetFile = "${createTargetFolder()}${config.artifactId}${config.classifier?'-'+config.classifier:''}.${config.packaging}"
    def artifactUrl = getArtifactUrl(nexusRepositoryUrl, config.group, config.artifactId, config.artifactVersion, config.packaging, config.classifier, config, script.commonPipelineEnvironment)
    // correct NPM url
    if(config.artifactType == 'npm') {
        if(config.fromStaging) {
            if(nexusRepositoryUrl || !script.commonPipelineEnvironment.getValue('stageBOM')) artifactUrl = artifactUrl.replace('tgz', 'tar.gz')
        }
        else if (config.classifier != 'bundle') {
            artifactUrl = artifactUrl.replaceAll("/${config.artifactVersion}/", '/-/')
        }
    }
    // download artifact
    loadArtifact(artifactUrl, targetFile, script.globalPipelineEnvironment, config)
}

private void handlePythonArtifactDownload(utils, config, nexusRepositoryUrl, script) {
    def gav = utils.getPipGAV(config.buildDescriptorFile)
    config.artifactId = gav.artifact
    config.artifactVersion = gav.version
    def targetFile = "${createTargetFolder()}${config.artifactId}.${config.packaging}"
    // create artifact URL
    def artifactUrl = getArtifactUrl(nexusRepositoryUrl, config.group, config.artifactId, config.artifactVersion, config.packaging, config.classifier, config, script.commonPipelineEnvironment)
    // download artifact
    loadArtifact(artifactUrl, targetFile, script.globalPipelineEnvironment, config)

    if (config.extractPackage) {
        extractPythonPackage(config)
    }
}

private void handleDubArtifactDownload(utils, config, nexusRepositoryUrl, script) {
    def gav = utils.getDubGAV(config.buildDescriptorFile)
    config.group = gav.group
    config.artifactId = gav.artifact
    config.artifactVersion = gav.version
    config.packaging = gav.packaging
    def targetFile = "${createTargetFolder()}${config.artifactId}.${config.packaging}"
    // create artifact URL
    def artifactUrl = getArtifactUrl(nexusRepositoryUrl, config.group, config.artifactId, config.artifactVersion, config.packaging, config.classifier, config, script.commonPipelineEnvironment)
    // download artifact
    loadArtifact(artifactUrl, targetFile, script.globalPipelineEnvironment, config)
}

private void handleZipArtifactDownload(utils, config, nexusRepositoryUrl, script) {
    def gav = utils.readMavenGAV(config.buildDescriptorFile, '')
    config.group = gav.group
    config.artifactId = gav.artifact
    config.artifactVersion = gav.version

    def targetFile = "${createTargetFolder()}${config.artifactId}.${config.packaging}"
    // create artifact URL
    def artifactUrl = getArtifactUrl(nexusRepositoryUrl, config.group, config.artifactId, config.artifactVersion, config.packaging, config.classifier, config, script.commonPipelineEnvironment)
    // download artifact
    loadArtifact(artifactUrl, targetFile, script.globalPipelineEnvironment, config)

    if (config.extractPackage) {
        extractZipPackage(config)
    }
}

private void handleMavenArtifactDownload(utils, config, nexusRepositoryUrl, script) {
    echo "helpEvaluateVersion::${config.helpEvaluateVersion}"

    def pom = readMavenPom(file: config.buildDescriptorFile)
    def gav = utils.readMavenGAV(config.buildDescriptorFile, '')
    def mavenBuildConfig = utils.readMavenBuildConfigurations(config.buildDescriptorFile, config.helpEvaluateVersion)

    config.group = gav.group
    config.artifactId = gav.artifact
    config.artifactVersion = gav.version
    config.packaging = gav.packaging

    config.packaging = mapPackaging(config.packaging)

    downloadAndRenameMavenArtifact(mavenBuildConfig, config, nexusRepositoryUrl, script)

    // Legacy / compatibility handling of artifact download.
    // !!! This should be removed after a given grace period!!!
    Boolean disableLegacyNaming = config.disableLegacyNaming ?: false
    if(!disableLegacyNaming) {
        LEGACY_downloadAndRenameMavenArtifact(mavenBuildConfig, config, nexusRepositoryUrl, script)
    }

    // handle multi-module POMs
    def modules = pom.modules
    if (modules) {
        modules.each { module ->
            echo "Downloading artifact for module: ${module}"
            Map moduleConfig = MapUtils.deepcopy(config)
            moduleConfig.buildDescriptorFile = "${mavenBuildConfig.basedir}/${module}/pom.xml"
            handleMavenArtifactDownload(utils, moduleConfig, nexusRepositoryUrl, script)
        }
    }
}

private String mapPackaging(String packaging) {
    echo "Original packaging type: ${packaging}"
    String mappedPackaging = packaging
    switch(packaging) {
        case "eclipse-target-definition":
            mappedPackaging = "target"
            break
        case ["maven-plugin", "eclipse-plugin", "eclipse-feature", "eclipse-test-plugin"]:
            mappedPackaging = "jar"
            break
        case ["eclipse-repository", "eclipse-application", "eclipse-p2-repository"]:
            mappedPackaging = "zip"
            break
        default:
            mappedPackaging = packaging
    }
    echo "Mapped packaging type: ${mappedPackaging}"
    return mappedPackaging
}

private void downloadAndRenameMavenArtifact(mavenBuildConfig, config, nexusRepositoryUrl, script) {
    // Resolve the final artifact name (which may contain variables) from the
    // information in the pom and based on the automatically generated version.
    // The original version will be retrieved from the pipeline environment in
    // case the artifactPrepareVersion step has set it beforehand.
    def artifactName = resolveFinalArtifactName(mavenBuildConfig, config, script.globalPipelineEnvironment)

    echo "Final artifact name that downloaded artifact will be renamed to: ${artifactName}"

    def targetFolder = createTargetFolder(mavenBuildConfig.targetFolder)
    def targetFile = "${targetFolder}${artifactName}"

    // create artifact URL
    def artifactUrl = getArtifactUrl(nexusRepositoryUrl, config.group, config.artifactId, config.artifactVersion, config.packaging, config.classifier, config, script.commonPipelineEnvironment)

    echo "Artifact URL: ${artifactUrl}"
    echo "Target File Name: ${targetFile}"

    // download artifact
    loadArtifact(artifactUrl, targetFile, script.globalPipelineEnvironment, config)
}

private void LEGACY_downloadAndRenameMavenArtifact(mavenBuildConfig, config, nexusRepositoryUrl, script) {
    // print a major warning for stakeholders to migrate to new version and away from this
    // legacy behavior.

    String warningMessage = "This build uses a legacy build artifact renaming behavior."
    String warningDetails = """Details:
    Previous versions of Piper used to download and rename Maven build artifacts from Nexus following a fixed / pre-defined naming convention.
    The convention was, that irrespective of the version you defined in your pom.xml, the build artifact's final name would always be <artifactId>.<packaging>, e.g. my-service.jar.
    Usually, developers were asked / required to define a <finalName> in their pom.xml files that specified this naming pattern.

    This convention is NO LONGER ENFORCED. You can now omit the <finalName> and the artifacts will be named the way Maven generates them.
    Piper will apply an automatic versioning, but make sure that after downloading artifacts from Nexus, they will be named again as Maven generated them.
    In other words, if your project's Maven GAV is com.sap.cloud:my-service:0.0.1 with packaging type jar, and you have not specified an explicit <finalName>,
    the artifact will be named my-service-0.0.1.jar before deployment to the Cloud, as expected. Piper will make sure of it.
    That artifact name then needs to be referenced in your manifest.yml or mta.yml, just as you would in a local development setup.

    Of course, you can still define a <finalName>, if you'd like to, and that name will be applied by Piper as well.

    If your project does NOT use a <finalName> in its pom.xml today, but implicitly relies on Piper's old artifact naming convention,
    the new behavior may break your builds in the future. This warning is to inform you, that this is a grace period and you MUST act!

    What do you have to do?
    - If you are using <finalName> in your project's pom.xml already, you don't need to do anything. Just know that this is not strictly required anymore.
    - If you are NOT using <finalName> in your project's pom.xml yet, you have two options:
        1) define a <finalName> in your pom.xml's <build> section, like this <finalName>\${project.artifactId}</finalName> ... or ...
        2) make sure that your manifest.yml or mta.yml references the build artifact as Maven created / named it.

    During this grace period, Piper will download build artifacts twice, renaming one using the new behavior, and one
    using the old behavior to stay backwards compatible. After the grace period only the new behavior will be used!

    In case you want to turn off the legacy behavior (e.g. because you are using a <finalName> configuration) and already
    actively opt in to the new behavior, you can do so via the .pipeline/config.yml setting "disableLegacyNaming: true"
    for step "downloadArtifactsFromNexus". This may speed up your build during the grace period.
    After the grace period has ended, this will be the default, and you can remove the additional configuration.
    """

    Notify.warning(this, "${warningMessage} - ${warningDetails}")
    setBuildUnstable(config, "Setting build to UNSTABLE for now.")

    // Resolve the final artifact name (which may contain variables) from the
    // information in the pom and based on the automatically generated version.
    // The original version will be retrieved from the pipeline environment in
    // case the artifactPrepareVersion step has set it beforehand.
    def artifactName = "${config.artifactId}.${config.packaging}"

    echo "Final artifact name that downloaded artifact will be renamed to: ${artifactName}"

    def targetFolder = createTargetFolder(mavenBuildConfig.targetFolder)
    def targetFile = "${targetFolder}${artifactName}"

    // create artifact URL
    def artifactUrl = getArtifactUrl(nexusRepositoryUrl, config.group, config.artifactId, config.artifactVersion, config.packaging, config.classifier, config, script.commonPipelineEnvironment)

    echo "Artifact URL: ${artifactUrl}"
    echo "Target File Name: ${targetFile}"

    // download artifact
    loadArtifact(artifactUrl, targetFile, script.globalPipelineEnvironment, config)
}

private void handleMavenMtaArtifactDownload(utils, config, nexusRepositoryUrl, script) {
    config.buildDescriptorFile = resolve(config)

    def gav = utils.readMavenGAV(config.buildDescriptorFile, '')

    config.group = gav.group
    config.artifactId = gav.artifact
    config.artifactVersion = gav.version

    def targetFile = "${createTargetFolder()}${config.artifactId}.${config.packaging}"
    // create artifact URL
    def artifactUrl = getArtifactUrl(nexusRepositoryUrl, config.group, config.artifactId, config.artifactVersion, config.packaging, config.classifier, config, script.commonPipelineEnvironment)
    // download artifact
    loadArtifact(artifactUrl, targetFile, script.globalPipelineEnvironment, config)
}

private void handleMtaArtifactDownload(utils, config, nexusRepositoryUrl, script) {
    def gav = utils.getMtaGAV(config.buildDescriptorFile)

    config.group = gav.group
    config.artifactId = gav.artifact
    config.artifactVersion = gav.version

    def targetFile = "${createTargetFolder()}${config.artifactId}.${config.packaging}"
    // create artifact URL
    def artifactUrl = getArtifactUrl(nexusRepositoryUrl, config.group, config.artifactId, config.artifactVersion, config.packaging, config.classifier, config, script.commonPipelineEnvironment)
    // download artifact
    loadArtifact(artifactUrl, targetFile, script.globalPipelineEnvironment, config)
}

private void handleSbtArtifactDownload(utils, config, nexusRepositoryUrl, script) {
    def gav = utils.getSbtGAV(config.buildDescriptorFile)

    config.group = gav.group
    config.artifactId = gav.artifact
    config.artifactVersion = gav.version
    config.packaging = gav.packaging

    def targetFile = "${config.artifactId}.${config.packaging}"
    // create artifact URL
    def artifactUrl = getArtifactUrl(nexusRepositoryUrl, config.group, config.artifactId, config.artifactVersion, config.packaging, config.classifier, config, script.commonPipelineEnvironment)
    // download artifact
    loadArtifact(artifactUrl, targetFile, script.globalPipelineEnvironment, config)
}

private void handleDockerBuildReleaseMetadataArtifactDownload(utils, config, nexusRepositoryUrl, script) {
    config.group = utils.getXmakeCfgField("buildplugin", "gid")
    config.artifactId = utils.getXmakeCfgField("buildplugin", "aid")
    config.artifactVersion = utils.getXmakeVersion() + (config.versionExtension ? "-${config.versionExtension}" : "")

    def targetFile = "${createTargetFolder()}${config.artifactId}.${config.packaging}"
    // create artifact URL
    def artifactUrl = getArtifactUrl(nexusRepositoryUrl, config.group, config.artifactId, config.artifactVersion, config.packaging, config.classifier, config, script.commonPipelineEnvironment)
    // download artifact
    loadArtifact(artifactUrl, targetFile, script.globalPipelineEnvironment, config)
}

private void handleNpmArtifactDownload(utils, config, nexusRepositoryUrl, script) {
    def gav = utils.getNpmGAV(config.buildDescriptorFile)

    if (config.fromStaging || config.classifier == 'bundle') {
        // in case the scope is set and not @sap the path contains the scope name without the leading '@'
        gav.group = 'com/sap/npm/' + (gav.group && gav.group != '@sap'?gav.group.substring(1):'')
    }

    if (!config.fromStaging && config.classifier != 'bundle') {
        // cut-off commitId in milestone/release quality
        if (gav.version.indexOf('+') != -1)
            gav.version = gav.version.substring(0, gav.version.indexOf('+'))
    }

    config.group = gav.group
    config.artifactId = gav.artifact
    config.artifactVersion = gav.version

    def targetFile = "${config.artifactId}.${config.packaging}"

    // create artifact URL
    def artifactUrl = getArtifactUrl(nexusRepositoryUrl, config.group, config.artifactId, config.artifactVersion, config.packaging, config.classifier, config, script.commonPipelineEnvironment)

    // correct NPM artifact URL
    if(config.fromStaging){
        if(nexusRepositoryUrl || !script.commonPipelineEnvironment.getValue('stageBOM')) artifactUrl = artifactUrl.replace('tgz', 'tar.gz')
    }
    else if (config.classifier != 'bundle'){
        artifactUrl = artifactUrl.replaceAll("/${config.artifactVersion}/", '/-/')
    }

    // download artifact
    loadArtifact(artifactUrl, targetFile, script.globalPipelineEnvironment, config)

    if (config.extractPackage) {
        extractNpmPackage(config)
    }
}

private void handleGolangArtifactDownload(utils, config, nexusRepositoryUrl, script) {
    def golangPackagePath =  utils.getXmakeCfgField("buildplugin", "golang-package-path")
    // last name in path is the artifact id
    def pathFragments = golangPackagePath.split("/")
    config.artifactId = pathFragments[pathFragments.length -1]
    // everything before the artifact id belongs to the group
    int index = golangPackagePath.lastIndexOf("/")

    // note: group needs to be lower case due to https://github.wdf.sap.corp/dtxmake/xmake-go-plugin/blob/1f80ad8744c3860de2ad1b7063cc730e1e397f18/src/externalplugin/buildplugin.py#L196
    config.group = "com.sap.golang." + golangPackagePath.substring(0, index).toLowerCase()
    def targetFile = "${config.artifactId}.${config.packaging}"
    // create artifact URL - it's expected that the artifactVersion was set before in the pipeline environment
    def artifactUrl = getArtifactUrl(nexusRepositoryUrl, config.group, config.artifactId, config.artifactVersion, config.packaging, config.classifier, config, script.commonPipelineEnvironment)
    // download artifact
    loadArtifact(artifactUrl, targetFile, script.globalPipelineEnvironment, config)
    if (config.extractPackage) {
        extractGolangPackage(config)
    }
}

private String resolveFinalArtifactName(mavenBuildConfig, config, environment) {

    // Final name from POM may have an auto-generated and complex version,
    // e.g. 0.0.1-<timestamp>_<commitId>.
    String finalName = mavenBuildConfig.finalName
    String automaticallyGeneratedVersion = config.artifactVersion

    // The final name used for uploading to CF needs to be the one that
    // the developer originally specified in Maven (and subsequently in
    // deployment files, e.g. manifest.yml)
    // It may or may not have contained a version, but if it did, it may have
    // been replaced by the automatically generated one of the artifactPrepareVersion step.
    // If that is the case, we simply switch it back.
    String versionBeforeAutoVersioning = environment.versionBeforeAutoVersioning
    if(versionBeforeAutoVersioning) {
        echo "Original artifact version was: ${versionBeforeAutoVersioning}"
        echo "Final name from pom (after automatic versioning): ${finalName}"
        echo "Replacing automatic version with original one."
        finalName = finalName.replace(automaticallyGeneratedVersion, versionBeforeAutoVersioning)
    }

    echo "Final name: ${finalName}"

    String packagingSuffix = ".${config.packaging}"

    finalName += packagingSuffix

    echo "Final artifact name: ${finalName}"

    return finalName
}

def compareWithPropertyValue(object, property, value) {
    return (object.containsKey(property) && object."$property"?object."$property".toString().toUpperCase():null)==(value?value.toString().toUpperCase():null)
}

private String getArtifactUrlFromStageBom(group, artifactId, artifactVersion, packaging, classifier, stageBom, artifactType) {
    for (repository in stageBom) {
        if(!repository.value.format?.toUpperCase().contains(artifactType.toUpperCase())) continue

        isMavenRepository = repository.value.format?.toUpperCase().equals("MAVEN2")
        for (component in repository.value.components) {
            if( component.containsKey("assets")
                && compareWithPropertyValue(component, "artifact", artifactId)
                && compareWithPropertyValue(component, "version", artifactVersion)
                && (!isMavenRepository || compareWithPropertyValue(component, "group", isMavenRepository && group?group.replaceAll('/', '.').replaceAll('\\.$',''):group)) // group & lower classifier & extension fields are maven2 specific & must be checked only if looking in maven repository
            ) {
                for (asset in component.assets) {
                    if(isMavenRepository && (!compareWithPropertyValue(asset, "classifier", classifier) || !compareWithPropertyValue(asset, "extension", packaging))) {
                        continue
                    }
                    // nexus migration : in future staging repo will be password protected and hence creds must be retained
                    stagingCredentialsMap["username"] = repository.value.credentials.user
                    stagingCredentialsMap["password"] = repository.value.credentials.password
                    return asset.url
                }
            }
        }
    }
}

private String getPromotedArtifactUrl(group, artifactId, artifactVersion, packaging, classifier = "", pipelineEnv, artifactType = ""){
    if (pipelineEnv.getValue('promotedArtifactUrls')){
        echo "Promoted Artifacts: ${pipelineEnv.getValue('promotedArtifactUrls')}"
    }
    if (group && !group.endsWith('/')) group += '/'
    group = group.replaceAll('\\.', '/')
    def matcher = "${group}${artifactId}/${artifactVersion}/${artifactId}-${artifactVersion}${classifier?'-'+classifier:''}.${packaging}"
    if(artifactType == 'npm' && classifier != 'bundle') {
       matcher = matcher.replaceAll("/${artifactVersion}/", '/-/')
    }
    echo "looking for artifact ${matcher}"
    for (artifactUrl in pipelineEnv.getValue('promotedArtifactUrls')) {
        if(artifactUrl.endsWith(matcher)){
            if(artifactUrl.contains("common.repositories.cloud.sap")){
                echo "skipping protected artifact from common repository"
                continue
            }
            echo "found ${artifactUrl}"
            // missing "repository/" in path to NPM type repository for NPM artifacts created with the xmake service
            //TODO: remove once NEXUS is sunsetted
            if(artifactType == 'npm' && artifactUrl.contains("nexus.wdf.sap.corp")
                && (artifactUrl.contains("/nexus/deploy.milestones.npm/") || artifactUrl.contains("/nexus/deploy.releases.npm/"))){
                artifactUrl = artifactUrl.replaceAll("/nexus/", "/nexus/repository/")
                echo "correct URL: ${artifactUrl}"
            }
            return artifactUrl
        }
    }
    return ""
}

private String getArtifactUrl(nexusUrl, group, artifactId, artifactVersion, packaging, classifier = "", config = null, pipelineEnv = null, artifactType = null) {
    // if no nexus url specified, then the artifact url must be resolved from the stage bom in staging mode
    if(!nexusUrl && config.fromStaging && pipelineEnv.getValue('stageBOM')) {
        // On first loop, search in priority for content in the appropriate repository build techno specific type except when classifier is specified which means, we want absolutly maven
        // This is quite also to ensure that callers asking for bundle as classifier will not get their requests ignored and returing something else
        def url = artifactType == null && !classifier ? getArtifactUrl(nexusUrl, group, artifactId, artifactVersion, packaging, classifier, config, pipelineEnv, config.artifactType):null // Search by default in a specific repository technology
        artifactType = artifactType == null ?"MAVEN2":artifactType

        if(!url) url = getArtifactUrlFromStageBom(group, artifactId, artifactType?.toUpperCase().equals("NPM")?artifactVersion.replaceAll('\\+\\w+$', ''):artifactVersion, packaging, classifier, pipelineEnv.getValue('stageBOM'), artifactType)
        return url?url:""
    } else if(config && config.fromStaging == false && pipelineEnv.getValue('promotedArtifactUrls')) {
        return getPromotedArtifactUrl(group, artifactId, artifactVersion, packaging, classifier, pipelineEnv, config.artifactType)
    } else {
        if (nexusUrl && !nexusUrl.endsWith('/')) nexusUrl += '/'
        if (group && !group.endsWith('/')) group += '/'
        group = group.replaceAll('\\.', '/')
        return "${nexusUrl}${group}${artifactId}/${artifactVersion}/${artifactId}-${artifactVersion}${classifier?'-'+classifier:''}.${packaging}"
    }
}

private void loadArtifact(url, target, gpe = [:], config) {
    def username = stagingCredentialsMap["username"]
    def password = stagingCredentialsMap["password"]

    wrap([$class: 'MaskPasswordsBuildWrapper', varPasswordPairs: [[password: username], [password: password]]]) {
         def auth = ""

        // nexus migration : if from staging add the authoriation header : stagingCredentialsMap is only filled when staging bom from build-result.json is filled up
        if (config.fromStaging && username && password) {
            auth =  "--basic --user " +  "${username}:${password}"
            // script = "curl -H Authorization: Basic " + auth.bytes.encodeBase64().toString() + " --silent --show-error --write-out '%{response_code}' --location --output '${target}' ${url}"
        }

        def statusCode = sh(
            returnStdout: true,
            script: "curl --insecure " + auth + " --silent --show-error --write-out '%{response_code}' --location --output '${target}' ${url}"
        ).trim()

        if (statusCode != '200') {
            error "Download artifacts from nexus failed: ${statusCode}"
        } else {
            gpe.nexusLastDownloadUrl = url
        }
    }
}

private void extractZipPackage(config, path = 'target/'){
    def file = "${config.artifactId}.${config.packaging}"
    unzip zipFile: "${path}${file}", dir: config.buildDescriptorFile.replace('pom.xml', ''), quiet: true
    sh "rm ${path}${file}"
}

private void extractNpmPackage(config){
    def file = "${config.artifactId}.${config.packaging}"
    def returnCode = sh(returnStatus: true, script: "tar -tf ${file} | grep -E \"^(\\.\\/)?package\\/package\\.json\"")
    if (returnCode == 0)
        sh "tar -xvf ${file} --strip-components=1"
    else
        sh "tar -xvf ${file} > /dev/null 2>&1"
    sh "rm ${file}"
}

private void extractPythonPackage(config, path = 'target/'){
    def file = "${config.artifactId}.${config.packaging}"
    sh "tar -xf ${path}${file} --strip-components=1"
    sh "rm ${path}${file}"
}

private void extractGolangPackage(config){
    def file = "${config.artifactId}.${config.packaging}"
    sh "tar -xf ${file} --strip-components=1"
    sh "rm ${file}"
}

private String getRepositoryUrl(config){
    def url  = config.nexusUrl
    if (config.fromStaging) {
        url  += config.stageEndpoint
        url  += config.stageRepository
        // handle legacy coding
        if(config.nexusStageFilePath)
            url = config.nexusStageFilePath
    } else {
        url  += config.promoteEndpoint
        url  += GStringTemplateEngine.newInstance().createTemplate(config.promoteRepository)
            .make([
                quality: config.xMakeBuildQuality.toLowerCase(),
                npmSuffix: config.artifactType == 'npm' && config.classifier != 'bundle'?".${config.artifactType}":''
            ]).toString()
    }
    return url
}

private String createTargetFolder(path = 'target/'){
    if(path){
        if(!path.endsWith('/')) path += '/'
        sh "mkdir --parents '${path}'"
    }
    return path
}

private String resolve(config){
    return GStringTemplateEngine.newInstance().createTemplate(config.buildDescriptorFile).make([assemblyPath: config.assemblyPath]).toString()
}

void setBuildUnstable(Map config, String message = null) {
    def failureMessage = message ? "[${config.stepName}] ${message}" : "[${config.stepName}] Error in step ${config.stepName} - Build result set to 'UNSTABLE'"
    try {
        //use new unstable feature if available: see https://jenkins.io/blog/2019/07/05/jenkins-pipeline-stage-result-visualization-improvements/
        unstable(failureMessage)
    } catch (java.lang.NoSuchMethodError nmEx) {
        if (config.stepParameters?.script) {
            config.stepParameters?.script.currentBuild.result = 'UNSTABLE'
        } else {
            currentBuild.result = 'UNSTABLE'
        }
        echo failureMessage
    }
}

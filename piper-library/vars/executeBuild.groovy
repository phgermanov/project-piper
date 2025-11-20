import hudson.AbortException

import com.sap.icd.jenkins.Utils
import com.sap.piper.GenerateDocumentation
import com.sap.piper.internal.Deprecate
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Notify

import groovy.text.GStringTemplateEngine
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'executeBuild'
@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * For xMake-free builds only: Defines the tool which is used for building the artifact.
     */
    'buildTool',
    /**
     * **For native builds**
     * Defines the build quality for the native build. For productive pipelines this should always be set to `Release`.
     * @possibleValues 'Milestone', 'Release'
     */
    'buildQuality',
    /**
     * **For xMake stage/promote builds**
     * Defines the build quality for the xMake build. For productive pipelines this should always be set to `Release`.
     * In order to setup a build in `Release` quality, please follow the respective [Documentation of the Non-ABAP Assembly team](https://wiki.wdf.sap.corp/wiki/display/NAAS/Simplified+Release+process+for+Github+based+Cloud+projects).
     * @possibleValues 'Milestone', 'Release'
     */
    'xMakeBuildQuality',
    /**
     * **For xMake stage/promote builds**
     * Defines the id of the Jenkins username/password credentials on your Jenkins instance for the **xMake-Dev landscape**.
     * This parameter does not need to be set explicitly if you maintain your credentials with the credential id `xmakeDev`.
     * The credentials contain a pair of user-id and [Jenkins API token](../build/xMake.md#jenkins-api-token).
     * For job execution on xMake these credentials are essential to successfully trigger the respective jobs.
     * Please see [xMake documentation](https://github.wdf.sap.corp/pages/xmake-ci/User-Guide/Setting_up_a_Build/Setting_up_a_Build/#step-2-create-and-configure-your-build) for details about your xMake build configuration.
     */
    'xMakeDevCredentialsId',
    /**
     * **For xMake stage/promote builds**
     * Defines the id of the Jenkins username/password credentials on your Jenkins instance for the **xMake-Nova landscape**.
     * This parameter does not need to be set explicitly if you maintain your credentials with the credential id `xmakeNova`.
     * The credentials contain a pair of user-id and [Jenkins API token](../build/xMake.md#jenkins-api-token).
     * For job execution on xMake these credentials are essential to successfully trigger the respective jobs.
     * Please see [xMake documentation](https://github.wdf.sap.corp/pages/xmake-ci/User-Guide/Setting_up_a_Build/Setting_up_a_Build/#step-2-create-and-configure-your-build) for details about your xMake build configuration.
     */
    'xMakeNovaCredentialsId',
    /**
     * **For xMake stage/promote builds**
     * Defines if projectArchive deployPackage.tar.gz files have to be unzipped in a subdirectory or not.
     * Really practical when you are working with multiple variants.
     * Please see [xMake documentation](https://github.wdf.sap.corp/pages/xmake-ci/User-Guide/Setting_up_a_Build/Setting_up_a_Build/#step-2-create-and-configure-your-build) for details about your xMake build configuration.
     */
    'xMakeDownloadDownstreamsProjectArchives',
    /**
     * Activate a "Piper native build" (e.g. maven, npm) in combination with SAP's staging service.
     * @possibleValues 'true', 'false'
     */
    'nativeBuild',
    /**
     * Activate a native build pull-request voting (e.g. maven, npm) only in combination with "Piper native build".
     * @possibleValues 'true', 'false'
     */
    'nativeVoting',
    /**
     * **Only for OLD functionality of NaaS release forks.**
     * Please consider switching to the [Simplified Release Process](https://wiki.wdf.sap.corp/wiki/display/NAAS/Simplified+Release+process+for+Github+based+Cloud+projects).
     * Defines the credentials id of your Jenkins credentials of type `SSH Username with private key`.
     */
    'gitSshKeyCredentialsId',
    /**
     * **Only for OLD functionality of NaaS release forks.**
     * Defines a specific git user name for the commit to the release fork.
     */
    'gitUserName',
    /**
     * **Only for OLD functionality of NaaS release forks.**
     * Defines a specific git user email for the commit to the release fork.
     */
    'gitUserEMail'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /** Performs upload of files to Cumulus. */
    'sapCumulusUpload',
    /**
     * Defines the type of build to be executed.
     * @possibleValues 'xMakeStage', 'xMakePromote', 'dockerInside', 'dockerLocal', 'kaniko'
     */
    'buildType',
    /** Defines if a container image should be created using Cloud Native Buildpacks using the artifact created defined by `buildTool`.
     * @possibleValues true, false
     */
    'cnbBuild',
    /**
     * **For `buildType: kaniko`**
     * Defines the build options for the [kaniko](https://github.com/GoogleContainerTools/kaniko) build.
     */
    'containerBuildOptions',
    /**
     * **For `buildType: kaniko`**
     * Kubernetes only:
     * Allows to specify start command for container created with dockerImage parameter to overwrite Piper default (`/usr/bin/tail -f /dev/null`).
     */
    'containerCommand',
    /**
     * **For `buildType: kaniko`**
     * Kubernetes only:
     * Allows to specify the shell to be used for execution of commands.
     */
    'containerShell',
    /**
     * **For `buildType: kaniko`**
     * Docker options to be set when starting the container (List or String).
     */
    'dockerOptions',
    /**
     * **For `buildType: kaniko`**
     * Name of the docker image that should be used. If empty, Docker is not used and the command is executed directly on the Jenkins system.
     */
    'dockerImage',
    /**
     * **For xMake stage/promote builds**
     * Defines the name of the GitHub organization
     */
    'githubOrg',
    /**
     * **For xMake stage/promote builds**
     * Defines the name of the GitHub repository
     */
    'githubRepo',
    /**
     * **For xMake stage/promote builds**
     * Defines the git commit id which should be the source of the xMake build.
     * It does not need to be provided during a typical pipeline run.
     * It defaults to `globalPipelineEnvironment.getGitCommitId()`
     */
    'gitCommitId',
    /** Defines if a helm package should be created.
     * @possibleValues true, false
     */
    'helmExecute',
    /** Defines if a container image should be created using Kaniko using the artifact created defined by `buildTool`.
     * @possibleValues true, false
     */
    'kanikoExecute',
    /**
     * **For xMake stage builds**
     * Defines if the generated build result (containing build result, test results, ...) should be downloaded.
     * **User with caution: If deactivated test results, etc. cannot be taken into consideration and pipeline might continue without noticing test and check failures.**
     * @possibleValues 'true', 'false'
     * @mandatory for buildType=xMakeStage/xMakePromote
     */
    'xMakeDownloadBuildResult',
    /**
     * **For xMake builds**
     * Defines the name of the xMake job name.
     * **Use with caution: Typically there is no need to change this since the name is calculated at runtime.**
     */
    'xMakeJobName',
    /**
     * **For xMake builds**
     * Defines how the name of the xMake job name is composed.
     * **Use with caution: Typically there is no need to change this**
     */
    'xMakeJobNameTemplate',
    /**
     * **For xMake builds**
     * Defines additional xMake configuration options that are passed as job parameters to the xMake job.
     * Definition is done as a list of `KEY=some xMake option` job parameters which are added besides the preconfigured MODE and TREEISH
     * Several parsing levels from Piper until the xMake plugin requires the corresponding escaping levels
     * Example to pass docker options to the xMake `dockerbuild` plugin from Piper
     * source: piper syntax is `['BUILD_OPTIONS=\"--buildplugin-option options=\\"--build-arg APP_VERSION=${BUILD_NUMBER}\\"\"']`
     * destination: dockerbuild plugin will receive the appropriate `--buildplugin-option options="--build-arg APP_VERSION=27"`
     * Please note: MODE and TREEISH will be fully controlled by the pipeline and cannot be overwritten.
     */
    'xMakeJobParameters',
    /**
     * **For xMake builds**
     * Defines the type of the shipment.
     * This will be part of the automatically generated job name.
     */
    'xMakeShipmentType',
    /**
     * **For xMake promote builds**
     * Defines the id of the repository containing the artifact created during the xMake stage build.
     * **Use with caution: This is handled automatically and retrieved from  `globalPipelineEnvironment.getXMakeProperties()`**
     */
    'xMakeStagingRepoId',
    /**
     * **Deprecated, please use `buildType: kaniko` or  step `dockerExecute` instead**
     * Only for `buildType: dockerInside`
     * Defines the arguments which should be passed to the Docker container which runs the build.
     * @mandatory for buildType=xMakePromote
     */
    'dockerArguments',
    /**
     * **Deprecated, please use `buildType: kaniko` or  step `dockerExecute` instead**
     * Only for `buildType: dockerInside`
     * Defines the command which should be executed inside the Docker image.
     */
    'dockerCommand', // used for dockerInside
    /**
     * Only for `buildType: dockerLocal`
     * Defines the name of the docker image (incl. tag) which will be built.
     * Defaults to `globalPipelineEnvironment.getDockerImageNameAndTag()`.
     * @mandatory for buildType=dockerLocal
     */
    'dockerImageNameAndTag',
    /**
     * **Deprecated**
     * It is only relevant if you build a Docker container using an application artifact which has been previously been build within the same pipeline.
     */
    'artifactType',
    /**
     * **Only for OLD functionality of NaaS release forks.**
     * Defines a specific git user email for the commit to the release fork.
     * Defaults to `globalPipelineEnvironment.getArtifactVersion()`
     */
    'artifactVersion',
    /**
     * **Only for OLD functionality of NaaS release forks.**
     * Defines git ssh url to the release fork repository.`
     */
    'xMakeNaasRepository',
    /**
     * **Deprecated - Do not use it any longer!**
     * xMake server resolution happens automatically now. This option is only kept for compatibility reasons.
     */
    'xMakeServer',
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * In this step a product standard compliant build will be executed according to [SLC-29](https://wiki.wdf.sap.corp/wiki/display/pssl/SLC-29).
 *
 * It fully leverages the capabilities of the stage-promote build pattern (either using a "native build on JaaS" or using the xMake landscape) and therefore ensures a productive build according to the **build only once** paradigm:
 *
 * * Software build after developer commits change to master
 * * Qualification of this build in multiple stages
 * * Release to customers (promote) without the necessity for building the software again and therefore making sure that we deliver exactly the artifact which has been tested.
 *
 * ### Details about possible build options
 *
 * 1. [Piper native build on JaaS](../build/native.md)
 * 2. [xMake](../build/xMake.md)
 *
 */
@GenerateDocumentation
void call(Map parameters = [:], body = '') {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, failOnError: true,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def script = checkScript(this, parameters) ?: this
        def stageName = parameters.stageName?:env.STAGE_NAME
        def utils = parameters.juStabUtils ?: new Utils()
        // handle deprecated parameters
        Deprecate.parameter(this, parameters, 'gitCredentialsId', 'gitSshKeyCredentialsId')
        Deprecate.parameter(this, parameters, 'gitSSHCredentialsId', 'gitSshKeyCredentialsId')
        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(
                dockerImageNameAndTag: script.globalPipelineEnvironment.getDockerImageNameAndTag()
            )
            .mixin(parameters, PARAMETER_KEYS)
            .dependingOn('buildType').mixin('containerBuildOptions')
            .dependingOn('buildType').mixin('containerCommand')
            .dependingOn('buildType').mixin('containerShell')
            .dependingOn('buildType').mixin('dockerImage')
            .dependingOn('buildType').mixin('dockerOptions')
            .addIfEmpty('cnbBuild', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.cnbBuild)
            .addIfEmpty('helmExecute', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.helmExecute)
            .addIfEmpty('kanikoExecute', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.kanikoExecute)
            .addIfEmpty('profiles', script.commonPipelineEnvironment.getStepConfiguration('mavenBuild', stageName)?.profiles)
            .addIfEmpty('sapCumulusUpload', script.commonPipelineEnvironment.getStepConfiguration('sapCumulusUpload', stageName)?.pipelineId ? true : false )
            .use()

        def globalSettingsFile = 'https://int.repositories.cloud.sap/artifactory/build-releases/settings.xml'
        def defaultNpmRegistry = (config.buildQuality == 'Milestone') ? 'https://int.repositories.cloud.sap/artifactory/api/npm/build-milestones-npm' : 'https://int.repositories.cloud.sap/artifactory/api/npm/build-releases-npm'

        // feature flag for adding functionality around xMake-free builds
        if (config.nativeBuild) {
            Notify.deprecatedStep(this, "executeBuild", "", script?.commonPipelineEnvironment)

            if (config.buildType == 'stage' || (config.nativeVoting && env.CHANGE_ID != null)) {
                try {
                    // open group via staging service
                    sapCallStagingService script: script, action: 'createGroup'

                    Map buildSettings = [:]
                    if (config.buildQuality == 'Milestone') {
                        buildSettings['buildQuality'] = 'Milestone'
                    } else {
                        buildSettings['buildQuality'] = 'Release'
                    }

                    // track that executeBuild was used
                    script.commonPipelineEnvironment.setValue('nativeBuild', true)
                    buildSettings['PiperNativeBuild'] = true

                    script.commonPipelineEnvironment.setValue('buildEnv', 'Hyperspace_Jenkins_native_BuildStep')
                    buildSettings['buildEnv'] = 'Hyperspace_Jenkins_native_BuildStep'

                    sapCallStagingService script: script, action: 'createRepositories'

                    // ToDo: allow multiple builds per group
                    switch (config.buildTool) {
                        case 'docker':
                            executeDockerBuildTool(config, globalSettingsFile, defaultNpmRegistry, script)
                            break
                        case 'maven':
                            builtinProfiles = (config.buildQuality == 'Milestone') ? ['!snapshot.build', 'milestone.build'] : ['!snapshot.build', '!milestone.build', 'release.build']
                            buildSettings['profiles'] = builtinProfiles

                            if (!config.profiles) {
                                echo "User-defined profiles not set in mavenBuild step settings"
                            } else {
                                buildSettings['profiles'].addAll(config.profiles)
                            }

                            buildSettings['globalSettingsFile'] = globalSettingsFile

                            Map mavenBuildSettings = [:]
                            mavenBuildSettings << buildSettings
                            mavenBuildSettings['script'] = script

                            mavenBuild mavenBuildSettings
                            break
                        case 'npm':
                            buildSettings['defaultNpmRegistry'] = defaultNpmRegistry

                            Map npmBuildSettings = [:]
                            npmBuildSettings << buildSettings
                            npmBuildSettings['script'] = script

                            npmExecuteScripts npmBuildSettings
                            break
                        case 'mta':
                            buildSettings['defaultNpmRegistry'] = defaultNpmRegistry
                            buildSettings['globalSettingsFile'] = globalSettingsFile
                            buildSettings['profiles'] = (config.buildQuality == 'Milestone') ? ['milestone.build'] : ['release.build']

                            Map mtaBuildSettings = [:]
                            mtaBuildSettings << buildSettings
                            mtaBuildSettings['script'] = script

                            mtaBuild mtaBuildSettings
                            break
                        case 'golang':
                            Map golangBuildSettings = [:]
                            golangBuildSettings << buildSettings
                            golangBuildSettings['script'] = script

                            golangBuild golangBuildSettings
                            break
                        case 'pip':
                            Map pythonBuildSettings = [:]
                            pythonBuildSettings << buildSettings
                            pythonBuildSettings['script'] = script

                            pythonBuild pythonBuildSettings
                            break
                        case 'gradle':
                            Map gradleBuildSettings = [:]
                            gradleBuildSettings << buildSettings
                            gradleBuildSettings['script'] = script

                            gradleExecuteBuild gradleBuildSettings
                            break
                        case 'helm':
                            config.helmExecute = true
                            break
                        default:
                            Notify.error(this, "invalid buildTool '${config.buildTool}' for native build - '${config.buildTool}' not supported")
                    }

                    // let's create a Docker image on top of the architecture build (e.g. maven, npm)
                    if (config.buildTool != 'docker' && (config.cnbBuild || config.kanikoExecute)) {
                        if (config.cnbBuild) {
                            Map cnbBuildSettings = ['script': script]

                            switch(config.buildTool){
                            case 'npm':
                                script.commonPipelineEnvironment.setContainerProperty('buildpacks', ["gcr.io/paketo-buildpacks/nodejs"])
                                break
                            case 'gradle':
                            case 'maven':
                                script.commonPipelineEnvironment.setContainerProperty('buildpacks', ["gcr.io/paketo-buildpacks/java"])
                                break
                            case 'mta':
                                // List of buildpacks and paths should be explicitly configured via 'config.yaml' file.
                                break
                            default:
                                throw new AbortException("ERROR - 'cnbBuild' does not support '${config.buildTool}' as a buildTool, consider using 'kanikoExecute' instead")
                            }

                            cnbBuild cnbBuildSettings
                        }
                        if (config.kanikoExecute) {
                            kanikoExecute script: script
                        }
                    }
                    if (config.helmExecute) {
                        Map helmExecuteSettings = [:]
                        helmExecuteSettings << buildSettings
                        helmExecuteSettings['script'] = script

                        helmExecute helmExecuteSettings
                    }
                } finally {
                    // cleanup / close staging group
                    sapCallStagingService script: script, action: 'close'

                    // only perform build settings creation and Cumulus upload for main pipeline runs - not for voting
                    if (!(config.nativeVoting && env.CHANGE_ID != null)) {
                        // generate and upload build-settings.json
                        sapGenerateEnvironmentInfo script: script
                        sapCumulusUpload script: script, filePattern: 'build-settings.json', stepResultType: 'settings'
                        // push HS assessment file to track assessment changes
                        sapCumulusUpload script: script, filePattern: 'hs-assessments.yaml', stepResultType: 'assessment'
                        sapCumulusUpload script: script, filePattern: '**/bom-*.xml', stepResultType: 'sbom'

                        // push http logs to Cumulus
                        sapCumulusUpload script: script, filePattern: '**/url-log.json', stepResultType: 'access-log'
                    }
                }
                return
            } else if (config.buildType == 'promote' || config.buildType == 'xMakePromote') {
                sapCallStagingService script: script, action: 'promote'
                return
            } else {
                Notify.error(this, "invalid buildType '${config.buildType}' for native build - only 'stage' and 'promote' supported")
            }
        }

        // from here only executed in case NOT nativeBuild

        if(config.artifactType == 'appContainer'){
            // appContainer option is DEPRECATED and will be removed by end of Q2/2021
            // Deprecate.value(script, config, 'appContainer')
            config = new ConfigurationHelper(config)
                .addIfEmpty('artifactVersion', script.globalPipelineEnvironment.getAppContainerProperty('artifactVersion'))
                .addIfEmpty('githubOrg', script.globalPipelineEnvironment.getAppContainerProperty('githubOrg'))
                .addIfEmpty('githubRepo', script.globalPipelineEnvironment.getAppContainerProperty('githubRepo'))
                .addIfEmpty('gitCommitId', script.globalPipelineEnvironment.getAppContainerProperty('gitCommitId'))
                .addIfEmpty('xMakeStagingRepoId', script.globalPipelineEnvironment.getAppContainerProperty('xMakeProperties')?.staging_repo_id)
                .use()
        }else{
            config = new ConfigurationHelper(config)
                .addIfEmpty('artifactVersion', script.globalPipelineEnvironment.getArtifactVersion())
                .addIfEmpty('githubOrg', script.globalPipelineEnvironment.getGithubOrg())
                .addIfEmpty('githubRepo', script.globalPipelineEnvironment.getGithubRepo())
                .addIfEmpty('gitCommitId', script.globalPipelineEnvironment.getGitCommitId())
                .addIfEmpty('xMakeStagingRepoId', script.globalPipelineEnvironment.getXMakeProperties()?.staging_repo_id)
                .use()
        }

        switch(config.buildType){
            case 'xMakeStage':
                sapXmakeExecuteBuild(parameters.plus([script: script, buildType: 'xMakeStage']))

                script.commonPipelineEnvironment.setValue('xMakeBuild', true)
                script.commonPipelineEnvironment.setValue('buildEnv', 'Hyperspace_Jenkins_xMake_BuildStep')

                def buildResultFile = "xmake_stage.json"
                def dockerMetadataFile = "docker.metadata.json"

                if (fileExists(buildResultFile)){
                    def buildResultJSON = readJSON(file: buildResultFile)
                    if (config.xMakeDownloadBuildResult)
                        downloadXMakeBuildResult(buildResultJSON, config.xMakeDownloadDownstreamsProjectArchives)

                    if (config.artifactType == 'appContainer') {
                        script.globalPipelineEnvironment.setAppContainerProperty('xMakeProperties', buildResultJSON)
                    } else {
                        script.globalPipelineEnvironment.setXMakeProperties(buildResultJSON)
                    }
                }

                if (fileExists(dockerMetadataFile)){
                    def dockerMedatadaJSON = readJSON(file: dockerMetadataFile)
                    if (config.artifactType == 'appContainer') {
                            script.globalPipelineEnvironment.setAppContainerDockerMetadata(dockerMedatadaJSON)
                    } else {
                            script.globalPipelineEnvironment.setDockerMetadata(dockerMedatadaJSON)
                    }
                }
                 // only perform piper config for ppms  creation and Cumulus upload for main pipeline runs - not for voting
                if (!(config.nativeVoting && env.CHANGE_ID != null) && config.sapCumulusUpload) {
                    sapGenerateEnvironmentInfo script: script, generateFiles:["piperConfig"]
                    sapCumulusUpload script: script, filePattern: 'piper-config.yaml', stepResultType: 'config'
                }

                break;
            case 'xMakePromote':
                sapXmakeExecuteBuild(parameters.plus([script: script, buildType: 'xMakePromote']))
                break;
            case 'dockerLocal':
                new ConfigurationHelper(config).withMandatoryProperty('dockerImageNameAndTag')
                def dockerBuildImage = docker.build(config.dockerImageNameAndTag, "${config.containerBuildOptions} .")
                script.globalPipelineEnvironment.setDockerBuildImage(dockerBuildImage)
                break;
            //option 'kaniko' so far only suitable for PR-voting since no push to a registry
            case 'kaniko':
                try {
                    dockerExecute(
                        script: script,
                        containerCommand: config.containerCommand,
                        containerShell: config.containerShell,
                        dockerImage: config.dockerImage,
                        dockerOptions: config.dockerOptions
                    ) {
                        sh """#!${config.containerShell}
mv /kaniko/.docker/config.json /kaniko/.docker/config.json.bak
mv /kaniko/.config/gcloud/docker_credential_gcr_config.json /kaniko/.config/gcloud/docker_credential_gcr_config.json.bak
/kaniko/executor --dockerfile ${env.WORKSPACE}/Dockerfile --context ${env.WORKSPACE} ${config.containerBuildOptions}"""
                    }
                } catch (ex) {
                    Notify.error(this, "The execution of the kaniko build failed, see the log for details.")
                }
                break;
            case 'dockerInside':
                Notify.warning(script, 'Option dockerInside is deprecated and will soon be removed.', STEP_NAME)
                def image = docker.image(config.dockerImage?:'')
                image.pull()
                image.inside(config.dockerArguments?:'') {
                    if (config.dockerCommand) {
                        try {
                            sh "${config.dockerCommand}"
                        } catch (ex) {
                            Notify.error(this,  "The execution of the dockerInside build failed, see the log for details.")
                        }
                    } else {
                        body()
                    }
                }
                break;
        }
    }
}

def executeDockerBuildTool(config, globalSettingsFile, defaultNpmRegistry, script) {
    if (config.cnbBuild) {
        Map cnbBuildSettings = [:]
        cnbBuildSettings["script"] = script
        cnbBuildSettings["bindings"] = [
            maven: [
                type: "maven",
                key: "settings.xml",
                fromUrl: globalSettingsFile
            ],
            npmrc: [
                type: "npmrc",
                key: ".npmrc",
                content: "registry=${defaultNpmRegistry}"
            ]
        ]
        cnbBuild cnbBuildSettings
    } else {
        kanikoExecute script: script
    }
}

def _downloadProjectArchiveFile(url, folderName=null) {
    if(url){
        try {
            sh "curl --insecure --silent --show-error --location --output build.tar ${url}"
            if(folderName) {
                sh "mkdir -p ${folderName}"
            }
            sh "tar -zxvf build.tar"+(folderName?" --directory ${folderName}":"")+" > extractProjectArchive.log"
            sh "rm -f build.tar"
            if(folderName && fileExists("${folderName}/docker.metadata.json") && !fileExists('docker.metadata.json')) {
                sh "cp '${folderName}/docker.metadata.json' 'docker.metadata.json'"
            }
        } catch (ex) {
            Notify.error(this,  "Downloading xMake build result failed, see the log for details.")
        }
    }else{
        Notify.error(this, "The url to the remote job's deploy package is not available in 'build-results.json'.")
    }

}

//
// Method to download projectArchive (deployPackage.tar.gz) files referenced in the build-results.json
// It will unzip them to the appropriate location looking to the xMakeDownloadDownstreamsProjectArchives mode
//
// @param buildResult build-results.json file of the build retrieved previously with for example: buildHandler.readJsonFileFromBuildArchive('build-results.json')
// @param xMakeDownloadDownstreamsProjectArchives false by default, will try to find a unique mandatory projectArchive entry and to download the corresponding file & unzip it in the current directory.
//     if set to true, will search for multiple projectArchive files in a projectArchiveFiles json map and unzip each of them in their corresponding downstream name subfolder
// @return nothing, void method
//
def downloadXMakeBuildResult(buildResult, xMakeDownloadDownstreamsProjectArchives=false) {
    if (buildResult == null) {
        Notify.error(this, "The xMake build did not return a 'build-results.json', please check your xMake configuration!")
    }
    if(xMakeDownloadDownstreamsProjectArchives) {
        //
        // New format when xMakeDownloadDownstreamsProjectArchives is enabled in the general settings. A subdirectory named with the map key for each entry (for example here: default_1_i051432_dtxmake_3374_SP_MS_linuxx86_64 will) will be created with the unzipped content
        // The build can return a projectArchiveFiles map with multiple entries if it has got multiple variants
        // {
        //  "projectArchiveFiles": {
        //   "default_1_i051432_dtxmake_3374_SP_MS_linuxx86_64": "http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.snapshots.blob/com/sap/production-services/build-service/i051432/DTXMAKE-3374/43083338f72764c28dac98fcd2f86ba597bc8185/2020_05_27__13_07_31/deployPackage.tar.gz"
        //   "default_1_i051432_dtxmake_3374_SP_MS_ntmad64": "http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.snapshots.blob/com/sap/production-services/build-service/i051432/DTXMAKE-3374/43083338f72764c28dac98fcd2f86ba597bc8185/2020_05_27__13_07_35/deployPackage.tar.gz"
        //  }
        // }
        //
        // This will produce:
        // 21:18:10  [Pipeline] sh
        // 21:18:10  + curl --insecure --silent --show-error --location --output build.tar http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.snapshots.blob/com/sap/production-services/build-service/i051432/DTXMAKE-3374/fd51fd98d31a272bff3a8fe08c8f5c03d398eaea/2020_05_28__19_17_44/deployPackage.tar.gz
        // 21:18:11  [Pipeline] sh
        // 21:18:11  + mkdir -p default_1_i051432_dtxmake_3374_SP_MS_linuxx86_64
        // 21:18:11  [Pipeline] sh
        // 21:18:12  + tar -zxvf build.tar --directory default_1_i051432_dtxmake_3374_SP_MS_linuxx86_64
        // 21:18:12  [Pipeline] sh
        // 21:18:13  + rm -f build.tar
        // 21:18:10  [Pipeline] sh
        // 21:18:10  + curl --insecure --silent --show-error --location --output build.tar http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.snapshots.blob/com/sap/production-services/build-service/i051432/DTXMAKE-3374/43083338f72764c28dac98fcd2f86ba597bc8185/2020_05_27__13_07_35/deployPackage.tar.gz
        // 21:18:11  [Pipeline] sh
        // 21:18:11  + mkdir -p default_1_i051432_dtxmake_3374_SP_MS_ntmad64
        // 21:18:11  [Pipeline] sh
        // 21:18:12  + tar -zxvf build.tar --directory default_1_i051432_dtxmake_3374_SP_MS_ntmad64
        // 21:18:12  [Pipeline] sh
        // 21:18:13  + rm -f build.tar
        //
        buildResult.projectArchiveFiles?.each  { downstream, url ->
            _downloadProjectArchiveFile(url, downstream)
        }
    } else {
        //
        // Legacy format containing only one archive file which will be downloaded & unzipped in the curruent workspace directory for backward compatibility
        // {
        //  "projectArchive": "http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.snapshots.blob/com/sap/production-services/build-service/i051432/DTXMAKE-3374/43083338f72764c28dac98fcd2f86ba597bc8185/2020_05_27__13_07_31/deployPackage.tar.gz",
        // }
        // This will produce:
        // 20:37:43  [Pipeline] sh
        // 20:37:43  + curl --insecure --silent --show-error --location --output build.tar http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.snapshots.blob/com/sap/production-services/build-service/i051432/DTXMAKE-3374/b3d84a845c7b52e010400b5b22610ee1b9565d3d/2020_05_28__18_37_16/deployPackage.tar.gz
        // 20:37:44  [Pipeline] sh
        // 20:37:44  + tar -zxvf build.tar
        // 20:37:44  [Pipeline] sh
        // 20:37:45  + rm -f build.tar
        //
        _downloadProjectArchiveFile(buildResult.projectArchive ?: buildResult.projectarchive)
    }
}

def getIdentifierFromSSHLink(url){
    return url // git@github.wdf.sap.corp:xyz/abc.git
        .tokenize(':').get(1) // -> xyz/abc.git
        .tokenize('.').get(0) // -> xyz/abc
        .tokenize('/') // -> [xyz, abc]
}

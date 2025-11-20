import hudson.AbortException

import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import groovy.transform.Field
import com.sap.piper.internal.Notify

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'sapPiperStageCentralBuild'

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * For native builds only: Defines the build quality for the native build. For productive pipelines this should always be set to `Release`.
     * @possibleValues 'Milestone', 'Release'
     */
    'buildQuality',
    /**
     * For native builds only: Defines the tool which is used for building the artifact.
     */
    'buildTool',
    /**
     * Activate a native build (e.g. maven, npm) in combination with SAP's staging service.
     * @possibleValues 'true', 'false'
     */
    'nativeBuild',
    /**
     * Activate a native build pull-request voting (e.g. maven, npm) only in combination with "Piper native build".
     * @possibleValues 'true', 'false'
     */
    'nativeVoting',
    /**
     * Only for DEBUG purposes, switches off execution of `executeBuild` step
     * @possibleValues `true`, `false`
     */
    'skipBuild',
     /**
     * Defines the main branch for your pipeline. **Typically this is the `master` branch, which does not need to be set explicitly.** Only change this in exceptional cases. Supports regular expression through Groovy Match operator, e.g. `master|develop`.
     */
    'productiveBranch',
    /**
     * Parameter in Beta mode.
     * To be set to true if env.json, bom xmls and build-settings.json are to be uploaded for a Pull Request.
     * @possibleValues `true`, `false`
     */
    'uploadCumulusFilesforPR',
    /*
     secretID, groupID to call systemTrust
    */
    'vaultAppRoleSecretTokenCredentialsId',
    'vaultBasePath',
    'vaultPipelineName'
]
@Field STAGE_STEP_KEYS = [
    /** Defines if a container image should be created using Cloud Native Buildpacks using the artifact created defined by `buildTool`.
     * @possibleValues true, false
     */
    'cnbBuild',
    /** Starts build execution, typically on xMake landscape in order to comply with SAP's corporate requirement [SLC-29](https://wiki.wdf.sap.corp/wiki/display/pssl/SLC-29). */
    'executeBuild',
    /** Executes Haskell Docker Linter in docker based scenario to analyse structural issues in the Dockerfile. */
    'hadolintExecute',
    /** Defines if a helm package should be created.
     * @possibleValues true, false
     */
    'helmExecute',
    /** Allows you to copy a Docker image from a source container registry  to a destination container registry. The imagePushToRegistry is not similar in functionality to containerPushToRegistry. Currently the imagePushToRegistry only supports copying a local image or image from source remote registry to destination registry.*/
    'imagePushToRegistry',
    /** Defines if a container image should be created using Kaniko using the artifact created defined by `buildTool`.
     * @possibleValues true, false
     */
    'kanikoExecute',
    /** Executes karma tests which is for example suitable for OPA5 testing as well as QUnit testing of SAP UI5 apps.*/
    'karmaExecuteTests',
    /** Executes static code checks for Maven based projects. The plugins SpotBugs and PMD are used. */
    'mavenExecuteStaticCodeChecks',
    /** Executes linting for npm projects. */
    'npmExecuteLint',
    /** Executes npm */
    'npmExecuteScripts',
    /** Executes API metadata validation. */
    'sapExecuteApiMetadataValidator',
    /** Executes stashing of files after build execution. Relevant for files which are only created in the build run, e.g. `*.js` files when using TypeScript.*/
    'pipelineStashFilesAfterBuild',
    /** For Docker builds: Automatically pushes created Docker image to a dedicated container registry. It will only be executed in case a `docker.metadata.json` is available from xMake since this contains the required image details.*/
    'pushToDockerRegistry',
    /** Executes a Sonar scan.*/
    'sonarExecuteScan',
    /** Publishes test results to Jenkins. It will always be active. */
    'testsPublishResults',
    /** Publishes check results to Jenkins. It will always be active. */
    'checksPublishResults',
    /** Performs upload of result files to cumulus. */
    'sapCumulusUpload',
    /** Creates an open-component-model (OCM) component-version. */
    'sapOcmCreateComponent',
    /** Analyse given proxy log file to detect insecure URLs.*/
    'sapURLScan'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    def stageName = parameters.stageName?:env.STAGE_NAME
    // check which stageName is configured in config.yml and use it
    // if both are configured, stick with 'Build'
    if (stageName == 'Central Build') {
        def stages = script.globalPipelineEnvironment.configuration.stages
        if (stages != null && (!stages.containsKey('Central Build') || stages.containsKey('Build'))) {
            stageName = 'Build'
        } else {
            stageName = 'Central Build'
        }
    }

    def isPR = env.CHANGE_ID != null

    // npmExecuteScripts is special , since this step is called in PR voting and in build stage as well. as npmExecuteScripts is also a generic step and not build specefic
    // in case when native voting and native build is true this step will run twice
    def isNPMExecuteScriptsRun = false

    //ToDo: Update stashing to be able to extend current stage
    piperStageWrapper (script: script, stageName: stageName) {

        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .addIfEmpty('cnbBuild', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.cnbBuild)
            .addIfEmpty('hadolintExecute', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.hadolintExecute)
            .addIfEmpty('helmExecute', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.helmExecute)
            .addIfEmpty('imagePushToRegistry', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.imagePushToRegistry)
            .addIfEmpty('kanikoExecute', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.kanikoExecute)
            .addIfEmpty('karmaExecuteTests', false)
            .addIfEmpty('mavenExecuteStaticCodeChecks', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.mavenExecuteStaticCodeChecks)
            .addIfEmpty('mavenBuild', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.mavenBuild)
            .addIfEmpty('profiles', script.commonPipelineEnvironment.getStepConfiguration('mavenBuild', stageName)?.profiles)
            .addIfEmpty('npmExecuteLint', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.npmExecuteLint)
            .addIfEmpty('npmExecuteScripts', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.npmExecuteScripts)
            .addIfEmpty('sapCallStagingService', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.sapCallStagingService)
            .addIfEmpty('sapExecuteApiMetadataValidator', false)
            .addIfEmpty('pushToDockerRegistry', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.pushToDockerRegistry)
            .addIfEmpty('sonarExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.sonarExecuteScan)
            .addIfEmpty('sapCumulusUpload', script.commonPipelineEnvironment.getStepConfiguration('sapCumulusUpload', stageName)?.pipelineId ? true : false )
            .addIfEmpty('sapOcmCreateComponent', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.sapOcmCreateComponent)
            .addIfEmpty('sapURLScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.sapURLScan)
            .addIfEmpty('nativeBuild', false)
            .use()

        def isProductiveBranch = env.BRANCH_NAME ==~ config.productiveBranch
        durationMeasure(script: script, measurementName: 'build_duration') {
            echo "Getting System trust token"
            def token = null
            try {
                def apiURLFromDefaults = script.commonPipelineEnvironment.getValue("hooks")?.systemtrust?.serverURL ?: ''
                token = sapGetSystemTrustToken(apiURLFromDefaults, config.vaultAppRoleSecretTokenCredentialsId, config.vaultPipelineName, config.vaultBasePath)
            } catch (Exception e) {
                echo "Couldn't get system trust token, will proceed with configured credentials: ${e.message}"
            }
            wrap([$class: 'MaskPasswordsBuildWrapper', varPasswordPairs: [[password: token]]]) {
                withEnv([
                    /*
                    Additional logic "?: ''" is necessary to ensure the environment
                    variable is set to an empty string if the value is null
                    Without this, the environment variable would be set to the string "null",
                    causing checks for an empty token in the Go application to fail.
                    */
                    "PIPER_systemTrustToken=${token ?: ''}",
                ]) {
                    if (!config.skipBuild) {
                        if (config.nativeBuild) {
                            // track that native build was used
                            script.commonPipelineEnvironment.setValue('nativeBuild', true)
                            script.commonPipelineEnvironment.setValue('buildEnv', 'Hyperspace_Jenkins_native_BuildStep')
                            // set default build quality
                            //TODO: move to defaults.yaml?
                            config.buildQuality = config.buildQuality == 'Milestone' ? 'Milestone' :'Release'

                            Map buildSettings = [:]
                            buildSettings['PiperNativeBuild'] = true
                            buildSettings['buildEnv'] = 'Hyperspace_Jenkins_native_BuildStep'
                            buildSettings['buildQuality'] = config.buildQuality

                            try {
                                if (!config.sapCallStagingService) {
                                    echo "sapCallStagingService step skipped (Central Build stage)"
                                } else {
                                    stagingServiceWrapper(script){
                                        def globalSettingsFile = 'https://int.repositories.cloud.sap/artifactory/build-releases/settings.xml'
                                        def defaultNpmRegistry = "https://int.repositories.cloud.sap/artifactory/api/npm/build-${config.buildQuality.toLowerCase()}s-npm"
                                        // ========> main build
                                        // ToDo: allow multiple builds per group
                                        // TODO: use checkIfStepActive
                                        switch (config.buildTool) {
                                            case ['maven', 'CAP']:
                                                if (config.mavenBuild) {
                                                    builtinProfiles = (config.buildQuality == 'Milestone') ? ['!snapshot.build', 'milestone.build'] : ['!snapshot.build', '!milestone.build', 'release.build']
                                                    buildSettings['profiles'] = builtinProfiles

                                                    if (!config.profiles) {
                                                        echo "User-defined profiles not set in mavenBuild step settings"
                                                    } else {
                                                        buildSettings['profiles'].addAll(config.profiles)
                                                    }

                                                    mavenBuild(['script': script].plus(buildSettings))
                                                }
                                                // fallthrough if build tool is CAP
                                                if (config.buildTool == "maven") {
                                                    break
                                                }
                                            case 'npm':
                                                if (config.npmExecuteScripts) {
                                                    buildSettings['defaultNpmRegistry'] = defaultNpmRegistry
                                                    npmExecuteScripts(['script': script].plus(buildSettings))
                                                    isNPMExecuteScriptsRun = true
                                                }
                                                break
                                            case 'mta':
                                                buildSettings['profiles'] = (config.buildQuality == 'Milestone') ? ['milestone.build'] : ['release.build']
                                                buildSettings['defaultNpmRegistry'] = defaultNpmRegistry
                                                mtaBuild(['script': script].plus(buildSettings))
                                                break
                                            case 'golang':
                                                golangBuild(['script': script].plus(buildSettings))
                                                break
                                            case 'pip':
                                                pythonBuild(['script': script].plus(buildSettings))
                                                break
                                            case 'gradle':
                                                gradleExecuteBuild(['script': script].plus(buildSettings))
                                                break
                                            case ['docker', 'custom']: // handled below
                                                break
                                            default:
                                                Notify.error(this, "invalid buildTool '${config.buildTool}' for native build - '${config.buildTool}' not supported")
                                        }

                                        // ========> package into Container (cnb)
                                        if (config.cnbBuild) {
                                            if (!['npm', 'gradle', 'maven', 'mta', 'docker', 'CAP'].contains(config.buildTool)) {
                                                throw new AbortException("ERROR - 'cnbBuild' does not support '${config.buildTool}' as a buildTool, consider using 'kanikoExecute' instead")
                                            }

                                            Map cnbBuildSettings = ['script': script]
                                            cnbBuildSettings.bindings = [
                                                maven: [type: "maven", data: [[key: "settings.xml", fromUrl: globalSettingsFile]]],
                                                npmrc: [type: "npmrc", data: [[key: ".npmrc", content: "registry=${defaultNpmRegistry}"]]]
                                            ]
                                            cnbBuild cnbBuildSettings
                                        }
                                        // ========> package into Container (kaniko)
                                        // TODO: check if kanikoExecute is already set in that case
                                        // TODO: migrate to stage condition
                                        if (config.kanikoExecute || (config.buildTool == 'docker' && !config.cnbBuild)) {
                                            kanikoExecute script: script
                                        }
                                        // ========> package into Helm
                                        if (config.helmExecute) {
                                            helmExecute(['script': script].plus(buildSettings))
                                        }
                                        if (config.sapOcmCreateComponent) {
                                            sapOcmCreateComponent script: script
                                        }
                                    }
                                }
                            } finally {
                                // BOM uploading needed for PR for ACT Hyperspace PR Voting
                                def isActPR = config.nativeVoting && isPR && config.sapCumulusUpload && config.uploadCumulusFilesforPR

                                // Uploading files either for a productive branch or an ACT PR
                                if ((isProductiveBranch || isActPR) && config.sapCumulusUpload){
                                    echo "Uploading hs-assessments, BOM, build-settings.json, piper config to cumulus"
                                    // push HS assessment file to track assessment changes
                                    sapCumulusUpload script: script, filePattern: 'hs-assessments.yaml', stepResultType: 'assessment'
                                    sapCumulusUpload script: script, filePattern: '**/bom-*.xml', stepResultType: 'sbom'
                                    // generate and upload build-settings.json
                                    sapGenerateEnvironmentInfo script: script
                                    sapCumulusUpload script: script, filePattern: 'build-settings.json', stepResultType: 'settings'
                                    // Upload build-settings.json for SLC-29 policy
                                    sapCumulusUpload script: script, filePattern: 'build-settings.json', stepResultType: 'policy-evidence/SLC-29-PNB'
                                    sapCumulusUpload script: script, filePattern: 'piper-config.yaml', stepResultType: 'config'

                                    // push http logs to Cumulus
                                    if (!(config.nativeVoting && isPR)) {
                                        sapCumulusUpload script: script, filePattern: '**/url-log.json', stepResultType: 'access-log'
                                    }
                                }
                            }
                        } else {
                            executeBuild script: script, stageName: stageName
                        }
                        pipelineStashFilesAfterBuild script: script, runCheckmarx: (script.globalPipelineEnvironment.configuration?.runStep?.Security?.executeCheckmarxScan ? true : null)
                    }

                    // nativeVoting related steps
                    try {
                        if (config.hadolintExecute) {
                            hadolintExecute script: script
                        }

                        if (config.karmaExecuteTests) {
                            durationMeasure(script: script, measurementName: 'opa_duration') {
                                karmaExecuteTests script: script
                            }
                        }

                        if (config.npmExecuteScripts && !isNPMExecuteScriptsRun) {
                            npmExecuteScripts script: script, publish: false, createBOM: false
                        }

                        if (config.sapCumulusUpload && !isPR) {
                            Map cumulusConfig = script.commonPipelineEnvironment.getStepConfiguration("testsPublishResults", stageName)

                            sapCumulusUpload script: script, filePattern: '**/requirement.mapping', stepResultType: 'requirement-mapping'

                            sapCumulusUpload script: script, filePattern: cumulusConfig.junit.pattern, stepResultType: 'junit'
                            sapCumulusUpload script: script, filePattern: '**/jacoco.xml', stepResultType: 'jacoco-coverage'
                            sapCumulusUpload script: script, filePattern: cumulusConfig.cobertura.pattern, stepResultType: 'cobertura-coverage'
                            sapCumulusUpload script: script, filePattern: '**/xmake_stage.json', stepResultType: 'xmake'
                            // Upload xmake_stage.json for SLC-29 Policy
                            sapCumulusUpload script: script, filePattern: '**/xmake_stage.json', stepResultType: 'policy-evidence/SLC-29-xMake'
                            sapCumulusUpload script: script, filePattern: '**/build-type.json', stepResultType: 'xmake'
                            sapCumulusUpload script: script, filePattern: '**/bom.xml', stepResultType: 'sbom'
                            // only needed for xmake builds. This is due to sboms from xmake generated in multiple formats
                            sapCumulusUpload script: script, filePattern: 'sbom/**/*', stepResultType: 'sbom'
                        }
                        //needs to run right after build, otherwise we may face "ERROR: Test reports were found but none of them are new"
                        testsPublishResults script: script, junit: [updateResults: true]

                        if (config.mavenExecuteStaticCodeChecks) {
                            mavenExecuteStaticCodeChecks(script: script)
                        }

                        if (config.npmExecuteLint) {
                            npmExecuteLint script: script
                        }

                        if (config.sapExecuteApiMetadataValidator) {
                            sapExecuteApiMetadataValidator script: script
                        }

                        if (config.sapURLScan) {
                            sapURLScan script: script
                        }

                        checksPublishResults script: script
                    } finally {
                        if (config.sonarExecuteScan) {
                            sonarExecuteScan script: script
                        }

                        if (config.sapCumulusUpload && !isPR) {
                            sapCumulusUpload script: script, filePattern: '**/sonarscan-result.json, **/sonarscan.json', stepResultType: 'sonarqube'
                            sapCumulusUpload script: script, filePattern: '**/sonarscan-result.json, **/sonarscan.json', stepResultType: 'policy-evidence/FC-1'
                        }

                        if (config.sapExecuteApiMetadataValidator && config.sapCumulusUpload && !isPR) {
                            sapCumulusUpload script: script, filePattern: '**/api-metadata-validator-results.json', stepResultType: 'api-metadata-validator'
                        }
                    }

                    if (fileExists('docker.metadata.json') && config.pushToDockerRegistry) {
                        pushToDockerRegistry script: script
                    }

                    if (config.imagePushToRegistry && !isPR) {
                        imagePushToRegistry script: script
                    }
                }
            }
        }
    }
}

void stagingServiceWrapper(script, body) {
    try{
        sapCallStagingService script: script, action: 'createGroup'
        sapCallStagingService script: script, action: 'createRepositories'
        body()
    } finally {
        sapCallStagingService script: script, action: 'close'
    }
}

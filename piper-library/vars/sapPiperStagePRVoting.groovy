import com.sap.icd.jenkins.Utils
import com.sap.icd.jenkins.TemplatingUtils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.ConfigurationLoader
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'sapPiperStagePRVoting'

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Defines the build tool used.
     * @possibleValues `maven`, `npm`, `mta`, ...
     */
    'buildTool',
    /** DEPRECATED */
    'dockerImageName',
    /**
     * Only for `buildTool: docker` **and not executed in a Kubernetes environment**.
     * Defines the name of the docker image (incl. tag) which will be build.
     * @mandatory for buildTool=docker & non-Kubernetes
     */
    'dockerImageNameAndTag',
    /**
     * Only for `buildTool: npm`
     * Defines the default registry used.
     */
    'defaultNpmRegistry',
    /*
     secretID, groupID to call systemTrust
    */
    'vaultAppRoleSecretTokenCredentialsId',
    'vaultBasePath',
    'vaultPipelineName'
]

@Field STAGE_STEP_KEYS = [
    /** Publishes check results to Jenkins. It will always be active. */
    'checksPublishResults',
    /** Executes a Checkmarx scan.<br />
     * This step is not active by default. It can be activated by:
     *
     * <ul>
     * <li>Using pull request comments or pull request lables (see [Advanced Pull-Request Voting](#advanced-pull-request-voting)</li>
     * <li>Explicit activation via stage configuration</li>
     * </ul>
     */
    'executeCheckmarxScan',
    /** Executes a Fortify scan.<br />
     * This step is not active by default. It can be activated by:
     *
     * <ul>
     * <li>Using pull request comments or pull request lables (see [Advanced Pull-Request Voting](#advanced-pull-request-voting)</li>
     * <li>Explicit activation via stage configuration</li>
     * </ul>
     */
    'executeFortifyScan',
    /**
     * Executes PPMS compliance check. This will also activate WhiteSource check since this is a pre-requisite.<br />
     * This step is not active by default. It can be activated by:
     *
     * <ul>
     * <li>Using pull request comments or pull request lables (see [Advanced Pull-Request Voting](#advanced-pull-request-voting)</li>
     * <li>Explicit activation via stage configuration</li>
     * </ul>
     */
    'executePPMSComplianceCheck',
    /**
     * Executes the open source vulnerability scan.<br />
     * This step is not active by default. It can be activated by:
     *
     * <ul>
     * <li>Using pull request comments or pull request lables (see [Advanced Pull-Request Voting](#advanced-pull-request-voting)</li>
     * <li>Explicit activation via stage configuration</li>
     * </ul>
     */
    'executeOpenSourceDependencyScan',
    /**
     * Performs Checkmarx scan.<br />
     * This step is not active by default. It can be activated by:
     *
     * <ul>
     * <li>Using pull request comments or pull request lables (see [Advanced Pull-Request Voting](#advanced-pull-request-voting)</li>
     * <li>Explicit activation via stage configuration</li>
     * </ul>
     */
    'checkmarxExecuteScan',
    /**
     * Performs Checkmarx One scan.<br />
     * This step is not active by default. It can be activated by:
     *
     * <ul>
     * <li>Using pull request comments or pull request lables (see [Advanced Pull-Request Voting](#advanced-pull-request-voting)</li>
     * <li>Explicit activation via stage configuration</li>
     * </ul>
     */
    'checkmarxOneExecuteScan',
    /**
     * Performs a Fortify scan.<br />
     * This step is not active by default. It can be activated by:
     *
     * <ul>
     * <li>Using pull request comments or pull request lables (see [Advanced Pull-Request Voting](#advanced-pull-request-voting)</li>
     * <li>Explicit activation via stage configuration</li>
     * </ul>
     */
    'fortifyExecuteScan',
    /**
     * Performs a Codeql scan.<br />
     * This step is not active by default. It can be activated by:
     *
     * <ul>
     * <li>Using pull request comments or pull request lables (see [Advanced Pull-Request Voting](#advanced-pull-request-voting)</li>
     * <li>Explicit activation via stage configuration</li>
     * </ul>
     */
    'codeqlExecuteScan',
    /** Runs backend integration tests via maven in the module integration-tests/pom.xml */
    'mavenExecuteIntegration',
    /**
     * Perform docker build with kaniko in case buildTool is set to `docker`.
     */
    'kanikoExecute',
    /**
     * Performs a golang build.
     */
    'golangBuild',
    /**
     * Executes karma tests. For example suitable for OPA5 testing as well as QUnit testing of SAP UI5 apps.<br />
     * This step is not active by default. It can be activated by:
     *
     * <ul>
     * <li>Using pull request comments or pull request lables (see [Advanced Pull-Request Voting](#advanced-pull-request-voting)</li>
     * <li>Explicit activation via stage configuration</li>
     * </ul>
     */
    'karmaExecuteTests',
    /** Executes static code checks for Maven based projects.<br />
     * The plugins SpotBugs and PMD are used.
     */
    'mavenExecuteStaticCodeChecks',
    /** Executes linting for npm projects. */
    'npmExecuteLint',
    /** Executes npm scripts to run frontend unit tests.<br />
     * If custom names for the npm scripts are configured via the `runScripts` parameter the step npmExecuteScripts needs **explicit activation via stage configuration**. */
    'npmExecuteScripts',
    /** Executes a Sonar scan.*/
    'sonarExecuteScan',
    /** Executes a PPMS compliance check.<br />
     * This step is not active by default. It can be activated by:
     *
     * <ul>
     * <li>Using pull request comments or pull request lables (see [Advanced Pull-Request Voting](#advanced-pull-request-voting)</li>
     * <li>Explicit activation via stage configuration</li>
     * </ul>
     */
    'sapCheckPPMSCompliance',
    /** Publishes test results to Jenkins.<br />
     * It will always be active.
     */
    'testsPublishResults',
    /** Executes a WhiteSource scan.<br />
     * This step is not active by default. It can be activated by:
     *
     * <ul>
     * <li>Using pull request comments or pull request lables (see [Advanced Pull-Request Voting](#advanced-pull-request-voting)</li>
     * <li>Explicit activation via stage configuration</li>
     * </ul>
     */
    'whitesourceExecuteScan'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Only for `buildTool: docker` **and not executed in a Kubernetes environment**.
     * Defines the build options for the Docker build.
     */
    'containerBuildOptions',
    /**
     * Command to be excecuted in `dockerBuildImage`.
     * The command can be of Groovy template style and `config` as well as `env` can be consumed in the template.
     * Only active if dockerBuildImage and dockerBuildCommand are both set.
     */
    'dockerBuildCommand',
    /**
     * Name of the docker image that should be used for a custom build.
     * Only active if dockerBuildImage and dockerBuildCommand are both set.
     */
    'dockerBuildImage',
    /**
     * Kubernetes only:
     * Specifies a dedicated user home directory for the container which will be passed as value for environment variable `HOME`.
     */
    'dockerBuildWorkspace',
    /**
     * Environment variables to set in the container.
     */
    'dockerBuildEnvVars',
    /** Username/Token credentials maintained in Jenkins for access to github repository.
     *  GITHUB_USERNAME & GITHUB_TOKEN be available in env to be consumed from dockerCommand.
     */
    'gitHttpsCredentialsId',
    /** Command to be executed in `dockerImage`.
     * The command can be of Groovy template style and `config` as well as `env` can be consumed in the template.
     */
    'dockerCommand',
    /** @see dockerExecute */
    'dockerImage',
    /** @see dockerExecute */
    'dockerWorkspace',
    /** @see dockerExecute */
    'dockerEnvVars',
     /**
     * maven PR voting only:
     * Specifies a dedicated global settings xml file used for maven pull request voting.
     */
    'globalSettingsFile'
].plus(STAGE_STEP_KEYS))
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    def stageName = parameters.stageName?:env.STAGE_NAME

    //ToDo: Update stashing in defaults to be able to extend current stage
    piperStageWrapper (script: script, stageName: stageName, stageLocking: false) {

        durationMeasure(script: script, measurementName: 'voter_duration') {

            // load default & individual configuration
            Map config = ConfigurationHelper
                .loadStepDefaults(this)
                .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
                .mixinStageConfig(script.globalPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
                .mixin(parameters, PARAMETER_KEYS)
                .dependingOn('buildTool').mixin('dockerCommand')
                .dependingOn('buildTool').mixin('dockerImage')
                .dependingOn('buildTool').mixin('dockerWorkspace')
                .addIfEmpty('executeFortifyScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.executeFortifyScan)
                .addIfEmpty('executeCheckmarxScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.executeCheckmarxScan)
                .addIfEmpty('executePPMSComplianceCheck', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.executePPMSComplianceCheck)
                .addIfEmpty('sonarExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.sonarExecuteScan)
                .addIfEmpty('karmaExecuteTests', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.karmaExecuteTests)
                .addIfEmpty('mavenExecuteStaticCodeChecks', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.mavenExecuteStaticCodeChecks)
                .addIfEmpty('npmExecuteLint', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.npmExecuteLint)
                .addIfEmpty('npmExecuteScripts', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.npmExecuteScripts)
                .addIfEmpty('mavenExecuteIntegration', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.mavenExecuteIntegration)
                .addIfEmpty('whitesourceExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.whitesourceExecuteScan)
                .addIfEmpty('sapCheckPPMSCompliance', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.sapCheckPPMSCompliance)
                .addIfEmpty('executeOpenSourceDependencyScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.executeOpenSourceDependencyScan)
                .addIfEmpty('checkmarxExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.checkmarxExecuteScan)
                .addIfEmpty('checkmarxOneExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.checkmarxOneExecuteScan)
                .addIfEmpty('fortifyExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.fortifyExecuteScan)
                .addIfEmpty('codeqlExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.codeqlExecuteScan)
                .addIfEmpty('kanikoExecute', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.kanikoExecute)
                .addIfEmpty('golangBuild', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.golangBuild)
                .addIfEmpty('githubTokenCredentialsAvailable', script.commonPipelineEnvironment.getStepConfiguration('githubSetCommitStatus', stageName).githubTokenCredentialsId ? true : false)
                .use()

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

                    config.dockerImageNameAndTag = config.dockerImageNameAndTag ?: config.dockerImageName

                    // initalize status checks
                    if (config.checkmarxExecuteScan && config.githubTokenCredentialsAvailable) githubSetCommitStatus script: script, commitId: commonPipelineEnvironment.getGitCommitId(), context: 'piper/checkmarx', status: 'pending', targetUrl: "${env.BUILD_URL}display/redirect"
                    if (config.fortifyExecuteScan && config.githubTokenCredentialsAvailable) githubSetCommitStatus script: script, commitId: commonPipelineEnvironment.getGitCommitId(), context: 'piper/fortify', status: 'pending', targetUrl: "${env.BUILD_URL}display/redirect"

                    durationMeasure(script: script, measurementName: 'voter_duration') {
                        if (config.dockerBuildImage && config.dockerBuildCommand){
                            dockerExecute(
                                script: script,
                                dockerImage: config.dockerBuildImage,
                                dockerWorkspace: config.dockerBuildWorkspace,
                                dockerEnvVars: config.dockerBuildEnvVars
                            ) {
                                try {
                                    sh TemplatingUtils.render(
                                        config.dockerBuildCommand.toString(),
                                        [env: env, config: config]
                                    )
                                } catch (e) {
                                    error "[PR Voting] ERROR: The execution of the voting build failed, see the log for details."
                                }
                            }
                        } else {
                            if (config.buildTool == 'docker') {
                                if (config.dockerCommand) {
                                    try {
                                        sh TemplatingUtils.render(
                                            config.dockerCommand.toString(),
                                            [env: env, config: config]
                                        )
                                    } catch (e) {
                                        error "[PR Voting] ERROR: The execution of the voting build failed, see the log for details."
                                    }
                                } else if (Boolean.valueOf(env.ON_K8S) || config.kanikoExecute){
                                    kanikoExecute script: script
                                } else {
                                    config = new ConfigurationHelper(config)
                                        .withMandatoryProperty('dockerImageNameAndTag')
                                        .use()
                                    docker.build(config.dockerImageNameAndTag, "${config.containerBuildOptions} .")
                                }
                            } else if (config.buildTool == 'mta'){
                                mtaBuild(script: script)
                            } else {
                                config = new ConfigurationHelper(config)
                                    .withMandatoryProperty('dockerCommand')
                                    .withMandatoryProperty('dockerImage')
                                    .use()

                                dockerExecute(script: script, dockerImage: config.dockerImage, dockerWorkspace: config.dockerWorkspace, dockerEnvVars: config.dockerEnvVars) {
                                    try {
                                        // when buildTool is maven, append the global settings file
                                        if (config.buildTool == 'maven' && config.globalSettingsFile){
                                            def mvnOpts = ""
                                            def globalSettingsFile = config.globalSettingsFile
                                            if (globalSettingsFile.startsWith("http:") || globalSettingsFile.startsWith("https:")){
                                                log("loading global settings file ${globalSettingsFile}")
                                                globalSettingsFile = downloadSettingsFromUrl(script, globalSettingsFile, ".pipeline/mavenGlobalSettings.xml")
                                            }
                                            mvnOpts = "--global-settings ${globalSettingsFile}"
                                            config.dockerCommand = config.dockerCommand.replaceAll("mvn", "mvn ${mvnOpts}")
                                        } else if (config.buildTool == 'npm' && config.defaultNpmRegistry) {
                                            log("setting npm registry to ${config.defaultNpmRegistry}")
                                            sh "npm config set registry ${config.defaultNpmRegistry}"
                                        }
                                        withCredentialsWrapper(config) {
                                            sh TemplatingUtils.render(
                                                config.dockerCommand.toString(),
                                                [env: env, config: config]
                                            )
                                        }
                                    } catch (e) {
                                        error "[PR Voting] ERROR: The execution of the voting build failed, see the log for details." +e
                                    }
                                }
                            }
                        }

                        try {
                            //needs to run right after build, otherwise we may face "ERROR: Test reports were found but none of them are new"
                            testsPublishResults script: script

                            if (config.mavenExecuteStaticCodeChecks) {
                                mavenExecuteStaticCodeChecks script: script, globalSettingsFile: config.globalSettingsFile
                            }

                            if (config.npmExecuteLint) {
                                npmExecuteLint script: script
                            }

                            checksPublishResults script: script

                            if (config.mavenExecuteIntegration) {
                                runMavenIntegrationTests(script, config)
                            }

                            if (config.karmaExecuteTests) {
                                karmaExecuteTests script: script
                                testsPublishResults script: script
                            }

                            if (config.npmExecuteScripts) {
                                npmExecuteScripts script: script
                                testsPublishResults script: script
                            }

                            if (config.golangBuild) {
                                golangBuild script: script
                                testsPublishResults script: script
                            }
                        } finally {
                            if (config.sonarExecuteScan) {
                                sonarExecuteScan script: script
                            }
                        }

                        // begin: OLD SAST handling
                        // ToDo: Remove once switch to new steps is done
                        if (!config.checkmarxExecuteScan && config.executeCheckmarxScan) {
                            executeCheckmarxScan script: script, pullRequestName: env.BRANCH_NAME, incremental: false
                        }
                        if (!config.fortifyExecuteScan && config.executeFortifyScan) {
                            executeFortifyScan script: script, pullRequestName: env.BRANCH_NAME, reporting: false
                        }
                        // end: OLD SAST handling

                        // begin: NEW SAST handling
                        if (config.checkmarxExecuteScan) {
                            try {
                                checkmarxExecuteScan script: script, pullRequestName: env.BRANCH_NAME, incremental: false
                                if (config.githubTokenCredentialsAvailable) githubSetCommitStatus script: script, commitId: commonPipelineEnvironment.getGitCommitId(), context: 'piper/checkmarx', status: 'success', targetUrl: "${env.BUILD_URL}display/redirect"
                            } catch (err) {
                                if (config.githubTokenCredentialsAvailable) githubSetCommitStatus script: script, commitId: commonPipelineEnvironment.getGitCommitId(), context: 'piper/checkmarx', status: 'failure', targetUrl: "${env.BUILD_URL}display/redirect"
                                throw err
                            }
                        }
                        if (config.checkmarxOneExecuteScan) {
                            try {
                                // TODO: use orchestrator package
                                checkmarxOneExecuteScan script: script, pullRequestName: env.BRANCH_NAME, incremental: false
                                if (config.githubTokenCredentialsAvailable) githubSetCommitStatus script: script, commitId: commonPipelineEnvironment.getGitCommitId(), context: 'piper/checkmarxOne', status: 'success', targetUrl: "${env.BUILD_URL}display/redirect"
                            } catch (err) {
                                if (config.githubTokenCredentialsAvailable) githubSetCommitStatus script: script, commitId: commonPipelineEnvironment.getGitCommitId(), context: 'piper/checkmarxone', status: 'failure', targetUrl: "${env.BUILD_URL}display/redirect"
                                throw err
                            }
                        }
                        if (config.fortifyExecuteScan) {
                            try {
                                fortifyExecuteScan script: script, pullRequestName: env.BRANCH_NAME, reporting: false
                                if (config.githubTokenCredentialsAvailable) githubSetCommitStatus script: script, commitId: commonPipelineEnvironment.getGitCommitId(), context: 'piper/fortify', status: 'success', targetUrl: "${env.BUILD_URL}display/redirect"
                            } catch (err) {
                                if (config.githubTokenCredentialsAvailable) githubSetCommitStatus script: script, commitId: commonPipelineEnvironment.getGitCommitId(), context: 'piper/fortify', status: 'failure', targetUrl: "${env.BUILD_URL}display/redirect"
                                throw err
                            }
                        }
                        if (config.codeqlExecuteScan) {
                            codeqlExecuteScan script: script
                        }
                        // end: NEW SAST handling

                        if (config.executeOpenSourceDependencyScan) {
                            executeOpenSourceDependencyScan script: script, pullRequestName: env.BRANCH_NAME
                        }

                        if (config.executePPMSComplianceCheck || config.sapCheckPPMSCompliance || config.whitesourceExecuteScan) {
                            whitesourceExecuteScan script: script, productVersion: env.BRANCH_NAME
                        }

                        if (config.sapCheckPPMSCompliance) {
                            sapCheckPPMSCompliance script: script, pullRequestMode: true
                        } else if (config.executePPMSComplianceCheck) {
                            executePPMSComplianceCheck script: script, whitesourceProjectNames: script.commonPipelineEnvironment.getValue('whitesourceProjectNames'), pullRequestMode: true
                        }
                    }
                }
            }
        }
    }
}

private runMavenIntegrationTests(script, config){
    boolean publishResults = false
    try {
        writeTemporaryCredentials(script: script) {
            publishResults = true
            mavenExecuteIntegration script: script, globalSettingsFile: config.globalSettingsFile
        }
    }
    finally {
        if (publishResults) {
            testsPublishResults script: script
        }
    }
}

void withCredentialsWrapper( config, body ) {
    if (config.gitHttpsCredentialsId) {
        withCredentials([usernamePassword(
            credentialsId: config.gitHttpsCredentialsId,
            passwordVariable: 'GITHUB_TOKEN',
            usernameVariable: 'GITHUB_USERNAME')]) {
            body()
        }
    }
    else {
        body()
    }
}

void log(msg) {
    echo "[${STEP_NAME}] ${msg}"
}

String downloadSettingsFromUrl(script, String url, String targetFile = 'settings.xml') {
    if (script.fileExists(targetFile)) {
        log("Global settings file ${targetFile} already exists. Skipping download from ${url}")
        return targetFile
    }

    def response = script.httpRequest(url)
    script.writeFile(file: targetFile, text: response.content)
    return targetFile
}

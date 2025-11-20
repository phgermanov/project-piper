import com.sap.piper.internal.JenkinsUtils
import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Notify
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'sapPiperStageInit'

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Defines the build tool used.
     * @possibleValues `maven`, `npm`, `mta`, ...
     */
    'buildTool',
    /**
     * Activate a native build pull-request voting (e.g. maven, npm) only in combination with "Piper native build".
     * @possibleValues 'true', 'false'
     */
    'nativeVoting',
    /**
     * Defines the main branch for your pipeline. **Typically this is the `master` branch, which does not need to be set explicitly.** Only change this in exceptional cases. Supports regular expression through Groovy Match operator, e.g. `master|develop`.
     */
    'productiveBranch',
    /**
     * Defines if the "Security" and "IPScan and PPMS" stages run only scheduled according to the schedule defined in parameter 'nightlySchedule' in step setupPipelineEnvironment.
     * @possibleValues `true`, `false`
     */
    'pipelineOptimization',
    /**
     * Parameter in Beta mode.
     * To be set to true if env.json, bom xmls and build-settings.json are to be uploaded for a Pull Request.
     * @possibleValues `true`, `false`
     */
    'uploadCumulusFilesforPR',
    /**
     * Defines the library resource containing the stash settings to be performed before and after each stage. **Caution: changing the default will break the standard behavior of the pipeline - thus only relevant when including `Init` stage into custom pipelines!**
     */
    'stashSettings',
    /**
     * Defines the library resource containing stage/step initialization settings. Those define conditions when certain steps/stages will be activated. **Caution: changing the default will break the standard behavior of the pipeline - thus only relevant when including `Init` stage into custom pipelines!**
     */
    'stageConfigResource',
    /**
     * Whether verbose output should be produced.
     * @possibleValues `true`, `false`
     */
    'verbose',
    /**
     * Defines the library resource containing the legacy configuration definition.
     */
    'legacyConfigSettings',
    /*
     secretID, groupID to call systemTrust
    */
    'vaultAppRoleSecretTokenCredentialsId',
    'vaultBasePath',
    'vaultPipelineName'
]
@Field STAGE_STEP_KEYS = [
    /** Always active: Initializes the environment for the pipeline run and creates a common object used for passing information between pipeline stages and steps.<br />One important part is reading the pipeline configuration file (located in your source code repository in `.pipeline/config.yml`).*/
    'setupPipelineEnvironment',
    /** Always active: Initializes stage and step execution based on pre-defined conditions. Conditions are documented for each stage (see respective stage documentation) */
    'piperInitRunStageConfiguration',
    /** Always active: Automatic versioning of artifact is triggered. */
    'artifactPrepareVersion',
    /** Always active: Executes stashing before build execution. This makes sure that files from source code repository can be made available for the individual pipeline stages. */
    'pipelineStashFilesBeforeBuild',
    /** Always active: Publishes pipelineRunStart and pipelineTaskRunFinished (artifactPrepareVersion) events to GCP. */
    'gcpPublishEvent'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /**
     * Optional skip of checkout if checkout was done before this step already.
     * @possibleValues `true`, `false`
     */
    'skipCheckout',
    /**
     * Mandatory if you skip the checkout. Then you need to unstash your workspace to get the e.g. configuration.
     */
    'stashContent',
    /**
     * The map returned from a Jenkins git checkout. Used to set the git information in the
     * common pipeline environment.
     */
    'scmInfo'
])

void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    def stageName = parameters.stageName?:env.STAGE_NAME

    piperStageWrapper (script: script, stageName: stageName, stashContent: [], ordinal: 1, nodeLabel: parameters.nodeLabel) {
        def skipCheckout = parameters.skipCheckout
        def scmInfoParam = parameters.scmInfo
        if (skipCheckout != null && !(skipCheckout instanceof Boolean)) {
            error "[${STEP_NAME}] Parameter skipCheckout has to be of type boolean. Instead got '${skipCheckout.class.getName()}'"
        }
        if (skipCheckout && !scmInfoParam) {
            echo "WARN [${STEP_NAME}] Need an scmInfo retrieved from a checkout. " +
                "If you want to skip the checkout the scm info needs to be provided by you with parameter scmInfo, " +
                "for example as follows:\n" +
                "  def scmInfo = checkout scm\n" +
                "  sapPipelineStageInit script:this, skipCheckout: true, stashContent: ['stash-name'], scmInfo: scmInfo"
        }
        if(skipCheckout) {
            def stashContent = parameters.stashContent
            stashContent = utils.unstashAll(stashContent)
            if(stashContent.size() == 0) {
                error "[${STEP_NAME}] needs stashes if you skip checkout"
            }
        }
        if(!skipCheckout && !scmInfoParam) {
            deleteDir()
            scmInfoParam = checkout scm
        }
        setupPipelineEnvironment script: script, customDefaults: parameters.customDefaults, scmInfo: scmInfoParam
        try {
            gcpPublishEvent script: script, eventType: "sap.hyperspace.pipelineRunStarted", topic: "hyperspace-pipelinerun-started"
        } catch (e) {
            echo "no pipelineRunStarted event published to GCP"
        }

        script.commonPipelineEnvironment.setGitCommitId(utils.getGitCommitIdOrNull())

        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .addIfEmpty('sapCumulusUpload', script.commonPipelineEnvironment.getStepConfiguration('sapCumulusUpload', stageName)?.pipelineId ? true : false )
            .withMandatoryProperty('buildTool')
            .use()

        //perform stashing based on library resource piper-stash-settings.yml if not configured otherwise
        initStashConfiguration(script, config)

        echo "Getting System trust token"
        def token = null
        try {
            def apiURLFromDefaults = script.commonPipelineEnvironment.getValue("hooks")?.systemtrust?.serverURL ?: ''
            token = sapGetSystemTrustToken(apiURLFromDefaults, config.vaultAppRoleSecretTokenCredentialsId, config.vaultPipelineName, config.vaultBasePath)
        } catch (Exception e) {
            echo "Couldn't get system trust token, will proceed with configured credentials: ${e.message}"
        }

        withEnv([
            /*
            Additional logic "?: ''" is necessary to ensure the environment
            variable is set to an empty string if the value is null
            Without this, the environment variable would be set to the string "null",
            causing checks for an empty token in the Go application to fail.
            */
            "PIPER_systemTrustToken=${token ?: ''}",
        ]) {
            if (config.verbose) {
                echo "piper-lib configuration: ${script.globalPipelineEnvironment.configuration}"
                echo "piper-lib-os  configuration: ${script.commonPipelineEnvironment.configuration}"
            }

            if (!script.globalPipelineEnvironment.getFlag('piper-lib-os')) {
                error "[${STEP_NAME}] Library 'piper-lib-os' not available. Please configure it according to https://go.sap.corp/piper/lib/setupLibrary/"
            }

            if (config.legacyConfigSettings) {
                Map legacyConfigSettings = readYaml(text: libraryResource(config.legacyConfigSettings))
                checkForLegacyConfiguration(script: script, legacyConfigSettings: legacyConfigSettings)
            }

            piperInitRunStageConfiguration script: script, stageConfigResource: config.stageConfigResource

            //ToDo: get rid of globalPipelineEnvironment.configuration.runStage & globalPipelineEnvironment.configuration.runStep
            script.globalPipelineEnvironment.configuration.runStage = script.commonPipelineEnvironment.configuration.runStage
            script.globalPipelineEnvironment.configuration.runStep = script.commonPipelineEnvironment.configuration.runStep

            // Pull-request functionality
            // CHANGE_ID is set only for pull requests
            if (env.CHANGE_ID) {
                List prActions = []

                //get trigger action from comment like /piper action
                def jenkinsUtils = new JenkinsUtils()
                def commentTriggerAction = jenkinsUtils.getIssueCommentTriggerAction()

                if (commentTriggerAction != null) prActions.add(commentTriggerAction)

                try {
                    prActions.addAll(pullRequest.getLabels().asList())
                } catch (ex) {
                    Notify.warning(this, 'GitHub labels could not be retrieved from Pull Request, please make sure that credentials are maintained on multi-branch job.', STEP_NAME)
                }

                setPullRequestStageStepActivation(script, config, prActions)

                if (config.nativeVoting) {
                    // need to prevent creation of tags for a PR run
                    def versioningType = script.globalPipelineEnvironment.getStepConfiguration('artifactPrepareVersion', stageName).versioningType

                    if ("cloud".equals(versioningType)) {
                        versioningType = "cloud_noTag"
                    }

                    artifactPrepareVersion script: this, versioningType: versioningType
                    // for Hyperspace PR voting service (ACT) native voting enabled PRs , upload env.json
                    sapGenerateEnvironmentInfo script: script
                    if (config.sapCumulusUpload && config.uploadCumulusFilesforPR) {
                        sapCumulusUpload script: script, filePattern: 'env*.json', stepResultType: 'root'
                        // Upload env.json file for SLC-29 policy
                        sapCumulusUpload script: script, filePattern: 'env*.json', stepResultType: 'policy-evidence/SLC-29-PI'
                    }
                }
            }
            if (env.BRANCH_NAME ==~ config.productiveBranch) {
                if (parameters.script.globalPipelineEnvironment.configuration.runStep?.get('Post Actions')?.slackSendNotification)
                    slackSendNotification script: script, message: "STARTED: Job <${env.BUILD_URL}|${URLDecoder.decode(env.JOB_NAME, java.nio.charset.StandardCharsets.UTF_8.name())} ${env.BUILD_DISPLAY_NAME}>", color: 'WARNING'

                Map versioningParams = [script: script]
                if (script.commonPipelineEnvironment.getValue('scheduledRun') && config.pipelineOptimization && script.globalPipelineEnvironment.getStepConfiguration('artifactPrepareVersion', stageName).versioningType != 'library') {
                    versioningParams.versioningType = 'cloud_noTag'
                    echo "new versioning type: ${versioningParams.versioningType}"
                }
                artifactPrepareVersion versioningParams
                // update environment based on step results
                parameters.script.globalPipelineEnvironment.setGitCommitId(parameters.script.commonPipelineEnvironment.gitCommitId)
                parameters.script.globalPipelineEnvironment.setArtifactVersion(parameters.script.commonPipelineEnvironment.artifactVersion)
                // remember versionBeforeAutoVersioning for downloadArtifactsFromNexus
                parameters.script.globalPipelineEnvironment.versionBeforeAutoVersioning = parameters.script.commonPipelineEnvironment.originalArtifactVersion

                // generate env.json and upload to cumulus
                // ToDo: should sapGenerateEnvironmentInfo run on non productive branch ?
                sapGenerateEnvironmentInfo script: script
                if (config.sapCumulusUpload) {
                    sapCumulusUpload script: script, filePattern: 'env*.json', stepResultType: 'root'
                    // Upload env.json file for SLC-25 policy
                    sapCumulusUpload script: script, filePattern: 'env*.json', stepResultType: 'policy-evidence/SLC-25'
                    // Upload env.json file for SLC-29 policy in case of not optimized and scheduled mode
                    if (!script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled')) {
                        sapCumulusUpload script: script, filePattern: 'env*.json', stepResultType: 'policy-evidence/SLC-29-PI'
                    }
                }
                // indicates that artifactPrepareVersion has been executed
                // can only run after sapCumulusUpload has exported the pipelineRunKey
                try {
                    gcpPublishEvent script: script, eventType: "sap.hyperspace.pipelineTaskRunFinished", topic: "hyperspace-pipelinetaskrun-finished", additionalEventData: "{\"taskName\": \"artifactPrepareVersion\",\"outcome\": \"success\"}"
                } catch (e) {
                    echo "no pipelineTaskRunFinished event published to GCP"
                }
            }

            pipelineStashFilesBeforeBuild script: script
        }
    }
}

private void initStashConfiguration (script, config) {
    Map stashConfiguration = readYaml(text: libraryResource(config.stashSettings))
    echo "Stash config: stashConfiguration"
    script.commonPipelineEnvironment.configuration.stageStashes = stashConfiguration
}

private void setPullRequestStageStepActivation(script, config, List actions) {

    if (script.globalPipelineEnvironment.configuration.runStep == null)
        script.globalPipelineEnvironment.configuration.runStep = [:]
    if (script.globalPipelineEnvironment.configuration.runStep[config.pullRequestStageName] == null)
        script.globalPipelineEnvironment.configuration.runStep[config.pullRequestStageName] = [:]

    actions.each {action ->
        if (action.startsWith(config.labelPrefix))
            action = action.minus(config.labelPrefix)

        def stepName = config.stepMappings[action]

        if (stepName) {

            // ensure compatibility
            // ToDo: Remove once switch to new steps is done
            if (stepName == 'checkmarxExecuteScan' && !script.globalPipelineEnvironment.configuration?.steps?.checkmarxExecuteScan) stepName = 'executeCheckmarxScan'
            if (stepName == 'fortifyExecuteScan' && !script.globalPipelineEnvironment.configuration?.steps?.fortifyExecuteScan) stepName = 'executeFortifyScan'
            // end - ensure compatibility

            script.globalPipelineEnvironment.configuration.runStep."${config.pullRequestStageName}"."${stepName}" = true
        }
    }
}

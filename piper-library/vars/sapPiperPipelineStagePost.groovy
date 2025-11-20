import com.sap.piper.internal.ConfigurationHelper
import groovy.transform.Field
import hudson.AbortException
import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.Notify

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Defines the main branch for your pipeline. **Typically this is the `master` branch, which does not need to be set explicitly.** Only change this in exceptional cases. Supports regular expression through Groovy Match operator, e.g. `master|develop`.
     */
    'productiveBranch',
    /**
     * Parameter in Beta mode. To be set to true if cumulus storage needs to be only with commit id.
    */
    'useCommitIdForCumulus',
    /**
     * Skips creating a daily summary GitHub issue.
     * @possibleValues `true`, `false`
    */
    'skipCreateSummmaryResultIssue',
    /*
     secretID, groupID to call systemTrust
    */
    'vaultAppRoleSecretTokenCredentialsId',
    'vaultBasePath',
    'vaultPipelineName'
]
@Field STAGE_STEP_KEYS = [
    /** Writes information about the build to InfluxDB. */
    'influxWriteData',
    /** Sends notifications to the Slack channel about the build status. */
    'slackSendNotification',
    /** Sends notifications for current or previous build failures. */
    'mailSendNotification',
    /** Publishes piper library messages on the Jenkins job run as *Piper warnings* via the warnings-ng plugin. */
    'piperPublishWarnings',
    /** Materializes the Jenkins log file of the running build. */
    'jenkinsMaterializeLog',
    /** Reports pipeline run status internally in case of success/failure. */
    'sapReportPipelineStatus',
    /** Performs upload of files to Cumulus. */
    'sapCumulusUpload'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    def stageName = parameters.stageName?:env.STAGE_NAME
    def utils = parameters.juStabUtils ?: new Utils()
    // ease handling extension
    stageName = stageName.replace('Declarative: ', '')
    Map config = ConfigurationHelper
        .loadStepDefaults(this)
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('sapCumulusUpload', script.commonPipelineEnvironment.getStepConfiguration('sapCumulusUpload', stageName)?.pipelineId ? true : false )
        .addIfEmpty('skipCreateSummmaryResultIssue', true)
        .addIfEmpty('useCommitIdForCumulus', false)
        .use()

    // executing the external post stage is already done inside general purpose pipeline template
    // documentation is handled by the above defined but 'unused' STAGE_STEP_KEYS

    piperStageWrapper (script: script, stageName: stageName, stageLocking: false) {
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
                //perform pipeline status reporting
                buildResult = script.currentBuild?.result.toLowerCase()

                sapReportPipelineStatus script: script, pipelineResult: buildResult

                // check filename with ....
                if (config.sapCumulusUpload && env.BRANCH_NAME ==~ config.productiveBranch) {
                    jenkinsLog = "jenkins-log.txt"
                    pipelineLog = "pipelineLog.log"

                    if (config.useCommitIdForCumulus && script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled')) {
                        jenkinsLog = "jenkins-log-scheduled.txt"
                        pipelineLog = "pipelineLog-scheduled.log"
                    }

                    utils.unstash 'tests'
                    sapCumulusUpload script: parameters.script, filePattern: '**/cumulus-configuration.json', stepResultType: 'root'
                    def exists = fileExists pipelineLog
                    if (exists){
                        // examine if the file exists and upload renamed file to cumulus
                        sh "mv $pipelineLog $jenkinsLog"
                        sapCumulusUpload script: parameters.script, filePattern: 'jenkins-log*.txt', stepResultType: 'log'
                    } else {
                        // this is the fallback, in case teams did not maintain their credentials and
                        // sapReportPipelineStatus could not retrieve the logFile
                        Notify.warning(this, "Using jenkinsMaterializeLog as fallback - please maintain your Jenkins credentials in Vault.")
                        Notify.warning(this, "https://go.sap.corp/piper/steps/sapReportPipelineStatus/#jenkinsuser_1")
                        jenkinsMaterializeLog script: parameters.script, { name ->
                            echo "log file: " + name
                            sh "mv '$name' $jenkinsLog"
                            sapCumulusUpload script: parameters.script, filePattern: 'jenkins-log*.txt', stepResultType: 'log'
                        }
                    }
                }

                if (env.BRANCH_NAME ==~ config.productiveBranch) {
                    try {
                        gcpPublishEvent script: script, eventType: "sap.hyperspace.pipelineRunFinished", topic: "hyperspace-pipelinerun-finished"
                    } catch (e) {
                        echo "no pipelineRunFinished event published to GCP"
                    }

                    // generate and upload release status file. It must be done after pipelineTaskRunFinished event (gcpPublishEvent).
                    // See for details: https://jira.tools.sap/browse/HSPIPER-508
                    if (config.sapCumulusUpload) {
                        // sapGenerateEnvironmentInfo generates release status file
                        sapGenerateEnvironmentInfo script: this, generateFiles: ["releaseStatus"]

                        sapCumulusUpload script: parameters.script, filePattern: "release-status-*.json", stepResultType: '.status-log/release'
                        sapCumulusUpload script: parameters.script, filePattern: "**/lock-run.json", stepResultType: 'root'
                    }
                }

                // create GitHub issue for unstable optimized pipelines
                if (buildResult == 'unstable' && script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled') && !config.skipCreateSummmaryResultIssue) {
                    // question: do we want to fail the build or keep it unstable?
                    echo "Creating a daily scan summary GitHub issue"
                    pipelineCreateScanSummary script: this, pipelineLink: "${env.BUILD_URL}" // creates file scanSummary.md
                    try {
                        githubCreateIssue script: this, bodyFilePath: 'scanSummary.md', title: "${(new Date()).format("yyyyMMdd")}: Daily scan results - follow-up required"
                    } catch (err) {
                        throw new AbortException("failed to write scan summary for scheduled pipeline run, please make sure to provide credentials for step githubCreateIssue. Error was ${err}")
                    }

                }
            }
        }
    }
}

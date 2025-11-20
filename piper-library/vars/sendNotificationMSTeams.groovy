import com.sap.piper.internal.ConfigurationHelper
import groovy.transform.Field
import groovy.text.GStringTemplateEngine

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'sendNotificationMSTeams'
@Field Set GENERAL_CONFIG_KEYS = ['gitSshKeyCredentialsId']
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'buildResult',
    'color',
    'gitCommitId',
    'gitUrl',
    'message',
    'notifyCulprits',
    'projectName',
    'status',
    // Office365 Plugin specific
    'webhookUrl'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Send a build notification to Microsoft teams through the Jenkins Office365 Connector
 * https://jenkins.io/doc/pipeline/steps/Office-365-Connector/#-office365connectorsend-%20office365connectorsend
 * Prerequisite:
 * Plugin: Office 365 Connector is installed https://plugins.jenkins.io/Office-365-Connector
 * Code is copied from https://github.wdf.sap.corp/ContinuousDelivery/piper-library/blob/master/vars/sendNotificationSlack.groovy
 *
 * For suggestions please get in touch with jens.layer@sap.com
 *
 * Limitations:
 * - It is not possible to mention individuals. See https://microsoftteams.uservoice.com/forums/555103-public/suggestions/32313991-allow-webhooks-extensions-to-mention-individuals
 *
 * Further ideas:
 * - Get and include list of Failed Steps. This might require a sleep 2 see https://github.wdf.sap.corp/ContinuousDelivery/piper-library/blob/18758d30ba5b90fb7bef565aac8a29efcb12808d/src/com/sap/icd/jenkins/JenkinsUtils.groovy#L304
 * - Include Fortify results
 * - Switch to plain JSON POST to be able to create a custom teams entry. See https://docs.microsoft.com/en-us/outlook/actionable-messages/actionable-messages-via-connectors#posting-more-complex-cards
 *
 */
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def script = checkScript(this, parameters) ?: this
        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .addIfEmpty('projectName', script.currentBuild.fullProjectName)
            .addIfEmpty('displayName', script.currentBuild.displayName)
            .addIfEmpty('buildResult', script.currentBuild.result)
            .addIfEmpty('gitUrl', script.globalPipelineEnvironment.getGitSshUrl())
            .addIfEmpty('gitCommitId', script.globalPipelineEnvironment.getGitCommitId())
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        //this takes care that terminated builds due to milestone-locking do not cause an error
        if (script.globalPipelineEnvironment.getBuildResult() == 'ABORTED') return

        if(!config.webhookUrl){
            echo "[${STEP_NAME}] webhookUrl is not set. Skipping MS Teams notification"
            return
        }


        def buildStatus = script.currentBuild.result
        // resolve templates
        config.color = GStringTemplateEngine.newInstance().createTemplate(config.color).make([buildStatus: buildStatus]).toString()

        //Get culprits
        def culpritCommitters = ''
        if(config.notifyCulprits == true){
                culpritCommitters = getCulpritCommitters(config, script.currentBuild)
        }

        if (!config?.message){
            if (!buildStatus) {
                echo "[${STEP_NAME}] currentBuild.result is not set. Skipping MS Teams notification"
                return
            }
            config.message = GStringTemplateEngine.newInstance().createTemplate(config.defaultMessage).make([buildStatus: buildStatus, env: env, culpritCommitters: culpritCommitters]).toString()
        }


        Map options = [:]
        options.put('webhookUrl', config.webhookUrl)
        options.put('color', config.get('color'))
        options.put('message', config.get('message'))
        options.put('status', buildStatus)
        office365ConnectorSend(options)
    }
}

//TODO Extract then following methods into a new class GitUtils
//TODO Adopt the sendNotificationMail Script accordingly
def getNumberOfCommits(buildList){
    def numCommits = 0
    if(buildList != null)
        for(actBuild in buildList) {
            def changeLogSetList = actBuild.getChangeSets()
            if(changeLogSetList != null)
                for(changeLogSet in changeLogSetList)
                        numCommits += changeLogSet.size()
        }
    return numCommits
}

def getCulpritCommitters(config, currentBuild) {
    def recipients
    def buildList = []
    def build = currentBuild

    if (build != null) {
        // At least add the current build
        buildList.add(build)

        // Now collect FAILED or ABORTED ones
        build = build.getPreviousBuild()
        while (build != null) {
            if (build.getResult() != 'SUCCESS') {
                buildList.add(build)
            } else {
                break
            }
            build = build.getPreviousBuild()
        }
    }
    def numberOfCommits = getNumberOfCommits(buildList)
    if(config.wrapInNode){
        node(){
            try{
                recipients = getCulprits(config, env.BRANCH_NAME, numberOfCommits)
            }finally{
                deleteDir()
            }
        }
    }else{
        recipients = getCulprits(config, env.BRANCH_NAME, numberOfCommits)
        deleteDir()
    }
    echo "[${STEP_NAME}] last ${numberOfCommits} commits revealed following responsibles ${recipients}"
    return recipients
}

def getCulprits(config, branch, numberOfCommits) {

    if (branch?.startsWith('PR-')) {
        deleteDir()
        sshagent(
            credentials: [config.gitSshKeyCredentialsId],
            ignoreMissing: true
        ) {
            sh 'git init'
            def pullRequestID = branch.replaceAll('PR-', '')
            def localBranchName = "pr" + pullRequestID;
            sh "git fetch ${config.gitUrl} pull/${pullRequestID}/head:${localBranchName} > /dev/null 2>&1"
            sh "git checkout -f ${localBranchName} > /dev/null 2>&1"
        }
    } else {
        if (config.gitCommitId) {
            deleteDir()
            sshagent(
                credentials: [config.gitSshKeyCredentialsId],
                ignoreMissing: true
            ) {
                sh "git clone ${config.gitUrl} ."
                sh "git checkout ${config.gitCommitId} > /dev/null 2>&1"
            }
        } else {
            def retCode = sh(returnStatus: true, script: 'git log > /dev/null 2>&1')
            if (retCode != 0) {
                echo "[${STEP_NAME}] No git context available to retrieve culprits"
                return ''
            }
        }
    }

    def recipients = sh(returnStdout: true, script: "git log -${numberOfCommits} --pretty=format:'%ae %ce'")
    return getDistinctRecipients(recipients)
}

def getDistinctRecipients(recipients){
    def result
    def recipientAddresses = recipients.split()
    def knownAddresses = new HashSet<String>()
    if(recipientAddresses != null) {
        for(address in recipientAddresses) {
            address = address.trim()
            if(address
                && address.contains("@")
                && !address.startsWith("noreply")
                && !knownAddresses.contains(address)) {
                knownAddresses.add(address)
            }
        }
        result = knownAddresses.join("<br/>")
    }
    return result
}

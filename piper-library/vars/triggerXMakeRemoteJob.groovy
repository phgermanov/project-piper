import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Notify

import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

import hudson.model.Result;
import static groovy.json.JsonOutput.*


@Field String STEP_NAME = 'triggerXMakeRemoteJob'
@Field Set GENERAL_CONFIG_KEYS = [
    'xMakeServer',
    'xMakeJobName',
    'xMakeJobParameters',
    'xMakeNovaCredentialsId',
    'xMakeDevCredentialsId',
    'xMakeAnalyzeBuildErrors',
    'verbose'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

def call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, failOnError: true,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()

        // notify about deprecated step usage
        Notify.deprecatedStep(this, "sapXmakeExecuteBuild", "removed", script?.commonPipelineEnvironment)
        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        echo "Parameters: xMakeServer: ${config.xMakeServer}, xMakeJobName: ${config.xMakeJobName}, xMakeJobParameters: ${config.xMakeJobParameters}, xMakeNovaCredentialsId: ${config.xMakeNovaCredentialsId}, xMakeDevCredentialsId: ${config.xMakeDevCredentialsId}"

        def urlResult = getXmakeJobUrl(config.xMakeJobName);
        def handle = null
        def credentials = null
        if (urlResult.xmakeNovaJobUrl.length() > 0) {
            credentials = config.xMakeNovaCredentialsId
            handle = triggerRemoteJob(
                auth: CredentialsAuth(credentials: config.xMakeNovaCredentialsId),
                job: urlResult.xmakeNovaJobUrl,
                parameters: config.xMakeJobParameters,
                shouldNotFailBuild: true,
                httpGetReadTimeout: 20000
            )
        } else if (urlResult.xmakeDevJobUrl.length() > 0) {
            if(doCredentialsExist(config.xMakeDevCredentialsId)) {
                credentials = config.xMakeDevCredentialsId
                handle = triggerRemoteJob(
                    auth: CredentialsAuth(credentials: config.xMakeDevCredentialsId),
                    job: urlResult.xmakeDevJobUrl,
                    parameters: config.xMakeJobParameters,
                    shouldNotFailBuild: true
                )
            } else {
                handle = triggerRemoteJob(
                    remoteJenkinsName: config.xMakeServer, //Take authentication from global config for backward compatibility
                    job: config.xMakeJobName,
                    parameters: config.xMakeJobParameters,
                    shouldNotFailBuild: true
                )
            }
        } else {
            def troubleshootData =  "\n[Troubleshooting data] result content: " + urlResult.content + " result status: " + urlResult.status

            if (urlResult.status.equals(200)) {
                // xmakeDevJobUrl is empty even though server returns status 200
                Notify.error(this, "CFG-JJEN-000: No xMake job found! Check in xMakeJobName in piper config or Project Portal configuration." + troubleshootData)
            } else {
                Notify.error(this, "SDE-JJEN-000: xMake Host not found! Visit the 'Cloud Availability Center' about xMake otherwise open a 'ServiceNow' ticket." + troubleshootData)
            }
        }

        if (handle?.getBuildStatus().toString() == 'FINISHED') { // error management
            if (handle.getBuildResult()!= Result.SUCCESS && handle.getBuildResult()!= Result.UNSTABLE) {
                script.commonPipelineEnvironment.setValue("xmakeJobUrl", handle.getBuildUrl().toString())
                if(credentials && config.xMakeAnalyzeBuildErrors) {
                    echo 'Analyzing build failure :('
                    def errors=checkTriggeredBuild(config, credentials, handle.getBuildUrl().toString())
                    if(errors.size()) {
                        errors.each{ job, err ->
                            Notify.error(this, "Remote build finished with status ${handle.getBuildResult()}: ${err}")
                        }
                    } else {
                        Notify.error(this, "Remote build finished with status "+handle.getBuildResult()+"(no available analysis).")
                    }
                } else {
                    Notify.error(this, 'The remote job did not succeed. Further information could be obtained by setting xMakeAnalyzeBuildErrors=true setting and provide xmake-dev/nova credentials.')
                }
            }
        } else {
            script.commonPipelineEnvironment.setValue("xmakeJobUrl", handle.getBuildUrl().toString())
            Notify.error(this, 'The remote job did not succeed.')
        }

        return handle
    }
}

def getXmakeJobUrl(xMakeJobName) {
    // double check nova landscape
    def xMakeJobFinderPluginUrl = 'https://xmake-nova.wdf.sap.corp/job_finder/api/xml?input='
    def result = [
        xmakeDevJobUrl: '',
        xmakeNovaJobUrl: '',
        content: null
    ];

    try {
        def req = httpRequest url: xMakeJobFinderPluginUrl+xMakeJobName, validResponseCodes: '0:600'

        result.content = req?.content
        result.status = req?.status

        if(req.status.equals(200)) {
            def jobs = new XmlParser().parseText(req.content)
            for (job in jobs) {
                if (job.landscape.text().equals("xmake-dev")){
                    result.xmakeDevJobUrl = job.url.text()
                } else if (job.landscape.text().equals("xmake-nova")) {
                    result.xmakeNovaJobUrl = job.url.text()
                }
            }
        }
    } catch(e) {
        println 'Checking xmake-nova job URL failed: ' + e.getMessage()
    }

    return result;
}

def doCredentialsExist(key) {
    try {
        withCredentials([usernamePassword(credentialsId: key, passwordVariable: 'testPass', usernameVariable: 'testUser')]) {
            return true;
        }
    } catch(e) {
        return false;
    }
}

/**
 * search for an appropriate error message based on a regular exrepression in piper resources/error_rules.yml
 * example format of error_rules.yml:
 * - "error message example":
 *   - "regex example 1"
 *   - "regex example 2"
 * - "Please ensure that your repository is public or that the xmake user has access in case of a private repository":
 *   - "^fatal: Could not read from remote repository.$"
 *
 * @param content full console output to parse
 * @return the found eror message or null
 */
def findErrorRule(content) {
    if(!content) return null;

    def errorRules  = readYaml text: libraryResource('error_rules.yml')
    for(line in content.split('\n')) {
        for(messageEntry in errorRules) {
            for(entry in messageEntry) {
                for(regex in entry.value) {
                    if (line =~ regex) return entry.key
                }
            }
        }
    }
    return null
}

/**
 * Generic method to browse a Jenkins job and drill into sub-jobs to find failed ones. If a failed sub-job is found the parent one is not analyzed considering its failure as consequence
 * It searchs for a xmake archived error_dictionary.json file (see here for more details: https://github.wdf.sap.corp/xmake-ci/documentation/blob/master/documents/how2_update_error_rules.md).
 * If that file is not found, then it gets console outputs of failed jobs and check them against the resources/error_rules.yml piper-libray file to find the appropriate error message
 * If clear error message is found, then it display console outputs of the corresponding failed jobs.
 *
 * @param config ConfigurationHelper instance to get verbose value
 * @param credentials Jenkins credentials ID cotaining the appropriate user/password to browse the Jenkins Job.
 * @param link Jenkins corresponding job url from where to start the analysis with an ending slash like the value returned by PRT handle.getBuildResult(), for ex: http://localhost:8080/jenkins/job/empty-job/24/
 * @return a map of all found failed jobs with the corresponding Jenkins job name as key and the error message or console text as value.
 */
def checkTriggeredBuild(config, credentials, link) {
    def errorMessages=[:] // will contain all current job error messages and child ones

    def response=httpRequest authentication: credentials, url: link+'api/json?pretty=true&tree=actions[triggeredBuilds[number,url,result]],result,fullDisplayName', quiet: config.verbose?false:true
    if(!(response?.status==200)) {
        errorMessages.put(link+'api/json?pretty=true&tree=actions[triggeredBuilds[number,url,result]],result,fullDisplayName', 'HTTP Failure: '+response?.status)
        return errorMessages
    }

    def jsonJobApiAnswer=readJSON text: response.content
    if(jsonJobApiAnswer?.result=='SUCCESS') {
        return errorMessages
    }
    // in case of failure, analyze the current job and all its childs
    jsonJobApiAnswer?.actions.each { it->
        if(it._class=='hudson.plugins.parameterizedtrigger.BuildInfoExporterAction') {
            it.triggeredBuilds.each { build ->
                if(build._class=='hudson.model.FreeStyleBuild' && build.result!='SUCCESS') {
                    errorMessages << checkTriggeredBuild(config, credentials, build.url)
                }
            }
        }
    }
    if(errorMessages.size()) { // on error messages in childs doesn't analyze this current parent considering its failure as a consequence of childs
        return errorMessages
    }

    response=httpRequest authentication: credentials, validResponseCodes: '200,403,404', url: link+'artifact/.xmake/error_dictionary.json', quiet: config.verbose?false:true
    if(response?.status==200) { // no error dictionary found, analyze the console text
        def errorDictionary=readJSON text: response.content
        def buildResults = errorDictionary?.find{ it.key == "BUILDRESULTS" }?.value
        if(buildResults) {
            errorMessages.put(jsonJobApiAnswer.fullDisplayName, buildResults)
        } else {
            errorMessages.put(jsonJobApiAnswer.fullDisplayName, 'Cannot find any information in error error_dictionary.json: '+errorDictionary)
        }
    } else { // error_dictionary.json, analyze it
        response=httpRequest authentication: credentials, url: link+'consoleText', quiet: true
        def errorRule=response?.status==200?findErrorRule(response.content):null
        if(errorRule) {
            errorMessages.put(jsonJobApiAnswer.fullDisplayName, errorRule)
        } else {
            errorMessages.put(jsonJobApiAnswer.fullDisplayName, 'No associated error message found for this job, please check directly in the build log: '+link+'consoleText')
        }
    }

    return errorMessages
}

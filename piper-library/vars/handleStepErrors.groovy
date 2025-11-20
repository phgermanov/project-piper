import com.cloudbees.groovy.cps.NonCPS

import com.sap.piper.internal.JenkinsUtils
import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Notify

import groovy.text.GStringTemplateEngine
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'handleStepErrors'
@Field Set PARAMETER_KEYS = [
    'allowBuildFailure',
    'echoDetails',
    'echoParameters',
    'stepName',
    'stepNameDoc',
    'stepParameters'
]

def call(Map parameters = [:], body) {
    def script = checkScript(this, parameters) ?: this
    // notify about deprecated step usage
    Notify.deprecatedStep(this, "handlePipelineStepErrors", "removed", script?.commonPipelineEnvironment)
    // load default & individual configuration
    Map config = ConfigurationHelper
        .loadStepDefaults(this)
        .mixin(parameters, PARAMETER_KEYS)
        .withMandatoryProperty('stepParameters')
        .withMandatoryProperty('stepName')
        .addIfEmpty('stepNameDoc' , parameters.stepName)
        .use()

    def jenkinsUtils = config.stepParameters.jenkinsUtilsStub ?: new JenkinsUtils()

    if (currentBuild.result == 'FAILURE' && !config.allowBuildFailure)
        error "Previous step has set the build status to FAILURE"

    def echoMessage = ''

    try {
        if (config.echoDetails)
            echo "--- BEGIN LIBRARY STEP: ${config.stepName}.groovy ---"
        return body()
    } catch (Throwable err) {
        logError(config, err)
        def log = ''
        if ("${err}".contains('exit code')){
            log = jenkinsUtils.getLastLogLine()
        }

        Utils utils = new Utils()
        echoMessage += getMessage(
            config,
            err,
            jenkinsUtils.getLibrariesInfoWithPiperLatest().toString(),
            libraryResource('com.sap.piper.internal/templates/error.log')
        )
        throw err
    } finally {
        if (config.echoDetails) {
            echoMessage += "--- END LIBRARY STEP: ${config.stepName}.groovy ---"
            echo echoMessage
        }
    }
}

@NonCPS
private String getMessage(config, error, jenkinsLibraries, template){
    def engine = new GStringTemplateEngine()
    def binding = [
        config: config,
        jenkinsLibraries: jenkinsLibraries,
        error: error
    ]

    return engine.createTemplate(template).make(binding).toString()
}

private void logError(config, error, key = 0, severity = 'FAILURE'){
    def script = config?.stepParameters?.script

    if(script && script.globalPipelineEnvironment?.getInfluxCustomDataMapTags().build_error_message == null){
        script.globalPipelineEnvironment?.setInfluxCustomDataMapTagsProperty('pipeline_data', 'build_error_step', config.stepName)
        script.globalPipelineEnvironment?.setInfluxCustomDataMapTagsProperty('pipeline_data', 'build_error_stage', script.env?.STAGE_NAME)
        script.globalPipelineEnvironment?.setInfluxCustomDataMapProperty('pipeline_data', 'build_error_message', error.getMessage())
        // DEPRECATED
        if (error instanceof org.jenkinsci.plugins.workflow.steps.FlowInterruptedException) {
            key = 2
            severity = 'ABORT'
        }
        script.globalPipelineEnvironment?.setInfluxCustomDataProperty('build_result_key', key)
        script.globalPipelineEnvironment?.setInfluxCustomDataProperty('build_result', severity)
    }
}

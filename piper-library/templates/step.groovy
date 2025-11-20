import com.cloudbees.groovy.cps.NonCPS
import com.sap.icd.jenkins.Utils

import groovy.transform.Field

//ToDo: Change parameter stepName
@Field def STEP_NAME = 'stepName'
@Field Set STEP_CONFIG_KEYS = [
    'param1Name',
    'param2Name'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Name of library step
 *
 * @param script global script environment of the Jenkinsfile run
 * @param others document all parameters
 */
def call(Map parameters = [:], body) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def script = parameters.script ?: [globalPipelineEnvironment: globalPipelineEnvironment]
        def utils = parameters.juStabUtils ?: new Utils()
        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            //mandatory parameter - default cannot be null
            .withMandatoryProperty('param1Name')
            .use()

        //use parameter
        def param1 = config.get('param2Name')

    }
}

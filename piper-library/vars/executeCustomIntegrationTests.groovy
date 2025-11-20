import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Notify
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'executeCustomIntegrationTests'
@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'extensionIntegrationTestScript'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

def call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, failOnError: true,
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
            .mixin(parameters, PARAMETER_KEYS)
            .use()
        // check mandatory parameters
        new ConfigurationHelper(config)
            .withMandatoryProperty('extensionIntegrationTestScript')

        def utils = parameters.juStabUtils
        if (utils == null) {
            utils = new Utils()
        }

        utils.unstash 'pipelineConfigAndTests'
        def customIntegrationCall = load(config.extensionIntegrationTestScript)
        durationMeasure(script: script, measurementName: 'custom_integration_duration') {
            try {
                customIntegrationCall(script)
            } catch (err) {
                Notify.error(script, "Failure in custom integration test: ${err}, please see log for details.")
            }
        }
    }
}

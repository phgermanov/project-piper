import com.sap.piper.internal.JenkinsUtils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Notify

import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'restartableSteps'
@Field Set STEP_CONFIG_KEYS = [
    'sendMail',
    'timeoutInSeconds'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

def call(Map parameters = [:], body) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, failOnError: true,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def script = checkScript(this, parameters) ?: this
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()
        // notify about deprecated step usage
        Notify.deprecatedStep(this, "pipelineRestartSteps", "removed", script?.commonPipelineEnvironment)
        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        def restart = true
        while (restart) {
            try {
                body()
                restart = false
            } catch (Throwable err) {
                echo "ERROR occured: ${err}"
                if (config.sendMail)
                    if (jenkinsUtils.nodeAvailable()) {
                        sendNotificationMail script: script, buildResult: 'UNSTABLE'
                    } else {
                        node {
                            sendNotificationMail script: script, buildResult: 'UNSTABLE'
                        }
                    }

                try {
                    timeout(time: config.timeoutInSeconds, unit: 'SECONDS') {
                        input message: 'Do you want to restart?', ok: 'Restart'
                    }
                } catch(e) {
                    restart = false
                    throw err
                }
            }
        }
    }
}

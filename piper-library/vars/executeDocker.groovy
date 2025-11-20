import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Notify

import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'executeDocker'
@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'dindImage',
    'dockerEnvVars',
    'dockerImage',
    'dockerOptions',
    'dockerVolumeBind',
    'dockerWorkspace',
    'stashBackConfig',
    'skipStashBack',
    'stashContent'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

def call(Map parameters = [:], body) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        // notify about deprecated step usage
        Notify.deprecatedStep(this, "dockerExecute", "removed", script?.commonPipelineEnvironment)
        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        echo utils.printStepParameters(config, PARAMETER_KEYS)

        if (config.dockerImage) {
            if (env.jaas_owner || config.dindImage) {
                //TODO: handling of docker options not implemented yet, see: https://github.wdf.sap.corp/ContinuousDelivery/jenkins-pipeline-library/issues/51
                executeDockerOnKubernetes(
                    script: script,
                    dockerImage: config.dockerImage,
                    dockerWorkspace: config.dockerWorkspace,
                    dockerEnvVars: config.dockerEnvVars,
                    stashContent: config.stashContent,
                    stashBackConfig: config.stashBackConfig,
                    skipStashBack: config.skipStashBack,
                    dindImage: config.dindImage
                ) {
                    body()
                }
            } else {
                utils.unstashAll(config.stashContent)

                def image = docker.image(config.dockerImage)
                image.pull()
                image.inside(
                    getDockerOptions(config.dockerEnvVars, config.dockerVolumeBind, config.dockerOptions)
                ) {
                    body()
                }
            }
        } else {
            body()
        }
    }
}

/**
 * Returns a string with docker options containing
 * environment variables (if set).
 * Possible to extend with further options.
 * @param dockerEnvVars Map with environment variables
 */
private getDockerOptions(dockerEnvVars, dockerVolumeBind, dockerOptions) {
    def result = []
    if(dockerEnvVars) {
        for (String k: dockerEnvVars.keySet()) {
            result.push("--env ${k}=${dockerEnvVars[k]}")
        }
    }
    if(dockerVolumeBind) {
        for (String k: dockerVolumeBind.keySet()) {
            result.push("--volume ${k}:${dockerVolumeBind[k]}")
        }
    }
    if(dockerOptions) {
        result.push(dockerOptions)
    }
    return result.join(' ')
}

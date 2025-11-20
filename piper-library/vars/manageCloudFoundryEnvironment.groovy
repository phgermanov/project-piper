import com.sap.icd.jenkins.EnvironmentManagerRunner
import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import groovy.text.GStringTemplateEngine
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'manageCloudFoundryEnvironment'
@Field Set GENERAL_CONFIG_KEYS = [
    'cfCredentialsId',
    'gitHttpsCredentialsId',
    'mhCredentialsId',
    'command',
    'dockerImage',
    'dockerWorkspace',
    'environmentDescriptorFile',
    'stashContent'
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

        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        config = new ConfigurationHelper(config)
            .mixin([
                command: GStringTemplateEngine.newInstance().createTemplate(config.command).make([config: config]).toString()
            ])
            .use()

        utils.unstashAll(config.stashContent)

        executeEnvironmentManager(script, config)
	}
}

private def executeEnvironmentManager(script, Map config) {
    EnvironmentManagerRunner emr = new EnvironmentManagerRunner(script, config.cfCredentialsId, config.mhCredentialsId, config.gitHttpsCredentialsId)
    def result
    if (config.dockerImage.isEmpty()) {
        result = emr.execute(config.command)
    } else {
        script.echo "[${STEP_NAME}] Calling EMR.executeWithDocker()"
        result = emr.executeWithDocker(config.command, config.dockerImage, config.dockerWorkspace)
    }
    return handleResult(result, script)
}

private String inDockerOrDirect(String dockerImage) {
	return dockerImage?'inDocker':'direct'
}

private String emCommandOf(String command) {
	return command.split(' ')[0]
}

private int wordCountOf(String command) {
	return command.split(' ').size()
}

private def handleResult(def result,def script) {
	def SUCCESSFUL = 0
	if (result == SUCCESSFUL) {
		script.echo "[${STEP_NAME}] EMR returned successful with: $result"
	}
	else {
		script.echo "[${STEP_NAME}] EMR returned unsuccessful with: $result"
		error ("EnvironmentManager encountered a problem! Stopping!!! ($result)")
	}
	return result
}

import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'sapPiperStageIntegration'

@Field Set GENERAL_CONFIG_KEYS = [
    /*
     secretID, groupID to call systemTrust
    */
    'vaultAppRoleSecretTokenCredentialsId',
    'vaultBasePath',
    'vaultPipelineName'
]
@Field STAGE_STEP_KEYS = [
    /** DEPRECATED: please use standard [stage extension](../extensibility.md) functionality.*/
    'executeCustomIntegrationTests',
    /** Runs backend integration tests via maven in the module integration-tests/pom.xml. */
    'mavenExecuteIntegration',
    /** Performs upload of result files to cumulus. */
    'sapCumulusUpload'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    def stageName = parameters.stageName?:env.STAGE_NAME

    Map config = ConfigurationHelper
        .loadStepDefaults(this)
        .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.globalPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('executeCustomIntegrationTests', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.executeCustomIntegrationTests)
        .addIfEmpty('mavenExecuteIntegration', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.mavenExecuteIntegration)
        .addIfEmpty('sapCumulusUpload', script.commonPipelineEnvironment.getStepConfiguration('sapCumulusUpload', stageName)?.pipelineId ? true : false )
        .use()

    piperStageWrapper (script: script, stageName: stageName) {
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
                if (config.executeCustomIntegrationTests) {
                    try {
                        executeCustomIntegrationTests script: script
                    } catch (e) {
                        error "[${STEP_NAME}] Integration tests failed: ${e}"
                    }

                }

                if (config.mavenExecuteIntegration) {
                    utils.unstashAll(['source', 'buildResult'])
                    boolean publishResults = false
                    try {
                        writeTemporaryCredentials(script: script) {
                            publishResults = true
                            mavenExecuteIntegration script: script
                            if (config.sapCumulusUpload) {
                                Map cumulusConfig = script.commonPipelineEnvironment.getStepConfiguration("testsPublishResults", stageName)
                                sapCumulusUpload script: script, filePattern: '**/requirement.mapping', stepResultType: 'requirement-mapping'
                                sapCumulusUpload script: script, filePattern: cumulusConfig.junit.pattern, stepResultType: 'junit'
                                sapCumulusUpload script: script, filePattern: '**/integration-test/*.xml', stepResultType: 'integration-test'
                                sapCumulusUpload script: script, filePattern: '**/jacoco.xml', stepResultType: 'jacoco-coverage'
                            }
                        }
                    }
                    finally {
                        if (publishResults) {
                            testsPublishResults script: script
                        }
                    }
                }
            }
        }
    }
}

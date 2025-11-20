import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'sapPiperStageAdditionalUnitTests'

@Field Set GENERAL_CONFIG_KEYS = [
    /*
     secretID, groupID to call systemTrust
    */
    'vaultAppRoleSecretTokenCredentialsId',
    'vaultBasePath',
    'vaultPipelineName'
]
@Field STAGE_STEP_KEYS = [
    /** Executes karma tests which is for example suitable for OPA5 testing as well as QUnit testing of SAP UI5 apps.*/
    'karmaExecuteTests',
    /** Executes npm scripts to run frontend unit tests.
     * If custom names for the npm scripts are configured via the `runScripts` parameter the step npmExecuteScripts needs **explicit activation via stage configuration**. */
    'npmExecuteScripts',
    /** Publishes test results to Jenkins. It will automatically be active in cases tests are executed. */
    'testsPublishResults',
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
        .addIfEmpty('karmaExecuteTests', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.karmaExecuteTests)
        .addIfEmpty('npmExecuteScripts', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.npmExecuteScripts)
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

                if (config.karmaExecuteTests) {
                    durationMeasure(script: script, measurementName: 'opa_duration') {
                        karmaExecuteTests script: script
                        if(config.sapCumulusUpload) {
                            Map cumulusConfig = script.commonPipelineEnvironment.getStepConfiguration("testsPublishResults", stageName)
                            sapCumulusUpload script: script, filePattern: '**/TEST-*.xml', stepResultType: 'karma'
                            sapCumulusUpload script: script, filePattern: cumulusConfig.cobertura.pattern, stepResultType: 'karma'
                        }
                        testsPublishResults script: script
                    }
                }

                if (config.npmExecuteScripts) {
                    durationMeasure(script: script, measurementName: 'frontendUnitTests_duration') {
                        npmExecuteScripts script: script
                        if(config.sapCumulusUpload) {
                            Map cumulusConfig = script.commonPipelineEnvironment.getStepConfiguration("testsPublishResults", stageName)
                            sapCumulusUpload script: script, filePattern: cumulusConfig.junit.pattern, stepResultType: 'junit'
                            sapCumulusUpload script: script, filePattern: cumulusConfig.cobertura.pattern, stepResultType: 'cobertura-coverage'
                            sapCumulusUpload script: script, filePattern: cumulusConfig.cucumber.pattern, stepResultType: 'cucumber'
                        }
                        testsPublishResults script: script
                    }
                }
            }
        }
    }
}

import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'sapPiperStagePromote'

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Generates a `lock-run.json` to automatically [lock](https://wiki.wdf.sap.corp/wiki/x/GUtWiw) a pipeline run in Cumulus when the artifacts are promoted, in order to keep its results for future reference.
     * Enabled by default. Requires a valid [sapCumulusUpload](../steps/sapCumulusUpload.md) configuration.
     */
    'lockPipelineRun',
    /**
     * Activate a native build (e.g. maven, npm) in combination with SAP's staging service.
     * @possibleValues 'true', 'false'
     */
    'nativeBuild',
    /**
     * Defines the main branch for your pipeline. **Typically this is the `master` branch, which does not need to be set explicitly.** Only change this in exceptional cases. Supports regular expression through Groovy Match operator, e.g. `master|develop`.
     */
    'productiveBranch',
    /*
     secretID, groupID to call systemTrust
    */
    'vaultAppRoleSecretTokenCredentialsId',
    'vaultBasePath',
    'vaultPipelineName'
]

@Field STAGE_STEP_KEYS = [
    /** Performs xMake promote build using the option `buildType: 'xMakePromote'`. This cannot be switched off. */
    'executeBuild'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    def stageName = parameters.stageName?:env.STAGE_NAME

    Map config = ConfigurationHelper
        .loadStepDefaults(this)
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('sapCallStagingService', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.sapCallStagingService)
        .addIfEmpty('nativeBuild', false)
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
                durationMeasure(script: script, measurementName: 'promote_duration') {
                    if(config.nativeBuild){
                        if (config.sapCallStagingService) {
                            sapCallStagingService script: script, action: 'promote'
                        } else {
                            echo "sapCallStagingService step skipped (Promote stage)"
                        }
                    }else{
                        sapXmakeExecuteBuild script: script, buildType: 'xMakePromote'
                    }
                }

                if (env.BRANCH_NAME ==~ config.productiveBranch && config.lockPipelineRun) {
                    sh 'touch lock-run.json'
                }
            }
        }
    }
}

import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Notify
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'sapPiperStageIPScanPPMS'

@Field Set GENERAL_CONFIG_KEYS = [
    /*
     secretID, groupID to call systemTrust
    */
    'vaultAppRoleSecretTokenCredentialsId',
    'vaultBasePath',
    'vaultPipelineName'
    ]
@Field STAGE_STEP_KEYS = [
    /** Executes compliance check against SAP's PPMS system.*/
    'executePPMSComplianceCheck',
    /** Executes compliance check against SAP's PPMS system.*/
    'executePPMSWhitesourceComplianceCheck',
    /** Executes compliance check against SAP's PPMS system, using the latest version of the step.*/
    'sapCheckPPMSCompliance',
    /** Performs upload of result files to Cumulus. */
    'sapCumulusUpload',
    /**Checks if the PPMS object is already ECCN classified.  */
    'sapCheckECCNCompliance',
    /** (BETA) Executes an IP compliance check executing the WhiteSource scan with the Piper Open Source step. */
    'whitesourceExecuteScan',
    /** Performs BlackDuck Detect scanning to identify IP compliance issues. */
    'detectExecuteScan',
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS.plus('newOSS'))
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    def stageName = parameters.stageName?:env.STAGE_NAME

    def unstableMessage = "optimized pipeline run - continue although error occurred"

    Map config = ConfigurationHelper
        .loadStepDefaults(this)
        .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.globalPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('executePPMSComplianceCheck', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.executePPMSComplianceCheck)
        .addIfEmpty('executePPMSWhitesourceComplianceCheck', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.executePPMSWhitesourceComplianceCheck)
        .addIfEmpty('sapCumulusUpload', script.commonPipelineEnvironment.getStepConfiguration('sapCumulusUpload', stageName)?.pipelineId ? true : false )
        .addIfEmpty('sapCheckPPMSCompliance', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.sapCheckPPMSCompliance)
        .addIfEmpty('sapCheckECCNCompliance', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.sapCheckECCNCompliance)
        .addIfEmpty('whitesourceExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.whitesourceExecuteScan)
        .addIfEmpty('detectExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.detectExecuteScan)
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
                def ipStepsExecuted = script.commonPipelineEnvironment.getValue('ipStepsExecuted') ?: []

                echo "[${STEP_NAME}] ipStepsExecuted: ${ipStepsExecuted}"

                // new behavior (usage of new steps) for optimized pipelines as a first step
                if (script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled') || config.newOSS) {
                    if (script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled'))
                        echo "[${STEP_NAME}] running in optimized mode"

                    // only allow one IP scanning tool
                    if (config.whitesourceExecuteScan && config.detectExecuteScan && ipStepsExecuted.size() == 2) {
                        Notify.error(this, "two scanners active for IP scanning but only one allowed, choose between WhiteSource and BlackDuck Detect as input for the PPMS scan")
                    }

                    // only run IP scanning if not yet done in Security stage - WhiteSource takes preference if both configured
                    if (ipStepsExecuted.size() == 0 || ipStepsExecuted.size() == 2) {
                        echo "[${STEP_NAME}] no successful IP scan from current pipeline found."
                        utils.unstableIfCondition(this, script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled'), unstableMessage) {
                            if (config.whitesourceExecuteScan) {
                                try{
                                    whitesourceExecuteScan script: script
                                }finally{
                                    if(config.sapCumulusUpload) {
                                        sapCumulusUpload script: script, filePattern: 'whitesource/**, whitesource/*risk-report.pdf', stepResultType: 'whitesource-ip'
                                        if (ipStepsExecuted.size() == 0) {
                                            sapCumulusUpload script: script, filePattern: 'whitesource/**, whitesource/*risk-report.pdf', stepResultType: 'whitesource-security'
                                        }
                                    }
                                }
                            } else if (config.detectExecuteScan) {
                                detectExecuteScan script: this
                            }
                        }
                    } else {
                        echo("[${STEP_NAME}] Info: Re-using successful ${ipStepsExecuted[0]} scan from Security stage for PPMS check")
                    }

                    if (config.sapCheckPPMSCompliance) {
                        utils.unstableIfCondition(this, script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled'), unstableMessage) {
                            try{
                                durationMeasure(script: script, measurementName: 'ppmscheck_duration') {
                                    sapCheckPPMSCompliance script: script
                                }
                            }finally{
                                if(config.sapCumulusUpload) {
                                    sapCumulusUpload script: script, filePattern: '**/piper_whitesource_ppms_report.*',  stepResultType: 'whitesource-ip'
                                    sapCumulusUpload script: script, filePattern: '**/piper_blackduck_ppms_report.*',  stepResultType: 'blackduck-ip'
                                }
                            }
                        }
                    }

                // DEPRECATED behavior still in place
                } else {
                    echo "[IPScan & PPMS] running in NON-optimized mode"
                    if (config.whitesourceExecuteScan) {
                        try{
                            durationMeasure(script: script, measurementName: 'whitesource_duration') {
                                whitesourceExecuteScan script: script
                            }
                        }finally{
                            if(config.sapCumulusUpload) {
                                sapCumulusUpload script: script, filePattern: 'whitesource/**, whitesource/*risk-report.pdf', stepResultType: 'whitesource-ip'
                            }
                        }
                    }

                    if (config.executePPMSComplianceCheck || config.executePPMSWhitesourceComplianceCheck || config.sapCheckPPMSCompliance) {
                        try{
                            durationMeasure(script: script, measurementName: 'ppmscheck_duration') {
                                if (!config.sapCheckPPMSCompliance) {
                                    executePPMSComplianceCheck script: script
                                } else {
                                    sapCheckPPMSCompliance script: script
                                }
                            }
                        }finally{
                            if(config.sapCumulusUpload) {
                                sapCumulusUpload script: script, filePattern: '**/piper_whitesource_ppms_report.*',  stepResultType: 'whitesource-ip'
                                sapCumulusUpload script: script, filePattern: '**/piper_whitesource_ppms_report.*',  stepResultType: 'whitesource-security'
                                sapCumulusUpload script: script, filePattern: '**/piper_blackduck_ppms_report.*',  stepResultType: 'blackduck-ip'
                                sapCumulusUpload script: script, filePattern: '**/piper_blackduck_ppms_report.*',  stepResultType: 'blackduck-security'
                                sapCumulusUpload script: script, filePattern: '**/blackduck-ip.json', stepResultType: 'policy-evidence/PSL-1'
                                sapCumulusUpload script: script, filePattern: '**/whitesource-ip.json', stepResultType: 'policy-evidence/PSL-1'
                            }
                        }
                    }
                }

                if (config.sapCheckECCNCompliance) {
                    durationMeasure(script: script, measurementName: 'eccncheck_duration') {
                        sapCheckECCNCompliance script: script
                    }
                }
            }
        }
    }
}

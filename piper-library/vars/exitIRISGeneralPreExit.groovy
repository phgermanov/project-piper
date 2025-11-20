import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Deprecate
import groovy.json.JsonBuilder
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'exitIRISGeneralPreExit'
@Field Set GENERAL_CONFIG_KEYS = [
    'callingScript',
    'dockerImage',
    'dockerWorkspace',
    'oDataBusinessType',
    'oDataFreeStyle',
    'oDataLifecycleStatus',
    'oDataTenantRole',
    'odataSystemRole',
    'run',
    'spc',
    'spcCredentialsId',
    'stashContent',
    'verbose'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

//////////////////////////////////////////////////////////////////////
//                                                                  //
//   IRIS - Immortal Repository for Integrated operations Scripts   //
//                                                                  //
//////////////////////////////////////////////////////////////////////

def call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        // handle deprecated parameters
        Deprecate.parameter(this, parameters, 'exitIRISdebugMessages', 'verbose')
        Deprecate.parameter(this, parameters, 'exitIRISoDataBusinessType', 'oDataBusinessType')
        Deprecate.parameter(this, parameters, 'exitIRISoDataFreeStyle', 'oDataFreeStyle')
        Deprecate.parameter(this, parameters, 'exitIRISoDataLifecycleStatus', 'oDataLifecycleStatus')
        Deprecate.parameter(this, parameters, 'exitIRISoDataTenantRole', 'oDataTenantRole')
        Deprecate.parameter(this, parameters, 'exitIRISsystemRoleCode', 'odataSystemRole')
        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('spcCredentialsId')
            .withMandatoryProperty('odataSystemRole')
            .use()

        config = new ConfigurationHelper(config)
            .mixin([
                run: config.run instanceof Boolean ? config.run : config.run.toBoolean(),
                verbose: config.verbose instanceof Boolean ? config.verbose : config.verbose.toBoolean()
            ])
            .use()

        if (config.run) {
            // Read and Secure User Credentials
            withCredentials([usernamePassword(
                credentialsId: config.spcCredentialsId,
                passwordVariable: 'password',
                usernameVariable: 'username'
            )]) {

                dockerExecute(script: script, dockerImage: config.dockerImage, dockerWorkspace: config.dockerWorkspace, stashContent: config.stashContent) {

                    // Build JSON with Input values
                    def inputValue = [
                        SystemRoleCode: config.odataSystemRole,
                        oDataLifecycleStatus: config.oDataLifecycleStatus,
                        oDataTenantRole: config.oDataTenantRole,
                        oDataBusinessType: config.oDataBusinessType,
                        oDataFreeStyle: config.oDataFreeStyle,
                        SPC: config.spc,
                        SPC_Username: username,
                        SPC_Password: password,
                        DebugMessages: config.verbose.toString()
                    ]
                    def inputValueJSON = new JsonBuilder(inputValue).toString().bytes.encodeBase64().toString()

                    sh "/usr/bin/groovy /home/root/scripts/dispatchIRIS.groovy ${config.callingScript} ${inputValueJSON}"

                }
            }

            if (config.verbose) {
                sh 'cat exitIRISTenants.json'
            }
        } else {
            echo 'Don\'t execute, based on configuration setting.'
        }
    }
}

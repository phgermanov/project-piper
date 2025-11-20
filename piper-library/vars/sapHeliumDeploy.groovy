import com.sap.icd.jenkins.Utils
import com.sap.piper.GenerateDocumentation
import com.sap.piper.internal.ConfigurationHelper
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'sapHeliumDeploy'
@Field Set GENERAL_CONFIG_KEYS = [
    /** CredentialsId you need to define in Jenkins */
    'heliumCredentialsId'
]

@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /** Helium account */
    'account',
    /** Application name */
    'application',
    /** Host URL of your deploy target */
    'host'
])

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Helium Deploy Step
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,
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
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName ?: env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(artifactVersion: script.globalPipelineEnvironment.getArtifactVersion())
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('heliumCredentialsId')
            .withMandatoryProperty('account')
            .withMandatoryProperty('application')
            .withMandatoryProperty('host')
            .use()

        withCredentials([usernamePassword(credentialsId: config.heliumCredentialsId,
            passwordVariable: 'password', usernameVariable: 'username')]) {
            mavenExecute(script: this,
                         goals: 'process-sources',
                         flags: '--batch-mode --update-snapshots --activate-profiles cloud-deployment',
                         defines: "-Dtarget=devsystem -Dcloud.application=${config.application} -Dcloud.username=${username} -Dcloud.password=${password} -Dcloud.account=${config.account} -Dcloud.landscape=${config.host}"
            )
        }
    }
}

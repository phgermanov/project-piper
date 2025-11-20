import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.BashUtils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Deprecate
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'manageUaaService'
@Field Set GENERAL_CONFIG_KEYS = [
    'cfApiEndpoint',
    'cfServiceInstance',
    'cfCredentialsId',
    'cfOrg',
    'cfServicePlan',
    'cfSpace',
    'cfSubaccountId',
    'dockerImage',
    'dockerWorkspace',
    'stashContent',
    'xsSecurityFile'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

def call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        // handle deprecated parameters
        Deprecate.parameter(this, parameters, 'instanceName', 'cfServiceInstance')
        Deprecate.parameter(this, parameters, 'servicePlan', 'cfServicePlan')
        Deprecate.parameter(this, parameters, 'subaccountId', 'cfSubaccountId')

        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('cfServiceInstance')
            .withMandatoryProperty('cfCredentialsId')
            .withMandatoryProperty('cfOrg')
            .withMandatoryProperty('cfSpace')
            .use()

        config = new ConfigurationHelper(config)
            .mixin([
                formattedSpaceName: config.cfSpace.toLowerCase().replaceAll("_", "-"),
                stashContent: utils.unstashAll(config.stashContent)
            ])
        .use()

        withCredentials([usernamePassword(
                credentialsId: config.cfCredentialsId,
                passwordVariable: 'cfPassword',
                usernameVariable: 'cfUser'
        )]) {
            dockerExecute(script: script, dockerImage: config.dockerImage, dockerWorkspace: config.dockerWorkspace, stashContent: config.stashContent) {
                def fileNameForModifiedSecurityJson = "xs-security-${config.formattedSpaceName}.json"

                def securityJson = readJSON file: config.xsSecurityFile
                echo "[${STEP_NAME}] Content of file ${config.xsSecurityFile}: " + securityJson.toString()
                securityJson['xsappname'] = securityJson['xsappname'] + "-" + config.formattedSpaceName
                echo "[${STEP_NAME}] Updated content of file ${config.xsSecurityFile}: " + securityJson.toString()

                echo "[${STEP_NAME}] Replace every occurence of \${space} with actual space name"
                def fileAsString = securityJson.toString()
                fileAsString = fileAsString.replaceAll(/\$\{space\}/, config.formattedSpaceName)
                if (config.cfSubaccountId) {
                    echo "[${STEP_NAME}] Replace every occurence of ${config.cfSubaccountId} with the id of the subaccount"
                    fileAsString = fileAsString.replaceAll(/\$\{subaccountId\}/, config.cfSubaccountId)
                }

                echo "[${STEP_NAME}] Final content of ${config.xsSecurityFile}: " + fileAsString
                writeFile file: fileNameForModifiedSecurityJson, text: fileAsString

                sh "cf login -u ${BashUtils.escape(cfUser)} -p ${BashUtils.escape(cfPassword)} -a ${config.cfApiEndpoint} -o '${config.cfOrg}' -s '${config.cfSpace}'"
                def retCode = sh script: "cf service ${config.cfServiceInstance}", returnStatus: true
                if (retCode == 0) {
                    // uaa service already present -> update
                    def retVal = sh script: "cf update-service ${config.cfServiceInstance} -c ${fileNameForModifiedSecurityJson}", returnStatus: true
                    if (!config.cfServicePlan.equals("broker") && retVal != 0) {
                        // (update of service plan broker is currently not working, cf. https://jtrack.wdf.sap.corp/browse/XSBUG-1973)
                        error 'Update of uaa service instance failed.'
                    }
                } else {
                    sh "cf create-service xsuaa ${config.cfServicePlan} ${config.cfServiceInstance} -c ${fileNameForModifiedSecurityJson}"
                }
            }
        }
    }
}

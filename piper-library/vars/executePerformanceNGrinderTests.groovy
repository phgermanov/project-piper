import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.BashUtils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Deprecate
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'executePerformanceNGrinderTests'
@Field Set GENERAL_CONFIG_KEYS = [
    'cfMetrics',
    'cfApiEndpoint',
    'cfAppGuid',
    'cfCredentialsId',
    'cfOrg',
    'cfSpace',
    'dockerImage',
    'dockerWorkspace',
    'grafanaCredentialsId',
    'influxCredentialsId',
    'ngrinderCredentialsId',
    'stashContent',
    'isNewNgrinderVersion'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        // handle deprecated parameters
        Deprecate.parameter(this, parameters, 'influxcredentialsId', 'influxCredentialsId')
        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('cfMetrics')
            .withMandatoryPropertyUponCondition('cfAppGuid',{ c -> return c.get('cfMetrics') })
            .withMandatoryPropertyUponCondition('cfCredentialsId',{ c -> return c.get('cfMetrics') })
            .withMandatoryPropertyUponCondition('cfOrg',{ c -> return c.get('cfMetrics') })
            .withMandatoryPropertyUponCondition('cfSpace',{ c -> return c.get('cfMetrics') })
            .withMandatoryProperty('grafanaCredentialsId')
            .withMandatoryProperty('influxCredentialsId')
            .withMandatoryProperty('ngrinderCredentialsId')
            .use()

        config = new ConfigurationHelper(config)
            .mixin([
                stashContent: utils.unstashAll(config.stashContent),
            ])
            .use()


        echo "[${STEP_NAME}] Tools credential ID's: cf Credentials ID=${config.cfCredentialsId}, influx Credentials ID=${config.influxCredentialsId}, grafana Credentials ID=${config.grafanaCredentialsId}, nGrinder Credentials ID=${config.ngrinderCredentialsId}"
        echo "[${STEP_NAME}] CF Parameters: CF Org=${config.cfOrg}, CF Api Endpoint=${config.cfApiEndpoint}, CF Space=${config.cfSpace}, CF Metrics=${config.cfMetrics} isNewNgrinderVersion=${config.isNewNgrinderVersion}"

        def creds = [
                usernamePassword(credentialsId: config.influxCredentialsId,passwordVariable: 'influx_password',usernameVariable: 'influx_username'),
                usernamePassword(credentialsId: config.grafanaCredentialsId,passwordVariable: 'grafana_password',usernameVariable: 'grafana_username'),
                usernamePassword(credentialsId: config.ngrinderCredentialsId,passwordVariable: 'ngrinder_password',usernameVariable: 'ngrinder_username')
        ]

         if (config.cfMetrics) {
             creds.push(usernamePassword(credentialsId: config.cfCredentialsId,passwordVariable: 'cf_password',usernameVariable: 'cf_username'))
        }

        withCredentials(creds) {
            dockerExecute(script: script, dockerImage: config.dockerImage, dockerWorkspace: config.dockerWorkspace, stashContent: config.stashContent) {
                if (config.cfMetrics) {
                    sh "cf login -u ${BashUtils.escape(cf_username)} -p ${BashUtils.escape(cf_password)} -a ${config.cfApiEndpoint} -o '${config.cfOrg}' -s '${config.cfSpace}'"
                }
                sh " /opt/bin/ci-perfscript.sh ${BashUtils.escape(grafana_password)} ${BashUtils.escape(ngrinder_password)} ${config.cfAppGuid} ${BashUtils.escape(grafana_username)} ${BashUtils.escape(ngrinder_username)} ${BUILD_NUMBER} ${BashUtils.escape(influx_username)} ${BashUtils.escape(influx_password)} ${config.cfMetrics} ${config.isNewNgrinderVersion}"
                if (config.cfMetrics) {
                    sh "cf logout"
                }
            }
        }

    }
}

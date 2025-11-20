import com.sap.icd.jenkins.Utils
import com.sap.piper.GenerateDocumentation
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.integration.Daster
import groovy.transform.Field
import hudson.AbortException

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'executeDasterScan'
@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Whether verbose output is generated to the log which is specifically helpful in error situations
     * @possibleValues `true`, `false`
     **/
    'verbose'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /** Number of retries to be attempted in case of HTTP connection instability */
    'maxRetries',
    /** ID referencing a user/pwd credentials to fetch an oAuth token for FioriDAST service testing, please encode client_id as username and client_secret as password */
    'oAuthCredentialsId',
    /** The grant type to use for fetching the token */
    'oAuthGrantType',
    /** The source used to fetch the token */
    'oAuthSource',
    /** The URL to the XSUAA used for fetching the token */
    'oAuthServiceUrl',
    /**
     * Whether to use the step in a synchronous or asynchronous mode.
     * Setting `synchronous: false` will reduce the step to only trigger the DASTer scan without polling for any results.
     * @possibleValues `true`, `false`
     */
    'synchronous',
    /**
     * The type of DASTer scan to trigger which actually corresponds to the API endpoints i.e. `'basicScan'`
     * @possibleValues `'basicScan'`, `'oDataScan'`, `'swaggerScan'`, `'fioriDASTScan'`, `'aemscan'`, `'oDataFuzzer'`, `'burpscan'`
     **/
    'scanType',
    /** The settings configuration object as required by DASTer with a tiny **deviation**: for any credentials do not provide the original parameter but a parameter named like it but having the suffix `CredentialsId` i.e. `userCredentialsCredentialsId` with a value pointing to a respective credential in the Jenkins Credentials Store */
    'settings',
    /**
     * Whether to finally delete the scan or not only supported for `fioriDASTScan`
     * @possibleValues `true`, `false`
     **/
    'deleteScan',
    /** The URL to DASTer */
    'serviceUrl',
    /** The thresholds used to fail the build. Negative values can be used to turn the threshold off i.e. `thresholds: [ fail: [ high: -1] ] would deactivate the default to not tolerate any high findings */
    'thresholds'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * The name DASTer is derived from **D**ynamic **A**pplication **S**ecurity **T**esting. As the name implies, the tool targets to provide black-box security testing capabilities for your solutions
 * in an automated fashion.
 *
 * DASTer itself ships with a [Swagger based frontend](https://daster.tools.sap/api-spec/viewer/) and a [Web UI](https://app.daster.tools.sap/ui5/) to generate tokens
 * required to record your consent and to authenticate. Please see the [documentation](https://github.wdf.sap.corp/pages/Security-Testing/doc/daster/) for
 * background information about the tool, its usage scenarios and channels to report problems.
 */
@GenerateDocumentation
void call(Map parameters = [:], Closure body = {}) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()

        script.globalPipelineEnvironment.setInfluxStepData('daster', false)

        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName ?: env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('settings')
            .use()

        // check mandatory parameters
        if (!config.settings.dasterTokenCredentialsId)
            throw new AbortException("ERROR - NO VALUE AVAILABLE FOR 'settings.dasterTokenCredentialsId'")

        // DASTer scan
        echo "[${STEP_NAME}] Running scan of type ${config.scanType} with settings ${config.settings}"

        if (config.oAuthCredentialsId && config.oAuthServiceUrl) {
            config = fetchOAuthToken(utils, config)
        }

        if (config.settings.userCredentialsCredentialsId) {
            runScanWithUserCredentials(parameters, utils, config, body)
        }
        else if (config.settings.targetAuthCredentialsId) {
            runScanWithTargetAuth(parameters, utils, config, body)
        } else {
            runScan(parameters, utils, config, body)
        }

        script.globalPipelineEnvironment.setInfluxStepData('daster', true)
    }
}

def runScanWithUserCredentials(parameters, utils, config, body) {
     withCredentials([string(
        credentialsId: config.settings.userCredentialsCredentialsId,
        variable: 'credential'
    )]) {
        def extendedConfig = [:].plus(config)
        extendedConfig.settings.remove('userCredentialsCredentialsId')
        extendedConfig.settings.userCredentials = credential
        runScan(parameters, utils, extendedConfig, body)
    }
}

def runScanWithTargetAuth(parameters, utils, config, body) {
    withCredentials([[$class: 'UsernamePasswordMultiBinding', credentialsId: config.settings.targetAuthCredentialsId, passwordVariable: 'password', usernameVariable: 'user']]) {
        def extendedConfig = [:].plus(config)
        extendedConfig.settings.remove('targetAuthCredentialsId')
        extendedConfig.settings.targetAuth = [user: user, password: password]
        runScan(parameters, utils, extendedConfig, body)
    }
}

def fetchOAuthToken(utils, config) {
    withCredentials([[$class: 'UsernamePasswordMultiBinding', credentialsId: config.oAuthCredentialsId, passwordVariable: 'clientSecret', usernameVariable: 'clientId']]) {
        def params = [
            url                    : "${config.oAuthServiceUrl}?grant_type=${URLEncoder.encode(config.oAuthGrantType, 'UTF-8')}&scope=${URLEncoder.encode(config.oAuthSource, 'UTF-8')}&client_id=${URLEncoder.encode(clientId, 'UTF-8')}&client_secret=${URLEncoder.encode(clientSecret, 'UTF-8')}",
            httpMode               : 'POST',
            acceptType             : 'APPLICATION_JSON',
            contentType            : 'APPLICATION_FORM',
            quiet                  : !config.verbose,
            consoleLogResponseBody : config.verbose
        ]
        def response = httpRequest(params)
        def responseJson = utils.parseJsonSerializable(response.content)
        def extendedConfig = [:].plus(config)
        extendedConfig.settings.parameterRules = extendedConfig.settings.parameterRules ?: []
        extendedConfig.settings.parameterRules += [name: 'Authorization', location: 'header', inject: true, value: "Bearer ${responseJson?.access_token}"]
        extendedConfig
    }
}

def runScan(parameters, utils, config, body) {
    withCredentials([string(
        credentialsId: config.settings.dasterTokenCredentialsId,
        variable: 'token',
    )]) {
        def extendedConfig = [:].plus(config)
        extendedConfig.settings.remove('dasterTokenCredentialsId')
        extendedConfig.settings.dasterToken = token

        def daster = parameters.dasterStub ?: new Daster(this, utils, extendedConfig)
        def scan = daster.triggerScan()
        echo "[${STEP_NAME}][INFO] Triggered scan of type ${config.scanType}${scan.message ? ' and received message: \'' + scan.message  + '\'' : ''}: ${scan.url ?: scan.scanId + ' and waiting for it to complete'}"
        if (scan?.scanId) {
            if(config.synchronous && config.scanType != 'burpscan') {
                try {
                    def scanResponse = [:]
                    while (scanResponse?.state?.terminated == null) {
                        scanResponse = daster.getScanResponse(scan?.scanId)
                        sleep(15)
                    }
                    def scanResult = daster.getScanResult(scanResponse)
                    def thresholdViolations = checkThresholdViolations(config, scanResult)
                    if (thresholdViolations) {
                        error "[${STEP_NAME}][ERROR] Threshold(s) ${thresholdViolations} violated by findings '${scanResult.summary}'"
                    } else if (scanResponse?.state?.terminated?.exitCode) {
                        error "[${STEP_NAME}][ERROR] Scan failed with code '${scanResponse?.state?.terminated?.exitCode}', reason '${scanResponse?.state?.terminated?.reason}' on container '${scanResponse?.state?.terminated?.containerID}'"
                    } else {
                        echo "Result of scan is ${scanResponse}"
                    }
                } finally {
                    daster.downloadAndAttachReportJSON(scan?.scanId, "result.json")
                    daster.downloadAndAttachReportJSON(scan?.scanId, "report.sarif.json")
                    daster.downloadAndAttachReportJSON(scan?.scanId, "report.xml")


                    if (config.deleteScan)
                        daster.deleteScan(scan?.scanId)
                }
            } else if (config.scanType == 'burpscan') {
                try {
                    withEnv(["BURP_PROXY=${scan.proxyURL}".toString()]) {
                        body()
                    }
                    def scanResponse = daster.getScanResponse(scan.scanId)
                    def scanResult = daster.getScanResult(scanResponse)
                    def thresholdViolations = checkThresholdViolations(config, scanResult)
                    if (thresholdViolations) {
                        error "[${STEP_NAME}][ERROR] Threshold(s) ${thresholdViolations} violated by findings '${scanResult.summary}'"
                    } else if (scanResponse?.state?.terminated?.exitCode) {
                        error "[${STEP_NAME}][ERROR] Scan failed with code '${scanResponse?.state?.terminated?.exitCode}', reason '${scanResponse?.state?.terminated?.reason}' on container '${scanResponse?.state?.terminated?.containerID}'"
                    } else {
                        echo "Result of scan is ${scanResponse}"
                    }
                } finally {
                    daster.stopBurpScan(scan.scanId)
                }
            }
        }
    }
}

def checkThresholdViolations(config, scanResult) {
    def violations = [:]
    if(config.thresholds.fail && scanResult.summary) {
        if(config.thresholds.fail.high >= 0 && config.thresholds.fail?.high < scanResult.summary.High)
            violations['thresholds.fail.high'] = config.thresholds.fail.high
        if(config.thresholds.fail.medium >= 0 && config.thresholds.fail?.medium < scanResult.summary.Medium)
            violations['thresholds.fail.medium'] = config.thresholds.fail.medium
        if(config.thresholds.fail.low >= 0 && config.thresholds.fail?.low < scanResult.summary.Low)
            violations['thresholds.fail.low'] = config.thresholds.fail.low
        if(config.thresholds.fail.info >= 0 && config.thresholds.fail?.info < scanResult.summary.Informational)
            violations['thresholds.fail.info'] = config.thresholds.fail.info
        if(config.thresholds.fail.all >= 0 && config.thresholds.fail?.all < (scanResult.summary.High + scanResult.summary.Medium + scanResult.summary.Low + scanResult.summary.Informational))
            violations['thresholds.fail.all'] = config.thresholds.fail.all
    }
    return violations
}

import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.integration.ZedAttackProxy
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'executeZAPScan'
@Field Set GENERAL_CONFIG_KEYS = [
    'verbose'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'addonInstallList',
    'alertThreshold',
    'context',
    'dockerImage',
    'dockerWorkspace',
    'scanner',
    'stashContent',
    'suppressedIssues',
    'targetUrls',
    'zapPort'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

def call(Map parameters = [:]) {
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
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('targetUrls')
            .use()

        config.stashContent = utils.unstashAll(config.stashContent)

        def zedAttackProxy = parameters.zedAttackProxyStub
        if(zedAttackProxy == null) {
            zedAttackProxy = new ZedAttackProxy(this, config.verbose)
        }

        dockerExecute(script: script, dockerImage: "${config.dockerImage}", dockerWorkspace: "${config.dockerWorkspace}", stashContent: config.stashContent) {

            // configure zap database
            sh "cp /zap/db/zapdb.script /zap/db/zapdb.script.original"
            sh "sed \'/SET FILES BACKUP INCREMENT TRUE/,/SET FILES CACHE ROWS 50000/s/SET FILES CACHE SIZE 10000/SET FILES CACHE SIZE 100000/\' /zap/db/zapdb.script.original > /zap/db/zapdb.script"

            // fiddle around with proxy settings
            def proxyParts = env.http_proxy?.split(':') ?: env.HTTP_PROXY?.split(':')
            def proxyEnabled = ''
            def proxyHost = ''
            def proxyPort = ''
            if(null != proxyParts && proxyParts.size() == 3) {
                proxyEnabled = " -config connection.proxyChain.enabled=true"
                proxyHost = " -config connection.proxyChain.hostName=${proxyParts[1].substring(2)}"
                proxyPort = " -config connection.proxyChain.port=${proxyParts[2]}"
            }
            def noProxyParts = env.no_proxy?.split(',') ?: env.NO_PROXY?.split(',')
            def noProxy = ''
            if(null != noProxyParts && noProxyParts.size() > 0) {
                def count = 0
                noProxyParts.each {
                    part ->
                        def domain = part.startsWith('*') ? '.' + part.replaceAll('\\.', '\\\\\\\\.') : '.*' + part.replaceAll('\\.', '\\\\\\\\.')
                        noProxy += " -config connection.proxyChain.confirmRemoveExcludedDomain=false -config connection.proxyChain.exclusions.exclusion\\(${count}\\).name=${domain} -config connection.proxyChain.exclusions.exclusion\\(${count}\\).regex=true -config connection.proxyChain.exclusions.exclusion\\(${count}\\).enabled=true"
                        count++
                }
            }

            def alerts = []
            def completeHTMLResult = null
            try {
                // start ZAP
                def command = "/zap/zap-x.sh -daemon -port ${config.zapPort} -host 0.0.0.0 -config api.addrs.addr.name=.* -config api.addrs.addr.regex=true -config api.disablekey=true -config database.recoverylog=false${proxyEnabled}${proxyHost}${proxyPort}${noProxy} &"
                if(config.verbose)
                    echo "ZAP start command: ${command}"
                sh command
                zedAttackProxy.checkServerStatus(config)

                // apply general settings
                zedAttackProxy.enableAddonUpdates(config)
                zedAttackProxy.installScannerRules(config)
                zedAttackProxy.checkAddonUpdates(config)
                zedAttackProxy.enablePassiveScan(config)
                zedAttackProxy.enablePassiveScanners(config)
                zedAttackProxy.enableForcedUserMode(config)

                // install addons
                for (def i = 0; i < config.addonInstallList.size(); i++) {
                    zedAttackProxy.installAddon(config, config.addonInstallList[i])
                }

                // load scripts
                def zapScriptFiles = findFiles(glob: "**${File.separator}zap${File.separator}scripts${File.separator}**${File.separator}*")
                echo "[${STEP_NAME}] Found ${zapScriptFiles.length} zap script files"
                for (def i = 0; i < zapScriptFiles.length; i++) {
                    def file = zapScriptFiles[i].path
                    def name = zapScriptFiles[i].name
                    def parts = zapScriptFiles[i].path.split(/\$File.separator/.toString())
                    def type = parts[parts.size() - 2]
                    zedAttackProxy.loadScript(file, type, name, config)
                }
                zedAttackProxy.enableAllScripts(utils, config)

                // load context
                def zapContextFiles = findFiles(glob: "**${File.separator}zap${File.separator}context${File.separator}**${File.separator}*")
                echo "[${STEP_NAME}] Found ${zapContextFiles.length} zap context files"
                def contextId = -1
                for (def i = 0; i < zapContextFiles.length; i++) {
                    def file = zapContextFiles[i].path
                    contextId = zedAttackProxy.loadContext(utils, file, config)
                    if(contextId >= 0)
                        i = zapContextFiles.length
                }

                // no context provided so create one on the fly
                if(contextId < 0)
                    contextId = zedAttackProxy.createContext(utils, config)
                // create user in case credentialsId supplied
                if(config.context.user.credentialsId && !config.context.user.credentialsId.isEmpty()) {
                    zedAttackProxy.configureAuthenticationMethod(config, contextId)
                    withCredentials([[$class: 'UsernamePasswordMultiBinding', credentialsId: config.context.user.credentialsId, passwordVariable: 'password', usernameVariable: 'user']]) {
                        def userId = zedAttackProxy.createJenkinsUserCredentials(utils, config, contextId, user, password)
                        if (userId >= 0) {
                            config.context.user.id = userId
                        }
                    }
                }

                if(config.context.sessionManagement.method)
                    zedAttackProxy.configureSessionManagement(config, contextId)

                if(config.context.authorization.headerRegex || config.context.authorization.bodyRegex || config.context.authorization.statusCode)
                    zedAttackProxy.configureAuthorizationDetection(config, contextId)

                // handle forced user mode
                if(config.context.forcedUserMode)
                    zedAttackProxy.forceUser(config, contextId)

                // DO THE SCANNING WORK and iterate URLs
                for (int i = 0; i < config.targetUrls.size(); i++) {
                    // add URL to scope if context is available
                    if(contextId >= 0 && config.context.name)
                        zedAttackProxy.includeInContext(config.targetUrls[i], config)

                    // spider scan
                    if(config.scanner.spiderScan.enabled) {
                        def spiderScanId = zedAttackProxy.startSpiderScan(utils, config.targetUrls[i], contextId, config)
                        zedAttackProxy.checkSpiderScanStatus(utils, spiderScanId, config)
                        zedAttackProxy.getSpiderScanResults(utils, spiderScanId, config)
                    }

                    //active scan
                    if(config.scanner.activeScan.enabled) {
                        // send initial request
                        zedAttackProxy.sendRequest(config, config.targetUrls[i])

                        // trigger active scan on initial request
                        def activeScanId = zedAttackProxy.startActiveScan(utils, config.targetUrls[i], contextId, config)
                        zedAttackProxy.checkActiveScanStatus(utils, activeScanId, config)
                        zedAttackProxy.getActiveScanResults(utils, activeScanId, config)
                        zedAttackProxy.getFinalScanProgress(utils, activeScanId, config)
                    }

                    //AJAX scan
                    if(config.scanner.ajaxSpiderScan.enabled) {
                        zedAttackProxy.startAJAXScan(config.targetUrls[i], config)
                        zedAttackProxy.checkAJAXScanStatus(utils, config)
                    }

                    alerts += zedAttackProxy.fetchAlerts(utils, config, config.targetUrls[i])
                }

                completeHTMLResult = zedAttackProxy.getCompleteHTMLReport(config)
                if(config.verbose) {
                    try {
                        zedAttackProxy.getCompleteListOfMessages(utils, config)
                    } catch (e) {
                        echo "Failed to load list of requests sent, giving up."
                    }
                }
            } finally {
                try {
                    zedAttackProxy.shutdownServer(config)
                } catch (e) {
                    echo "Failed to shutdown ZAP instance explicitly, tearing down container now."
                }

                if (null != completeHTMLResult) {
                    def reportFile = 'zap_report.html'
                    writeFile file: reportFile, text: completeHTMLResult
                    publishTestResults([html: [ active: true,
                                                allowEmptyResults: false,
                                                archive: true,
                                                file: reportFile,
                                                name: 'ZAP Scan Report -',
                                                path: '']
                    ])
                }

                zedAttackProxy.filterAlertsAndCheckFailConditions(config, alerts)
            }
        }
    }
}

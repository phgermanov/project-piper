package com.sap.piper.internal.integration

import hudson.AbortException

class ZedAttackProxy implements Serializable {

    def script
    def verbose

    ZedAttackProxy(def script, def verbose) {
        this.script = script
        this.verbose = verbose
    }

    void checkServerStatus(config) {
        def serverStatusUrl = "localhost:${config.zapPort}/JSON/core/view/version/"
        def available = false

        // let's wait some time to let it begin starting
        script.sleep(15)

        def errorCount = 0
        def attempts = 0
        while(!available) {
            try {
                attempts++
                def response = triggerApiRequest(serverStatusUrl)
                if(response != null) {
                    available = true
                } else if(attempts <= 20) {
                    script.sleep(15)
                } else {
                    throw new AbortException("Giving up after 5 minutes polling")
                }
            }
            catch (error) {
                errorCount++
                if(errorCount > 3)
                script.error "ZAP daemon failed to start within the allowed time frame"

                script.sleep(15)
            }
        }
    }

    void shutdownServer(config) {
        def shutdownServerURL = "localhost:${config.zapPort}/JSON/core/action/shutdown/"

        def response = triggerApiRequest(shutdownServerURL)
        if (response != null) {
            script.echo "Server will shutdown!"
        }
    }

    String startSpiderScan(utils, targetUrl, contextId, config) {
        def spiderScanStartURL = "localhost:${config.zapPort}/JSON/spider/action/${config.scanner.scanMode}/?recurse=${config.scanner.recurse}${buildContextPart(config, contextId)}&url=${urlEncode(targetUrl)}${buildUserPart(config)}${config.scanner.additionalParameters}"

        script.echo "Trying to spider scan ${targetUrl}"
        def response = triggerApiRequest(spiderScanStartURL)
        if (response != null) {
            def responseBody = utils.parseJsonSerializable(response)
            script.echo "Site available and scan started!"
            script.echo "Scan ID: '${responseBody[config.scanner.scanMode]}'"
            return responseBody[config.scanner.scanMode]
        }
    }

    void checkSpiderScanStatus(utils, scanId, config) {
        def spiderScanStatusURL = "localhost:${config.zapPort}/JSON/spider/view/status/?scanId=${scanId}"
        def finished = false

        while (!finished) {
            def response = triggerApiRequest(spiderScanStatusURL)
            if (response != null) {
                def responseBody = utils.parseJsonSerializable(response)
                script.echo "Scanned ${responseBody.status} %"
                if (responseBody.status == "100") {
                    finished = true
                } else {
                    script.sleep(15)
                }
            }
        }
    }

    void getSpiderScanResults(utils, scanId, config) {
        def spiderScanResultURL = "localhost:${config.zapPort}/JSON/spider/view/results/?scanId=${scanId}"

        def response = triggerApiRequest(spiderScanResultURL)
        if (response != null) {
            def spiderScanResult = utils.parseJsonSerializable(response)
            script.echo "Spiderscan result: '${spiderScanResult}'"
        }
    }

    String startActiveScan(utils, targetUrl, contextId, config) {
        def activeScanStartURL = "localhost:${config.zapPort}/JSON/ascan/action/${config.scanner.scanMode}/?recurse=${config.scanner.recurse}&inScopeOnly=${config.scanner.inScope}&scanPolicyName=&method=${config.scanner.activeScan.method}&postData=${java.net.URLEncoder.encode(config.scanner.activeScan.postData, 'UTF-8')}${buildContextPart(config, contextId)}&url=${urlEncode(targetUrl)}${buildUserPart(config)}${config.scanner.additionalParameters}"

        script.echo "Trying to active Scan ${targetUrl}"
        def response = triggerApiRequest(activeScanStartURL)
        if (response != null) {
            def responseBody = utils.parseJsonSerializable(response)
            script.echo "Site available and scan started!"
            script.echo "Scan ID: '${responseBody[config.scanner.scanMode]}'"
            return responseBody[config.scanner.scanMode]
        }
    }

    void checkActiveScanStatus ( utils, scanId, config) {
        def activeScanStatusURL = "localhost:${config.zapPort}/JSON/ascan/view/status/?scanId=${scanId}"
        def finished = false

        while (!finished) {
            def response = triggerApiRequest(activeScanStatusURL)
            if (response != null) {
                def responseBody = utils.parseJsonSerializable(response)
                script.echo "Scanned ${responseBody.status} %"
                if (responseBody.status == "100") {
                    finished = true
                } else {
                    script.sleep(15)
                }
            }
        }
    }

    void getActiveScanResults(utils, scanId, config) {
        def activeScanResultURL = "localhost:${config.zapPort}/JSON/ascan/view/alertsIds/?scanId=${scanId}"
        def response = triggerApiRequest(activeScanResultURL)
        if (response != null) {
            def activeScanResult = utils.parseJsonSerializable(response)
            script.echo "Activescan Result: '${activeScanResult}'"
        }
    }

    void getFinalScanProgress(utils, scanId, config) {
        def activeScanResultURL = "localhost:${config.zapPort}/JSON/ascan/view/scanProgress/?scanId=${scanId}"
        def response = triggerApiRequest(activeScanResultURL)
        if (response != null) {
            def json = utils.parseJsonSerializable(response)
            def output = "=============================================================================================================\n"
            output += "Active scan status results for URL: ${json.scanProgress[0]}\n"
            output += "-------------------------------------------------------------------------------------------------------------\n"
            def pluginResults = json.scanProgress[1].HostProcess
            for(int i = 0; i < pluginResults.size(); i++) {
                output += "Plugin: '${pluginResults[i].Plugin[0]}', Elapsed (ms): '${pluginResults[i].Plugin[4]}', Requests: '${pluginResults[i].Plugin[5]}', Alerts: '${pluginResults[i].Plugin[6]}', Status: '${pluginResults[i].Plugin[3]}'\n"
            }
            output += "=============================================================================================================\n"
            script.echo output
        }
    }

    void sendRequest(config, url) {
        def request = "${config.scanner.activeScan.method} ${url} HTTP/1.1\r\n${config.scanner.activeScan.header}${config.scanner.activeScan.postData ? '\r\n\r\n' + config.scanner.activeScan.postData + '\r\n' : ''}"
        def apiRequestURL = "localhost:${config.zapPort}/JSON/core/action/sendRequest/?followRedirects=true&request=${urlEncode(request)}"
        script.echo "Initiating target URL request to '${url}'"
        triggerApiRequest(apiRequestURL)
    }

    void startAJAXScan(targetUrl, config) {
        def ajaxScanStartURL = "localhost:${config.zapPort}/JSON/ajaxSpider/action/${config.scanner.scanMode}/?inScope=${config.scanner.inScope}&subtreeOnly=&url=${urlEncode(targetUrl)}${buildUserPart(config)}${buildContextPart(config, null)}${config.scanner.additionalParameters}"

        script.echo "Trying to AJAX Scan '${targetUrl}'"
        def response = triggerApiRequest(ajaxScanStartURL)
        if (response != null) {
            script.echo "Site available and Scan started!"
        }
    }

    void checkAJAXScanStatus(utils, config) {
        def ajaxScanStatusURL = "localhost:${config.zapPort}/JSON/ajaxSpider/view/status/"
        def finished = false

        while (!finished) {
            def response = triggerApiRequest(ajaxScanStatusURL)
            if (response != null) {
                def responseBody = utils.parseJsonSerializable(response)
                script.echo "AJAX Scan is '${responseBody.status}'"
                if (responseBody.status == "stopped") {
                    finished = true
                } else {
                    script.sleep(15)
                }
            }
        }
    }

    String getCompleteHTMLReport(config) {
        def ajaxScanResultURL = "localhost:${config.zapPort}/OTHER/core/other/htmlreport/"
        triggerApiRequest(ajaxScanResultURL)
    }

    void getCompleteListOfMessages(utils, config) {
        def activeScanResultURL = "localhost:${config.zapPort}/JSON/core/view/messages/"
        def response = triggerApiRequest(activeScanResultURL)
        if (response != null) {
            def json = utils.parseJsonSerializable(response)
            script.echo "List of messages sent: ${json}"
        }
    }

    void loadScript(scriptPath, scriptType, scriptName, config) {
        def scriptPathUrlEncoded = urlEncode(script.pwd() + File.separator + scriptPath)
        def apiRequestURL = "localhost:${config.zapPort}/JSON/script/action/load/?scriptName=${scriptName}&scriptType=${scriptType}&scriptEngine=Oracle+Nashorn&fileName=${scriptPathUrlEncoded}&scriptDescription=&charset="
        script.echo "Loading script '${scriptName}' at '${scriptPath}'"
        triggerApiRequest(apiRequestURL)
    }

    void enableAllScripts(utils, config) {
        def apiRequestURL = "localhost:${config.zapPort}/JSON/script/view/listScripts"
        script.echo "Loading list of installed scripts"
        def response = triggerApiRequest(apiRequestURL)
        if(null != response) {
            def json = utils.parseJsonSerializable(response)
            def list = json.listScripts
            for( int i = 0; i < list.size(); i++) {
                def currentScript = list[i]
                if(currentScript.enabled != null && currentScript.enabled == false) {
                    apiRequestURL = "localhost:${config.zapPort}/JSON/script/action/enable/?scriptName=${currentScript.name}"
                    script.echo "Enabling script '${scriptName}'"
                    triggerApiRequest(apiRequestURL)
                }
            }
        }
    }

    int loadContext(utils, contextPath, config) {
        def contextPathUrlEncoded = urlEncode(script.pwd() + File.separator + contextPath)
        def apiRequestURL = "localhost:${config.zapPort}/JSON/context/action/importContext/?contextFile=${contextPathUrlEncoded}"
        script.echo "Loading context '${contextPath}'"
        def response = triggerApiRequest(apiRequestURL)
        if (response != null) {
            def json = utils.parseJsonSerializable(response)
            script.echo "Context successfully loaded with ID '${json.contextId}'"
            return json.contextId
        }
        return -1
    }

    void includeInContext(urlPattern, config) {
        def urlPatternUrlEncoded = urlEncode(urlPattern + ".*")
        def apiRequestURL = "localhost:${config.zapPort}/JSON/context/action/includeInContext/?regex=${urlPatternUrlEncoded}${buildContextPart(config, null)}"
        script.echo "Setting inclusion pattern for URL '${urlPattern}'"
        triggerApiRequest(apiRequestURL)
    }

    void enablePassiveScan(config) {
        def apiRequestURL = "localhost:${config.zapPort}/JSON/pscan/action/setEnabled/?enabled=${config.scanner.passiveScan.enabled}"
        script.echo "Setting passive scanning enabled to '${config.scanner.passiveScan.enabled}'"
        triggerApiRequest(apiRequestURL)
    }

    void enablePassiveScanners(config) {
        def apiRequestURL
        if(config.scanner.passiveScan.scanners.size() == 0) {
            apiRequestURL = "localhost:${config.zapPort}/JSON/pscan/action/enableAllScanners/"
            script.echo "Enabling *ALL* passive scanners"
        } else {
            apiRequestURL = "localhost:${config.zapPort}/JSON/pscan/action/enableScanners/?ids=${config.scanner.passiveScan.scanners}"
            script.echo "Enabling dedicated passive scanners '${config.scanner.passiveScan.scanners}'"
        }
        triggerApiRequest(apiRequestURL)
    }

    void enableAddonUpdates(config) {
        def apiRequestURL = "localhost:${config.zapPort}/JSON/autoupdate/action/setOptionInstallAddonUpdates/?Boolean=true"
        script.echo "Setting option to enable addon updates"
        triggerApiRequest(apiRequestURL)
    }

    void installScannerRules(config) {
        def apiRequestURL = "localhost:${config.zapPort}/JSON/autoupdate/action/setOptionInstallScannerRules/?Boolean=true"
        script.echo "Setting option to install scanner rules"
        triggerApiRequest(apiRequestURL)
    }

    void checkAddonUpdates(config) {
        def apiRequestURL = "localhost:${config.zapPort}/JSON/autoupdate/action/setOptionCheckAddonUpdates/?Boolean=true"
        script.echo "Setting option to check add on updates"
        triggerApiRequest(apiRequestURL)
    }

    void installAddon(config, addonId) {
        def apiRequestURL = "localhost:${config.zapPort}/JSON/autoupdate/action/installAddon/?id=${addonId}"
        script.echo "Installing addon with ID ${addonId}"
        def response = triggerApiRequest(apiRequestURL)
        if(null != response)
            script.echo "Installation of addon with ID '${addonId}' resulted in '${response}'"
    }

    void enableForcedUserMode(config) {
        def apiRequestURL = "localhost:${config.zapPort}/JSON/forcedUser/action/setForcedUserModeEnabled/?boolean=${config.context.forcedUserMode}"
        script.echo "${config.context.forcedUserMode ? 'Enabling' : 'Disabling'} forced user mode"
        triggerApiRequest(apiRequestURL)
    }

    void forceUser(config, contextId) {
        def apiRequestURL = "localhost:${config.zapPort}/JSON/forcedUser/action/setForcedUser/?userId=${config.context.user.id}${buildContextPart(config, contextId)}"
        script.echo "Forcing user with ID '${config.context.user.id}' for context with ID '${contextId}'"
        triggerApiRequest(apiRequestURL)
    }

    def createContext(utils, config) {
        if(!config.context.name)
            config.context.name = 'Piper Default'
        def apiRequestURL = "localhost:${config.zapPort}/JSON/context/action/newContext/?contextName=${urlEncode(config.context.name)}"
        script.echo "Creating new context with name '${config.context.name}'"
        def response = triggerApiRequest(apiRequestURL)
        if(null != response) {
            def json = utils.parseJsonSerializable(response)
            script.echo "Context with ID '${json.contextId}' successfully created"
            return json.contextId
        }
        return -1
    }

    void configureAuthenticationMethod(config, contextId) {
        def authenticationMethodConfig = ''
        for(int i = 0; i < config.context.authentication.methodConfiguration.keySet().size(); i++) {
            def key = config.context.authentication.methodConfiguration.keySet()[i]
            if(i > 0)
                authenticationMethodConfig += "&"
            authenticationMethodConfig += "${key}=${config.context.authentication.methodConfiguration[key]}"
        }
        def apiRequestURL = "localhost:${config.zapPort}/JSON/authentication/action/setAuthenticationMethod/?authMethodName=${config.context.authentication.method}${buildContextIdPart(contextId)}&authMethodConfigParams=${urlEncode(authenticationMethodConfig)}"
        script.echo "Setting authentication configuration for context with ID '${contextId}'"
        triggerApiRequest(apiRequestURL)
        apiRequestURL = "localhost:${config.zapPort}/JSON/authentication/action/setLoggedInIndicator/?loggedInIndicatorRegex=${urlEncode(config.context.authentication.loggedInIndicator)}${buildContextIdPart(contextId)}"
        script.echo "Setting logged-in indicator '${config.context.authentication.loggedInIndicator}' for context with ID '${contextId}'"
        triggerApiRequest(apiRequestURL)
        apiRequestURL = "localhost:${config.zapPort}/JSON/authentication/action/setLoggedOutIndicator/?loggedOutIndicatorRegex=${urlEncode(config.context.authentication.loggedOutIndicator)}${buildContextIdPart(contextId)}"
        script.echo "Setting logged-out indicator '${config.context.authentication.loggedOutIndicator}' for context with ID '${contextId}'"
        triggerApiRequest(apiRequestURL)
    }

    def createJenkinsUserCredentials(utils, config, contextId, user, password) {
        def userId = -1
        def apiRequestURL = "localhost:${config.zapPort}/JSON/users/action/newUser/?name=${urlEncode(config.context.user.name)}${buildContextPart(config, contextId)}"
        script.echo "Creating new user with name '${config.context.user.name}' in context with ID '${contextId}'"
        def response = triggerApiRequest(apiRequestURL)
        if (null != response) {
            def json = utils.parseJsonSerializable(response)
            script.echo "User with ID '${json.userId}' successfully created"
            userId = json.userId
        }
        if(userId >= 0) {
            def credentials = "${urlEncode(config.context.user.userParameterName)}=${urlEncode(user.replaceAll('\\+', '%2B'))}&${urlEncode(config.context.user.passwordParameterName)}=${urlEncode(password.replaceAll('\\+', '%2B'))}"
            for(int i = 0; i < config.context.user.additionalCreationParameters.keySet().size(); i++) {
                def key = config.context.user.additionalCreationParameters.keySet()[i]
                credentials += "&${key}=${urlEncode(config.context.user.additionalCreationParameters[key])}"
            }
            apiRequestURL = "localhost:${config.zapPort}/JSON/users/action/setAuthenticationCredentials/?userId=${userId}${buildContextIdPart(contextId)}&authCredentialsConfigParams=${urlEncode(credentials)}"
            script.echo "Setting authentication parameters for '${config.context.user.name}'"
            triggerApiRequest(apiRequestURL)
            apiRequestURL = "localhost:${config.zapPort}/JSON/users/action/setUserEnabled/?userId=${userId}${buildContextIdPart(contextId)}&enabled=true"
            script.echo "Setting '${config.context.user.name}' to enabled"
            triggerApiRequest(apiRequestURL)
            return userId
        }
        return -1
    }

    void configureSessionManagement(config, contextId) {
        def sessionMethodConfig = ''
        for(int i = 0; i < config.context.sessionManagement.methodConfiguration.keySet().size(); i++) {
            def key = config.context.sessionManagement.methodConfiguration.keySet()[i]
            if(i > 0)
                sessionMethodConfig += "&"
            sessionMethodConfig += "${key}=${config.context.sessionManagement.methodConfiguration[key]}"
        }
        def apiRequestURL = "localhost:${config.zapPort}/JSON/sessionManagement/action/setSessionManagementMethod/?methodName=${config.context.sessionManagement.method}${buildContextIdPart(contextId)}&methodConfigParams=${urlEncode(sessionMethodConfig)}"
        script.echo "Setting session management configuration for context with ID '${contextId}'"
        triggerApiRequest(apiRequestURL)
    }

    void configureAuthorizationDetection(config, contextId) {
        def apiRequestURL = "localhost:${config.zapPort}/JSON/authorization/action/setBasicAuthorizationDetectionMethod/?headerRegex=${urlEncode(config.context.authorization.headerRegex)}&bodyRegex =${urlEncode(config.context.authorization.bodyRegex)}&statusCode=${config.context.authorization.statusCode}&logicalOperator=${config.context.authorization.logicalOperator}${buildContextIdPart(contextId)}"
        script.echo "Setting authorization detection for context with ID '${contextId}'"
        triggerApiRequest(apiRequestURL)
    }

    def fetchAlerts(utils, config, url) {
        def apiRequestURL = "localhost:${config.zapPort}/JSON/core/view/alerts/?baseurl=${urlEncode(url)}"
        script.echo "Fetching alerts for '${url}'"
        def response = triggerApiRequest(apiRequestURL)
        if(null != response) {
            def json = utils.parseJsonSerializable(response)
            return json.alerts
        }
        return []
    }

    void filterAlertsAndCheckFailConditions(config, alerts) {
        def surpressedAlerts = []
        for (int i = 0; i < config.suppressedIssues.size(); i++) {
            def key = [:]
            key.cwe = config.suppressedIssues[i].cwe
            key.url = config.suppressedIssues[i].url
            surpressedAlerts += key
        }

        def filteredAlerts = []
        for (int i = 0; i < alerts.size(); i++) {
            def key = [:]
            key.cwe = alerts[i].cweid
            key.url = alerts[i].url
            if(!surpressedAlerts.contains(key))
                filteredAlerts += alerts[i]
        }

        if(config.alertThreshold >= 0 && filteredAlerts.size() > config.alertThreshold)
            script.error "Detected '${alerts.size()}' alert which is above configured threshold '${config.alertThreshold}': Active alerts are '${filteredAlerts}', suppressed alerts are '${config.suppressedIssues}'"
    }

    private String triggerApiRequest(url) {
        def response
        try {
            response = script.sh returnStdout: true, script: "${verbose ? '' : '#!/bin/sh -e\n'}wget --quiet --tries=1 -O - \"${url}\""
            if (response != null) {
                if (verbose)
                    script.echo "Operation succeeded: '${response}'"
            }
        } catch( e ) {
            try {
                response = script.sh returnStdout: true, script: "wget --tries=1 --content-on-error -O - \"${url}\""
            } catch( ex ) {
                script.error "Requested operation failed with exception '${ex}': response was '${response}'"
            }
        }
        return response
    }

    private String buildUserPart(config) {
        def userPart = ''
        if((config.context.user.id instanceof String && config.context.user.id?.length()
            || config.context.user.id instanceof Integer && config.context.user.id >= 0) && config.context.user.name) {
            userPart = "&userId=${config.context.user.id}&userName=${urlEncode(config.context.user.name)}"
        }
        return userPart
    }

    private String buildContextPart(config, contextId) {
        def contextPart = ''
        if(config.context.name && !config.context.name.isEmpty())
            contextPart = "&contextName=${urlEncode(config.context.name)}"
        contextPart += buildContextIdPart(contextId)
        return contextPart
    }

    private String buildContextIdPart(contextId) {
        def contextIdPart = ''
        if(contextId && contextId > 0)
            contextIdPart += "&contextId=${contextId}"
        return contextIdPart
    }

    private String urlEncode(value) {
        return java.net.URLEncoder.encode(value, 'UTF-8')
    }
}

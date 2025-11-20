package com.sap.piper.internal.integration

import com.sap.icd.jenkins.Utils

class Daster implements Serializable {

    final List<String> TRIGGER_ENDPOINTS = ['basicScan', 'oDataScan', 'swaggerScan', 'fioriDASTScan', 'aemscan', 'oDataFuzzer', 'burpscan']
    final List<String> STATUS_ENDPOINTS = ['fioriDASTScan', 'aemscan', 'oDataFuzzer', 'burpscan']
    final List<String> DELETE_ENDPOINTS = ['fioriDASTScan', 'oDataFuzzer']
    final List<Integer> RETRY_CODES = [100, 101, 102, 103, 404, 408, 425, /* not really common but a DASTer specific issue*/ 500, 503, 504]
    final Utils utils
    final Script script
    final Map config

    Daster(Script script, Utils utils, Map config) {
        this.script = script
        this.utils = utils
        this.config = config
        if (!this.config.serviceUrl)
            this.script.error "Parameter 'serviceUrl' must be provided as part of the configuration."
    }

    def triggerScan() {
        if (this.TRIGGER_ENDPOINTS.contains(this.config.scanType)) {
            def body = transformConfiguration()
            def parsedResponse = callApi("${this.config.scanType}", body)
            return parsedResponse
        }
        return [:]
    }

    def getScanResponse(scanId) {
        if (this.STATUS_ENDPOINTS.contains(this.config.scanType)) {
            return getScanStatus(scanId)
        }
        return [:]
    }

    def downloadAndAttachReportJSON(scanId, fileName) {
        if (this.config.scanType != 'burpscan') {
            try {
                def resultFile = callApi("report/${scanId}/${fileName}", null, 'GET', 'APPLICATION_JSON', false, [[name: 'Authorization', value: this.config.settings.dasterToken]])
                if (resultFile) {
                    this.script.writeFile file: fileName, text: resultFile
                } else {
                    this.script.error "Failed to download the result file."
                }

                this.script.archiveArtifacts artifacts: fileName, allowEmptyArchive: false
                this.script.echo "File downloaded and attached as artifact: ${fileName}"
            } catch (Exception e) {
                this.script.echo "Failed to download or attach the file: ${e.message}"
            }
        }
    }

    def getScanResult(scanResponse) {
        def result = [:]
        switch (this.config.scanType) {
            case 'fioriDASTScan':
                result.summary = scanResponse?.riskSummary
                result. details = scanResponse?.riskReport
                break
            case  'aemscan':
                result.details = scanResponse?.log
                break
        }
        return result
    }

    def deleteScan(scanId) {
        if (this.DELETE_ENDPOINTS.contains(this.config.scanType)) {
            def parsedResponse = callApi("${this.config.scanType}/${scanId}", null, 'DELETE', 'APPLICATION_JSON', false)
            return parsedResponse
        }
        return [:]
    }

    def stopBurpScan(scanId) {
        def body = transformConfiguration()
        def parsedResponse = callApi("burpscan/${scanId}/stop", body)
        return parsedResponse
    }

    private def getScanStatus(scanId) {
        def parsedResponse = callApi("${this.config.scanType}/${scanId}", null, 'GET')
        return parsedResponse
    }

    private def transformConfiguration() {
        def requestBody = [:].plus(config.settings)
        return requestBody
    }

    private def callApi(endpoint, requestBody = null, mode = 'POST', contentType = 'APPLICATION_JSON', parseJsonResult = true, customHeaders = []){
        def params = [
            url                    : "${this.config.serviceUrl}${endpoint}",
            httpMode               : mode,
            acceptType             : 'APPLICATION_JSON',
            contentType            : contentType,
            quiet                  : !this.config.verbose,
            consoleLogResponseBody : this.config.verbose,
            validResponseCodes     : '100:499',
            customHeaders          : customHeaders
        ]
        if (requestBody) {
            def requestBodyString = utils.jsonToString(requestBody)
            if (this.config.verbose) this.script.echo "Request with body ${requestBodyString} being sent."
            params.put('requestBody', requestBodyString)
        }
        def response = [status: 0]
        def attempts = 0
        while ((!response.status || RETRY_CODES.contains(response.status)) && attempts < this.config.maxRetries) {
            response = httpResource(params)
            attempts++
        }
        if (parseJsonResult)
            return this.utils.parseJsonSerializable(response.content)
        else
            return response.content
    }

    def httpResource(params) {
        this.script.httpRequest(params)
    }
}

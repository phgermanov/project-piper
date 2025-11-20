package com.sap.piper.internal.integration

import com.sap.icd.jenkins.Utils

class Vulas implements Serializable {

    final Utils utils
    final Script script
    final Map config
    String baseUrl
    String tenant = ''

    Vulas(Script script, Utils utils, Map config) {
        this.script = script
        this.utils = utils
        this.config = config
        if (!config.serverUrl || !config.backendEndpoint)
            this.script.error "Parameters 'serverUrl' and 'backendEndpoint' must be provided as part of the configuration."
        this.baseUrl = "${config.serverUrl}${config.backendEndpoint}/"
    }

    def initializeSpaceToken() {
        if (this.config.vulasSpaceToken && !this.config.pullRequestName) {
            this.script.echo "[Vulas] Using space token ${this.config.vulasSpaceToken}${this.config.ppmsID ? ', please ensure proper binding to PPMS object ' + this.config.ppmsID + ' as part of your workspace configuration in Vulas...' : ''}"
        } else {
            def errMessage = "[Vulas] Failed to either lookup or create vulas space for "
            lookupDefaultTenantToken()
            if (this.config.pullRequestName) {
                this.config.space.name = this.config.space.name ? "${this.config.space.name}_${this.config.pullRequestName}" : "PiperVulasSpace_${this.config.projectGroup}_${this.config.pullRequestName}"
                if (!this.config.space.description) this.config.space.description = "Piper managed Vulas space to scan projects with group ${this.config.projectGroup} and ${this.config.pullRequestName}"
                lookupSpaceByName()
                errMessage += "project group ${this.config.projectGroup} and ${this.config.pullRequestName}..."
            } else if (this.config.ppmsID) {
                if (!this.config.space.name) this.config.space.name = "PiperVulasSpace_${this.config.ppmsID}"
                if (!this.config.space.description) this.config.space.description = "Piper managed Vulas space to scan projects bound to PPMS Object ${this.config.ppmsID}"
                lookupSpaceByPPMSId()
                errMessage += "PPMS object ${this.config.ppmsID}..."
            } else {
                if (!this.config.space.name) this.config.space.name = "PiperVulasSpace_${this.config.projectGroup}"
                if (!this.config.space.description) this.config.space.description = "Piper managed Vulas space to scan projects with group ${this.config.projectGroup}"
                lookupSpaceByName()
                errMessage +=  "project group ${this.config.projectGroup}..."
            }

            // Lookup failed, therefore create a new workspace
            if (!this.config.vulasSpaceToken) {
                createSpace()
                if (!this.config.vulasSpaceToken)
                    this.script.error errMessage
            }

            this.script.echo "[Vulas] Using space token ${this.config.vulasSpaceToken}${!this.config.ppmsID && !config.pullRequestName?', please configure the missing PPMS object binding via config parameter \'ppmsID\' in Piper':''}"
        }
        return this.config.vulasSpaceToken
    }

    private lookupSpaceByPPMSId() {
        def parsedResponse = httpVulasJson("spaces/search?propertyName=ppmsObjNumber&value=${this.config.ppmsID}")
        if (parsedResponse && parsedResponse.size() == 1)
            this.config.vulasSpaceToken = parsedResponse[0].spaceToken
        else {
            for(int i = 0; i < parsedResponse.size(); i++) {
                def item = parsedResponse.get(i)
                if (item.spaceName == this.config.space.name) {
                    this.config.vulasSpaceToken = item.spaceToken
                }
            }
        }
        if (this.config.vulasSpaceToken)
            this.script.echo "[Vulas] Successfully looked up space with token ${this.config.vulasSpaceToken} for PPMS object ${this.config.ppmsID}"
    }

    private lookupSpaceByName() {
        def parsedResponse = httpVulasJson("spaces")
        for(int i = 0; i < parsedResponse.size(); i++) {
            def item = parsedResponse.get(i)
            if (item.spaceName == this.config.space.name) {
                this.config.vulasSpaceToken = item.spaceToken
            }
        }
        if (this.config.vulasSpaceToken)
            this.script.echo "[Vulas] Successfully looked up space with token ${this.config.vulasSpaceToken} for project group ${this.config.projectGroup}"
    }

    private lookupSpaceNameByToken() {
        def parsedResponse = httpVulasJson("spaces/${this.config.vulasSpaceToken}")
        this.config.space.name = parsedResponse.spaceName
    }

    def lookupVulnerabilities() {
        lookupSpaceNameByToken()
        def token = this.config.vulasSpaceToken
        def name = URLEncoder.encode(this.config.space.name, "UTF-8")
        return httpVulasJson("hubIntegration/apps/${name}%20(${token})/vulndeps")
    }

    def lookupVulnerabilitiesByGAV(gav) {
        lookupSpaceNameByToken()
        def token = this.config.vulasSpaceToken
        def name = URLEncoder.encode(this.config.space.name, "UTF-8")
        def gav_enc = URLEncoder.encode(gav, "UTF-8")
        return httpVulasJson("hubIntegration/apps/${name}%20(${token})${gav_enc}/vulndeps")
    }

    private createSpace() {
        def body = [
            spaceName           : this.config.space.name,
            spaceDescription    : this.config.space.description,
            exportConfiguration : this.config.space.exportConfiguration,
            public              : this.config.space.public,
            default             : false,
            spaceOwners         : this.config.space.owners
        ]
        if(this.config.ppmsID && !this.config.pullRequestName) {
            body += [ properties : [[
                                       source  : "USER",
                                       name    : "ppmsObjNumber",
                                       value   : this.config.ppmsID
                                   ]]
            ]
        }
        def parsedResponse = httpVulasJson("spaces", body, 'POST')
        this.config.vulasSpaceToken = parsedResponse.spaceToken
        this.script.echo "[Vulas] Successfully created new space with token ${this.config.vulasSpaceToken}"
    }

    private lookupDefaultTenantToken() {
        if (!this.tenant) {
            def parsedResponse = httpVulasJson("tenants/default")
            this.tenant = parsedResponse.tenantToken
        }
    }

    private httpVulasJson(url, body = null, mode = 'GET') {
        def response = httpVulas(url, body, mode)
        def parsedResponse = this.utils.parseJsonSerializable(response.content)
        if (this.config.verbose)
            this.script.echo "Parsed response is ${parsedResponse}"
        return parsedResponse
    }

    private httpVulas(path, body = null, mode = 'GET') {
        def params = [
            url                    : "${this.baseUrl}${path}",
            httpMode               : mode,
            acceptType             : 'APPLICATION_JSON',
            contentType            : 'APPLICATION_JSON',
            quiet                  : !this.config.verbose,
            ignoreSslErrors        : true,
            validResponseCodes     : "100:400",
            consoleLogResponseBody : config.verbose
        ]
        if (this.tenant) params += [customHeaders  : [[name: 'X-Vulas-Tenant', value: this.tenant]]]
        if(body) params += [requestBody : this.utils.jsonToString(body)]

        def response = this.script.httpRequest(params)

        if (this.config.verbose)
            this.script.echo "Received response ${response.content} with status ${response.status}"

        if (response.status == 400) {
            def message = "HTTP 400 - Bad request received as answer from Vulas backend: ${response.content}"
            this.script.error message
        }

        return response
    }
}

package com.sap.piper.internal.integration

import com.sap.icd.jenkins.Utils

class Checkmarx implements Serializable {

    final Utils utils
    final Script script
    final Map config

    List authHeaders

    Checkmarx(Script script, Utils utils, Map config) {
        this.script = script
        this.utils = utils
        this.config = config
        if(!this.config.checkmarxServerUrl)
            script.error "Parameter 'checkmarxServerUrl' must be provided as part of the configuration."
    }

    def initAuth(username, password) {
        this.authHeaders = fetchAuthToken(username, password)
    }

    def fetchProjectPreset(project) {
        def filteredProjectSettings
        project.links.each {
            link ->
                if(link.rel == 'scansettings') {
                    filteredProjectSettings = link.uri
                }
        }
        if(null == filteredProjectSettings)
            this.script.error "Unable to retrieve settings for project ${this.config.checkmarxProject}"

        def parsedResponse = callApi(filteredProjectSettings)
        return String.valueOf(parsedResponse?.preset?.id)
    }

    def fetchProjectByName(projectName) {
        def parsedResponse = callApi('/projects')
        return filterProject(parsedResponse, projectName)
    }

    def fetchProjectById(projectId) {
        return callApi("/projects/${projectId}")
    }

    def branchProject(projectId, branchName) {
        def parsedResponse = callApi("/projects/${projectId}/branch", "{ \"name\" : \"${branchName}\" }", 'POST', 'APPLICATION_JSON')
        return parsedResponse.id
    }

    private def filterProject(projectsResponse, projectName) {
        def result = [:]
        projectsResponse.each {
            project ->
                if(projectName == project.name) {
                    result = project
                }
        }
        return result
    }

    private def fetchAuthToken(username, password) {
        def body = "grant_type=password&" +
            "scope=sast_rest_api&" +
            "client_id=resource_owner_client&" +
            "client_secret=014DF517-39D1-4453-B7B3-9930C563627C&" +
            "username=${urlEncode(username)}&" +
            "password=${urlEncode(password)}"
        def parsedResponse = callApi('/auth/identity/connect/token', body, 'POST', 'APPLICATION_FORM', 3)
        return createAuthHeader(parsedResponse)
    }

    private def createAuthHeader(response) {
        return [[name: "Authorization", value: "${response.token_type} ${response.access_token}"]]
    }

    private def callApi(endpoint, requestBody = null, mode = 'GET', contentType = 'NOT_SET', maxAttempts = 1){
        def params = [
            url                    : "${this.config.checkmarxServerUrl}/CxRestAPI${endpoint}",
            httpMode               : mode,
            acceptType             : 'APPLICATION_JSON',
            contentType            : contentType,
            quiet                  : !this.config.verbose,
            validResponseCodes     : "100:400",
            consoleLogResponseBody : this.config.verbose
        ]
        if(requestBody) params.put('requestBody', requestBody)
        if(this.authHeaders) params.put('customHeaders', this.authHeaders)
        def attempts = 0
        def response
        def error = true
        while (attempts < maxAttempts && error) {
            try {
                attempts++
                response = httpResource(params)
                if (attempts >= maxAttempts && response?.status == 400 && endpoint == '/auth/identity/connect/token') {
                    this.script.error "Invalid credentials provided. Please verify whether you can manually log into Checkmarx using the supplied credentials."
                } else {
                    error = false
                }
            } catch (e) {
                if (attempts >= maxAttempts) {
                    throw e
                }
            }
        }
        return this.utils.parseJsonSerializable(response.content)
    }

    def httpResource(params) {
        this.script.httpRequest(params)
    }

    private String urlEncode(value) {
        return java.net.URLEncoder.encode(value, 'UTF-8')
    }
}

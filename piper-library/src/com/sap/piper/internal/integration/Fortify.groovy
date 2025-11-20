package com.sap.piper.internal.integration

import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.BashUtils
import com.sap.piper.internal.Notify

class Fortify implements Serializable{

    def config
    Utils utils

    Fortify(config, utils) {
        this.config = config
        this.utils = utils
    }

    def fetchProjectIDFor(script, projectName) {
        def fetchedProjects = httpProjectID(script, projectName)
        def projectID = 0
        for( int i = 0; i < fetchedProjects.size(); i++) {
        final project = fetchedProjects.get(i)
            if (project.name == projectName) {
                projectID = project.id
            }
        }
        if (projectID == 0)
            script.error "[Fortify] Could not find Fortify project with project name $projectName"

        projectID
    }

    def httpProjectID(script, projectName) {
        if(this.config.verbose)
            script.echo "[Fortify] Fetching project ID for ${projectName}"

        def encodedProjectName = URLEncoder.encode(projectName, "UTF-8")

        def response =
            sendFortifyApiRequest(script, "/projects?q=name%3A$encodedProjectName".toString())
        def parsedResponse = parseResponse(script, response)
        return parsedResponse.data
    }

    def fetchProjectVersionID(script, projectID, version, failOnUnavailable = true) {
        def fetchedVersions = httpVersionID(script, projectID)
        return getFortifyVersionIDFor(script, version, fetchedVersions, failOnUnavailable)
    }

    def httpVersionID(script, projectID) {
        if(this.config.verbose)
            script.echo "[Fortify] fetching project version ID for project with ID ${projectID}"

        def response = sendFortifyApiRequest(script, "/projects/$projectID/versions".toString())
        def parsedResponse = parseResponse(script, response)
        return parsedResponse.data
    }

    def getFortifyVersionIDFor(script, version, fetchedVersions, failOnUnavailable = true) {
        def projectVersionID = null
        for( int i = 0; i < fetchedVersions.size(); i++) {
        final listEntry = fetchedVersions.get(i)
            if (listEntry.name == version)
                projectVersionID = listEntry.id
        }

        if (projectVersionID == null && failOnUnavailable)
            script.error "[Fortify] Could not find Fortify project with project version $version"

        if (config.verbose)
            script.echo "[Fortify] Resolved project version ID as '${projectVersionID}'"

        projectVersionID
    }

    def lookupOrCreateProjectVersionIDForPR(script, projectID, masterProjectVersionID, prName) {
        def prProjectVersionId = fetchProjectVersionID(script, projectID, prName, false)
        if (prProjectVersionId) {
            return prProjectVersionId
        } else {
            def master = httpProjectVersionDetails(script, masterProjectVersionID)
            def newVersion = [:]
            newVersion.name = prName
            newVersion.description = master.description
            newVersion.active = master.active
            newVersion.committed = master.committed
            newVersion.project = [:]
            newVersion.project.name = master.project.name
            newVersion.project.description = master.project.description
            newVersion.project.id = master.project.id
            newVersion.project.issueTemplateId = master.project.issueTemplateId
            newVersion.issueTemplateId = master.issueTemplateId

            def newVersionId = httpCreateProjectVersion(script, newVersion)

            def attributes = httpProjectVersionAttributes(script, masterProjectVersionID)
            def newAttributes = []
            for( int i = 0; i < attributes.size(); i++) {
                def item = attributes.get(i)
                def newAttribute = [:]
                newAttribute.attributeDefinitionId = item.attributeDefinitionId
                newAttribute.value = item.value
                if (item.values) {
                    newAttribute.values = []
                    for( int r = 0; r < item.values.size(); r++) {
                        def m = [:]
                        m.guid = item.values.get(r).guid
                        newAttribute.values.add(m)
                    }
                } else {
                    newAttribute.values = null
                }
                newAttributes.add(newAttribute)
            }


            httpUpdateProjectVersionAttributes(script, newVersionId, newAttributes)

            httpCopyFromPartial(script, masterProjectVersionID, newVersionId)

            httpCommitProjectVersion(script, newVersionId)

            httpCopyCurrentState(script, masterProjectVersionID, newVersionId)

            httpCopyPermissions(script, masterProjectVersionID, newVersionId)

            newVersionId
        }
    }

    def mergeProjectVersionStateOfPRIntoMaster(script, projectID, masterProjectVersionID, prName) {
        def path = config.buildDescriptorFile?.replace('pom.xml', '')?.replace('setup.py', '')
        def fprFile = "${path}${prName}-transfer.fpr"
        def prProjectVersionId = fetchProjectVersionID(script, projectID, prName, false)
        if (prProjectVersionId) {
            def status = fetchFPR(script, prProjectVersionId, fprFile)
            if (this.config.verbose)
                script.echo "[Fortify] Download of latest PR FPR file of project version ${prProjectVersionId} ended with status ${status}"
            status = uploadFPR(script, masterProjectVersionID, fprFile)
            if (this.config.verbose)
                script.echo "[Fortify] Upload of PR FPR file into master project version ${masterProjectVersionID} ended with result ${status}"
            if(!status.contains(":code>-10001<"))
                script.error "[Fortify] Failed to transfer PR fpr into master project version ${masterProjectVersionID} on Fortify SSC at ${config.fortifyServerUrl}"
            inactivateProjectVersion(script, prProjectVersionId)
        } else {
            script.echo "[Fortify] Unable to detect project version for PR branch, skipping merge attempt"
        }
    }

    def httpProjectVersionDetails(script, projectVersionID) {
        def response =
            sendFortifyApiRequest(script, "/projectVersions/$projectVersionID".toString())
        def parsedResponse = parseResponse(script, response)
        return parsedResponse.data
    }

    def inactivateProjectVersion(script, projectVersionID) {
        def body = [ "active": false, "committed": true ]
        httpUpdateProjectVersionDetails(script, projectVersionID, body)
    }

    def httpUpdateProjectVersionDetails(script, projectVersionID, requestBody) {
        def serializedBody = utils.jsonToString(requestBody)
        def response =
            sendFortifyApiRequest(script, "/projectVersions/$projectVersionID".toString(), 'PUT', serializedBody)
        def parsedResponse = parseResponse(script, response)
        return parsedResponse.data
    }

    def httpProjectVersionAttributes(script, projectVersionID) {
        def response =
            sendFortifyApiRequest(script, "/projectVersions/$projectVersionID/attributes".toString())
        def parsedResponse = parseResponse(script, response)
        return parsedResponse.data
    }

    def httpUpdateProjectVersionAttributes(script, projectVersionID, requestBody) {
        def serializedBody = utils.jsonToString(requestBody)
        def response =
            sendFortifyApiRequest(script, "/projectVersions/$projectVersionID/attributes".toString(),'PUT', serializedBody)
        def parsedResponse = parseResponse(script, response)
        return parsedResponse.data
    }

    def httpCreateProjectVersion(script, requestBody) {
        def serializedBody = utils.jsonToString(requestBody)
        def response =
            sendFortifyApiRequest(script, "/projectVersions".toString(), 'POST', serializedBody)
        def parsedResponse = parseResponse(script, response)
        return parsedResponse.data.id
    }

    def httpCopyFromPartial(script, sourceProjectVersionID, targetProjectVersionID) {
        def requestBody = [
            "type": "COPY_FROM_PARTIAL",
            "values": [
                "projectVersionId": targetProjectVersionID,
                "previousProjectVersionId": sourceProjectVersionID,
                "copyAnalysisProcessingRules": true,
                "copyBugTrackerConfiguration": true,
                "copyCustomTags": true
            ]
        ]
        httpSendProjectVersionAction(script, targetProjectVersionID, requestBody)
    }

    def httpCopyCurrentState(script, sourceProjectVersionID, targetProjectVersionID) {
        def requestBody = [
            "type": "COPY_CURRENT_STATE",
            "values": [
                "projectVersionId": targetProjectVersionID,
                "previousProjectVersionId": sourceProjectVersionID
            ]
        ]
        httpSendProjectVersionAction(script, targetProjectVersionID, requestBody)
    }

    def httpCopyPermissions(script, sourceProjectVersionID, targetProjectVersionID) {
        def response = sendFortifyApiRequest(script, "/projectVersions/$sourceProjectVersionID/authEntities?embed=roles".toString())
        def parsedResponse = parseResponse(script, response)
        def serializedBody = utils.jsonToString(parsedResponse.data)
        sendFortifyApiRequest(script, "/projectVersions/$targetProjectVersionID/authEntities", 'PUT', serializedBody)
    }

    def httpCommitProjectVersion(script, projectVersionID) {
        if(this.config.verbose)
            script.echo "[Fortify] Commiting project version $projectVersionID"

        def response = sendFortifyApiRequest(script, "/projectVersions/$projectVersionID?hideProgress=true".toString(),'PUT','{ "committed": true }')
        def parsedResponse = parseResponse(script, response)
        return parsedResponse.data.token
    }

    def httpSendProjectVersionAction(script, projectVersionID, requestBody) {
        def serializedBody = this.utils.jsonToString(requestBody)
        def response =
            sendFortifyApiRequest(script, "/projectVersions/$projectVersionID/action".toString(), 'POST', serializedBody)
        def parsedResponse = parseResponse(script, response)
        return parsedResponse.data
    }

    def httpUploadStatus(script, projectVersionID) {
        def response =
            sendFortifyApiRequest(script, "/projectVersions/$projectVersionID/artifacts?embed=scans".toString())
        def parsedResponse = parseResponse(script, response)
        return parsedResponse.data
    }

    def httpFilterSets(script, projectVersionID) {
        def response =
            sendFortifyApiRequest(script, "/projectVersions/$projectVersionID/filterSets".toString())
        def parsedResponse = parseResponse(script, response)
        def result = [:]
        if(parsedResponse.data.get(0).title == "SAP") {
            result.templateId = parsedResponse.data.get(0).guid
            for( int i = 0; i < parsedResponse.data.get(0).folders.size(); i++) {
                final folder = parsedResponse.data.get(0).folders.get(i)
                if( folder.name == "Corporate Security Requirements") {
                    result.corporateRequirements = [:]
                    result.corporateRequirements.id = folder.guid
                    result.corporateRequirements.name = folder.name
                    result.corporateRequirements.type = "FOLDER"
                } else if( folder.name == "Audit All") {
                    result.auditAll = [:]
                    result.auditAll.id = folder.guid
                    result.auditAll.name = folder.name
                    result.auditAll.type = "FOLDER"
                } else if( folder.name == "Spot Checks of Each Category") {
                    result.spotChecks = [:]
                    result.spotChecks.id = folder.guid
                    result.spotChecks.name = folder.name
                    result.spotChecks.type = "FOLDER"
                }
            }
        }
        return result
    }

    def httpGroupings(script, projectVersionID) {
        def response =
            sendFortifyApiRequest(script, "/projectVersions/$projectVersionID/issueSelectorSet?type=GROUP".toString())
        def parsedResponse = parseResponse(script, response)
        def result = [:]
        for( int i = 0; i < parsedResponse.data.groupBySet.size(); i++) {
            final groupBySet = parsedResponse.data.groupBySet.get(i)
                if( groupBySet.displayName == "Analysis") {
                    result.analysis = [:]
                    result.analysis.id = groupBySet.guid
                    result.analysis.name = groupBySet.displayName
                    result.analysis.type = groupBySet.entityType
                } else if( groupBySet.displayName == "Category") {
                    result.category = [:]
                    result.category.id = groupBySet.guid
                    result.category.name = groupBySet.displayName
                    result.category.type = groupBySet.entityType
                } else if( groupBySet.displayName == "Folder") {
                    result.folder = [:]
                    result.folder.id = groupBySet.guid
                    result.folder.name = groupBySet.displayName
                    result.folder.type = groupBySet.entityType
                }
        }
        return result
    }

    def httpUnauditedIssues(script, projectVersionID, filterGroupMap) {
        def encodedFilterSetKey = URLEncoder.encode(filterGroupMap.templateId, "UTF-8")
        def urlValue = "/projectVersions/$projectVersionID/issueGroups?groupingtype=${filterGroupMap.folder.id}&filterset=$encodedFilterSetKey&showsuppressed=true".toString()
        def response = sendFortifyApiRequest(script, urlValue)
        def parsedResponse = parseResponse(script, response)
        return parsedResponse.data
    }

    def httpSpotIssues(script, projectVersionID, filterGroupMap) {
        def encodedFilterKey = URLEncoder.encode("${filterGroupMap.spotChecks.type}:${filterGroupMap.spotChecks.id}", "UTF-8")
        def encodedFilterSetKey = URLEncoder.encode(filterGroupMap.templateId, "UTF-8")
        def urlValue = "/projectVersions/$projectVersionID/issueGroups?groupingtype=${filterGroupMap.category.id}&filterset=$encodedFilterSetKey&filter=$encodedFilterKey&showsuppressed=true".toString()
        def response = sendFortifyApiRequest(script,urlValue)
        def parsedResponse = parseResponse(script, response)
        return parsedResponse.data
    }

    def httpSuspiciousExploitable(script, projectVersionID, filterGroupMap) {
        def encodedFilterSetKey = URLEncoder.encode(filterGroupMap.templateId, "UTF-8")
        def urlValue = "/projectVersions/$projectVersionID/issueGroups?groupingtype=${filterGroupMap.analysis.id}&filterset=$encodedFilterSetKey&showsuppressed=true".toString()
        def response = sendFortifyApiRequest(script, urlValue)
        def parsedResponse = parseResponse(script, response)
        return parsedResponse.data
    }

    def httpIssueStatistics(script, projectVersionID) {
        def result = [:]
        try {
            def urlValue = "/projectVersions/$projectVersionID/issueStatistics".toString()
            def response = sendFortifyApiRequest(script, urlValue)
            def parsedResponse = parseResponse(script, response)
            if (parsedResponse.data?.get(0))
                result = parsedResponse.data?.get(0)
        } catch (e) {
            // NOP
        }
        return result
    }

    def generateFortifyReport(script, projectID, projectVersionID, fortifyProjectName, projectVersion, reportType) {
        if(this.config.verbose)
            script.echo "[Fortify] Generating fortify report for ${fortifyProjectName}"

        def encodedProjectName = URLEncoder.encode(fortifyProjectName, "UTF-8")
        def encodedProjectFormat = URLEncoder.encode(reportType, "UTF-8")
        def reportTemplate = '{"name":"FortifyReport","note":"","type":"PORTFOLIO","reportDefinitionId":18,"format":"PDF","project":{"id":2837,"name":"Fortify","version":{"id":17540,"name":"develop"}},"projectVersionDisplayName":"develop","inputReportParameters":[{"name":"Q-gate-report","identifier":"projectVersionId","paramValue":17540,"type":"SINGLE_PROJECT"}]}'
        def jsonBody = utils.parseJsonSerializable(reportTemplate)
        jsonBody.name = 'FortifyReport: ' + encodedProjectName
        jsonBody.format = encodedProjectFormat
        jsonBody.project.id = projectID
        jsonBody.project.name = encodedProjectName
        jsonBody.project.version.id = projectVersionID
        jsonBody.project.version.name = projectVersion
        jsonBody.projectVersionDisplayName = projectVersion
        jsonBody.inputReportParameters[0].paramValue = projectVersionID
        reportTemplate = this.utils.jsonToString(jsonBody)

        def response = sendFortifyApiRequest(script, "/reports", 'POST', reportTemplate)
        def parsedResponse = parseResponse(script, response)

        return parsedResponse.data.id
    }

    void pollReportProcessingStatus(script, reportId) {
        if(this.config.verbose)
            script.echo "[Fortify] Polling for fortify report with ID ${reportId}"
        def response = sendFortifyApiRequest(script, "/reports/${reportId}".toString())
        def parsedResponse = parseResponse(script, response)
        def status = parsedResponse.data.status
        if( status == 'PROCESSING' || status == 'SCHED_PROCESSING') {
            script.sleep(10)
            pollReportProcessingStatus(script, reportId)
        }
    }

    private fetchFortifyReportToken(script) {
        if(this.config.verbose)
            script.echo "[Fortify] Fetching report file token"

        def parsedResponse = fetchFileToken(script, '{"fileTokenType": "REPORT_FILE"}')
        return parsedResponse.data.token
    }

    private fetchDownloadToken(script) {
        if(this.config.verbose)
            script.echo "[Fortify] Fetching file download token"

        def parsedResponse = fetchFileToken(script, '{"fileTokenType": "DOWNLOAD"}')
        return parsedResponse.data.token
    }

    private fetchUploadToken(script) {
        if(this.config.verbose)
            script.echo "[Fortify] Fetching file upload token"

        def parsedResponse = fetchFileToken(script, '{"fileTokenType": "UPLOAD"}')
        return parsedResponse.data.token
    }


    private fetchFileToken(script, String tokenRequest) {
        def response = sendFortifyApiRequest(script, "/fileTokens", 'POST', tokenRequest)
        def parsedResponse = parseResponse(script, response)
        if (parsedResponse.responseCode != 201)
            script.error "[Fortify] Failed to fetch token for request ${tokenRequest}"
        parsedResponse
    }

    void invalidateFileTokens(script) {
        try {
            sendFortifyApiRequest(script, "/fileTokens",'DELETE')
        } catch (e) {
            // NOP
        }
    }

    def fetchFPR(script, id, filePath) {
        def token = fetchDownloadToken(script)
        fetchFile(script, config.fortifyFprDownloadEndpoint, token, id, filePath, 'POST', 'APPLICATION_ZIP')
    }

    def fetchReport(script, id, filePath) {
        def token = fetchFortifyReportToken(script)
        fetchFile(script, config.fortifyReportDownloadEndpoint, token, id, filePath)
    }

    private fetchFile(script, urlPath, token, id, filePath, httpMode = 'GET', acceptType = 'APPLICATION_OCTETSTREAM') {
        try {
            def customHeaders = [[name: 'Cache-Control', value: 'no-cache, no-store, must-revalidate'], [name: 'Pragma', value: 'no-cache']]
            def serializedRequestContent = "mat=${token}&id=${id}"

            def response
            if (httpMode == 'GET') {
                def urlPathQuery = "${urlPath}?${serializedRequestContent}"
                response = sendFortifyRequest(script, urlPathQuery, httpMode, null, customHeaders,'APPLICATION_FORM', acceptType, filePath)
            } else {
                response = sendFortifyRequest(script, urlPath, httpMode, serializedRequestContent, customHeaders,'APPLICATION_FORM', acceptType, filePath)
            }
            return response.status
        } finally {
            invalidateFileTokens(script)
        }
    }

    def uploadFPR(script, id, filePath) {
        uploadFile(script, config.fortifyFprUploadEndpoint, id, filePath)
    }

    private uploadFile(script, urlPath, id, filePath) {
        def token = fetchUploadToken(script)
        /*
        try {
            def customHeaders = [[name: 'Cache-Control', value: 'no-cache, no-store, must-revalidate'], [name: 'Pragma', value: 'no-cache']]
            def requestContent = "file=@${filePath}&entityId=${id}&mat=${token}"

            def response = sendFortifyRequest(script, urlPath, 'POST', requestContent.toString(), customHeaders,'APPLICATION_OCTETSTREAM')
            return response.content
        } finally {
            invalidateFileTokens(script)
        }*/
        try {
            def response = executeScript(script, true, "curl --insecure -H 'Authorization:FortifyToken ${BashUtils.escape(this.config.fortifyToken, false)}' -X \"POST\" \"${config.fortifyServerUrl}${urlPath}?mat=${token}\" -F \"entityId=${id}\" -F \"file=@${filePath}\"")
            return response
        } finally {
            invalidateFileTokens(script)
        }
    }

    def sendFortifyApiRequest(script, urlPath, httpMode = 'GET', requestBody = null, headers = [], contentType = 'APPLICATION_JSON', acceptType = 'APPLICATION_JSON', outputFile = null) {
        sendFortifyRequest(script, "${config.fortifyApiEndpoint}${urlPath}", httpMode, requestBody, headers, contentType, acceptType, outputFile)
    }

    def sendFortifyRequest(script, urlPath, httpMode = 'GET', requestBody = null, headers = [], contentType = 'APPLICATION_JSON', acceptType = 'APPLICATION_JSON', outputFile = null) {
        headers += resolveHTTPAuthHeader()
        def parameters = [
            url                    : "${this.config.fortifyServerUrl}${urlPath}".toString(),
            acceptType             : acceptType,
            contentType            : contentType,
            ignoreSslErrors        : true,
            customHeaders          : headers,
            httpMode               : httpMode,
            quiet                  : !this.config.verbose,
            consoleLogResponseBody : this.config.verbose
        ]
        if (requestBody) parameters += [requestBody: requestBody]
        if (outputFile) parameters += [outputFile: outputFile]

        internalSendFortifyApiRequest(script, parameters)
    }

    def internalSendFortifyApiRequest(script, parameters) {
        Object rawResponse = script.httpRequest(parameters)
        return rawResponse
    }

    private resolveHTTPAuthHeader() {
        [[name: "Authorization", value: "FortifyToken ${this.config.fortifyToken}"]]
    }

    def parseResponse(script, response) {
        def jsonResponse = this.utils.parseJsonSerializable(response.content)
        def responseCode = jsonResponse?.responseCode ?: 200
        if (responseCode >= 200 && responseCode < 400) {
            return jsonResponse
        } else {
            Notify.error(script, "Request failed with code '${jsonResponse.responseCode}' and message '${jsonResponse.message}'", "Fortify")
        }
    }

    private executeScript(script, returnStdout, cmd) {
        script.isUnix()
            ? script.sh(returnStdout: returnStdout, script: cmd)
            : script.bat(returnStdout: returnStdout, script: cmd)
    }
}

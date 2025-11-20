package com.sap.piper.internal.integration

import com.cloudbees.groovy.cps.NonCPS
import com.sap.icd.jenkins.Utils
import hudson.AbortException

class PPMSRepository implements Serializable {

    final Script script
    final Utils utils
    def config
    def resourceID
    def buildVersionId

    PPMSRepository(Script script, Utils utils, config) {
        this.script = script
        this.utils = utils
        this.config = config
        this.resourceID = config.ppmsID.toString()
    }

    //required since direct function calls in constructor lead to 'hudson.remoting.ProxyException: com.cloudbees.groovy.cps.impl.CpsCallableInvocation'
    private String getBuildVersionId() {
        //do not retrieve if already available
        if (buildVersionId) return buildVersionId

        if (config.ppmsBuildVersion) {
            buildVersionId = getScvBuildVersionId(config.ppmsBuildVersion)
            if (!buildVersionId) script.error ("[ppmsCheck] Build version not available")
        } else {
            //use latest build version if build versions are available
            buildVersionId = getLatestScvBuildVersionId()
        }
        return buildVersionId
   }

    def fetchFOSSResponse() {
        def response = httpFossListFromPPMS(getBuildVersionId()?:resourceID)
        def parsedResponse = utils.parseJsonSerializable(response.content)
        return parsedResponse
    }

    def mapFossIDToFossObjects(response) {
        def fossLibraryRiskResponse = response.riskRating.fossRiskInfo
        return mapFossIDToFossObject(fossLibraryRiskResponse)
    }

    Map mapFossIDToFossObject(ppmsFossList) {
        def ppmsFossIDObjectMap = [:]

        for (int i = 0; i < ppmsFossList.size(); i++) {

            def fossItem = ppmsFossList[i]

            ppmsFossIDObjectMap[fossItem.fossnr] = fossItem
        }

        return ppmsFossIDObjectMap
    }

    def resolveMatchingChannelDescription(parsedResponse, ppmsChannelID) {

        def gtmcNames = parsedResponse.riskRating.gtmcName

        def fittingDescription = null

        for (int i = 0; i < gtmcNames.size(); i++) {
            def gtmc = gtmcNames[i]

            if (gtmc.gtmcId == ppmsChannelID) {
                fittingDescription = gtmc.gtmcName
                break
            }
        }

        return fittingDescription
    }

    def httpFossListFromPPMS(ppmsResourceID) {
        sendApiRequest(url: "${config.ppmsServerUrl}/ppmslight/rest/${ppmsResourceID}/riskrating".toString())
    }

    def httpPPMSProductVersionMetaInformation() {
        sendApiRequest(url: "${config.ppmsServerUrl}/ppmslight/rest/PV/${resourceID}/header?sap-language=en".toString())
    }

    // ---- new stuff -----

    def fetchResourceDetails() {

        //check with SCV first
        try {
            return getScvDetails()
        } catch (ignored) {
        }

        //no SCV found, try with PV now
        try {
            def response = httpPPMSProductVersionMetaInformation()
            def parsedResponse = script.readJSON(text: response.content)
            return [
                name: parsedResponse.header.pvname,
                overviewLink: "${config.ppmsServerUrl}/ppmslight/#/details/pv/${resourceID}/overview".toString()
            ]
        } catch (ignored) {
            throw new AbortException("Illegal PPMS ID, cannot determine whether is component version or product version")
        }

    }

    Map getScvDetails() {
        def response = sendApiRequest url: "${config.ppmsServerUrl}/odataint/borm/odataforosrcy/SoftwareComponentVersions('${resourceID}')?\$format=json", httpMode: 'GET', acceptType: 'APPLICATION_JSON'
        def scvDetails = script.readJSON(text: response.content).d ?: []

        return [
            overviewLink:  "${config.ppmsServerUrl}/ppmslight/#/details/cv/${resourceID}/overview".toString(),
            name: scvDetails.Name,
            technicalName: scvDetails.TechnicalName,
            technicalRelease: scvDetails.TechnicalRelease
        ]
    }

    String getScvBuildVersionId(ppmsBuildVersionName) {
        def buildVersions = getScvBuildVersions()
        def buildVersionId = null
        for(int i = 0; i < buildVersions.size(); i++) {
            def bv = buildVersions.get(i)
            if (bv.Name == ppmsBuildVersionName) {
                buildVersionId = bv.Id
            }
        }
        return buildVersionId
    }

    def getLatestScvBuildVersionId() {
        def buildVersions = getScvBuildVersions()
        if (buildVersions.size() == 0) {
            return null
        } else {
            sortBVs(buildVersions)
            return buildVersions[0].Id
        }
    }

    private List getScvBuildVersions() {
        def response = sendApiRequest url: "${config.ppmsServerUrl}/odataint/borm/odataforosrcy/SoftwareComponentVersions('${resourceID}')/BuildVersions?\$format=json", httpMode: 'GET', acceptType: 'APPLICATION_JSON'
        return script.readJSON(text: response.content).d?.results ?: []
    }

    @NonCPS
    private void sortBVs(m) {
        m.sort { v1, v2 ->
            int i1 = v1.SortSequence as Integer
            int i2 = v2.SortSequence as Integer
            i2 <=> i1

        }
    }

    Map getChangeRequestDocumentV1(userId, source, scvName = null) {

        def uniqueId = UUID.randomUUID().toString()

        def changeRequestDocument = [
            schemaVer        : "1-0-0",
            changeRequestData: [
                id                : uniqueId,
                scan              : [
                    source: source,
                    tool  : 'WhiteSource'
                ],
                comparison        : [
                    tool      : 'Piper',
                    timeStamp : (new Date()).format("yyyyMMddHHmmss"),
                    reviewedBy: userId
                ],
                target            : [
                    softwareComponentVersionNumber: resourceID,
                    softwareComponentVersionName  : scvName
                ],
                comment           : 'Auto-generated by Piper - Please add to PPMS model.',
                fossToAdd         : [],
                fossToRemove      : [],
                sendConfirmationTo: [[userId: userId]]
            ]
        ]

        if (getBuildVersionId()) {
            //changeRequestDocument.changeRequestData.target.buildVersionName = config.ppmsBuildVersion
            changeRequestDocument.changeRequestData.target.buildVersionNumber = getBuildVersionId()
        }

        return changeRequestDocument
    }

    Map getChangeRequestDocumentV2(userId, source, scvName = null) {

        def uniqueId = UUID.randomUUID().toString()

        def changeRequestDocument = [
            schemaVer: "2-0-0",
            changeRequestData: [
                id: uniqueId,
                provider: [
                    timeStamp: (new Date()).format("yyyyMMddHHmmss"),
                    tool: 'Piper',
                    reviewedBy: userId
                ],
                scan: [
                    source: source,
                    tool: 'WhiteSource'
                ],
                target: [
                    softwareComponentVersionNumber: resourceID,
                    softwareComponentVersionName: scvName
                ],
                comment: 'Auto-generated by Piper.',
                fossComprised: [],
                sendConfirmationTo: [[userId: userId]]
            ]
        ]

        if (getBuildVersionId()) {
            //changeRequestDocument.changeRequestData.target.buildVersionName = config.ppmsBuildVersion
            changeRequestDocument.changeRequestData.target.buildVersionNumber = getBuildVersionId()
        }

        return changeRequestDocument
    }

    String createChangeRequest(changeRequestDocument) {

        def xcsrfResponse = sendApiRequest(
            url: "${config.ppmsServerUrl}${config.ppmsChangeRequestEndpoint}",
            httpMode: 'HEAD',
            customHeaders: [[maskValue: true, name: 'x-csrf-token', value: 'Fetch']]
        )
        def token = xcsrfResponse.getHeaders().'x-csrf-token'[0]
        def cookies = xcsrfResponse.getHeaders().'set-cookie'

        def cookieList = []
        for(int i = 0; i < cookies.size(); i++) {
            def cookie = cookies.get(i)
            cookieList.add(splitCookie(cookie))
        }

        def crResponse = sendApiRequest(
            url: "${config.ppmsServerUrl}${config.ppmsChangeRequestEndpoint}",
            httpMode: 'POST',
            customHeaders: [
                [maskValue: true, name: 'x-csrf-token', value: token],
                [name: 'Cookie', value: cookieList.join('; ')]
            ],
            requestBody: changeRequestDocument
        )

        return (script.readJSON(text: crResponse.content)).crId
    }

    Map getChangeRequestStatus(changeRequestId) {
        def crStatus = getChangeRequestHeaderInfo(changeRequestId)
        return [id: crStatus.status, description: crStatus.statusText]
    }

    def getChangeRequestHeaderInfo(changeRequestId) {
        def crResponse = sendApiRequest(
            url: "${config.ppmsServerUrl}${config.ppmsChangeRequestEndpoint}/${changeRequestId}",
            httpMode: 'GET'
        )

        return script.readJSON(text: crResponse.content)
    }

    Map getBuildVersionCreationDocument(userId, buildVersionName, buildVersionDescription, copyPredecessorFoss = false, copyPredecessorCvBv = false) {

        def uniqueId = UUID.randomUUID().toString()

        def bvCreationDocument = [
            schemaVer: "1-0-0",
            changeRequestData: [
                id: uniqueId,
                provider: [
                    timeStamp: (new Date()).format("yyyyMMddHHmmss"),
                    tool: 'Piper',
                    reviewedBy: userId
                ],
                target: [
                    softwareComponentVersionNumber: resourceID,
                ],
                buildVersion: [
                    name: buildVersionName,
                    description: buildVersionDescription,
                    predecessorBuildVersionNumber: getLatestScvBuildVersionId()
                ],
                options: [
                    copyPredecessorFoss: copyPredecessorFoss,
                    copyPredecessorCvBv: copyPredecessorCvBv
                ]
            ]
        ]

        return bvCreationDocument
    }


    String createBuildVersion(buildVersionDocument) {

        def xcsrfResponse = sendApiRequest(
            url: "${config.ppmsServerUrl}${config.ppmsBuildVersionEndpoint}",
            httpMode: 'HEAD',
            customHeaders: [[maskValue: true, name: 'x-csrf-token', value: 'Fetch']]
        )
        def token = xcsrfResponse.getHeaders().'x-csrf-token'[0]
        def cookies = xcsrfResponse.getHeaders().'set-cookie'

        def cookieList = []
        for(int i = 0; i < cookies.size(); i++) {
            def cookie = cookies.get(i)
            cookieList.add(splitCookie(cookie))
        }

        def bvResponse = sendApiRequest(
            url: "${config.ppmsServerUrl}${config.ppmsBuildVersionEndpoint}",
            httpMode: 'POST',
            customHeaders: [
                [maskValue: true, name: 'x-csrf-token', value: token],
                [name: 'Cookie', value: cookieList.join('; ')]
            ],
            requestBody: buildVersionDocument
        )

        return (script.readJSON(text: bvResponse.content)).crId
    }

    Map getBuildVersionCreationStatus(bvRequestId) {
        def bvCrStatus = getBuildVersionCreationHeaderInfo(bvRequestId)
        return [id: bvCrStatus.status, description: bvCrStatus.statusText]
    }

    def getBuildVersionCreationHeaderInfo(bvRequestId) {
        def bvResponse = sendApiRequest(
            url: "${config.ppmsServerUrl}${config.ppmsBuildVersionEndpoint}/${bvRequestId}",
            httpMode: 'GET'
        )

        return script.readJSON(text: bvResponse.content)
    }

    private def sendApiRequest(params) {
        params.put('authentication', config.ppmsCredentialsId)
        params.put('quiet', !config.verbose)
        params.put('consoleLogResponseBody', config.verbose)
        return script.httpRequest(params)
    }


    private String splitCookie(setCookieHeader) {
        return setCookieHeader.split(';')[0]
    }

}

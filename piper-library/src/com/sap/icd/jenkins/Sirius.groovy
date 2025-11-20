package com.sap.icd.jenkins

class Sirius implements Serializable {
    def credentialsId = ''
    def siriusApiUrl = ''
    def siriusUploadUrl = ''
    def script

    Sirius(script, String credentialsId) {
        this(script, credentialsId, '', '')
    }

    Sirius(script, String credentialsId, siriusApiUrl, siriusUploadUrl) {
        this.credentialsId = credentialsId
        this.script = script
        this.siriusApiUrl = siriusApiUrl
        this.siriusUploadUrl = siriusUploadUrl
    }

    def setApiUrl(String url) {
        siriusApiUrl = url
    }

    def getApiUrl() {
        return siriusApiUrl
    }

    def setUploadUrl(String url) {
        siriusUploadUrl = url
    }

    def getUploadUrl() {
        return siriusUploadUrl
    }

    def getPrograms() {
        def response = script.httpRequest httpMode: 'GET', authentication: credentialsId, url: "${siriusApiUrl}/program"
        def programs = script.readJSON text: response.content
        return programs
    }

    def getProgramIntGuid(String programName) {
        def programs = getPrograms()

        //get int_guid for specific program
        def programIntGuid
        for (def i=0; i<programs.size(); i++) {
            if (programs[i].PROGRAM_NAME == programName) {
                programIntGuid = programs[i].PROGRAM_INT_GUID
                break
            }
        }
        return programIntGuid
    }

    //get delivery with specific name
    def getDeliveryByName(String programName, String deliveryName) {

        def programIntGuid = getProgramIntGuid(programName)

        deliveryName = java.net.URLEncoder.encode(deliveryName, "UTF-8")

        def response = script.httpRequest httpMode: 'GET', authentication: credentialsId, url: "${siriusApiUrl}/delivery?programGuid=${programIntGuid}&name=${deliveryName}"
        def delivery = script.readJSON text: response.content
        return delivery
    }

    def getDeliveryExtGuidByName(String programName, String deliveryName) {
        def delivery = getDeliveryByName(programName, deliveryName)
        return delivery[0].DELIVERY_EXT_GUID
    }

    def uploadDocument(String deliveryExtGuid, String taskGuid, String fileName, String documentName, String documentFamily, boolean confidential=false) {

        documentName = documentName?:fileName

        //Base64 encode file
        def fileNameMain = fileName.tokenize(".")[0]
        script.sh "base64 --wrap=0 ${fileName} >${fileNameMain}.b64"
        def fileContentBase64 = script.readFile file: "${fileNameMain}.b64"

        def data = """{
\"deliveryGuid\": \"${deliveryExtGuid}\",
\"originalPmtGuid\": \"${taskGuid}\",
\"fileName\": \"${fileName}\",
\"fileContent\": \"${fileContentBase64}\",
\"documentName\": \"${documentName}\",
\"documentFamily\": \"${documentFamily}\",
\"confidential\": \"${confidential?'X':''}\"
}"""

        script.httpRequest httpMode: 'POST', requestBody: data, authentication: credentialsId, url: "${siriusUploadUrl}"
    }
}

import groovy.json.JsonSlurper

/**
 * To get the system trust token, run:
 * def apiURLFromDefaults = script.commonPipelineEnvironment.getValue("hooks")?.systemtrust?.serverURL ?: ''
 * def token = sapGetSystemTrustToken(apiURLFromDefaults, config.vaultAppRoleSecretTokenCredentialsId, config.vaultPipelineName, config.vaultPath)
 */

def call(String apiURL, String credsID, String pipelineId, String groupId) {
    echo "Starting to request a token"

    // stop execution if input parameters are missing and proceed with secrets defined in Vault or Jenkins credentials
    if (!credsID || !pipelineId || !groupId) {
        echo '''
        Warning: Missing input parameters (credsID, pipelineId, groupId).
        Will continue with secrets defined in Vault or Jenkins credentials.
        Consider using System Trust integration for credentialless access as described here:
        https://github.wdf.sap.corp/pages/ContinuousDelivery/piper-doc/news/2025/02/11/system-trust-integration-now-available-on-jaas/
        '''
        return null
    }
    if (!apiURL) {
        echo '''
        Warning: Missing System Trust API URL. If you want to enable System Trust integration please put the URL into 'hooks' section of your defaults.
        Here you can find how it should look like: https://github.tools.sap/project-piper/resources/blob/b3e0f356bbee87c9c7b03bba2d88d50276eaeb17/gen/piper-defaults-azure-tools.yml#L974-L975
        '''
        return null
    }

    def token = null
    def gcpToken = readGcpToken()
    echo "Successfully retrieved GCP token"

    // Only attempt to get a token if the necessary data is provided
    if (credsID && pipelineId && groupId) {
        def groupIdTrimmed = groupId.replace("piper/", "")
        trustApiUrl = "${apiURL}/auth"
        wrap([$class: 'MaskPasswordsBuildWrapper', varPasswordPairs: [[password: gcpToken]]]) {
            // Make an API request to get the trust token
            def response = makeTrustApiRequest(trustApiUrl, pipelineId, groupIdTrimmed, credsID, gcpToken)
            token = processTrustApiResponse(response)
        }
    }
    // Return the token, which could be null if retrieval failed
    return token
}

def readGcpToken() {
    return sh(script: "cat /var/run/secrets/kubernetes.io/serviceaccount/token", returnStdout: true).trim()
}

def makeTrustApiRequest(String url, String pipelineId, String groupId, String secretId, String authToken) {
    writeFile file: ".pipeline/trust_api_request.sh", text: libraryResource("scripts/trust_api_request.sh")
    def scriptPath = ".pipeline/trust_api_request.sh"
    sh "chmod +x ${scriptPath}"
    // Secure handling of secrets using environment variables configured by Jenkins
    withCredentials([string(credentialsId: secretId, variable: 'vaultSecretID')]) {
        return sh(script: "${scriptPath} ${url} \$vaultSecretID ${pipelineId} ${groupId} ${authToken}", returnStdout: true).trim()
    }
}

def processTrustApiResponse(String response) {
    def httpStatus = response.replaceAll(/.*HTTPSTATUS:/, '')
    def body = response.replaceAll(/HTTPSTATUS:.*/, '')

    if (httpStatus.toInteger() != 200) {
        throw new RuntimeException("HTTP Status: $httpStatus, Failed to obtain System Trust session token, Response body: $body")
    }

    def jsonSlurper = new JsonSlurper()
    echo "Successfully retrieved System Trust session token"
    return jsonSlurper.parseText(body).token
}

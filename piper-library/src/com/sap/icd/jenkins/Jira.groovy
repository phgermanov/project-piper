package com.sap.icd.jenkins

class Jira implements Serializable {
    def credentialsId = ''
    def jiraApiUrl = ''
    def script

    Jira(script, String credentialsId) {
        this.credentialsId = credentialsId
        this.script = script
    }

    def setApiUrl(String url) {
        jiraApiUrl = url
    }

    def getApiUrl() {
        return jiraApiUrl
    }

    def setCredentialsId(String credentialsId) {
        this.credentialsId = credentialsId
    }

    def getCredentialsId() {
        return credentialsId
    }

    def getServerUrl() {
        def urlParts = jiraApiUrl.tokenize('/')
        return "${urlParts[0]}//${urlParts[1]}"
    }

    def searchIssuesWithFilterId(filterId) {
        def jql = "filter = ${filterId}"
        return searchIssuesWithJql(jql)
    }

    def searchIssuesWithJql(String query) {
        def jql = java.net.URLEncoder.encode(query, "UTF-8")
        def response = script.httpRequest httpMode: 'GET', acceptType: 'APPLICATION_JSON', authentication: credentialsId, url: "${jiraApiUrl}/search?jql=${jql}"
        def jiraIssues = script.readJSON text: response.content
        return jiraIssues
    }

    def getIssue(String jiraKey) {
        def response = script.httpRequest httpMode: 'GET', acceptType: 'APPLICATION_JSON', authentication: credentialsId, url: "${jiraApiUrl}/issue/${jiraKey}"
        def jiraIssue = script.readJSON text: response.content
        return jiraIssue
    }

}

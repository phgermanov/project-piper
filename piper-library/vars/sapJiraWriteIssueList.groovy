import com.sap.icd.jenkins.Jira
import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Deprecate
import groovy.text.GStringTemplateEngine
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'sapJiraWriteIssueList'
@Field Set GENERAL_CONFIG_KEYS = [
    'fileName',
    'jiraApiUrl',
    'jiraCredentialsId',
    'jiraFilterId',
    'jql',
    'reportTitle',
    'style'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

def call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        // handle deprecated parameters
        Deprecate.parameter(this, parameters, 'credentialsId', 'jiraCredentialsId')
        Deprecate.parameter(this, parameters, 'apiUrl', 'jiraApiUrl')
        Deprecate.parameter(this, parameters, 'filterId', 'jiraFilterId')
        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(style: libraryResource('piper.css'))
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('jiraCredentialsId')
            .use()

        def jiraObj = parameters.juStabJira
        if (jiraObj == null) {
            jiraObj = new Jira(script, config.jiraCredentialsId)
        } else {
            jiraObj.setCredentialsId(config.jiraCredentialsId)
        }

        jiraObj.setApiUrl(config.jiraApiUrl)

        def jiraIssues = [:]
        def searchFilter = ''

        if (config.jiraFilterId) {
            def response = httpRequest httpMode: 'GET', acceptType: 'APPLICATION_JSON', authentication: config.jiraCredentialsId, url: "${jiraObj.getApiUrl()}/filter/${config.jiraFilterId}"
            def jiraFilter = readJSON text: response.content
            searchFilter = jiraFilter.jql
            jiraIssues = jiraObj.searchIssuesWithFilterId(config.jiraFilterId)
        } else {
            config = new ConfigurationHelper(config)
                .withMandatoryProperty('jql')
                .use()
            searchFilter = config.jql
            jiraIssues = jiraObj.searchIssuesWithJql(searchFilter)
        }

        def now = new Date().format( 'MMM dd, yyyy - HH:mm:ss' )

        def issueTable = ''

        if (jiraIssues.issues.size() == 0)
            issueTable += '<tr><td colspan=7>No issues found</td></tr>'
        for (i = 0; i < jiraIssues.issues.size(); i++) {

            def creator = jiraIssues.issues[i].fields.creator?.displayName?jiraIssues.issues[i].fields.creator.displayName:''

            //required due to groovy.lang.MissingPropertyException: No such property: displayName for class: net.sf.json.JSONNull if assignee/creator is not set
            def assignee = ''
            try {
                assignee = jiraIssues.issues[i].fields.assignee?.displayName?jiraIssues.issues[i].fields.assignee.displayName:''
            } catch (groovy.lang.MissingPropertyException e) {
                echo "[${STEP_NAME}] Sub task is not assigned."
            }

            issueTable += "<tr><td>${i+1}</td><td><a href=\"${jiraObj.getServerUrl()}/browse/${jiraIssues.issues[i].key}\" target=\"_blank\">${jiraIssues.issues[i].key}</a></td><td>${jiraIssues.issues[i].fields.issuetype.name}</td><td>${jiraIssues.issues[i].fields.summary}</td><td>${jiraIssues.issues[i].fields.status.name}</td><td>${creator}</td><td>${assignee}</td></tr>"
        }

        writeFile file: config.fileName, text: GStringTemplateEngine.newInstance().createTemplate(libraryResource('com.sap.piper.internal/templates/jiraIssue.html')).make(
            [
                issueTable: issueTable,
                jiraIssues: jiraIssues,
                now: now,
                reportTitle: config.reportTitle,
                searchFilter: searchFilter,
                style: config.style
            ]).toString()

    }
}

import com.sap.icd.jenkins.Jira
import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Deprecate
import groovy.text.GStringTemplateEngine
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'sapJiraWriteTaskStatus'
@Field Set GENERAL_CONFIG_KEYS = [
    'fileName',
    'jiraApiUrl',
    'jiraCredentialsId',
    'jiraIssueKey',
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
        Deprecate.parameter(this, parameters, 'issueKey', 'jiraIssueKey')
        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(style: libraryResource('piper.css'))
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('jiraCredentialsId')
            .withMandatoryProperty('jiraIssueKey')
            .use()

        def jiraObj = parameters.juStabJira
        if (jiraObj == null) {
            jiraObj = new Jira(script, config.jiraCredentialsId)
        } else {
            jiraObj.setCredentialsId(config.jiraCredentialsId)
        }

        jiraObj.setApiUrl(config.jiraApiUrl)

        def jiraIssue = jiraObj.getIssue(config.jiraIssueKey)

        def jiraFixVersion = ''
        for (def i=0; i < jiraIssue.fields.fixVersions.size(); i++) {
            if (i > 0) jiraFixVersion = jiraFixVersion + ', '
            jiraFixVersion = jiraFixVersion + jiraIssue.fields.fixVersions[i].name
        }

        def assignee = (jiraIssue.fields.assignee!= null)?jiraIssue.fields.assignee.displayName:''

        def now = new Date().format( 'MMM dd, yyyy - HH:mm:ss' )

        def subTaskTable = ''

        for (def i=0; i < jiraIssue.fields.subtasks.size(); i++) {
            def response = httpRequest httpMode: 'GET', acceptType: 'APPLICATION_JSON', authentication: config.jiraCredentialsId, url: "${jiraIssue.fields.subtasks[i].self}"
            def subTask = readJSON text: response.content
            def subAssignee = ''
            //required due to groovy.lang.MissingPropertyException: No such property: displayName for class: net.sf.json.JSONNull if assignee is not set
            try {
                subAssignee = subTask.fields.assignee?subTask.fields.assignee.displayName:''
            } catch (groovy.lang.MissingPropertyException e) {
                echo "[${STEP_NAME}] Sub task is not assigned."
            }

            subTaskTable += "<tr><td><a href=\"${jiraObj.getServerUrl()}/browse/${jiraIssue.fields.subtasks[i].key}\" target=\"_blank\">${jiraIssue.fields.subtasks[i].key}</a></td><td>${jiraIssue.fields.subtasks[i].fields.summary}</td><td>${jiraIssue.fields.subtasks[i].fields.status.name}</td><td>${subAssignee}</td></tr>"
        }

        writeFile file: config.fileName, text: GStringTemplateEngine.newInstance().createTemplate(libraryResource('com.sap.piper.internal/templates/jiraTask.html')).make(
            [
                assignee: assignee,
                jiraFixVersion: jiraFixVersion,
                jiraIssue: jiraIssue,
                jiraServerUrl: jiraObj.getServerUrl(),
                now: now,
                reportTitle: config.reportTitle,
                style: config.style,
                subTaskTable: subTaskTable
            ]).toString()

    }
}

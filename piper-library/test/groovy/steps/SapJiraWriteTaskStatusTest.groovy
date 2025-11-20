#!groovy
 package steps

import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.junit.Assert.assertTrue

class SapJiraWriteTaskStatusTest extends BasePiperTest {

    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsWriteFileRule jwfr = new JenkinsWriteFileRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jwfr)
        .around(jsr) // needs to be activated after jedr, otherwise executeDocker is not mocked

    private class JiraMock {
        def credentialsId = ''
        def jiraApiUrl = ''
        JiraMock() {}
        def setCredentialsId(String credentialsId) {
            this.credentialsId = credentialsId
        }

        def setApiUrl(jiraApiUrl) {
            this.jiraApiUrl = jiraApiUrl
        }

        def getServerUrl() {
            return 'https://jira.test'
        }

        def getIssue(key) {
            return [
                key: 'issueKey',
                fields: [
                    description: 'Task Description',
                    summary: 'Issue Summary',
                    status: [
                        name: 'completed'
                    ],
                    fixVersions: [
                        [name: 'version1'],
                        [name: 'version2']
                    ],
                    assignee: [
                        displayName: 'TestAssignee'
                    ],
                    subtasks: [
                        [
                            key: 'task1',
                            self: 'link1',
                            fields: [
                                summary: 'Summary 1',
                                status: [
                                    name: 'completed'
                                ]
                            ]
                        ],
                        [
                            key: 'task2',
                            self: 'link2',
                            fields: [
                                summary: 'Summary 2',
                                status: [
                                    name: 'completed'
                                ]
                            ]
                        ]
                    ]
                ]
            ]
        }
    }

    @Test
    void testWriteTask() {

        def httpRequestParams = [:]

        helper.registerAllowedMethod("httpRequest", [Map.class], {params ->
            httpRequestParams = params
            def content = ''
            if (params.url.contains('1')) {
                content = '1'
            } else {
                content = '2'
            }
            return [status : 200, content: content]
        })

        helper.registerAllowedMethod("readJSON", [Map.class], {params ->
            if (params.text == '1') {
                return [fields: [assignee: [displayName: 'assignee 1']]]
            } else {
                return [fields: [assignee: [displayName: 'assignee 2']]]
            }
        })


        def jiraMock = new JiraMock()

        jsr.step.sapJiraWriteTaskStatus([
            script: nullScript,
            jiraCredentialsId: 'JiraId',
            jiraIssueKey: 'testKey',
            juStabJira: jiraMock
        ])

        assertTrue(jwfr.files.get('taskStatus.html').contains('<title>Jira Task Status</title>'))
        assertTrue(jwfr.files.get('taskStatus.html').contains('<h1>Jira Task Status</h1>'))
        assertTrue(jwfr.files.get('taskStatus.html').contains('<p>As documented in Jira task <a href="https://jira.test/browse/issueKey">Issue Summary</a></p>'))
        assertTrue(jwfr.files.get('taskStatus.html').contains('<p>Relevant for version/delivery: version1, version2</p>'))
        assertTrue(jwfr.files.get('taskStatus.html').contains('<p>Overall Status: completed</p>'))
        assertTrue(jwfr.files.get('taskStatus.html').contains('<p>Assignee: TestAssignee</p>'))
        assertTrue(jwfr.files.get('taskStatus.html').contains('<p>Description: <br />Task Description</p>'))
        assertTrue(jwfr.files.get('taskStatus.html').contains('<tr><td><a href="https://jira.test/browse/task1" target="_blank">task1</a></td><td>Summary 1</td><td>completed</td><td>assignee 1</td></tr>'))
        assertTrue(jwfr.files.get('taskStatus.html').contains('<tr><td><a href="https://jira.test/browse/task2" target="_blank">task2</a></td><td>Summary 2</td><td>completed</td><td>assignee 2</td></tr>'))

        assertJobStatusSuccess()
    }
}

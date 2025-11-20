#!groovy
 package steps

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.junit.Assert.assertTrue

class SapJiraWriteIssueListTest extends BasePiperTest {

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

        def getApiUrl() {
            return 'https://jira.test/v1'
        }

        def searchIssuesWithFilterId(filterId) {
            return searchIssuesWithJql(filterId)
        }

        def searchIssuesWithJql(searchFilter) {
            return [
                total: 2,
                issues: [
                    [
                        key: 'issue1',
                        fields: [
                            summary: 'Issue Summary 1',
                            status: [
                                name: 'closed'
                            ],
                            issuetype: [
                                name: 'TypeName 1'
                            ],
                            creator: [
                                displayName: 'TestCreator 1'
                            ],
                            assignee: [
                                displayName: 'TestAssignee 1'
                            ]
                        ]
                    ],
                    [
                        key: 'issue2',
                        fields: [
                            summary: 'Issue Summary 2',
                            status: [
                                name: 'closed'
                            ],
                            issuetype: [
                                name: 'TypeName 2'
                            ],
                            creator: [
                                displayName: 'TestCreator 2'
                            ],
                            assignee: [
                                displayName: 'TestAssignee 2'
                            ]
                        ]
                    ]
                ],

            ]

        }
    }

    @Test
    void testWriteIssuesJql() {

        def jiraMock = new JiraMock()

        jsr.step.sapJiraWriteIssueList([
            script: nullScript,
            jiraCredentialsId: 'JiraId',
            jql: 'testJQL',
            juStabJira: jiraMock
        ])

        assertTrue(jwfr.files.get('issueList.html').contains('<title>Jira Issue List</title>'))
        assertTrue(jwfr.files.get('issueList.html').contains('<h1>Jira Issue List</h1>'))
        assertTrue(jwfr.files.get('issueList.html').contains('<p>Search filter:</p><div class="code">testJQL</div>'))
        assertTrue(jwfr.files.get('issueList.html').contains('<p>Overall number of issues: 2</p>'))
        assertTrue(jwfr.files.get('issueList.html').contains('<tr><td>1</td><td><a href="https://jira.test/browse/issue1" target="_blank">issue1</a></td><td>TypeName 1</td><td>Issue Summary 1</td><td>closed</td><td>TestCreator 1</td><td>TestAssignee 1</td></tr>'))
        assertTrue(jwfr.files.get('issueList.html').contains('<tr><td>2</td><td><a href="https://jira.test/browse/issue2" target="_blank">issue2</a></td><td>TypeName 2</td><td>Issue Summary 2</td><td>closed</td><td>TestCreator 2</td><td>TestAssignee 2</td></tr>'))

        assertJobStatusSuccess()
    }

    @Test
    void testWriteIssuesFilter() {

        def httpRequestParams = [:]

        helper.registerAllowedMethod("httpRequest", [Map.class], {params ->
            httpRequestParams = params
            return [status : 200, content: '']
        })

        helper.registerAllowedMethod("readJSON", [Map.class], {params ->
            return [jql : 'TestJQL2']
        })

        def jiraMock = new JiraMock()

        jsr.step.sapJiraWriteIssueList([
            script: nullScript,
            jiraCredentialsId: 'JiraId',
            jiraFilterId: 'testId',
            juStabJira: jiraMock
        ])

        assertTrue(jwfr.files.get('issueList.html').contains('<p>Search filter:</p><div class="code">TestJQL2</div>'))
        assertTrue(jwfr.files.get('issueList.html').contains('<p>Overall number of issues: 2</p>'))
        assertTrue(jwfr.files.get('issueList.html').contains('<tr><td>1</td><td><a href="https://jira.test/browse/issue1" target="_blank">issue1</a></td><td>TypeName 1</td><td>Issue Summary 1</td><td>closed</td><td>TestCreator 1</td><td>TestAssignee 1</td></tr>'))
        assertTrue(jwfr.files.get('issueList.html').contains('<tr><td>2</td><td><a href="https://jira.test/browse/issue2" target="_blank">issue2</a></td><td>TypeName 2</td><td>Issue Summary 2</td><td>closed</td><td>TestCreator 2</td><td>TestAssignee 2</td></tr>'))

        assertJobStatusSuccess()
    }
}

package steps

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.startsWith
import static org.hamcrest.Matchers.arrayWithSize
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.hasItem

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import util.BasePiperTest

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertNull
import static org.junit.Assert.assertTrue
import static org.junit.Assert.assertThat
import org.junit.rules.RuleChain

import util.Rules
import util.JenkinsLoggingRule
import util.JenkinsStepRule
import util.JenkinsShellCallRule

class SendNotificationMSTeamsTest extends BasePiperTest {
    def teamsCallMap = [:]

    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jscr)
        .around(jsr)

    @Before
    void init() throws Exception {
        // register Jenkins commands with mock values
        helper.registerAllowedMethod("office365ConnectorSend", [Map.class], {m -> teamsCallMap = m})
        
        helper.registerAllowedMethod("deleteDir", [], null)
        helper.registerAllowedMethod("sshagent", [Map.class, Closure.class], null)

        nullScript.globalPipelineEnvironment.configuration = nullScript.globalPipelineEnvironment.configuration ?: [:]
        nullScript.globalPipelineEnvironment.configuration['general'] = nullScript.globalPipelineEnvironment.configuration['general'] ?: [:]
        nullScript.globalPipelineEnvironment.configuration['steps'] = nullScript.globalPipelineEnvironment.configuration['steps'] ?: [:]
        nullScript.globalPipelineEnvironment.configuration['steps']['sendNotificationMail'] = nullScript.globalPipelineEnvironment.configuration['steps']['sendNotificationMSTeams'] ?: [:]

        helper.registerAllowedMethod('requestor', [], { -> return [$class: 'RequesterRecipientProvider']})

    }

    @Test
    void testNotificationBuildSuccess() throws Exception {
        nullScript.currentBuild = [
        		result: 'SUCCESS'
        ]

        jsr.step.sendNotificationMSTeams(script: nullScript, webhookUrl: 'http://office.url', notifyCulprits: false)
        // asserts
        assertThat(teamsCallMap.message.toString(), startsWith('SUCCESS: Job p <http://build.url>'))
        assertEquals('Color not set correctly', '#008000', teamsCallMap.color)
        assertEquals('WebHook not set', 'http://office.url', teamsCallMap.webhookUrl)
        assertJobStatusSuccess()
    }

    @Test
    void testNotificationBuildFailed() throws Exception {
        nullScript.currentBuild = [
        		result: 'FAILURE'
        ]
        jsr.step.sendNotificationMSTeams(script: nullScript, webhookUrl: 'http://office.url', notifyCulprits: false)
        // asserts
        assertThat(teamsCallMap.message.toString(), startsWith('FAILURE: Job p <http://build.url>'))
        assertEquals('Color not set correctly', '#E60000', teamsCallMap.color)
    }

    @Test
    void testNotificationBuildStatusNull() throws Exception {
        nullScript.currentBuild = [:]
        jsr.step.sendNotificationMSTeams(script: nullScript, webhookUrl: 'http://office.url', notifyCulprits: false)
        // asserts
        assertTrue('Missing build status not detected', jlr.log.contains('currentBuild.result is not set. Skipping MS Teams notification'))
        assertJobStatusSuccess()
    }

    @Test
    void testNotificationCustomMessageAndColor() throws Exception {
        nullScript.currentBuild = [
        		result: 'FAILURE'
        ]
        jsr.step.sendNotificationMSTeams(script: nullScript, webhookUrl: 'http://office.url', message: 'Custom Message', color: '#AAAAAA', notifyCulprits: false)
        // asserts
        assertEquals('Custom message not set correctly', 'Custom Message', teamsCallMap.message.toString())
        assertEquals('Custom color not set correctly', '#AAAAAA', teamsCallMap.color)
        assertJobStatusSuccess()
    }

    @Test
    void testNotificationWithOutWebHookUrl() throws Exception {
        nullScript.currentBuild = [
        		result: 'SUCCESS'
        ]
        jsr.step.sendNotificationMSTeams(script: nullScript, notifyCulprits: false)
        assertTrue('Missing webhookUrl not detected', jlr.log.contains('webhookUrl is not set. Skipping MS Teams notification'))
        assertJobStatusSuccess()
    }

		// Copied from SendNotificationMailTest.groovy
   @Test
    void testGetDistinctRecipients() throws Exception {
        // git log -10 --pretty=format:"%ae %ce"
        def input = '''s.merk@sap.com noreply+github@sap.corp
s.merk@sap.com noreply+github@sap.corp
Christopher.Fenner@sap.com noreply+github@sap.corp
oliver.nocon@sap.com s.merk@sap.com
s.merk@sap.com noreply+github@sap.corp
Christopher.Fenner@sap.com noreply+github@sap.corp
Christopher.Fenner@sap.com noreply+github@sap.corp
Christopher.Fenner@sap.com noreply+github@sap.corp
Christopher.Fenner@sap.com noreply+github@sap.corp
Christopher.Fenner@sap.com noreply+github@sap.corp'''

        def result = jsr.step.getDistinctRecipients(input)
        // asserts
        assertThat(result.split('<br/>'), arrayWithSize(3))
        assertThat(result, containsString('s.merk@sap.com'))
        assertThat(result, containsString('oliver.nocon@sap.com'))
        assertThat(result, containsString('Christopher.Fenner@sap.com'))
    }

    @Test
    void testCulpritsFromGitCommit() throws Exception {
        def gitCommand = "git log -2 --pretty=format:'%ae %ce'"
        def expected = "oliver.nocon@sap.com Christopher.Fenner@sap.com"

        jscr.setReturnValue("git log -2 --pretty=format:'%ae %ce'", 'oliver.nocon@sap.com Christopher.Fenner@sap.com')

        def result = jsr.step.getCulprits(
            [
                gitSSHCredentialsId: '',
                gitUrl: 'git@github.wdf.sap.corp:IndustryCloudFoundation/pipeline-test-node.git',
                gitCommitId: 'f0973368a35a2b973612acb86f932c61f2635f6e'
            ],
            'master',
            2)
        println("LOGS: ${jlr.log}")
        println("RESULT: ${result}")
        // asserts
        assertThat(result, containsString('oliver.nocon@sap.com'))
        assertThat(result, containsString('Christopher.Fenner@sap.com'))
    }

    @Test
    void testCulpritsWithEmptyGitCommit() throws Exception {

        jscr.setReturnStatus('git log > /dev/null 2>&1',1)

        jsr.step.getCulprits(
            [
                gitSSHCredentialsId: '',
                gitUrl: 'git@github.wdf.sap.corp:IndustryCloudFoundation/pipeline-test-node.git',
                gitCommitId: ''
            ],
            'master',
            2)
        // asserts
        assertThat(jlr.log, containsString('[sendNotificationMSTeams] No git context available to retrieve culprits'))
    }

    @Test
    void testCulpritsWithoutGitCommit() throws Exception {

        jscr.setReturnStatus('git log > /dev/null 2>&1',1)

        jsr.step.getCulprits(
            [
                gitSSHCredentialsId: '',
                gitUrl: 'git@github.wdf.sap.corp:IndustryCloudFoundation/pipeline-test-node.git',
                gitCommitId: null
            ],
            'master',
            2)
        // asserts
        assertThat(jlr.log, containsString('[sendNotificationMSTeams] No git context available to retrieve culprits'))
    }

    @Test
    void testCulpritsWithoutBranch() throws Exception {

        jscr.setReturnStatus('git log > /dev/null 2>&1',1)

        jsr.step.getCulprits(
            [
                gitSSHCredentialsId: '',
                gitUrl: 'git@github.wdf.sap.corp:IndustryCloudFoundation/pipeline-test-node.git',
                gitCommitId: ''
            ],
            null,
            2)
        // asserts
        assertThat(jlr.log, containsString('[sendNotificationMSTeams] No git context available to retrieve culprits'))
    }		    
}

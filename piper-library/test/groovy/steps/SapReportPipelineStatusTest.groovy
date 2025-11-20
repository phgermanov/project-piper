#!groovy
package steps

import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.MockHelper
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class SapReportPipelineStatusTest extends BasePiperTest{

    private JenkinsStepRule stepRule = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(stepRule)

    @Test
    void testSapReportPipelineStatus() {
        boolean sapReportPipelineStatusCalled
        helper.registerAllowedMethod("sapReportPipelineStatus", [Map], { Map m ->
                sapReportPipelineStatusCalled = true
         })
         stepRule.step.sapReportPipelineStatus(
                     script: nullScript,
                 )
        assertThat(sapReportPipelineStatusCalled, is(true))
    }

    @Test
    void testErrorCategory() {
        List tt = [
            [stepName: '', reason: '', expectedCategory: 'undefined'],
            [stepName: 'mavenExecute', reason: '', expectedCategory: 'build'],
            [stepName: '', reason: 'The execution of the voting build failed', expectedCategory: 'build'],
            [stepName: '', reason: 'NO VALUE AVAILABLE FOR', expectedCategory: 'config'],
            [stepName: '', reason: 'FOSS to PPMS Entry mappings are missing', expectedCategory: 'compliance'],
            [stepName: 'callOnboarding', reason: '', expectedCategory: 'custom'],
            [stepName: 'Central Build (extended)', reason: '', expectedCategory: 'custom'],
            [stepName: '', reason: 'org.jenkinsci.plugins.workflow.steps.FlowInterruptedException', expectedCategory: 'custom'],
            [stepName: 'checksPublishResults', reason: '', expectedCategory: 'test'],
            [stepName: '', reason: 'The execution of the karma tests failed', expectedCategory: 'test'],
            [stepName: '', reason: 'Step execution failed. Error: Download of sonar-scanner\' failed: HTTP GET request: connect: connection refused', expectedCategory: 'service'],
        ]

        tt.each {test ->
            def result = stepRule.step.sapReportPipelineStatus.getErrorCategory(test.stepName, test.reason)
            assertThat(result, is(test.expectedCategory))
        }
    }

    @Test
    void testFilterDynamicMessages() {
        List tt = [
            [reason: 'Error', expectedReason: 'Error'],
            [reason: 'ERROR - NO VALUE AVAILABLE FOR gitCommitId', expectedReason: 'ERROR - NO VALUE AVAILABLE'],
            [reason: 'java.io.FileNotFoundException: master/workspace@2/pom.xml does not exist.', expectedReason: 'java.io.FileNotFoundException'],
            [reason: 'java.lang.NullPointerException: Cannot get property tenantId on null object', expectedReason: 'Cannot get property  on null object'],
            [reason: 'org.jenkinsci.plugins.scriptsecurity.sandbox.RejectedAccessException: No such field found: field java.lang.String downloadArtifactsFromNexus', expectedReason: 'RejectedAccessException: No such field found'],
            [reason: '[sonarExecuteScan] Step execution failed. Error: running command \'/home/jenkins/agent/workspace/<pipeline>/.sonar-scanner/bin/sonar-scanner\' failed: cmd.Run() failed: exit status 1', expectedReason: '[sonarExecuteScan] Step execution failed. Error: running command sonar-scanner failed: cmd.Run() failed: exit status'],
            [reason: '[sonarExecuteScan] Step execution failed (category: infrastructure). Error: running command /home/jenkins/agent/workspace/<pipeline>/.sonar-scanner/bin/sonar-scanner failed: cmd.Run() failed: exit status 1', expectedReason: '[sonarExecuteScan] Step execution failed (category: infrastructure). Error: running command sonar-scanner failed: cmd.Run() failed: exit status 1'],
            [reason: '[artifactPrepareVersion] Step execution failed (category: undefined). Error: failed to push changes for version 0.0.1-123456: ssh: handshake failed: ssh: unable to authenticate, attempted methods [none publickey], no supported methods remain', expectedReason: '[artifactPrepareVersion] Step execution failed (category: ). Error: failed to push changes: ssh: handshake failed'],
            [reason: '[artifactPrepareVersion] Step execution failed (category: undefined). Error: failed to push changes for version 1.0.0-123456: knownhosts: /var/jenkins_home/.ssh/known_hosts:6: illegal base64 data at input byte 8', expectedReason: '[artifactPrepareVersion] Step execution failed (category: ). Error: failed to push changes: knownhosts: illegal base64 data at input byte 8']
        ]

        tt.each {test ->
            def result = stepRule.step.sapReportPipelineStatus.filterDynamicMessages(test.reason)
            assertThat(result, is(test.expectedReason))
        }
    }

    @Test
    void testWildcardRuleMatch() {
        List tt = [
            [rule: '', reason: '', expectedResult: true],
            [rule: 'no wildcard', reason: 'test', expectedResult: false],
            [rule: 'no wildcard', reason: 'test no wildcard match', expectedResult: true],
            [rule: 'has * wildcard', reason: 'test wildcard no match', expectedResult: false],
            [rule: 'has * wildcard', reason: 'test has one wildcard match', expectedResult: true],
            [rule: 'has * wildcard * version', reason: 'test has multiple wildcard e.g. for version', expectedResult: true],
        ]

        tt.each {test ->
            def result = stepRule.step.sapReportPipelineStatus.wildcardRuleMatch(test.rule, test.reason)
            assertThat(result, is(test.expectedResult))
        }
    }
}

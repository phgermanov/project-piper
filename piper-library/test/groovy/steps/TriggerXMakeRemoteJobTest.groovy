#!groovy
package steps

import hudson.AbortException

import org.junit.Before
import org.junit.Ignore
import org.junit.Test
import org.junit.Rule
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest

import static org.hamcrest.Matchers.containsString

import util.Rules
import util.JenkinsStepRule
import util.JenkinsLoggingRule

@Ignore("step disabled")
class TriggerXMakeRemoteJobTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jsr)
        .around(thrown)

    @Test
    void testXMakeRemoteCallNoHostFound() throws Exception {
        helper.registerAllowedMethod("httpRequest", [String.class], { s ->
            return ['status': 404,'content': '{"status":"404"}']
        })
        thrown.expect(AbortException)
        thrown.expectMessage(containsString('xMake Host not found!'))

        jsr.step.triggerXMakeRemoteJob(
            script: nullScript,
            juStabUtils : utils,
            xMakeServer: '',
            xMakeJobName: 'NAAS-pipeline-test-SP-REL-common_indirectshipment',
            xMakeJobParameters: 'MODE=stage'
        )
        println("LOG: ${jlr.log}")
        // asserts
        assertJobStatusFailure()
    }

    @Test
    void testXMakeRemoteCallJobNotFound() throws Exception {
        helper.registerAllowedMethod("httpRequest", [Map], { s ->
            return ['status': 200, 'content': '<com.sap.prd.jenkins.plugins.jobfinder.JobFinder/>']
        })

        thrown.expect(AbortException)
        thrown.expectMessage(containsString('No xMake job found!'))

        jsr.step.triggerXMakeRemoteJob(
            script: nullScript,
            juStabUtils : utils,
            xMakeServer: '',
            xMakeJobName: 'NAAS-pipeline-test-SP-REL-common_indirectshipment',
            xMakeJobParameters: 'MODE=stage'
        )
        println("LOG: ${jlr.log}")
        // asserts
        assertJobStatusFailure()
    }
}

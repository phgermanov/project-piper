#!groovy
package steps

import org.jenkinsci.plugins.workflow.steps.FlowInterruptedException
import org.junit.Ignore
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsErrorRule
import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.Rules

import static org.hamcrest.CoreMatchers.containsString
import static org.hamcrest.CoreMatchers.hasItem
import static org.hamcrest.CoreMatchers.is
import static org.hamcrest.CoreMatchers.isA
import static org.junit.Assert.assertThat

@Ignore("step disabled")
class RestartableStepsTest extends BasePiperTest {

    private JenkinsErrorRule jer = new JenkinsErrorRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain chain = Rules.getCommonRules(this)
        .around(jer)
        .around(jlr)
        .around(jsr)

    @Test
    void testError() throws Exception {

        def mailBuildResult = ''
        helper.registerAllowedMethod('sendNotificationMail', [Map.class], { m ->
            mailBuildResult = m.buildResult
            return null
        })

        helper.registerAllowedMethod('timeout', [Map.class, Closure.class], { m, closure ->
            assertThat(m.time, is(1))
            assertThat(m.unit, is('SECONDS'))
            return closure()
        })

        def iterations = 0
        helper.registerAllowedMethod('input', [Map.class], { m ->
            iterations ++
            assertThat(m.message, is('Do you want to restart?'))
            assertThat(m.ok, is('Restart'))
            if (iterations > 1) {
                throw new FlowInterruptedException()
            } else {
                return null
            }
        })

        try {
            jsr.step.restartableSteps ([
                script: nullScript,
                jenkinsUtilsStub: jenkinsUtils,
                sendMail: true,
                timeoutInSeconds: 1

            ]) {
                throw new hudson.AbortException('I just created an error')
            }
        } catch(err) {
            assertThat(jlr.log, containsString('ERROR occured: hudson.AbortException: I just created an error'))
            assertThat(mailBuildResult, is('UNSTABLE'))
        }
    }

    @Test
    void testErrorNoMail() throws Exception {

        def mailBuildResult = ''
        helper.registerAllowedMethod('sendNotificationMail', [Map.class], { m ->
            mailBuildResult = m.buildResult
            return null
        })

        helper.registerAllowedMethod('timeout', [Map.class, Closure.class], { m, closure ->
            assertThat(m.time, is(1))
            assertThat(m.unit, is('SECONDS'))
            return closure()
        })

        def iterations = 0
        helper.registerAllowedMethod('input', [Map.class], { m ->
            iterations ++
            assertThat(m.message, is('Do you want to restart?'))
            assertThat(m.ok, is('Restart'))
            if (iterations > 1) {
                throw new FlowInterruptedException()
            } else {
                return null
            }
        })

        try {
            jsr.step.restartableSteps ([
                script: nullScript,
                jenkinsUtilsStub: jenkinsUtils,
                sendMail: false,
                timeoutInSeconds: 1

            ]) {
                throw new hudson.AbortException('I just created an error')
            }
        } catch(err) {
            assertThat(jlr.log, containsString('ERROR occured: hudson.AbortException: I just created an error'))
            assertThat(mailBuildResult, is(''))
        }
    }

    @Test
    void testSuccess() throws Exception {

        jsr.step.restartableSteps ([
            script: nullScript,
            jenkinsUtilsStub: jenkinsUtils,
            sendMail: false,
            timeoutInSeconds: 1

        ]) {
            nullScript.echo 'This is a test'
        }

        assertThat(jlr.log, containsString('This is a test'))
    }
}

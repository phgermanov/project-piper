#!groovy
package steps

import org.junit.Before
import org.junit.Ignore
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsEnvironmentRule
import util.JenkinsLoggingRule
import util.JenkinsStepRule
import util.Rules

import static com.lesfurets.jenkins.unit.MethodCall.callArgsToString
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertTrue
import static org.junit.Assert.assertFalse

@Ignore("step disabled")
class HandleStepErrorsTest extends BasePiperTest {

    private ExpectedException thrown = ExpectedException.none()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)


    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jlr)
        .around(jsr) // needs to be activated after jedr, otherwise executeDocker is not mocked

    @Before
    void init() throws Throwable {

    }

    @Test
    void testHandeleErrorsNoError() {
        def bodyExecuted = false
        jsr.step.handleStepErrors([
            stepName: 'test',
            stepParameters: [jenkinsUtilsStub: jenkinsUtils]
        ]) {
            bodyExecuted = true
        }

        assertEquals(true, bodyExecuted)
        assertTrue(jlr.log.contains('--- BEGIN LIBRARY STEP: test.groovy'))
        assertTrue(jlr.log.contains('--- END LIBRARY STEP: test.groovy'))
    }

    @Test
    void testHandleErrorsPreviousFailure() {
        binding.setVariable('currentBuild', [
            result: 'FAILURE'
        ])
        thrown.expect(Exception)
        jsr.step.handleStepErrors([
            stepName: 'test',
            stepParameters: [jenkinsUtilsStub: jenkinsUtils]
        ]) {
            def bodyExecuted = false
        }
    }

    @Test
    @Ignore("step has been deprecated")
    void testHandleErrorsDefault() {
        def errorReported = false
        try {
            jsr.step.handleStepErrors([
                stepName: 'test',
                stepParameters: [jenkinsUtilsStub: jenkinsUtils]
            ]) {
                throw new Exception('TestError')
            }
        } catch (err) {
            errorReported = true
        }

        assertTrue(errorReported)
        assertTrue(jlr.log.contains('ERROR OCCURED IN LIBRARY STEP: test'))
    }

    @Test
    @Ignore("step has been deprecated")
    void testHandleErrorsCustomDocumentation() {
        try {
            jsr.step.handleStepErrors([
                stepName: 'test',
                stepNameDoc: 'customDoc',
                stepParameters: [jenkinsUtilsStub: jenkinsUtils]
            ]) {
                throw new Exception('TestError')
            }
        } catch (err) {
            //do nothing
        }
        assertTrue(jlr.log.contains('Documentation of step test: https://go.sap.corp/piper/steps/customDoc/'))
    }

    @Test
    void testHandleErrorsDoNotPrintErrorDetails() {
        try {
            jsr.step.handleStepErrors([
                echoDetails: false,
                stepName: 'test',
                stepParameters: [jenkinsUtilsStub: jenkinsUtils]
            ]) {
                throw new Exception('TestError')
            }
        } catch (err) {
            //do nothing
        }
        assertFalse(jlr.log.contains('--- BEGIN LIBRARY STEP: test.groovy'))
        assertFalse(jlr.log.contains('--- END LIBRARY STEP: test.groovy'))
        assertFalse(jlr.log.contains('ERROR OCCURED IN LIBRARY STEP: test'))
    }
}

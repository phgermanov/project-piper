#!groovy
package stages

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsStepRule
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class SapPiperStageAdditionalUnitTestsTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private ExpectedException thrown = new ExpectedException()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jsr)

    private List stepsCalled = []

    @Before
    void init()  {

        binding.variables.env.STAGE_NAME = 'Additional Unit Tests'

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Additional Unit Tests'))
            return body()
        })

        helper.registerAllowedMethod('durationMeasure', [Map.class, Closure.class], {m, body ->
            return body()
        })

        helper.registerAllowedMethod('karmaExecuteTests', [Map.class], {m ->
            stepsCalled.add('karmaExecuteTests')
        })

        helper.registerAllowedMethod('npmExecuteScripts', [Map.class], {m ->
            stepsCalled.add('npmExecuteScripts')
        })

        helper.registerAllowedMethod('testsPublishResults', [Map.class], {m ->
            stepsCalled.add('testsPublishResults')
        })
    }

    @Test
    void testAdditionalUnitTestsDefault() {

        jsr.step.sapPiperStageAdditionalUnitTests(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, not(hasItems('karmaExecuteTests', 'npmExecuteScripts', 'testsPublishResults')))
    }

    @Test
    void testAdditionalUnitTestsithKarmaConfig() {

        nullScript.globalPipelineEnvironment.configuration = [runStep: ['Additional Unit Tests': [karmaExecuteTests: true]]]

        jsr.step.sapPiperStageAdditionalUnitTests(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItems('karmaExecuteTests', 'testsPublishResults'))
    }

    @Test
    void testAdditionalUnitTestsOverwriteDefault() {

        jsr.step.sapPiperStageAdditionalUnitTests(script: nullScript, juStabUtils: utils, karmaExecuteTests: false)

        assertThat(stepsCalled, not(hasItems('karmaExecuteTests', 'testsPublishResults')))
    }

    @Test
    void testAdditionalUnitTestsWithNpm() {

        nullScript.globalPipelineEnvironment.configuration = [runStep: ['Additional Unit Tests': [npmExecuteScripts: true]]]

        jsr.step.sapPiperStageAdditionalUnitTests(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItems('npmExecuteScripts', 'testsPublishResults'))
    }
}

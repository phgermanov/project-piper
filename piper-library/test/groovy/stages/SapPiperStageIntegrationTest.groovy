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

class SapPiperStageIntegrationTest extends BasePiperTest {
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

        binding.variables.env.STAGE_NAME = 'Integration'

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Integration'))
            return body()
        })

        helper.registerAllowedMethod('durationMeasure', [Map.class, Closure.class], {m, body ->
            return body()
        })

        helper.registerAllowedMethod('executeCustomIntegrationTests', [Map.class], {m ->
            stepsCalled.add('executeCustomIntegrationTests')
        })
    }

    @Test
    void testIntegrationDefault() {

        jsr.step.sapPiperStageIntegration(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, not(hasItem('executeCustomIntegrationTests')))
    }

    @Test
    void testIntegrationOverwriteDefault() {

        jsr.step.sapPiperStageIntegration(script: nullScript, juStabUtils: utils, executeCustomIntegrationTests: true)

        assertThat(stepsCalled, hasItem('executeCustomIntegrationTests'))
    }

    @Test
    void testIntegration() {

        nullScript.globalPipelineEnvironment.configuration = [runStep: ['Integration': [executeCustomIntegrationTests: true]]]
        jsr.step.sapPiperStageIntegration(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItem('executeCustomIntegrationTests'))
    }
}

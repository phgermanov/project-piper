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

class SapPiperStagePromoteTest extends BasePiperTest {
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
        binding.variables.env.STAGE_NAME = 'Promote'

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Promote'))
            return body()
        })

        helper.registerAllowedMethod('durationMeasure', [Map.class, Closure.class], {m, body ->
            return body()
        })

        helper.registerAllowedMethod('executeBuild', [Map.class], {m ->
            stepsCalled.add('executeBuild')
        })

        helper.registerAllowedMethod('sapCallStagingService', [Map.class], {m ->
            stepsCalled.add('sapCallStagingService')
        })

        helper.registerAllowedMethod('sapXmakeExecuteBuild', [Map.class], {m ->
            stepsCalled.add('sapXmakeExecuteBuild')
        })
    }

    @Test
    void testPromoteDefault() {
        jsr.step.sapPiperStagePromote(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, not(hasItems('executeBuild')))
        assertThat(stepsCalled, not(hasItems('sapCallStagingService')))
        assertThat(stepsCalled, hasItems('sapXmakeExecuteBuild'))
    }

    @Test
    void testPromoteNativeBuild() {
        jsr.step.sapPiperStagePromote(script: nullScript, juStabUtils: utils, nativeBuild: true)

        assertThat(stepsCalled, not(hasItems('executeBuild')))
        assertThat(stepsCalled, hasItems('sapCallStagingService'))
        assertThat(stepsCalled, not(hasItems('sapXmakeExecuteBuild')))
    }
}

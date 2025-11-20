#!groovy
import com.lesfurets.jenkins.unit.BasePipelineTest

import static org.hamcrest.Matchers.*
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import static org.junit.Assert.assertThat
import org.junit.rules.RuleChain

import util.Rules
import util.JenkinsStepRule

class StepTestTemplateTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jsr)

    @Before
    void init() throws Exception {
    }

    @Test
    void testStepTestTemplate() throws Exception {
        //ToDo: replace call with step name
        jsr.step.call()
        // asserts
        assertThat('xyz', is('xyz'))
        assertThat(true, is(true))
        assertJobStatusSuccess()
    }
}

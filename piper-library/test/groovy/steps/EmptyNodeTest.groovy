#!groovy
package steps

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import util.BasePiperTest

import static org.junit.Assert.assertTrue
import org.junit.rules.RuleChain

import util.Rules
import util.JenkinsStepRule

class EmptyNodeTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jsr)

    @Before
    void init() throws Exception {
        // register Jenkins commands with mock values
        helper.registerAllowedMethod( "deleteDir", [], null )
    }

    @Test
    void testEmptyNode() throws Exception {
        def bodyExecuted = false
        jsr.step.emptyNode() {
            bodyExecuted = true
        }

        assertTrue(bodyExecuted)
        assertJobStatusSuccess()
    }
}

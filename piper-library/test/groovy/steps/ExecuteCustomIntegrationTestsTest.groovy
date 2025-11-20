package steps

import com.sap.icd.jenkins.Utils
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.containsString
import static org.junit.Assert.assertThat

class ExecuteCustomIntegrationTestsTest extends BasePiperTest {

    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsScriptLoaderRule jslr = new JenkinsScriptLoaderRule(this, 'test/resources/customIntegrationTest/.pipeline')

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jslr)
        .around(jsr) // needs to be activated after jedr, otherwise executeDocker is not mocked

    @Test
    void testRunExecuteCustomIntegrationTest() {

        helper.registerAllowedMethod('durationMeasure', [Map.class, Closure.class], {m, body ->
            return body()
        })

        def utils = new Utils()

        prepareObjectInterceptors(utils)

        jsr.step.executeCustomIntegrationTests(script: nullScript, juStabUtils: utils, extensionIntegrationTestScript: 'integration.groovy')

        assertJobStatusSuccess()

        assertThat(jlr.log, containsString("Integration Test"))
        assertThat(jlr.log, containsString("Test var: null"))
    }
}

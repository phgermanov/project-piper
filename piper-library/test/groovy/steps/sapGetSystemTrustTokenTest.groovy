package steps

import org.codehaus.groovy.GroovyException
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue
import static org.junit.Assert.fail

class CallFunctionTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain rules = RuleChain.outerRule(jlr).around(jscr).around(jsr)

    @Before
    void init() throws Exception {
        helper.registerAllowedMethod("sh", [Map.class], (Map map) -> {
            if (map.script.contains("trust_api_request.sh")) {
                return "HTTPSTATUS:200{\"token\":\"expectedToken\"}"
            } else if (map.script.contains("cat")) {
                return "mockGCPtoken"
            }
            return ""
        })
        helper.registerAllowedMethod("echo", [String.class], { String _ -> })
    }

    @Test
    void testSuccessfulTokenRetrieval() {
        def result = jsr.step.call('validCredsId', 'validPipelineId', 'validGroupId')
        assertThat(result, equalTo("expectedToken"))
    }

    @Test(expected = IllegalArgumentException.class)
    void testCallWithMissingParameters() {
        jsr.step.call(null, 'validPipelineId', 'validGroupId')
    }

    @Test(expected = RuntimeException.class)
    void testErrorHandlingOnBadApiResponse() {
        helper.registerAllowedMethod("sh", [Map.class], (Map map) -> {
            return "HTTPSTATUS:400{\"message\":\"Bad Request\"}"
        })
        
        try {
            jsr.step.call('validCredsId', 'validPipelineId', 'validGroupId')
        } catch (RuntimeException e) {
            assertThat(e.getMessage(), containsString("Failed to obtain System Trust session token"))
            throw e
        }
    }

    @Test
    void testReadGcpToken() {
        def token = jsr.step.readGcpToken()
        assertThat(token, is("mockGCPtoken"))
    }

    @Test
    void testFailureOnReadingGcpToken() {
        helper.registerAllowedMethod("sh", [Map.class], { Map _ ->
            throw new GroovyException("Failed to read GCP token")
        })
        
        try {
            jsr.step.readGcpToken()
            fail("Should have thrown an exception for missing GCP token.")
        } catch (RuntimeException e) {
            assertThat(e.getMessage(), containsString("Failed to read GCP token"))
        }
    }
}
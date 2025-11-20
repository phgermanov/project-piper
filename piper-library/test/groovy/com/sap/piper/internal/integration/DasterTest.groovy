package com.sap.piper.internal.integration

import com.sap.icd.jenkins.Utils
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.Rules

import static org.hamcrest.Matchers.is
import static org.junit.Assert.assertThat

class DasterTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(thrown)

    @Test
    void testCallApiTwice() {

        def counter = 0
        helper.registerAllowedMethod('httpRequest', [Map.class], {
            def result = [status: 200, content: '{"test": "abcd"}']
            if (counter == 0) {
                result = [status: 408]
            }
            counter++
            return result
        })

        Daster myDaster = new Daster(nullScript, new Utils(), [serviceUrl: 'https://bcd.def', maxRetries: 3, scanType: 'basicScan', settings: [:]])

        def result = myDaster.triggerScan()

        assertThat(counter, is(2))
        assertThat(result, is([test: 'abcd']))
    }

    @Test
    void testCallApiTillEnd() {

        def counter = 0
        helper.registerAllowedMethod('httpRequest', [Map.class], {
            def result = [status: 404, content: '{}']
            counter++
            return result
        })

        Daster myDaster = new Daster(nullScript, new Utils(), [serviceUrl: 'https://bcd.def', maxRetries: 3, scanType: 'basicScan', settings: [:]])

        def result = myDaster.triggerScan()

        assertThat(counter, is(3))
        assertThat(result, is([:]))
    }
}

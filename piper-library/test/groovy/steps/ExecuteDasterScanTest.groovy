#!groovy

package steps

import hudson.AbortException
import org.junit.Before
import org.junit.Ignore
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat
import static org.junit.Assert.fail

class ExecuteDasterScanTest extends BasePiperTest {
    ExpectedException thrown = ExpectedException.none()
    private JenkinsErrorRule jer = new JenkinsErrorRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    def deleteCalled = false, statusFetchCalled = false, stopCalled = false
    def env = []

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jer)
        .around(jlr)
        .around(jscr)
        .around(jsr) // needs to be activated after jedr, otherwise executeDocker is not mocked

    @Before
    void init() throws Exception {
        def credentialsStore = ['ID-abcdefg': ['testtoken287652765'], 'daster_app_usercredentials': ['someusername&somepassword']]
        def withCredentialsBindings
        helper.registerAllowedMethod('withEnv', [List, Closure], {
            l, c ->
                env.addAll(l)
                c()
        })
        helper.registerAllowedMethod('string', [Map], {
            m ->
                withCredentialsBindings = ["${m.credentialsId}": ["${m.variable}"]]
                return m
        })
        helper.registerAllowedMethod('withCredentials', [List.class, Closure.class], {
            l, body ->
                def index = 0
                withCredentialsBindings.each {
                    entry ->
                        if(entry.value instanceof List) {
                            entry.value.each {
                                subEntry ->
                                    def value = credentialsStore[entry.key]
                                    getBinding().setProperty(subEntry, value[index])
                                    index++

                            }
                        } else {
                            getBinding().setProperty(entry.value, credentialsStore[entry.value])
                        }
                }
                try {
                    body()
                } finally {
                    withCredentialsBindings.each {
                        entry ->
                            if(entry.value instanceof List) {
                                entry.value.each {
                                    subEntry ->
                                        getBinding().setProperty(subEntry, null)

                                }
                            } else {
                                getBinding().setProperty(entry.value, null)
                            }
                    }
                }
        })

        helper.registerAllowedMethod('httpRequest', [Map.class], {
            m ->
                if(m.url == 'https://daster.tools.sap/basicScan') {
                    assertThat(utils.parseJsonSerializable(m.requestBody), is(utils.parseJsonSerializable("{\n" +
                        "  \"url\": \"https://test.cfapps.hana.ondemand.com/\",\n" +
                        "  \"dasterToken\": \"testtoken287652765\",\n" +
                        "}")))
                    return [content: '{\n' +
                        '    "message": "You can open the result in a Silverlight enabled browser",' +
                        '    "url": "https://daster.tools.sap/12345dasdas"' +
                        '}']
                }
                if(m.url.startsWith('https://daster.tools.sap/fioriDASTScan')) {
                    if(m.httpMode == 'POST') {
                        assertThat(utils.parseJsonSerializable(m.requestBody), is(utils.parseJsonSerializable("{\n" +
                            "    \"targetUrl\": \"https://test.cfapps.hana.ondemand.com/ui\",\n" +
                            "    \"recipients\": [\n" +
                            "        \"some-test@sap.com\"\n" +
                            "    ],\n" +
                            "    \"userCredentials\": \"someusername&somepassword\",\n" +
                            "    \"dasterToken\": \"testtoken287652765\"\n" +
                            "}")))
                        return [content: '{"scanId":"fiori-dast-1561528940394"}']
                    }
                    if(m.httpMode == 'GET' && m.url.endsWith('/fiori-dast-1561528940394')) {
                        statusFetchCalled = true
                        return [content: '{"state":{"terminated":{"exitCode":0,"reason":"","startedAt":"2019-06-26T09:16:37Z","finishedAt":"2019-06-26T09:24:32Z","containerID":"docker://b8613dabb9d91cca05e0c75454c62e157c3be8584a1022117fde5d5a4ea83126"}},"riskSummary":{"High":3,"Medium":5,"Low":6,"Informational":2},"riskReport":[\"Just a test\"]}']
                    }
                    if(m.httpMode == 'DELETE') {
                        deleteCalled = true
                        return [content: null]
                    }
                    fail("Unexpected invocation of httpRequest with map ${m}")
                }
                if(m.url.startsWith('https://daster.tools.sap/burpscan')) {
                    if(m.httpMode == 'POST' && m.url.endsWith('/burpscan')) {
                        assertThat(utils.parseJsonSerializable(m.requestBody), is(utils.parseJsonSerializable("{\n" +
                            "    \"targetUrl\": \"https://test.cfapps.hana.ondemand.com/ui\",\n" +
                            "    \"dasterToken\": \"testtoken287652765\"\n" +
                            "}")))
                        return [content: '{"scanId":"burp-1561528940394", "proxyURL":"https://192.168.0.1:8080"}']
                    }
                    if(m.httpMode == 'GET' && m.url.endsWith('/burp-1561528940394')) {
                        statusFetchCalled = true
                        return [content: '{"state":{"terminated":{"exitCode":0,"reason":"","startedAt":"2019-06-26T09:16:37Z","finishedAt":"2019-06-26T09:24:32Z","containerID":"docker://b8613dabb9d91cca05e0c75454c62e157c3be8584a1022117fde5d5a4ea83126"}},"riskSummary":{"High":3,"Medium":5,"Low":6,"Informational":2},"riskReport":[\"Just a test\"]}']
                    }
                    if(m.url.endsWith('/burp-1561528940394/stop')) {
                        stopCalled = true
                        return [content: '{}']
                    }
                    fail("Unexpected invocation of httpRequest with map ${m}")
                }
        })
    }

    @Test
    void testDasterBasicDefaults() throws Exception {
        def parameters = [juStabUtils: utils,
                          script     : nullScript,
                          settings: [
                              dasterTokenCredentialsId: 'ID-abcdefg',
                              url: 'https://test.cfapps.hana.ondemand.com/'
                          ]
        ]

        jsr.step.executeDasterScan(parameters)

        assertThat(jlr.log, containsString("You can open the result in a Silverlight enabled browser"))
        assertThat(jlr.log, containsString("https://daster.tools.sap/12345dasdas"))

        assertThat(statusFetchCalled, is(false))

        assertThat(deleteCalled, is(false))
    }

    @Test
    void testBurpDefaults() throws Exception {
        def parameters = [juStabUtils: utils,
                          script     : nullScript,
                          scanType   : 'burpscan',
                          settings   : [
                              dasterTokenCredentialsId: 'ID-abcdefg',
                              targetUrl: 'https://test.cfapps.hana.ondemand.com/ui'
                          ]
        ]

        jsr.step.executeDasterScan(parameters) {
            nullScript.sh 'runTest.sh'
        }

        assertThat(statusFetchCalled, is(true))
        assertThat(stopCalled, is(true))

        assertThat(env, hasItem('BURP_PROXY=https://192.168.0.1:8080'))
        assertThat(jscr.shell, hasItem('runTest.sh'))
    }

    @Test
    void testDasterFioriDASTDefaults() throws Exception {
        def parameters = [juStabUtils: utils,
                          script     : nullScript,
                          scanType   : 'fioriDASTScan',
                          settings   : [
                              dasterTokenCredentialsId: 'ID-abcdefg',
                              targetUrl: 'https://test.cfapps.hana.ondemand.com/ui',
                              userCredentialsCredentialsId: 'daster_app_usercredentials',
                              recipients: ['some-test@sap.com']
                          ]
        ]

        thrown.expect(AbortException)
        thrown.expectMessage('[executeDasterScan][ERROR] Threshold(s) [thresholds.fail.high:0] violated by findings \'[High:3, Low:6, Medium:5, Informational:2]\'')

        jsr.step.executeDasterScan(parameters)

        assertThat(statusFetchCalled, is(true))

        assertThat(deleteCalled, is(true))
    }

    @Test
    void testDasterFioriDASTAsynchronous() throws Exception {
        def parameters = [juStabUtils: utils,
                          script     : nullScript,
                          scanType   : 'fioriDASTScan',
                          synchronous: false,
                          settings   : [
                              dasterTokenCredentialsId: 'ID-abcdefg',
                              targetUrl: 'https://test.cfapps.hana.ondemand.com/ui',
                              userCredentialsCredentialsId: 'daster_app_usercredentials',
                              recipients: ['some-test@sap.com']
                          ]
        ]

        jsr.step.executeDasterScan(parameters)

        // If entering the result fetching and evaluation
        assertThat(jlr.log, containsString('''[executeDasterScan] Running scan of type fioriDASTScan with settings [dasterTokenCredentialsId:ID-abcdefg, targetUrl:https://test.cfapps.hana.ondemand.com/ui, userCredentialsCredentialsId:daster_app_usercredentials, recipients:[some-test@sap.com]]
[executeDasterScan][INFO] Triggered scan of type fioriDASTScan: fiori-dast-1561528940394 and waiting for it to complete'''))

        assertThat(statusFetchCalled, is(false))

        assertThat(deleteCalled, is(false))
    }

    @Test
    void testDasterFioriDASTNoDelete() throws Exception {
        def parameters = [juStabUtils  : utils,
                          script       : nullScript,
                          scanType     : 'fioriDASTScan',
                          deleteCalled : false,
                          settings     : [
                              dasterTokenCredentialsId: 'ID-abcdefg',
                              targetUrl: 'https://test.cfapps.hana.ondemand.com/ui',
                              userCredentialsCredentialsId: 'daster_app_usercredentials',
                              recipients: ['some-test@sap.com']
                          ]
        ]

        thrown.expect(AbortException)
        thrown.expectMessage('[executeDasterScan][ERROR] Threshold(s) [thresholds.fail.high:0] violated by findings \'[High:3, Low:6, Medium:5, Informational:2]\'')

        jsr.step.executeDasterScan(parameters)

        assertThat(statusFetchCalled, is(true))

        assertThat(deleteCalled, is(false))
    }

    @Test
    void testDasterFioriDASTFailMedium() throws Exception {
        def parameters = [juStabUtils: utils,
                          script     : nullScript,
                          scanType   : 'fioriDASTScan',
                          thresholds : [ fail: [ high: -1, medium: 0 ]],
                          settings   : [
                              dasterTokenCredentialsId: 'ID-abcdefg',
                              targetUrl: 'https://test.cfapps.hana.ondemand.com/ui',
                              userCredentialsCredentialsId: 'daster_app_usercredentials',
                              recipients: ['some-test@sap.com']
                          ]
        ]

        thrown.expect(AbortException)
        thrown.expectMessage('[executeDasterScan][ERROR] Threshold(s) [thresholds.fail.medium:0] violated by findings \'[High:3, Low:6, Medium:5, Informational:2]\'')

        jsr.step.executeDasterScan(parameters)
    }

    @Test
    void testDasterFioriDASTFailLow() throws Exception {
        def parameters = [juStabUtils: utils,
                          script     : nullScript,
                          scanType   : 'fioriDASTScan',
                          thresholds : [ fail: [ high: -1, low: 0 ]],
                          settings   : [
                              dasterTokenCredentialsId: 'ID-abcdefg',
                              targetUrl: 'https://test.cfapps.hana.ondemand.com/ui',
                              userCredentialsCredentialsId: 'daster_app_usercredentials',
                              recipients: ['some-test@sap.com']
                          ]
        ]

        thrown.expect(AbortException)
        thrown.expectMessage('[executeDasterScan][ERROR] Threshold(s) [thresholds.fail.low:0] violated by findings \'[High:3, Low:6, Medium:5, Informational:2]\'')

        jsr.step.executeDasterScan(parameters)
    }

    @Test
    void testDasterFioriDASTFailInfo() throws Exception {
        def parameters = [juStabUtils: utils,
                          script     : nullScript,
                          scanType   : 'fioriDASTScan',
                          thresholds : [ fail: [ high: -1, info: 0 ]],
                          settings   : [
                              dasterTokenCredentialsId: 'ID-abcdefg',
                              targetUrl: 'https://test.cfapps.hana.ondemand.com/ui',
                              userCredentialsCredentialsId: 'daster_app_usercredentials',
                              recipients: ['some-test@sap.com']
                          ]
        ]

        thrown.expect(AbortException)
        thrown.expectMessage('[executeDasterScan][ERROR] Threshold(s) [thresholds.fail.info:0] violated by findings \'[High:3, Low:6, Medium:5, Informational:2]\'')

        jsr.step.executeDasterScan(parameters)
    }

    @Test
    void testDasterFioriDASTFailAll() throws Exception {
        def parameters = [juStabUtils: utils,
                          script     : nullScript,
                          scanType   : 'fioriDASTScan',
                          thresholds : [ fail: [ high: -1, all: 0 ]],
                          settings   : [
                              dasterTokenCredentialsId: 'ID-abcdefg',
                              targetUrl: 'https://test.cfapps.hana.ondemand.com/ui',
                              userCredentialsCredentialsId: 'daster_app_usercredentials',
                              recipients: ['some-test@sap.com']
                          ]
        ]

        thrown.expect(AbortException)
        thrown.expectMessage('[executeDasterScan][ERROR] Threshold(s) [thresholds.fail.all:0] violated by findings \'[High:3, Low:6, Medium:5, Informational:2]\'')

        jsr.step.executeDasterScan(parameters)
    }

    @Test
    void testDasterFioriDASTNotFail() throws Exception {
        def parameters = [juStabUtils: utils,
                          script     : nullScript,
                          scanType   : 'fioriDASTScan',
                          thresholds : [ fail: [ high: -1 ] ],
                          settings   : [
                              dasterTokenCredentialsId: 'ID-abcdefg',
                              targetUrl: 'https://test.cfapps.hana.ondemand.com/ui',
                              userCredentialsCredentialsId: 'daster_app_usercredentials',
                              recipients: ['some-test@sap.com']
                          ]
        ]

        jsr.step.executeDasterScan(parameters)

        assertThat(deleteCalled, is(true))
    }
}

package com.sap.piper.internal

import org.junit.After

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not

import static org.junit.Assert.assertThat
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.Rules

class NotifyTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jscr)
        .around(thrown)

    def httpParams

    @Before
    void init() throws Exception {
        httpParams = null
        // prepare
        utils.env.JOB_NAME = 'testJob'
        utils.env.JOB_URL = 'https://test.jenkins.com/test/'
        utils.env.BUILD_URL = 'https://test.jenkins.com/test/123'
        Notify.instance = utils

        helper.registerAllowedMethod("timeout", [Map.class, Closure.class], { m,c ->
            c()
        })
        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            httpParams = m
        })
    }

    @After
    void cleanup() {
        utils.env.JOB_NAME = null
        utils.env.JOB_URL = null
        utils.env.BUILD_URL = null
    }

    @Test
    void testWarning() {
        // execute test
        Notify.warning(nullScript, "test message", "anyStep")
        // asserts
        assertThat(jlr.log, containsString('[WARNING] test message (piper-lib/anyStep)'))
    }

    @Test
    void testError() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[ERROR] test message (piper-lib/anyStep)')
        // execute test
        try{
            Notify.error(nullScript, "test message", "anyStep")
        }finally{
            // asserts
        }
    }
}

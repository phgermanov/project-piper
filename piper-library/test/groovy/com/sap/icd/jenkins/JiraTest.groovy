#!groovy
package com.sap.icd.jenkins

import org.junit.Rule
import org.junit.Test
import util.BasePiperTest

import static org.junit.Assert.assertEquals
import org.junit.rules.RuleChain

import util.Rules

class JiraTest extends BasePiperTest {
    @Rule
    public RuleChain rules = Rules.getCommonRules(this)

    @Test
    void testJiraBasic() throws Exception {
        // load Jira
        Jira jira = new Jira(this, 'test')
        jira.setApiUrl('https://test/rest/api/2')
        assertEquals(jira.getApiUrl(), 'https://test/rest/api/2')
        assertEquals(jira.getServerUrl(), "${'https://test'}")
        // test setApiUrl
        jira.setApiUrl('https://test/rest/api/2')
        assertEquals(jira.getApiUrl(), 'https://test/rest/api/2')

        assertJobStatusSuccess()
    }
}

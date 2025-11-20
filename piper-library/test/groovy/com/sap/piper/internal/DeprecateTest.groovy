#!groovy
package com.sap.piper.internal

import org.codehaus.groovy.runtime.InvokerHelper
import util.BasePiperTest

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain

import util.Rules
import util.JenkinsLoggingRule
import util.JenkinsEnvironmentRule

class DeprecateTest extends BasePiperTest {

    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jer)

    @Test
    void testDeprecateParameter() throws Exception {
        Map map = [oldKey: 'justATest']
        Deprecate.parameter(nullScript, map, 'oldKey', 'newKey')
        // asserts
        assertEquals('Map does not contain the new key', 2, map.size())
        assertTrue('Map does not contain the new key',
            map.containsKey('newKey'))
        assertEquals('Map does not contain the expected value for newKey',
            'justATest', map.get('newKey'))
        assertTrue('Log does not contain the expected message',
            jlr.log.contains('[WARNING] The parameter oldKey is DEPRECATED, use newKey instead'))
    }

    @Test
    void testDeprecateParameterWithoutReplacement() throws Exception {
        Map map = [oldKey: 'justATest']
        // asserts
        assertEquals('', 1, map.size())
        // test
        Deprecate.parameter(nullScript, map, 'oldKey', null)
        // asserts
        assertEquals('', 1, map.size())
        assertTrue('Log does not contain the expected message',
            jlr.log.contains('[WARNING] The parameter oldKey is DEPRECATED'))
    }

    @Test
    void testDeprecateParameterWithType() throws Exception {
        Map map = [oldKey: 'justATest']
        // asserts
        assertEquals('', 1, map.size())
        // test
        Deprecate.parameter(nullScript, map, 'oldKey', null, 'step')
        // asserts
        assertEquals('', 1, map.size())
        assertTrue('Log does not contain the expected message',
            jlr.log.contains('[WARNING] The step parameter oldKey is DEPRECATED'))
    }

    @Test
    void testDeprecateParameterWithoutUpdate() throws Exception {
        Map map = [oldKey: 'justATest', newKey: 'anotherTest']
        // asserts
        assertEquals('', 2, map.size())
        // test
        Deprecate.parameter(nullScript, map, 'oldKey', 'newKey')
        // asserts
        assertEquals('', 2, map.size())
        assertEquals('Map does not contain the expected value for oldKey',
            'justATest', map.get('oldKey'))
        assertEquals('Map does not contain the expected value for newKey',
            'anotherTest', map.get('newKey'))
        assertTrue('Log does not contain the expected message',
            jlr.log.contains('[WARNING] The parameter oldKey is DEPRECATED, use newKey instead'))
    }

    @Test
    void testDeprecateParameterValue() throws Exception {
        Map map = [key: 'oldValue']
        // test
        Deprecate.value(nullScript, map, 'key', 'oldValue', 'newValue')
        // asserts
        assertEquals('Map does not contain the expected value for key',
            'newValue', map.get('key'))
        assertTrue('Log does not contain the expected message',
            jlr.log.contains('[WARNING] The value \'oldValue\' for the parameter key is DEPRECATED, use \'newValue\' instead'))
    }

    @Test
    void testDeprecateParameterValueWithoutReplacement() throws Exception {
        Map map = [key: 'oldValue']
        // test
        Deprecate.value(nullScript, map, 'key', 'oldValue')
        // asserts
        assertEquals('Map does not contain the expected value for key',
            'oldValue', map.get('key'))
        assertTrue('Log does not contain the expected message',
            jlr.log.contains('[WARNING] The value \'oldValue\' for the parameter key is DEPRECATED'))
    }
}

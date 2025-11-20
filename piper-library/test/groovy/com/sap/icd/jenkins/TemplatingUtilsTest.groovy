package com.sap.icd.jenkins

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.JenkinsLoggingRule
import util.JenkinsSetupRule
import util.SharedLibraryCreator

import static org.junit.Assert.assertEquals

class TemplatingUtilsTest extends BasePipelineTest {

    @Rule
    public ExpectedException exception = ExpectedException.none()
    public JenkinsSetupRule setUpRule = new JenkinsSetupRule(this, SharedLibraryCreator.lazyLoadedLibrary)
    public JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain =
        RuleChain.outerRule(setUpRule)
            .around(loggingRule)


    @Before
    void init() throws Exception {
    }

    @Test
    void testRenderImpl() {
        Map context = [a: 'AAA', b: 'BBB']
        String template = 'my a="${a}", and b="${b}"'
        String expected = 'my a="AAA", and b="BBB"'
        def actual = TemplatingUtils.renderImpl(template, context)
        assertEquals(expected, actual)
        context = [:]
        template = 'my a="${a}"'
        exception.expect(Exception.class)
        exception.expectMessage("No such property: a for class")
        TemplatingUtils.renderImpl(template, context)
        // nested
        context = [a: [b: 'BBB']]
        template = 'my a.b="${a.b}"'
        expected = 'my a.b="BBB"'
        TemplatingUtils.renderImpl(template, context)
        assertEquals(expected, actual)
    }

    @Test
    void testRender() {
        String template = null
        String expected = null
        // no varName in template
        def actual = TemplatingUtils.render(template, [:])
        assertEquals(expected, actual)
        template = 'quak'
        expected = template
        // no context
        actual = TemplatingUtils.render(template, [:])
        assertEquals(expected, actual)
        // existing context and all is happy
        def kik = 'quak'
        template = 'quak: ${kik}'
        expected = "quak: ${kik}".toString()
        actual = TemplatingUtils.render(template, [kik: kik])
        assertEquals(expected, actual)
    }

    void prepareObjectInterceptors(object) {
        object.metaClass.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.static.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.methodMissing = helper.getMethodMissingInterceptor()
    }
}

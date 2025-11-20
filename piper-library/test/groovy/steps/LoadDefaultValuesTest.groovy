#!groovy
package steps

import org.junit.After
import org.junit.Before
import org.junit.Rule;
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import org.springframework.test.annotation.DirtiesContext;

import static org.junit.Assert.assertThat

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.hasSize
import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.iterableWithSize

import org.springframework.test.annotation.DirtiesContext;

import com.sap.piper.internal.DefaultValueCache

import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule;
import util.Rules

@DirtiesContext(classMode = DirtiesContext.ClassMode.AFTER_CLASS)
public class LoadDefaultValuesTest extends BasePiperTest {

    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jsr)
        .around(jlr)

    @Before
    public void setup() {
        helper.registerAllowedMethod("libraryResource", [String], { fileName-> return fileName })
        helper.registerAllowedMethod("readYaml", [Map], { m ->
            switch(m.text) {
                case 'piper-defaults.yml': return [default: 'config']
                case 'custom.yml': return [custom: 'myConfig']
                case 'not_found': throw new hudson.AbortException('No such library resource not_found could be found')
                default: return [the:'end']
            }
        })
    }

    @Test
    public void testDefaultPipelineEnvironmentOnly() {
        jsr.step.loadDefaultValues(script: nullScript)
        // asserts
        assertThat(DefaultValueCache.getInstance().getDefaultValues().keySet(), hasSize(1))
        assertThat(DefaultValueCache.getInstance().getDefaultValues(), hasEntry('default', 'config'))
    }

    @Test
    public void testReInitializeOnCustomConfig() {
        DefaultValueCache.createInstance([key:'value'])
        // existing instance is dropped in case a custom config is provided.
        jsr.step.loadDefaultValues(script: nullScript, customDefaults: 'custom.yml')
        // asserts
        assertThat(DefaultValueCache.getInstance().getDefaultValues().keySet(), hasSize(2))
        assertThat(DefaultValueCache.getInstance().getDefaultValues(), hasEntry('default', 'config'))
        assertThat(DefaultValueCache.getInstance().getDefaultValues(), hasEntry('custom', 'myConfig'))
    }

    @Test
    public void testNoReInitializeWithoutCustomConfig() {
        DefaultValueCache.createInstance([key:'value'])
        jsr.step.loadDefaultValues(script: nullScript)
        // asserts
        assertThat(DefaultValueCache.getInstance().getDefaultValues().keySet(), hasSize(1))
        assertThat(DefaultValueCache.getInstance().getDefaultValues(), hasEntry('key', 'value'))
    }

    @Test
    public void testAttemptToLoadNonExistingConfigFile() {
        // Behavior documented here based on reality check
        thrown.expect(hudson.AbortException.class)
        thrown.expectMessage('No such library resource not_found could be found')

        jsr.step.loadDefaultValues(script: nullScript, customDefaults: 'not_found')
    }

    @Test
    public void testDefaultPipelineEnvironmentWithCustomConfigReferencedAsString() {
        jsr.step.loadDefaultValues(script: nullScript, customDefaults: 'custom.yml')
        // asserts
        assertThat(DefaultValueCache.getInstance().getDefaultValues().keySet(), hasSize(2))
        assertThat(DefaultValueCache.getInstance().getDefaultValues(), hasEntry('default', 'config'))
        assertThat(DefaultValueCache.getInstance().getDefaultValues(), hasEntry('custom', 'myConfig'))
    }

    @Test
    public void testDefaultPipelineEnvironmentWithCustomConfigReferencedAsList() {
        jsr.step.loadDefaultValues(script: nullScript, customDefaults: ['custom.yml'])
        // asserts
        assertThat(DefaultValueCache.getInstance().getDefaultValues().keySet(), hasSize(2))
        assertThat(DefaultValueCache.getInstance().getDefaultValues(), hasEntry('default', 'config'))
        assertThat(DefaultValueCache.getInstance().getDefaultValues(), hasEntry('custom', 'myConfig'))
    }

    @Test
    public void testAssertNoLogMessageInCaseOfNoAdditionalConfigFiles() {
        jsr.step.loadDefaultValues(script: nullScript)
        // asserts
        assertThat(jlr.log, containsString("Loading library configuration file 'piper-defaults.yml'"))
    }

    @Test
    public void testAssertLogMessageInCaseOfMoreThanOneConfigFile() {
        jsr.step.loadDefaultValues(script: nullScript, customDefaults: ['custom.yml'])
        // asserts
        assertThat(jlr.log, containsString("Loading library configuration file 'piper-defaults.yml'"))
        assertThat(jlr.log, containsString("Loading library configuration file 'custom.yml'"))
    }
}

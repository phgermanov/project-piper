#!groovy
package steps

import org.codehaus.groovy.GroovyException
import org.junit.Before
import org.junit.Ignore
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

@Ignore("step disabled")
class ExecuteDockerOnKubernetesTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    def stashList = []
    def unstashList = []
    def containerTemplateMap = [:]

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jscr)
        .around(jsr)

    @Before
    void init() throws Exception {
        // register Jenkins commands with mock values
        helper.registerAllowedMethod("stash", [String.class], {string ->
            stashList.add([name: string])
        })
        helper.registerAllowedMethod("stash", [Map.class], {map ->
            stashList.add(map)
        })
        helper.registerAllowedMethod("unstash", [String.class], {string ->
            unstashList.add(string.toString())
        })
        helper.registerAllowedMethod("containerTemplate", [Map.class], {map ->
            containerTemplateMap = map
        })
        helper.registerAllowedMethod("isOldKubePluginVersion", [], {return false})
    }

    @Test
    void testExecuteWithError() throws Exception {
        binding.setVariable('params', null)

        helper.registerAllowedMethod("envVar", [Map.class], {map ->
            return [:]
        })
        helper.registerAllowedMethod("podTemplate", [Map.class, Closure.class], {  map, closure ->
            return closure()
        })
        helper.registerAllowedMethod("container", [Map.class, Closure.class], {  map, closure ->
            return closure()
        })

        try {
            jsr.step.executeDockerOnKubernetes(
                script: nullScript,
                juStabUtils: utils,
                uniqueId: 'test-unique',
                dockerImage: 'testImage',
                dockerEnvVars: [:],
                jenkinsUtilsStub: jenkinsUtils
            ) {
                throw new GroovyException("Failure in pipeline")
            }
        }
        catch (GroovyException e) {
            assertThat(stashList, hasItem(hasEntry('name', "${'workspace-test-unique'}")))
            assertThat(stashList, hasItem(hasEntry('name', "${'container-test-unique'}")))
            assertThat(unstashList, hasItem('container-test-unique'))
        }
    }
}

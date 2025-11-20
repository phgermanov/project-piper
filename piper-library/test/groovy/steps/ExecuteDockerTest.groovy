#!groovy
package steps

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is

import org.junit.Before
import org.junit.Ignore
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import static org.junit.Assert.assertThat

//import org.springframework.test.annotation.DirtiesContext

import util.BasePiperTest
import util.Rules
import util.JenkinsStepRule
import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.JenkinsExecuteDockerRule

@Ignore("step disabled")
//@DirtiesContext(classMode = DirtiesContext.ClassMode.AFTER_CLASS, hierarchyMode = DirtiesContext.HierarchyMode.EXHAUSTIVE)
class ExecuteDockerTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jscr)
        .around(jsr)

    @Before
    void init() throws Exception {
        // register Jenkins commands with mock values
        helper.registerAllowedMethod("executeDockerOnKubernetes", [Map.class], null)
    }

    @Test
    void testExecuteWithoutDocker() throws Exception {
        jsr.step.executeDocker(
            script: nullScript,
            juStabUtils: utils
        ) {
            jscr.shell.push("echo 'This is a test!'")
        }
        // asserts
        assertThat('the body is executed', jscr.shell, hasItem("echo 'This is a test!'"))
        assertThat(jlr.log, containsString('Parameters: dindImage: null, dockerEnvVars: null, dockerImage: null, dockerOptions: null, dockerVolumeBind: null, dockerWorkspace: null, stashBackConfig: [excludes:nohup.out], skipStashBack: null, stashContent: []'))
    }

    @Test
    void testExecuteWithDocker() throws Exception {
        def dockerParams = [:]
        // mock docker object
		binding.setVariable('docker', [
            image: { name ->
                dockerParams.image = name
                return [
                    pull: { dockerParams.pull = true},
                    inside: { options, body ->
                        dockerParams.options = options
                        body()
                    }
                ]
            }
        ])

        jsr.step.executeDocker(
            script: nullScript,
            juStabUtils: utils,
            dockerImage: 'piper.int.repositories.cloud.sap/piper/maven',
            dockerEnvVars: ['MY_ENV_ARG': 'ENV_ARG_VAL']
        ) {
            jscr.shell.push("echo 'This is a test!'")
        }

        // asserts
        assertThat('Docker image is set to the custom value', dockerParams.image, is('piper.int.repositories.cloud.sap/piper/maven'))
        assertThat('new Docker image is pulled', dockerParams.pull, is(true))
        assertThat('Docker env vars is set to the custom value', dockerParams.options, is('--env MY_ENV_ARG=ENV_ARG_VAL'))
        assertThat(jlr.log, containsString('Parameters: dindImage: null, dockerEnvVars: [MY_ENV_ARG:ENV_ARG_VAL], dockerImage: piper.int.repositories.cloud.sap/piper/maven, dockerOptions: null, dockerVolumeBind: null, dockerWorkspace: null, stashBackConfig: [excludes:nohup.out], skipStashBack: null, stashContent: []'))
    }

    @Ignore('test implementation needed')
    @Test
    void testExecuteWithDockerOnKubernetes() throws Exception {}

    @Test
    void testGetDockerOptions() throws Exception {
        def result = jsr.step.getDockerOptions(['MY_ENV_ARG': '105'], ['~': '/home/piper/home'], '--endpoint bash')
        // asserts
        assertThat(result, is('--env MY_ENV_ARG=105 --volume ~:/home/piper/home --endpoint bash'))
    }
}

package steps

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class DeployToCloudFoundryWithIRISTest extends BasePiperTest {

    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsReadJsonRule jrjr = new JenkinsReadJsonRule(this, 'test/resources/deployToCFWithIris/')
    private JenkinsReadYamlRule jryr = new JenkinsReadYamlRule(this, 'test/resources/deployToCFWithIris/')

    private deployCFCalls = []

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jrjr)
        .around(jryr)
        .around(jsr) // needs to be activated after jedr, otherwise executeDocker is not mocked

    @Before
    void init() {
        helper.registerAllowedMethod('deployToCloudFoundry', [Map.class], {m ->
            deployCFCalls.add(m)
        })
    }

    @Test
    void testSuccess() {
        helper.registerAllowedMethod('fileExists', [String.class], {s -> return true})

        jsr.step.deployToCloudFoundryWithIRIS([
            script: nullScript,
            run: true
        ])

        assertThat(deployCFCalls[0].cfAppName, is('testApp'))
        assertThat(deployCFCalls[0].cfManifest, is('manifest.yml'))
        assertThat(deployCFCalls[0].cfApiEndpoint, is('https://api1.test.com'))
        assertThat(deployCFCalls[0].cfOrg, is('testOrg1'))
        assertThat(deployCFCalls[0].cfSpace, is('testSpace1'))

        assertThat(deployCFCalls[1].cfApiEndpoint, is('https://api2.test.com'))
        assertThat(deployCFCalls[1].cfOrg, is('testOrg2'))
        assertThat(deployCFCalls[1].cfSpace, is('testSpace2'))
    }

    @Test
    void testFailure() {
        helper.registerAllowedMethod('fileExists', [String.class], {s -> return false})
        jsr.step.deployToCloudFoundryWithIRIS([
            script: nullScript,
            run: true
        ])
        assertThat(jlr.log, containsString('Transferfile exitIRISTenants.json does not exist. Nothing to deploy!'))
    }

    @Test
    void testDontExecute() {
        jsr.step.deployToCloudFoundryWithIRIS([
            script: nullScript,
            run: false
        ])

        assertThat(jlr.log, containsString('Don\'t execute, based on configuration setting.'))
    }
}

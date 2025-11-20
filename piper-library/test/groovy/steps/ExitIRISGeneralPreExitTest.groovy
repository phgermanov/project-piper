package steps

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class ExitIRISGeneralPreExitTest extends BasePiperTest {

    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsExecuteDockerRule jedr = new JenkinsExecuteDockerRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jscr)
        .around(jedr)
        .around(jsr) // needs to be activated after jedr, otherwise executeDocker is not mocked

    @Before
    void init() {
        helper.registerAllowedMethod('usernamePassword', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            binding.setProperty('username', 'test_cf')
            binding.setProperty('password', '********')
            try {
                c()
            } finally {
                binding.setProperty('username', null)
                binding.setProperty('password', null)
            }
        })
    }

    @Test
    void testNoDebug() {
        jsr.step.exitIRISGeneralPreExit([
            script: nullScript,
            run: true,
            odataSystemRole: 'testRoleCode',
            spcCredentialsId: 'testCredentials'
        ])

        assertThat(jedr.dockerParams.dockerImage, is('docker.wdf.sap.corp:51148/test/iris:latest'))
        assertThat(jedr.dockerParams.dockerWorkspace, is('/home/root/scripts'))
        //assertThat(jedr.dockerParams.stashContent, hasItem(''))

        assertThat(jscr.shell, hasItem('/usr/bin/groovy /home/root/scripts/dispatchIRIS.groovy piperExits/exitIRISGeneralPreExit_Docker.groovy eyJTeXN0ZW1Sb2xlQ29kZSI6InRlc3RSb2xlQ29kZSIsIm9EYXRhTGlmZWN5Y2xlU3RhdHVzIjoiTElWRSIsIm9EYXRhVGVuYW50Um9sZSI6IjA0Iiwib0RhdGFCdXNpbmVzc1R5cGUiOiJaSDQyMSIsIm9EYXRhRnJlZVN0eWxlIjoiIiwiU1BDIjoiTlpBIiwiU1BDX1VzZXJuYW1lIjoidGVzdF9jZiIsIlNQQ19QYXNzd29yZCI6IioqKioqKioqIiwiRGVidWdNZXNzYWdlcyI6ImZhbHNlIn0='))
    }

    @Test
    void testDebug() {
        jsr.step.exitIRISGeneralPreExit([
            script: nullScript,
            run: true,
            exitIRISdebugMessages: 'true',
            odataSystemRole: 'testRoleCode',
            spcCredentialsId: 'testCredentials'
        ])

        assertThat(jscr.shell, hasItem('cat exitIRISTenants.json'))
    }

    @Test
    void testDontExecute() {
        jsr.step.exitIRISGeneralPreExit([
            script: nullScript,
            run: false,
            odataSystemRole: 'testRoleCode',
            spcCredentialsId: 'testCredentials'
        ])

        assertThat(jlr.log, containsString('Don\'t execute, based on configuration setting.'))
    }
}
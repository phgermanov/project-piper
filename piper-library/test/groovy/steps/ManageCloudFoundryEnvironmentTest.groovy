#!groovy
package steps

import com.sap.icd.jenkins.EnvironmentManagerRunner
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.containsString
import static org.junit.Assert.assertThat

class ManageCloudFoundryEnvironmentTest extends BasePiperTest {

    public static final String MH_CREDENTIAL = 'MH_CREDENTIAL'
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsErrorRule jer = new JenkinsErrorRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsExecuteDockerRule jedr = new JenkinsExecuteDockerRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsWriteYamlRule jwyr = new JenkinsWriteYamlRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jer)
        .around(jedr)
        .around(jscr)
        .around(jwyr)
        .around(jlr)
        .around(jsr)

    static final String DOCKER_IMAGE = 'piper.int.repositories.cloud.sap/piper/cf-cli'
    static final String DOCKER_WORKSPACE = '/home/piper'
    static final String CF_CREDENTIAL = "CF_CREDENTIAL"
    static final String EM_COMMAND = "setup-environment"
    static final String COMMAND = "$EM_COMMAND -y myEnv.yml"

    @Before
    public void init() {
        helper.registerAllowedMethod('usernamePassword', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            binding.setProperty('CF_USERNAME', 'test_cf')
            binding.setProperty('CF_PASSWORD', '********')
            try {
                c()
            } finally {
                binding.setProperty('CF_USERNAME', null)
                binding.setProperty('CF_PASSWORD', null)
            }
        })
    }

    @After
    public void cleanupMetaClass() {
        EnvironmentManagerRunner.metaClass = null
    }

    @Test
    public void executeEnvManInDockerByDefault() {
        helper.registerAllowedMethod('fileExists', [String], { s -> return false })
        EnvironmentManagerRunner.metaClass.executeWithDocker = { String s, String d, String w ->
            assert s.contains('setup-environment -y environment.yml')
            assert d.contains(DOCKER_IMAGE)
            assert w.contains(DOCKER_WORKSPACE)
            return 0
        }

        jsr.step.manageCloudFoundryEnvironment([
            script         : nullScript,
            cfCredentialsId: CF_CREDENTIAL,
            juStabUtils    : utils,
        ])

        assertThat(jlr.log, containsString('EMR returned successful with:'))

    }

    @Test
    public void shouldStopIfEnvManReturnsUnsuccessful() throws Exception {
        helper.registerAllowedMethod('fileExists', [String], { s -> return false })
        EnvironmentManagerRunner.metaClass.executeWithDocker = { String s, String d, String w ->
            assert cfCredentialsId == CF_CREDENTIAL
            assert muenchhausenCredentialsId == null
            assert s.contains(COMMAND)
            assert d.contains(DOCKER_IMAGE)
            assert w.contains(DOCKER_WORKSPACE)
            return -1
        }

        thrown.expectMessage('EnvironmentManager encountered a problem! Stopping!!! (-1)')

        jsr.step.manageCloudFoundryEnvironment([
            script         : nullScript,
            cfCredentialsId: CF_CREDENTIAL,
            command        : COMMAND,
            juStabUtils    : utils
        ])
    }

    @Test
    public void shouldExecuteDirectlyIfDockerImageIsEmptyString() throws Exception {
        helper.registerAllowedMethod('fileExists', [String], { s -> return false })
        EnvironmentManagerRunner.metaClass.execute = { String s ->
            assert cfCredentialsId == CF_CREDENTIAL
            assert muenchhausenCredentialsId == null
            assert s.contains(COMMAND)
            return 0
        }

        jsr.step.manageCloudFoundryEnvironment([
            script     : nullScript,
            command    : COMMAND,
            dockerImage: '',
            juStabUtils: utils,
        ])

        assertThat(jlr.log, containsString('EMR returned successful with:'))
    }

    @Test
    public void cfCredentialsIdDefaultIsSet() {
        helper.registerAllowedMethod('fileExists', [String], { s -> return false })
        EnvironmentManagerRunner.metaClass.executeWithDocker = { String s, String d, String w ->
            assert cfCredentialsId == CF_CREDENTIAL
            return 0
        }

        jsr.step.manageCloudFoundryEnvironment([
            script     : nullScript,
            juStabUtils: utils,
        ])

        assertThat(jlr.log, containsString('EMR returned successful with:'))
    }


    @Test
    public void mhCredentialsIdParameterIsPassedToEMT() {
        helper.registerAllowedMethod('fileExists', [String], { s -> return false })
        EnvironmentManagerRunner.metaClass.executeWithDocker = { String s, String d, String w ->
            assert muenchhausenCredentialsId == MH_CREDENTIAL
            return 0
        }

        jsr.step.manageCloudFoundryEnvironment([
            script         : nullScript,
            juStabUtils    : utils,
            mhCredentialsId: MH_CREDENTIAL
        ])

        assertThat(jlr.log, containsString('EMR returned successful with:'))
    }


    @Test
    public void stepConfigShouldOverwriteDefaultCfCredentialsId() {

        def customId = 'customId'
        helper.registerAllowedMethod('fileExists', [String], { s -> return false })
        EnvironmentManagerRunner.metaClass.executeWithDocker = { String s, String d, String w ->
            assert cfCredentialsId == customId
            return 0
        }

        nullScript.globalPipelineEnvironment.configuration = [
            general: [:],
            steps  : [manageCloudFoundryEnvironment: [cfCredentialsId: customId]],
            stages : [:]
        ]

        jsr.step.manageCloudFoundryEnvironment([
            script     : nullScript,
            juStabUtils: utils,
        ])

        assertThat(jlr.log, containsString('EMR returned successful with:'))
    }

    @Test
    public void parameterShouldOverwriteDefaultCfCredentialsId() {

        def customId = 'customId'
        helper.registerAllowedMethod('fileExists', [String], { s -> return false })
        EnvironmentManagerRunner.metaClass.executeWithDocker = { String s, String d, String w ->
            assert cfCredentialsId == customId
            return 0
        }

        jsr.step.manageCloudFoundryEnvironment([
            script         : nullScript,
            cfCredentialsId: customId,
            juStabUtils    : utils,
        ])

        assertThat(jlr.log, containsString('EMR returned successful with:'))
    }
}

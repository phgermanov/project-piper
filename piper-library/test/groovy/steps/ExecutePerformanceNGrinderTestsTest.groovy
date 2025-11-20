#!groovy
 package steps

import org.junit.Rule
import org.junit.rules.RuleChain
import util.JenkinsExecuteDockerRule
import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.Rules

import static org.hamcrest.Matchers.*

import org.junit.Before
import org.junit.Ignore
import org.junit.Test
import util.BasePiperTest

import static org.junit.Assert.assertThat

class ExecutePerformanceNGrinderTestsTest extends BasePiperTest {

    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsExecuteDockerRule jedr = new JenkinsExecuteDockerRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jscr)
        .around(jedr)
        .around(jsr)

    @Before
    void init() {
        jscr.setReturnValue('git ls-remote https://github.wdf.sap.corp/ContinuousDelivery/piper-library.git refs/heads/master', '1.4.3')

        binding.setProperty('BUILD_NUMBER', 7)

        helper.registerAllowedMethod('usernamePassword', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            binding.setProperty('cf_username', 'test_cf')
            binding.setProperty('cf_password', 'cf******')

            binding.setProperty('influx_username', 'test_influx')
            binding.setProperty('influx_password', 'in******')

            binding.setProperty('grafana_username', 'test_grafana')
            binding.setProperty('grafana_password', 'gr******')

            binding.setProperty('ngrinder_username', 'test_ngrinder')
            binding.setProperty('ngrinder_password', 'ng******')

            try {
                c()
            } finally {
                binding.setProperty('username', null)
                binding.setProperty('password', null)

                binding.setProperty('influx_username', null)
                binding.setProperty('influx_password', null)

                binding.setProperty('grafana_username', null)
                binding.setProperty('grafana_password', null)

                binding.setProperty('ngrinder_username', null)
                binding.setProperty('ngrinder_password', null)
            }
        })
    }


    @Test
    @Ignore("config.properties support removed")
    void testNGrinderExecution() throws Exception {

        jsr.step.executePerformanceNGrinderTests([
            juStabUtils: utils,
            script     : nullScript,
            cfMetrics  : true,
            cfCredentialsId: 'test_cfCredentialsId',
            cfApiEndpoint: 'https://api.cf.sap.hana.ondemand.com',
            cfAppGuid: 'testGuid',
            cfOrg: 'testOrg',
            cfSpace: 'perfSpace',
            configFile : 'resources/config.properties',
            influxcredentialsId: 'influx_1234567890',
            grafanaCredentialsId: 'grafana_1234567890',
            ngrinderCredentialsId: 'ngrinder_1234567890',
            isNewNgrinderVersion: true
        ])

        assertThat(jlr.log, containsString('Tools credential ID\'s: cf Credentials ID=test_cfCredentialsId, influx Credentials ID=influx_1234567890, grafana Credentials ID=grafana_1234567890, nGrinder Credentials ID=ngrinder_1234567890'))
        assertThat(jlr.log, containsString('CF Parameters: CF Org=testOrg, CF Api Endpoint=https://api.cf.sap.hana.ondemand.com, CF Space=perfSpace'))

        assertThat(jedr.dockerParams.dockerImage, is('piper.int.repositories.cloud.sap/piper/ngrinder'))
        assertThat(jedr.dockerParams.dockerWorkspace, is('/home/piper'))
        assertThat(jedr.dockerParams.stashContent, hasItem('buildDescriptor'))
        assertThat(jedr.dockerParams.stashContent, hasItem('tests'))

        assertThat(jscr.shell, hasItem("cf login -u 'test_cf' -p 'cf******' -a https://api.cf.sap.hana.ondemand.com -o 'testOrg' -s 'perfSpace'"))
        assertThat(jscr.shell, hasItem("/opt/bin/ci-perfscript.sh 'gr******' 'ng******' testGuid 'test_grafana' 'test_ngrinder' 7 'test_influx' 'in******' true true"))
        assertThat(jscr.shell, hasItem('cf logout'))

        assertJobStatusSuccess()

    }
}

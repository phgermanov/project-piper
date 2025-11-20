#!groovy
package steps

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class ManageUaaServiceTest extends BasePiperTest {

    private ExpectedException thrown = ExpectedException.none()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsReadJsonRule jrjr = new JenkinsReadJsonRule(this, 'test/resources/manageUaa/')
    private JenkinsWriteFileRule jwfr = new JenkinsWriteFileRule(this)
    private JenkinsExecuteDockerRule jedr = new JenkinsExecuteDockerRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jlr)
        .around(jscr)
        .around(jrjr)
        .around(jwfr)
        .around(jedr)
        .around(jsr)

    @Before
    void init() {
        helper.registerAllowedMethod('usernamePassword', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            if (l[0].credentialsId == 'testCredentialsId') {
                binding.setProperty('cfUser', 'testUser')
                binding.setProperty('cfPassword', '********')
            }
            try {
                c()
            } finally {
                binding.setProperty('cfUser', null)
                binding.setProperty('cfPassword', null)
            }
        })
    }

    @Test
    void testNoSubAccountUaaExisting() {

        jscr.setReturnStatus('cf service testInstance', 0)

        jsr.step.manageUaaService(
            script: nullScript,
            juStabUtils: utils,
            //remove cfApiEndpoint
            cfApiEndpoint: 'https://api.cf.sap.hana.ondemand.com',
            cfOrg: 'testOrg',
            cfSpace: 'test_Space',
            cfCredentialsId: 'testCredentialsId',
            cfServiceInstance: 'testInstance'
        )

        assertThat(jlr.log, containsString('Unstash content: securityDescriptor'))

        assertThat(jedr.dockerParams.dockerImage, is('piper.int.repositories.cloud.sap/piper/cf-cli'))
        assertThat(jedr.dockerParams.dockerWorkspace, is('/home/piper'))

        assertThat(jwfr.files['xs-security-test-space.json'], containsString('testApp-test-space'))

        assertThat(jscr.shell, hasItem("cf login -u 'testUser' -p '********' -a https://api.cf.sap.hana.ondemand.com -o 'testOrg' -s 'test_Space'"))
        assertThat(jscr.shell, hasItem('cf update-service testInstance -c xs-security-test-space.json'))

    }

    @Test
    void testNoSubAccountUaaExistingError() {
        jscr.setReturnStatus('cf service testInstance', 0)
        jscr.setReturnStatus('cf update-service testInstance -c xs-security-test-space.json', 1)

        thrown.expectMessage('Update of uaa service instance failed.')
        jsr.step.manageUaaService(
            script: nullScript,
            juStabUtils: utils,
            //remove cfApiEndpoint
            cfApiEndpoint: 'https://api.cf.sap.hana.ondemand.com',
            cfOrg: 'testOrg',
            cfSpace: 'test_Space',
            cfCredentialsId: 'testCredentialsId',
            cfServiceInstance: 'testInstance',
            servicePlan: 'test'
        )
    }

    @Test
    void testNoSubAccountUaaNonExisting() {

        jscr.setReturnStatus('cf service testInstance', 1)

        jsr.step.manageUaaService(
            script: nullScript,
            juStabUtils: utils,
            //remove cfApiEndpoint
            cfApiEndpoint: 'https://api.cf.sap.hana.ondemand.com',
            cfOrg: 'testOrg',
            cfSpace: 'test_Space',
            cfCredentialsId: 'testCredentialsId',
            cfServiceInstance: 'testInstance'
        )

        assertThat(jscr.shell, hasItem('cf create-service xsuaa broker testInstance -c xs-security-test-space.json'))
    }
}

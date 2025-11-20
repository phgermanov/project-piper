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

class DeployMultipartAppToCloudFoundryTest extends BasePiperTest {

    private ExpectedException thrown = ExpectedException.none()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsReadFileRule jrfr = new JenkinsReadFileRule(this, 'test/resources/deployMultipartApp')
    private JenkinsWriteFileRule jwfr = new JenkinsWriteFileRule(this)
    private JenkinsExecuteDockerRule jedr = new JenkinsExecuteDockerRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jlr)
        .around(jscr)
        .around(jrfr)
        .around(jwfr)
        .around(jedr)
        .around(jsr) // needs to be activated after jedr, otherwise executeDocker is not mocked

    @Before
    void init() throws Throwable {
        helper.registerAllowedMethod('usernamePassword', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            binding.setProperty('cfUser', 'test_cf')
            binding.setProperty('cfPassword', '********')
            try {
                c()
            } finally {
                binding.setProperty('cfUser', null)
                binding.setProperty('cfPassword', null)
            }
        })
    }

    @Test
    void testDefault() {

        def cfModules = [
            [
                cfAppName                 : 'module1Name',
                cfManifestPath            : 'testManifest1.yml',
            ],
            [
                cfAppName                 : 'module2Name',
                cfManifestPath            : 'testManifest2.yml',
            ]
        ]

        jscr.setReturnStatus('cf create-domain test_Org cfapps.sap.hana.ondemand.com', 0)
        jscr.setReturnStatus('cf app module1Name', 1)
        jscr.setReturnStatus('cf app module2Name', 1)


        jsr.step.deployMultipartAppToCloudFoundry([
            script         : nullScript,
            juStabUtils    : utils,
            cfCredentialsId: 'testCredentials',
            cfOrg          : 'test_Org',
            cfSpace        : 'test_Space',
            modules: cfModules
        ])

        assertThat(jedr.dockerParams.dockerImage, is('piper.int.repositories.cloud.sap/piper/cf-cli'))
        assertThat(jedr.dockerParams.dockerWorkspace, is('/home/piper'))

        assertThat(jscr.shell, hasItem("cf login -u 'test_cf' -p '********' -a https://api.cf.sap.hana.ondemand.com -o 'test_Org' -s 'test_Space'"))
        assertThat(jscr.shell, hasItem('cf create-domain test_Org cfapps.sap.hana.ondemand.com'))
        assertThat(jscr.shell, hasItem('cf push module1Name -n module1Name-test-space -f testManifest1.yml -d cfapps.sap.hana.ondemand.com'))
        assertThat(jscr.shell, hasItem('cf push module2Name -n module2Name-test-space -f testManifest2.yml -d cfapps.sap.hana.ondemand.com'))

        assertThat(jscr.shell, hasItem('cf logout'))

        assertThat(jlr.log, containsString('Custom domain created.'))
        assertThat(jlr.log, containsString('Not needed. No blue-green-deployment for this module.'))

    }

    @Test
    void testDefaultWithEnv() {

        def cfModules = [
            [
                cfAppName                 : 'module1Name',
                cfManifestPath            : 'testManifest1.yml',
                cfEnvVariables: [
                    env1: 'testEnv1',
                    env2: 'testEnv2'
                ]
            ],
            [
                cfAppName                 : 'module2Name',
                cfManifestPath            : 'testManifest2.yml',
            ]
        ]

        jscr.setReturnStatus('cf create-domain test_Org cfapps.sap.hana.ondemand.com', 1)

        jsr.step.deployMultipartAppToCloudFoundry([
            script         : nullScript,
            juStabUtils    : utils,
            cfCredentialsId: 'testCredentials',
            cfOrg          : 'test_Org',
            cfSpace        : 'test_Space',
            modules: cfModules
        ])

        assertThat(jscr.shell, hasItem('cf push module1Name -n module1Name-test-space -f testManifest1.yml -d cfapps.sap.hana.ondemand.com --no-start'))
        assertThat(jscr.shell, hasItem('cf set-env module1Name env1 "testEnv1"'))
        assertThat(jscr.shell, hasItem('cf set-env module1Name env2 "testEnv2"'))
        assertThat(jscr.shell, hasItem('cf start module1Name'))
        assertThat(jscr.shell, hasItem('cf push module2Name -n module2Name-test-space -f testManifest2.yml -d cfapps.sap.hana.ondemand.com'))

        assertThat(jlr.log, containsString('Domain already existing.'))
    }

    @Test
    void testBlueGreenAppExisting() {

        def cfModules = [
            [
                cfAppName                 : 'module1Name',
                cfManifestPath            : 'testManifest1.yml'
            ],
            [
                cfAppName                 : 'module2Name',
                cfManifestPath            : 'testManifest2.yml'
            ]
        ]

        jscr.setReturnStatus('cf create-domain testOrg cfapps.sap.hana.ondemand.com', 1)

        jscr.setReturnStatus('cf app module1Name', 0)
        jscr.setReturnStatus('cf app module2Name', 0)
        jscr.setReturnStatus('cf app module2Name-old', 0)


        jsr.step.deployMultipartAppToCloudFoundry([
            script         : nullScript,
            juStabUtils    : utils,
            cfCredentialsId: 'testCredentials',
            cfOrg          : 'testOrg',
            cfSpace        : 'testSpace',
            deployType: 'blue-green',
            modules: cfModules
        ])

        assertThat(jscr.shell, hasItem('cf push module1Name-new -n module1Name-testspace-new -f testManifest1.yml -d cfapps.sap.hana.ondemand.com'))
        assertThat(jscr.shell, hasItem('cf push module2Name-new -n module2Name-testspace-new -f testManifest2.yml -d cfapps.sap.hana.ondemand.com'))

        assertThat(jscr.shell, hasItem('cf map-route module1Name-new cfapps.sap.hana.ondemand.com -n \'module1Name-testspace\''))
        assertThat(jscr.shell, hasItem('cf map-route module2Name-new cfapps.sap.hana.ondemand.com -n \'module2Name-testspace\''))

        assertThat(jscr.shell, hasItem('cf unmap-route module1Name cfapps.sap.hana.ondemand.com -n \'module1Name-testspace\''))
        assertThat(jscr.shell, hasItem('cf delete-route cfapps.sap.hana.ondemand.com -n module1Name-testspace-new -f'))
        assertThat(jscr.shell, hasItem('cf rename module1Name module1Name-old'))
        assertThat(jscr.shell, hasItem('cf rename module1Name-new module1Name'))
        assertThat(jscr.shell, hasItem('cf delete module1Name-old -f'))

        assertThat(jscr.shell, hasItem('cf unmap-route module2Name cfapps.sap.hana.ondemand.com -n \'module2Name-testspace\''))
        assertThat(jscr.shell, hasItem('cf delete-route cfapps.sap.hana.ondemand.com -n module2Name-testspace-new -f'))
        assertThat(jscr.shell, hasItem('cf delete module2Name-old -f'))
        assertThat(jscr.shell, hasItem('cf rename module2Name module2Name-old'))
        assertThat(jscr.shell, hasItem('cf rename module2Name-new module2Name'))
        assertThat(jscr.shell, hasItem('cf delete module2Name-old -f'))

        assertThat(jlr.log, containsString('Domain already existing.'))
    }

    @Test
    void testServiceBroker(){

        jscr.setReturnValue('cf service-brokers', '')
        jscr.setReturnValue('</dev/urandom tr -dc A-Za-z0-9_ | head -c16', 'generatedPwd')

        def cfModules = [
            [
                cfAppName                 : 'module1Name',
                cfManifestPath            : 'testManifest1.yml',
                registerAsServiceBroker: true
            ],
            [
                cfAppName                 : 'module2Name',
                cfManifestPath            : 'testManifest2.yml',
                registerAsServiceBroker: true,
                serviceBrokerHasSpaceScope: false
            ]
        ]

        jsr.step.deployMultipartAppToCloudFoundry([
            script         : nullScript,
            juStabUtils    : utils,
            cfCredentialsId: 'testCredentials',
            cfOrg          : 'testOrg',
            cfSpace        : 'testSpace',
            modules: cfModules
        ])

        assertThat(jscr.shell, hasItem('cf service-brokers'))
        assertThat(jscr.shell, hasItem('cf push module1Name -n module1Name-testspace -f testManifest1.yml -d cfapps.sap.hana.ondemand.com --no-start'))
        assertThat(jscr.shell, hasItem('</dev/urandom tr -dc A-Za-z0-9_ | head -c16'))
        assertThat(jscr.shell, hasItem('cf set-env module1Name SBF_BROKER_CREDENTIALS "{\\"module1Name-user\\" : \\"generatedPwd\\"}"'))
        assertThat(jscr.shell, hasItem('cf set-env module1Name SBF_CATALOG_SUFFIX testspace'))
        assertThat(jscr.shell, hasItem('cf start module1Name'))
        assertThat(jscr.shell, hasItem('cf create-service-broker module1Name-testspace module1Name-user generatedPwd https://module1Name-testspace.cfapps.sap.hana.ondemand.com --space-scoped'))

        assertThat(jscr.shell, hasItem('cf push module2Name -n module2Name-testspace -f testManifest2.yml -d cfapps.sap.hana.ondemand.com --no-start'))
        assertThat(jscr.shell, hasItem('cf start module2Name'))
        assertThat(jscr.shell, hasItem('cf create-service-broker module2Name-testspace module2Name-user generatedPwd https://module2Name-testspace.cfapps.sap.hana.ondemand.com'))

    }
}

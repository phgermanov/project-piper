#!groovy
package steps

import hudson.AbortException
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertEquals

class SapHeliumDeployTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsEnvironmentRule envRule = new JenkinsEnvironmentRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)

    def heliumDeployCredentialsId = 'heliumCredentialsId'
    def mavenExecuteArgMap = [:]

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(stepRule)
        .around(shellRule)
        .around(envRule)
        .around(new JenkinsErrorRule(this))
        .around(loggingRule)
        .around(thrown)

    @Before
    void init() {
        registerAllowedMethods()
    }

    @After
    void tearDown() {
        mavenExecuteArgMap.clear()

    }

    @Test
    void correctHeliumDeployConfiguration_isSuccessful(){
      stepRule.step.sapHeliumDeploy(script              :  nullScript,
                                   heliumCredentialsId  :  heliumDeployCredentialsId,
                                   host                 :  'ahost.com',
                                   application          :  'myApp',
                                   account              :  'myAccount')
       assertEquals('Goals not set correctly', 'process-sources',mavenExecuteArgMap.goals.toString())
       assertEquals('Flags not set correctly', '--batch-mode --update-snapshots --activate-profiles cloud-deployment',mavenExecuteArgMap.flags.toString())
       assertEquals('Defines not set correctly', '-Dtarget=devsystem -Dcloud.application=myApp -Dcloud.username=myUser -Dcloud.password=******** -Dcloud.account=myAccount -Dcloud.landscape=ahost.com',mavenExecuteArgMap.defines.toString())
   }

   @Test
    void wrongHeliumDeployConfiguration_missingParameter(){
      thrown.expect(AbortException.class)
      stepRule.step.sapHeliumDeploy(script              :   nullScript,
                                   heliumCredentialsId  :   heliumDeployCredentialsId,
                                   host                 :   'ahost.com',
                                   account              :   'myAccount')

   }

def registerAllowedMethods() {
        helper.registerAllowedMethod('usernamePassword', [Map], { m ->
            return m
        })

        helper.registerAllowedMethod('mavenExecute', [Map], { m->
            mavenExecuteArgMap = m
        })

        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
           l.each{
               if (it.credentialsId == heliumDeployCredentialsId) {
                binding.setProperty('username', 'myUser')
                binding.setProperty('password', '********')
              }
            }
            try {
                c()
            } finally {
                binding.setProperty('username', null)
                binding.setProperty('password', null)
            }
        })

    }

}

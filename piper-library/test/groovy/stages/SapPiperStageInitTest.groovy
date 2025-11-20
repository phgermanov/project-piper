#!groovy
package stages

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsStepRule
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertFalse
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue

class SapPiperStageInitTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private ExpectedException thrown = new ExpectedException()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jlr)
        .around(jsr)

    private List stepsCalled = []

    @Before
    void init()  {

        binding.variables.env.STAGE_NAME = 'Init'

        nullScript.commonPipelineEnvironment = this.loadScript('test/resources/openSource/commonPipelineEnvironment.groovy').commonPipelineEnvironment

        helper.registerAllowedMethod('deleteDir', [], null)

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Init'))
            return body()
        })

        helper.registerAllowedMethod('durationMeasure', [Map.class, Closure.class], {m, body ->
            return body()
        })

        helper.registerAllowedMethod('checkout', [Closure.class], {c ->
            stepsCalled.add('checkout')
            return null
        })
        binding.setVariable('scm', {})

        helper.registerAllowedMethod('setupPipelineEnvironment', [Map.class], {m ->
            stepsCalled.add('setupPipelineEnvironment')
        })

        helper.registerAllowedMethod('piperInitRunStageConfiguration', [Map.class], {m ->
            stepsCalled.add('piperInitRunStageConfiguration')
        })

        helper.registerAllowedMethod('artifactPrepareVersion', [Map.class], {m ->
            stepsCalled.add('artifactPrepareVersion')
        })

        helper.registerAllowedMethod('sapGenerateEnvironmentInfo', [Map.class], {m ->
            stepsCalled.add('sapGenerateEnvironmentInfo')
        })
        helper.registerAllowedMethod('sapCumulusUpload', [Map.class], {m ->
            stepsCalled.add('sapCumulusUpload')
        })

        helper.registerAllowedMethod('pipelineStashFilesBeforeBuild', [Map.class], {m ->
            stepsCalled.add('pipelineStashFilesBeforeBuild')
        })

        helper.registerAllowedMethod('gcpPublishEvent', [Map.class], {m ->
            stepsCalled.add('gcpPublishEvent')
        })

        helper.registerAllowedMethod('fileExists', [String.class], {return null})

    }

    @Test
    void testInitNoBuildTool() {

        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR buildTool')
        jsr.step.sapPiperStageInit(script: nullScript, juStabUtils: utils)

    }

    @Test
    void testInitDefault() {

        nullScript.globalPipelineEnvironment.setFlag('piper-lib-os')
        jsr.step.sapPiperStageInit(script: nullScript, juStabUtils: utils, buildTool: 'maven')

        assertThat(stepsCalled, hasItems('checkout', 'setupPipelineEnvironment', 'piperInitRunStageConfiguration', 'artifactPrepareVersion', 'sapGenerateEnvironmentInfo','pipelineStashFilesBeforeBuild', 'gcpPublishEvent'))

    }

    @Test
    void testInitOverwriteDefault() {

        binding.variables.env.BRANCH_NAME = 'testBranch'

        nullScript.globalPipelineEnvironment.setFlag('piper-lib-os')
        jsr.step.sapPiperStageInit(script: nullScript, juStabUtils: utils, buildTool: 'maven')

        assertThat(stepsCalled, hasItems('checkout', 'setupPipelineEnvironment', 'piperInitRunStageConfiguration', 'pipelineStashFilesBeforeBuild'))
        assertThat(stepsCalled, not(hasItems('artifactPrepareVersion', 'gcpPublishEvent')))

    }

    @Test
    void testInitCustomPRActions() {

        binding.variables.env.CHANGE_ID = 'myChangeId'

        binding.variables.pullRequest = new Object() {
            def getLabels() {
                return new Object() {
                    def asList() {
                        return ['pr_karma']
                    }
                }
            }
        }

        nullScript.globalPipelineEnvironment.setFlag('piper-lib-os')
        jsr.step.sapPiperStageInit(script: nullScript, juStabUtils: utils, buildTool: 'maven')

        assertThat(nullScript.globalPipelineEnvironment.configuration.runStep."Pull-Request Voting"."karmaExecuteTests", is(true))
    }

    @Test
    void testInitNoPiperLibOs() {

        thrown.expectMessage('[sapPiperStageInit] Library \'piper-lib-os\' not available. Please configure it according to https://go.sap.corp/piper/lib/setupLibrary/')
        jsr.step.sapPiperStageInit(script: nullScript, juStabUtils: utils, buildTool: 'maven')

    }

    @Test
    void testPullRequestStageStepActivation() {

        nullScript.globalPipelineEnvironment.configuration = [
            runStep: [:]
        ]
        def config = [
            pullRequestStageName: 'Pull-Request Voting',
            stepMappings: [
                karma: 'karmaExecuteTests',
                opa5: 'opa5ExecuteTests'
            ],
            labelPrefix: 'pr_'
        ]

        def actions = ['karma', 'pr_opa5']
        jsr.step.sapPiperStageInit.setPullRequestStageStepActivation(nullScript, config, actions)

        assertThat(nullScript.globalPipelineEnvironment.configuration.runStep."Pull-Request Voting".karmaExecuteTests, is(true))
        assertThat(nullScript.globalPipelineEnvironment.configuration.runStep."Pull-Request Voting".opa5ExecuteTests, is(true))
    }
    @Test
    void testInitWithProductiveBranchPattern() {
        binding.variables.env.BRANCH_NAME = 'anyOtherbranch'
        nullScript.globalPipelineEnvironment.configuration = [
            general: [productiveBranch: '.*branch'],
            runStep: [Init: [slackSendNotification: true]]
        ]
        nullScript.globalPipelineEnvironment.setFlag('piper-lib-os')

        jsr.step.sapPiperStageInit(script: nullScript, juStabUtils: utils, buildTool: 'maven')

        assertThat(stepsCalled, hasItems('artifactPrepareVersion'))
    }

    @Test
    void testLegacyConfigSettings() {
        boolean checkForLegacyConfigurationCalled = false
        helper.registerAllowedMethod('checkForLegacyConfiguration', [Map.class], {
            checkForLegacyConfigurationCalled = true
        })
        nullScript.globalPipelineEnvironment.configuration = [
            general: [legacyConfigSettings: 'com.sap.piper.internal/pipeline/cloudSdkToSAPGppConfigSettings.yml']
        ]
        nullScript.globalPipelineEnvironment.setFlag('piper-lib-os')

        jsr.step.sapPiperStageInit(script: nullScript, juStabUtils: utils, buildTool: 'maven')

        assertTrue(checkForLegacyConfigurationCalled)
    }

    @Test
    void testLegacyConfigSettingsNotEnabled() {
        boolean checkForLegacyConfigurationCalled = false
        helper.registerAllowedMethod('checkForLegacyConfiguration', [Map.class], {
            checkForLegacyConfigurationCalled = true
        })
        nullScript.globalPipelineEnvironment.configuration = [
            general: []
        ]
        nullScript.globalPipelineEnvironment.setFlag('piper-lib-os')

        jsr.step.sapPiperStageInit(script: nullScript, juStabUtils: utils, buildTool: 'maven')

        assertFalse(checkForLegacyConfigurationCalled)
    }

    @Test
    void "Parameter skipCheckout skips the checkout call"() {
        nullScript.globalPipelineEnvironment.setFlag('piper-lib-os')

        jsr.step.sapPiperStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven',
            skipCheckout: true,
            stashContent: 'code'
        )

        assertThat(stepsCalled, hasItems('setupPipelineEnvironment', 'piperInitRunStageConfiguration', 'artifactPrepareVersion', 'pipelineStashFilesBeforeBuild'))
        assertThat(stepsCalled, not(hasItem('checkout')))
    }

    @Test
    void "Try to skip checkout with parameter skipCheckout not boolean throws error"() {
        thrown.expectMessage('[sapPiperStageInit] Parameter skipCheckout has to be of type boolean. Instead got \'java.lang.String\'')

        jsr.step.sapPiperStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven',
            skipCheckout: 'false',
            stashContent: ['code']
        )
    }

    @Test
    void "Try to skip checkout with parameter skipCheckout and missing stashContent"() {
        thrown.expectMessage('[sapPiperStageInit] needs stashes if you skip checkout')

        jsr.step.sapPiperStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven',
            skipCheckout: true
        )
    }

    @Test
    void "Try to skip checkout with parameter skipCheckout and empty array for stashContent"() {
        thrown.expectMessage('[sapPiperStageInit] needs stashes if you skip checkout')

        jsr.step.sapPiperStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven',
            skipCheckout: true,
            stashContent: []
        )
    }

    @Test
    void "Pass parameter scmInfo, skipCheckout skips the checkout call"() {
        nullScript.globalPipelineEnvironment.setFlag('piper-lib-os')

        jsr.step.sapPiperStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven',
            skipCheckout: true,
            stashContent: 'code',
            scmInfo: ["dummyScmKey":"dummyScmKey"]
        )

        assertThat(stepsCalled, hasItems('setupPipelineEnvironment', 'piperInitRunStageConfiguration', 'artifactPrepareVersion', 'pipelineStashFilesBeforeBuild'))
        assertThat(stepsCalled, not(hasItem('checkout')))
    }

    @Test
    void "Passed parameter scmInfo skips the checkout call"() {
        nullScript.globalPipelineEnvironment.setFlag('piper-lib-os')

        jsr.step.sapPiperStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven',
            scmInfo: ["dummyScmKey":"dummyScmKey"]
        )

        assertThat(stepsCalled, hasItems('setupPipelineEnvironment', 'piperInitRunStageConfiguration', 'artifactPrepareVersion', 'pipelineStashFilesBeforeBuild'))
        assertThat(stepsCalled, not(hasItem('checkout')))
    }

    @Test
    void "Not passed parameters scmInfo and skipCheckout don't skip the checkout call"() {
        nullScript.globalPipelineEnvironment.setFlag('piper-lib-os')

        jsr.step.sapPiperStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven'
        )

        assertThat(stepsCalled, hasItems('setupPipelineEnvironment', 'piperInitRunStageConfiguration', 'artifactPrepareVersion', 'pipelineStashFilesBeforeBuild'))
        assertThat(stepsCalled, hasItem('checkout'))
    }
}

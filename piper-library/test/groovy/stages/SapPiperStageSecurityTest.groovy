#!groovy
package stages

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsStepRule
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class SapPiperStageSecurityTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private ExpectedException thrown = new ExpectedException()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jsr)

    private List stepsCalled = []

    @Before
    void init()  {

        binding.variables.env.STAGE_NAME = 'Security'

        helper.registerAllowedMethod("deleteDir", [], null)

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Security'))
            return body()
        })

        helper.registerAllowedMethod('durationMeasure', [Map.class, Closure.class], {m, body ->
            return body()
        })

        def parallelMap = [:]
        helper.registerAllowedMethod("parallel", [Map.class], { map ->
            parallelMap = map
            parallelMap.each {key, value ->
                if (key != 'failFast') {
                    value()
                }
            }
        })

        helper.registerAllowedMethod('executeFortifyScan', [Map.class], {m ->
            stepsCalled.add('executeFortifyScan')
        })

        helper.registerAllowedMethod('executeCheckmarxScan', [Map.class], {m ->
            stepsCalled.add('executeCheckmarxScan')
        })

        helper.registerAllowedMethod('executeOpenSourceDependencyScan', [Map.class], {m ->
            stepsCalled.add('executeOpenSourceDependencyScan')
        })

        helper.registerAllowedMethod('fortifyExecuteScan', [Map.class], {m ->
            stepsCalled.add('fortifyExecuteScan')
        })

        helper.registerAllowedMethod('checkmarxExecuteScan', [Map.class], {m ->
            stepsCalled.add('checkmarxExecuteScan')
        })

        helper.registerAllowedMethod('malwareExecuteScan', [Map.class], {m ->
            stepsCalled.add('malwareExecuteScan')
        })

        helper.registerAllowedMethod('githubCheckBranchProtection', [Map.class], {m ->
            stepsCalled.add('githubCheckBranchProtection')
        })
    }

    @Test
    void testSecurityMta() {

        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'mta'],
            runStep: [Security: [executeFortifyScan: true, executeCheckmarxScan: true]]
        ]

        jsr.step.sapPiperStageSecurity(script: nullScript, juStabUtils: utils, newOSS: false)

        assertThat(stepsCalled, hasItems('executeFortifyScan', 'executeCheckmarxScan', 'executeOpenSourceDependencyScan'))
    }

    @Test
    void testSecurityMaven() {

        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'maven'],
            runStep: [Security: [executeFortifyScan: true, executeCheckmarxScan: true]]
        ]
        jsr.step.sapPiperStageSecurity(script: nullScript, juStabUtils: utils, newOSS: false)

        assertThat(stepsCalled, hasItems('executeFortifyScan', 'executeOpenSourceDependencyScan'))
        assertThat(stepsCalled, not(hasItems('executeCheckmarxScan')))
    }

    @Test
    void testSecurityNpm() {

        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'npm'],
            runStep: [Security: [executeFortifyScan: true, executeCheckmarxScan: true]]
        ]
        jsr.step.sapPiperStageSecurity(script: nullScript, juStabUtils: utils, newOSS: false)

        assertThat(stepsCalled, hasItems('executeCheckmarxScan', 'executeOpenSourceDependencyScan'))
        assertThat(stepsCalled, not(hasItems('executeFortifyScan')))
    }

    @Test
    void testSecurityPip() {

        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'pip'],
            runStep: [Security: [executeFortifyScan: true, executeCheckmarxScan: true]]
        ]
        jsr.step.sapPiperStageSecurity(script: nullScript, juStabUtils: utils, newOSS: false)

        assertThat(stepsCalled, hasItems('executeFortifyScan', 'executeOpenSourceDependencyScan'))
        assertThat(stepsCalled, not(hasItems('executeCheckmarxScan')))
    }

    @Test
    void testSecuritySbt() {

        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'golang'],
            runStep: [Security: [executeFortifyScan: true, executeCheckmarxScan: true]]
        ]
        jsr.step.sapPiperStageSecurity(script: nullScript, juStabUtils: utils, newOSS: false)

        assertThat(stepsCalled, hasItems('executeCheckmarxScan', 'executeOpenSourceDependencyScan'))
        assertThat(stepsCalled, not(hasItems('executeFortifyScan')))
    }

    @Test
    void testSecurityGolang() {

        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'golang'],
            runStep: [Security: [executeFortifyScan: true, executeCheckmarxScan: true]]
        ]
        jsr.step.sapPiperStageSecurity(script: nullScript, juStabUtils: utils, newOSS: false)

        assertThat(stepsCalled, hasItems('executeCheckmarxScan', 'executeOpenSourceDependencyScan'))
        assertThat(stepsCalled, not(hasItems('executeFortifyScan')))
    }

    @Test
    void testOverwriteDefaults() {

        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'testIt'],
            runStep: [Security: [executeFortifyScan: true, executeCheckmarxScan: true]]
        ]
        jsr.step.sapPiperStageSecurity(script: nullScript, juStabUtils: utils, newOSS: false, buildTool: 'testIt', executeFortifyScan: true, executeCheckmarxScan: true, executeOpenSourceDependencyScan: false)

        assertThat(stepsCalled, hasItems('executeCheckmarxScan', 'executeFortifyScan'))
        assertThat(stepsCalled, not(hasItems('executeOpenSourceDependencyScan')))
    }

    @Test
    void testSecurityDockerMultiStage() {

        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'docker'],
            runStep: [Security: [executeFortifyScan: true, executeCheckmarxScan: true]]
        ]
        jsr.step.sapPiperStageSecurity(script: nullScript, juStabUtils: utils, newOSS: false)

        assertThat(stepsCalled, hasItems('executeCheckmarxScan', 'executeOpenSourceDependencyScan', 'executeFortifyScan'))
    }

    @Test
    void testNewFortifyActive() {

        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'maven'],
            runStep: [Security: [fortifyExecuteScan: true, executeFortifyScan: true]]
        ]

        jsr.step.sapPiperStageSecurity(script: nullScript, juStabUtils: utils, newOSS: false)

        assertThat(stepsCalled, hasItem('fortifyExecuteScan'))
        assertThat(stepsCalled, not(hasItem('githubCheckBranchProtection')))
    }

    @Test
    void testNewFortifyVerifyOnly() {

        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'maven', githubTokenCredentialsId: 'theToken'],
            runStep: [Security: [fortifyExecuteScan: true, executeFortifyScan: true]]
        ]

        jsr.step.sapPiperStageSecurity(script: nullScript, juStabUtils: utils, newOSS: false)

        assertThat(stepsCalled, allOf(hasItem('fortifyExecuteScan'), hasItem('githubCheckBranchProtection')))
    }

    @Test
    void testNewCheckmarxActive() {

        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'npm'],
            runStep: [Security: [checkmarxExecuteScan: true, executeCheckmarxScan: true]]
        ]

        jsr.step.sapPiperStageSecurity(script: nullScript, juStabUtils: utils, newOSS: false)

        assertThat(stepsCalled, hasItems('checkmarxExecuteScan'))
        assertThat(stepsCalled, not(hasItem('githubCheckBranchProtection')))
    }

    @Test
    void testNewCheckmarxVerifyOnly() {

        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'npm', githubTokenCredentialsId: 'theToken'],
            runStep: [Security: [checkmarxExecuteScan: true, executeCheckmarxScan: true]]
        ]


        jsr.step.sapPiperStageSecurity(script: nullScript, juStabUtils: utils, newOSS: false)

        assertThat(stepsCalled, allOf(hasItem('checkmarxExecuteScan'), hasItem('githubCheckBranchProtection')))
    }
}

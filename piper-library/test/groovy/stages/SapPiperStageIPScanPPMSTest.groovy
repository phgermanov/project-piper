#!groovy
package stages

import org.junit.After
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

class SapPiperStageIPScanPPMSTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private ExpectedException thrown = new ExpectedException()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jsr)

    private List stepsCalled = []

    @Before
    void init() {

        binding.variables.env.STAGE_NAME = 'IPScan and PPMS'

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('IPScan and PPMS'))
            return body()
        })

        helper.registerAllowedMethod('durationMeasure', [Map.class, Closure.class], {m, body ->
            return body()
        })

        helper.registerAllowedMethod('executePPMSComplianceCheck', [Map.class], {m ->
            stepsCalled.add('executePPMSComplianceCheck')
        })

        helper.registerAllowedMethod('sapCheckPPMSCompliance', [Map.class], {m ->
            stepsCalled.add('sapCheckPPMSCompliance')
        })

        helper.registerAllowedMethod('whitesourceExecuteScan', [Map.class], {m ->
            stepsCalled.add('whitesourceExecuteScan')
        })
    }

    @After
    void cleanUp() {
        nullScript.globalPipelineEnvironment.configuration = [steps: [:]]
    }


    @Test
    void testIPScanDefault() {

        jsr.step.sapPiperStageIPScanPPMS(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, not(hasItems('whitesourceExecuteScan', 'executePPMSComplianceCheck')))
    }

    @Test
    void testIPScanOverwriteDefault() {

        jsr.step.sapPiperStageIPScanPPMS(script: nullScript, juStabUtils: utils, executePPMSComplianceCheck: true, whitesourceExecuteScan: true)

        assertThat(stepsCalled, hasItems('whitesourceExecuteScan', 'executePPMSComplianceCheck'))
    }

    @Test
    void testPPMSActive() {

        nullScript.globalPipelineEnvironment.configuration = [runStep: ['IPScan and PPMS': [
            whitesourceExecuteScan: true,
            executePPMSComplianceCheck: true
        ]]]

        jsr.step.sapPiperStageIPScanPPMS(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItems('whitesourceExecuteScan', 'executePPMSComplianceCheck'))
    }

    @Test
    void testNewPPMSActive() {

        nullScript.globalPipelineEnvironment.configuration = [
            runStep: ['IPScan and PPMS': [
                whitesourceExecuteScan: true,
                sapCheckPPMSCompliance: true,
                executePPMSComplianceCheck: true
            ]],
        ]

        jsr.step.sapPiperStageIPScanPPMS(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItems('whitesourceExecuteScan', 'sapCheckPPMSCompliance'))
        assertThat(stepsCalled, not(hasItem('executePPMSComplianceCheck')))
    }

    @Test
    void testNewWhiteSource() {

        nullScript.globalPipelineEnvironment.configuration = [
            steps: [whitesourceExecuteScan: [test: 'test']],
            runStep: ['IPScan and PPMS': [
                whitesourceExecuteScan: true
            ]],
        ]

        jsr.step.sapPiperStageIPScanPPMS(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, (hasItem('whitesourceExecuteScan')))
    }
}

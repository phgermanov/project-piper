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

class SapPiperStageAcceptanceTest extends BasePiperTest {
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

        binding.variables.env.STAGE_NAME = 'Acceptance'

        helper.registerAllowedMethod("findFiles", [Map.class], { map ->
            return [].toArray()
        })

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Acceptance'))
            return body()
        })

        helper.registerAllowedMethod('durationMeasure', [Map.class, Closure.class], {m, body ->
            return body()
        })

        helper.registerAllowedMethod('multicloudDeploy', [Map.class], {m ->
            stepsCalled.add('multicloudDeploy')
        })

        helper.registerAllowedMethod('cloudFoundryCreateService', [Map.class], {m ->
            stepsCalled.add('cloudFoundryCreateService')
        })

        helper.registerAllowedMethod('manageCloudFoundryEnvironment', [Map.class], {m ->
            stepsCalled.add('manageCloudFoundryEnvironment')
        })

        helper.registerAllowedMethod('downloadArtifactsFromNexus', [Map.class], {m ->
            stepsCalled.add('downloadArtifactsFromNexus')
        })

        helper.registerAllowedMethod('cloudFoundryDeploy', [Map.class], {m ->
            stepsCalled.add('cloudFoundryDeploy')
        })

        helper.registerAllowedMethod('kubernetesDeploy', [Map.class], {m ->
            stepsCalled.add('kubernetesDeploy')
        })

        helper.registerAllowedMethod('healthExecuteCheck', [Map.class], {m ->
            stepsCalled.add('healthExecuteCheck')
        })

        helper.registerAllowedMethod('newmanExecute', [Map.class], {m ->
            stepsCalled.add('newmanExecute')
        })

        helper.registerAllowedMethod('uiVeri5ExecuteTests', [Map.class], {m ->
            stepsCalled.add('uiVeri5ExecuteTests')
        })

        helper.registerAllowedMethod('gaugeExecuteTests', [Map.class], {m ->
            stepsCalled.add('gaugeExecuteTests')
        })

        helper.registerAllowedMethod('npmExecuteEndToEndTests', [Map.class], {m ->
            stepsCalled.add('npmExecuteEndToEndTests')
        })

        helper.registerAllowedMethod('npmExecuteScripts', [Map.class], {m ->
            stepsCalled.add('npmExecuteScripts')
        })

        helper.registerAllowedMethod('testsPublishResults', [Map.class], {m ->
            stepsCalled.add('testsPublishResults')
        })

        helper.registerAllowedMethod('sapCreateTraceabilityReport', [Map.class], {m ->
            stepsCalled.add('sapCreateTraceabilityReport')
        })

    }

    @Test
    void testAcceptanceDefault() {

        jsr.step.sapPiperStageAcceptance(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, not(anyOf(hasItem('multicloudDeploy'), hasItem('manageCloudFoundryEnvironment'), hasItem('cloudFoundryDeploy'), hasItem('healthExecuteCheck'), hasItem('newmanExecute'), hasItem('uiVeri5ExecuteTests'), hasItem('gaugeExecuteTests'), hasItem('testsPublishResults'), hasItem('sapCreateTraceabilityReport'), hasItem('npmExecuteScripts'))))
    }

    @Test
    void testAcceptanceEndToEndTests() {

        jsr.step.sapPiperStageAcceptance(script: nullScript, juStabUtils: utils, npmExecuteEndToEndTests: true)

        assertThat(stepsCalled, hasItem('npmExecuteEndToEndTests'))
        assertThat(stepsCalled, not(anyOf(hasItem('manageCloudFoundryEnvironment'), hasItem('cloudFoundryDeploy'), hasItem('healthExecuteCheck'), hasItem('newmanExecute'), hasItem('uiVeri5ExecuteTests'), hasItem('gaugeExecuteTests'), hasItem('sapCreateTraceabilityReport'), hasItem('testsPublishResults'), hasItem('npmExecuteScripts'))))

    }

    @Test
    void testAcceptanceTestsWithNpmExecuteScripts() {

        jsr.step.sapPiperStageAcceptance(script: nullScript, juStabUtils: utils, npmExecuteScripts: true)

        assertThat(stepsCalled, hasItem('npmExecuteScripts'))
        assertThat(stepsCalled, not(anyOf(hasItem('manageCloudFoundryEnvironment'), hasItem('cloudFoundryDeploy'), hasItem('healthExecuteCheck'), hasItem('newmanExecute'), hasItem('uiVeri5ExecuteTests'), hasItem('gaugeExecuteTests'), hasItem('sapCreateTraceabilityReport'), hasItem('npmExecuteEndToEndTests'))))

    }

    @Test
    void testAcceptanceOverwriteDefaults() {

        jsr.step.sapPiperStageAcceptance(script: nullScript, juStabUtils: utils, downloadArtifactsFromNexus: true, manageCloudFoundryEnvironment: true, cloudFoundryDeploy: true, newmanExecute: true, uiVeri5ExecuteTests: true, gaugeExecuteTests: true, sapCreateTraceabilityReport: true)

        assertThat(stepsCalled, allOf(hasItem('downloadArtifactsFromNexus'), hasItem('manageCloudFoundryEnvironment'), hasItem('cloudFoundryDeploy'), hasItem('healthExecuteCheck'), hasItem('newmanExecute'), hasItem('uiVeri5ExecuteTests'), hasItem('gaugeExecuteTests'), hasItem('testsPublishResults'), hasItem('sapCreateTraceabilityReport')))
    }

    @Test
    void testPerformanceRunsCFCreateServiceDefault() {

        jsr.step.sapPiperStageAcceptance(script: nullScript, juStabUtils: utils, downloadArtifactsFromNexus: true, cloudFoundryDeploy: true, cloudFoundryCreateService: true)

        assertThat(stepsCalled, allOf(hasItem('downloadArtifactsFromNexus'), hasItem('cloudFoundryCreateService'), hasItem('cloudFoundryDeploy'), hasItem('healthExecuteCheck')))
    }

    @Test
    void testPerformanceUsesConfigToTriggerCreateService() {

        nullScript.globalPipelineEnvironment.configuration = [
            runStep: [Acceptance: [cloudFoundryCreateService: true, downloadArtifactsFromNexus: true]]
        ]

        jsr.step.sapPiperStageAcceptance(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, allOf(hasItem('cloudFoundryCreateService'),hasItem('downloadArtifactsFromNexus')))

    }

    @Test
    void testAcceptanceCFDeploy() {

        nullScript.globalPipelineEnvironment.configuration = [
            runStep: [Acceptance: [cloudFoundryDeploy: true, downloadArtifactsFromNexus: true]]]


        jsr.step.sapPiperStageAcceptance(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, allOf(hasItem('downloadArtifactsFromNexus'), hasItem('cloudFoundryDeploy'), hasItem('healthExecuteCheck')))
        assertThat(stepsCalled, not(anyOf(hasItem('newmanExecute'), hasItem('uiVeri5ExecuteTests'), hasItem('gaugeExecuteTests'), hasItem('testsPublishResults'), hasItem('sapCreateTraceabilityReport'))))

    }

    @Test
    void testAcceptanceKubernetesDeploy() {

        helper.registerAllowedMethod('sleep', [Integer.class], null)
        helper.registerAllowedMethod('retry', [Integer.class, Closure.class], null)

        nullScript.globalPipelineEnvironment.configuration = [runStep: [Acceptance: [kubernetesDeploy: true]]]

        jsr.step.sapPiperStageAcceptance(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItem('kubernetesDeploy'))
        assertThat(stepsCalled, not(anyOf(hasItem('healthExecuteCheck'), hasItem('cloudFoundryDeploy'), hasItem('newmanExecute'), hasItem('uiVeri5ExecuteTests'), hasItem('gaugeExecuteTests'), hasItem('testsPublishResults'), hasItem('sapCreateTraceabilityReport'))))

    }

    @Test
    void testSkipHealthCheckDuringAcceptanceKubernetesDeployDefinedInStagesAcceptance() {

        helper.registerAllowedMethod('sleep', [Integer.class], null)
        helper.registerAllowedMethod('retry', [Integer.class, Closure.class], null)

        nullScript.globalPipelineEnvironment.configuration = [
            general: [k8sHealthExecuteCheck:false],
            runStep: [Acceptance: [kubernetesDeploy: true]]
        ]

        jsr.step.sapPiperStageAcceptance(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItem('kubernetesDeploy'))
        assertThat(stepsCalled, not(anyOf(hasItem('deployToKubernetes'),hasItem('healthExecuteCheck'), hasItem('cloudFoundryDeploy'), hasItem('newmanExecute'), hasItem('uiVeri5ExecuteTests'), hasItem('gaugeExecuteTests'), hasItem('testsPublishResults'), hasItem('sapCreateTraceabilityReport'))))
    }

    @Test
    void testAcceptanceNewman() {

        nullScript.globalPipelineEnvironment.configuration = [
            runStep: [Acceptance: [newmanExecute: true, downloadArtifactsFromNexus: true]]]

        jsr.step.sapPiperStageAcceptance(script: nullScript, juStabUtils: utils, cloudFoundryDeploy: true)

        assertThat(stepsCalled, allOf(hasItem('downloadArtifactsFromNexus'), hasItem('cloudFoundryDeploy'), hasItem('healthExecuteCheck'), hasItem('newmanExecute'), hasItem('testsPublishResults')))
        assertThat(stepsCalled, not(anyOf(hasItem('uiVeri5ExecuteTests'), hasItem('gaugeExecuteTests'), hasItem('sapCreateTraceabilityReport'))))

    }

    @Test
    void testAcceptanceUiVeri5() {

        nullScript.globalPipelineEnvironment.configuration = [
            runStep: [Acceptance: [uiVeri5ExecuteTests: true, downloadArtifactsFromNexus: true]]]

        jsr.step.sapPiperStageAcceptance(script: nullScript, juStabUtils: utils, cloudFoundryDeploy: true)

        assertThat(stepsCalled, allOf(hasItem('downloadArtifactsFromNexus'), hasItem('cloudFoundryDeploy'), hasItem('healthExecuteCheck'), hasItem('uiVeri5ExecuteTests'), hasItem('testsPublishResults')))
        assertThat(stepsCalled, not(anyOf(hasItem('newmanExecute'), hasItem('gaugeExecuteTests'), hasItem('sapCreateTraceabilityReport'))))

    }

    @Test
    void testAcceptanceTraceability() {

        nullScript.globalPipelineEnvironment.configuration = [
            runStep: [Acceptance: [sapCreateTraceabilityReport: true, downloadArtifactsFromNexus: true]]]

        jsr.step.sapPiperStageAcceptance(script: nullScript, juStabUtils: utils, cloudFoundryDeploy: true)

        assertThat(stepsCalled, allOf(hasItem('downloadArtifactsFromNexus'), hasItem('cloudFoundryDeploy'), hasItem('healthExecuteCheck'), hasItem('sapCreateTraceabilityReport')))
        assertThat(stepsCalled, not(anyOf(hasItem('newmanExecute'), hasItem('gaugeExecuteTests'), hasItem('uiVeri5ExecuteTests'), hasItem('testsPublishResults'))))

    }

    @Test
    void testAcceptanceMultiCloud() {

        jsr.step.sapPiperStageAcceptance(script: nullScript, juStabUtils: utils, multicloudDeploy: true)

        assertThat(stepsCalled, hasItem('multicloudDeploy'))
        assertThat(stepsCalled, not(anyOf(hasItem('manageCloudFoundryEnvironment'), hasItem('cloudFoundryDeploy'), hasItem('healthExecuteCheck'), hasItem('newmanExecute'), hasItem('uiVeri5ExecuteTests'), hasItem('gaugeExecuteTests'), hasItem('testsPublishResults'), hasItem('sapCreateTraceabilityReport'))))
    }

    @Test
    void testDisableHealthCheck() {

        jsr.step.sapPiperStageRelease(script: nullScript, juStabUtils: utils, healthExecuteCheck: false)

        assertThat(stepsCalled, not(hasItem('healthExecuteCheck')))
    }
}

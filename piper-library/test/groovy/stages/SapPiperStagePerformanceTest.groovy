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

class SapPiperStagePerformanceTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private ExpectedException thrown = new ExpectedException()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jsr)

    private List stepsCalled = []
    private Map testPublishing = [:]

    @Before
    void init()  {

        binding.variables.env.STAGE_NAME = 'Performance'

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Performance'))
            return body()
        })

        helper.registerAllowedMethod('durationMeasure', [Map.class, Closure.class], {m, body ->
            return body()
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

    }

    @Test
    void testPerformanceTestsDefault() {

        jsr.step.sapPiperStagePerformance(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItem('downloadArtifactsFromNexus'))
        assertThat(stepsCalled, not(anyOf(hasItem('manageCloudFoundryEnvironment'), hasItem('cloudFoundryDeploy'), hasItem('healthExecuteCheck'))))

    }

    @Test
    void testPerformanceTestsOverwriteDefault() {

        jsr.step.sapPiperStagePerformance(script: nullScript, juStabUtils: utils, downloadArtifactsFromNexus: false, manageCloudFoundryEnvironment: true, cloudFoundryDeploy: true)

        assertThat(stepsCalled, allOf(hasItem('manageCloudFoundryEnvironment'), hasItem('cloudFoundryDeploy'), hasItem('healthExecuteCheck')))
        assertThat(stepsCalled, not(hasItem('downloadArtifactsFromNexus')))

    }

    @Test
    void testPerformanceRunsCFCreateServiceDefault() {

        jsr.step.sapPiperStagePerformance(script: nullScript, juStabUtils: utils, downloadArtifactsFromNexus: false, cloudFoundryDeploy: true, cloudFoundryCreateService: true)

        assertThat(stepsCalled, allOf(hasItem('cloudFoundryCreateService'), hasItem('cloudFoundryDeploy'), hasItem('healthExecuteCheck')))
        assertThat(stepsCalled, not(hasItem('downloadArtifactsFromNexus')))

    }

    @Test
    void testPerformanceUsesConfigToTriggerCreateService() {

        nullScript.globalPipelineEnvironment.configuration = [
            runStep: [Performance: [cloudFoundryCreateService: true]]
        ]

        jsr.step.sapPiperStagePerformance(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, allOf(hasItem('cloudFoundryCreateService'), hasItem('downloadArtifactsFromNexus')))

    }

    @Test
    void testPerformanceCFDeploy() {

        nullScript.globalPipelineEnvironment.configuration = [runStep: [Performance: [cloudFoundryDeploy: true]]]

        jsr.step.sapPiperStagePerformance(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, allOf(hasItem('downloadArtifactsFromNexus'), hasItem('cloudFoundryDeploy'), hasItem('healthExecuteCheck')))
        assertThat(stepsCalled, not(anyOf(hasItem('manageCloudFoundryEnvironment'), hasItem('deployToKubernetes'))))

    }

    @Test
    void testPerformanceKubernetesDeploy() {

        helper.registerAllowedMethod('sleep', [Integer.class], null)
        helper.registerAllowedMethod('retry', [Integer.class, Closure.class], null)

        nullScript.globalPipelineEnvironment.configuration = [runStep: [Performance: [kubernetesDeploy: true]]]

        jsr.step.sapPiperStagePerformance(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItem('kubernetesDeploy'))
        assertThat(stepsCalled, not(anyOf(hasItem('healthExecuteCheck'), hasItem('deployToKubernetes'), hasItem('manageCloudFoundryEnvironment'), hasItem('cloudFoundryDeploy'))))
    }

    @Test
    void testSkipHealthCheckDuringPerformanceKubernetesDeploy() {

        helper.registerAllowedMethod('sleep', [Integer.class], null)
        helper.registerAllowedMethod('retry', [Integer.class, Closure.class], null)

        nullScript.globalPipelineEnvironment.configuration = [
            general: [k8sHealthExecuteCheck:false],
            runStep: [Performance: [kubernetesDeploy: true]]
        ]

        jsr.step.sapPiperStagePerformance(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItem('kubernetesDeploy'))
        assertThat(stepsCalled, not(anyOf(hasItem('deployToKubernetes'), hasItem('manageCloudFoundryEnvironment'), hasItem('cloudFoundryDeploy'))))
    }

    @Test
    void testDisableHealthCheck() {

        jsr.step.sapPiperStageRelease(script: nullScript, juStabUtils: utils, healthExecuteCheck: false)

        assertThat(stepsCalled, not(hasItem('healthExecuteCheck')))
    }

}

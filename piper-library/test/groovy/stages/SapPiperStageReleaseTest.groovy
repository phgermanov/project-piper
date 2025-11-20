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

class SapPiperStageReleaseTest extends BasePiperTest {
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

        binding.variables.env.STAGE_NAME = 'Release'

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Release'))
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

        helper.registerAllowedMethod('githubPublishRelease', [Map.class], {m ->
            stepsCalled.add('githubPublishRelease')
        })

    }

    @Test
    void testReleaseDefault() {

        jsr.step.sapPiperStageRelease(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItem('downloadArtifactsFromNexus'))
        assertThat(stepsCalled, not(anyOf(hasItem('multicloudDeploy'), hasItem('manageCloudFoundryEnvironment'), hasItem('cloudFoundryDeploy'), hasItem('healthExecuteCheck'), hasItem('githubPublishRelease'))))

    }

    @Test
    void testReleaseOverwriteDefault() {

        jsr.step.sapPiperStageRelease(script: nullScript, juStabUtils: utils, downloadArtifactsFromNexus: false, cloudFoundryDeploy: true, manageCloudFoundryEnvironment: true, githubPublishRelease: true)

        assertThat(stepsCalled, allOf(hasItem('manageCloudFoundryEnvironment'), hasItem('cloudFoundryDeploy'), hasItem('healthExecuteCheck'), hasItem('githubPublishRelease')))
        assertThat(stepsCalled, not(hasItem('downloadArtifactsFromNexus')))

    }

    @Test
    void testReleaseRunsCFCreateServiceDefault() {

        jsr.step.sapPiperStageRelease(script: nullScript, juStabUtils: utils, downloadArtifactsFromNexus: false, cloudFoundryDeploy: true, cloudFoundryCreateService: true, githubPublishRelease: true)

        assertThat(stepsCalled, allOf(hasItem('cloudFoundryCreateService'), hasItem('cloudFoundryDeploy'), hasItem('healthExecuteCheck'), hasItem('githubPublishRelease')))
        assertThat(stepsCalled, not(anyOf(hasItem('downloadArtifactsFromNexus'), hasItem('kubernetesDeploy'))))

    }

    @Test
    void testReleaseUsesConfigToTriggerCreateService() {

        nullScript.globalPipelineEnvironment.configuration = [
            runStep: [Release: [cloudFoundryCreateService: true]]
        ]

        jsr.step.sapPiperStageRelease(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, allOf(hasItem('cloudFoundryCreateService'), hasItem('downloadArtifactsFromNexus')))

    }

    @Test
    void testReleaseUseConfig() {

        nullScript.globalPipelineEnvironment.configuration = [
            runStep: [Release: [cloudFoundryDeploy: true]]
        ]

        jsr.step.sapPiperStageRelease(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, allOf(hasItem('downloadArtifactsFromNexus'), hasItem('cloudFoundryDeploy'), hasItem('healthExecuteCheck')))

    }

    @Test
    void testReleaseCFDeploy() {

        nullScript.globalPipelineEnvironment.configuration = [runStep: [Release: [cloudFoundryDeploy: true]]]

        jsr.step.sapPiperStageRelease(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, allOf(hasItem('downloadArtifactsFromNexus'), hasItem('cloudFoundryDeploy'), hasItem('healthExecuteCheck')))
        assertThat(stepsCalled, not(anyOf(hasItem('manageCloudFoundryEnvironment'), hasItem('deployToKubernetes'))))

    }

    @Test
    void testReleaseKubernetesDeploy() {

        helper.registerAllowedMethod('sleep', [Integer.class], null)
        helper.registerAllowedMethod('retry', [Integer.class, Closure.class], null)

        nullScript.globalPipelineEnvironment.configuration = [runStep: [Release: [kubernetesDeploy: true]]]

        jsr.step.sapPiperStageRelease(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItem('kubernetesDeploy'))
        assertThat(stepsCalled, not(anyOf(hasItem('healthExecuteCheck'), hasItem('deployToKubernetes'), hasItem('downloadArtifactsFromNexus'), hasItem('manageCloudFoundryEnvironment'), hasItem('cloudFoundryDeploy'))))

    }

    @Test
    void testSkipHealthCheckDuringReleaseKubernetesDeploy() {

        helper.registerAllowedMethod('sleep', [Integer.class], null)
        helper.registerAllowedMethod('retry', [Integer.class, Closure.class], null)

        nullScript.globalPipelineEnvironment.configuration = [
            general: [k8sHealthExecuteCheck:false],
            runStep: [Release: [kubernetesDeploy: true]]
        ]

        jsr.step.sapPiperStageRelease(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItem('kubernetesDeploy'))
        assertThat(stepsCalled, not(anyOf(hasItem('healthExecuteCheck'), hasItem('deployToKubernetes'), hasItem('downloadArtifactsFromNexus'), hasItem('manageCloudFoundryEnvironment'), hasItem('cloudFoundryDeploy'))))
    }

    @Test
    void testReleaseMultiCloud() {

        jsr.step.sapPiperStageRelease(script: nullScript, juStabUtils: utils, multicloudDeploy: true)

        assertThat(stepsCalled, allOf(hasItem('multicloudDeploy'), hasItem('healthExecuteCheck')))
        assertThat(stepsCalled, not(anyOf(hasItem('manageCloudFoundryEnvironment'), hasItem('cloudFoundryDeploy'), hasItem('githubPublishRelease'))))
    }

    @Test
    void testDisableHealthCheck() {

        jsr.step.sapPiperStageRelease(script: nullScript, juStabUtils: utils, healthExecuteCheck: false)

        assertThat(stepsCalled, not(hasItem('healthExecuteCheck')))
    }

}

#!groovy
package stages

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsStepRule
import util.PipelineWhenException
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class SapPiperStageCentralBuildTest extends BasePiperTest {
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

        binding.variables.env.STAGE_NAME = 'Central Build'

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Central Build'))
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

        helper.registerAllowedMethod("fileExists", [String.class], { s ->
            return false
        })

        helper.registerAllowedMethod('pipelineStashFilesAfterBuild', [Map.class], {m ->
            stepsCalled.add('pipelineStashFilesAfterBuild')
        })

        helper.registerAllowedMethod('hadolintExecute', [Map.class], {m ->
            stepsCalled.add('hadolintExecute')
        })

        helper.registerAllowedMethod('executeBuild', [Map.class], {m ->
            stepsCalled.add('executeBuild')
        })

        helper.registerAllowedMethod('mavenExecuteStaticCodeChecks', [Map.class], {m ->
            stepsCalled.add('mavenExecuteStaticCodeChecks')
        })

        helper.registerAllowedMethod('npmExecuteLint', [Map.class], {m ->
            stepsCalled.add('npmExecuteLint')
        })

        helper.registerAllowedMethod('pushToDockerRegistry', [Map.class], {m ->
            stepsCalled.add('pushToDockerRegistry')
        })

        helper.registerAllowedMethod('checksPublishResults', [Map.class], {m ->
            stepsCalled.add('checksPublishResults')
        })

        helper.registerAllowedMethod('testsPublishResults', [Map.class], {m ->
            stepsCalled.add('testsPublishResults')
        })

        helper.registerAllowedMethod('parallel', [Map], { Map m -> m.entrySet().each { e -> if (e.key != 'failFast') e.value.call(this) } } )
    }

    @Test
    void testCentralBuildDefault() {

        jsr.step.sapPiperStageCentralBuild(script: nullScript, juStabUtils: utils, buildTool: 'maven')

        assertThat(stepsCalled, hasItems('executeBuild', 'pipelineStashFilesAfterBuild', 'checksPublishResults', 'testsPublishResults'))

    }

    @Test
    void testCentralBuildOverwriteDefault() {

        jsr.step.sapPiperStageCentralBuild(script: nullScript, juStabUtils: utils, buildTool: 'maven')

        assertThat(stepsCalled, hasItems('executeBuild', 'pipelineStashFilesAfterBuild', 'checksPublishResults', 'testsPublishResults'))

    }

    @Test
    void testCentralBuildWithLinting() {


        nullScript.globalPipelineEnvironment.configuration = [stages: ['Central Build': [npmExecuteLint: true, mavenExecuteStaticCodeChecks: true]]]

        jsr.step.sapPiperStageCentralBuild(script: nullScript, juStabUtils: utils, buildTool: 'maven')

        assertThat(stepsCalled, hasItem('npmExecuteLint'))
        assertThat(stepsCalled, hasItem('mavenExecuteStaticCodeChecks'))

    }

    @Test
    void testCentralBuildCheckmarxStashing() {

        helper.registerAllowedMethod('pipelineStashFilesAfterBuild', [Map.class], {m ->
            assertThat(m.runCheckmarx, is(true))
        })

        nullScript.globalPipelineEnvironment.configuration = [runStep: [Security: [executeCheckmarxScan: true]]]
        nullScript.globalPipelineEnvironment.setFlag('piper-lib-os')

        jsr.step.sapPiperStageCentralBuild(script: nullScript, juStabUtils: utils, buildTool: 'maven')

    }

    @Test
    void testCentralBuildNoCheckmarxStashing() {

        helper.registerAllowedMethod('pipelineStashFilesAfterBuild', [Map.class], {m ->
            assertThat(m.runCheckmarx, isEmptyOrNullString())
        })

        nullScript.globalPipelineEnvironment.setFlag('piper-lib-os')
        jsr.step.sapPiperStageCentralBuild(script: nullScript, juStabUtils: utils, buildTool: 'maven')

    }

    @Test
    void testCentralBuildWithPushToDockerregistry() {

        helper.registerAllowedMethod("fileExists", [String.class], { s ->
            if (s == 'docker.metadata.json')
                return true
            else
                return false
        })

        nullScript.globalPipelineEnvironment.configuration = [runStep: ['Central Build': [pushToDockerRegistry: true, hadolintExecute: true]]]

        helper.registerAllowedMethod('pipelineStashFilesAfterBuild', [Map.class], {m ->
            assertThat(m.runCheckmarx, isEmptyOrNullString())
        })

        nullScript.globalPipelineEnvironment.setFlag('piper-lib-os')
        jsr.step.sapPiperStageCentralBuild(script: nullScript, juStabUtils: utils, buildTool: 'maven')

        assertThat(stepsCalled, hasItems('hadolintExecute', 'pushToDockerRegistry'))

    }

}

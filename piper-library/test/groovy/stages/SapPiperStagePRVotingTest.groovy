#!groovy
package stages

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsEnvironmentRule
import util.JenkinsExecuteDockerRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class SapPiperStagePRVotingTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private ExpectedException thrown = new ExpectedException()
    private JenkinsExecuteDockerRule jedr = new JenkinsExecuteDockerRule(this)
    private JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jer)
        .around(jscr)
        .around(jedr)
        .around(jsr)

    private List stepsCalled = []
    private List credentials = []
    private Map stepParameters = [:]

    @Before
    void init()  {

        helper.registerAllowedMethod("fileExists", [String.class], {map -> return false})
        helper.registerAllowedMethod("httpRequest", [String.class], {string ->return [status: 200, content: 'testContent']})
        helper.registerAllowedMethod("writeFile", [HashMap.class], {map -> return})


        binding.variables.env = [
            STAGE_NAME: 'Pull-Request Voting',
            BRANCH_NAME: 'PR-1'
        ]

        helper.registerAllowedMethod('usernamePassword', [Map], { m ->
            return m
        })

        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            credentials.add(l)
            l.each {
                if (it.credentialsId == 'testCredentials') {
                    binding.setProperty('GITHUB_USERNAME', 'myUser')
                    binding.setProperty('GITHUB_TOKEN', '*****')
                }
            }

            try {
                c()
            } finally {
                binding.setProperty('GITHUB_USERNAME', null)
                binding.setProperty('GITHUB_TOKEN', null)
            }
        })

        helper.registerAllowedMethod("findFiles", [Map.class], { map ->
            switch (map.glob) {
                case 'pom.xml':
                    return [new File('pom.xml')].toArray()
                default:
                    return [].toArray()
            }
        })

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Pull-Request Voting'))
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

        helper.registerAllowedMethod('mavenExecuteStaticCodeChecks', [Map.class], {m ->
            stepsCalled.add('mavenExecuteStaticCodeChecks')
        })

        helper.registerAllowedMethod('mavenExecuteIntegration', [Map.class], {m ->
            stepsCalled.add('mavenExecuteIntegration')
        })

        helper.registerAllowedMethod('npmExecuteLint', [Map.class], {m ->
            stepsCalled.add('npmExecuteLint')
        })

        helper.registerAllowedMethod('npmExecuteScripts', [Map.class], {m ->
            stepsCalled.add('npmExecuteScripts')
        })

        helper.registerAllowedMethod('checksPublishResults', [Map.class], {m ->
            stepsCalled.add('checksPublishResults')
        })

        helper.registerAllowedMethod('testsPublishResults', [Map.class], {m ->
            stepsCalled.add('testsPublishResults')
        })

        helper.registerAllowedMethod('kanikoExecute', [Map.class], {m ->
            stepsCalled.add('kanikoExecute')
        })

        helper.registerAllowedMethod('golangBuild', [Map.class], {m ->
            stepsCalled.add('golangBuild')
        })

        helper.registerAllowedMethod('karmaExecuteTests', [Map.class], {m ->
            stepsCalled.add('karmaExecuteTests')
        })

        helper.registerAllowedMethod('executeFortifyScan', [Map.class], {m ->
            stepsCalled.add('executeFortifyScan')
            stepParameters.executeFortifyScan = m
        })

        helper.registerAllowedMethod('executePPMSComplianceCheck', [Map.class], {m ->
            stepsCalled.add('executePPMSComplianceCheck')
            stepParameters.executePPMSComplianceCheck = m
        })

        helper.registerAllowedMethod('whitesourceExecuteScan', [Map.class], {m ->
            stepsCalled.add('whitesourceExecuteScan')
            stepParameters.whitesourceExecuteScan = m
            m.script.commonPipelineEnvironment.setValue('whitesourceProjectNames', ['ws project - PR1'])

        })

        helper.registerAllowedMethod('writeTemporaryCredentials', [Map.class, Closure.class], {m, body ->
            stepsCalled.add('writeTemporaryCredentials')
            body()
        })

    }

    @Test
    void testPRVotingDefault() {

        nullScript.globalPipelineEnvironment.setFlag('piper-lib-os')
        nullScript.globalPipelineEnvironment.configuration = [general: [buildTool: 'maven']]
        jsr.step.sapPiperStagePRVoting(script: nullScript, juStabUtils: utils, dockerEnvVars: [Test: 'test'], stashBackConfig: [excludes:'test.out'])

        assertThat(stepsCalled, hasItems('checksPublishResults', 'testsPublishResults'))
        assertThat(stepsCalled, not(hasItems('karmaExecuteTests', 'npmExecuteScripts', 'mavenExecuteIntegration')))
        assertThat(jscr.shell, hasItem(containsString('mvn --global-settings .pipeline/mavenGlobalSettings.xml --batch-mode clean verify')))
        assertThat(jedr.dockerParams.dockerImage, is('maven:3.6-jdk-8'))
        assertThat(jedr.dockerParams.dockerEnvVars, is([Test: 'test']))
    }

    @Test
    void testPRVotingWithCustomSteps() {

        nullScript.globalPipelineEnvironment.setFlag('piper-lib-os')
        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'maven'],
            runStep: ['Pull-Request Voting': [karmaExecuteTests: true]]
        ]

        jsr.step.sapPiperStagePRVoting(
            script: nullScript,
            juStabUtils: utils,
        )

        assertThat(stepsCalled, hasItems('karmaExecuteTests'))
    }

    @Test
    void testPRVotingWithLinting() {

        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'maven'],
            stages: ['Pull-Request Voting': [npmExecuteLint: true, mavenExecuteStaticCodeChecks: true]]
        ]

        jsr.step.sapPiperStagePRVoting(
            script: nullScript,
            juStabUtils: utils,
        )

        assertThat(stepsCalled, hasItem('npmExecuteLint'))
        assertThat(stepsCalled, hasItem('mavenExecuteStaticCodeChecks'))

    }

    @Test
    void testPRVotingWithFortify() {

        nullScript.globalPipelineEnvironment.setFlag('piper-lib-os')
        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'maven'],
            runStep: ['Pull-Request Voting': [executeFortifyScan: true]]
        ]

        jsr.step.sapPiperStagePRVoting(
            script: nullScript,
            juStabUtils: utils,
        )

        assertThat(stepsCalled, hasItems('executeFortifyScan'))
        assertThat(stepParameters.executeFortifyScan.pullRequestName, is('PR-1'))
    }

    @Test
    void testPRVotingWithWhiteSource() {

        nullScript.globalPipelineEnvironment.setFlag('piper-lib-os')
        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'maven'],
            runStep: ['Pull-Request Voting': [whitesourceExecuteScan: true]]
        ]

        jsr.step.sapPiperStagePRVoting(
            script: nullScript,
            juStabUtils: utils,
        )

        assertThat(stepsCalled, hasItem('whitesourceExecuteScan'))
        assertThat(stepsCalled, not(hasItem('executePPMSComplianceCheck')))
        assertThat(stepParameters.whitesourceExecuteScan.productVersion, is('PR-1'))
    }

    @Test
    void testPRVotingWithPPMS() {

        nullScript.globalPipelineEnvironment.setFlag('piper-lib-os')
        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'maven'],
            runStep: ['Pull-Request Voting': [executePPMSComplianceCheck: true]]
        ]

        jsr.step.sapPiperStagePRVoting(
            script: nullScript,
            juStabUtils: utils,
        )

        assertThat(stepsCalled, hasItems('executePPMSComplianceCheck', 'whitesourceExecuteScan'))
        assertThat(stepParameters.executePPMSComplianceCheck.whitesourceProjectNames, hasItem('ws project - PR1'))
        assertThat(stepParameters.executePPMSComplianceCheck.pullRequestMode, is(true))
    }

    @Test
    void testNoDockerCommandAvailable() {

        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR dockerCommand')
        jsr.step.sapPiperStagePRVoting(script: nullScript, juStabUtils: utils)

    }

    @Test
    void testNoDockerImageAvailable() {

        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR dockerImage')
        jsr.step.sapPiperStagePRVoting(script: nullScript, juStabUtils: utils, dockerCommand: 'testCommand')

    }

    @Test
    void testDockerCommandCustomized() {
        nullScript.globalPipelineEnvironment.configuration = [general: [buildTool: 'docker']]
        nullScript.env = binding.variables.env
        jsr.step.sapPiperStagePRVoting(script: nullScript, juStabUtils: utils, dockerImage: 'testImage', dockerCommand: 'testCommand ${env.STAGE_NAME}')
        assertThat(jscr.shell, hasItem(containsString("testCommand ${binding.variables.env.STAGE_NAME}")))
    }

    @Test
    void testDockerBuild() {

        def image = ''
        def args = ''
        binding.setVariable('docker', [build: {i, a ->
            image = i
            args = a.toString()
        }])

        nullScript.globalPipelineEnvironment.configuration = [general: [buildTool: 'docker']]
        jsr.step.sapPiperStagePRVoting(script: nullScript, juStabUtils: utils, dockerImageName: 'testImage')

        assertThat(image, is('testImage'))
        assertThat(args, is(' .'))
    }

    @Test
    void testDockerBuildNoImageName() {

        nullScript.globalPipelineEnvironment.configuration = [general: [buildTool: 'docker']]
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR dockerImageName')
        jsr.step.sapPiperStagePRVoting(script: nullScript, juStabUtils: utils)

    }

    @Test
    void testDockerBuildWithScript() {

        nullScript.globalPipelineEnvironment.configuration = [general: [buildTool: 'docker']]
        jsr.step.sapPiperStagePRVoting(script: nullScript, juStabUtils: utils, dockerCommand: 'testCommand')

        assertThat(jscr.shell, hasItem(containsString('testCommand')))

    }

    @Test
    void testDockerBuildK8S() {

        binding.variables.env.ON_K8S = 'true'

        boolean kanikoCalled = false
        helper.registerAllowedMethod('kanikoExecute', [Map.class], {m ->
            kanikoCalled = true
        })

        nullScript.globalPipelineEnvironment.configuration = [general: [buildTool: 'docker']]
        jsr.step.sapPiperStagePRVoting(script: nullScript, juStabUtils: utils, dockerImageName: 'testImage')

        assertThat(kanikoCalled, is(true))
    }

    @Test
    void testDockerBuildK8Sish() {
        boolean kanikoCalled = false
        helper.registerAllowedMethod('kanikoExecute', [Map.class], {m ->
            kanikoCalled = true
        })

        nullScript.globalPipelineEnvironment.configuration = [general: [buildTool: 'docker'], stages: ['Pull-Request Voting': [kanikoExecute: true]]]
        jsr.step.sapPiperStagePRVoting(script: nullScript, juStabUtils: utils, dockerImageName: 'testImage')

        assertThat(kanikoCalled, is(true))
    }

    @Test
    void testDockerCommandWrapWithHttpsCredentialsUserToken() {
        nullScript.globalPipelineEnvironment.configuration = [general: [buildTool: 'npm']]
        jsr.step.sapPiperStagePRVoting(script: nullScript, juStabUtils: utils, gitHttpsCredentialsId: 'testCredentials', dockerCommand: 'testCommand ${env.GITHUB_USERNAME} ${env.GITHUB_TOKEN}' )

        assertThat(credentials[0].credentialsId, hasItem('testCredentials'))
        assertThat(jscr.shell, hasItem(containsString("testCommand ${binding.variables.env.GITHUB_USERNAME} ${binding.variables.env.GITHUB_TOKEN}")))
    }

    @Test
    void testDockerCommandWrapByWithoutHttpsCredentials() {
        nullScript.globalPipelineEnvironment.configuration = [general: [buildTool: 'npm']]
        jsr.step.sapPiperStagePRVoting(script: nullScript, juStabUtils: utils, dockerCommand: "testCommand" )

        assertThat(credentials[0], not(hasItem('credentialsId')))
        assertThat(binding.variables.env, not(hasItems('GITHUB_USERNAME','GITHUB_TOKEN')))
        assertThat(jscr.shell, hasItem(containsString('testCommand')))
    }

    void prepareObjectInterceptors(object) {
        object.metaClass.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.static.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.methodMissing = helper.getMethodMissingInterceptor()
    }

    @Test
    void testPRVotingWithNpmExecuteScripts() {

        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'npm'],
            stages: ['Pull-Request Voting': [npmExecuteScripts: true]]
        ]

        jsr.step.sapPiperStagePRVoting(
            script: nullScript,
            juStabUtils: utils,
        )

        assertThat(stepsCalled, hasItems('npmExecuteScripts', 'testsPublishResults'))
    }

    @Test
    void testPRVotingWithMavenExecuteIntegration() {

        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'maven'],
            stages: ['Pull-Request Voting': [mavenExecuteIntegration: true]]
        ]

        jsr.step.sapPiperStagePRVoting(
            script: nullScript,
            juStabUtils: utils,
        )

        assertThat(stepsCalled, hasItems('mavenExecuteIntegration', 'writeTemporaryCredentials'))
    }

}

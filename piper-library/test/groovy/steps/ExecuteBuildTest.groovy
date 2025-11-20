#!groovy
package steps

import com.sap.icd.jenkins.Utils

import static org.hamcrest.Matchers.contains
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.hasItemInArray
import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.hasKey
import static org.hamcrest.Matchers.hasSize
import static org.hamcrest.Matchers.is

import org.junit.Before
import org.junit.Test
import org.junit.Rule
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import static org.hamcrest.Matchers.startsWith
import static org.hamcrest.Matchers.stringContainsInOrder
import static org.junit.Assert.assertThat

import util.BasePiperTest
import util.Rules
import util.JenkinsStepRule
import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.JenkinsEnvironmentRule
import util.JenkinsExecuteDockerRule
import util.JenkinsWriteJsonRule

import org.yaml.snakeyaml.Yaml

import util.MockHelper
import util.Rules
import util.SharedLibraryCreator

import static com.lesfurets.jenkins.unit.MethodCall.callArgsToString
import static com.lesfurets.jenkins.unit.MethodSignature.method
import static org.junit.Assert.assertTrue

class ExecuteBuildTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)
    private JenkinsExecuteDockerRule jedr = new JenkinsExecuteDockerRule(this)
    private JenkinsWriteJsonRule jwjr = new JenkinsWriteJsonRule(this)
    private ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jedr)
        .around(jscr)
        .around(jlr)
        .around(jer)
        .around(jsr) // needs to be activated after jedr, otherwise executeDocker is not mocked
        .around(jwjr)
        .around(thrown)

    def xmakeOptions = [:]
    def sshagentOptions

    @Before
    void init() throws Exception {
        // set jenkins mock commands for Utils.groovy
        binding.setVariable('steps', [
            stash    : {m -> println "stashName = ${m.name}"},
            unstash: {println "unstashName called"}
        ])
        nullScript.globalPipelineEnvironment.configuration = nullScript.globalPipelineEnvironment.configuration ?: [:]
        nullScript.globalPipelineEnvironment.configuration['steps'] = nullScript.globalPipelineEnvironment.configuration['steps'] ?: [:]
        nullScript.globalPipelineEnvironment.configuration['steps']['executeBuild'] = nullScript.globalPipelineEnvironment.configuration['steps']['executeBuild'] ?: [:]

        helper.registerAllowedMethod("writeJSON", [Map], null)
        helper.registerAllowedMethod("archiveArtifacts", [Map], null)
        helper.registerAllowedMethod('fileExists', [String], {String s -> return (s == 'docker.metadata.json')})
        // register Jenkins commands with mock values
        //helper.registerAllowedMethod( 'sh', [Map], { map -> return 'value was sent to sh' } )
        helper.registerAllowedMethod('readJSON', [Map], { map ->
            if (map.file == 'docker.metadata.json')
                return [tag_name: 'docker.wdf.sap.corp:51116/com.sap.piper/k8s-test-app:1.0.0-20180329121243_788c252e1db53f91efd3eb1eb7c73990e0c8bf37', image_name: 'docker.wdf.sap.corp:51116/com.sap.piper/k8s-test-app:1.0.0-20180329121243_788c252e1db53f91efd3eb1eb7c73990e0c8bf37']
            else
                return mockHelper.loadJSON('resources/configMapping.json')
        })
        helper.registerAllowedMethod("readYaml", [Map], { Map m -> return new Yaml().load(m.text)})
        helper.registerAllowedMethod("sshagent", [Map, Closure], { Map m, Closure c->
            sshagentOptions = m
            c()
            return null
        })
    }

    // Test executeXMakeJob

    @Test
    void testXMakeNoBuildResults() {
        thrown.expectMessage(containsString("The xMake build did not return a 'build-results.json', please check your xMake configuration!"))
        jsr.step.executeBuild.downloadXMakeBuildResult(null)
    }

    @Test
    void testXMakeEmptyBuildResults() {
        thrown.expectMessage(containsString("The url to the remote job's deploy package is not available in 'build-results.json'"))
        jsr.step.executeBuild.downloadXMakeBuildResult([:])
    }

    @Test
    void testKanikoDefault() throws Exception {

        binding.variables.env.WORKSPACE = '/my/test/workspace'
        jsr.step.executeBuild(
            script: nullScript,
            buildType: 'kaniko',
        )

        // asserts

        assertThat(jedr.dockerParams.dockerImage, startsWith('gcr.io/kaniko-project/executor:debug'))

        assertThat(jedr.dockerParams.containerCommand, is('/busybox/tail -f /dev/null'))
        assertThat(jedr.dockerParams.containerShell, is('/busybox/sh'))
        assertThat(jedr.dockerParams.dockerOptions, is('-u 0 --entrypoint=\'\''))
        assertThat(jscr.shell, hasItem(stringContainsInOrder([
            '#!/busybox/sh',
            'mv /kaniko/.docker/config.json /kaniko/.docker/config.json.bak',
            'mv /kaniko/.config/gcloud/docker_credential_gcr_config.json /kaniko/.config/gcloud/docker_credential_gcr_config.json.bak',
            '/kaniko/executor --dockerfile /my/test/workspace/Dockerfile',
            '--context /my/test/workspace',
            '--no-push --skip-tls-verify-pull'
            ])))

    }

    // Test getIdentifierFromSSHLink
    @Test
    void testGetIdentifierFromSSHLink() throws Exception {
        def result = jsr.step.executeBuild.getIdentifierFromSSHLink('git@github.wdf.sap.corp:ContinuousDelivery/piper-library.git')
        // asserts
        assertThat(result, hasItem('ContinuousDelivery'))
        assertThat(result, hasItem('piper-library'))
        assertThat(result, hasSize(2))
        assertThat(result[0], is('ContinuousDelivery'))
        assertThat(result[1], is('piper-library'))
    }

    @Test
    void testNativeMavenBuild() throws Exception {
        nullScript.globalPipelineEnvironment.setArtifactVersion('1.0.0-20170217')
        nullScript.globalPipelineEnvironment.setGitCommitId('ffa402b965d5537d593e653d661a077f3aafe9d3')

        List stagingCalls = []
        List mavenBuildCalls = []
        List sapCumulusUploadCalls = []
        List sapGenerateEnvironmentInfo = []

        helper.registerAllowedMethod("sapCallStagingService", [Map], { Map m ->
            stagingCalls.add(m)
        })

        helper.registerAllowedMethod("mavenBuild", [Map], { Map m ->
            mavenBuildCalls.add(m)
        })

        helper.registerAllowedMethod("sapCumulusUpload", [Map], { Map m ->
            sapCumulusUploadCalls.add(m)
        })

        helper.registerAllowedMethod("sapGenerateEnvironmentInfo", [Map], { Map m ->
            sapGenerateEnvironmentInfo.add(m)
        })

        jsr.step.executeBuild(
            juStabUtils: utils,
            script: nullScript,
            nativeBuild: true,
            buildTool: 'maven',
            buildType: 'stage'
        )

        // asserts
        assertThat(stagingCalls, hasSize(3))
        assertThat(mavenBuildCalls, hasSize(1))
        assertThat(sapCumulusUploadCalls, hasSize(4))
        assertThat(sapGenerateEnvironmentInfo, hasSize(1))
    }

    @Test
    void testNativeMavenCnbBuild() throws Exception {
        nullScript.globalPipelineEnvironment.setArtifactVersion('1.0.0-20170217')
        nullScript.globalPipelineEnvironment.setGitCommitId('ffa402b965d5537d593e653d661a077f3aafe9d3')

        List stagingCalls = []
        List mavenBuildCalls = []
        List sapCumulusUploadCalls = []
        List cnbBuildCalls = []
        List sapGenerateEnvironmentInfo = []

        helper.registerAllowedMethod("sapCallStagingService", [Map], { Map m ->
            stagingCalls.add(m)
        })

        helper.registerAllowedMethod("mavenBuild", [Map], { Map m ->
            mavenBuildCalls.add(m)
        })

        helper.registerAllowedMethod("cnbBuild", [Map], { Map m ->
            cnbBuildCalls.add(m)
        })

        helper.registerAllowedMethod("sapCumulusUpload", [Map], { Map m ->
            sapCumulusUploadCalls.add(m)
        })

         helper.registerAllowedMethod("sapGenerateEnvironmentInfo", [Map], { Map m ->
            sapGenerateEnvironmentInfo.add(m)
        })

        jsr.step.executeBuild(
            juStabUtils: utils,
            script: nullScript,
            nativeBuild: true,
            buildTool: 'maven',
            buildType: 'stage',
            cnbBuild: true
        )

        // asserts
        assertThat(stagingCalls, hasSize(3))
        assertThat(stagingCalls[1], hasKey('action'))
        assertThat(stagingCalls[1]['action'], is('createRepositories'))
        assertThat(mavenBuildCalls, hasSize(1))
        assertThat(cnbBuildCalls, hasSize(1))
        assertThat(sapCumulusUploadCalls, hasSize(4))
        assertThat(sapGenerateEnvironmentInfo, hasSize(1))
    }

    @Test
    void testNativeNodeJSCnbBuild() throws Exception {
        nullScript.globalPipelineEnvironment.setArtifactVersion('1.0.0-20170217')
        nullScript.globalPipelineEnvironment.setGitCommitId('ffa402b965d5537d593e653d661a077f3aafe9d3')

        List stagingCalls = []
        List npmExecuteScriptsCalls = []
        List sapCumulusUploadCalls = []
        List cnbBuildCalls = []
        List sapGenerateEnvironmentInfo = []

        helper.registerAllowedMethod("sapCallStagingService", [Map], { Map m ->
            stagingCalls.add(m)
        })

        helper.registerAllowedMethod("npmExecuteScripts", [Map], { Map m ->
            npmExecuteScriptsCalls.add(m)
        })

        helper.registerAllowedMethod("cnbBuild", [Map], { Map m ->
            cnbBuildCalls.add(m)
        })

        helper.registerAllowedMethod("sapCumulusUpload", [Map], { Map m ->
            sapCumulusUploadCalls.add(m)
        })

         helper.registerAllowedMethod("sapGenerateEnvironmentInfo", [Map], { Map m ->
            sapGenerateEnvironmentInfo.add(m)
        })

        jsr.step.executeBuild(
            juStabUtils: utils,
            script: nullScript,
            nativeBuild: true,
            buildTool: 'npm',
            buildType: 'stage',
            cnbBuild: true
        )

        // asserts
        assertThat(stagingCalls, hasSize(3))
        assertThat(stagingCalls[1], hasKey('action'))
        assertThat(stagingCalls[1]['action'], is('createRepositories'))
        assertThat(npmExecuteScriptsCalls, hasSize(1))
        assertThat(cnbBuildCalls, hasSize(1))
        assertThat(sapCumulusUploadCalls, hasSize(4))
        assertThat(sapGenerateEnvironmentInfo, hasSize(1))
    }

    @Test
    void testNativeNpmBuild() throws Exception {
        List stagingCalls = []
        List npmExecuteScriptsCalls = []
        List sapCumulusUploadCalls = []
        List sapGenerateEnvironmentInfo = []

        helper.registerAllowedMethod("sapCallStagingService", [Map], { Map m ->
            stagingCalls.add(m)
        })

        helper.registerAllowedMethod("npmExecuteScripts", [Map], { Map m ->
            npmExecuteScriptsCalls.add(m)
        })

        helper.registerAllowedMethod("sapCumulusUpload", [Map], { Map m ->
            sapCumulusUploadCalls.add(m)
        })

        helper.registerAllowedMethod("sapGenerateEnvironmentInfo", [Map], { Map m ->
            sapGenerateEnvironmentInfo.add(m)
        })

        jsr.step.executeBuild(
            juStabUtils: utils,
            script: nullScript,
            nativeBuild: true,
            buildTool: 'npm',
            buildType: 'stage'
        )

        // asserts
        assertThat(stagingCalls, hasSize(3))
        assertThat(npmExecuteScriptsCalls, hasSize(1))
        assertThat(sapCumulusUploadCalls, hasSize(4))
        assertThat(sapGenerateEnvironmentInfo, hasSize(1))
    }

    @Test
    void testNativeMtaBuild() throws Exception {
        List stagingCalls = []
        List mtaBuild = []
        List sapCumulusUploadCalls = []
        List sapGenerateEnvironmentInfo = []

        helper.registerAllowedMethod("sapCallStagingService", [Map], { Map m ->
            stagingCalls.add(m)
        })

        helper.registerAllowedMethod("mtaBuild", [Map], { Map m ->
            mtaBuild.add(m)
        })

        helper.registerAllowedMethod("sapCumulusUpload", [Map], { Map m ->
            sapCumulusUploadCalls.add(m)
        })

        helper.registerAllowedMethod("sapGenerateEnvironmentInfo", [Map], { Map m ->
            sapGenerateEnvironmentInfo.add(m)
        })

        jsr.step.executeBuild(
            juStabUtils: utils,
            script: nullScript,
            nativeBuild: true,
            buildTool: 'mta',
            buildType: 'stage'
        )

        // asserts
        assertThat(stagingCalls, hasSize(3))
        assertThat(mtaBuild, hasSize(1))
        assertThat(sapCumulusUploadCalls, hasSize(4))
        assertThat(sapGenerateEnvironmentInfo, hasSize(1))
    }

    @Test
    void testNativeMtaBuildCnbBuild() throws Exception {
        List stagingCalls = []
        List mtaBuild = []
        List sapCumulusUploadCalls = []
        List sapGenerateEnvironmentInfo = []
        List cnbBuildCalls = []

        helper.registerAllowedMethod("sapCallStagingService", [Map], { Map m ->
            stagingCalls.add(m)
        })

        helper.registerAllowedMethod("mtaBuild", [Map], { Map m ->
            mtaBuild.add(m)
        })

        helper.registerAllowedMethod("sapCumulusUpload", [Map], { Map m ->
            sapCumulusUploadCalls.add(m)
        })

        helper.registerAllowedMethod("sapGenerateEnvironmentInfo", [Map], { Map m ->
            sapGenerateEnvironmentInfo.add(m)
        })

        helper.registerAllowedMethod("cnbBuild", [Map], { Map m ->
            cnbBuildCalls.add(m)
        })

        jsr.step.executeBuild(
            juStabUtils: utils,
            script: nullScript,
            nativeBuild: true,
            cnbBuild: true,
            buildTool: 'mta',
            buildType: 'stage'
        )

        // asserts
        assertThat(stagingCalls, hasSize(3))
        assertThat(mtaBuild, hasSize(1))
        assertThat(cnbBuildCalls, hasSize(1))
        assertThat(sapCumulusUploadCalls, hasSize(4))
        assertThat(sapGenerateEnvironmentInfo, hasSize(1))
    }


    @Test
    void testNativeDockerBuildKaniko() throws Exception {
        List stagingCalls = []
        List kanikoExecute = []
        List sapCumulusUploadCalls = []
        List sapGenerateEnvironmentInfo = []

        helper.registerAllowedMethod("sapCallStagingService", [Map], { Map m ->
            stagingCalls.add(m)
        })

        helper.registerAllowedMethod("kanikoExecute", [Map], { Map m ->
            kanikoExecute.add(m)
        })

        helper.registerAllowedMethod("sapCumulusUpload", [Map], { Map m ->
            sapCumulusUploadCalls.add(m)
        })

        helper.registerAllowedMethod("sapGenerateEnvironmentInfo", [Map], { Map m ->
            sapGenerateEnvironmentInfo.add(m)
        })

        jsr.step.executeBuild(
            juStabUtils: utils,
            script: nullScript,
            nativeBuild: true,
            buildTool: 'docker',
            buildType: 'stage'
        )

        // asserts
        assertThat(stagingCalls, hasSize(3))
        assertThat(kanikoExecute, hasSize(1))
        assertThat(sapCumulusUploadCalls, hasSize(4))
        assertThat(sapGenerateEnvironmentInfo, hasSize(1))
    }

    @Test
    void testNativeDockerBuildCnbBuild() throws Exception {
        nullScript.globalPipelineEnvironment.configuration['steps']['executeBuild']['cnbBuild'] = true
        List stagingCalls = []
        List cnbBuild = []
        List sapCumulusUploadCalls = []
        List sapGenerateEnvironmentInfo = []

        helper.registerAllowedMethod("sapCallStagingService", [Map], { Map m ->
            stagingCalls.add(m)
        })

        helper.registerAllowedMethod("cnbBuild", [Map], { Map m ->
            cnbBuild.add(m)
        })

        helper.registerAllowedMethod("sapCumulusUpload", [Map], { Map m ->
            sapCumulusUploadCalls.add(m)
        })

        helper.registerAllowedMethod("sapGenerateEnvironmentInfo", [Map], { Map m ->
            sapGenerateEnvironmentInfo.add(m)
        })

        jsr.step.executeBuild(
            juStabUtils: utils,
            script: nullScript,
            nativeBuild: true,
            buildTool: 'docker',
            buildType: 'stage'
        )

        // asserts
        assertThat(stagingCalls, hasSize(3))
        assertThat(cnbBuild, hasSize(1))
        assertThat(sapCumulusUploadCalls, hasSize(4))
        assertThat(sapGenerateEnvironmentInfo, hasSize(1))
    }

    @Test
    void testNativeGolangBuild() throws Exception {
        List stagingCalls = []
        List golangBuild = []
        List sapCumulusUploadCalls = []
        List sapGenerateEnvironmentInfo = []

        helper.registerAllowedMethod("sapCallStagingService", [Map], { Map m ->
            stagingCalls.add(m)
        })

        helper.registerAllowedMethod("golangBuild", [Map], { Map m ->
            golangBuild.add(m)
        })

        helper.registerAllowedMethod("sapCumulusUpload", [Map], { Map m ->
            sapCumulusUploadCalls.add(m)
        })

        helper.registerAllowedMethod("sapGenerateEnvironmentInfo", [Map], { Map m ->
            sapGenerateEnvironmentInfo.add(m)
        })

        jsr.step.executeBuild(
            juStabUtils: utils,
            script: nullScript,
            nativeBuild: true,
            buildTool: 'golang',
            buildType: 'stage'
        )

        // asserts
        assertThat(stagingCalls, hasSize(3))
        assertThat(golangBuild, hasSize(1))
        assertThat(sapCumulusUploadCalls, hasSize(4))
        assertThat(sapGenerateEnvironmentInfo, hasSize(1))
    }
}

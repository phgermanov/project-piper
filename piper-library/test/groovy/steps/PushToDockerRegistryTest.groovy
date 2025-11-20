#!groovy
package steps

import org.junit.Before
import org.junit.Test
import org.junit.Rule
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsExecuteDockerRule
import util.JenkinsShellCallRule

import static org.hamcrest.CoreMatchers.hasItem
import static org.hamcrest.CoreMatchers.is
import static org.hamcrest.CoreMatchers.not

import static org.junit.Assert.assertThat

import util.Rules
import util.JenkinsStepRule
import util.JenkinsEnvironmentRule

class PushToDockerRegistryTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)
    private JenkinsExecuteDockerRule jedr = new JenkinsExecuteDockerRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jer)
        .around(jscr)
        .around(jedr)
        .around(jsr)

    def dockerMockArgs = [:]
    class DockerMock {
        DockerMock(name){
            dockerMockArgs.name = name
        }
        def withRegistry(paramRegistry, paramCredentials, paramClosure){
            dockerMockArgs.paramRegistry = paramRegistry
            dockerMockArgs.paramCredentials = paramCredentials
            return paramClosure()
        }
        def withRegistry(paramRegistry, paramClosure){
            dockerMockArgs.paramRegistryAnonymous = paramRegistry.toString()
            return paramClosure()
        }

        def image(name) {
            dockerMockArgs.name = name
            return new DockerImageMock()
        }
    }

    def dockerMockPushes = []
    def dockerMockPull = false
    class DockerImageMock {
        DockerImageMock(){}
        def push(tag){
            dockerMockPushes.add(tag)
        }
        def push(){
            push('default')
        }

        def pull(){
            dockerMockPull = true
        }
    }

    @Before
    void init() throws Exception {
        binding.setVariable('docker', new DockerMock('test'))
    }

    @Test
    void testPushToDockerRegistryLatest() throws Exception {
        def dockerBuildImage = new DockerImageMock()
        jsr.step.pushToDockerRegistry(
            script: nullScript,
            dockerRegistryUrl: 'https://testRegistry',
            dockerCredentialsId: 'testCredentialsId',
            dockerBuildImage: dockerBuildImage,
            tagLatest: true
        )

        assertThat(dockerMockArgs.paramRegistry, is('https://testRegistry'))
        assertThat(dockerMockArgs.paramCredentials, is('testCredentialsId'))
        assertThat(dockerMockArgs.paramRegistryAnonymous, is(null))
        assertThat(dockerMockArgs.name, is('test'))
        assertThat(dockerMockPushes, hasItem('default'))
        assertThat(dockerMockPushes, hasItem('latest'))
        assertJobStatusSuccess()
    }

    @Test
    void testPushToDockerRegistryArtifactVersion() throws Exception {
        nullScript.globalPipelineEnvironment.setArtifactVersion('1.0.0')
        def dockerBuildImage = new DockerImageMock()
        jsr.step.pushToDockerRegistry(
            script: nullScript,
            dockerRegistryUrl: 'https://testRegistry',
            dockerCredentialsId: 'testCredentialsId',
            dockerBuildImage: dockerBuildImage,
            tagArtifactVersion: true,
            tagLatest: false
        )

        assertThat(dockerMockArgs.paramRegistry, is('https://testRegistry'))
        assertThat(dockerMockArgs.paramCredentials, is('testCredentialsId'))
        assertThat(dockerMockArgs.paramRegistryAnonymous, is(null))
        assertThat(dockerMockArgs.name, is('test'))
        assertThat(dockerMockPushes, hasItem('default'))
        assertThat(dockerMockPushes, not(hasItem('latest')))
        assertThat(dockerMockPushes, hasItem('1.0.0'))
        assertJobStatusSuccess()
    }

    @Test
    void testPushToDockerRegistryLatestAndArtifactVersion() throws Exception {
        nullScript.globalPipelineEnvironment.setArtifactVersion('1.0.0')
        def dockerBuildImage = new DockerImageMock()
        jsr.step.pushToDockerRegistry(
            script: nullScript,
            dockerRegistryUrl: 'https://testRegistry',
            dockerCredentialsId: 'testCredentialsId',
            dockerBuildImage: dockerBuildImage,
            tagLatest: true,
            tagArtifactVersion: true
        )

        assertThat(dockerMockArgs.paramRegistry, is('https://testRegistry'))
        assertThat(dockerMockArgs.paramCredentials, is('testCredentialsId'))
        assertThat(dockerMockArgs.paramRegistryAnonymous, is(null))
        assertThat(dockerMockArgs.name, is('test'))
        assertThat(dockerMockPushes, hasItem('default'))
        assertThat(dockerMockPushes, hasItem('latest'))
        assertThat(dockerMockPushes, hasItem('1.0.0'))
        assertJobStatusSuccess()
    }

    @Test
    void testPushToDockerRegistryWithDefaultImage() throws Exception {
        nullScript.globalPipelineEnvironment.setDockerBuildImage(new DockerImageMock())
        jsr.step.pushToDockerRegistry(
            script: nullScript,
            dockerRegistryUrl: 'https://testRegistry',
            dockerCredentialsId: 'testCredentialsId'
        )

        assertThat(dockerMockPushes, hasItem('default'))
        assertJobStatusSuccess()
    }

    @Test
    void testPushToDockerRegistryImageNameNoLatest() throws Exception {
        jsr.step.pushToDockerRegistry(
            script: nullScript,
            dockerRegistryUrl: 'https://testRegistry',
            dockerCredentialsId: 'testCredentialsId',
            dockerImage: 'testImage:tag',
        )

        assertThat(dockerMockArgs.paramRegistry, is('https://testRegistry'))
        assertThat(dockerMockArgs.paramCredentials, is('testCredentialsId'))
        assertThat(dockerMockArgs.name, is('testImage:tag'))
        assertThat(dockerMockPushes, hasItem('default'))
        assertThat(dockerMockPushes, not(hasItem('latest')))
        assertJobStatusSuccess()
    }

    @Test
    void testWithDockerMetadata() {
        nullScript.globalPipelineEnvironment.setDockerMetadata([
            repo: 'testRegistry:55555',
            tag_name: 'testRegistry:55555/path/testImage:tag',
            image_name: 'testRegistry:55555/path/testImage:tag'
        ])
        jsr.step.pushToDockerRegistry(
            script: nullScript,
            dockerRegistryUrl: 'https://testRegistry',
            dockerCredentialsId: 'testCredentialsId',
        )

        assertThat(dockerMockArgs.paramRegistryAnonymous, is('https://testRegistry:55555'))
        assertThat(dockerMockArgs.name, is('path/testImage:tag'))
        assertThat(jscr.shell, hasItem('docker tag testRegistry:55555/path/testImage:tag path/testImage:tag'))
        assertThat(dockerMockPull, is(true))

    }

    @Test
    void testWithAppContainerDockerMetadata() {
        nullScript.globalPipelineEnvironment.setDockerMetadata([
            repo: 'testRegistry:55555',
            tag_name: 'testRegistry:55555/path/testImage:tag',
            image_name: 'testRegistry:55555/path/testImage:tag'
        ])
        nullScript.globalPipelineEnvironment.setAppContainerDockerMetadata([
            repo: 'testRegistryX:55555',
            tag_name: 'testRegistryX:55555/path/testImageX:tag',
            image_name: 'testRegistryX:55555/path/testImageX:tag'
        ])
        jsr.step.pushToDockerRegistry(
            script: nullScript,
            dockerRegistryUrl: 'https://testRegistry',
            dockerCredentialsId: 'testCredentialsId',
        )

        assertThat(dockerMockArgs.paramRegistryAnonymous, is('https://testRegistryX:55555'))
        assertThat(dockerMockArgs.name, is('path/testImageX:tag'))
        assertThat(jscr.shell, hasItem('docker tag testRegistryX:55555/path/testImageX:tag path/testImageX:tag'))
        assertThat(dockerMockPull, is(true))
    }

    @Test
    void testPushToDockerRegistryWithSourceImageAndRegistry() {
        jsr.step.pushToDockerRegistry(
            script: nullScript,
            dockerCredentialsId: 'testCredentialsId',
            dockerRegistryUrl: 'https://testRegistry',
            sourceImage: 'testSourceName:testSourceTag',
            sourceRegistryUrl: 'http://testSourceRegistry'
        )

        assertThat(dockerMockArgs.paramRegistryAnonymous, is('http://testSourceRegistry'))
        assertThat(dockerMockArgs.name, is('testSourceName:testSourceTag'))
        assertThat(jscr.shell, hasItem('docker tag testSourceRegistry/testSourceName:testSourceTag testSourceName:testSourceTag'))
        assertThat(dockerMockPull, is(true))
        assertJobStatusSuccess()
    }

    @Test
    void testPushToDockerRegistryWithTargetImage() {
        jsr.step.pushToDockerRegistry(
            script: nullScript,
            dockerCredentialsId: 'testCredentialsId',
            dockerImage: 'testImage:tag',
            dockerRegistryUrl: 'https://testRegistry',
            sourceImage: 'testSourceName:testSourceTag',
            sourceRegistryUrl: 'http://testSourceRegistry'
        )

        assertThat(dockerMockArgs.paramRegistryAnonymous, is('http://testSourceRegistry'))
        assertThat(dockerMockArgs.name, is('testSourceName:testSourceTag'))
        assertThat(jscr.shell, hasItem('docker tag testSourceRegistry/testSourceName:testSourceTag testImage:tag'))
        assertThat(dockerMockPull, is(true))
        assertJobStatusSuccess()
    }

    @Test
    void testPushToDockerRegistryMoveOnKubernetes() {
        helper.registerAllowedMethod('usernamePassword', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            if (l[0].credentialsId == 'testCredentialsId') {
                binding.setProperty('userid', 'registryUser')
                binding.setProperty('password', '********')
            }
            try {
                c()
            } finally {
                binding.setProperty('userid', null)
                binding.setProperty('password', null)
            }
        })

        binding.setVariable('docker', null)
        jscr.setReturnStatus('docker ps -q > /dev/null', 1)

        jsr.step.pushToDockerRegistry(
            script: nullScript,
            dockerCredentialsId: 'testCredentialsId',
            dockerImage: 'testImage:tag',
            dockerRegistryUrl: 'https://my.registry:55555',
            sourceImage: 'sourceImage:sourceTag',
            sourceRegistryUrl: 'https://my.source.registry:44444'
        )

        assertThat(jscr.shell, hasItem('skopeo copy --src-tls-verify=false --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/testImage:tag'))

    }

    @Test
    void testPushToDockerRegistryMoveOnKubernetesTagLatest() {
        helper.registerAllowedMethod('usernamePassword', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            if (l[0].credentialsId == 'testCredentialsId') {
                binding.setProperty('userid', 'registryUser')
                binding.setProperty('password', '********')
            }
            try {
                c()
            } finally {
                binding.setProperty('userid', null)
                binding.setProperty('password', null)
            }
        })

        binding.setVariable('docker', null)
        jscr.setReturnStatus('docker ps -q > /dev/null', 1)

        jsr.step.pushToDockerRegistry(
            script: nullScript,
            dockerCredentialsId: 'testCredentialsId',
            dockerImage: 'some.path/testImage:tag',
            dockerRegistryUrl: 'https://my.registry:55555',
            sourceImage: 'some.path/sourceImage:sourceTag',
            sourceRegistryUrl: 'https://my.source.registry:44444',
            tagLatest: true
        )

        assertThat(jscr.shell, hasItem('skopeo copy --src-tls-verify=false --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/some.path/sourceImage:sourceTag docker://my.registry:55555/some.path/testImage:tag'))
        assertThat(jscr.shell, hasItem('skopeo copy --src-tls-verify=false --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/some.path/sourceImage:sourceTag docker://my.registry:55555/some.path/testImage:latest'))

    }

    @Test
    void testPushToDockerRegistryMoveOnKubernetesTagArtifactVersion() {
        nullScript.globalPipelineEnvironment.setArtifactVersion('1.0.0')
        helper.registerAllowedMethod('usernamePassword', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            if (l[0].credentialsId == 'testCredentialsId') {
                binding.setProperty('userid', 'registryUser')
                binding.setProperty('password', '********')
            }
            try {
                c()
            } finally {
                binding.setProperty('userid', null)
                binding.setProperty('password', null)
            }
        })

        binding.setVariable('docker', null)
        jscr.setReturnStatus('docker ps -q > /dev/null', 1)

        jsr.step.pushToDockerRegistry(
            script: nullScript,
            dockerCredentialsId: 'testCredentialsId',
            dockerImage: 'some.path/testImage:tag',
            dockerRegistryUrl: 'https://my.registry:55555',
            sourceImage: 'some.path/sourceImage:sourceTag',
            sourceRegistryUrl: 'https://my.source.registry:44444',
            tagLatest: false,
            tagArtifactVersion: true
        )

        assertThat(jscr.shell, hasItem('skopeo copy --src-tls-verify=false --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/some.path/sourceImage:sourceTag docker://my.registry:55555/some.path/testImage:tag'))
        assertThat(jscr.shell, hasItem('skopeo copy --src-tls-verify=false --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/some.path/sourceImage:sourceTag docker://my.registry:55555/some.path/testImage:1.0.0'))

    }

    @Test
    void testPushToDockerRegistryMoveOnKubernetesTagLatestAndArtifactVersion() {
        nullScript.globalPipelineEnvironment.setArtifactVersion('1.0.0')
        helper.registerAllowedMethod('usernamePassword', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            if (l[0].credentialsId == 'testCredentialsId') {
                binding.setProperty('userid', 'registryUser')
                binding.setProperty('password', '********')
            }
            try {
                c()
            } finally {
                binding.setProperty('userid', null)
                binding.setProperty('password', null)
            }
        })

        binding.setVariable('docker', null)
        jscr.setReturnStatus('docker ps -q > /dev/null', 1)

        jsr.step.pushToDockerRegistry(
            script: nullScript,
            dockerCredentialsId: 'testCredentialsId',
            dockerImage: 'some.path/testImage:tag',
            dockerRegistryUrl: 'https://my.registry:55555',
            sourceImage: 'some.path/sourceImage:sourceTag',
            sourceRegistryUrl: 'https://my.source.registry:44444',
            tagLatest: true,
            tagArtifactVersion: true
        )

        assertThat(jscr.shell, hasItem('skopeo copy --src-tls-verify=false --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/some.path/sourceImage:sourceTag docker://my.registry:55555/some.path/testImage:tag'))
        assertThat(jscr.shell, hasItem('skopeo copy --src-tls-verify=false --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/some.path/sourceImage:sourceTag docker://my.registry:55555/some.path/testImage:latest'))
        assertThat(jscr.shell, hasItem('skopeo copy --src-tls-verify=false --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/some.path/sourceImage:sourceTag docker://my.registry:55555/some.path/testImage:1.0.0'))

    }

    @Test
    void testPushToDockerRegistryMoveOnKubernetesSourceOnly() {
        helper.registerAllowedMethod('usernamePassword', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            if (l[0].credentialsId == 'testCredentialsId') {
                binding.setProperty('userid', 'registryUser')
                binding.setProperty('password', '********')
            }
            try {
                c()
            } finally {
                binding.setProperty('userid', null)
                binding.setProperty('password', null)
            }
        })

        binding.setVariable('docker', null)
        jscr.setReturnStatus('docker ps -q > /dev/null', 1)

        jsr.step.pushToDockerRegistry(
            script: nullScript,
            dockerCredentialsId: 'testCredentialsId',
            dockerRegistryUrl: 'https://my.registry:55555',
            sourceImage: 'sourceImage:sourceTag',
            sourceRegistryUrl: 'https://my.source.registry:44444'
        )

        assertThat(jscr.shell, hasItem('skopeo copy --src-tls-verify=false --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/sourceImage:sourceTag'))

    }

    @Test
    void testPushToDockerRegistryMoveOnKubernetesSourceRegistryFromEnv() {

        helper.registerAllowedMethod('usernamePassword', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            if (l[0].credentialsId == 'testCredentialsId') {
                binding.setProperty('userid', 'registryUser')
                binding.setProperty('password', '********')
            }
            try {
                c()
            } finally {
                binding.setProperty('userid', null)
                binding.setProperty('password', null)
            }
        })

        binding.setVariable('docker', null)
        jscr.setReturnStatus('docker ps -q > /dev/null', 1)

        nullScript.globalPipelineEnvironment.setDockerMetadata([
            repo: 'my.source.registry:44444',
            tag_name: 'my.source.registry:44444/sourceImage:sourceTag',
            image_name: 'my.source.registry:44444/sourceImage:sourceTag'
        ])

        jsr.step.pushToDockerRegistry(
            script: nullScript,
            dockerCredentialsId: 'testCredentialsId',
            dockerRegistryUrl: 'https://my.registry:55555',
            sourceImage: 'sourceImage:sourceTag',
        )

        assertThat(jscr.shell, hasItem('skopeo copy --src-tls-verify=false --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/sourceImage:sourceTag'))
    }
}

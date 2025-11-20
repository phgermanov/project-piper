package com.sap.piper.internal

import hudson.AbortException
import org.junit.Before
import org.junit.Ignore
import org.junit.rules.ExpectedException
import util.JenkinsShellCallRule

import static org.hamcrest.CoreMatchers.hasItem
import static org.hamcrest.CoreMatchers.is
import static org.hamcrest.CoreMatchers.nullValue
import static org.junit.Assert.assertThat

import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.Rules

class DockerUtilsTest extends BasePiperTest {

    public ExpectedException exception = ExpectedException.none()
    public JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    def dockerMockArgs = [:]
    class DockerMock {
        def withRegistry(paramRegistry, paramClosure){
            dockerMockArgs.paramRegistryAnonymous = paramRegistry.toString()
            return paramClosure()
        }
    }

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(jscr)
        .around(exception)

    @Before
    void init() {
        nullScript.binding.setVariable('docker', new DockerMock())
    }

    @Test
    void testWithDockerDeamon() {
        DockerUtils dockerUtils = new DockerUtils(nullScript)
        assertThat(dockerUtils.withDockerDeamon(), is(true))
    }

    @Test
    void testWithoutDockerDeamon() {
        jscr.setReturnStatus('docker ps -q > /dev/null', 1)
        DockerUtils dockerUtils = new DockerUtils(nullScript)
        assertThat(dockerUtils.withDockerDeamon(), is(false))
    }

    @Test
    void testOnKubernetesJaaS() {
        nullScript.env.jaas_owner = 'Test'
        nullScript.env.ON_K8S = null
        DockerUtils dockerUtils = new DockerUtils(nullScript)
        assertThat(dockerUtils.onKubernetes(), is(true))
    }

    @Test
    void testOnKubernetesOS() {
        nullScript.env.ON_K8S = 'true'
        nullScript.env.jaas_owner = null
        DockerUtils dockerUtils = new DockerUtils(nullScript)
        assertThat(dockerUtils.onKubernetes(), is(true))
    }

    @Test
    void testSaveImageDocker() {
        DockerUtils dockerUtils = new DockerUtils(nullScript)
        dockerUtils.saveImage('testPath', 'testImage:tag', 'https://my.registry:55555')
        assertThat(dockerMockArgs.paramRegistryAnonymous, is('https://my.registry:55555'))
        assertThat(jscr.shell, hasItem('docker pull my.registry:55555/testImage:tag && docker save --output testPath my.registry:55555/testImage:tag'))
    }

    @Test
    void testSaveImageDockerNoRegistry() {
        DockerUtils dockerUtils = new DockerUtils(nullScript)
        dockerUtils.saveImage('testPath', 'testImage:tag')
        assertThat(dockerMockArgs.paramRegistryAnonymous, nullValue())
    }

    @Test
    void testSaveImageKubernetes() {
        jscr.setReturnStatus('docker ps -q > /dev/null', 1)
        nullScript.env.ON_K8S = 'true'
        DockerUtils dockerUtils = new DockerUtils(nullScript)
        dockerUtils.saveImage('testPath', 'testImage:tag', 'https://my.registry:55555')
        assertThat(jscr.shell, hasItem('skopeo copy --src-tls-verify=false docker://my.registry:55555/testImage:tag docker-archive:testPath:testImage:tag'))
    }

    @Test
    void testSaveImageNoContainer() {
        jscr.setReturnStatus('docker ps -q > /dev/null', 1)
        helper.registerAllowedMethod('sh', [String.class], {s ->
            if (s == 'skopeo copy --src-tls-verify=false docker://my.registry:55555/testImage:tag docker-archive:testPath:testImage:tag')
            throw new AbortException('Error')
        })
        DockerUtils dockerUtils = new DockerUtils(nullScript)
        exception.expectMessage('No Kubernetes container provided for running Skopeo ...')
        dockerUtils.saveImage('testPath', 'testImage:tag', 'https://my.registry:55555')
    }

    @Ignore
    @Test
    void testMoveImageDocker() {
        //not relevant yet since not implemented
        helper.registerAllowedMethod('usernamePassword', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            if (l[0].credentialsId == 'registryCredentials') {
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

        DockerUtils dockerUtils = new DockerUtils(nullScript)
        dockerUtils.moveImage(
            [
                registryUrl: 'https://my.source.registry:44444',
                image: 'sourceImage:sourceTag'
            ],
            [
                registryUrl: 'https://my.registry:55555',
                image: 'testImage:tag',
                credentialsId: 'registryCredentials'
            ]
        )
    }

    @Test
    void testMoveImageKubernetes() {
        helper.registerAllowedMethod('usernamePassword', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            if (l[0].credentialsId == 'testCredentialsId') {
                nullScript.binding.setProperty('userid', 'registryUser')
                nullScript.binding.setProperty('password', '********')
            }
            try {
                c()
            } finally {
                nullScript.binding.setProperty('userid', null)
                nullScript.binding.setProperty('password', null)
            }
        })

        jscr.setReturnStatus('docker ps -q > /dev/null', 1)
        DockerUtils dockerUtils = new DockerUtils(nullScript)
        dockerUtils.moveImage(
            [
                registryUrl: 'https://my.source.registry:44444',
                image: 'sourceImage:sourceTag'
            ],
            [
                registryUrl: 'https://my.registry:55555',
                image: 'testImage:tag',
                credentialsId: 'testCredentialsId'
            ]
        )

        assertThat(jscr.shell, hasItem('skopeo copy --src-tls-verify=false --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/testImage:tag'))
    }

    @Test
    void testGetRegistryFromUrl() {
        DockerUtils dockerUtils = new DockerUtils(nullScript)
        assertThat(dockerUtils.getRegistryFromUrl('https://my.registry.com:55555'), is('my.registry.com:55555'))
        assertThat(dockerUtils.getRegistryFromUrl('http://my.registry.com:55555'), is('my.registry.com:55555'))
    }

    @Test
    void testGetProtocolFromUrl() {
        DockerUtils dockerUtils = new DockerUtils(nullScript)
        assertThat(dockerUtils.getProtocolFromUrl('https://my.registry.com:55555'), is('https'))
        assertThat(dockerUtils.getProtocolFromUrl('http://my.registry.com:55555'), is('http'))
    }

    @Test
    void testGetNameFromImageUrl() {

        def testList = [
            ["image", "image"],
            ["image:tag", "image"],
            ["path/image", "path/image"],
            ["path/image:tag", "path/image"],
            ["my.registry/path/image", "path/image"],
            ["my.registry/path/image:tag", "path/image"],
            ["my.registry:50000/path/image", "path/image"],
            ["my.registry:50000/path/image:tag", "path/image"],
            ["my.registry:50000/path/image@sha256:44092b2ea3da5b9adc3c51c2ff6b399ae487094183a3746dbb8918d450d52ac5", "path/image"],
            ["my.registry:50000/path/image:tag@sha256:44092b2ea3da5b9adc3c51c2ff6b399ae487094183a3746dbb8918d450d52ac5", "path/image"]
        ]

        DockerUtils dockerUtils = new DockerUtils(nullScript)

        testList.each {test ->
            assertThat(dockerUtils.getNameFromImageUrl(test[0]), is(test[1]))
        }
    }
}

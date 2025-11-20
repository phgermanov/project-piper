#!groovy
package steps

import com.sap.piper.internal.JenkinsUtils

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasKey
import static org.hamcrest.Matchers.is

import org.junit.Before
import org.junit.Ignore
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import static org.junit.Assert.assertThat
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.Rules
import util.JenkinsLoggingRule
import util.JenkinsStepRule
import util.JenkinsEnvironmentRule

@Ignore("step disabled")
class WriteInfluxTest extends BasePiperTest {
    Map fileMap
    Map stepMap

    public JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)
    private ExpectedException thrown = ExpectedException.none()

    String influxVersion

    class JenkinsUtilsMock extends JenkinsUtils {
        def getPlugin(name){
            return [getVersion:{influxVersion}]
        }
    }

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jsr)
        .around(jer)
        .around(thrown)

    @Before
    void init() throws Exception {
        //reset stepMap
        stepMap = [:]
        //reset fileMap
        fileMap = [:]

        influxVersion = '1.15'

        helper.registerAllowedMethod('writeFile', [Map.class],{m -> fileMap[m.file] = m.text})
        helper.registerAllowedMethod('step', [Map.class],{m -> stepMap = m})
        helper.registerAllowedMethod('influxDbPublisher', [Map.class],{m -> stepMap = m})

        jer.env.setArtifactVersion('1.2.3')
        jer.env.setGithubOrg('PiperTestOrg')
        jer.env.setGithubRepo('pipeline-test')
        nullScript.globalPipelineEnvironment = jer.env
    }

    @Test
    void testInfluxInactive() {
        jsr.step.writeInflux(
            juStabUtils: utils,
            script: nullScript,
        )
        // asserts
        assertThat(jlr.log, containsString('Artifact version: 1.2.3'))
        assertThat(stepMap.size(), is(0))
        assertThat(fileMap, hasKey('jenkins_data.json'))
        assertThat(fileMap, hasKey('pipeline_data.json'))
        assertJobStatusSuccess()
    }

    @Test
    void testWriteInfluxWithDefaults() throws Exception {
        jer.env.configuration['steps'] = jer.env.configuration['steps'] ?: [:]
        jer.env.configuration['steps']['writeInflux'] = jer.env.configuration['steps']['writeInflux'] ?: [:]
        jer.env.configuration['steps']['writeInflux']['influxServer'] = 'testInflux'
        jsr.step.writeInflux(
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            script: nullScript,
        )
        // asserts
        assertThat(jlr.log, containsString('Artifact version: 1.2.3'))
        assertThat(stepMap.selectedTarget, is('testInflux'))
        assertThat(stepMap.customPrefix, is("${'PiperTestOrg_pipeline-test'}"))
        assertThat(stepMap.customData, is(null))
        assertThat(stepMap.customDataMap, is([pipeline_data:[:], step_data: [:]]))
        assertThat(fileMap, hasKey('jenkins_data.json'))
        assertThat(fileMap, hasKey('pipeline_data.json'))
        assertJobStatusSuccess()
    }

    @Test
    void testWriteInfluxWithCustomValues() throws Exception {
        jsr.step.writeInflux(
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            script: nullScript,
            influxServer: 'myInstance',
            influxPrefix: 'myPrefix'
        )
        // asserts
        assertThat(stepMap.selectedTarget, is('myInstance'))
        assertThat(stepMap.customPrefix, is('myPrefix'))
        assertJobStatusSuccess()
    }

    @Test
    void testWriteInfluxWithoutArtifactVersion() throws Exception {
        jer.env.setArtifactVersion(null)
        jsr.step.writeInflux(
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            script: nullScript
        )
        // asserts
        assertThat(stepMap.size(), is(0))
        assertThat(fileMap.size(), is(0))
        assertThat(jlr.log, containsString('no artifact version available -> exiting writeInflux without writing data'))
        assertJobStatusSuccess()
    }

    @Test
    void testCustomValues() {
        nullScript.globalPipelineEnvironment.setInfluxCustomDataProperty('build_result', 'SUCCESS')
        nullScript.globalPipelineEnvironment.setInfluxCustomDataProperty('build_result_key', 1)
        jsr.step.writeInflux(
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            script: nullScript,
            influxServer: 'myInstance',
            customData: [:],
            customDataTags: [:],
            customDataMap: [deployment_data: [key1: 'keyValue1']],
            customDataMapTags: [deployment_data: [tag1: 'tagValue1']]
        )
        assertThat(fileMap.size(), is(0))
        assertThat(stepMap.customData, is(null))
        assertThat(stepMap.customDataTags, is(null))
        assertThat(stepMap.customDataMap, hasKey('deployment_data'))
        assertThat(stepMap.customDataMapTags, hasKey('deployment_data'))
    }

    @Test
    @Ignore
    void testNPE() {
        thrown.expect(hudson.AbortException)
        //thrown.expectMessage('Error: zippedFile or jdroidAPiUrl is not available')
        //nullScript.globalPipelineEnvironment.setInfluxCustomDataProperty('build_result', 'SUCCESS')
        //nullScript.globalPipelineEnvironment.setInfluxCustomDataProperty('build_result_key', 1)
        helper.registerAllowedMethod('step', [Map.class],{m ->
            throw new NullPointerException()
        })
        // execute tests
        try{
            jsr.step.writeInflux(
                juStabUtils: utils,
                jenkinsUtilsStub: new JenkinsUtilsMock(),
                script: nullScript,
                influxServer: 'testInflux')
        }finally{
            assertThat(jlr.log, containsString('NullPointerException occured, is the correct target defined?'))
            assertThat(jlr.log, containsString('java.lang.NullPointerException'))
        }
    }

    @Test
    void testInfluxWriteDataPluginVersion2() {
        influxVersion = '2.0'

        nullScript.globalPipelineEnvironment.setInfluxCustomDataProperty('build_result', 'SUCCESS')
        nullScript.globalPipelineEnvironment.setInfluxCustomDataProperty('build_result_key', 1)
        jsr.step.writeInflux(
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            script: nullScript,
            influxServer: 'myInstance',
            customData: [:],
            customDataTags: [:],
            customDataMap: [deployment_data: [key1: 'keyValue1']],
            customDataMapTags: [deployment_data: [tag1: 'tagValue1']]
        )
        assertThat(fileMap.size(), is(0))
        assertThat(stepMap.customData, is(null))
        assertThat(stepMap.customDataTags, is(null))
        assertThat(stepMap.customDataMap, hasKey('deployment_data'))
        assertThat(stepMap.customDataMapTags, hasKey('deployment_data'))
    }
}

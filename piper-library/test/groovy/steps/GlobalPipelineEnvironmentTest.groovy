#!groovy
package steps

import hudson.AbortException
import hudson.model.Action
import org.jenkinsci.plugins.workflow.actions.LabelAction
import org.jenkinsci.plugins.workflow.actions.TagsAction
import org.jenkinsci.plugins.workflow.graph.FlowNode
import org.jenkinsci.plugins.workflow.graph.ForkNode
import org.junit.Before
import org.junit.Ignore
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsReadJsonRule
import util.JenkinsStepRule
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class GlobalPipelineEnvironmentTest extends BasePiperTest {

    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsReadJsonRule jrjr = new JenkinsReadJsonRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jsr) // needs to be activated after jedr, otherwise executeDocker is not mocked
        .around(jrjr)

    @Test
    void testInfluxHandling() {
        nullScript.globalPipelineEnvironment.setInfluxCustomDataProperty('testKey1', 'testValue1')
        nullScript.globalPipelineEnvironment.setInfluxCustomDataProperty('testKey2', 'testValue2')
        assertThat(nullScript.globalPipelineEnvironment.getInfluxCustomData(), is([testKey1: 'testValue1', testKey2: 'testValue2']))

        nullScript.globalPipelineEnvironment.setInfluxCustomDataTagsProperty('testTag3', 'testValue3')
        assertThat(nullScript.globalPipelineEnvironment.getInfluxCustomDataTags(), is([testTag3: 'testValue3']))

        nullScript.globalPipelineEnvironment.setInfluxCustomDataMapProperty('pipeline_data', 'testKey1', 'testValue1')
        nullScript.globalPipelineEnvironment.setInfluxCustomDataMapProperty('step_data', 'testKey2', 'testValue2')
        assertThat(nullScript.globalPipelineEnvironment.getInfluxCustomDataMap(), is([pipeline_data: [testKey1: 'testValue1'], step_data: [testKey2: 'testValue2']]))

        nullScript.globalPipelineEnvironment.setInfluxCustomDataMapTagsProperty('pipeline_data', 'testTag1', 'testValue1')
        nullScript.globalPipelineEnvironment.setInfluxCustomDataMapTagsProperty('step_data', 'testTag2', 'testValue2')
        assertThat(nullScript.globalPipelineEnvironment.getInfluxCustomDataMapTags(), is([pipeline_data: [testTag1: 'testValue1'], step_data: [testTag2: 'testValue2']]))

        nullScript.globalPipelineEnvironment.setInfluxCustomDataMapTagsProperty('testMeasurement', 'testTag4', 'testValue4')
        assertThat(nullScript.globalPipelineEnvironment.getInfluxCustomDataMapTags().testMeasurement, is([testTag4: 'testValue4']))

        nullScript.globalPipelineEnvironment.setInfluxCustomDataProperty('testProperty', 'testValue')
        assertThat(nullScript.globalPipelineEnvironment.getInfluxCustomData().get('testProperty'), is('testValue'))

        nullScript.globalPipelineEnvironment.setPipelineMeasurement('testMeasurement', 8)
        assertThat(nullScript.globalPipelineEnvironment.getInfluxCustomDataMap().get('pipeline_data').get('testMeasurement'), is(8))

        nullScript.globalPipelineEnvironment.setInfluxStepData('testKey', 'testValue')
        assert nullScript.globalPipelineEnvironment.getInfluxCustomDataMap().get('step_data').get('testKey') == 'testValue'
    }

    @Test
    void testHasLabelAction() {
        def flowNode = new ForkNode(null, null, []) {
            @Override
            protected String getTypeDisplayName() {
                return "My Test Flow"
            }
        }
        flowNode.actions = [new TagsAction()]
        assertThat(nullScript.globalPipelineEnvironment.hasLabelAction(flowNode), is(false))
        flowNode.actions += [new LabelAction()]
        assertThat(nullScript.globalPipelineEnvironment.hasLabelAction(flowNode), is(true))
    }


    @Test
    void testXMakeProperties() throws Exception {

        def jsonProperties = mockHelper.loadJSON('test/resources/build_results.json')

        nullScript.globalPipelineEnvironment.setXMakeProperties(jsonProperties)
        assert nullScript.globalPipelineEnvironment.getXMakeProperties().staging_repo_id == 'xmakedeploymilestonesprofile-30449'
        assert nullScript.globalPipelineEnvironment.getXMakeProperty('staging_repo_id') == 'xmakedeploymilestonesprofile-30449'
        assert nullScript.globalPipelineEnvironment.getXMakeProperties().projectArchive == 'http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.snapshots.blob/com/sap/production-services/build-service/IndustryCloudFoundation/pipeline-test-node/bc7d12de355c96d68fb5429023da2e68acecfd0e/2017_03_08__08_03_08/deployPackage.tar.gz'
        assert nullScript.globalPipelineEnvironment.getXMakeProperty('projectArchive') == 'http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.snapshots.blob/com/sap/production-services/build-service/IndustryCloudFoundation/pipeline-test-node/bc7d12de355c96d68fb5429023da2e68acecfd0e/2017_03_08__08_03_08/deployPackage.tar.gz'
    }

    @Test
    void testGetStepConfiguration() {

        nullScript.loadDefaultValues()
        nullScript.globalPipelineEnvironment.configuration = [
            general: [test1: 'general'],
            stages: [Acceptance: [], Integration:[test1: 'integration']],
            steps: [step1: [test1: 'step'], step2: [test2: 'step2']]]

        //check with defaults
        assertThat(nullScript.globalPipelineEnvironment.getStepConfiguration('step1', 'Acceptance').get('pythonVersion'), is('python3'))

        //check without defaults
        assertThat(nullScript.globalPipelineEnvironment.getStepConfiguration('step1', 'Acceptance', false).get('pythonVersion'), isEmptyOrNullString())

        //check correct hierarchy
        assertThat(nullScript.globalPipelineEnvironment.getStepConfiguration('step1', 'Integration').get('test1'), is('integration'))
        assertThat(nullScript.globalPipelineEnvironment.getStepConfiguration('step1', 'Acceptance').get('test1'), is('step'))
        assertThat(nullScript.globalPipelineEnvironment.getStepConfiguration('step2', 'Acceptance').get('test1'), is('general'))
    }

    @Test
    void testCompatibilityWithCommonPipelineEnvironment() {

        nullScript.commonPipelineEnvironment = this.loadScript('test/resources/openSource/commonPipelineEnvironment.groovy').commonPipelineEnvironment

        //Bring the script/step object into globalPipelineEnvironment in order to be able to access commonPipelineEnvironment
        nullScript.globalPipelineEnvironment.cpe = nullScript.commonPipelineEnvironment

        nullScript.globalPipelineEnvironment.setAppContainerProperty('appProperty', 'testVal')

        nullScript.globalPipelineEnvironment.setArtifactVersion('1.2.3')
        nullScript.globalPipelineEnvironment.setGitCommitId('myCommitId')
        nullScript.globalPipelineEnvironment.setBuildResult('TEST')

        nullScript.globalPipelineEnvironment.setInfluxCustomDataProperty('key1', 'val1')
        nullScript.globalPipelineEnvironment.setInfluxCustomDataProperty('key2', 'val2')
        nullScript.globalPipelineEnvironment.setInfluxCustomDataMapProperty('pipeline_data', 'p1', 'vp1')
        nullScript.globalPipelineEnvironment.setInfluxCustomDataMapProperty('step_data', 's1', 'vs1')
        nullScript.globalPipelineEnvironment.setInfluxCustomDataProperty('key3', 'val3')
        nullScript.globalPipelineEnvironment.setInfluxStepData('step1', true)
        nullScript.globalPipelineEnvironment.setPipelineMeasurement ('measure1', 2)

        nullScript.globalPipelineEnvironment.setGithubOrg ('TestOrg')
        nullScript.globalPipelineEnvironment.setGithubRepo ('TestRepo')
        nullScript.globalPipelineEnvironment.setGitBranch ('testBranch')
        nullScript.globalPipelineEnvironment.setGitSshUrl ('git@test:test')

        assertThat(nullScript.commonPipelineEnvironment.getAppContainerProperty('appProperty'), is('testVal'))

        assertThat(nullScript.commonPipelineEnvironment.getArtifactVersion(), is('1.2.3'))
        assertThat(nullScript.commonPipelineEnvironment.getGitCommitId(), is('myCommitId'))
        assertThat(nullScript.commonPipelineEnvironment.getBuildResult(), is('TEST'))

        //TODO: activate once TODO in setInfluxCustomDataProperty is done (https://github.com/SAP/jenkins-library/pull/420)
        //assertThat(nullScript.commonPipelineEnvironment.getInfluxCustomData(), is([key1: 'val1', key2: 'val2', key3: 'val3']))
        //TODO: activate once TODO in setInfluxCustomDataMapProperty is done (https://github.com/SAP/jenkins-library/pull/420)
        //assertThat(nullScript.commonPipelineEnvironment.getInfluxCustomDataMap().pipeline_data.p1, is('vp1'))
        //assertThat(nullScript.commonPipelineEnvironment.getInfluxCustomDataMap().pipeline_data.measure1, is(2))
        //assertThat(nullScript.commonPipelineEnvironment.getInfluxCustomDataMap().step_data.step1, is(true))

        assertThat(nullScript.commonPipelineEnvironment.getGithubOrg(), is('TestOrg'))
        assertThat(nullScript.commonPipelineEnvironment.getGithubRepo(), is('TestRepo'))
        assertThat(nullScript.commonPipelineEnvironment.getGitBranch(), is('testBranch'))
        assertThat(nullScript.commonPipelineEnvironment.getGitSshUrl(), is('git@test:test'))

    }

    @Test
    void testSetGithubStatisticsAuthenticated() {
        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            if (!m.authentication) {
                throw new AbortException('authentication denied')
            }
            return [content: '{"stats": {"total":3,"additions":2,"deletions":1},"files":["file1"]}']
        })

        nullScript.globalPipelineEnvironment.setGithubStatistics(nullScript, 'https://github.api.url', 'githubOrg', 'githubRepo', 'gitCommit', 'credentialId')
        assertThat(nullScript.globalPipelineEnvironment.influxCustomDataMap.pipeline_data.github_changes, is(3))
        assertThat(nullScript.globalPipelineEnvironment.influxCustomDataMap.pipeline_data.github_additions, is(2))
        assertThat(nullScript.globalPipelineEnvironment.influxCustomDataMap.pipeline_data.github_deletions, is(1))
        assertThat(nullScript.globalPipelineEnvironment.influxCustomDataMap.pipeline_data.github_filesChanged, is(1))
    }

    @Test
    void testSetGithubStatisticsNotAuthenticated() {
        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            if (!m.authentication) {
                throw new AbortException('authentication denied')
            }
            return [content: '{"stats": {"total":3,"additions":2,"deletions":1},"files":["file1"]}']
        })

        nullScript.globalPipelineEnvironment.setGithubStatistics(nullScript, 'https://github.api.url', 'githubOrg', 'githubRepo', 'gitCommit', null)
        assertThat(jlr.log, containsString('failed to connect to GitHub API'))
    }
}

#!groovy
package steps

import org.codehaus.groovy.GroovyException
import util.JenkinsReadYamlRule
import util.JenkinsWriteFileRule

import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.contains
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.hasSize
import static org.hamcrest.Matchers.hasKey
import static org.hamcrest.Matchers.hasValue
import static org.hamcrest.Matchers.is
import static org.junit.Assert.assertThat

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsReadJsonRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.Rules

class SetupPipelineEnvironmentTest extends BasePiperTest {

    private JenkinsReadJsonRule jrjr = new JenkinsReadJsonRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsReadYamlRule jryr = new JenkinsReadYamlRule(this, 'test/resources/setupPipelineEnvironment/')
    private JenkinsWriteFileRule jwfr = new JenkinsWriteFileRule(this)

    def yamlFiles = [:]
    def discarderSettings = [:]
    def commonSettings = [:]
    def removeScheduleCalled = false
    def oldSchedule = ""
    def scheduleJobCalled = false

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jrjr)
        .around(jryr)
        .around(jscr)
        .around(jsr)
        .around(jwfr)

    @Before
    void init()  {

        // load configuration file and mock the return value of readProperties method
        helper.registerAllowedMethod( 'readProperties', [Map.class], {map ->
            return mockHelper.loadProperties('test/resources/config.properties')
        })

        helper.registerAllowedMethod('fileExists', [String.class], {s ->
            if (s == '.pipeline/config.properties') {
                return true
            } else {
                return false
            }
        })

        helper.registerAllowedMethod('writeYaml', [Map.class], {m ->
            yamlFiles[m.file] = m.data
        })

        helper.registerAllowedMethod("addBuildDiscarder", [int, int, int, int], { a,b,c,d -> discarderSettings['daysToKeep'] = a; discarderSettings['numToKeep'] = b;discarderSettings['artifactDaysToKeep'] = c; discarderSettings['artifactNumToKeep'] = d})
        helper.registerAllowedMethod('getLibrariesInfo', [], {return [
            [name: 'piper-lib'],
            [name: 'piper-lib-os']
        ]})
        helper.registerAllowedMethod('setupCommonPipelineEnvironment', [Map.class], { m -> commonSettings = m })
        nullScript.commonPipelineEnvironment = this.loadScript('test/resources/openSource/commonPipelineEnvironment.groovy').commonPipelineEnvironment
        nullScript.commonPipelineEnvironment.configuration = [steps: [:]]

        helper.registerAllowedMethod("removeJobSchedule", [String.class], {
            schedule ->
                removeScheduleCalled = true
                oldSchedule = schedule.toString()
        })

        helper.registerAllowedMethod("removeJobSchedule", [], {
                removeScheduleCalled = true
        })

        helper.registerAllowedMethod("scheduleJob", [String.class], {
            schedule ->
                scheduleJobCalled = true
                if(!'H(0-59) H(18-23) * * *'.equals(schedule.toString()))
                    throw new GroovyException("scheduleJob called with wrong schedule")
        })

        helper.registerAllowedMethod("pipelineIsScheduled", [], {return false})

        helper.registerAllowedMethod("library", [String.class], {return null})

        helper.registerAllowedMethod("sapPipelineInit", [Map], { Map m ->
                        sapReportPipelineStatusCalled = true
        })
        // TODO check if sapPipelineInit has been executed with sapReportPipelineStatusCalled
    }

    @Test
    void testNightlyExecution() throws Exception {
        jsr.step.setupPipelineEnvironment([
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            script: nullScript,
            githubOrg: 'IndustryCloudFoundation',
            githubRepo: 'pipeline-test',
            gitBranch: 'master',
            runNightly: true
        ])

        assertThat(scheduleJobCalled, is(true))
    }

    @Test
    void testContinuousExecution() throws Exception {
        jsr.step.setupPipelineEnvironment([
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            script: nullScript,
            githubOrg: 'IndustryCloudFoundation',
            githubRepo: 'pipeline-test',
            gitBranch: 'master'
        ])

        assertThat(removeScheduleCalled, is(true))
        assertThat(scheduleJobCalled, is(false))
    }

    @Test
    void testPipelineEnvironmentInitialization() throws Exception {
        jsr.step.setupPipelineEnvironment([
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            script: nullScript,
            githubOrg: 'IndustryCloudFoundation',
            githubRepo: 'pipeline-test',
            gitBranch: 'master'
        ])

        assertThat(discarderSettings['daysToKeep'], is(-1))
        assertThat(discarderSettings['numToKeep'], is(10))
        assertThat(discarderSettings['artifactDaysToKeep'], is(-1))
        assertThat(discarderSettings['artifactNumToKeep'], is(-1))
    }

    @Test
    void testSpecificBuildDiscarderSettings() {
        def parameters = [juStabUtils: utils, jenkinsUtilsStub: jenkinsUtils, script: nullScript, configFile: 'resources/config.properties', githubOrg: 'IndustryCloudFoundation',
                          githubRepo: 'pipeline-test', gitBranch: 'master', buildDiscarder: [daysToKeep: 7, numToKeep: -1]]

        helper.registerAllowedMethod("addBuildDiscarder", [int, int, int, int], { a,b,c,d -> discarderSettings['daysToKeep'] = a; discarderSettings['numToKeep'] = b;discarderSettings['artifactDaysToKeep'] = c; discarderSettings['artifactNumToKeep'] = d})

        jsr.step.setupPipelineEnvironment(parameters)
        assertJobStatusSuccess()

        assertThat(discarderSettings['daysToKeep'], is(7))
        assertThat(discarderSettings['numToKeep'], is(-1))
        assertThat(discarderSettings['artifactDaysToKeep'], is(-1))
        assertThat(discarderSettings['artifactNumToKeep'], is(-1))
    }

    @Test
    void testPipelineEnvironmentConvertLists() {
        helper.registerAllowedMethod('fileExists', [String.class], {s ->
            if (s == '.pipeline/myConfig.yml')
                return true
            else
                return false
        })

        jsr.step.setupPipelineEnvironment([
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            script: nullScript,
            githubOrg: 'IndustryCloudFoundation',
            githubRepo: 'pipeline-test',
            gitBranch: 'master',
            customDefaults: 'test/resources/setupPipelineEnvironment/myDefaults.yaml'
        ])

        assertThat(nullScript?.globalPipelineEnvironment?.configuration?.steps?.executeWhitesourceScan?.buildDescriptorExcludeList, is(null))
        assertThat(nullScript?.globalPipelineEnvironment?.configuration?.steps?.executeFortifyScan?.buildDescriptorExcludeList, allOf(hasSize(1), hasItem('moduleExclude')))
    }

    @Test
    void testPiperOsStepConfigMapping() {
        nullScript.commonPipelineEnvironment.configuration = [steps: [newmanExecute: [newmanKey: 'testVal']]]

        helper.registerAllowedMethod('fileExists', [String.class], {s ->
            if (s == '.pipeline/config.yml')
                return true
            else
                return false
        })

        //config.configYmlFile = .pipeline/config.yml
        jsr.step.setupPipelineEnvironment([
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            script: nullScript,
            githubOrg: 'IndustryCloudFoundation',
            githubRepo: 'pipeline-test',
            gitBranch: 'master'
        ])

        assertThat(nullScript.commonPipelineEnvironment.configuration.steps.newmanExecute.newmanKey, is('testVal'))
        assertThat(nullScript.commonPipelineEnvironment.configuration.steps.gaugeExecuteTests.gaugeKey1, is('gaugeVal1'))
        assertThat(nullScript.commonPipelineEnvironment.configuration.steps.gaugeExecuteTests.gaugeKey2, is('gaugeVal2'))
    }

    @Test
    void testPiperOsStepConfigMappingNoOsConfig() {
//        nullScript.commonPipelineEnvironment.configuration = [steps: [newmanExecute: [newmanKey: 'testVal']]]

        helper.registerAllowedMethod('fileExists', [String.class], {s ->
            if (s == '.pipeline/config.yml')
                return true
            else
                return false
        })

        //config.configYmlFile = .pipeline/config.yml
        jsr.step.setupPipelineEnvironment([
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            script: nullScript,
            githubOrg: 'IndustryCloudFoundation',
            githubRepo: 'pipeline-test',
            gitBranch: 'master'
        ])

        //no assert required. Just need to make sure that no error occurs
    }


    @Test
    void testConfigYamlHandlling() {
        nullScript.commonPipelineEnvironment.configuration = [steps: [newmanExecute: [newmanKey: 'testVal']]]
        // mock existance of config.yaml
        helper.registerAllowedMethod('fileExists', [String.class], {s -> return s == '.pipeline/config.yaml' })

        //config.configYmlFile = .pipeline/config.yml
        jsr.step.setupPipelineEnvironment([
                juStabUtils: utils,
                jenkinsUtilsStub: jenkinsUtils,
                script: nullScript,
                githubOrg: 'IndustryCloudFoundation',
                githubRepo: 'pipeline-test',
                gitBranch: 'master'
        ])

        assertThat(commonSettings, allOf(
                hasKey('customDefaults'),
                hasValue(['piper-defaults.yml'])
        ))
    }

}

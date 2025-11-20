package com.sap.icd.jenkins

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.JenkinsLoggingRule
import util.JenkinsSetupRule
import util.SharedLibraryCreator

import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.contains
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.hasSize
import static org.hamcrest.Matchers.nullValue
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertThat

class UtilsTest extends BasePipelineTest {

    @Rule
    public ExpectedException exception = ExpectedException.none()
    public JenkinsSetupRule setUpRule = new JenkinsSetupRule(this, SharedLibraryCreator.lazyLoadedLibrary)
    public JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain =
        RuleChain.outerRule(setUpRule)
            .around(loggingRule)

    Utils utils

    @Before
    void init() throws Exception {
        utils = new Utils()
        prepareObjectInterceptors(utils)
    }

    @Test
    void testGetMandatoryParameterValid() {

        def sourceMap = [test1: 'value1', test2: 'value2']

        def defaultFallbackMap = [myDefault1: 'default1']

        assertEquals('value1', utils.getMandatoryParameter(sourceMap, 'test1', null))
        assertEquals('value1', utils.getMandatoryParameter(sourceMap, 'test1', defaultFallbackMap.test))

        assertEquals('value1', utils.getMandatoryParameter(sourceMap, 'test1', ''))

        assertEquals('value1', utils.getMandatoryParameter(sourceMap, 'test1', 'customValue'))

    }

    @Test
    void testGetMandatoryParameterDefaultFallback() {

        def myMap = [test1: 'value1', test2: 'value2']

        assertEquals('', utils.getMandatoryParameter(myMap, 'test3', ''))
        assertEquals('customValue', utils.getMandatoryParameter(myMap, 'test3', 'customValue'))
    }


    @Test
    void testGetMandatoryParameterFail() {

        def myMap = [test1: 'value1', test2: 'value2']

        exception.expect(Exception.class)

        exception.expectMessage("ERROR - NO VALUE AVAILABLE FOR")

        utils.getMandatoryParameter(myMap, 'test3', null)
    }

    @Test
    void testGetSSHGitFolder() {

        assertEquals(utils.getFolderFromGitUrl('git@github.wdf.sap.corp:test/test-pipeline.git'), 'test')
        assertEquals(utils.getFolderFromGitUrl('git@github.wdf.sap.corp:test1/test2/test-pipeline.git'), 'test1/test2')
        assertEquals(utils.getFolderFromGitUrl('git@github.wdf.sap.corp:test-pipeline.git'), '')

    }

    @Test
    void testGetSSHGitRepository() {
        assertEquals(utils.getRepositoryFromGitUrl('git@github.wdf.sap.corp:test/test-pipeline.git'), 'test-pipeline')
        assertEquals(utils.getRepositoryFromGitUrl('git@github.wdf.sap.corp:test1/test2/test-pipeline.git'), 'test-pipeline')
        assertEquals(utils.getRepositoryFromGitUrl('git@github.wdf.sap.corp:test-pipeline.git'), 'test-pipeline')
        assertEquals(utils.getRepositoryFromGitUrl('git@github.wdf.sap.corp:test-pipeline'), 'test-pipeline')
    }

    @Test
    void testGetHTTPSGitFolder() {

        assertEquals(utils.getFolderFromGitUrl('https://github.wdf.sap.corp/test-pipeline.git'), '')
        assertEquals(utils.getFolderFromGitUrl('https://github.wdf.sap.corp/test/test-pipeline.gitt'), 'test')
        assertEquals(utils.getFolderFromGitUrl('https://github.wdf.sap.corp/test1/test2/test-pipeline.git'), 'test1/test2')

    }

    @Test
    void testGetHTTPSGitRepository() {
        assertEquals(utils.getRepositoryFromGitUrl('https://github.wdf.sap.corp/test/test-pipeline.git'), 'test-pipeline')
        assertEquals(utils.getRepositoryFromGitUrl('https://github.wdf.sap.corp/test1/test2/test-pipeline.git'), 'test-pipeline')
        assertEquals(utils.getRepositoryFromGitUrl('https://github.wdf.sap.corp/test-pipeline.git'), 'test-pipeline')
        assertEquals(utils.getRepositoryFromGitUrl('https://github.wdf.sap.corp/test-pipeline'), 'test-pipeline')
    }

    @Test
    void testAddBuildDiscarder() {
        prepareObjectInterceptors(utils.getJenkinsUtilsInstance())
        def a, b, c, d
        helper.registerAllowedMethod("addBuildDiscarder", [int.class, int.class, int.class, int.class], {
            one, two, three, four -> a = one; b = two; c = three; d = four
        })

        utils.addBuildDiscarder(-1, 35)

        assertEquals(a, -1)
        assertEquals(b, 35)
        assertEquals(c, -1)
        assertEquals(d, -1)
    }

    @Test
    void testStash() {

        def localUtils = new Utils()

        prepareObjectInterceptors(localUtils)

        def mockedStepsProperty = new Object()

        prepareObjectInterceptors(mockedStepsProperty)

        localUtils.steps = mockedStepsProperty

        helper.registerAllowedMethod("stash", [Map.class], {
            stashInput ->
        })

        localUtils.stash('test1-1.log')

        assertThat(loggingRule.log, containsString("Stash content:"))

    }

    @Test
    void testStashWithOwnMessage() {

        def localUtils = new Utils()

        prepareObjectInterceptors(localUtils)

        def mockedStepsProperty = new Object()

        prepareObjectInterceptors(mockedStepsProperty)

        localUtils.steps = mockedStepsProperty

        helper.registerAllowedMethod("stash", [Map.class], {
            stashInput ->

                if (stashInput.name == 'test2') {
                    throw new Exception("Cannot stash file test2")
                }
        })

        localUtils.stashWithMessage('test1-1.log', 'My own message')

        localUtils.stashWithMessage('test2', 'My own message')

        assertThat(loggingRule.log, containsString("Stash content: test1-1.log"))
        assertThat(loggingRule.log, containsString("My own message"))
    }

    @Test
    void testUnstash() {

        def stashedInput = []

        def localUtils = new Utils()

        prepareObjectInterceptors(localUtils)

        def mockedStepsProperty = new Object()

        prepareObjectInterceptors(mockedStepsProperty)

        localUtils.steps = mockedStepsProperty

        helper.registerAllowedMethod("stash", [Map.class], {
            stashInput ->
                stashedInput.add(stashInput.name)
        })

        helper.registerAllowedMethod("unstash", [String.class], {
            stashFileName ->
                if (!stashedInput.contains(stashFileName)) {
                    throw new Exception("Stash not found")
                }
                return stashFileName
        })

        def resultUnstashedEmpty = localUtils.unstashAll(['log1', 'log2'])

        assertThat(resultUnstashedEmpty, hasSize(0))

        localUtils.stash('log1')
        localUtils.stash('log2')

        def resultUnstashed = localUtils.unstashAll(['log1', 'log2'])

        assertThat(resultUnstashed, hasSize(2))
    }

    @Test
    void testNPMRead() {

        helper.registerAllowedMethod("readJSON", [Map.class], {
            searchConfig ->
                def packageJsonFile = new File("test/resources/downloadFromNexusTest/npmScope/${searchConfig.file}")
                return utils.parseJsonSerializable(packageJsonFile.text)
        })

        def gav = utils.getNpmGAV('package.release.json')

        assertEquals(gav.group, '@sap')
        assertEquals(gav.artifact, 'hdi-deploy')
        assertEquals(gav.version, '2.3.0')
    }

    @Test
    void testDubRead() {

        helper.registerAllowedMethod("readJSON", [Map.class], {
            searchConfig ->
                def packageJsonFile = new File("test/resources/downloadFromNexusTest/dub/${searchConfig.file}")
                return utils.parseJsonSerializable(packageJsonFile.text)
        })

        def gav = utils.getDubGAV('dub.json')

        assertEquals(gav.group, 'com.sap.dlang')
        assertEquals(gav.artifact, 'hdi-deploy')
        assertEquals(gav.version, '2.3.0')
    }

    @Test
    void testGetPipGAV() {

        helper.registerAllowedMethod("sh", [Map.class], {
            map ->
                def descriptorFile = new File("test/resources/utilsTest/${map.script.substring(4, map.script.size())}")
                return descriptorFile.text
        })

        def gav = utils.getPipGAV('setup.py')

        assertEquals('', gav.group)
        assertEquals('py_connect', gav.artifact)
        assertEquals('1.0', gav.version)
    }

    @Test
    void testReadMavenGAVComplete() {
        def parameters = []

        helper.registerAllowedMethod("readMavenPom", [Map.class], {
            searchConfig ->
                return new Object(){
                    def groupId = 'test.group', artifactId = 'test-artifact', version = '1.2.4', packaging = 'jar'
                }
        })

        helper.registerAllowedMethod("sh", [Map.class], {
            mvnHelpCommand ->
                def scriptCommand = mvnHelpCommand['script']
                parameters.add(scriptCommand)
                if(scriptCommand.contains('project.groupId'))
                    return 'test.group'
                if(scriptCommand.contains('project.artifactId'))
                    return 'test-artifact'
                if(scriptCommand.contains('project.version'))
                    return '1.2.4'
                if(scriptCommand.contains('project.packaging'))
                    return 'jar'
                if(scriptCommand.contains('project.build.finalName'))
                    return 'test-artifact'
                if(scriptCommand.contains('project.build.directory'))
                    return 'target/'
                if(scriptCommand.contains('project.basedir'))
                    return '.'
        })

        def gav = utils.readMavenGAV('pom.xml', '')

        assertEquals(gav.group, 'test.group')
        assertEquals(gav.artifact, 'test-artifact')
        assertEquals(gav.version, '1.2.4')
        assertEquals(gav.packaging, 'jar')
    }

    @Test
    void testReadMavenGAVPartial() {
        def parameters = []

        helper.registerAllowedMethod("readMavenPom", [Map.class], {
            searchConfig ->
                return new Object(){
                    def groupId = null, artifactId = null, version = null, packaging = 'jar'
                }
        })

        helper.registerAllowedMethod("sh", [Map.class], {
            mvnHelpCommand ->
                def scriptCommand = mvnHelpCommand['script']
                parameters.add(scriptCommand)
                if(scriptCommand.contains('project.groupId'))
                   return 'test.group'
                if(scriptCommand.contains('project.artifactId'))
                    return 'test-artifact'
                if(scriptCommand.contains('project.version'))
                    return '1.2.4'
                if(scriptCommand.contains('project.packaging'))
                    return 'jar'
                if(scriptCommand.contains('project.build.finalName'))
                    return 'test-artifact'
                if(scriptCommand.contains('project.build.directory'))
                    return 'target/'
                if(scriptCommand.contains('project.basedir'))
                    return '.'
        })

        def gav = utils.readMavenGAV('pom.xml')

        assertEquals(gav.group, 'test.group')
        assertEquals(gav.artifact, 'test-artifact')
        assertEquals(gav.version, '1.2.4')
        assertEquals(gav.packaging, 'jar')
    }

    @Test
    void testReadMavenGAVWithVersionPlaceholder() {
        def parameters = []

        helper.registerAllowedMethod("readMavenPom", [Map.class], {
            searchConfig ->
                return new Object(){
                    def groupId = 'test.group', artifactId = 'test-artifact', version = '${revision}', packaging = 'jar'
                }
        })

        helper.registerAllowedMethod("sh", [Map.class], {
            mvnHelpCommand ->
                def scriptCommand = mvnHelpCommand['script']
                parameters.add(scriptCommand)
                if(scriptCommand.contains('project.groupId'))
                    return 'test.group'
                if(scriptCommand.contains('project.artifactId'))
                    return 'test-artifact'
                if(scriptCommand.contains('project.version'))
                    return '1.2.4'
                if(scriptCommand.contains('project.packaging'))
                    return 'jar'
                if(scriptCommand.contains('project.build.finalName'))
                    return 'test-artifact'
                if(scriptCommand.contains('project.build.directory'))
                    return 'target/'
                if(scriptCommand.contains('project.basedir'))
                    return '.'
        })

        def gav = utils.readMavenGAV('pom.xml')

        assertEquals('test.group', gav.group)
        assertEquals('test-artifact', gav.artifact)
        assertEquals('1.2.4', gav.version)
        assertEquals('jar', gav.packaging)
    }

    @Test
    void testGetVersionElements() {
        assertThat(utils.getVersionElements('xyz'), nullValue())
        assertThat(utils.getVersionElements('1'), allOf(hasEntry('all', '1'), hasEntry('full', '1'), hasEntry('major', '1')))
        assertThat(utils.getVersionElements('1.22'), allOf(hasEntry('all', '1.22'), hasEntry('full', '1.22'), hasEntry('major', '1'), hasEntry('minor', '22')))
        assertThat(utils.getVersionElements('1.22.333'), allOf(hasEntry('all', '1.22.333'), hasEntry('full', '1.22.333'), hasEntry('major', '1'), hasEntry('minor', '22'), hasEntry('patch', '333')))
        assertThat(utils.getVersionElements('1.22.333-20000101+sha1'), allOf(hasEntry('all', '1.22.333-20000101+sha1'), hasEntry('full', '1.22.333-20000101'), hasEntry('major', '1'), hasEntry('minor', '22'), hasEntry('patch', '333'), hasEntry('timestamp', '20000101')))
        assertThat(utils.getVersionElements('1.22.333.20000101+sha1'), allOf(hasEntry('all', '1.22.333.20000101+sha1'), hasEntry('full', '1.22.333.20000101'), hasEntry('major', '1'), hasEntry('minor', '22'), hasEntry('patch', '333'), hasEntry('timestamp', '20000101')))

    }

    void prepareObjectInterceptors(object) {
        object.metaClass.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.static.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.methodMissing = helper.getMethodMissingInterceptor()
    }
}

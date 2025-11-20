#!groovy
package steps

import org.junit.Before
import org.junit.Ignore

import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.junit.Assert.assertThat

import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest

import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.Rules

@Ignore("step disabled")
class ExecuteOpenSourceDependencyScanTest extends BasePiperTest {

    JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    private stepsCalled = []
    private stepParams = []
    private parallelMap = [:]
    private parallelKeys = []
    private boolean failFast

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jscr)
        .around(jsr)

    @Before
    void init() {
        helper.registerAllowedMethod('parallel', [Map.class], { m ->
            parallelMap = m
            parallelMap.each {key, value ->
                if(key != 'failFast'){
                    parallelKeys.add(key)
                    value()
                }else{
                    failFast = value
                }
            }
        })

        helper.registerAllowedMethod('protecodeExecuteScan', [Map.class], {m ->
            stepsCalled.add('protecodeExecuteScan')
            stepParams.add(m)
        })

        helper.registerAllowedMethod('executeWhitesourceScan', [Map.class], {m ->
            stepsCalled.add('executeWhitesourceScan')
            stepParams.add(m)
        })

        helper.registerAllowedMethod('executeVulasScan', [Map.class], {m ->
            stepsCalled.add('executeVulasScan')
            stepParams.add(m)
        })

        helper.registerAllowedMethod('whitesourceExecuteScan', [Map.class], {m ->
            stepsCalled.add('whitesourceExecuteScan')
            stepParams.add(m)
        })

        helper.registerAllowedMethod('detectExecuteScan', [Map.class], {m ->
            stepsCalled.add('detectExecuteScan')
            stepParams.add(m)
        })
    }



    @Test
    void testPlainProject() {

        jsr.step.executeOpenSourceDependencyScan([
            script: nullScript,
            juStabUtils: utils,
            executeVulasScan: true,
            executeWhitesourceScan: true,
        ])

        assertThat(parallelMap.size(), is(3))
        assertThat(stepsCalled, allOf(hasItem('executeVulasScan'), hasItem('executeWhitesourceScan')))
        assertThat(failFast, is(false))
    }

    @Test
    void testPlainProjectWithVulasAndBlackDuck() {

        jsr.step.executeOpenSourceDependencyScan([
            script: nullScript,
            juStabUtils: utils,
            executeVulasScan: true,
            executeWhitesourceScan: true,
            detectExecuteScan: true
        ])

        assertThat(parallelMap.size(), is(3))
        assertThat(stepsCalled, allOf(hasItem('executeVulasScan'), hasItem('executeWhitesourceScan')))
        assertThat(stepsCalled, not(hasItem('detectExecuteScan')))
        assertThat(failFast, is(false))
    }

    @Test
    void testPlainProjectWithVulasDeactivatedBlackDuckActive() {

        jsr.step.executeOpenSourceDependencyScan([
            script: nullScript,
            juStabUtils: utils,
            executeVulasScan: false,
            executeWhitesourceScan: true,
            detectExecuteScan: true
        ])

        assertThat(parallelMap.size(), is(3))
        assertThat(stepsCalled, allOf(hasItem('executeWhitesourceScan'), hasItem('detectExecuteScan')))
        assertThat(stepsCalled, not(hasItem('executeVulasScan')))
        assertThat(failFast, is(false))
    }

    @Test
    void testPlainProjectDeprecated() {

        jsr.step.executeOpenSourceDependencyScan([
            script: nullScript,
            juStabUtils: utils,
            retire: true,
            whitesource: true,
            whitesourceJava: true,
            vulas: true,
            protecode: true
        ])

        assertThat(parallelMap.size(), is(3))
        assertThat(stepsCalled, allOf(hasItem('executeProtecodeScan'), hasItem('executeVulasScan'), hasItem('executeWhitesourceScan')))
    }


    @Test
    void testProjectWithMavenWhiteSource() {

        jsr.step.executeOpenSourceDependencyScan([
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven',
            whitesourceJava: true
        ])

        assertThat(parallelMap.size(), is(3))
        assertThat(stepsCalled, allOf(hasItem('executeVulasScan'), hasItem('executeWhitesourceScan')))
    }

    @Test
    void testProjectWithBuildToolMaven() {

        jsr.step.executeOpenSourceDependencyScan([
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven'
        ])

        assertThat(parallelMap.size(), is(2))
        assertThat(stepsCalled, allOf(hasItem('executeVulasScan')))
    }

    @Test
    void testProjectWithBuildToolNpm() {

        jsr.step.executeOpenSourceDependencyScan([
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'npm'
        ])

        assertThat(parallelMap.size(), is(2))
        assertThat(stepsCalled, allOf(hasItem('executeWhitesourceScan')))
    }

    @Test
    void testOldWhitesourceStepIsConfigured() {

        jsr.step.executeOpenSourceDependencyScan([
            script: nullScript,
            juStabUtils: utils,
            executeWhitesourceScan: true
        ])

        assertThat(parallelMap.size(), is(2))
        assertThat(stepsCalled, allOf(hasItem('executeWhitesourceScan')))
    }

    @Test
    void testNewWhitesourceStepIsConfiguredOldStepIsNull() {

        jsr.step.executeOpenSourceDependencyScan([
            script: nullScript,
            juStabUtils: utils,
            executeWhitesourceScan: null,
            whitesourceExecuteScan: true,
        ])

        assertThat(parallelMap.size(), is(2))
        assertThat(stepsCalled, allOf(hasItem('whitesourceExecuteScan')))
    }

    @Test
    void testMtaProject() {
        helper.registerAllowedMethod("findFiles", [Map.class], {
            map ->
                if(map.glob.equals("**${File.separator}pom.xml"))
                    return [new File('pom.xml'), new File('assembly/pom.xml'), new File('some-ui/pom.xml'),
                            new File('some-service/pom.xml'), new File('some-other-service/pom.xml')].toArray()
                if(map.glob.equals("**${File.separator}package.json"))
                    return [new File('package.json'), new File('some-other-ui/package.json')].toArray()
                if(map.glob.equals("**${File.separator}sbtDescriptor.json"))
                    return [new File('some-scala-module/sbtDescriptor.json'), new File('some-other-scala-module/sbtDescriptor.json')].toArray()
                if(map.glob.equals("**${File.separator}setup.py"))
                    return [new File('some-python-module/setup.py')].toArray()
        })

        helper.registerAllowedMethod("fileExists", [String.class], {
            path ->
                assertThat(path, is('.git/index'))
                return false
        })

        jsr.step.executeOpenSourceDependencyScan([
            script: nullScript,
            juStabUtils: utils,
            exclude: ['pom.xml', 'package.json' ,'some-other-scala-module/sbtDescriptor.json'],
            scanType: 'mta',
            whitesource: true,
            whitesourceJava: false,
            retire: true,
            vulas: true
        ])

        assertThat(parallelMap.size(), is(3))
        assertThat(parallelKeys, allOf(
            hasItem('OpenSourceDependency [VULAS]'),
            hasItem('OpenSourceDependency [whitesource]'))
        )
        assertThat(stepsCalled, allOf(hasItem('executeVulasScan'), hasItem('executeWhitesourceScan')))


        assertThat(jlr.log, containsString('Unstash content: buildDescriptor'))

        assertThat(jlr.log, not(containsString('Adding pom.xml to exclude list')))
        assertThat(jlr.log, containsString("Adding assembly${File.separator}pom.xml to exclude list"))
        assertThat(jlr.log, containsString("Adding some-ui${File.separator}pom.xml to exclude list"))
        assertThat(jlr.log, containsString("Adding some-service${File.separator}pom.xml to exclude list"))
        assertThat(jlr.log, containsString("Adding some-other-service${File.separator}pom.xml to exclude list"))

        assertThat(stepParams, hasItem(hasEntry('scanType', 'mta')))
    }

    @Test
    void testMtaProjectWithBlackDuck() {
        helper.registerAllowedMethod("findFiles", [Map.class], {
            map ->
                if(map.glob.equals("**${File.separator}pom.xml"))
                    return [new File("assembly${File.separator}pom.xml"), new File("some-ui${File.separator}pom.xml"),
                            new File("some-service${File.separator}pom.xml"), new File("some-other-service${File.separator}pom.xml")].toArray()
        })

        helper.registerAllowedMethod("fileExists", [String.class], {
            path ->
                assertThat(path, is('.git/index'))
                return false
        })

        jsr.step.executeOpenSourceDependencyScan([
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'mta',
            executeVulasScan: false,
            executeWhitesourceScan: true,
            detectExecuteScan: true,
            whitesourceJava: false,
            retire: true,
        ])

        assertThat(parallelMap.size(), is(4))
        assertThat(parallelKeys, allOf(
            hasItem('OpenSourceDependency [BlackDuck]'),
            hasItem('OpenSourceDependency [Protecode]'),
            hasItem('OpenSourceDependency [whitesource]'))
        )
        assertThat(stepsCalled, allOf(hasItem('detectExecuteScan'), hasItem('executeWhitesourceScan')))

        assertThat(jlr.log, containsString('Unstash content: buildDescriptor'))

        assertThat(jlr.log, containsString("WhiteSource for Java not activated. Ignoring all maven descriptor files during WhiteSource scan."))
        assertThat(jlr.log, containsString("Adding assembly${File.separator}pom.xml to exclude list"))
        assertThat(jlr.log, containsString("Adding some-ui${File.separator}pom.xml to exclude list"))
        assertThat(jlr.log, containsString("Adding some-service${File.separator}pom.xml to exclude list"))
        assertThat(jlr.log, containsString("Adding some-other-service${File.separator}pom.xml to exclude list"))

        assertThat(jlr.log, containsString("BlackDuck activated. Including only maven and gradle builds during Blackduck detector scan"))
        assertThat(jlr.log, containsString("BlackDuck activated. Adding maven descriptor directory assembly${File.separator} to be included during Blackduck signature scan"))
        assertThat(jlr.log, containsString("BlackDuck activated. Adding maven descriptor directory some-ui${File.separator} to be included during Blackduck signature scan"))
        assertThat(jlr.log, containsString("BlackDuck activated. Adding maven descriptor directory some-service${File.separator} to be included during Blackduck signature scan"))
        assertThat(jlr.log, containsString("BlackDuck activated. Adding maven descriptor directory some-other-service${File.separator} to be included during Blackduck signature scan"))
    }

    @Test
    void testDubProject() {
        helper.registerAllowedMethod("findFiles", [Map.class], {
            map ->
                if(map.glob.equals("**${File.separator}dub.json"))
                    return [new File('dub.json')].toArray()
        })

        jsr.step.executeOpenSourceDependencyScan([
            script: nullScript,
            juStabUtils: utils,
            scanType: 'dub',
            whitesource: true,
            whitesourceJava: false,
            retire: true,
            vulas: false,
            protecode: true
        ])

        assertThat(parallelMap.size(), is(3))
        assertThat(stepsCalled, allOf(hasItem('executeWhitesourceScan')))
    }

    @Test
    void testGolangProject() {

        binding.variables.env.STAGE_NAME = 'Security'
        nullScript.globalPipelineEnvironment.configuration = [runStep: [Security: [whitesourceExecuteScan: true]]]

        jsr.step.executeOpenSourceDependencyScan([
            script                : nullScript,
            juStabUtils           : utils,
            scanType              : 'golang',
            whitesourceExecuteScan: true,
        ])

        assertThat(parallelMap.size(), is(3))
        assertThat(stepsCalled, allOf(hasItem('whitesourceExecuteScan'), not(hasItem('executeWhitesourceScan'))))
    }

    @Test
    void testProtecodeExecuteScan() {

        nullScript.globalPipelineEnvironment.configuration = [
            runStep: [Security: [protecodeExecuteScan: true]]
        ]
        jsr.step.executeOpenSourceDependencyScan([
            script: nullScript,
            juStabUtils: utils,
            stageName: 'Security'
        ])

        assertThat(stepsCalled, hasItem('protecodeExecuteScan'))
    }

    @Test
    void testProtecodeExecuteScanDocker() {

        nullScript.globalPipelineEnvironment.configuration = [
            general: [buildTool: 'docker'],
            runStep: [Security: [protecodeExecuteScan: true]]
        ]
        jsr.step.executeOpenSourceDependencyScan([
            script: nullScript,
            juStabUtils: utils,
            stageName: 'Security'
        ])

        assertThat(stepsCalled, hasItem('protecodeExecuteScan'))
    }
}

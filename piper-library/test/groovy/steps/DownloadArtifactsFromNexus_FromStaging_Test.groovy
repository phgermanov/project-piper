#!groovy
package steps

import util.JenkinsUnstableRule

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.containsInAnyOrder
import static org.hamcrest.Matchers.hasItems
import static org.hamcrest.Matchers.hasKey
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.hamcrest.CoreMatchers.isA

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.Rule
import org.junit.Ignore
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue

import org.yaml.snakeyaml.Yaml

import util.BasePiperTest
import util.Rules
import util.JenkinsStepRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsExecuteDockerRule

import static com.lesfurets.jenkins.unit.MethodCall.callArgsToString
import static com.lesfurets.jenkins.unit.MethodSignature.method
import static org.junit.Assert.fail

class DownloadArtifactsFromNexus_FromStaging_Test extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsReadYamlRule jryr = new JenkinsReadYamlRule(this, 'test/resources/downloadFromNexusTest/mta/')
    protected JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsUnstableRule jur = new JenkinsUnstableRule(this)
    private ExpectedException thrown = ExpectedException.none()
    protected Map xmakeProperties = ['staging_repo_id': '123456789']

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jur)
        .around(jscr)
        .around(jlr)
        .around(jsr)
        .around(jryr)
        .around(thrown)

    @Before
    void init() throws Exception {
        // set jenkins mock commands for Utils.groovy
        binding.setVariable('steps', [
            stash  : { m -> println "stashName = ${m.name}" },
            unstash: { println "unstashName called" }
        ])

        // register Jenkins commands with mock values
        helper.registerAllowedMethod('sh', [Map.class], { m ->
            jscr.shell.add(m.script.toString())
            return m.script.contains('/404')?'404':'200'
        })
        helper.registerAllowedMethod('unstash', [String.class], { s -> println "unstashName called"})
        // load POM file and mock the return value from readMavenPom method
        helper.registerAllowedMethod(method('readMavenPom', Map.class), { m -> return mockHelper.loadPom('test/resources/downloadFromNexusTest/milestone/pom.xml') })
        // load POM file and mock the return value from readMavenPom method
        //helper.registerAllowedMethod(method('readMavenPom', Map.class), { m -> return mockHelper.loadPom('test/resources/downloadFromNexusTest/release/pom.xml') })
        // load JSON file and mock the return value of readJSON method
        helper.registerAllowedMethod(method('readJSON', Map.class), { m -> return mockHelper.loadJSON('test/resources/downloadFromNexusTest/milestone/package.json') })
        // load JSON file and mock the return value of readJSON method
        //helper.registerAllowedMethod(method('readJSON', Map.class), { m -> return mockHelper.loadJSON('test/resources/downloadFromNexusTest/release/package.json') })
        // load configuration file and mock the return value of readProperties method
        // helper.registerAllowedMethod(method('readProperties', Map.class), { m -> return mockHelper.loadProperties('test/resources/config.properties') })
        helper.registerAllowedMethod("wrap", [Map.class, Closure.class], { m, body ->
            body()
        })
        nullScript.commonPipelineEnvironment.setValue("stageBOM", xmakeProperties['stage-bom'])
        nullScript.commonPipelineEnvironment.setValue("xmakeStagingRepositoryId", xmakeProperties['staging_repo_id'])
        nullScript.globalPipelineEnvironment.setXMakeProperty('stage_repourl', xmakeProperties['stage_repourl'])
    }

    @Test
    void testMavenFromStaging() throws Exception {
        helper.registerAllowedMethod('sh', [Map.class], { m ->
            if (m.script.startsWith("curl")) {
                jscr.shell.add(m.script.toString())
                return m.script.contains('/404')?'404':'200'
            }
            else if (m.script.startsWith("mvn")) {
                jscr.shell.add(m.script.toString())

                def scriptCommand = m.script
                if(scriptCommand.contains('project.groupId'))
                    return 'com.sap.suite.cloud.foundation'
                if(scriptCommand.contains('project.artifactId'))
                    return 'pipeline-test'
                if(scriptCommand.contains('project.version'))
                    return '0.0.1-20170301-104821_8ffcbe7'
                if(scriptCommand.contains('project.packaging'))
                    return 'war'
                if(scriptCommand.contains('project.build.finalName'))
                    return 'pipeline-test'
                if(scriptCommand.contains('project.build.directory'))
                    return 'target/'
                if(scriptCommand.contains('project.basedir'))
                    return '.'
            }
        })

        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            fromStaging: true
        )
        // asserts
        assertThat(jlr.log, containsString(String.format('Nexus repository: %s',xmakeProperties.containsKey('stage-bom')?null:'http://nexus.wdf.sap.corp:8081/stage/repository/123456789')))
        assertThat(jscr.shell, hasItem('mkdir --parents \'target/\''))
        assertThat(jscr.shell, hasItem('curl --insecure --basic --user pK2XyGGhvVy4mMa:******** --silent --show-error --write-out \'%{response_code}\' --location --output \'target/pipeline-test.war\' http://nexus.wdf.sap.corp:8081/stage/repository/123456789/com/sap/suite/cloud/foundation/pipeline-test/0.0.1-20170301-104821_8ffcbe7/pipeline-test-0.0.1-20170301-104821_8ffcbe7.war'))
        assertJobStatusSuccess()
    }

    @Test
    void testMavenFromStagingBuildTool() {
        helper.registerAllowedMethod('sh', [Map.class], { m ->
            if (m.script.startsWith("curl")) {
                jscr.shell.add(m.script.toString())
                return m.script.contains('/404')?'404':'200'
            }
            else if (m.script.startsWith("mvn")) {
                jscr.shell.add(m.script.toString())

                def scriptCommand = m.script
                if(scriptCommand.contains('project.groupId'))
                    return 'com.sap.suite.cloud.foundation'
                if(scriptCommand.contains('project.artifactId'))
                    return 'pipeline-test'
                if(scriptCommand.contains('project.version'))
                    return '0.0.1-20170301-104821_8ffcbe7'
                if(scriptCommand.contains('project.packaging'))
                    return 'war'
                if(scriptCommand.contains('project.build.finalName'))
                    return 'pipeline-test'
                if(scriptCommand.contains('project.build.directory'))
                    return 'target/'
                if(scriptCommand.contains('project.basedir'))
                    return '.'
            }
        })

        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven',
            fromStaging: true
        )
        // asserts
        assertThat(jlr.log, containsString(String.format('Nexus repository: %s',xmakeProperties.containsKey('stage-bom')?null:'http://nexus.wdf.sap.corp:8081/stage/repository/123456789')))
        assertThat(jscr.shell, hasItem('mkdir --parents \'target/\''))
        assertThat(jscr.shell, hasItem('curl --insecure --basic --user pK2XyGGhvVy4mMa:******** --silent --show-error --write-out \'%{response_code}\' --location --output \'target/pipeline-test.war\' http://nexus.wdf.sap.corp:8081/stage/repository/123456789/com/sap/suite/cloud/foundation/pipeline-test/0.0.1-20170301-104821_8ffcbe7/pipeline-test-0.0.1-20170301-104821_8ffcbe7.war'))
        assertJobStatusSuccess()
    }

    @Test
    void testMavenFromStagingLegacy() throws Exception {
        helper.registerAllowedMethod('sh', [Map.class], { m ->
            if (m.script.startsWith("curl")) {
                jscr.shell.add(m.script.toString())
                return m.script.contains('/404')?'404':'200'
            }
            else if (m.script.startsWith("mvn")) {
                jscr.shell.add(m.script.toString())

                def scriptCommand = m.script
                if(scriptCommand.contains('project.groupId'))
                    return 'com.sap.suite.cloud.foundation'
                if(scriptCommand.contains('project.artifactId'))
                    return 'pipeline-test'
                if(scriptCommand.contains('project.version'))
                    return '0.0.1-20170301-104821_8ffcbe7'
                if(scriptCommand.contains('project.packaging'))
                    return 'war'
                if(scriptCommand.contains('project.build.finalName'))
                    return 'pipeline-test'
                if(scriptCommand.contains('project.build.directory'))
                    return 'target/'
                if(scriptCommand.contains('project.basedir'))
                    return '.'
            }
        })

        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            fromStaging: true,
            nexusStageFilePath: 'http://test.wdf.sap.corp/myRepository'
        )
        // asserts
        assertThat(jlr.log, containsString(String.format('Nexus repository: %s', 'http://test.wdf.sap.corp/myRepository')))
        assertThat(jscr.shell, hasItem('mkdir --parents \'target/\''))
        assertThat(jscr.shell, hasItem(String.format('curl --insecure --basic --user pK2XyGGhvVy4mMa:******** --silent --show-error --write-out \'%%{response_code}\' --location --output \'target/pipeline-test.war\' %s', 'http://test.wdf.sap.corp/myRepository/com/sap/suite/cloud/foundation/pipeline-test/0.0.1-20170301-104821_8ffcbe7/pipeline-test-0.0.1-20170301-104821_8ffcbe7.war')))
        assertJobStatusSuccess()
    }

    @Test
    void testJavaFromStaging() throws Exception {
        // register Jenkins commands with mock values
        helper.registerAllowedMethod('sh', [Map.class], { m ->
            if (m.script.startsWith("curl")) {
                jscr.shell.add(m.script.toString())
                return m.script.contains('/404')?'404':'200'
            }
            else if (m.script.startsWith("mvn")) {
                jscr.shell.add(m.script.toString())

                def scriptCommand = m.script
                if(scriptCommand.contains('project.groupId'))
                    return 'com.sap.suite.cloud.foundation'
                if(scriptCommand.contains('project.artifactId'))
                    return 'pipeline-test'
                if(scriptCommand.contains('project.version'))
                    return '0.0.1-20170301-104821_8ffcbe7'
                if(scriptCommand.contains('project.packaging'))
                    return 'war'
                if(scriptCommand.contains('project.build.finalName'))
                    return 'pipeline-test'
                if(scriptCommand.contains('project.build.directory'))
                    return 'target/'
                if(scriptCommand.contains('project.basedir'))
                    return '.'
            }
        })

        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            fromStaging: true,
            artifactType: 'java'
        )
        // asserts
        assertThat(jlr.log, containsString(String.format('Nexus repository: %s',xmakeProperties.containsKey('stage-bom')?null:'http://nexus.wdf.sap.corp:8081/stage/repository/123456789')))
        assertThat(jscr.shell, hasItem('mkdir --parents \'target/\''))
        assertThat(jscr.shell, hasItem('curl --insecure --basic --user pK2XyGGhvVy4mMa:******** --silent --show-error --write-out \'%{response_code}\' --location --output \'target/pipeline-test.war\' http://nexus.wdf.sap.corp:8081/stage/repository/123456789/com/sap/suite/cloud/foundation/pipeline-test/0.0.1-20170301-104821_8ffcbe7/pipeline-test-0.0.1-20170301-104821_8ffcbe7.war'))
        assertJobStatusSuccess()
    }

    @Test
    void testUserDefinedGAVFromStaging() throws Exception {
        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            fromStaging: true,
            artifactType: 'nothing',
            group: 'com.sap.suite.cloud.foundation',
            artifactId: 'pipeline-test',
            artifactVersion: '0.0.1-20170301-104821_8ffcbe7',
            packaging: 'war',
        )
        println("LOG: ${jlr.log}")
        // asserts
        assertThat(jlr.log, containsString(String.format('Nexus repository: %s',xmakeProperties.containsKey('stage-bom')?null:'http://nexus.wdf.sap.corp:8081/stage/repository/123456789')))
        assertThat(jscr.shell, hasItem('mkdir --parents \'target/\''))
        assertThat(jscr.shell, hasItem(containsString('--output \'target/pipeline-test.war\' http://nexus.wdf.sap.corp:8081/stage/repository/123456789/com/sap/suite/cloud/foundation/pipeline-test/0.0.1-20170301-104821_8ffcbe7/pipeline-test-0.0.1-20170301-104821_8ffcbe7.war')))
        assertJobStatusSuccess()
    }

    @Test
    void testNpmFromStaging() throws Exception {
        // load JSON file and mock the return value of readJSON method
        helper.registerAllowedMethod(method('readJSON', Map.class), { m -> return mockHelper.loadJSON('test/resources/downloadFromNexusTest/staging/package.json') })

        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            fromStaging: true,
            artifactType: 'npm'
        )
        println("LOG: ${jlr.log}")
        // asserts
        assertThat(jlr.log, containsString(String.format('Nexus repository: %s',xmakeProperties.containsKey('stage-bom')?null:'http://nexus.wdf.sap.corp:8081/stage/repository/123456789')))
        assertThat(jscr.shell, not(hasItem('mkdir --parents \'target/\'')))
        def version = '1.1.11-20170222101742'+(nullScript.commonPipelineEnvironment.getValue("stageBOM")?'':'+a6181bdecc50cc829329d3b38671585ac5c99ed0')
        assertThat(jscr.shell, hasItem(containsString("--output 'sap-notification-hub.tgz' http://nexus.wdf.sap.corp:8081/stage/repository/123456789/com/sap/npm/sap-notification-hub/${version}/sap-notification-hub-${version}.tar.gz")))
        assertJobStatusSuccess()
    }

    @Test
    void testNpmBundleFromStaging() throws Exception {
        // load JSON file and mock the return value of readJSON method
        helper.registerAllowedMethod(method('readJSON', Map.class), { m -> return mockHelper.loadJSON('test/resources/downloadFromNexusTest/staging/package.json') })

        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            fromStaging: true,
            artifactType: 'npm',
            classifier: 'bundle'
        )
        println("LOG: ${jlr.log}")
        // asserts
        assertThat(jlr.log, containsString(String.format('Nexus repository: %s',xmakeProperties.containsKey('stage-bom')?null:'http://nexus.wdf.sap.corp:8081/stage/repository/123456789')))
        assertThat(jscr.shell, not(hasItem('mkdir --parents \'target/\'')))
        assertThat(jscr.shell, hasItem(containsString('--output \'sap-notification-hub.tar.gz\' http://nexus.wdf.sap.corp:8081/stage/repository/123456789/com/sap/npm/sap-notification-hub/1.1.11-20170222101742+a6181bdecc50cc829329d3b38671585ac5c99ed0/sap-notification-hub-1.1.11-20170222101742+a6181bdecc50cc829329d3b38671585ac5c99ed0-bundle.tar.gz')))
        assertJobStatusSuccess()
    }

    @Test
    void testNpmFromStagingBuildTool(){
        // load JSON file and mock the return value of readJSON method
        helper.registerAllowedMethod(method('readJSON', Map.class), { m -> return mockHelper.loadJSON('test/resources/downloadFromNexusTest/staging/package.json') })

        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            fromStaging: true,
            buildTool: 'npm'
        )
        println("LOG: ${jlr.log}")
        // asserts
        def version = '1.1.11-20170222101742'+(nullScript.commonPipelineEnvironment.getValue("stageBOM")?'':'+a6181bdecc50cc829329d3b38671585ac5c99ed0')
        assertThat(jlr.log, containsString(String.format('Nexus repository: %s',xmakeProperties.containsKey('stage-bom')?null:'http://nexus.wdf.sap.corp:8081/stage/repository/123456789')))
        assertThat(jscr.shell, not(hasItem('mkdir --parents \'target/\'')))
        assertThat(jscr.shell, hasItem(containsString("--output 'sap-notification-hub.tgz' http://nexus.wdf.sap.corp:8081/stage/repository/123456789/com/sap/npm/sap-notification-hub/${version}/sap-notification-hub-${version}.tar.gz")))
        assertJobStatusSuccess()
    }

    @Test
    void testLegacyStagingRepository() throws Exception {
    	def testXmakeProperties = xmakeProperties.containsKey('stage_repourl')?xmakeProperties:[ 'stage_repourl': 'http://nexus.wdf.sap.corp:8081/nexus/content/repositories/123456789', 'staging_repo_id': '123456789' ]

        nullScript.globalPipelineEnvironment.setXMakeProperties(testXmakeProperties)

        nullScript.globalPipelineEnvironment.setXMakeProperty('stage_repourl', testXmakeProperties['stage_repourl'])
        nullScript.commonPipelineEnvironment.setValue("xmakeStagingRepositoryId", testXmakeProperties['staging_repo_id'])

        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            fromStaging: true
        )
        // asserts
        assertThat(jlr.log, containsString(String.format('Nexus repository: %s',xmakeProperties.containsKey('stage-bom')?null:testXmakeProperties.stage_repourl)))
        assertJobStatusSuccess()
    }

    @Test
    void testMultiModuleMavenProjects() {
        // This test checks that for multi-module maven projects
        // the step downloads all deployable module artifacts and properly
        // renames them.

        HashSet<String> pomFilesLoaded = []
        def counter = 0
        // register mock readMavenPom method that reads the test pom.xml
        helper.registerAllowedMethod('readMavenPom', [Map.class], { m ->
            counter++
            String pomFile = m.file
            pomFilesLoaded.add(pomFile)
            println("POM: ${pomFile}")
            if(counter % 2 == 0)
                return null
            if("pom.xml" == pomFile) {
                return mockHelper.loadPom('test/resources/downloadFromNexusTest/maven/pom.xml')
            }
            else if ("./cds-build/pom.xml" == pomFile) {
                return mockHelper.loadPom('test/resources/downloadFromNexusTest/maven/cds-build/pom.xml')
            }
            else if ("./srv/pom.xml" == pomFile) {
                return mockHelper.loadPom('test/resources/downloadFromNexusTest/maven/srv/pom.xml')
            }
            else {
                fail("Unsupported pom.xml location found. This is most likely an error in the test.")
            }
        })

        int moduleCounter = 0
        List<String> curledArtifacts = []
        // register Jenkins commands with mock values
        helper.registerAllowedMethod('sh', [Map.class], { m ->
            if (m.script.startsWith("curl")) {
                String curlCommand = m.script
                jscr.shell.add(curlCommand.toString())

                if(curlCommand.contains("dummy-${moduleCounter}")) {
                    curledArtifacts.add("dummy-${moduleCounter}")
                }

                return m.script.contains('/404')?'404':'200'
            }
            else if (m.script.contains("help:evaluate")) { // mock the mvn help:evaluate call.
                String scriptCommand = m.script
                if(scriptCommand.contains('project.groupId'))
                    return 'dummy'
                if(scriptCommand.contains('project.artifactId'))
                    return 'dummy'
                if(scriptCommand.contains('project.version'))
                    return '1.2.34'
                if(scriptCommand.contains('project.packaging'))
                    return 'jar'
                if(scriptCommand.contains('project.build.finalName')) { // once for every module we should evaluate the finalName.
                    moduleCounter++
                    return "dummy-${moduleCounter}"
                }
                if(scriptCommand.contains('project.build.directory'))
                    return 'target/'
                if(scriptCommand.contains('project.basedir'))
                    return '.'
            }
        })

        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            fromStaging: true,
            artifactType: 'java'
        )
        // asserts
        assertThat(moduleCounter, is(3)) // check that all modules were processed.
        assertThat(pomFilesLoaded?.size(), is(3))
        assertThat(pomFilesLoaded, hasItems("pom.xml", "./cds-build/pom.xml", "./srv/pom.xml"))
        assertThat(jlr.log, containsString(String.format('Nexus repository: %s',xmakeProperties.containsKey('stage-bom')?null:'http://nexus.wdf.sap.corp:8081/stage/repository/123456789')))
        assertThat(jscr.shell, hasItem('mkdir --parents \'target/\''))
        assertThat(jscr.shell, hasItem('curl --insecure --basic --user pK2XyGGhvVy4mMa:******** --silent --show-error --write-out \'%{response_code}\' --location --output \'target/dummy-1.jar\' http://nexus.wdf.sap.corp:8081/stage/repository/123456789/dummy/dummy/1.2.34/dummy-1.2.34.jar'))
        assertThat(jscr.shell, hasItem('curl --insecure --basic --user pK2XyGGhvVy4mMa:******** --silent --show-error --write-out \'%{response_code}\' --location --output \'target/dummy-2.jar\' http://nexus.wdf.sap.corp:8081/stage/repository/123456789/dummy/dummy/1.2.34/dummy-1.2.34.jar'))
        assertThat(jscr.shell, hasItem('curl --insecure --basic --user pK2XyGGhvVy4mMa:******** --silent --show-error --write-out \'%{response_code}\' --location --output \'target/dummy-3.jar\' http://nexus.wdf.sap.corp:8081/stage/repository/123456789/dummy/dummy/1.2.34/dummy-1.2.34.jar'))
        //assertTrue(["dummy-1", "dummy-2", "dummy-3"] == curledArtifacts) // check that exactly three artifacts (and not more) were curled
        assertJobStatusSuccess()
    }

    @Test
    void testVersionReplacementInFinalName() {
        // This test checks that the version in the final
        // name read / evaluated from the POM is replaced
        // with the original version, if it was changed
        // (e.g. by artifactPrepareVersion step)

        String versionBeforeAutoVersioning = "0.0.1"
        String versionAfterAutoVersioning = "0.0.1-<Timestamp>_<CommitId>"
        String finalNameAfterAutomaticVersioning = "dummy-0.0.1-<Timestamp>_<CommitId>"
        String expectedFinalName = "dummy-0.0.1"

        def counter = 0
        helper.registerAllowedMethod('readMavenPom', [Map.class], {
            m ->
                counter++
                String pomFile = m.file
                println("POM: ${pomFile}")
                if(counter % 2 == 0)
                    return null
                if("pom.xml" == pomFile) {
                    return mockHelper.loadPom('test/resources/downloadFromNexusTest/milestone/pom.xml')
                }
                else {
                    fail("Unsupported pom.xml location found. This is most likely an error in the test.")
                }
        })

        // register Jenkins commands with mock values
        helper.registerAllowedMethod('sh', [Map.class], { m ->
            if (m.script.startsWith("curl")) {
                String curlCommand = m.script
                jscr.shell.add(curlCommand.toString())
                return m.script.contains('/404')?'404':'200'
            }
            else if (m.script.contains("help:evaluate")) { // mock the mvn help:evaluate call.
                String scriptCommand = m.script
                if(scriptCommand.contains('project.groupId'))
                    return 'dummy'
                if(scriptCommand.contains('project.artifactId'))
                    return 'dummy'
                if(scriptCommand.contains('project.version'))
                    return versionAfterAutoVersioning
                if(scriptCommand.contains('project.packaging'))
                    return 'jar'
                if(scriptCommand.contains('project.build.finalName')) { // once for every module we should evaluate the finalName.
                    return finalNameAfterAutomaticVersioning
                }
                if(scriptCommand.contains('project.build.directory'))
                    return 'target/'
                if(scriptCommand.contains('project.basedir'))
                    return '.'
            }
        })

        nullScript.globalPipelineEnvironment.versionBeforeAutoVersioning = versionBeforeAutoVersioning

        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            fromStaging: true,
            artifactType: 'java'
        )
        // asserts
        String expectedCurlCommand = "curl --insecure --basic --user pK2XyGGhvVy4mMa:******** --silent --show-error --write-out \'%{response_code}\' --location --output \'target/${expectedFinalName}.jar\' http://nexus.wdf.sap.corp:8081/stage/repository/123456789/dummy/dummy/${versionAfterAutoVersioning}/${finalNameAfterAutomaticVersioning}.jar"

        assertThat(jlr.log, containsString(String.format('Nexus repository: %s',xmakeProperties.containsKey('stage-bom')?null:'http://nexus.wdf.sap.corp:8081/stage/repository/123456789')))
        assertThat(jscr.shell, hasItem('mkdir --parents \'target/\''))
        assertThat(jscr.shell, hasItem(expectedCurlCommand))

        assertJobStatusSuccess()
    }

    @Test
    void testGolangFromStaging() {
        // register Jenkins commands with mock values
        helper.registerAllowedMethod('sh', [Map.class], { m ->
            if (m.script.startsWith("curl")) {
                jscr.shell.add(m.script.toString())
                return m.script.contains('/404')?'404':'200'
            }
        })
        // load .xmake.cfg and mock return value from readFile method
        helper.registerAllowedMethod(method('readFile', String.class), { m -> return mockHelper.loadFile('test/resources/downloadFromNexusTest/golang/.xmake.cfg') })
        // mock fileExists method - true for files loaded from src/test/resources...
        helper.registerAllowedMethod(method('fileExists', String.class), { s -> return s != 'cfg/VERSION'})

        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            fromStaging: true,
            artifactType: 'golang',
            artifactVersion: '1.0.0'
        )
        // asserts
        assertThat(jlr.log, containsString(String.format('Nexus repository: %s',xmakeProperties.containsKey('stage-bom')?null:'http://nexus.wdf.sap.corp:8081/stage/repository/123456789')))
        assertThat(jscr.shell, hasItem('curl --insecure --basic --user pK2XyGGhvVy4mMa:******** --silent --show-error --write-out \'%{response_code}\' --location --output \'Pipeline-Test.tar.gz\' http://nexus.wdf.sap.corp:8081/stage/repository/123456789/com/sap/golang/github/wdf/sap/corp/test-org/Pipeline-Test/1.0.0/Pipeline-Test-1.0.0-linux-amd64.tar.gz'))
        assertJobStatusSuccess()
    }
}

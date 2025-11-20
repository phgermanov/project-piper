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

class DownloadArtifactsFromNexusTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsReadYamlRule jryr = new JenkinsReadYamlRule(this, 'test/resources/downloadFromNexusTest/mta/')
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsUnstableRule jur = new JenkinsUnstableRule(this)
    private ExpectedException thrown = ExpectedException.none()

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
    }

    @Test
    void testGetArtifactUrl() throws Exception {
        def result = jsr.step.downloadArtifactsFromNexus.getArtifactUrl(
            'http://nexus.wdf.sap.corp:8081/stage/repository/123456789',
            'very.long.groupId',
            'artifactId',
            'artifactVersion',
            'packaging',
            'classifier'
        ).toString()
        // asserts
        assertThat(result, is('http://nexus.wdf.sap.corp:8081/stage/repository/123456789/very/long/groupId/artifactId/artifactVersion/artifactId-artifactVersion-classifier.packaging'))
        assertJobStatusSuccess()
    }

    @Test
    void testDownloadSpecificArtifact() throws Exception {
        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            nexusUrl: 'http://test.com/',
            group: 'com/sap/anything',
            artifactId: 'sap-pipeline-test',
            artifactVersion: '1.2.3',
            packaging: 'tar.gz'
        )
        // asserts
        assertThat(jscr.shell, hasItem(containsString('http://test.com/nexus/content/repositories/deploy.milestones/com/sap/anything/sap-pipeline-test/1.2.3/sap-pipeline-test-1.2.3.tar.gz')))
        assertThat(nullScript.globalPipelineEnvironment.nexusLastDownloadUrl, is('http://test.com/nexus/content/repositories/deploy.milestones/com/sap/anything/sap-pipeline-test/1.2.3/sap-pipeline-test-1.2.3.tar.gz'))
        assertJobStatusSuccess()
    }

    //verify correction for issue https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues/1138
    @Test
    void testDownloadSpecificArtifactNpm() throws Exception {
        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            artifactType: 'npm',
            nexusUrl: 'http://test.com/',
            group: '',
            artifactId: 'sap-pipeline-test',
            artifactVersion: '1.2.3',
            packaging: 'tar.gz'
        )
        // asserts
        assertThat(jscr.shell, hasItem(containsString('http://test.com/nexus/content/repositories/deploy.milestones.npm/sap-pipeline-test/-/sap-pipeline-test-1.2.3.tar.gz')))
        assertJobStatusSuccess()
    }

    //verify correction for issue https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues/2197
    @Test
    void testDownloadSpecificArtifactNpmBundle() throws Exception {
        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            artifactType: 'npm',
            nexusUrl: 'http://test.com/',
            group: null,
            artifactId: 'sap-notification-hub',
            artifactVersion: '1.1.9-20161022151707+c7ce1d979f49b83ffcbdb9d3603e3236361a64ed',
            packaging: 'tar.gz',
            classifier: 'bundle' // with bundle, download from deploy.milestones rather than deploy.milestones.npm
        )
        // asserts
        assertThat(jscr.shell, hasItem(containsString('http://test.com/nexus/content/repositories/deploy.milestones/com/sap/npm/sap-notification-hub/1.1.9-20161022151707+c7ce1d979f49b83ffcbdb9d3603e3236361a64ed/sap-notification-hub-1.1.9-20161022151707+c7ce1d979f49b83ffcbdb9d3603e3236361a64ed-bundle.tar.gz')))
        assertJobStatusSuccess()
    }

    @Test
    void testMavenFromRelease() throws Exception {
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
            juStabUtils: utils
        )
        // asserts
        assertThat(jlr.log, containsString('Nexus repository: http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.milestones/'))
        assertThat(jscr.shell, hasItem('mkdir --parents \'target/\''))
        assertThat(jscr.shell, hasItem('curl --insecure  --silent --show-error --write-out \'%{response_code}\' --location --output \'target/pipeline-test.war\' http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.milestones/com/sap/suite/cloud/foundation/pipeline-test/0.0.1-20170301-104821_8ffcbe7/pipeline-test-0.0.1-20170301-104821_8ffcbe7.war'))
        assertJobStatusSuccess()
    }

    @Test
    void testMavenMTA() throws Exception {
        helper.registerAllowedMethod('sh', [Map.class], { m ->
            if (m.script.startsWith("curl")) {
                jscr.shell.add(m.script.toString())
                return m.script.contains('/404')?'404':'200'
            }
            else if (m.script.startsWith("mvn")) {
                jscr.shell.add(m.script.toString())

                def scriptCommand = m.script
                if(scriptCommand.contains('project.groupId'))
                    return 'com.sap.icf.samples.cc'
                if(scriptCommand.contains('project.artifactId'))
                    return 'currency-rates-assembly'
                if(scriptCommand.contains('project.version'))
                    return '0.1.0-20170215-113917_198ad63f21a122c6137cdd2402f3e4a59aac6e60'
                if(scriptCommand.contains('project.packaging'))
                    return 'pom'
                if(scriptCommand.contains('project.build.finalName'))
                    return 'currency-rates-assembly'
                if(scriptCommand.contains('project.build.directory'))
                    return 'target/'
                if(scriptCommand.contains('project.basedir'))
                    return '.'
            }
        })

        // load POM file and mock the return value from readMavenPom method
        helper.registerAllowedMethod(method('readMavenPom', Map.class), { m -> return mockHelper.loadPom('test/resources/downloadFromNexusTest/assembly/pom.xml') })
        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            artifactType: 'mta'
        )
        // asserts
        assertThat(jlr.log, containsString('from assembly/pom.xml'))
        assertThat(jlr.log, containsString('Nexus repository: http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.milestones/'))
        assertThat(jscr.shell, hasItem(containsString('--output \'target/currency-rates-assembly.mtar\' http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.milestones/com/sap/icf/samples/cc/currency-rates-assembly/0.1.0-20170215-113917_198ad63f21a122c6137cdd2402f3e4a59aac6e60/currency-rates-assembly')))
        assertJobStatusSuccess()
    }

    @Test
    void testMavenMTAWithCustomAssemblyPath() throws Exception {
        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            artifactType: 'mta',
            assemblyPath: 'my-fancy-assembly-module'
        )
        // asserts
        assertThat(jlr.log, containsString('from my-fancy-assembly-module/pom.xml'))
        assertJobStatusSuccess()
    }

    @Test
    void testNpmFromReleaseWithMilestoneQuality() throws Exception {
        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            artifactType: 'npm'
        )
        // asserts
        assertThat(jlr.log, containsString('Nexus repository: http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.milestones.npm/'))
        assertThat(jscr.shell, not(hasItem('mkdir --parents \'target/\'')))
        assertThat(jscr.shell, hasItem(containsString('--output \'sap-notification-hub.tgz\' http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.milestones.npm/sap-notification-hub/-/sap-notification-hub-1.1.9-20161022151707.tgz')))
        assertJobStatusSuccess()
    }

    //verify correction for issue https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues/2197
    @Test
    void testNpmFromReleaseWithMilestoneQualityAndBundleClassifier() throws Exception {
        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            artifactType: 'npm',
            classifier: 'bundle'
        )
        // asserts
        assertThat(jlr.log, containsString('Nexus repository: http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.milestones/'))
        assertThat(jscr.shell, not(hasItem('mkdir --parents \'target/\'')))
        assertThat(jscr.shell, hasItem(containsString('--output \'sap-notification-hub.tar.gz\' http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.milestones/com/sap/npm/sap-notification-hub/1.1.9-20161022151707+c7ce1d979f49b83ffcbdb9d3603e3236361a64ed/sap-notification-hub-1.1.9-20161022151707+c7ce1d979f49b83ffcbdb9d3603e3236361a64ed-bundle.tar.gz')))
        assertJobStatusSuccess()
    }

    @Test
    void testNpmDownloadReleaseWithReleaseQuality() throws Exception {
        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            artifactType: 'npm',
            xMakeBuildQuality: 'Release'
        )
        // asserts
        assertThat(jlr.log, containsString('Nexus repository: http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.releases.npm/'))
        assertThat(jscr.shell, not(hasItem('mkdir --parents \'target/\'')))
        assertThat(jscr.shell, hasItem(containsString('--output \'sap-notification-hub.tgz\' http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.releases.npm/sap-notification-hub/-/sap-notification-hub-1.1.9-20161022151707.tgz')))
        assertJobStatusSuccess()
    }

    //verify correction for issue https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues/2197
    @Test
    void testNpmDownloadReleaseWithReleaseQualityAndBundleClassifier() throws Exception {
        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            artifactType: 'npm',
            xMakeBuildQuality: 'Release',
            classifier: 'bundle'
        )
        // asserts
        assertThat(jlr.log, containsString('Nexus repository: http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.releases/'))
        assertThat(jscr.shell, not(hasItem('mkdir --parents \'target/\'')))
        assertThat(jscr.shell, hasItem(containsString('--output \'sap-notification-hub.tar.gz\' http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.releases/com/sap/npm/sap-notification-hub/1.1.9-20161022151707+c7ce1d979f49b83ffcbdb9d3603e3236361a64ed/sap-notification-hub-1.1.9-20161022151707+c7ce1d979f49b83ffcbdb9d3603e3236361a64ed-bundle.tar.gz')))
        assertJobStatusSuccess()
    }

    @Test
    void testMta() throws Exception {
        // load .xmake.cfg and mock return value from readProperties method
        helper.registerAllowedMethod(method('readProperties', Map.class), { m -> return mockHelper.loadProperties('test/resources/downloadFromNexusTest/mta/.xmake.cfg') })
        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            fromStaging: false,
            artifactType: 'mta',
            buildTool: 'mta'
        )
        // asserts
        assertThat(jlr.log, containsString('from mta.yaml and .xmake.cfg'))
        assertThat(jscr.shell, hasItem(containsString('--output \'target/com.sap.icf.samples.shoppinglist.mtar\' http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.milestones/com/sap/icf/samples/com.sap.icf.samples.shoppinglist/0.3.1/com.sap.icf.samples.shoppinglist-0.3.1.mtar')))
        assertJobStatusSuccess()
    }

    @Test
    void testMtaWithDefaultGroup() throws Exception {
        // load .xmake.cfg and mock return value from readProperties method
        helper.registerAllowedMethod(method('readProperties', Map.class), { m -> return [:] })
        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            fromStaging: false,
            artifactType: 'mta',
            buildTool: 'mta'
        )
        // asserts
        assertThat(jlr.log, containsString('[WARNING] No groupID set in \'.xmake.cfg\', using default groupID \'com.sap.prd.xmake.example.mtars\'.'))
        assertThat(jlr.log, containsString('group:com.sap.prd.xmake.example.mtars'))
        assertJobStatusSuccess()
    }

    @Test
    void testDockerFromReleaseWithReleaseQuality() throws Exception {
        // load .xmake.cfg and mock return value from readFile method
        helper.registerAllowedMethod(method('readFile', String.class), { m -> return mockHelper.loadFile('test/resources/downloadFromNexusTest/dockerbuild-releaseMetadata/.xmake.cfg') })
        // mock fileExists method - true for files loaded from src/test/resources...
        helper.registerAllowedMethod(method('fileExists', String.class), { s -> return s != 'cfg/VERSION'})

        jsr.step.downloadArtifactsFromNexus(
            script: nullScript,
            juStabUtils: utils,
            artifactType: 'dockerbuild-releaseMetadata',
            xMakeBuildQuality: 'Milestone'
        )
        // asserts
        assertThat(jlr.log, containsString('Nexus repository: http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.milestones/'))
        assertThat(jscr.shell, hasItem('mkdir --parents \'target/\''))
        assertThat(jscr.shell, hasItem(containsString('--output \'target/uwes-test.zip\' http://nexus.wdf.sap.corp:8081/nexus/content/repositories/deploy.milestones/com/sap/prd/test/uwes-test/1.0.7/uwes-test-1.0.7-releaseMetadata.zip')))
        assertJobStatusSuccess()
    }
}

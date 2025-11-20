#!groovy
package steps

import org.hamcrest.Matchers
import org.junit.Assert
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

class ExecutePitTestsTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsExecuteDockerRule jedr = new JenkinsExecuteDockerRule(this)
    private ExpectedException exception = ExpectedException.none()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(exception)
        .around(jedr)
        .around(jscr)
        .around(jlr)
        .around(jsr) // needs to be activated after jedr, otherwise executeDocker is not mocked

    @Test
    void testPitCalled() {
        def parameters = [script: nullScript, juStabUtils: utils, jenkinsUtilsStub: jenkinsUtils]

        helper.registerAllowedMethod("fileExists", [String.class], {map -> return false})
        helper.registerAllowedMethod("httpRequest", [String.class], {string ->return [status: 200, content: 'testContent']})
        helper.registerAllowedMethod("writeFile", [HashMap.class], {map -> return})

        helper.registerAllowedMethod("isJobStartedByTimer", [], {
            return false
        })

        def pitCalled = false
        helper.registerAllowedMethod("sh", [String.class], {
            string ->
                Assert.assertThat(string.toString(), Matchers.startsWith("mvn --global-settings .pipeline/mavenGlobalSettings.xml --batch-mode --file ./pom.xml -DtimestampedReports=false -DcoverageThreshold=50 -DmutationThreshold=50 clean process-test-classes org.pitest:pitest-maven:mutationCoverage"))
                pitCalled = true
        })

        def publishCalled = false
        helper.registerAllowedMethod("publishHTML", [HashMap.class], {
            map ->
                publishCalled = true
                Assert.assertEquals('index.html'.toString() , map.reportFiles.toString())
                Assert.assertEquals('PIT Report'.toString(), map.reportName.toString())
                Assert.assertEquals("${File.separator}target${File.separator}pit-reports${File.separator}".toString(), map.reportDir.toString())
        })

        def archiveCalled = false
        helper.registerAllowedMethod("archiveArtifacts", [Map.class], {
            map ->
                archiveCalled = true
                Assert.assertEquals('**/target/pit-reports/index.html'.toString() , map.artifacts.toString())
                Assert.assertTrue(map.allowEmptyArchive)
        })

        def findFilesCalled = false
        helper.registerAllowedMethod("findFiles", [Map.class], {
            map ->
                findFilesCalled = true
                return [new File("/target/pit-reports/index.html")].toArray()
        })


        jsr.step.executePitTests(parameters)

        Assert.assertTrue(pitCalled)
        Assert.assertTrue(publishCalled)
        Assert.assertTrue(archiveCalled)
        Assert.assertTrue(findFilesCalled)
        Assert.assertEquals("maven:3.6-jdk-8", jedr.dockerParams.dockerImage)
        Assert.assertThat(jedr.dockerParams.stashContent.size(), Matchers.is(2))
        Assert.assertThat(jedr.dockerParams.stashContent, Matchers.hasItem('tests'))
        Assert.assertThat(jedr.dockerParams.stashContent, Matchers.hasItem('buildDescriptor'))
    }

    @Test
    void testRunOnlyScheduledDoesNotRunOnNormalPipelineRun() {
        def parameters = [script: nullScript, juStabUtils: utils, jenkinsUtilsStub: jenkinsUtils, runOnlyScheduled: true]

        helper.registerAllowedMethod("fileExists", [String.class], {map -> return false})
        helper.registerAllowedMethod("httpRequest", [String.class], {string ->return [status: 200, content: 'testContent']})
        helper.registerAllowedMethod("writeFile", [HashMap.class], {map -> return})

        helper.registerAllowedMethod("isJobStartedByTimer", [], {
            return false
        })

        def pitNotCalled = true
        helper.registerAllowedMethod("sh", [String.class], {
            string ->
                pitNotCalled = false
        })

        jsr.step.executePitTests(parameters)

        Assert.assertTrue(pitNotCalled)
    }

    @Test
    void testRunOnlyScheduledRunsWhenScheduled() {
        def parameters = [script: nullScript, juStabUtils: utils, jenkinsUtilsStub: jenkinsUtils, runOnlyScheduled: true]

        helper.registerAllowedMethod("fileExists", [String.class], {map -> return false})
        helper.registerAllowedMethod("httpRequest", [String.class], {string ->return [status: 200, content: 'testContent']})
        helper.registerAllowedMethod("writeFile", [HashMap.class], {map -> return})

        helper.registerAllowedMethod("isJobStartedByTimer", [], {
            return true
        })

        def pitCalled = false
        helper.registerAllowedMethod("sh", [String.class], {
            string ->
                Assert.assertThat(string.toString(), Matchers.startsWith("mvn --global-settings .pipeline/mavenGlobalSettings.xml --batch-mode --file ./pom.xml -DtimestampedReports=false -DcoverageThreshold=50 -DmutationThreshold=50 clean process-test-classes org.pitest:pitest-maven:mutationCoverage"))
                pitCalled = true
        })

        def publishCalled = false
        helper.registerAllowedMethod("publishHTML", [HashMap.class], {
            map ->
                publishCalled = true
                Assert.assertEquals('index.html'.toString() , map.reportFiles.toString())
                Assert.assertEquals('PIT Report'.toString(), map.reportName.toString())
                Assert.assertEquals("${File.separator}target${File.separator}pit-reports${File.separator}".toString(), map.reportDir.toString())
        })

        def findFilesCalled = false
        helper.registerAllowedMethod("findFiles", [Map.class], {
            map ->
                findFilesCalled = true
                return [new File("/target/pit-reports/index.html")].toArray()
        })

        def archiveCalled = false
        helper.registerAllowedMethod("archiveArtifacts", [Map.class], {
            map ->
                archiveCalled = true
                Assert.assertEquals('**/target/pit-reports/index.html'.toString() , map.artifacts.toString())
                Assert.assertTrue(map.allowEmptyArchive)
        })

        jsr.step.executePitTests(parameters)

        Assert.assertTrue(pitCalled)
    }
}

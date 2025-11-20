#!groovy
package steps

import minimatch.Minimatch
import org.junit.After
import util.BasePiperTest

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasKey
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.isEmptyString
import static org.hamcrest.Matchers.not
import static org.hamcrest.Matchers.nullValue
import static org.hamcrest.Matchers.notNullValue

import org.junit.Before
import org.junit.Ignore
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import org.junit.rules.ExpectedException
import static org.junit.Assert.assertThat

import util.Rules
import util.JenkinsLoggingRule
import util.JenkinsStepRule

@Ignore("step disabled")
class PublishTestResultsTest extends BasePiperTest {
    Map publisherStepOptions
    List archiveStepPatterns

    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jlr)
        .around(jsr)

    @Before
    void init() throws Exception {
        publisherStepOptions = [:]
        archiveStepPatterns = []
        // add handler for generic step call
        helper.registerAllowedMethod("step", [Map.class], {
            parameters -> publisherStepOptions[parameters.$class] = parameters
        })
        helper.registerAllowedMethod("perfReport", [Map.class], {
            parameters -> publisherStepOptions['perfReport'] = parameters
        })
        helper.registerAllowedMethod("junit", [Map.class], {
            parameters -> publisherStepOptions['junit'] = parameters
        })
        helper.registerAllowedMethod("jacoco", [Map.class], {
            parameters -> publisherStepOptions['jacoco'] = parameters
        })
        helper.registerAllowedMethod("cobertura", [Map.class], {
            parameters -> publisherStepOptions['cobertura'] = parameters
        })
        helper.registerAllowedMethod("publishHTML", [Map.class], {
            parameters -> publisherStepOptions['publishHTML'] = parameters
        })
        helper.registerAllowedMethod("archiveArtifacts", [Map.class], {
            parameters -> archiveStepPatterns.push(parameters.artifacts)
        })

        helper.registerAllowedMethod("findFiles", [Map.class], { map ->
            println(map.glob)
            return [
                new File("target${File.separator}test.mtar")
            ].toArray()
        })

        nullScript.currentBuild.getRawBuild = {
            return [getAction: { type ->
                return null
            }]
        }
    }

    @After
    void cleanup() {
        nullScript.currentBuild.result = 'SUCCESS'
    }

    @Test
    void testPublishNothing() throws Exception {
        jsr.step.publishTestResults(script: nullScript)
        // asserts
        assertThat(publisherStepOptions.junit, is(nullValue()))
        assertThat(publisherStepOptions.jacoco, is(nullValue()))
        assertThat(publisherStepOptions.cobertura, is(nullValue()))
        assertThat(publisherStepOptions.jmeter, is(nullValue()))
        assertJobStatusSuccess()
    }

    @Test
    void testJunitWithDefaultSettings() throws Exception {
        jsr.step.publishTestResults(script: nullScript, junit: true)
        // asserts
        assertThat(publisherStepOptions.junit, is(notNullValue()))
        assertThat(publisherStepOptions.junit.testResults, is('**/target/surefire-reports/*.xml'))
        assertThat(publisherStepOptions.junit.allowEmptyResults, is(true))
        assertJobStatusSuccess()
    }

    @Test
    void testJunitWithArchive() throws Exception {
        jsr.step.publishTestResults(script: nullScript, junit: [archive: true])
        // asserts
        assertThat(publisherStepOptions.junit, is(notNullValue()))
        assertThat(jlr.log, containsString('[publishTestResults] archive **/target/surefire-reports/*.xml'))
        assertThat(archiveStepPatterns, hasItem('**/target/surefire-reports/*.xml'))
        assertJobStatusSuccess()
    }

    @Test
    void testJunitWithResultUpdate() throws Exception {
        jsr.step.publishTestResults(script: nullScript, junit: [updateResults: true])
        // asserts
        assertThat(publisherStepOptions.junit, is(notNullValue()))
        assertThat(jlr.log, containsString('[publishTestResults] update test results'))
        assertJobStatusSuccess()
    }

    @Test
    void testJunitWithCustomPattern() throws Exception {
        jsr.step.publishTestResults(script: nullScript, junit: [pattern: '**/test.xml'])
        // asserts
        assertThat(publisherStepOptions.junit, is(notNullValue()))
        assertThat(publisherStepOptions.junit.testResults, is('**/test.xml'))
        assertJobStatusSuccess()
    }

    @Test
    void testJunitWithMultipleCustomPattern() throws Exception {
        jsr.step.publishTestResults(script: nullScript, junit: [pattern: '**/target/surefire-reports/*.xml, **/target/failsafe-reports/*.xml'])
        // asserts
        assertThat(publisherStepOptions.junit, is(notNullValue()))
        assertThat(publisherStepOptions.junit.testResults, is('**/target/surefire-reports/*.xml, **/target/failsafe-reports/*.xml'))
        assertJobStatusSuccess()
    }

    @Test
    void testJacocoWithDefaultSettings() throws Exception {
        helper.registerAllowedMethod("findFiles", [Map.class], { map -> return [new File("service/target/jacoco.exec"),new File("service/target/jacoco-it.exec")].toArray()})
        jsr.step.publishTestResults(script: nullScript, jacoco: true)
        // asserts
        assertThat(publisherStepOptions, hasKey('jacoco'))
        assertThat(publisherStepOptions.jacoco, is(notNullValue()))
        assertThat(publisherStepOptions.jacoco.execPattern, is('**/target/*.exec'))
        assertThat(publisherStepOptions.jacoco.inclusionPattern, isEmptyString())
        assertThat(publisherStepOptions.jacoco.exclusionPattern, isEmptyString())
        assertJobStatusSuccess()
    }

    @Test
    void testJacocoWithArchive() throws Exception {
        helper.registerAllowedMethod("findFiles", [Map.class], { map -> return [new File("service/target/jacoco.exec"),new File("service/target/jacoco-it.exec")].toArray()})
        jsr.step.publishTestResults(script: nullScript, jacoco: [archive: true])
        // asserts
        assertThat(publisherStepOptions, hasKey('jacoco'))
        assertThat(publisherStepOptions.jacoco, is(notNullValue()))
        assertThat(jlr.log, containsString('[publishTestResults] archive **/target/*.exec'))
        assertThat(archiveStepPatterns, hasItem('**/target/*.exec'))
        assertJobStatusSuccess()
    }

    @Test
    void testJacocoWithCustomPattern() throws Exception {
        helper.registerAllowedMethod("findFiles", [Map.class], { map -> return [new File("service/target/myReportFile.exec"),new File("service/target/jacoco-it.exec")].toArray()})
        jsr.step.publishTestResults(script: nullScript, jacoco: [archive: true, pattern: 'myReportFile.exec', allowEmptyResults: false])
        // asserts
        assertThat(publisherStepOptions, hasKey('jacoco'))
        assertThat(publisherStepOptions.jacoco, is(notNullValue()))
        assertThat(publisherStepOptions.jacoco.execPattern, is('myReportFile.exec'))
        assertThat(jlr.log, containsString('[publishTestResults] archive myReportFile.exec'))
        assertThat(archiveStepPatterns, hasItem('myReportFile.exec'))
        assertJobStatusSuccess()
    }

    @Test
    void testJacocoWithExclusionsInclusions() throws Exception {
        helper.registerAllowedMethod("findFiles", [Map.class], { map -> return [new File("service/target/jacoco.exec"),new File("service/target/jacoco-it.exec")].toArray()})
        jsr.step.publishTestResults(script: nullScript, jacoco: [exclude: 'exclude.*', include: 'include.*'])
        // asserts
        assertThat(publisherStepOptions, hasKey('jacoco'))
        assertThat(publisherStepOptions.jacoco, is(notNullValue()))
        assertThat(publisherStepOptions.jacoco.inclusionPattern, is('include.*'))
        assertThat(publisherStepOptions.jacoco.exclusionPattern, is('exclude.*'))
        assertJobStatusSuccess()
    }

    @Test
    void testLcovWithArchive() throws Exception {
        helper.registerAllowedMethod("findFiles", [Map.class], { map -> return [new File("target/coverage/lcov-report/index.html")].toArray()})
        jsr.step.publishTestResults(script: nullScript, lcov: [archive: true])
        // asserts
        assertThat(publisherStepOptions, hasKey('publishHTML'))
        assertThat(publisherStepOptions['publishHTML'].reportDir.toString(), is("target${File.separator}coverage${File.separator}lcov-report${File.separator}".toString()))
        assertThat(publisherStepOptions['publishHTML'].reportFiles.toString(), is('index.html'))
        assertThat(publisherStepOptions['publishHTML'].allowMissing, is(true))
        assertThat(jlr.log, containsString('[publishTestResults] archive **/target/coverage/lcov-report/index.html'))
        assertThat(jlr.log, containsString('[publishTestResults] found 1 file(s) to publish for pattern \'**/target/coverage/lcov-report/index.html\''))
        assertJobStatusSuccess()
    }

    @Test
    void testSuptWithDefaults() throws Exception {
        helper.registerAllowedMethod("findFiles", [Map.class], { map -> return [new File("test.csv")].toArray()})
        jsr.step.publishTestResults(script: nullScript, supt: [archive: true])
        // asserts
        assertThat(publisherStepOptions, hasKey('publishHTML'))
        assertThat(publisherStepOptions['publishHTML'].reportDir.toString(), is(''))
        assertThat(publisherStepOptions['publishHTML'].reportFiles.toString(), is('test.csv'))
        assertThat(publisherStepOptions['publishHTML'].allowMissing, is(true))
        assertThat(jlr.log, containsString('[publishTestResults] archive *.csv'))
        assertThat(jlr.log, containsString('[publishTestResults] found 1 file(s) to publish for pattern \'*.csv\''))
        assertJobStatusSuccess()
    }

    @Test
    void testSupaWithDefaults() throws Exception {
        helper.registerAllowedMethod("findFiles", [Map.class], { map -> return [new File("target${File.separator}supa${File.separator}supa_result.html")].toArray()})
        jsr.step.publishTestResults(script: nullScript, supa: [archive: true])
        // asserts
        assertThat(publisherStepOptions, hasKey('publishHTML'))
        assertThat(publisherStepOptions['publishHTML'].reportDir.toString(), is("target${File.separator}supa${File.separator}".toString()))
        assertThat(publisherStepOptions['publishHTML'].reportFiles.toString(), is('supa_result.html'))
        assertThat(jlr.log, containsString('[publishTestResults] archive **/target/supa/supa_result.html'))
        assertThat(jlr.log, containsString('[publishTestResults] found 1 file(s) to publish for pattern \'**/target/supa/supa_result.html\''))
        assertJobStatusSuccess()
    }

    @Test
    void testHtml() throws Exception {
        helper.registerAllowedMethod("findFiles", [Map.class], { map -> return [new File("reports${File.separator}index.html")].toArray()})
        jsr.step.publishTestResults(script: nullScript, html: [archive: true, path: 'reports'])
        // asserts
        assertThat(publisherStepOptions, hasKey('publishHTML'))
        assertThat(publisherStepOptions['publishHTML'].reportDir.toString(), is("reports${File.separator}".toString()))
        assertThat(publisherStepOptions['publishHTML'].reportFiles.toString(), is('index.html'))
        assertThat(jlr.log, containsString('[publishTestResults] archive reports/index.html'))
        assertThat(jlr.log, containsString('[publishTestResults] found 1 file(s) to publish for pattern \'reports/index.html\''))
        assertJobStatusSuccess()
    }



    @Test
    void testPublishNothingWithDefaultSettings() throws Exception {
        jsr.step.publishTestResults(script: nullScript)

        // ensure nothing is published
        assertThat(publisherStepOptions.junit, is(nullValue()))
        assertThat(publisherStepOptions.jacoco, is(nullValue()))
        assertThat(publisherStepOptions.cobertura, is(nullValue()))
        assertThat(publisherStepOptions.jmeter, is(nullValue()))
    }

    @Test
    void testPublishNothingWithAllDisabled() throws Exception {
        jsr.step.publishTestResults(script: nullScript, junit: false, jacoco: false, cobertura: false, jmeter: false)

        // ensure nothing is published
        assertThat(publisherStepOptions.junit, is(nullValue()))
        assertThat(publisherStepOptions.jacoco, is(nullValue()))
        assertThat(publisherStepOptions.cobertura, is(nullValue()))
        assertThat(publisherStepOptions.jmeter, is(nullValue()))
    }

    @Test
    void testPublishUnitTestsWithDefaultSettings() throws Exception {
        jsr.step.publishTestResults(script: nullScript, junit: true)
        // asserts
        assertThat(publisherStepOptions.junit, is(notNullValue()))
        // ensure default patterns are set
        assertThat(publisherStepOptions.junit.testResults, is('**/target/surefire-reports/*.xml'))
        // ensure nothing else is published
        assertThat(publisherStepOptions.jacoco, is(nullValue()))
        assertThat(publisherStepOptions.cobertura, is(nullValue()))
        assertThat(publisherStepOptions.jmeter, is(nullValue()))
    }

    @Test
    void testPublishCoverageWithDefaultSettings() throws Exception {
        jsr.step.publishTestResults(script: nullScript, jacoco: true, cobertura: true)
        // asserts
        assertThat(publisherStepOptions.cobertura, is(notNullValue()))
        assertThat(publisherStepOptions.cobertura.coberturaReportFile, is(notNullValue()))
        String sampleCoberturaPathForJava = 'my/workspace/my/project/target/coverage/cobertura-coverage.xml'
        assertThat(Minimatch.minimatch(sampleCoberturaPathForJava, publisherStepOptions.cobertura.coberturaReportFile), is(true))
        String sampleCoberturaPathForKarma = 'my/workspace/my/project/target/coverage/Chrome 78.0.3904 (Mac OS X 10.14.6)/cobertura-coverage.xml'
        assertThat(Minimatch.minimatch(sampleCoberturaPathForKarma, publisherStepOptions.cobertura.coberturaReportFile), is(true))

        assertThat(publisherStepOptions.jacoco, is(notNullValue()))
        assertThat(publisherStepOptions.jacoco.execPattern, is('**/target/*.exec'))
        // ensure nothing else is published
        assertThat(publisherStepOptions.junit, is(nullValue()))
        assertThat(publisherStepOptions.jmeter, is(nullValue()))
    }

    @Test
    void testPublishJMeterWithDefaultSettings() throws Exception {
        jsr.step.publishTestResults(script: nullScript, jmeter: true)
        // asserts
        assertThat(publisherStepOptions.perfReport, is(notNullValue()))
        assertThat(publisherStepOptions.perfReport.sourceDataFiles, is('**/*.jtl'))
        // ensure nothing else is published
        assertThat(publisherStepOptions.junit, is(nullValue()))
        assertThat(publisherStepOptions.jacoco, is(nullValue()))
        assertThat(publisherStepOptions.cobertura, is(nullValue()))
    }

    @Test
    void testPublishUnitTestsWithCustomSettings() throws Exception {
        jsr.step.publishTestResults(script: nullScript, junit: [pattern: 'fancy/file/path', archive: true, active: true])
        // asserts
        assertThat(publisherStepOptions.junit, is(notNullValue()))
        // ensure that custom patterns are set
        assertThat(publisherStepOptions.junit.testResults, is('fancy/file/path'))
        // ensure nothing else is published
        assertThat(publisherStepOptions.jacoco, is(nullValue()))
        assertThat(publisherStepOptions.cobertura, is(nullValue()))
        assertThat(publisherStepOptions.jmeter, is(nullValue()))
    }

    @Test
    void testBuildResultStatus() throws Exception {
        jsr.step.publishTestResults(script: nullScript)
        assertJobStatusSuccess()
    }

    @Test
    void testBuildWithTestFailuresAndWithoutFailOnError() throws Exception {
        nullScript.currentBuild.getRawBuild = {
            return [getAction: { type ->
                return [getFailCount: {
                    return 6
                }]
            }]
        }

        jsr.step.publishTestResults(script: nullScript, failOnError: false)
        assertJobStatusSuccess()
    }

    @Test
    void testBuildWithTestFailuresAndWithFailOnError() throws Exception {
        nullScript.currentBuild.getRawBuild = {
            return [getAction: { type ->
                return [getFailCount: {
                    return 6
                }]
            }]
        }

        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[publishTestResults] Some tests failed!')

        jsr.step.publishTestResults(script: nullScript, failOnError: true)
    }

    @Test
    void testBuildWithTestFailuresAndWithDefaultFailOnError() throws Exception {
        nullScript.currentBuild.getRawBuild = {
            return [getAction: { type ->
                return [getFailCount: {
                    return 6
                }]
            }]
        }

        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[publishTestResults] Some tests failed!')

        jsr.step.publishTestResults(script: nullScript)
    }
}

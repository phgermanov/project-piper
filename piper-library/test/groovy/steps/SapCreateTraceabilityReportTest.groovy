#!groovy
package steps

import com.sap.piper.internal.JenkinsUtils
import com.sap.icd.jenkins.Utils
import hudson.tasks.junit.TestResult
import util.JenkinsReadFileRule
import util.JenkinsReadJsonRule

import static org.hamcrest.Matchers.*

import org.junit.Before
import org.junit.Test
import org.junit.Rule
import static org.junit.Assert.assertThat
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.Rules

class SapCreateTraceabilityReportTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException()
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsWriteFileRule jwfr = new JenkinsWriteFileRule(this)
    private JenkinsReadJsonRule jrjr = new JenkinsReadJsonRule(this, 'test/resources/traceability/')
    private JenkinsReadFileRule jrfr = new JenkinsReadFileRule(this, 'test/resources/traceability/')

    Map httpRequestMap = [:]


    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jlr)
        .around(jrjr)
        .around(jscr)
        .around(jwfr)
        .around(jsr)
        .around(jrfr)

    @Before
    void init() throws Exception {

        httpRequestMap = [:]
        jscr.setReturnValue('git ls-remote https://github.wdf.sap.corp/ContinuousDelivery/piper-library.git refs/heads/master', '1.4.3')

        binding.setVariable('currentBuild', [
            result: 'SUCCESS',
            rawBuild: mockHelper.loadMockBuild(getSampleTestResult())
        ])

        // register Jenkins commands with mock values
        helper.registerAllowedMethod('archiveArtifacts', [String.class], { s -> null } )

        // register Jenkins commands with mock values
        helper.registerAllowedMethod('getRequirementResultMapping', [Object.class], { o -> return new JenkinsUtils().getMappingWithTestResults( o, getSampleTestResult()) } )

        helper.registerAllowedMethod('string', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            if (l[0].credentialsId == 'terCredentials') {
                binding.setProperty('terToken', 'terTestToken')
                try {
                    c()
                } finally {
                    binding.setProperty('terToken', null)
                }
            }
        })

        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            m.each {entry ->
                if (entry.getValue() instanceof GString)
                    httpRequestMap[entry.getKey()] = entry.getValue().toString()
                else
                    httpRequestMap[entry.getKey()] = entry.getValue()
            }
            if (m.url.contains('validation')) {
                return [status: 200, content: '''{
  "validationPassed": true
}
''']
            }
        })

    }

    private TestResult getSampleTestResult() {

        org.apache.tools.ant.DirectoryScanner ds = new org.apache.tools.ant.DirectoryScanner()
        String[] includes = ["**/resources/traceability/TEST-*.xml"]
        ds.setIncludes(includes)
        ds.setBasedir(".")
        ds.scan()
        def now = new Date()
        def fileList = ds.getIncludedFiles() as List

        fileList.each {fileName ->
            File file = new File(fileName)
            file.setLastModified(now.getTime())
        }

        TestResult testResult = new TestResult()
        testResult.parse(now.getTime(), ds)
        testResult.tally()

        return testResult
    }

    @Test
    void defaultValuesTest() throws Exception {

        jsr.step.sapCreateTraceabilityReport([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            echoDetails: false,
            failOnError: false
        ])

        assertJobStatusSuccess()

        assertThat(jlr.log, containsString('Unstash content: traceabilityMapping'))

        //check that html is generated
        assert jwfr.files['piper_traceability_delivery.html']?.length()>0
        assert jwfr.files['piper_traceability_all.html']?.length()>0

        assert jwfr.files['piper_traceability_delivery.html'].contains('<title>Full Software Requirement Test Report</title>')
        assert jwfr.files['piper_traceability_delivery.html'].contains('<span><i>Program: My sirius program name</i></span>')
        assert jwfr.files['piper_traceability_delivery.html'].contains('<span><i>Delivery: My sirius delivery name</i></span>')

        //check that delivery json is created correctly
        def deliveryJson = new Utils().parseJson(jwfr.files['piper_traceability_delivery.json'])

        assert deliveryJson["JENKINSBCKLG-3"].test_cases.size() == 0
        assert deliveryJson["JENKINSBCKLG-3"].link == "https://jira.tools.sap/browse/JENKINSBCKLG-3"

        assert deliveryJson.size() == 5 
        assert deliveryJson["JENKINSBCKLG-4"].link == "https://jira.tools.sap/browse/JENKINSBCKLG-4"
        assert deliveryJson["JENKINSBCKLG-4"].test_cases.size() == 2
        assert deliveryJson["JENKINSBCKLG-4"].test_cases.contains([
            "test_fullname": "com.sap.suite.cloud.foundation.currencyconversion.CurrencyConversionUnitTest.skeletonTest",
            "test_name": "skeletonTest",
            "test_class": "com.sap.suite.cloud.foundation.currencyconversion.CurrencyConversionUnitTest",
            "passed": true,
            "skipped": false
        ])
        assert deliveryJson['JENKINSBCKLG-4'].test_cases.contains([
            "test_fullname": "com.sap.suite.cloud.foundation.currencyconversion.CurrencyCodeListServiceTest.notExistingTest()",
            "test_name": "",
            "test_class": "",
            "passed": false,
            "skipped": true
        ])
        assert deliveryJson["JENKINSBCKLG-7"].link == 'https://jira.tools.sap/browse/JENKINSBCKLG-7'
        assertThat(deliveryJson["JENKINSBCKLG-7"].test_cases, hasSize(3))
        assert deliveryJson["JENKINSBCKLG-7"].test_cases.contains([
            "test_fullname": "com.sap.suite.cloud.foundation.createautomatedprocedures.CreateAutomatedProceduresActivateIT.procedureCanBeActivated{CreationOrigin}[1]",
            "test_name": "procedureCanBeActivated{CreationOrigin}[1]",
            "test_class": "com.sap.suite.cloud.foundation.createautomatedprocedures.CreateAutomatedProceduresActivateIT",
            "passed": true,
            "skipped": false
        ])
        assert deliveryJson["JENKINSBCKLG-7"].test_cases.contains([
            "test_fullname": "com.sap.suite.cloud.foundation.createautomatedprocedures.CreateAutomatedProceduresActivateIT.procedureCanBeActivated{CreationOrigin}[2]",
            "test_name": "procedureCanBeActivated{CreationOrigin}[2]",
            "test_class": "com.sap.suite.cloud.foundation.createautomatedprocedures.CreateAutomatedProceduresActivateIT",
            "passed": true,
            "skipped": false
        ])
        assert deliveryJson["JENKINSBCKLG-7"].test_cases.contains([
            "test_fullname": "com.sap.suite.cloud.foundation.createautomatedprocedures.CreateAutomatedProceduresActivateIT.procedureCanBeActivated{CreationOrigin}[3]",
            "test_name": "procedureCanBeActivated{CreationOrigin}[3]",
            "test_class": "com.sap.suite.cloud.foundation.createautomatedprocedures.CreateAutomatedProceduresActivateIT",
            "passed": true,
            "skipped": false
        ])    
        assert deliveryJson['org/repo#5'].link == 'https://github.wdf.sap.corp/org/repo/issues/5'
        assert deliveryJson['org/repo#5'].test_cases.size() == 1
        assert deliveryJson['org/repo#5'].test_cases.contains([
            "test_fullname": "com.sap.suite.cloud.foundation.currencyconversion.CurrencyCodeListServiceTest.notExistingTest()",
            "test_name": "",
            "test_class": "",
            "passed": false,
            "skipped": true
        ])

        //check that full json is created correctly
        def fullJson = new Utils().parseJson(jwfr.files['piper_traceability_all.json'])

        assert fullJson.size() == 11
        assert fullJson['JENKINSBCKLG-1'].link == 'https://jira.tools.sap/browse/JENKINSBCKLG-1'
        assertThat(fullJson["JENKINSBCKLG-1"].test_cases, hasSize(3))
        //assert fullJson['JENKINSBCKLG-1'].test_cases.size() == 3
        assert fullJson['JENKINSBCKLG-1'].test_cases.contains([
            "test_fullname": "com.sap.suite.cloud.foundation.currencyconversion.CurrencyConversionUnitTest.convert_currency",
            "test_name": "convert_currency",
            "test_class": "com.sap.suite.cloud.foundation.currencyconversion.CurrencyConversionUnitTest",
            "passed": true,
            "skipped": false
        ])
        assert fullJson['JENKINSBCKLG-5'].test_cases.contains([
            "test_fullname": "com.sap.suite.cloud.foundation.currencyconversion.CurrencyCodeListServiceTest.notExistingTest()",
            "test_name": "",
            "test_class": "",
            "passed": false,
            "skipped": true
        ])
        assert fullJson['org/repo#3'].link == 'https://github.wdf.sap.corp/org/repo/issues/3'
        assert fullJson['org/repo#3'].test_cases.size() == 3
        assert fullJson['org/repo#3'].test_cases.contains([
            "test_fullname": "com.sap.suite.cloud.foundation.currencyconversion.CurrencyConversionUnitTest.convert_currency",
            "test_name": "convert_currency",
            "test_class": "com.sap.suite.cloud.foundation.currencyconversion.CurrencyConversionUnitTest",
            "passed": true,
            "skipped": false
        ])
        assert fullJson["JENKINSBCKLG-7"].link == 'https://jira.tools.sap/browse/JENKINSBCKLG-7'
        assertThat(fullJson["JENKINSBCKLG-7"].test_cases, hasSize(3))
        assert fullJson["JENKINSBCKLG-7"].test_cases.contains([
            "test_fullname": "com.sap.suite.cloud.foundation.createautomatedprocedures.CreateAutomatedProceduresActivateIT.procedureCanBeActivated{CreationOrigin}[1]",
            "test_name": "procedureCanBeActivated{CreationOrigin}[1]",
            "test_class": "com.sap.suite.cloud.foundation.createautomatedprocedures.CreateAutomatedProceduresActivateIT",
            "passed": true,
            "skipped": false
        ])
        assert fullJson["JENKINSBCKLG-7"].test_cases.contains([
            "test_fullname": "com.sap.suite.cloud.foundation.createautomatedprocedures.CreateAutomatedProceduresActivateIT.procedureCanBeActivated{CreationOrigin}[2]",
            "test_name": "procedureCanBeActivated{CreationOrigin}[2]",
            "test_class": "com.sap.suite.cloud.foundation.createautomatedprocedures.CreateAutomatedProceduresActivateIT",
            "passed": true,
            "skipped": false
        ])
        assert fullJson["JENKINSBCKLG-7"].test_cases.contains([
            "test_fullname": "com.sap.suite.cloud.foundation.createautomatedprocedures.CreateAutomatedProceduresActivateIT.procedureCanBeActivated{CreationOrigin}[3]",
            "test_name": "procedureCanBeActivated{CreationOrigin}[3]",
            "test_class": "com.sap.suite.cloud.foundation.createautomatedprocedures.CreateAutomatedProceduresActivateIT",
            "passed": true,
            "skipped": false
        ])
    }

    @Test
    void jiraOnlyTest() throws Exception {

        jsr.step.sapCreateTraceabilityReport([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            echoDetails: false,
            failOnError: false,
            requirementMappingFile: 'jira_requirement.mapping',
            deliveryMappingFile: 'jira_delivery.mapping'
        ])

        assertJobStatusSuccess()

    }

    @Test
    void githubOnlyTest() throws Exception {

        jsr.step.sapCreateTraceabilityReport([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            echoDetails: false,
            failOnError: false,
            requirementMappingFile: 'github_requirement.mapping',
            deliveryMappingFile: 'github_delivery.mapping'
        ])

        assertJobStatusSuccess()

    }

    @Test
    void untestedRequirementTest() {

        thrown.expectMessage('[sapCreateTraceabilityReport] only 1 of 5 requirements fulfilled')

        jsr.step.sapCreateTraceabilityReport([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            echoDetails: false,
            failOnError: true,
            requirementMappingFile: 'requirement.mapping',
            deliveryMappingFile: 'delivery_error.mapping'
        ])

        assertJobStatusSuccess()

    }

    @Test
    void terServiceTestDefault() {

        jsr.step.sapCreateTraceabilityReport([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            failOnError: false,
            terUpload: true,
            terTokenCredentialsId: 'terCredentials',
            verbose: false
        ])

        assertThat(httpRequestMap.url, is('https://ter.tools.sap.corp/v1/api/traceability/fc2/update/twb'))
        assertThat(httpRequestMap.contentType, is('APPLICATION_JSON'))
        assertThat(httpRequestMap.httpMode, is('POST'))
        assertThat(httpRequestMap.customHeaders, hasItem([name: 'Authorization', value: 'terTestToken']))
        assertThat(httpRequestMap.consoleLogResponseBody, is(false))
        assertThat(httpRequestMap.responseHandle, is('NONE'))
        assertThat(httpRequestMap.requestBody, allOf(
            containsString('"siriusProgramName": "My sirius program name"'),
            containsString('"deliveryName": "My sirius delivery name"'),
            containsString('"sourceId": "p"'),
            containsString('"fc2Content":'),
            containsString('"JENKINSBCKLG-3":'),
            containsString('"JENKINSBCKLG-4":'),
            containsString('"org/repo#3":'),
            not(containsString('"testSet":'))
        ))
    }

    @Test
    void terServiceTestCustom() {

        jsr.step.sapCreateTraceabilityReport([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            failOnError: false,
            terUpload: true,
            terServerUrl: 'https://ter-qa.tools.sap.corp',
            terTokenCredentialsId: 'terCredentials',
            verbose: true
        ])

        assertThat(httpRequestMap.url, is('https://ter-qa.tools.sap.corp/v1/api/traceability/fc2/update/twb'))
        assertThat(httpRequestMap.consoleLogResponseBody, is(true))
    }

    @Test
    void testTestSetAvailable() {
        jsr.step.sapCreateTraceabilityReport([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
            failOnError: false,
            terUpload: true,
            terTestSet: 'myTestSet',
            terTokenCredentialsId: 'terCredentials',
            verbose: false
        ])

        assertThat(httpRequestMap.requestBody, containsString('"testSet": "myTestSet"'))
    }

    @Test
    void terServiceValidationFailureTest() {

        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            m.each {entry ->
                if (entry.getValue() instanceof GString)
                    httpRequestMap[entry.getKey()] = entry.getValue().toString()
                else
                    httpRequestMap[entry.getKey()] = entry.getValue()
            }
            if (m.url.contains('validation')) {
                return [status: 200, content: '''{
  "validationPassed": false,
  "message": "test message",
  "errors": [
    "error 1"
  ]
}
''']
            }
        })

        boolean errorOccured = false
        try {
            jsr.step.sapCreateTraceabilityReport([
                script: nullScript,
                juStabUtils: utils,
                jenkinsUtilsStub: jenkinsUtils,
                terUpload: true,
                terTokenCredentialsId: 'terCredentials',
                verbose: true
            ])
        } catch (err) {
            assertThat(httpRequestMap.url, is('https://ter.tools.sap.corp/v1/api/validation/fc2input'))
            assertThat(httpRequestMap.contentType, is('APPLICATION_JSON'))
            assertThat(httpRequestMap.httpMode, is('POST'))
            assertThat(httpRequestMap.customHeaders, hasItem([name: 'Authorization', value: 'terTestToken']))
            assertThat(httpRequestMap.requestBody, not(containsString('JENKINSBCKLG-1')))

            assertThat(jlr.log, containsString('validation failed'))
            errorOccured = true
        }

        assertThat(errorOccured, is(true))
    }

    @Test
    void terServiceValidationFailureTestAll() {

        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            m.each {entry ->
                if (entry.getValue() instanceof GString)
                    httpRequestMap[entry.getKey()] = entry.getValue().toString()
                else
                    httpRequestMap[entry.getKey()] = entry.getValue()
            }
            if (m.url.contains('validation')) {
                return [status: 200, content: '''{
  "validationPassed": false,
  "message": "test message",
  "errors": [
    "error 1"
  ]
}
''']
            }
        })

        boolean errorOccured = false
        try {
            jsr.step.sapCreateTraceabilityReport([
                script: nullScript,
                juStabUtils: utils,
                jenkinsUtilsStub: jenkinsUtils,
                terUpload: true,
                terTokenCredentialsId: 'terCredentials',
                checkAllRequirements: true
            ])
        } catch (err) {
            assertThat(httpRequestMap.requestBody, containsString('JENKINSBCKLG-1'))
            errorOccured = true
        }
        assertThat(errorOccured, is(true))
    }


    // https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues/1789
    @Test
    void testNoTestResultsAvailable() {
        helper.registerAllowedMethod('getRequirementResultMapping', [Object.class], { o ->
            throw new java.lang.NullPointerException('Cannot invoke method getResult() on null object')
        })
        thrown.expectMessage(containsString('Failed to retrieve test results'))
        thrown.expectMessage(containsString('Cannot invoke method getResult() on null object'))

        jsr.step.sapCreateTraceabilityReport([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtils,
        ])

    }
}

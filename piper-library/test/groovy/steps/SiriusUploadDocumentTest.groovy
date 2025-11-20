#!groovy
package steps

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsReadJsonRule
import util.JenkinsStepRule
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue

class SiriusUploadDocumentTest extends BasePiperTest {

    private static class Sirius {
        def programName = ''
        def deliveryName = ''

        def deliveryGuid = ''
        def siriusTaskGuid = ''
        def fileName = ''
        def documentFamily = ''

        def getDeliveryExtGuidByName(programName, deliveryName) {
            this.programName = programName
            this.deliveryName = deliveryName
            return 'deliveryTestGUID'
        }
        def uploadDocument(deliveryGuid, siriusTaskGuid, fileName, documentName, documentFamily) {
            this.deliveryGuid =  deliveryGuid
            this.siriusTaskGuid = siriusTaskGuid
            this.fileName = fileName
            this.documentFamily = documentFamily
        }
    }

    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsReadJsonRule jrjr = new JenkinsReadJsonRule(this, 'test/resources/traceability/')

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jrjr)
        .around(jsr) // needs to be activated after jedr, otherwise executeDocker is not mocked

    @Before
    void init() {
        helper.registerAllowedMethod('fileExists', [String.class], {s -> return false})
    }

    @Test
    void testSiriusUploadDocumentBackwardCompatibility() {

        Sirius sirius = new Sirius()

        jsr.step.siriusUploadDocument([
            script: nullScript,
            juStabUtils: utils,
            siriusStub: sirius,
            credentialsId: 'TestCredentials',
            fileName: 'file_name',
            siriusDeliveryName: 'sirius_delivery_name',
            siriusTaskGuid: 'sirius_task_guid',
            siriusProgramName: 'sirius_program_name',
            apiUrl: 'https://api.test'
        ])
        assertEquals('sirius_program_name',sirius.programName)
        assertEquals('sirius_delivery_name', sirius.deliveryName)

        assertEquals('deliveryTestGUID', sirius.deliveryGuid)
        assertEquals('sirius_task_guid', sirius.siriusTaskGuid)
        assertEquals('file_name', sirius.fileName)
        assertEquals('TEST', sirius.documentFamily)

        assertTrue(jlr.log.contains('Using credentialsId: \'TestCredentials\', apiUrl: https://api.test'))
    }

    @Test
    void testSiriusUploadDocument() {

        Sirius sirius = new Sirius()

        jsr.step.siriusUploadDocument([
            script: nullScript,
            juStabUtils: utils,
            siriusStub: sirius,
            siriusCredentialsId: 'TestCredentials',
            fileName: 'file_name',
            siriusDeliveryName: 'sirius_delivery_name',
            siriusTaskGuid: 'sirius_task_guid',
            siriusProgramName: 'sirius_program_name',
        ])
        assertEquals('sirius_program_name',sirius.programName)
        assertEquals('sirius_delivery_name', sirius.deliveryName)

        assertEquals('deliveryTestGUID', sirius.deliveryGuid)
        assertEquals('sirius_task_guid', sirius.siriusTaskGuid)
        assertEquals('file_name', sirius.fileName)
        assertEquals('TEST', sirius.documentFamily)

        assertTrue(jlr.log.contains('Using credentialsId: \'TestCredentials\', apiUrl: https://ifp.bss.net.sap/zprs/api/v1'))
    }

    @Test
    void testWithDeliveryMapping() {
        helper.registerAllowedMethod('fileExists', [String.class], {s -> return true})
        Sirius sirius = new Sirius()

        jsr.step.siriusUploadDocument([
            script: nullScript,
            juStabUtils: utils,
            siriusStub: sirius,
            siriusCredentialsId: 'TestCredentials',
            fileName: 'file_name',
            siriusTaskGuid: 'sirius_task_guid',
        ])
        assertThat(sirius.programName, is('My sirius program name'))
        assertThat(sirius.deliveryName, is('My sirius delivery name'))
    }

    @Test
    void testWithDeliveryMappingCustomFile() {
        helper.registerAllowedMethod('fileExists', [String.class], {s -> return true})
        Sirius sirius = new Sirius()
        nullScript.globalPipelineEnvironment.configuration = [steps: [sapCreateTraceabilityReport: [deliveryMappingFile: 'delivery_error.mapping']]]

        jsr.step.siriusUploadDocument([
            script: nullScript,
            juStabUtils: utils,
            siriusStub: sirius,
            siriusCredentialsId: 'TestCredentials',
            fileName: 'file_name',
            siriusTaskGuid: 'sirius_task_guid',
        ])
        assertThat(sirius.programName, is('My sirius program name 2'))
        assertThat(sirius.deliveryName, is('My sirius delivery name 2'))
    }
}

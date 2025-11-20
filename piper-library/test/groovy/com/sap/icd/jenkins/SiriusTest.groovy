#!groovy
package com.sap.icd.jenkins

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import util.BasePiperTest
import util.JenkinsShellCallRule

import static org.junit.Assert.assertEquals
import org.junit.rules.RuleChain

import util.Rules

class SiriusTest extends BasePiperTest {

    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jscr)

    def httpRequestMap = [:]

    @Before
    void init() {
        nullScript.metaClass.httpRequest = {Map m ->
            httpRequestMap = m
            def response = mockHelper.createResponse('{"content":[]}')
            return response
        }

    }

    @Test
    void testSiriusBasic() throws Exception {
        // load Sirius
        Sirius sirius = new Sirius(nullScript, 'test', 'https://ifp.wdf.sap.corp/zprs/api/v1','https://ifp.wdf.sap.corp/zprs/api')
        // test default getApiUrl
        assertEquals(sirius.getApiUrl(), 'https://ifp.wdf.sap.corp/zprs/api/v1')
        // test setApiUrl
        sirius.setApiUrl('https://test/zprs/api/v1')
        assertEquals(sirius.getApiUrl(), 'https://test/zprs/api/v1')
        assertEquals(sirius.getUploadUrl(), 'https://ifp.wdf.sap.corp/zprs/api')
        // test setUploadUrl
        sirius.setUploadUrl('https://test/zprs/api')
        assertEquals(sirius.getUploadUrl(), 'https://test/zprs/api')
    }

    @Test
    void testSiriusGetProgramIntGuid() {
        helper.registerAllowedMethod('readJSON', [Map.class], { m -> return [[PROGRAM_NAME: 'Program1', PROGRAM_INT_GUID: 'Program1GUID'], [PROGRAM_NAME: 'Program2', PROGRAM_INT_GUID: 'Program2GUID']]} )
        Sirius sirius = new Sirius(nullScript, 'testId', 'https://ifp.wdf.sap.corp/zprs/api/v1','https://ifp.wdf.sap.corp/zprs/api')
        assertEquals('Program2GUID', sirius.getProgramIntGuid('Program2'))
        assertEquals ('GET', httpRequestMap.httpMode)
        assertEquals ('testId', httpRequestMap.authentication)
        assertEquals('https://ifp.wdf.sap.corp/zprs/api/v1/program', httpRequestMap.url.toString())

    }

    @Test
    void testGetDeliveryExtGuidByName() {
        helper.registerAllowedMethod('readJSON', [Map.class], { m -> return [["DELIVERY_EXT_GUID":"TestGUID"]] } )
        Sirius sirius = new Sirius(nullScript, 'testId', 'https://ifp.wdf.sap.corp/zprs/api/v1','https://ifp.wdf.sap.corp/zprs/api')
        assertEquals('TestGUID', sirius.getDeliveryExtGuidByName('TestProgram', 'TestDelivery'))
        assertEquals ('GET', httpRequestMap.httpMode)
        assertEquals ('testId', httpRequestMap.authentication)
        assertEquals('https://ifp.wdf.sap.corp/zprs/api/v1/delivery?programGuid=null&name=TestDelivery', httpRequestMap.url.toString())
    }

    @Test
    void testUploadDocument() {
        def fileName = ''
        helper.registerAllowedMethod('readFile', [Map.class], { m ->
            fileName = m.file
            return 'FileContent' } )
        Sirius sirius = new Sirius(nullScript, 'testId', 'https://ifp.wdf.sap.corp/zprs/api/v1','https://ifp.wdf.sap.corp/zprs/api/r2d2/saveDocumentInRDTask')
        sirius.uploadDocument('deliveryExtGuid', 'taskGuid', 'fileName.txt', 'documentName', 'documentFamily')
        assertEquals('base64 --wrap=0 fileName.txt >fileName.b64', jscr.shell[0])
        assertEquals('fileName.b64', fileName.toString())
        assertEquals ('POST', httpRequestMap.httpMode)
        assertEquals ('''{
"deliveryGuid": "deliveryExtGuid",
"originalPmtGuid": "taskGuid",
"fileName": "fileName.txt",
"fileContent": "FileContent",
"documentName": "documentName",
"documentFamily": "documentFamily",
"confidential": ""
}''', httpRequestMap.requestBody.toString())
        assertEquals('https://ifp.wdf.sap.corp/zprs/api/r2d2/saveDocumentInRDTask', httpRequestMap.url.toString())
    }

    @Test
    void testUploadDocumentConfidential() {
        def fileName = ''
        helper.registerAllowedMethod('readFile', [Map.class], { m ->
            fileName = m.file
            return 'FileContent' } )
        Sirius sirius = new Sirius(nullScript, 'testId', 'https://ifp.wdf.sap.corp/zprs/api/v1','https://ifp.wdf.sap.corp/zprs/api/r2d2/saveDocumentInRDTask')
        sirius.uploadDocument('deliveryExtGuid', 'taskGuid', 'fileName.txt', 'documentName', 'documentFamily', true)
        assertEquals ('''{
"deliveryGuid": "deliveryExtGuid",
"originalPmtGuid": "taskGuid",
"fileName": "fileName.txt",
"fileContent": "FileContent",
"documentName": "documentName",
"documentFamily": "documentFamily",
"confidential": "X"
}''', httpRequestMap.requestBody.toString())
    }

}

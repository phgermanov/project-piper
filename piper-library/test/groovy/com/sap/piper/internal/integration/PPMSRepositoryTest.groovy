package com.sap.piper.internal.integration

import hudson.AbortException
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.JenkinsEnvironmentRule
import util.JenkinsLoggingRule
import util.JenkinsReadJsonRule
import util.Rules
import util.BasePiperTest

import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.instanceOf
import static org.hamcrest.Matchers.nullValue
import static org.hamcrest.Matchers.startsWith
import static org.junit.Assert.assertThat
import static org.hamcrest.Matchers.hasKey
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not

class PPMSRepositoryTest extends BasePiperTest {

    private ExpectedException exception = ExpectedException.none()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsReadJsonRule jrjr = new JenkinsReadJsonRule(this)
    private JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(exception)
        .around(jlr)
        .around(jrjr)
        .around(jer)

    PPMSRepository repository
    Map config

    @Before
    void init() throws Exception {
        config = [
            ppmsBuildVersionEndpoint: '/bvEndpoint',
            ppmsChangeRequestEndpoint: '/crEndpoint',
            ppmsCredentialsId: 'ppmsCredentialsId',
            ppmsID: '101',
            ppmsServerUrl: 'https://test.server'
        ]

        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            if (m.url.contains('BuildVersions')) {
                return [content: '{}']
            }
        })

        //default PPMS repository object to use in tests
        repository = new PPMSRepository(nullScript, utils, config)
    }

    @Test
    void testMapFossIDToFossObject() {

        def mappingResponse = [
            [
                fossnr    : 'test',
                moreParams: 'testParams'
            ],
            [
                fossnr    : 'test2',
                moreParams: 'testParams'
            ]
        ]

        assertThat(repository.mapFossIDToFossObject(mappingResponse), allOf(hasKey("test"), hasKey("test2")))
    }

    @Test
    void testResolveFittingChannelDescription() {

        def ppmsResponse = [
            riskRating: [
                gtmcName: [
                    [
                        gtmcId  : 'GMTC_1',
                        gtmcName: 'test'
                    ],
                    [
                        gtmcId  : 'GMTC_2',
                        gtmcName: 'ABAP OO'
                    ],
                    [
                        gtmcId  : 'GMTC_10',
                        gtmcName: 'Correct Name Cloud'
                    ]
                ]
            ]
        ]

        def result = repository.resolveMatchingChannelDescription(ppmsResponse, 'GMTC_10')

        assertThat(result, is('Correct Name Cloud'))
    }

    //-------------- NEW TESTS ------------

    def scvContent = '''{
"d": {
"Id": "101",
"Name": "TEST SCV 1.0",
"TechnicalName": "TEST_SCV",
"TechnicalRelease": "1"
}
}
'''

    def pvContent = '''{
"header": {
"pvname": "TEST PV 1.0"
}
}'''

    def bvContent = '''{
"d": {
"results": [
{
"__metadata": {
"id": "https://test.server/odataint/borm/odataforosrcy/BuildVersions('1001')",
"uri": "https://test.server/odataint/borm/odataforosrcy/BuildVersions('1001')",
"type": "/BORM/ODATA_FOR_OSRCY_SRV.BuildVersion"
},
"Id": "1001",
"Name": "BV 1",
"Description": "Description BV 1",
"SoftwareComponentVersionId": "101",
"SoftwareComponentVersionsName": "TEST SCV 1.0",
"SortSequence": "0000000001",
"ReviewModelRiskRatings": {
"__deferred": {
"uri": "https://test.server/odataint/borm/odataforosrcy/BuildVersions('1001')/ReviewModelRiskRatings"
}
},
"FreeOpenSourceSoftwares": {
"__deferred": {
"uri": "https://test.server/odataint/borm/odataforosrcy/BuildVersions('1001')/FreeOpenSourceSoftwares"
}
}
},
{
"__metadata": {
"id": "https://test.server/odataint/borm/odataforosrcy/BuildVersions('1002')",
"uri": "https://test.server/odataint/borm/odataforosrcy/BuildVersions('1002')",
"type": "/BORM/ODATA_FOR_OSRCY_SRV.BuildVersion"
},
"Id": "1002",
"Name": "BV 2",
"Description": "Description BV 2",
"SoftwareComponentVersionId": "101",
"SoftwareComponentVersionsName": "TEST SCV 1.0",
"SortSequence": "0000000002",
"ReviewModelRiskRatings": {
"__deferred": {
"uri": "https://test.server/odataint/borm/odataforosrcy/BuildVersions('1002')/ReviewModelRiskRatings"
}
},
"FreeOpenSourceSoftwares": {
"__deferred": {
"uri": "https://test.server/odataint/borm/odataforosrcy/BuildVersions('1002')/FreeOpenSourceSoftwares"
}
}
},
{
"__metadata": {
"id": "https://test.server/odataint/borm/odataforosrcy/BuildVersions('1003')",
"uri": "https://test.server/odataint/borm/odataforosrcy/BuildVersions('1003')",
"type": "/BORM/ODATA_FOR_OSRCY_SRV.BuildVersion"
},
"Id": "1003",
"Name": "BV 3",
"Description": "Description BV 3",
"SoftwareComponentVersionId": "101",
"SoftwareComponentVersionsName": "TEST SCV 1.0",
"SortSequence": "0000000003",
"ReviewModelRiskRatings": {
"__deferred": {
"uri": "https://test.server/odataint/borm/odataforosrcy/BuildVersions('1003')/ReviewModelRiskRatings"
}
},
"FreeOpenSourceSoftwares": {
"__deferred": {
"uri": "https://test.server/odataint/borm/odataforosrcy/BuildVersions('1003')/FreeOpenSourceSoftwares"
}
}
}
]
}
}
'''

    def crStatus = '''{
    "id": "2222",
    "status": "PENDING",
    "statusText": "Not yet applied",
    "createdAt": "2000-01-01T00:00:00Z",
    "createdBy": "testUser",
    "processedAt": null,
    "processedBy": null,
    "links": [
        {
            "rel": "self",
            "href": "https://i7d.wdf.sap.corp/sap/internal/ppms/api/changerequest/v1/cvpart/2222"
        },
        {
            "rel": "document",
            "href": "https://i7d.wdf.sap.corp/sap/internal/ppms/api/changerequest/v1/cvpart/2222/document"
        },
        {
            "rel": "response",
            "href": "https://i7d.wdf.sap.corp/sap/internal/ppms/api/changerequest/v1/cvpart/2222/response"
        }
    ]
}
'''

    @Test
    void testFetchResourceDetails_scv() {
        def httpRequestMap
        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            httpRequestMap = m
            return [content: scvContent]
        })

        config.ppmsID = '101'
        PPMSRepository ppmsRepository = new PPMSRepository(nullScript, utils, config)
        def scvDetails = ppmsRepository.fetchResourceDetails()

        assertThat(scvDetails.overviewLink, is('https://test.server/ppmslight/#/details/cv/101/overview'))
        assertThat(scvDetails.name, is('TEST SCV 1.0'))
        assertThat(scvDetails.technicalName, is('TEST_SCV'))
        assertThat(scvDetails.technicalRelease, is('1'))
    }

    @Test
    void testFetchResourceDetails_pv() {
        def httpRequestMap
        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            httpRequestMap = m
            if (m.url.contains('BuildVersions')) {
                return [content: '{}']
            } else if (m.url.contains('SoftwareComponentVersions')) {
                throw new AbortException('Error occurred')
            }
            return [content: pvContent]
        })

        config.ppmsID = '201'
        PPMSRepository ppmsRepository = new PPMSRepository(nullScript, utils, config)
        def pvDetails = ppmsRepository.fetchResourceDetails()

        assertThat(pvDetails.overviewLink, is('https://test.server/ppmslight/#/details/pv/201/overview'))
        assertThat(pvDetails.name, is('TEST PV 1.0'))
    }

    @Test
    void testFetchResourceDetails_v2_notFound() {
        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            if (m.url.contains('BuildVersions')) {
                return [content: '{}']
            }
            throw new AbortException('Error occurred')
        })
        exception.expectMessage('Illegal PPMS ID, cannot determine whether is component version or product version')


        config.ppmsID = '301'
        PPMSRepository ppmsRepository = new PPMSRepository(nullScript, utils, config)
        ppmsRepository.fetchResourceDetails()

    }

    @Test
    void testGetScvDetails() {
        def httpRequestMap
        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            httpRequestMap = m
            return [content: scvContent]
        })

        PPMSRepository ppmsRepository = new PPMSRepository( nullScript, utils, config)
        def scvDetails = ppmsRepository.getScvDetails()

        assertThat(httpRequestMap.url.toString(), is("https://test.server/odataint/borm/odataforosrcy/SoftwareComponentVersions('101')?\$format=json"))

        assertThat(scvDetails.overviewLink, is('https://test.server/ppmslight/#/details/cv/101/overview'))
        assertThat(scvDetails.name, is('TEST SCV 1.0'))
        assertThat(scvDetails.technicalName, is('TEST_SCV'))
        assertThat(scvDetails.technicalRelease, is('1'))

    }

    @Test
    void testGetScvBuildVersionId() {
        def httpRequestMap
        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            httpRequestMap = m
            return [content: bvContent]
        })

        PPMSRepository ppmsRepository = new PPMSRepository( nullScript, utils, config)
        def bvId = ppmsRepository.getScvBuildVersionId('BV 2')

        assertThat(bvId, is('1002'))
        assertThat(httpRequestMap.url.toString(), is("https://test.server/odataint/borm/odataforosrcy/SoftwareComponentVersions('101')/BuildVersions?\$format=json"))
    }

    @Test
    void testGetScvBuildVersionIdNotFound() {
        def httpRequestMap
        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            httpRequestMap = m
            return [content: bvContent]
        })

        PPMSRepository ppmsRepository = new PPMSRepository( nullScript, utils, config)
        def bvId = ppmsRepository.getScvBuildVersionId('BV 4')

        assertThat(bvId, nullValue())
    }

    @Test
    void testGetScvBuildVersionIdNoBV() {
        def httpRequestMap
        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            httpRequestMap = m
            return [content: '''{
"d": {
"results": []
}
}
''']
        })

        PPMSRepository ppmsRepository = new PPMSRepository( nullScript, utils, config)
        def bvId = ppmsRepository.getScvBuildVersionId('BV 2')

        assertThat(bvId, nullValue())
    }

    @Test
    void testGetLatestScvBuildVersionId() {
        def httpRequestMap
        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            httpRequestMap = m
            return [content: bvContent]
        })

        PPMSRepository ppmsRepository = new PPMSRepository( nullScript, utils, config)
        def bvId = ppmsRepository.getLatestScvBuildVersionId()

        assertThat(bvId, is('1003'))
    }

    @Test
    void testGetLatestScvBuildVersionIdNoBV() {
        def httpRequestMap
        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            httpRequestMap = m
            return [content: '''{
"d": {
"results": []
}
}
''']
        })

        PPMSRepository ppmsRepository = new PPMSRepository( nullScript, utils, config)
        def bvId = ppmsRepository.getLatestScvBuildVersionId()

        assertThat(bvId, nullValue())
    }

    @Test
    void testChangeRequestDocumentV1() {
        def testDoc = repository.getChangeRequestDocumentV1('testUser', 'TestSource', 'scvTestName')

        assertThat(testDoc.schemaVer, is('1-0-0'))
        assertThat(testDoc.changeRequestData.id, not(nullValue()))
        assertThat(testDoc.changeRequestData.scan.source, is('TestSource'))
        assertThat(testDoc.changeRequestData.scan.tool, is('WhiteSource'))
        assertThat(testDoc.changeRequestData.comparison.tool, is('Piper'))
        assertThat(testDoc.changeRequestData.comparison.reviewedBy, is('testUser'))
        assertThat(testDoc.changeRequestData.target.softwareComponentVersionName, is('scvTestName'))
        assertThat(testDoc.changeRequestData.target.softwareComponentVersionNumber, is('101'))
        assertThat(testDoc.changeRequestData.sendConfirmationTo[0].userId, is('testUser'))
        assertThat(testDoc.changeRequestData.target.buildVersionName, nullValue())
        assertThat(testDoc.changeRequestData.target.buildVersionNumber, nullValue())
    }

    @Test
    void testCreateChangeRequestDocumentV1TypeHandling() {

        PPMSRepository ppmsRepository = new PPMSRepository(nullScript, utils, config)

        def testReport = ppmsRepository.getChangeRequestDocumentV1('testUser', 'TestSource', 'scvTestName')
        assertThat(testReport.changeRequestData.target.softwareComponentVersionNumber, instanceOf(String.class))
    }

    @Test
    void testGetChangeRequestDocumentV1BV() {

        def httpRequestMap
        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            httpRequestMap = m
            return [content: bvContent]
        })

        config.ppmsBuildVersion = 'BV 1'
        PPMSRepository ppmsRepository = new PPMSRepository(nullScript, utils, config)

        def testDoc = ppmsRepository.getChangeRequestDocumentV1('testUser', 'TestSource')
        assertThat(testDoc.changeRequestData.target.buildVersionNumber, is('1001'))
    }

    @Test
    void testGetChangeRequestDocumentV2() {
        PPMSRepository ppmsRepository = new PPMSRepository(nullScript, utils, config)

        def testReport = ppmsRepository.getChangeRequestDocumentV2('testUser', 'TestSource', 'scvTestName')
        assertThat(testReport.schemaVer, is('2-0-0'))
        assertThat(testReport.changeRequestData.provider.reviewedBy, is('testUser'))
        assertThat(testReport.changeRequestData.target.softwareComponentVersionNumber, is('101'))
        assertThat(testReport.changeRequestData.sendConfirmationTo, is([[userId: 'testUser']]))
    }

    @Test
    void testGetChangeRequestDocumentV2BV() {

        def httpRequestMap
        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            httpRequestMap = m
            return [content: bvContent]
        })

        config.ppmsBuildVersion = 'BV 1'

        PPMSRepository ppmsRepository = new PPMSRepository( nullScript, utils, config)
        def testReport = ppmsRepository.getChangeRequestDocumentV2('testUser', 'TestSource')

        assertThat(testReport.changeRequestData.target.buildVersionNumber, is('1001'))
    }

    @Test
    void testCreateChangeRequest() {

        def httpRequestMap = []
        helper.registerAllowedMethod("httpRequest", [Map.class], {m ->
            m.url = m.url.toString()
            def myStatus = 500
            def myContent = '{}'
            if (m.customHeaders && m.customHeaders[0] == [maskValue: true, name: 'x-csrf-token', value: 'Fetch'] && m.httpMode == 'HEAD') {
                myStatus = 200
            }

            if (m.customHeaders && m.customHeaders[0] == [maskValue: true, name: 'x-csrf-token', value: 'testToken'] && m.httpMode == 'POST') {
                myStatus = 201
                myContent = '{"crId": "1000"}'
            }

            httpRequestMap.add(m)
            //check for xscrf header and return value
            return new Object(){
                def content = myContent
                def status = myStatus
                def getHeaders() {
                    def headers = [
                        'cache-control': ['No-Cache'],
                        'set-cookie': [
                            'cookie1=piper-client=001; path=/',
                            'SESSION=piperSession; path=/; secure; HttpOnly',
                        ]
                    ]
                    if (m.customHeaders[0] == [maskValue: true, name: 'x-csrf-token', value: 'Fetch']) {
                        headers.'x-csrf-token' = ['testToken']
                    }
                    return headers
                }
            }
        })

        config.ppmsID = '301'

        PPMSRepository ppmsRepository = new PPMSRepository(nullScript, utils, config)

        def crId = ppmsRepository.createChangeRequest('myCDDoc')

        assertThat(httpRequestMap[0], allOf(
            hasEntry('url','https://test.server/crEndpoint'),
            hasEntry('httpMode', 'HEAD'),
            hasEntry('authentication', 'ppmsCredentialsId'))
        )
        assertThat(httpRequestMap[1], allOf(
            hasEntry('url','https://test.server/crEndpoint'),
            hasEntry('httpMode', 'POST'),
            hasEntry('authentication', 'ppmsCredentialsId'),
            hasEntry('requestBody', 'myCDDoc'))
        )

        assertThat(httpRequestMap[1].customHeaders[1], is([name: 'Cookie', value: 'cookie1=piper-client=001; SESSION=piperSession']))

        assertThat(crId, is('1000'))
    }

    @Test
    void testGetChangeRequestHeaderInfoAndStatus() {

        def httpRequestMap
        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            httpRequestMap = m
            return [content: crStatus]
        })

        def crHeaderInfo = repository.getChangeRequestHeaderInfo('2222')

        assertThat(httpRequestMap.url.toString(), is('https://test.server/crEndpoint/2222'))
        assertThat(crHeaderInfo.id, is('2222'))
        assertThat(crHeaderInfo.status, is('PENDING'))

        def crStatus = repository.getChangeRequestStatus('2222')
        assertThat(crStatus.id, is('PENDING'))
        assertThat(crStatus.description, is('Not yet applied'))

    }

    @Test
    void testGetBuildVersionCreationDocument() {
        def httpRequestMap
        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            httpRequestMap = m
            return [content: bvContent]
        })

        def bvDoc = repository.getBuildVersionCreationDocument('testUser', 'BVName', 'BVDescription')

        assertThat(bvDoc.changeRequestData.id, not(nullValue()))
        assertThat(bvDoc.changeRequestData.provider.reviewedBy, is('testUser'))
        assertThat(bvDoc.changeRequestData.target.softwareComponentVersionNumber, is('101'))
        assertThat(bvDoc.changeRequestData.buildVersion.predecessorBuildVersionNumber, is('1003'))
        assertThat(bvDoc.changeRequestData.options.copyPredecessorFoss, is(false))
        assertThat(bvDoc.changeRequestData.options.copyPredecessorCvBv, is(false))

        assertThat(bvDoc.changeRequestData.buildVersion.name, is('BVName'))
        assertThat(bvDoc.changeRequestData.buildVersion.description, is('BVDescription'))
    }

    @Test
    void testGetBuildVersionCreationDocumentBV() {
        def httpRequestMap
        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            httpRequestMap = m
            return [content: bvContent]
        })

        PPMSRepository ppmsRepository = new PPMSRepository(nullScript, utils, config)
        def bvDoc = ppmsRepository.getBuildVersionCreationDocument('testUser', 'BVName', 'BVDescription')

        assertThat(bvDoc.changeRequestData.buildVersion.predecessorBuildVersionNumber, is('1003'))
    }

    @Test
    void testCreateBuildVersion() {

        def httpRequestMap = []
        helper.registerAllowedMethod("httpRequest", [Map.class], {m ->
            m.url = m.url.toString()
            def myStatus = 500
            def myContent = '{}'
            if (m.customHeaders && m.customHeaders[0] == [maskValue: true, name: 'x-csrf-token', value: 'Fetch'] && m.httpMode == 'HEAD') {
                myStatus = 200
            }

            if (m.customHeaders && m.customHeaders[0] == [maskValue: true, name: 'x-csrf-token', value: 'testToken'] && m.httpMode == 'POST') {
                myStatus = 201
                myContent = '{"crId": "1000"}'
            }

            /*if (m.url.contains('BuildVersions')) {
                myStatus = 200
                myContent = '{}'
            }*/

            httpRequestMap.add(m)
            //check for xscrf header and return value
            return new Object(){
                def content = myContent
                def status = myStatus
                def getHeaders() {
                    def headers = [
                        'cache-control': ['No-Cache'],
                        'set-cookie': [
                            'cookie1=piper-client=001; path=/',
                            'SESSION=piperSession; path=/; secure; HttpOnly',
                        ]
                    ]
                    if (m.customHeaders[0] == [maskValue: true, name: 'x-csrf-token', value: 'Fetch']) {
                        headers.'x-csrf-token' = ['testToken']
                    }
                    return headers
                }
            }
        })

        PPMSRepository ppmsRepository = new PPMSRepository(nullScript, utils, config)

        def crId = ppmsRepository.createBuildVersion('myBVDoc')

        assertThat(httpRequestMap[0], allOf(
            hasEntry('url','https://test.server/bvEndpoint'),
            hasEntry('httpMode', 'HEAD'),
            hasEntry('authentication', 'ppmsCredentialsId'))
        )
        assertThat(httpRequestMap[1], allOf(
            hasEntry('url','https://test.server/bvEndpoint'),
            hasEntry('httpMode', 'POST'),
            hasEntry('authentication', 'ppmsCredentialsId'),
            hasEntry('requestBody', 'myBVDoc'))
        )

        assertThat(httpRequestMap[1].customHeaders[1], is([name: 'Cookie', value: 'cookie1=piper-client=001; SESSION=piperSession']))

        assertThat(crId, is('1000'))
    }

    @Test
    void testGetBuildVersionCreationHeaderInfoAndStatus() {

        def httpRequestMap
        helper.registerAllowedMethod('httpRequest', [Map.class], {m ->
            httpRequestMap = m
            return [content: crStatus]
        })

        def crHeaderInfo = repository.getBuildVersionCreationHeaderInfo('2222')

        assertThat(httpRequestMap.url.toString(), is('https://test.server/bvEndpoint/2222'))
        assertThat(crHeaderInfo.id, is('2222'))
        assertThat(crHeaderInfo.status, is('PENDING'))

        def crStatus = repository.getBuildVersionCreationStatus('2222')
        assertThat(crStatus.id, is('PENDING'))
        assertThat(crStatus.description, is('Not yet applied'))

    }

}

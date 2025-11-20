package com.sap.piper.internal.integration

import hudson.AbortException
import org.junit.Assert
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import org.springframework.beans.factory.annotation.Autowired
import util.BasePiperTest
import util.JenkinsEnvironmentRule
import util.JenkinsLoggingRule
import util.Rules

import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.is
import static org.hamcrest.CoreMatchers.isA


class FortifyTest extends BasePiperTest {

    private ExpectedException exception = ExpectedException.none()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(exception)
        .around(jlr)
        .around(jer)


    @Autowired
    Fortify fortify

    @Test
    void testGetFortifyVersion() throws IOException {

        List<LinkedHashMap> fetchedVersions = new ArrayList<>()

        LinkedHashMap<Object, Object> searchedEntry = new LinkedHashMap<>()

        searchedEntry.put("name", "1.0")
        searchedEntry.put("id", "123-Fortify")

        LinkedHashMap<Object, Object> obsoleteEntry = new LinkedHashMap<>()

        obsoleteEntry.put("name", "0.9")
        obsoleteEntry.put("id", "000-Fortify")

        fetchedVersions.add(obsoleteEntry)
        fetchedVersions.add(searchedEntry)

        def result = fortify.getFortifyVersionIDFor(nullScript, "1.0", fetchedVersions)

        Assert.assertEquals("123-Fortify", result.toString())
    }

    @Test
    void testlookupOrCreateProjectVersionIDForPR() {

        def getVersionsCalled, getVersionDetailsCalled, createVersionCalled, getAttributesCalled, setAttributesCalled, copyFromPartialCalled, commitCalled, copyCurrentStateCalled, getAuthEntitiesCalled, setAuthEntitiesCalled = false

        helper.registerAllowedMethod('httpRequest', [Map], {
            m ->
                if (m.url.endsWith("/projects/4711/versions")) {
                    getVersionsCalled = true
                    return [content: "{\"data\":[], \"responseCode\": 200}"]
                }
                if (m.url.endsWith("/projectVersions/10172")) {
                    getVersionDetailsCalled = true
                    return [content: "{\"data\":{\"name\": \"master\", \"description\": \"master desc\", \"active\": true, \"committed\": true, \"project\": {\"name\": \"master\", \"description\": \"master desc\", \"id\": 4711, \"issueTemplateId\": 18}, \"issueTemplateId\": 18}, \"responseCode\": 200}"]
                }
                if (m.url.endsWith("/projectVersions")) {
                    createVersionCalled = true
                    return [content: "{\"data\":{\"name\": \"PR-78\", \"id\": 10173, \"description\": \"master desc\", \"active\": true, \"committed\": true, \"project\": {\"name\": \"master\", \"description\": \"master desc\", \"id\": 4711, \"issueTemplateId\": 18}, \"issueTemplateId\": 18}, \"responseCode\": 200}"]
                }
                if (m.url.endsWith("/projectVersions/10172/attributes")) {
                    getAttributesCalled = true
                    return [content: "{\"data\": [{\"_href\": \"https://fortify.mo.sap.corp/ssc/api/v1/projectVersions/10172/attributes/4712\",\"attributeDefinitionId\": 31, \"values\": null,\"guid\": \"gdgfdgfdgfdgfd\",\"id\": 4712,\"value\": \"abcd\"}],\"count\": 8,\"responseCode\": 200}"]
                }
                if (m.url.endsWith("/projectVersions/10173/attributes")) {
                    setAttributesCalled = true
                    return [content: "{\"data\": [{\"_href\": \"https://fortify.mo.sap.corp/ssc/api/v1/projectVersions/10173/attributes/4713\",\"attributeDefinitionId\": 31, \"values\": null," +
                        "\"guid\": \"gdgfdgfdgfdgfd\",\"id\": 4713,\"value\": \"abcd\"}],\"count\": 8,\"responseCode\": 200}"]
                }
                if (m.url.endsWith("/projectVersions/10173/action") && m.requestBody.contains('COPY_FROM_PARTIAL')) {
                    copyFromPartialCalled = true
                    return [content: "{\"data\":[{\"latestScanId\":null,\"serverVersion\":17.2,\"tracesOutOfDate\":false,\"attachmentsOutOfDate\":false,\"description\":\"\",\n" +
                        "\t\t\t\t\"project\":{\"id\":4711,\"name\":\"python-test\",\"description\":\"\",\"creationDate\":\"2018-12-03T06:29:38.197+0000\",\"createdBy\":\"someUser\",\n" +
                        "\t\t\t\t\"issueTemplateId\":\"dasdasdasdsadasdasdasdasdas\"},\"sourceBasePath\":null,\"mode\":\"BASIC\",\"masterAttrGuid\":\"sddasdasda\",\"obfuscatedId\":null,\n" +
                        "\t\t\t\t\"id\":10172,\"customTagValuesAutoApply\":null,\"issueTemplateId\":\"dasdasdasdsadasdasdasdasdas\",\"loadProperties\":null,\"predictionPolicy\":null,\n" +
                        "\t\t\t\t\"bugTrackerPluginId\":null,\"owner\":\"admin\",\"_href\":\"https://fortify.mo.sap.corp/ssc/api/v1/projectVersions/10172\",\n" +
                        "\t\t\t\t\"committed\":true,\"bugTrackerEnabled\":false,\"active\":true,\"snapshotOutOfDate\":false,\"issueTemplateModifiedTime\":1578411924701,\n" +
                        "\t\t\t\t\"securityGroup\":null,\"creationDate\":\"2018-02-09T16:59:41.297+0000\",\"refreshRequired\":false,\"issueTemplateName\":\"someTemplate\",\n" +
                        "\t\t\t\t\"migrationVersion\":null,\"createdBy\":\"admin\",\"name\":\"0\",\"siteId\":null,\"staleIssueTemplate\":false,\"autoPredict\":null,\n" +
                        "\t\t\t\t\"currentState\":{\"id\":10172,\"committed\":true,\"attentionRequired\":false,\"analysisResultsExist\":true,\"auditEnabled\":true,\n" +
                        "\t\t\t\t\"lastFprUploadDate\":\"2018-02-09T16:59:53.497+0000\",\"extraMessage\":null,\"analysisUploadEnabled\":true,\"batchBugSubmissionExists\":false,\n" +
                        "\t\t\t\t\"hasCustomIssues\":false,\"metricEvaluationDate\":\"2018-03-10T00:02:45.553+0000\",\"deltaPeriod\":7,\"issueCountDelta\":0,\"percentAuditedDelta\":0.0,\n" +
                        "\t\t\t\t\"criticalPriorityIssueCountDelta\":0,\"percentCriticalPriorityIssuesAuditedDelta\":0.0},\"assignedIssuesCount\":0,\"status\":null}],\n" +
                        "\t\t\t\t\"count\":1,\"responseCode\":200,\"links\":{\"last\":{\"href\":\"https://fortify.mo.sap.corp/ssc/api/v1/projects/4711/versions?start=0\"},\n" +
                        "\t\t\t\t\"first\":{\"href\":\"https://fortify.mo.sap.corp/ssc/api/v1/projects/4711/versions?start=0\"}}}"]
                }
                if (m.url.endsWith("/projectVersions/10173?hideProgress=true")) {
                    commitCalled = true
                    return [content: "{\"data\": {\"active\": true, \"bugTrackerEnabled\": false, \"bugTrackerPluginId\": null, \"committed\": true, \"createdBy\": null, \"creationDate\": null," +
                        "\"description\": null, \"issueTemplateId\": null, \"issueTemplateModifiedTime\": null, \"issueTemplateName\": null, \"latestScanId\": null, \"masterAttrGuid\": null," +
                        "\"name\": \"PR-78\", \"owner\": null, \"serverVersion\": 19.2, \"snapshotOutOfDate\": null, \"staleIssueTemplate\": null}, \"responseCode\": 200}"]
                }
                if (m.url.endsWith("/projectVersions/10173/action") && m.requestBody.contains('COPY_CURRENT_STATE')) {
                    copyCurrentStateCalled = true
                    return [content: "{\"data\":[{\"latestScanId\":null,\"serverVersion\":17.2,\"tracesOutOfDate\":false,\"attachmentsOutOfDate\":false,\"description\":\"\",\n" +
                        "\t\t\t\t\"project\":{\"id\":4711,\"name\":\"python-test\",\"description\":\"\",\"creationDate\":\"2018-12-03T06:29:38.197+0000\",\"createdBy\":\"someUser\",\n" +
                        "\t\t\t\t\"issueTemplateId\":\"dasdasdasdsadasdasdasdasdas\"},\"sourceBasePath\":null,\"mode\":\"BASIC\",\"masterAttrGuid\":\"sddasdasda\",\"obfuscatedId\":null,\n" +
                        "\t\t\t\t\"id\":10172,\"customTagValuesAutoApply\":null,\"issueTemplateId\":\"dasdasdasdsadasdasdasdasdas\",\"loadProperties\":null,\"predictionPolicy\":null,\n" +
                        "\t\t\t\t\"bugTrackerPluginId\":null,\"owner\":\"admin\",\"_href\":\"https://fortify.mo.sap.corp/ssc/api/v1/projectVersions/10172\",\n" +
                        "\t\t\t\t\"committed\":true,\"bugTrackerEnabled\":false,\"active\":true,\"snapshotOutOfDate\":false,\"issueTemplateModifiedTime\":1578411924701,\n" +
                        "\t\t\t\t\"securityGroup\":null,\"creationDate\":\"2018-02-09T16:59:41.297+0000\",\"refreshRequired\":false,\"issueTemplateName\":\"someTemplate\",\n" +
                        "\t\t\t\t\"migrationVersion\":null,\"createdBy\":\"admin\",\"name\":\"0\",\"siteId\":null,\"staleIssueTemplate\":false,\"autoPredict\":null,\n" +
                        "\t\t\t\t\"currentState\":{\"id\":10172,\"committed\":true,\"attentionRequired\":false,\"analysisResultsExist\":true,\"auditEnabled\":true,\n" +
                        "\t\t\t\t\"lastFprUploadDate\":\"2018-02-09T16:59:53.497+0000\",\"extraMessage\":null,\"analysisUploadEnabled\":true,\"batchBugSubmissionExists\":false,\n" +
                        "\t\t\t\t\"hasCustomIssues\":false,\"metricEvaluationDate\":\"2018-03-10T00:02:45.553+0000\",\"deltaPeriod\":7,\"issueCountDelta\":0,\"percentAuditedDelta\":0.0,\n" +
                        "\t\t\t\t\"criticalPriorityIssueCountDelta\":0,\"percentCriticalPriorityIssuesAuditedDelta\":0.0},\"assignedIssuesCount\":0,\"status\":null}],\n" +
                        "\t\t\t\t\"count\":1,\"responseCode\":200,\"links\":{\"last\":{\"href\":\"https://fortify.mo.sap.corp/ssc/api/v1/projects/4711/versions?start=0\"},\n" +
                        "\t\t\t\t\"first\":{\"href\":\"https://fortify.mo.sap.corp/ssc/api/v1/projects/4711/versions?start=0\"}}}"]
                }
                if(m.url.endsWith("/projectVersions/10172/authEntities?embed=roles")) {
                    getAuthEntitiesCalled = true
                    return [content: "{\"data\": [{\"displayName\":\"some user\",\"email\":\"some.one@test.com\",\"entityName\":\"some_user\",\"firstName\":\"some\",\"id\":589," +
                        "\"lastName\":\"user\",\"type\":\"User\"}], \"responseCode\": 200}"]
                }
                if(m.url.endsWith("/projectVersions/10173/authEntities")) {
                    setAuthEntitiesCalled = true
                    return [content: "{\"data\": [{\"displayName\":\"some user\",\"email\":\"some.one@test.com\",\"entityName\":\"some_user\",\"firstName\":\"some\",\"id\":589," +
                        "\"lastName\":\"user\",\"type\":\"User\"}], \"responseCode\": 200}"]
                }
                Assert.fail("Unexpected HTTP request ${m.url}")
        })

        def result = fortify.lookupOrCreateProjectVersionIDForPR(nullScript, "4711", "10172", "PR-78")
        Assert.assertEquals("10173", result.toString())
        Assert.assertThat(true, allOf(is(getVersionsCalled) , is(getVersionDetailsCalled), is(createVersionCalled), is(getAttributesCalled), is(setAttributesCalled),
            is(copyFromPartialCalled), is(commitCalled), is(copyCurrentStateCalled), is(getAuthEntitiesCalled), is(setAuthEntitiesCalled)))
    }

    @Test
    void testGetFortifyVersionExpectNotFound() throws Throwable {

        List<LinkedHashMap> fetchedVersions = new ArrayList<>()

        exception.expect(isA(AbortException.class))

        fortify.getFortifyVersionIDFor(nullScript, "1.0", fetchedVersions)
    }

    @Test
    void testHttpFilterSets() throws Throwable {
        helper.registerAllowedMethod("httpRequest", [Map.class], {
            def result = [:]
            result.content = "{\"data\":[{\"defaultFilterSet\":true,\"folders\":[{\"id\":19709,\"guid\":\"b968f72f-1810-03b5-976e-ad4c13920c21\",\"name\":\"Corporate Security Requirements\",\"color\":\"000000\"},{\"id\":19710,\"guid\":\"5b50bb77-1810-08ed-fdba-1213fa90ac5a\",\"name\":\"Audit All\",\"color\":\"ff0000\"},{\"id\":19711,\"guid\":\"d5f55910-1810-a775-e91f-191d1f5608a4\",\"name\":\"Spot Checks of Each Category\",\"color\":\"ff8000\"},{\"id\":19712,\"guid\":\"a36ce828-1810-49cb-b351-7051609e26af\",\"name\":\"Optional\",\"color\":\"808080\"}],\"description\":\"\",\"guid\":\"a243b195-0a59-3f8b-1403-d55b7a7d78e6\",\"title\":\"SAP\"}],\"count\":1,\"responseCode\":200,\"links\":{\"last\":{\"href\":\"https://fortify.mo.sap.corp/ssc/api/v1/projectVersions/13772/filterSets?start=0\"},\"first\":{\"href\":\"https://fortify.mo.sap.corp/ssc/api/v1/projectVersions/13772/filterSets?start=0\"}}}"
            return result
        })

        def filterSets = fortify.httpFilterSets(nullScript, '7635675')
        Assert.assertNotNull(filterSets.corporateRequirements)
        Assert.assertEquals("b968f72f-1810-03b5-976e-ad4c13920c21", filterSets.corporateRequirements.id)
        Assert.assertEquals("Audit All", filterSets.auditAll.name)
    }

    @Test
    void testHttpGroupings() throws Throwable {
        helper.registerAllowedMethod("httpRequest", [Map.class], {
            def result = [:]
            result.content = "{\"data\":{\"groupBySet\":[{\"entityType\":\"CUSTOMTAG\",\"guid\":\"87f2364f-dcd4-49e6-861d-f8d3f351686b\",\"displayName\":\"Analysis\"," +
                "\"value\":\"87f2364f-dcd4-49e6-861d-f8d3f351686b\",\"description\":\"\"},{\"entityType\":\"CUSTOMTAG\",\"guid\":\"0aafa386-766b-484e-90c7-4b9397998dfc\",\"displayName\":\"Analysis State\"," +
                "\"value\":\"0aafa386-766b-484e-90c7-4b9397998dfc\",\"description\":\"\"},{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111151\",\"displayName\":\"Analysis Type\"," +
                "\"value\":\"11111111-1111-1111-1111-111111111151\",\"description\":\"\"},{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111141\",\"displayName\":\"Analyzer\"," +
                "\"value\":\"11111111-1111-1111-1111-111111111141\",\"description\":\"\"},{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111176\",\"displayName\":\"App Defender Protected\"," +
                "\"value\":\"11111111-1111-1111-1111-111111111176\",\"description\":\"\"},{\"entityType\":\"EXTERNALLIST\",\"guid\":\"3ADB9EE4-5761-4289-8BD3-CBFCC593EBBC\",\"displayName\":\"CWE\"," +
                "\"value\":\"3ADB9EE4-5761-4289-8BD3-CBFCC593EBBC\",\"description\":\"\"},{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111165\",\"displayName\":\"Category\"," +
                "\"value\":\"11111111-1111-1111-1111-111111111165\",\"description\":\"\"},{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111209\",\"displayName\":\"Correlated\"," +
                "\"value\":\"11111111-1111-1111-1111-111111111209\",\"description\":\"\"},{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111208\",\"displayName\":\"Correlation Group\"," +
                "\"value\":\"11111111-1111-1111-1111-111111111208\",\"description\":\"\"},{\"entityType\":\"EXTERNALLIST\",\"guid\":\"B40F9EE0-3824-4879-B9FE-7A789C89307C\",\"displayName\":\"FISMA\"," +
                "\"value\":\"B40F9EE0-3824-4879-B9FE-7A789C89307C\",\"description\":\"\"},{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111132\",\"displayName\":\"File Name\"," +
                "\"value\":\"11111111-1111-1111-1111-111111111132\",\"description\":\"\"},{\"entityType\":\"ISSUE\",\"guid\":\"FOLDER\",\"displayName\":\"Folder\",\"value\":\"FOLDER\",\"description\":\"\"}," +
                "{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111150\",\"displayName\":\"Fortify Priority Order\",\"value\":\"11111111-1111-1111-1111-111111111150\",\"description\":\"\"}," +
                "{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111602\",\"displayName\":\"Introduced date\",\"value\":\"11111111-1111-1111-1111-111111111602\",\"description\":\"\"}," +
                "{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111125\",\"displayName\":\"Issue State\",\"value\":\"11111111-1111-1111-1111-111111111125\",\"description\":\"\"}," +
                "{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111136\",\"displayName\":\"Kingdom\",\"value\":\"11111111-1111-1111-1111-111111111136\",\"description\":\"\"}," +
                "{\"entityType\":\"EXTERNALLIST\",\"guid\":\"555A3A66-A0E1-47AF-910C-3F19A6FB2506\",\"displayName\":\"MISRA C 2012\",\"value\":\"555A3A66-A0E1-47AF-910C-3F19A6FB2506\",\"description\":\"\"}," +
                "{\"entityType\":\"EXTERNALLIST\",\"guid\":\"5D4B75A1-FC91-4B4B-BD4D-C81BBE9604FA\",\"displayName\":\"MISRA C++ 2008\",\"value\":\"5D4B75A1-FC91-4B4B-BD4D-C81BBE9604FA\",\"description\":\"\"}," +
                "{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111170\",\"displayName\":\"Manual\",\"value\":\"11111111-1111-1111-1111-111111111170\",\"description\":\"\"}," +
                "{\"entityType\":\"EXTERNALLIST\",\"guid\":\"1114583B-EA24-45BE-B7F8-B61201BACDD0\",\"displayName\":\"NIST SP 800-53 Rev.4\",\"value\":\"1114583B-EA24-45BE-B7F8-B61201BACDD0\"," +
                "\"description\":\"\"},{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111167\",\"displayName\":\"New Issue\",\"value\":\"11111111-1111-1111-1111-111111111167\"," +
                "\"description\":\"\"},{\"entityType\":\"EXTERNALLIST\",\"guid\":\"EEE3F9E7-28D6-4456-8761-3DA56C36F4EE\",\"displayName\":\"OWASP Mobile 2014\",\"value\":\"EEE3F9E7-28D6-4456-8761-3DA56C36F4EE\"," +
                "\"description\":\"\"},{\"entityType\":\"EXTERNALLIST\",\"guid\":\"771C470C-9274-4580-8556-C023E4D3ADB4\",\"displayName\":\"OWASP Top 10 2004\",\"value\":\"771C470C-9274-4580-8556-C023E4D3ADB4\"," +
                "\"description\":\"\"},{\"entityType\":\"EXTERNALLIST\",\"guid\":\"1EB1EC0E-74E6-49A0-BCE5-E6603802987A\",\"displayName\":\"OWASP Top 10 2007\",\"value\":\"1EB1EC0E-74E6-49A0-BCE5-E6603802987A\"," +
                "\"description\":\"\"},{\"entityType\":\"EXTERNALLIST\",\"guid\":\"FDCECA5E-C2A8-4BE8-BB26-76A8ECD0ED59\",\"displayName\":\"OWASP Top 10 2010\",\"value\":\"FDCECA5E-C2A8-4BE8-BB26-76A8ECD0ED59\"," +
                "\"description\":\"\"},{\"entityType\":\"EXTERNALLIST\",\"guid\":\"1A2B4C7E-93B0-4502-878A-9BE40D2A25C4\",\"displayName\":\"OWASP Top 10 2013\",\"value\":\"1A2B4C7E-93B0-4502-878A-9BE40D2A25C4\"," +
                "\"description\":\"\"},{\"entityType\":\"EXTERNALLIST\",\"guid\":\"CBDB9D4D-FC20-4C04-AD58-575901CAB531\",\"displayName\":\"PCI 1.1\",\"value\":\"CBDB9D4D-FC20-4C04-AD58-575901CAB531\"," +
                "\"description\":\"\"},{\"entityType\":\"EXTERNALLIST\",\"guid\":\"57940BDB-99F0-48BF-BF2E-CFC42BA035E5\",\"displayName\":\"PCI 1.2\",\"value\":\"57940BDB-99F0-48BF-BF2E-CFC42BA035E5\"," +
                "\"description\":\"\"},{\"entityType\":\"EXTERNALLIST\",\"guid\":\"8970556D-7F9F-4EA7-8033-9DF39D68FF3E\",\"displayName\":\"PCI 2.0\",\"value\":\"8970556D-7F9F-4EA7-8033-9DF39D68FF3E\"," +
                "\"description\":\"\"},{\"entityType\":\"EXTERNALLIST\",\"guid\":\"E2FB0D38-0192-4F03-8E01-FE2A12680CA3\",\"displayName\":\"PCI 3.0\",\"value\":\"E2FB0D38-0192-4F03-8E01-FE2A12680CA3\"," +
                "\"description\":\"\"},{\"entityType\":\"EXTERNALLIST\",\"guid\":\"AC0D18CF-C1DA-47CF-9F1A-E8EC0A4A717E\",\"displayName\":\"PCI 3.1\",\"value\":\"AC0D18CF-C1DA-47CF-9F1A-E8EC0A4A717E\"," +
                "\"description\":\"\"},{\"entityType\":\"EXTERNALLIST\",\"guid\":\"4E8431F9-1BA1-41A8-BDBD-087D5826751A\",\"displayName\":\"PCI 3.2\",\"value\":\"4E8431F9-1BA1-41A8-BDBD-087D5826751A\"," +
                "\"description\":\"\"},{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111144\",\"displayName\":\"Package\",\"value\":\"11111111-1111-1111-1111-111111111144\",\"description\":\"\"}," +
                "{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111164\",\"displayName\":\"Primary Context\",\"value\":\"11111111-1111-1111-1111-111111111164\",\"description\":\"\"}," +
                "{\"entityType\":\"EXTERNALLIST\",\"guid\":\"939EF193-507A-44E2-ABB7-C00B2168B6D8\",\"displayName\":\"SANS Top 25 2009\",\"value\":\"939EF193-507A-44E2-ABB7-C00B2168B6D8\",\"description\":\"\"}," +
                "{\"entityType\":\"EXTERNALLIST\",\"guid\":\"72688795-4F7B-484C-88A6-D4757A6121CA\",\"displayName\":\"SANS Top 25 2010\",\"value\":\"72688795-4F7B-484C-88A6-D4757A6121CA\",\"description\":\"\"}," +
                "{\"entityType\":\"EXTERNALLIST\",\"guid\":\"92EB4481-1FD9-4165-8E16-F2DE6CB0BD63\",\"displayName\":\"SANS Top 25 2011\",\"value\":\"92EB4481-1FD9-4165-8E16-F2DE6CB0BD63\",\"description\":\"\"}," +
                "{\"entityType\":\"EXTERNALLIST\",\"guid\":\"F2FA57EA-5AAA-4DDE-90A5-480BE65CE7E7\",\"displayName\":\"STIG 3.1\",\"value\":\"F2FA57EA-5AAA-4DDE-90A5-480BE65CE7E7\",\"description\":\"\"}," +
                "{\"entityType\":\"EXTERNALLIST\",\"guid\":\"788A87FE-C9F9-4533-9095-0379A9B35B12\",\"displayName\":\"STIG 3.10\",\"value\":\"788A87FE-C9F9-4533-9095-0379A9B35B12\",\"description\":\"\"}," +
                "{\"entityType\":\"EXTERNALLIST\",\"guid\":\"58E2C21D-C70F-4314-8994-B859E24CF855\",\"displayName\":\"STIG 3.4\",\"value\":\"58E2C21D-C70F-4314-8994-B859E24CF855\",\"description\":\"\"}," +
                "{\"entityType\":\"EXTERNALLIST\",\"guid\":\"DD18E81F-3507-41FA-9DFA-2A9A15B5479F\",\"displayName\":\"STIG 3.5\",\"value\":\"DD18E81F-3507-41FA-9DFA-2A9A15B5479F\",\"description\":\"\"}," +
                "{\"entityType\":\"EXTERNALLIST\",\"guid\":\"000CA760-0FED-4374-8AA2-6FA3968A07B1\",\"displayName\":\"STIG 3.6\",\"value\":\"000CA760-0FED-4374-8AA2-6FA3968A07B1\",\"description\":\"\"}," +
                "{\"entityType\":\"EXTERNALLIST\",\"guid\":\"E69C07C0-81D8-4B04-9233-F3E74167C3D2\",\"displayName\":\"STIG 3.7\",\"value\":\"E69C07C0-81D8-4B04-9233-F3E74167C3D2\",\"description\":\"\"}," +
                "{\"entityType\":\"EXTERNALLIST\",\"guid\":\"1A9D736B-2D4A-49D1-88CA-DF464B40D732\",\"displayName\":\"STIG 3.9\",\"value\":\"1A9D736B-2D4A-49D1-88CA-DF464B40D732\",\"description\":\"\"}," +
                "{\"entityType\":\"EXTERNALLIST\",\"guid\":\"95227C50-A9E4-4C9D-A8AF-FD98ABAE1F3C\",\"displayName\":\"STIG 4.1\",\"value\":\"95227C50-A9E4-4C9D-A8AF-FD98ABAE1F3C\",\"description\":\"\"}," +
                "{\"entityType\":\"EXTERNALLIST\",\"guid\":\"672C15F8-8822-4E05-8C9E-1A4BAAA7A373\",\"displayName\":\"STIG 4.2\",\"value\":\"672C15F8-8822-4E05-8C9E-1A4BAAA7A373\",\"description\":\"\"}," +
                "{\"entityType\":\"EXTERNALLIST\",\"guid\":\"A0B313F0-29BD-430B-9E34-6D10F1178506\",\"displayName\":\"STIG 4.3\",\"value\":\"A0B313F0-29BD-430B-9E34-6D10F1178506\",\"description\":\"\"}," +
                "{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111163\",\"displayName\":\"Sink\",\"value\":\"11111111-1111-1111-1111-111111111163\",\"description\":\"\"}," +
                "{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111161\",\"displayName\":\"Source\",\"value\":\"11111111-1111-1111-1111-111111111161\",\"description\":\"\"}," +
                "{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111162\",\"displayName\":\"Source Context\",\"value\":\"11111111-1111-1111-1111-111111111162\",\"description\":\"\"}," +
                "{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111166\",\"displayName\":\"Source File\",\"value\":\"11111111-1111-1111-1111-111111111166\",\"description\":\"\"}," +
                "{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111123\",\"displayName\":\"Status\",\"value\":\"11111111-1111-1111-1111-111111111123\",\"description\":\"\"}," +
                "{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111143\",\"displayName\":\"Taint Flag\",\"value\":\"11111111-1111-1111-1111-111111111143\",\"description\":\"\"}," +
                "{\"entityType\":\"ISSUE\",\"guid\":\"11111111-1111-1111-1111-111111111168\",\"displayName\":\"URL\",\"value\":\"11111111-1111-1111-1111-111111111168\",\"description\":\"\"}," +
                "{\"entityType\":\"EXTERNALLIST\",\"guid\":\"74f8081d-dd49-49da-880f-6830cebe9777\",\"displayName\":\"WASC 2.00\",\"value\":\"74f8081d-dd49-49da-880f-6830cebe9777\"," +
                "\"description\":\"\"},{\"entityType\":\"EXTERNALLIST\",\"guid\":\"9DC61E7F-1A48-4711-BBFD-E9DFF537871F\",\"displayName\":\"WASC 24 + 2\",\"value\":\"9DC61E7F-1A48-4711-BBFD-E9DFF537871F\"," +
                "\"description\":\"\"}],\"filterBySet\":[{\"entityType\":\"FOLDER\",\"filterSelectorType\":\"LIST\",\"guid\":\"userAssignment\",\"displayName\":\"Folder\",\"value\":\"FOLDER\"," +
                "\"description\":\"\",\"selectorOptions\":[{\"guid\":\"b968f72f-1810-03b5-976e-ad4c13920c21\",\"displayName\":\"Corporate Security Requirements\"," +
                "\"value\":\"b968f72f-1810-03b5-976e-ad4c13920c21\"},{\"guid\":\"5b50bb77-1810-08ed-fdba-1213fa90ac5a\",\"displayName\":\"Audit All\",\"value\":\"5b50bb77-1810-08ed-fdba-1213fa90ac5a\"}," +
                "{\"guid\":\"d5f55910-1810-a775-e91f-191d1f5608a4\",\"displayName\":\"Spot Checks of Each Category\",\"value\":\"d5f55910-1810-a775-e91f-191d1f5608a4\"}," +
                "{\"guid\":\"a36ce828-1810-49cb-b351-7051609e26af\",\"displayName\":\"Optional\",\"value\":\"a36ce828-1810-49cb-b351-7051609e26af\"}]}," +
                "{\"entityType\":\"ISSUE\",\"filterSelectorType\":\"LIST\",\"guid\":\"11111111-1111-1111-1111-111111111151\",\"displayName\":\"Analysis Type\"" +
                ",\"value\":\"11111111-1111-1111-1111-111111111151\",\"description\":\"\",\"selectorOptions\":[{\"guid\":\"SCA\",\"displayName\":\"SCA\",\"value\":\"SCA\"}]}," +
                "{\"entityType\":\"CUSTOMTAG\",\"filterSelectorType\":\"LIST\",\"guid\":\"87f2364f-dcd4-49e6-861d-f8d3f351686b\",\"displayName\":\"Analysis\"," +
                "\"value\":\"87f2364f-dcd4-49e6-861d-f8d3f351686b\",\"description\":\"The analysis tag must be set for an issue to be counted as 'Audited.' " +
                "This is encouraged to be the final action performed by an auditor.\",\"selectorOptions\":[{\"guid\":\"\",\"displayName\":\"NONE\",\"value\":\"\"}," +
                "{\"guid\":\"Not an Issue\",\"displayName\":\"Not an Issue\",\"value\":\"0\"},{\"guid\":\"Reliability Issue\",\"displayName\":\"Reliability Issue\",\"value\":\"1\"}," +
                "{\"guid\":\"Bad Practice\",\"displayName\":\"Bad Practice\",\"value\":\"2\"},{\"guid\":\"Suspicious\",\"displayName\":\"Suspicious\",\"value\":\"3\"}," +
                "{\"guid\":\"Exploitable\",\"displayName\":\"Exploitable\",\"value\":\"4\"}]},{\"entityType\":\"CUSTOMTAG\",\"filterSelectorType\":\"LIST\"," +
                "\"guid\":\"0aafa386-766b-484e-90c7-4b9397998dfc\",\"displayName\":\"Analysis State\",\"value\":\"0aafa386-766b-484e-90c7-4b9397998dfc\"," +
                "\"description\":\"Additional information about the issue\",\"selectorOptions\":[{\"guid\":\"\",\"displayName\":\"NONE\",\"value\":\"\"},{\"guid\":\"In Process\"," +
                "\"displayName\":\"In Process\",\"value\":\"0\"},{\"guid\":\"Completed\",\"displayName\":\"Completed\",\"value\":\"1\"},{\"guid\":\"Code Quality Issue\"," +
                "\"displayName\":\"Code Quality Issue\",\"value\":\"2\"},{\"guid\":\"Functional Bug\",\"displayName\":\"Functional Bug\",\"value\":\"3\"},{\"guid\":\"False Positive\"," +
                "\"displayName\":\"False Positive\",\"value\":\"4\"}]},{\"entityType\":\"HYBRIDTAG\",\"filterSelectorType\":\"LIST\",\"guid\":\"userAssignment\",\"displayName\":\"Assignments\"," +
                "\"value\":\"userAssignment\",\"description\":\"\",\"selectorOptions\":[{\"guid\":\"d054628\",\"displayName\":\"My Assignments\",\"value\":\"d054628\"}]}]}," +
                "\"responseCode\":200}"
            return result
        })

        def groupings = fortify.httpGroupings(nullScript, '7635675')
        Assert.assertNotNull(groupings.analysis)
        Assert.assertEquals("87f2364f-dcd4-49e6-861d-f8d3f351686b", groupings.analysis.id)
        Assert.assertEquals("Analysis", groupings.analysis.name)
        Assert.assertEquals("CUSTOMTAG", groupings.analysis.type)
    }
}

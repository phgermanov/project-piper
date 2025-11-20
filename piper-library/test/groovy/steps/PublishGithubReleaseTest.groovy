#!groovy
package steps

import groovy.json.JsonSlurperClassic

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.everyItem
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.hamcrest.Matchers.startsWith

import org.junit.Before
import org.junit.Ignore
import org.junit.Test
import org.junit.Rule
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import static org.junit.Assert.assertThat

import util.BasePiperTest

import util.Rules
import util.JenkinsStepRule
import util.JenkinsLoggingRule

@Ignore("step disabled")
class PublishGithubReleaseTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jsr)
        .around(thrown)

    def data
    def requestList = []

    @Before
    void init() throws Exception {
        // register Jenkins commands with mock values
        helper.registerAllowedMethod( "deleteDir", [], null )
        helper.registerAllowedMethod("httpRequest", [], null)
        helper.registerAllowedMethod("readJSON", [Map.class], { m ->
            return new JsonSlurperClassic().parseText(m.text)
        })
        helper.registerAllowedMethod('usernamePassword', [Map], { m -> return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            try {
                l.each {Map settings ->
                    binding.setProperty(settings.usernameVariable, 'foo')
                    binding.setProperty(settings.passwordVariable, '********')
                }
                c()
            } finally {
                l.each {Map settings ->
                    binding.setProperty(settings.usernameVariable, null)
                    binding.setProperty(settings.passwordVariable, null)
                }
            }
        })

        def responseLatestRelease = '{"url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/releases/26581","assets_url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/releases/26581/assets","upload_url":"https://github.wdf.sap.corp/api/uploads/repos/ContinuousDelivery/piper-library/releases/26581/assets{?name,label}","html_url":"https://github.wdf.sap.corp/ContinuousDelivery/piper-library/releases/tag/1.11.0-20180409-074550","id":26581,"tag_name":"1.11.0-20180409-074550","target_commitish":"master","name":"1.11.0-20180409-074550","draft":false,"author":{"login":"D032835","id":1809,"avatar_url":"https://github.wdf.sap.corp/avatars/u/1809?","gravatar_id":"","url":"https://github.wdf.sap.corp/api/v3/users/D032835","html_url":"https://github.wdf.sap.corp/D032835","followers_url":"https://github.wdf.sap.corp/api/v3/users/D032835/followers","following_url":"https://github.wdf.sap.corp/api/v3/users/D032835/following{/other_user}","gists_url":"https://github.wdf.sap.corp/api/v3/users/D032835/gists{/gist_id}","starred_url":"https://github.wdf.sap.corp/api/v3/users/D032835/starred{/owner}{/repo}","subscriptions_url":"https://github.wdf.sap.corp/api/v3/users/D032835/subscriptions","organizations_url":"https://github.wdf.sap.corp/api/v3/users/D032835/orgs","repos_url":"https://github.wdf.sap.corp/api/v3/users/D032835/repos","events_url":"https://github.wdf.sap.corp/api/v3/users/D032835/events{/privacy}","received_events_url":"https://github.wdf.sap.corp/api/v3/users/D032835/received_events","type":"User","site_admin":false},"prerelease":false,"created_at":"2018-04-09T07:45:38Z","published_at":"2018-04-09T07:52:49Z","assets":[],"tarball_url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/tarball/1.11.0-20180409-074550","zipball_url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/zipball/1.11.0-20180409-074550","body":"<br /><br />**List of closed pull-requests since last release**<br />[# 887](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/887): Add tests for checkmarx step<br />[# 907](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/907): Enable triaging<br />[# 908](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/908): SourceClear support reporting into fixed version<br />[# 909](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/909): add UserTriggerCause<br />[# 910](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/910): deployToKubernetes support of kubectl based deployment<br />[# 912](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/912): Speed up tests<br />[# 914](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/914): update config usage in writeInflux <br />[# 915](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/915): update config usage in executeVulasScan<br />[# 918](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/918): switch to new slack channel name<br />[# 921](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/921): correct jenkinsUtils object for restartableSteps step<br />[# 922](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/922): add default influx server<br />[# 924](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/924): update config usage in setupPipelineEnvironment<br />[# 925](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/925): Unstash content earlier to avoid FileNotFoundException<br />[# 927](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/927): Revert SourceClear support reporting into fixed version<br />[# 928](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/928): Fix rolling back mock behavior<br />[# 929](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/929): Add post-deploy actions via body<br />[# 930](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/930): parameters passed to resolveFortifyCredentialsID<br />[# 931](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/931): simplify bower installation for source clear <br />[# 932](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/932): use descriptive message for nodeAvailable<br />[# 934](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/934): Cease support  for fortify technical user<br />[# 937](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/937): setVersion - allow extension of maven parameters<br />[# 938](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/938): Improve Protecode vulnerability processing<br />[# 939](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/939): xMake Docker metadata available in global pipeline environment<br />[# 940](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/940): Add missing hand-over of globalPipelineEnvironment<br />[# 944](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/944): fix: translate gh issues in traceability reports to proper urls<br />[# 945](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/pull/945): Bump Version<br /><br />**List of closed issues since last release**<br />[# 382](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues/382): Snyk for security testing<br />[# 579](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues/579): Evaluate required enhancements for Kubernetes <br />[# 638](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues/638): Integrate Protecode for Docker scanning<br />[# 878](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues/878): setVersion - allow parametrization of Maven call<br />[# 920](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues/920): restartableSteps refers to wrong jenkinsUtils object<br />[# 941](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues/941): issue in deep config merge <br /><br />**Changes**<br />[1.10.0-20180326-070201...1.11.0-20180409-074550](https://github.wdf.sap.corp/ContinuousDelivery/piper-library/compare/1.10.0-20180326-070201...1.11.0-20180409-074550) <br />"}'
        def responseIssues = '[{"url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/issues/13","repository_url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library","labels_url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/issues/13/labels{/name}","comments_url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/issues/13/comments","events_url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/issues/13/events","html_url":"https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues/13","id":422536,"number":13,"title":"influx: add function to include performance result file (CSV)","user":{"login":"D065687","id":6991,"avatar_url":"https://github.wdf.sap.corp/avatars/u/6991?","gravatar_id":"","url":"https://github.wdf.sap.corp/api/v3/users/D065687","html_url":"https://github.wdf.sap.corp/D065687","followers_url":"https://github.wdf.sap.corp/api/v3/users/D065687/followers","following_url":"https://github.wdf.sap.corp/api/v3/users/D065687/following{/other_user}","gists_url":"https://github.wdf.sap.corp/api/v3/users/D065687/gists{/gist_id}","starred_url":"https://github.wdf.sap.corp/api/v3/users/D065687/starred{/owner}{/repo}","subscriptions_url":"https://github.wdf.sap.corp/api/v3/users/D065687/subscriptions","organizations_url":"https://github.wdf.sap.corp/api/v3/users/D065687/orgs","repos_url":"https://github.wdf.sap.corp/api/v3/users/D065687/repos","events_url":"https://github.wdf.sap.corp/api/v3/users/D065687/events{/privacy}","received_events_url":"https://github.wdf.sap.corp/api/v3/users/D065687/received_events","type":"User","site_admin":false},"labels":[{"id":541874,"url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/labels/enhancement","name":"enhancement","color":"84b6eb","default":true}],"state":"closed","locked":false,"assignee":null,"assignees":[],"milestone":null,"comments":1,"created_at":"2017-03-17T10:23:08Z","updated_at":"2017-08-02T08:39:37Z","closed_at":"2017-08-02T08:39:37Z","author_association":"OWNER","body":""},{"url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/issues/21","repository_url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library","labels_url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/issues/21/labels{/name}","comments_url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/issues/21/comments","events_url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/issues/21/events","html_url":"https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues/21","id":422768,"number":21,"title":"environment: provide convenient method to get property as boolean","user":{"login":"D065687","id":6991,"avatar_url":"https://github.wdf.sap.corp/avatars/u/6991?","gravatar_id":"","url":"https://github.wdf.sap.corp/api/v3/users/D065687","html_url":"https://github.wdf.sap.corp/D065687","followers_url":"https://github.wdf.sap.corp/api/v3/users/D065687/followers","following_url":"https://github.wdf.sap.corp/api/v3/users/D065687/following{/other_user}","gists_url":"https://github.wdf.sap.corp/api/v3/users/D065687/gists{/gist_id}","starred_url":"https://github.wdf.sap.corp/api/v3/users/D065687/starred{/owner}{/repo}","subscriptions_url":"https://github.wdf.sap.corp/api/v3/users/D065687/subscriptions","organizations_url":"https://github.wdf.sap.corp/api/v3/users/D065687/orgs","repos_url":"https://github.wdf.sap.corp/api/v3/users/D065687/repos","events_url":"https://github.wdf.sap.corp/api/v3/users/D065687/events{/privacy}","received_events_url":"https://github.wdf.sap.corp/api/v3/users/D065687/received_events","type":"User","site_admin":false},"labels":[{"id":541874,"url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/labels/enhancement","name":"enhancement","color":"84b6eb","default":true}],"state":"closed","locked":false,"assignee":{"login":"D065687","id":6991,"avatar_url":"https://github.wdf.sap.corp/avatars/u/6991?","gravatar_id":"","url":"https://github.wdf.sap.corp/api/v3/users/D065687","html_url":"https://github.wdf.sap.corp/D065687","followers_url":"https://github.wdf.sap.corp/api/v3/users/D065687/followers","following_url":"https://github.wdf.sap.corp/api/v3/users/D065687/following{/other_user}","gists_url":"https://github.wdf.sap.corp/api/v3/users/D065687/gists{/gist_id}","starred_url":"https://github.wdf.sap.corp/api/v3/users/D065687/starred{/owner}{/repo}","subscriptions_url":"https://github.wdf.sap.corp/api/v3/users/D065687/subscriptions","organizations_url":"https://github.wdf.sap.corp/api/v3/users/D065687/orgs","repos_url":"https://github.wdf.sap.corp/api/v3/users/D065687/repos","events_url":"https://github.wdf.sap.corp/api/v3/users/D065687/events{/privacy}","received_events_url":"https://github.wdf.sap.corp/api/v3/users/D065687/received_events","type":"User","site_admin":false},"assignees":[{"login":"D065687","id":6991,"avatar_url":"https://github.wdf.sap.corp/avatars/u/6991?","gravatar_id":"","url":"https://github.wdf.sap.corp/api/v3/users/D065687","html_url":"https://github.wdf.sap.corp/D065687","followers_url":"https://github.wdf.sap.corp/api/v3/users/D065687/followers","following_url":"https://github.wdf.sap.corp/api/v3/users/D065687/following{/other_user}","gists_url":"https://github.wdf.sap.corp/api/v3/users/D065687/gists{/gist_id}","starred_url":"https://github.wdf.sap.corp/api/v3/users/D065687/starred{/owner}{/repo}","subscriptions_url":"https://github.wdf.sap.corp/api/v3/users/D065687/subscriptions","organizations_url":"https://github.wdf.sap.corp/api/v3/users/D065687/orgs","repos_url":"https://github.wdf.sap.corp/api/v3/users/D065687/repos","events_url":"https://github.wdf.sap.corp/api/v3/users/D065687/events{/privacy}","received_events_url":"https://github.wdf.sap.corp/api/v3/users/D065687/received_events","type":"User","site_admin":false}],"milestone":null,"comments":1,"created_at":"2017-03-17T13:24:21Z","updated_at":"2017-08-03T09:32:45Z","closed_at":"2017-08-03T09:32:45Z","author_association":"OWNER","body":""}]'
        def responseRelease = '{"url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/releases/27149","assets_url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/releases/27149/assets","upload_url":"https://github.wdf.sap.corp/api/uploads/repos/ContinuousDelivery/piper-library/releases/27149/assets{?name,label}","html_url":"https://github.wdf.sap.corp/ContinuousDelivery/piper-library/releases/tag/test","id":27149,"tag_name":"test","target_commitish":"master","name":"v1.0.0","draft":false,"author":{"login":"D065687","id":6991,"avatar_url":"https://github.wdf.sap.corp/avatars/u/6991?","gravatar_id":"","url":"https://github.wdf.sap.corp/api/v3/users/D065687","html_url":"https://github.wdf.sap.corp/D065687","followers_url":"https://github.wdf.sap.corp/api/v3/users/D065687/followers","following_url":"https://github.wdf.sap.corp/api/v3/users/D065687/following{/other_user}","gists_url":"https://github.wdf.sap.corp/api/v3/users/D065687/gists{/gist_id}","starred_url":"https://github.wdf.sap.corp/api/v3/users/D065687/starred{/owner}{/repo}","subscriptions_url":"https://github.wdf.sap.corp/api/v3/users/D065687/subscriptions","organizations_url":"https://github.wdf.sap.corp/api/v3/users/D065687/orgs","repos_url":"https://github.wdf.sap.corp/api/v3/users/D065687/repos","events_url":"https://github.wdf.sap.corp/api/v3/users/D065687/events{/privacy}","received_events_url":"https://github.wdf.sap.corp/api/v3/users/D065687/received_events","type":"User","site_admin":false},"prerelease":false,"created_at":"2018-04-18T11:00:17Z","published_at":"2018-04-18T11:32:34Z","assets":[],"tarball_url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/tarball/test","zipball_url":"https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/zipball/test","body":"Description of the release"}'

        helper.registerAllowedMethod("httpRequest", [String.class], { s ->
            def result = ''
            requestList.push(s.toString())
            if(s.contains('/releases/latest?'))
                result = responseLatestRelease
            else if(s.contains('/issues?'))
                result = responseIssues
            return [content: result]
        })
        helper.registerAllowedMethod("httpRequest", [Map.class], { m ->
            def result = ''
            requestList.push(m?.url?.toString())
            if(m?.url?.contains('/releases?')){
                data = new JsonSlurperClassic().parseText(m?.requestBody?.toString())
                result = responseRelease
            }
            return [content: result]
        })
    }

    @Test
    void testPublishGithubReleaseWithDefaults() throws Exception {
        jsr.step.publishGithubRelease(
            script: nullScript,
            githubOrg: 'TestOrg',
            githubRepo: 'TestRepo',
            githubCredentialsId: 'TestCredentials',
            version: '1.2.3'
        )
        // asserts
        assertThat('this is not handled as a first release', jlr.log, not(containsString('[publishGithubRelease] This is the first release - no previous releases available')))
        assertThat('every request starts with the github api url', requestList, everyItem(startsWith('https://github.wdf.sap.corp/api/v3')))
        assertThat('every request contains the github org & repo', requestList, everyItem(containsString('/TestOrg/TestRepo/')))
        // test githubCredentialsId
        assertThat('every request has an access token', requestList, everyItem(containsString('access_token=********')))
        // test releaseBodyHeader
        assertThat('the header is not set', data.body, startsWith(''))
        // test addClosedIssues
        assertThat('the list of closed PR is not present', data.body, not(containsString('**List of closed pull-requests since last release**')))
        assertThat('the list of closed issues is not present', data.body, not(containsString('**List of closed issues since last release**')))
        // test addDeltaToLastRelease
        assertThat('the compare link is not present', data.body, not(containsString('[1.11.0-20180409-074550...1.2.3]')))

        assertThat(data.name, is('1.2.3'))
        assertThat(data.tag_name, is('1.2.3'))
        assertThat(data.draft, is(false))
        assertThat(data.prerelease, is(false))
        assertJobStatusSuccess()
    }

    @Test
    void testPublishGithubRelease() throws Exception {
        jsr.step.publishGithubRelease(
            script: nullScript,
            githubOrg: 'TestOrg',
            githubRepo: 'TestRepo',
            githubCredentialsId: 'TestCredentials',
            version: '1.2.3',
            releaseBodyHeader: '**TestHeader**',
            addClosedIssues: true,
            addDeltaToLastRelease: true
        )
        // asserts
        assertThat('this is not handled as a first release', jlr.log, not(containsString('[publishGithubRelease] This is the first release - no previous releases available')))
        assertThat('every request starts with the github api url', requestList, everyItem(startsWith('https://github.wdf.sap.corp/api/v3')))
        assertThat('every request contains the github org & repo', requestList, everyItem(containsString('/TestOrg/TestRepo/')))
        // test githubCredentialsId
        assertThat('every request has an access token', requestList, everyItem(containsString('access_token=********')))
        // test releaseBodyHeader
        assertThat('the header is set', data.body, startsWith('**TestHeader**'))
        // test addClosedIssues
        assertThat('the list of closed PR is present', data.body, containsString('**List of closed pull-requests since last release**'))
        assertThat('the list of closed issues is present', data.body, containsString('**List of closed issues since last release**'))
        // test addDeltaToLastRelease
        assertThat('the compare link is present', data.body, containsString('[1.11.0-20180409-074550...1.2.3]'))
        assertThat('the default github url is used', data.body, containsString('https://github.wdf.sap.corp'))

        //test fix for https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues/1047
        assertThat(requestList[1].toString(), is('https://github.wdf.sap.corp/api/v3/repos/TestOrg/TestRepo/issues?access_token=********&per_page=100&state=closed&direction=asc&since=2018-04-09T07:52:49Z'))

        assertThat(data.name, is('1.2.3'))
        assertThat(data.tag_name, is('1.2.3'))
        assertThat(data.draft, is(false))
        assertThat(data.prerelease, is(false))
        assertJobStatusSuccess()
    }

    @Test
    void testExcludeLabels() throws Exception {
        jsr.step.publishGithubRelease(
            script: nullScript,
            githubOrg: 'TestOrg',
            githubRepo: 'TestRepo',
            githubCredentialsId: 'TestCredentials',
            version: '1.2.3',
            releaseBodyHeader: '**TestHeader**',
            addClosedIssues: true,
            addDeltaToLastRelease: true,
            excludeLabels: ['enhancement']
        )
        // asserts
        assertThat('issues with excluded labels are not listed', data.body, not(containsString('influx: add function to include performance result file (CSV)')))
        assertJobStatusSuccess()
    }

    @Test
    void testIsExcluded() throws Exception {
        def item = new JsonSlurperClassic().parseText('''{
            "id": 422536,
            "number": 13,
            "title": "influx: add function to include performance result file (CSV)",
            "user": {
                "login": "D065687",
                "id": 6991,
                "type": "User",
                "site_admin": false
            },
            "labels": [{
                "id": 541874,
                "url": "https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/labels/enhancement",
                "name": "enhancement",
                "color": "84b6eb",
                "default": true
            }],
            "state": "closed",
            "locked": false,
            "body": ""
        }''')
        // asserts
        assertThat(jsr.step.isExcluded(item, ['enhancement', 'won\'t fix']), is(true))
        assertThat(jsr.step.isExcluded(item, ['won\'t fix']), is(false))
        assertJobStatusSuccess()
    }
}

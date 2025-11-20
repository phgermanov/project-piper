package com.sap.piper.internal

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsReadJsonRule
import util.JenkinsShellCallRule
import util.Rules

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.is
import static org.junit.Assert.assertThat

class PiperGoUtilsTest extends BasePiperTest {

    public ExpectedException exception = ExpectedException.none()
    public JenkinsShellCallRule shellCallRule = new JenkinsShellCallRule(this)
    public JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    public JenkinsReadJsonRule readJsonRule = new JenkinsReadJsonRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(shellCallRule)
        .around(exception)
        .around(loggingRule)
        .around(readJsonRule)

    @Before
    void init() {
        helper.registerAllowedMethod("withCredentials", [List.class, Closure.class], { credentials, body ->
            body()
        })
    }

    @Test
    void testUnstashPiperBinAvailable() {

        def piperBinStash = 'sap-piper-bin'

        // this mocks utils.unstash
        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            if (stashFileName != piperBinStash) {
                return []
            }
            return [piperBinStash]
        })

        def piperGoUtils = new PiperGoUtils(nullScript, utils)

        piperGoUtils.unstashPiperBin()
    }

    @Test
    void testUnstashPiperBinLatestArtifactory() {

        def piperGoUtils = new PiperGoUtils(nullScript, utils)
        piperGoUtils.metaClass.getLibrariesInfo = {-> return [[name: 'piper-lib', version: 'latest']]}

        // this mocks utils.unstash - mimic stash not existing
        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            return []
        })

        shellCallRule.setReturnStatus('[ -x ./sap-piper ]', 1)
        final String jsonResponse = '200'
        String script1 = 'curl --silent --location --write-out \'%{http_code}\' --output sap-piper \'https://sap-piper.int.repositories.cloud.sap/artifactory/sap-piper/releases/latest/sap-piper\''
        shellCallRule.setReturnValue(script1, jsonResponse)

        piperGoUtils.unstashPiperBin()
        assertThat(shellCallRule.shell.size(), is(3))
        assertThat(shellCallRule.shell[1].toString(), is(script1))
        assertThat(shellCallRule.shell[2].toString(), is('chmod +x sap-piper'))
    }

    @Test
    void testUnstashPiperBinBranchArtifactory() {

        def piperGoUtils = new PiperGoUtils(nullScript, utils)
        piperGoUtils.metaClass.getLibrariesInfo = {-> return [[name: 'piper-lib', version: 'testTag']]}

        // this mocks utils.unstash - mimic stash not existing
        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            return []
        })

        shellCallRule.setReturnStatus('[ -x ./sap-piper ]', 1)
        final String jsonResponse = '200'
        String script1 = 'curl --silent --location --write-out \'%{http_code}\' --output sap-piper \'https://sap-piper.int.repositories.cloud.sap/artifactory/sap-piper/releases/testTag/sap-piper\''
        shellCallRule.setReturnValue(script1, jsonResponse)

        piperGoUtils.unstashPiperBin()
        assertThat(shellCallRule.shell.size(), is(3))
        assertThat(shellCallRule.shell[1].toString(), is(script1))
        assertThat(shellCallRule.shell[2].toString(), is('chmod +x sap-piper'))
    }

    @Test
    void testUnstashPiperBinFallbackArtifactory() {

        def piperGoUtils = new PiperGoUtils(nullScript, utils)
        piperGoUtils.metaClass.getLibrariesInfo = {-> return [[name: 'piper-lib', version: 'notAvailable']]}

        shellCallRule.setReturnStatus('[ -x ./sap-piper ]', 1)
        shellCallRule.setReturnStatus('[ -x ./sap-piper ]', 1)
        final String jsonResponse1 = '404'
        String script1 = 'curl --silent --location --write-out \'%{http_code}\' --output sap-piper \'https://sap-piper.int.repositories.cloud.sap/artifactory/sap-piper/releases/notAvailable/sap-piper\''
        shellCallRule.setReturnValue(script1, jsonResponse1)
        shellCallRule.setReturnStatus('[ -x ./sap-piper ]', 1)
        final String jsonResponse2 = '200'
        String script2 = 'curl --silent --location --write-out \'%{http_code}\' --output sap-piper \'https://sap-piper.int.repositories.cloud.sap/artifactory/sap-piper/releases/latest/sap-piper\''
        shellCallRule.setReturnValue(script2, jsonResponse2)

        // this mocks utils.unstash - mimic stash not existing
        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            return []
        })

        piperGoUtils.unstashPiperBin()
        assertThat(shellCallRule.shell.size(), is(4))
        assertThat(shellCallRule.shell[0].toString(), is('[ -x ./sap-piper ]'))
        assertThat(shellCallRule.shell[1].toString(), is(script1))
        assertThat(shellCallRule.shell[2].toString(), is(script2))
        assertThat(shellCallRule.shell[3].toString(), is('chmod +x sap-piper'))
    }

    @Test
    void testUnstashPiperBinLatest() {

        def piperGoUtils = new PiperGoUtils(nullScript, utils)
        piperGoUtils.metaClass.getLibrariesInfo = {-> return [[name: 'piper-lib', version: 'latest']]}
        piperGoUtils.metaClass.gitHubAPIRequest = {String token, String params -> return [assets: [[name: 'sap-piper', id: '42']]]}
        piperGoUtils.metaClass.curlWithOutput = {String url, String outputName -> return false}

        // this mocks utils.unstash - mimic stash not existing
        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            return []
        })

        shellCallRule.setReturnStatus('[ -x ./sap-piper ]', 1)

        String script = 'curl --header "Authorization: token my-token" --insecure --silent --location --header "Accept: application/octet-stream" --write-out \'-DeLiMiTeR-status=%{http_code}\' --output "./sap-piper" \'https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/releases/assets/42\''
        shellCallRule.setReturnValue(script, 'this is some return -DeLiMiTeR-status=200')

        helper.registerAllowedMethod("usernamePassword", [LinkedHashMap.class], { return null})

        helper.registerAllowedMethod("withCredentials", [List.class, Closure.class], { credentials, body ->
            nullScript.env.TOKEN = 'my-token'
            body()
            nullScript.env.TOKEN = null
        })

        piperGoUtils.setSapPiperDownloadCredentialsId('testCredentials')
        piperGoUtils.unstashPiperBin()
        assertThat(shellCallRule.shell.size(), is(6))
        assertThat(shellCallRule.shell[3].toString(), is("cat sap-piper"))
        assertThat(shellCallRule.shell[4].toString(), is(script))
        assertThat(shellCallRule.shell[5].toString(), is('chmod +x sap-piper'))
    }

    @Test
    void testUnstashPiperBinNonLatest() {

        def piperGoUtils = new PiperGoUtils(nullScript, utils)
        piperGoUtils.metaClass.getLibrariesInfo = {-> return [[name: 'piper-lib', version: 'testTag']]}
        piperGoUtils.metaClass.curlWithOutput = {String url, String outputName -> return false}

        // this mocks utils.unstash - mimic stash not existing
        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            return []
        })

        shellCallRule.setReturnStatus('[ -x ./sap-piper ]', 1)

        final String jsonResponse = '{"assets":[{"name":"sap-piper","id":"42"}]}'
        String script1 = 'curl --header "Authorization: token my-token" --insecure --silent --location --header "Accept: application/vnd.github.v3+json" \'https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/releases/tags/testTag\''
        shellCallRule.setReturnValue(script1, jsonResponse)

        String script2 = 'curl --header "Authorization: token my-token" --insecure --silent --location --header "Accept: application/octet-stream" --write-out \'-DeLiMiTeR-status=%{http_code}\' --output "./sap-piper" \'https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/releases/assets/42\''
        shellCallRule.setReturnValue(script2, 'this is some return -DeLiMiTeR-status=200')

        helper.registerAllowedMethod("usernamePassword", [LinkedHashMap.class], { return null})

        helper.registerAllowedMethod("withCredentials", [List.class, Closure.class], { credentials, body ->
            nullScript.env.TOKEN = 'my-token'
            body()
            nullScript.env.TOKEN = null
        })

        piperGoUtils.setSapPiperDownloadCredentialsId('testCredentials')
        piperGoUtils.unstashPiperBin()
        assertThat(shellCallRule.shell.size(), is(7))
        assertThat(shellCallRule.shell[3].toString(), is("cat sap-piper"))
        assertThat(shellCallRule.shell[4].toString(), is(script1))
        assertThat(shellCallRule.shell[5].toString(), is(script2))
        assertThat(shellCallRule.shell[6].toString(), is('chmod +x sap-piper'))
    }

    @Test
    void testUnstashPiperBinFallback() {

        def piperGoUtils = new PiperGoUtils(nullScript, utils)
        piperGoUtils.metaClass.getLibrariesInfo = {-> return [[name: 'piper-lib', version: 'notAvailable']]}
        piperGoUtils.metaClass.curlWithOutput = {String url, String outputName -> return false}

        shellCallRule.setReturnStatus('[ -x ./sap-piper ]', 1)

        final String jsonResponse1 = '{"message":"Not Found"}'
        String script1 = 'curl --header "Authorization: token my-token" --insecure --silent --location --header "Accept: application/vnd.github.v3+json" \'https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/releases/tags/notAvailable\''
        shellCallRule.setReturnValue(script1, jsonResponse1)

        final String jsonResponse2 = '{"assets":[{"name":"sap-piper","id":"42"}]}'
        String script2 = 'curl --header "Authorization: token my-token" --insecure --silent --location --header "Accept: application/vnd.github.v3+json" \'https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/releases/latest\''
        shellCallRule.setReturnValue(script2, jsonResponse2)

        String script3 = 'curl --header "Authorization: token my-token" --insecure --silent --location --header "Accept: application/octet-stream" --write-out \'-DeLiMiTeR-status=%{http_code}\' --output "./sap-piper" \'https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/releases/assets/42\''
        shellCallRule.setReturnValue(script3, 'this is some return -DeLiMiTeR-status=200')

        // this mocks utils.unstash - mimic stash not existing
        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            return []
        })

        helper.registerAllowedMethod("usernamePassword", [LinkedHashMap.class], { return null})

        helper.registerAllowedMethod("withCredentials", [List.class, Closure.class], { credentials, body ->
            nullScript.env.TOKEN = 'my-token'
            body()
            nullScript.env.TOKEN = null
        })

        piperGoUtils.setSapPiperDownloadCredentialsId('testCredentials')
        piperGoUtils.unstashPiperBin()

        assertThat(shellCallRule.shell.size(), is(8))
        assertThat(shellCallRule.shell[3].toString(), is("cat sap-piper"))
        assertThat(shellCallRule.shell[4].toString(), is(script1))
        assertThat(shellCallRule.shell[5].toString(), is(script2))
        assertThat(shellCallRule.shell[6].toString(), is(script3))
        assertThat(shellCallRule.shell[7].toString(), is('chmod +x sap-piper'))
    }

    @Test
    void testDownloadFailed() {
        def piperGoUtils = new PiperGoUtils(nullScript, utils)
        piperGoUtils.metaClass.getLibrariesInfo = {-> return [[name: 'piper-lib', version: 'latest']]}
        piperGoUtils.metaClass.gitHubAPIRequest = {String token, String params -> return [assets: [[name: 'sap-piper', id: '42']]]}
        piperGoUtils.metaClass.curlWithOutput = {String url, String outputName -> return false}

        shellCallRule.setReturnStatus('[ -x ./sap-piper ]', 1)

        String script = 'curl --insecure --silent --location --header "Accept: application/octet-stream" --write-out \'-DeLiMiTeR-status=%{http_code}\' --output "./sap-piper" \'https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/releases/assets/42\''
        shellCallRule.setReturnValue(script, 'this is some return -DeLiMiTeR-status=500')

        helper.registerAllowedMethod("withCredentials", [List.class, Closure.class], { credentials, body ->
            body()
        })

        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            return []
        })

        exception.expectMessage(containsString('Downloading the Piper binary from Artifactory failed'))
        piperGoUtils.unstashPiperBin()
    }

    @Test
    void testUnstashPiperBinWithToken() {

        def piperGoUtils = new PiperGoUtils(nullScript, utils)
        piperGoUtils.setSapPiperDownloadCredentialsId('myCredentialsId')
        piperGoUtils.metaClass.getLibrariesInfo = {-> return [[name: 'piper-lib', version: 'latest']]}
        piperGoUtils.metaClass.curlWithOutput = {String url, String outputName -> return false}

        // this mocks utils.unstash - mimic stash not existing
        helper.registerAllowedMethod("unstash", [String.class], { stashFileName ->
            return []
        })
        helper.registerAllowedMethod("usernamePassword", [LinkedHashMap.class], { map ->
            assertThat(map.credentialsId, is('myCredentialsId'))
            assertThat(map.passwordVariable, is('TOKEN'))
        })
        helper.registerAllowedMethod("withCredentials", [List.class, Closure.class], { credentials, body ->
            nullScript.env.TOKEN = 'my-token'
            body()
            nullScript.env.TOKEN = null
        })

        shellCallRule.setReturnStatus('[ -x ./sap-piper ]', 1)

        final String jsonResponse = '{"assets":[{"name":"sap-piper","id":"42"}]}'
        String script1 = 'curl --header "Authorization: token my-token" --insecure --silent --location --header "Accept: application/vnd.github.v3+json" \'https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/releases/latest\''
        shellCallRule.setReturnValue(script1, jsonResponse)

        String script2 = 'curl --header "Authorization: token my-token" --insecure --silent --location --header "Accept: application/octet-stream" --write-out \'-DeLiMiTeR-status=%{http_code}\' --output "./sap-piper" \'https://github.wdf.sap.corp/api/v3/repos/ContinuousDelivery/piper-library/releases/assets/42\''
        shellCallRule.setReturnValue(script2, 'this is some return -DeLiMiTeR-status=200')

        piperGoUtils.unstashPiperBin()
        assertThat(shellCallRule.shell.size(), is(7))
        assertThat(shellCallRule.shell[3].toString(), is("cat sap-piper"))
        assertThat(shellCallRule.shell[4].toString(), is(script1))
        assertThat(shellCallRule.shell[5].toString(), is(script2))
        assertThat(shellCallRule.shell[6].toString(), is('chmod +x sap-piper'))
    }

}


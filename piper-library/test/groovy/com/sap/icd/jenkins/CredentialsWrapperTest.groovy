package com.sap.icd.jenkins

import org.codehaus.groovy.runtime.InvokerHelper
import org.junit.After
import org.junit.Before
import org.junit.Test

class CredentialsWrapperTest {

    Binding binding = new Binding()
    Script script

    static final int RETURNCODE = 111

    static final String CF_USERNAME = "CF_USERNAME"
    static final String MH_USERNAME = "MH_USERNAME"

    static final String CF_PASSWORD_ENCODED = "Q0ZfUEFTU1dPUkQK"
    static final String MH_PASSWORD_ENCODED = "TUhfUEFTU1dPUkQK"

    static final String CF_PASSWORD = "CF_PASSWORD"
    static final String MH_PASSWORD = "MH_PASSWORD"

    static final String CF_PASSWORD_STORE = "CF_PASSWORD_STORE"
    static final String MH_PASSWORD_STORE = "MH_PASSWORD_STORE"

    static final String CF_CREDENTIALS_ID = "cfCredentialsId"
    static final String MH_CREDENTIALS_ID = "mhCredentialsId"


    @Before
    void setUp() throws Exception {
        script = new InvokerHelper.NullScript(binding)
    }

    @After
    public void cleanup() {
        Script.metaClass=null
    }

    private CallCounter setupWithCredentialsUsernamePasswordMetaClass()  {
        CallCounter callCounter = new CallCounter()
        Script.metaClass.usernamePassword = { Map map ->
            if (map.credentialsId==CF_CREDENTIALS_ID) {
                assert map.credentialsId == CF_CREDENTIALS_ID
                assert map.passwordVariable == CF_PASSWORD_STORE
                assert map.usernameVariable == CF_USERNAME
                return [CF_PASSWORD_STORE: CF_PASSWORD, CF_USERNAME: CF_USERNAME]
            }
            else {
                assert map.credentialsId == MH_CREDENTIALS_ID
                assert map.passwordVariable == MH_PASSWORD_STORE
                assert map.usernameVariable == MH_USERNAME
                return [MH_PASSWORD_STORE: MH_PASSWORD, MH_USERNAME: MH_USERNAME]
            }
        }

        Script.metaClass.withCredentials = { List list, Closure closure ->
            list.forEach { it.each { k, v ->
                binding.setVariable(k,"$v")
            } }
            def res = closure.call()
            list.forEach { it.each { k, v ->
                binding.setVariable(k,null)
            } }
            callCounter.called()
            return res
        }
        return callCounter
    }

    private CallCounter setupWithEnv() {
        CallCounter callCounter = new CallCounter()

        Script.metaClass.withEnv = { List<String> list, Closure closure ->

            assert ['MH_PASSWORD_ENCODED=true','CF_PASSWORD_ENCODED=true'].contains(list[0])
            assert ["CF_PASSWORD=$CF_PASSWORD_ENCODED","MH_PASSWORD=$MH_PASSWORD_ENCODED"].contains(list[1])
            callCounter.called()
            return closure.call()
        }
        return callCounter
    }

    @Test
    void shouldSetCFCredentialsAndCallBody() {

        CallCounter credentialsCounter = setupWithCredentialsUsernamePasswordMetaClass()
        CallCounter withEnvCounter = setupWithEnv()
        CallCounter bodyCounter = new CallCounter()

        assert new CredentialsWrapper(script,CF_CREDENTIALS_ID).asCFCredentialsEncoded({bodyCounter.called()
            return RETURNCODE
        }) == RETURNCODE
        assert bodyCounter.gotCalled(1)
        assert credentialsCounter.gotCalled(1)
        assert withEnvCounter.gotCalled(1)
    }

    @Test
    void shouldSetMHCredentialsAndCallBody() {

        CallCounter credentialsCounter = setupWithCredentialsUsernamePasswordMetaClass()
        CallCounter withEnvCounter = setupWithEnv()
        CallCounter bodyCounter = new CallCounter()

        assert new CredentialsWrapper(script,MH_CREDENTIALS_ID).asMuenchhausenCredentialsEncoded ({bodyCounter.called()
            return RETURNCODE
        }) == RETURNCODE
        assert bodyCounter.gotCalled(1)
        assert credentialsCounter.gotCalled(1)
        assert withEnvCounter.gotCalled(1)

    }

    @Test
    void shouldOnlyCallBodyIfCredentialsAreNull() {

        CallCounter credentialsCounter = setupWithCredentialsUsernamePasswordMetaClass()
        CallCounter withEnvCounter = setupWithEnv()
        CallCounter bodyCounter = new CallCounter()

        assert new CredentialsWrapper(script,null).asMuenchhausenCredentialsEncoded ({bodyCounter.called()
            return RETURNCODE
        }) == RETURNCODE
        assert bodyCounter.gotCalled(1)
        assert credentialsCounter.gotCalled(0)
        assert withEnvCounter.gotCalled(0)
    }
}

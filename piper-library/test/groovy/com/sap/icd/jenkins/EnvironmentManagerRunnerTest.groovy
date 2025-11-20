package com.sap.icd.jenkins

import org.codehaus.groovy.runtime.InvokerHelper
import org.junit.After
import org.junit.Before
import org.junit.Test

class EnvironmentManagerRunnerTest {

    static final String A_COMMAND = "aCommand"
    static final String CF_CREDENTIALS_ID = "cfCredentialsId"
    static final String MH_CREDENTIALS_ID = "mhCredentialsId"

    static final int SUCCESSFUL = 0
    public static final String PATH_TO_WORK_SPACE = "PathToWorkSpace"
    public static final String DOCKER_IMAGE = "DockerImage"

    Binding binding = new Binding()

    Script script

    @Before
    void setUp() throws Exception {
        Script.metaClass.dir = { String s, Closure cl ->
            cl()
        }
        Script.metaClass.echo = { String s ->
        }
        script = new InvokerHelper.NullScript(binding)
    }

    @After
    public void cleanup() {
        Script.metaClass = null
        CredentialsWrapper.metaClass = null
    }

    private CallCounter setupExecuteDockerMetaClass() {
        CallCounter callCounter = new CallCounter()
        Script.metaClass.dockerExecute = { Map m, Closure cl ->
            assert m.dockerWorkspace == PATH_TO_WORK_SPACE
            assert m.dockerImage == DOCKER_IMAGE
            callCounter.called()
            cl()
        }
        return callCounter
    }

    private CallCounter setupShMetaClass(List calls) {
        CallCounter callCounter = new CallCounter()
        Script.metaClass.sh = { Map m ->
            callCounter.called()
            calls.add(m.script)
            return SUCCESSFUL
        }
        Script.metaClass.sh = { String s ->
            callCounter.called()
            calls.add(s)
            return SUCCESSFUL
        }
        return callCounter
    }

    private CallCounter setupCFCredentialsWrapper() {
        CallCounter cc = new CallCounter()
        CredentialsWrapper.metaClass.asCFCredentialsEncoded = { Closure c ->
            cc.called()
            assert credentialsId == CF_CREDENTIALS_ID
            return c.call()
        }
        return cc
    }

    private CallCounter setupMuenchhausenCredentialsWrapper(String credentialsValue) {
        CallCounter cc = new CallCounter()
        CredentialsWrapper.metaClass.asMuenchhausenCredentialsEncoded = { Closure c ->
            cc.called()
            assert credentialsId == credentialsValue
            return c.call()
        }
        return cc
    }

    private CallCounter setupScriptGitMock() {
        CallCounter cc = new CallCounter()
        Script.metaClass.git = { Map map ->
            cc.called()
            assert map.branch == "master"
            assert map.url == "https://github.wdf.sap.corp/cc-devops-envman/EnvMan"
        }
        return cc
    }

    @Test
    public void shouldCallGroovyWithCommand() {
        List<String> callStringsToSh = []
        CallCounter callsToSh = setupShMetaClass(callStringsToSh)
        CallCounter callsToCFCW = setupCFCredentialsWrapper()
        CallCounter callsToMHCW = setupMuenchhausenCredentialsWrapper(null)
        CallCounter callsToExecuteDocker = setupExecuteDockerMetaClass()
        CallCounter callsToGit = setupScriptGitMock()
        Script.metaClass.getEnv = { -> [jaas_owner: true] }

        EnvironmentManagerRunner emr = new EnvironmentManagerRunner(script, CF_CREDENTIALS_ID, null, null)
        emr.execute(A_COMMAND)

        assert callsToSh.gotCalled(2)
        assert callStringsToSh[0].contains(A_COMMAND)
        assert callStringsToSh[1].contains("rm -r")
        assert callsToExecuteDocker.gotCalled(0)
        assert callsToCFCW.gotCalled(1)
        assert callsToMHCW.gotCalled(1)
        assert callsToGit.gotCalled(1)
    }

    @Test
    public void shouldSetupMHCredentialsIfNotNull() {
        List<String> callStringsToSh = []
        CallCounter callsToSh = setupShMetaClass(callStringsToSh)
        CallCounter callsToCFCW = setupCFCredentialsWrapper()
        CallCounter callsToMHCW = setupMuenchhausenCredentialsWrapper(MH_CREDENTIALS_ID)
        CallCounter callsToExecuteDocker = setupExecuteDockerMetaClass()
        CallCounter callsToGit = setupScriptGitMock()
        Script.metaClass.getEnv = { -> [jaas_owner: true] }

        EnvironmentManagerRunner emr = new EnvironmentManagerRunner(script, CF_CREDENTIALS_ID, MH_CREDENTIALS_ID, null)
        emr.execute(A_COMMAND)

        assert callsToSh.gotCalled(2)
        assert callStringsToSh[0].contains(A_COMMAND)
        assert callStringsToSh[1].contains("rm -r")
        assert callsToExecuteDocker.gotCalled(0)
        assert callsToCFCW.gotCalled(1)
        assert callsToMHCW.gotCalled(1)
        assert callsToGit.gotCalled(1)
    }


    @Test
    public void shouldCallGroovyWithCommandInDocker() {
        List<String> callStringsToSh = []
        CallCounter callsToSh = setupShMetaClass(callStringsToSh)
        CallCounter callsToCFCW = setupCFCredentialsWrapper()
        CallCounter callsToMHCW = setupMuenchhausenCredentialsWrapper(null)
        CallCounter callsToExecuteDocker = setupExecuteDockerMetaClass()
        CallCounter callsToGit = setupScriptGitMock()
        Script.metaClass.getEnv = { -> [jaas_owner: false] }

        EnvironmentManagerRunner emr = new EnvironmentManagerRunner(script, CF_CREDENTIALS_ID, null, null)
        emr.executeWithDocker(A_COMMAND, DOCKER_IMAGE, PATH_TO_WORK_SPACE)

        assert callsToSh.gotCalled(2)
        assert callStringsToSh[0].contains(A_COMMAND)
        assert callStringsToSh[1].contains("rm -r")
        assert callsToExecuteDocker.gotCalled(1)
        assert callsToCFCW.gotCalled(1)
        assert callsToMHCW.gotCalled(1)
        assert callsToGit.gotCalled(1)
    }

    @Test
    public void shouldCallGroovyWithCommandInDockerInKubernetes() {
        List<String> callStringsToSh = []
        CallCounter callsToSh = setupShMetaClass(callStringsToSh)
        CallCounter callsToCFCW = setupCFCredentialsWrapper()
        CallCounter callsToMHCW = setupMuenchhausenCredentialsWrapper(null)
        CallCounter callsToExecuteDocker = setupExecuteDockerMetaClass()
        CallCounter callsToGit = setupScriptGitMock()
        Script.metaClass.getEnv = { -> [jaas_owner: true] }

        EnvironmentManagerRunner emr = new EnvironmentManagerRunner(script, CF_CREDENTIALS_ID, null, null)
        emr.executeWithDocker(A_COMMAND, DOCKER_IMAGE, PATH_TO_WORK_SPACE)

        assert callsToSh.gotCalled(2)
        assert callStringsToSh[0].contains(A_COMMAND)
        assert callStringsToSh[0].contains('/src/com/sap/icd/environmentManager/EnvironmentManager.groovy')
        assert callStringsToSh[1].contains('rm -r')
        assert callsToExecuteDocker.gotCalled(1)
        assert callsToCFCW.gotCalled(1)
        assert callsToMHCW.gotCalled(1)
        assert callsToGit.gotCalled(1)
    }
}

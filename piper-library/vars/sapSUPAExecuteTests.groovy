import com.sap.piper.internal.PiperGoUtils
import com.sap.icd.jenkins.Utils
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/sapSUPAExecuteTests.yaml'

void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    Map stepParams = PiperGoUtils.prepare(this, script).plus(parameters)
    stepParams.stashNoDefaultExcludes = true
    List credentials = [
        [type: 'token', id: 'githubTokenCredentialsId', env: ['PIPER_githubToken']],
        [type: 'token', id: 'supaKeystoreKeyId', env: ['PIPER_supaKeystoreKey']]
    ]

    utils.unstashAll(['source'])
    piperExecuteBin(stepParams, STEP_NAME, METADATA_FILE, credentials)
}

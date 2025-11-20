import com.sap.piper.internal.PiperGoUtils
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/sapCallStagingService.yaml'

void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this

    Map stepParams = PiperGoUtils.prepare(this, script).plus(parameters)
    List credentials = [
        [type: 'usernamePassword', id: 'stagingUserCredentialsId', env: ['PIPER_username', 'PIPER_password']],
        [type: 'usernamePassword', id: 'stagingTenantCredentialsId', env: ['PIPER_tenantId', 'PIPER_tenantSecret']],
        [type: 'file', id: 'dockerConfigJsonCredentialsId', env: ['PIPER_dockerConfigJSON']]
    ]

    piperExecuteBin(stepParams, STEP_NAME, METADATA_FILE, credentials)
}

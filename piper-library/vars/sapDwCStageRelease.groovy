import com.sap.piper.internal.PiperGoUtils
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/sapDwCStageRelease.yaml'

void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this

    Map stepParams = PiperGoUtils.prepare(this, script).plus(parameters)
    List credentials = [
        [type: 'file', id: 'themistoInstanceCertificateCredentialsId', env: ['PIPER_themistoInstanceCertificatePath']],
        [type: 'file', id: 'gatewayCertificateCredentialsId', env: ['PIPER_gatewayCertificatePath']],
        [type: 'token', id: 'githubTokenCredentialsId', env: ['PIPER_githubToken']]
    ]

    piperExecuteBin(stepParams, STEP_NAME, METADATA_FILE, credentials)
}

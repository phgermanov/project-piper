import com.sap.piper.internal.PiperGoUtils
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/sapCumulusUpload.yaml'

//Metadata maintained in file resources/metadata/sapCumulusUpload.yaml

void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    Map stepParams = PiperGoUtils.prepare(this, script).plus(parameters)
    List credentials = [
        [type: 'file', id: 'cumulusFileCredentialsId', env: ['PIPER_jsonKeyFilePath']]
    ]
    piperExecuteBin(stepParams, STEP_NAME, METADATA_FILE, credentials)
}

import com.sap.piper.BuildTool
import com.sap.piper.internal.PiperGoUtils
import com.sap.piper.DownloadCacheUtils
import groovy.transform.Field
import static com.sap.piper.Prerequisites.checkScript
import com.sap.icd.jenkins.Utils

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/sapAccessContinuumExecuteTests.yaml'
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    Map stepParams = PiperGoUtils.prepare(this, script).plus(parameters)
    def utils = parameters.juStabUtils ?: new Utils()

    List credentials = [
        [type: 'token', id: 'AMPTokenCredentialsID', env: ['PIPER_token']]
    ]
    //Get the workspace source i.e. continuum related files 
    utils.unstashAll(['source'])
    piperExecuteBin(stepParams, STEP_NAME, METADATA_FILE, credentials)
}

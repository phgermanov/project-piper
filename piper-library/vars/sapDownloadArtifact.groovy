import com.sap.piper.BuildTool
import com.sap.piper.DownloadCacheUtils
import com.sap.piper.internal.PiperGoUtils
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/sapDownloadArtifact.yaml'

void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    parameters = DownloadCacheUtils.injectDownloadCacheInParameters(script, parameters, BuildTool.MAVEN)

    Map stepParams = PiperGoUtils.prepare(this, script).plus(parameters)
    List credentials = [[type: 'token', id: 'artifactoryTokenCredentialsId', env: ['PIPER_artifactoryToken']]]

    piperExecuteBin(stepParams, STEP_NAME, METADATA_FILE, credentials)
    // adding downloaded artifacts to stash for helm use cases (for kubernetesDeploy step)
    stash name: 'downloadedArtifact', includes: '**/*.tgz', allowEmpty: true
}

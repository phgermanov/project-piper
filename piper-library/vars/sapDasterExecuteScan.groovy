import com.sap.piper.internal.PiperGoUtils
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript


@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/sapDasterExecuteScan.yaml'

/**
 * The name DASTer is derived from **D**ynamic **A**pplication **S**ecurity **T**esting. As the name implies, the tool targets to provide black-box security testing capabilities for your solutions
 * in an automated fashion.
 *
 * DASTer itself ships with a [Swagger based frontend](https://daster.tools.sap/api-spec/viewer/) and a [Web UI](https://app.daster.tools.sap/ui5/) to generate tokens
 * required to record your consent and to authenticate. Please see the [documentation](https://github.wdf.sap.corp/pages/Security-Testing/doc/daster/) for
 * background information about the tool, its usage scenarios and channels to report problems.
 */
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this

    Map stepParams = PiperGoUtils.prepare(this, script).plus(parameters)
    List credentials = [
        [type: 'usernamePassword', id: 'oAuthCredentialsId', env: ['PIPER_clientId', 'PIPER_clientSecret']],
        [type: 'token', id: 'dasterTokenCredentialsId', env: ['PIPER_dasterToken']],
        [type: 'token', id: 'userCredentialsId', env: ['PIPER_user']],
        [type: 'usernamePassword', id: 'targetAuthCredentialsId', env: ['PIPER_dasterTargetUser', 'PIPER_dasterTargetPassword']]
    ]
    piperExecuteBin(stepParams, STEP_NAME, METADATA_FILE, credentials)
}
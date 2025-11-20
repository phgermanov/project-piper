import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.MapUtils
import com.sap.piper.internal.PiperGoUtils
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/sapCreateFosstarsReport.yaml'

//Metadata maintained in file resources/metadata/sapCreateFosstarsReport.yaml

void call(Map parameters = [:]) {
  handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()
    parameters.juStabUtils = null

    //TODO: reuse with piperExecuteBin
    Map stepParams = PiperGoUtils.prepare(this, script, utils).plus(parameters)
    stepParams.piperGoUtils.unstashPiperBin()
    utils.unstash('pipelineConfigAndTests')

    piperExecuteBin.prepareMetadataResource(script, METADATA_FILE)
    Map stepParameters = piperExecuteBin.prepareStepParameters(parameters)

    List credentialInfo = [[type: 'usernamePassword', id: 'gitHttpsCredentialsId', env: ['PIPER_username', 'PIPER_password']]]

    withEnv([
        "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(stepParameters)}",
        "PIPER_correlationID=${env.BUILD_URL}",
        "PIPER_sCMUrl=${scm?.userRemoteConfigs?.get(0)?.url}",
        "PIPER_branch=${env.CHANGE_BRANCH?:env.BRANCH_NAME}"
    ]) {
        String customDefaultConfig = piperExecuteBin.getCustomDefaultConfigsArg()
        String customConfigArg = piperExecuteBin.getCustomConfigArg(script)
        // get context configuration
        Map config
        piperExecuteBin.handleErrorDetails(STEP_NAME) {
            config = piperExecuteBin.getStepContextConfig(script, "./sap-piper", METADATA_FILE, customDefaultConfig, customConfigArg)
            echo "Context Config: ${config}"
        }
        piperExecuteBin.dockerWrapper(script, STEP_NAME, config){
            try {
                piperExecuteBin.credentialWrapper(config, credentialInfo){
                    sh "./sap-piper sapCreateFosstarsReport${customDefaultConfig}${customConfigArg}"
                }
            } finally {
                handleStepResults(STEP_NAME, false, false)
            }
        }
    }
  }
}

// Remove this method once piperExecuteBin is extended with
void handleStepResults(String stepName, boolean failOnMissingReports, boolean failOnMissingLinks) {
    String reportsFileName = "${stepName}_reports.json"
    def reportsFileExists = fileExists(reportsFileName)
    if (failOnMissingReports && !reportsFileExists) {
        error "Expected to find ${reportsFileName} in workspace but it is not there"
    } else if (reportsFileExists) {
        def reports = readJSON(file: reportsFileName)
        for (report in reports) {
            String target = report['target'] as String
            if (target != null && target.startsWith("./")) {
                // archiveArtifacts does not match any files when they start with "./",
                // even though that is a correct relative path.
                target = target.substring(2)
            }
            archiveArtifacts artifacts: target, allowEmptyArchive: !report['mandatory']
        }
    }

    String linksFileName = "${stepName}_links.json"
    def linksFileExists = fileExists(linksFileName)
    if (failOnMissingLinks && !linksFileExists) {
        error "Expected to find ${linksFileName} in workspace but it is not there"
    } else if (linksFileExists) {
        def links = readJSON(file: linksFileName)
        for (link in links) {
            if(link['scope'] == 'job') {
                removeJobSideBarLinks(link['target'])
                addJobSideBarLink(link['target'], link['name'], "images/24x24/graph.png")
            }
            addRunSideBarLink(link['target'], link['name'], "images/24x24/graph.png")
        }
    }
}

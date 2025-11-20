import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.JenkinsUtils
import com.sap.piper.internal.Notify

import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'writeInflux'
@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'artifactVersion',
    'customData',
    'customDataTags',
    'customDataMap',
    'customDataMapTags',
    'influxPrefix',
    'influxServer',
    'wrapInNode'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()

        // notify about deprecated step usage
        Notify.deprecatedStep(this, "influxWriteData", "removed", script?.commonPipelineEnvironment)
        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(
                artifactVersion: script.globalPipelineEnvironment.getArtifactVersion(),
                influxPrefix: script.globalPipelineEnvironment.getGithubOrg() && script.globalPipelineEnvironment.getGithubRepo()
                    ? "${script.globalPipelineEnvironment.getGithubOrg()}_${script.globalPipelineEnvironment.getGithubRepo()}"
                    : null
            )
            .mixin(parameters, PARAMETER_KEYS)
            .addIfNull('customData', script.globalPipelineEnvironment.getInfluxCustomData())
            .addIfNull('customDataTags', script.globalPipelineEnvironment.getInfluxCustomDataTags())
            .addIfNull('customDataMap', script.globalPipelineEnvironment.getInfluxCustomDataMap())
            .addIfNull('customDataMapTags', script.globalPipelineEnvironment.getInfluxCustomDataMapTags())
            .use()

        if (config.artifactVersion == null)  {
            //this takes care that terminated builds due to milestone-locking do not cause an error
            echo "no artifact version available -> exiting ${STEP_NAME} without writing data"
            return
        }

        echo """----------------------------------------------------------
Artifact version: ${config.artifactVersion}
Influx server: ${config.influxServer}
Influx prefix: ${config.influxPrefix}
InfluxDB data: ${config.customData}
InfluxDB data tags: ${config.customDataTags}
InfluxDB data map: ${config.customDataMap}
InfluxDB data map tags: ${config.customDataMapTags}
----------------------------------------------------------"""

        if(config.wrapInNode){
            node(''){
                try{
                    writeToInflux(config, utils, jenkinsUtils)
                }finally{
                    deleteDir()
                }
            }
        } else {
            writeToInflux(config, utils, jenkinsUtils)
        }
    }
}

private void writeToInflux(config, utils, jenkinsUtils){
    if (config.influxServer) {
        try {
            def influxPluginVersion = jenkinsUtils.getPlugin('influxdb').getVersion()
            def influxParams = [
                selectedTarget: config.influxServer,
                customPrefix: config.influxPrefix,
                customData: config.customData.size()>0 ? config.customData : null,
                customDataTags: config.customDataTags.size()>0 ? config.customDataTags : null,
                customDataMap: config.customDataMap.size()>0 ? config.customDataMap : null,
                customDataMapTags: config.customDataMapTags.size()>0 ? config.customDataMapTags : null
            ]

            if (!influxPluginVersion || influxPluginVersion.startsWith('1.')) {
                influxParams['$class'] = 'InfluxDbPublisher'
                step(influxParams)
            } else {
                influxDbPublisher(influxParams)
            }
        } catch (NullPointerException e){
             if(!e.getMessage()){
                 //TODO: catch NPEs as long as https://issues.jenkins-ci.org/browse/JENKINS-55594 is not fixed & released
                StringWriter writer = new StringWriter()
                e.printStackTrace(new PrintWriter(writer))
                error "[$STEP_NAME] NullPointerException occured, is the correct target defined?\n\n${writer.toString()}"
             }
             throw e
        }
    }

    // only print files when pipeline_data is available, i.e. typically when called at the end of the pipeline
    if (config.customDataMap.pipeline_data != null) {
        //write results into json file for archiving
        writeFile file: 'jenkins_data.json', text: utils.getPrettyJsonString(config.customData)
        writeFile file: 'pipeline_data.json', text: utils.getPrettyJsonString(config.customDataMap)
        writeFile file: 'jenkins_data_tags.json', text: utils.getPrettyJsonString(config.customDataTags)
        writeFile file: 'pipeline_data_tags.json', text: utils.getPrettyJsonString(config.customDataMapTags)
        archiveArtifacts artifacts: '*data*.json', allowEmptyArchive: true
    }

}

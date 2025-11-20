import com.sap.piper.internal.JenkinsUtils
import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'executePitTests'
@Field Set GENERAL_CONFIG_KEYS = [
    'globalSettingsFile'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'buildDescriptorFile',
    'dockerImage',
    'dockerWorkspace',
    'stashContent',
    'runOnlyScheduled',
    'coverageThreshold',
    'mutationThreshold'
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

        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName ?: env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        if (config.runOnlyScheduled == false || (jenkinsUtils.isJobStartedByTimer() && config.runOnlyScheduled == true)) {
            try {
                script.globalPipelineEnvironment.setInfluxStepData('pit', false)
                config.stashContent = utils.unstashAll(config.stashContent)
                dockerExecute(
                    script: script,
                    dockerImage: config.dockerImage,
                    dockerWorkspace: config.dockerWorkspace,
                    stashContent: config.stashContent
                ) {
                    def mvnOpts = ""
                    if (config.globalSettingsFile){
                        def globalSettingsFile = config.globalSettingsFile
                        if (globalSettingsFile.startsWith("http:") || globalSettingsFile.startsWith("https:")){
                            log("loading global settings file ${globalSettingsFile}")
                            globalSettingsFile = downloadSettingsFromUrl(script, globalSettingsFile, ".pipeline/mavenGlobalSettings.xml")
                        }
                        mvnOpts = "--global-settings ${globalSettingsFile}"
                    }

                    sh "mvn ${mvnOpts} --batch-mode --file ${config.buildDescriptorFile} -DtimestampedReports=false -DcoverageThreshold=${config.coverageThreshold} -DmutationThreshold=${config.mutationThreshold} clean process-test-classes org.pitest:pitest-maven:mutationCoverage -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn"
                }
            }
            finally {
                publishHtmlReport(config.pitHtml)
                script.globalPipelineEnvironment.setInfluxStepData('pit', true)
            }
        }
    }
}

void publishHtmlReport(Map settings = [:]) {
    if (settings.active) {
        if (settings.path && !settings.path.endsWith('/')) settings.path += '/'
        def pattern = "${settings.path}${settings.file}"
        def resultFiles = findFiles(glob: pattern)

        log("found ${resultFiles.length} file(s) to publish for '${pattern}")
        for (File file : resultFiles) {
            def fileDir = "${file.path.replaceAll(file.name, '')}"
            publishHTML([
                allowMissing         : settings.allowEmptyResults,
                alwaysLinkToLastBuild: true,
                keepAll              : true,
                reportDir            : "${fileDir}",
                reportFiles          : "${file.name}",
                reportName           : "${settings.name}"
            ])
        }
        // archive Pit results
        archiveResults(settings.archive, pattern, settings.allowEmptyResults)
    }
}

void archiveResults(archive, pattern, allowEmpty) {
    if (archive) {
        log("archive ${pattern}")
        archiveArtifacts artifacts: pattern, allowEmptyArchive: allowEmpty
    }
}

void log(msg) {
    echo "[${STEP_NAME}] ${msg}"
}

String downloadSettingsFromUrl(script, String url, String targetFile = 'settings.xml') {
    if (script.fileExists(targetFile)) {
        log("Global settings file ${targetFile} already exists. Skipping download from ${url}")
        return targetFile
    }

    def response = script.httpRequest(url)
    script.writeFile(file: targetFile, text: response.content)
    return targetFile
}

import com.sap.piper.internal.JenkinsUtils
import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Notify

import groovy.transform.Field
import groovy.text.GStringTemplateEngine

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'setupPipelineEnvironment'
@Field Set GENERAL_CONFIG_KEYS = [
    'githubApiUrl',
    'gitHttpsCredentialsId',
    /**
     * Defines the main branch for your pipeline. **Typically this is the `master` branch, which does not need to be set explicitly.** Only change this in exceptional cases. Supports regular expression through Groovy Match operator, e.g. `master|develop`.
     */
    'productiveBranch',
    /**
     * Defines if the "Security" and "IPScan and PPMS" stages run only scheduled according to the schedule defined in parameter 'nightlySchedule' in step setupPipelineEnvironment.
     * @possibleValues `true`, `false`
     */
    'pipelineOptimization'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'appContainer',
    'buildDiscarder',
    'configFile',
    'configYmlFile',
    'customDefaultsCredentialsId',
    'gitBranch',
    'gitCommitId',
    'gitHttpsUrl',
    'githubOrg',
    'githubRepo',
    'gitSshUrl',
    'nightlySchedule',
    'relatedLibraries',
    'runNightly',
    'storeGithubStatistics'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    'scmInfo'
])

void call(Map parameters = [:]) {
    library 'piper-lib-os'
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, failOnError: true,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()

        def gitRemoteUrl
        try {
            gitRemoteUrl = parameters.githubRepo ? null : utils.getGitRemoteUrl()
        } catch (e) {
            Notify.error(this, "git call to retrieve 'remote.origin.url' failed - step needs to run in a git context.")
        }

        // propagate custom defaults
        loadDefaultValues customDefaults: parameters.customDefaults

        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        // load yaml config
        script.globalPipelineEnvironment.configuration = getYamlConfig(config)

        //Load defaults including custom default files defined in the configuration file
        List customDefaultsFiles = script.globalPipelineEnvironment?.configuration?.customDefaults
        String customDefaultsCredentialsId = script.globalPipelineEnvironment?.configuration?.general?.customDefaultsCredentialsId
        if(customDefaultsFiles) {
            loadDefaultValues customDefaults: parameters.customDefaults, customDefaultsFromFiles: customDefaultsFiles, customDefaultsCredentialsId: customDefaultsCredentialsId
        }

        config = new ConfigurationHelper(config, STEP_NAME)
            .loadStepDefaults(this)
            .mixinHooksConfig()
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(
                gitSshUrl: (parameters.githubRepo || gitRemoteUrl?.startsWith('http')) ? null : gitRemoteUrl,
                githubOrg: parameters.githubRepo ? null : utils.getFolderFromGitUrl(gitRemoteUrl),
                githubRepo: parameters.githubRepo ? null : utils.getRepositoryFromGitUrl(gitRemoteUrl),
                gitBranch: env.BRANCH_NAME,
                gitCommitId: utils.getGitCommitId()
            )
            .mixin(parameters, PARAMETER_KEYS)
            // check mandatory parameters
            .withMandatoryProperty('githubOrg')
            .withMandatoryProperty('githubRepo')
            .use()

        // resolve templates
        config.gitSshUrl = GStringTemplateEngine.newInstance().createTemplate(config.gitSshUrl).make([githubOrg: config.githubOrg, githubRepo: config.githubRepo]).toString()
        config.gitHttpsUrl = GStringTemplateEngine.newInstance().createTemplate(config.gitHttpsUrl).make([githubOrg: config.githubOrg, githubRepo: config.githubRepo]).toString()

        loadRelatedLibraries(config)

        //get library information
        def sapPiperLib = 'piper-lib:unknown'
        def osPiperLib = 'piper-lib-os:unknown'
        def additionalLibs = ''
        def piperLibs = jenkinsUtils.getLibrariesInfo()
        piperLibs.each { lib ->
            if (lib.name == 'piper-lib') {
                sapPiperLib = "${lib.name}:${lib.version}"
            } else if (lib.name == 'piper-lib-os') {
                osPiperLib = "${lib.name}:${lib.version}"
            } else {
                additionalLibs += "${lib.name}:${lib.version}, "
            }
        }

        //check if piper-lib-os is available properly
        try {
            // init piper-lib-os with SAP Piper defaults
            if (parameters.customDefaults instanceof String)
                parameters.customDefaults = [parameters.customDefaults]
            // write "piper-defaults.yml" as "defaults.yaml"
            writeFile file: ".pipeline/defaults.yaml", text: libraryResource("piper-defaults.yml")

            List customDefaults = ['piper-defaults.yml'].plus(parameters.customDefaults?:[])
            script.setupCommonPipelineEnvironment([
                script: script,
                customDefaults: customDefaults,
                scmInfo: parameters.scmInfo
            ].plus(
                config.configYmlFile.endsWith('.yaml')?[configFile: config.configYmlFile]:[:]
            ))

            script.globalPipelineEnvironment.setFlag('piper-lib-os')
            script.globalPipelineEnvironment.cpe = script.commonPipelineEnvironment

            //map step configuration if not available
            mapStepConfiguration(script)

        } catch (NoSuchMethodError err) {
            osPiperLib = 'n/a'
            Notify.warning(this, "Open Source Piper library not available, please make sure to load 'piper-lib-os'. see https://go.sap.corp/piper/lib/setupLibrary/")
        }
        script.globalPipelineEnvironment.setGithubOrg(config.githubOrg)
        script.globalPipelineEnvironment.setGithubRepo(config.githubRepo)
        script.globalPipelineEnvironment.setGitBranch(config.gitBranch)
        script.globalPipelineEnvironment.setGitSshUrl(config.gitSshUrl)
        if (gitRemoteUrl != null && gitRemoteUrl.startsWith('https')){
            script.globalPipelineEnvironment.setGitHttpsUrl(gitRemoteUrl)
        }

        //properties when an appContainer is built - will not be exposed in configuration for now
        //appContainer is a container which contains an app previously built in the same pipeline
        if (config.appContainer) {
            script.globalPipelineEnvironment.setAppContainerProperty('githubOrg', config.appContainer.githubOrg)
            script.globalPipelineEnvironment.setAppContainerProperty('githubRepo', config.appContainer.githubRepo)
            script.globalPipelineEnvironment.setAppContainerProperty('gitBranch', config.appContainer.gitBranch)
            script.globalPipelineEnvironment.setAppContainerProperty('gitSshUrl', config.appContainer.gitSshUrl?gitSshUrl:"git@github.wdf.sap.corp:${config.appContainer.githubOrg}/${config.appContainer.githubRepo}.git")
        }
        script.globalPipelineEnvironment.setInfluxCustomDataProperty('build_result', 'SUCCESS')
        script.globalPipelineEnvironment.setInfluxCustomDataProperty('build_result_key', 1)
        script.globalPipelineEnvironment.setInfluxStepData('build_url', env.BUILD_URL)
        script.globalPipelineEnvironment.setInfluxPipelineData('build_url', env.BUILD_URL)

        if (config.storeGithubStatistics) {
            script.globalPipelineEnvironment.setGithubStatistics(script, config.githubApiUrl, config.githubOrg, config.githubRepo, config.gitCommitId, config.gitHttpsCredentialsId)
        }

        if (config.buildDiscarder)
            jenkinsUtils.addBuildDiscarder(config.buildDiscarder.daysToKeep, config.buildDiscarder.numToKeep, config.buildDiscarder.artifactDaysToKeep, config.buildDiscarder.artifactNumToKeep)


        sapPipelineInit(script: script, isScheduled: jenkinsUtils.pipelineIsScheduled())
        def isOptimized = script.commonPipelineEnvironment.getValue('pipelineOptimization')

        script.commonPipelineEnvironment.setValue('scheduledRun', jenkinsUtils.pipelineIsScheduled())
        script.commonPipelineEnvironment.setValue('hooks', config.hooks)

        config.pipelineOptimization = parameters.script.commonPipelineEnvironment.getStepConfiguration('', '').pipelineOptimization
        echo "Pipeline is scheduled: ${parameters.script.commonPipelineEnvironment.getValue('scheduledRun')}"
        echo "Pipeline schedule security: ${config.pipelineOptimization}"

        // do not allow optimized runs with pipeline resiliency switched on
        // we will check step configuration for Security stage to cover most cases
        if (config.pipelineOptimization && !script.commonPipelineEnvironment.getStepConfiguration('handlePipelineStepErrors', 'Security').failOnError) {
            Notify.error(this, "failOnError: false not supported for optimized pipeline runs, please change your configuration")
        }

        // only allow scheduling for productiveBranch, NOT for PRs, feature branches, etc.
        if (env.BRANCH_NAME ==~ config.productiveBranch) {
            if (config.runNightly || config.pipelineOptimization) {
                jenkinsUtils.scheduleJob(config.nightlySchedule)
            } else {
                // remove all schedules in case all scheduling is disabled
                // this makes sure that the job does not run accidentally
                jenkinsUtils.removeJobSchedule()
            }
        }
    }
}

private Map getYamlConfig(Map config){
    Map configMap = [:]
    if (fileExists(config.configYmlFile)) {
        try {
            configMap = readYaml(file: config.configYmlFile)
        } catch (e) {
            Notify.error(this, "Invalid yml configuration in ${config.configYmlFile} - ${e}")
        }
    } else if (fileExists(config.configYmlFile.replace('.yml','.yaml'))) {
        config.configYmlFile = config.configYmlFile.replace('.yml','.yaml')
        try {
            configMap = readYaml(file: config.configYmlFile)
        } catch (e) {
            Notify.error(this, "Invalid yml configuration in ${config.configYmlFile} - ${e}")
        }
    }
    return configMap
}

private void loadRelatedLibraries(config) {
    config.relatedLibraries.each {lib ->
        try {
            library lib
            echo "[$STEP_NAME] Loaded related library '${lib}'"
        } catch (err) {
            echo "[$STEP_NAME] Failed to load related library '${lib}'"
        }

    }
}

private void mapStepConfiguration(script) {
    if (!script.commonPipelineEnvironment.configuration?.steps) return
    def stepMapping = readJSON text: libraryResource('piperOsStepMapping.json')
    stepMapping.each { osStepName, internalStepName ->
        if (!script.commonPipelineEnvironment.configuration.steps[osStepName]) {
            script.commonPipelineEnvironment.configuration.steps[osStepName] = script.globalPipelineEnvironment.configuration.steps[internalStepName]
        }
    }
}

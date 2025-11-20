import com.cloudbees.groovy.cps.NonCPS

import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.MapUtils
import com.sap.piper.internal.Notify

import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'publishCheckResults'
@Field Set TOOLS = ['aggregation','checkstyle','cpd','eslint','findbugs','fiori','pmd','pylint','tasks']
@Field Set GENERAL_CONFIG_KEYS = TOOLS.plus([
    'archive'
])
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        // notify about deprecated step usage
        Notify.deprecatedStep(this, "checksPublishResults", "removed", script?.commonPipelineEnvironment)
        // harmonize parameters
        prepare(parameters)
        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()
        // fix path issues with pmd & cpd
        fixFilePathsInReport(config)
        // JAVA
        report('PmdPublisher', config.pmd, config.archive)
        report('DryPublisher', config.cpd, config.archive)
        report('FindBugsPublisher', config.findbugs, config.archive)
        report('CheckStylePublisher', config.checkstyle, config.archive)
        // JAVA SCRIPT
        reportWarnings('JSLint', config.eslint, config.archive)
        reportWarnings('Fiori Analysis Parser', config.fiori, config.archive)
        // PYTHON
        reportWarnings('PyLint', config.pylint, config.archive)
        // GENERAL
        reportTasks(config.tasks)
        aggregateReports(config.aggregation)
    }
}

void aggregateReports(settings){
    if (settings.active) {
        def options = createCommonOptionsMap('AnalysisPublisher', settings)
        // publish
        step(options)
    }
}

void reportTasks(settings){
    if (settings.active) {
        def options = createCommonOptionsMap('TasksPublisher', settings)
        options.put('pattern', settings.get('pattern'))
        options.put('high', settings.get('high'))
        options.put('normal', settings.get('normal'))
        options.put('low', settings.get('low'))
        // publish
        step(options)
    }
}

void report(publisherName, settings, doArchive){
    if (settings.active) {
        def pattern = settings.get('pattern')
        def options = createCommonOptionsMap(publisherName, settings)
        options.put('pattern', pattern)
        // publish
        step(options)
        // archive check results
        archiveResults(doArchive && settings.get('archive'), pattern, true)
    }
}

void reportWarnings(parserName, settings, doArchive){
    if (settings.active) {
        def pattern = settings.get('pattern')
        def options = createCommonOptionsMap('WarningsPublisher', settings)
        options.put('parserConfigurations', [[
            parserName: parserName,
            pattern: pattern
        ]])
        // publish
        step(options)
        // archive check results
        archiveResults(doArchive && settings.get('archive'), pattern, true)
    }
}

void archiveResults(archive, pattern, allowEmpty){
  if(archive){
    echo("[${STEP_NAME}] archive ${pattern}")
    archiveArtifacts artifacts: pattern, allowEmptyArchive: allowEmpty
  }
}

void fixFilePathsInReport(config){
    def reportFileList = []
    def search = '\\/data\\/xmake\\/.*\\/gen\\/tmp\\/src\\/'
    def replacement = ''

    if(config.pmd.active) reportFileList.push('pmd.xml')
    if(config.cpd.active) reportFileList.push('cpd.xml')

    for(String reportFile : reportFileList)
        try {
            sh "find . -type f -name '${reportFile}' -exec sed -i 's/${search}/${replacement}/g' {} \\;"
        } catch (ignore) {}
}

@NonCPS
Map createCommonOptionsMap(publisherName, settings){
    Map result = [:]
    def thresholds = settings.get('thresholds', [:])
    def fail = thresholds.get('fail', [:])
    def unstable = thresholds.get('unstable', [:])

    result.put('$class', publisherName)
    result.put('healthy', settings.get('healthy'))
    result.put('unHealthy', settings.get('unHealthy'))
    result.put('canRunOnFailed', true)
    result.put('failedTotalAll', fail.get('all')?.toString())
    result.put('failedTotalHigh', fail.get('high')?.toString())
    result.put('failedTotalNormal', fail.get('normal')?.toString())
    result.put('failedTotalLow', fail.get('low')?.toString())
    result.put('unstableTotalAll', unstable.get('all')?.toString())
    result.put('unstableTotalHigh', unstable.get('high')?.toString())
    result.put('unstableTotalNormal', unstable.get('normal')?.toString())
    result.put('unstableTotalLow', unstable.get('low')?.toString())
    // filter empty values
    result = result.findAll {
        return it.value != null && it.value != ''
    }
    return result
}

Map prepare(Map parameters){
    // ensure tool maps are initialized correctly
    for(String tool : TOOLS){
        parameters[tool] = toMap(parameters[tool])
    }
    return parameters
}

Map toMap(parameter){
    if(MapUtils.isMap(parameter)) {
        parameter.put('active', parameter.active == null?true:parameter.active)
    } else if(Boolean.TRUE.equals(parameter)) {
        parameter = [active: true]
    } else if(Boolean.FALSE.equals(parameter)) {
        parameter = [active: false]
    } else {
        parameter = [:]
    }
    return parameter
}

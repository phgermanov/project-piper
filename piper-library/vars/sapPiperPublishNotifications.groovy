import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.JenkinsUtils
import com.sap.piper.internal.Notify

import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field def STEP_NAME = getClass().getName()

void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        // notify about deprecated step usage
        Notify.deprecatedStep(this, "piperPublishWarnings")
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        // notify about deprecated step usage
        Notify.deprecatedStep(this, "piperPublishWarnings", "removed", script?.commonPipelineEnvironment)

        Map piperNotificationsSettings = [
            parserName: 'Piper Notifications Parser',
            parserLinkName: 'Piper Notifications',
            parserTrendName: 'Piper Notifications',
            parserRegexp: '\\[(INFO|WARNING|ERROR)\\] (.*) \\(([^) ]*)\\/([^) ]*)\\)',
            parserExample: ''
        ]
        piperNotificationsSettings.parserScript = '''import hudson.plugins.warnings.parser.Warning
        import hudson.plugins.analysis.util.model.Priority

        Priority priority = Priority.LOW
        String message = matcher.group(2)
        String libraryName = matcher.group(3)
        String stepName = matcher.group(4)
        String fileName = 'Jenkinsfile'

        switch(matcher.group(1)){
            case 'WARNING': priority = Priority.NORMAL; break;
            case 'ERROR': priority = Priority.HIGH; break;
        }

        return new Warning(fileName, 0, libraryName, stepName, message, priority);
        '''

        // add Piper Notifications parser to config if missing
        if(JenkinsUtils.addWarningsParser(script, piperNotificationsSettings)){
            echo "[${STEP_NAME}] New Warnings plugin parser '${piperNotificationsSettings.parserName}' configuration added."
        }

        node(){
            try{
                // parse log for Piper Notifications
                warnings(canRunOnFailed: true, consoleParsers: [[ parserName: piperNotificationsSettings.parserName ]])
            }finally{
                deleteDir()
            }
        }
    }
}

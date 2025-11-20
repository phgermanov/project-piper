package com.sap.piper.internal

import com.sap.icd.jenkins.Utils

import groovy.text.GStringTemplateEngine

class Notify implements Serializable {
    private static enum Severity { ERROR, WARNING }
    private final static String LIBRARY_NAME = 'piper-lib'
    private final static String MESSAGE_PATTERN = '[${severity}] ${message} (${libName}/${stepName})'

    protected static Utils instance = null

    protected static Utils getUtilsInstance(){
        instance = instance ?: new Utils()
        return instance
    }

    static void deprecatedStep(Script step, String successor = null, String deprecationState = null, pipelineEnv = null){
        def msg = "The step ${step.STEP_NAME} is deprecated."
        if (successor)
            msg += " Please use it's successor (see https://go.sap.corp/piper/steps/${successor}/)."

        switch(deprecationState) {
            case "removed":
                msg += " This step has been removed, for details see: https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues/2844"
                error(step, msg)
            break
            case "announced":
                msg += " This step will be removed by the end of the quarter, for details see: https://github.wdf.sap.corp/ContinuousDelivery/piper-library/issues/2844"
                if(!pipelineEnv?.configuration?.general?.acknowledgeDeprecatedStepUsage){
                    List unstableSteps = pipelineEnv?.getValue('unstableSteps') ?: []
                    // add information about unstable steps to pipeline environment
                    // this helps to bring this information to users in a consolidated manner inside a pipeline
                    unstableSteps.add(step.STEP_NAME)
                    pipelineEnv?.setValue('unstableSteps', unstableSteps)
                    step.unstable(msg)
                }
                warning(step, msg)
            break
            default:
                warning(step, msg)
            break
        }
    }

    static void warning(Script step, String msg, String stepName = null){
        log(step, msg, stepName)
    }

    static void error(Script step, String msg, String stepName = null) {
        log(step, msg, stepName, Severity.ERROR)
    }

    private static void log(Script step, String msg, String stepName, Severity severity = Severity.WARNING){
        stepName = stepName ?: step.STEP_NAME
        def logEntry = GStringTemplateEngine.newInstance().createTemplate(
            MESSAGE_PATTERN
        ).make([
            libName: LIBRARY_NAME,
            stepName: stepName,
            message: msg,
            severity: severity
        ]).toString()

        if (severity == Severity.ERROR){
            step.error(logEntry)
        }
        step.echo(logEntry)
    }
}

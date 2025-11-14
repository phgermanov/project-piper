@Library('piper-lib') _

import com.sap.piper.internal.ConfigurationHelper

import groovy.text.SimpleTemplateEngine
import groovy.transform.Field
import hudson.model.Run
import org.jenkinsci.plugins.workflow.cps.GlobalVariable

@Field String STEP_NAME = ''

node {
    stage('Generate Docs') {
        deleteDir()

        git 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library.git'

        loadDefaultValues()

        List stepNames = []

        if (params.STEPNAME) {
            //respect documentation creation for an individual step - passed as parameter STEPNAME
            stepNames.add(params.STEPNAME)
        } else {
            //create documentation for all steps
            def files = findFiles(glob: 'vars/*.groovy')
            files.each {file ->
                stepNames.add(file.name.substring(0, file.name.indexOf('.')))
            }
        }
        echo "Creting documentation for following steps: ${stepNames}"

        createRawStepDocumentation(getStepDetails(stepNames))

        zip zipFile: 'stepDocs.zip', archive: true, dir: './stepDocs'
    }
}

def createRawStepDocumentation(stepDetails) {
    def docTemplate = '''# ${stepName}

## Description

## Prerequisites

## Example

## Parameters

${parameterTable}

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).  

In following sections the configuration is possible:

${stepConfiguration}

'''
    sh 'mkdir stepDocs'
    stepDetails.each{step ->
        String parameterTable = '''| parameter | mandatory | default | possible values |
| ----------|-----------|---------|-----------------|
'''
        String stepConfiguration = '''| parameter | general | step | stage |
| ----------|-----------|---------|-----------------|
'''

        step.getValue().doc.each {stepParam ->
            parameterTable += "|${stepParam.getKey()}|${stepParam.getValue().isMandatory}|${stepParam.getValue().defaults}||\n"

            String stepActive = step.getValue().STEP_CONFIG_KEYS != null ? (step.getValue().STEP_CONFIG_KEYS.contains(stepParam.getKey()) ? 'X' : '') : ''
            def stepScriptFileContent = readFile file: "vars/${step.getKey()}.groovy"
            String generalActive = ''
            if (stepScriptFileContent.contains('mixinGeneralConfig')) {
                generalActive = step.getValue().GENERAL_CONFIG_KEYS.size() > 0 ? (step.getValue().GENERAL_CONFIG_KEYS.contains(stepParam.getKey()) ? 'X' : '') : stepActive
            }
            String stageActive = stepActive
            stepConfiguration += "|${stepParam.getKey()}|${generalActive}|${stepActive}|${stageActive}|\n"
        }
        String stepDocFileContent = SimpleTemplateEngine.newInstance().createTemplate(docTemplate).make([stepName: step.getKey(), parameterTable: parameterTable, stepConfiguration: stepConfiguration]).toString()
        writeFile file: "stepDocs/${step.getKey()}.md", text: stepDocFileContent

    }

}

def getStepDetails(stepNames) {
    def stepParams = getStepParameters(stepNames)

    stepParams.each {step ->
        def stepKeys = stepParams[step.getKey()]
        def stepScriptFileContent = readFile file: "vars/${step.getKey()}.groovy"
        def stepName = step.getKey()
        stepParams[stepName].doc = [:]

        //add default information for script
        stepParams[stepName].doc.script = [
                defaults: '',
                isMandatory: 'yes',
                description: 'defines the global script environment of the Jenkinsfile run. Typically `this` is passed to this parameter. This allows the function to access the [`globalPipelineEnvironment`](../objects/globalPipelineEnvironment.md) for retrieving e.g. configuration parameters.'
        ]

        stepKeys.ALL_KEYS.each {paramName ->
            stepParams[stepName].doc[paramName] =  [:]
            stepParams[stepName].doc[paramName].description = getParameterDescription(stepScriptFileContent, paramName)
            stepParams[stepName].doc[paramName].isMandatory = isParameterMandatory(stepScriptFileContent, paramName)
            stepParams[stepName].doc[paramName].defaults = getParameterDefaults(this, stepScriptFileContent, stepName, paramName)
            //ToDo: how to handle possible values?
        }
    }
}

def getParameterDescription(stepScriptFileContent, paramName) {
    //ToDo: implement retrieval of parameter details
    //Parameter Details:
    //add parameter descriptions from groovy file if available as comment after parameter name, like <paramName>, // Parameter description
}


String isParameterMandatory(stepScriptFileContent, paramName) {
    return stepScriptFileContent.contains(".withMandatoryProperty('${paramName}')")? 'yes' : 'no'
}

def getParameterDefaults(script, stepScriptFileContent, stepName, paramName) {
    STEP_NAME = stepName
    Map config = ConfigurationHelper.loadStepDefaults(script).use()
    if (config[paramName] != null) {
        if (config[paramName] instanceof List) {
            return getListHtml(config[paramName])
        }
        config[paramName] ?: "''"
        return "`${config[paramName]}`"
    } else {
        //default could still be inside a sub-structre merged with dependingOn
        def dependingOnList = stepScriptFileContent.findAll(/dependingOn\(.*\).mixin\(.*\)/)
        String deepDefaults = ''
        dependingOnList.each {dependingOn ->
            if (dependingOn.contains(paramName)) {
                def subKey = dependingOn.substring(dependingOn.indexOf("'") + 1, dependingOn.indexOf("')"))
                config.each {item ->
                    if (item.getValue() instanceof Map && item.getValue().containsKey(paramName)) {
                        deepDefaults += "${subKey}=`${item.getKey()}`: "
                        if (item.getValue()[paramName] instanceof List) {
                            deepDefaults += getListHtml(item.getValue()[paramName])
                        } else {
                            deepDefaults += "`${item.getValue()[paramName]}`<br />"
                        }
                    }
                }
            }
        }
        return deepDefaults
    }

}

String getListHtml(myList) {
    if (myList.size() == 0) return '`[]`'
    def returnValue = '<ul>'
    myList.each {item ->
        returnValue += "<li>`${item}`</li>"
    }
    returnValue += '</ul>'
    return returnValue
}


Map getStepParameters(stepNames) {
    Map stepParams = [:]

    def globals = []
    for (GlobalVariable var : GlobalVariable.forRun(currentBuild.rawBuild instanceof Run ? (Run) currentBuild.rawBuild : null)) {
        globals.add(var)
        if (var.getName() in stepNames) {
            def stepScript = var.getValue(this)
            def stepName = var.getName()

            stepParams[stepName] = [:]

            stepParams[stepName].PARAMETER_KEYS = getConfigKeys(stepScript, 'PARAMETERS')
            stepParams[stepName].GENERAL_CONFIG_KEYS = getConfigKeys(stepScript, 'GENERAL')
            stepParams[stepName].STEP_CONFIG_KEYS = getConfigKeys(stepScript, 'STEP')

            stepParams[stepName].ALL_KEYS = stepParams[stepName].PARAMETER_KEYS + stepParams[stepName].GENERAL_CONFIG_KEYS + stepParams[stepName].STEP_CONFIG_KEYS
            stepParams[stepName].ALL_KEYS = stepParams[stepName].ALL_KEYS.sort()
        }
    }
    return stepParams
}

Map getConfigKeys(stepScript, type) {

    def configKeys
    try {
        switch (type) {
            case 'STEP':
                configKeys = stepScript.STEP_CONFIG_KEYS
                break
            case 'PARAMETERS':
                configKeys = stepScript.PARAMETER_KEYS
                break
            case 'GENERAL':
                configKeys = stepScript.GENERAL_CONFIG_KEYS
                break
            default:
                configKeys = []
        }
    } catch (err) {
        configKeys = []
    }
    return configKeys
}

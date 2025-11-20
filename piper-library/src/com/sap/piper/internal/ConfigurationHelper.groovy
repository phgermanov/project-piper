package com.sap.piper.internal

import hudson.AbortException

class ConfigurationHelper implements Serializable {
    static def loadStepDefaults(Script step){
        return new ConfigurationHelper(step)
            .initDefaults(step)
            .loadDefaults(step)
    }

    private Map config = [:]
    private String name
    private dependingOn
    private enforced

    ConfigurationHelper(Script step){
        name = step.STEP_NAME
        if(!name) throw new IllegalArgumentException('Step has no public STEP_NAME property!')
    }

    private final ConfigurationHelper initDefaults(Script step){
        step.loadDefaultValues()
        return this
    }

    private final ConfigurationHelper loadDefaults(Script step){
        config = ConfigurationLoader.defaultGeneralConfiguration()
        mixin(ConfigurationLoader.defaultStepConfiguration(null, name))
        mixin(ConfigurationLoader.defaultStageConfiguration(null, step.env.STAGE_NAME))
        return this
    }

    ConfigurationHelper mixinGeneralConfig(globalPipelineEnvironment, Set filter = null){
        Map stepConfiguration = ConfigurationLoader.generalConfiguration([globalPipelineEnvironment: globalPipelineEnvironment])
        return mixin(stepConfiguration, filter)
    }

    ConfigurationHelper mixinHooksConfig(){
        Map hooksConfiguration = ConfigurationLoader.defaultHooksConfiguration()
        return mixin(hooksConfiguration)
    }

    ConfigurationHelper mixinStageConfig(globalPipelineEnvironment, stageName, Set filter = null){
        Map stageConfiguration = ConfigurationLoader.stageConfiguration([globalPipelineEnvironment: globalPipelineEnvironment], stageName)
        return mixin(stageConfiguration, filter)
    }

    ConfigurationHelper mixinStepConfig(globalPipelineEnvironment, Set filter = null){
        if(!name) throw new IllegalArgumentException('Step has no public STEP_NAME property!')
        Map stepConfiguration = ConfigurationLoader.stepConfiguration([globalPipelineEnvironment: globalPipelineEnvironment], name)
        return mixin(stepConfiguration, filter)
    }

    ConfigurationHelper mixin(Map parameters, Set filter = null){
        config = ConfigurationMerger.merge(parameters, filter, config)
        return this
    }

    ConfigurationHelper mixin(String key){
        def dependentValue = config[dependingOn]
        if((config[key] == null || enforced) && dependentValue && config[dependentValue] && config[dependentValue][key] != null)
            config[key] = config[dependentValue][key]

        dependingOn = null
        enforced = null
        return this
    }

    ConfigurationHelper dependingOn(dependentKey, enforce=false){
        dependingOn = dependentKey
        enforced = enforce
        return this
    }

    ConfigurationHelper addIfEmpty(key, value){
        if (config[key] instanceof Boolean) {
            return this
        } else if (!config[key]){
            config[key] = value
        }
        return this
    }

    ConfigurationHelper addIfNull(key, value){
        if (config[key] instanceof Boolean) {
            return this
        } else if (config[key] == null){
            config[key] = value
        }
        return this
    }

    Map use(){ return config }

    ConfigurationHelper(Map config = [:], String stepName = null){
        this.config = config
        this.name = stepName
    }

    def getConfigProperty(key) {
        if (config[key] != null && config[key].class == String) {
            return config[key].trim()
        }
        return config[key]
    }

    def getConfigProperty(key, defaultValue) {
        def value = getConfigProperty(key)
        if (value == null) {
            return defaultValue
        }
        return value
    }

    def isPropertyDefined(key){

        def value = getConfigProperty(key)

        if(value == null){
            return false
        }

        if(value.class == String){
            return value?.isEmpty() == false
        }

        if(value){
            return true
        }

        return false
    }

    def getMandatoryProperty(key, defaultValue = null) {

        def paramValue = config[key]

        if (paramValue == null)
            paramValue = defaultValue

        if (paramValue == null)
            throw new AbortException("ERROR - NO VALUE AVAILABLE FOR ${key}")
        return paramValue
    }

    def withMandatoryProperty(key) {
        getMandatoryProperty(key)
        return this
    }

    def withMandatoryPropertyUponCondition(key, condition) {
        if(condition(this.config)) {
            getMandatoryProperty(key)
        }
        return this
    }
}

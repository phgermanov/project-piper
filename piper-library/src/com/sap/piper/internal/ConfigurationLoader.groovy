package com.sap.piper.internal


class ConfigurationLoader implements Serializable {
    def static stepConfiguration(script, String stepName) {
        return loadConfiguration(script, 'steps', stepName, ConfigurationType.CUSTOM_CONFIGURATION)
    }

    def static stageConfiguration(script, String stageName) {
        return loadConfiguration(script, 'stages', stageName, ConfigurationType.CUSTOM_CONFIGURATION)
    }

    def static generalConfiguration(script){
        return script?.globalPipelineEnvironment?.configuration?.general ?: [:]
    }

    def static defaultStepConfiguration(script, String stepName) {
        return loadConfiguration(script, 'steps', stepName, ConfigurationType.DEFAULT_CONFIGURATION)
    }

    def static defaultStageConfiguration(script, String stageName) {
        return loadConfiguration(script, 'stages', stageName, ConfigurationType.DEFAULT_CONFIGURATION)
    }

    def static defaultGeneralConfiguration(){
        return DefaultValueCache.getInstance()?.getDefaultValues()?.general ?: [:]
    }

    def static defaultHooksConfiguration(){
        def hooks = DefaultValueCache.getInstance()?.getDefaultValues()?.hooks
        if (hooks == null) {
            return [:]
        }
        return ['hooks': hooks]
    }

    private static loadConfiguration(script, String type, String entryName, ConfigurationType configType){
        switch (configType) {
            case ConfigurationType.CUSTOM_CONFIGURATION:
                return script?.globalPipelineEnvironment?.configuration?.get(type)?.get(entryName) ?: [:]
            case ConfigurationType.DEFAULT_CONFIGURATION:
                return DefaultValueCache.getInstance()?.getDefaultValues()?.get(type)?.get(entryName) ?: [:]
            default:
                throw new IllegalArgumentException("Unknown configuration type: ${configType}")
        }
    }
}

package com.sap.piper.internal

class ConfigurationMerger {
    def static merge(Map configs, Set configKeys, Map defaults) {
        Map filteredConfig = configKeys?configs.subMap(configKeys):configs

        return MapUtils.merge(
            MapUtils.pruneNulls(defaults),
            MapUtils.pruneNulls(filteredConfig)
        )
    }

    def static merge(
        Map parameters, Set parameterKeys,
        Map configuration, Set configurationKeys,
        Map defaults=[:]
    ){
        Map merged
        merged = merge(configuration, configurationKeys, defaults)
        merged = merge(parameters, parameterKeys, merged)
        return merged
    }

    def static merge(
        def script, def stepName,
        Map parameters, Set parameterKeys,
        Map pipelineData,
        Set stepConfigurationKeys
    ) {
        Map stepDefaults = ConfigurationLoader.defaultStepConfiguration(script, stepName)
        Map stepConfiguration = ConfigurationLoader.stepConfiguration(script, stepName)

        mergeWithPipelineData(parameters, parameterKeys, pipelineData ?: [:], stepConfiguration, stepConfigurationKeys, stepDefaults)
    }

    def static mergeWithPipelineData(Map parameters, Set parameterKeys,
                            Map pipelineDataMap,
                            Map configurationMap, Set configurationKeys,
                            Map stepDefaults=[:]
    ){
        Map merged
        merged = merge(configurationMap, configurationKeys, stepDefaults)
        merged = merge(pipelineDataMap, null, merged)
        merged = merge(parameters, parameterKeys, merged)

        return merged
    }
}

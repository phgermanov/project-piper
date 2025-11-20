import com.sap.piper.internal.DefaultValueCache
import com.sap.piper.internal.MapUtils

def call(Map parameters = [:]) {
    if (!DefaultValueCache.getInstance() || parameters.customDefaults || parameters.customDefaultsFromFiles) {
        Map defaultValues = [:]
        def configFileListFromResources = ['piper-defaults.yml']
        def configFileListFromFiles = parameters.customDefaultsFromFiles ?: []
        def customDefaults = parameters.customDefaults

        if (customDefaults in String) customDefaults = [customDefaults]
        if (customDefaults in List) configFileListFromResources += customDefaults
        for (def configFileName : configFileListFromResources) {
            echo "Loading library configuration file '${configFileName}'"
            def configuration = readYaml text: libraryResource(configFileName)
            defaultValues = MapUtils.merge(
                MapUtils.pruneNulls(defaultValues),
                MapUtils.pruneNulls(configuration)
            )
        }
        for (String configFileName : configFileListFromFiles) {
            String customDefaultsCredentialsId = parameters.customDefaultsCredentialsId
            defaultValues = MapUtils.merge(
                MapUtils.pruneNulls(defaultValues),
                MapUtils.pruneNulls(downloadCustomDefaultValues(configFileName, customDefaultsCredentialsId))
            )
        }
        DefaultValueCache.createInstance(defaultValues)
    }
}


Map downloadCustomDefaultValues(String configFileName, String credentialsId) {
    if (configFileName.startsWith('http://') || configFileName.startsWith('https://')) {

        Map httpRequestParameters = [
            url               : configFileName,
            validResponseCodes: '100:399,404' // Allow a more specific error message for 404 case
        ]
        if (credentialsId) {
            httpRequestParameters.authentication = credentialsId
        }
        def response = httpRequest(httpRequestParameters)
        if (response.status == 404) {
            error "URL for remote custom defaults (${customDefaults[i]}) appears to be incorrect. " +
                "Server returned HTTP status code 404. " +
                "Please make sure that the path is correct and no authentication is required to retrieve the file."
        }

        return readYaml(text: response.content)
    } else if (fileExists(configFileName)) {
        return readYaml(file: configFileName)
    }
    echo "WARNING: Custom default entry not found: '${configFileName}', it will be ignored"
    return [:]

}

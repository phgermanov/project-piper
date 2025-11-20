import com.sap.icd.jenkins.Utils
import com.sap.icd.jenkins.Sirius
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.ConfigurationLoader
import com.sap.piper.internal.Deprecate
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'siriusUploadDocument'
@Field Set GENERAL_CONFIG_KEYS = [
    'confidential',
    'documentName',
    'fileName',
    'siriusApiUrl',
    'siriusCredentialsId',
    'siriusDeliveryName',
    'siriusDocumentFamily',
    'siriusProgramName',
    'siriusTaskGuid',
    'siriusUploadUrl'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        // handle deprecated parameters
        Deprecate.parameter(this, parameters, 'credentialsId', 'siriusCredentialsId')
        Deprecate.parameter(this, parameters, 'apiUrl', 'siriusApiUrl')
        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixin(readInfoFromDeliveryMapping(ConfigurationLoader.defaultStepConfiguration(script, 'sapCreateTraceabilityReport').deliveryMappingFile))
            .mixin(readInfoFromDeliveryMapping(script.globalPipelineEnvironment.configuration?.steps?.sapCreateTraceabilityReport?.deliveryMappingFile))
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('siriusCredentialsId')
            .withMandatoryProperty('fileName')
            .withMandatoryProperty('siriusDeliveryName')
            .withMandatoryProperty('siriusProgramName')
            .withMandatoryProperty('siriusTaskGuid')
            .use()

        echo "[${STEP_NAME}] Using credentialsId: '${config.siriusCredentialsId}', apiUrl: ${config.siriusApiUrl}"

        def sirius = parameters.siriusStub
        if (sirius == null) {
            sirius = new Sirius(script, config.siriusCredentialsId, config.siriusApiUrl, config.siriusUploadUrl)
        }

        def deliveryGuid = sirius.getDeliveryExtGuidByName(config.siriusProgramName, config.siriusDeliveryName)
        sirius.uploadDocument(deliveryGuid, config.siriusTaskGuid, config.fileName, config.documentName, config.siriusDocumentFamily)
    }
}

private Map readInfoFromDeliveryMapping (String fileName) {

    deliveryInfo = [:]
    if (fileName && fileExists(fileName)) {
        def deliveryMapping = readJSON file: fileName
        if (deliveryMapping.sirius_program) deliveryInfo.siriusProgramName = deliveryMapping.sirius_program
        if (deliveryMapping.sirius_delivery) deliveryInfo.siriusDeliveryName = deliveryMapping.sirius_delivery
    }
    return deliveryInfo
}

import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'sapPiperStagePerformance'

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * **Alpha**: Activate a native build (e.g. maven, npm) in combination with SAP's staging service.
     * This is not yet product standard compliant (SLC-29). Efforts are on the way to provide this as alternative option to building on xMake within Q2/2021.
     * (default value: `false`)
     * @possibleValues 'true', 'false'
     */
    'nativeBuild',
    /*
     secretID, groupID to call systemTrust
    */
    'vaultAppRoleSecretTokenCredentialsId',
    'vaultBasePath',
    'vaultPipelineName'
]
@Field STAGE_STEP_KEYS = [
    /** For Cloud Foundry use-cases: Creates CF Services based on information in yml-format.*/
    'cloudFoundryCreateService',
    /** For Cloud Foundry use-cases: Sets up the required test environment based on a infrastructure definition in yml-format. */
    'manageCloudFoundryEnvironment',
    /** For Cloud Foundry use-cases: Performs deployment to Cloud Foundry space/org. */
    'cloudFoundryDeploy',
    /** For Docker/Kubernetes use-cases: Performs deployment to Kubernetes cluster. */
    'kubernetesDeploy',
    /** For non-Docker use-cases, downloads artifact from Nexus which should be deployed. */
    'downloadArtifactsFromNexus',
    /**
     * **Active for scenarios where no deployment to Kubernetes is performed** AND `nativeBuild: true` unless explicitly deactivated with configuration `sapDownloadArtifact: false`.<br />
     * Downloads artifact from Nexus/Artifactory which should be deployed.
     */
    'sapDownloadArtifact',
    /** End-to-end performance tests using Gatling tool. */
    'gatlingExecuteTests',
    /** Single user performance tests using SUPA. */
    'sapSUPAExecuteTests',
    /**
     * Only Cloud Foundry: Performs health check in order to prove one aspect of operational readiness.
     * In order to be able to respond to health checks from infrastructure components (like load balancers) it is important to provide one unprotected application endpoint which allows a judgement about the health of your application.
     * For Kubernetes cases this is not required since it is typically done via Kubernetes means (liveness/readiness probes).
     */
    'healthExecuteCheck',
    /** Performs upload of result files of the previous steps of this stage to Cumulus. */
    'sapCumulusUpload',
    /** Executes Terraform to apply the desired state of configuration. */
    'terraformExecute',
    /** Publishes test results to Jenkins. It will automatically be active in cases tests are executed. */
    'testsPublishResults',
    /** Updates Kubernetes Deployment Manifest in an Infrastructure Git Repository. */
    'gitopsUpdateDeployment',
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Defines if artifact should be downloaded from Nexus staging environment. **Only change default of `true` in exceptional cases!**
     * @possibleValues `true`, `false`
     */
    'fromStaging'
].plus(STAGE_STEP_KEYS))
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    def stageName = parameters.stageName?:env.STAGE_NAME

    Map config = ConfigurationHelper
        .loadStepDefaults(this)
        .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.globalPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('cloudFoundryDeploy', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.cloudFoundryDeploy)
        .addIfEmpty('kubernetesDeploy', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.kubernetesDeploy)
        .addIfEmpty('downloadArtifactsFromNexus', true)
        .addIfEmpty('healthExecuteCheck', true)
        .addIfEmpty('sapDownloadArtifact', true)
        .addIfEmpty('testsPublishResults', false)
        .addIfEmpty('cloudFoundryCreateService', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.cloudFoundryCreateService)
        .addIfEmpty('manageCloudFoundryEnvironment', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.manageCloudFoundryEnvironment)
        .addIfEmpty('gatlingExecuteTests', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.gatlingExecuteTests)
        .addIfEmpty('sapCumulusUpload', script.commonPipelineEnvironment.getStepConfiguration('sapCumulusUpload', stageName)?.pipelineId ? true : false )
        .addIfEmpty('terraformExecute', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.terraformExecute)
        .addIfEmpty('gitopsUpdateDeployment', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.gitopsUpdateDeployment)
        .addIfEmpty('fromStaging', true)
        .use()

    piperStageWrapper (script: script, stageName: stageName) {
        echo "Getting System trust token"
        def token = null
        try {
            def apiURLFromDefaults = script.commonPipelineEnvironment.getValue("hooks")?.systemtrust?.serverURL ?: ''
            token = sapGetSystemTrustToken(apiURLFromDefaults, config.vaultAppRoleSecretTokenCredentialsId, config.vaultPipelineName, config.vaultBasePath)
        } catch (Exception e) {
            echo "Couldn't get system trust token, will proceed with configured credentials: ${e.message}"
        }
        wrap([$class: 'MaskPasswordsBuildWrapper', varPasswordPairs: [[password: token]]]) {
            withEnv([
                /*
                Additional logic "?: ''" is necessary to ensure the environment
                variable is set to an empty string if the value is null
                Without this, the environment variable would be set to the string "null",
                causing checks for an empty token in the Go application to fail.
                */
                "PIPER_systemTrustToken=${token ?: ''}",
            ]) {

                if (config.cloudFoundryCreateService) {
                    cloudFoundryCreateService script: script
                }

                if (config.manageCloudFoundryEnvironment) {
                    durationMeasure(script: script, measurementName: 'envsetup_perf_duration') {
                        manageCloudFoundryEnvironment script: script
                    }
                }

                if (config.terraformExecute) {
                    terraformExecute script: script
                }

                if (config.gitopsUpdateDeployment) {
                    gitopsUpdateDeployment script: script
                }

                if (!config.kubernetesDeploy) {
                    if (config.nativeBuild && config.sapDownloadArtifact) {
                        sapDownloadArtifact script: script, fromStaging: true
                    } else if (!config.nativeBuild && config.downloadArtifactsFromNexus) {
                        downloadArtifactsFromNexus script: script, fromStaging: config.fromStaging
                    }
                }

                if (config.cloudFoundryDeploy) {
                    durationMeasure(script: script, measurementName: 'deploy_perf_duration') {
                        cloudFoundryDeploy script: script
                        if (config.healthExecuteCheck) {
                            healthExecuteCheck script: script
                        }
                    }
                } else if (config.kubernetesDeploy) {
                    if (script.commonPipelineEnvironment.getValue('helmChartUrl') && config.sapDownloadArtifact) {
                        sapDownloadArtifact script: script, fromStaging: true
                    }
                    durationMeasure(script: script, measurementName: 'deploy_perf_duration') {
                        kubernetesDeploy script: script
                    }
                }

                def publishResults = config.testsPublishResults

                if (config.sapSUPAExecuteTests) {
                    durationMeasure(script: script, measurementName: 'supa_duration') {
                        publishResults = true
                        sapSUPAExecuteTests script: script
                    }
                }

                if (config.gatlingExecuteTests) {
                    durationMeasure(script: script, measurementName: 'gatling_duration') {
                        publishResults = true
                        gatlingExecuteTests script: script
                    }
                }

                if (publishResults) {
                    testsPublishResults script: script
                }

                if(config.sapCumulusUpload) {
                    sapCumulusUpload script: script, stepResultType: 'e2e-test'
                }
            }
        }
    }
}

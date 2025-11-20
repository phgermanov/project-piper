import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'sapPiperStageAcceptance'

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
    /** Can perform both deployments to cloud foundry and neo targets. Preferred over cloudFoundryDeploy and neoDeploy, if configured. */
    'multicloudDeploy',
    /** For Cloud Foundry use-cases: Performs deployment to Cloud Foundry space/org. */
    'cloudFoundryDeploy',
    /** For Docker/Kubernetes use-cases: Performs deployment to Kubernetes cluster. */
    'kubernetesDeploy',
    /**
     * **Active for scenarios where no deployment to Kubernetes is performed** unless explicitly deactivated with configuration `downloadArtifactsFromNexus: false`.<br />
     * Downloads artifact from Nexus which should be deployed.
     */
    'downloadArtifactsFromNexus',
    /**
     * **Active for scenarios where no deployment to Kubernetes is performed** AND `nativeBuild: true` unless explicitly deactivated with configuration `sapDownloadArtifact: false`.<br />
     * Downloads artifact from Nexus/Artifactory which should be deployed.
     */
    'sapDownloadArtifact',
    /** Performs behavior-driven tests using Gauge test framework against the deployed application/service. */
    'gaugeExecuteTests',
    /**
     * Only Cloud Foundry: Performs health check in order to prove one aspect of operational readiness.
     * In order to be able to respond to health checks from infrastructure components (like load balancers) it is important to provide one unprotected application endpoint which allows a judgement about the health of your application.
     * For Kubernetes cases this is not required since it is typically done via Kubernetes means (liveness/readiness probes).
     */
    'healthExecuteCheck',
    /** For Cloud Foundry use-cases: Creates CF Services based on information in yml-format.*/
    'cloudFoundryCreateService',
    /** For Cloud Foundry use-cases: Sets up the required test environment based on a infrastructure definition in yml-format. */
    'manageCloudFoundryEnvironment',
    /** Performs API testing using Newman against the deployed application/service. */
    'newmanExecute',
    /** Generate traceability report (mapping of software requirements against tests). This step helps to ensure compliance to SAP's corporate requirement _FC-2_. */
    'sapCreateTraceabilityReport',
    /** Executes Terraform to apply the desired state of configuration. */
    'terraformExecute',
    /** Publishes test results to Jenkins. It will automatically be active in cases tests are executed. */
    'testsPublishResults',
    /** Performs end-to-end UI testing using UIVeri5 test framework against the deployed application/service. */
    'uiVeri5ExecuteTests',
    /** Executes end to end tests by running the npm script 'ci-e2e' defined in the project's package.json file. */
    'npmExecuteEndToEndTests',
    /** Executes npm scripts to run frontend unit tests.<br />
     * If custom names for the npm scripts are configured via the `runScripts` parameter the step npmExecuteScripts needs **explicit activation via stage configuration**. */
    'npmExecuteScripts',
    /** Performs upload of result files of the previous steps of this stage to Cumulus. */
    'sapCumulusUpload',
    /** Perform accessibility related tests as part of acceptance stage. It is disabled by default*/
    'sapAccessContinuumExecuteTests',
    /** Updates Kubernetes Deployment Manifest in an Infrastructure Git Repository. */
    'gitopsUpdateDeployment',
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Defines if artifact should be downloaded from Nexus staging environment. **Only change default in exceptional cases!**
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
        .addIfEmpty('multicloudDeploy', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.multicloudDeploy)
        .addIfEmpty('cloudFoundryDeploy', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.cloudFoundryDeploy)
        .addIfEmpty('kubernetesDeploy', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.kubernetesDeploy)
        .addIfEmpty('testsPublishResults', false)
        .addIfEmpty('downloadArtifactsFromNexus', true)
        .addIfEmpty('healthExecuteCheck', true)
        .addIfEmpty('sapDownloadArtifact', true)
        .addIfEmpty('gaugeExecuteTests', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.gaugeExecuteTests)
        .addIfEmpty('newmanExecute', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.newmanExecute)
        .addIfEmpty('uiVeri5ExecuteTests', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.uiVeri5ExecuteTests)
        .addIfEmpty('npmExecuteEndToEndTests', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.npmExecuteEndToEndTests)
        .addIfEmpty('npmExecuteScripts', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.npmExecuteScripts)
        .addIfEmpty('fromStaging', true)
        .addIfEmpty('cloudFoundryCreateService', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.cloudFoundryCreateService)
        .addIfEmpty('manageCloudFoundryEnvironment', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.manageCloudFoundryEnvironment)
        .addIfEmpty('sapCreateTraceabilityReport', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.sapCreateTraceabilityReport)
        .addIfEmpty('sapCumulusUpload', script.commonPipelineEnvironment.getStepConfiguration('sapCumulusUpload', stageName)?.pipelineId ? true : false )
        .addIfEmpty('terraformExecute', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.terraformExecute)
        .addIfEmpty('gitopsUpdateDeployment', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.gitopsUpdateDeployment)
        .addIfEmpty('sapAccessContinuumExecuteTests', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.sapAccessContinuumExecuteTests)
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

                if (config.multicloudDeploy) {
                    durationMeasure(script: script, measurementName: 'deploy_test_duration') {
                        Closure downloadArtifacts = {
                            downloadArtifactsFromNexus script: script, fromStaging: config.fromStaging
                        }
                        multicloudDeploy script: script, preDeploymentHook: downloadArtifacts
                    }
                } else {
                    if (config.cloudFoundryCreateService) {
                        cloudFoundryCreateService script: script
                    }

                    if (config.manageCloudFoundryEnvironment) {
                        durationMeasure(script: script, measurementName: 'envsetup_test_duration') {
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
                        durationMeasure(script: script, measurementName: 'deploy_test_duration') {
                            cloudFoundryDeploy script: script
                            if (config.healthExecuteCheck) {
                                healthExecuteCheck script: script
                            }
                        }
                    } else if (config.kubernetesDeploy) {
                        if (script.commonPipelineEnvironment.getValue('helmChartUrl') && config.sapDownloadArtifact) {
                            sapDownloadArtifact script: script, fromStaging: true
                        }
                        durationMeasure(script: script, measurementName: 'deploy_test_duration') {
                            kubernetesDeploy script: script
                        }
                    }
                }

                def publishMap = [script: script]
                def publishResults = config.testsPublishResults

                if (config.newmanExecute) {
                    durationMeasure(script: script, measurementName: 'newman_duration') {
                        publishResults = true
                        newmanExecute script: script
                    }
                }

                if (config.uiVeri5ExecuteTests) {
                    durationMeasure(script: script, measurementName: 'uiveri5_duration') {
                        publishResults = true
                        uiVeri5ExecuteTests script: script
                    }
                }

                if (config.gaugeExecuteTests) {
                    durationMeasure(script: script, measurementName: 'gauge_duration') {
                        publishResults = true
                        gaugeExecuteTests script: script
                        publishMap += [gauge: [archive: true]]
                    }
                }

                if (config.npmExecuteEndToEndTests) {
                    durationMeasure(script: script, measurementName: 'npmExecuteEndToEndTests_duration') {
                        npmExecuteEndToEndTests script: script, stageName: stageName, runScript: 'ci-e2e'
                    }
                }

                if (config.npmExecuteScripts) {
                    durationMeasure(script: script, measurementName: 'npmExecuteScripts_duration') {
                        publishResults = true
                        npmExecuteScripts script: script
                    }
                }

                if (publishResults) {
                    if(config.sapCumulusUpload) {
                        sapCumulusUpload script: script, filePattern: '**/TEST-*.xml', stepResultType: 'acceptance-test'
                    }
                    testsPublishResults publishMap
                }

                if (config.sapCreateTraceabilityReport) {
                    try{
                        sapCreateTraceabilityReport script: script
                    }finally{
                        if(config.sapCumulusUpload) {
                            sapCumulusUpload script: script, filePattern: '**/piper_traceability*', stepResultType: 'traceability-report'
                        }
                    }
                }

                if(config.sapCumulusUpload) {
                    sapCumulusUpload script: script, filePattern: '**/requirement.mapping', stepResultType: 'requirement-mapping'
                    sapCumulusUpload script: script, filePattern: '**/delivery.mapping', stepResultType: 'delivery-mapping'
                }

                if(config.sapAccessContinuumExecuteTests) {
                    sapAccessContinuumExecuteTests script: script
                }
            }
        }
    }
}

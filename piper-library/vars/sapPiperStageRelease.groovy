import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'sapPiperStageRelease'

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * **Alpha**: Activate a native build (e.g. maven, npm) in combination with SAP's staging service.
     * This is not yet product standard compliant (SLC-29). Efforts are on the way to provide this as alternative option to building on xMake within Q2/2021.
     * (default value: `false`)
     * @possibleValues 'true', 'false'
     */
    'nativeBuild',
    'artifactType',
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
    /**
     * Only Cloud Foundry and Neo Deploy: Performs health check in order to prove one aspect of operational readiness.
     * In order to be able to respond to health checks from infrastructure components (like load balancers) it is important to provide one unprotected application endpoint which allows a judgement about the health of your application.
     * For Kubernetes cases this is not required since it is typically done via Kubernetes means (liveness/readiness probes).
     */
    'healthExecuteCheck',
    /** Publishes release information on GitHub. */
    'githubPublishRelease',
    /** For Neo use-cases: Performs deployment to SAP Cloud Platoform Neo. */
    'neoDeploy',
    /** Executes Terraform to apply the desired state of configuration. */
    'terraformExecute',
    /** Perform an upload to Deploy with Confidence */
    'sapDwCStageRelease',
    /** Collect DevOps metrics including DORA */
    'sapCollectInsights',
    /** Updates Kubernetes Deployment Manifest in an Infrastructure Git Repository. */
    'gitopsUpdateDeployment',
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Optional skip of checkout if checkout was done before this step already.
     * @possibleValues `true`, `false`
     */
    'skipCheckout'
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
        .addIfEmpty('healthExecuteCheck', true)
        .addIfEmpty('downloadArtifactsFromNexus', true)
        .addIfEmpty('sapDownloadArtifact', true)
        .addIfEmpty('cloudFoundryCreateService', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.cloudFoundryCreateService)
        .addIfEmpty('manageCloudFoundryEnvironment', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.manageCloudFoundryEnvironment)
        .addIfEmpty('githubPublishRelease', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.githubPublishRelease || (script.globalPipelineEnvironment.configuration?.steps?.githubPublishRelease? true : false))
        .addIfEmpty('neoDeploy', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.neoDeploy)
        .addIfEmpty('terraformExecute', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.terraformExecute)
        .addIfEmpty('sapDwCStageRelease', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.sapDwCStageRelease)
        .addIfEmpty('gitopsUpdateDeployment', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.gitopsUpdateDeployment)
        .addIfEmpty('sapCollectInsights', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.sapCollectInsights)
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
                    durationMeasure(script: script, measurementName: 'deploy_release_duration') {
                        Closure downloadArtifacts = {
                            downloadArtifactsFromNexus script: script, fromStaging: config.fromStaging
                        }
                        multicloudDeploy script: script, preDeploymentHook: downloadArtifacts
                    }
                } else {
                    def skipCheckout = config.skipCheckout
                    if (skipCheckout != null && !(skipCheckout instanceof Boolean)) {
                        error "[${STEP_NAME}] Parameter skipCheckout has to be of type boolean. Instead got '${skipCheckout.class.getName()}'"
                    }
                    if (config.sapDwCStageRelease && !skipCheckout) {
                        checkout scm
                    }
                    if (config.cloudFoundryCreateService) {
                        cloudFoundryCreateService script: script
                    }

                    if (config.manageCloudFoundryEnvironment) {
                        durationMeasure(script: script, measurementName: 'envsetup_release_duration') {
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
                            sapDownloadArtifact script: script, fromStaging: false
                        } else if (!config.nativeBuild && config.downloadArtifactsFromNexus) {
                            downloadArtifactsFromNexus script: script, fromStaging: config.fromStaging
                        }
                    }

                    if (config.neoDeploy) {
                        durationMeasure(script: script, measurementName: 'deploy_release_duration') {
                            neoDeploy script: script
                        }
                    }

                    if (config.cloudFoundryDeploy) {
                        durationMeasure(script: script, measurementName: 'deploy_release_duration') {
                            cloudFoundryDeploy script: script
                        }
                    } else if (config.kubernetesDeploy) {
                        if (script.commonPipelineEnvironment.getValue('helmChartUrl') && config.sapDownloadArtifact) {
                            sapDownloadArtifact script: script, fromStaging: false
                        }
                        durationMeasure(script: script, measurementName: 'deploy_release_duration') {
                            kubernetesDeploy script: script
                        }
                    }

                    if (config.sapDwCStageRelease) {
                        durationMeasure(script: script, measurementName: 'deploy_release_duration') {
                            sapDwCStageRelease script: script
                        }
                    }
                }

                if (config.healthExecuteCheck) {
                    if (config.multicloudDeploy || config.neoDeploy || config.cloudFoundryDeploy) {
                        healthExecuteCheck script: script
                    } else {
                        echo "healthExecuteCheck is skipped because none of following steps are active: multicloudDeploy, neoDeploy, cloudFoundryDeploy"
                    }
                }

                if (config.sapCollectInsights){
                    try {
                        sapCollectInsights script: script
                    } catch (err) {
                        echo "Dora reporting exceptional: ${err}"
                    }
                }

                if (config.githubPublishRelease) {
                    githubPublishRelease script: script
                }
            }
        }
    }
}

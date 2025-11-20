import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Deprecate
import com.sap.piper.internal.DockerUtils
import com.sap.piper.internal.Notify
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'pushToDockerRegistry'
@Field Set GENERAL_CONFIG_KEYS = [
    'dockerBuildImage',
    'dockerCredentialsId',
    'dockerImage',
    'dockerRegistryUrl',
    'sourceImage',
    'sourceRegistryUrl',
    'tagLatest',
    'tagArtifactVersion',
    'skopeoImage'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'dockerArchive'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()

        // notify about deprecated step usage
        Notify.deprecatedStep(this, "containerPushToRegistry") //, "announced", script?.commonPipelineEnvironment)
        // handle deprecated parameters
        Deprecate.parameter(this, parameters, 'dockerCredentials', 'dockerCredentialsId')
        Deprecate.parameter(this, parameters, 'dockerRegistry', 'dockerRegistryUrl')
        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(
                dockerBuildImage: script.globalPipelineEnvironment.getDockerBuildImage(),
                sourceImage: script.globalPipelineEnvironment.getAppContainerDockerMetadata()?.imageNameTag?:script.globalPipelineEnvironment.getDockerMetadata().imageNameTag,
                artifactVersion: script.globalPipelineEnvironment.getArtifactVersion()
            )
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('dockerCredentialsId')
            .withMandatoryProperty('dockerRegistryUrl')
            .addIfEmpty('skopeoImage', 'piper.int.repositories.cloud.sap/piper/skopeo')
            .use()

        DockerUtils dockerUtils = new DockerUtils(this)

        if (config.sourceRegistryUrl) {
            config.sourceRegistry = dockerUtils.getRegistryFromUrl(config.sourceRegistryUrl)
        } else {
            //calculate sourceRegistryUrl based on information set by xMake build into globalPipelineEnvironment
            config.sourceRegistry = script.globalPipelineEnvironment.getAppContainerDockerMetadata()?.repo?:script.globalPipelineEnvironment.getDockerMetadata().repo
            config.sourceRegistryUrl = config.sourceRegistry ? "${config.sourceRegistryProtocol}://${config.sourceRegistry}" : null
        }

        if (!config.dockerImage)
            config.dockerImage = config.sourceImage

        if (dockerUtils.withDockerDeamon()) {

            //ToDo: evaluate if option dockerBuildImage can be removed, if not make dockerImage mandatory if no dockerBuildImage is available, otherwise we will run into a NullPointerException here!
            config.dockerBuildImage = config.dockerBuildImage?:docker.image(config.dockerImage)
            new ConfigurationHelper(config)
                .withMandatoryProperty('dockerBuildImage')

            if (config.sourceRegistry && config.sourceImage) {

                def sourceBuildImage = docker.image(config.sourceImage)
                docker.withRegistry(config.sourceRegistryUrl) {
                    sourceBuildImage.pull()
                }
                sh "docker tag ${config.sourceRegistry}/${config.sourceImage} ${config.dockerImage}"
            }

            docker.withRegistry(
                config.dockerRegistryUrl,
                config.dockerCredentialsId
            ) {
                config.dockerBuildImage.push()
                if (config.tagLatest)
                    config.dockerBuildImage.push('latest')
                if (config.tagArtifactVersion )
                    config.dockerBuildImage.push(config.artifactVersion)
            }
        } else {
            //handling for Kubernetes case
            dockerExecute(
                script: script,
                // if the skopeoImage is not set, the default value is used (see line 54)
                dockerImage: config.skopeoImage
            ) {
                //since no Docker deamon is available we can only push from a Docker tar archive or move from one registry to another
                if (config.dockerArchive) {
                    //to be implemented later - not relevant yet
                } else {
                    dockerUtils.moveImage([image: config.sourceImage, registryUrl: config.sourceRegistryUrl], [image: config.dockerImage, registryUrl: config.dockerRegistryUrl, credentialsId: config.dockerCredentialsId])
                    if (config.tagLatest) {
                        def dockerImageName = dockerUtils.removeTagFromImageName(config.dockerImage)
                        def dockerImageNameLatest = dockerImageName + ':latest'

                        dockerUtils.moveImage([image: config.sourceImage, registryUrl: config.sourceRegistryUrl], [image: dockerImageNameLatest, registryUrl: config.dockerRegistryUrl, credentialsId: config.dockerCredentialsId])
                    }
                    if (config.tagArtifactVersion){
                        def dockerImageName = dockerUtils.removeTagFromImageName(config.dockerImage)
                        def dockerImageNameArtifactVersion =  dockerImageName + ':' + config.artifactVersion

                        dockerUtils.moveImage([image: config.sourceImage, registryUrl: config.sourceRegistryUrl], [image: dockerImageNameArtifactVersion, registryUrl: config.dockerRegistryUrl, credentialsId: config.dockerCredentialsId])
                    }
                }
            }
        }
    }
}

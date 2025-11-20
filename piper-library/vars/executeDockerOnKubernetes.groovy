import com.sap.piper.internal.JenkinsUtils
import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Notify

import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'executeDockerOnKubernetes' // needed for ConfigurationHelper
@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'dindImage',
    'dockerImage',
    'dockerWorkspace',
    'dockerEnvVars',
    'stashContent',
    'stashBackConfig',
    'skipStashBack',
    'uniqueId'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

def call(Map parameters = [:], body) {
    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()
    def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()
    // notify about deprecated step usage
    Notify.deprecatedStep(this, "dockerExecuteOnKubernetes", "removed", script?.commonPipelineEnvironment)
    // load default & individual configuration
    Map config = ConfigurationHelper
        .loadStepDefaults(this)
        .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
        .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
        .mixin(
            uniqueId: UUID.randomUUID().toString(),
            isDeprecatedPlugin: jenkinsUtils.isOldKubePluginVersion()
        )
        .mixin(parameters, PARAMETER_KEYS)
        .use()

    if (!config.dockerImage)
        Notify.error(this, "Docker image not specified.")

    def options = [
        name: env.jaas_owner + '-jaas',
        label: config.uniqueId,
        containers: getContainerList(config)
    ]
    if(!config.isDeprecatedPlugin)
        options.put('nodeUsageMode', 'EXCLUSIVE')

    stashWorkspace(config)
    try {
        podTemplate(options) {
            node(config.uniqueId) {
                try {
                    echo "Execute container content in Kubernetes pod"
                    utils.unstashAll(config.stashContent)
                    container(name: 'container-exec') {
                        body()
                    }
                }
                finally {
                    stashContainer(config)
                }
            }
        }
    }
    finally {
        unstashContainer(config)
    }
}

private stashWorkspace(config){
    if (config.stashContent.size() == 0) {
        try {
            sh "chmod -R u+w ."
            stash name: "workspace-${config.uniqueId}", useDefaultExcludes: false
            config.stashContent += 'workspace-' + config.uniqueId
        } catch (hudson.AbortException e) {
            echo "${e.getMessage()}"
        } catch (java.io.IOException ioe) {
            echo "${ioe.getMessage()}"
        }
    }
}

private stashContainer(config){
    if (config.skipStashBack){
        echo "skip stash container"
        return;
    }
    def stashBackConfig = config.stashBackConfig
    try {
        stashBackConfig.name = "container-${config.uniqueId}"
        stash stashBackConfig
    } catch (hudson.AbortException e) {
        echo "${e.getMessage()}"
    } catch (java.io.IOException ioe) {
        echo "${ioe.getMessage()}"
    }
}

private unstashContainer(config){
    if (config.skipStashBack){
        echo "skip unstash container"
        return;
    }
    try {
        unstash "container-${config.uniqueId}"
    } catch (hudson.AbortException e) {
        echo "${e.getMessage()}"
    } catch (java.io.IOException ioe) {
        echo "${ioe.getMessage()}"
    }
}

private getContainerList(config){
    def envVars
    /**
     * Check kubernetes-plugin version for backwards compatibility.
     * Impacting changes from version 1.0+.
     */
    if(config.isDeprecatedPlugin){
        /**
         * @Deprecated
         * Configuration for usage of kubernetes-plugin 0.12, 0.11, etc.
         */
        envVars = getContainerEnvsLegacy(config.dockerEnvVars, config.dockerWorkspace)
        if(config.dindImage) {
            envVars <<  containerEnvVar(key: 'DOCKER_HOST', value: '2375')
        }
    }else{
        /**
         * Configuration for usage of kubernetes-plugin 1.0+
         */
        envVars = getContainerEnvs(config.dockerEnvVars, config.dockerWorkspace)
        if(config.dindImage) {
            envVars <<  envVar(key: 'DOCKER_HOST', value: '2375')
        }
    }

    result = []
    result.push(containerTemplate(
        name: 'jnlp',
        image: 'docker.wdf.sap.corp:50001/sap-production/jnlp-alpine:4.3.4-sap-01',
        args: '${computer.jnlpmac} ${computer.name}'
    ))
    result.push(containerTemplate (
        name: 'container-exec',
        image: config.dockerImage,
        alwaysPullImage: true,
        command: '/usr/bin/tail -f /dev/null',
        envVars: envVars
    ))
    if(config.dindImage)
        result.push(containerTemplate(
            name: 'container-dind',
            image: config.dindImage,
            privileged: true
        ))
    return result
}

/**
 * Returns a list of envVar object consisting of set
 * environment variables, params (Parametrized Build) and working directory.
 * (Kubernetes-Plugin only!)
 * @param dockerEnvVars Map with environment variables
 * @param dockerWorkspace Path to working dir
 */
private getContainerEnvs(dockerEnvVars, dockerWorkspace) {
    def containerEnv = []

    if (dockerEnvVars) {
        for (String k: dockerEnvVars.keySet()) {
            containerEnv << envVar(key: k, value: dockerEnvVars[k].toString())
        }
    }
    if (params) {
        for (String k: params.keySet()) {
            containerEnv << envVar(key: k, value: params[k].toString())
        }
    }
    if (dockerWorkspace) containerEnv << envVar(key: "HOME", value: dockerWorkspace)
    // ContainerEnv array can't be empty. Using a stub to avoid failure.
    if (!containerEnv) containerEnv << envVar(key: "EMPTY_VAR", value: " ")

    return containerEnv
}

/**
 @Deprecated use for backwards compatibility (kubernetes-plugin version < 1.0)
 */
private getContainerEnvsLegacy(dockerEnvVars, dockerWorkspace) {
    def containerEnv = []

    if (dockerEnvVars) {
        for (String k: dockerEnvVars.keySet()) {
            containerEnv << containerEnvVar(key: k, value: dockerEnvVars[k].toString())
        }
    }
    if (params) {
        for (String k: params.keySet()) {
            containerEnv << containerEnvVar(key: k, value: params[k].toString())
        }
    }
    if (dockerWorkspace) containerEnv << containerEnvVar(key: "HOME", value: dockerWorkspace)
    // ContainerEnv array can't be empty. Using a stub to avoid failure.
    if (!containerEnv) containerEnv << containerEnvVar(key: "EMPTY_VAR", value: " ")

    return containerEnv
}

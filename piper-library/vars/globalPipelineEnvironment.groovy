import com.sap.piper.internal.ConfigurationLoader
import com.sap.piper.internal.ConfigurationMerger
import com.sap.piper.internal.Notify
import org.jenkinsci.plugins.workflow.actions.LabelAction
import org.jenkinsci.plugins.workflow.graph.FlowNode
import org.jenkinsci.plugins.workflow.steps.FlowInterruptedException

class globalPipelineEnvironment implements Serializable {
    private Map xMakeProperties = [:]
    private Map dockerMetadata = [:]
    Map configuration = [:]

    private String artifactVersion
    private String buildResult = 'SUCCESS'

    private String githubOrg
    private String githubRepo

    private String gitBranch
    private String gitSshUrl
    private String gitHttpsUrl
    private String gitCommitId
    private String gitCommitMessage

    private String nexusLastDownloadUrl
    private String nexusStagingFilePath
    private String nexusStagingGroup
    private String nexusPromoteRepositories

    String versionBeforeAutoVersioning

    private cpe

    private Map appContainerProperties = [:]
    // possible values
    /*
    [
        githubOrg: '',
        githubRepo: '',
        gitBranch: '',
        gitSshUrl: '',
        gitHttpsUrl: '',
        artifactVersion: '',
        gitCommitId: '',
        xMakeProperties: [:]
    ]
    */

    private dockerBuildImage

    private Map influxCustomData = [:]
    private Map influxCustomDataTags = [:]
    private Map influxCustomDataMap = [pipeline_data: [:], step_data: [:]]
    private Map influxCustomDataMapTags = [pipeline_data: [:]]

    private Map flags = [:]

    def reset() {
        influxCustomData = [:]
        influxCustomDataTags = [:]
        influxCustomDataMap = [pipeline_data: [:], step_data: [:]]
        influxCustomDataMapTags = [pipeline_data: [:]]
        dockerBuildImage = null
        appContainerProperties = [:]
        xMakeProperties = [:]
        dockerMetadata = [:]

        configuration = [:]
        flags = [:]

        artifactVersion = null
        githubOrg = null
        githubRepo = null
        gitBranch = null
        gitSshUrl = null
        gitHttpsUrl = null
        gitCommitId = null
        gitCommitMessage = null
        nexusStagingFilePath = null
        buildResult = 'SUCCESS'
        nexusStagingGroup=null
        nexusPromoteRepositories=null
        versionBeforeAutoVersioning = null
    }

    def setAppContainerProperty(property, value) {
        appContainerProperties[property] = value
        //piper-lib-os compatibility
        cpe?.setAppContainerProperty(property, value)
    }

    def getAppContainerProperty(property) {
        return appContainerProperties[property]
    }

    def setGithubOrg(org) {
        githubOrg = org
        //piper-lib-os compatibility
        cpe?.setGithubOrg(org)
    }
    def getGithubOrg() {
        return githubOrg
    }
    def setGithubRepo(repo) {
        githubRepo = repo
        //piper-lib-os compatibility
        cpe?.setGithubRepo(repo)
    }
    def getGithubRepo() {
        return githubRepo
    }
    def setGitBranch(branch) {
        gitBranch = branch
        //piper-lib-os compatibility
        cpe?.setGitBranch(branch)
    }
    def getGitBranch() {
        return gitBranch
    }
    def setGitSshUrl(url) {
        gitSshUrl = url
        //piper-lib-os compatibility
        cpe?.setGitSshUrl(url)
    }
    def getGitSshUrl() {
        return gitSshUrl
    }
    def setGitHttpsUrl(url) {
        gitHttpsUrl = url
        //piper-lib-os compatibility
        cpe?.setGitHttpsUrl(url)
    }
    def getGitHttpsUrl() {
        return gitHttpsUrl
    }
    def setXMakeProperties(map) {
        xMakeProperties = map
    }
    def getXMakeProperties() {
        return xMakeProperties
    }
    def setXMakeProperty(property, value) {
        xMakeProperties[property] = value
    }
    def getXMakeProperty(property) {
        return xMakeProperties[property]
    }
    def setDockerMetadata(map) {
        dockerMetadata = map
        dockerMetadata.imageNameTag = map.tag_name.split('/', 2)[1]
        dockerMetadata.imageName = dockerMetadata.imageNameTag.split(':', 2)[0]
        dockerMetadata.imageTag = dockerMetadata.imageNameTag.split(':', 2)[1]
        dockerMetadata.imageRegistry = map.tag_name.split('/', 2)[0]

        // set commonPipelineEnvironment container properties
        cpe?.setContainerProperty('image', map.image_name) // full image name incl. registry
        cpe?.setContainerProperty('imageNameTag', map.image_name.split('/', 2)[1]) // image name & tag without registry
        cpe?.setContainerProperty('registryUrl', "https://${map.repo}") // registry https url
    }
    def getDockerMetadata() {
        return dockerMetadata
    }
    def setAppContainerDockerMetadata(map) {
        appContainerProperties.dockerMetadata = map
        appContainerProperties.dockerMetadata.imageNameTag = map.tag_name.split('/', 2)[1]
        appContainerProperties.dockerMetadata.imageName = appContainerProperties.dockerMetadata.imageNameTag.split(':', 2)[0]
        appContainerProperties.dockerMetadata.imageTag = appContainerProperties.dockerMetadata.imageNameTag.split(':', 2)[1]
        appContainerProperties.dockerMetadata.imageRegistry = map.tag_name.split('/', 2)[0]
    }
    def getAppContainerDockerMetadata() {
        return appContainerProperties.dockerMetadata
    }
    def setNexusStagingFilePath(path) {
         nexusStagingFilePath = path
    }
    def getNexusStagingFilePath() {
        return nexusStagingFilePath
    }
    def setArtifactVersion(version) {
        artifactVersion = version
        //piper-lib-os compatibility
        cpe?.setArtifactVersion(version)
    }
    def getArtifactVersion() {
        return artifactVersion
    }
    def setGitCommitId(commitId) {
        gitCommitId = commitId
        //piper-lib-os compatibility
        cpe?.setGitCommitId(commitId)
    }
    def getGitCommitId() {
        return gitCommitId
    }
    def setGitCommitMessage(commitMessage) {
        gitCommitMessage = commitMessage
        //piper-lib-os compatibility
        if (commitMessage)
            cpe?.setGitCommitMessage(commitMessage)
    }
    def getGitCommitMessage() {
        return gitCommitMessage
    }
    def setBuildResult(result) {
        buildResult = result
        //piper-lib-os compatibility
        cpe?.setBuildResult(result)
    }
    def getBuildResult() {
        return buildResult
    }
    //not used - check if still used in templates
    def addError(script, error){
        if (error != null) {
            //error
            script.echo "Error: ${error}"
            if (error instanceof FlowInterruptedException) {
                setBuildResult('ABORTED')
            } else {
                setBuildResult('FAILURE')
                script.currentBuild.result = 'FAILURE'
            }
        }
    }

    def getGithubORB(){
      return "${getGithubOrg()}/${getGithubRepo()}/${getGitBranch()}"
    }

    def setInfluxCustomDataProperty(field, value) {
        influxCustomData[field] = value
        //piper-lib-os compatibility
        cpe?.setInfluxCustomDataEntry(field, value)
    }
    def getInfluxCustomData() {
        return influxCustomData
    }
    def setInfluxCustomDataTagsProperty(tag, value) {
        influxCustomDataTags[tag] = value
        //piper-lib-os compatibility
        cpe?.setInfluxCustomDataTagsEntry(tag, value)
    }
    def getInfluxCustomDataTags() {
        return influxCustomDataTags
    }
    def setInfluxCustomDataMapProperty(measurement, field, value){
        if (!influxCustomDataMap[measurement]) {
            influxCustomDataMap[measurement] = [:]
        }
        influxCustomDataMap[measurement][field] = value
        //piper-lib-os compatibility
        cpe?.setInfluxCustomDataMapEntry(measurement, field, value)
    }
    def getInfluxCustomDataMap() {
        return influxCustomDataMap
    }
    def setInfluxCustomDataMapTagsProperty(measurement, tag, value) {
        if (!influxCustomDataMapTags[measurement]) {
            influxCustomDataMapTags[measurement] = [:]
        }
        influxCustomDataMapTags[measurement][tag] = value
        //piper-lib-os compatibility
        cpe?.setInfluxCustomDataMapTagsEntry(measurement, tag, value)
    }
    def getInfluxCustomDataMapTags() {
        return influxCustomDataMapTags
    }
    def setPipelineMeasurement(field, value) {
        setInfluxPipelineData(field, value)
    }
    def setInfluxPipelineData(field, value) {
        setInfluxCustomDataMapProperty('pipeline_data', field, value)
    }
    def setInfluxStepData(field, value) {
        setInfluxCustomDataMapProperty('step_data', field, value)
    }
    def setGithubStatistics(script, githubApiUrl, githubOrg, githubRepo, gitCommit, credentialId) {
        //collecting GitHub statistics
        def githubRequestParams = [
            url: "${githubApiUrl}/repos/${githubOrg}/${githubRepo}/commits/${gitCommit}",
            timeout: 5,
            authentication: credentialId
        ]
        try {
            def response = script.httpRequest githubRequestParams

            def content = script.readJSON text: response.content

            setPipelineMeasurement('github_changes', content.stats.total)
            setPipelineMeasurement('github_additions', content.stats.additions)
            setPipelineMeasurement('github_deletions', content.stats.deletions)
            setPipelineMeasurement('github_filesChanged', content.files.size())
        } catch (err) {
            // ignore error due to newly introduced mandatory authentication
            Notify.warning(script, 'failed to connect to GitHub API - GitHub statistics were not retrieved. Please configure gitHttpsCredentialsId', 'globalPipelineEnvironment')
        }
    }
    def getDockerImageNameAndTag() {
        def defaultDockerImageName = configuration.general?.dockerImageName
        def defaultDockerImageVersion = getArtifactVersion()
        def dockerImageAndTag = (defaultDockerImageName != null && defaultDockerImageVersion != null) ? "${defaultDockerImageName}:${defaultDockerImageVersion}" : null
        return dockerImageAndTag
    }
    def setDockerBuildImage(image) {
        dockerBuildImage = image
    }
    def getDockerBuildImage() {
        return dockerBuildImage
    }

    def setNexusPromoteRepositories(String nexusPromoteRepositories) {
        this.nexusPromoteRepositories = nexusPromoteRepositories
    }

    def getNexusPromoteRepositories() {
        return nexusPromoteRepositories
    }

    def setNexusStagingGroup(String nexusStagingGroup) {
        this.nexusStagingGroup = nexusStagingGroup
    }

    def getNexusStagingGroup() {
        return nexusStagingGroup
    }

    def hasLabelAction(FlowNode flowNode){
        def actions = flowNode.getActions()
        def result = false
        actions.each {
            action ->
                if (action instanceof LabelAction) {
                    result = true
                    return
                }
        }
        return result
    }

    void setFlag(flagName) {
        flags[flagName] = true
    }

    void removeFlag(flagName) {
        flags[flagName] = false
    }

    boolean getFlag(flagName) {
        return flags[flagName] == true
    }

    Map getStepConfiguration(stepName, stageName, includeDefaults = true) {
        Map defaults = [:]
        if (includeDefaults) {
            defaults = ConfigurationLoader.defaultGeneralConfiguration()
            defaults = ConfigurationMerger.merge(ConfigurationLoader.defaultStepConfiguration(null, stepName), null, defaults)
            defaults = ConfigurationMerger.merge(ConfigurationLoader.defaultStageConfiguration(null, stageName), null, defaults)
        }
        Map config = ConfigurationMerger.merge(configuration.get('general') ?: [:], null, defaults)
        config = ConfigurationMerger.merge(configuration.get('steps')?.get(stepName) ?: [:], null, config)
        config = ConfigurationMerger.merge(configuration.get('stages')?.get(stageName) ?: [:], null, config)
        return config
    }
}

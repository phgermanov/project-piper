package com.sap.piper.internal

import com.sap.icd.jenkins.Utils

import com.sap.piper.internal.ConfigurationHelper

class PiperGoUtils implements Serializable {

    static Map prepare(Script step, Script script, Utils utils = null){
        PiperGoUtils piperGoUtils = new PiperGoUtils(step, utils ?: new Utils())
        Map stepParams = [piperGoPath: './sap-piper', piperGoUtils: piperGoUtils]

        Map generalConfig = ConfigurationHelper
            .loadStepDefaults(step)
            .mixinGeneralConfig(script.globalPipelineEnvironment)
            .use()
        piperGoUtils.setSapPiperDownloadCredentialsId(generalConfig.sapPiperDownloadCredentialsId)

        return stepParams
    }

    private Script steps
    private Utils utils
    private String sapPiperDownloadCredentialsId

    private static String DELIMITER = '-DeLiMiTeR-'

    PiperGoUtils(Script steps) {
        this.steps = steps
        this.utils = new Utils()
    }

    PiperGoUtils(Script steps, Utils utils) {
        this.steps = steps
        this.utils = utils
    }

    void setSapPiperDownloadCredentialsId(sapPiperDownloadCredentialsId){
        this.sapPiperDownloadCredentialsId = sapPiperDownloadCredentialsId
    }

    void unstashPiperBin() {
        // Check if the sap-piper binary is already present
        if (steps.sh(script: "[ -x ./sap-piper ]", returnStatus: true) == 0) {
            steps.echo "Found sap-piper binary in the workspace - skipping unstash"
            return
        }

        if (utils.unstash('sap-piper-bin').size() > 0) return

        def libraries = getLibrariesInfo()
        String version
        libraries.each {lib ->
            if (lib.name == 'piper-lib' || lib.name == 'piper-library') {
                version = lib.version
            }
        }

        def binaryVersion = (version == 'master') ? "latest" : version

        def binaryOutputName = 'sap-piper'
        def binaryURL = "https://sap-piper.int.repositories.cloud.sap/artifactory/sap-piper/releases/$binaryVersion/sap-piper"
        def fallbackURL = "https://sap-piper.int.repositories.cloud.sap/artifactory/sap-piper/releases/latest/sap-piper"
        def downloadSuccess = curlWithOutput(binaryURL, binaryOutputName)
        if (downloadSuccess) {
            steps.sh(script: "chmod +x $binaryOutputName")
            utils.stashWithMessage('sap-piper-bin', 'failed to stash sap-piper binary', 'sap-piper')
            return
        }
        Notify.warning(steps, "Failed to download sap-piper binary for version $binaryVersion. Trying to download latest version.", 'PiperGoUtils')

        def fallbackDownloadSuccess = curlWithOutput(fallbackURL, binaryOutputName)
        if (fallbackDownloadSuccess) {
            steps.sh(script: "chmod +x $binaryOutputName")
            utils.stashWithMessage('sap-piper-bin', 'failed to stash sap-piper binary', 'sap-piper')
            return
        }

        def errorResponse = steps.sh(script: "cat $binaryOutputName", returnStdout: true)
        Notify.warning(steps, "Failed to download sap-piper binary from Artifactory. Error: $errorResponse . Trying to dowload binary from GitHub releases.", 'PiperGoUtils')

        List credentials = []
        if (sapPiperDownloadCredentialsId) {
            credentials.add(steps.usernamePassword(
                credentialsId: sapPiperDownloadCredentialsId,
                usernameVariable: 'USERNAME',
                passwordVariable: 'TOKEN'
            ))
        }
        steps.withCredentials(credentials) {
            boolean downloaded = downloadGoBinary(steps.env.TOKEN, version)
            if (!downloaded) {
                Notify.error(steps, "Download of Piper go binary failed.", 'PiperGoUtils')
            }
            steps.sh(script: "chmod +x $binaryOutputName")
        }
    }

    List getLibrariesInfo() {
        return utils.getJenkinsUtilsInstance().getLibrariesInfo()
    }

    private boolean downloadGoBinary(String token, String version) {
        String gitHub = "https://github.wdf.sap.corp"
        String apiEndpoint = "$gitHub/api/v3" as String
        String repo = "ContinuousDelivery/piper-library"

        if (!token) {
            Notify.error(steps, "Downloading the Piper binary from Artifactory failed, and it might be a temporary issue.\n"
                + "As a workaround, you can configure 'sapPiperDownloadCredentialsId' in the general section of your .pipeline/config.yml file.\n"
                + "The value should be a Jenkins credentials ID of type username + password, where the password is a GitHub Personal Access Token for $gitHub.",
                'PiperGoUtils')
        }

        def response = null
        if (version != "latest") {
            // NOTE: In case the "version" is a branch, this will also not find a release, and the fall-back to "latest" is used.
            response = gitHubAPIRequest(token, "$apiEndpoint/repos/$repo/releases/tags/$version" as String)
            if (response?.message == "Not Found") {
                Notify.warning(steps, "No sap-piper release found for version '$version', falling back to 'latest'.", 'PiperGoUtils')
                version = "latest"
            }
        }
        if (version == "latest") {
            response = gitHubAPIRequest(token, "$apiEndpoint/repos/$repo/releases/latest" as String)
        }

        String assetName = "sap-piper"
        String assetId = findAssetIdByName(response as Map, assetName)

        boolean success = gitHubDownloadReleaseAsset(token, "$apiEndpoint/repos/$repo/releases/assets/$assetId" as String, "./sap-piper")
        if (!success) {
            Notify.warning(steps, "Failed to download sap-piper release asset (ID $assetId) via GitHub API", 'PiperGoUtils')
        }
        return success
    }

    private String findAssetIdByName(Map release, String assetName) {
        for (def asset in release.assets) {
            if (asset.name == assetName) {
                return asset.id
            }
        }
        Notify.error(steps, "Did not find sap-piper release asset in version ${release.tag_name}", 'PiperGoUtils')
        return "" // Not reached, suppresses warning
    }

    // Not private due to https://issues.apache.org/jira/browse/GROOVY-7368
    Map gitHubAPIRequest(String token, String url) {
        List<String> curlCommand = ["curl"]
        if (token) {
            curlCommand.addAll(["--header", "\"Authorization: token $token\""])
        }
        curlCommand.addAll([
            "--insecure",
            "--silent",
            "--location",
            "--header", "\"Accept: application/vnd.github.v3+json\"",
            "'$url'"
        ])
        def response = steps.readJSON(text: steps.sh(returnStdout: true, script: curlCommand.join(" ")))
        if (response?.message == "Bad credentials") {
            Notify.error(steps, "Invalid Personal Access Token provided in 'sapPiperDownloadCredentialsId'.\n"
                + "Its value needs to be a Jenkins secret store credentials ID of type username + password, where the password is a GitHub Personal Access Token for $url.",
                'PiperGoUtils')
        }
        return response as Map
    }

    private boolean curlWithOutput(String url, String outputName) {
        List curlCommand = [
            "curl",
            "--silent",
            "--location",
            "--write-out", "'%{http_code}'",
            "--output", "$outputName",
            "'$url'"
            ]

        def response = steps.sh(returnStdout: true, script: curlCommand.join(" "))
        return response == '200'
    }

    private boolean gitHubDownloadReleaseAsset(String token, String url, String file) {
        List curlCommand = ["curl"]
        if (token) {
            curlCommand.addAll(["--header", "\"Authorization: token $token\""])
        }
        curlCommand.addAll([
            "--silent",
            "--location",
            "--header", "\"Accept: application/octet-stream\"",
            "--write-out", "'${DELIMITER}status=%{http_code}'",
            "--output", "\"$file\"",
            "'$url'",
        ])
        def response = steps.sh(returnStdout: true, script: curlCommand.join(" "))
        def parts = response.split(DELIMITER)
        return parts.size() > 1 && parts[1] == 'status=200'
    }
}

package com.sap.piper.internal

class GitUtils implements Serializable {
    static void handleTestRepository(Script steps, Map config){
        def stashName = "testContent-${UUID.randomUUID()}".toString()
        def options = [url: config.testRepository]
        if (config.gitSshKeyCredentialsId)
            options.put('credentialsId', config.gitSshKeyCredentialsId)
        if (config.gitBranch)
            options.put('branch', config.gitBranch)
        // checkout test repository
        steps.git options
        // stash test content
        steps.stash stashName
        // alter stashContent
        config.stashContent = [stashName]
    }
}

package com.sap.icd.jenkins

class EnvironmentManagerRunner implements Serializable {

    private Script script
    private String cfCredentialsId
    private String muenchhausenCredentialsId
    private String gitHttpsCredentialsId

    public EnvironmentManagerRunner(Script script, String cfCredentialsId, String muenchhausenCredentialsId, String gitHttpsCredentialsId) {
        this.script = script
        this.cfCredentialsId = cfCredentialsId
        this.muenchhausenCredentialsId = muenchhausenCredentialsId
        this.gitHttpsCredentialsId = gitHttpsCredentialsId
    }

    public int execute(String groovyCommand) {
        String envManPathWithUUID = UUID.randomUUID().toString()
        script.dir(envManPathWithUUID) { script.git(branch: 'master', url: 'https://github.wdf.sap.corp/cc-devops-envman/EnvMan', credentialsId: this.gitHttpsCredentialsId) }

        return withMHCredentialsWrapper({ callGroovyEnvironmentManager(groovyCommand, envManPathWithUUID) })
    }

    public int executeWithDocker(String groovyCommand, String dockerImage, dockerWorkspace) {
        String envManPathWithUUID = UUID.randomUUID().toString()
        script.dir(envManPathWithUUID) { script.git(branch: 'master', url: 'https://github.wdf.sap.corp/cc-devops-envman/EnvMan', credentialsId: this.gitHttpsCredentialsId) }

        return script.dockerExecute(script: script, dockerImage: dockerImage, dockerWorkspace: dockerWorkspace) {
            withMHCredentialsWrapper({ callGroovyEnvironmentManager(groovyCommand, envManPathWithUUID) })
        }
    }

    private int withCredentialsWrapper(Closure body) {
        return new CredentialsWrapper(script, cfCredentialsId).asCFCredentialsEncoded(body)
    }

    private int withMHCredentialsWrapper(Closure body) {
        return new CredentialsWrapper(script, muenchhausenCredentialsId).asMuenchhausenCredentialsEncoded({ withCredentialsWrapper(body) })
    }

    private int callGroovyEnvironmentManager(String groovyCommand, String envManPathWithUUID) {
        String command = "groovy --classpath $envManPathWithUUID/src:$envManPathWithUUID/lib/* $envManPathWithUUID/src/com/sap/icd/environmentManager/EnvironmentManager.groovy " + groovyCommand
        if (command.toLowerCase().contains('--debug true')) {
            script.echo "$command"
        }
        int exitValue = script.sh(script: """#!/bin/bash
set +x
$command
""", returnStatus: true)
        script.sh("rm -r $envManPathWithUUID")

        return exitValue
    }
}

package com.sap.piper.internal

import hudson.AbortException

class DockerUtils implements Serializable {

    private static Script script

    DockerUtils(Script script) {
        this.script = script
    }

    public boolean withDockerDeamon() {
        def returnCode = script.sh script: 'docker ps -q > /dev/null', returnStatus: true
        return (returnCode == 0)
    }

    public boolean onKubernetes() {
        return (Boolean.valueOf(script.env.ON_K8S) || (script.env.jaas_owner != null))
    }

    public void saveImage(filePath, dockerImage, dockerRegistryUrl = '', dockerCredentialsId = null) {
        def dockerRegistry = dockerRegistryUrl ? "${getRegistryFromUrl(dockerRegistryUrl)}/" : ''
        def imageFullName = dockerRegistry + dockerImage
        if (withDockerDeamon()) {
            if (dockerRegistry) {
                script.docker.withRegistry(dockerRegistryUrl) {
                    script.sh "docker pull ${imageFullName} && docker save --output ${filePath} ${imageFullName}"
                }
            } else {
                script.sh "docker pull ${imageFullName} && docker save --output ${filePath} ${imageFullName}"
            }
        } else {
            try {
                //assume that we are on Kubernetes
                //needs to run inside an existing pod in order to not move around heavy images
                if(dockerCredentialsId){
                    script.withCredentials([script.usernamePassword(
                        credentialsId: dockerCredentialsId,
                        passwordVariable: 'password',
                        usernameVariable: 'user'
                    )]) {
                        skopeoSaveImage(imageFullName, dockerImage, filePath, script.user, script.password)
                    }
                }else{
                    skopeoSaveImage(imageFullName, dockerImage, filePath)
                }
            } catch (err) {
                throw new AbortException('No Kubernetes container provided for running Skopeo ...')
            }
        }
    }

    public void moveImage(Map source, Map target) {
        //expects source/target in the format [image: '', registryUrl: '', credentialsId: '']
        def sourceDockerRegistry = source.registryUrl ? "${getRegistryFromUrl(source.registryUrl)}/" : ''
        def sourceImageFullName = sourceDockerRegistry + source.image
        def targetDockerRegistry = target.registryUrl ? "${getRegistryFromUrl(target.registryUrl)}/" : ''
        def targetImageFullName = targetDockerRegistry + target.image

        if (withDockerDeamon()) {
            //not yet implemented here - available directly via pushToDockerRegistry
        } else {
            script.withCredentials([script.usernamePassword(
                credentialsId: target.credentialsId,
                passwordVariable: 'password',
                usernameVariable: 'userid'
            )]) {
                skopeoMoveImage(sourceImageFullName, targetImageFullName, script.userid, script.password)
            }
        }
    }

    private void runSkopeoWithRetry( skopeoCmd ) {
        try {
            script.sh skopeoCmd
        } catch (err) {
            // The skopeo tool can fail with "i/o timeout".
            // Often a retry will work.
            // No root cause has been found yet, so we should retry.
            // See: https://sapjira.wdf.sap.corp/browse/DTSOF-10024
            script.echo "skopeo failed. Retrying..."
            script.sh skopeoCmd
        }
    }

    private void skopeoSaveImage(imageFullName, dockerImage, filePath, user = null, password = null) {
        def source = "docker://${imageFullName}"
        def target = "docker-archive:${filePath}:${dockerImage.replace("@","_")}"
        List options = [
            '--src-tls-verify=false'
        ]
        if(user) options.add("--src-creds=${BashUtils.escape(user)}:${BashUtils.escape(password)}")
        runSkopeoWithRetry("skopeo copy ${options.join(' ')} ${source} ${target}")
    }

    private void skopeoMoveImage(sourceImageFullName, targetImageFullName, targetUserId, targetPassword) {
        runSkopeoWithRetry("skopeo copy --src-tls-verify=false --dest-tls-verify=false --dest-creds=${BashUtils.escape(targetUserId)}:${BashUtils.escape(targetPassword)} docker://${sourceImageFullName} docker://${targetImageFullName}")
    }

    public String getRegistryFromUrl(dockerRegistryUrl) {
        return dockerRegistryUrl.split(/^https?:\/\//)[1]
    }

    public String getProtocolFromUrl(dockerRegistryUrl) {
        return dockerRegistryUrl.split(/:\/\//)[0]
    }

    public String getNameFromImageUrl(imageUrl) {

        def imageNameAndTag

        //remove digest if present
        imageUrl = imageUrl.split('@')[0]

        //remove registry part if present
        def pattern = /\.(?:[^\/]*)\/(.*)/
        def matcher = imageUrl =~ pattern
        if (matcher.size() == 0) {
            imageNameAndTag = imageUrl
        } else {
            imageNameAndTag = matcher[0][1]
        }

        //remove tag if present
        return removeTagFromImageName(imageNameAndTag)
    }

    public String removeTagFromImageName(imageNameAndTag) {
        //remove tag if present
        return imageNameAndTag.split(':')[0]
    }
}

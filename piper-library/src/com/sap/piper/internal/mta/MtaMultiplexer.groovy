package com.sap.piper.internal.mta

import com.cloudbees.groovy.cps.NonCPS

class MtaMultiplexer implements Serializable {
    static Map createJobs(Script step, Map parameters, List excludeList, String jobPrefix, String buildDescriptorFile, String scanType, Closure worker) {
        Map jobs = [:]
        def filesToScan = []

        // avoid java.io.IOException: /var/jenkins_home/workspace/... does not exist.
        // see https://issues.jenkins-ci.org/browse/JENKINS-62077
        if(step.findFiles().size() == 0){
            step.echo "Found 0 ${scanType} descriptor files, the directory is empty."
            return jobs
        }
        // avoid java.io.NotSerializableException: org.codehaus.groovy.util.ArrayIterator
        // see https://issues.jenkins-ci.org/browse/JENKINS-47730
        filesToScan.addAll(step.findFiles(glob: "**${File.separator}${buildDescriptorFile}")?:[])
        step.echo "Found ${filesToScan?.size()} ${scanType} descriptor files: ${filesToScan}"

        filesToScan = removeNodeModuleFiles(step, filesToScan)
        filesToScan = removeExcludedFiles(step, filesToScan, excludeList)

        for (int i = 0; i < filesToScan.size(); i++){
            def file = filesToScan.get(i)
            def options = [:]
            options.putAll(parameters)
            options.scanType = scanType
            options.buildDescriptorFile = file.path
            jobs["${jobPrefix} - ${file.path.replace("${File.separator}${buildDescriptorFile}",'')}"] = {worker(options)}
        }
        return jobs
    }

    @NonCPS
    static def removeNodeModuleFiles(Script step, filesToScan){
        step.echo "Excluding node_modules:"
        return filesToScan.findAll({
            if(it.path.contains("node_modules${File.separator}")){
                step.echo "- Skipping ${it.path}"
                return false
            }
            return true
        })
    }

    @NonCPS
    static def removeExcludedFiles(Script step, filesToScan, List<String> filesToExclude){
        step.echo "Applying exclude list: ${filesToExclude}"
        return filesToScan.findAll({
            if(filesToExclude.contains(it.path)){
                step.echo "- Skipping ${it.path}"
                return false
            }
            return true
        })
    }
}

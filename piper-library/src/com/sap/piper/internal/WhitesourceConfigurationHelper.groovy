package com.sap.piper.internal

import com.cloudbees.groovy.cps.NonCPS
import com.sap.icd.jenkins.Utils

import java.security.MessageDigest

class WhitesourceConfigurationHelper implements Serializable {

    private static def SCALA_CONTENT_KEY = "@__content"

    static def extendConfigurationFile(script, utils, config, path) {
        def mapping = [:]
        def parsingClosure
        def serializationClosure
        def inputFile = config.configFilePath.replaceFirst('\\./', '')
        def suffix = MessageDigest.getInstance("MD5").digest("${path}${inputFile}".bytes).encodeHex().toString()
        def targetFile = "${inputFile}.${suffix}"
        switch (config.scanType) {
            case 'unifiedAgent':
            case 'fileAgent':
                mapping = [
                    [name: 'apiKey', value: config.orgToken, warnIfPresent: true],
                    [name: 'productName', value: config.whitesourceProductName],
                    [name: 'productToken', value: config.whitesourceProductToken, omitIfPresent: 'projectToken'],
                    [name: 'userKey', value: config.userKey, warnIfPresent: true]
                ]
                parsingClosure = { fileReadPath -> return script.readProperties (file: fileReadPath) }
                serializationClosure = { configuration -> serializeUAConfig(configuration) }
                break
            case 'yarn':
                mapping = [
                    [name: 'apiKey', value: config.orgToken, warnIfPresent: true],
                    [name: 'productName', value: config.whitesourceProductName],
                    [name: 'productVer', value: config.whitesourceProductVersion],
                    [name: 'productToken', value: config.whitesourceProductToken, omitIfPresent: 'projectToken'],
                    [name: 'userKey', value: config.userKey, warnIfPresent: true],
                    [name: 'whitesourceYarnVersion', value: config.whitesourceYarnVersion]
                ]
                parsingClosure = { fileReadPath -> return script.readJSON (file: fileReadPath) }
                serializationClosure = { configuration -> return new Utils().getPrettyJsonString(configuration) }
                break
            case 'npm':
                mapping = [
                    [name: 'apiKey', value: config.orgToken, warnIfPresent: true],
                    [name: 'productName', value: config.whitesourceProductName],
                    [name: 'productVer', value: config.whitesourceProductVersion],
                    [name: 'productToken', value: config.whitesourceProductToken, omitIfPresent: 'projectToken'],
                    [name: 'userKey', value: config.userKey, warnIfPresent: true]
                ]
                parsingClosure = { fileReadPath -> return script.readJSON (file: fileReadPath) }
                serializationClosure = { configuration -> return new Utils().getPrettyJsonString(configuration) }
                break
            case 'pip':
                mapping = [
                    [name: "'org_token'", value: "\'${config.orgToken}\'", warnIfPresent: true],
                    [name: "'product_name'", value: "\'${config.whitesourceProductName}\'"],
                    [name: "'product_version'", value: "\'${config.whitesourceProductVersion}\'"],
                    [name: "'product_token'", value: "\'${config.whitesourceProductToken}\'"],
                    [name: "'user_key'", value: "\'${config.userKey}\'", warnIfPresent: true]
                ]
                parsingClosure = { fileReadPath -> return readPythonConfig (script, fileReadPath) }
                serializationClosure = { configuration -> serializePythonConfig(configuration) }
                targetFile = "${inputFile}.${suffix}.py"
                break
            case 'sbt':
                mapping = [
                    [name: "whitesourceOrgToken in ThisBuild", value: "\"${config.orgToken}\"", warnIfPresent: true],
                    [name: "whitesourceProduct in ThisBuild", value: "\"${config.whitesourceProductName}\""],
                    [name: "whitesourceServiceUrl in ThisBuild", value: "uri(\"${config.whitesourceAgentUrl}\")"]
                    // actually not supported [name: "whitesourceUserKey in ThisBuild", value: config.userKey]
                ]
                parsingClosure = { fileReadPath -> return readScalaConfig (script, mapping, fileReadPath) }
                serializationClosure = { configuration -> serializeScalaConfig (configuration) }
                targetFile = inputFile
                break
        }

        rewriteConfiguration(script, utils, config, mapping, suffix, path, inputFile, targetFile, parsingClosure, serializationClosure)
    }

    static private def rewriteConfiguration(script, utils, config, mapping, suffix, path, inputFile, targetFile, parsingClosure, serializationClosure) {
        def inputFilePath = "${path}${inputFile}"
        def outputFilePath = "${path}${targetFile}"
        def moduleSpecificFile

        // Handle IO exceptions while trying to read module specific file
        try {
            moduleSpecificFile = parsingClosure(inputFilePath)
        } catch(e) {
            if(config.verbose)
                script.echo "Could not read config file ${inputFilePath}. ${e}"
            moduleSpecificFile = false
        }

        if (!moduleSpecificFile) {
            try {
                moduleSpecificFile = parsingClosure(config.configFilePath)
            } catch(e) {
                if(config.verbose)
                    script.echo "Could not read config file ${config.configFilePath}. ${e}"
                moduleSpecificFile = false
            }
        }

        if (!moduleSpecificFile)
            moduleSpecificFile = [:]

        for(int i = 0; i < mapping.size(); i++) {
            def entry = mapping.get(i)
            if (entry.warnIfPresent && moduleSpecificFile[entry.name])
                Notify.warning(script, "Obsolete configuration ${entry.name} detected, please omit its use and rely on configuration via Piper.", 'WhitesourceConfigurationHelper')
            def dependentValue = entry.omitIfPresent ? moduleSpecificFile[entry.omitIfPresent] : null
            if ((entry.omitIfPresent && !dependentValue || !entry.omitIfPresent) && entry.value && entry.value != 'null' && entry.value != '' && entry.value != "'null'")
                moduleSpecificFile[entry.name] = entry.value.toString()
        }

        def output = serializationClosure(moduleSpecificFile)

        if(config.verbose)
            script.echo "Writing config file ${outputFilePath} with content:\n${output}"
        script.writeFile file: outputFilePath, text: output
        if(config.stashContent && config.stashContent.size() > 0) {
            def stashName = "modified whitesource config ${suffix}".toString()
            utils.stashWithMessage (
                stashName,
                "Stashing modified Whitesource configuration",
                outputFilePath.replaceFirst('\\./', '')
            )
            config.stashContent += [stashName]
        }
        config.configFilePath = targetFile
    }

    static private def readPythonConfig(script, filePath) {
        def contents = script.readFile file: filePath
        def lines = contents.split('\n')
        def resultMap = [:]
        for(int i = 0; i < lines.length; i++) {
            def line = lines[i]
            List parts = line?.replaceAll(',$', '')?.split(':')
            def key = parts[0]?.trim()
            parts.removeAt(0)
            resultMap[key] = parts.size() > 0 ? (parts as String[]).join(':').trim() : null
        }
        return resultMap
    }

    static private def serializePythonConfig(configuration) {
        StringBuilder result = new StringBuilder()
        for(int i = 0; i < configuration.entrySet().size(); i++) {
            def entry = configuration.entrySet().getAt(i)
            if(entry.key != '}')
                result.append(entry.value ? '    ' : '').append(entry.key).append(entry.value ? ': ' : '').append(entry.value ?: '').append(entry.value ? ',' : '').append('\r\n')
        }
        return result.toString().replaceAll(',$', '\r\n}')
    }

    static private def readScalaConfig(script, mapping, filePath) {
        def contents = script.readFile file: filePath
        def lines = contents.split('\n')
        def resultMap = [:]
        resultMap[SCALA_CONTENT_KEY] = []
        def keys = []
        for(int i = 0; i < mapping.size(); i++) {
            keys.add(mapping.get(i).name)
        }
        for(int i = 0; i < lines.length; i++) {
            def line = lines[i]
            def parts = line?.split(':=').toList()
            def key = parts[0]?.trim()
            if (keys.contains(key)) {
                resultMap[key] = parts[1]?.trim()
            } else if (line != null) {
                resultMap[SCALA_CONTENT_KEY].add(line)
            }
        }
        return resultMap
    }

    static private def serializeScalaConfig(configuration) {
        StringBuilder result = new StringBuilder()

        // write the general content
        for(int i = 0; i < configuration[SCALA_CONTENT_KEY]?.size(); i++) {
            def line = configuration[SCALA_CONTENT_KEY]?.get(i)
            result.append(line)
            result.append('\r\n')
        }

        // write the mappings
        def confKeys = configuration.keySet()
        confKeys.remove(SCALA_CONTENT_KEY)

        for(int i = 0; i < confKeys.size(); i++) {
            def key = confKeys.getAt(i)
            def value = configuration[key]
            result.append(key)
            if (value != null) {
                result.append(' := ').append(value)
            }
            result.append('\r\n')
        }

        return result.toString()
    }

    @NonCPS
    static private def serializeUAConfig(configuration) {
        Properties p = new Properties()
        for(int i = 0; i < configuration.entrySet().size(); i++) {
            def entry = configuration.entrySet().getAt(i)
            p.setProperty(entry.key, entry.value)
        }

        new StringWriter().with{ w -> p.store(w, null); w }.toString()
    }
}

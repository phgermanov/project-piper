package com.sap.icd.jenkins

import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.internal.JenkinsUtils
import com.sap.piper.internal.Notify

import groovy.json.JsonBuilder
import groovy.json.JsonSlurper
import groovy.json.JsonSlurperClassic
import groovy.transform.Field
import groovy.xml.XmlUtil

import java.nio.charset.StandardCharsets
import java.security.MessageDigest
import java.util.regex.Matcher
import java.util.regex.Pattern

import org.jenkinsci.plugins.workflow.steps.FlowInterruptedException

@Field
def transient jenkinsUtils

@Field
def name = Pattern.compile("(.*)name=['\"](.*?)['\"](.*)", Pattern.DOTALL)

@Field
def version = Pattern.compile("(.*)version=['\"](.*?)['\"](.*)", Pattern.DOTALL)

@Field
def method = Pattern.compile("(.*)\\(\\)", Pattern.DOTALL)

def getJenkinsUtilsInstance() {
    if(null == jenkinsUtils)
        jenkinsUtils = new JenkinsUtils()
    return jenkinsUtils
}

def getGitCommitMessage() {
    def message = ''
    if (fileExists('.git')) {
        message = sh(returnStdout: true, script: 'git log -1')?.trim()
    }
    return message
}

def getGitCommitId() {
	return sh(returnStdout: true, script: 'git rev-parse HEAD')?.trim()
}

def getGitCommitIdOrNull() {
    if (fileExists('.git')) {
        getGitCommitId()
    } else {
        return null
    }
}

//@Deprecated('unused')
//TODO: should make use of githubApiUrl from general default settings
def getCommitInfo(org, repo, commit, String githubApiUrl = 'https://github.wdf.sap.corp/api/v3') {
	def response = httpRequest "${githubApiUrl}/repos/${org}/${repo}/commits/${commit}"
	readJSON text: response.content
}

def getGitRemoteUrl() {
	return sh(returnStdout: true, script: 'git config --get remote.origin.url').trim()
}

def getFolderFromGitUrl(gitUrl) {
	def folderStart
	if (gitUrl.startsWith('http')) {
		folderStart = gitUrl.indexOf('/', 8) + 1
	} else {
		folderStart = gitUrl.lastIndexOf(":") + 1
	}
	def folderEnd = gitUrl.lastIndexOf("/")

	return (folderEnd < 0 || folderStart == folderEnd + 1) ? '' :  gitUrl.substring(folderStart, folderEnd)
}

def getRepositoryFromGitUrl(gitUrl) {
	def repoStart = gitUrl.lastIndexOf("/") + 1
	def repoEnd = gitUrl.lastIndexOf(".git")
    if (repoEnd < 0) {
        repoEnd = gitUrl.length()
    }

	//for ssh git url without folder like git@github.wdf.sap.corp:test-pipeline.git
	if (repoStart == 0)
		repoStart = gitUrl.lastIndexOf(":") + 1

	return gitUrl.substring(repoStart, repoEnd)
}

def decodeBase64(text) {
	return sh(returnStdout: true, script: "echo ${text} | base64 --decode").trim()
}

/**
* @deprecated directly use <code>new JenkinsUtils().isJobStartedByTimer()</code> instead
**/
@Deprecated
def isJobStartedByTimer() {
    Notify.warning(this, "Utils.isJobStartedByTimer() has been deprecated and will be removed, please directly use JenkinsUtils.isJobStartedByTimer().", 'com.sap.icd.jenkins.Utils')
	return getJenkinsUtilsInstance().isJobStartedByTimer()
}

/**
* @deprecated directly use <code>new JenkinsUtils().scheduleJob(...)</code> instead
**/
@Deprecated
void scheduleJob(String spec) {
    Notify.warning(this, "Utils.scheduleJob(...) has been deprecated and will be removed, please directly use JenkinsUtils.scheduleJob(...).", 'com.sap.icd.jenkins.Utils')
    getJenkinsUtilsInstance().scheduleJob(spec)
}

/**
* @deprecated directly use <code>new JenkinsUtils().removeJobSchedule(...)</code> instead
**/
@Deprecated
void removeJobSchedule(String spec = null) {
    Notify.warning(this, "Utils.removeJobSchedule(...) has been deprecated and will be removed, please directly use JenkinsUtils.removeJobSchedule(...).", 'com.sap.icd.jenkins.Utils')
    getJenkinsUtilsInstance().removeJobSchedule(spec)
}

/**
* @deprecated directly use <code>new JenkinsUtils().addBuildDiscarder(...)</code> instead
**/
@Deprecated
void addBuildDiscarder(int daysToKeep = -1, int numToKeep = -1, int artifactDaysToKeep = -1, int artifactNumToKeep = -1) {
    Notify.warning(this, "Utils.addBuildDiscarder(...) has been deprecated and will be removed, please directly use JenkinsUtils.addBuildDiscarder(...).", 'com.sap.icd.jenkins.Utils')
    getJenkinsUtilsInstance().addBuildDiscarder(daysToKeep, numToKeep, artifactDaysToKeep, artifactNumToKeep)
}

/**
 * @deprecated directly use <code>new JenkinsUtils().removeBuildDiscarder()</code> instead
 **/
@Deprecated
void removeBuildDiscarder() {
    Notify.warning(this, "Utils.removeBuildDiscarder() has been deprecated and will be removed, please directly use JenkinsUtils.removeBuildDiscarder().", 'com.sap.icd.jenkins.Utils')
    getJenkinsUtilsInstance().removeBuildDiscarder()
}

/**
* @deprecated directly use <code>new JenkinsUtils().getLibrariesInfoWithPiperLatest()</code> instead
**/
@Deprecated
def getLibrariesInfoWithPiperLatest() {
    Notify.warning(this, "Utils.getLibrariesInfoWithPiperLatest() has been deprecated and will be removed, please directly use JenkinsUtils.getLibrariesInfoWithPiperLatest().", 'com.sap.icd.jenkins.Utils')
    return getJenkinsUtilsInstance().getLibrariesInfoWithPiperLatest()
}

@NonCPS
def generateHash(input) {
    return MessageDigest
        .getInstance("SHA-1")
        .digest(input.getBytes(StandardCharsets.UTF_8))
        .encodeHex().toString()
}

@NonCPS
private String getPayloadString(Map payload){
    return payload
        .collect { key, value -> return "&${key}=${URLEncoder.encode(value.toString(), "UTF-8")}"}
        .join('')
        .replaceFirst('&','?')
}

@NonCPS
def getParameter(Map map, paramName, defaultValue = null) {

	def paramValue = map[paramName]

	if (paramValue == null)
		paramValue = defaultValue

	return paramValue
}

@NonCPS
def getMandatoryParameter(Map map, paramName, defaultValue = null) {

	def paramValue = map[paramName]

	if (paramValue == null)
		paramValue = defaultValue

	if (paramValue == null)
		throw new IllegalArgumentException("ERROR - NO VALUE AVAILABLE FOR ${paramName}")

	return paramValue
}

@NonCPS
def parseJson(text) {
	def object = new JsonSlurper().parseText(text)
	return serializeMap(object)
}

@NonCPS
def parseJsonSerializable(text) {
	return new JsonSlurperClassic().parseText(text)
}

@NonCPS
def jsonToString(content) {
	return new JsonBuilder(content).toString()
}

@NonCPS
def getPrettyJsonString(object) {
	return groovy.json.JsonOutput.prettyPrint(groovy.json.JsonOutput.toJson(object))
}

@NonCPS
def serializeMap(object) {
	if (object instanceof groovy.json.internal.LazyMap) {
		for (item in object) {
			// println "${item.key}: ${item.value}"
			item.value = serializeMap(item.value)
		}
		return new HashMap<>(object)
	}
	return object
}

def unstash(name, msg = "Unstash failed:") {
	def unstashedContent = []
	try {
		echo "Unstash content: ${name}"
		steps.unstash name
		unstashedContent += name
	} catch (e) {
		echo "$msg $name (${e.getMessage()})"
	}
	return unstashedContent
}

def unstashAll(stashContent) {
	def unstashedContent = []
	if (stashContent) {
		for (i = 0; i < stashContent.size(); i++) {
			unstashedContent += unstash(stashContent[i])
		}
	}
	return unstashedContent
}

def stash(name, include = '**/*.*', exclude = '', useDefaultExcludes = true) {
	echo "Stash content: ${name} (include: ${include}, exclude: ${exclude})"
	steps.stash name: name, includes: include, excludes: exclude, useDefaultExcludes: useDefaultExcludes
}

def stashWithMessage(name, msg, include = '**/*.*', exclude = '', useDefaultExcludes = true) {
	try {
		stash(name, include, exclude, useDefaultExcludes)
	} catch (e) {
		echo msg + name + " (${e.getMessage()})"
	}
}


@NonCPS
def getSCMBranchName() {
	if (scm) {
		return scm.branches[0].name
	} else {
		return null
	}
}

def readMavenBuildConfigurations(pomFileName, helpEvaluateVersion) {
    // Maven Pom Model API: http://maven.apache.org/components/ref/3.3.9/maven-model/apidocs/org/apache/maven/model/Model.html
    def result = [:]
    result['finalName']    = evaluateFromMavenPom(pomFileName, "project.build.finalName", helpEvaluateVersion)
    result['targetFolder'] = evaluateFromMavenPom(pomFileName, "project.build.directory", helpEvaluateVersion)
    result['basedir']      = evaluateFromMavenPom(pomFileName, "project.basedir", helpEvaluateVersion)
    return result
}

def readMavenGAV(pomFileName, helpEvaluateVersion = '') {
    // Maven Pom Model API: http://maven.apache.org/components/ref/3.3.9/maven-model/apidocs/org/apache/maven/model/Model.html
    def result = [:]

    def descriptor = readMavenPom(file: pomFileName)

    result['packaging'] = readFromMavenDescriptor(descriptor, 'packaging') ?: evaluateFromMavenPom(pomFileName, "project.packaging", helpEvaluateVersion)
    result['group']     = readFromMavenDescriptor(descriptor, 'groupId') ?: evaluateFromMavenPom(pomFileName, "project.groupId", helpEvaluateVersion)
    result['artifact']  = readFromMavenDescriptor(descriptor, 'artifactId') ?: evaluateFromMavenPom(pomFileName, "project.artifactId", helpEvaluateVersion)
    result['version']   = readFromMavenDescriptor(descriptor, 'version') ?: evaluateFromMavenPom(pomFileName, "project.version", helpEvaluateVersion)

    echo "loaded ${result} from ${pomFileName}"
    return result
}

def readFromMavenDescriptor(descriptor, propertyName) {
    if (descriptor && propertyName) {
        def value = descriptor[propertyName]
        if (value?.matches(/.*?\$\{.*?\}.*/))
            return null
        return value
    }
    return null
}

/*
 * Uses the Maven Help plugin to evaluate the given expression into the resolved values
 * that maven sees at / generates at runtime. This way, the exact Maven coordinates and
 * variables can be used.
 */
def evaluateFromMavenPom(String pomFileName, String pomPathExpression, String helpEvaluateVersion) {
    def helpEvaluate = "help:evaluate"
    if (helpEvaluateVersion) {
        helpEvaluate = "org.apache.maven.plugins:maven-help-plugin:${helpEvaluateVersion}:evaluate"
    }
    // The regular expression used with grep matches all of the contents we don't want, and grep then
    // inverts this selection (-v flag) to grep only what we are interested in.
    // Especially the regex matches / filters out any strings that come from downloading artifacts.
    return sh(returnStdout: true, script: /mvn -f '${pomFileName}' ${helpEvaluate} -Dexpression=${pomPathExpression} | grep -Ev '(^\s*\[|Download|Java\w+:)'/).trim()
}

def getNpmGAV(file = 'package.json') {
	def result = [:]
	def descriptor = readJSON(file: file)

	if (descriptor.name.startsWith('@')) {
		def packageNameArray = descriptor.name.split('/')
		if (packageNameArray.length != 2)
			error "Unable to parse package name '${descriptor.name}'"
		result['group'] = packageNameArray[0]
		result['artifact'] = packageNameArray[1]
	} else {
		result['group'] = ''
		result['artifact'] = descriptor.name
	}
	result['version'] = descriptor.version
	echo "loaded ${result} from ${file}"
	return result
}

def getDubGAV(file = 'dub.json') {
    def result = [:]
    def descriptor = readJSON(file: file)

    result['group'] = 'com.sap.dlang'
    result['artifact'] = descriptor.name
    result['version'] = descriptor.version
    result['packaging'] = 'tar.gz'
    echo "loaded ${result} from ${file}"
    return result
}

def getSbtGAV(file = 'sbtDescriptor.json') {
	def result = [:]
	def descriptor = readJSON(file: file)

	result['group'] = descriptor.group
	result['artifact'] = descriptor.artifactId
	result['version'] = descriptor.version
	result['packaging'] = descriptor.packaging
	echo "loaded ${result} from ${file}"
	return result
}

def getMtaGAV(file = 'mta.yaml', xmakeConfigFile = '.xmake.cfg') {
    def result = [:]
    def descriptor = readYaml(file: file)
    def xmakeConfig = readProperties(file: xmakeConfigFile)

    result['group'] = xmakeConfig['mtar-group']
    result['artifact'] = descriptor.ID
    result['version'] = descriptor.version
    result['packaging'] = 'mtar'
    // using default value: https://github.wdf.sap.corp/dtxmake/xmake-mta-plugin#complete-list-of-default-values
    if(!result['group']){
        result['group'] = 'com.sap.prd.xmake.example.mtars'
        Notify.warning(this, "No groupID set in '.xmake.cfg', using default groupID '${result['group']}'.", 'com.sap.icd.jenkins.Utils')
    }
    echo "loaded ${result} from ${file} and ${xmakeConfigFile}"
    return result
}

def getPipGAV(file = 'setup.py') {
    def result = [:]
    result['group'] = ''
    result['packaging'] = ''

    def descriptor = sh(returnStdout: true, script: "cat ${file}")
    result['artifact'] = matches(name, descriptor)
    result['version'] = matches(version, descriptor)

    if (!result['version'] || matches(method, result['version'])) {
        def versionFile = file.replace('setup.py', 'version.txt')
        def versionString = sh(returnStdout: true, script: "cat ${versionFile}")
        if (versionString) {
            result['version'] = versionString.trim()
        }
        echo "loaded ${result} from ${file} and ${versionFile}"
    }else{
        echo "loaded ${result} from ${file}"
    }

    return result
}

Map getVersionElements(version) {
    //triggerPattern e.g. like '.* /piper ([a-z]*) .*'
    def matcher = version =~ '(\\d+)\\.?(\\d+)?\\.?(\\d+)?[-\\.]?(\\d+)?'
    if (matcher.size() == 0) {
        return null
    }

    return [all: version, full: matcher[0][0], major: matcher[0][1], minor: matcher[0][2], patch: matcher[0][3], timestamp: matcher[0][4]]
}

@NonCPS
def matches(Pattern pattern, String input, int group) {
    def m = pattern.matcher(input)
    return m.matches() ? m.group(group) : ''
}

@NonCPS
def matches(String regex, String input, int group) {
    def m = new Matcher(regex, input)
    return m.matches() ? m.group(group) : ''
}

@NonCPS
def matches(regex, input) {
   return matches(regex, input, 2)
}

void rewriteSettings(Object script, String artifactUrl, String newSettingsFile, String oldSettingsFile) {
    def settingsFile = script.sh(returnStdout: true, script: "cat ${oldSettingsFile}")
    if(settingsFile != null) {
        def rootNodeAsText = manipulateXml(settingsFile, artifactUrl)
        script.writeFile file: newSettingsFile, text: rootNodeAsText
    }
}

@NonCPS
def manipulateXml(String settingsFile, String artifactUrl) {
    def rootNode  = new XmlSlurper(false,false).parseText(settingsFile)
    def newFragmentNode = new XmlSlurper(false,false).parseText("<repository><id>staging</id><url>${artifactUrl}</url></repository>")
    rootNode.profiles.profile[0].repositories.appendNode(newFragmentNode)
    def rootNodeAsString = XmlUtil.serialize(rootNode)
    rootNode = null
    newFragmentNode = null
    return rootNodeAsString
}

/* Process order is
 * 1° Reading cfg/VERSION file containing directly a string describing the version
 * 2° search for a version field in the xmake section of cfg/xmake.cfg
 * 3° search for a version field in the xmake section of .xmake.cfg
 * MissingPropertyException is returned if no version found
 * if no file found, version will be saved in cfg/VERSION file
 */
def getXmakeVersion(filePath='') {
    def versionFile=filePath+'cfg/VERSION'
    if(fileExists(versionFile)) {
        return readFile(versionFile).trim()
    } else {
        try {
            return getXmakeCfgField("xmake", "version", filePath)
        } catch(MissingPropertyException e) {
            throw new MissingPropertyException("no cfg/VERSION file or version field in cfg/xmake.cfg or .xmake.cfg")
        }
    }
}

def setXmakeVersion(version, filePath='', buildArgs=[:]) {
    dir(filePath+'cfg') {
        writeFile file:'VERSION', text: version
    }
    if (buildArgs?.size()>0)
        updateXMakeDockerBuildArgs(version, buildArgs)
}

def getSbtVersion(filePath) {
    def sbtDescriptorJson = readJSON file: filePath
    return sbtDescriptorJson.version
}

def setSbtVersion(version, filePath) {
    def sbtDescriptorJson = readJSON file: filePath
    sbtDescriptorJson.version = new String(version)
    writeJSON file: filePath, json: sbtDescriptorJson
}

def getMtaVersion(filePath = 'mta.yaml'){
    def mtaYaml = readYaml file: filePath
    return mtaYaml.version
}

def setMtaVersion(baseVersion, newVersion, filePath = 'mta.yaml'){
    def search = "version: ${baseVersion}"
    def replacement = "version: ${newVersion}"
    sh "sed -i 's/${search}/${replacement}/g' ${filePath}"
}

def getMavenVersion(filePath) {
    def mavenPom = readMavenPom (file: filePath)
    return mavenPom.getVersion()
}

def getNpmVersion(filePath) {
    def packageJson = readJSON file: filePath
    return packageJson.version
}

def setNpmVersion(version, filePath) {
    def packageJson = readJSON file: filePath
    packageJson.version = new String(version)
    writeJSON file: filePath, json: packageJson
}

def getPythonVersion(filePath) {
    def lines = readFile(filePath).split('\n')
    return lines[0].trim()
}

def setPythonVersion(version, filePath) {
    writeFile file: filePath, text: version
}

def getGolangVersion(filePath) {
    def lines = readFile(filePath).split('\n')
    return lines[0].trim()
}

def setGolangVersion(version, filePath) {
    writeFile file: filePath, text:version
}

def getDockerVersion(filePath, dockerVersionSource) {
    if (dockerVersionSource) {
        if (dockerVersionSource == 'FROM'){
            return getVersionFromDockerBaseImageTag(filePath)
        }else{
            return getVersionFromDockerEnvVariable(filePath, dockerVersionSource)
        }
    } else {
        return getXmakeVersion('')
    }
}

def getVersionFromDockerEnvVariable(filePath, name) {
    def lines = readFile(filePath).split('\n')
    def version = ''
    for (def i = 0; i < lines.size(); i++) {
        if (lines[i].startsWith('ENV') && lines[i].split(' ')[1] == name) {
            version = lines[i].split(' ')[2]
            break
        }
    }
    echo "Version from Docker environment ${name}: ${version}"
    return version.trim()
}

def getVersionFromDockerBaseImageTag(filePath) {
    def lines = readFile(filePath).split('\n')
    def version = null
    for (def i = 0; i < lines.size(); i++) {
        if (lines[i].startsWith('FROM') && lines[i].indexOf(':') > 0) {
            version = lines[i].split(':')[1]
            break
        }
    }
    echo "Version from Docker base image tag: ${version}"
    return version.trim()
}

/* Will return a map containing Properties. Each section name is a map entry
 * on no sections part, section name will be "_java_properties_"
 */
def readIniFile(file) {
    def ini = [:]

    def section="_java_properties_"
    def content=""

    def lines=readFile(file).readLines()
    for(def pos=0;pos<lines.size;pos++) {
        def line=lines[pos]
        def m = (line =~ /\s*\[\s*(.*)\s*\]\s*/)
        if(m) {
            if(content.isEmpty()==false) {
                final Properties p = new Properties();
                p.load(new StringReader(content));
                ini.put(section,p)
                content=""
            }
            section=m.group(1)
        } else {
            line=line.trim()
            if(line.isEmpty()==false) content+=(line+"\n")
        }
    }
    if(content.isEmpty()==false) {
        final Properties p = new Properties();
        p.load(new StringReader(content));
        ini.put(section,p)
    }
    return ini
}

def getXmakeCfgField(section, field, filePath="") {
    try {
        def file=filePath+'cfg/xmake.cfg'
        if(!fileExists(file)) throw new MissingPropertyException("")
        def ini=readIniFile(file)
        if(ini.get(section)==null) throw new MissingPropertyException("")
        return ini.get(section)[field]
    } catch(MissingPropertyException) {
        try {
            def file=filePath+'.xmake.cfg'
            if(!fileExists(file)) throw new MissingPropertyException("")
            def ini=readIniFile(file)
            if(ini.get(section)==null) throw new MissingPropertyException("")
            return ini.get(section)[field]
        } catch(MissingPropertyException e) {
        }
    }
    throw new MissingPropertyException("no ${field} field in section ${section} of cfg/xmake.cfg or .xmake.cfg")
}

def updateXMakeDockerBuildArgs(version, buildArgs) {
    def filePath = ''

    if (fileExists('cfg/xmake.cfg'))
        filePath = 'cfg/xmake.cfg'
    else if (fileExists('.xmake.cfg'))
        filePath = '.xmake.cfg'
    if (!filePath.isEmpty()) {
        def lines = readFile(filePath).tokenize('\n')
        def result = ''
        lines.each {line ->
            if (line.startsWith('options=')) {
                def newOptions = line
                buildArgs.each {arg ->
                    def queryString = "--build-arg ${arg.key}="
                    def target = "--build-arg ${arg.key}=${arg.value}"
                    def matcher = newOptions =~ /$queryString[\S]*/
                    if (matcher.getCount() > 0){
                        newOptions = newOptions.replaceFirst(/$queryString[\S]*/, target)
                    }else{
                        newOptions = newOptions + ' ' + target
                    }
                }
                result = result + newOptions + '\n'
            } else if (line.startsWith('version=')){
                result = result + "version=${version}" + '\n'
            } else {
                result = result + line + '\n'
            }
        }
        writeFile file: filePath, text: result
    }
}

def printStepParameters(configMap, parameterNames) {
    def sb = new StringBuilder()
    sb.append("Parameters").append(": ")
    def count = 0
    parameterNames.each { name ->
        if(count > 0)
            sb.append(", ")
        sb.append(name).append(": ")
        def value = configMap[name]
        if(value instanceof Map) {
            sb.append("[")
            if (value.values().size() > 0) {
                value.entrySet().each { entry ->
                    sb.append(entry.key).append(":").append(entry.value)
                }
            } else {
                sb.append(":")
            }
            sb.append("]")
        } else {
            sb.append(value)
        }
        count++
    }
    return sb.toString()
}


void unstableIfCondition(Script step, boolean condition, String message, Closure body) {
    try {
        body()
    } catch (err) {
        if (condition) {
            Notify.warning(step, "${message}: ${err}")
            unstable(err.message)
        } else {
            throw err
        }
    }
}

package com.sap.piper.internal

import com.cloudbees.groovy.cps.NonCPS
import hudson.FilePath
import hudson.Functions
import hudson.tasks.junit.TestResultAction
import hudson.util.Secret
import jenkins.model.Jenkins
import org.jenkinsci.plugins.workflow.libs.LibrariesAction
import org.jenkinsci.plugins.workflow.steps.MissingContextVariableException


@NonCPS
static boolean addWarningsParser(Script script, Map parserSettings){
    def descriptorClass
    try{
        descriptorClass = script.class.classLoader.loadClass('hudson.plugins.warnings.WarningsDescriptor', true, false)
    }catch(java.lang.ClassNotFoundException e){
        script.echo "Could not load warnings descriptor class ($e)"
        return false
    }
    def isMissing = true
    def warningsSettings = Jenkins.instance.getExtensionList(descriptorClass)[0]

    warningsSettings.getParsers().each{ parser ->
        if (parser.getName() == parserSettings.parserName) isMissing = false
    }

    if(isMissing){
        def parserClass
        try{
            parserClass = script.class.classLoader.loadClass('hudson.plugins.warnings.GroovyParser', true, false)
        }catch(java.lang.ClassNotFoundException e){
            script.echo "Could not load parser class ($e)"
            return false
        }

        warningsSettings.addGroovyParser(
            parserClass?.newInstance(
                parserSettings.parserName,
                parserSettings.parserRegexp,
                parserSettings.parserScript,
                parserSettings.parserExample,
                parserSettings.parserLinkName,
                parserSettings.parserTrendName
            )
        )
        return true
    }
    return false
}

@NonCPS
static boolean hasTestFailures(build){
    //build: https://javadoc.jenkins.io/plugin/workflow-support/org/jenkinsci/plugins/workflow/support/steps/build/RunWrapper.html
    //getRawBuild: https://javadoc.jenkins.io/plugin/workflow-job/org/jenkinsci/plugins/workflow/job/WorkflowRun.html
    //getAction: http://www.hudson-ci.org/javadoc/hudson/tasks/junit/TestResultAction.html
    def action = build?.getRawBuild()?.getAction(TestResultAction.class)
    return action && action.getFailCount() != 0
}

@NonCPS
def getCurrentBuildInstance() {
    return currentBuild
}

@NonCPS
def getRawBuild() {
    return getCurrentBuildInstance().rawBuild
}

@NonCPS
def getParentJob() {
    return getRawBuild().getParent()
}

@NonCPS
def getActiveJenkinsInstance() {
    return Jenkins.getActiveInstance()
}

@NonCPS
void addGlobalSideBarLink(String url, String displayName, String relativeIconPath) {
    try {
        def linkActionClass = this.class.classLoader.loadClass("hudson.plugins.sidebar_link.LinkAction")
        if (url != null && displayName != null) {
            def iconPath = (null != relativeIconPath) ? "${Functions.getResourcePath()}/${relativeIconPath}" : null
            def action = linkActionClass.newInstance(url, displayName, iconPath)
            echo "Added Jenkins global sidebar link to '$url' with name '$displayName' and icon '$iconPath'"
            getActiveJenkinsInstance().getActions().add(action)
        }
    } catch (e) {
        e.printStackTrace()
    }
}

@NonCPS
void removeGlobalSideBarLinks(String url = null) {
    try {
        def linkActionClass = this.class.classLoader.loadClass("hudson.plugins.sidebar_link.LinkAction")
        def listToRemove = new ArrayList()
        getActiveJenkinsInstance()?.getActions()?.each {
            action ->
                if (linkActionClass.isAssignableFrom(action.getClass()) && (null == url || action.getUrlName().endsWith(url))) {
                    listToRemove.add(action)
                }
        }
        getActiveJenkinsInstance().getActions().removeAll(listToRemove)
        echo "Removed Jenkins global sidebar links ${listToRemove}"
    } catch (Exception e) {
        e.printStackTrace()
    }
}

@NonCPS
void addJobSideBarLink(String relativeUrl, String displayName, String relativeIconPath) {
    try {
        def linkActionClass = this.class.classLoader.loadClass("hudson.plugins.sidebar_link.LinkAction")
        if (relativeUrl != null && displayName != null) {
            def parentJob = getParentJob()
            def buildNumber = getCurrentBuildInstance().number
            def iconPath = (null != relativeIconPath) ? "${Functions.getResourcePath()}/${relativeIconPath}" : null
            def action = linkActionClass.newInstance("${buildNumber}/${relativeUrl}", displayName, iconPath)
            echo "Added job level sidebar link to '${action.getUrlName()}' with name '${action.getDisplayName()}' and icon '${action.getIconFileName()}'"
            parentJob.getActions().add(action)
        }
    } catch (e) {
        e.printStackTrace()
    }
}

@NonCPS
void removeJobSideBarLinks(String relativeUrl = null) {
    try {
        def linkActionClass = this.class.classLoader.loadClass("hudson.plugins.sidebar_link.LinkAction")
        def parentJob = getParentJob()
        def listToRemove = new ArrayList()
        parentJob?.getActions()?.each {
            action ->
                if (linkActionClass.isAssignableFrom(action.getClass()) && (null == relativeUrl || action.getUrlName().endsWith(relativeUrl))) {
                    echo "Removing job level sidebar link to '${action.getUrlName()}' with name '${action.getDisplayName()}' and icon '${action.getIconFileName()}'"
                    listToRemove.add(action)
                }
        }
        parentJob.getActions().removeAll(listToRemove)
        echo "Removed Jenkins global sidebar links ${listToRemove}"
    } catch (e) {
        e.printStackTrace()
    }
}

@NonCPS
void addRunSideBarLink(String relativeUrl, String displayName, String relativeIconPath) {
    try {
        def linkActionClass = this.class.classLoader.loadClass("hudson.plugins.sidebar_link.LinkAction")
        if (relativeUrl != null && displayName != null) {
            def run = getRawBuild()
            def iconPath = (null != relativeIconPath) ? "${Functions.getResourcePath()}/${relativeIconPath}" : null
            def action = linkActionClass.newInstance(relativeUrl, displayName, iconPath)
            echo "Added run level sidebar link to '${action.getUrlName()}' with name '${action.getDisplayName()}' and icon '${action.getIconFileName()}'"
            run.getActions().add(action)
        }
    } catch (e) {
        e.printStackTrace()
    }
}

@NonCPS
def isJobStartedByTimer() {
    return isJobStartedByCause(hudson.triggers.TimerTrigger.TimerTriggerCause.class)
}

@NonCPS
def isJobStartedByUser() {
    return isJobStartedByCause(hudson.model.Cause.UserIdCause.class)
}

@NonCPS
def isJobStartedByCause(Class cause) {
    def startedByGivenCause = false
    def detectedCause = getRawBuild().getCause(cause)
    if (null != detectedCause) {
        startedByGivenCause = true
        echo "Found build cause ${detectedCause}"
    }
    return startedByGivenCause
}

@NonCPS
void scheduleJob(String spec) {
    if (spec == null || spec.length() == 0) {
        echo "scheduleJob - Missing TimerTrigger spec, therefore nothing to schedule"
        return
    }

    def parentJob = getParentJob()
    def pipelineTriggersJobProperty = parentJob.getTriggersJobProperty()

    // Check for existing TimerTrigger
    def foundTrigger = false
    def specChanged = false
    def newParts = spec.tokenize("\n")
    def newTriggers = new ArrayList<hudson.triggers.Trigger<?>>()
    pipelineTriggersJobProperty?.getTriggers()?.each {
        item ->
            if (item instanceof hudson.triggers.TimerTrigger) {
                if(foundTrigger){
                    echo "scheduleJob - Omit additional TimerTrigger"
                    return
                }
                foundTrigger = true
                def newSpec = item.getSpec()
                newParts?.each {
                    newPart ->
                        if (!newSpec.contains(newPart)) {
                            newSpec += "\n" + newPart
                            specChanged = true
                        }
                }
                newSpec = newSpec.trim()
                if (specChanged == true) {
                    def timerTrigger = new hudson.triggers.TimerTrigger(newSpec)
                    newTriggers.add(timerTrigger)
                    echo "scheduleJob - Modified existing TimerTrigger to spec '${newSpec}'"
                } else {
                    echo "scheduleJob - Nothing to do, existing job with same spec found"
                }
                return
            } else {
                newTriggers.add(item)
            }
    }

    if (foundTrigger == false) {
        def timerTrigger = new hudson.triggers.TimerTrigger(spec)
        newTriggers.add(timerTrigger)
        echo "scheduleJob - Added new TimerTrigger with spec '${spec}'"
    }

    if (foundTrigger == false || specChanged == true)
        parentJob.setTriggers(newTriggers)
}

@NonCPS
void removeJobSchedule(String spec = null) {
    def parentJob = getParentJob()
    def pipelineTriggersJobProperty = parentJob.getTriggersJobProperty()

    def existingTriggers = pipelineTriggersJobProperty.getTriggers()
    def newTriggers = new ArrayList<hudson.triggers.Trigger<?>>()
    existingTriggers?.each {
        item ->
            if (item instanceof hudson.triggers.TimerTrigger) {
                if (spec != null) {
                    def existingSpecParts = item.getSpec().split("\n")
                    def partsToRemove = spec.split("\n")
                    def existingSpec = ""
                    existingSpecParts?.each {
                        existingPart ->
                            def remove = false
                            partsToRemove?.each {
                                partToRemove ->
                                    if (existingPart.trim().equals(partToRemove.trim()))
                                        remove = true
                            }
                            if (!remove) {
                                existingSpec = existingSpec + "\n" + existingPart
                            }
                    }
                    if (existingSpec.length() > 0) {
                        def timerTrigger = new hudson.triggers.TimerTrigger(existingSpec.trim())
                        newTriggers.add(timerTrigger)
                        echo "removeJobSchedule - Modified existing TimerTrigger to spec '${existingSpec}'"
                    } else {
                        echo "removeJobSchedule - Removed existing TimerTrigger with spec '${item.getSpec()}'"
                    }
                } else {
                    echo "removeJobSchedule - Removed existing TimerTrigger with spec '${item.getSpec()}'"
                }
            } else {
                newTriggers.add(item)
            }
    }
    parentJob.setTriggers(newTriggers)
}

@NonCPS
void addBuildDiscarder(int daysToKeep = -1, int numToKeep = -1, int artifactDaysToKeep = -1, int artifactNumToKeep = -1) {
    def parentJob = getParentJob()
    def oldDiscarder = parentJob.getBuildDiscarder()

    def foundExisting = false
    if (oldDiscarder instanceof hudson.tasks.LogRotator
        && oldDiscarder.getDaysToKeep() == daysToKeep
        && oldDiscarder.getNumToKeep() == numToKeep
        && oldDiscarder.getArtifactDaysToKeep() == artifactDaysToKeep
        && oldDiscarder.getArtifactNumToKeep() == artifactNumToKeep) {
        echo "Found existing LogRotator with same spec, nothing to do"
        foundExisting = true
    }

    if (!foundExisting) {
        def newDiscarder = new hudson.tasks.LogRotator(daysToKeep, numToKeep, artifactDaysToKeep, artifactNumToKeep)
        parentJob.setBuildDiscarder(newDiscarder)
        echo "Set new LogRotator with spec daysToKeep: ${daysToKeep}, numToKeep: ${numToKeep}, artifactDaysToKeep: ${artifactDaysToKeep}, artifactNumToKeep: ${artifactNumToKeep}"
    }
}

@NonCPS
void removeBuildDiscarder() {
    def parentJob = getParentJob()
    if (null != parentJob.getBuildDiscarder()) {
        parentJob.setBuildDiscarder(null)
        echo "Removed existing BuildDiscarder"
    }
}

def getLibrariesInfoWithPiperLatest() {
    def libraries = []
    try {
        libraries = getLibrariesInfo()
        def versionFragment = sh(returnStdout: true, script: 'git ls-remote https://github.wdf.sap.corp/ContinuousDelivery/piper-library.git refs/heads/master')
        def version = "${null != versionFragment ? versionFragment.trim().split()[0] : ''}"
        libraries.add([name: 'piper-lib@master', version: version])
    } catch (MissingContextVariableException noNode) {
        echo "Determination of Piper library version skipped, no node available!"
    } catch (ignored) {}
    return libraries
}

def nodeAvailable() {
    try {
        sh "echo 'Node is available!'"
    } catch (MissingContextVariableException e) {
        echo "No node context"
        return false
    }
    return true
}

@NonCPS
def getLibrariesInfo() {
    def libraries = []
    def build = getRawBuild()
    def libs = build.getAction(LibrariesAction.class).getLibraries()

    for (def i = 0; i < libs.size(); i++) {
        Map lib = [:]

        lib['name'] = libs[i].name
        lib['version'] = libs[i].version
        lib['trusted'] = libs[i].trusted
        libraries.add(lib)
    }

    return libraries
}

@NonCPS
FilePath getLibraryPath(String libraryName) {
    def build = getRawBuild()
    FilePath libDir = new FilePath(build.getRootDir()).child("libs/" + libraryName)
    return libDir
}

def getLastLogLine() {
    //sleep required due to log synchronization - 2 secs is current estimate which proved to be working
    sleep 2
    //filter out sleep entrry in log
    def logLine = getLogLines(4)[1]
    return logLine
}

@NonCPS
def getLogLines(lines) {
    getRawBuild().getLog(lines)
}


def encryptPassword(String password) {
    def encyptedPassword = Secret.fromString(password).getEncryptedValue()
    return encyptedPassword
}

@NonCPS
def getRequirementResultMapping(mapping) {
    def result = getRawBuild().getAction(hudson.tasks.junit.TestResultAction.class).getResult()
    return getMappingWithTestResults(mapping, result)
}

@NonCPS
def getMappingWithTestResults(mapping, result) {
    result.each { child ->
        if (child.class.getName() == "hudson.tasks.junit.ClassResult") {
            mapping.each {
                if (it.source_reference == child.getFullName()) {
                    it.test_cases = []
                    for (def i=0; i<child.getChildren().size(); i++) {
                        def caseResult = [
                            test_fullname: child.getChildren()[i].getFullName(),
                            test_name: child.getChildren()[i].getName(),
                            test_class: child.getChildren()[i].getClassName(),
                            passed: child.getChildren()[i].isPassed(),
                            skipped: child.getChildren()[i].isSkipped()
                        ]
                        it.test_cases.add(caseResult)
                    }
                }
            }
            mapping = getMappingWithTestResults(mapping, child.getChildren())
        } else if (child.class.getName() == "hudson.tasks.junit.CaseResult") {
            mapping.each {
                if (it.source_reference == child.getFullName() || it.source_reference.startsWith(child.getFullName().split(/(\([\w, ]+\)|\{[\w, ]+\})\[\d+\]/)[0] + "(")) {
                    if (it.test_cases == null) {
                        it.test_cases = []
                    }
                    def caseResult = [
                        test_fullname: child.getFullName(),
                        test_name: child.getName(),
                        test_class: child.getClassName(),
                        passed: child.isPassed(),
                        skipped: child.isSkipped()
                    ]
                    it.test_cases.add(caseResult)
                }
            }
        } else if (child.class.getName() == "hudson.tasks.junit.PackageResult") {
            mapping = getMappingWithTestResults(mapping, child.getChildren())
        } else {
            mapping = getMappingWithTestResults(mapping, child.getChildren())
        }
    }
    return mapping
}

@NonCPS
def getPlugin(name){
    def result = null
    getActiveJenkinsInstance().getPluginManager().getPlugins().each {
        plugin ->
            if (name == plugin.getShortName()) {
                result = plugin
                return
            }
    }
    return result
}

@NonCPS
boolean isOldKubePluginVersion() {
    def plugin = getPlugin('kubernetes')

    if (plugin && plugin.getVersion().startsWith('0.')) {
        echo "[${STEP_NAME}] DEPRECATED: You are using an old version of the Kubernetes-Plugin. Please update to a version >= 1.0!"
        plugin = null
        return true
    }
    return false
}


def getCheckmarxResults() {
    def resultMap = [
        High: [
            Issues: 0,
            NotFalsePositive: 0,
            NotExploitable: 0,
            Confirmed: 0,
            Urgent: 0,
            ProposedNotExploitable: 0,
            ToVerify: 0
        ],
        Medium: [
            Issues: 0,
            NotFalsePositive: 0,
            NotExploitable: 0,
            Confirmed: 0,
            Urgent: 0,
            ProposedNotExploitable: 0,
            ToVerify: 0
        ],
        Low: [
            Issues: 0,
            NotFalsePositive: 0,
            NotExploitable: 0,
            Confirmed: 0,
            Urgent: 0,
            ProposedNotExploitable: 0,
            ToVerify: 0
        ],
        Information: [
            Issues: 0,
            NotFalsePositive: 0,
            NotExploitable: 0,
            Confirmed: 0,
            Urgent: 0,
            ProposedNotExploitable: 0,
            ToVerify: 0
        ]]
    def sastResultFiles = findFiles(glob: "**/ScanReport.xml")

    if (sastResultFiles.size() == 0 ) {
        error 'No Checkmarx result available, aborting ...'
    }
    def sastResultFile = sastResultFiles[0]

    def xmlFileContent = readFile(file: sastResultFile.path).replace("\uFEFF", "")
    readScanResultXmlIntoMap(xmlFileContent, resultMap)
    return  resultMap
}

@NonCPS
private void readScanResultXmlIntoMap(xmlFileContent, resultMap) {
    def xml = new XmlSlurper(false, false).parseText(xmlFileContent)
    def summary = xml?.getAt(0)
    if (summary) {
        resultMap.InitiatorName = summary?.attributes?.InitiatorName?.toString()
        resultMap.Owner = summary?.attributes?.Owner?.toString()
        resultMap.ScanId = summary?.attributes?.ScanId?.toString()
        resultMap.ProjectId = summary?.attributes?.ProjectId?.toString()
        resultMap.ProjectName = summary?.attributes?.ProjectName?.toString()
        resultMap.Team = summary?.attributes?.Team?.toString()
        resultMap.TeamFullPathOnReportDate = summary?.attributes?.TeamFullPathOnReportDate?.toString()
        resultMap.ScanStart = summary?.attributes?.ScanStart?.toString()
        resultMap.ScanTime = summary?.attributes?.ScanTime?.toString()
        resultMap.LinesOfCodeScanned = summary?.attributes?.LinesOfCodeScanned?.toString()
        resultMap.FilesScanned = summary?.attributes?.FilesScanned?.toString()
        resultMap.CheckmarxVersion = summary?.attributes?.CheckmarxVersion?.toString()
        resultMap.ScanType = summary?.attributes?.ScanType?.toString()
        resultMap.Preset = summary?.attributes?.Preset?.toString()
        resultMap.DeepLink = summary?.attributes?.DeepLink?.toString()
        resultMap.ReportCreationTime = summary?.attributes?.ReportCreationTime?.toString()
        summary?.children?.each { query ->
            query?.children?.each { result ->
                def key = result?.attributes?.Severity?.toString()
                def issueCount = resultMap[key]['Issues']
                resultMap[key]['Issues'] = ++issueCount

                def auditState
                switch (result?.attributes?.state?.toString()) {
                    case '1':
                        auditState = "NotExploitable"
                        break
                    case '2':
                        auditState = "Confirmed"
                        break
                    case '3':
                        auditState = "Urgent"
                        break
                    case '4':
                        auditState = "ProposedNotExploitable"
                        break
                    case '0':
                    default:
                        auditState = "ToVerify"
                        break
                }
                def stateCount = resultMap[key][auditState]
                resultMap[key][auditState] = ++stateCount

                if (result?.attributes?.FalsePositive?.toString() != 'True') {
                    def falsePositiveCount = resultMap[key]['NotFalsePositive']
                    resultMap[key]['NotFalsePositive'] = ++falsePositiveCount
                }
            }
        }
    }
}

@NonCPS
String getIssueCommentTriggerAction() {
    try {
        def triggerCause = getRawBuild().getCause(org.jenkinsci.plugins.pipeline.github.trigger.IssueCommentCause)
        if (triggerCause) {
            //triggerPattern e.g. like '.* /piper ([a-z]*) .*'
            def matcher = triggerCause.comment =~ triggerCause.triggerPattern
            if (matcher) {
                return matcher[0][1]
            }
        }
        return null
    } catch (err) {
        return null
    }
}

@NonCPS
boolean pipelineIsScheduled() {
    def causes = getCurrentBuildInstance().getBuildCauses()
    boolean scheduled = false
    causes.each {cause ->
        if (cause.shortDescription == 'Started by timer') {
            scheduled = true
        }
    }
    return scheduled
}

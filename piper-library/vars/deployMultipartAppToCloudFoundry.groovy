import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.BashUtils
import com.sap.piper.internal.ConfigurationHelper
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'deployMultipartAppToCloudFoundry'
@Field Set GENERAL_CONFIG_KEYS = [
    'cfApiEndpoint',
    'cfCredentialsId',
    'cfOrg',
    'cfDomain',
    'cfSpace',
    'deployType',
    'dockerImage',
    'dockerWorkspace',
    'modules'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

def call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('cfCredentialsId')
            .withMandatoryProperty('cfOrg')
            .withMandatoryProperty('cfSpace')
            .use()

        config.formattedSpaceName = config.cfSpace.toLowerCase().replaceAll("_", "-")

        withCredentials([usernamePassword(credentialsId: config.cfCredentialsId, passwordVariable: 'cfPassword', usernameVariable: 'cfUser')]) {
            dockerExecute(script: script, dockerImage: config.dockerImage, dockerWorkspace: config.dockerWorkspace) {
                unstash 'deployDescriptor'
                login(config.cfApiEndpoint, config.cfOrg, config.cfSpace)
                pushAll(config.modules, config.formattedSpaceName, config.cfDomain, config.deployType)
                bindProductiveRoutesToNewApps(config.modules, config.formattedSpaceName, config.cfDomain, config.cfOrg, config.deployType)
                disconnectOldAppsAndCleanup(config.modules, config.formattedSpaceName, config.cfDomain, config.deployType)
                logout()
            }
        }
    }
}

def isAppExisting(cfAppName) {
    def retVal = sh script: "cf app ${cfAppName}", returnStatus: true
    if (retVal == 0) {
        return true
    } else {
        return false
    }
}

def pushAll(modules, formattedSpaceName, cfDomain, deployType) {
    def numberOfModules = modules.size()
    def i = 1
    writeLog("Starting to push ${numberOfModules} parts.")
    for (Map module : modules) {
        def cfAppName = module.cfAppName
        def cfManifestPath = module.cfManifestPath
        def cfHostname = module.get('cfHostname', "${cfAppName}-${formattedSpaceName}")
        def cfEnvVariables = module.get('cfEnvVariables', [])
        def registerAsServiceBroker = module.get('registerAsServiceBroker', false)
        def serviceBrokerHasSpaceScope = module.get('serviceBrokerHasSpaceScope', true)
        def actualDeployType = module.get('deployType', deployType)
        def actualDomain = module.get('cfDomain', cfDomain)

        cfManifestPath = replaceReferencesToOtherModulesInManifest(cfManifestPath, modules, formattedSpaceName, cfDomain)

        writeLog("Pushing ${cfAppName} from file ${cfManifestPath}. (Part ${i}/${numberOfModules})")

        if (actualDeployType == 'blue-green' && registerAsServiceBroker == false) {
            cfAppName = getAppNameForNewAppInBlueGreenDeployment(cfAppName)
            cfHostname = getHostnameForNewAppInBlueGreenDeployment(cfHostname)
        }

        if (registerAsServiceBroker == true) {
            if (isServiceBrokerAlreadyRegistered(cfAppName, formattedSpaceName)) {
                cfPush(cfAppName, cfHostname, cfManifestPath, actualDomain)
            } else {
                cfPush(cfAppName, cfHostname, cfManifestPath, actualDomain, true)
                registerServiceBroker(cfAppName, cfHostname, formattedSpaceName, serviceBrokerHasSpaceScope, cfDomain)
            }
        } else {
            if (cfEnvVariables.size() > 0) {
                cfPush(cfAppName, cfHostname, cfManifestPath, actualDomain, true)
                def keySet = cfEnvVariables.keySet() //for each didn't work here - jenkins bug
                echo "${keySet.size()} variables to set for ${cfAppName}"
                def j = 0
                while (j < keySet.size()) {
                    def variable = keySet[j]
                    def value = cfEnvVariables[variable]
                    if (value instanceof String || value instanceof GString) {
                        value = value.replaceAll('"', '\\\\"')
                        value = value.replaceAll('\n', '')
                    }
                    sh script: "cf set-env ${cfAppName} ${variable} \"${value}\""
                    j = j + 1
                }
                sh "cf start ${cfAppName}"
            } else {
                cfPush(cfAppName, cfHostname, cfManifestPath, actualDomain)
            }
        }
        i++
    }
}



def cfPush(cfAppName, cfHostname, cfManifestPath, cfDomain, doNotStart = false) {
    def noStartOption = ''
    if (doNotStart == true) {
        noStartOption = '--no-start'
    }
    sh "cf push ${cfAppName} -n ${cfHostname} -f ${cfManifestPath} -d ${cfDomain} ${noStartOption}"
}

def replaceReferencesToOtherModulesInManifest(cfManifestPath, modules, formattedSpaceName, cfDomain) {
    def manifestContent = readFile file: cfManifestPath
    def updatedManifestPath = cfManifestPath
    for (Map otherModule : modules) {
        if (otherModule.cfManifestPath != cfManifestPath) {
            def cfOtherAppName = otherModule.cfAppName
            def cfOtherHostname = otherModule.get('cfHostname', "${cfOtherAppName}-${formattedSpaceName}")
            def cfOtherDomain = otherModule.get('cfDomain', cfDomain)
            def otherUrl = "https://${cfOtherHostname}.${cfOtherDomain}"
            if (manifestContent.contains("\${${cfOtherAppName}-url")) {
                manifestContent = manifestContent.replaceAll(/\$\{${cfOtherAppName}-url\}/, otherUrl)
                updatedManifestPath = createFileNameForUpdatedRessource(cfManifestPath)
                writeLog("The following changed manifest content will be written to ${updatedManifestPath}: ${manifestContent}")
                writeFile file: updatedManifestPath, text: manifestContent
            }
        }
    }
    return updatedManifestPath
}

def createFileNameForUpdatedRessource(path) {
    def i = path.lastIndexOf(".")
    if (i == -1) {
        i = path.length()
    }
    path = path.substring(0, i) + "_updated" + path.substring(i, path.length())
    return path
}

def bindProductiveRoutesToNewApps(modules, formattedSpaceName, cfDomain, cfOrg, deployType) {
    def numberOfModules = modules.size()
    def i = 1
    writeLog("All parts have been pushed. Starting with route-mapping.")
    for (Map module : modules) {
        def cfAppName = module.cfAppName
        def cfHostname = module.get('cfHostname', "${cfAppName}-${formattedSpaceName}")
        def actualDomain = module.get('cfDomain', cfDomain)
        def actualDeployType = module.get('deployType', deployType)
        def registerAsServiceBroker = module.get('registerAsServiceBroker', false)

        writeLog("Mapping the route for ${cfAppName}. (Part ${i}/${numberOfModules})")
        createDomainIfNotAlreadyExisting(cfOrg, actualDomain)
        if (registerAsServiceBroker == true || actualDeployType == 'standard') {
            echo 'Not needed. No blue-green-deployment for this module.'
            continue
        }
        def cfNewAppName = getAppNameForNewAppInBlueGreenDeployment(cfAppName)
        sh "cf map-route ${cfNewAppName} ${actualDomain} -n '${cfHostname}'"
        i++
    }
}

def disconnectOldAppsAndCleanup(modules, formattedSpaceName, cfDomain, deployType) {
    def numberOfModules = modules.size()
    def i = 1
    for (Map module : modules) {
        def cfAppName = module.cfAppName
        def cfHostname = module.get('cfHostname', "${cfAppName}-${formattedSpaceName}")
        def registerAsServiceBroker = module.get('registerAsServiceBroker', false)
        def actualDomain = module.get('cfDomain', cfDomain)
        def actualDeployType = module.get('deployType', deployType)
        writeLog("Cleanup old version of ${cfAppName}. (Part ${i}/${numberOfModules})")
        if (registerAsServiceBroker == true || actualDeployType == 'standard') {
            echo 'Not needed. No blue-green-deployment for this module.'
            continue
        }
        def cfNewAppName = getAppNameForNewAppInBlueGreenDeployment(cfAppName)
        def cfNewHostname = getHostnameForNewAppInBlueGreenDeployment(cfHostname)
        if (isAppExisting(cfAppName)) {
            sh "cf unmap-route ${cfAppName} ${actualDomain} -n '${cfHostname}'"
            sh "cf delete-route ${actualDomain} -n ${cfNewHostname} -f"
            if (isAppExisting("${cfAppName}-old")) {
                sh "cf delete ${cfAppName}-old -f"
            }
            sh "cf rename ${cfAppName} ${cfAppName}-old"
            sh "cf rename ${cfNewAppName} ${cfAppName}"
            sh "cf delete ${cfAppName}-old -f"
        } else {
            sh "cf delete-route ${actualDomain} -n ${cfNewHostname} -f"
            sh "cf rename ${cfNewAppName} ${cfAppName}"
        }
        i++
    }
    writeLog("Finished deployment of ${numberOfModules} parts.")
}

def login(cfApiEndpoint, cfOrg, cfSpace) {
    writeLog("Log in to ${cfOrg} - ${cfSpace} as ${cfUser}.")
    sh "cf login -u ${BashUtils.escape(cfUser)} -p ${BashUtils.escape(cfPassword)} -a ${cfApiEndpoint} -o '${cfOrg}' -s '${cfSpace}'"
}

def logout() {
    writeLog('Log out from CF.')
    sh 'cf logout'
}

def writeLog(String message) {
    def timestamp = (new Date()).format("yyyy-MM-dd HH:mm:ss", TimeZone.getTimeZone('CET'))
    echo """########################################Deploy Multipart App################################################
    [${timestamp}] - ${message}
############################################################################################################"""
}

def isServiceBrokerAlreadyRegistered(cfAppName, formattedSpaceName) {
    def serviceBrokers = sh script: 'cf service-brokers', returnStdout: true
    if (serviceBrokers.contains("${cfAppName}-${formattedSpaceName}")) {
        return true
    } else {
        return false
    }
}

def registerServiceBroker(cfAppName, cfHostname, formattedSpaceName, isSpaceScoped, cfDomain) {
    def spaceScopedParameter = ''
    if (isSpaceScoped == true) {
        spaceScopedParameter = '--space-scoped'
    }
    def userName = "${cfAppName}-user"
    def generatedPassword = sh script: '</dev/urandom tr -dc A-Za-z0-9_ | head -c16', returnStdout: true

    sh """ cf set-env ${cfAppName} SBF_BROKER_CREDENTIALS \"{\\"${userName}\\" : \\"${
        generatedPassword
    }\\"}\"  """
    sh "cf set-env ${cfAppName} SBF_CATALOG_SUFFIX ${formattedSpaceName}"
    sh "cf start ${cfAppName}"
    sh "cf create-service-broker ${cfAppName}-${formattedSpaceName} ${userName} ${generatedPassword} https://${cfHostname}.${cfDomain} ${spaceScopedParameter}"
}

def createDomainIfNotAlreadyExisting(cfOrg, cfDomain) {
    def retVal = sh script: "cf create-domain ${cfOrg} ${cfDomain}", returnStatus: true
    if (retVal == 0) {
        echo 'Custom domain created.'
    } else {
        echo 'Domain already existing.'
    }
}

def getAppNameForNewAppInBlueGreenDeployment(cfAppName) {
    return cfAppName + '-new'
}

def getHostnameForNewAppInBlueGreenDeployment(cfHostname) {
    if (cfHostname == '*'){
        return 'new'
    }
    return cfHostname + '-new'
}

#!/usr/bin/env groovy

///////////////////////////////////////////////////////////////////////////////
//                                                                           //
//   Gate @ Piper into IRIS                                                  //
//                                                                           //
//   IRIS - An SAP inner source platform for central operations procedures   //
//                                                                           //
//     Check out project page:            https://go.sap.corp/iris           //
//               technical documentation: https://go.sap.corp/irisdoc        //
//                                                                           //
///////////////////////////////////////////////////////////////////////////////

import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import static com.sap.piper.internal.Prerequisites.checkScript

//import groovy.json.JsonException // It's a different class: "Error: net.sf.json.JSONException: Invalid JSON String"
import groovy.transform.Field
import hudson.AbortException

// Parameter/Config configuration
@Field String STEP_NAME = 'callIRIS'

@Field String LIBRARY_MANIFEST = 'IRISManifest.properties' // Name of Manifest inside of IRIS Libraries

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = [
    'dockerImage',                     // Location of Docker image
    'dockerWorkspace',                 // Main folder inside Docker image (based on Oliver Nocon -> not relevant)
    'irisManifest',                    // Manifest content (JSON)
    'irisManifestFile',                // Name of Manifest file inside Caller (like Jenkins-Workspace)
    'libraryFrameworkJenkinsName',     // Name of Library inside of Jenkins-Configuration (for indirect access, needed to be usable in Jenkins... Sandboxing)
    'libraryFrameworkRepository',      // Name of Repostiory in GitHub                    (for direct access)
    'libraryFrameworkOrganization',    // Name of Organization in GitHub                  (for direct access)
    'libraryContentBaseJenkinsName',
    'libraryContentBaseRepository',
    'libraryContentBaseOrganization',
    'libraryContentJenkinsName',
    'libraryContentRepository',
    'libraryContentOrganization',
    'useCaseVersion',                  // Use defined version of useCase means defined version of Content Repository
    'verbose'                          // Print more information in Log
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([ // UseCases
    'CLDReader',
    'Demultiplexer',
    'ITPCaller',
    'ITPStatusGetter',
    'ITPSynchronousCaller',
    'AvailabilityTester',
    'AvailabilityReader',
    'AvailabilityWriter',
    'HealthChecker',
    'BlueGreenCheck',
    'PollScript',
    'OpenStackDNSCreator',
    'SQLExecutor',
    'PingdomReader'
])

// Name of callable UseCase
@Field Map UseCase = [
    // Pipeline enhancement - distributed deployment
    CLDReader:            'CLDReader',            // Read SCP environment / instance information from SPC's CLD
    Demultiplexer:        'Demultiplexer',        // Enhancement of Piper to deploy your software to multiple environments / instances

    // "Jenkins -> SPC" integration
    ITPCaller:            'ITPCaller',            // Call SPC's IT processes via Service Request
    ITPStatusGetter:      'ITPStatusGetter',      // Get status of SPC's IT processes
    ITPSynchronousCaller: 'ITPSynchronousCaller', // Call SPC's IT processes via Service Request and wait

    // Pipeline enhancement - Smoke test
    AvailabilityTester:   'AvailabilityTester',   // Reuse availability metrics before switching from Green to Blue
    AvailabilityReader:   'AvailabilityReader',   // Read from AvS
    AvailabilityWriter:   'AvailabilityWriter',   // Write to AvS
    HealthChecker:        'HealthChecker',        // Reuse health checks before switching from Green to Blue
    BlueGreenCheck:       'BlueGreenCheck',       // Wrapper to init & call Blue Green Check
    PingdomReader:        'PingdomReader',

    // "CIS -> SPC" integration
    PollScript:           'PollScript',           // Call SPC's IT processes via Service Request with data sent by CIS

    // "Jenkins -> Monsoon3" automation
    OpenStackDNSCreator:  'OpenStackDNSCreator',   // Create DNS Zones in Monsoon3 from Jenkins

    // Automated SQL Execution
    SQLExecutor:          'SQLExecutor'
]

// Define SPC-System - see https://github.wdf.sap.corp/ProjectIRIS/iris/blob/master/src/com/sap/iris/SPC.groovy
@Field Map SPC = [
    Development: 'ACE',
    Test:        'VNS',
    Productive:  'NZA'
]

// Define Result - https://github.wdf.sap.corp/ProjectIRIS/iris/blob/master/src/com/sap/iris/Result.groovy
@Field Map Result = [
    Ok:       0, // Everything is fine
    Warning:  4, // Something happen, but not critical
    pError:   6, // Error for an optional thing happened
    Error:    8, // Error happened
    Abort:   12  // Hard error - is not possible to proceed
]

// Main program
def call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        // Common Piper header - instantiate script & utils
        def script = checkScript(this, parameters) ?: this

        def utils = parameters.juStabUtils ?: new Utils()

        // Common Piper header - Get and mix parameter
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('dockerImage')
            .withMandatoryProperty('dockerWorkspace')
            .use()

        // Determine useCase based on parameters handed over
        config.useCase = ''
        parameters.each{ key, value ->
            if (value instanceof Map) {
                if (config.useCase == '') {
                    config.useCase = key
                } else {
                    throw new IllegalArgumentException("${STEP_NAME}: ERROR - Already one useCase '${config.useCase}' defined. You cannot define another useCase '${key}' within same call!")
                }
            }
        }

        // Check if any useCase has been identified
        if (config.useCase == '') {
            throw new IllegalArgumentException("${STEP_NAME}: No useCase relevant parameter given!")
        }

        // Select and repack useCase specific parameter
        if (config[config.useCase]) {
            config.useCaseValue = config[config.useCase]
            config.remove(config.useCase)
        } else {
            config.useCaseValue = [:]
        }

        // If not given, use global 'verbose'
        config.useCaseValue.verbose = config.useCaseValue.get('verbose', config.get('verbose', false)).toBoolean()

        // HINT: Following Source Code needs to be in sync with docker.executeIRISDockerNoCredentials (IRIS Library)

        // Identify correct Version to use for Content Library
        String versionContent = 'master' // Default value

        // Version was handed over as Parameter
        if (config.useCaseVersion != '') {
            versionContent = config.useCaseVersion
            echo "Using Content Version '${versionContent}' for useCase '${config.useCase}' (Version passed as parameter)."

        } else { // Version seems to be inside Manifest content
            if (config.irisManifest != '') {
                versionContent = readVersionFromText(config.irisManifest, config.useCase, true)

                if (versionContent) {
                    if (versionContent != 'master') {
                        echo "Using Content Version '${versionContent}' for useCase '${config.useCase}' (Version read from Manifest parameter)."
                    } else {
                        echo "Using Content Version '${versionContent}' for useCase '${config.useCase}' (default Version)."
                    }
                }

            } else { // Version seems to be inside Manifest file
                String pathIRISManifest = ''

                // Check for Manifest file
                if (fileExists(config.irisManifestFile)) {
                    pathIRISManifest = config.irisManifestFile
                } else { // Pipeline script from SCM
                    if (fileExists("${pwd()}@script/${config.irisManifestFile}")) {
                        pathIRISManifest = "${pwd()}@script/${config.irisManifestFile}"
                    }
                }

                // Read Version from Manifest
                if (pathIRISManifest != '') {
                    versionContent = readVersionFromFile(pathIRISManifest, config.useCase, true)

                    if (versionContent) {
                        if (versionContent != 'master') {
                            echo "Using Content Version '${versionContent}' for useCase '${config.useCase}' (Version read from Manifest file '${pathIRISManifest}')."
                        } else {
                            echo "Using Content Version '${versionContent}' for useCase '${config.useCase}' (default Version)."
                        }
                    }
                }
            }
        }

        // Version could be read
        if (versionContent) {
            // Default Version found
            if (versionContent == 'master') {
                echo "Developer Notice: You did not specify a Version or simply 'master' for the useCase '${config.useCase}'. In future releases this will cause IRIS to fail." //Now it will just set Pipeline status to 'unstable'."// Documentation on how to set the useCaseVersion can be found: https://go.sap.corp/irisdoc/lib/setupLibrary/"
                //script.currentBuild.result = 'UNSTABLE'
                //throw new IllegalArgumentException("${STEP_NAME}: No useCaseVersion parameter given! Documentation on how to set the useCaseVersion can be found: https://go.sap.corp/irisdoc/lib/setupLibrary/")
            }

            // Split Version information for Reporting
            String versionMajor      = 'master'
            String versionMinor      = 'master'
            String versionCorrection = 'master'

            if (versionContent != 'master') {
                String[] versionContentSplit = versionContent.split('\\.')
                Integer versionContentSplitSize = versionContentSplit.size()

                versionMajor      = ((versionContentSplitSize>0) && versionContentSplit[0]) ?:'n/a'
                versionMinor      = ((versionContentSplitSize>1) && versionContentSplit[1]) ?:'n/a'
                versionCorrection = ((versionContentSplitSize>2) && versionContentSplit[2]) ?:'n/a'
            }

            // Save UseCase version details
            config.useCaseDetails = [useCase: config.useCase, versionMajor: versionMajor, versionMinor: versionMinor, versionCorrection: versionCorrection]

            // Load IRIS Content Library
            if (loadLibrary(config.libraryContentJenkinsName, versionContent)) {

                // Read Version information from Content to be used to load ContentBase Library
                String versionContentBase = readVersionFromFile("../workspace@libs/${config.libraryContentJenkinsName}/"+LIBRARY_MANIFEST, 'ContentBase', false)
                if (versionContentBase) {
                    echo "Using ContentBase Version '${versionContentBase}' (Version read from Manifest '${LIBRARY_MANIFEST}')."

                    // Load IRIS ContentBase Library
                    if (loadLibrary(config.libraryContentBaseJenkinsName, versionContentBase)) {

                        // Read Version information from ContentBase to be used to load Framework Library
                        String versionFramework = readVersionFromFile("../workspace@libs/${config.libraryContentBaseJenkinsName}/"+LIBRARY_MANIFEST, 'Framework', false)
                        if (versionFramework) {
                            echo "Using Framework Version '${versionFramework}' (Version read from Manifest '${LIBRARY_MANIFEST}')."

                            // Load IRIS Framework Library
                            if (loadLibrary(config.libraryFrameworkJenkinsName, versionFramework)) {

                                // All Libraries are available now, lets go...
                                if (config.useCaseValue.verbose) {
                                    echo "Calling useCase '${config.useCase}' with data '${config.useCaseValue}'."
                                }

                                // Jump into IRIS, still outside docker (catching is not possible, for unkown reason)
                                String resultJSON = callIRISScript(config)

                                try {
                                    // Extract result
                                    Map result = readJSON text: resultJSON

                                    // Print text of result
                                    if (result.resultText) {
                                        echo "ResultText = ${result.resultText}"
                                    }

                                    // Set status of Build based on result
                                    switch (result.resultCode.toInteger()) {
                                        case this.Result.Ok:           break // Don't touch

                                        case this.Result.Error:
                                        case this.Result.Abort:        script.currentBuild.result = 'FAILURE'
                                                                       break

                                        default: /* Warning, pError */ script.currentBuild.result = 'UNSTABLE'
                                    }

                                } catch (/*JsonException*/ ex) {
                                    echo "Trouble while reading resultJSON '${resultJSON}': ${ex.toString()}"
                                    script.currentBuild.result = 'FAILURE'
                                }
                            } else {
                                script.currentBuild.result = 'FAILURE'
                            }
                        } else {
                            script.currentBuild.result = 'FAILURE'
                        }
                    } else {
                        script.currentBuild.result = 'FAILURE'
                    }
                } else {
                    script.currentBuild.result = 'FAILURE'
                }
            } else {
                script.currentBuild.result = 'FAILURE'
            }
        } else {
            script.currentBuild.result = 'FAILURE'
        }
    }
}

// Read Version from Manifest File (JSON)
String readVersionFromFile(String filePath, String keyName, Boolean useCaseLayer) {
    String version = ''

    if (fileExists(filePath)) {
        try {
            Map manifest = readJSON file: filePath

            version = readVersionFromMap(manifest, keyName, useCaseLayer)

        } catch (/*JsonException | FileNotFoundException*/ ex) {
            echo "Trouble while reading manifest from '${filePath}': ${ex.toString()}"
        }
    } else {
        echo "Manifest '${filePath}' does not exist!"
    }

    return version
}

// Read Version from Manifest Text (JSON)
String readVersionFromText(String manifestText, String keyName, Boolean useCaseLayer) {
    String version = ''

    try {
        Map manifest = readJSON text: manifestText

        version = readVersionFromMap(manifest, keyName, useCaseLayer)

    } catch (/*JsonException*/ ex) {
        echo "Trouble while interpreting manifest: ${ex.toString()}"
    }

    return version
}

// Read Version from Manifest Map
String readVersionFromMap(Map manifest, String keyName, Boolean useCaseLayer) {
    String version = 'master' // Manifest available and readable, but maybe no settings inside, so defaults to 'master'

    // Is it a manifest of IRIS-caller like Pipeline script?
    if (useCaseLayer) {
        if (manifest.UseCase) {
            manifest = manifest.UseCase // Jump one layer up

            echo 'Manifest: Jump to useCase Level.' // Needed for automated tests ONLY
        }
    }

    // Read version
    if (manifest."${keyName}") {
        if (manifest."${keyName}".Version) {
            version = manifest."${keyName}".Version
        }
    }

    return version
}

// Load Library
Boolean loadLibrary(String libraryName, String libraryVersion = 'master') {
    def loadedLibrary = null // defaults to 'no-library'

    try {
        // Try to load Library with specified Version
        loadedLibrary = library libraryName+'@'+libraryVersion

    } catch (AbortException ex) {
        if (ex.toString().contains('No library named')) {
            echo "Library '${libraryName}' is not defined in Jenkins-Settings. See documentation, how to enable: https://go.sap.corp/irisdoc/lib/setupLibrary/#setup"
        } else if (ex.toString().contains('No version')) {
            echo "Library '${libraryName}' with Version '${libraryVersion}' does not exist. Check for correct Version in repository https://github.wdf.sap.corp/ProjectIRIS/iris/releases"
        } else {
            echo "Trouble while loading Library '${libraryName}' with Version '${libraryVersion}': ${ex.toString()}."
        }

        // RELOAD with different version tag DOES NOT WORK... reported as bug: https://issues.jenkins-ci.org/browse/JENKINS-54742 -> commented out as would fail

//        try {
//            // Try to load Library with Version plus suffix '-inc' because it was an incompatible update...
//            loadedLibrary = library libraryName+'@'+libraryVersion+'-inc'
//
//        } catch (AbortException ex2) {
//            try {
//                // Try to load Library with default Version, as other Versions have not been found
//                loadedLibrary = library libraryName
//
//                echo "Library '${libraryName}' with Version ${libraryVersion} does not exist. Latest version used instead. In future releases this will cause IRIS to fail."
//                //throw new IllegalArgumentException("${STEP_NAME}: Library '${libraryName}' with Version ${libraryVersion} does not exist! Documentation on how to set the libraryVersion can be found: https://go.sap.corp/irisdoc/lib/setupLibrary/")
//
//            } catch (AbortException ex3) {
//                echo "Library '${libraryName}' does not exist."
//                //throw new IllegalArgumentException("${STEP_NAME}: Library '${libraryName}' does not exist! Documentation can be found: https://go.sap.corp/irisdoc/lib/setupLibrary/")
//            }
//        }
    }

    return loadedLibrary != null
}

import com.sap.icd.jenkins.Utils
import com.sap.piper.GenerateDocumentation
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Deprecate
import com.sap.piper.internal.Notify
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'executeOpenSourceDependencyScan'
@Field Set GENERAL_CONFIG_KEYS = [
    /** DEPRECATED */
    'artifactType',
    /**
     * Defines the tool which is used for building the artifact.
     * @possibleValues `dub`, `docker`, `golang`, `maven`, `mta`, `npm`, `pip`, `sbt`
     */
    'buildTool',
    /**
     * For MTAs only: Defines a list of module build descriptors (including relative path to project root) to exclude from the scan and assessment activities.
     */
    'buildDescriptorExcludeList',
    /** DEPRECATED */
    'scanType'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Explicitly activates/deactivates `protecodeExecuteScan`
     * @possibleValues `true`, `false`
     */
    'protecodeExecuteScan',
    /**
     * Explicitly activates/deactivates `executeVulasScan`
     * @possibleValues `true`, `false`
     */
    'executeVulasScan',
    /**
     * Explicitly activates/deactivates `executeWhitesourceScan`
     * @possibleValues `true`, `false`
     */
    'executeWhitesourceScan',
    /**
     * Explicitly activates/deactivates `detectExecuteScan`
     * @possibleValues `true`, `false`
     */
    'detectExecuteScan',
    /** List of _buildTool_ values for which the step `protecodeExecuteScan` will be active */
    'protecodeActive',
    /** List of _buildTool_ values for which the step `executeVulasScan` will be active */
    'vulasActive',
    /** List of _buildTool_ values for which the step `whitesourceExecuteScan` will be active */
    'whitesourceActive',
    /** List of _buildTool_ values for which the step `detectExecuteScan` will be active */
    'detectActive',
    /** List of _buildTool_ values for which the step `executeWhitesourceScan` will be active */
    'whitesourceActiveOld',
    /**
     * Explicitly activates/deactivates `whitesourceExecuteScan`
     * @possibleValues `true`, `false`
     */
    'whitesourceExecuteScan',
    /** DEPRECATED */
    'whitesourceJava'
])

@Field Set PARAMETER_KEYS = null

/**
 * This step conducts Open Source dependency analysis. With this step you will be able to identify if there are known security vulnerabilities (e.g. published CVEs) within any of your direct or indirect dependencies.
 *
 * In addition to plain vulnerability scanning, please also subscribe to your product in mandatory [Software Vulnerability Monitor (SVM)](https://mo-d9c00d87e.mo.sap.corp:4300/vulmon/ui/monitor/WebContent/). <br />
 * Ideally your Security Expert claims ownership for your product in this tool and developers start viewing it.
 *
 * !!! note
 *     Open source dependency scanning is one important aspect in the [Rugged DevOps](https://youtu.be/dogofef4HWg?list=PLEx5khR4g7PIBIQHkNnOyRy2kclkq25Bh) practice and supports a clean software supply chain management.
 *
 * !!! note
 *     "Using Components with Known Vulnerabilities" is now a part of the [OWASP Top 10](https://www.owasp.org/index.php/Top_10_2013-A9-Using_Components_with_Known_Vulnerabilities) and insecure libraries can pose a huge risk for your webapp.
 *
 * Depending on the technology of your application different tools are used (see [full overview](https://jam4.sapjam.com/groups/XgeUs0CXItfeWyuI4k7lM3/overview_page/VpPnhWUTipYV4eg57rSmXk) from Open Source Security team):
 *
 * | Technology | Tool | Remarks |
 * |------------|------|---------|
 * | Java | [Vulas](https://go.sap.corp/vulas)| SAP internal tool coming from SAP Security Research |
 * | Python | [Vulas](https://go.sap.corp/vulas)| SAP internal tool coming from SAP Security Research |
 * | Node.js, Scala | [Whitesource](https://go.sap.corp/whitesource) | external tool |
 * | golang | [Whitesource](https://www.whitesourcesoftware.com/) \| [Protecode](https://protecode.mo.sap.corp/) | external tool |
 * | Docker | [Protecode](https://go.sap.corp/protecode) | external tool |
 * | D | [Whitesource](https://www.whitesourcesoftware.com/) \| [Protecode](https://protecode.mo.sap.corp/) | external tool |
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this

    // notify about deprecated step usage
    Notify.deprecatedStep(this, null, "removed", script?.commonPipelineEnvironment)

    def scanJobs = [:]

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, failOnError: true,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def utils = parameters.juStabUtils ?: new Utils()

        // handle deprecated parameters
        Deprecate.parameter(this, parameters, 'exclude', 'buildDescriptorExcludeList')
        Deprecate.parameter(this, parameters, 'mtaExcludeModules', 'buildDescriptorExcludeList')
        Deprecate.parameter(this, parameters, 'vulas', 'executeVulasScan')
        Deprecate.parameter(this, parameters, 'whitesource', 'executeWhitesourceScan')

        script.globalPipelineEnvironment.setInfluxStepData('opensourcedependency', false)

        def stageName = parameters.stageName?:env.STAGE_NAME

        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS) //consider all parameters - detailed steps will conduct filtering of parameters
            .use()

        if (config.whitesourceJava) {
            Notify.warning(this, "Parameter 'whitesourceJava' is deprecated, please use parameter 'whitesourceActiveOld'.", STEP_NAME)
            config.whitesourceActiveOld.add('maven')
        }

        config = new ConfigurationHelper(config)
            .addIfEmpty('scanType', config.buildTool)
            .addIfEmpty('protecodeExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.protecodeExecuteScan)
            .addIfEmpty('protecodeExecuteScan', (config.artifactType in config.protecodeActive || config.buildTool in config.protecodeActive) ?: null)
            .addIfEmpty('executeVulasScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.executeVulasScan)
            .addIfEmpty('executeVulasScan',  (config.artifactType in config.vulasActive || config.buildTool in config.vulasActive) ?: null)
            .addIfEmpty('detectExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.detectExecuteScan)
            .addIfEmpty('detectExecuteScan',  (config.artifactType in config.detectActive || config.buildTool in config.detectActive) ?: null)
            .addIfEmpty('executeWhitesourceScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.executeWhitesourceScan)
            .addIfEmpty('executeWhitesourceScan',  (config.artifactType in config.whitesourceActiveOld || config.buildTool in config.whitesourceActiveOld) ?: null)
            .addIfEmpty('whitesourceExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.whitesourceExecuteScan)
            .use()

        def excludedFiles = []
        def includeSignaturePathBlackDuck = []
        def defaultDetectExecuteScanProperties = ['--blackduck.signature.scanner.memory=4096','--blackduck.timeout=6000','--blackduck.trust.cert=true',
                                                '--detect.report.timeout=4800','--logging.level.com.synopsys.integration=DEBUG',
                                                '--detect.maven.excluded.scopes=test']
        def mtaDetectExecuteScanProperties = []

        if (config.scanType == 'mta') {
            utils.unstash('buildDescriptor')

            config = new ConfigurationHelper(config)
                .mixin([
                    buildDescriptorExcludeList: config.buildDescriptorExcludeList instanceof List ? config.buildDescriptorExcludeList : config.buildDescriptorExcludeList?.tokenize(','),
                    scanPaths: config.scanPaths instanceof List ? config.scanPaths : config.scanPaths?.tokenize(',')
                ])
                .use()

            config.buildDescriptorExcludeList?.each {
                item ->
                    excludedFiles.add(new File(item).path)
            }

            config.scanPaths?.each {
                item ->
                    includeSignaturePathBlackDuck.add(item)
            }

            // remove pom.xml file from executeWhitesourceScan in case it is not explicitly set to treat Java projects
            if (config.executeWhitesourceScan && !config.whitesourceJava) {
                def mavenDescriptorFiles = findFiles(glob: "**${File.separator}pom.xml")
                echo "[${STEP_NAME}] WhiteSource for Java not activated. Ignoring all maven descriptor files during WhiteSource scan."
                for (def i = 0; i < mavenDescriptorFiles.length; i++) {
                    def file = mavenDescriptorFiles[i].path
                    if (!excludedFiles.contains(file)) {
                        echo "[${STEP_NAME}] Adding ${file} to exclude list"
                        excludedFiles.add(file)
                    }
                }
            }

            // in case of MTA only include java projects for detectExecute.
            if (config.detectExecuteScan){
                if(!config.whitesourceJava){
                    //detector scanning is configured for maven and gradle build and directory depth is 5
                    echo "[${STEP_NAME}] BlackDuck activated. Including only maven and gradle builds during Blackduck detector scan"
                    mtaDetectExecuteScanProperties.add('--detect.detector.search.depth=5')
                    mtaDetectExecuteScanProperties.add('--detect.included.detector.types=MAVEN,GRADLE')
                    def mavenDescriptorFilesForBlackDuck = findFiles(glob: "**${File.separator}pom.xml")
                    if (mavenDescriptorFilesForBlackDuck.size() == 0){
                          mtaDetectExecuteScanProperties.add('--detect.tools.excluded=SIGNATURE_SCAN')
                    }
                    //signature scanning is configured for directory where POM xml files are present
                    for (def i = 0; i < mavenDescriptorFilesForBlackDuck.length; i++) {
                        def fileForBlackDuck = mavenDescriptorFilesForBlackDuck[i].path
                        def pathForBlackDuck = fileForBlackDuck.substring(0,fileForBlackDuck.lastIndexOf(File.separator)+1)
                        if (!includeSignaturePathBlackDuck.contains(pathForBlackDuck)) {
                            echo "[${STEP_NAME}] BlackDuck activated. Adding maven descriptor directory ${pathForBlackDuck} to be included during Blackduck signature scan"
                            includeSignaturePathBlackDuck.add(pathForBlackDuck)
                        }
                    }
                } else {
                    //signature scanning and detector scanning is not configured when whitesource for java is true
                    mtaDetectExecuteScanProperties.add('--detect.included.detector.types=NONE')
                    mtaDetectExecuteScanProperties.add('--detect.tools.excluded=SIGNATURE_SCAN')
                }
            }
        }

        // TODO to be removed once new whitesourceExecuteScan is fully functional
        if (config.executeWhitesourceScan && !config.whitesourceExecuteScan) {
            def wsScanType = (config.scanType == 'dub') ? 'fileAgent' : (config.scanType ?: 'npm')
            def whitesourceConfig = [:].plus(config)
            whitesourceConfig.buildDescriptorExcludeList = [].plus(config.buildDescriptorExcludeList).plus(excludedFiles)
            whitesourceConfig.securityVulnerabilities = true
            whitesourceConfig.scanType = wsScanType
            scanJobs["OpenSourceDependency [whitesource]"] = {
                executeWhitesourceScan(whitesourceConfig)
            }
        } else if (config.whitesourceExecuteScan) {
            scanJobs["OpenSourceDependency [whitesource]"] = {
                whitesourceExecuteScan script: script, securityVulnerabilities: true
            }
        }

        if (config.protecodeExecuteScan) {
            def protecodeConfig = [:].plus(config)
            scanJobs["OpenSourceDependency [Protecode]"] = {
                protecodeExecuteScan(protecodeConfig)
            }
        }

        if (config.detectExecuteScan && !config.executeVulasScan) {
            def detectExecuteConfig = [:].plus(config)
            if (config.scanType == 'mta') {
                mergeDetectScanProperties(detectExecuteConfig, mtaDetectExecuteScanProperties, defaultDetectExecuteScanProperties, includeSignaturePathBlackDuck, script, stageName)
            }
            scanJobs["OpenSourceDependency [BlackDuck]"] = {
                detectExecuteScan(detectExecuteConfig)
            }
        }

        if (config.executeVulasScan) {
            def vulasConfig = [:].plus(config)
            scanJobs["OpenSourceDependency [VULAS]"] = {
                executeVulasScan(vulasConfig)
            }
        }
    }
    // execute OpenSourceDependency scans
    if (scanJobs.size()>0) {
        scanJobs.failFast = false
        parallel scanJobs
        script.globalPipelineEnvironment.setInfluxStepData('opensourcedependency', true)
    }
}

void mergeDetectScanProperties(detectExecuteConfig, mtaDetectExecuteScanProperties, defaultDetectExecuteScanProperties, includeSignaturePathBlackDuck, script, stageName){
    // if there are scanProperties passed to step detectExecuteScan then we need to include those scanProperties along with mtaDetectExecuteScanProperties needed for mta
    // if any scanProperty passed to step detectExecuteScan is the same as in mtaDetectExecuteScanProperties ,
    // then mtaDetectExecuteScanProperties takes precedence
    Map currentDetectExecuteScanConfiguration = script.commonPipelineEnvironment.getStepConfiguration("detectExecuteScan", stageName)
    if(currentDetectExecuteScanConfiguration.containsKey('scanProperties')){
        currentDetectExecuteScanConfiguration.scanProperties.each{property ->
            if (!mtaDetectExecuteScanProperties.any{ it.contains(property.substring(0,property.indexOf('='))) }) {
                mtaDetectExecuteScanProperties.add(property)
            }
        }
        detectExecuteConfig.scanProperties = mtaDetectExecuteScanProperties
    } else {
        detectExecuteConfig.scanProperties = [].plus(defaultDetectExecuteScanProperties).plus(mtaDetectExecuteScanProperties)
    }
    if(includeSignaturePathBlackDuck.size() > 0 && !currentDetectExecuteScanConfiguration.containsKey('scanPaths')) {
        detectExecuteConfig.scanPaths = includeSignaturePathBlackDuck
    }
}

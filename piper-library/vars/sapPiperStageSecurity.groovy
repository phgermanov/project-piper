import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'sapPiperStageSecurity'

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Defines the build tool used.
     * @possibleValues `maven`, `npm`, `mta`, ...
     */
    'buildTool',
    /**
     * Defines the labels of the Jenkins execution nodes which are started for the individual scans in a parallel mode.
     */
    'nodeLabel',
    /*
     secretID, groupID to call systemTrust
    */
    'vaultAppRoleSecretTokenCredentialsId',
    'vaultBasePath',
    'vaultPipelineName'
]
@Field STAGE_STEP_KEYS = [
    /**
     * Performs BlackDuck Detect scanning to identify Open Source Security vulnerabilities.<br />
     * This tool is typically used for following `buildTool`s: `maven`
     */
    'detectExecuteScan',
    /**
     * Performs Checkmarx scanning.<br />
     * This tool is used for following `buildTool`s: `docker`, `golang`, `mta`, `npm`, `sbt`
     */
    'executeCheckmarxScan',
    /**
     * Performs Fortify scanning.<br />
     * This tool is used for following `buildTool`s: `docker`, `maven`, `mta`, `pip`
     */
    'executeFortifyScan',
    /**
     * Always active unless pipeline runs in optimized mode, then individual scan steps are triggered<br />
     * This tool triggers an Open Source Dependency scan which then selects the appropriate tool.
     */
    'executeOpenSourceDependencyScan',
    /**
     * Performs DASTer scanning.<br />
     */
    'executeDasterScan',
    /**
     * Performs DASTer scanning.<br />
     */
    'executeZAPScan',
    /**
     * Performs Checkmarx scanning.<br />
     * This tool is typically used for following `buildTool`s: `docker`, `golang`, `mta`, `npm`, `sbt`<br />
     * Currently, explicit activation is required of this updated step by providing configuration under `steps` - `checkmarxExecuteScan` in .pipeline/config.yml
     */
    'checkmarxExecuteScan',
    /**
     * Performs CheckmarxOne scanning.<br />
     * This tool is typically used for following `buildTool`s: `docker`, `golang`, `mta`, `npm`, `sbt`<br />
     * Currently, explicit activation is required of this updated step by providing configuration under `steps` - `checkmarxOneExecuteScan` in .pipeline/config.yml
     */
    'checkmarxOneExecuteScan',
    /**
     * Performs Fortify scanning.<br />
     * This tool is typically used for following `buildTool`s: `docker`, `maven`, `mta`, `pip`
     *
     * Currently, explicit activation is required of this updated step by providing configuration under `steps` - `fortifyExecuteScan` in .pipeline/config.yml.
     */
    'fortifyExecuteScan',
    /**
     * Performs CodeQL scanning.<br />
     * This step executes a CodeQL scan on the specified project to perform static code analysis and check the source code for security flaws.
     *
     * Currently, explicit activation is required of this updated step by providing configuration under `steps` - `codeqlExecuteScan` in .pipeline/config.yml.
     */
    'codeqlExecuteScan',
    /**
     * Performs Malware scanning.<br />
     * This tool is typically used for following `buildTool`s: `docker`
     *
     * Currently, explicit ctivation is required of this updated step by providing configuration under `steps` - `malwareExecuteScan` in .pipeline/config.yml.
     */
    'malwareExecuteScan',
    /**
     * Performs upload of result files to cumulus.<br />
     */
    'sapCumulusUpload',
    /**
     * Create Fosstars OSS rating report.<br />
     */
    'sapCreateFosstarsReport',
    /**
     * Performs WhiteSource scanning to identify Open Source Security vulnerabilities.<br />
     * This tool is typically used for following `buildTool`s: `golang`, `mta`, `pip`, , `npm`
     */
    'whitesourceExecuteScan'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS.plus(['newOSS']))
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    def stageName = parameters.stageName?:env.STAGE_NAME

    def unstableMessage = "optimized pipeline run - continue although error occurred"

    def securityScanMap = [:]

    List ipStepsExecuted = []

    Map config = ConfigurationHelper
        .loadStepDefaults(this)
        .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.globalPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .withMandatoryProperty('buildTool')
        .use()

    config = new ConfigurationHelper(config)
        // option 'docker' available for both scans to properly support multistage Docker builds
        .addIfEmpty('executeFortifyScan', config.buildTool in ['docker', 'maven', 'mta', 'pip'] && script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.executeFortifyScan)
        .addIfEmpty('executeCheckmarxScan', config.buildTool in ['docker', 'golang', 'mta', 'npm', 'sbt'] && script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.executeCheckmarxScan)
        .addIfEmpty('executeOpenSourceDependencyScan', true)
        .addIfEmpty('executeDasterScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.executeDasterScan)
        .addIfEmpty('executeZAPScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.executeZAPScan)
        .addIfEmpty('checkmarxExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.checkmarxExecuteScan)
        .addIfEmpty('checkmarxOneExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.checkmarxOneExecuteScan)
        .addIfEmpty('fortifyExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.fortifyExecuteScan)
        .addIfEmpty('codeqlExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.codeqlExecuteScan)
        .addIfEmpty('malwareExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.malwareExecuteScan)
        .addIfEmpty('detectExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.detectExecuteScan)
        .addIfEmpty('whitesourceExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.whitesourceExecuteScan)
        .addIfEmpty('protecodeExecuteScan', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.protecodeExecuteScan)
        .addIfEmpty('sapCumulusUpload', script.commonPipelineEnvironment.getStepConfiguration('sapCumulusUpload', stageName)?.pipelineId ? true : false )
        .addIfEmpty('sapCreateFosstarsReport', script.globalPipelineEnvironment.configuration.runStep?.get(stageName)?.sapCreateFosstarsReport)
        .use()


            // new behavior (usage of new steps and no more executeOpenSourceDependencyScan) for optimized pipelines as a first step
            // also allow pilot pipelines without executeOpenSourceDependencyScan
            if (script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled') || config.newOSS) {

                if (config.checkmarxExecuteScan) {
                    securityScanMap['SAST [Checkmarx]'] = {
                        node(config.nodeLabel) {
                            utils.unstableIfCondition(this, script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled'), unstableMessage) {
                                try {
                                    durationMeasure(script: script, measurementName: 'checkmarx_duration') {
                                        checkmarxExecuteScan script: script
                                    }
                                } finally {
                                    if (config.sapCumulusUpload) {
                                        sapCumulusUpload script: script, filePattern: '**/CxSASTReport_*.pdf, **/*CxSAST*.html, **/ScanReport.*, **/toolrun_checkmarx_*.json, **/piper_checkmarx_report.json, **/*.sarif', stepResultType: 'checkmarx'
                                        sapCumulusUpload script: script, filePattern: '**/piper_checkmarx_report.json', stepResultType: 'policy-evidence/SDOL-009-SAST'
                                    }
                                    deleteDir()
                                }
                            }
                        }
                    }
                }

                if (config.checkmarxOneExecuteScan) {
                    securityScanMap['SAST [CheckmarxOne]'] = {
                        node(config.nodeLabel) {
                            utils.unstableIfCondition(this, script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled'), unstableMessage) {
                                try {
                                    durationMeasure(script: script, measurementName: 'checkmarxone_duration') {
                                        checkmarxOneExecuteScan script: script
                                    }
                                } finally {
                                    if (config.sapCumulusUpload) {
                                        sapCumulusUpload script: script, filePattern: '**/Cx1_SASTReport_*.pdf, **/Cx1_SASTReport*.json, **/toolrun_checkmarxOne_*.json, **/checkmarxOne/*.sarif', stepResultType: 'checkmarxOne'
                                        sapCumulusUpload script: script, filePattern: '**/piper_checkmarxone_report.json', stepResultType: 'policy-evidence/SDOL-009-SAST'
                                    }
                                    deleteDir()
                                }
                            }
                        }
                    }
                }

                if (config.fortifyExecuteScan) {
                    securityScanMap['SAST [Fortify]'] = {
                        node(config.nodeLabel) {
                            utils.unstableIfCondition(this, script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled'), unstableMessage) {
                                try {
                                    durationMeasure(script: script, measurementName: 'fortify_duration') {
                                        fortifyExecuteScan script: script
                                    }
                                } finally {
                                    if (config.sapCumulusUpload) {
                                        sapCumulusUpload script: script, filePattern: '**/*.PDF, **/*.fpr, **/fortify-scan.*, **/toolrun_fortify_*.json, **/piper_fortify_report.json, **/*.sarif, **/*.sarif.gz', stepResultType: 'fortify'
                                        sapCumulusUpload script: script, filePattern: '**/piper_fortify_report.json', stepResultType: 'policy-evidence/SDOL-009-SAST'
                                    }
                                    deleteDir()
                                }
                            }
                        }
                    }
                }

                if (config.codeqlExecuteScan) {
                    securityScanMap['SAST [CodeQL]'] = {
                        node(config.nodeLabel) {
                            utils.unstableIfCondition(this, script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled'), unstableMessage) {
                                try {
                                    durationMeasure(script: script, measurementName: 'codeql_duration') {
                                        codeqlExecuteScan script: script
                                    }
                                } finally {
                                    if (config.sapCumulusUpload) {
                                        sapCumulusUpload script: script, filePattern: '**/toolrun_codeql_*.json, **/codeqlReport.sarif, **/codeqlReport.csv', stepResultType: 'codeql'
                                        sapCumulusUpload script: script, filePattern: '**/piper_codeql_report.json', stepResultType: 'policy-evidence/SDOL-009-SAST'
                                    }
                                    deleteDir()
                                }
                            }
                        }
                    }
                }

                if (config.malwareExecuteScan) {
                    securityScanMap['MalwareScan [SAP Cloud]'] = {
                        node(config.nodeLabel) {
                            utils.unstableIfCondition(this, script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled'), unstableMessage) {
                                try {
                                    durationMeasure(script: script, measurementName: 'malwarescan_duration') {
                                        malwareExecuteScan script: script
                                    }
                                } finally {
                                    if (config.sapCumulusUpload) {
                                        sapCumulusUpload script: script, filePattern: '**/toolrun_malwarescan_*.json, **/malwarescan_report.json', stepResultType: 'malwarescan'
                                    }
                                    deleteDir()
                                }
                            }
                        }
                    }
                }

                if (config.detectExecuteScan) {
                    securityScanMap["OpenSourceSecurity [BlackDuck]"] = {
                        node(config.nodeLabel) {
                            utils.unstableIfCondition(this, script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled'), unstableMessage) {
                                try {
                                    durationMeasure(script: script, measurementName: 'detect_duration') {
                                        detectExecuteScan script: script
                                        ipStepsExecuted.add('detectExecuteScan')
                                    }
                                } finally {
                                    if (config.sapCumulusUpload) {
                                        sapCumulusUpload script: script, filePattern: '**/*BlackDuck_RiskReport.pdf, **/blackduck-ip.json, **/toolrun_detectExecute_*.json, **/piper_detect_policy_violation_report.html', stepResultType: 'blackduck-ip'
                                        sapCumulusUpload script: script, filePattern: '**/*BlackDuck_RiskReport.pdf, **/detectExecuteScan_policy_*.json, **/piper_detect_vulnerability_report.html, **/toolrun_detectExecute_*.json, **/piper_detect_vulnerability.sarif, **/piper_hub_detect_sbom.xml', stepResultType: 'blackduck-security'
                                        sapCumulusUpload script: script, filePattern: '**/blackduck-ip.json', stepResultType: 'policy-evidence/PSL-1'
                                    }
                                    deleteDir()
                                }
                            }
                        }
                    }
                }

                if (config.whitesourceExecuteScan) {
                    securityScanMap["OpenSourceSecurity [WhiteSource]"] = {
                        node(config.nodeLabel) {
                            utils.unstableIfCondition(this, script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled'), unstableMessage) {
                                try {
                                    durationMeasure(script: script, measurementName: 'whitesource_duration') {
                                        whitesourceExecuteScan script: script, securityVulnerabilities: true
                                        ipStepsExecuted.add('whitesourceExecuteScan')
                                    }
                                }finally{
                                    if(config.sapCumulusUpload) {
                                        sapCumulusUpload script: script, filePattern: 'whitesource/*risk-report.pdf, **/toolrun_whitesource_*.json, **/piper_whitesource_vulnerability_report.html, **/piper_whitesource_vulnerability.sarif, **/piper_whitesource_sbom.xml', stepResultType: 'whitesource-security'
                                        sapCumulusUpload script: script, filePattern: '**/whitesource-ip.json, whitesource/*risk-report.pdf, **/toolrun_whitesource_*.json', stepResultType: 'whitesource-ip'
                                        sapCumulusUpload script: script, filePattern: '**/whitesource-ip.json', stepResultType: 'policy-evidence/PSL-1'
                                    }
                                    deleteDir()
                                }
                            }
                        }
                    }
                }

                if (config.protecodeExecuteScan) {
                    securityScanMap["OpenSourceSecurity [Protecode]"] = {
                        node(config.nodeLabel) {
                            utils.unstableIfCondition(this, script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled'), unstableMessage) {
                                try {
                                    durationMeasure(script: script, measurementName: 'protecode_duration') {
                                        protecodeExecuteScan script: script
                                    }
                                } finally {
                                    if (config.sapCumulusUpload) {
                                        sapCumulusUpload script: script, filePattern: script.commonPipelineEnvironment.getStepConfiguration("protecodeExecuteScan", stageName).reportFileName+', **/toolrun_protecode_*.json', stepResultType: 'protecode'
                                    }
                                    deleteDir()
                                }
                            }
                        }
                    }
                }

                if (config.executeDasterScan) {
                    securityScanMap['DAST [DASTer]'] = {
                        node(config.nodeLabel) {
                            utils.unstableIfCondition(this, script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled'), unstableMessage) {
                                try{
                                    durationMeasure(script: script, measurementName: 'daster_duration') {
                                        executeDasterScan script: script
                                    }
                                }finally{
                                    deleteDir()
                                }
                            }
                        }
                    }
                }

                if (config.executeZAPScan) {
                    securityScanMap['DAST [ZAP]'] = {
                        node(config.nodeLabel) {
                            utils.unstableIfCondition(this, script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled'), unstableMessage) {
                                try{
                                    durationMeasure(script: script, measurementName: 'zap_duration') {
                                        executeZAPScan script: script
                                    }
                                }finally{
                                    if(config.sapCumulusUpload) {
                                        sapCumulusUpload script: script, filePattern: 'zap_report.html', stepResultType: 'zap'
                                    }
                                    deleteDir()
                                }
                            }
                        }
                    }
                }

                if (config.sapCreateFosstarsReport) {
                    securityScanMap['FOSSTARS'] = {
                        node(config.nodeLabel) {
                            utils.unstableIfCondition(this, script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled'), unstableMessage) {
                                try{
                                    durationMeasure(script: script, measurementName: 'fosstars_duration') {
                                        sapCreateFosstarsReport script: script
                                    }
                                }finally{
                                    if(config.sapCumulusUpload) {
                                        sapCumulusUpload script: script, filePattern: 'fosstars-report/*.json', stepResultType: 'fosstars'
                                    }
                                    deleteDir()
                                }
                            }
                        }
                    }
                }

            // DEPRECATED behavior still in place
            // planned removal March 31st, 2021
            // currently default - will change with broad rollout of pipeline optimizations
            } else {

                if (config.executeFortifyScan || config.fortifyExecuteScan) {
                    securityScanMap['Fortify'] = {
                        node(config.nodeLabel) {
                            try{
                                durationMeasure(script: script, measurementName: 'fortify_duration') {
                                    if (!config.fortifyExecuteScan) {
                                        executeFortifyScan script: script
                                    } else {
                                        boolean verifyOnly = false
                                        if (script.globalPipelineEnvironment.getStepConfiguration('', stageName).githubTokenCredentialsId) {
                                            try {
                                                githubCheckBranchProtection script: script, branch: script.globalPipelineEnvironment.getStepConfiguration('', '').productiveBranch, requiredChecks: ['piper/fortify'], requireEnforceAdmins: true
                                                verifyOnly = true
                                            } catch (err) {
                                                verifyOnly = false
                                            }
                                        }
                                        fortifyExecuteScan script: script, verifyOnly: verifyOnly
                                    }
                                }
                            }finally{
                                if(config.sapCumulusUpload) {
                                    sapCumulusUpload script: script, filePattern: '**/*.PDF, **/*.fpr, **/fortify-scan.*', stepResultType: 'fortify'
                                }
                                deleteDir()
                            }
                        }
                    }
                }

                if (config.executeCheckmarxScan || config.checkmarxExecuteScan) {
                    securityScanMap['Checkmarx'] = {
                        node(config.nodeLabel) {
                            try{
                                durationMeasure(script: script, measurementName: 'checkmarx_duration') {
                                    if (!config.checkmarxExecuteScan) {
                                        executeCheckmarxScan script: script
                                    } else {
                                        boolean verifyOnly = false
                                        if (script.globalPipelineEnvironment.getStepConfiguration('', stageName).githubTokenCredentialsId) {
                                            try {
                                                githubCheckBranchProtection script: script, branch: script.globalPipelineEnvironment.getStepConfiguration('', '').productiveBranch, requiredChecks: ['piper/checkmarx'], requireEnforceAdmins: true
                                                verifyOnly = true
                                            } catch (err) {
                                                verifyOnly = false
                                            }
                                        }
                                        checkmarxExecuteScan script: script, verifyOnly: verifyOnly
                                    }
                                }
                            }finally{
                                if(config.sapCumulusUpload) {
                                    sapCumulusUpload script: script, filePattern: '**/CxSASTReport_*.pdf, **/*CxSAST*.html, **/CxSASTResults_*.xml, **/ScanReport.*', stepResultType: 'checkmarx'
                                }
                                deleteDir()
                            }
                        }
                    }
                }

                if (config.executeOpenSourceDependencyScan) {
                    securityScanMap['OpenSourceVulnerability'] = {
                        node(config.nodeLabel) {
                            try{
                                durationMeasure(script: script, measurementName: 'opensourcevulnerability_duration') {
                                    executeOpenSourceDependencyScan script: script
                                }
                            }finally{
                                if(config.sapCumulusUpload) {
                                    sapCumulusUpload script: script, filePattern: 'whitesource/*risk-report.pdf, **/toolrun_whitesource_*.json, **/piper_whitesource_vulnerability_report.html, **/piper_whitesource_vulnerability.sarif, **/piper_whitesource_sbom.xml', stepResultType: 'whitesource-security'
                                    sapCumulusUpload script: script, filePattern: script.commonPipelineEnvironment.getStepConfiguration("protecodeExecuteScan", stageName).reportFileName+", **/toolrun_protecode_*.json", stepResultType: 'protecode'
                                }
                                deleteDir()
                            }
                        }
                    }
                }

                if (config.executeDasterScan) {
                    securityScanMap['DASTer'] = {
                        node(config.nodeLabel) {
                            try{
                                durationMeasure(script: script, measurementName: 'daster_duration') {
                                    executeDasterScan script: script
                                }
                            }finally{
                                deleteDir()
                            }
                        }
                    }
                }

                if (config.executeZAPScan) {
                    securityScanMap['ZAP'] = {
                        node(config.nodeLabel) {
                            try{
                                durationMeasure(script: script, measurementName: 'zap_duration') {
                                    executeZAPScan script: script
                                }
                            }finally{
                                if(config.sapCumulusUpload) {
                                    sapCumulusUpload script: script, filePattern: 'zap_report.html', stepResultType: 'zap'
                                }
                                deleteDir()
                            }
                        }
                    }
                }

                if (config.sapCreateFosstarsReport) {
                    securityScanMap['FOSSTARS'] = {
                        node(config.nodeLabel) {
                            try{
                                durationMeasure(script: script, measurementName: 'fosstars_duration') {
                                    sapCreateFosstarsReport script: script
                                }
                            }finally{
                                if(config.sapCumulusUpload) {
                                    sapCumulusUpload script: script, filePattern: 'fosstars-report/*.json', stepResultType: 'fosstars'
                                }
                                deleteDir()
                            }
                        }
                    }
                }
            }

            if (securityScanMap.size() > 0) {
                piperStageWrapper (script: script, stageName: stageName) {
                    echo "Getting System trust token"
                    def token = null
                    try {
                        def apiURLFromDefaults = script.commonPipelineEnvironment.getValue("hooks")?.systemtrust?.serverURL ?: ''
                        token = sapGetSystemTrustToken(apiURLFromDefaults, config.vaultAppRoleSecretTokenCredentialsId, config.vaultPipelineName, config.vaultBasePath)  
                    } catch (Exception e) {
                        echo "Couldn't get system trust token, will proceed with configured credentials: ${e.message}"
                    }
                    wrap([$class: 'MaskPasswordsBuildWrapper', varPasswordPairs: [[password: token]]]) {
                        withEnv([
                            /*
                            Additional logic "?: ''" is necessary to ensure the environment
                            variable is set to an empty string if the value is null
                            Without this, the environment variable would be set to the string "null",
                            causing checks for an empty token in the Go application to fail.
                            */
                            "PIPER_systemTrustToken=${token ?: ''}",
                            ]) {
                                parallel securityScanMap.plus([failFast: false])

                                // save information about successful scans which can be re-used in IPScan & PPMS stage
                                script.commonPipelineEnvironment.setValue('ipStepsExecuted', ipStepsExecuted)
                            }
                }
            }
        }
    }

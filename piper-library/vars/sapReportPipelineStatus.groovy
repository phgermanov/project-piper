import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.PiperGoUtils
import static com.sap.piper.internal.Prerequisites.checkScript
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/sapReportPipelineStatus.yaml'

void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this

    Map stepParams = PiperGoUtils.prepare(this, script).plus(parameters)

    List credentials = [
        [type: 'usernamePassword', id: 'jenkinsCredentialsId', env: ['PIPER_jenkinsUser', 'PIPER_jenkinsToken']],
    ]
    try {
        piperExecuteBin(stepParams, STEP_NAME, METADATA_FILE, credentials)
    } catch (Exception ex) {
        //noop
    }
    def utils = parameters.juStabUtils ?: new Utils()
    parameters.juStabUtils = null
}

private String getErrorCategory(String stepName, String reason) {

    //known error categories so far (see resources/error_categories.yml):
    // - build, compliance, config, custom, test, service, infrastructure, undefined (fall back)

    if (stepName.contains('(extended)')) return 'custom'
    def errorCategories  = readYaml text: libraryResource('error_categories.yml')
    for(errCategory in errorCategories) {
        for(level in errCategory.value) {
            for (item in level.value) {
                if ((level.key == "step" && item == stepName) || reason.contains(item) ) {
                    return errCategory.key
                }
                if (item.contains("*")) {
                    if (wildcardRuleMatch(item, reason)) {
                        return errCategory.key
                    }
                }
            }
        }
    }
    return 'undefined'
}

private String filterDynamicMessages(String reason) {
    def staticMessageParts = [
        'ERROR - NO VALUE AVAILABLE',
        'Health check failed',
        'No such library resource',
        '[ERROR] Vulas detected Open Source Software Security vulnerabilities, the project is not compliant.',
        '[ERROR] Invalid yml configuration',
        '[executeFortifyAuditStatusCheck] There are artifacts that require manual approval',
        '[newmanExecute] No collection found with pattern',
        '[setupPipelineEnvironment] ERROR: Invalid yml configuration',
        'groovy.lang.MissingMethodException: No signature of method',
        'groovy.lang.MissingPropertyException: No such property',
        'java.io.FileNotFoundException',
        'java.io.IOException: java.nio.file.AccessDeniedException',
        'java.lang.IllegalStateException: KubernetesSlave',
        'java.lang.NoSuchMethodError: No such DSL method',
        'missing workspace',
        'org.jenkinsci.plugins.credentialsbinding.impl.CredentialNotFoundException',
        'org.jenkinsci.plugins.workflow.cps.CpsCompilationErrorsException: startup failed',
        'CFG-JJEN-000: No xMake job found! Check in xMakeJobName in piper config or Project Portal configuration',
        'SDE-JJEN-000: xMake Host not found! Visit the \'Cloud Availability Center\' about xMake otherwise open a \'ServiceNow\' ticket'
    ]
    def dynamicMessageParts = [
        '[sonarExecuteScan] Step execution failed (category: infrastructure). Error: running command *sonar-scanner failed: cmd.Run() failed: exit status 1',
        '[sonarExecuteScan] Step execution failed (category: service). Error: running command *sonar-scanner failed: cmd.Run() failed: exit status 1',
        '[sonarExecuteScan] Step execution failed*. Error: running command *sonar-scanner failed: cmd.Run() failed: exit status*',
        '[artifactPrepareVersion] Step execution failed (category: *). Error: failed to push changes*: ssh: handshake failed*',
        '[artifactPrepareVersion] Step execution failed (category: *). Error: failed to push changes*: reference already exists',
        '[artifactPrepareVersion] Step execution failed (category: *). Error: failed to push changes*: connection timed out',
        '[artifactPrepareVersion] Step execution failed (category: *). Error: failed to push changes*: authorization failed',
        '[artifactPrepareVersion] Step execution failed (category: *). Error: failed to push changes*: knownhosts: *illegal base64 data at input byte 8',
        '[artifactPrepareVersion] Step execution failed (category: *). Error: failed to retrieve version: failed to read file*: no such file or directory',
        '[artifactPrepareVersion] Step execution failed (category: config). Error: failed to retrieve version*',
        '[artifactPrepareVersion] Step execution failed (category: infrastructure). Error: failed to push changes*: connection timed out',
        '[githubSetCommitStatus] Step execution failed (category: *). Error: failed to set status \'pending\' on commitId \'*\': POST *: 422 No commit found for SHA: ',
        '[uiVeri5ExecuteTests] Step execution failed (category: *). Error: failed to execute run command: * failed: exit status 1',
        '[whitesourceExecuteScan] Step execution failed (category: compliance). Error: failed to execute WhiteSource scan*',
        'failed to execute WhiteSource scan: failed to run scan for npm module *: failed to execute WhiteSource scan with exit code 255',
        '[sapCheckPPMSCompliance] Step execution failed (category: config)', // combine all config errors
        '[sapXmakeExecuteBuild] Step execution failed (category: config). Error: No jobs found with name \'*\'',
        '[sapXmakeExecuteBuild] Step execution failed (category: *). Error: Failed to trigger job \'*\': Could not invoke job "": 401',
        '[sapXmakeExecuteBuild] Step execution failed (category: *). Error: Failed to trigger job \'*\': Could not invoke job "": 403',
        '[sapXmakeExecuteBuild] Step execution failed (category: *). Error: Failed to trigger job \'*\': Could not invoke job "": 500',
        '[sapXmakeExecuteBuild] Step execution failed (category: *). Error: Failed to trigger job \'*\': Unable to queue build',
        '[sapXmakeExecuteBuild] Step execution failed (category: service). Error: Failed to trigger job \'*\': build did not start in a reasonable amount of time',
        'No artifacts found that match the file pattern*. Configuration error?',
        'Remote build finished with status *: No associated error message found for this job*',
        'Cannot get property * on null object',
        'RejectedAccessException: No such field found',
        'IllegalArgumentException: Could not instantiate'
    ]
    staticMessageParts.each{ m ->
        if (reason.contains(m)) {
            reason = m
            return reason
        }
    }
    dynamicMessageParts.find{ rule ->
        if (wildcardRuleMatch(rule, reason)) {
            reason = rule.replace('*', '')
            return reason
        }
    }
    return reason
}

private boolean wildcardRuleMatch(String rule, String reason) {
    def allRulePartsMatch = true
    def rule_arr = rule.split("\\*")
    rule_arr.each{ e ->
        if (!reason.replace('\'', '').contains(e.replace('\'', '').trim())) {
            allRulePartsMatch = false
        }
    }
    return allRulePartsMatch
}

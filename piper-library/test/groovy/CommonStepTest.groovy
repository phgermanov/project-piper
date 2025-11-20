import org.junit.Before
import org.junit.ClassRule
import org.junit.Rule
import org.junit.rules.MethodRule
import org.junit.runner.RunWith
import org.junit.runners.Parameterized
import org.springframework.test.context.junit4.rules.SpringClassRule
import org.springframework.test.context.junit4.rules.SpringMethodRule
import util.Rules

import static java.util.stream.Collectors.toList
import static org.hamcrest.Matchers.empty
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.hamcrest.Matchers.startsWith
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue
import static org.junit.Assert.fail
import static org.junit.Assume.assumeThat

import java.lang.reflect.Field

import org.junit.Test

import groovy.io.FileType
import hudson.AbortException
import util.BasePiperTest

/*
 * Intended for collecting generic checks applied to all steps.
 */
@RunWith(Parameterized.class)
class CommonStepTest extends BasePiperTest {

    @ClassRule
    public static final SpringClassRule SPRING_CLASS_RULE = new SpringClassRule()

    private static final Map<String, Script> cache = new HashMap<>()

    @Rule
    public final SpringMethodRule springMethodRule = new SpringMethodRule()

    @Rule
    public MethodRule ruleChain = Rules.getCommonMethodRule(this)

    private String stepName

    CommonStepTest(String stepName) {
       this.stepName = stepName
    }

    @Before
    void preLoadScript(){
        if(!cache.get(this.stepName))
            cache.put(this.stepName, loadScript("${this.stepName}.groovy"))
    }

    /*
     * With that test we ensure the very first action inside a method body of a call method
     * for a not white listed step is the check for the script handed over properly.
     * Actually we assert for the exception type (AbortException) and for the exception message.
     * In case a new step is added this step will fail. It is the duty of the author of the
     * step to either follow the pattern of checking the script first or to add the step
     * to the white list.
     */
    @Test
    void scriptReferenceNotHandedOverTest() {
        List blacklist = ['handleStepErrors', 'sapCumulusUpload', 'sapCumulusDownload', 'sapCreateFosstarsReport', 'sapReportPipelineStatus', 'sapCheckPPMSCompliance', 'sapCallStagingService', 'sapCollectInsights','sapAccessContinuumExecuteTests', 'sapExecuteCustomPolicy', 'sapCollectPolicyResults']

        // assumptions
        assumeThat(blacklist, not(hasItem(this.stepName)))

        try {
            helper.registerAllowedMethod('echo', [String], { m -> System.out.println(m) })
            helper.registerAllowedMethod("library", [String.class], {return null})

            def script = cache.get(this.stepName)
            try {
                System.setProperty('com.sap.piper.featureFlag.failOnMissingScript', 'true')
                try {
                    script.call([:])
                } catch(AbortException | MissingMethodException e) {
                    throw e
                }  catch(Exception e) {
                    fail "Unexpected exception ${e.getClass().getName()} caught from step '${this.stepName}': ${e.getMessage()}"
                }
                fail("Expected AbortException not raised by step '${this.stepName}'")
            } catch(MissingMethodException e) {
                // can be improved: exception handling as some kind of control flow.
                // we can also check for the methods and call the appropriate one.
                try {
                    script.call([:]) {}
                } catch(AbortException e1) {
                    throw e1
                }  catch(Exception e1) {
                    fail "Unexpected exception ${e1.getClass().getName()} caught from step '${this.stepName}': ${e1.getMessage()}"
                }
                fail("Expected AbortException not raised by step '${this.stepName}'")
            }
        } catch(AbortException e) {
            assertThat("Step '${this.stepName}' does not fail with expected error message in case mandatory parameter 'script' is not provided.",
                e.getMessage(),  startsWith("[ERROR] No reference to surrounding script provided with key 'script', e.g. 'script: this'."))
        } finally {
            System.clearProperty('com.sap.piper.featureFlag.failOnMissingScript')
        }
    }

    private static additionalFieldBlacklist = [
        'restartableSteps',
        'sapPiperPublishNotifications',
        'sapPiperStagePromote',
        'sapReportPipelineStatus',
        'sapDwCStageRelease', //implementing new golang pattern without fields
        'sapCumulusUpload', //implementing new golang pattern without fields
        'sapCumulusDownload', //implementing new golang pattern without fields
        'sapCreateFosstarsReport', //implementing new golang pattern without fields
        'sapCheckPPMSCompliance', //implementing new golang pattern without fields
        'sapXmakeExecuteBuild', //implementing new golang pattern without fields
        'sapCallFossService', //implementing new golang pattern without fields
        'sapCheckECCNCompliance', //implementing new golang pattern without fields
        'sapCallStagingService', //implementing new golang pattern without fields
        'sapURLScan', //implementing new golang pattern without fields
        'sapDownloadArtifact', //implementing new golang pattern without fields
        'sapExecuteFastlane', //implementing new golang pattern without fields
        'sapSUPAExecuteTests', //implementing new golang pattern without fields
        'sapGenerateEnvironmentInfo', //implementing new golang pattern without fields
        'sapPipelineInit', //implementing new golang pattern without fields
        'sapCollectInsights', //implementing new golang pattern without fields
        'sapExecuteApiMetadataValidator', //implementing new golang pattern without fields
        'sapAccessContinuumExecuteTests', //implementing new golang pattern without fields
        'sapExecuteCustomPolicy', //implementing new golang pattern without fields
        'sapCollectPolicyResults', //implementing new golang pattern without fields
        'sapExecuteCentralPolicy', //implementing new golang pattern without fields
        'sapDasterExecuteScan', //implementing new golang pattern without fields
    ]

    @Test
    void generalConfigKeysSetPresentTest() {
        def fieldName = 'GENERAL_CONFIG_KEYS'
        List blacklist = additionalFieldBlacklist.plus([
            'handleStepErrors'
        ])
        // assumptions
        assumeThat(blacklist, not(hasItem(this.stepName)))
        // assertions
        assertTrue("Step ${this.stepName} has no ${fieldName} field (or that field is not a Set).",
            fieldCheck(fieldName))
    }

    @Test
    void stepConfigKeysSetPresentTest() {
        def fieldName = 'STEP_CONFIG_KEYS'
        List blacklist = additionalFieldBlacklist.plus([
            'handleStepErrors'
        ])
        // assumptions
        assumeThat(blacklist, not(hasItem(this.stepName)))
        // assertions
        assertTrue("Step ${this.stepName} has no ${fieldName} field (or that field is not a Set).",
            fieldCheck(fieldName))
    }

    @Test
    void parametersKeysSetPresentTest() {
        def fieldName = 'PARAMETER_KEYS'
        List blacklist = additionalFieldBlacklist.plus([
            'executeOpenSourceDependencyScan'
        ])
        // assumptions
        assumeThat(blacklist, not(hasItem(this.stepName)))
        // assertions
        assertTrue("Step ${this.stepName} has no ${fieldName} field (or that field is not a Set).",
            fieldCheck(fieldName))
    }

    private fieldCheck(fieldName) {
        def fields = cache.get(this.stepName).getClass().getDeclaredFields() as Set
        Field generalConfigKeyField = fields.find{ it.getName() == fieldName}
        if(! generalConfigKeyField ||
            ! generalConfigKeyField
                .getType()
                .isAssignableFrom(Set.class)) {
                    return false
        }
        return true
    }

    @Test
    void stepsWithWrongFieldNameTest() {
        def blacklist = [
            'conditionalSteps',
            'emptyNode',
            'globalPipelineEnvironment',
            'loadDefaultValues',
            'milestoneLock',
            'setBuildStatus',
            'sapPiperPipeline',
            'sapReportPipelineStatus',
            'sapCollectInsights'
        ]

        // assumptions
        assumeThat(blacklist, not(hasItem(this.stepName)))

        def script = cache.get(this.stepName)
        def fields = script.getClass().getDeclaredFields() as Set
        Field stepNameField = fields.find { it.getName() == 'STEP_NAME'}

        assertThat("Step ${this.stepName} has no STEP_NAME field.",
            stepNameField, is(not(null)))

        boolean notAccessible = false;
        def fieldName

        if(!stepNameField.isAccessible()) {
            stepNameField.setAccessible(true)
            notAccessible = true
        }

        try {
            fieldName = stepNameField.get(script)
        } finally {
            if(notAccessible) stepNameField.setAccessible(false)
        }
        assertThat("Step ${this.stepName} has no correct STEP_NAME value: ${fieldName}",
            fieldName, is(this.stepName))
    }

    /*
     * With that test we ensure that all return types of the call methods of all the steps
     * are void. Return types other than void are not possible when running inside declarative
     * pipelines. Parameters shared between several steps needs to be shared via the commonPipelineEnvironment.
     */
    @Test
    @org.junit.Ignore("too many findings")
    void returnTypeForCallMethodsIsVoidTest() {
        List blacklist = []

        // assumptions
        assumeThat(blacklist, not(hasItem(this.stepName)))

        def methods = cache.get(this.stepName).getClass().getDeclaredMethods() as List
        Collection callMethodsWithReturnTypeOtherThanVoid = methods.stream()
            .filter { it.getName() == 'call' && it.getReturnType() != Void.TYPE }
            .collect(toList())

        // assertions
        assertThat("Step ${this.stepName} has call method with return types other than void",
            callMethodsWithReturnTypeOtherThanVoid, is(empty()))
    }

    private static commonBlacklist = [
        'conditionalSteps',
        'emptyNode',
        'globalPipelineEnvironment',
        'loadDefaultValues',
        'milestoneLock',
        'sapPiperPipeline',
        'setBuildStatus'
    ]


    @Parameterized.Parameters(name = "testing {0}")
    static Collection getSteps() {
        List steps = []
        new File('vars').traverse(type: FileType.FILES, maxDepth: 0)
            { if(it.getName().endsWith('.groovy')) steps << (it =~ /vars[\\\/](.*)\.groovy/)[0][1] }
        steps -= commonBlacklist
        return steps
    }
}

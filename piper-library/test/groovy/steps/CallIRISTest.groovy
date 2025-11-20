package steps

import org.junit.Before
import org.junit.After
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import hudson.AbortException
import java.io.FileNotFoundException
import net.sf.json.JSONException

import static org.hamcrest.Matchers.*
import static org.junit.Assert.*

class CallIRISTest extends BasePiperTest {

    private Boolean verbose = true

    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsReadJsonRule jrjr = new JenkinsReadJsonRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jrjr)
        .around(jsr) // needs to be activated after jedr, otherwise executeDocker is not mocked

    @Before
    void init() {
    }

    @After
    public void cleanupMetaClass() {
        jsr.step.callIRIS.metaClass=null        // Reset metaClass.
        nullScript.currentBuild.result=null     // Reset BuildState.
    }

/**
*
*   Various successes / different version passing.
*
**/
    @Test
    void testSuccessVersionAsParam() {
        helper.registerAllowedMethod('callIRISScript', [Map.class], {m -> return '{ "resultCode": "0", "resultText": "testSuccessVersionAsParam" }'})
        helper.registerAllowedMethod('fileExists', [String.class], {m -> return true})

        jsr.step.callIRIS.metaClass.readVersionFromFile = {String s,String d, Boolean e ->
          return "0.0.0";
        }

        jsr.step.callIRIS.metaClass.loadLibrary = {String a, String b ->
            return true
        }

        jsr.step.callIRIS([
            script: nullScript, (jsr.step.callIRIS.UseCase.ITPStatusGetter): [credentialsID: 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', processID: '12345', timeOut: 123], useCaseVersion: '1.2.3'
        ])

        if (verbose) {
            println("Log's of 'testSuccessVersionAsParam'\n${jlr.log}")
        }

        assertThat(jlr.log, containsString("Using Content Version '1.2.3' for useCase 'ITPStatusGetter' (Version passed as parameter)."))
        assertThat(nullScript.currentBuild.result, is(null))
    }


    @Test
    void testSuccessVersionFromManifest() {
        helper.registerAllowedMethod('callIRISScript', [Map.class], {m -> return '{ "resultCode": "0", "resultText": "testSuccessVersionFromManifest" }'})
        helper.registerAllowedMethod('fileExists', [String.class], {m -> return true})

        jsr.step.callIRIS.metaClass.readVersionFromFile = {String s,String d, Boolean e ->
          return "1.2.3";
        }

        jsr.step.callIRIS.metaClass.loadLibrary = {String a, String b ->
            return true
        }

        jsr.step.callIRIS([
            script: nullScript, (jsr.step.callIRIS.UseCase.ITPStatusGetter): [credentialsID: 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', processID: '12345', timeOut: 123]
        ])

        if (verbose) {
            println("Log's of 'testSuccessVersionFromManifest'\n${jlr.log}")
        }

        assertThat(jlr.log, containsString("Using Content Version '1.2.3' for useCase 'ITPStatusGetter' (Version read from Manifest file 'IRISManifest.properties')."))
        assertThat(nullScript.currentBuild.result, is(null))
    }

    @Test
    void testSuccessWarningNoVersion() {
        helper.registerAllowedMethod('callIRISScript', [Map.class], {m -> return '{ "resultCode": "4", "resultText": "Test -> warning" }'})
        helper.registerAllowedMethod('fileExists', [String.class], {m -> return true})

        jsr.step.callIRIS.metaClass.readVersionFromFile = {String s,String d, Boolean e ->
          return "master";
        }

        jsr.step.callIRIS.metaClass.loadLibrary = {String a, String b ->
            return true
        }

        jsr.step.callIRIS([
            script: nullScript, (jsr.step.callIRIS.UseCase.ITPStatusGetter): [credentialsID: 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', processID: '12345', timeOut: 123]
        ])

        if (verbose) {
            println("Log's of 'testSuccessWarningNoVersion'\n${jlr.log}")
        }

        assertThat(jlr.log, containsString("Developer Notice: You did not specify a Version or simply 'master' for the useCase 'ITPStatusGetter'. In future releases this will cause IRIS to fail."))
       // assertThat(nullScript.currentBuild.result, is('UNSTABLE'))
    }

/**
*
*   Test number of useCases.
*
**/
    @Test
    void testFailureNoUseCase() {

        Boolean errorOccured = false
        try {
            jsr.step.callIRIS([
                script: nullScript
            ])
        } catch (IllegalArgumentException ex) {
            if (ex.getMessage().contains('No useCase relevant parameter given')) {
                errorOccured = true
            }
        }

        if (verbose) {
            println("Log's of 'testFailureNoUseCase'\n${jlr.log}")
        }

        assertTrue(errorOccured)
    }

    @Test
    void testFailureMultipleUseCases() {

        Boolean errorOccured = false
        try {
            jsr.step.callIRIS([
                script: nullScript, (jsr.step.callIRIS.UseCase.ITPStatusGetter):      [credentialsID: 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', processID: '12345', timeOut: 123],
                                    (jsr.step.callIRIS.UseCase.ITPSynchronousCaller): [credentialsID: 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', systemRoleCode: 'ABC', customerITP: 'TEST',
                                                                                       contactID: 'd000000', contactName: 'Max Mustermann', contactEmail: 'max.mustermann@testhausen.de']
            ])
        } catch (IllegalArgumentException ex) {
            if (ex.getMessage().contains('You cannot define another useCase')) {
                errorOccured = true
            }
        }

        if (verbose) {
            println("Log's of 'testFailureMultipleUseCases'\n${jlr.log}")
        }

        assertTrue(errorOccured)
    }

/**
*
*   Tests for method loadLibrary.
*
**/
    @Test
    void testFailureNoLibraryNamedIris() {
        helper.registerAllowedMethod('library', [String.class], {throw new AbortException("No library named iris found")})
        helper.registerAllowedMethod('fileExists', [String.class], {m -> return true})

        jsr.step.callIRIS.metaClass.readVersionFromFile = {String s,String d, Boolean e ->
          return "1.2.3";
        }

        jsr.step.callIRIS([
            script: nullScript, (jsr.step.callIRIS.UseCase.ITPStatusGetter): [credentialsID: 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', processID: '12345', timeOut: 123]
        ])

        if (verbose) {
            println("Log's of 'testFailureNoLibraryNamedIris'\n${jlr.log}")
        }

        assertThat(jlr.log, containsString("Library 'iris' is not defined in Jenkins-Settings. See documentation, how to enable: https://go.sap.corp/irisdoc/lib/setupLibrary/#setup"))
        assertThat(nullScript.currentBuild.result, is('FAILURE'))
    }

    @Test
    void testFailureNoLibraryWithVersion() {
        helper.registerAllowedMethod('library', [String.class], {throw new AbortException("No version")})
        helper.registerAllowedMethod('fileExists', [String.class], {m -> return true})

        jsr.step.callIRIS.metaClass.readVersionFromFile = {String s,String d, Boolean e ->
          return "1.2.3";
        }

        jsr.step.callIRIS([
            script: nullScript, (jsr.step.callIRIS.UseCase.ITPStatusGetter): [credentialsID: 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', processID: '12345', timeOut: 123]
        ])

        if (verbose) {
            println("Log's of 'testFailureNoLibraryWithVersion'\n${jlr.log}")
        }

        assertThat(jlr.log, containsString("Library 'iris' with Version '1.2.3' does not exist. Check for correct Version in repository https://github.wdf.sap.corp/ProjectIRIS/iris/releases"))
        assertThat(nullScript.currentBuild.result, is('FAILURE'))
    }

    @Test
    void testFailureLibraryUnknownProblem() {
        helper.registerAllowedMethod('library', [String.class], {throw new AbortException("Unknown Problem")})
        helper.registerAllowedMethod('fileExists', [String.class], {m -> return true})

        jsr.step.callIRIS.metaClass.readVersionFromFile = {String s,String d, Boolean e ->
          return "1.2.3";
        }

        jsr.step.callIRIS([
            script: nullScript, (jsr.step.callIRIS.UseCase.ITPStatusGetter): [credentialsID: 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', processID: '12345', timeOut: 123]
        ])

        if (verbose) {
            println("Log's of 'testFailureLibraryUnknownProblem'\n${jlr.log}")
        }

        assertThat(jlr.log, containsString("Trouble while loading Library 'iris' with Version '1.2.3':"))
        assertThat(nullScript.currentBuild.result, is('FAILURE'))
    }

/**
*
*   Tests for method readVersionFromFile.
*   NOTE: These tests need to be done with direct method call because otherwise we can not mock "fileExists" correctly.
*
**/

    @Test
    void testReadVersionManifestNotFound() {
        helper.registerAllowedMethod('fileExists', [String.class], {m -> return false});

        String version = jsr.step.callIRIS.readVersionFromFile("path", "keyName", true);

        if (verbose) {
            println("Log's of 'testReadVersionManifestNotFound'\n${jlr.log}");
        }

        assertThat(jlr.log, containsString("Manifest 'path' does not exist!"));
        assertTrue(version.equals(''));
    }

    @Test
    void testReadVersionManifestNotReadable() {
        helper.registerAllowedMethod('readJSON', [Map.class], {throw new JSONException("ManifestNotReadable")})
        helper.registerAllowedMethod('fileExists', [String.class], {m -> return true});

        String version = jsr.step.callIRIS.readVersionFromFile("path", "keyName", true);

        if (verbose) {
            println("Log's of 'testReadVersionManifestNotReadable'\n${jlr.log}");
        }

        assertThat(jlr.log, containsString("Trouble while reading manifest from 'path': "));
        assertTrue(version.equals(''));
    }

    @Test
    void testReadVersionUnknownKey() {
        helper.registerAllowedMethod('readJSON', [Map.class], {m -> return [useCase: [version: "0.0.0"], useCaseNextLayer: [useCase1: [version: "1.2.3"], useCase2: [notVersion: "1.0.0"]]]})

        helper.registerAllowedMethod('fileExists', [String.class], {m -> return true});

        String version = jsr.step.callIRIS.readVersionFromFile("path", "unknownKeyName", true);

        if (verbose) {
            println("Log's of 'testReadVersionUnknownKey'\n${jlr.log}");
        }

        assertTrue(version.equals('master'));
    }

    @Test
    void testReadVersionKeyNoVersion() {
        helper.registerAllowedMethod('readJSON', [Map.class], {m -> return [keyName: [notVersion: "0.0.0"]]})

        helper.registerAllowedMethod('fileExists', [String.class], {m -> return true});

        String version = jsr.step.callIRIS.readVersionFromFile("path", "keyName", false);

        if (verbose) {
            println("Log's of 'testReadVersionKeyNoVersion'\n${jlr.log}");
        }

        assertTrue(version.equals('master'));
    }

    @Test
    void testReadVersionKeyWithVersion() {
        helper.registerAllowedMethod('readJSON', [Map.class], {m -> return [keyName: [Version: "0.0.0"]]})

        helper.registerAllowedMethod('fileExists', [String.class], {m -> return true});

        String version = jsr.step.callIRIS.readVersionFromFile("path", "keyName", true);

        if (verbose) {
            println("Log's of 'testReadVersionKeyWithVersion'\n${jlr.log}");
        }

        assertTrue(version.equals('0.0.0'));
    }

    @Test
    void testReadVersionUseCaseLayerUp() {
        helper.registerAllowedMethod('readJSON', [Map.class], {m -> return [UseCase: [keyName: [Version: "0.0.0"]]]})

        helper.registerAllowedMethod('fileExists', [String.class], {m -> return true});

        String version = jsr.step.callIRIS.readVersionFromFile("path", "keyName", true);

        if (verbose) {
            println("Log's of 'testReadVersionUseCaseLayerUp'\n${jlr.log}");
        }

        assertTrue(version.equals('0.0.0'));
        assertThat(jlr.log, containsString("Manifest: Jump to useCase Level."));
    }
}

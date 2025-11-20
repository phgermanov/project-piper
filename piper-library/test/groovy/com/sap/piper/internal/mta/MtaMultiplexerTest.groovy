package com.sap.piper.internal.mta

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.hasKey
import static org.hamcrest.Matchers.hasSize
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not

import static org.junit.Assert.assertThat
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsLoggingRule
import util.Rules

class MtaMultiplexerTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(thrown)

    @Test
    void testFilterFiles() {
        // prepare test data
        def files = [
            new File("pom.xml"),
            new File("some-ui${File.separator}pom.xml"),
            new File("some-service${File.separator}pom.xml"),
            new File("some-other-service${File.separator}pom.xml")
        ].toArray()
        // execute test
        def result = MtaMultiplexer.removeExcludedFiles(nullScript, files, ['pom.xml'])
        // asserts
        assertThat(result, not(hasItem('pom.xml')))
        assertThat(result, hasSize(3))
        assertThat(jlr.log, containsString('- Skipping pom.xml'))
    }

    @Test
    void testIgnoreFiles() {
        // prepare test data
        def files = [
            new File("pom.xml"),
            new File("some-ui${File.separator}pom.xml"),
            new File("some-ui${File.separator}node_modules${File.separator}pom.xml"),
            new File("some-service${File.separator}pom.xml")
        ].toArray()
        // execute test
        def result = MtaMultiplexer.removeNodeModuleFiles(nullScript, files)
        // asserts
        assertThat(result, not(hasItem("some-ui${File.separator}node_modules${File.separator}pom.xml".toString())))
        assertThat(result, hasSize(3))
        assertThat(jlr.log, containsString("- Skipping some-ui${File.separator}node_modules${File.separator}pom.xml".toString()))
    }

    @Test
    void testCreateJobs() {
        def optionsList = []
        // prepare test data
        helper.registerAllowedMethod("findFiles", [], { return [new File("anyFile.xml")].toArray() })
        helper.registerAllowedMethod("findFiles", [Map.class], { map ->
            if (map.glob == "**${File.separator}pom.xml") {
                return [new File("some-service${File.separator}pom.xml"), new File("some-other-service${File.separator}pom.xml")].toArray()
            }
            if (map.glob == "**${File.separator}package.json") {
                return [new File("some-ui${File.separator}package.json"), new File("somer-service-broker${File.separator}package.json")].toArray()
            }
            return [].toArray()
        })
        // execute test
        def result = MtaMultiplexer.createJobs(nullScript, ['myParameters':'value'], [], 'TestJobs', 'pom.xml', 'maven'){
            options -> optionsList.push(options)
        }
        // invoke jobs
        for(Closure c : result.values()) c()
        // asserts
        assertThat(result.size(), is(2))
        assertThat(result, hasKey('TestJobs - some-other-service'))
        assertThat(jlr.log, containsString("Found 2 maven descriptor files: [some-service${File.separator}pom.xml, some-other-service${File.separator}pom.xml]".toString()))
        assertThat(optionsList.get(0), hasEntry('myParameters', 'value'))
        assertThat(optionsList.get(0), hasEntry('scanType', 'maven'))
        assertThat(optionsList.get(0), hasEntry('buildDescriptorFile', "some-service${File.separator}pom.xml".toString()))
        assertThat(optionsList.get(1), hasEntry('myParameters', 'value'))
        assertThat(optionsList.get(1), hasEntry('scanType', 'maven'))
        assertThat(optionsList.get(1), hasEntry('buildDescriptorFile', "some-other-service${File.separator}pom.xml".toString()))
    }

    @Test
    void ignoreNodeModuleTest() {
        def optionsList = []
        // prepare test data
        helper.registerAllowedMethod("findFiles", [], { return [new File("anyFile.xml")].toArray() })
        helper.registerAllowedMethod("findFiles", [Map.class], { map ->
            if (map.glob == "**${File.separator}package.json") {
                return [
                    new File("some-ui${File.separator}package.json"),
                    new File("some-service${File.separator}package.json"),
                    new File("some-service${File.separator}node_modules${File.separator}lodash${File.separator}package.json")
                ].toArray()
            }
            return [].toArray()
        })
        // execute test
        def result = MtaMultiplexer.createJobs(nullScript, ['myParameters':'value'], ["some-ui${File.separator}package.json".toString()], 'TestJobs', 'package.json', 'npm'){
            options -> optionsList.push(options)
        }
        // asserts
        assertThat(jlr.log, containsString("Found 3 npm descriptor files: [some-ui${File.separator}package.json, some-service${File.separator}package.json, some-service${File.separator}node_modules${File.separator}lodash${File.separator}package.json]".toString()))
        assertThat(result.size(), is(1))
        assertThat(result, hasKey('TestJobs - some-service'))
        assertThat(result, not(hasKey('TestJobs - some-ui')))
    }

    @Test
    void emptyFilesArrayTest() {
        def optionsList = []
        // prepare test data
        helper.registerAllowedMethod("findFiles", [], { return [new File("any.file")].toArray() })
        helper.registerAllowedMethod("findFiles", [Map.class], { [] })
        // execute test
        def result = MtaMultiplexer.createJobs(nullScript, ['myParameters':'value'], ["some-ui${File.separator}package.json".toString()], 'TestJobs', 'package.json', 'npm'){
            options -> optionsList.push(options)
        }
        // asserts
        assertThat(result.size(), is(0))
    }

    @Test
    void emptyFolderTest() {
        def optionsList = []
        // prepare test data
        helper.registerAllowedMethod("findFiles", [], { return [].toArray() })
        // execute test
        def result = MtaMultiplexer.createJobs(nullScript, ['myParameters':'value'], ["some-ui${File.separator}package.json".toString()], 'TestJobs', 'package.json', 'npm'){
            options -> optionsList.push(options)
        }
        // asserts
        assertThat(result.size(), is(0))
        assertThat(jlr.log, containsString("Found 0 npm descriptor files, the directory is empty.".toString()))
    }
}

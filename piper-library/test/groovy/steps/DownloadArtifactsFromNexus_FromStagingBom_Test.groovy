#!groovy
package steps

import util.JenkinsUnstableRule

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.containsInAnyOrder
import static org.hamcrest.Matchers.hasItems
import static org.hamcrest.Matchers.hasKey
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.hamcrest.CoreMatchers.isA

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.Rule
import org.junit.Ignore
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue

import org.yaml.snakeyaml.Yaml

import util.BasePiperTest
import util.MockHelper
import util.Rules
import util.JenkinsStepRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsExecuteDockerRule

import static com.lesfurets.jenkins.unit.MethodCall.callArgsToString
import static com.lesfurets.jenkins.unit.MethodSignature.method
import static org.junit.Assert.fail

class DownloadArtifactsFromNexus_FromStagingBom_Test extends DownloadArtifactsFromNexus_FromStaging_Test {
	@Override
    @Before
    void setUp() {
        xmakeProperties=mockHelper.loadJSON('test/resources/build_results_stage_bom.json')
        super.setUp()
    }
}

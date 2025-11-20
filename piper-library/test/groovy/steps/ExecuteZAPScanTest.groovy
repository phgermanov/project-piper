#!groovy
package steps

import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

class ExecuteZAPScanTest extends BasePiperTest {

    public ExpectedException exception = ExpectedException.none()
    public JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    public JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    public JenkinsExecuteDockerRule jedr = new JenkinsExecuteDockerRule(this)
    public JenkinsStepRule jsr = new JenkinsStepRule(this)
    public JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(exception)
        .around(jlr)
        .around(jscr)
        .around(jedr)
        .around(jsr)
        .around(jer)

    @Test
    void testDefaultValues() {
        def parameters = [script: nullScript, juStabUtils: utils, jenkinsUtilsStub: jenkinsUtils, zedAttackProxyStub: zedAttackProxy, targetUrls: ['https://server.com/A', 'https://server.com/B']]

        def stashContent
        helper.registerAllowedMethod("unstashAll", [Object.class], {
            list ->
                stashContent = list
                return list
        })

        helper.registerAllowedMethod("checkServerStatus", [Object.class], {
            map ->
                return map
        })

        helper.registerAllowedMethod("pwd", [], {
           return ''
        })

        helper.registerAllowedMethod("checkSpiderScanStatus", [Object.class,Object.class,Object.class], {})

        helper.registerAllowedMethod("checkActiveScanStatus", [Object.class,Object.class,Object.class], {})

        helper.registerAllowedMethod("checkAJAXScanStatus", [Object.class,Object.class], {})

        helper.registerAllowedMethod("writeFile", [Map.class], {
            map ->
                return map
        })

        helper.registerAllowedMethod("publishTestResults", [Map.class], {
            map ->
                return map
        })

        helper.registerAllowedMethod("findFiles", [Map.class], {
            map ->
                if(map.glob.contains("scripts"))
                    return [[path: "some${File.separator}zap${File.separator}scripts${File.separator}httpfuzzer${File.separator}fuzzer.js", name: 'fuzzer.js']].toArray()
                if(map.glob.contains("context"))
                    return [[path: "some${File.separator}zap${File.separator}context${File.separator}SomeContext.context", name: 'SomeContext.context']].toArray()
        })

        jsr.step.executeZAPScan(parameters)
    }
}

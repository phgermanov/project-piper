package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsShellCallRule implements TestRule {

    final BasePipelineTest testInstance

    List shell = []

    def returnValues = [:]
    def returnStatus = [:]

    JenkinsShellCallRule(BasePipelineTest testInstance) {
        this.testInstance = testInstance
    }

    def setReturnValue(script, value) {
        returnValues[script.replaceAll(/\s+/, " ").trim()] = value
    }

    def setReturnStatus(script, value) {
        returnStatus[script.replaceAll(/\s+/, " ").trim()] = value
    }

    @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                testInstance.helper.registerAllowedMethod("sh", [String.class], {
                    command ->
                        shell.add(command.replaceAll(/\s+/, " ").trim())
                })

                testInstance.helper.registerAllowedMethod("sh", [Map.class], {
                    m ->
                        def cleanedScript = m.script.replaceAll(/\s+/, " ").trim()
                        shell.add(cleanedScript)
                        if (m.returnStdout) {
                            return returnValues[cleanedScript]

                        } else if (m.returnStatus) {
                            if (returnStatus[cleanedScript]) {
                                return returnStatus[cleanedScript]
                            } else {
                                return 0
                            }
                        } else {
                            return null
                        }
                })

                base.evaluate()
            }
        }
    }
}

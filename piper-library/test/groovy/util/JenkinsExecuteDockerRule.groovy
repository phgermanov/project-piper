package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsExecuteDockerRule implements TestRule {

    final BasePipelineTest testInstance

    def dockerParams = [:]

    JenkinsExecuteDockerRule(BasePipelineTest testInstance) {
        this.testInstance = testInstance
    }

    @Override
    Statement apply(Statement base, Description description) {
        return statement(base)
    }

    private Statement statement(final Statement base) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {

                testInstance.helper.registerAllowedMethod("executeDocker", [Map.class, Closure.class], {map, closure ->
                    dockerParams = map
                    return closure()
                })

                //take care of steps using Piper Open Source version of the step
                testInstance.helper.registerAllowedMethod("dockerExecute", [Map.class, Closure.class], {map, closure ->
                    dockerParams = map
                    return closure()
                })

                base.evaluate()
            }
        }
    }
}

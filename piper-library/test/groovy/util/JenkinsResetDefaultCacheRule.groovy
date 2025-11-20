package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import com.sap.piper.internal.DefaultValueCache
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsResetDefaultCacheRule implements TestRule {


    JenkinsResetDefaultCacheRule() {
        this(null)
    }

    //
    // Actually not needed. Only provided for the sake of consistency
    // with our other rules which comes with an constructor having the
    // test case contained in the signature.
    JenkinsResetDefaultCacheRule(BasePipelineTest testInstance) {
    }

    @Override
    Statement apply(Statement base, Description description) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {
                DefaultValueCache.reset()
                base.evaluate()
            }
        }
    }
}

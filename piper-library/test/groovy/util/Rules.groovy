package util

import org.junit.rules.MethodRule
import org.junit.rules.RuleChain

import com.lesfurets.jenkins.unit.BasePipelineTest
import com.lesfurets.jenkins.unit.global.lib.LibraryConfiguration
import org.junit.runners.model.FrameworkMethod
import org.junit.runners.model.Statement

class Rules {

    static RuleChain getCommonRules(BasePipelineTest testCase) {
        return getCommonRules(testCase, null)
    }

    static RuleChain getCommonRules(BasePipelineTest testCase, LibraryConfiguration libConfig) {
        return RuleChain.outerRule(new JenkinsSetupRule(testCase, libConfig))
            .around(new JenkinsReadYamlRule(testCase))
            .around(new JenkinsResetDefaultCacheRule())
            .around(new JenkinsErrorRule(testCase))
            .around(new JenkinsHandlePipelineStepErrorsRule(testCase))
            .around(new JenkinsEnvironmentRule(testCase))
    }

    static MethodRule getCommonMethodRule(BasePipelineTest testCase) {
        return new MethodRule() {
            @Override
            Statement apply(Statement base, FrameworkMethod method, Object target) {
                return getCommonRules(testCase).apply(base, null)
            }
        }
    }
}

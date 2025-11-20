package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.codehaus.groovy.control.CompilerConfiguration
import org.codehaus.groovy.runtime.DefaultGroovyMethods
import org.junit.rules.TestRule
import org.junit.runner.Description
import org.junit.runners.model.Statement

class JenkinsEnvironmentRule implements TestRule {
    final BasePipelineTest testInstance

    def env

    JenkinsEnvironmentRule(BasePipelineTest testInstance) {
        this.testInstance = testInstance
    }

    @Override
    Statement apply(Statement base, Description description) {
        return new Statement() {
            @Override
            void evaluate() throws Throwable {
                env = testInstance.loadScript('globalPipelineEnvironment.groovy').globalPipelineEnvironment
                testInstance?.nullScript.globalPipelineEnvironment = env

                def groovyLoader = new GroovyClassLoader(this.getClass().getClassLoader(), new CompilerConfiguration())
                testInstance.helper.scriptRoots.each { scriptRoot ->
                    def file = new File(scriptRoot)
                    groovyLoader.addURL(file.toPath().toUri().toURL())
                }
                def clazz = groovyLoader.loadClass('commonPipelineEnvironment')
                def object = DefaultGroovyMethods.newInstance(clazz)
                testInstance.binding.setVariable('commonPipelineEnvironment', object)
                testInstance?.nullScript.commonPipelineEnvironment = object

                base.evaluate()
            }
        }
    }
}

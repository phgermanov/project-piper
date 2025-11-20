#!/usr/bin/env groovy
// If there's a call method, you can just load the file, say, as "foo", and then invoke that call method with foo(...)
def call(script) {
    node {
        echo "Integration Test"
        echo "Test var: ${script.globalPipelineEnvironment.getStepConfiguration('step1', 'Integration').get('test1')}"
    }
}
return this;

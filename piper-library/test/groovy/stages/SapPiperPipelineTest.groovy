#!groovy
package stages

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertThat

class SapPiperPipelineTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jsr)

    private skipDefaultCheckout = false
    private timestamps = false
    private stagesExecuted = []
    private stepsCalled = []

    @Before
    void init() {

        helper.registerAllowedMethod('library', [String.class], null)

        helper.registerAllowedMethod('pipeline', [Closure.class], null)

        helper.registerAllowedMethod('agent', [Closure.class], null)

        helper.registerAllowedMethod('script', [Closure.class], {c -> c() })

        binding.setVariable('any', {})
        binding.setVariable('none', {})

        helper.registerAllowedMethod('options', [Closure.class], null)

        helper.registerAllowedMethod('skipDefaultCheckout', [], {skipDefaultCheckout = true})
        helper.registerAllowedMethod('timestamps', [], {timestamps = true})

        helper.registerAllowedMethod('triggers', [Closure.class], null)
        helper.registerAllowedMethod('issueCommentTrigger', [String.class], { s ->
            assertThat(s, is('.*/piper ([a-z]*).*'))
        })

        helper.registerAllowedMethod('stages', [Closure.class], null)

        helper.registerAllowedMethod('stage', [String.class, Closure.class], {stageName, body ->

            def stageResult

            binding.variables.env.STAGE_NAME = stageName

            helper.registerAllowedMethod('when', [Closure.class], {cWhen ->

                helper.registerAllowedMethod('beforeAgent', [Boolean.class], {agentBoolean  ->
                    return agentBoolean
                })

                helper.registerAllowedMethod('allOf', [Closure.class], {cAllOf ->
                    def branchResult = false
                    helper.registerAllowedMethod('branch', [String.class], {branchName  ->
                        if (!branchResult)
                            branchResult = (branchName == env.BRANCH_NAME)
                        if( !branchResult) {
                            throw new PipelineWhenException("Stage '${stageName}' skipped - expression: '${branchResult}'")
                        }
                    })
                    helper.registerAllowedMethod('expression', [Closure.class], { Closure cExp ->
                        def result = cExp()
                        if(!result) {
                            throw new PipelineWhenException("Stage '${stageName}' skipped - expression: '${result}'")
                        }
                        return result
                    })
                    return cAllOf()
                })

                helper.registerAllowedMethod('anyOf', [Closure.class], {cAnyOf ->
                    def result = false
                    helper.registerAllowedMethod('branch', [String.class], {branchName  ->
                        if (!result)
                            result = (branchName == env.BRANCH_NAME)
                    })
                    helper.registerAllowedMethod('expression', [Closure.class], { Closure cExp ->
                        if (!result)
                            result = cExp()
                    })
                    helper.registerAllowedMethod('allOf', [Closure.class], {cAllOf ->
                        def allOfResult = false
                        helper.registerAllowedMethod('branch', [String.class], {branchName  ->
                            if (branchName) {
                                allOfResult = (branchName == env.BRANCH_NAME)
                                if( !allOfResult) {
                                    throw new PipelineWhenException("Stage '${stageName}' skipped - expression: '${allOfResult}'")
                                }
                                return true
                            }
                        })
                        helper.registerAllowedMethod('expression', [Closure.class], { Closure cExp ->
                            allOfResult = cExp()
                            if( !allOfResult) {
                                throw new PipelineWhenException("Stage '${stageName}' skipped - expression: '${allOfResult}'")
                            }
                            return true
                        })
                        try {
                            cAllOf()
                        } catch (ex) {
                            if (!result)
                               return false
                        }
                        if (!result) {
                            result = allOfResult
                        }
                    })
                    cAnyOf()
                    if(!result) {
                        throw new PipelineWhenException("Stage '${stageName}' skipped - anyOf: '${result}'")
                    }
                    return cAnyOf()
                })

                helper.registerAllowedMethod('branch', [String.class], {branchName  ->
                    def result =  (branchName == env.BRANCH_NAME)
                    if(result == false) {
                        throw new PipelineWhenException("Stage '${stageName}' skipped - expected branch: '${branchName}' while current branch: '${env.BRANCH_NAME}'")
                    }
                    return result
                })

                helper.registerAllowedMethod('expression', [Closure.class], { Closure cExp ->
                    def result = cExp()
                    if(!result) {
                        throw new PipelineWhenException("Stage '${stageName}' skipped - expression: '${result}'")
                    }
                    return result
                })
                return cWhen()
            })

            // Stage is not executed if build fails or aborts
            def status = currentBuild.result
            switch (status) {
                case 'FAILURE':
                case 'ABORTED':
                    break
                default:
                    try {
                        stageResult = body()
                        stagesExecuted.add(stageName)
                    }
                    catch (PipelineWhenException pwe) {
                        //skip stage due to not met when expression
                    }
                    catch (Exception e) {
                        throw e
                    }
            }
            return stageResult
        })

        helper.registerAllowedMethod('steps', [Closure], null)
        helper.registerAllowedMethod('post', [Closure], null)
        helper.registerAllowedMethod('always', [Closure], {c -> c()})
        helper.registerAllowedMethod('success', [Closure], {c -> c()})
        helper.registerAllowedMethod('failure', [Closure], {c -> c()})
        helper.registerAllowedMethod('aborted', [Closure], {c -> c()})
        helper.registerAllowedMethod('unstable', [Closure], {c -> c()})
        helper.registerAllowedMethod('cleanup', [Closure], {c -> c()})

        helper.registerAllowedMethod('sapPiperStageInit', [Map.class], {m ->
            stepsCalled.add('sapPiperStageInit')
        })
        helper.registerAllowedMethod('sapPiperStagePRVoting', [Map.class], {m ->
            stepsCalled.add('sapPiperStagePRVoting')
        })
        helper.registerAllowedMethod('sapPiperStageCentralBuild', [Map.class], {m ->
            stepsCalled.add('sapPiperStageCentralBuild')
        })
        helper.registerAllowedMethod('sapPiperStageAdditionalUnitTests', [Map.class], {m ->
            stepsCalled.add('sapPiperStageAdditionalUnitTests')
        })
        helper.registerAllowedMethod('sapPiperStageIntegration', [Map.class], {m ->
            stepsCalled.add('sapPiperStageIntegration')
        })
        helper.registerAllowedMethod('sapPiperStageAcceptance', [Map.class], {m ->
            stepsCalled.add('sapPiperStageAcceptance')
        })
        helper.registerAllowedMethod('sapPiperStageSecurity', [Map.class], {m ->
            stepsCalled.add('sapPiperStageSecurity')
        })
        helper.registerAllowedMethod('sapPiperStagePerformance', [Map.class], {m ->
            stepsCalled.add('sapPiperStagePerformance')
        })
        helper.registerAllowedMethod('sapPiperStageIPScanPPMS', [Map.class], {m ->
            stepsCalled.add('sapPiperStageIPScanPPMS')
        })
        helper.registerAllowedMethod('piperPipelineStageConfirm', [Map.class], {m ->
            stepsCalled.add('piperPipelineStageConfirm')
        })
        helper.registerAllowedMethod('sapPiperStagePromote', [Map.class], {m ->
            stepsCalled.add('sapPiperStagePromote')
        })
        helper.registerAllowedMethod('sapPiperStageRelease', [Map.class], {m ->
            stepsCalled.add('sapPiperStageRelease')
        })
        helper.registerAllowedMethod('piperPipelineStagePost', [Map.class], {m ->
            stepsCalled.add('piperPipelineStagePost')
        })
        helper.registerAllowedMethod('sapPiperPipelineStagePost', [Map.class], {m ->
            stepsCalled.add('sapPiperPipelineStagePost')
        })


        nullScript.loadDefaultValues(script: nullScript)

    }

    @Test
    void testPRVoting() {

        helper.registerAllowedMethod('sapPiperStageInit', [Map], null)

        binding.variables.env.BRANCH_NAME = 'PR-*'

        nullScript.globalPipelineEnvironment.configuration = [runStage:[Integration:[test: 'test']]]
        jsr.step.sapPiperPipeline(script: nullScript)

        assertThat(skipDefaultCheckout, is(true))
        assertThat(timestamps, is(true))

        assertThat(stagesExecuted.size(), is(2))
        assertThat(stagesExecuted, allOf(hasItem('Init'), hasItem('Pull-Request Voting')))

        assertThat(stepsCalled, hasItem('sapPiperStagePRVoting'))
    }

    @Test
    void testNotProductiveBranch() {

        helper.registerAllowedMethod('sapPiperStageInit', [Map], null)

        binding.variables.env.BRANCH_NAME = 'master'

        nullScript.globalPipelineEnvironment.configuration = [  general: [productiveBranch: 'not_master']]
//                                                                runStage:[Integration:[test: 'test']]]
        jsr.step.sapPiperPipeline(script: nullScript)

        assertThat(skipDefaultCheckout, is(true))
        assertThat(timestamps, is(true))

        assertThat(stagesExecuted.size(), is(1))
        assertThat(stagesExecuted, allOf(hasItem('Init')))
    }

    @Test
    void testProductiveBranchNotMaster() {

        helper.registerAllowedMethod('sapPiperStageInit', [Map], null)

        binding.variables.env.BRANCH_NAME = 'myProductiveBranch'

        nullScript.globalPipelineEnvironment.configuration = [general: [productiveBranch: 'myProductiveBranch'],
                                                                runStage:[Integration:[test: 'test']]]
        jsr.step.sapPiperPipeline(script: nullScript)

        assertThat(skipDefaultCheckout, is(true))
        assertThat(timestamps, is(true))
        assertEquals(['Init','Central Build','Integration','Confirm','Promote'], stagesExecuted)
        assertThat(stagesExecuted.size(), is(5))
    }


    @Test
    void testProductiveBranchAsRegex() {

        helper.registerAllowedMethod('sapPiperStageInit', [Map], null)

        binding.variables.env.BRANCH_NAME = 'live'

        nullScript.globalPipelineEnvironment.configuration = [  general: [productiveBranch: 'master|rc|live'],
                                                                runStage:['Additional Unit Tests':[test: 'test']]]
        jsr.step.sapPiperPipeline(script: nullScript)

        assertThat(skipDefaultCheckout, is(true))
        assertThat(timestamps, is(true))
        assertEquals(['Init','Central Build','Additional Unit Tests','Confirm','Promote'], stagesExecuted )
        assertThat(stagesExecuted.size(), is(5))
    }

    @Test
    void testConfirm() {
        jsr.step.sapPiperPipeline(script: nullScript)

        assertThat(stepsCalled, hasItem('piperPipelineStageConfirm'))

    }

    @Test
    void testConfirmUnstable() {
        nullScript.globalPipelineEnvironment.configuration = [
            general: [
                manualConfirmation: false
            ]
        ]
        binding.setVariable('currentBuild', [
            result: 'UNSTABLE'
        ])
        jsr.step.sapPiperPipeline(script: nullScript)

        assertThat(stepsCalled, hasItem('piperPipelineStageConfirm'))

    }

    @Test
    void testNoConfirm() {
        nullScript.globalPipelineEnvironment.configuration = [
            general: [
                manualConfirmation: false
            ]
        ]
        jsr.step.sapPiperPipeline(script: nullScript)

        assertThat(stepsCalled, not(hasItem('piperPipelineStageConfirm')))
    }

    @Test
    void testMasterPipeline() {
        jsr.step.sapPiperPipeline(script: nullScript)

        assertThat(stepsCalled, hasItem('piperPipelineStagePost'))
        assertThat(stepsCalled, hasItem('sapPiperPipelineStagePost'))
    }

    @Test
    void testMasterPipelineWithProductiveBranchPattern() {
        binding.variables.env.BRANCH_NAME = 'anyOtherbranch'
        nullScript.globalPipelineEnvironment.configuration = [
            general: [productiveBranch: '.*branch'],
            runStage: [
                Build: true,
                'Additional Unit Tests': true,
                Integration: true,
                Acceptance: true,
                Security: true,
                Performance: true,
                Promote: true,
                Release: true
            ]
        ]

        jsr.step.sapPiperPipeline(script: nullScript)

        assertEquals([
            'sapPiperStageInit',
              'sapPiperStageCentralBuild',
              'sapPiperStageAdditionalUnitTests',
              'sapPiperStageIntegration',
              'sapPiperStageAcceptance',
              'sapPiperStageSecurity',
              'sapPiperStagePerformance',
              'piperPipelineStageConfirm',
              'sapPiperStagePromote',
              'sapPiperStageRelease',
              'sapPiperPipelineStagePost',
              'piperPipelineStagePost'
        ], stepsCalled)
    }

    @Test
    void testOptimizedCommit() {
        nullScript.commonPipelineEnvironment.configuration = [
            general: [pipelineOptimization: true]
        ]
        nullScript.globalPipelineEnvironment.configuration = [
            runStage: [
                Build: true,
                'Additional Unit Tests': true,
                Integration: true,
                Acceptance: true,
                Security: true,
                Performance: true,
                "IPScan and PPMS": true,
                Promote: true,
                Release: true
            ]
        ]

        nullScript.commonPipelineEnvironment.setValue('isOptimizedAndScheduled', false)

        jsr.step.sapPiperPipeline(script: nullScript)

        assertEquals([
            'sapPiperStageInit',
            'sapPiperStageCentralBuild',
            'sapPiperStageAdditionalUnitTests',
            'sapPiperStageIntegration',
            'sapPiperStageAcceptance',
            'sapPiperStagePerformance',
            'piperPipelineStageConfirm',
            'sapPiperStagePromote',
            'sapPiperStageRelease',
            'sapPiperPipelineStagePost',
            'piperPipelineStagePost'
        ], stepsCalled)
    }

    @Test
    void testOptimizedScheduled() {
        nullScript.commonPipelineEnvironment.configuration = [
            general: [pipelineOptimization: true]
        ]
        nullScript.globalPipelineEnvironment.configuration = [
            runStage: [
                Build: true,
                'Additional Unit Tests': true,
                Integration: true,
                Acceptance: true,
                Security: true,
                Performance: true,
                "IPScan and PPMS": true,
                Promote: true,
                Release: true
            ]
        ]

        nullScript.commonPipelineEnvironment.setValue('isOptimizedAndScheduled', true)

        jsr.step.sapPiperPipeline(script: nullScript)

        assertEquals([
            'sapPiperStageInit',
            'sapPiperStageSecurity',
            'sapPiperStageIPScanPPMS',
            'sapPiperPipelineStagePost',
            'piperPipelineStagePost'
        ], stepsCalled)
    }
}

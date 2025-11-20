void call(parameters) {
    pipeline {
        agent none
        triggers {
            issueCommentTrigger('.*/piper ([a-z]*).*')
        }
        options {
            skipDefaultCheckout()
            timestamps()
        }
        environment {
            // PIPER_PIPELINE_TEMPLATE_NAME is for internal use by Piper team only.
            // If you've copied/forked this pipeline, please remove this line.
            PIPER_PIPELINE_TEMPLATE_NAME = 'hyperspace-piper-gpp'
        }
        stages {
            stage('Init') {
                steps {
                    library 'piper-lib-os'
                    sapPiperStageInit script: parameters.script, customDefaults: parameters.customDefaults, nodeLabel: parameters.initNodeLabel, skipCheckout: parameters.skipCheckout, stashContent: parameters.stashContent
                }
            }
            stage('Pull-Request Voting') {
                when {
                    beforeAgent true
                    anyOf {
                        allOf {
                            branch 'PR-*';
                            expression { return !parameters.script.globalPipelineEnvironment.getStepConfiguration('', '').nativeVoting };
                        }
                        branch parameters.script.globalPipelineEnvironment.getStepConfiguration('sapPiperStagePRVoting', 'Pull-Request Voting').customVotingBranch;
                    }
                }
                steps {
                    sapPiperStagePRVoting script: parameters.script
                }
            }
            stage('Central Build') {
                when {
                    beforeAgent true
                    anyOf {
                        // build for commit run for productive branch
                        allOf {
                            expression { env.BRANCH_NAME ==~ parameters.script.globalPipelineEnvironment.getStepConfiguration('', '').productiveBranch};
                            expression { return !parameters.script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled') };
                        }
                        // native PR voting capability if activated
                        allOf {
                            expression { return parameters.script.globalPipelineEnvironment.getStepConfiguration('', '').nativeBuild };
                            expression { return parameters.script.globalPipelineEnvironment.getStepConfiguration('', '').nativeVoting};
                            branch 'PR-*';
                        }
                    }
                }
                steps {
                    sapPiperStageCentralBuild script: parameters.script
                }
            }
            stage('Additional Unit Tests') {
                when {
                    beforeAgent true
                    allOf {
                        expression { env.BRANCH_NAME ==~ parameters.script.globalPipelineEnvironment.getStepConfiguration('', '').productiveBranch };
                        expression { return parameters.script.globalPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME) };
                        expression { return !parameters.script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled') };
                    }
                }
                steps {
                    sapPiperStageAdditionalUnitTests script: parameters.script
                }
            }
            stage('Integration') {
                when {
                    beforeAgent true
                    allOf {
                        expression { env.BRANCH_NAME ==~ parameters.script.globalPipelineEnvironment.getStepConfiguration('', '').productiveBranch };
                        expression { return parameters.script.globalPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME) };
                        expression { return !parameters.script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled') }
                    }
                }
                steps {
                    sapPiperStageIntegration script: parameters.script
                }
            }
            stage('Acceptance') {
                when {
                    beforeAgent true
                    allOf {
                        expression { env.BRANCH_NAME ==~ parameters.script.globalPipelineEnvironment.getStepConfiguration('', '').productiveBranch };
                        expression { return parameters.script.globalPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME) };
                        expression { return !parameters.script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled') };
                    }
                }
                steps {
                    sapPiperStageAcceptance script: parameters.script
                }
            }
            stage('Security') {
                when {
                    beforeAgent true
                    allOf {
                        expression { env.BRANCH_NAME ==~ parameters.script.globalPipelineEnvironment.getStepConfiguration('', '').productiveBranch };
                        expression { return parameters.script.globalPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME) };
                        anyOf {
                            expression { return !parameters.script.commonPipelineEnvironment.getStepConfiguration('', '').pipelineOptimization};
                            expression { return parameters.script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled') }
                        }
                    }
                }
                steps {
                    sapPiperStageSecurity script: parameters.script
                }
            }
            stage('Performance') {
                when {
                    beforeAgent true
                    allOf {
                        expression { env.BRANCH_NAME ==~ parameters.script.globalPipelineEnvironment.getStepConfiguration('', '').productiveBranch };
                        expression { return parameters.script.globalPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME) };
                        expression { return !parameters.script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled') }
                    }
                }
                steps {
                    sapPiperStagePerformance script: parameters.script
                }
            }
            stage('IPScan and PPMS') {
                when {
                    beforeAgent true
                    allOf {
                        expression { env.BRANCH_NAME ==~ parameters.script.globalPipelineEnvironment.getStepConfiguration('', '').productiveBranch };
                        expression { return parameters.script.globalPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME) };
                        anyOf {
                            expression { return !parameters.script.commonPipelineEnvironment.getStepConfiguration('', '').pipelineOptimization};
                            expression { return parameters.script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled') }
                        }
                    }
                }
                steps {
                    sapPiperStageIPScanPPMS script: parameters.script
                }
            }
            stage('Confirm') {
                agent none
                when {
                    beforeAgent true
                    allOf {
                        expression { env.BRANCH_NAME ==~ parameters.script.globalPipelineEnvironment.getStepConfiguration('', '').productiveBranch };
                        expression { return !parameters.script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled') };
                        anyOf {
                            expression { return (currentBuild.result == 'UNSTABLE') };
                            expression { return parameters.script.globalPipelineEnvironment.getStepConfiguration('sapPiperInitRunStageConfiguration', env.STAGE_NAME).manualConfirmation };
                        }
                    }
                }
                steps {
                    piperPipelineStageConfirm script: parameters.script
                }
            }
            stage('Promote') {
                when {
                    beforeAgent true
                    allOf {
                        expression { env.BRANCH_NAME ==~ parameters.script.globalPipelineEnvironment.getStepConfiguration('', '').productiveBranch };
                        expression { return !parameters.script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled') };
                    }
                }
                steps {
                    sapPiperStagePromote script: parameters.script
                }
            }
            stage('Release') {
                when {
                    beforeAgent true
                    allOf {
                        expression { env.BRANCH_NAME ==~ parameters.script.globalPipelineEnvironment.getStepConfiguration('', '').productiveBranch };
                        expression { return parameters.script.globalPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME) };
                        expression { return !parameters.script.commonPipelineEnvironment.getValue('isOptimizedAndScheduled') }
                    }
                }
                steps {
                    sapPiperStageRelease script: parameters.script
                }
            }
        }
        post {
            /* https://jenkins.io/doc/book/pipeline/syntax/#post */
            success {
                setBuildStatus(currentBuild)
            }
            aborted {setBuildStatus(currentBuild, 'ABORTED')}
            failure {
                setBuildStatus(currentBuild, 'FAILURE')
            }
            unstable {setBuildStatus(currentBuild, 'UNSTABLE')}
            cleanup {
                // sapReportPipelineStatus shouldn't report errors in Post stage, so run that step first
                sapPiperPipelineStagePost script: parameters.script
                piperPipelineStagePost script: parameters.script
            }
        }
    }
}

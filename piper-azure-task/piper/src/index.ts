'use strict'

import path from 'node:path'
import * as taskLib from 'azure-pipelines-task-lib/task'
import { Telemetry } from './telemetry'
import { cacheStd } from './std'
import { fetchBinaryVersionTagWrapper, pickBinary, preparePiperBinaryWrapper } from './binary'
import { initializeGlobals } from './globals'
import {
  checkIfStepActive,
  createCheckIfStepActiveMaps,
  preserveDefaultConfig,
  preserveStageConditions,
  readContextConfig,
  restorePipelineDefaults,
  restorePipelineStageConditions
} from './config'
import { loadPipelineEnvFromDisk, loadPipelineEnvToDisk } from './environment'
import { startContainers, stopContainers } from './docker'
import { executePiperStep } from './piper'
import {
  warnAboutDeprecatedInputs,
  handleDeprecatedCachingVariables,
  handleDeprecatedCachingVariablesOS
} from './legacy'

async function run (): Promise<void> {
  initializeGlobals()
  warnAboutDeprecatedInputs()

  const cache = cacheStd()
  const telemetry = new Telemetry()

  taskLib.setResourcePath(path.join(__dirname, '..', 'task.json'))

  exposeSystemTrustTokenToEnv()

  const fetchVersionAndExit = taskLib.getBoolInput('fetchPiperBinaryVersionTag', false)
  if (fetchVersionAndExit) {
    await fetchBinaryVersionTagWrapper()  // fetch and cache version of binary and write to Azure variable
    return Promise.resolve()
  }

  await handleDeprecatedCachingVariables()
  await handleDeprecatedCachingVariablesOS()

  await preparePiperBinaryWrapper()  // fetch and cache version of binary
    .then(pickBinary)  // determine which binary to use
    .then(restorePipelineStageConditions)
    .then(loadPipelineEnvToDisk)
    .then(restorePipelineDefaults)
    .then(checkIfStepActive)
    .then(createCheckIfStepActiveMaps)
    .then(readContextConfig)
    .then(startContainers)
    .then(executePiperStep)
    .finally(stopContainers)
    .then(() => taskLib.setResult(taskLib.TaskResult.Succeeded, ''))
    .catch((error) => {
      if (taskLib.getVariable('__stepActive') === 'false') {
        taskLib.setResult(taskLib.TaskResult.Skipped, '')
      } else {
        taskLib.setResult(taskLib.TaskResult.Failed, error)

        const contents = cache.content
        const stdLines = contents.split('\n')
        const telemetryOccurrences = (contents.match(/Sending telemetry data/g) || []).length
        taskLib.debug(`Contents lines: ${stdLines.length}, occurrences of "Sending telemetry data" ${telemetryOccurrences}`)

        // These logged data will be processed by sapReportPipelineStatus step in Post stage
        if (telemetryOccurrences === 0) {
          telemetry.setPostExecutionData(error.message)
          console.log('Logging telemetry data from Azure instead of Piper')
          console.log(`Step telemetry data:${JSON.stringify(telemetry.data())}`)
        }
      }
    })
    .finally(preserveDefaultConfig)
    .finally(preserveStageConditions)
    .finally(loadPipelineEnvFromDisk)
}

// PIPER_systemTrustToken is an envvar that is used to pass the system trust token to the piper binary and for getting
// artifactory token, that is used to pull images from common registry.
function exposeSystemTrustTokenToEnv () {
  const token = taskLib.getVariable('systemTrustToken')
  if (token) {
    taskLib.debug('using system trust token from environment variables')
    process.env.PIPER_systemTrustToken = token
  } else {
    taskLib.debug('no system trust token detected')
  }
}

run()

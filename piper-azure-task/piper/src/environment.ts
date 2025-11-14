import * as taskLib from 'azure-pipelines-task-lib/task'
import { decode, encode } from './utils'
import fs from 'fs'
import { BINARY_NAME_VAR } from './globals'

// loadPipelineEnvToDisk reads base64 encoded 'pipelineEnvironment_b64' task variable, decodes it and
// writes result to '.pipeline/commonPipelineEnvironment' file on disk.
// Does not run if '.pipeline/commonPipelineEnvironment' already exists or pipelineEnvironment_b64 is unset
export async function loadPipelineEnvToDisk (): Promise<void> {
  const restorePipelineEnv = taskLib.getInput('restorePipelineEnv')
  if (restorePipelineEnv !== undefined) { // TODO: delete this condition after deleting 'restorePipelineEnv' input from task.json
    taskLib.warning('"restorePipelineEnv" is a deprecated input and does not have any impact. ' +
        'Please remove this input from the task, as it will be removed in February 2024.')
  }

  if (fs.existsSync('.pipeline/commonPipelineEnvironment')) {
    taskLib.debug('loading pipelineEnv to disk is skipped as ".pipeline/commonPipelineEnvironment" already exists')
    return
  }
  const cpeB64String = taskLib.getVariable('pipelineEnvironment_b64')
  if (cpeB64String == null) {
    taskLib.debug('loading pipelineEnv to disk is skipped as "pipelineEnvironment_b64" variable is not set')
    return
  }

  taskLib.debug('loading pipelineEnv to disk')
  return decodePipelineEnv(cpeB64String).then(writePipelineEnv)
}

async function decodePipelineEnv (cpeB64String: any): Promise<string> {
  taskLib.debug('decoding pipeline env')
  const cpeString = decode(cpeB64String)
  taskLib.debug('decoded pipeline env: ' + cpeString)
  return Promise.resolve(cpeString)
}

async function writePipelineEnv (cpeString: any): Promise<void> {
  const binaryName = taskLib.getVariable(BINARY_NAME_VAR) ?? ''
  if (!binaryName) {
    return Promise.reject(new Error('piper binary not available'))
  }
  taskLib.debug('writing pipeline env')

  let size = Buffer.byteLength(cpeString, 'utf-8');  // size in bytes for string
  taskLib.debug(`Size of cpeString: ${size} bytes`);

  taskLib
    .tool(binaryName)
    .arg('writePipelineEnv')
    // TODO: defaults needed?
    .execSync({ env: { PIPER_pipelineEnv: cpeString } })
  return Promise.resolve()
}

// loadPipelineEnvFromDisk reads content of '.pipeline/commonPipelineEnvironment', encodes to base64 string and
// result is set into 'PipelineEnv' stage variable.
export async function loadPipelineEnvFromDisk (): Promise<void> {
  let exportPipelineEnv: boolean
  const preservePipelineEnv = taskLib.getBoolInput('preservePipelineEnv')
  if (preservePipelineEnv) { // Note: delete this condition after deleting 'preservePipelineEnv' input from task.json
    exportPipelineEnv = true
    taskLib.warning('"preservePipelineEnv" is a deprecated input. ' +
        'Please use "exportPipelineEnv" instead, as "preservePipelineEnv" will be removed in will be removed in February 2024.')
  } else {
    exportPipelineEnv = taskLib.getBoolInput('exportPipelineEnv', false)
  }

  if (!exportPipelineEnv) {
    return Promise.resolve()
  }
  taskLib.debug('loading pipelineEnv from disk')

  return readPipelineEnv()
    .then(encodePipelineEnv)
    .then(publishPipelineEnv)
    .catch(taskLib.error)
}

async function readPipelineEnv (): Promise<string> {
  const binaryName = taskLib.getVariable(BINARY_NAME_VAR) ?? ''
  if (!binaryName) {
    return Promise.reject(new Error('piper binary not available'))
  }
  taskLib.debug('reading pipeline env')

  // read pipeline env
  const cpeString = taskLib
    .tool(binaryName)
    .arg('readPipelineEnv')
    // TODO: defaults needed?
    .execSync().stdout
  taskLib.debug('read pipeline env: ' + cpeString)

  return Promise.resolve(cpeString)
}

async function encodePipelineEnv (cpeString: any): Promise<string> {
  taskLib.debug('encoding pipeline env')
  const cpeB64String = encode(cpeString)
  taskLib.debug('encoded pipeline env: ' + cpeB64String)
  return Promise.resolve(cpeB64String)
}

async function publishPipelineEnv (cpeB64String: any): Promise<void> {
  taskLib.debug('publishing pipeline env')
  taskLib.setVariable('PipelineEnv', cpeB64String, false, true)
  return Promise.resolve()
}

export const testExports = {
  publishPipelineEnv,
  encodePipelineEnv,
  decodePipelineEnv,
  readPipelineEnv,
  writePipelineEnv,
  loadPipelineEnvFromDisk,
  loadPipelineEnvToDisk
}

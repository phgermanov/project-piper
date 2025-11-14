'use strict'

import path from 'node:path'
import * as os from 'os'
import * as taskLib from 'azure-pipelines-task-lib/task'
import { prepareVaultEnvVars } from './vault'
import { getCustomConfigLocation, getStageName } from './configHelper'
import { DEFAULTS_FILE_LIST } from './config'
import { CONTAINER_VARIABLE_NAME } from './docker'
import { makeCWDWriteable } from './utils'
import { BINARY_NAME_VAR, STEP_NAME } from './globals'

function movePiperToCWD (binaryName: string) {
  const piperDownloadPath = taskLib.which(binaryName, true)
  const cwd = taskLib.cwd()
  // TODO: check if it makes sense to avoid a copy and directly reference the binary from the cache or map into container
  taskLib.cp(piperDownloadPath, path.join(cwd, binaryName))
}

export async function executePiperStep () {
  if (taskLib.getVariable('__stepActive') === 'false') {
    return Promise.resolve()
  }
  const stepName = STEP_NAME
  const defaultsFileList = (taskLib.getTaskVariable(DEFAULTS_FILE_LIST) || '').split(' ')
  const flags = taskLib.getInput('flags', false) || ''

  const stageName = getStageName()
  const customConfig = getCustomConfigLocation()
  const containerID = taskLib.getTaskVariable(CONTAINER_VARIABLE_NAME) || ''
  const binaryName = taskLib.getVariable(BINARY_NAME_VAR) || ''
  if (containerID) {
    movePiperToCWD(binaryName)

    if (process.env.SYSTEM_DEBUG) {
      console.log(taskLib.tool('env').execSync())
      console.log(
        taskLib
          .tool('docker')
          .arg('exec')
          .arg(containerID)
          .arg(['env'])
          .execSync().stdout
      )
    }

    await makeCWDWriteable()

    return taskLib
      .tool('docker')
      .arg(['exec', containerID])
      .arg([`./${binaryName}`, stepName])
      .arg(defaultsFileList)
      .argIf(customConfig, ['--customConfig', customConfig])
      .argIf(stageName, ['--stageName', stageName])
      .line(flags)
      .exec()
  } else {
    prepareVaultEnvVars()

    // Check if a different architecture for macOS is requested
    const arch = taskLib.getVariable('hyperspace.piper.enforcedOSArch') || ''
    if (arch && os.platform() == 'darwin') {
      return taskLib
        .tool('arch')
        .arg(`-${arch}`)
        .arg(binaryName)
        .arg(stepName)
        .arg(defaultsFileList)
        .argIf(stageName, ['--stageName', stageName])
        .argIf(customConfig, ['--customConfig', customConfig])
        .line(flags)
        .exec()
    }
    return taskLib
      .tool(binaryName)
      .arg(stepName)
      .arg(defaultsFileList)
      .argIf(stageName, ['--stageName', stageName])
      .argIf(customConfig, ['--customConfig', customConfig])
      .line(flags)
      .exec()
  }
}

export async function executePiper (stepName: string, flags?: string[]): Promise<void> {
  const piperBinaryName = taskLib.getVariable(BINARY_NAME_VAR)
  if (piperBinaryName === undefined) {
    await Promise.reject(new Error('couldn\'t execute Piper: no binary available'))
    return
  }

  let exec = taskLib
    .tool(piperBinaryName)
    .arg(stepName)

  if (flags !== undefined) {
    exec = exec.arg(flags)
  }

  exec.exec()
}

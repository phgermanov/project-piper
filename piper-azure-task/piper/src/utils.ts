'use strict'

import { Octokit } from '@octokit/rest'
import { retry } from '@octokit/plugin-retry'
import { OctokitOptions } from '@octokit/core/dist-types/types'
import * as taskLib from 'azure-pipelines-task-lib'
import { v4 as uuidv4 } from 'uuid'
import fs from 'fs'
import path from 'node:path'
import { STEP_NAME } from './globals'

export const msgWarnMasterBinary = 'Using _master binaries is no longer supported, fetching latest release instead.'

export const encode = (str: string): string =>
  Buffer.from(str, 'binary').toString('base64')

export const decode = (str: string): string =>
  Buffer.from(str, 'base64').toString('binary')

function wait (delay: number) {
  return new Promise((resolve) => setTimeout(resolve, delay))
}

export function fetchRetry (url: string, tries: number, reqOpts?: any): Promise<Response> {
  function onError (err: Error) {
    const delay = 1000
    const triesLeft = tries - 1
    console.log(`Error during fetch: ${err}. Retrying ${triesLeft} more time(s)...`)
    if (!triesLeft) {
      throw err
    }
    return wait(delay).then(() => fetchRetry(url, triesLeft))
  }

  return fetch(url, reqOpts).catch((err) => onError(err))
}

export function isInternalStep (): boolean {
  return STEP_NAME.startsWith('sap')
}

export function retryableOctokitREST (opts: OctokitOptions): Octokit {
  const RetryOctokit = Octokit.plugin(retry)
  return new RetryOctokit(opts)
}

export async function makeCWDWriteable () {
  const contents = 'umask 000\nchmod --silent --recursive +w ' + taskLib.cwd()

  // borrowed from https://github.com/microsoft/azure-pipelines-tasks/blob/master/Tasks/BashV3/bash.ts
  // Write the script to disk.
  taskLib.assertAgent('2.115.0')
  const tempDirectory = taskLib.getVariable('agent.tempDirectory') || ''
  taskLib.checkPath(tempDirectory, `${tempDirectory} (agent.tempDirectory)`)
  const fileName = uuidv4() + '.sh'
  const filePath = path.join(tempDirectory, fileName)
  fs.writeFileSync(filePath, contents, { encoding: 'utf8' })

  const bash = taskLib.tool('bash')
  bash.arg(filePath)

  await bash.exec({ ignoreReturnCode: true })
}

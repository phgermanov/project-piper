import * as taskLib from 'azure-pipelines-task-lib'
import { msgWarnMasterBinary, retryableOctokitREST } from './utils'
import {
  LATEST,
  OS_BINARY_NAME,
  OS_BINARY_NAME_VAR,
  OS_PIPER_OWNER,
  OS_PIPER_REPOSITORY,
  OS_PIPER_VERSION,
  pickVersionInput, setOsPiperVersion,
} from './globals'
import { OctokitOptions } from '@octokit/core/dist-types/types'
import * as toolLib from 'azure-pipelines-tool-lib'
import path from 'node:path'
import fs, { chmodSync, writeFileSync } from 'fs'
import { v4 as uuidv4 } from 'uuid'

import { getPlatform } from './github'

export async function fetchOsBinaryVersionTag () {
  taskLib.debug('fetching version of OS binary')

  let version = pickVersionInput('hyperspace.piper.version')
  if (['main', 'master'].includes(version)) {
    version = LATEST
    taskLib.warning(msgWarnMasterBinary)
  }

  if (version === LATEST) {
    taskLib.debug(`resolving latest version tag for '${OS_BINARY_NAME}' binary`)
    version = await fetchLatestOSVersionTag()
  }

  taskLib.setVariable('osBinaryVersion', version) // for usage withing stage
  taskLib.setVariable('osBinaryVersion', version, false, true) // for passing between stages

  // to avoid switching to legacy mode. Should be removed with all other legacy code
  setOsPiperVersion(version)
}

export async function prepareOsPiperBinary () {
  taskLib.debug('preparing OS binary')

  let binaryName = OS_BINARY_NAME
  if (taskLib.getTaskVariable('runningInLegacyMode') === 'true') {
    binaryName = 'piper'
  }

  taskLib.setVariable(OS_BINARY_NAME_VAR, binaryName)

  // at that point osBinaryVersion will contain only exact version tag or branch name (for development tests).
  // No 'latest' or 'master/main'
  const version = OS_PIPER_VERSION

  // First look in tool cache
  taskLib.debug(`looking for ${binaryName} version '${version}' in tool cache`)
  const toolPath = toolLib.findLocalTool(binaryName, version) || ''
  if (toolPath) {
    taskLib.debug(`binary ${binaryName} version ${version} found in tool cache, path: '${toolPath}'`)
    toolLib.prependPath(toolPath)
    return Promise.resolve()
  }

  // Build from source in specified branch
  if (taskLib.getVariable('hyperspace.piper.isBranch') || '') {
    const branchName = version
    console.log(`Piper branch ${branchName} specified`)
    const branchBinaryName = `${binaryName}-${branchName.replace('/', '-')}`
    const workspaceDir = taskLib.getVariable('Pipeline.Workspace') || ''
    const branchCachedBinaryPath = path.join(workspaceDir, OS_PIPER_REPOSITORY, 'binaries', branchBinaryName)

    // check job cache
    let toolPath = toolLib.findLocalTool(branchBinaryName, '1.0.0')
    if (!toolPath && !fs.existsSync(branchCachedBinaryPath)) {
      await buildOsBinaryFromBranch(branchName, branchBinaryName, branchCachedBinaryPath, workspaceDir)
        .catch(err => Promise.reject(new Error(`Could not build Piper from branch: ${err}`)))
    }

    // cache (within job only) and add to PATH
    // TODO: use correct version to cache
    taskLib.debug(`adding '${branchBinaryName}' from branch '${version}' to tool cache`)
    toolPath = await toolLib.cacheFile(branchCachedBinaryPath, branchBinaryName, branchBinaryName, '1.0.0')
      .catch(err => Promise.reject(new Error(`Could not cache file: ${err}`)))
    toolLib.prependPath(toolPath)

    taskLib.setVariable(OS_BINARY_NAME_VAR, branchBinaryName)
    return Promise.resolve()
  }

  // same purpose as binaryNameInAssets in preparePiperBinary().
  // Currently, OS binaries are published with name 'piper'
  const binaryNameInAssets = 'piper'
  const binaryTmpPath = await downloadOsBinary(version, binaryNameInAssets)

  taskLib.debug(`adding ${binaryName} version ${version} to tool cache`)
  const downloadedToolPath = await toolLib.cacheFile(binaryTmpPath, binaryName, binaryName, version)
  toolLib.prependPath(downloadedToolPath)
}

// fetchLatestVersion will fetch the latest release from GitHub releases and returns tag_name field
export async function fetchLatestOSVersionTag (): Promise<string> {
  const options: OctokitOptions = { baseUrl: 'https://github.com' }
  const octokit = retryableOctokitREST(options)
  const response = await octokit.request(`GET /${OS_PIPER_OWNER}/${OS_PIPER_REPOSITORY}/releases/latest`)
    .catch((error: any) => {
      return Promise.reject(new Error(`Getting latest OS release failed. Status: ${error.response.status}. Message: ${error.response.data.message}`))
    })

  const redirectURL = response.url
  const tag = redirectURL.split('/').pop() || ''
  if (tag === '') {
    return Promise.reject(new Error(`Could not get latest release tag from redirect url ${redirectURL}`))
  }

  return Promise.resolve(tag)
}

async function buildOsBinaryFromBranch (branchName: string, branchBinaryName: string, branchCachedBinaryPath: string, workspaceDir: string): Promise<void> {
  // build from source. Requires GoTool@0 task to be present in job
  // https://learn.microsoft.com/en-us/azure/devops/pipelines/tasks/reference/go-tool-v0?view=azure-pipelines
  const repoUrl = `https://github.com/${OS_PIPER_OWNER}/${OS_PIPER_REPOSITORY}.git`
  const sourceCodePath = path.join(workspaceDir, OS_PIPER_REPOSITORY)
  taskLib.debug(`Cloning '${branchName}' branch from ${repoUrl} into ${workspaceDir}`)
  const cloneRes = await taskLib.tool('git').arg(['clone', '-b', branchName, '--depth', '1', repoUrl, sourceCodePath]).exec()
  if (cloneRes !== 0) {
    await Promise.reject(new Error(`clone failed`))
  }

  const gitCommit = taskLib.tool('git').arg(['rev-parse', 'HEAD']).execSync({ cwd: sourceCodePath }).stdout

  taskLib.debug(`Building ${branchBinaryName}`)
  taskLib.tool('go').arg(['build', '-o', branchBinaryName])
    .arg(['-ldflags', `-X github.com/SAP/jenkins-library/cmd.GitCommit=${gitCommit}`])
    .execSync({ cwd: sourceCodePath })

  // the golangBuild step appends the OS and architecture of the environment to the output filename
  // remove that part and move binary to folder that can be used by a pipeline caching step
  // https://docs.microsoft.com/en-us/azure/devops/pipelines/release/caching
  const cacheDir = branchCachedBinaryPath.substring(0, branchCachedBinaryPath.lastIndexOf('/'))
  const builtBinaryPath = path.join(workspaceDir, OS_PIPER_REPOSITORY, branchBinaryName)
  if (!fs.existsSync(cacheDir)) {
    taskLib.debug(`Creating branch binary cache dir ${cacheDir}`)
    fs.mkdirSync(cacheDir)
  }
  taskLib.debug(`Renaming ${builtBinaryPath} to ${branchCachedBinaryPath}`)
  fs.renameSync(builtBinaryPath, branchCachedBinaryPath)
  return Promise.resolve()
}

async function downloadOsBinary (version: string, binaryNameInAssets: string): Promise<string> {
  let assetName = binaryNameInAssets
  const platform = getPlatform()
  if (platform !== '') {
    assetName += `-${platform}`
  }

  taskLib.debug(`downloading '${assetName}' version '${version}' from ${OS_PIPER_OWNER}/${OS_PIPER_REPOSITORY} without API call`)
  const options: OctokitOptions = { baseUrl: 'https://github.com' }
  const octokit = retryableOctokitREST(options)
  const response = await octokit.request(`GET /${OS_PIPER_OWNER}/${OS_PIPER_REPOSITORY}/releases/download/${version}/${assetName}`)
    .catch((error: any) => {
      return Promise.reject(new Error(`Could not download OS binary. Status: ${error.response.status}. Message: ${error.response.data}`))
    })

  // write to temporary file with uuid name
  const tempDir = taskLib.getVariable('Agent.TempDirectory') || ''
  const destinationPath = path.join(tempDir, uuidv4())
  writeFileSync(destinationPath, Buffer.from(response.data as unknown as ArrayBuffer))
  taskLib.debug(`github asset written to ${destinationPath}`)
  chmodSync(destinationPath, '777')
  return Promise.resolve(destinationPath)
}

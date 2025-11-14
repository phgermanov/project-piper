import * as taskLib from 'azure-pipelines-task-lib'
import { msgWarnMasterBinary } from './utils'
import * as toolLib from 'azure-pipelines-tool-lib'
import {
  BINARY_NAME,
  BINARY_NAME_VAR,
  GITHUB_TOOLS_API_URL,
  GITHUB_TOOLS_TOKEN,
  LATEST,
  OS_BINARY_NAME_VAR,
  pickVersionInput,
  PIPER_OWNER,
  PIPER_REPOSITORY,
  PIPER_VERSION,
  setPiperVersion,
  STEP_NAME
} from './globals'
import { downloadBinaryViaAPI, fetchLatestVersionTag } from './github'
import { fetchOsBinaryVersionTag, prepareOsPiperBinary } from './osBinary'

// fetchBinaryVersionTagWrapper is a temporary wrapper and should be removed after full migration to single binary usage
// and replaced with fetchBinaryVersionTag()
export async function fetchBinaryVersionTagWrapper () {
  await fetchBinaryVersionTag()
  await fetchOsBinaryVersionTag()
}

// preparePiperBinaryWrapper is a temporary wrapper and should be removed after full migration to single binary usage
// and replaced with preparePiperBinary()
export async function preparePiperBinaryWrapper () {
  await preparePiperBinary()
  await prepareOsPiperBinary()
}

// pickBinary is a temporary function that determines which binary to use when calling a step.
// It invokes the checkStep command built into the internal binary. If this command exits with code 0,
// it indicates that the step exists in the internal binary and will be used to call the step.
export async function pickBinary (): Promise<void> {
  const innerSourceBinary = taskLib.getVariable(BINARY_NAME_VAR) || ''
  if (innerSourceBinary === '') {
    return Promise.reject(new Error('Binary is not available'))
  }

  const openSourceBinary = taskLib.getVariable(OS_BINARY_NAME_VAR) || ''
  if (openSourceBinary === '') {
    return Promise.reject(new Error('OS binary is not available'))
  }

  // execute ./piper <stepName> --help to check if step exists in inner source binary. It fails if step is not there
  // and exits with 0 code if it is.
  const stepName = STEP_NAME
  const res = taskLib.tool(innerSourceBinary).arg(stepName).arg('--help').execSync({ silent: true })
  if (res.code === 0) {
    taskLib.debug(`step ${stepName} is in inner source binary.`)
    return Promise.resolve()
  }

  taskLib.debug(`step ${stepName} doesn't exist in inner source binary. Switching to open source.`)
  taskLib.setVariable(BINARY_NAME_VAR, openSourceBinary)
}

// fetchBinaryVersionTag will fetch version tag from binary releases and sets in Azure variable for later use
// by subsequent steps, like for caching binary
// If version is not latest, it will simply set the variable to that version
export async function fetchBinaryVersionTag () {
  taskLib.debug('fetching version of binary')

  let version = pickVersionInput('hyperspace.sappiper.version')
  if (['main', 'master'].includes(version)) {
    version = LATEST
    taskLib.warning(msgWarnMasterBinary)
  }

  if (version === LATEST) {
    taskLib.debug(`resolving latest version tag for '${BINARY_NAME}' binary`)
    version = await fetchLatestVersionTag(PIPER_OWNER, PIPER_REPOSITORY, GITHUB_TOOLS_TOKEN)
  }

  taskLib.setVariable('binaryVersion', version)
  taskLib.setVariable('binaryVersion', version, false, true)

  // to avoid switching to legacy mode. Should be removed with all other legacy code
  setPiperVersion(version)
}

export async function preparePiperBinary () {
  taskLib.debug('preparing binary')

  let binaryName = BINARY_NAME
  if (taskLib.getTaskVariable('runningInLegacyMode') === 'true') {
    binaryName = 'sap-piper'
  }

  taskLib.setVariable(BINARY_NAME_VAR, binaryName)

  // at that point PIPER_VERSION will contain only exact version tag or branch name (for development tests).
  // No 'latest' or 'master/main'
  const version = PIPER_VERSION

  taskLib.debug(`looking for ${binaryName} version '${version}' in tool cache`)
  const toolPath = toolLib.findLocalTool(binaryName, version)
  if (toolPath) {
    taskLib.debug(`binary ${binaryName} version ${version} found in tool cache, path: '${toolPath}'`)
    toolLib.prependPath(toolPath)
    return
  }

  // TODO: implement building from source for testing purposes. Must be here (after cache check and before release download)

  // binaryNameInAssets constant is a temporary switch introduced for compatibility reasons.
  // Currently, our internal binaries are published under the name 'sap-piper'. Once the OS binary is fully deprecated,
  // the single binary should be published under the name 'piper', replacing 'sap-piper'.
  // This will mean that the binaryNameInAssets should be removed and the BINARY_NAME variable should be used instead.
  const binaryNameInAssets = 'sap-piper'
  const binaryTmpPath = await downloadBinaryViaAPI(version, binaryNameInAssets, GITHUB_TOOLS_API_URL, PIPER_OWNER, PIPER_REPOSITORY, GITHUB_TOOLS_TOKEN)

  taskLib.debug(`adding ${binaryName} version ${version} to tool cache`)
  const downloadedToolPath = await toolLib.cacheFile(binaryTmpPath, binaryName, binaryName, version)
  toolLib.prependPath(downloadedToolPath)
}

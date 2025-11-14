import * as taskLib from 'azure-pipelines-task-lib'

// TODO: move other globals here as well

export let STEP_NAME = ''

export let GITHUB_TOOLS_TOKEN = ''

export const BINARY_NAME = 'piper'
// PIPER_VERSION holds version value that is used to decide which version of binary to use.
// By default, comes from cached variable 'binaryVersion' or from sapPiperVersion input parameter, if set.
// Piper task input parameter overwrites the cached variable.
export let PIPER_VERSION = ''
export let PIPER_OWNER = ''
export let PIPER_REPOSITORY = ''

export const OS_BINARY_NAME = 'osPiper'
export let OS_PIPER_VERSION = ''  // same as PIPER_VERSION, but for open source binary
export const OS_PIPER_OWNER = 'SAP'
export const OS_PIPER_REPOSITORY = 'jenkins-library'

export const LATEST = 'latest'
export const GITHUB_TOOLS_API_URL = 'https://github.tools.sap/api/v3'

export const BINARY_NAME_VAR = '__binaryName'
export const OS_BINARY_NAME_VAR = '__osBinaryName'
const githubToolsConnectionInputName = 'gitHubConnection'

// initializeGlobals ensures that global variables are initialized at runtime, rather than during module imports.
export function initializeGlobals () {
  STEP_NAME = taskLib.getInput('stepName', true)!

  GITHUB_TOOLS_TOKEN = getGithubConnectionToken(githubToolsConnectionInputName)

  // TODO(reminder): after fully deprecating open source repository and completely moving to single binary
  //  we should use 'hyperspace.piper.**' for setting variables and remove/deprecate hyperspace.sappiper.**
  //  Also, it will be a breaking change, so needs an announcement.
  //  same goes about sapPiperVersion and piperVersion task inputs.
  PIPER_VERSION = pickVersionToUse('sapPiperVersion', 'binaryVersion')
  PIPER_OWNER = taskLib.getVariable('hyperspace.sappiper.owner') || 'project-piper'
  PIPER_REPOSITORY = taskLib.getVariable('hyperspace.sappiper.repository') || 'sap-piper'

  OS_PIPER_VERSION = pickVersionToUse('piperVersion', 'osBinaryVersion')
}

export function getGithubConnectionToken (inputName: string): string {
  let serviceConnName = taskLib.getInput(inputName, false) || ''
  if (serviceConnName === '') {
    taskLib.debug('no github service connection provided')
    serviceConnName = 'github.tools.sap'  // Hyperspace created pipelines has this service connection set up by default
  }

  const pat = taskLib.getEndpointAuthorizationParameter(serviceConnName, 'apitoken', true) || ''
  const oauth = taskLib.getEndpointAuthorizationParameter(serviceConnName, 'AccessToken', true) || ''
  // Use PAT or OAuth token
  const token = pat === '' ? oauth : pat

  if (token === '') {
    if (serviceConnName === 'github.tools.sap') {
      taskLib.debug('Token is empty on a Hyperspace configured service connection')
    } else {
      taskLib.error('Your GitHub service connection ' + serviceConnName + ' does not contain a personal access token or OAuth token.')
    }
    return ''
  }

  return token
}

export function pickVersionInput (varName: string): string {
  const version = taskLib.getVariable(varName) || ''
  if (version) {
    taskLib.debug(`using version '${version}' from pipeline variable`)
    return version
  }

  taskLib.debug(`using version '${LATEST}' as fallback default`)
  return LATEST
}

function pickVersionToUse (taskInputName: string, cacheVariableName: string): string {
  const versionFromInput = taskLib.getInput(taskInputName) || ''
  if (versionFromInput === LATEST) {
    console.log(`'${LATEST}' is not allowed value for '${taskInputName}' task input. Specify exact version or omit the input. Falling back to cached version`)
  } else if (versionFromInput) {
    taskLib.debug(`using version '${versionFromInput}' from task input`)
    return versionFromInput
  }

  const cachedVersion = taskLib.getVariable(cacheVariableName) || ''
  taskLib.debug(`using version '${cachedVersion}' from pipeline cache`)
  // empty cache here means that Piper Azure task is running for the first time in the pipeline,
  // and it will be resolved in handleDeprecatedCachingVariables function.

  return cachedVersion
}

// Legacy functions used by backward compatibility logic
export function setPiperVersion (version: string) {
  PIPER_VERSION = version
}

export function setOsPiperVersion (version: string) {
  OS_PIPER_VERSION = version
}

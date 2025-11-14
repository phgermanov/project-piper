// checkDeprecatedInputs checks if deprecated inputs are used and logs a warning if they are
import * as taskLib from 'azure-pipelines-task-lib'
import {
  BINARY_NAME,
  GITHUB_TOOLS_TOKEN,
  LATEST,
  OS_BINARY_NAME,
  OS_PIPER_VERSION,
  pickVersionInput,
  PIPER_OWNER,
  PIPER_REPOSITORY,
  PIPER_VERSION,
  setOsPiperVersion,
  setPiperVersion
} from './globals'
import { fetchLatestVersionTag } from './github'
import { fetchLatestOSVersionTag } from './osBinary'

// warnAboutDeprecatedInputs checks if deprecated inputs are used and logs a message.
// console.log is used intentionally to avoid annoying warning messages in the logs for pipelines
// that use these inputs. Must be removed with when switching to single binary usage and releasing a major version.
export function warnAboutDeprecatedInputs (): void {
  const var1 = taskLib.getVariable('hyperspace.sappiper.binaryname')
  if (var1) {
    console.log('The variable hyperspace.sappiper.binaryname is no longer used in this Azure task. Please remove it from your pipeline')
  }

  const input1 = taskLib.getInput('gitHubComConnection')
  if (input1) {
    console.log('The input gitHubComConnection is no longer used in this Azure task. Please remove it from Piper task input')
  }

  const input2 = taskLib.getBoolInput('getSapPiperLatestVersion')
  if (input2) {
    console.log('The input getSapPiperLatestVersion is no longer used in this Azure task. Please remove it from Piper task input')
  }
}

// handleDeprecatedCachingVariables is a temporary function that handles deprecated caching variables and
// here only for backward compatibility. It should be removed after full migration to single binary usage.
export async function handleDeprecatedCachingVariables () {
  // By this point, if the version variable is empty, it means that the task is running in older version of pipeline.
  // So we can consider it as legacy mode.
  if (PIPER_VERSION !== '') {
    return
  }
  taskLib.debug('[legacy] running in legacy mode')

  // to signal binary switching logic to use old binary names
  taskLib.setTaskVariable('runningInLegacyMode', 'true')

  // There might be stages without stage level variable sapPiperCacheVersion, for example Init stage. Custom pipelines could have it too.
  // That's why we need to refer to variable sapPiperVersion exported by sap_piper_cache_key named step
  const legacyCachedVersion = taskLib.getVariable('sapPiperCacheVersion') || taskLib.getVariable('sap_piper_cache_key.sapPiperVersion') || ''
  if (legacyCachedVersion !== '') {
    taskLib.debug(`[legacy] using ${legacyCachedVersion} from cached variable`)
    setPiperVersion(legacyCachedVersion)
    return
  }

  // There is no cached version -> resolve and cache it.
  const versionInput = pickVersionInput('hyperspace.sappiper.version')
  let piperVersion = ['main', 'master'].includes(versionInput) ? LATEST : versionInput
  taskLib.debug('[legacy] caching SAP binary version as \'sapPiperVersion\' variable')

  if (piperVersion === LATEST) {
    taskLib.debug(`[legacy] resolving latest version tag for '${BINARY_NAME}' binary`)
    piperVersion = await fetchLatestVersionTag(PIPER_OWNER, PIPER_REPOSITORY, GITHUB_TOOLS_TOKEN)
  }

  // `sapPiperVersion` is kept for backward compatibility. It is exported to ensure that pipelines,
  // which still reference `sapPiperVersion` in custom stages, continue to function correctly.
  taskLib.setVariable('sapPiperVersion', piperVersion, false, true)
  taskLib.setVariable('sapPiperLatestVersion', piperVersion, false, true)  // it was set when getSapPiperLatestVersion=true
  setPiperVersion(piperVersion)
}

// handleDeprecatedCachingVariablesOS is a temporary function that handles deprecated caching variables and
// here only for backward compatibility. It should be removed after full migration to single binary usage.
export async function handleDeprecatedCachingVariablesOS () {
  // By this point, if the version variable is empty, it means that the task is running in older version of pipeline.
  // So we can consider it as legacy mode.
  if (OS_PIPER_VERSION !== '') {
    return
  }
  taskLib.debug('[legacy] running in legacy mode')

  // to signal binary switching logic to use old binary names
  taskLib.setTaskVariable('runningInLegacyMode', 'true')

  // There might be stages without stage level variable piperCacheVersion, for example Init stage. Custom pipelines could have it too.
  // That's why we need to refer to variable piperVersion exported by piper_cache_key named step
  const legacyCachedVersion = taskLib.getVariable('piperCacheVersion') || taskLib.getVariable('piper_cache_key.piperVersion') || ''
  if (legacyCachedVersion !== '') {
    taskLib.debug(`[legacy] using ${legacyCachedVersion} from piper_cache_key.piperVersion variable`)
    setOsPiperVersion(legacyCachedVersion)
    return
  }

  // There is no cached version -> resolve and cache it.

  const osVersionInput = pickVersionInput('hyperspace.piper.version')
  let osPiperVersion = ['main', 'master'].includes(osVersionInput) ? LATEST : osVersionInput
  taskLib.debug('[legacy] caching OS binary version as \'piperVersion\' variable')

  if (osPiperVersion === LATEST) {
    taskLib.debug(`[legacy] resolving latest version OS tag for '${OS_BINARY_NAME}' binary`)
    osPiperVersion = await fetchLatestOSVersionTag()
  }

  // `piperVersion` is kept for backward compatibility. It is exported to ensure that pipelines,
  // which still reference `piperVersion` in custom stages, continue to function correctly.
  taskLib.setVariable('piperVersion', osPiperVersion, false, true)
  setOsPiperVersion(osPiperVersion)
}

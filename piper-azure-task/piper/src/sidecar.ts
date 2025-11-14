import * as taskLib from 'azure-pipelines-task-lib'
import { getAzureEnvVars, getProxyEnvVars, getVaultEnvVars, getVolumes} from './docker'
import { v4 as uuidv4 } from 'uuid'

export const SIDECAR_VARIABLE_NAME = '__sidecarContainerId'

export async function startSidecar (sidecarImage: string, config: any) {
  await createNetwork()
  taskLib.debug(`Starting sidecar container with image ${sidecarImage}`)
  await startContainer(sidecarImage, config)
}

async function startContainer (containerImage: string, config: any) {
  const containerOptions =
        taskLib.getInput('sidecarOptions', false) ||
        config.sidecarOptions?.join(' ') ||
        ''
  const containerID = uuidv4()
  const networkName = taskLib.getVariable(NETWORK_VARIABLE_NAME) || ''
  const networkAlias = config.sidecarName || ''

  taskLib.setTaskVariable(SIDECAR_VARIABLE_NAME, containerID)
  console.log(`Starting image ${containerImage} as sidecar ${containerID}`)
  const result = taskLib
    .tool('docker')
    .arg('run')
    .arg('--detach')
    .arg('--rm')
    .line(containerOptions)
    .arg(['--name', containerID])
    .argIf(networkName, ['--network', networkName])
    .argIf(networkName && networkAlias, ['--network-alias', networkAlias])
    .arg(getSidecarEnvVars(config))
    .arg(getProxyEnvVars())
    .arg(getVaultEnvVars())
    .arg(getAzureEnvVars())
    .arg(getVolumes(config.sidecarVolumeBind))
    .line(containerImage)
    .execSync()

  if (result.error) {
    taskLib.error('failed to start container: ' + result.error.message)
  } else {
    taskLib.debug('container started')
  }
}

export async function stopSidecar () {
  const containerID = taskLib.getTaskVariable(SIDECAR_VARIABLE_NAME) || ''
  if (!containerID) {
    taskLib.debug('no sidecar to stop')
    return
  }

  console.log(`Stopping sidecar ${containerID}`)
  const result = taskLib
    .tool('docker')
    .arg('stop')
    .arg('--time=1')
    .arg(containerID)
    .execSync()

  if (result.error) {
    taskLib.error('failed to stop container: ' + result.error.message)
  } else {
    taskLib.debug('container stopped')
  }

  await removeNetwork()
}

const NETWORK_PREFIX = 'sidecar-'
export const NETWORK_VARIABLE_NAME = '__sidecarNetworkId'

export async function removeNetwork () {
  const networkName = taskLib.getVariable(NETWORK_VARIABLE_NAME) || ''
  if (!networkName) {
    taskLib.debug('no network to remove')
    return
  }

  taskLib.setVariable(NETWORK_VARIABLE_NAME, '')
  console.log(`Removing network ${networkName}`)
  const result = taskLib
    .tool('docker')
    .arg('network')
    .arg('remove')
    .arg(networkName)
    .execSync()

  if (result.error) {
    taskLib.error('failed to remove network: ' + result.error.message)
  } else {
    taskLib.debug('network removed')
  }
}

export async function createNetwork () {
  const networkName = NETWORK_PREFIX + uuidv4()

  console.log(`Creating network ${networkName}`)
  const result = taskLib
    .tool('docker')
    .arg('network')
    .arg('create')
    .arg(networkName)
    .execSync()

  if (result.error) {
    taskLib.error('failed to create network: ' + result.error.message)
  } else {
    taskLib.debug('network created')
    taskLib.setVariable(NETWORK_VARIABLE_NAME, networkName)
  }
}

export function getSidecarEnvVars (config: any) {
  const result: string[] = []
  let sidecarEnvVars = taskLib.getInput('sidecarEnvVars', false) || config.sidecarEnvVars || {}

  if (typeof sidecarEnvVars === 'string') {
    try {
      sidecarEnvVars = JSON.parse(sidecarEnvVars)
    } catch (err) {
      console.log(`sidecarEnvVars value ${sidecarEnvVars} is not a JSON-formatted string, therefore ignore it`)
      sidecarEnvVars = {}
    }
  }

  Object.entries(sidecarEnvVars)
    .forEach(([key, value]) => {
      result.push('--env')
      result.push(`${key}=${value}`)
    })

  return result
}

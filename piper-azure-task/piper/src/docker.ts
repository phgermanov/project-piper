import * as taskLib from 'azure-pipelines-task-lib/task'
import { prepareVaultEnvVars } from './vault'
import { NETWORK_VARIABLE_NAME, startSidecar, stopSidecar } from './sidecar'
import { v4 as uuidv4 } from 'uuid'
import { fetchRetry } from './utils'

export const CONTAINER_VARIABLE_NAME = '__containerID'
export const COMMON_REPO_HOST = 'docker-hub.common.repositories.cloud.sap'

export async function stopContainers () {
  if (taskLib.getVariable('__stepActive') === 'false') {
    return Promise.resolve()
  }

  stopContainer()
  await stopSidecar()
}

function stopContainer () {
  const containerID = taskLib.getTaskVariable(CONTAINER_VARIABLE_NAME) || ''
  if (!containerID) {
    taskLib.debug('no container to stop')
    return
  }

  console.log(`Stopping container ${containerID}`)
  const result = taskLib
    .tool('docker')
    .arg('stop')
    .arg('--time=1')
    .arg(containerID)
    .execSync()

  if (result.code != 0) {
    taskLib.debug(`could not stop container with ID ${containerID}!`)
    taskLib.debug(result.stderr)
  }
}

export async function startContainers (config: any) {
  if (taskLib.getVariable('__stepActive') === 'false') {
    return Promise.resolve()
  }

  const dockerImage = taskLib.getInput('dockerImage', false) || config.dockerImage || ''
  taskLib.debug('dockerImage: [' + dockerImage + ']')
  if (dockerImage === '' || dockerImage === 'none') {
    return
  }

  const sidecarImage = taskLib.getInput('sidecarImage', false) || config.sidecarImage || ''
  if (sidecarImage !== '') {
    const pulledImage = await handleImagePull(sidecarImage)
    await startSidecar(pulledImage, config)
  }

  const pulledImage = await handleImagePull(dockerImage)
  taskLib.debug(`Starting container with image ${pulledImage}`)
  await startContainer(pulledImage, config)
}

async function handleImagePull (image: string): Promise<string> {
  const [registry, imageNameAndTag] = parseImageName(image)
  const imageToPull = await handleAuthentication(registry, imageNameAndTag)

  const initialPullSuccess = pullWithRetry(imageToPull)
  if (initialPullSuccess) {
    return imageToPull
  }

  if (registry !== COMMON_REPO_HOST) {
    return Promise.reject(`Failed to pull image ${imageToPull}`)
  }

  taskLib.debug('Falling back to pull from docker hub')
  const dockerHubPullSuccess = pullWithRetry(imageNameAndTag)
  if (dockerHubPullSuccess) {
    return imageNameAndTag
  }

  return Promise.reject(`Failed to pull image ${imageNameAndTag} from docker hub`)
}

async function handleAuthentication (registry: string, imageNameAndTag: string): Promise<string> {
  const imageToPull = registry === '' ? imageNameAndTag : `${registry}/${imageNameAndTag}`

  // Service connection authentication
  const registryConnectionID = taskLib.getInput('dockerRegistryConnection', false) || ''
  if (registryConnectionID !== '') {
    taskLib.debug('Authenticating with credentials from dockerRegistryConnection input')
    const registryCreds = taskLib.getEndpointAuthorization(registryConnectionID, true)
    if (registryCreds === undefined) {
      taskLib.debug(`Service connection ${registryConnectionID} configuration invalid, proceeding without docker authentication`)
      return imageToPull
    }

    const loginExitCode = dockerLogin(registry, registryCreds.parameters.username, registryCreds.parameters.password)
    if (loginExitCode !== 0) {
      taskLib.debug('Service connection authentication failed, proceeding without docker authentication')
      return imageToPull
    }

    return imageToPull
  }

  // System Trust authentication
  if (registry === COMMON_REPO_HOST) {
    taskLib.debug('Running in a MS-hosted agent and image is from docker hub')
    const [user, password] = await getArtifactoryCredentialsFromSystemTrust()
    if (user === '' || password === '') {
      taskLib.debug('Artifactory credentials are empty, proceeding without docker authentication')
      return imageNameAndTag
    }

    const loginExitCode = dockerLogin(registry, user, password)
    if (loginExitCode !== 0) {
      taskLib.debug('System trust authentication failed, proceeding without docker authentication')
      return imageNameAndTag
    }

    taskLib.debug("docker authentication with System Trust successful")
    return imageToPull
  }

  taskLib.debug('Proceeding without docker authentication')
  return imageToPull
}

async function getArtifactoryCredentialsFromSystemTrust (): Promise<[string, string]> {
  const systemTrustToken = process.env.PIPER_systemTrustToken || ''
  if (systemTrustToken === '') {
    taskLib.debug('System trust token is not set')
    return ['', '']
  }

  let systemTrustURL = taskLib.getVariable('systemTrustURL') || ''
  if (systemTrustURL === '') {
    taskLib.debug('systemTrustURL variable is empty. Using default value')
    systemTrustURL = 'https://api.trust.tools.sap'
  }
  const response = await fetchRetry(`${systemTrustURL}/tokens?system=artifactory`, 3, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${systemTrustToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify([{ system: 'artifactory', scope: 'pipeline' }])
  })
  if (!response.ok) {
    taskLib.debug(`Artifactory token request failed with status ${response.status}: ${await response.text()}`)
    return ['', '']
  }

  const data = await response.json()
  const artifactoryToken = data.artifactory
  const payload = artifactoryToken.split('.')[1]
  const decodedPayload = Buffer.from(payload, 'base64').toString('utf-8')
  const user = JSON.parse(decodedPayload).sub
  return [user, artifactoryToken]
}

function pullWithRetry (image: string): boolean {
  taskLib.debug(`Pulling image ${image}`)
  let attempts = 3
  let execRes
  while (attempts > 0) {
    execRes = taskLib.tool('docker').arg('pull').arg(image).execSync()
    if (execRes.code === 0) {
      taskLib.debug(`Image ${image} pulled successfully`)
      return true
    }
    taskLib.debug(`Failed to pull image ${image}: ${execRes.stderr}`)
    taskLib.debug(`Retrying... (${3 - attempts + 1}/3)`)
    attempts--
  }

  taskLib.debug(`Failed to pull image ${image} after ${attempts} attempts: ${execRes?.stderr}`)
  return false
}

async function startContainer (dockerImage: string, config: any) {
  let dockerOptions =
    taskLib.getInput('dockerOptions', false) ||
    config.dockerOptions ||
    ''
  if (Array.isArray(dockerOptions)) {
    dockerOptions = dockerOptions.join(' ')
  }

  const containerID = uuidv4()
  const networkName = taskLib.getVariable(NETWORK_VARIABLE_NAME) || ''
  const networkAlias = config.dockerName || ''
  taskLib.setTaskVariable(CONTAINER_VARIABLE_NAME, containerID)
  console.log(`Starting image ${dockerImage} as container ${containerID}`)
  const cwd = taskLib.cwd()
  prepareVaultEnvVars()
  return (
    taskLib
      .tool('docker')
      .arg('run')
      .arg('--tty')
      .arg('--detach')
      .arg('--rm')
      // TODO: # on azure we need to re-consider -u setting with uid, here we deal with 1001
      .arg(['--user', '1000:1000'])
      .arg(['--volume', `${cwd}:${cwd}`])
      .arg(['--workdir', cwd])
      .arg(getVolumes(config.dockerVolumeBind))
      .line(dockerOptions)
      .arg(['--name', containerID])
      .argIf(networkName, ['--network', networkName])
      .argIf(networkName && networkAlias, [
        '--network-alias',
        networkAlias
      ])
      .arg(getDockerEnvVars(config))
      .arg(getProxyEnvVars())
      .arg(getVaultEnvVars())
      .arg(getAzureEnvVars())
      .arg(getSystemTrustEnvVars())
      .arg(getTelemetryEnvVars())
      .arg(getDockerImageFromEnvVar(dockerImage))
      .line(dockerImage)
      .arg('cat')
      .exec()
  )
}

export function getProxyEnvVars () {
  return [
    '--env', 'http_proxy',
    '--env', 'https_proxy',
    '--env', 'no_proxy',
    '--env', 'HTTP_PROXY',
    '--env', 'HTTPS_PROXY',
    '--env', 'NO_PROXY'
  ]
}

export function getVaultEnvVars () {
  return [
    '--env', 'PIPER_vaultAppRoleID',
    '--env', 'PIPER_vaultAppRoleSecretID'
  ]
}

export function getAzureEnvVars () {
  return [
    '--env', 'SYSTEM_STAGENAME',
    '--env', 'SYSTEM_STAGEDISPLAYNAME',
    // needed for Piper orchestrator detection
    '--env', 'AZURE_HTTP_USER_AGENT',
    '--env', 'TF_BUILD',
    // Build Info (needed for sapGenerateEnvironmentInfo)
    '--env', 'BUILD_BUILDID',
    '--env', 'BUILD_BUILDNUMBER',
    '--env', 'BUILD_SOURCEBRANCH',
    '--env', 'BUILD_SOURCEVERSION',
    '--env', 'BUILD_REPOSITORY_URI',
    '--env', 'BUILD_REASON',
    '--env', 'SYSTEM_TEAMFOUNDATIONCOLLECTIONURI',
    '--env', 'SYSTEM_TEAMPROJECT',
    '--env', 'SYSTEM_DEFINITIONNAME',
    '--env', 'SYSTEM_DEFINITIONID',
    // Pull Request Info (needed for sonarExecuteScan)
    '--env', 'SYSTEM_PULLREQUEST_SOURCEBRANCH',
    '--env', 'SYSTEM_PULLREQUEST_TARGETBRANCH',
    '--env', 'SYSTEM_PULLREQUEST_PULLREQUESTNUMBER',
    '--env', 'SYSTEM_PULLREQUEST_PULLREQUESTID'
  ]
}

export function getSystemTrustEnvVars (): string[] {
  return [
    '--env', 'PIPER_systemTrustToken'
  ]
}

export function getTelemetryEnvVars (): string[] {
  return [
    '--env', 'PIPER_PIPELINE_TEMPLATE_NAME'
  ]
}

export function getDockerImageFromEnvVar (dockerImage: string): string[] {
  return [
    '--env', `PIPER_dockerImage=${dockerImage}`
  ]
}

export function getDockerEnvVars (config: any) {
  const result: string[] = []
  let dockerEnvVars = taskLib.getInput('dockerEnvVars', false) || config.dockerEnvVars || {}

  if (typeof dockerEnvVars === 'string') {
    try {
      dockerEnvVars = JSON.parse(dockerEnvVars)
    } catch (err) {
      console.log(`dockerEnvVars value ${dockerEnvVars} is not a JSON-formatted string, therefore ignore it`)
      dockerEnvVars = {}
    }
  }

  Object.entries(dockerEnvVars)
    .forEach(([key, value]) => {
      result.push('--env')
      result.push(`${key}=${value}`)
    })

  return result
}

export const testExports = {
  startContainer,
  stopContainer,
  parseImageName
}

export function getVolumes (volumeBinds: any) {
  const dockerVolumeBinds: string[] = []

  if (volumeBinds != null && (['Map', 'Object', 'Set'].includes(volumeBinds.constructor.name))) {
    Object.entries<string>(volumeBinds).forEach(([name, path]) => {
      dockerVolumeBinds.push('--volume', `${name}:${path}`)
    })
  }

  return dockerVolumeBinds
}

// Regex that captures registry name and image+tag/digest
// Test this regex here: https://regex101.com/r/PizBla/3 or check unit test cases for function
const dockerImageRegex = /^(?:([\w-]+(?:\.[\w-]+)+(?::\d+)?|\w+:\d+)\/)?(.*)$/

// parseImageName parses the incoming image name to get the registry and image name (without the registry part).
// If the image is from Docker Hub and the step is running on a Microsoft-hosted agent,
// it returns the common repository host and the image name. Otherwise, it returns the registry and image name.
function parseImageName (image: string): [string, string] {
  const match = image.match(dockerImageRegex)
  if (!match) {
    throw new Error(`Image name ${image} is not a valid docker image name`)
  }

  let [, registry, imageWithTag] = match
  if (registry) {
    taskLib.debug(`Image ${image} is not from docker hub`)
    return [registry, imageWithTag]
  }

  if (onMsHostedAgent()) {
    taskLib.debug(`Image ${image} is from docker-hub and step is running on a MS-hosted agent`)
    return [COMMON_REPO_HOST, imageWithTag]
  }

  taskLib.debug(`Image ${image} is from docker-hub and step is running on a self-hosted agent`)
  return ['', imageWithTag]
}

function dockerLogin (registry: string, user: string, password: string): number {
  taskLib.setSecret(password)
  const execRes = taskLib.tool('docker')
    .arg('login')
    .arg(['--username', user])
    .arg(['--password', password]) // password value is masked in logs
    .arg(registry)
    .execSync()
  return execRes.code
}

// To identify if agent is one of the MS Hosted ones from here
// https://dev.azure.com/hyperspace-pipelines/_settings/agentpools?poolId=9&view=agents
function onMsHostedAgent (): boolean {
  const agentName = taskLib.getVariable('Agent.Name') || ''
  return agentName.startsWith("Azure Pipelines") || agentName === 'Hosted Agent'
}

import * as taskLib from 'azure-pipelines-task-lib/task'
import {
  checkStageConditions,
  decodeConfig,
  encodeConfig,
  getCustomConfigLocation,
  getStageName,
  publishConfig,
  readConfig,
  writeConfig
} from './configHelper'
import { getReleaseAsset } from './github'
import { writeFileSync } from 'fs'
import { join } from 'path'
import { executePiper } from './piper'
import {
  BINARY_NAME_VAR,
  GITHUB_TOOLS_API_URL,
  GITHUB_TOOLS_TOKEN,
  PIPER_OWNER,
  PIPER_REPOSITORY,
  PIPER_VERSION,
  STEP_NAME,
} from './globals'

export const DEFAULTS_FILE_LIST = '__defaultsFileList'
const STAGE_CONDITIONS_FILE_LIST = '__stageConditionsFileList'

// corresponding parameter names in the general purpose pipeline
const defaultConfigParamName = 'DefaultConfig'
const stageConditionsParamName = 'StageConditions'

const SAP_DEFAULT_CONFIG_HOST = 'github.tools.sap'
const SAP_STAGE_CONDITIONS_FILE = 'piper-stage-config.yml'
const SAP_DEFAULTS_FILE_URL = `https://${SAP_DEFAULT_CONFIG_HOST}/api/v3/repos/project-piper/sap-piper/contents/resources/piper-defaults.yml`
const SAP_STAGE_CONDITIONS_URL = `https://${SAP_DEFAULT_CONFIG_HOST}/api/v3/repos/project-piper/sap-piper/contents/resources/${SAP_STAGE_CONDITIONS_FILE}`
const SAP_DEFAULTS_FILE = 'piper-defaults-azure.yml'

export async function readContextConfig() {
  if (taskLib.getVariable('__stepActive') === 'false') {
    return Promise.resolve({})
  }
  const stepName = STEP_NAME
  const defaultsFileList = (taskLib.getTaskVariable(DEFAULTS_FILE_LIST) || '').split(' ')
  // exit when version or help is called
  if (['version', 'help', 'getConfig', 'writePipelineEnv'].includes(stepName)) {
    return Promise.resolve({})
  }

  const binaryName = taskLib.getVariable(BINARY_NAME_VAR) || ''
  if (binaryName) {
    taskLib.debug('reading context config')
  } else {
    return Promise.reject('piper binary not available')
  }

  const customConfig = getCustomConfigLocation()
  const stageName = getStageName()
  // read config
  const result = taskLib
    .tool(binaryName)
    .arg('getConfig')
    .arg('--contextConfig')
    .arg(defaultsFileList)
    .argIf(customConfig, ['--customConfig', customConfig])
    .argIf(stageName, ['--stageName', stageName])
    .arg(['--stepName', `${stepName}`])
    // TODO: respect "parameters"
    .execSync()
  // result handling
  taskLib.debug('RESULT:' + result.code)
  if (result.code !== 0) {
    taskLib.error(result.stderr)
    if (result.error) {
      taskLib.error('error happened while calling binary ' + result.error.name + ': ' + result.error.message)
    }
    return Promise.reject('piper binary can\'t get config')
  }
  // parse config
  taskLib.debug('Context Config: ' + JSON.stringify(result.stdout))
  return Promise.resolve(JSON.parse(result.stdout))
}

// this function is deprecated and will be removed
export async function checkIfStepActive() {
  const newCheckIfStepActive = taskLib.getVariable('activeStepsMap')
  if (newCheckIfStepActive !== undefined) {
    taskLib.debug('skipping checkIfStepActive because activeStepsMap is used in this stage')
    return
  }

  console.log(`Checking if step ${STEP_NAME} is active`)
  let stageConditions = taskLib.getInput('restorePipelineStageConditions', false)
  if (!stageConditions) {
    stageConditions = taskLib.getInput('stageConditions', false)
  }
  if (stageConditions) {
    // TODO do we need default.yaml as well?
    return checkStageConditions()
    // .catch((error) => taskLib.error("ERROR" + error));
  } else {
    console.log('YES')
    taskLib.setVariable('__stepActive', 'true')
    // return Promise.resolve(true);
  }
}

export async function createCheckIfStepActiveMaps(): Promise<void> {
  if (!taskLib.getBoolInput('createCheckIfStepActiveMaps', false)) {
    return
  }

  taskLib.debug('Creating checkIfStepActive maps')
  const stageConditionsPath = await downloadStageConditions()
  taskLib.debug('Stage conditions path: ' + stageConditionsPath)
  const defaultsFileList = (taskLib.getTaskVariable(DEFAULTS_FILE_LIST) ?? '').split(' ')

  const flags: string[] = []
  flags.push('--stageConfig', stageConditionsPath)
  flags.push('--stageOutputFile', '.pipeline/stage_out.json')
  flags.push('--stepOutputFile', '.pipeline/step_out.json')
  flags.push('--stage', '_')
  flags.push('--step', '_')
  flags.push(...defaultsFileList)

  await executePiper('checkIfStepActive', flags)
}

async function downloadStageConditions(): Promise<string> {
  const stageConditionsURL = await getStageConditionsURL()
  taskLib.debug('Stage conditions URL: ' + stageConditionsURL)
  const stageConditions = await readConfig(stageConditionsParamName, SAP_DEFAULT_CONFIG_HOST, [stageConditionsURL])
  const config = JSON.parse(stageConditions).content

  const stageConditionsPath = join('.pipeline', SAP_STAGE_CONDITIONS_FILE)
  try {
    writeFileSync(stageConditionsPath, config)
  } catch (err) {
    return await Promise.reject(new Error(`Couldn't download stage conditions: ${(err as Error).message}`))
  }
  return stageConditionsPath
}

async function getStageConditionsURL(): Promise<string> {
  const customStageConditions = taskLib.getInput('customStageConditions')
  if (customStageConditions !== undefined) {
    taskLib.debug(`using custom stage config from ${customStageConditions}`)
    return customStageConditions
  }

  const asset = await getReleaseAsset(PIPER_VERSION, SAP_STAGE_CONDITIONS_FILE, GITHUB_TOOLS_API_URL, PIPER_OWNER, PIPER_REPOSITORY, GITHUB_TOOLS_TOKEN)
  if (asset != null) {
    taskLib.debug('using stage config from release assets')
    return asset.url
  }

  taskLib.debug('using stage config from sap-piper')
  // for older releases there is no stage config file in release assets.
  return SAP_STAGE_CONDITIONS_URL
}

// this function is deprecated and will be removed
export async function preserveStageConditions() {
  const preserve = taskLib.getBoolInput('preserveStageConditions', false)
  if (preserve) {
    // taskLib.warning('There is a new way of determining which steps and stages are active. preserveStageConditions will be deprecated in a future version of piper-azure-task.' +
    //   'Please refer to https://github.tools.sap/project-piper/piper-pipeline-azure/pull/304 and ask the maintainer of your pipeline definition to incorporate the changes mentioned there.')

    let config: string
    const customStageConditions = taskLib.getInput('customStageConditions') || ''
    if (customStageConditions !== '') {
      taskLib.debug(`using custom stage config from ${customStageConditions}`)
      config = customStageConditions
    } else {
      const asset = await getReleaseAsset(PIPER_VERSION, SAP_STAGE_CONDITIONS_FILE, GITHUB_TOOLS_API_URL, PIPER_OWNER, PIPER_REPOSITORY, GITHUB_TOOLS_TOKEN)
      if (asset != null) {
        taskLib.debug('using stage config from release assets')
        config = asset.url
      } else {
        taskLib.debug('using stage config from sap-piper')
        // for older releases there is no stage config file in release assets.
        config = SAP_STAGE_CONDITIONS_URL
      }
    }
    return readConfig(stageConditionsParamName, SAP_DEFAULT_CONFIG_HOST, [config])
      .then(conditions => encodeConfig(conditions, stageConditionsParamName))
      .then(conditionsB64 => publishConfig(conditionsB64, stageConditionsParamName))
      .catch(taskLib.error)
  } else {
    return Promise.resolve()
  }
}

export async function preserveDefaultConfig() {
  const preserve = taskLib.getBoolInput('preserveDefaultConfig', false)
  if (preserve) {
    let defaults: string[]
    const asset = await getReleaseAsset(PIPER_VERSION, SAP_DEFAULTS_FILE, GITHUB_TOOLS_API_URL, PIPER_OWNER, PIPER_REPOSITORY, GITHUB_TOOLS_TOKEN)
    if (asset != null) {
      taskLib.debug('using defaults from release assets')
      defaults = [asset.url]
    } else {
      taskLib.debug('using defaults from sap-piper')
      // for older releases there is no defaults file in release assets.
      defaults = [SAP_DEFAULTS_FILE_URL]
    }

    const customDefaultsInput = taskLib.getDelimitedInput('customDefaults', '\n') || []
    defaults = defaults.concat(customDefaultsInput)

    return readConfig(defaultConfigParamName, SAP_DEFAULT_CONFIG_HOST, defaults)
      .then(config => encodeConfig(config, defaultConfigParamName))
      .then(configB64 => publishConfig(configB64, defaultConfigParamName))
      .catch(taskLib.error)
  } else {
    return Promise.resolve()
  }
}

export async function restorePipelineDefaults() {
  const defaultConfig = taskLib.getInput('restorePipelineDefaults', false)
  if (defaultConfig) {
    return decodeConfig(defaultConfig, 'DefaultConfig')
      .then(configString => writeConfig(configString, 'DefaultConfig'))
      .then(configFiles => taskLib.setTaskVariable(DEFAULTS_FILE_LIST, configFiles.map((file, i) => '--defaultConfig ' + file).join(' ')))
  }
  return Promise.resolve()
}

// this function is deprecated and will be removed
export async function restorePipelineStageConditions() {
  let stageConditions = taskLib.getInput('restorePipelineStageConditions', false)
  if (!stageConditions) {
    stageConditions = taskLib.getInput('stageConditions', false)
  }
  if (stageConditions) {
    // taskLib.warning('There is a new way of determining which steps and stages are active. restorePipelineStageConditions will be deprecated in a future version of piper-azure-task.' +
    //   'Please refer to https://github.tools.sap/project-piper/piper-pipeline-azure/pull/304 and ask the maintainer of your pipeline definition to incorporate the changes mentioned there.')
    return decodeConfig(stageConditions, 'StageConditions')
      .then(configString => writeConfig(configString, 'StageConditions'))
      .then(configFileFlags => taskLib.setTaskVariable(STAGE_CONDITIONS_FILE_LIST, configFileFlags.join(' ')))
  }
  return Promise.resolve()
}

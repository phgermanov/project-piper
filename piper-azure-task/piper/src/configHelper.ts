import * as fs from 'fs'
import * as taskLib from 'azure-pipelines-task-lib/task'
import * as yaml from 'js-yaml'
import { decode, encode } from './utils'
import { DEFAULTS_FILE_LIST } from './config'
import { BINARY_NAME_VAR, GITHUB_TOOLS_TOKEN, STEP_NAME } from './globals'

const PIPELINE_CONFIG_DIR = '.pipeline'
const DEFAULTS_FILE = 'defaults.yaml'
const STAGE_CONDITIONS_FILE = 'stage_conditions.yaml'

interface Config {
  filepath: string
  content: string
}

export async function decodeConfig (configB64: any, configType: string) {
  taskLib.debug('decoding' + configType)
  const configString = decode(configB64)
  taskLib.debug('decoded pipeline defaults: ' + configString)
  return Promise.resolve(configString)
}

export async function encodeConfig (configString: string, configType: string) {
  taskLib.debug(`encoding ${configType}`)
  const configB64 = encode(configString)
  taskLib.debug(`encoded ${configType}: ${configString}`)
  return Promise.resolve(configB64)
}

// TODO: find out how to avoid export for testing
export function loadYaml (configString: any, configType: string): Config[] {
  const yml = yaml.load(configString)

  if (yml === undefined) {
    taskLib.debug('config yaml is undefined')
    return []
  }

  if (Array.isArray(yml)) {
    taskLib.debug("multiple config yamls from Piper's getDefaults")
    return yml as Config[]
  }

  const data: Config[] = []
  if (typeof yml === 'object') {
    const ymlObject = yml as Record<string, unknown>
    if ('filepath' in ymlObject && 'content' in ymlObject) {
      taskLib.debug("single config yaml from Piper's getDefaults")
      data.push(yml as Config)
      return data
    }
  }

  taskLib.debug('single config yaml')
  const outputPath = configType == 'DefaultConfig' ? DEFAULTS_FILE : STAGE_CONDITIONS_FILE
  return [{
    filepath: outputPath,
    content: configString
  }]
}

export async function writeConfig (configString: any, configType: string) {
  const configFileList: string[] = []

  if (!fs.existsSync(PIPELINE_CONFIG_DIR)) {
    taskLib.debug(`creating ${PIPELINE_CONFIG_DIR} folder`)
    fs.mkdirSync(PIPELINE_CONFIG_DIR)
  }

  if (fs.lstatSync(PIPELINE_CONFIG_DIR).isDirectory()) {
    const data = loadYaml(configString, configType)

    data.forEach(element => {
      let filename: string
      if (configType === 'StageConditions') {
        filename = STAGE_CONDITIONS_FILE
      } else {
        filename = element.filepath
        filename = filename.substring(filename.lastIndexOf('/') + 1)
      }
      // TODO: check if we have an default object or a map of config objects and write to different files
      const configFile = `${PIPELINE_CONFIG_DIR}/${filename}`
      if (!fs.existsSync(configFile)) {
        taskLib.debug(`writing ${configFile} file with content: \n${element.content}`)
        fs.writeFileSync(configFile, element.content)
        // return promisify(fs.writeFile)(configFile, element.content);
      } else {
        taskLib.debug(`${configFile} already exists`)
      }

      configFileList.push(configFile)
    })
  }
  return configFileList
}

export async function readConfig (configType: string, configHost: string, urls: string[]) {
  const binaryName = taskLib.getVariable(BINARY_NAME_VAR) || ''
  if (binaryName) {
    taskLib.debug(`reading ${configType}`)
  } else {
    return Promise.reject('Piper binary not available')
  }

  const token = GITHUB_TOOLS_TOKEN
  if (!token) {
    return Promise.reject(`Can't read ${configType}: no GitHub token available. Make sure to add a GitHub service connection!`)
  }

  const urlArgs = urls.map((url) => '--defaultsFile ' + url).join(' ')

  const configString = taskLib
    .tool(binaryName)
    .arg('getDefaults')
    .argIf(configType === 'StageConditions', ['--useV1'])
    .line(urlArgs)
    .arg(['--gitHubTokens', `${configHost}:${token}`])
    .execSync().stdout
  taskLib.debug(`read ${configType}: ${configString}`)
  return Promise.resolve(configString)
}

export async function publishConfig (configB64: string, configType: string) {
  taskLib.debug(`publishing ${configType}`)
  taskLib.setVariable(configType, configB64, false, true)
  return Promise.resolve()
}

// this function is deprecated and will be removed
export async function checkStageConditions () {
  const binaryName = taskLib.getVariable(BINARY_NAME_VAR) || ''
  const stepName = STEP_NAME
  const stageName = getStageName()
  const defaultsFileList = (taskLib.getTaskVariable(DEFAULTS_FILE_LIST) || '').split(' ')
  const customConfig = getCustomConfigLocation()

  const result = taskLib
    .tool(binaryName)
    .arg('checkIfStepActive')
    .arg(['--stageConfig', `${PIPELINE_CONFIG_DIR}/${STAGE_CONDITIONS_FILE}`])
    .arg(defaultsFileList)
    .argIf(customConfig, ['--customConfig', customConfig])
    .argIf(stageName, ['--stage', stageName])
    .argIf(stepName, ['--step', stepName])
  // .arg(["--gitHubTokens", "${{ parameters.gitHubTokens }}"])
    .execSync()

  if (result.error) return Promise.reject(result.error)
  taskLib.debug('RESULT: ' + result.code)
  taskLib.setVariable('__stepActive', `${result.code === 0}`)
  if (result.code !== 0) { return Promise.reject(`step ${stepName} is deactivated`) }
}

export function getCustomConfigLocation (): string {
  const customConfig = taskLib.getInput('customConfigLocation', false) || ''
  if (customConfig) {
    taskLib.debug(`using Piper config location from task input: ${customConfig}`)
  }
  return customConfig
}

export function getStageName (): string {
  const stageNameFromInput = taskLib.getInput('stageName', false) || ''
  if (stageNameFromInput) {
    taskLib.debug('using stage name from task input')
    return stageNameFromInput
  }
  // FIXE: only necessary as log as https://github.com/SAP/jenkins-library/issues/3148 is not fixed
  const stageNameFromEnv = process.env.SYSTEM_STAGEDISPLAYNAME || ''
  if (stageNameFromEnv && stageNameFromEnv !== '__default') {
    taskLib.debug('using stage name from environment variables')
    return stageNameFromEnv
  }
  taskLib.debug('no stage name detected')
  return ''
}

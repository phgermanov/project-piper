import * as taskLib from 'azure-pipelines-task-lib'
import * as environment from '../src/environment'
import * as utils from '../src/utils'
import { type IExecSyncResult, ToolRunner } from 'azure-pipelines-task-lib/toolrunner'
import fs from 'fs'
import { BINARY_NAME_VAR } from '../src/globals'

describe('Environment', () => {
  beforeEach(() => {
    jest.spyOn(taskLib, 'debug').mockImplementation((message: string) => {
      console.log(message)
    })
    jest.spyOn(taskLib, 'warning').mockImplementation((message: string) => {
      console.log(message)
    })
    jest.spyOn(taskLib, 'tool').mockReturnValue(
      ToolRunner.prototype
    )
    jest.spyOn(ToolRunner.prototype, 'arg').mockReturnValue(
      ToolRunner.prototype
    )
  })

  afterEach(() => {
    jest.resetAllMocks()
    jest.clearAllMocks()
  })

  test('loadPipelineEnvToDisk when ".pipeline/commonPipelineEnvironment" already exist and ' +
    'pipelineEnvironment_b64 is empty', async () => {
    jest.spyOn(fs, 'existsSync').mockReturnValue(true)
    jest.spyOn(taskLib, 'getVariable').mockReturnValue(undefined)

    await environment.testExports.loadPipelineEnvToDisk()
    expect(taskLib.debug).toHaveBeenCalledWith(expect.stringContaining('already exists'))
    expect(taskLib.debug).not.toHaveBeenCalledWith('loading pipelineEnv to disk')
  })

  test('loadPipelineEnvToDisk when ".pipeline/commonPipelineEnvironment" does not exist and ' +
    'pipelineEnvironment_b64 is empty', async () => {
    jest.spyOn(fs, 'existsSync').mockReturnValue(false)
    jest.spyOn(taskLib, 'getVariable').mockReturnValue(undefined)

    await environment.testExports.loadPipelineEnvToDisk()
    expect(taskLib.debug).toHaveBeenCalledWith(expect.stringContaining('variable is not set'))
    expect(taskLib.debug).not.toHaveBeenCalledWith('loading pipelineEnv to disk')
  })

  test('loadPipelineEnvToDisk when ".pipeline/commonPipelineEnvironment" does not exist and ' +
    'pipelineEnvironment_b64 stage variable exist', async () => {
    jest.spyOn(fs, 'existsSync').mockReturnValue(false)
    jest.spyOn(taskLib, 'getVariable').mockImplementation((name: string): string | undefined => {
      if (name === 'pipelineEnvironment_b64') {
        return 'eyJ0MSI6ICJ2MSJ9'
      }
      if (name === BINARY_NAME_VAR) {
        return ''
      }
    })

    expect.assertions(1)
    try {
      await environment.testExports.loadPipelineEnvToDisk()
      expect(taskLib.debug).toHaveBeenCalledWith('loading pipelineEnv to disk')
    } catch (e) {
      expect(e).toEqual(Error('piper binary not available'))
    }
  })

  test('loadPipelineEnvToDisk when deprecated input is used', async () => {
    jest.spyOn(taskLib, 'getInput').mockImplementation((name: string): string | undefined => {
      if (name === 'restorePipelineEnv') {
        return 'test'
      }
    })

    await environment.testExports.loadPipelineEnvToDisk()
    expect(taskLib.warning).toHaveBeenCalledWith(expect.stringContaining('"restorePipelineEnv" is a deprecated input'))
  })

  test('decodePipelineEnv', async () => {
    const input = 'eyJ0MSI6ICJ2MSJ9'
    const want = '{"t1": "v1"}'

    const result = await environment.testExports.decodePipelineEnv(input)
    expect(result).toBe(want)
  })

  test('writePipelineEnv when binary not available', async () => {
    jest.spyOn(taskLib, 'getVariable').mockImplementation(() => {
      return ''
    })

    expect.assertions(1)
    try {
      await environment.testExports.writePipelineEnv('some-env')
    } catch (e) {
      expect(e).toEqual(Error('piper binary not available'))
    }
  })

  test('writePipelineEnv when succeed', async () => {
    const expected = '{"t1": "v1"}'
    jest.spyOn(taskLib, 'getVariable').mockReturnValue('piper-library')
    jest.spyOn(ToolRunner.prototype, 'execSync').mockReturnValue(
      { stdout: expected } as IExecSyncResult
    )

    await environment.testExports.writePipelineEnv(expected)
    expect(taskLib.debug).toHaveBeenCalledWith('writing pipeline env')
  })

  test('loadPipelineEnvFromDisk: exportPipelineEnv = false and preservePipelineEnv = false', async () => {
    jest.spyOn(taskLib, 'getBoolInput').mockImplementation((name: string): boolean => {
      switch (name) {
        case 'exportPipelineEnv':
          return false
        case 'preservePipelineEnv':
          return false
        default:
          return false
      }
    })

    await environment.testExports.loadPipelineEnvFromDisk()
    expect(taskLib.debug).not.toHaveBeenCalledWith('loading pipelineEnv from disk')
  })

  test('loadPipelineEnvFromDisk: exportPipelineEnv = false and preservePipelineEnv = true', async () => {
    jest.spyOn(taskLib, 'getBoolInput').mockImplementation((name: string): boolean => {
      switch (name) {
        case 'exportPipelineEnv':
          return false
        case 'preservePipelineEnv':
          return true
        default:
          return false
      }
    })

    await environment.testExports.loadPipelineEnvFromDisk()
    expect(taskLib.warning).toHaveBeenCalledWith(expect.stringContaining('"preservePipelineEnv" is a deprecated input.'))
    expect(taskLib.debug).toHaveBeenCalledWith('loading pipelineEnv from disk')
  })

  test('loadPipelineEnvFromDisk: exportPipelineEnv = true and preservePipelineEnv = false', async () => {
    jest.spyOn(taskLib, 'getBoolInput').mockImplementation((name: string): boolean => {
      switch (name) {
        case 'exportPipelineEnv':
          return true
        case 'preservePipelineEnv':
          return false
        default:
          return false
      }
    })

    await environment.testExports.loadPipelineEnvFromDisk()
    expect(taskLib.warning).not.toHaveBeenCalledWith(expect.stringContaining('"preservePipelineEnv" is a deprecated input.'))
    expect(taskLib.debug).toHaveBeenCalledWith('loading pipelineEnv from disk')
  })

  test('readPipelineEnv when succeed', async () => {
    const expected = '{"t1": "v1"}'
    jest.spyOn(taskLib, 'getVariable').mockReturnValue('piper-library')
    jest.spyOn(ToolRunner.prototype, 'execSync').mockReturnValue(
      { stdout: expected } as IExecSyncResult
    )

    const result = await environment.testExports.readPipelineEnv()

    expect(taskLib.debug).toHaveBeenCalledWith('reading pipeline env')
    expect(result).toBe(expected)
  })

  test('readPipelineEnv when binary not available', async () => {
    jest.spyOn(taskLib, 'getVariable').mockReturnValue('')

    expect.assertions(1)
    try {
      await environment.testExports.readPipelineEnv()
    } catch (e) {
      expect(e).toEqual(Error('piper binary not available'))
    }
  })

  test('encodePipelineEnv', async () => {
    const testEnv = 'test'
    const result = await environment.testExports.encodePipelineEnv(testEnv)
    const t = utils.encode(testEnv)
    expect(taskLib.debug).toHaveBeenCalledWith('encoding pipeline env')
    expect(taskLib.debug).toHaveBeenCalledWith('encoded pipeline env: ' + t)
    expect(result).toBe(t)
  })

  test('publishPipelineEnv', async () => {
    const testEnv = 'test'
    jest.spyOn(taskLib, 'setVariable').mockImplementation(() => {
    })
    await environment.testExports.publishPipelineEnv(testEnv)

    expect(taskLib.debug).toHaveBeenCalledWith('publishing pipeline env')
  })
})

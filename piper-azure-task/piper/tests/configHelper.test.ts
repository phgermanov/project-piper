import * as helper from '../src/configHelper'
import * as taskLib from 'azure-pipelines-task-lib'
import { type IExecSyncResult, ToolRunner } from 'azure-pipelines-task-lib/toolrunner'

describe('Helper', () => {
  beforeEach(() => {
    jest.spyOn(taskLib, 'debug').mockImplementation((name: string) => {
      console.log(name)
    })
    jest.spyOn(ToolRunner.prototype, 'arg').mockReturnValue(
      ToolRunner.prototype
    )
    jest.spyOn(ToolRunner.prototype, 'argIf').mockReturnValue(
      ToolRunner.prototype
    )
    jest.spyOn(taskLib, 'tool').mockReturnValue(
      ToolRunner.prototype
    )
  })

  afterEach(() => {
    jest.resetAllMocks()
    jest.clearAllMocks()
  })

  test('test decodeConfig', async () => {
    const configType = 'test type'
    const configB64 = 'test configuration'
    const configString = await helper.decodeConfig(configB64, configType)

    expect(taskLib.debug).toHaveBeenCalledWith('decoding' + configType)
    expect(taskLib.debug).toHaveBeenCalledWith('decoded pipeline defaults: ' + configString)
  })

  test('test encodeConfig', async () => {
    const configType = 'test type'
    const configB64 = 'test configuration'
    await helper.encodeConfig(configB64, configType)

    expect(taskLib.debug).toHaveBeenCalledWith(`encoding ${configType}`)
    expect(taskLib.debug).toHaveBeenCalledWith(`encoded ${configType}: ${configB64}`)
  })

  test('test publishConfig', async () => {
    const configType = 'test type'
    const configB64 = 'test configuration'

    await helper.publishConfig(configB64, configType)

    expect(taskLib.debug).toHaveBeenCalledWith(`publishing ${configType}`)
  })

  test('test getStageName using stage name from task input', async () => {
    const expectedResult = 'test'
    jest.spyOn(taskLib, 'getInput').mockImplementation(() => {
      return expectedResult
    })

    const result = helper.getStageName()

    expect(result).toBe(expectedResult)
    expect(taskLib.debug).toHaveBeenCalledWith('using stage name from task input')
  })

  test('test getStageName using stage name from environment variables', async () => {
    const expectedResult = 'test'
    jest.spyOn(taskLib, 'getInput').mockImplementation(() => {
      return ''
    })
    process.env.SYSTEM_STAGEDISPLAYNAME = expectedResult
    const result = helper.getStageName()

    expect(result).toBe(expectedResult)
    expect(taskLib.debug).toHaveBeenCalledWith('using stage name from environment variables')
  })

  test('test getStageName no stage name', async () => {
    const expectedResult = ''
    process.env.SYSTEM_STAGEDISPLAYNAME = ''
    const result = helper.getStageName()

    expect(taskLib.debug).toHaveBeenCalledWith('no stage name detected')
    expect(result).toBe(expectedResult)
  })

  test('test checkStageConditions when succeed', async () => {
    jest.spyOn(taskLib, 'getVariable').mockImplementation(() => {
      return 'testVariable'
    })
    jest.spyOn(taskLib, 'getInput').mockImplementation(() => {
      return 'testVariable'
    })
    jest.spyOn(helper, 'getStageName').mockImplementation(() => {
      return 'testStageName'
    })
    jest.spyOn(taskLib, 'getTaskVariable').mockImplementation(() => {
      return 'test Task Variable'
    })
    jest.spyOn(ToolRunner.prototype, 'execSync').mockReturnValue(
      { code: 0 } as IExecSyncResult
    )

    await helper.checkStageConditions()

    expect(taskLib.debug).toHaveBeenCalledWith('RESULT: 0')
  })

  test('test checkStageConditions when fails', async () => {
    jest.spyOn(taskLib, 'getVariable').mockImplementation(() => {
      return 'testVariable'
    })
    jest.spyOn(taskLib, 'getInput').mockImplementation(() => {
      return 'testVariable'
    })
    jest.spyOn(helper, 'getStageName').mockImplementation(() => {
      return 'testStageName'
    })
    jest.spyOn(taskLib, 'getTaskVariable').mockImplementation(() => {
      return 'test Task Variable'
    })
    jest.spyOn(ToolRunner.prototype, 'execSync').mockReturnValue(
      { code: 1 } as IExecSyncResult
    )

    expect.assertions(1)
    try {
      await helper.checkStageConditions()
    } catch (e) {
      expect(e).toContain('is deactivated')
    }
  })

  describe('test loadYaml', () => {
    test('with empty yaml', async () => {
      // test
      const result = helper.loadYaml('', 'DefaultConfig')
      // assert
      expect(result).toHaveLength(0)
    })

    test('with multi yaml', async () => {
      // test
      const result = helper.loadYaml('[{}, {}]', 'DefaultConfig')
      // assert
      expect(result).toHaveLength(2)
    })

    test('with single yaml', async () => {
      // test
      const result = helper.loadYaml('{"filepath": "anything", "content": some base-64 string}', 'DefaultConfig')
      // assert
      expect(result).toHaveLength(1)
      expect(result[0]).toHaveProperty('filepath', 'anything')
      expect(result[0]).toHaveProperty('content', 'some base-64 string')
    })

    test('with config yaml', async () => {
      // test
      const result = helper.loadYaml('some base-64 string', 'DefaultConfig')
      // assert
      expect(result).toHaveLength(1)
      expect(result[0]).toHaveProperty('filepath', 'defaults.yaml')
      expect(result[0]).toHaveProperty('content', 'some base-64 string')
    })
  })
})

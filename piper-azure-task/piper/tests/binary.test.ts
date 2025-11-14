import * as binaryTs from '../src/binary'
import * as taskLib from 'azure-pipelines-task-lib'
import * as toolLib from 'azure-pipelines-tool-lib'
import * as utilsTs from '../src/utils'
import * as globalsTs from '../src/globals'
import * as githubTs from '../src/github'

jest.mock('azure-pipelines-task-lib')
jest.mock('azure-pipelines-tool-lib')
jest.mock('../src/github')

describe('binary.ts', () => {
  beforeEach(() => {
  })

  afterEach(() => {
    jest.resetAllMocks()
    jest.clearAllMocks()
  })

  test('fetchBinaryVersionTag. default latest version', async () => {
    // mock
    jest.spyOn(taskLib, 'getBoolInput').mockReturnValue(true)  // fetchPiperBinaryVersionTag
    jest.spyOn(taskLib, 'getInput').mockReturnValue('latest')  // sapPiperVersion
    jest.spyOn(githubTs, 'fetchLatestVersionTag').mockResolvedValue('2.3.4')
    globalsTs.initializeGlobals()

    // execute
    await binaryTs.fetchBinaryVersionTag()

    // assert
    expect(taskLib.warning).not.toHaveBeenCalledWith(utilsTs.msgWarnMasterBinary)
    expect(taskLib.debug).toHaveBeenCalledWith(expect.stringContaining('resolving latest version tag for'))
  })
  test('fetchBinaryVersionTag. master version warning and fetch latest', async () => {
    // mock
    jest.spyOn(taskLib, 'getBoolInput').mockReturnValue(true)  // fetchPiperBinaryVersionTag
    jest.spyOn(taskLib, 'getVariable').mockReturnValue('master')  // sapPiperVersion
    jest.spyOn(githubTs, 'fetchLatestVersionTag').mockResolvedValue('2.3.4')
    globalsTs.initializeGlobals()

    // execute
    await binaryTs.fetchBinaryVersionTag()

    // assert
    expect(taskLib.warning).toHaveBeenCalledWith(utilsTs.msgWarnMasterBinary)
    expect(taskLib.debug).toHaveBeenCalledWith(expect.stringContaining('resolving latest version tag for'))
  })
  test('fetchBinaryVersionTag. exact version', async () => {
    // mock
    jest.spyOn(taskLib, 'getBoolInput').mockReturnValue(true)  // fetchPiperBinaryVersionTag
    jest.spyOn(taskLib, 'getVariable').mockReturnValue('v1.2.3')  // sapPiperVersion
    globalsTs.initializeGlobals()

    // execute
    await binaryTs.fetchBinaryVersionTag()

    // assert
    expect(taskLib.setVariable).toHaveBeenCalledWith('binaryVersion', 'v1.2.3')
    expect(taskLib.setVariable).toHaveBeenCalledWith('binaryVersion', 'v1.2.3', false, true)
  })

  test('preparePiperBinary. found in tool cache', async () => {
    // mock
    jest.spyOn(taskLib, 'getVariable').mockImplementation((key: string) => {
      if (key === 'binaryVersion') {
        return 'v1.2.3'
      }
    })
    jest.spyOn(toolLib, 'findLocalTool').mockReturnValue('/path/to/piper')  // sapPiperVersion
    globalsTs.initializeGlobals()

    // execute
    await binaryTs.preparePiperBinary()

    // assert
    expect(taskLib.debug).toHaveBeenCalledWith(expect.stringContaining('binary piper version v1.2.3 found in tool cache'))
    expect(taskLib.debug).not.toHaveBeenCalledWith(expect.stringContaining('downloading sap-piper version v1.2.3 from'))
    expect(toolLib.prependPath).toHaveBeenCalledWith('/path/to/piper')
  })
  test('preparePiperBinary. download from via API', async () => {
    // mock
    jest.spyOn(taskLib, 'getVariable').mockImplementation((key: string) => {
      if (key === 'binaryVersion') {
        return 'v1.2.3'
      }
    })
    jest.spyOn(toolLib, 'findLocalTool').mockReturnValue('')  // sapPiperVersion
    jest.spyOn(githubTs, 'downloadBinaryViaAPI').mockResolvedValue('/tmp/path/to/piper')
    globalsTs.initializeGlobals()

    // execute
    await binaryTs.preparePiperBinary()

    // assert
    expect(taskLib.debug).toHaveBeenCalledWith('adding piper version v1.2.3 to tool cache')
    expect(toolLib.cacheFile).toHaveBeenCalledWith('/tmp/path/to/piper', 'piper', 'piper', 'v1.2.3')
  })
})

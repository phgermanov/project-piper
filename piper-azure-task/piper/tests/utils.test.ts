import * as utils from '../src/utils'
import * as taskLib from 'azure-pipelines-task-lib'
import { initializeGlobals } from '../src/globals'

describe('Utils', () => {
  test('Test encoding', async () => {
    const input = 'test'
    const expected = 'dGVzdA=='
    const tag = utils.encode(input)

    expect(tag).toBe(expected)
  })

  test('Test encoding', async () => {
    const input = 'test'
    const expected = 'µë-'
    const tag = utils.decode(input)
    expect(tag).toBe(expected)
  })

  test('is Internal Step when false', async () => {
    jest.spyOn(taskLib, 'getInput').mockImplementation(() => {
      return 'github.com'
    })
    initializeGlobals()

    const isInternal = utils.isInternalStep()

    expect(isInternal).toBe(false)
  })

  test('is Internal Step when true', async () => {
    jest.spyOn(taskLib, 'getInput').mockImplementation(() => {
      return 'sap.github.com'
    })
    initializeGlobals()

    const isInternal = utils.isInternalStep()

    expect(isInternal).toBe(true)
  })
})

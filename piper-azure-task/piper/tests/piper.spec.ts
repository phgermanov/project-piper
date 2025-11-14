import * as path from 'path'
import * as fs from 'fs'

import * as ttm from 'azure-pipelines-task-lib/mock-test'

describe('Piper', function () {
  beforeEach(() => {
    jest.setTimeout(10000)
    const dir = './tmp'

    try {
      if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir)
        // console.log("Directory is created.");
      } else {
        // console.log("Directory already exists.");
      }
    } catch (err) {
      console.log(err)
    }
  })

  afterEach(() => {
    delete process.env.AGENT_TOOLSDIRECTORY
    delete process.env.AGENT_TEMPDIRECTORY
  })

  describe('Binary', function () {
    const taskJSONPath = path.join(__dirname, '..', '..', 'piper', 'task.json')
    it.skip('fetch from tool cache', () => {
      // init
      const taskPath = path.join(__dirname, '..', 'dist', 'tests', 'TestToolCache.js')
      const tr = new ttm.MockTestRunner(taskPath, taskJSONPath)
      // test
      tr.runAsync()
      console.log(tr.stderr)
      console.log(tr.stdout)
      // assert
      expect(tr.invokedToolCount).toBe(1)
      expect(tr.stderr).toBe('')
      expect(tr.stdout).not.toContain('downloading')
      expect(tr.stdout).toContain('looking for')
      expect(tr.stdout).toContain('piper arg: version')
      expect(tr.stdout).toContain('checking cache')
      expect(tr.stdout).not.toContain('cache not found')
      expect(tr.stdout).toContain('prepending path')
      expect(tr.stdout).toContain('exec tool: piper')
      expect(tr.succeeded).toBeTruthy()
    })
    it.skip('downloads from releases', async () => {
      // init
      const taskPath = path.join(__dirname, '..', 'dist', 'tests', 'TestToolDownload.js')
      const tr = new ttm.MockTestRunner(taskPath, taskJSONPath)
      // test
      await tr.runAsync()
      console.log(tr.stderr)
      console.log(tr.stdout)
      // assert
      expect(tr.invokedToolCount).toBe(1)
      expect(tr.stderr).toBe('')
      expect(tr.stdout).toContain('not found in tool cache')
      expect(tr.stdout).not.toContain('requesting tool piper')
      expect(tr.stdout).toContain('downloading piper version')
      expect(tr.stdout).toContain('caching tool piper')
      expect(tr.stdout).toContain('exec tool: piper')
      expect(tr.succeeded).toBeTruthy()
    })
  })
})

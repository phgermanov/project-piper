import * as tmrm from 'azure-pipelines-task-lib/mock-run'
import type * as ma from 'azure-pipelines-task-lib/mock-answer'
import path from 'node:path'

const taskPath = path.join(__dirname, '..', 'src', 'index.js')
const tr: tmrm.TaskMockRunner = new tmrm.TaskMockRunner(taskPath)
const binaryName = 'piper'
const binaryVersion = '1.260.0'
const toolDirectory = 'tools'
const tempDirectory = 'tmp'

tr.setInput('stepName', 'version')
tr.setInput('piperVersion', binaryVersion)
console.log('Inputs have been set')

process.env.AGENT_TOOLSDIRECTORY = toolDirectory
process.env.AGENT_TEMPDIRECTORY = tempDirectory
console.log('Envs have been set')

// provide answers for task mock
const a: ma.TaskLibAnswers = <ma.TaskLibAnswers>{
  cwd: {
    cwd: '.'
  },
  checkPath: {} as any,
  exec: {
    'piper getDefaults': {
      code: 0,
      stdout: 'empty'
    },
    'piper version ': {
      code: 0,
      stdout: 'piper-version:\n\t\tcommit: "<n/a>"\n\t\ttag: "<n/a>"'
    },
    // FIXME: on Azure DevOps stage name is detected
    'piper version  --stageName Build': {
      code: 0,
      stdout: 'piper-version:\n\t\tcommit: "<n/a>"\n\t\ttag: "<n/a>"'
    }
  },
  which: {
    piper: `${toolDirectory}/piper/${binaryVersion}/x64/${binaryName}`
  }
}

// Add extra answer definitions that need to be dynamically generated
a.checkPath![`${toolDirectory}/piper/${binaryVersion}/x64/${binaryName}`] = true

tr.setAnswers(<any>a)

tr.registerMock('azure-pipelines-tool-lib/tool', {
  findLocalTool: (toolName: string, version: string) => {
    const cachePath = path.join(toolDirectory, toolName, version, 'x64')
    if (toolName && version) {
      console.log(`checking cache: ${cachePath}`)
      return cachePath
    } else {
      console.log('cache not found')
    }
  },
  prependPath: (pathToTool: string) => {
    console.log(`prepending path ${pathToTool}`)
  }
})
console.log('Mocks have been prepared')

tr.run()

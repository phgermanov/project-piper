import * as docker from '../src/docker'
import * as taskLib from 'azure-pipelines-task-lib/task'
import { type IExecSyncResult, ToolRunner } from 'azure-pipelines-task-lib/toolrunner'

describe('Docker', () => {
  let mockInput: any = {}
  const mockVariables: any = {}

  beforeEach(() => {
    jest.spyOn(taskLib, 'tool').mockReturnValue(
      ToolRunner.prototype
    )
    jest.spyOn(ToolRunner.prototype, 'arg').mockReturnValue(
      ToolRunner.prototype
    )
    jest.spyOn(ToolRunner.prototype, 'line').mockReturnValue(
      ToolRunner.prototype
    )
    jest.spyOn(ToolRunner.prototype, 'argIf').mockReturnValue(
      ToolRunner.prototype
    )
    jest.spyOn(ToolRunner.prototype, 'exec').mockReturnValue(
      Promise.resolve(0)
    )
    jest.spyOn(ToolRunner.prototype, 'execSync').mockReturnValue(
      { code: 0 } as IExecSyncResult
    )

    jest.spyOn(taskLib, 'getInput').mockImplementation((name: string) => {
      if (name in mockInput) {
        return mockInput[name]
      }
    })

    jest.spyOn(taskLib, 'getTaskVariable').mockImplementation((name: string) => {
      if (name in mockVariables) {
        return mockVariables[name]
      }
    })

    jest.spyOn(taskLib, 'getEndpointAuthorization').mockReturnValue({
      parameters: {
        username: 'mockUsername',
        password: 'mockPassword',
        registry: 'mockRegistry'
      },
      scheme: ''
    } as taskLib.EndpointAuthorization)
  })

  afterEach(() => {
    mockInput = {}
    jest.resetAllMocks()
    jest.clearAllMocks()
  })

  test.skip('Start container without Docker options', async () => {
    const mockContextConfig = {}
    const exitCode = await docker.testExports.startContainer('golang:1', mockContextConfig)

    expect(exitCode).toBe(0)
    // --name could be an empty string if uuidv4 fails
    expect(ToolRunner.prototype.arg).not.toHaveBeenCalledWith(expect.arrayContaining(['--name', '']))
    expect(ToolRunner.prototype.line).toHaveBeenCalledWith('golang:1')
  })

  test.skip('Start container with Docker options from config', async () => {
    const mockContextConfig = { dockerOptions: ['-u 0'] }
    const exitCode = await docker.testExports.startContainer('golang:1', mockContextConfig)

    expect(exitCode).toBe(0)
    expect(ToolRunner.prototype.line).toHaveBeenCalledWith('-u 0')
    expect(ToolRunner.prototype.line).toHaveBeenCalledWith('golang:1')
  })

  test.skip('Start container with Docker options from task input', async () => {
    mockInput.dockerOptions = '-u 0'

    const mockContextConfig = {}
    process.env.INPUT_DOCKEROPTIONS = '-u 0'
    const exitCode = await docker.testExports.startContainer('golang:1', mockContextConfig)

    expect(exitCode).toBe(0)
    expect(ToolRunner.prototype.line).toHaveBeenCalledWith('-u 0')
    expect(ToolRunner.prototype.line).toHaveBeenCalledWith('golang:1')
  })

  test.skip('Start container with Docker Volume binds', async () => {
    const mockContextConfig = {
      dockerVolumeBind: {
        volume1: 'path1',
        volume2: 'path2'
      }
    }
    const exitCode = await docker.testExports.startContainer('golang:1', mockContextConfig)

    expect(exitCode).toBe(0)
    expect(ToolRunner.prototype.arg).toHaveBeenCalledWith(['--volume', 'volume1:path1', '--volume', 'volume2:path2'])
    expect(ToolRunner.prototype.line).toHaveBeenCalledWith('golang:1')
  })

  test.skip('Stop container', async () => {
    mockVariables.__containerID = 'testID'

    docker.testExports.stopContainer()

    expect(ToolRunner.prototype.execSync).toHaveBeenCalled()
    expect(ToolRunner.prototype.arg).toHaveBeenCalledWith(mockVariables.__containerID)
  })

  test.skip('No container to stop', async () => {
    mockVariables.__containerID = ''

    docker.testExports.stopContainer()

    expect(ToolRunner.prototype.execSync).not.toHaveBeenCalled()
  })

  test.skip('Azure env vars to be formatted correctly', async () => {
    const envVars = docker.getAzureEnvVars()

    expect(envVars.length % 2).toBe(0)
    const hasCorrectFlags = checkEnvVars(envVars)
    expect(hasCorrectFlags).toBe(true)
  })

  test.skip('Proxy env vars to be formatted correctly', async () => {
    const envVars = docker.getProxyEnvVars()

    expect(envVars.length % 2).toBe(0)
    const hasCorrectFlags = checkEnvVars(envVars)
    expect(hasCorrectFlags).toBe(true)
  })

  test.skip('Vault env vars to be formatted correctly', async () => {
    const envVars = docker.getVaultEnvVars()

    expect(envVars.length % 2).toBe(0)
    const hasCorrectFlags = checkEnvVars(envVars)
    expect(hasCorrectFlags).toBe(true)
  })
})

describe('docker.parseImageName', () => {
  const testCases = [
    { input: 'alpine', agentName: 'Hosted Agent', expected: ['docker-hub.common.repositories.cloud.sap', 'alpine'] },
    {
      input: 'alpine:latest',
      agentName: 'Azure Pipelines 123',
      expected: ['docker-hub.common.repositories.cloud.sap', 'alpine:latest']
    },
    {
      input: 'library/alpine',
      agentName: 'Azure Pipelines 123',
      expected: ['docker-hub.common.repositories.cloud.sap', 'library/alpine']
    },
    { input: 'test:1234/blaboon', agentName: 'Azure Pipelines 123', expected: ['test:1234', 'blaboon'] },
    {
      input: 'alpine:3.7',
      agentName: 'Azure Pipelines 123',
      expected: ['docker-hub.common.repositories.cloud.sap', 'alpine:3.7']
    },
    {
      input: 'docker.example.edu/gmr/alpine:3.7',
      agentName: 'Azure Pipelines 123',
      expected: ['docker.example.edu', 'gmr/alpine:3.7']
    },
    {
      input: 'docker.example.com:5000/gmr/alpine@sha256:5a156ff125e5a12ac7ff43ee5120fa249cf62248337b6d04abc574c8',
      agentName: 'Azure Pipelines 123',
      expected: ['docker.example.com:5000', 'gmr/alpine@sha256:5a156ff125e5a12ac7ff43ee5120fa249cf62248337b6d04abc574c8']
    },
    {
      input: 'docker.example.co.uk/gmr/alpine/test2:latest',
      agentName: 'Azure Pipelines 123',
      expected: ['docker.example.co.uk', 'gmr/alpine/test2:latest']
    },
    {
      input: 'registry.dobby.org/dobby/dobby-servers/arthound:2019-08-08',
      agentName: 'Azure Pipelines 123',
      expected: ['registry.dobby.org', 'dobby/dobby-servers/arthound:2019-08-08']
    },
    {
      input: 'owasp/zap:3.8.0',
      agentName: 'Azure Pipelines 123',
      expected: ['docker-hub.common.repositories.cloud.sap', 'owasp/zap:3.8.0']
    },
    {
      input: 'registry.dobby.co/dobby/dobby-servers/github-run:2021-10-04',
      agentName: 'Azure Pipelines 123',
      expected: ['registry.dobby.co', 'dobby/dobby-servers/github-run:2021-10-04']
    },
    {
      input: 'docker.elastic.co/kibana/kibana:7.6.2',
      agentName: 'Azure Pipelines 123',
      expected: ['docker.elastic.co', 'kibana/kibana:7.6.2']
    },
    {
      input: 'registry.dobby.org/dobby/dobby-servers/lerphound:latest',
      agentName: 'Azure Pipelines 123',
      expected: ['registry.dobby.org', 'dobby/dobby-servers/lerphound:latest']
    },
    {
      input: 'registry.dobby.org/dobby/dobby-servers/marbletown-poc:2021-03-29',
      agentName: 'Azure Pipelines 123',
      expected: ['registry.dobby.org', 'dobby/dobby-servers/marbletown-poc:2021-03-29']
    },
    {
      input: 'marbles/marbles:v0.38.1',
      agentName: 'Azure Pipelines 123',
      expected: ['docker-hub.common.repositories.cloud.sap', 'marbles/marbles:v0.38.1']
    },
    {
      input: 'registry.dobby.org/dobby/dobby-servers/loophole@sha256:5a156ff125e5a12ac7ff43ee5120fa249cf62248337b6d04abc574c8',
      agentName: 'Azure Pipelines 123',
      expected: ['registry.dobby.org', 'dobby/dobby-servers/loophole@sha256:5a156ff125e5a12ac7ff43ee5120fa249cf62248337b6d04abc574c8']
    },
    {
      input: 'sonatype/nexon:3.30.0',
      agentName: 'Azure Pipelines 123',
      expected: ['docker-hub.common.repositories.cloud.sap', 'sonatype/nexon:3.30.0']
    },
    {
      input: 'prom/node-exporter:v1.1.1',
      agentName: 'Azure Pipelines 123',
      expected: ['docker-hub.common.repositories.cloud.sap', 'prom/node-exporter:v1.1.1']
    },
    {
      input: 'sosedoff/pgweb@sha256:5a156ff125e5a12ac7ff43ee5120fa249cf62248337b6d04abc574c8',
      agentName: 'Azure Pipelines 123',
      expected: ['docker-hub.common.repositories.cloud.sap', 'sosedoff/pgweb@sha256:5a156ff125e5a12ac7ff43ee5120fa249cf62248337b6d04abc574c8']
    },
    {
      input: 'sosedoff/pgweb:latest',
      agentName: 'Azure Pipelines 123',
      expected: ['docker-hub.common.repositories.cloud.sap', 'sosedoff/pgweb:latest']
    },
    {
      input: 'registry.dobby.org/dobby/dobby-servers/arpeggio:2021-06-01',
      agentName: 'Azure Pipelines 123',
      expected: ['registry.dobby.org', 'dobby/dobby-servers/arpeggio:2021-06-01']
    },
    {
      input: 'registry.dobby.org/dobby/antique-penguin:release-production',
      agentName: 'Azure Pipelines 123',
      expected: ['registry.dobby.org', 'dobby/antique-penguin:release-production']
    },
    {
      input: 'dalprodictus/halcon:6.7.5',
      agentName: 'Azure Pipelines 123',
      expected: ['docker-hub.common.repositories.cloud.sap', 'dalprodictus/halcon:6.7.5']
    },
    {
      input: 'antigua/antigua:v31',
      agentName: 'Azure Pipelines 123',
      expected: ['docker-hub.common.repositories.cloud.sap', 'antigua/antigua:v31']
    },
    {
      input: 'weblate/weblate:4.7.2-1',
      agentName: 'Azure Pipelines 123',
      expected: ['docker-hub.common.repositories.cloud.sap', 'weblate/weblate:4.7.2-1']
    },
    {
      input: 'redis:4.0.01-alpine',
      agentName: 'Azure Pipelines 123',
      expected: ['docker-hub.common.repositories.cloud.sap', 'redis:4.0.01-alpine']
    },
    {
      input: 'registry.dobby.com/dobby/dobby-servers/github-run:latest',
      agentName: 'Azure Pipelines 123',
      expected: ['registry.dobby.com', 'dobby/dobby-servers/github-run:latest']
    },
    {
      input: 'build-milestones.common.repositories.cloud.sap/com.sap.prd.sonar/sonar-scanner-cache:11.1.1-sap-01',
      agentName: 'Azure Pipelines 123',
      expected: ['build-milestones.common.repositories.cloud.sap', 'com.sap.prd.sonar/sonar-scanner-cache:11.1.1-sap-01']
    },
    {
      input: 'sapperf.int.repositories.cloud.sap/performance_scalability/supa_performance_test_min:dev',
      agentName: 'asdqwe',
      expected: ['sapperf.int.repositories.cloud.sap', 'performance_scalability/supa_performance_test_min:dev']
    },
    {
      input: 'node:22.14.0@sha256:c7fd844945a76eeaa83cb372e4d289b4a30b478a1c80e16c685b62c54156285b',
      agentName: 'asdqwe',
      expected: ['', 'node:22.14.0@sha256:c7fd844945a76eeaa83cb372e4d289b4a30b478a1c80e16c685b62c54156285b']
    },
    {
      input: 'node:22.14.0@sha256:c7fd844945a76eeaa83cb372e4d289b4a30b478a1c80e16c685b62c54156285b',
      agentName: 'Azure Pipelines 123',
      expected: ['docker-hub.common.repositories.cloud.sap', 'node:22.14.0@sha256:c7fd844945a76eeaa83cb372e4d289b4a30b478a1c80e16c685b62c54156285b']
    },
  ]

  testCases.forEach(({ input, agentName, expected }) => {
    test(`parseImageName(${input})`, () => {
      jest.spyOn(taskLib, 'getVariable').mockReturnValue(agentName)
      expect(docker.testExports.parseImageName(input)).toEqual(expected)
    })
  })

  test('System trust token env vars to be formatted correctly', async () => {
    const envVars = docker.getSystemTrustEnvVars()

    expect(envVars.length % 2).toBe(0)
    const hasCorrectFlags = checkEnvVars(envVars)
    expect(hasCorrectFlags).toBe(true)
  })
})

function checkEnvVars (envVars: string[]) {
  for (let i = 0; i < envVars.length; i += 2) {
    if (envVars[i] !== '--env') {
      return false
    }
  }
  return true
}

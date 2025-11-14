import * as vault from '../src/vault'
import * as taskLib from 'azure-pipelines-task-lib'

describe('Vault', () => {
  let mockVariables: any = {}
  beforeEach(() => {
    jest.spyOn(taskLib, 'debug').mockImplementation((name: string) => {
      console.log(name)
    })

    jest.spyOn(taskLib, 'getVariable').mockImplementation((name: string) => {
      if (name in mockVariables) {
        return mockVariables[name]
      }
    })
  })

  afterEach(() => {
    jest.resetAllMocks()
    jest.clearAllMocks()
    process.env.PIPER_vaultAppRoleID = ''
    process.env.PIPER_vaultAppRoleSecretID = ''
    mockVariables = {}
  })

  test('Test when get vault secret from env', async () => {
    process.env.PIPER_vaultAppRoleID = 'test-role-id'
    process.env.PIPER_vaultAppRoleSecretID = 'test-role-secret-id'
    vault.prepareVaultEnvVars()

    expect(taskLib.debug).toHaveBeenCalledWith('using Vault secrets from environment variables')
  })

  test('Test when get vault secret from pipeline variables', async () => {
    mockVariables['hyperspace.vault.roleId'] = 'testRoleId'
    mockVariables['hyperspace.vault.secretId'] = 'testSecretId'
    vault.prepareVaultEnvVars()

    expect(taskLib.debug).toHaveBeenCalledWith('using Vault secrets from pipeline variables')
  })

  test('Test when secrets is empty', async () => {
    vault.prepareVaultEnvVars()

    expect(taskLib.debug).toHaveBeenCalledWith('no Vault secrets detected')
  })
})

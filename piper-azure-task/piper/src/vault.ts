import * as taskLib from 'azure-pipelines-task-lib/task'

export function prepareVaultEnvVars () {
  // checking env vars
  let appRoleID = process.env.PIPER_vaultAppRoleID
  let appRoleSecretID = process.env.PIPER_vaultAppRoleSecretID
  if (appRoleID && appRoleSecretID) {
    taskLib.debug('using Vault secrets from environment variables')
  } else {
    // checking pipline vars
    appRoleID = taskLib.getVariable('hyperspace.vault.roleId') || ''
    appRoleSecretID =
            taskLib.getVariable('hyperspace.vault.secretId') || ''
    if (appRoleID && appRoleSecretID) {
      taskLib.debug('using Vault secrets from pipeline variables')
      process.env.PIPER_vaultAppRoleID = appRoleID
      process.env.PIPER_vaultAppRoleSecretID = appRoleSecretID
    } else {
      taskLib.debug('no Vault secrets detected')
    }
  }
}

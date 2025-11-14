import { OctokitOptions } from '@octokit/core/dist-types/types'
import * as taskLib from 'azure-pipelines-task-lib'
import { retryableOctokitREST } from './utils'
import { chmodSync, writeFileSync } from 'fs'
import { v4 as uuidv4 } from 'uuid'
import path from 'node:path'
import { GITHUB_TOOLS_API_URL } from './globals'
import { arch, platform } from 'os'

export async function downloadBinaryViaAPI (version: string, binaryName: string, apiURL: string, owner: string, repo: string, token: string) {
  const options: OctokitOptions = {
    auth: token,
    baseUrl: apiURL,
  }

  let assetName = binaryName
  const platform = getPlatform()
  if (platform) {
    assetName += `-${platform}`
  }

  taskLib.debug(`downloading ${assetName} version ${version} from ${owner}/${repo} via API`)
  const asset = await getReleaseAsset(version, assetName, apiURL, owner, repo, token)
  const octokit = retryableOctokitREST(options)
  const response = await octokit.rest.repos.getReleaseAsset({
    owner, repo, asset_id: asset.id, headers: { Accept: 'application/octet-stream' }
  }).catch((error: any) => {
    return Promise.reject(new Error(`Could not download GitHub asset. Status: ${error.response.status}. Message: ${error.response.data.message}`))
  })

  // write to temporary file with uuid name
  const tempDir = taskLib.getVariable('Agent.TempDirectory') || ''
  const destinationPath = path.join(tempDir, uuidv4())

  taskLib.debug(`writing into downloaded binary as ${destinationPath}`)
  writeFileSync(destinationPath, Buffer.from(response.data as unknown as ArrayBuffer))
  chmodSync(destinationPath, '777')

  return Promise.resolve(destinationPath)
}

export async function getReleaseAsset (
  version: string,
  assetName: string,
  apiURL: string,
  owner: string,
  repo: string,
  token: string
): Promise<any> {
  const options: OctokitOptions = {
    auth: token,
    baseUrl: apiURL,
  }

  taskLib.debug(`fetching release '${version}' from ${owner}/${repo} to find '${assetName}'`)
  const octokit = retryableOctokitREST(options)
  const response = await octokit.rest.repos.getReleaseByTag({ owner, repo, tag: version })
    .catch((error: any) => {
      return Promise.reject(new Error(`Getting release failed. Status: ${error.response.status}. Message: ${error.response.data.message}`))
    })

  taskLib.debug(`looking for asset '${assetName}' in release ${version} of ${owner}/${repo} repository`)
  const asset = response.data.assets.find((value): boolean => {
    return value.name === assetName
  })
  if (asset) {
    return Promise.resolve(asset)
  }

  return Promise.resolve(null)
}

// fetchLatestVersionTag will fetch the latest release from GitHub releases and returns tag_name field
export async function fetchLatestVersionTag (owner: string, repo: string, token: string): Promise<string> {
  if (token === '') {
    return Promise.reject(new Error('GitHub token is required. Did you set up service connection?'))
  }

  const options: OctokitOptions = {
    auth: token,
    baseUrl: GITHUB_TOOLS_API_URL,
  }

  const octokit = retryableOctokitREST(options)
  const response = await octokit.rest.repos.getLatestRelease({ owner, repo })
    .catch((error: any) => {
      return Promise.reject(new Error(`Getting latest release failed. Status: ${error.response.status}. Message: ${error.response.data.message}`))
    })

  if (response.data.tag_name === '') {
    return Promise.reject(new Error('Response didn\'t contain tag name'))
  }

  return Promise.resolve(response.data.tag_name)
}

export function getPlatform (): string {
  switch (platform()) {
    case 'darwin':
      if (arch().startsWith('arm')) { // could be arm or arm64
        return 'darwin.arm64' // Apple Silicon
      } else {
        return 'darwin.x86_64' // Intel Mac
      }
    default:
      return ''
  }
}

import { downloadBinaryViaAPI, fetchLatestVersionTag, getReleaseAsset } from '../src/github'
import * as taskLib from 'azure-pipelines-task-lib'
import { retryableOctokitREST } from '../src/utils'
import path from 'node:path'

jest.mock('../src/utils', () => ({
  ...jest.requireActual('../src/utils'),
  retryableOctokitREST: jest.fn(),
}))

jest.mock('azure-pipelines-task-lib')

jest.mock('fs', () => ({
  chmodSync: jest.fn(),
  writeFileSync: jest.fn(),
}))

jest.mock('uuid', () => ({
  v4: jest.fn(() => 'mock-uuid'),
}))

describe('getReleaseAsset', () => {
  const mockGetReleaseByTag = jest.fn()
  const mockOctokit = {
    rest: {
      repos: {
        getReleaseByTag: mockGetReleaseByTag,
      },
    },
  }

  const version = 'v1.2.3'
  const assetName = 'sap-piper'
  const apiURL = 'https://api.github.com'
  const owner = 'test-owner'
  const repo = 'test-repo'
  const token = 'test-token'

  beforeEach(() => {
    (retryableOctokitREST as jest.Mock).mockReturnValue(mockOctokit)
    jest.clearAllMocks()
  })

  it('should return the asset when it is found', async () => {
    const mockAsset = { name: assetName, url: 'https://example.com/asset' }
    mockGetReleaseByTag.mockResolvedValue({
      data: {
        assets: [mockAsset],
      },
    })

    const result = await getReleaseAsset(version, assetName, apiURL, owner, repo, token)

    expect(taskLib.debug).toHaveBeenCalledWith(`fetching release '${version}' from ${owner}/${repo} to find '${assetName}'`)
    expect(taskLib.debug).toHaveBeenCalledWith(`looking for asset '${assetName}' in release ${version} of ${owner}/${repo} repository`)
    expect(result).toEqual(mockAsset)
  })
  it('should return null when the asset is not found', async () => {
    mockGetReleaseByTag.mockResolvedValue({
      data: {
        assets: [],
      },
    })

    const result = await getReleaseAsset(version, assetName, apiURL, owner, repo, token)

    expect(taskLib.debug).toHaveBeenCalledWith(`fetching release '${version}' from ${owner}/${repo} to find '${assetName}'`)
    expect(taskLib.debug).toHaveBeenCalledWith(`looking for asset '${assetName}' in release ${version} of ${owner}/${repo} repository`)
    expect(result).toBeNull()
  })
  it('should throw an error when the API call fails', async () => {
    const errorResponse = {
      response: {
        status: 404,
        data: { message: 'Not Found' },
      },
    }
    mockGetReleaseByTag.mockRejectedValue(errorResponse)

    await expect(getReleaseAsset(version, assetName, apiURL, owner, repo, token))
      .rejects.toThrow('Getting release failed. Status: 404. Message: Not Found')

    expect(taskLib.debug).toHaveBeenCalledWith(`fetching release '${version}' from ${owner}/${repo} to find '${assetName}'`)
  })
})

describe('fetchLatestVersionTag', () => {
  const mockOctokit = {
    rest: {
      repos: {
        getLatestRelease: jest.fn(),
      },
    },
  }

  const owner = 'test-owner'
  const repo = 'test-repo'
  const token = 'test-token'

  beforeEach(() => {
    jest.clearAllMocks();
    (retryableOctokitREST as jest.Mock).mockReturnValue(mockOctokit)
  })

  it('should return the latest version tag', async () => {
    const mockTagName = 'v1.2.3'
    mockOctokit.rest.repos.getLatestRelease.mockResolvedValue({
      data: { tag_name: mockTagName },
    })

    const result = await fetchLatestVersionTag(owner, repo, token)

    expect(result).toBe(mockTagName)
  })

  it('should throw an error if the token is empty', async () => {
    await expect(fetchLatestVersionTag(owner, repo, '')).rejects.toThrow(
      'GitHub token is required. Did you set up service connection?'
    )
  })

  it('should throw an error if the API call fails', async () => {
    const errorResponse = {
      response: {
        status: 500,
        data: { message: 'Internal Server Error' },
      },
    }
    mockOctokit.rest.repos.getLatestRelease.mockRejectedValue(errorResponse)

    await expect(fetchLatestVersionTag(owner, repo, token)).rejects.toThrow(
      'Getting latest release failed. Status: 500. Message: Internal Server Error'
    )
  })

  it('should throw an error if the response doesn\'t contain a tag name', async () => {
    mockOctokit.rest.repos.getLatestRelease.mockResolvedValue({
      data: { tag_name: '' },
    })

    await expect(fetchLatestVersionTag(owner, repo, token)).rejects.toThrow(
      'Response didn\'t contain tag name'
    )
  })
})

describe('downloadBinaryViaAPI', () => {
  const mockGetReleaseAsset = jest.fn()
  const mockGetReleaseAssetData = { id: 123, name: 'test-binary' }
  const mockOctokit = {
    rest: {
      repos: {
        getReleaseAsset: jest.fn(),
        getReleaseByTag: jest.fn(),
      },
    },
  }

  const version = 'v1.0.0'
  const binaryName = 'test-binary'
  const apiURL = 'https://api.github.com'
  const owner = 'test-owner'
  const repo = 'test-repo'
  const token = 'test-token'

  beforeEach(() => {
    jest.clearAllMocks();
    (retryableOctokitREST as jest.Mock).mockReturnValue(mockOctokit);
    (taskLib.getVariable as jest.Mock).mockReturnValue('temp/dir');
    (mockGetReleaseAsset as jest.Mock).mockResolvedValue(mockGetReleaseAssetData)
  })

  it('should download the binary and save it to a file', async () => {
    const mockResponseData = Buffer.from('mock-binary-data')
    mockOctokit.rest.repos.getReleaseAsset.mockResolvedValue({ data: mockResponseData })
    mockOctokit.rest.repos.getReleaseByTag.mockResolvedValue({ data: { assets: [mockGetReleaseAssetData] } })

    const result = await downloadBinaryViaAPI(version, binaryName, apiURL, owner, repo, token)

    expect(taskLib.debug).toHaveBeenCalledWith('downloading test-binary version v1.0.0 from test-owner/test-repo via API')
    expect(taskLib.debug).toHaveBeenCalledWith(expect.stringContaining('writing into downloaded binary as '))
    expect(result).toBe(path.join('temp', 'dir', 'mock-uuid'))
  })

  it('should throw an error if the API call to download fails', async () => {
    const errorResponse = {
      response: {
        status: 404,
        data: { message: 'Not Found' },
      },
    }
    mockOctokit.rest.repos.getReleaseAsset.mockRejectedValue(errorResponse)

    await expect(
      downloadBinaryViaAPI(version, binaryName, apiURL, owner, repo, token)
    ).rejects.toThrow('Could not download GitHub asset. Status: 404. Message: Not Found')
  })
})

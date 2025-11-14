import * as taskLib from 'azure-pipelines-task-lib/task'
import { isInternalStep } from './utils'
import { BINARY_NAME_VAR, OS_PIPER_OWNER, OS_PIPER_REPOSITORY, STEP_NAME } from './globals'

import crypto from 'crypto'

export class Telemetry {
  Utils: TelemetryUtils
  private StartTimeMS: number

  StepStartTime: string
  PipelineURLHash: string
  BuildURLHash: string
  StageName: string
  StepName: string
  ErrorCode: string
  ErrorCategory: string
  CorrelationID: string
  ErrorDetail: ErrorDetail
  StepDuration = ''
  PiperCommitHash = ''

  constructor () {
    this.Utils = new TelemetryUtils()
    this.StartTimeMS = new Date().getTime()

    this.StepStartTime = this.Utils.getTimeStamp(false)
    this.PipelineURLHash = this.Utils.getPipelineURLHash()
    this.BuildURLHash = this.Utils.getBuildURLHash()
    this.StageName = this.Utils.getStageName()
    this.StepName = this.Utils.getStepName()
    this.ErrorCode = '1'
    this.ErrorCategory = 'infrastructure'
    this.CorrelationID = this.Utils.getBuildURL()
    this.ErrorDetail = new ErrorDetail(this.ErrorCategory, this.CorrelationID, this.Utils.getLibrary(), this.StepName)
  }

  private setStepDuration () {
    this.StepDuration = (new Date().getTime() - this.StartTimeMS).toString()
  }

  private setPiperCommitHash () {
    this.PiperCommitHash = this.Utils.getPiperCommitHash()
  }

  setPostExecutionData (errorMessage: string) {
    this.setStepDuration()
    this.setPiperCommitHash()
    this.ErrorDetail.error = errorMessage
    this.ErrorDetail.time = this.Utils.getTimeStamp(true)
  }

  data () {
    const { Utils: utils, StartTimeMS: startTimeMS, ...data } = this
    return data
  }
}

class ErrorDetail {
  category: string
  correlationId: string
  error: string = ''
  library: string
  message: string
  result: string
  stepName: string
  time: string = ''

  constructor (category: string, correlationId: string, library: string, stepName: string) {
    this.category = category
    this.correlationId = correlationId
    this.library = library
    this.message = 'Azure Task execution failed'
    this.result = 'failure'
    this.stepName = stepName
  }
}

class TelemetryUtils {
  getTimeStamp (iso: boolean) {
    // data from Piper's telemetry.go is in nanoseconds, so we pad
    const date = new Date()
    if (iso) {
      const isoString = date.toISOString()
      return isoString.slice(0, -1) + '000000' + isoString.slice(isoString.length - 1)
    }

    const dd = String(date.getUTCDate()).padStart(2, '0')
    const mm = String(date.getUTCMonth() + 1).padStart(2, '0')
    const yyyy = date.getUTCFullYear()

    const hours = String(date.getUTCHours()).padStart(2, '0')
    const minutes = String(date.getUTCMinutes()).padStart(2, '0')
    const seconds = String(date.getUTCSeconds()).padStart(2, '0')
    const nanoseconds = String(date.getUTCMilliseconds()).padEnd(9, '0')

    const dateString = [yyyy, mm, dd].join('-')
    const timeString = [hours, minutes, seconds].join(':') + '.' + nanoseconds
    const utcString = '+0000 UTC'

    return [dateString, timeString, utcString].join(' ')
  }

  getPiperCommitHash () {
    const binaryName = taskLib.getVariable(BINARY_NAME_VAR) || ''
    if (!binaryName) {
      return ''
    }

    const versionExec = taskLib
      .tool(binaryName)
      .arg('version')
      .execSync({ silent: true })

    if (versionExec.error) {
      return ''
    }

    const versionOutput = versionExec.stdout
    const commitHashLine = versionOutput.split('\n')[1]
    const commitHashMatch = commitHashLine.match(/"([^"]+)"/)

    if (commitHashMatch && commitHashMatch.length > 1) {
      return commitHashMatch[1]
    }

    return ''
  }

  getPipelineURLHash () {
    const jobURL = this.getJobURL()
    return this.toSha1OrNA(jobURL)
  }

  getBuildURLHash () {
    const buildURL = this.getBuildURL()
    return this.toSha1OrNA(buildURL)
  }

  getJobURL () {
    return this.getEnvVar('SYSTEM_TEAMFOUNDATIONCOLLECTIONURI') + this.getEnvVar('SYSTEM_TEAMPROJECT') + '/' + this.getEnvVar('SYSTEM_DEFINITIONNAME') + '/_build?definitionId=' + this.getEnvVar('SYSTEM_DEFINITIONID')
  }

  getBuildURL () {
    return this.getEnvVar('SYSTEM_TEAMFOUNDATIONCOLLECTIONURI') + this.getEnvVar('SYSTEM_TEAMPROJECT') + '/' + this.getEnvVar('SYSTEM_DEFINITIONNAME') + '/_build/results?buildId=' + this.getAzureBuildID()
  }

  getAzureBuildID () {
    return this.getEnvVar('BUILD_BUILDID') || 'n/a'
  }

  getStageName () {
    return this.getEnvVar('SYSTEM_STAGEDISPLAYNAME') || 'n/a'
  }

  getStepName () {
    return STEP_NAME
  }

  getLibrary () {
    if (isInternalStep() && !taskLib.getVariable('hyperspace.sappiper.owner') && !taskLib.getVariable('hyperspace.sappiper.repository')) {
      return 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library.git'
    } else if (!isInternalStep()) {
      return [OS_PIPER_OWNER, OS_PIPER_REPOSITORY].join('/')
    }
    return ''
  }

  getEnvVar (key: string) {
    if (key in process.env) {
      return process.env[key]!
    }
    taskLib.debug(`telemetry data: environment variable ${key} does not exist`)
    return ''
  }

  toSha1OrNA (input: string) {
    if (!input) {
      return 'n/a'
    }
    return crypto.createHash('sha1').update(input).digest('hex')
  }
}

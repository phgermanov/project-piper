# Logging messages related to pulling docker images for running Piper steps

All below-mentioned log messages are available only when System Diagnostics is enabled on the pipeline (when Azure debug logs are printed)

## Scenarios

* System Trust is unavailable or there was an error obtaining system trust token. This means that docker will fall back to pulling from docker hub. Messages in logs:
  * `System trust token is not set`
  * `Artifactory credentials are empty, proceeding without docker authentication`
  * followed by pull logs described [below](#pull-logs)

* Normal flow, when system trust token is available and docker hub image must be pulled from Artifactory. Messages in logs:
  * `Running in a MS-hosted agent and image is from docker hub`
  * `docker authentication with System Trust successful`
  * followed by pull logs described [below](#pull-logs)

* Normal flow, but an error happens when retrieving Artifactory token from System trust. Messages:
  * `Running in a MS-hosted agent and image is from docker hub`
  * `Artifactory token request failed with status ${httpStatusCode}: ${responseBody}`
  * `Artifactory credentials are empty, proceeding without docker authentication`
  * followed by pull logs described [below](#pull-logs)

* Normal flow, but an error happens when authenticating with Artifactory token. Messages:
  * `Running in a MS-hosted agent and image is from docker hub`
  * An error message describing why docker login failed, followed by:
  * `System trust authentication failed, proceeding without docker authentication`
  * followed by pull logs described [below](#pull-logs)

* When step is not running in MS hosted agents you will see only one message
  * `Proceeding without docker authentication`
  * followed by pull logs described [below](#pull-logs)

  This means that Common repo is not used.

## Pull logs

* In all above scenarios pulls will be retried maximum 3 times on failures. In case of artifactory pulls, Piper task will retry 3 times on Common repo and fallback to pulling from docker hub, also with 3 times retry.
  * Message of successful pull (imageName will contain registry name too):
    * `Image ${imageName} pulled successfully`
  * Messages of pull retries:
    * `Failed to pull image ${imageName}: ${errorMsg}`
    * `Retrying... (${attemptNum}/3)`           -> for example, `Retrying... (2/3)`
    * `Failed to pull image ${imageName} after ${numOfAttempts} attempts: ${errorMsg}`
  * Message when fallback to docker hub happens after 3 retries on Artifactory:
    * `Falling back to pull from docker hub`

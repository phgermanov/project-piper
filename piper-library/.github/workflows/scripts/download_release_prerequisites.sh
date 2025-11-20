#!/usr/bin/env bash

resourcesBaseURL="https://github.tools.sap/api/v3/repos/project-piper/resources/contents"

githubToolsToken=""
if [[ -z "${PIPER_VAULTCREDENTIAL_GITHUB_TOOLS_TOKEN}" ]]; then
  echo "GitHub tools token is not set in PIPER_VAULTCREDENTIAL_GITHUB_TOOLS_TOKEN"
  exit 1
else
  githubToolsToken="${PIPER_VAULTCREDENTIAL_GITHUB_TOOLS_TOKEN}"
fi

declare -a fileNames=(
  "stageconfig/piper-stage-config.yml"
  "gen/piper-defaults-azure-tools.yml"
  "gen/piper-defaults-github-tools.yml"
  "gen/piper-defaults-github-wdf.yml"
  "gen/piper-defaults-jenkins-wdf.yml"
)

for i in "${fileNames[@]}"
do
  # remove folder and slash prefix
  outputName=${i##*/}

  status_code=$(curl -L -o "$outputName" --write-out %{http_code} \
    -H "Accept: application/vnd.github.raw+json" \
    -H "Authorization: Bearer $githubToolsToken" \
    $resourcesBaseURL/$i)

  if [[ "$status_code" -ne 200 ]] ; then
    echo "Failed to download $outputName with code $status_code"
    cat "$outputName"
    exit 1
  fi
done

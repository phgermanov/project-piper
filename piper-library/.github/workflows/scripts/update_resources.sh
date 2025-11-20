#!/usr/bin/env bash

githubWDFUser=github-wdf-piper-serviceuser
githubWDFToken=""
if [[ -z "${PIPER_VAULTCREDENTIAL_GITHUB_WDF_TOKEN}" ]]; then
  echo "GitHub WDF token is not set in PIPER_VAULTCREDENTIAL_GITHUB_WDF_TOKEN"
  exit 1
else
  githubWDFToken="${PIPER_VAULTCREDENTIAL_GITHUB_WDF_TOKEN}"
fi

cp piper-defaults-jenkins.yml resources/piper-defaults.yml
cp piper-stage-config.yml resources/piper-stage-config.yml

declare -a fileNames=(
  "resources/piper-defaults.yml"
  "resources/piper-stage-config.yml"
)
hasChanges=0
for i in "${fileNames[@]}"
do
  if git diff --quiet --exit-code "$i"; then
    echo "$i has not been changed"
  else
    echo "found changes in $i"
    git add "$i"
    hasChanges=1
  fi
done

if [[ "$hasChanges" == 1 ]]; then
  git config user.name ${githubWDFUser}
  git config user.email dl_6287ae4dec3ca802990e86e5@global.corp.sap
  
  basicAuth=$(echo -n "${githubWDFUser}:${githubWDFToken}" | base64 -w0)
  git config http.https://github.wdf.sap.corp/.extraheader "AUTHORIZATION: basic ${basicAuth}"

  git commit -m "Update resources from release pipeline"
  git push
fi

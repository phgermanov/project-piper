#!/usr/bin/env bash

# Exposes variables passed as arguments to $GITHUB_ENV using shell param expansion
# See: https://www.gnu.org/software/bash/manual/html_node/Shell-Parameter-Expansion.html#Shell-Parameter-Expansion
for vaultSecretName in "$@"
do
  envVarName="PIPER_VAULTCREDENTIAL_$vaultSecretName"
  echo "Exposing $envVarName to \$GITHUB_ENV"
  echo "$envVarName=${!envVarName}" >> "$GITHUB_ENV"
done

#!/bin/bash

# This script is used to rotate the password for the staging service.
# https://github.wdf.sap.corp/pages/Repository-services/staging-service/howto/rotate_user_password/

# Take the credentials from vault (PIPELINE-GROUP-6133/GROUP-SECRETS/staging-service)
# {
#   "password": "...",
#   "tenantId": "ocmpiper",
#   "tenantSecret": "...",
#   "username": "ocmpiper"
# }

# ... store them in a file and provide the path as an argument.
# Check if the script received the required argument
if [[ "$#" -ne 1 ]]; then
  echo "Usage: $0 <path-to-credentials-json>"
  exit 1
fi

# Take the credentials file path from the first argument
credentialsJson=$(realpath "$1")

# First read the current credentials from json file
username=$(jq -r '.username' "${credentialsJson}")
password=$(jq -r '.password' "${credentialsJson}")
tenantId=$(jq -r '.tenantId' "${credentialsJson}")
tenantSecret=$(jq -r '.tenantSecret' "${credentialsJson}")

# Get a new JWT token
access_token=$(curl -sSL -u ${tenantId}:${tenantSecret} -X POST https://staging.repositories.cloud.sap/stage/api/login -d "grant_type=password&username=${username}&password=${password}" | jq -r .access_token)
echo "JWT Token: ${access_token}"

# Get the new password
newPassword=$(curl -sSL -X POST -H "Authorization: Bearer ${access_token}" https://staging.repositories.cloud.sap/stage/api/admin/user/rotateownpassword | jq -r .password)
echo "New Password: ${newPassword}"

# Update the json file with the new password
jq --arg newPassword "${newPassword}" '.password = $newPassword' ${credentialsJson} > tmp.json && mv tmp.json ${credentialsJson}

# Now you can create a new version for credentials in Vault

# We keep two backups in Vault, to be able to switch easier between the different tenants.
# staging-service-hyperspace
# staging-service-ocmpiper

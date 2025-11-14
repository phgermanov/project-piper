#!/bin/bash

# This script retrieves a session token from the system trust API and sets it as an environment variable
# Usage: ./set-system-trust-token.sh <url> <runtime_header> <secret_id> <pipeline_id> <group_id>

url=$1
runtime_header=$2
secret_id=$3
pipeline_id=$4
group_id=$5

# Function to check if a parameter is empty
check_param() {
    if [[ -z "$1" ]]; then
        echo "Error: Missing parameter - $2"
        exit 1
    fi
}

# Check that none of the parameters are empty
check_param "$url" "URL"
check_param "$runtime_header" "Runtime Header"
check_param "$secret_id" "Secret ID"
check_param "$pipeline_id" "Pipeline ID"
check_param "$group_id" "Group ID"

# Perform the API request to get the session token
echo "Retrieving system trust token from $url"
response=$(curl --retry 3 --retry-delay 5 --retry-connrefused --retry-max-time 15 --silent --show-error --write-out "HTTPSTATUS:%{http_code}" \
    -X POST ${url} \
    --header "X-Trust-Authorization-Runtime: ${runtime_header}" \
    --header "Content-Type: application/json" \
    --data "{\"secret_id\":\"${secret_id}\", \"pipeline_id\": \"${pipeline_id}\", \"group_id\": \"${group_id}\"}")

# Extract the HTTP status code from the response
http_status=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
echo "HTTP status: $http_status"

# Extract the token from the response
token=$(echo $response | sed -e 's/HTTPSTATUS\:.*//g' | jq -r '.token')

# Check if the token was retrieved successfully
if [ "$http_status" -ne 200 ] || [ -z "$token" ]; then
    echo "Failed to retrieve system trust token. Will continue with secrets defined in Vault or Jenkins credentials. \
          Consider using System Trust integration for credentialless access as described here: \
          https://github.wdf.sap.corp/pages/ContinuousDelivery/piper-doc/news/2024/11/14/credentialless-usage-of-cumulus-in-pipeline-with-hyperspace-system-trust/ \
          https://github.wdf.sap.corp/pages/ContinuousDelivery/piper-doc/news/2024/10/17/credentialless-usage-of-sonar-in-pipeline-with-hyperspace-system-trust/"
    echo "response: $response"
    echo "http_status: $http_status"
    exit 1
fi

# Mask the token in the logs
echo "System trust token retrieved successfully"
echo "##vso[task.setvariable variable=systemTrustSessionToken;isOutput=true;issecret=true]$token"

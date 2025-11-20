#!/bin/bash

# This script makes a request to a Trust API with provided parameters
# Usage: ./trust_api_request.sh <url> <secret_id> <pipeline_id> <group_id> <auth_token>

url=$1
secret_id=$2
pipeline_id=$3
group_id=$4
auth_token=$5

# Function to check if a parameter is empty
check_param() {
    if [[ -z "$1" ]]; then
        echo "Error: Missing parameter - $2"
        exit 1
    fi
}

# Check that none of the parameters are empty
check_param "$url" "URL"
check_param "$secret_id" "Secret ID"
check_param "$pipeline_id" "Pipeline ID"
check_param "$group_id" "Group ID"
check_param "$auth_token" "Authorization Token"

# Perform the API request
curl --retry 3 --retry-delay 5 --retry-connrefused --retry-max-time 15 --silent --show-error --write-out "HTTPSTATUS:%{http_code}" \
    -X POST -d "{\"secret_id\":\"${secret_id}\", \"pipeline_id\": \"${pipeline_id}\", \"group_id\": \"${group_id}\"}" ${url} \
    --header "X-Trust-Authorization-Runtime: ${auth_token}"

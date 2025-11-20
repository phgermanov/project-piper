#!/bin/bash

# Script to upload assets to a GitHub release
# Usage: upload_release_assets.sh <api_url> <owner> <repo> <release_tag> <token> <assets...>
#
# Parameters:
#   api_url: GitHub API URL (e.g., https://api.github.com or https://github.tools.sap/api/v3)
#   owner: Repository owner
#   repo: Repository name
#   release_tag: Release tag to find/create
#   token: GitHub API token
#   assets: Space-separated list of asset files to upload
#
# Environment variables:
#   CREATE_DRAFT: Set to "true" to create a draft release if not found (default: false)
#   RELEASE_NAME: Custom release name (default: same as tag)
#   RELEASE_BODY: Custom release body (default: "Release <tag>")
#   TARGET_COMMITISH: Target branch/commit (default: "master")

set -euo pipefail

# Check required parameters
if [ $# -lt 5 ]; then
    echo "Usage: $0 <api_url> <owner> <repo> <release_tag> <token> <assets...>"
    echo "Example: $0 https://api.github.com myorg myrepo v1.0.0 \$GITHUB_TOKEN file1.txt file2.bin"
    exit 1
fi

API_URL="$1"
OWNER="$2"
REPO="$3"
RELEASE_TAG="$4"
TOKEN="$5"
shift 5
ASSETS=("$@")

# Default values
CREATE_DRAFT="${CREATE_DRAFT:-true}"
RELEASE_NAME="${RELEASE_NAME:-$RELEASE_TAG}"
RELEASE_BODY="${RELEASE_BODY:-Release $RELEASE_TAG}"
TARGET_COMMITISH="${TARGET_COMMITISH:-master}"

echo "=== Upload Release Assets Script ==="
echo "Uploading assets to release $RELEASE_TAG in $OWNER/$REPO"
echo "API URL: $API_URL"
echo "Assets: ${ASSETS[*]}"
echo "CREATE_DRAFT: $CREATE_DRAFT"
echo "TARGET_COMMITISH: $TARGET_COMMITISH"
echo "===================================="

# Function to validate JSON
validate_json() {
    local response="$1"
    if echo "$response" | jq empty > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Function to handle API error responses
handle_api_error() {
    local response="$1"
    local operation="$2"

    if validate_json "$response"; then
        if echo "$response" | jq -e '.message' > /dev/null 2>&1; then
            echo "API Error during $operation: $(echo "$response" | jq -r '.message')" >&2
            if echo "$response" | jq -e '.errors' > /dev/null 2>&1; then
                echo "Details: $(echo "$response" | jq -c '.errors')" >&2
            fi
        else
            echo "Response during $operation: $response" >&2
        fi
    else
        echo "Error: Invalid JSON response during $operation" >&2
        echo "Raw response (first 500 chars): ${response:0:500}" >&2
    fi
}

# Function to find release by tag
find_release() {
    local tag="$1"
    local response
    local http_code

    response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
         -H "Accept: application/vnd.github+json" \
         -H "Authorization: Bearer $TOKEN" \
         "$API_URL/repos/$OWNER/$REPO/releases")

    http_code=$(echo "$response" | tail -1 | cut -d: -f2)
    response=$(echo "$response" | sed '$d')

    # Debug: Show first 500 chars of response if it's not JSON
    if ! validate_json "$response"; then
        echo "Debug: Response is not valid JSON in find_release. First 500 chars:" >&2
        echo "${response:0:500}" >&2
        echo "" >&2
        echo "HTTP Code: $http_code" >&2
        return 1
    fi

    if [ "$http_code" != "200" ]; then
        echo "Error: Failed to fetch releases (HTTP $http_code)" >&2
        handle_api_error "$response" "fetching releases"
        # Return empty string instead of failing
        echo ""
        return 0
    fi

    # Safely parse JSON and find release ID
    if validate_json "$response"; then
        echo "$response" | jq -r ".[] | select(.tag_name == \"$tag\") | .id" | head -1
    else
        echo ""
    fi
}

# Function to find draft release
find_draft_release() {
    local response
    local http_code

    response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
         -H "Accept: application/vnd.github+json" \
         -H "Authorization: Bearer $TOKEN" \
         "$API_URL/repos/$OWNER/$REPO/releases")

    http_code=$(echo "$response" | tail -1 | cut -d: -f2)
    response=$(echo "$response" | sed '$d')

    if [ "$http_code" != "200" ]; then
        echo "Error: Failed to fetch releases (HTTP $http_code)" >&2
        handle_api_error "$response" "fetching releases"
        # Return empty string instead of failing
        echo ""
        return 0
    fi

    # Safely parse JSON and find draft release ID
    if validate_json "$response"; then
        echo "$response" | jq -r '.[] | select(.draft == true) | .id' | head -1
    else
        echo ""
    fi
}

# Function to create draft release
create_draft_release() {
    local tag="$1"
    local name="$2"
    local body="$3"

    local response
    local http_code

    echo "Creating draft release $tag" >&2
    response=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST \
         -H "Accept: application/vnd.github+json" \
         -H "Authorization: Bearer $TOKEN" \
         -H "Content-Type: application/json" \
         "$API_URL/repos/$OWNER/$REPO/releases" \
         -d "{
           \"tag_name\": \"$tag\",
           \"target_commitish\": \"$TARGET_COMMITISH\",
           \"name\": \"$name\",
           \"body\": \"$body\",
           \"draft\": true
         }")

    http_code=$(echo "$response" | tail -1 | cut -d: -f2)
    response=$(echo "$response" | sed '$d')

    if [ "$http_code" != "201" ] && [ "$http_code" != "200" ]; then
        echo "Error: Failed to create release (HTTP $http_code)" >&2
        handle_api_error "$response" "creating release"
        return 1
    fi

    echo "$response"
}

# Function to get upload URL for release
get_upload_url() {
    local release_id="$1"
    local response
    
    response=$(curl -s -H "Accept: application/vnd.github+json" \
         -H "Authorization: Bearer $TOKEN" \
         "$API_URL/repos/$OWNER/$REPO/releases/$release_id")
    
    if validate_json "$response"; then
        echo "$response" | jq -r '.upload_url' | sed 's/{?name,label}//'
    else
        echo "Error: Invalid JSON when getting upload URL" >&2
        echo "Response (first 500 chars): ${response:0:500}" >&2
        echo ""
    fi
}

# Function to upload asset
upload_asset() {
    local upload_url="$1"
    local asset_file="$2"

    if [ ! -f "$asset_file" ]; then
        echo "Warning: $asset_file not found, skipping"
        return 0
    fi

    echo "Uploading $asset_file..." >&2
    local response
    response=$(curl -s -X POST \
         -H "Accept: application/vnd.github+json" \
         -H "Authorization: Bearer $TOKEN" \
         -H "Content-Type: application/octet-stream" \
         --data-binary "@$asset_file" \
         "$upload_url?name=$(basename "$asset_file")")

    if validate_json "$response" && echo "$response" | jq -e '.id' > /dev/null 2>&1; then
        echo "  ✓ Successfully uploaded $asset_file"
        return 0
    else
        echo "  ⚠ Failed to upload $asset_file: $response"
        return 1
    fi
}

# Main logic
RELEASE_ID=""

# Try to find existing release by tag
RELEASE_ID=$(find_release "$RELEASE_TAG")

if [ -z "$RELEASE_ID" ] || [ "$RELEASE_ID" == "null" ]; then
    if [ "$CREATE_DRAFT" == "true" ]; then
        # Create new draft release
        RELEASE_RESPONSE=$(create_draft_release "$RELEASE_TAG" "$RELEASE_NAME" "$RELEASE_BODY")
        
        if validate_json "$RELEASE_RESPONSE"; then
            RELEASE_ID=$(echo "$RELEASE_RESPONSE" | jq -r '.id')
        else
            echo "Error: Invalid JSON response when creating draft release" >&2
            echo "Response (first 500 chars): ${RELEASE_RESPONSE:0:500}" >&2
            RELEASE_ID=""
        fi

        if [ -z "$RELEASE_ID" ] || [ "$RELEASE_ID" == "null" ]; then
            echo "Error: Failed to create draft release"
            echo "Response: $RELEASE_RESPONSE"
            exit 1
        fi
        echo "Created draft release with ID: $RELEASE_ID" >&2
    else
        # Try to find any draft release if tag-specific release not found
        RELEASE_ID=$(find_draft_release)

        if [ -z "$RELEASE_ID" ] || [ "$RELEASE_ID" == "null" ]; then
            echo "Error: No release found with tag $RELEASE_TAG and CREATE_DRAFT is not enabled"
            exit 1
        fi
        echo "Using draft release with ID: $RELEASE_ID" >&2
    fi
else
    echo "Found existing release with ID: $RELEASE_ID" >&2
fi

# Get upload URL
UPLOAD_URL=$(get_upload_url "$RELEASE_ID")

if [ -z "$UPLOAD_URL" ] || [ "$UPLOAD_URL" == "null" ]; then
    echo "Error: Failed to get upload URL for release $RELEASE_ID"
    exit 1
fi

echo "Upload URL: $UPLOAD_URL" >&2

# Upload all assets
UPLOAD_FAILURES=0
for asset in "${ASSETS[@]}"; do
    if ! upload_asset "$UPLOAD_URL" "$asset"; then
        ((UPLOAD_FAILURES++))
    fi
done

# Output release ID for further processing
echo "release_id=$RELEASE_ID"

if [ $UPLOAD_FAILURES -gt 0 ]; then
    echo "Warning: $UPLOAD_FAILURES asset(s) failed to upload"
    exit 1
fi

echo "All assets uploaded successfully" >&2

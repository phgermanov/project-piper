# shellcheck disable=SC2086

github_api_url=$1
owner=$2
repository=$3

# call_github_api will make request and exits with code 1 in case of request failure
# args:
#   1 - URL of the api resource
#   2 - 'Accept' header value
#   3 - output file name
function call_github_api () {
  url=$1
  accept_header=$2
  outputName=$3

  echo -e "curl -L -s -o $outputName -H \"Authorization: Bearer <token>\" -H \"Accept: $accept_header\" $url \n"
  status_code=$(curl --write-out %{http_code} -L -s -o "$outputName" \
              -H "Authorization: Bearer ${GH_TOKEN}" \
              -H "Accept: $accept_header" \
              $url)
  if [[ "$status_code" -ne 200 ]] ; then
    echo "request to $url failed with code $status_code"
    cat $outputName
    exit 1
  fi
}

release_response_json="release_response.json"
call_github_api "$github_api_url/repos/$owner/$repository/releases" "application/vnd.github.raw+json" "$release_response_json"

# assets to be downloaded from release
declare -a assetNames=(
  "sap-piper"
  "piper-stage-config.yml"
  "piper-defaults-jenkins.yml"
  "piper-defaults-azure.yml"
  "piper-defaults-github.yml"
)

# For each asset, search for url in release_obj and download it
release_obj=$(jq 'first(.[] | select(.draft == true))' $release_response_json)
for asset_name in "${assetNames[@]}"; do
  asset_url=$(jq ".assets[] | select(.name == \"$asset_name\").url" <<< "$release_obj" | tr -d '"' )

  if [ -z "${asset_url}" ]; then continue; fi  # File is not part of the release -> skip download

  call_github_api ${asset_url} "application/octet-stream" "$asset_name"
done

echo "release_id=$(jq .id <<< "$release_obj")" >> "$GITHUB_OUTPUT"

#!/bin/bash
# This script is a cut-down version of https://github.com/reecetech/version-increment
# and works only with semver versions like 'v1.2.3' or '1.2.3'
# it simply increments minor version
# Environment variable CURRENT_VERSION must be set.

set -euo pipefail
export LC_ALL=C.UTF-8

## SemVer regex
pcre_master_ver='^v{0,1}(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)$'

echo "ℹ️ envvar current_version=${current_version}"

errors='false'
if [[ -z ${current_version:-} ]] ; then
    echo "🛑 'current_version' is unset or empty" 1>&2
    errors='true'
elif [[ -z "$(echo "${current_version}" | grep -P "${pcre_master_ver}")" ]] ; then
    echo "🛑 'current_version' is not a valid normal version (M.m.p)" 1>&2
    errors='true'
fi

if [[ "${errors}" == 'true' ]] ; then
    exit 8
fi

# clear v prefix from input version
# shellcheck disable=SC2001
version=$(echo "${current_version}" | sed 's/^v//g')
echo "ℹ️ The current normal version is ${version}"


# Parse current version into version_array
IFS=" " read -r -a version_array <<< "${version//./ }"

# increment patch version:  1.2.3 -> 1.2.4
#(( ++version_array[2] ))

# increment minor version and set patch version to 0:  1.2.3 -> 1.3.0
(( ++version_array[1] ))
version_array[2]='0'

# increment major version, set minor and patch versions to 0:  1.2.3 -> 2.0.0
#(( ++version_array[0] ))
#version_array[1]='0'
#version_array[2]='0'

new_version="${version_array[0]}.${version_array[1]}.${version_array[2]}"

if [[ -z ${new_version} ]] ; then
    echo "🛑 Version incrementing has failed to produce a semver compliant version" 1>&2
    echo "ℹ️ Failed version string: '${new_version}'" 1>&2
    exit 12
fi

echo "ℹ️ The new version is ${new_version}"

# shellcheck disable=SC2129
echo "VERSION=${new_version}" >> "${GITHUB_OUTPUT}"
echo "V_VERSION=v${new_version}" >> "${GITHUB_OUTPUT}"

#!/bin/bash

# Copyright 2023 Linka Cloud  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

USER="{{ .User }}"
PASSWORD="{{ .Password }}"

SCHEME="{{ .Scheme }}"
REPO_HOST="{{ .Host }}"
REPO_PATH="{{ .Path }}"

REPO_NAME="$(basename "${REPO_PATH}")"

already_exists() {
    echo "Repository already configured."
    echo "Use --force to overwrite."
}

[ -n "${DEBUG}" ] && set -x

if [ -n "${USER}" ]; then
    REPO_URL="${SCHEME}://${USER}:${PASSWORD}@${REPO_HOST}${REPO_PATH}"
else
    REPO_URL="${SCHEME}://${REPO_HOST}${REPO_PATH}"
fi

if [ "$UID" -ne 0 ]; then
    echo "Please run as root or sudo"
    exit 1
fi

if [ -f "/etc/yum.repos.d/${REPO_NAME}.repo" ] && [ "$1" != "--force" ]; then
    echo "Repository already configured."
    echo "Use --force to overwrite."
    exit 1
fi

if which dnf >/dev/null 2>&1; then
    PROG=dnf
else
    PROG=yum
fi
if dnf config-manager --help >/dev/null 2>&1; then
    HAS_CONFIG_MANAGER=1
fi

if [ "${PROG}" = "dnf" ] && [ -n "${HAS_CONFIG_MANAGER}" ]; then
    if ! which curl >/dev/null 2>&1; then
        echo "curl is required to setup the repository."
        exit 1
    fi
    curl -s "${REPO_URL}.repo" -o "/etc/yum.repos.d/${REPO_NAME}.repo"
else
    dnf config-manager --add-repo "${REPO_URL}.repo"
fi

echo "yum setup complete."
echo "You can now run 'yum update' to update the package list."

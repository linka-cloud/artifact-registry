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
REPO_NAME="{{ .Name }}"

DIST="{{ .Dist }}"
COMPONENT="{{ .Component }}"

[ -n "${DEBUG}" ] && set -x

if [ "$(id -u)" -ne 0 ]; then
    echo "Please run as root or sudo"
    exit 1
fi

if [ -f "/etc/apt/sources.list.d/${REPO_NAME}.list" ] && [ "$1" != "--force" ]; then
    echo "Repository already configured."
    echo "Use --force to overwrite."
    exit 1
fi

if ! which curl >/dev/null 2>&1; then
    echo "curl is required to setup the repository."
    exit 1
fi

REPO="${SCHEME}://${REPO_HOST}${REPO_PATH}"

if [ -n "${USER}" ]; then
    REPO_AUTH="${SCHEME}://${USER}:${PASSWORD}@${REPO_HOST}${REPO_PATH}"
    echo "machine ${REPO} login $USER password $PASSWORD" > "/etc/apt/auth.conf.d/${REPO_NAME}.conf"
else
    REPO_AUTH="${SCHEME}://${REPO_HOST}${REPO_PATH}"
fi

curl -s "${REPO_AUTH}/repository.key" -o "/etc/apt/trusted.gpg.d/${REPO_NAME}.asc"
echo "deb ${REPO} ${DIST} ${COMPONENT}" > "/etc/apt/sources.list.d/${REPO_NAME}.list"

echo "deb repository setup complete."
echo "You can now run 'apt update' to update the package list."


#!/bin/sh

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

USER="{{ .User }}"
PASSWORD="{{ .Password }}"

SCHEME="{{ .Scheme }}"
REPO_HOST="{{ .Host }}"
REPO_PATH="{{ .Path }}"

BRANCH="{{ .Branch }}"
REPOSITORY="{{ .Repository }}"

set -e

[ -n "${DEBUG}" ] && set -x
[ "$1" = "--force" ] && FORCE=1

if [ -n "${USER}" ]; then
    REPO_URL="${SCHEME}://${USER}:${PASSWORD}@${REPO_HOST}${REPO_PATH}/${BRANCH}/${REPOSITORY}"
else
    REPO_URL="${SCHEME}://${REPO_HOST}${REPO_PATH}/${BRANCH}/${REPOSITORY}"
fi

if [ "$(id -u)" -ne 0 ]; then
    echo "Please run as root or sudo"
    exit 1
fi

if grep "${REPO_HOST}${REPO_PATH}/${BRANCH}/${REPOSITORY}" /etc/apk/repositories >/dev/null 2>&1 && [ -z "${FORCE}" ]; then
    echo "Repository already configured."
    echo "Use --force to overwrite."
    exit 1
fi

if ! which curl >/dev/null 2>&1; then
    echo "curl is required to setup the repository."
    exit 1
fi

PATTERN="$(echo "${REPO_HOST}${REPO_PATH}/${BRANCH}/${REPOSITORY}"|sed 's/\//\\\//g')"
sed -i "/${PATTERN}/d" /etc/apk/repositories

if [ -n "${FORCE}" ]; then
    NAME="$(curl -sIX GET "${REPO_URL}"/key|grep "Content-Disposition:"|cut -d'=' -f2|sed 's/\r$//')"
    if [ -n "${NAME}" ]; then
        rm -f "/etc/apk/keys/${NAME}"
    fi
fi

(cd /etc/apk/keys && curl -sJO "${REPO_URL}/key")

echo "${REPO_URL}" >> /etc/apk/repositories

echo "apk repository setup complete."
echo "You can now run 'apk update' to update the package list or install packages directly using 'apk add --no-cache <package>'."

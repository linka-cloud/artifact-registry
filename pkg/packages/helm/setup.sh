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

[ "$1" = "--force" ] && FORCE=1

ARGS="repo add ${REPO_NAME} ${SCHEME}://${REPO_HOST}/${REPO_PATH}"

if ! which helm >/dev/null 2>&1; then
    echo "helm is required to setup the repository."
    exit 1
fi

if [[ -n "${USER}" ]]; then
    ARGS="${ARGS} --username ${USER}"
fi

if [[ -n "${PASSWORD}" ]]; then
    ARGS="${ARGS} --password ${PASSWORD}"
fi

if [[ -n "${FORCE}" ]]; then
    ARGS="${ARGS} --force-update"
fi

helm ${ARGS}

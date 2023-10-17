// Copyright 2023 Linka Cloud  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { RepositoryType } from '../api/repository'
import { Credentials } from '../api/schemas/login'

const typeUrl = (type: RepositoryType) => window.location.host.split('.')[0] === type.toString() ? window.location.host : `${window.location.host}/${type}`

export const lkar = {
  login: (repo?: string, creds?: Credentials) => `lkar login -u ${creds?.user ?? '$USER'} -p ${creds?.password ?? '$PASSWORD'} ${window.location.protocol === 'http:' ? '--plain-http ' : ''}${window.location.host}${repo ? '/' + repo : ''}`,
  setup: (type: RepositoryType, repo?: string, sub?: string, _?: Credentials) => `lkar ${type} setup ${window.location.protocol === 'http:' ? '--plain-http ' : ''}${repo ? `${window.location.host}/${repo}` : window.location.host} ${(sub ? `${sub.split('/')[0]} ${sub.split('/')[1]}` : '')}`,
  push: (type: RepositoryType, repo?: string, sub?: string, _?: Credentials) => `lkar ${type} push ${window.location.protocol === 'http:' ? '--plain-http ' : ''}${repo ? `${window.location.host}/${repo}` : window.location.host} ${(sub ? `${sub.split('/')[0]} ${sub.split('/')[1]} ` : '')}# my-package.${type}`,
  delete: (type: RepositoryType, filePath: string, repo?: string, _?: Credentials) => `lkar ${type} delete ${window.location.protocol === 'http:' ? '--plain-http ' : ''}${repo ? `${window.location.host}/${repo}` : window.location.host} ${filePath}`,
}

export const curl = {
  setup: (type: RepositoryType, repo?: string, sub?: string, creds?: Credentials) => `curl --user "${creds?.user ?? '$USER'}:${creds?.password ?? '$PASSWORD'}" ${window.location.protocol}//${typeUrl(type)}${repo ? '/' + repo : '' + (sub ? '/' + sub : '')}/setup | sudo sh`,
  push: (type: RepositoryType, repo?: string, sub?: string, creds?: Credentials) => `curl --user "${creds?.user ?? '$USER'}:${creds?.password ?? '$PASSWORD'}" ${window.location.protocol}//${typeUrl(type)}${repo ? '/' + repo : '' + (sub ? '/' + sub : '')}/push --upload-file # my-package.${type}`,
  delete: (type: RepositoryType, filePath: string, repo?: string, creds?: Credentials) => `curl --user "${creds?.user ?? '$USER'}:${creds?.password ?? '$PASSWORD'}" -X DELETE ${window.location.protocol}//${typeUrl(type)}${repo ? '/' + repo : ''}/${filePath}`,
}

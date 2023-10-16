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

import { AsyncVoid } from '../utils'
import { APKPackage } from './apk'
import { DEBPackage } from './deb'
import { makePackage, Package, Repository, RepositoryType } from './repository'
import { RPMPackage } from './rpm'
import { Credentials } from './schemas/login'

export interface API {
  login: (user: string, password: string, repo?: string) => Promise<[boolean, Error?]>
  logout: () => Promise<void>
  credentials: () => Promise<[Partial<Credentials>, Error?]>

  repositories: (repo?: string) => Promise<[Repository[], Error?]>
  packages: (type: RepositoryType, repo?: string) => Promise<[Package[], Error?]>
}


export const api: API = {
  login: async (user: string, password: string, repo: string | undefined = '') => {
    try {
      const res = await fetch(repo ? `/_auth/${repo}/login` : '/_auth/login', { headers: { 'Authorization': `Basic ${btoa(`${user}:${password}`)}` } })
      if (!res.ok) {
        return [false, new Error(res.statusText)]
      }
      return [true]
    } catch (e) {
      return [false, e as Error]
    }
  },
  logout: async () => fetch(`/_auth/logout`, { method: 'POST' }).then(AsyncVoid),

  credentials: async () => {
    const res = await fetch('/_auth/credentials')
    if (!res.ok) {
      return [{}, new Error(res.statusText)]
    }
    return [await res.json()] as [Credentials, Error?]
  },

  repositories: async (repo: string | undefined = '') => {
    const res = await fetch(`/_repositories/${repo}`)
    if (!res.ok) {
      return [[], new Error(res.statusText)]
    }
    return res.json()
      .then(d => [d.map((d: any) => ({ ...d, lastUpdated: new Date(d.lastUpdated) }))
        .sort((a: Repository, b: Repository) => a.type.localeCompare(b.type))
        .sort((a: Repository, b: Repository) => a.name?.localeCompare(b.name??''))] as [Repository[], Error?])
      .catch(e => [[], e] as [Repository[], Error])
  },
  packages: async (type: RepositoryType, repo: string | undefined = '') => {
    const url = window.location.host.split('.')[0] === type.toString() ? `/_packages/${repo}` : `/_packages/${type}/${repo}`
    const res = await fetch(url)
    if (!res.ok) {
      return [[], new Error(res.statusText)]
    }
    return res.json()
      .then(d => [d.map((d: APKPackage | DEBPackage | RPMPackage) => makePackage(type, d)).sort((a: Package, b: Package) => a.name.localeCompare(b.name))] as [Package[], Error?])
      .catch(e => [[], e] as [Package[], Error])
  },
}

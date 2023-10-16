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

import { APKPackage, fromAPK } from './apk'
import { DEBPackage, fromDEB } from './deb'
import { fromHelm, HelmPackage } from './helm'
import { fromRPM, RPMPackage } from './rpm'

export interface Stats {
  count: number
  size: number
}

export interface Repository {
  name?: string
  type: RepositoryType
  size: number
  lastUpdated: Date
  metadata: Stats
  packages: Stats
}

export enum RepositoryType {
  APK = 'apk',
  DEB = 'deb',
  RPM = 'rpm',
  HELM = 'helm',
}

export interface Package {
  type: RepositoryType
  name: string
  size: number
  version: string
  architecture: string
  license?: string
  description: string
  summary?: string
  projectURL: string
  lastUpdated?: Date
  filePath: string
}

export const makePackage = (type: RepositoryType, d: APKPackage | DEBPackage | RPMPackage | HelmPackage) => {
  switch (type) {
    case RepositoryType.APK:
      return fromAPK(d as APKPackage)
    case RepositoryType.DEB:
      return fromDEB(d as DEBPackage)
    case RepositoryType.RPM:
      return fromRPM(d as RPMPackage)
    case RepositoryType.HELM:
      return fromHelm(d as HelmPackage)
  }
}

export const subRepositories = (packages: Package[], type: RepositoryType) => type !== RepositoryType.RPM && type !== RepositoryType.HELM
  ? packages.map(p => p.filePath.replace('pool/', '').split('/').slice(0, 2).join('/')).filter((p, i, arr) => arr.indexOf(p) === i)
  : []

export const subRepositoryPackages = (packages: Package[], type: RepositoryType, sub: string) => packages.filter(p => sub === '' || (type === RepositoryType.DEB ? p.filePath.startsWith(`pool/${sub}`) : p.filePath.startsWith(sub)))

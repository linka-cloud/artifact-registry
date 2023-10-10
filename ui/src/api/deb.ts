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

import { Package, RepositoryType } from './repository'

export const fromDEB = (deb: DEBPackage): Package => ({
  type: RepositoryType.DEB,
  name: deb.name,
  architecture: deb.architecture,
  size: deb.size,
  version: deb.version,
  description: deb.metadata.description,
  projectURL: deb.metadata.projectURL,
  filePath: deb.filePath,
})

export interface DEBPackage {
  name: string
  version: string
  size: number
  architecture: string
  control: string
  metadata: Metadata
  component: string
  distribution: string
  md5: string
  sha1: string
  sha256: string
  sha512: string
  filePath: string
}

export interface Metadata {
  maintainer: string
  projectURL: string
  description: string
  dependencies: string[]
}

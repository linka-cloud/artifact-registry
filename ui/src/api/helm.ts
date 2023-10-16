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

export const fromHelm = (chart: HelmPackage): Package => ({
  type: RepositoryType.HELM,
  name: chart.name,
  architecture: "noarch",
  size: chart.size,
  version: chart.version,
  description: chart.description,
  projectURL: chart.home || '',
  filePath: chart.filePath,
})

export interface HelmPackage {
  $type: 'helm'
  name: string
  version: string
  description: string
  home: string
  size: number
  filePath: string
}

// Copyright 2022 Linka Cloud  All rights reserved.
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

import { CheckBoxOutlineBlank } from '@mui/icons-material'
import React from 'react'
import { RepositoryType } from '../api/repository'
import PackagesPage from '../Components/Packages/PackagesPage'
import { MainRoutesRegistry } from '../routes'

MainRoutesRegistry['packages'] = {
  path: '/:repo',
  component: <PackagesPage />,
  icon: <CheckBoxOutlineBlank />,
  priority: 200,
  public: false,
  label: 'Packages',
  show: false,
  navigate: ([type, repo]: [type: RepositoryType, repo: string]) => `/${type}/${repo}`,
}

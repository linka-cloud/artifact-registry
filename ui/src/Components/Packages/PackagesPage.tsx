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

import { Box, Card, CardContent, CardHeader, MenuItem, Stack, Typography } from '@mui/material'
import React, { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { Package, RepositoryType, subRepositories, subRepositoryPackages } from '../../api/repository'
import { useAPI } from '../../api/useAPI'
import { useAsync } from '../../hooks'
import { packageTypeIcon } from '../../icons/packageTypeIcon'
import { useSnackbar } from '../../snackbar'

import { defaultPadding, defaultSpacing } from '../../theme/theme'
import { Loading } from '../Loading'
import { MultiLangCode, MultiLangCodeItem } from '../MultiLangCode'
import { SimpleSelect } from '../SimpleSelect'
import { PackageCard } from './PackageCard'

const PackagesPage = () => {
  const api = useAPI()
  const [packages, setPackages] = useState<Package[]>([])
  const {errorSnackbar} = useSnackbar()
  const [loading, setLoading] = useState(false)
  const { repo: _repo } = useParams<{ repo: string }>()
  const [repo, type] = decodeURIComponent(_repo!!).split(':') as [string, RepositoryType]
  const subs = subRepositories(packages, type)
  const [sub, setSub] = useState<string>(subs.length > 0 ? subs[0] : '')
  useEffect(() => {
    setSub(subs.length > 0 ? subs[0] : '')
  }, [packages])
  console.log(sub, packages, subs)
  useAsync(async () => {
    setLoading(true)
    const [packages, error] = await api.packages(repo!!, type!!)
    console.log(packages)
    setPackages(packages)
    if (error) {
      errorSnackbar(error)
    }
    setLoading(false)
  }, [])
  return loading ? <Loading /> : (
    <Stack padding={defaultPadding} spacing={defaultSpacing}>
      <Stack direction='row'>
        <Box flex={1}>
          <Typography variant='h6'>Repository</Typography>
        </Box>
        {
          !!subs.length && (
            <SimpleSelect
              sx={{ m: 1, minWidth: 120 }}
              value={sub}
              onChange={e => setSub(e.target.value as string)}
            >
              {
                subs.map(p => (
                  <MenuItem key={p} value={p}>{p}</MenuItem>
                ))
              }
            </SimpleSelect>
          )
        }
      </Stack>
      <Stack>
        <Card>
          <CardHeader avatar={packageTypeIcon(type)} title={repo + (sub ? '/' + sub : '')}
                      titleTypographyProps={{ variant: 'h5' }} />
          <CardContent>
            <Typography variant='body2'>Run this command to setup the repository on your machine :</Typography>
            <MultiLangCode key='lang'>
              <MultiLangCodeItem
                label='lkar'
                code={`lkar ${type} setup ${window.location.host}/${repo} ${(sub ? `${sub.split("/")[0]} ${sub.split("/")[1]}` : '')}`}
                language='bash'
              />
              <MultiLangCodeItem
                label='curl'
                code={`curl --user "$USER:$PASSWORD" ${window.location.protocol}//${window.location.host}/${type}/${repo + (sub ? '/' + sub : '')}/setup | sudo sh`}
                language='bash'
              />
            </MultiLangCode>
          </CardContent>
        </Card>
      </Stack>
      <Stack>
        <Box flex={1}>
          <Typography variant='h6'>Packages</Typography>
        </Box>
      </Stack>
      <Stack>
        {
          subRepositoryPackages(packages, type, sub).map(({ name, ...rest }) => (
            <PackageCard key={name} package={{ name, ...rest }} />
          ))
        }
      </Stack>
    </Stack>
  )
}

export default PackagesPage

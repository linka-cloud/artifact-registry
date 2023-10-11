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

import { Box, Card, CardContent, CardHeader, Stack, Typography } from '@mui/material'
import React, { useState } from 'react'
import { Repository } from '../../api/repository'
import { useAPI } from '../../api/useAPI'
import { lkar } from '../../cli/cli'
import { useAsync } from '../../hooks'
import { useSnackbar } from '../../snackbar'

import { defaultPadding, defaultSpacing } from '../../theme/theme'
import { Loading } from '../Loading'
import { MultiLangCode, MultiLangCodeItem } from '../MultiLangCode'
import { RepositoryCard } from './RepositoryCard'

const HomePage = () => {
  const api = useAPI()
  const [repos, setRepos] = useState<Repository[]>([])
  const { errorSnackbar } = useSnackbar()
  const [loading, setLoading] = useState(false)
  useAsync(async () => {
    setLoading(true)
    const [repos, error] = await api.repositories(api.baseRepo)
    setRepos(repos)
    if (error) {
      errorSnackbar(error.message)
    }
    setLoading(false)
  }, [])
  return loading ? <Loading /> : (
    <Stack padding={defaultPadding} spacing={defaultSpacing}>
      <Stack direction='row'>
        <Box flex={1}>
          <Typography variant='h6'>Repositories</Typography>
        </Box>
      </Stack>
      <Stack>
        <Card>
          {
            api.baseRepo
            && <CardHeader
              title={api.baseRepo}
              titleTypographyProps={{ variant: 'h5' }}
            />
          }
          <CardContent>
            <MultiLangCode storageKey='lang' title='Run this command to log into the repository on your machine :'>
              <MultiLangCodeItem
                label='lkar'
                code={lkar.login(api.baseRepo)}
                hiddenCode={lkar.login(api.baseRepo, api.credentials)}
                language='bash'
              />
            </MultiLangCode>
          </CardContent>
        </Card>
      </Stack>
      <Stack>
        {
          repos.map(({ name, type, ...rest }) => (
            <RepositoryCard key={`${name}:${type}`} repository={{ name, type, ...rest }} />
          ))
        }
      </Stack>
    </Stack>
  )
}

export default HomePage

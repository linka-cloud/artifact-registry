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

import { ExpandLessOutlined, ExpandMoreOutlined } from '@mui/icons-material'
import { Card, CardActions, CardContent, CardHeader, Collapse, IconButton, Typography } from '@mui/material'
import React, { useState } from 'react'
import { RepositoryType } from '../../api/repository'
import { useAPI } from '../../api/useAPI'
import { curl, lkar } from '../../cli/cli'
import { packageTypeIcon } from '../../icons/packageTypeIcon'
import { MultiLangCode, MultiLangCodeItem } from '../MultiLangCode'

export interface RepoCardProps {
  type: RepositoryType
  repo?: string;
  sub?: string;
}

export const RepoCard = ({type, repo = '', sub}: RepoCardProps) => {
  const {credentials} = useAPI();
  const [expanded, setExpanded] = useState(false)
  return (
    <Card>
      <CardHeader avatar={packageTypeIcon(type)} title={repo ? (repo + (sub ? '/' + sub : '')) : sub ? sub : type}
                  titleTypographyProps={{ variant: 'h5' }} />
      <CardContent sx={{pt: 0, pb: 0}}>
        <Typography variant='h6'>Setup</Typography>
        <MultiLangCode storageKey='lang' title='Run this command to setup the repository on your machine :'>
          <MultiLangCodeItem
            label='lkar'
            code={lkar.setup(type, repo, sub)}
            hiddenCode={lkar.setup(type, repo, sub, credentials)}
            language='bash'
          />
          <MultiLangCodeItem
            label='curl'
            code={curl.setup(type, repo, sub)}
            hiddenCode={curl.setup(type, repo, sub, credentials)}
            language='bash'
          />
        </MultiLangCode>
      </CardContent>
      <CardActions sx={{justifyContent: 'end'}}>
        <IconButton onClick={() => setExpanded(!expanded)}>
          {expanded ? <ExpandLessOutlined /> : <ExpandMoreOutlined /> }
        </IconButton>
      </CardActions>
      <Collapse in={expanded} timeout="auto" unmountOnExit>
        <CardContent sx={{pt: 0}}>
          <Typography variant='h6'>Push</Typography>
          <MultiLangCode storageKey='lang' title='Run this command on your machine to push a package to the repository :'>
            <MultiLangCodeItem
              label='lkar'
              code={lkar.push(type, repo, sub)}
              hiddenCode={lkar.push(type, repo, sub, credentials)}
              language='bash'
            />
            <MultiLangCodeItem
              label='curl'
              code={curl.push(type, repo, sub)}
              hiddenCode={curl.push(type, repo, sub, credentials)}
              language='bash'
            />
          </MultiLangCode>
        </CardContent>
      </Collapse>
    </Card>
  )
}

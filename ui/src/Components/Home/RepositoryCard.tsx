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

import { CheckBoxOutlineBlank, InsertDriveFileOutlined, UpdateOutlined } from '@mui/icons-material'
import { Card, CardContent, CardHeader, Chip, Link, Stack, Typography } from '@mui/material'
import moment from 'moment/moment'
import React from 'react'
import { Repository, RepositoryType } from '../../api/repository'
import { KubernetesIcon } from '../../icons/KubernetesIcon'
import { LinuxIcon } from '../../icons/LinuxIcon'
import { packageTypeIcon } from '../../icons/packageTypeIcon'
import { defaultPadding } from '../../theme/theme'
import { humanSize } from '../../utils'

export interface RepositoryCardProps {
  repository: Repository
}

export const RepositoryCard = ({ repository: { name, type, size, lastUpdated, metadata, packages } }: RepositoryCardProps) => (
  <Card key={name ? `${name}:${type}` : type} component={Link}
        href={'/' + (name ? encodeURIComponent(`${name.split('/').slice(1).join('/')}:${type}`) : type)}>
    <CardHeader
      avatar={packageTypeIcon(type)}
      title={name || type} subheader={humanSize(size)}
      action={(
        <Stack direction='row' padding={defaultPadding} alignItems='center'>
          <UpdateOutlined />
          <Typography
            sx={{ marginLeft: '4px !important' }}
            variant='body2'>{moment(lastUpdated.getTime()).fromNow()}</Typography>
        </Stack>
      )} />
    <CardContent>
      <Stack direction='row'>
        <Chip icon={type === RepositoryType.HELM ? <KubernetesIcon /> : <LinuxIcon sx={{ padding: 0.15 }} />}
              label={type === RepositoryType.HELM ? 'kubernetes' : 'linux'} />
        <Chip icon={<InsertDriveFileOutlined />} label={type} />
        <Chip icon={<CheckBoxOutlineBlank />} label={`${packages.count} packages`} />
      </Stack>
    </CardContent>
  </Card>
)

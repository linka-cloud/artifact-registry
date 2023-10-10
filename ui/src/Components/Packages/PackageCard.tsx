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

import { Balance, MemoryOutlined, UpdateOutlined } from '@mui/icons-material'
import { Card, CardContent, CardHeader, Chip, Stack, Typography } from '@mui/material'
import React from 'react'
import { Package } from '../../api/repository'
import { LinuxIcon } from '../../icons/LinuxIcon'
import { packageTypeIcon } from '../../icons/packageTypeIcon'
import { defaultPadding, defaultSpacing } from '../../theme/theme'
import { humanSize } from '../../utils'
import { ExternalLink } from '../ExternalLink'

export interface PackageCardProps {
  package: Package
}

export const PackageCard = ({ package: {name, type, size, version, architecture, license, projectURL, description} }: PackageCardProps) => (
  <Card>
    <CardHeader
      avatar={packageTypeIcon(type)}
      title={name} subheader={humanSize(size)}
      action={(
        <Stack direction='row' padding={defaultPadding}>
          <UpdateOutlined />
          <Typography
            sx={{ marginLeft: '4px !important' }}
            variant='body2'>{version}</Typography>
        </Stack>
      )} />
    <CardContent sx={{ paddingTop: 0 }}>
      <Stack direction='row' marginTop={0}>
        <Chip icon={<LinuxIcon />} label='linux' />
        <Chip icon={<MemoryOutlined />} label={architecture} />
        {license && <Chip icon={<Balance />} label={license} />}
      </Stack>
      <Stack sx={{ marginTop: defaultSpacing }}>
        <Typography variant='body2' fontStyle='italic'>{description}</Typography>
        {projectURL && <ExternalLink href={projectURL}>{projectURL}</ExternalLink>}
      </Stack>
    </CardContent>
  </Card>
)

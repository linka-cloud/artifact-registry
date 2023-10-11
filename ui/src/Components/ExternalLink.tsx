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

import { BoxProps } from '@mui/material'
import Box from '@mui/material/Box'
import { LinkProps } from '@mui/material/Link'
import { styled } from '@mui/material/styles'
import { forwardRef } from 'react'

const Link = forwardRef<{}, Omit<BoxProps, 'component' | 'target'> & LinkProps>((props, ref) => (
  <Box
    {...props} target='_blank'
    component='a'
    ref={ref}
  />
))
export const ExternalLink = styled(Link, { shouldForwardProp: prop => prop != 'component' })(({ theme }) => ({
  color: theme.palette.text.secondary,
  textDecoration: 'none',
  '&:hover': {
    color: theme.palette.primary.main,
  },
}))

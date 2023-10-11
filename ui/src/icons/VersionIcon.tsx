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

import { SvgIcon, SvgIconProps, useTheme } from '@mui/material'

export const VersionIcon = (props: SvgIconProps) => {
  const theme = useTheme()
  return (
    <SvgIcon {...props} viewBox='0 0 21 21'>
      <g fill='none' fillRule='evenodd' stroke={theme.palette.text.primary}
         strokeLinecap='round' strokeLinejoin='round' transform='translate(2 4)'>
        <path d='m.5 8.5 8 4 8.017-4'></path>
        <path d='m.5 4.657 8.008 3.843 8.009-3.843-8.009-4.157z'></path>
      </g>
    </SvgIcon>
  )
}

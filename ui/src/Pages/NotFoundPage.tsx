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

import { Stack } from '@mui/material'
import { TypingMessage } from '../Components/TypingMessage'
import { NotFoundIcon } from '../icons/NotFoundIcon'
import { MainRoutesRegistry } from '../routes'

export const NotFoundPage = () => (
  <Stack alignItems='center' justifyContent='center' paddingTop={8}>
    <Stack>
      <NotFoundIcon
        sx={{ fontSize: '24rem', padding: theme => theme.spacing(8), color: theme => theme.palette.text.secondary }} />
    </Stack>
    <TypingMessage
      text={[
        '404, Page Not Found.',
        'Sorry... 🙀',
      ]}
      speed={100}
      eraseDelay={3000}
      eraseSpeed={50}
      typingDelay={1000}
    />
  </Stack>
)

MainRoutesRegistry['notFound'] = {
  path: '/*',
  component: <NotFoundPage />,
  priority: 0,
  public: false,
  show: false,
  navigate: () => '/404',
}


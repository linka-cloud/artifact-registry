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

import { PowerSettingsNewOutlined } from '@mui/icons-material'
import { useNavigate } from 'react-router-dom'
import { useAPI } from '../api/useAPI'
import { useAsyncOnce } from '../hooks'
import { MainRoutesRegistry } from '../routes'

export const LogoutPage = () => {
  const navigate = useNavigate()
  const { logout } = useAPI()
  useAsyncOnce(async () => {
    await logout()
    navigate(MainRoutesRegistry.login.navigate(), { replace: true })
  })
  return null
}

MainRoutesRegistry['logout'] = {
  label: 'Logout',
  path: '/logout',
  component: <LogoutPage />,
  icon: <PowerSettingsNewOutlined />,
  priority: 0,
  public: false,
  show: false,
  bottomEnd: true,
  navigate: () => '/logout',
}

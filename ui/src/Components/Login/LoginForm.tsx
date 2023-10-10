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

import { LockOutlined, PersonOutlineOutlined, VisibilityOffOutlined, VisibilityOutlined } from '@mui/icons-material'
import { Button, IconButton, InputAdornment, Stack } from '@mui/material'
import { Formik } from 'formik'
import { useState } from 'react'
import { Credentials, credentialsSchema } from '../../api/schemas/login'
import { defaultSpacing } from '../../theme/theme'
import { FTextField } from '../Form'

export const LoginForm = ({ onLogin }: { onLogin: (credentials: Credentials) => Promise<void> }) => {
  const [showPassword, setShowPassword] = useState(false)
  return (
    <Formik<Credentials> initialValues={{ user: '', password: '' }} validationSchema={credentialsSchema}
                         onSubmit={onLogin}>
      {({ isSubmitting, handleReset, handleSubmit }) => (
        <Stack component='form' onReset={handleReset} onSubmit={handleSubmit}>
          <Stack>
            <FTextField
              name='user'
              label='Username'
              autoFocus={true}
              InputProps={{
                startAdornment: (
                  <InputAdornment position='start'>
                    <PersonOutlineOutlined />
                  </InputAdornment>
                ),
              }}
            />
            <FTextField
              name='password'
              label='Password'
              type={showPassword ? 'text' : 'password'}
              InputProps={{
                startAdornment: (
                  <InputAdornment position='start'>
                    <LockOutlined />
                  </InputAdornment>
                ),
                endAdornment: (
                  <InputAdornment position='end'>
                    <IconButton aria-label='toggle password visibility' onClick={() => setShowPassword(!showPassword)}>
                      {showPassword ? <VisibilityOffOutlined /> : <VisibilityOutlined />}
                    </IconButton>
                  </InputAdornment>
                ),
              }}
            />
          </Stack>
          <Stack direction='row' spacing={defaultSpacing} justifyContent='flex-end'>
            <Button disabled={isSubmitting} type='submit'>
              Login
            </Button>
          </Stack>
        </Stack>
      )}
    </Formik>
  )
}

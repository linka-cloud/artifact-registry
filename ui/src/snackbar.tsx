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

import { Button, Stack } from '@mui/material'
import { styled } from '@mui/material/styles'
import {
  OptionsObject,
  SnackbarKey,
  SnackbarMessage,
  SnackbarProvider as NSnackbarProvider,
  SnackbarProviderProps,
  useSnackbar as useNSnackbar,
} from 'notistack'
import React, { useRef } from 'react'
import { useIsMobile } from './hooks'

// import "./snackbar.css";

const StyledSnackbarProvider = styled(NSnackbarProvider)(({ theme }) => ({
  '&.SnackbarItem-variantSuccess': {
    color: theme.palette.success.main,
    borderColor: theme.palette.success.main,
    border: '2px solid',
    background: theme.palette.background.paper,
    boxShadow: 'none',
    '& button': {
      color: theme.palette.success.main,
    },
  },
  '&.SnackbarItem-variantError': {
    color: theme.palette.error.main,
    borderColor: theme.palette.error.main,
    border: '2px solid',
    background: theme.palette.background.paper,
    boxShadow: 'none',
    '& button': {
      color: theme.palette.error.main,
    },
  },
  '&.SnackbarItem-variantInfo': {
    color: theme.palette.info.main,
    borderColor: theme.palette.info.main,
    border: '2px solid',
    background: theme.palette.background.paper,
    boxShadow: 'none',
    '& button': {
      color: theme.palette.info.main,
    },
  },
  '&.SnackbarItem-variantWarning': {
    color: theme.palette.warning.main,
    borderColor: theme.palette.warning.main,
    border: '2px solid',
    background: theme.palette.background.paper,
    boxShadow: 'none',
    '& button': {
      color: theme.palette.warning.main,
    },
  },
}))

export const SnackbarProvider = ({ children, ...rest }: SnackbarProviderProps) => {
  const isMobile = useIsMobile()
  const notistackRef = useRef<NSnackbarProvider>(null)
  const dismiss = (key: SnackbarKey) => notistackRef.current?.closeSnackbar(key)
  return (
    <StyledSnackbarProvider
      anchorOrigin={{ vertical: isMobile ? 'bottom' : 'top', horizontal: isMobile ? 'center' : 'right' }}
      maxSnack={isMobile ? 1 : 3}
      preventDuplicate
      {...rest}
      ref={notistackRef}
      action={(key) => <Button onClick={() => dismiss(key)}>Dismiss</Button>}
    >
      {children}
    </StyledSnackbarProvider>
  )
}

export type Snackbar = (message: SnackbarMessage, options?: OptionsObject) => SnackbarKey

export const useSnackbar = (defaultOptions?: OptionsObject) => {
  const { enqueueSnackbar, closeSnackbar } = useNSnackbar()
  return {
    enqueueSnackbar,
    closeSnackbar,
    defaultSnackbar: (message: SnackbarMessage, options?: OptionsObject): SnackbarKey =>
      enqueueSnackbar(message, {
        ...defaultOptions,
        ...options,
        variant: 'default',
      }),
    infoSnackbar: (message: SnackbarMessage, options?: OptionsObject): SnackbarKey =>
      enqueueSnackbar(message, {
        ...defaultOptions,
        ...options,
        variant: 'info',
      }),
    successSnackbar: (message: SnackbarMessage, options?: OptionsObject): SnackbarKey =>
      enqueueSnackbar(message, {
        ...defaultOptions,
        ...options,
        variant: 'success',
      }),
    warnSnackbar: (message: SnackbarMessage, options?: OptionsObject): SnackbarKey =>
      enqueueSnackbar(message, {
        ...defaultOptions,
        ...options,
        variant: 'warning',
      }),
    errorSnackbar: (message: SnackbarMessage, options?: OptionsObject): SnackbarKey =>
      enqueueSnackbar(message, {
        ...defaultOptions,
        ...options,
        variant: 'error',
      }),
  }
}

type ConfirmParams<T> = {
  confirm?: string
  success?: string | ((v: T) => string)
  error?: string | ((e: any) => string)
  action: () => Promise<T>
}

export const useConfirmSnackbar = <T, >() => {
  const { successSnackbar, errorSnackbar, infoSnackbar, closeSnackbar } = useSnackbar()
  return (params: ConfirmParams<T>) => new Promise<void>((resolve) => {
    const { confirm = 'Are you sure ?', success = 'Done.', error = (e: any) => e.message, action } = params
    const doWork = async (key: SnackbarKey) => {
      closeSnackbar(key)
      try {
        const v = await action()
        successSnackbar(typeof success === 'function' ? success(v) : success)
        resolve()
      } catch (e: any) {
        errorSnackbar(typeof error === 'function' ? error(e.message) : error)
        resolve()
      }
    }
    const dismiss = async (key: SnackbarKey) => {
      closeSnackbar(key)
      resolve()
    }
    infoSnackbar(confirm, {
      onClose: () => resolve(),
      action: (key) => (
        <Stack direction='row'>
          <Button onClick={async () => dismiss(key)}>Cancel</Button>
          <Button onClick={async () => doWork(key)}>Confirm</Button>
        </Stack>
      ),
    })
  })

}

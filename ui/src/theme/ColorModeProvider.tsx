// Copyright 2021 Linka Cloud  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { CssBaseline, PaletteMode, ThemeProvider, useMediaQuery } from '@mui/material'
import React, { PropsWithChildren, useContext, useEffect, useState } from 'react'
import { Helmet, HelmetProvider } from 'react-helmet-async'
import { configureTheme } from './configureTheme'
import { UiMode } from './theme'

const LOCAL_STORAGE_KEY = 'ui-mode'

const modeFromLocalStorage = (): UiMode => {
  const mode = localStorage.getItem(LOCAL_STORAGE_KEY)
  switch (mode) {
    case 'light':
    case 'dark':
    case undefined:
      return mode
    case '':
    default:
      return undefined
  }
}

interface ModeContext {
  mode: UiMode
  setMode: (mode: UiMode) => void
}

const modeContext = React.createContext<ModeContext>({
  mode: undefined,
  setMode: () => {
  },
})

export const useUiMode = () => {
  const ctx = useContext(modeContext)
  return { mode: ctx.mode, setMode: ctx.setMode }
}

export const ColorModeThemeProvider = ({ children }: PropsWithChildren<any>) => {
  const prefersDarkMode = useMediaQuery('(prefers-color-scheme: dark)')

  const restoredMode = modeFromLocalStorage()
  // console.log('mode: restore', restoredMode || 'system')
  const [storedMode, _setStoredMode] = useState<UiMode>(restoredMode)
  const setStoredMode = (mode: UiMode) => {
    localStorage.setItem(LOCAL_STORAGE_KEY, mode || '')
    _setStoredMode(mode)
  }

  const [appliedMode, setAppliedMode] = useState<PaletteMode>(restoredMode || (prefersDarkMode ? 'dark' : 'light'))
  // console.log('mode: apply', restoredMode || 'system')
  const setMode = (mode: UiMode) => {
    const m = mode || (prefersDarkMode ? 'dark' : 'light')
    setStoredMode(mode)
    setAppliedMode(m)
  }

  useEffect(() => {
    if (storedMode) {
      return
    }
    setMode(undefined)
  }, [prefersDarkMode])

  const theme = React.useMemo(() => configureTheme(appliedMode), [appliedMode])
  return (
    <modeContext.Provider value={{ mode: appliedMode, setMode: setMode }}>
      <HelmetProvider>
        <ThemeProvider theme={theme}>
          <Helmet>
            <meta name='theme-color' content={appliedMode === 'dark' ? '#000000' : '#FFFFFF'} />
          </Helmet>
          <CssBaseline />
          {children}
        </ThemeProvider>
      </HelmetProvider>
    </modeContext.Provider>
  )
}

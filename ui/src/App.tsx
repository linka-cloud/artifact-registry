import React, { Suspense } from 'react'

import { BrowserRouter } from 'react-router-dom'
import { APIProvider } from './api/useAPI'
import './App.css'
import { ErrorBoundary } from './Components/ErrorBoundary'
import { Layout } from './Components/Layout'
import { Loading } from './Components/Loading'
import { MultiLangCodeProvider } from './Components/MultiLangCode'

import './Pages'
import { Router } from './Router'
import { SnackbarProvider } from './snackbar'
import { ColorModeThemeProvider } from './theme/ColorModeProvider'

const App = () => (
  <ColorModeThemeProvider>
    <APIProvider>
      <MultiLangCodeProvider storageKey="lkar" value="lkar">
        <SnackbarProvider>
          <BrowserRouter basename="/ui">
            <Layout>
              <ErrorBoundary>
                <Suspense fallback={<Loading />}>
                  <Router />
                </Suspense>
              </ErrorBoundary>
            </Layout>
          </BrowserRouter>
        </SnackbarProvider>
      </MultiLangCodeProvider>
    </APIProvider>
  </ColorModeThemeProvider>
)

export default App

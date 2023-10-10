// Copyright 2021 Linka Cloud  All rights reserved.
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

import React, { PropsWithChildren, useCallback } from 'react'
import { Navigate, Route, Routes, useLocation, useParams } from 'react-router-dom'
import { useAPI } from './api/useAPI'
import { Loading } from './Components/Loading'
import { MainRoutesRegistry, RouteDefinition, RoutesRegistry } from './routes'

export const Router = () => {
  const makeRoutes = useCallback(
    (reg?: RoutesRegistry, parent?: RouteDefinition) =>
      Object.values(reg ?? {}).map((r, i) => {
        return r.public ? (
          <Route key={i} path={parent ? parent.path + '/' + r.path : r.path} element={r.component}>
            {makeRoutes(r.subRoutes, r)}
          </Route>
        ) : (
          <Route key={i} path={parent ? parent.path + '/' + r.path : r.path}
                 element={<ProtectedRoute>{r.component}</ProtectedRoute>}>
            {makeRoutes(r.subRoutes, r)}
          </Route>
        )
      }),
    [],
  )
  return <Routes>{makeRoutes(MainRoutesRegistry)}</Routes>
}

export const ProtectedRoute = ({ children }: PropsWithChildren<any>) => {
  const from = useLocation()
  const fromParams = useParams()
  const { authenticated } = useAPI()
  if (authenticated === undefined) {
    return <Loading />
  }
  if (!authenticated) {
    return <Navigate to={MainRoutesRegistry.login.navigate()} state={{ from, fromParams }} replace />
  }
  return children
}
